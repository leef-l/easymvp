package chat

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

var eventLabelMap = map[string]string{
	"workflow.created":              "工作流已创建",
	"workflow.paused":               "工作流已暂停",
	"workflow.resumed":              "工作流已恢复",
	"workflow.canceled":             "工作流已取消",
	"workflow.completed":            "工作流已完成",
	"stage.started":                 "阶段已启动",
	"stage.completed":               "阶段已完成",
	"stage.failed":                  "阶段失败",
	"plan_version.created":          "方案版本已创建",
	"plan_version.submitted":        "方案已提交审核",
	"plan_version.approved":         "方案审核通过",
	"plan_version.rejected":         "方案被驳回",
	"review.issue_created":          "发现审核问题",
	"review.decision_ready":         "审核决策就绪",
	"review.issue_replan_requested": "审核问题已回流方案修订",
	"task.created":                  "任务已创建",
	"task.started":                  "任务已启动",
	"task.completed":                "任务已完成",
	"task.delivery_prepared":        "任务交付物已生成",
	"task.failed":                   "任务失败",
	"task.escalated":                "任务已升级",
	"task.review_required":          "任务进入人工审核闸门",
	"task.retried":                  "任务已重试",
	"task.manual_updated":           "任务已人工修改",
	"task.sync_applied":             "任务变更已自动回写",
	"replan.completed":              "重规划完成",
	"replan.failed":                 "重规划失败",
	"replan.aborted":                "重规划中止",
	"accept.issue_rework_requested": "验收问题已触发返工",
	"verification.started":          "验证已启动",
	"verification.completed":        "验证已完成",
	"verification.failed":           "验证失败",
	"verification.repair_requested": "验证问题已触发返工",
	"workflow.force_stage":          "工作流已人工切换阶段",
}

func formatTimelineLabel(eventType, payload string) string {
	label := eventLabelMap[eventType]
	if label == "" {
		label = eventType
	}
	if payload == "" || payload == "null" {
		return label
	}

	var pm map[string]string
	if json.Unmarshal([]byte(payload), &pm) != nil {
		return label
	}
	if st, ok := pm["stage_type"]; ok {
		stageLabel := map[string]string{"design": "设计", "review": "审核", "execute": "执行", "accept": "验收", "rework": "返工", "complete": "完成"}[st]
		if stageLabel != "" {
			switch {
			case strings.HasPrefix(label, "阶段"):
				label = stageLabel + label
			case strings.Contains(label, "切换阶段"):
				label = strings.Replace(label, "切换阶段", "切换到"+stageLabel+"阶段", 1)
			default:
				label = stageLabel + "阶段 " + label
			}
		}
	}
	if reason, ok := pm["reason"]; ok && reason != "" {
		label += "：" + reason
	}
	return label
}

func buildStageHistoryItem(record gdb.Record) v1.StageHistoryItem {
	return v1.StageHistoryItem{
		ID:         snowflake.JsonInt64(record["id"].Int64()),
		StageType:  record["stage_type"].String(),
		StageNo:    record["stage_no"].Int(),
		Status:     record["status"].String(),
		StartedAt:  normalizeDBUTCGTime(record["started_at"].GTime()),
		FinishedAt: normalizeDBUTCGTime(record["finished_at"].GTime()),
		Error:      record["error_message"].String(),
	}
}

// Timeline 工作流事件时间线
func (c *cWorkflow) Timeline(ctx context.Context, req *v1.WorkflowTimelineReq) (res *v1.WorkflowTimelineRes, err error) {
	var (
		workflowEventRepo = repo.NewWorkflowEventRepo()
		workflowRunRepo   = repo.NewWorkflowRunRepo()
	)

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	wfRuns, err := workflowRunRepo.ListByProject(ctx, projectID, "id")
	if err != nil || len(wfRuns) == 0 {
		return &v1.WorkflowTimelineRes{Events: []v1.TimelineEvent{}}, nil
	}

	wfRunIDs := make([]int64, 0, len(wfRuns))
	for _, r := range wfRuns {
		wfRunIDs = append(wfRunIDs, mapToDBRecord(r)["id"].Int64())
	}

	events, err := workflowEventRepo.ListByWorkflowIDs(ctx, wfRunIDs, limit)
	if err != nil {
		return nil, err
	}

	list := make([]v1.TimelineEvent, 0, len(events))
	for _, e := range events {
		list = append(list, buildTimelineEvent(mapToDBRecord(e)))
	}

	return &v1.WorkflowTimelineRes{Events: list}, nil
}

