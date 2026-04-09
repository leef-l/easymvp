package chat

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workspace"
	"easymvp/utility/snowflake"
)

// ProjectTrace 项目级轨迹总览。
func (c *cWorkflow) ProjectTrace(ctx context.Context, req *v1.WorkflowProjectTraceReq) (res *v1.WorkflowProjectTraceRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return &v1.WorkflowProjectTraceRes{
			ProjectID:              req.ProjectID,
			DeliveryModes:          map[string]int{},
			SyncStatuses:           map[string]int{},
			Stages:                 []v1.ProjectTraceStageItem{},
			RecentEvents:           []v1.TimelineEvent{},
			PendingDeliveryReviews: 0,
			HighRiskTasks:          0,
			PRDraftTasks:           0,
			ManualSyncTasks:        0,
		}, nil
	}
	workflowRunID := wfRun["id"].Int64()

	res = &v1.WorkflowProjectTraceRes{
		WorkflowRunID:          snowflake.JsonInt64(workflowRunID),
		ProjectID:              req.ProjectID,
		WorkflowStatus:         wfRun["status"].String(),
		DeliveryModes:          map[string]int{},
		SyncStatuses:           map[string]int{},
		Stages:                 []v1.ProjectTraceStageItem{},
		RecentEvents:           []v1.TimelineEvent{},
		PendingDeliveryReviews: 0,
		HighRiskTasks:          0,
		PRDraftTasks:           0,
		ManualSyncTasks:        0,
	}

	if currentStageRunID := wfRun["current_stage_run_id"].Int64(); currentStageRunID > 0 {
		stageType, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
			Where("id", currentStageRunID).
			WhereNull("deleted_at").
			Value("stage_type")
		if stageErr == nil {
			res.CurrentStage = stageType.String()
		}
	}

	if res.CurrentStage == "" {
		stageRecord, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			WhereNull("deleted_at").
			OrderDesc("stage_no").
			Fields("stage_type").
			One()
		if stageErr == nil && !stageRecord.IsEmpty() {
			res.CurrentStage = stageRecord["stage_type"].String()
		}
	}

	res.TotalEvents, _ = g.DB().Model("mvp_workflow_event").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Count()
	res.TotalTasks, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Count()
	res.TotalStages, _ = g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Count()
	res.ReworkRounds, _ = g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", "rework").
		WhereNull("deleted_at").
		Count()
	res.OpenReviewIssues, _ = g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", "open").
		WhereNull("deleted_at").
		Count()
	res.PendingCheckpoints, _ = g.DB().Model("mvp_human_checkpoint").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "open").
		WhereNull("deleted_at").
		Count()
	res.PendingActions, _ = g.DB().Model("mvp_decision_action").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("action_status", g.Slice{"pending", "waiting_human"}).
		WhereNull("deleted_at").
		Count()

	latestAcceptRun, acceptErr := g.DB().Model("mvp_accept_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("accept_round").
		Fields("id, decision, score").
		One()
	if acceptErr == nil && !latestAcceptRun.IsEmpty() {
		res.AcceptDecision = latestAcceptRun["decision"].String()
		res.AcceptScore = latestAcceptRun["score"].Float64()
		res.OpenAcceptIssues, _ = g.DB().Model("mvp_accept_issue").Ctx(ctx).
			Where("accept_run_id", latestAcceptRun["id"].Int64()).
			Where("status", "open").
			WhereNull("deleted_at").
			Count()
	}

	stageRecords, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("stage_no").
		Fields("id, stage_type, stage_no, status, started_at, finished_at, error_message").
		All()
	if stageErr == nil {
		for _, record := range stageRecords {
			res.Stages = append(res.Stages, v1.ProjectTraceStageItem{
				ID:         snowflake.JsonInt64(record["id"].Int64()),
				StageType:  record["stage_type"].String(),
				StageNo:    record["stage_no"].Int(),
				Status:     record["status"].String(),
				StartedAt:  normalizeDBUTCGTime(record["started_at"].GTime()),
				FinishedAt: normalizeDBUTCGTime(record["finished_at"].GTime()),
				Error:      record["error_message"].String(),
			})
		}
	}

	if wsGroups, wsErr := loadProjectTraceWorkspaceSummary(ctx, workflowRunID); wsErr == nil {
		res.DeliveryModes = wsGroups.deliveryModes
		res.SyncStatuses = wsGroups.syncStatuses
		res.PendingDeliveryReviews = wsGroups.pendingDeliveryReviews
		res.HighRiskTasks = wsGroups.highRiskTasks
		res.PRDraftTasks = wsGroups.prDraftTasks
		res.ManualSyncTasks = wsGroups.manualSyncTasks
	}

	eventRecords, eventErr := g.DB().Model("mvp_workflow_event").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		OrderDesc("created_at").
		Limit(12).
		All()
	if eventErr == nil {
		res.RecentEvents = make([]v1.TimelineEvent, 0, len(eventRecords))
		for _, record := range eventRecords {
			res.RecentEvents = append(res.RecentEvents, buildTimelineEvent(record))
		}
	}

	return res, nil
}

