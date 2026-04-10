package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

const (
	manualApproveModePending  = "pending"
	manualApproveModeRejected = "rejected"
)

func buildReviewIssueItem(issue gdb.Record) v1.ReviewIssueItem {
	return v1.ReviewIssueItem{
		ID:         snowflake.JsonInt64(issue["id"].Int64()),
		Severity:   issue["severity"].String(),
		IssueCode:  issue["issue_code"].String(),
		SourceRole: issue["source_role"].String(),
		TaskName:   issue["task_name"].String(),
		Message:    issue["message"].String(),
		Suggestion: issue["suggestion"].String(),
		Status:     issue["status"].String(),
		CreatedAt:  normalizeDBUTCGTime(issue["created_at"].GTime()),
	}
}

func buildReviewStageTask(record gdb.Record) v1.ReviewStageTask {
	return v1.ReviewStageTask{
		ID:           snowflake.JsonInt64(record["id"].Int64()),
		TaskType:     record["task_type"].String(),
		RoleType:     record["role_type"].String(),
		Status:       record["status"].String(),
		StartedAt:    normalizeDBUTCGTime(record["started_at"].GTime()),
		CompletedAt:  normalizeDBUTCGTime(record["completed_at"].GTime()),
		ErrorMessage: record["error_message"].String(),
	}
}

func manualApproveModeForPlanVersion(record gdb.Record) string {
	status := strings.ToLower(strings.TrimSpace(record["status"].String()))
	reviewStatus := strings.ToLower(strings.TrimSpace(record["review_status"].String()))
	switch {
	case status == "active" && reviewStatus == "pending":
		return manualApproveModePending
	case status == "draft" && reviewStatus == "rejected":
		return manualApproveModeRejected
	default:
		return ""
	}
}

func loadManualApprovablePlanVersion(ctx context.Context, projectID int64) (gdb.Record, string, error) {
	plans, err := repo.NewPlanVersionRepo().ListByProjectStatuses(ctx, projectID, []string{"active", "draft"},
		"id", "status", "review_status", "version_no")
	if err != nil {
		return nil, "", err
	}
	for _, plan := range plans {
		if mode := manualApproveModeForPlanVersion(plan); mode != "" {
			return plan, mode, nil
		}
	}
	return nil, "", nil
}

func restoreRejectedPlanVersionForManualApprove(ctx context.Context, planVersionID int64) error {
	if err := repo.NewPlanVersionRepo().RestoreRejectedForManualApprove(ctx, planVersionID); err != nil {
		return fmt.Errorf("恢复被驳回方案版本失败: %w", err)
	}
	return nil
}

func latestReviewStageRunForProject(ctx context.Context, projectID int64, fields ...string) (gdb.Record, error) {
	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return repo.NewStageRunRepo().GetLatestByWorkflowAndType(ctx, wfRun["id"].Int64(), "review", fields...)
}

func latestActiveOrDraftPlanVersion(ctx context.Context, projectID int64, fields ...string) (gdb.Record, error) {
	plans, err := repo.NewPlanVersionRepo().ListByProjectStatuses(ctx, projectID, []string{"draft", "active"}, fields...)
	if err != nil || len(plans) == 0 {
		return nil, err
	}
	return plans[0], nil
}

