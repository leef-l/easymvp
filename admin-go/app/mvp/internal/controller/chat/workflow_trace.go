package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workspace"
	"easymvp/utility/snowflake"
)

// ProjectTrace 项目级轨迹总览。
func (c *cWorkflow) ProjectTrace(ctx context.Context, req *v1.WorkflowProjectTraceReq) (res *v1.WorkflowProjectTraceRes, err error) {
	var (
		acceptIssueRepo     = repo.NewAcceptIssueRepo()
		acceptRunRepo       = repo.NewAcceptRunRepo()
		decisionActionRepo  = repo.NewDecisionActionRepo()
		domainTaskRepo      = repo.NewDomainTaskRepo()
		humanCheckpointRepo = repo.NewHumanCheckpointRepo()
		reviewIssueRepo     = repo.NewReviewIssueRepo()
		stageRunRepo        = repo.NewStageRunRepo()
		workflowEventRepo   = repo.NewWorkflowEventRepo()
	)

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
		stageRecord, stageErr := stageRunRepo.GetByIDMap(ctx, currentStageRunID, "stage_type")
		if stageErr == nil && len(stageRecord) > 0 {
			res.CurrentStage = g.NewVar(stageRecord["stage_type"]).String()
		}
	}

	if res.CurrentStage == "" {
		stageRecord, stageErr := stageRunRepo.GetLatestByWorkflow(ctx, workflowRunID, "stage_type")
		if stageErr == nil && !stageRecord.IsEmpty() {
			res.CurrentStage = stageRecord["stage_type"].String()
		}
	}

	res.TotalEvents, _ = workflowEventRepo.CountByWorkflow(ctx, workflowRunID)
	res.TotalTasks, _ = domainTaskRepo.CountByWorkflow(ctx, workflowRunID)
	res.TotalStages, _ = stageRunRepo.CountByWorkflow(ctx, workflowRunID)
	res.ReworkRounds, _ = stageRunRepo.CountByWorkflowAndType(ctx, workflowRunID, "rework")
	res.OpenReviewIssues, _ = reviewIssueRepo.CountOpenByWorkflow(ctx, workflowRunID)
	res.PendingCheckpoints, _ = humanCheckpointRepo.CountOpenByProject(ctx, projectID)
	res.PendingActions, _ = decisionActionRepo.CountPendingByProject(ctx, projectID)

	latestAcceptRun, acceptErr := acceptRunRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if acceptErr == nil && len(latestAcceptRun) > 0 {
		res.AcceptDecision = g.NewVar(latestAcceptRun["decision"]).String()
		res.AcceptScore = g.NewVar(latestAcceptRun["score"]).Float64()
		res.OpenAcceptIssues, _ = acceptIssueRepo.CountOpenByAcceptRun(ctx, g.NewVar(latestAcceptRun["id"]).Int64())
	}

	stageRecords, stageErr := stageRunRepo.ListByWorkflowMaps(ctx, workflowRunID, "id", "stage_type", "stage_no", "status", "started_at", "finished_at", "error_message")
	if stageErr == nil {
		for _, record := range stageRecords {
			res.Stages = append(res.Stages, v1.ProjectTraceStageItem{
				ID:         snowflake.JsonInt64(g.NewVar(record["id"]).Int64()),
				StageType:  g.NewVar(record["stage_type"]).String(),
				StageNo:    g.NewVar(record["stage_no"]).Int(),
				Status:     g.NewVar(record["status"]).String(),
				StartedAt:  normalizeDBUTCGTime(g.NewVar(record["started_at"]).GTime()),
				FinishedAt: normalizeDBUTCGTime(g.NewVar(record["finished_at"]).GTime()),
				Error:      g.NewVar(record["error_message"]).String(),
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

	eventRecords, eventErr := workflowEventRepo.ListRecentByWorkflow(ctx, workflowRunID, 12)
	if eventErr == nil {
		res.RecentEvents = make([]v1.TimelineEvent, 0, len(eventRecords))
		for _, record := range eventRecords {
			res.RecentEvents = append(res.RecentEvents, buildTimelineEvent(mapToDBRecord(record)))
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

	records, err := repo.NewTaskWorkspaceRepo().ListWorkspaceSummaryByWorkflow(ctx, workflowRunID)
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "unknown column") {
			return result, err
		}
		return result, nil
	}

	for _, record := range records {
		deliveryStatus := g.NewVar(record["delivery_status"]).String()
		if deliveryMode := g.NewVar(record["delivery_mode"]).String(); deliveryMode != "" {
			result.deliveryModes[deliveryMode]++
			if deliveryMode == "pr" && deliveryStatus == "ready" {
				result.prDraftTasks++
			}
		}
		if syncStatus := g.NewVar(record["sync_status"]).String(); syncStatus != "" {
			result.syncStatuses[syncStatus]++
			if syncStatus == "pending" && deliveryStatus == "ready" {
				result.manualSyncTasks++
			}
		}
		riskLevel := g.NewVar(record["risk_level"]).String()
		deliveryMode := g.NewVar(record["delivery_mode"]).String()
		syncStatus := g.NewVar(record["sync_status"]).String()
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
	domainTaskRepo := repo.NewDomainTaskRepo()

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return &v1.WorkflowDeliveryReviewsRes{Items: []v1.DeliveryReviewItem{}}, nil
	}
	workflowRunID := wfRun["id"].Int64()

	taskRecords, err := domainTaskRepo.ListByWorkflowOrdered(ctx, workflowRunID, "id", "name", "status", "role_type", "execution_mode", "batch_no", "sort")
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}
	if len(taskRecords) == 0 {
		return &v1.WorkflowDeliveryReviewsRes{Items: []v1.DeliveryReviewItem{}}, nil
	}

	taskIDs := make([]int64, 0, len(taskRecords))
	for _, record := range taskRecords {
		taskIDs = append(taskIDs, g.NewVar(record["id"]).Int64())
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
		taskID := g.NewVar(record["id"]).Int64()
		ws := workspaceMeta[taskID]
		reasons := buildDeliveryReviewReasons(ws)
		if len(reasons) == 0 {
			continue
		}
		items = append(items, v1.DeliveryReviewItem{
			WorkspaceID:    snowflake.JsonInt64(ws.WorkspaceID),
			TaskID:         snowflake.JsonInt64(taskID),
			WorkflowRunID:  snowflake.JsonInt64(workflowRunID),
			TaskName:       g.NewVar(record["name"]).String(),
			TaskStatus:     g.NewVar(record["status"]).String(),
			RoleType:       g.NewVar(record["role_type"]).String(),
			ExecutionMode:  g.NewVar(record["execution_mode"]).String(),
			BatchNo:        g.NewVar(record["batch_no"]).Int(),
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
	taskWorkspaceRepo := repo.NewTaskWorkspaceRepo()

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

	records, err := taskWorkspaceRepo.ListReadyPendingSyncByWorkflow(ctx, workflowRunID, taskIDs)
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
		taskID := g.NewVar(record["task_id"]).Int64()
		item := v1.DeliveryApplyItem{
			WorkspaceID:    snowflake.JsonInt64(g.NewVar(record["id"]).Int64()),
			TaskID:         snowflake.JsonInt64(taskID),
			TaskName:       g.NewVar(record["name"]).String(),
			DeliveryStatus: g.NewVar(record["delivery_status"]).String(),
			SyncStatus:     g.NewVar(record["sync_status"]).String(),
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
	var (
		acceptEvidenceRepo = repo.NewAcceptEvidenceRepo()
		acceptIssueRepo    = repo.NewAcceptIssueRepo()
		decisionActionRepo = repo.NewDecisionActionRepo()
		domainTaskRepo     = repo.NewDomainTaskRepo()
		handoffRecordRepo  = repo.NewHandoffRecordRepo()
		stageRunRepo       = repo.NewStageRunRepo()
		taskLogRepo        = repo.NewTaskLogRepo()
		taskWorkspaceRepo  = repo.NewTaskWorkspaceRepo()
		workflowEventRepo  = repo.NewWorkflowEventRepo()
		workflowRunRepo    = repo.NewWorkflowRunRepo()
	)

	projectID := int64(req.ProjectID)
	taskID := int64(req.TaskID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	taskRecordMap, err := domainTaskRepo.GetByIDMap(ctx, taskID, "id", "workflow_run_id", "stage_run_id", "name", "description", "status", "role_type", "role_level", "batch_no", "sort", "execution_mode", "affected_resources", "started_at", "completed_at", "result", "retry_count", "error_message")
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	if len(taskRecordMap) == 0 {
		return nil, fmt.Errorf("任务不存在: %d", taskID)
	}
	taskRecord := mapToDBRecord(taskRecordMap)

	workflowRunID := taskRecord["workflow_run_id"].Int64()
	runRecord, ownErr := workflowRunRepo.GetByIDMap(ctx, workflowRunID, "project_id")
	if ownErr != nil {
		return nil, fmt.Errorf("校验任务归属失败: %w", ownErr)
	}
	if len(runRecord) == 0 || g.NewVar(runRecord["project_id"]).Int64() != projectID {
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
		stageRecord, stageErr := stageRunRepo.GetByIDMap(ctx, stageRunID, "stage_type")
		if stageErr == nil && len(stageRecord) > 0 {
			res.StageType = g.NewVar(stageRecord["stage_type"]).String()
		}
	}

	logRecords, logErr := taskLogRepo.ListByTask(ctx, taskID)
	if logErr == nil {
		for _, record := range logRecords {
			res.Logs = append(res.Logs, v1.TaskReplayLogItem{
				ID:         snowflake.JsonInt64(g.NewVar(record["id"]).Int64()),
				Action:     g.NewVar(record["action"]).String(),
				FromStatus: g.NewVar(record["from_status"]).String(),
				ToStatus:   g.NewVar(record["to_status"]).String(),
				Message:    g.NewVar(record["message"]).String(),
				Operator:   g.NewVar(record["operator"]).String(),
				CreatedAt:  normalizeDBUTCGTime(g.NewVar(record["created_at"]).GTime()),
			})
		}
	}

	eventRecords, eventErr := workflowEventRepo.ListByWorkflowAndEntity(ctx, workflowRunID, "domain_task", taskID)
	if eventErr == nil {
		for _, record := range eventRecords {
			res.Events = append(res.Events, buildTimelineEvent(mapToDBRecord(record)))
		}
	}

	issueRecords, issueErr := acceptIssueRepo.ListByWorkflowAndTask(ctx, workflowRunID, taskID)
	if issueErr == nil {
		for _, issue := range issueRecords {
			res.Issues = append(res.Issues, v1.AcceptIssueItem{
				ID:              snowflake.JsonInt64(g.NewVar(issue["id"]).Int64()),
				IssueType:       g.NewVar(issue["issue_type"]).String(),
				RuleCode:        g.NewVar(issue["rule_code"]).String(),
				Severity:        g.NewVar(issue["severity"]).String(),
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
	}

	workspaceID, workspaceErr := taskWorkspaceRepo.GetLatestIDByTask(ctx, taskID)
	if workspaceErr != nil {
		workspaceID = 0
	}

	evidenceRecords, evidenceErr := acceptEvidenceRepo.ListByTaskSources(ctx, taskID, workspaceID)
	if evidenceErr == nil {
		for _, record := range evidenceRecords {
			res.Evidence = append(res.Evidence, v1.AcceptEvidenceItem{
				ID:           snowflake.JsonInt64(g.NewVar(record["id"]).Int64()),
				EvidenceType: g.NewVar(record["evidence_type"]).String(),
				SourceType:   g.NewVar(record["source_type"]).String(),
				SourceID:     snowflake.JsonInt64(g.NewVar(record["source_id"]).Int64()),
				ContentRef:   g.NewVar(record["content_ref"]).String(),
				Summary:      g.NewVar(record["summary"]).String(),
				CreatedAt:    normalizeDBUTCGTime(g.NewVar(record["created_at"]).GTime()),
			})
		}
	}

	handoffRecords, handoffErr := handoffRecordRepo.ListByWorkflowAndTask(ctx, workflowRunID, taskID)
	if handoffErr == nil {
		for _, record := range handoffRecords {
			mode, createdTaskIDs := decodeTaskReplayHandoffPayload(g.NewVar(record["payload"]).String())
			res.Handoffs = append(res.Handoffs, v1.TaskReplayHandoffItem{
				ID:             snowflake.JsonInt64(g.NewVar(record["id"]).Int64()),
				HandoffType:    g.NewVar(record["handoff_type"]).String(),
				FromTaskID:     snowflake.JsonInt64(g.NewVar(record["from_task_id"]).Int64()),
				ToTaskID:       snowflake.JsonInt64(g.NewVar(record["to_task_id"]).Int64()),
				Mode:           mode,
				CreatedTaskIDs: createdTaskIDs,
				Reason:         g.NewVar(record["reason"]).String(),
				Payload:        g.NewVar(record["payload"]).String(),
				CreatedAt:      normalizeDBUTCGTime(g.NewVar(record["created_at"]).GTime()),
			})
		}
	}

	actionRecords, actionErr := decisionActionRepo.ListByWorkflowAndTask(ctx, workflowRunID, taskID)
	if actionErr == nil {
		for _, record := range actionRecords {
			res.Actions = append(res.Actions, mapToDecisionActionDTO(record))
		}
	}

	return res, nil
}

func mapToDBRecord(data g.Map) gdb.Record {
	record := make(gdb.Record, len(data))
	for key, value := range data {
		record[key] = g.NewVar(value)
	}
	return record
}

func decodeTaskReplayHandoffPayload(payload string) (string, []snowflake.JsonInt64) {
	payload = strings.TrimSpace(payload)
	if payload == "" || !json.Valid([]byte(payload)) {
		return "", nil
	}

	var decoded struct {
		Mode           string  `json:"mode"`
		CreatedTaskIDs []int64 `json:"created_task_ids"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return "", nil
	}
	taskIDs := make([]snowflake.JsonInt64, 0, len(decoded.CreatedTaskIDs))
	for _, taskID := range decoded.CreatedTaskIDs {
		if taskID <= 0 {
			continue
		}
		taskIDs = append(taskIDs, snowflake.JsonInt64(taskID))
	}
	return strings.TrimSpace(decoded.Mode), taskIDs
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
