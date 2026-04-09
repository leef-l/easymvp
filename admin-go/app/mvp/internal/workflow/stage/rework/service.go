// Package rework 统一承接 bug 修复与失败升级的返工阶段。
// 当 execute 阶段任务失败超限时，由 orchestrator 触发 rework stage。
package rework

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

var reworkJsonBlockRe = regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")

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
}

// NewService 创建返工阶段服务。
func NewService() *Service { return &Service{} }

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
	stageRun, err := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).WhereNull("deleted_at").One()
	if err != nil || stageRun.IsEmpty() {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := stageRun["workflow_run_id"].Int64()

	// 2. 查询失败任务详情
	failedTask, err := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", failedTaskID).WhereNull("deleted_at").One()
	if err != nil || failedTask.IsEmpty() {
		return fmt.Errorf("failed domain_task(%d) 不存在", failedTaskID)
	}

	// 3. 检查返工轮次是否超限
	maxRounds := engine.GetConfigInt(ctx, "failure_handoff.max_rounds", "engine.failureHandoff.maxRounds", 3)
	reworkCount, countErr := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("from_task_id", failedTaskID).
		Count()
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
	_, err = g.DB().Model("mvp_domain_task").Ctx(ctx).Insert(g.Map{
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
		"role_type":       "architect",
		"role_level":      "max",
		"execution_mode":  "chat",
		"status":          domainTask.StatusPending,
		"source_task_id":  failedTaskID,
		"root_task_id":    rootTaskID,
		"batch_no":        0, // 高优先级
		"sort":            0,
		"retry_count":     0,
		"created_by":      scope.CreatedBy,
		"dept_id":         scope.DeptID,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return fmt.Errorf("创建分析任务失败: %w", err)
	}

	// 6. 写 handoff_record（payload 中记录来源阶段，用于返工完成后决定回流目标）
	handoffID := int64(snowflake.Generate())
	payloadJSON, jsonErr := json.Marshal(map[string]string{"source_stage": sourceStage})
	if jsonErr != nil {
		g.Log().Warningf(ctx, "[ReworkStage] 序列化 handoff payload 失败: %v", jsonErr)
	}
	if _, insErr := g.DB().Model("mvp_handoff_record").Ctx(ctx).Insert(g.Map{
		"id":              handoffID,
		"workflow_run_id": workflowRunID,
		"from_task_id":    failedTaskID,
		"to_task_id":      analysisTaskID,
		"handoff_type":    "failure_escalation",
		"reason":          failedTask["result"].String(),
		"payload":         string(payloadJSON),
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
	analysisTask, err := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", analysisTaskID).WhereNull("deleted_at").One()
	if err != nil || analysisTask.IsEmpty() {
		return fmt.Errorf("分析任务(%d) 不存在", analysisTaskID)
	}

	// 2. 获取原失败任务 ID
	failedTaskID := analysisTask["source_task_id"].Int64()
	if failedTaskID == 0 {
		return fmt.Errorf("分析任务没有关联的来源任务")
	}

	failedTask, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", failedTaskID).
		WhereNull("deleted_at").
		Fields("id, name").
		One()
	if err != nil || failedTask.IsEmpty() {
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
		resJSON, jsonErr := json.Marshal(patch.AffectedResources)
		if jsonErr != nil {
			g.Log().Warningf(ctx, "[ReworkStage] 序列化 affected_resources 失败: %v", jsonErr)
		} else {
			updateData["affected_resources"] = string(resJSON)
		}
	}

	res, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", failedTaskID).
		Where("status", domainTask.StatusFailed).
		Update(updateData)
	if err != nil {
		return fmt.Errorf("回写原任务失败: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		// 原任务已不在 failed 状态（被手动重试或其他流程改状态），标记 rework 失败
		reason := fmt.Sprintf("原任务(%d)已不在 failed 状态，回写跳过", failedTaskID)
		g.Log().Warningf(ctx, "[ReworkStage] %s", reason)
		if s.stageCompleter != nil {
			_ = s.stageCompleter.FailStage(ctx, stageRunID, reason)
		}
		return nil
	}

	// 5. 写 handoff_record（分析 → 原任务）
	handoffID := int64(snowflake.Generate())
	if _, insErr := g.DB().Model("mvp_handoff_record").Ctx(ctx).Insert(g.Map{
		"id":              handoffID,
		"workflow_run_id": analysisTask["workflow_run_id"].Int64(),
		"from_task_id":    analysisTaskID,
		"to_task_id":      failedTaskID,
		"handoff_type":    "rework",
		"reason":          patch.Reason,
		"payload":         patchContent,
		"created_at":      gtime.Now(),
	}); insErr != nil {
		g.Log().Errorf(ctx, "[ReworkStage] 写入 handoff_record(rework) 失败: analysis=%d target=%d err=%v", analysisTaskID, failedTaskID, insErr)
	}

	g.Log().Infof(ctx, "[ReworkStage] 回写完成: analysis=%d → original=%d reason=%s", analysisTaskID, failedTaskID, patch.Reason)

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
	payload, err := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("from_task_id", failedTaskID).
		Where("handoff_type", "failure_escalation").
		OrderDesc("created_at").
		Value("payload")
	if err != nil || payload.IsEmpty() {
		return "execute" // 默认回 execute
	}
	var data map[string]string
	if json.Unmarshal([]byte(payload.String()), &data) == nil {
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
		conv, err := g.DB().Model("mvp_conversation").Ctx(ctx).
			Where("task_id", analysisTask["id"].Int64()).
			WhereNull("deleted_at").
			OrderDesc("created_at").
			Fields("id").
			One()
		if err != nil || conv.IsEmpty() {
			return ""
		}
		conversationID = conv["id"].Int64()
	}
	if conversationID == 0 {
		return ""
	}

	reply, err := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("role", "assistant").
		Where("status", "completed").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		OrderDesc("id").
		Fields("content").
		One()
	if err != nil || reply.IsEmpty() {
		return ""
	}
	return strings.TrimSpace(reply["content"].String())
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

// parseTaskPatch 解析架构师输出的 JSON patch。
func parseTaskPatch(content string, taskName string) (*taskPatch, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("架构师输出为空")
	}

	if patch, err := parseTaskPatchPayload(content, taskName); err == nil {
		return patch, nil
	}

	re := reworkJsonBlockRe
	match := re.FindStringSubmatch(content)
	if len(match) == 2 {
		if patch, err := parseTaskPatchPayload(match[1], taskName); err == nil {
			return patch, nil
		}
	}
	return nil, fmt.Errorf("未解析到有效 JSON patch")
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
