// Package rework 统一承接 bug 修复与失败升级的返工阶段。
// 当 execute 阶段任务失败超限时，由 orchestrator 触发 rework stage。
package rework

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/resourcepath"
	"easymvp/utility/snowflake"
)

var reworkJsonBlockRe = regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")

func marshalHandoffPayload(payload interface{}) string {
	switch v := payload.(type) {
	case nil:
		return ""
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return ""
		}
		if json.Valid([]byte(text)) {
			return text
		}
		data, err := json.Marshal(map[string]string{"content": text})
		if err != nil {
			return ""
		}
		return string(data)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(data)
	}
}

func classifyOriginalTaskStatusForRework(status string) (allowReset bool, alreadyRecovered bool) {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case domainTask.StatusFailed, domainTask.StatusEscalated:
		return true, false
	case domainTask.StatusPending, domainTask.StatusRunning, domainTask.StatusCompleted:
		return false, true
	default:
		return false, false
	}
}

// StageCompleter 阶段操作回调（避免循环依赖）。
type StageCompleter interface {
	CompleteStage(ctx context.Context, stageRunID int64) error
	FailStage(ctx context.Context, stageRunID int64, reason string) error
}

// ExecuteTriggerFn 触发执行阶段回调。
type ExecuteTriggerFn func(ctx context.Context, workflowRunID, planVersionID int64) error

// AcceptTriggerFn 触发验收阶段回调（返工完成后回验收）。
type AcceptTriggerFn func(ctx context.Context, workflowRunID int64) error

// Service 返工阶段服务。
type Service struct {
	stageCompleter StageCompleter
	executeTrigger ExecuteTriggerFn
	acceptTrigger  AcceptTriggerFn
	stageRunRepo   *repo.StageRunRepo
	taskRepo       *repo.DomainTaskRepo
	handoffRepo    *repo.HandoffRecordRepo
	convRepo       *repo.ConversationRepo
	messageRepo    *repo.MessageRepo
}

// NewService 创建返工阶段服务。
func NewService() *Service {
	return &Service{
		stageRunRepo: repo.NewStageRunRepo(),
		taskRepo:     repo.NewDomainTaskRepo(),
		handoffRepo:  repo.NewHandoffRecordRepo(),
		convRepo:     repo.NewConversationRepo(),
		messageRepo:  repo.NewMessageRepo(),
	}
}

// SetStageCompleter 注册阶段完成回调。
func (s *Service) SetStageCompleter(sc StageCompleter) { s.stageCompleter = sc }

// SetExecuteTrigger 注册执行阶段触发回调。
func (s *Service) SetExecuteTrigger(fn ExecuteTriggerFn) { s.executeTrigger = fn }

// SetAcceptTrigger 注册验收阶段触发回调（返工完成后回验收）。
func (s *Service) SetAcceptTrigger(fn AcceptTriggerFn) { s.acceptTrigger = fn }

// HandleRework 处理返工流程。
// 接收失败的 domain_task，创建架构师分析任务，分析完成后回写原任务。
// sourceStage 标记返工来源阶段（"execute"/"accept"），决定返工完成后回流目标。
func (s *Service) HandleRework(ctx context.Context, stageRunID int64, failedTaskID int64) error {
	return s.HandleReworkWithSource(ctx, stageRunID, failedTaskID, "execute")
}