type traceWorkspaceSummary struct {
	deliveryModes          map[string]int
	syncStatuses           map[string]int
	pendingDeliveryReviews int
	highRiskTasks          int
	prDraftTasks           int
	manualSyncTasks        int
}

func loadProjectTraceWorkspaceSummary(ctx context.Context, workflowRunID int64) (traceWorkspaceSummary, error) {
	result := traceWorkspaceSummary{
		deliveryModes: map[string]int{},
		syncStatuses:  map[string]int{},
	}

	records, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("delivery_mode, delivery_status, sync_status, risk_level").
		All()
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "unknown column") {
			return result, err
		}
		return result, nil
	}

	for _, record := range records {
		deliveryStatus := record["delivery_status"].String()
		if deliveryMode := record["delivery_mode"].String(); deliveryMode != "" {
			result.deliveryModes[deliveryMode]++
			if deliveryMode == "pr" && deliveryStatus == "ready" {
				result.prDraftTasks++
			}
		}
		if syncStatus := record["sync_status"].String(); syncStatus != "" {
			result.syncStatuses[syncStatus]++
			if syncStatus == "pending" && deliveryStatus == "ready" {
				result.manualSyncTasks++
			}
		}
		riskLevel := record["risk_level"].String()
		deliveryMode := record["delivery_mode"].String()
		syncStatus := record["sync_status"].String()
		if riskLevel == "high" {
			result.highRiskTasks++
		}
		if deliveryStatus == "ready" && (deliveryMode == "pr" || deliveryMode == "manual" || syncStatus == "pending" || riskLevel == "high") {
			result.pendingDeliveryReviews++
		}
	}
	return result, nil
}

// DeliveryReviews 交付闸门明细。
func (c *cWorkflow) DeliveryReviews(ctx context.Context, req *v1.WorkflowDeliveryReviewsReq) (res *v1.WorkflowDeliveryReviewsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return &v1.WorkflowDeliveryReviewsRes{Items: []v1.DeliveryReviewItem{}}, nil
	}
	workflowRunID := wfRun["id"].Int64()

	taskRecords, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("id, name, status, role_type, execution_mode, batch_no, sort").
		OrderAsc("batch_no").
		OrderAsc("sort").
		All()
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}
	if len(taskRecords) == 0 {
		return &v1.WorkflowDeliveryReviewsRes{Items: []v1.DeliveryReviewItem{}}, nil
	}

	taskIDs := make([]int64, 0, len(taskRecords))
	for _, record := range taskRecords {
		taskIDs = append(taskIDs, record["id"].Int64())
	}
	workspaceMeta, wsErr := loadTaskWorkspaceMeta(ctx, taskIDs)
	if wsErr != nil {
		if strings.Contains(strings.ToLower(wsErr.Error()), "unknown column") {
			return &v1.WorkflowDeliveryReviewsRes{Items: []v1.DeliveryReviewItem{}}, nil
		}
		return nil, fmt.Errorf("查询任务工作空间失败: %w", wsErr)
	}

	items := make([]v1.DeliveryReviewItem, 0)
	for _, record := range taskRecords {
		taskID := record["id"].Int64()
		ws := workspaceMeta[taskID]
		reasons := buildDeliveryReviewReasons(ws)
		if len(reasons) == 0 {
			continue
		}
		items = append(items, v1.DeliveryReviewItem{
			WorkspaceID:    snowflake.JsonInt64(ws.WorkspaceID),
			TaskID:         snowflake.JsonInt64(taskID),
			WorkflowRunID:  snowflake.JsonInt64(workflowRunID),
			TaskName:       record["name"].String(),
			TaskStatus:     record["status"].String(),
			RoleType:       record["role_type"].String(),
			ExecutionMode:  record["execution_mode"].String(),
			BatchNo:        record["batch_no"].Int(),
			DeliveryMode:   ws.DeliveryMode,
			DeliveryStatus: ws.DeliveryStatus,
			SyncStrategy:   ws.SyncStrategy,
			SyncStatus:     ws.SyncStatus,
			RiskLevel:      ws.RiskLevel,
			PatchRef:       ws.PatchRef,
			DeliveryRef:    ws.DeliveryRef,
			DeliveryTitle:  ws.DeliveryTitle,
			DiffSummary:    ws.DiffSummary,
			Reasons:        reasons,
			UpdatedAt:      ws.UpdatedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].BatchNo != items[j].BatchNo {
			return items[i].BatchNo < items[j].BatchNo
		}
		if items[i].RiskLevel != items[j].RiskLevel {
			return deliveryReviewRiskRank(items[i].RiskLevel) > deliveryReviewRiskRank(items[j].RiskLevel)
		}
		return int64(items[i].TaskID) < int64(items[j].TaskID)
	})

	return &v1.WorkflowDeliveryReviewsRes{Items: items}, nil
}