// buildConfirmPlanResult 查询最新审核结果，组装 ConfirmPlan 响应。
func buildConfirmPlanResult(ctx context.Context, projectID int64) *v1.WorkflowConfirmPlanRes {
	res := &v1.WorkflowConfirmPlanRes{}

	stageRun, err := latestReviewStageRunForProject(ctx, projectID, "id", "status", "error_message")
	if err != nil || stageRun == nil || stageRun.IsEmpty() {
		return res
	}
	stageRunID := stageRun["id"].Int64()

	var countErr error
	reviewIssueRepo := repo.NewReviewIssueRepo()
	res.ErrorCount, countErr = reviewIssueRepo.CountOpenByStageRunAndSeverity(ctx, stageRunID, "error")
	if countErr != nil {
		g.Log().Warningf(ctx, "[ReviewStatus] 统计 error issue 失败: stageRun=%d err=%v", stageRunID, countErr)
	}
	res.WarningCount, countErr = reviewIssueRepo.CountOpenByStageRunAndSeverity(ctx, stageRunID, "warning")
	if countErr != nil {
		g.Log().Warningf(ctx, "[ReviewStatus] 统计 warning issue 失败: stageRun=%d err=%v", stageRunID, countErr)
	}

	issues, _ := reviewIssueRepo.ListByStageRun(ctx, stageRunID, true, 50)

	for _, issue := range issues {
		res.Issues = append(res.Issues, buildReviewIssueItem(issue))
	}

	if stageRun["status"].String() == "failed" {
		res.RejectReason = stageRun["error_message"].String()
	}

	return res
}

// parseAndCreateBlueprints V2 专用：解析 AI 回复并创建蓝图。
func parseAndCreateBlueprints(ctx context.Context, projectID, conversationID, messageID int64, aiReply string) (int, error) {
	project, _ := repo.NewProjectRepo().GetByID(ctx, projectID, "project_category")
	projectCategory := g.NewVar(project["project_category"]).String()

	tasks, report, err := engine.GetParser().ExtractAndNormalizeWithReport(ctx, aiReply, projectCategory)
	if report != nil && report.HasBlockingIssue() {
		engine.NotifyProjectArchitectConversation(ctx, projectID, report.BuildContinuationPrompt())
		return 0, fmt.Errorf("任务清单存在阻断问题: %s", report.Summary())
	}
	if err != nil || len(tasks) == 0 {
		return 0, err
	}

	return createBlueprints(ctx, projectID, conversationID, messageID, tasks)
}

// createBlueprints 将已提取的任务列表写入 plan_version + task_blueprint。
func createBlueprints(ctx context.Context, projectID, conversationID, messageID int64, tasks []engine.ArchitectTask) (int, error) {
	wfRun, _ := repo.NewWorkflowRunRepo().GetLatestByProjectExcludingStatuses(ctx, projectID, []string{"completed", "canceled"}, "id")
	var wfRunID int64
	if len(wfRun) != 0 {
		wfRunID = g.NewVar(wfRun["id"]).Int64()
	}

	pvSvc := orchestrator.GetPlanVersionService()
	_, bpCount, err := pvSvc.CreateFromArchitectReply(ctx, projectID, wfRunID, conversationID, messageID, tasks)
	if err != nil {
		return 0, err
	}
	return bpCount, nil
}

// ReviewStatus 获取项目审核状态（V2 专用）
func (c *cWorkflow) ReviewStatus(ctx context.Context, req *v1.WorkflowReviewStatusReq) (res *v1.WorkflowReviewStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowReviewStatusRes{}

	pv, err := latestActiveOrDraftPlanVersion(ctx, projectID, "id", "review_status")
	if err != nil || pv == nil || pv.IsEmpty() {
		return res, nil
	}
	pvID := pv["id"].Int64()
	res.PlanVersionID = snowflake.JsonInt64(pvID)
	res.ReviewStatus = pv["review_status"].String()

	bpCount, _ := repo.NewBlueprintRepo().CountByPlanVersion(ctx, pvID)
	res.BlueprintCount = bpCount

	stageRun, _ := latestReviewStageRunForProject(ctx, projectID)
	if stageRun != nil && !stageRun.IsEmpty() {
		stageRunID := stageRun["id"].Int64()
		res.StageRunID = snowflake.JsonInt64(stageRunID)
		res.StageStatus = stageRun["status"].String()

		var stageTasks []v1.ReviewStageTask
		tasks, _ := repo.NewStageTaskRepo().ListByStageRun(ctx, stageRunID)
		for _, t := range tasks {
			stageTasks = append(stageTasks, buildReviewStageTask(t))
		}
		res.StageTasks = stageTasks

		reviewIssueRepo := repo.NewReviewIssueRepo()
		res.ErrorCount, _ = reviewIssueRepo.CountOpenByStageRunAndSeverity(ctx, stageRunID, "error")
		res.WarningCount, _ = reviewIssueRepo.CountOpenByStageRunAndSeverity(ctx, stageRunID, "warning")
	}

	return res, nil
}