// ReworkStatus 返工阶段状态
func (c *cWorkflow) ReworkStatus(ctx context.Context, req *v1.WorkflowReworkStatusReq) (res *v1.WorkflowReworkStatusRes, err error) {
	var (
		domainTaskRepo    = repo.NewDomainTaskRepo()
		handoffRecordRepo = repo.NewHandoffRecordRepo()
		stageRunRepo      = repo.NewStageRunRepo()
	)

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return &v1.WorkflowReworkStatusRes{HasRework: false, History: []v1.ReworkRoundInfo{}}, nil
	}
	wfRunID := wfRun["id"].Int64()

	handoffs, err := handoffRecordRepo.ListByWorkflowAndType(ctx, wfRunID, "failure_escalation")
	if err != nil {
		return nil, err
	}
	if len(handoffs) == 0 {
		return &v1.WorkflowReworkStatusRes{HasRework: false, History: []v1.ReworkRoundInfo{}}, nil
	}

	var currentStage *v1.ReworkStageInfo
	reworkStage, rsErr := stageRunRepo.GetLatestByWorkflowAndType(ctx, wfRunID, "rework")
	if rsErr != nil {
		g.Log().Warningf(ctx, "[ReworkStatus] 查询 rework stage 失败: %v", rsErr)
	}
	if !reworkStage.IsEmpty() {
		currentStage = &v1.ReworkStageInfo{
			StageRunID: snowflake.JsonInt64(reworkStage["id"].Int64()),
			Status:     reworkStage["status"].String(),
			StartedAt:  normalizeDBUTCGTime(reworkStage["started_at"].GTime()),
		}
	}

	history := make([]v1.ReworkRoundInfo, 0, len(handoffs))
	for i, h := range handoffs {
		record := mapToDBRecord(h)
		fromTaskID := record["from_task_id"].Int64()
		toTaskID := record["to_task_id"].Int64()

		failedTask, ftErr2 := domainTaskRepo.GetByIDMap(ctx, fromTaskID, "name", "result")
		if ftErr2 != nil {
			g.Log().Warningf(ctx, "[ReworkStatus] 查询失败任务详情失败: taskID=%d err=%v", fromTaskID, ftErr2)
		}
		failedName := ""
		failedReason := record["reason"].String()
		if len(failedTask) > 0 {
			failedName = mapToDBRecord(failedTask)["name"].String()
		}

		var analysisID *snowflake.JsonInt64
		analysisResult := ""
		if toTaskID > 0 {
			v := snowflake.JsonInt64(toTaskID)
			analysisID = &v
			analysisTask, atErr := domainTaskRepo.GetByIDMap(ctx, toTaskID, "result")
			if atErr != nil {
				g.Log().Warningf(ctx, "[ReworkStatus] 查询分析任务结果失败: taskID=%d err=%v", toTaskID, atErr)
			}
			if len(analysisTask) > 0 {
				analysisResult = mapToDBRecord(analysisTask)["result"].String()
			}
		}

		history = append(history, v1.ReworkRoundInfo{
			Round:          i + 1,
			FailedTaskID:   snowflake.JsonInt64(fromTaskID),
			FailedTaskName: failedName,
			FailedReason:   failedReason,
			AnalysisTaskID: analysisID,
			AnalysisResult: analysisResult,
			HandoffType:    record["handoff_type"].String(),
			CreatedAt:      normalizeDBUTCGTime(record["created_at"].GTime()),
		})
	}

	return &v1.WorkflowReworkStatusRes{
		HasRework:    true,
		ReworkRounds: len(history),
		CurrentStage: currentStage,
		History:      history,
	}, nil
}

// StageHistory 工作流阶段历史
func (c *cWorkflow) StageHistory(ctx context.Context, req *v1.WorkflowStageHistoryReq) (res *v1.WorkflowStageHistoryRes, err error) {
	stageRunRepo := repo.NewStageRunRepo()

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return &v1.WorkflowStageHistoryRes{Stages: []v1.StageHistoryItem{}}, nil
	}

	stages, err := stageRunRepo.ListByWorkflowMaps(ctx, wfRun["id"].Int64(), "id", "stage_type", "stage_no", "status", "started_at", "finished_at", "error_message")
	if err != nil {
		return nil, err
	}

	list := make([]v1.StageHistoryItem, 0, len(stages))
	for _, s := range stages {
		list = append(list, buildStageHistoryItem(mapToDBRecord(s)))
	}

	return &v1.WorkflowStageHistoryRes{Stages: list}, nil
}

// CompletionSummary 获取项目完成总结
func (c *cWorkflow) CompletionSummary(ctx context.Context, req *v1.WorkflowCompletionSummaryReq) (res *v1.WorkflowCompletionSummaryRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	svc := orchestrator.GetCompleteStageService()
	summary, err := svc.GetSummary(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowCompletionSummaryRes{
		WorkflowRunID:   snowflake.JsonInt64(summary.WorkflowRunID),
		ProjectID:       snowflake.JsonInt64(summary.ProjectID),
		TotalTasks:      summary.TotalTasks,
		CompletedTasks:  summary.CompletedTasks,
		FailedTasks:     summary.FailedTasks,
		EscalatedTasks:  summary.EscalatedTasks,
		SkippedTasks:    summary.SkippedTasks,
		SuccessRate:     summary.SuccessRate,
		TotalDuration:   summary.TotalDuration,
		AvgTaskDuration: summary.AvgTaskDuration,
		StageDurations:  summary.StageDurations,
		ReworkRounds:    summary.ReworkRounds,
		HandoffCount:    summary.HandoffCount,
		StartedAt:       summary.StartedAt,
		FinishedAt:      summary.FinishedAt,
	}, nil
}
