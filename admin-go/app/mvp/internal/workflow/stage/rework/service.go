// Package rework 统一承接 bug 修复与失败升级的返工阶段。
// 当 execute 阶段任务失败超限时，由 orchestrator 触发 rework stage。
package rework

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/utility/snowflake"
)

// StageCompleter 阶段操作回调（避免循环依赖）。
type StageCompleter interface {
	CompleteStage(ctx context.Context, stageRunID int64) error
	FailStage(ctx context.Context, stageRunID int64, reason string) error
}

// ExecuteTriggerFn 触发执行阶段回调。
type ExecuteTriggerFn func(ctx context.Context, workflowRunID, planVersionID int64) error

// Service 返工阶段服务。
type Service struct {
	stageCompleter StageCompleter
	executeTrigger ExecuteTriggerFn
}

// NewService 创建返工阶段服务。
func NewService() *Service { return &Service{} }

// SetStageCompleter 注册阶段完成回调。
func (s *Service) SetStageCompleter(sc StageCompleter) { s.stageCompleter = sc }

// SetExecuteTrigger 注册执行阶段触发回调。
func (s *Service) SetExecuteTrigger(fn ExecuteTriggerFn) { s.executeTrigger = fn }

// HandleRework 处理返工流程。
// 接收失败的 domain_task，创建架构师分析任务，分析完成后回写原任务。
func (s *Service) HandleRework(ctx context.Context, stageRunID int64, failedTaskID int64) error {
	// 1. 查询 stage_run 和 workflow_run 信息
	stageRun, err := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).One()
	if err != nil || stageRun.IsEmpty() {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := stageRun["workflow_run_id"].Int64()

	// 2. 查询失败任务详情
	failedTask, err := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", failedTaskID).One()
	if err != nil || failedTask.IsEmpty() {
		return fmt.Errorf("failed domain_task(%d) 不存在", failedTaskID)
	}

	// 3. 查项目 ID（用于日志）
	projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	_ = projectID

	// 4. 检查返工轮次是否超限
	maxRounds := engine.GetConfigInt(ctx, "failure_handoff.max_rounds", "engine.failureHandoff.maxRounds", 3)
	reworkCount, _ := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("from_task_id", failedTaskID).
		Count()
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
	_, err = g.DB().Model("mvp_domain_task").Ctx(ctx).Insert(g.Map{
		"id":              analysisTaskID,
		"workflow_run_id": workflowRunID,
		"stage_run_id":    stageRunID,
		"task_kind":       "failure_analysis",
		"name":            fmt.Sprintf("失败分析: %s", failedTask["name"].String()),
		"description": fmt.Sprintf(
			"请分析任务失败原因，并给出可直接回写到原任务的修复方案。\n\n"+
				"关联任务ID：%d\n角色：%s\n错误信息：\n%s\n\n"+
				"原任务名称：%s\n原任务描述：\n%s\n\n"+
				"请严格输出 JSON：\n{\"description\":\"修订后的任务描述\",\"affected_resources\":[\"路径\"],\"reason\":\"修订原因\"}",
			failedTaskID, failedTask["role_type"].String(),
			failedTask["result"].String(),
			failedTask["name"].String(), failedTask["description"].String(),
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
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return fmt.Errorf("创建分析任务失败: %w", err)
	}

	// 6. 写 handoff_record
	handoffID := int64(snowflake.Generate())
	_, _ = g.DB().Model("mvp_handoff_record").Ctx(ctx).Insert(g.Map{
		"id":              handoffID,
		"workflow_run_id": workflowRunID,
		"from_task_id":    failedTaskID,
		"to_task_id":      analysisTaskID,
		"handoff_type":    "failure_escalation",
		"reason":          failedTask["result"].String(),
		"created_at":      now,
	})

	g.Log().Infof(ctx, "[ReworkStage] 创建分析任务: stageRunID=%d failedTask=%d analysisTask=%d round=%d/%d",
		stageRunID, failedTaskID, analysisTaskID, reworkCount+1, maxRounds)

	return nil
}

// OnAnalysisCompleted 架构师分析任务完成后的回调。
// 解析分析结果，回写原失败任务，推进回 execute stage。
func (s *Service) OnAnalysisCompleted(ctx context.Context, stageRunID int64, analysisTaskID int64) error {
	// 1. 获取分析任务
	analysisTask, err := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", analysisTaskID).One()
	if err != nil || analysisTask.IsEmpty() {
		return fmt.Errorf("分析任务(%d) 不存在", analysisTaskID)
	}

	// 2. 获取原失败任务 ID
	failedTaskID := analysisTask["source_task_id"].Int64()
	if failedTaskID == 0 {
		return fmt.Errorf("分析任务没有关联的来源任务")
	}

	// 3. 解析架构师修复方案
	patch, err := parseTaskPatch(analysisTask["result"].String())
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
		resJSON, _ := json.Marshal(patch.AffectedResources)
		updateData["affected_resources"] = string(resJSON)
	}

	_, err = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", failedTaskID).
		Where("status", domainTask.StatusFailed).
		Update(updateData)
	if err != nil {
		return fmt.Errorf("回写原任务失败: %w", err)
	}

	// 5. 写 handoff_record（分析 → 原任务）
	handoffID := int64(snowflake.Generate())
	_, _ = g.DB().Model("mvp_handoff_record").Ctx(ctx).Insert(g.Map{
		"id":              handoffID,
		"workflow_run_id": analysisTask["workflow_run_id"].Int64(),
		"from_task_id":    analysisTaskID,
		"to_task_id":      failedTaskID,
		"handoff_type":    "rework",
		"reason":          patch.Reason,
		"payload":         analysisTask["result"].String(),
		"created_at":      gtime.Now(),
	})

	g.Log().Infof(ctx, "[ReworkStage] 回写完成: analysis=%d → original=%d reason=%s", analysisTaskID, failedTaskID, patch.Reason)

	// 6. 完成 rework stage，推回 execute/review
	if s.stageCompleter != nil {
		_ = s.stageCompleter.CompleteStage(ctx, stageRunID)
	}

	return nil
}

// taskPatch 架构师修复方案。
type taskPatch struct {
	Description       string   `json:"description"`
	AffectedResources []string `json:"affected_resources"`
	Reason            string   `json:"reason"`
}

// parseTaskPatch 解析架构师输出的 JSON patch。
func parseTaskPatch(content string) (*taskPatch, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("架构师输出为空")
	}

	var patch taskPatch
	if err := json.Unmarshal([]byte(content), &patch); err == nil {
		return &patch, nil
	}

	re := regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")
	match := re.FindStringSubmatch(content)
	if len(match) == 2 {
		if err := json.Unmarshal([]byte(match[1]), &patch); err == nil {
			return &patch, nil
		}
	}
	return nil, fmt.Errorf("未解析到有效 JSON patch")
}