// ReviewIssues 获取审核问题列表
func (c *cWorkflow) ReviewIssues(ctx context.Context, req *v1.WorkflowReviewIssuesReq) (res *v1.WorkflowReviewIssuesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	stageRun, _ := latestReviewStageRunForProject(ctx, projectID, "id")
	if stageRun == nil || stageRun.IsEmpty() {
		return &v1.WorkflowReviewIssuesRes{Issues: []v1.ReviewIssueItem{}}, nil
	}

	issues, _ := repo.NewReviewIssueRepo().ListByStageRun(ctx, stageRun["id"].Int64(), false, 0)

	items := make([]v1.ReviewIssueItem, 0, len(issues))
	for _, issue := range issues {
		items = append(items, buildReviewIssueItem(issue))
	}

	return &v1.WorkflowReviewIssuesRes{Issues: items}, nil
}

// ManualApprove 手动审批通过
func (c *cWorkflow) ManualApprove(ctx context.Context, req *v1.WorkflowManualApproveReq) (res *v1.WorkflowManualApproveRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	pv, approveMode, err := loadManualApprovablePlanVersion(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if pv == nil || pv.IsEmpty() {
		return nil, fmt.Errorf("没有可人工审批通过的方案版本")
	}

	pvSvc := orchestrator.GetPlanVersionService()
	planVersionID := pv["id"].Int64()
	switch approveMode {
	case manualApproveModePending:
		if err := pvSvc.Approve(ctx, planVersionID); err != nil {
			return nil, err
		}
	case manualApproveModeRejected:
		if err := restoreRejectedPlanVersionForManualApprove(ctx, planVersionID); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("方案版本当前状态不支持人工审批通过")
	}

	wfRun, _ := repo.NewWorkflowRunRepo().GetLatestByProjectExcludingStatuses(ctx, projectID, []string{"completed", "canceled"}, "id", "current_stage_run_id")
	if len(wfRun) != 0 {
		wfRunID := g.NewVar(wfRun["id"]).Int64()
		currentStageRunID := g.NewVar(wfRun["current_stage_run_id"]).Int64()
		if approveMode == manualApproveModeRejected {
			if _, upErr := repo.NewWorkflowRunRepo().UpdateStatus(ctx, wfRunID, "designing", "reviewing", nil); upErr != nil {
				return nil, fmt.Errorf("恢复工作流审核状态失败: %w", upErr)
			}
			if _, upErr := repo.NewProjectRepo().UpdateStatusIfCurrent(ctx, projectID, "designing", "reviewing"); upErr != nil {
				return nil, fmt.Errorf("恢复项目审核状态失败: %w", upErr)
			}
		}

		if currentStageRunID > 0 {
			stgSvc := orchestrator.GetStageService()
			_ = stgSvc.CompleteStage(ctx, currentStageRunID)
		}

		execSvc := orchestrator.GetExecuteStageService()
		stgSvc := orchestrator.GetStageService()
		execStageRunID, err2 := stgSvc.StartStage(ctx, wfRunID, "execute")
		if err2 != nil {
			return nil, fmt.Errorf("审核已通过，但创建执行阶段失败: %w", err2)
		}
		if err3 := execSvc.InstantiateAndStart(ctx, execStageRunID, planVersionID); err3 != nil {
			_ = stgSvc.FailStage(ctx, execStageRunID, err3.Error())
			return nil, fmt.Errorf("审核已通过，但执行阶段启动失败: %w", err3)
		}
	}

	return &v1.WorkflowManualApproveRes{}, nil
}

// ManualReject 手动驳回
func (c *cWorkflow) ManualReject(ctx context.Context, req *v1.WorkflowManualRejectReq) (res *v1.WorkflowManualRejectRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	pv, err := repo.NewPlanVersionRepo().GetLatestByProjectStatusAndReviewStatus(ctx, projectID, "active", "pending", "id")
	if err != nil || pv == nil || pv.IsEmpty() {
		return nil, fmt.Errorf("没有待审核的方案版本")
	}

	pvSvc := orchestrator.GetPlanVersionService()
	if err := pvSvc.Reject(ctx, pv["id"].Int64()); err != nil {
		return nil, err
	}

	if upErr := repo.NewProjectRepo().UpdateStatus(ctx, projectID, "designing"); upErr != nil {
		g.Log().Errorf(ctx, "[ManualReject] 项目状态回退失败: project=%d err=%v", projectID, upErr)
	}

	return &v1.WorkflowManualRejectRes{}, nil
}

// ReviewIssueReplan 将审核问题转为方案修订。
func (c *cWorkflow) ReviewIssueReplan(ctx context.Context, req *v1.WorkflowReviewIssueReplanReq) (res *v1.WorkflowReviewIssueReplanRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	if len(req.IssueIDs) == 0 {
		return nil, fmt.Errorf("请选择至少一条审核问题")
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	workflowRunID := wfRun["id"].Int64()

	stageRun, stageErr := repo.NewStageRunRepo().GetLatestByWorkflowAndType(ctx, workflowRunID, "review", "id")
	if stageErr != nil || stageRun == nil || stageRun.IsEmpty() {
		return nil, fmt.Errorf("当前工作流没有审核阶段记录")
	}
	stageRunID := stageRun["id"].Int64()

	issueIDs := make([]int64, 0, len(req.IssueIDs))
	for _, id := range req.IssueIDs {
		if int64(id) > 0 {
			issueIDs = append(issueIDs, int64(id))
		}
	}
	issues, err := repo.NewReviewIssueRepo().ListOpenByStageRunAndIDs(ctx, stageRunID, issueIDs)
	if err != nil {
		return nil, fmt.Errorf("查询审核问题失败: %w", err)
	}
	if len(issues) == 0 {
		return nil, fmt.Errorf("未找到可回流的审核问题")
	}

	reason := buildReviewIssueReplanReason(issues, req.Reason)
	if _, err := c.ManualReject(ctx, &v1.WorkflowManualRejectReq{
		ProjectID: req.ProjectID,
		Reason:    reason,
	}); err != nil {
		return nil, err
	}

	recordWorkflowEvent(ctx, workflowRunID, "review_issue", "review.issue_replan_requested", nil, &stageRunID, map[string]interface{}{
		"issue_ids": jsonInt64SliceToInt64(issueIDs),
		"reason":    reason,
	})

	return &v1.WorkflowReviewIssueReplanRes{
		Message: fmt.Sprintf("已基于 %d 条审核问题发起方案修订", len(issues)),
	}, nil
}

func buildReviewIssueReplanReason(issues gdb.Result, extraReason string) string {
	lines := []string{"基于审核问题发起方案修订："}
	for i, issue := range issues {
		if i >= 5 {
			lines = append(lines, fmt.Sprintf("其余 %d 条问题请查看审核问题列表。", len(issues)-i))
			break
		}
		line := fmt.Sprintf(
			"%d. [%s/%s] %s",
			i+1,
			issue["severity"].String(),
			issue["issue_code"].String(),
			issue["message"].String(),
		)
		if suggestion := strings.TrimSpace(issue["suggestion"].String()); suggestion != "" {
			line += "；建议：" + suggestion
		}
		if taskName := strings.TrimSpace(issue["task_name"].String()); taskName != "" {
			line += "；关联蓝图：" + taskName
		}
		lines = append(lines, line)
	}
	if extraReason = strings.TrimSpace(extraReason); extraReason != "" {
		lines = append(lines, "附加说明："+extraReason)
	}
	return strings.Join(lines, "\n")
}
