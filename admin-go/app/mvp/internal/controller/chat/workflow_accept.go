package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// AcceptStatus 验收状态总览
func (c *cWorkflow) AcceptStatus(ctx context.Context, req *v1.WorkflowAcceptStatusReq) (res *v1.WorkflowAcceptStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused", "completed"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return nil, fmt.Errorf("项目无工作流运行")
	}
	workflowRunID := wfRun["id"].Int64()

	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(acceptRun) == 0 {
		return &v1.WorkflowAcceptStatusRes{Status: "none"}, nil
	}

	acceptRunID := acceptRun["id"]
	acceptRunIDInt := g.NewVar(acceptRunID).Int64()

	issueRepo := repo.NewAcceptIssueRepo()
	issues, _ := issueRepo.ListByAcceptRun(ctx, acceptRunIDInt)
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

	evidenceRepo := repo.NewAcceptEvidenceRepo()
	evidenceList, _ := evidenceRepo.ListByAcceptRun(ctx, acceptRunIDInt)

	res = &v1.WorkflowAcceptStatusRes{
		AcceptRunID:   snowflake.JsonInt64(acceptRunIDInt),
		WorkflowRunID: snowflake.JsonInt64(workflowRunID),
		AcceptRound:   g.NewVar(acceptRun["accept_round"]).Int(),
		Status:        g.NewVar(acceptRun["status"]).String(),
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

// AcceptIssues 验收问题列表
func (c *cWorkflow) AcceptIssues(ctx context.Context, req *v1.WorkflowAcceptIssuesReq) (res *v1.WorkflowAcceptIssuesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused", "completed"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return &v1.WorkflowAcceptIssuesRes{Issues: []v1.AcceptIssueItem{}}, nil
	}

	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, wfRun["id"].Int64())
	if err != nil || len(acceptRun) == 0 {
		return &v1.WorkflowAcceptIssuesRes{Issues: []v1.AcceptIssueItem{}}, nil
	}

	issueRepo := repo.NewAcceptIssueRepo()
	issues, err := issueRepo.ListByAcceptRun(ctx, g.NewVar(acceptRun["id"]).Int64())
	if err != nil {
		return nil, err
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

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused", "completed"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return &v1.WorkflowAcceptEvidenceRes{Evidence: []v1.AcceptEvidenceItem{}}, nil
	}

	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, wfRun["id"].Int64())
	if err != nil || len(acceptRun) == 0 {
		return &v1.WorkflowAcceptEvidenceRes{Evidence: []v1.AcceptEvidenceItem{}}, nil
	}

	evidenceRepo := repo.NewAcceptEvidenceRepo()
	evidenceList, err := evidenceRepo.ListByAcceptRun(ctx, g.NewVar(acceptRun["id"]).Int64())
	if err != nil {
		return nil, err
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
	issues, err := g.DB().Model("mvp_accept_issue").Ctx(ctx).
		Where("accept_run_id", g.NewVar(acceptRun["id"]).Int64()).
		WhereIn("id", issueIDs).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("severity").
		OrderDesc("created_at").
		All()
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

func buildAcceptIssueReworkReason(issues gdb.Result, extraReason string) string {
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