// DeliveryApply 人工确认并回写交付结果。
func (c *cWorkflow) DeliveryApply(ctx context.Context, req *v1.WorkflowDeliveryApplyReq) (res *v1.WorkflowDeliveryApplyRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("项目无工作流运行")
	}
	workflowRunID := wfRun["id"].Int64()

	taskIDs := normalizeDeliveryApplyTaskIDs(req.TaskIDs)

	query := g.DB().Model("mvp_task_workspace w").Ctx(ctx).
		InnerJoin("mvp_domain_task t", "t.id = w.task_id").
		Where("w.workflow_run_id", workflowRunID).
		WhereNull("w.deleted_at").
		WhereNull("t.deleted_at").
		Where("w.delivery_status", workspace.DeliveryStatusReady).
		WhereIn("w.sync_status", g.Slice{workspace.SyncStatusPending, workspace.SyncStatusFailed}).
		Fields("w.id, w.task_id, w.sync_status, w.delivery_status, t.name, t.batch_no, t.sort")
	if len(taskIDs) > 0 {
		query = query.WhereIn("w.task_id", taskIDs)
	}

	records, err := query.OrderAsc("t.batch_no").OrderAsc("t.sort").All()
	if err != nil {
		return nil, fmt.Errorf("查询待回写交付失败: %w", err)
	}
	if len(records) == 0 {
		return &v1.WorkflowDeliveryApplyRes{Items: []v1.DeliveryApplyItem{}}, nil
	}

	wsMgr := workspace.NewGitWorktreeManager()
	items := make([]v1.DeliveryApplyItem, 0, len(records))
	var (
		appliedCount int
		failedCount  int
	)
	for _, record := range records {
		taskID := record["task_id"].Int64()
		item := v1.DeliveryApplyItem{
			WorkspaceID:    snowflake.JsonInt64(record["id"].Int64()),
			TaskID:         snowflake.JsonInt64(taskID),
			TaskName:       record["name"].String(),
			DeliveryStatus: record["delivery_status"].String(),
			SyncStatus:     record["sync_status"].String(),
			Status:         "applied",
		}

		if applyErr := wsMgr.ApplyDelivery(ctx, taskID); applyErr != nil {
			failedCount++
			item.Status = "failed"
			item.Message = applyErr.Error()
			item.SyncStatus = workspace.SyncStatusFailed
		} else {
			appliedCount++
			item.Message = strings.TrimSpace(req.Reason)
			item.SyncStatus = workspace.SyncStatusApplied
		}
		items = append(items, item)
	}

	return &v1.WorkflowDeliveryApplyRes{
		AppliedCount: appliedCount,
		FailedCount:  failedCount,
		Items:        items,
	}, nil
}

