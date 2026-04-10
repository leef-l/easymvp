package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	verificationsvc "easymvp/app/mvp/internal/workflow/verification"
	"easymvp/utility/snowflake"
)

// VerificationStart 启动 Docker-first 项目验证。
func (c *cWorkflow) VerificationStart(ctx context.Context, req *v1.WorkflowVerificationStartReq) (res *v1.WorkflowVerificationStartRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)
	runID, workflowRunID, err := startVerificationRun(ctx, projectID, userID, deptID, "manual", req.Reason)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowVerificationStartRes{
		VerificationRunID: snowflake.JsonInt64(runID),
		WorkflowRunID:     snowflake.JsonInt64(workflowRunID),
		Status:            "running",
		Message:           "验证已启动，系统将优先尝试 Docker 环境并回写问题与证据",
	}, nil
}

// VerificationStatus 验证状态总览。
func (c *cWorkflow) VerificationStatus(ctx context.Context, req *v1.WorkflowVerificationStatusReq) (res *v1.WorkflowVerificationStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	workflowRunID, verificationRun, issues, evidence, err := loadLatestVerificationBundle(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(verificationRun) == 0 {
		return &v1.WorkflowVerificationStatusRes{Status: "none"}, nil
	}

	blockers, errorsCount, warns, infos := countVerificationSeverities(issues)
	return &v1.WorkflowVerificationStatusRes{
		VerificationRunID: snowflake.JsonInt64(g.NewVar(verificationRun["id"]).Int64()),
		WorkflowRunID:     snowflake.JsonInt64(workflowRunID),
		VerificationRound: g.NewVar(verificationRun["verification_round"]).Int(),
		Status:            g.NewVar(verificationRun["status"]).String(),
		Decision:          g.NewVar(verificationRun["decision"]).String(),
		RunnerType:        g.NewVar(verificationRun["runner_type"]).String(),
		TriggerSource:     g.NewVar(verificationRun["trigger_source"]).String(),
		Summary:           g.NewVar(verificationRun["summary"]).String(),
		StartedAt:         normalizeDBUTCGTime(g.NewVar(verificationRun["started_at"]).GTime()),
		FinishedAt:        normalizeDBUTCGTime(g.NewVar(verificationRun["finished_at"]).GTime()),
		BlockerCount:      blockers,
		ErrorCount:        errorsCount,
		WarnCount:         warns,
		InfoCount:         infos,
		EvidenceCount:     len(evidence),
	}, nil
}

// VerificationIssues 验证问题列表。
func (c *cWorkflow) VerificationIssues(ctx context.Context, req *v1.WorkflowVerificationIssuesReq) (res *v1.WorkflowVerificationIssuesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	_, verificationRun, issues, _, err := loadLatestVerificationBundle(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(verificationRun) == 0 {
		return &v1.WorkflowVerificationIssuesRes{Issues: []v1.VerificationIssueItem{}}, nil
	}

	items := make([]v1.VerificationIssueItem, 0, len(issues))
	for _, issue := range issues {
		severity := g.NewVar(issue["severity"]).String()
		if req.Severity != "" && severity != req.Severity {
			continue
		}
		items = append(items, v1.VerificationIssueItem{
			ID:              snowflake.JsonInt64(g.NewVar(issue["id"]).Int64()),
			IssueType:       g.NewVar(issue["issue_type"]).String(),
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
		items = []v1.VerificationIssueItem{}
	}
	return &v1.WorkflowVerificationIssuesRes{Issues: items}, nil
}

// VerificationEvidence 验证证据列表。
func (c *cWorkflow) VerificationEvidence(ctx context.Context, req *v1.WorkflowVerificationEvidenceReq) (res *v1.WorkflowVerificationEvidenceRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	_, verificationRun, _, evidenceList, err := loadLatestVerificationBundle(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(verificationRun) == 0 {
		return &v1.WorkflowVerificationEvidenceRes{Evidence: []v1.VerificationEvidenceItem{}}, nil
	}

	items := make([]v1.VerificationEvidenceItem, 0, len(evidenceList))
	for _, item := range evidenceList {
		items = append(items, v1.VerificationEvidenceItem{
			ID:           snowflake.JsonInt64(g.NewVar(item["id"]).Int64()),
			EvidenceType: g.NewVar(item["evidence_type"]).String(),
			SourceType:   g.NewVar(item["source_type"]).String(),
			SourceID:     snowflake.JsonInt64(g.NewVar(item["source_id"]).Int64()),
			ContentRef:   g.NewVar(item["content_ref"]).String(),
			Summary:      g.NewVar(item["summary"]).String(),
			CreatedAt:    normalizeDBUTCGTime(g.NewVar(item["created_at"]).GTime()),
		})
	}
	if items == nil {
		items = []v1.VerificationEvidenceItem{}
	}
	return &v1.WorkflowVerificationEvidenceRes{Evidence: items}, nil
}

// VerificationRepair 将验证问题转为返工。
func (c *cWorkflow) VerificationRepair(ctx context.Context, req *v1.WorkflowVerificationRepairReq) (res *v1.WorkflowVerificationRepairRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	if len(req.IssueIDs) == 0 {
		return nil, fmt.Errorf("请选择至少一条验证问题")
	}

	issueIDs := make([]int64, 0, len(req.IssueIDs))
	for _, id := range req.IssueIDs {
		if int64(id) > 0 {
			issueIDs = append(issueIDs, int64(id))
		}
	}
	message, err := requestVerificationRepair(ctx, projectID, issueIDs, req.Reason)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowVerificationRepairRes{Message: message}, nil
}

func startVerificationRun(ctx context.Context, projectID, userID, deptID int64, triggerSource, reason string) (int64, int64, error) {
	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return 0, 0, err
	}
	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Fields("created_by, dept_id").
		One()
	if err != nil {
		return 0, 0, fmt.Errorf("查询项目信息失败: %w", err)
	}
	if userID == 0 {
		userID = project["created_by"].Int64()
	}
	if deptID == 0 {
		deptID = project["dept_id"].Int64()
	}

	svc := verificationsvc.NewService(nil, nil, nil)
	runID, err := svc.Start(ctx, verificationsvc.StartRequest{
		ProjectID:     projectID,
		WorkflowRunID: wfRun["id"].Int64(),
		CreatedBy:     userID,
		DeptID:        deptID,
		TriggerSource: triggerSource,
		Reason:        reason,
	})
	if err != nil {
		return 0, 0, err
	}
	return runID, wfRun["id"].Int64(), nil
}

func loadLatestVerificationBundle(ctx context.Context, projectID int64) (int64, g.Map, []g.Map, []g.Map, error) {
	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	workflowRunID := wfRun["id"].Int64()

	runRepo := repo.NewVerificationRunRepo()
	verificationRun, err := runRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(verificationRun) == 0 {
		return workflowRunID, nil, nil, nil, nil
	}

	issueRepo := repo.NewVerificationIssueRepo()
	issues, err := issueRepo.ListByVerificationRun(ctx, g.NewVar(verificationRun["id"]).Int64())
	if err != nil {
		return 0, nil, nil, nil, err
	}

	evidenceRepo := repo.NewVerificationEvidenceRepo()
	evidence, err := evidenceRepo.ListByVerificationRun(ctx, g.NewVar(verificationRun["id"]).Int64())
	if err != nil {
		return 0, nil, nil, nil, err
	}
	return workflowRunID, verificationRun, issues, evidence, nil
}

func requestVerificationRepair(ctx context.Context, projectID int64, issueIDs []int64, extraReason string) (string, error) {
	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return "", err
	}
	workflowRunID := wfRun["id"].Int64()

	runRepo := repo.NewVerificationRunRepo()
	verificationRun, err := runRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(verificationRun) == 0 {
		return "", fmt.Errorf("当前工作流没有验证记录")
	}
	verificationRunID := g.NewVar(verificationRun["id"]).Int64()

	issues, err := g.DB().Model("mvp_verification_issue").Ctx(ctx).
		Where("verification_run_id", verificationRunID).
		WhereIn("id", issueIDs).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if err != nil {
		return "", fmt.Errorf("查询验证问题失败: %w", err)
	}
	if len(issues) == 0 {
		return "", fmt.Errorf("未找到可返工的验证问题")
	}

	failedTaskID, resolvedIssueTaskIDs, err := resolveVerificationRepairTask(ctx, workflowRunID, issues)
	if err != nil {
		return "", err
	}
	if len(resolvedIssueTaskIDs) > 0 {
		for issueID, taskID := range resolvedIssueTaskIDs {
			if taskID <= 0 {
				continue
			}
			if _, upErr := g.DB().Model("mvp_verification_issue").Ctx(ctx).
				Where("id", issueID).
				Where("verification_run_id", verificationRunID).
				Data(g.Map{"domain_task_id": taskID}).
				Update(); upErr != nil {
				g.Log().Warningf(ctx, "[VerificationRepair] 回写 issue 任务归属失败: issue=%d task=%d err=%v", issueID, taskID, upErr)
			}
		}
	}

	task, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", failedTaskID).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("id, name, status").
		One()
	if err != nil {
		return "", fmt.Errorf("查询返工任务失败: %w", err)
	}
	if task.IsEmpty() {
		return "", fmt.Errorf("返工任务 %d 不存在", failedTaskID)
	}
	if task["status"].String() == "running" {
		return "", fmt.Errorf("任务 %d 当前仍在运行，不能直接进入返工", failedTaskID)
	}

	reason := buildVerificationRepairReason(issues, extraReason)
	now := gtime.Now()
	if _, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", failedTaskID).
		Data(g.Map{
			"status":       "failed",
			"result":       reason,
			"started_at":   nil,
			"completed_at": nil,
			"updated_at":   now,
		}).
		Update(); err != nil {
		return "", fmt.Errorf("标记返工任务失败: %w", err)
	}

	stageSvc := orchestrator.GetStageService()
	reworkSvc := orchestrator.GetReworkStageService()
	if stageSvc == nil || reworkSvc == nil {
		return "", fmt.Errorf("返工服务未初始化")
	}

	stageRunID, err := stageSvc.ForceStartStage(ctx, workflowRunID, "rework", reason)
	if err != nil {
		return "", fmt.Errorf("创建 rework 阶段失败: %w", err)
	}
	if err := reworkSvc.HandleReworkWithSource(ctx, stageRunID, failedTaskID, "execute"); err != nil {
		_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
		return "", fmt.Errorf("启动 rework 阶段失败: %w", err)
	}
	if err := orchestrator.ActivateReworkStage(ctx, workflowRunID, stageRunID); err != nil {
		_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
		return "", fmt.Errorf("启动 rework 调度器失败: %w", err)
	}

	if _, err := g.DB().Model("mvp_verification_issue").Ctx(ctx).
		Where("verification_run_id", verificationRunID).
		WhereIn("id", issueIDs).
		Where("status", "open").
		Update(g.Map{"status": "rework_requested", "updated_at": now}); err != nil {
		g.Log().Warningf(ctx, "[VerificationRepair] 更新 issue 状态失败: run=%d err=%v", verificationRunID, err)
	}

	recordWorkflowEvent(ctx, workflowRunID, "verification_issue", "verification.repair_requested", nil, &stageRunID, map[string]interface{}{
		"verification_run_id": verificationRunID,
		"issue_ids":           jsonInt64SliceToInt64(issueIDs),
		"failed_task_id":      failedTaskID,
		"reason":              reason,
	})

	return fmt.Sprintf("已基于 %d 条验证问题触发返工，关联任务 %d", len(issues), failedTaskID), nil
}

func resolveVerificationRepairTask(ctx context.Context, workflowRunID int64, issues gdb.Result) (int64, map[int64]int64, error) {
	resolved := make(map[int64]int64, len(issues))
	targetSet := make(map[int64]struct{}, len(issues))

	for _, issue := range issues {
		taskID, err := verificationsvc.ResolveIssueTaskID(
			ctx,
			workflowRunID,
			issue["domain_task_id"].Int64(),
			issue["title"].String(),
			issue["detail"].String(),
			issue["resource_ref"].String(),
		)
		if err != nil {
			g.Log().Warningf(ctx, "[VerificationRepair] 解析 issue 任务归属失败: issue=%d err=%v", issue["id"].Int64(), err)
			taskID = issue["domain_task_id"].Int64()
		}
		if taskID <= 0 {
			continue
		}
		resolved[issue["id"].Int64()] = taskID
		targetSet[taskID] = struct{}{}
	}

	if len(targetSet) == 0 {
		return 0, resolved, fmt.Errorf("选中的验证问题没有关联可返工任务")
	}
	if len(targetSet) == 1 {
		for taskID := range targetSet {
			return taskID, resolved, nil
		}
	}

	taskIDs := make([]int64, 0, len(targetSet))
	for taskID := range targetSet {
		taskIDs = append(taskIDs, taskID)
	}
	names := loadVerificationRepairTaskNames(ctx, taskIDs)
	return 0, resolved, fmt.Errorf("选中的验证问题分属多个任务，请分别返工：%s", strings.Join(names, " / "))
}

func loadVerificationRepairTaskNames(ctx context.Context, taskIDs []int64) []string {
	if len(taskIDs) == 0 {
		return nil
	}
	rows, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		WhereIn("id", taskIDs).
		WhereNull("deleted_at").
		Fields("id, name").
		All()
	if err != nil {
		return nil
	}
	nameByID := make(map[int64]string, len(rows))
	for _, row := range rows {
		nameByID[row["id"].Int64()] = strings.TrimSpace(row["name"].String())
	}
	names := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		if name := nameByID[taskID]; name != "" {
			names = append(names, fmt.Sprintf("%s(%d)", name, taskID))
		} else {
			names = append(names, fmt.Sprintf("%d", taskID))
		}
	}
	return names
}

func latestVerificationRunForWorkflow(ctx context.Context, workflowRunID int64) (g.Map, error) {
	return repo.NewVerificationRunRepo().GetLatestByWorkflow(ctx, workflowRunID)
}

func countVerificationSeverities(issues []g.Map) (blockers, errorsCount, warns, infos int) {
	for _, issue := range issues {
		switch g.NewVar(issue["severity"]).String() {
		case "blocker":
			blockers++
		case "error":
			errorsCount++
		case "warn":
			warns++
		case "info":
			infos++
		}
	}
	return
}

func pickVerificationRepairTaskID(issues gdb.Result) int64 {
	for _, issue := range issues {
		if taskID := issue["domain_task_id"].Int64(); taskID > 0 {
			return taskID
		}
	}
	return 0
}

func buildVerificationRepairReason(issues gdb.Result, extraReason string) string {
	lines := []string{"基于验证问题触发返工："}
	for i, issue := range issues {
		if i >= 5 {
			lines = append(lines, fmt.Sprintf("其余 %d 条问题请查看验证问题列表。", len(issues)-i))
			break
		}
		line := fmt.Sprintf("%d. [%s] %s", i+1, issue["severity"].String(), issue["title"].String())
		if detail := strings.TrimSpace(issue["detail"].String()); detail != "" {
			line += " - " + detail
		}
		if action := strings.TrimSpace(issue["suggested_action"].String()); action != "" {
			line += "；建议：" + action
		}
		lines = append(lines, line)
	}
	if extraReason = strings.TrimSpace(extraReason); extraReason != "" {
		lines = append(lines, "附加说明："+extraReason)
	}
	return strings.Join(lines, "\n")
}
