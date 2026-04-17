// Package rework 统一承接 bug 修复与失败升级的返工阶段。
// 当 execute 阶段任务失败超限时，由 orchestrator 触发 rework stage。
package rework

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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
				"修复方案不得新增 .gitignore 这类仓库级控制文件，除非原任务 affected_resources 已明确包含该文件。\n"+
				"修复方案必须严格围绕当前任务，不能改变原任务目标、不能脱离原有任务设定、不能把范围扩展到无关任务。\n"+
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
	patch, patchContent, err := s.resolveAnalysisPatch(ctx, analysisTask, failedTask)
	if err != nil {
		g.Log().Warningf(ctx, "[ReworkStage] 解析修复方案失败: task=%d err=%v", analysisTaskID, err)
		// 解析失败也要完成 rework stage
		if s.stageCompleter != nil {
			_ = s.stageCompleter.FailStage(ctx, stageRunID, "架构师修复方案解析失败: "+err.Error())
		}
		return nil
	}

	// 4. 回写原失败任务
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
			return fmt.Errorf("解析原任务 affected_resources 失败: %w", decodeErr)
		}
		if introduced := resourcepath.FindNewlyIntroducedGovernedRootFiles(currentResources, patch.AffectedResources); len(introduced) > 0 {
			return fmt.Errorf("返工 patch 不允许新增受治理仓库文件: %s", strings.Join(introduced, ", "))
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
		return fmt.Errorf("回写原任务失败: %w", err)
	}
	reconciledWithoutWrite := false
	if rows == 0 {
		currentTask, statusErr := s.taskRepo.GetByIDMap(ctx, failedTaskID, "status")
		if statusErr != nil {
			return fmt.Errorf("查询原任务当前状态失败: %w", statusErr)
		}
		currentStatus := g.NewVar(currentTask["status"]).String()

		_, alreadyRecovered := classifyOriginalTaskStatusForRework(currentStatus)
		if !alreadyRecovered {
			reason := fmt.Sprintf("原任务(%d)状态异常(%s)，回写跳过", failedTaskID, currentStatus)
			g.Log().Warningf(ctx, "[ReworkStage] %s", reason)
			if s.stageCompleter != nil {
				_ = s.stageCompleter.FailStage(ctx, stageRunID, reason)
			}
			return nil
		}

		reconciledWithoutWrite = true
		g.Log().Infof(ctx, "[ReworkStage] 原任务(%d)已由其他链路推进到 %s，视为返工已收敛",
			failedTaskID, currentStatus)
	}

	// 5. 写 handoff_record（分析 → 原任务）
	handoffID := int64(snowflake.Generate())
	payloadJSON := marshalHandoffPayload(map[string]interface{}{
		"content": patchContent,
		"reason":  patch.Reason,
	})
	if insErr := s.handoffRepo.Create(ctx, g.Map{
		"id":              handoffID,
		"workflow_run_id": analysisTask["workflow_run_id"].Int64(),
		"from_task_id":    analysisTaskID,
		"to_task_id":      failedTaskID,
		"handoff_type":    "rework",
		"reason":          patch.Reason,
		"payload":         payloadJSON,
		"created_at":      gtime.Now(),
	}); insErr != nil {
		g.Log().Errorf(ctx, "[ReworkStage] 写入 handoff_record(rework) 失败: analysis=%d target=%d err=%v", analysisTaskID, failedTaskID, insErr)
	}

	if reconciledWithoutWrite {
		g.Log().Infof(ctx, "[ReworkStage] 返工收敛完成: analysis=%d target=%d reason=%s", analysisTaskID, failedTaskID, patch.Reason)
	} else {
		g.Log().Infof(ctx, "[ReworkStage] 回写完成: analysis=%d → original=%d reason=%s", analysisTaskID, failedTaskID, patch.Reason)
	}

	// 6. 完成 rework stage
	if s.stageCompleter != nil {
		_ = s.stageCompleter.CompleteStage(ctx, stageRunID)
	}

	// 7. 按来源阶段决定回流目标
	workflowRunID := analysisTask["workflow_run_id"].Int64()
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

func (s *Service) resolveAnalysisPatch(ctx context.Context, analysisTask gdb.Record, failedTask gdb.Record) (*taskPatch, string, error) {
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
		patch, err := parseTaskPatch(candidate, taskName)
		if err == nil {
			return patch, candidate, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("未找到可解析的修复方案")
	}
	return nil, "", lastErr
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
	AffectedResources []string `json:"affected_resources"`
}

// parseTaskPatch 解析架构师输出的 JSON patch。
func parseTaskPatch(content string, taskName string) (*taskPatch, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("架构师输出为空")
	}

	candidates := []string{content}
	if normalized := normalizeEscapedPatchContent(content); normalized != "" && normalized != content {
		candidates = append(candidates, normalized)
	}

	for _, candidate := range candidates {
		if patch, err := parseTaskPatchPayload(candidate, taskName); err == nil {
			return patch, nil
		}
		re := reworkJsonBlockRe
		match := re.FindStringSubmatch(candidate)
		if len(match) == 2 {
			if patch, err := parseTaskPatchPayload(match[1], taskName); err == nil {
				return patch, nil
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

func parseTaskPatchPayload(content string, taskName string) (*taskPatch, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("patch 内容为空")
	}

	var patch taskPatch
	if err := json.Unmarshal([]byte(content), &patch); err == nil && taskPatchHasContent(&patch) {
		return &patch, nil
	}

	var repair taskRepairEnvelope
	if err := json.Unmarshal([]byte(content), &repair); err == nil && architectTaskPatchMatchesTask(repair.TaskRepair, taskName) {
		return &taskPatch{
			Description:       repair.TaskRepair.Description,
			AffectedResources: repair.TaskRepair.AffectedResources,
			Reason:            repair.TaskRepair.Reason,
		}, nil
	}

	var envelope taskPatchEnvelope
	if err := json.Unmarshal([]byte(content), &envelope); err == nil && len(envelope.TaskPatches) > 0 {
		selected, ok := selectTaskPatch(envelope.TaskPatches, taskName)
		if !ok {
			return nil, fmt.Errorf("task_patches 中未找到任务 %q 的修订项", taskName)
		}
		return &taskPatch{
			Description:       selected.Description,
			AffectedResources: selected.AffectedResources,
			Reason:            selected.Reason,
		}, nil
	}

	var planEnvelope planTaskEnvelope
	if err := json.Unmarshal([]byte(content), &planEnvelope); err == nil && len(planEnvelope.Tasks) > 0 {
		patch := buildTaskPatchFromPlanTasks(planEnvelope.Tasks)
		if taskPatchHasContent(patch) {
			return patch, nil
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