// HandleReworkWithSource 处理返工流程（指定来源阶段）。
func (s *Service) HandleReworkWithSource(ctx context.Context, stageRunID int64, failedTaskID int64, sourceStage string) error {
	// 1. 查询 stage_run 和 workflow_run 信息
	stageRun, err := s.stageRunRepo.GetByIDMap(ctx, stageRunID, "workflow_run_id")
	if err != nil || len(stageRun) == 0 {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := g.NewVar(stageRun["workflow_run_id"]).Int64()

	// 2. 查询失败任务详情
	failedTask, err := s.taskRepo.GetRecordByID(ctx, failedTaskID)
	if err != nil || failedTask == nil || failedTask.IsEmpty() {
		return fmt.Errorf("failed domain_task(%d) 不存在", failedTaskID)
	}

	// 3. 检查返工轮次是否超限
	maxRounds := engine.GetConfigInt(ctx, "failure_handoff.max_rounds", "engine.failureHandoff.maxRounds", 3)
	reworkCount, countErr := s.handoffRepo.CountByWorkflowAndFromTask(ctx, workflowRunID, failedTaskID)
	if countErr != nil {
		g.Log().Warningf(ctx, "[ReworkStage] 查询返工次数失败: wfRun=%d task=%d err=%v", workflowRunID, failedTaskID, countErr)
	}
	if reworkCount >= maxRounds {
		reason := fmt.Sprintf("任务 %d 返工已达上限 %d 次", failedTaskID, maxRounds)
		g.Log().Errorf(ctx, "[ReworkStage] %s, 标记阶段失败", reason)
		if s.stageCompleter != nil {
			_ = s.stageCompleter.FailStage(ctx, stageRunID, reason)
		}
		return nil
	}

	// 5. 创建架构师分析 domain_task
	analysisTaskID := int64(snowflake.Generate())
	rootTaskID := failedTask["root_task_id"].Int64()
	if rootTaskID == 0 {
		rootTaskID = failedTaskID
	}

	now := gtime.Now()
	scope := repo.GetProjectScopeByWorkflowRun(ctx, workflowRunID)
	err = s.taskRepo.Insert(ctx, g.Map{
		"id":              analysisTaskID,
		"workflow_run_id": workflowRunID,
		"stage_run_id":    stageRunID,
		"task_kind":       "failure_analysis",
		"name":            fmt.Sprintf("失败分析: %s", failedTask["name"].String()),
		"description": fmt.Sprintf(
			"请分析任务失败的具体原因，并给出可直接回写到原任务的修复方案。\n"+
				"必须明确指出失败是由哪条命令、哪条路径、哪个资源冲突、哪次越界修改或哪个依赖缺失引起；不要只给泛泛结论。\n\n"+
				"关联任务ID：%d\n角色：%s\n错误信息：\n%s\n\n"+
				"原任务名称：%s\n原任务描述：\n%s\n\n"+
				"如果失败原因是 token limit、上下文过大、任务范围过宽或单任务文件集过大，禁止继续把原任务揉成一个更长描述；必须重新设计为 N 个更小的同批次子任务，并显式给出 depends_on 关系。\n"+
				"修复方案不得新增 .gitignore 这类仓库级控制文件，除非原任务 affected_resources 已明确包含该文件。\n"+
				"修复方案必须严格围绕当前任务，不能改变原任务目标、不能脱离原有任务设定、不能把范围扩展到无关任务。\n"+
				"推荐输出多任务重设计 JSON：\n"+
				"{\"plan_meta\":{\"plan_id\":\"repair-保持一致\",\"declared_total\":3},\"tasks\":[{\"name\":\"子任务1\",\"description\":\"更小可执行任务\",\"role_type\":\"implementer\",\"role_level\":\"pro\",\"batch_no\":原任务批次,\"affected_resources\":[\"路径\"],\"depends_on\":[]},{\"name\":\"子任务2\",\"description\":\"后续子任务\",\"role_type\":\"implementer\",\"role_level\":\"pro\",\"batch_no\":原任务批次,\"affected_resources\":[\"路径\"],\"depends_on\":[\"子任务1\"]}]}\n\n"+
				"推荐输出任务级修复 JSON：\n"+
				"{\"task_repair\":{\"task_name\":%q,\"description\":\"修订后的任务描述\",\"affected_resources\":[\"路径\"],\"reason\":\"修订原因，必须写清具体失败原因\"}}\n\n"+
				"兼容旧格式：也可以直接输出\n"+
				"{\"description\":\"修订后的任务描述\",\"affected_resources\":[\"路径\"],\"reason\":\"修订原因，必须写清具体失败原因\"}\n"+
				"或输出只包含当前任务 %q 的 {\"task_patches\":[...]}。\n"+
				"无论使用哪种格式，都只能修当前任务，不得偏离项目既定方案。",
			failedTaskID, failedTask["role_type"].String(),
			failedTask["result"].String(),
			failedTask["name"].String(), failedTask["description"].String(), failedTask["name"].String(), failedTask["name"].String(),
		),
		"role_type":      "architect",
		"role_level":     "max",
		"execution_mode": "chat",
		"status":         domainTask.StatusPending,
		"source_task_id": failedTaskID,
		"root_task_id":   rootTaskID,
		"batch_no":       0, // 高优先级
		"sort":           0,
		"retry_count":    0,
		"created_by":     scope.CreatedBy,
		"dept_id":        scope.DeptID,
		"created_at":     now,
		"updated_at":     now,
	})
	if err != nil {
		return fmt.Errorf("创建分析任务失败: %w", err)
	}

	// 6. 写 handoff_record（payload 中记录来源阶段，用于返工完成后决定回流目标）
	handoffID := int64(snowflake.Generate())
	payloadJSON := marshalHandoffPayload(map[string]string{"source_stage": sourceStage})
	if payloadJSON == "" {
		g.Log().Warningf(ctx, "[ReworkStage] 序列化 handoff payload 失败: sourceStage=%s", sourceStage)
	}
	if insErr := s.handoffRepo.Create(ctx, g.Map{
		"id":              handoffID,
		"workflow_run_id": workflowRunID,
		"from_task_id":    failedTaskID,
		"to_task_id":      analysisTaskID,
		"handoff_type":    "failure_escalation",
		"reason":          failedTask["result"].String(),
		"payload":         payloadJSON,
		"created_at":      now,
	}); insErr != nil {
		g.Log().Errorf(ctx, "[ReworkStage] 写入 handoff_record 失败: wfRun=%d task=%d err=%v", workflowRunID, failedTaskID, insErr)
	}

	g.Log().Infof(ctx, "[ReworkStage] 创建分析任务: stageRunID=%d failedTask=%d analysisTask=%d round=%d/%d",
		stageRunID, failedTaskID, analysisTaskID, reworkCount+1, maxRounds)

	return nil
}

// OnAnalysisCompleted 架构师分析任务完成后的回调。
// 解析分析结果，回写原失败任务，推进回 execute stage。
func (s *Service) OnAnalysisCompleted(ctx context.Context, stageRunID int64, analysisTaskID int64) error {
	// 1. 获取分析任务
	analysisTask, err := s.taskRepo.GetRecordByID(ctx, analysisTaskID)
	if err != nil || analysisTask == nil || analysisTask.IsEmpty() {
		return fmt.Errorf("分析任务(%d) 不存在", analysisTaskID)
	}

	// 2. 获取原失败任务 ID
	failedTaskID := analysisTask["source_task_id"].Int64()
	if failedTaskID == 0 {
		return fmt.Errorf("分析任务没有关联的来源任务")
	}

	failedTask, err := s.taskRepo.GetRecordByID(ctx, failedTaskID, "id", "name", "affected_resources")
	if err != nil || failedTask == nil || failedTask.IsEmpty() {
		return fmt.Errorf("原失败任务(%d) 不存在", failedTaskID)
	}

	// 3. 解析架构师修复方案
	resolution, err := s.resolveAnalysisResolution(ctx, analysisTask, failedTask)
	if err != nil {
		g.Log().Warningf(ctx, "[ReworkStage] 解析修复方案失败: task=%d err=%v", analysisTaskID, err)
		// 解析失败也要完成 rework stage
		if s.stageCompleter != nil {
			_ = s.stageCompleter.FailStage(ctx, stageRunID, "架构师修复方案解析失败: "+err.Error())
		}
		return nil
	}

	reconciledWithoutWrite := false
	handoffPayload := map[string]interface{}{}
	if resolution.SplitPlan != nil {
		reconciledWithoutWrite, handoffPayload, err = s.applySplitPlan(ctx, analysisTask, failedTaskID, resolution.SplitPlan)
		if err != nil {
			if s.stageCompleter != nil {
				_ = s.stageCompleter.FailStage(ctx, stageRunID, err.Error())
			}
			return nil
		}
	} else {
		reconciledWithoutWrite, err = s.applyTaskPatch(ctx, failedTaskID, failedTask, resolution.Patch)
		if err != nil {
			if s.stageCompleter != nil {
				_ = s.stageCompleter.FailStage(ctx, stageRunID, err.Error())
			}
			return nil
		}
		handoffPayload["content"] = resolution.Content
		handoffPayload["reason"] = resolution.Patch.Reason
	}

	// 5. 写 handoff_record（分析 → 原任务）
	handoffID := int64(snowflake.Generate())
	payloadJSON := marshalHandoffPayload(handoffPayload)
	if insErr := s.handoffRepo.Create(ctx, g.Map{
		"id":              handoffID,
		"workflow_run_id": analysisTask["workflow_run_id"].Int64(),
		"from_task_id":    analysisTaskID,
		"to_task_id":      failedTaskID,
		"handoff_type":    "rework",
		"reason":          resolveHandoffReason(resolution.Patch, resolution.SplitPlan),
		"payload":         payloadJSON,
		"created_at":      gtime.Now(),
	}); insErr != nil {
		g.Log().Errorf(ctx, "[ReworkStage] 写入 handoff_record(rework) 失败: analysis=%d target=%d err=%v", analysisTaskID, failedTaskID, insErr)
	}

	if reconciledWithoutWrite {
		g.Log().Infof(ctx, "[ReworkStage] 返工收敛完成: analysis=%d target=%d reason=%s", analysisTaskID, failedTaskID, resolveHandoffReason(resolution.Patch, resolution.SplitPlan))
	} else {
		g.Log().Infof(ctx, "[ReworkStage] 回写完成: analysis=%d → original=%d reason=%s", analysisTaskID, failedTaskID, resolveHandoffReason(resolution.Patch, resolution.SplitPlan))
	}

	workflowRunID := analysisTask["workflow_run_id"].Int64()
	if nextEscalatedTaskID, nextFound, nextErr := s.findNextEscalatedTask(ctx, workflowRunID, failedTaskID); nextErr != nil {
		return fmt.Errorf("查询后续 escalated 任务失败: %w", nextErr)
	} else if nextFound {
		sourceStage := s.resolveSourceStage(ctx, workflowRunID, nextEscalatedTaskID)
		g.Log().Infof(ctx, "[ReworkStage] 当前返工已收敛，继续接管剩余 escalated 任务: workflowRunID=%d nextTask=%d sourceStage=%s",
			workflowRunID, nextEscalatedTaskID, sourceStage)
		return s.HandleReworkWithSource(ctx, stageRunID, nextEscalatedTaskID, sourceStage)
	}

	// 6. 完成 rework stage
	if s.stageCompleter != nil {
		_ = s.stageCompleter.CompleteStage(ctx, stageRunID)
	}

	// 7. 按来源阶段决定回流目标
	sourceStage := s.resolveSourceStage(ctx, workflowRunID, failedTaskID)

	if sourceStage == "accept" && s.acceptTrigger != nil && workflowRunID > 0 {
		// 来自 accept 阶段的返工 → 回 accept 重新验收
		g.Log().Infof(ctx, "[ReworkStage] 返工完成，回流 accept: workflowRunID=%d", workflowRunID)
		if err := s.acceptTrigger(ctx, workflowRunID); err != nil {
			g.Log().Errorf(ctx, "[ReworkStage] 回流 accept 失败: workflowRunID=%d err=%v", workflowRunID, err)
		}
	} else if s.executeTrigger != nil && workflowRunID > 0 {
		// 默认（来自 execute 阶段的返工）→ 回 execute 重启调度
		g.Log().Infof(ctx, "[ReworkStage] 返工完成，回流 execute: workflowRunID=%d", workflowRunID)
		if err := s.executeTrigger(ctx, workflowRunID, 0); err != nil {
			g.Log().Errorf(ctx, "[ReworkStage] 重启调度失败: workflowRunID=%d err=%v", workflowRunID, err)
		}
	}

	return nil
}

// resolveSourceStage 从 handoff_record payload 中解析返工来源阶段。
func (s *Service) resolveSourceStage(ctx context.Context, workflowRunID, failedTaskID int64) string {
	payload, err := s.handoffRepo.GetLatestPayloadByWorkflowFromTaskType(ctx, workflowRunID, failedTaskID, "failure_escalation")
	if err != nil || strings.TrimSpace(payload) == "" {
		return "execute" // 默认回 execute
	}
	var data map[string]string
	if json.Unmarshal([]byte(payload), &data) == nil {
		if stage, ok := data["source_stage"]; ok && stage != "" {
			return stage
		}
	}
	return "execute"
}

func (s *Service) resolveAnalysisResolution(ctx context.Context, analysisTask gdb.Record, failedTask gdb.Record) (*analysisResolution, error) {
	taskName := failedTask["name"].String()
	candidates := make([]string, 0, 2)

	if raw := strings.TrimSpace(analysisTask["result"].String()); raw != "" {
		candidates = append(candidates, raw)
	}

	if latestReply := strings.TrimSpace(s.loadLatestAnalysisReply(ctx, analysisTask)); latestReply != "" {
		duplicate := false
		for _, candidate := range candidates {
			if candidate == latestReply {
				duplicate = true
				break
			}
		}
		if !duplicate {
			candidates = append([]string{latestReply}, candidates...)
		}
	}

	var lastErr error
	for _, candidate := range candidates {
		resolution, err := parseAnalysisResolution(candidate, taskName)
		if err == nil {
			resolution.Content = candidate
			return resolution, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("未找到可解析的修复方案")
	}
	return nil, lastErr
}

func (s *Service) loadLatestAnalysisReply(ctx context.Context, analysisTask gdb.Record) string {
	conversationID := analysisTask["conversation_id"].Int64()
	if conversationID == 0 {
		conv, err := s.convRepo.GetLatestByTask(ctx, analysisTask["id"].Int64(), "id")
		if err != nil || len(conv) == 0 {
			return ""
		}
		conversationID = g.NewVar(conv["id"]).Int64()
	}
	if conversationID == 0 {
		return ""
	}

	for i := 0; i < 20; i++ {
		if reply := s.loadLatestAnalysisReplyContent(ctx, conversationID); reply != "" {
			return reply
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

func (s *Service) loadLatestAnalysisReplyContent(ctx context.Context, conversationID int64) string {
	reply, err := s.messageRepo.GetLatestAssistantContentByConversation(ctx, conversationID)
	if err != nil {
		return ""
	}
	return reply
}

// taskPatch 架构师修复方案。
type taskPatch struct {
	Description       string   `json:"description"`
	AffectedResources []string `json:"affected_resources"`
	Reason            string   `json:"reason"`
}

type analysisResolution struct {
	Patch     *taskPatch
	SplitPlan *taskSplitPlan
	Content   string
}

type taskPatchEnvelope struct {
	TaskPatches []engine.ArchitectTaskPatch `json:"task_patches"`
}

type taskRepairEnvelope struct {
	TaskRepair engine.ArchitectTaskPatch `json:"task_repair"`
}

type planTaskEnvelope struct {
	Tasks []planTaskPatch `json:"tasks"`
}

type planTaskPatch struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	RoleType          string   `json:"role_type"`
	RoleLevel         string   `json:"role_level"`
	BatchNo           *int     `json:"batch_no"`
	Sort              *int     `json:"sort"`
	AffectedResources []string `json:"affected_resources"`
	DependsOn         []string `json:"depends_on"`
}

type taskSplitPlan struct {
	Tasks  []splitTaskSpec
	Reason string
}

type splitTaskSpec struct {
	Name              string
	Description       string
	RoleType          string
	RoleLevel         string
	BatchNo           *int
	Sort              *int
	AffectedResources []string
	DependsOn         []string
}

// parseTaskPatch 解析架构师输出的 JSON patch。
func parseTaskPatch(content string, taskName string) (*taskPatch, error) {
	resolution, err := parseAnalysisResolution(content, taskName)
	if err != nil {
		return nil, err
	}
	if resolution.SplitPlan != nil {
		return buildTaskPatchFromSplitPlan(resolution.SplitPlan), nil
	}
	return resolution.Patch, nil
}

func parseAnalysisResolution(content string, taskName string) (*analysisResolution, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("架构师输出为空")
	}

	candidates := []string{content}
	if normalized := normalizeEscapedPatchContent(content); normalized != "" && normalized != content {
		candidates = append(candidates, normalized)
	}

	for _, candidate := range candidates {
		if resolution, err := parseTaskPatchPayload(candidate, taskName); err == nil {
			return resolution, nil
		}
		re := reworkJsonBlockRe
		match := re.FindStringSubmatch(candidate)
		if len(match) == 2 {
			if resolution, err := parseTaskPatchPayload(match[1], taskName); err == nil {
				return resolution, nil
			}
		}
	}
	return nil, fmt.Errorf("未解析到有效 JSON patch")
}

func normalizeEscapedPatchContent(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		`\r\n`, "\n",
		`\n`, "\n",
		`\t`, "\t",
		`\"`, `"`,
	)
	return replacer.Replace(content)
}

func decodeAffectedResourcesJSON(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var resources []string
	if err := json.Unmarshal([]byte(raw), &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

func parseTaskPatchPayload(content string, taskName string) (*analysisResolution, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("patch 内容为空")
	}

	var patch taskPatch
	if err := json.Unmarshal([]byte(content), &patch); err == nil && taskPatchHasContent(&patch) {
		return &analysisResolution{Patch: &patch}, nil
	}

	var repair taskRepairEnvelope
	if err := json.Unmarshal([]byte(content), &repair); err == nil && architectTaskPatchMatchesTask(repair.TaskRepair, taskName) {
		return &analysisResolution{Patch: &taskPatch{
			Description:       repair.TaskRepair.Description,
			AffectedResources: repair.TaskRepair.AffectedResources,
			Reason:            repair.TaskRepair.Reason,
		}}, nil
	}

	var envelope taskPatchEnvelope
	if err := json.Unmarshal([]byte(content), &envelope); err == nil && len(envelope.TaskPatches) > 0 {
		selected, ok := selectTaskPatch(envelope.TaskPatches, taskName)
		if !ok {
			return nil, fmt.Errorf("task_patches 中未找到任务 %q 的修订项", taskName)
		}
		return &analysisResolution{Patch: &taskPatch{
			Description:       selected.Description,
			AffectedResources: selected.AffectedResources,
			Reason:            selected.Reason,
		}}, nil
	}

	var planEnvelope planTaskEnvelope
	if err := json.Unmarshal([]byte(content), &planEnvelope); err == nil && len(planEnvelope.Tasks) > 0 {
		if splitPlan := buildTaskSplitPlanFromPlanTasks(planEnvelope.Tasks, taskName); splitPlan != nil {
			return &analysisResolution{SplitPlan: splitPlan}, nil
		}
		patch := buildTaskPatchFromPlanTasks(planEnvelope.Tasks)
		if taskPatchHasContent(patch) {
			return &analysisResolution{Patch: patch}, nil
		}
	}

	return nil, fmt.Errorf("未解析到有效 JSON patch")
}

func taskPatchHasContent(patch *taskPatch) bool {
	return patch != nil && (strings.TrimSpace(patch.Description) != "" || len(patch.AffectedResources) > 0 || strings.TrimSpace(patch.Reason) != "")
}

func architectTaskPatchMatchesTask(patch engine.ArchitectTaskPatch, taskName string) bool {
	name := strings.TrimSpace(patch.TaskName)
	taskName = strings.TrimSpace(taskName)
	if name == "" || taskName == "" {
		return taskPatchHasContent(&taskPatch{
			Description:       patch.Description,
			AffectedResources: patch.AffectedResources,
			Reason:            patch.Reason,
		})
	}
	return name == taskName && taskPatchHasContent(&taskPatch{
		Description:       patch.Description,
		AffectedResources: patch.AffectedResources,
		Reason:            patch.Reason,
	})
}

func selectTaskPatch(patches []engine.ArchitectTaskPatch, taskName string) (engine.ArchitectTaskPatch, bool) {
	taskName = strings.TrimSpace(taskName)
	if len(patches) == 0 {
		return engine.ArchitectTaskPatch{}, false
	}
	if len(patches) == 1 {
		if !architectTaskPatchMatchesTask(patches[0], taskName) {
			return engine.ArchitectTaskPatch{}, false
		}
		return patches[0], true
	}
	for _, patch := range patches {
		if strings.TrimSpace(patch.TaskName) == taskName {
			return patch, true
		}
	}
	return engine.ArchitectTaskPatch{}, false
}

func buildTaskPatchFromPlanTasks(tasks []planTaskPatch) *taskPatch {
	if len(tasks) == 0 {
		return nil
	}

	var descriptionParts []string
	seenResources := make(map[string]struct{}, len(tasks))
	resources := make([]string, 0, len(tasks))

	for idx, item := range tasks {
		name := strings.TrimSpace(item.Name)
		desc := strings.TrimSpace(item.Description)
		switch {
		case name != "" && desc != "":
			descriptionParts = append(descriptionParts, fmt.Sprintf("步骤%d：%s。%s", idx+1, name, desc))
		case name != "":
			descriptionParts = append(descriptionParts, fmt.Sprintf("步骤%d：%s。", idx+1, name))
		case desc != "":
			descriptionParts = append(descriptionParts, fmt.Sprintf("步骤%d：%s", idx+1, desc))
		}

		for _, resource := range item.AffectedResources {
			resource = strings.TrimSpace(resource)
			if resource == "" {
				continue
			}
			if _, exists := seenResources[resource]; exists {
				continue
			}
			seenResources[resource] = struct{}{}
			resources = append(resources, resource)
		}
	}

	return &taskPatch{
		Description:       strings.Join(descriptionParts, "\n"),
		AffectedResources: resources,
		Reason:            fmt.Sprintf("失败分析返回了 plan-style tasks，共 %d 个子步骤，已自动折叠回单任务返工 patch", len(tasks)),
	}
}

func buildTaskSplitPlanFromPlanTasks(tasks []planTaskPatch, taskName string) *taskSplitPlan {
	if len(tasks) <= 1 {
		return nil
	}

	specs := make([]splitTaskSpec, 0, len(tasks))
	for _, item := range tasks {
		specs = append(specs, splitTaskSpec{
			Name:              strings.TrimSpace(item.Name),
			Description:       strings.TrimSpace(item.Description),
			RoleType:          strings.TrimSpace(item.RoleType),
			RoleLevel:         strings.TrimSpace(item.RoleLevel),
			BatchNo:           item.BatchNo,
			Sort:              item.Sort,
			AffectedResources: sanitizeAffectedResources(item.AffectedResources),
			DependsOn:         sanitizeDependsOn(item.DependsOn),
		})
	}
	return &taskSplitPlan{
		Tasks:  specs,
		Reason: fmt.Sprintf("任务 %q 命中 token limit/范围过大，返工分析已重设计为 %d 个同批次子任务", taskName, len(specs)),
	}
}

func buildTaskPatchFromSplitPlan(plan *taskSplitPlan) *taskPatch {
	if plan == nil {
		return nil
	}
	planTasks := make([]planTaskPatch, 0, len(plan.Tasks))
	for _, task := range plan.Tasks {
		planTasks = append(planTasks, planTaskPatch{
			Name:              task.Name,
			Description:       task.Description,
			AffectedResources: task.AffectedResources,
		})
	}
	patch := buildTaskPatchFromPlanTasks(planTasks)
	if patch != nil && strings.TrimSpace(plan.Reason) != "" {
		patch.Reason = plan.Reason
	}
	return patch
}

func sanitizeAffectedResources(resources []string) []string {
	seen := make(map[string]struct{}, len(resources))
	result := make([]string, 0, len(resources))
	for _, resource := range resources {
		resource = strings.TrimSpace(resource)
		if resource == "" {
			continue
		}
		if _, exists := seen[resource]; exists {
			continue
		}
		seen[resource] = struct{}{}
		result = append(result, resource)
	}
	return result
}

func sanitizeDependsOn(dependsOn []string) []string {
	seen := make(map[string]struct{}, len(dependsOn))
	result := make([]string, 0, len(dependsOn))
	for _, dep := range dependsOn {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		if _, exists := seen[dep]; exists {
			continue
		}
		seen[dep] = struct{}{}
		result = append(result, dep)
	}
	return result
}

func resolveHandoffReason(patch *taskPatch, splitPlan *taskSplitPlan) string {
	if splitPlan != nil {
		return splitPlan.Reason
	}
	if patch != nil {
		return patch.Reason
	}
	return ""
}

func (s *Service) findNextEscalatedTask(ctx context.Context, workflowRunID, excludeTaskID int64) (int64, bool, error) {
	if workflowRunID == 0 {
		return 0, false, nil
	}
	records, err := s.taskRepo.ListByWorkflowAndStatuses(ctx, workflowRunID, []string{domainTask.StatusEscalated},
		"id", "batch_no", "sort", "updated_at", "created_at")
	if err != nil {
		return 0, false, err
	}
	if len(records) == 0 {
		return 0, false, nil
	}

	bestBatch := 0
	bestSort := 0
	bestID := int64(0)
	found := false
	for _, record := range records {
		taskID := g.NewVar(record["id"]).Int64()
		if taskID == 0 || taskID == excludeTaskID {
			continue
		}
		batchNo := g.NewVar(record["batch_no"]).Int()
		sortNo := g.NewVar(record["sort"]).Int()
		if !found || batchNo < bestBatch || (batchNo == bestBatch && sortNo < bestSort) || (batchNo == bestBatch && sortNo == bestSort && taskID < bestID) {
			bestBatch = batchNo
			bestSort = sortNo
			bestID = taskID
			found = true
		}
	}
	if !found {
		return 0, false, nil
	}
	return bestID, true, nil
}

func (s *Service) applyTaskPatch(ctx context.Context, failedTaskID int64, failedTask gdb.Record, patch *taskPatch) (bool, error) {
	updateData := g.Map{
		"status":     domainTask.StatusPending,
		"result":     nil,
		"started_at": nil,
		"updated_at": gtime.Now(),
	}
	if strings.TrimSpace(patch.Description) != "" {
		updateData["description"] = patch.Description
	}
	if len(patch.AffectedResources) > 0 {
		currentResources, decodeErr := decodeAffectedResourcesJSON(failedTask["affected_resources"].String())
		if decodeErr != nil {
			return false, fmt.Errorf("解析原任务 affected_resources 失败: %w", decodeErr)
		}
		if introduced := resourcepath.FindNewlyIntroducedGovernedRootFiles(currentResources, patch.AffectedResources); len(introduced) > 0 {
			return false, fmt.Errorf("返工 patch 不允许新增受治理仓库文件: %s", strings.Join(introduced, ", "))
		}
		resJSON, jsonErr := json.Marshal(patch.AffectedResources)
		if jsonErr != nil {
			g.Log().Warningf(ctx, "[ReworkStage] 序列化 affected_resources 失败: %v", jsonErr)
		} else {
			updateData["affected_resources"] = string(resJSON)
		}
	}

	rows, err := s.taskRepo.UpdateFieldsIfStatuses(ctx, failedTaskID, []string{domainTask.StatusFailed, domainTask.StatusEscalated}, updateData)
	if err != nil {
		return false, fmt.Errorf("回写原任务失败: %w", err)
	}
	if rows > 0 {
		return false, nil
	}
	return s.resolveAlreadyRecovered(ctx, failedTaskID)
}

func (s *Service) resolveAlreadyRecovered(ctx context.Context, failedTaskID int64) (bool, error) {
	currentTask, statusErr := s.taskRepo.GetByIDMap(ctx, failedTaskID, "status")
	if statusErr != nil {
		return false, fmt.Errorf("查询原任务当前状态失败: %w", statusErr)
	}
	currentStatus := g.NewVar(currentTask["status"]).String()

	_, alreadyRecovered := classifyOriginalTaskStatusForRework(currentStatus)
	if !alreadyRecovered {
		return false, fmt.Errorf("原任务(%d)状态异常(%s)，回写跳过", failedTaskID, currentStatus)
	}

	g.Log().Infof(ctx, "[ReworkStage] 原任务(%d)已由其他链路推进到 %s，视为返工已收敛",
		failedTaskID, currentStatus)
	return true, nil
}

func (s *Service) applySplitPlan(ctx context.Context, analysisTask gdb.Record, failedTaskID int64, plan *taskSplitPlan) (bool, map[string]interface{}, error) {
	if plan == nil || len(plan.Tasks) == 0 {
		return false, nil, fmt.Errorf("拆分方案为空")
	}

	failedTask, err := s.taskRepo.GetRecordByID(ctx, failedTaskID,
		"id", "workflow_run_id", "plan_version_id", "blueprint_id", "parent_task_id", "depends_on_task_ids",
		"root_task_id", "task_kind", "name", "description", "role_type", "role_level", "execution_mode",
		"model_id", "batch_no", "sort", "created_by", "dept_id", "affected_resources",
	)
	if err != nil || failedTask == nil || failedTask.IsEmpty() {
		return false, nil, fmt.Errorf("原失败任务(%d) 不存在", failedTaskID)
	}

	currentResources, decodeErr := decodeAffectedResourcesJSON(failedTask["affected_resources"].String())
	if decodeErr != nil {
		return false, nil, fmt.Errorf("解析原任务 affected_resources 失败: %w", decodeErr)
	}
	for _, task := range plan.Tasks {
		if introduced := resourcepath.FindNewlyIntroducedGovernedRootFiles(currentResources, task.AffectedResources); len(introduced) > 0 {
			return false, nil, fmt.Errorf("返工拆分任务 %q 不允许新增受治理仓库文件: %s", task.Name, strings.Join(introduced, ", "))
		}
	}

	createdTaskIDs := make([]int64, 0, len(plan.Tasks))
	createdTaskNames := make([]string, 0, len(plan.Tasks))
	err = repo.WithTx(ctx, func(ctx context.Context, tx gdb.TX) error {
		createdTaskIDs = createdTaskIDs[:0]
		createdTaskNames = createdTaskNames[:0]
		record, txErr := tx.Model("mvp_domain_task").Ctx(ctx).
			Where("id", failedTaskID).
			WhereNull("deleted_at").
			Fields("status").
			One()
		if txErr != nil {
			return fmt.Errorf("查询原任务状态失败: %w", txErr)
		}
		status := record["status"].String()
		if status != domainTask.StatusFailed && status != domainTask.StatusEscalated {
			_, alreadyRecovered := classifyOriginalTaskStatusForRework(status)
			if alreadyRecovered {
				return nil
			}
			return fmt.Errorf("原任务(%d)状态异常(%s)，拆分跳过", failedTaskID, status)
		}

		baseDeps, depErr := decodeDependsOnTaskIDs(failedTask["depends_on_task_ids"].String(), failedTask["parent_task_id"].Int64())
		if depErr != nil {
			return fmt.Errorf("解析原任务依赖失败: %w", depErr)
		}
		rootTaskID := failedTask["root_task_id"].Int64()
		if rootTaskID == 0 {
			rootTaskID = failedTaskID
		}
		baseSort := failedTask["sort"].Int()
		if baseSort < 1 {
			baseSort = 100
		}
		now := gtime.Now()
		nameToID := make(map[string]int64, len(plan.Tasks))
		childDependsOnNames := make(map[int64][]string, len(plan.Tasks))

		for idx, task := range plan.Tasks {
			taskID := int64(snowflake.Generate())
			taskName := firstNonEmpty(task.Name, fmt.Sprintf("%s / split-%d", failedTask["name"].String(), idx+1))
			roleType := firstNonEmpty(task.RoleType, failedTask["role_type"].String())
			roleLevel := firstNonEmpty(task.RoleLevel, failedTask["role_level"].String())
			batchNo := failedTask["batch_no"].Int()
			if batchNo < 1 {
				batchNo = 1
			}
			sortValue := baseSort*100 + idx + 1
			if task.Sort != nil && *task.Sort > 0 {
				sortValue = baseSort*100 + *task.Sort
			}
			resourcesJSON := "[]"
			if len(task.AffectedResources) > 0 {
				data, jsonErr := json.Marshal(task.AffectedResources)
				if jsonErr != nil {
					return fmt.Errorf("序列化拆分任务资源失败: %w", jsonErr)
				}
				resourcesJSON = string(data)
			}
			if _, insErr := tx.Model("mvp_domain_task").Ctx(ctx).Insert(g.Map{
				"id":                 taskID,
				"workflow_run_id":    failedTask["workflow_run_id"].Int64(),
				"stage_run_id":       analysisTask["stage_run_id"].Int64(),
				"plan_version_id":    failedTask["plan_version_id"].Int64(),
				"blueprint_id":       failedTask["blueprint_id"].Int64(),
				"source_task_id":     failedTaskID,
				"root_task_id":       rootTaskID,
				"task_kind":          failedTask["task_kind"].String(),
				"name":               taskName,
				"description":        firstNonEmpty(task.Description, failedTask["description"].String()),
				"role_type":          roleType,
				"role_level":         roleLevel,
				"execution_mode":     failedTask["execution_mode"].String(),
				"status":             domainTask.StatusPending,
				"model_id":           failedTask["model_id"].Int64(),
				"batch_no":           batchNo,
				"sort":               sortValue,
				"affected_resources": resourcesJSON,
				"retry_count":        0,
				"created_by":         failedTask["created_by"].Int64(),
				"dept_id":            failedTask["dept_id"].Int64(),
				"created_at":         now,
				"updated_at":         now,
			}); insErr != nil {
				return fmt.Errorf("创建拆分任务失败: %w", insErr)
			}
			nameToID[taskName] = taskID
			createdTaskIDs = append(createdTaskIDs, taskID)
			createdTaskNames = append(createdTaskNames, taskName)
			childDependsOnNames[taskID] = append([]string{}, task.DependsOn...)
		}

		for _, taskID := range createdTaskIDs {
			deps := append([]int64{}, baseDeps...)
			for _, depName := range childDependsOnNames[taskID] {
				if depID, ok := nameToID[depName]; ok {
					deps = append(deps, depID)
				}
			}
			deps = uniqueInt64s(deps)
			parentTaskID := int64(0)
			if len(deps) > 0 {
				parentTaskID = deps[0]
			}
			depsJSON, jsonErr := json.Marshal(deps)
			if jsonErr != nil {
				return fmt.Errorf("序列化拆分任务依赖失败: %w", jsonErr)
			}
			if _, upErr := tx.Model("mvp_domain_task").Ctx(ctx).
				Where("id", taskID).
				WhereNull("deleted_at").
				Data(g.Map{
					"parent_task_id":      parentTaskID,
					"depends_on_task_ids": string(depsJSON),
					"updated_at":          now,
				}).Update(); upErr != nil {
				return fmt.Errorf("更新拆分任务依赖失败: %w", upErr)
			}
		}

		if err := rewriteDependentTasksForSplit(ctx, tx, failedTask["workflow_run_id"].Int64(), failedTaskID, createdTaskIDs, now); err != nil {
			return err
		}

		if _, upErr := tx.Model("mvp_domain_task").Ctx(ctx).
			Where("id", failedTaskID).
			WhereIn("status", []string{domainTask.StatusFailed, domainTask.StatusEscalated}).
			WhereNull("deleted_at").
			Data(g.Map{
				"status":       domainTask.StatusCompleted,
				"result":       fmt.Sprintf("reworked_split:%v", createdTaskIDs),
				"completed_at": now,
				"updated_at":   now,
			}).Update(); upErr != nil {
			return fmt.Errorf("标记原任务为已拆分完成失败: %w", upErr)
		}
		return nil
	})
	if err != nil {
		return false, nil, err
	}
	if len(createdTaskIDs) == 0 {
		reconciled, err := s.resolveAlreadyRecovered(ctx, failedTaskID)
		return reconciled, nil, err
	}
	return false, map[string]interface{}{
		"mode":               "split_tasks",
		"reason":             plan.Reason,
		"created_task_ids":   createdTaskIDs,
		"created_task_names": createdTaskNames,
		"source_task_id":     failedTaskID,
	}, nil
}

func decodeDependsOnTaskIDs(raw string, fallbackParentID int64) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" || raw == "[]" {
		if fallbackParentID > 0 {
			return []int64{fallbackParentID}, nil
		}
		return nil, nil
	}
	var depIDs []int64
	if err := json.Unmarshal([]byte(raw), &depIDs); err != nil {
		return nil, err
	}
	if len(depIDs) == 0 && fallbackParentID > 0 {
		return []int64{fallbackParentID}, nil
	}
	return uniqueInt64s(depIDs), nil
}

func rewriteDependentTasksForSplit(ctx context.Context, tx gdb.TX, workflowRunID, failedTaskID int64, childTaskIDs []int64, now *gtime.Time) error {
	records, err := tx.Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("id <> ?", failedTaskID).
		WhereNull("deleted_at").
		Fields("id", "parent_task_id", "depends_on_task_ids").
		All()
	if err != nil {
		return fmt.Errorf("查询待重挂依赖任务失败: %w", err)
	}
	for _, record := range records {
		taskID := record["id"].Int64()
		parentTaskID := record["parent_task_id"].Int64()
		deps, depErr := decodeDependsOnTaskIDs(record["depends_on_task_ids"].String(), parentTaskID)
		if depErr != nil {
			return fmt.Errorf("解析依赖任务 %d 失败: %w", taskID, depErr)
		}
		if !slices.Contains(deps, failedTaskID) && parentTaskID != failedTaskID {
			continue
		}
		rewritten := make([]int64, 0, len(deps)+len(childTaskIDs))
		for _, depID := range deps {
			if depID == failedTaskID {
				rewritten = append(rewritten, childTaskIDs...)
				continue
			}
			rewritten = append(rewritten, depID)
		}
		if !slices.Contains(deps, failedTaskID) {
			rewritten = append(rewritten, childTaskIDs...)
		}
		rewritten = uniqueInt64s(rewritten)
		parentTaskID = 0
		if len(rewritten) > 0 {
			parentTaskID = rewritten[0]
		}
		depsJSON, jsonErr := json.Marshal(rewritten)
		if jsonErr != nil {
			return fmt.Errorf("序列化重挂依赖失败: %w", jsonErr)
		}
		if _, upErr := tx.Model("mvp_domain_task").Ctx(ctx).
			Where("id", taskID).
			WhereNull("deleted_at").
			Data(g.Map{
				"parent_task_id":      parentTaskID,
				"depends_on_task_ids": string(depsJSON),
				"updated_at":          now,
			}).Update(); upErr != nil {
			return fmt.Errorf("重挂依赖失败: task=%d err=%w", taskID, upErr)
		}
	}
	return nil
}

func uniqueInt64s(items []int64) []int64 {
	seen := make(map[int64]struct{}, len(items))
	result := make([]int64, 0, len(items))
	for _, item := range items {
		if item == 0 {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
