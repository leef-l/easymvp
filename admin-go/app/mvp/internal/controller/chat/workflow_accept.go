package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

func acceptRelatedWorkflowStatuses() g.Slice {
	return g.Slice{
		consts.WorkflowRunStatusAccepting,
		consts.WorkflowRunStatusExecuting,
		consts.WorkflowRunStatusReworking,
		consts.WorkflowRunStatusPaused,
		consts.WorkflowRunStatusCompleted,
		"running", // 兼容历史脏数据
	}
}

func loadLatestAcceptBundle(ctx context.Context, projectID int64) (int64, g.Map, []g.Map, []g.Map, error) {
	workflowRun, err := repo.NewWorkflowRunRepo().GetLatestByProjectStatuses(ctx, projectID, []string{
		consts.WorkflowRunStatusAccepting,
		consts.WorkflowRunStatusExecuting,
		consts.WorkflowRunStatusReworking,
		consts.WorkflowRunStatusPaused,
		consts.WorkflowRunStatusCompleted,
		"running",
	}, "id")
	if err != nil {
		return 0, nil, nil, nil, err
	}
	if len(workflowRun) == 0 {
		return 0, nil, nil, nil, nil
	}

	workflowRunID := g.NewVar(workflowRun["id"]).Int64()
	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(acceptRun) == 0 {
		return workflowRunID, nil, nil, nil, err
	}

	acceptRunID := g.NewVar(acceptRun["id"]).Int64()
	issueRepo := repo.NewAcceptIssueRepo()
	issues, err := issueRepo.ListByAcceptRun(ctx, acceptRunID)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	evidenceRepo := repo.NewAcceptEvidenceRepo()
	evidence, err := evidenceRepo.ListByAcceptRun(ctx, acceptRunID)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	return workflowRunID, acceptRun, issues, evidence, nil
}