func buildDeliveryReviewReasons(ws taskWorkspaceMeta) []string {
	if ws.DeliveryStatus != "ready" {
		return nil
	}

	reasons := make([]string, 0, 4)
	switch ws.DeliveryMode {
	case "pr":
		reasons = append(reasons, "PR 草稿待人工确认")
	case "manual":
		reasons = append(reasons, "人工交付待处理")
	}
	if ws.SyncStatus == "pending" {
		reasons = append(reasons, "变更待人工回写")
	}
	if ws.RiskLevel == "high" {
		reasons = append(reasons, "高风险任务需人工复核")
	}
	return reasons
}

func normalizeDeliveryApplyTaskIDs(taskIDs []snowflake.JsonInt64) []int64 {
	result := make([]int64, 0, len(taskIDs))
	seen := make(map[int64]struct{}, len(taskIDs))
	for _, raw := range taskIDs {
		taskID := int64(raw)
		if taskID <= 0 {
			continue
		}
		if _, exists := seen[taskID]; exists {
			continue
		}
		seen[taskID] = struct{}{}
		result = append(result, taskID)
	}
	return result
}

func deliveryReviewRiskRank(riskLevel string) int {
	switch riskLevel {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// TaskReplay 任务执行回放。
func (c *cWorkflow) TaskReplay(ctx context.Context, req *v1.WorkflowTaskReplayReq) (res *v1.WorkflowTaskReplayRes, err error) {
	projectID := int64(req.ProjectID)
	taskID := int64(req.TaskID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	taskRecord, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Fields("id, workflow_run_id, stage_run_id, name, description, status, role_type, role_level, batch_no, sort, execution_mode, affected_resources, started_at, completed_at, result, retry_count, error_message").
		One()
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	if taskRecord.IsEmpty() {
		return nil, fmt.Errorf("任务不存在: %d", taskID)
	}

	workflowRunID := taskRecord["workflow_run_id"].Int64()
	owned, ownErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Count()
	if ownErr != nil {
		return nil, fmt.Errorf("校验任务归属失败: %w", ownErr)
	}
	if owned == 0 {
		return nil, fmt.Errorf("任务 %d 不属于项目 %d", taskID, projectID)
	}

	workspaceMeta, wsErr := loadTaskWorkspaceMeta(ctx, []int64{taskID})
	if wsErr != nil {
		g.Log().Warningf(ctx, "[TaskReplay] 查询工作空间信息失败: task=%d err=%v", taskID, wsErr)
	}

	res = &v1.WorkflowTaskReplayRes{
		WorkflowRunID: snowflake.JsonInt64(workflowRunID),
		Task:          buildDomainTaskItem(taskRecord, workspaceMeta[taskID]),
		Logs:          []v1.TaskReplayLogItem{},
		Events:        []v1.TimelineEvent{},
		Issues:        []v1.AcceptIssueItem{},
		Evidence:      []v1.AcceptEvidenceItem{},
		Handoffs:      []v1.TaskReplayHandoffItem{},
		Actions:       []v1.DecisionActionDTO{},
	}

	if stageRunID := taskRecord["stage_run_id"].Int64(); stageRunID > 0 {
		res.StageRunID = snowflake.JsonInt64(stageRunID)
		stageType, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
			Where("id", stageRunID).
			WhereNull("deleted_at").
			Value("stage_type")
		if stageErr == nil {
			res.StageType = stageType.String()
		}
	}

	logRecords, logErr := g.DB().Model("mvp_task_log").Ctx(ctx).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		OrderAsc("created_at").
		All()
	if logErr == nil {
		for _, record := range logRecords {
			res.Logs = append(res.Logs, v1.TaskReplayLogItem{
				ID:         snowflake.JsonInt64(record["id"].Int64()),
				Action:     record["action"].String(),
				FromStatus: record["from_status"].String(),
				ToStatus:   record["to_status"].String(),
				Message:    record["message"].String(),
				Operator:   record["operator"].String(),
				CreatedAt:  normalizeDBUTCGTime(record["created_at"].GTime()),
			})
		}
	}

	eventRecords, eventErr := g.DB().Model("mvp_workflow_event").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("entity_type", "domain_task").
		Where("entity_id", taskID).
		OrderAsc("created_at").
		All()
	if eventErr == nil {
		for _, record := range eventRecords {
			res.Events = append(res.Events, buildTimelineEvent(record))
		}
	}

	issueRecords, issueErr := g.DB().Model("mvp_accept_issue").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("domain_task_id", taskID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if issueErr == nil {
		for _, issue := range issueRecords {
			res.Issues = append(res.Issues, v1.AcceptIssueItem{
				ID:              snowflake.JsonInt64(issue["id"].Int64()),
				IssueType:       issue["issue_type"].String(),
				RuleCode:        issue["rule_code"].String(),
				Severity:        issue["severity"].String(),
				Title:           issue["title"].String(),
				Detail:          issue["detail"].String(),
				ExpectedValue:   issue["expected_value"].String(),
				ActualValue:     issue["actual_value"].String(),
				SuggestedAction: issue["suggested_action"].String(),
				DomainTaskID:    snowflake.JsonInt64(issue["domain_task_id"].Int64()),
				ResourceRef:     issue["resource_ref"].String(),
				Status:          issue["status"].String(),
				CreatedAt:       normalizeDBUTCGTime(issue["created_at"].GTime()),
			})
		}
	}

	workspaceID := int64(0)
	workspaceRecord, workspaceErr := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		Fields("id").
		OrderDesc("created_at").
		One()
	if workspaceErr == nil && !workspaceRecord.IsEmpty() {
		workspaceID = workspaceRecord["id"].Int64()
	}

	evidenceQuery := g.DB().Model("mvp_accept_evidence").Ctx(ctx).
		WhereNull("deleted_at")
	if workspaceID > 0 {
		evidenceQuery = evidenceQuery.Where(
			"(source_type = ? AND source_id = ?) OR (source_type = ? AND source_id = ?)",
			"domain_task", taskID, "workspace", workspaceID,
		)
	} else {
		evidenceQuery = evidenceQuery.Where("source_type", "domain_task").Where("source_id", taskID)
	}
	evidenceRecords, evidenceErr := evidenceQuery.OrderDesc("created_at").All()
	if evidenceErr == nil {
		for _, record := range evidenceRecords {
			res.Evidence = append(res.Evidence, v1.AcceptEvidenceItem{
				ID:           snowflake.JsonInt64(record["id"].Int64()),
				EvidenceType: record["evidence_type"].String(),
				SourceType:   record["source_type"].String(),
				SourceID:     snowflake.JsonInt64(record["source_id"].Int64()),
				ContentRef:   record["content_ref"].String(),
				Summary:      record["summary"].String(),
				CreatedAt:    normalizeDBUTCGTime(record["created_at"].GTime()),
			})
		}
	}

	handoffRecords, handoffErr := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("(from_task_id = ? OR to_task_id = ?)", taskID, taskID).
		OrderAsc("created_at").
		All()
	if handoffErr == nil {
		for _, record := range handoffRecords {
			res.Handoffs = append(res.Handoffs, v1.TaskReplayHandoffItem{
				ID:          snowflake.JsonInt64(record["id"].Int64()),
				HandoffType: record["handoff_type"].String(),
				FromTaskID:  snowflake.JsonInt64(record["from_task_id"].Int64()),
				ToTaskID:    snowflake.JsonInt64(record["to_task_id"].Int64()),
				Reason:      record["reason"].String(),
				Payload:     record["payload"].String(),
				CreatedAt:   normalizeDBUTCGTime(record["created_at"].GTime()),
			})
		}
	}

	actionRecords, actionErr := g.DB().Model("mvp_decision_action").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("domain_task_id", taskID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if actionErr == nil {
		for _, record := range actionRecords {
			res.Actions = append(res.Actions, mapToDecisionActionDTO(record.Map()))
		}
	}

	return res, nil
}

func buildTimelineEvent(record gdb.Record) v1.TimelineEvent {
	eventType := record["event_type"].String()
	payload := record["payload"].String()
	item := v1.TimelineEvent{
		ID:            snowflake.JsonInt64(record["id"].Int64()),
		WorkflowRunID: snowflake.JsonInt64(record["workflow_run_id"].Int64()),
		EntityType:    record["entity_type"].String(),
		EventType:     eventType,
		Label:         formatTimelineLabel(eventType, payload),
		Payload:       payload,
		CreatedAt:     normalizeDBUTCGTime(record["created_at"].GTime()),
	}
	if sid := record["stage_run_id"].Int64(); sid > 0 {
		v := snowflake.JsonInt64(sid)
		item.StageRunID = &v
	}
	if eid := record["entity_id"].Int64(); eid > 0 {
		v := snowflake.JsonInt64(eid)
		item.EntityID = &v
	}
	return item
}