// AcceptStatus 验收状态总览
func (c *cWorkflow) AcceptStatus(ctx context.Context, req *v1.WorkflowAcceptStatusReq) (res *v1.WorkflowAcceptStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	workflowRunID, acceptRun, issues, evidenceList, err := loadLatestAcceptBundle(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if workflowRunID == 0 {
		return nil, fmt.Errorf("项目无工作流运行")
	}
	if len(acceptRun) == 0 {
		return &v1.WorkflowAcceptStatusRes{Status: "none"}, nil
	}

	acceptRunID := acceptRun["id"]
	acceptRunIDInt := g.NewVar(acceptRunID).Int64()

	var blockers, errors, warns, infos int
	for _, issue := range issues {
		switch g.NewVar(issue["severity"]).String() {
		case "blocker":
			blockers++
		case "error":
			errors++
		case "warn":
			warns++
		case "info":
			infos++
		}
	}

	res = &v1.WorkflowAcceptStatusRes{
		AcceptRunID:   snowflake.JsonInt64(acceptRunIDInt),
		WorkflowRunID: snowflake.JsonInt64(workflowRunID),
		AcceptRound:   g.NewVar(acceptRun["accept_round"]).Int(),
		Status:        normalizeAcceptRunStatus(g.NewVar(acceptRun["status"]).String(), g.NewVar(acceptRun["decision"]).String(), g.NewVar(acceptRun["finished_at"]).GTime()),
		Decision:      g.NewVar(acceptRun["decision"]).String(),
		Score:         g.NewVar(acceptRun["score"]).Float64(),
		Summary:       g.NewVar(acceptRun["summary"]).String(),
		RulesSnapshot: g.NewVar(acceptRun["rules_snapshot_ref"]).String(),
		StartedAt:     normalizeDBUTCGTime(g.NewVar(acceptRun["started_at"]).GTime()),
		FinishedAt:    normalizeDBUTCGTime(g.NewVar(acceptRun["finished_at"]).GTime()),
		BlockerCount:  blockers,
		ErrorCount:    errors,
		WarnCount:     warns,
		InfoCount:     infos,
		EvidenceCount: len(evidenceList),
	}
	return res, nil
}

func normalizeAcceptRunStatus(status, decision string, finishedAt *gtime.Time) string {
	status = strings.ToLower(strings.TrimSpace(status))
	decision = strings.ToLower(strings.TrimSpace(decision))
	if status == "running" && finishedAt != nil {
		switch decision {
		case "passed", "failed", "manual_review":
			return "completed"
		}
	}
	return status
}

// AcceptIssues 验收问题列表
func (c *cWorkflow) AcceptIssues(ctx context.Context, req *v1.WorkflowAcceptIssuesReq) (res *v1.WorkflowAcceptIssuesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	_, acceptRun, issues, _, err := loadLatestAcceptBundle(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(acceptRun) == 0 {
		return &v1.WorkflowAcceptIssuesRes{Issues: []v1.AcceptIssueItem{}}, nil
	}

	var items []v1.AcceptIssueItem
	for _, issue := range issues {
		severity := g.NewVar(issue["severity"]).String()
		if req.Severity != "" && severity != req.Severity {
			continue
		}
		items = append(items, v1.AcceptIssueItem{
			ID:              snowflake.JsonInt64(g.NewVar(issue["id"]).Int64()),
			IssueType:       g.NewVar(issue["issue_type"]).String(),
			RuleCode:        g.NewVar(issue["rule_code"]).String(),
			Severity:        severity,
			Title:           g.NewVar(issue["title"]).String(),
			Detail:          g.NewVar(issue["detail"]).String(),
			ExpectedValue:   g.NewVar(issue["expected_value"]).String(),
			ActualValue:     g.NewVar(issue["actual_value"]).String(),
			SuggestedAction: g.NewVar(issue["suggested_action"]).String(),
			DomainTaskID:    snowflake.JsonInt64(g.NewVar(issue["domain_task_id"]).Int64()),
			ResourceRef:     g.NewVar(issue["resource_ref"]).String(),
			Status:          g.NewVar(issue["status"]).String(),
			CreatedAt:       normalizeDBUTCGTime(g.NewVar(issue["created_at"]).GTime()),
		})
	}
	if items == nil {
		items = []v1.AcceptIssueItem{}
	}
	return &v1.WorkflowAcceptIssuesRes{Issues: items}, nil
}

// AcceptEvidence 验收证据列表
func (c *cWorkflow) AcceptEvidence(ctx context.Context, req *v1.WorkflowAcceptEvidenceReq) (res *v1.WorkflowAcceptEvidenceRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	_, acceptRun, _, evidenceList, err := loadLatestAcceptBundle(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(acceptRun) == 0 {
		return &v1.WorkflowAcceptEvidenceRes{Evidence: []v1.AcceptEvidenceItem{}}, nil
	}

	var items []v1.AcceptEvidenceItem
	for _, e := range evidenceList {
		items = append(items, v1.AcceptEvidenceItem{
			ID:           snowflake.JsonInt64(g.NewVar(e["id"]).Int64()),
			EvidenceType: g.NewVar(e["evidence_type"]).String(),
			SourceType:   g.NewVar(e["source_type"]).String(),
			SourceID:     snowflake.JsonInt64(g.NewVar(e["source_id"]).Int64()),
			ContentRef:   g.NewVar(e["content_ref"]).String(),
			Summary:      g.NewVar(e["summary"]).String(),
			CreatedAt:    normalizeDBUTCGTime(g.NewVar(e["created_at"]).GTime()),
		})
	}
	if items == nil {
		items = []v1.AcceptEvidenceItem{}
	}
	return &v1.WorkflowAcceptEvidenceRes{Evidence: items}, nil
}

// AcceptApprove 人工放行
func (c *cWorkflow) AcceptApprove(ctx context.Context, req *v1.WorkflowAcceptApproveReq) (res *v1.WorkflowAcceptApproveRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualApprove(ctx, projectID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptApproveRes{}, nil
}

// AcceptReject 驳回验收
func (c *cWorkflow) AcceptReject(ctx context.Context, req *v1.WorkflowAcceptRejectReq) (res *v1.WorkflowAcceptRejectRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualReject(ctx, projectID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptRejectRes{}, nil
}

// AcceptRerun 重新验收
func (c *cWorkflow) AcceptRerun(ctx context.Context, req *v1.WorkflowAcceptRerunReq) (res *v1.WorkflowAcceptRerunRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.Rerun(ctx, projectID); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptRerunRes{}, nil
}

// AcceptRework 驳回并返工
func (c *cWorkflow) AcceptRework(ctx context.Context, req *v1.WorkflowAcceptReworkReq) (res *v1.WorkflowAcceptReworkRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualRework(ctx, projectID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptReworkRes{}, nil
}

// AcceptIssueRework 将验收问题转为正式返工。
func (c *cWorkflow) AcceptIssueRework(ctx context.Context, req *v1.WorkflowAcceptIssueReworkReq) (res *v1.WorkflowAcceptIssueReworkRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	if len(req.IssueIDs) == 0 {
		return nil, fmt.Errorf("请选择至少一条验收问题")
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	workflowRunID := wfRun["id"].Int64()

	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(acceptRun) == 0 {
		return nil, fmt.Errorf("当前工作流没有验收记录")
	}

	issueIDs := make([]int64, 0, len(req.IssueIDs))
	for _, id := range req.IssueIDs {
		if int64(id) > 0 {
			issueIDs = append(issueIDs, int64(id))
		}
	}
	issues, err := repo.NewAcceptIssueRepo().ListOpenByAcceptRunAndIDs(ctx, g.NewVar(acceptRun["id"]).Int64(), issueIDs)
	if err != nil {
		return nil, fmt.Errorf("查询验收问题失败: %w", err)
	}
	if len(issues) == 0 {
		return nil, fmt.Errorf("未找到可回流的验收问题")
	}

	reason := buildAcceptIssueReworkReason(issues, req.Reason)
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualRework(ctx, projectID, reason); err != nil {
		return nil, err
	}

	stageRunID := g.NewVar(acceptRun["stage_run_id"]).Int64()
	recordWorkflowEvent(ctx, workflowRunID, "accept_issue", "accept.issue_rework_requested", nil, &stageRunID, map[string]interface{}{
		"issue_ids": jsonInt64SliceToInt64(issueIDs),
		"reason":    reason,
	})

	return &v1.WorkflowAcceptIssueReworkRes{
		Message: fmt.Sprintf("已基于 %d 条验收问题触发返工", len(issues)),
	}, nil
}

func buildAcceptIssueReworkReason(issues []g.Map, extraReason string) string {
	lines := []string{"基于验收问题触发返工："}
	for i, issue := range issues {
		if i >= 5 {
			lines = append(lines, fmt.Sprintf("其余 %d 条问题请查看验收问题列表。", len(issues)-i))
			break
		}
		line := fmt.Sprintf(
			"%d. [%s] %s",
			i+1,
			g.NewVar(issue["rule_code"]).String(),
			g.NewVar(issue["title"]).String(),
		)
		if detail := strings.TrimSpace(g.NewVar(issue["detail"]).String()); detail != "" {
			line += " - " + detail
		}
		if action := strings.TrimSpace(g.NewVar(issue["suggested_action"]).String()); action != "" {
			line += "；建议：" + action
		}
		lines = append(lines, line)
	}
	if extraReason = strings.TrimSpace(extraReason); extraReason != "" {
		lines = append(lines, "附加说明："+extraReason)
	}
	return strings.Join(lines, "\n")
}

func jsonInt64SliceToInt64(ids []int64) []int64 {
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			result = append(result, id)
		}
	}
	return result
}
