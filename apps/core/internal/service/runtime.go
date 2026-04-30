package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type StartBrainRunCommand struct {
	ProjectID      string
	TaskID         string
	BrainKind      string
	Prompt         string
	Workdir        string
	MaxTurns       int
	Timeout        string
	Provider       string
	PermissionMode string // 任务级权限: restricted/default/accept-edits/auto/bypass-permissions
}

type BrainRunStartResult struct {
	BindingID string `json:"binding_id"`
	RunID     string `json:"run_id"`
	Status    string `json:"status"`
}

type BrainRunState struct {
	RunID       string          `json:"run_id"`
	ExecutionID string          `json:"execution_id,omitempty"`
	Status      string          `json:"status"`
	Brain       string          `json:"brain,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	CreatedAt   string          `json:"created_at,omitempty"`
}

// ToRunResult normalizes BrainRunState into the formal RunResult DTO.
// This is the adapter boundary: raw brain-v3 output → structured domain object.
func (s *BrainRunState) ToRunResult(taskID string) braincontracts.RunResult {
	return buildRunResultFromBrainState(s, taskID, s.Brain)
}

// ToDeliveryResult attempts to derive a DeliveryResult from the brain run state.
func (s *BrainRunState) ToDeliveryResult(taskID string) braincontracts.DeliveryResult {
	return buildDeliveryResultFromBrainState(s, taskID)
}

type BrainRunResumeResult struct {
	RunID      string `json:"run_id"`
	StoreRunID string `json:"store_run_id,omitempty"`
	ResumedAs  string `json:"resumed_as,omitempty"`
	Status     string `json:"status"`
	Turns      int    `json:"turns,omitempty"`
	Reply      string `json:"reply,omitempty"`
}

type IRuntime interface {
	Healthz(ctx context.Context) (*runtimev1.HealthzRes, error)
	GetExecutionView(ctx context.Context, projectID string, bindingID string, eventLimit int, replayLimit int, logLimit int) (*runtimev1.ExecutionViewRes, error)
	CheckHealth(ctx context.Context) error
	StartRunCommand(ctx context.Context, req StartBrainRunCommand) (*runtimev1.StartRunRes, error)
	StartBrainRun(ctx context.Context, req StartBrainRunCommand) (*BrainRunStartResult, error)
	GetRunBindingView(ctx context.Context, bindingID string) (*runtimev1.GetRunBindingRes, error)
	GetRunBindingDetail(ctx context.Context, bindingID string, eventLimit int) (*runtimev1.GetRunBindingDetailRes, error)
	ListRunBindingEvents(ctx context.Context, bindingID string, limit int) (*runtimev1.ListRunBindingEventsRes, error)
	GetBrainRun(ctx context.Context, runID string) (*BrainRunState, error)
	SyncRunBindingCommand(ctx context.Context, bindingID string) (*runtimev1.SyncRunBindingRes, error)
	ResumeBrainRun(ctx context.Context, runID string) (*BrainRunResumeResult, error)
	ResumeRunBindingCommand(ctx context.Context, bindingID string) (*runtimev1.ResumeRunBindingRes, error)
	CancelBrainRun(ctx context.Context, runID string) error
	CancelRunBindingCommand(ctx context.Context, bindingID string) (*runtimev1.CancelRunBindingRes, error)
	SyncBrainRunBinding(ctx context.Context, bindingID string) (*BrainRunState, error)
}

var localRuntime IRuntime

type sRuntime struct {
	httpClient *http.Client
	loadGuard  *SystemLoadGuard
}

func Runtime() IRuntime {
	if localRuntime == nil {
		localRuntime = &sRuntime{
			httpClient: &http.Client{Timeout: 15 * time.Second},
			loadGuard:  NewSystemLoadGuard(),
		}
	}
	return localRuntime
}

func (s *sRuntime) CheckHealth(ctx context.Context) error {
	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return wrapRuntimeError(runtimeErrorCodeUnavailable, "resolve brain serve base url failed", err)
	}
	if err = runtimeHealthCheck(ctx, s.httpClient, baseURL); err != nil {
		return wrapRuntimeError(runtimeErrorCodeUnavailable, "brain serve is unavailable", err)
	}
	return nil
}

// validateReworkReadyForExecution checks whether a project in "reworking" status
// is ready to return to "executing". Maps Engineering Cybernetics ch.4 / ch.11:
// reworking → executing requires RepairPlanDraft to be ready.
func validateReworkReadyForExecution(ctx context.Context, db *sql.DB, projectID string) error {
	draft, err := getLatestRepairDraftForProject(ctx, projectID)
	if err != nil {
		return gerror.Wrap(err, "load repair draft for rework validation failed")
	}
	if draft == nil {
		return gerror.New("project is in reworking status but no RepairPlanDraft exists; cannot continue to execution")
	}
	if strings.TrimSpace(draft.RepairPlanJSON) == "" || draft.RepairPlanJSON == "{}" {
		return gerror.New("RepairPlanDraft exists but updated_tasks are not yet defined; cannot continue to execution")
	}
	// P0-03: hard rule — human checkpoint must be cleared before re-execution.
	if draft.HumanCheckpointRequired {
		return gerror.New("RepairPlanDraft requires human_checkpoint; cannot auto-continue to execution")
	}
	return nil
}

func (s *sRuntime) Healthz(ctx context.Context) (*runtimev1.HealthzRes, error) {
	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeUnavailable, "resolve brain serve base url failed", err)
	}
	if err = runtimeHealthCheck(ctx, s.httpClient, baseURL); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeUnavailable, "brain serve is unavailable", err)
	}
	return &runtimev1.HealthzRes{
		Status:  "ok",
		BaseURL: baseURL,
	}, nil
}

func (s *sRuntime) StartBrainRun(ctx context.Context, req StartBrainRunCommand) (*BrainRunStartResult, error) {
	// P1-08: load guard check before starting any new resource-intensive run.
	if s.loadGuard != nil && !s.loadGuard.AllowRun() {
		return nil, wrapRuntimeError(runtimeErrorCodeUnavailable, "system load too high: new run start throttled by load guard", nil)
	}

	normalized, err := normalizeStartBrainRunCommand(req)
	if err != nil {
		return nil, err
	}
	reusable, err := findReusableRunBindingForTask(ctx, normalized.ProjectID, normalized.TaskID, normalized.BrainKind)
	if err != nil {
		return nil, err
	}
	if reusable != nil {
		return &BrainRunStartResult{
			BindingID: reusable.Id,
			RunID:     reusable.BrainRunId,
			Status:    reusable.RunStatus,
		}, nil
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()
	project, err := getProjectByID(ctx, db, normalized.ProjectID)
	if err != nil {
		return nil, err
	}
	if err = validateProjectStatusTransition(project.Status, []string{"compiled", "execution_ready", "executing", "running", "acceptance", "reworking"}); err != nil {
		return nil, err
	}
	// Engineering Cybernetics ch.4: reworking → executing requires RepairPlanDraft to be ready
	if project.Status == "reworking" {
		if err = validateReworkReadyForExecution(ctx, db, project.Id); err != nil {
			return nil, err
		}
	}

	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeUnavailable, "resolve brain serve base url failed", err)
	}
	if err = runtimeHealthCheck(ctx, s.httpClient, baseURL); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeUnavailable, "brain serve is unavailable", err)
	}

	started, err := runtimeCreateRun(ctx, s.httpClient, baseURL, normalized)
	if err != nil {
		recordDiagnostic(ctx, "runtime.create_run", "warning", runtimeErrorCodeCreateRun, "create brain run failed", map[string]any{
			"project_id": normalized.ProjectID,
			"task_id":    normalized.TaskID,
			"brain_kind": normalized.BrainKind,
			"provider":   normalized.Provider,
			"error":      err.Error(),
		})
		return nil, wrapRuntimeError(runtimeErrorCodeCreateRun, "create brain run failed", err)
	}

	bindingID := newResourceID("runbind")
	now := nowText()
	mappedStatus := mapBrainRunStatus(started.Status)
	bindingRow := &do.BrainRunBindings{
		Id:         bindingID,
		ProjectId:  normalized.ProjectID,
		TaskId:     normalized.TaskID,
		BrainKind:  normalized.BrainKind,
		BrainRunId: started.RunID,
		RunStatus:  mappedStatus,
		StartedAt:  now,
		LastSyncAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err = insertBrainRunBinding(ctx, bindingRow); err != nil {
		return nil, err
	}
	if err = appendRunEventIndex(
		ctx,
		normalized.ProjectID,
		bindingID,
		runEventTypeForStatus(mappedStatus),
		eventLevelForRunStatus(mappedStatus),
		runEventSummaryForStatus(mappedStatus),
		map[string]any{
			"run_id":         started.RunID,
			"execution_id":   started.ExecutionID,
			"runtime_status": started.Status,
			"mapped_status":  mappedStatus,
			"brain_kind":     normalized.BrainKind,
		},
	); err != nil {
		return nil, err
	}
	if err = updateProjectTaskRuntimeStatus(ctx, normalized.TaskID, mappedStatus); err != nil {
		return nil, err
	}
	if err = maybeUpdateProjectStatusForExecution(ctx, normalized.ProjectID); err != nil {
		return nil, err
	}

	return &BrainRunStartResult{
		BindingID: bindingID,
		RunID:     started.RunID,
		Status:    mappedStatus,
	}, nil
}

func (s *sRuntime) StartRunCommand(ctx context.Context, req StartBrainRunCommand) (*runtimev1.StartRunRes, error) {
	started, err := s.StartBrainRun(ctx, req)
	if err != nil {
		return nil, err
	}
	if auditErr := insertAuditLog(ctx, req.ProjectID, "runtime.run.started", "user:local_operator", "Brain run started", map[string]any{
		"binding_id": started.BindingID,
		"task_id":    req.TaskID,
		"brain_kind": req.BrainKind,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	recordSecurityAuditNoFail(ctx, SecurityAuditEvent{
		ProjectID:   req.ProjectID,
		EventType:   AuditEventBrainRunStart,
		Severity:    SeverityInfo,
		Operator:    "user:local_operator",
		Resource:    started.BindingID,
		Description: "Brain run started",
		Metadata: map[string]string{
			"task_id":    req.TaskID,
			"brain_kind": req.BrainKind,
			"run_id":     started.RunID,
		},
	})
	bindingView, err := s.GetRunBindingView(ctx, started.BindingID)
	if err != nil {
		return nil, err
	}
	return &runtimev1.StartRunRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: started.BindingID,
		NextAction: "sync_runtime_run_binding",
		RunBinding: bindingView.RunBinding,
	}, nil
}

func (s *sRuntime) GetBrainRun(ctx context.Context, runID string) (*BrainRunState, error) {
	if runID == "" {
		return nil, gerror.New("run id is required")
	}
	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "resolve brain serve base url failed", err)
	}
	runState, err := runtimeGetRun(ctx, s.httpClient, baseURL, runID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "query brain run status failed", err)
	}
	return runState, nil
}

func (s *sRuntime) GetRunBindingView(ctx context.Context, bindingID string) (*runtimev1.GetRunBindingRes, error) {
	binding, err := getBrainRunBindingByID(ctx, bindingID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "load brain run binding failed", err)
	}
	return &runtimev1.GetRunBindingRes{
		RunBinding: mapRunBindingView(binding),
	}, nil
}

func (s *sRuntime) GetRunBindingDetail(ctx context.Context, bindingID string, eventLimit int) (*runtimev1.GetRunBindingDetailRes, error) {
	if bindingID == "" {
		return nil, gerror.New("binding id is required")
	}
	binding, err := getBrainRunBindingByID(ctx, bindingID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "load brain run binding failed", err)
	}

	eventsRes, err := s.ListRunBindingEvents(ctx, bindingID, eventLimit)
	if err != nil {
		return nil, err
	}

	res := &runtimev1.GetRunBindingDetailRes{
		RunBinding:   mapRunBindingView(binding),
		RecentEvents: eventsRes.Events,
		RefreshHint:  "sync_runtime_run_binding",
		RuntimeStale: true,
	}

	runState, err := s.GetBrainRun(ctx, binding.BrainRunId)
	if err != nil {
		res.RuntimeError = err.Error()
		return res, nil
	}
	res.RuntimeState = &runtimev1.BrainRunStateView{
		RunID:       runState.RunID,
		ExecutionID: runState.ExecutionID,
		Status:      runState.Status,
		Brain:       runState.Brain,
		Prompt:      runState.Prompt,
		CreatedAt:   runState.CreatedAt,
	}
	res.RuntimeStale = false
	return res, nil
}

func (s *sRuntime) ListRunBindingEvents(ctx context.Context, bindingID string, limit int) (*runtimev1.ListRunBindingEventsRes, error) {
	if bindingID == "" {
		return nil, gerror.New("binding id is required")
	}
	events, err := listRunBindingEvents(ctx, bindingID, limit)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "list brain run binding events failed", err)
	}
	items := make([]runtimev1.RunEventListItem, 0, len(events))
	for _, item := range events {
		items = append(items, runtimev1.RunEventListItem{
			EventID:    item.Id,
			SequenceNo: item.SequenceNo,
			EventType:  item.EventType,
			EventLevel: item.EventLevel,
			Summary:    item.Summary,
			Payload:    item.PayloadJson,
			CreatedAt:  item.CreatedAt,
		})
	}
	return &runtimev1.ListRunBindingEventsRes{
		RunBindingID: bindingID,
		Events:       items,
	}, nil
}

func (s *sRuntime) CancelBrainRun(ctx context.Context, runID string) error {
	if runID == "" {
		return gerror.New("run id is required")
	}
	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return wrapRuntimeError(runtimeErrorCodeSyncRun, "resolve brain serve base url failed", err)
	}
	if err = runtimeCancelRun(ctx, s.httpClient, baseURL, runID); err != nil {
		return wrapRuntimeError(runtimeErrorCodeSyncRun, "cancel brain run failed", err)
	}
	return nil
}

func (s *sRuntime) CancelRunBindingCommand(ctx context.Context, bindingID string) (*runtimev1.CancelRunBindingRes, error) {
	binding, err := getBrainRunBindingByID(ctx, bindingID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "load brain run binding failed", err)
	}
	if err = s.CancelBrainRun(ctx, binding.BrainRunId); err != nil {
		return nil, err
	}
	if err = updateBrainRunBindingStatus(ctx, bindingID, "run_cancelled"); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "update cancelled binding status failed", err)
	}
	if err = appendRunEventIndex(
		ctx,
		binding.ProjectId,
		bindingID,
		"run.cancel_requested",
		"warning",
		"brain run cancel requested",
		map[string]any{
			"run_id":          binding.BrainRunId,
			"previous_status": binding.RunStatus,
			"mapped_status":   "run_cancelled",
		},
	); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "append cancel event failed", err)
	}
	if err = updateProjectTaskRuntimeStatus(ctx, binding.TaskId, "run_cancelled"); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "update task cancelled status failed", err)
	}
	bindingView, err := s.GetRunBindingView(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	if auditErr := insertAuditLog(ctx, binding.ProjectId, "runtime.run.cancelled", "user:local_operator", "Brain run cancelled", map[string]any{"binding_id": bindingID}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	recordSecurityAuditNoFail(ctx, SecurityAuditEvent{
		ProjectID:   binding.ProjectId,
		EventType:   AuditEventBrainRunStop,
		Severity:    SeverityWarning,
		Operator:    "user:local_operator",
		Resource:    bindingID,
		Description: "Brain run cancelled",
		Metadata:    map[string]string{"reason": "user_cancelled"},
	})
	return &runtimev1.CancelRunBindingRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: bindingID,
		NextAction: "refresh_project_workspace",
		RunBinding: bindingView.RunBinding,
	}, nil
}

func (s *sRuntime) SyncBrainRunBinding(ctx context.Context, bindingID string) (*BrainRunState, error) {
	if bindingID == "" {
		return nil, gerror.New("binding id is required")
	}

	binding, err := getBrainRunBindingByID(ctx, bindingID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "load brain run binding failed", err)
	}
	runState, err := s.GetBrainRun(ctx, binding.BrainRunId)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "sync brain run status failed", err)
	}
	nextStatus := mapBrainRunStatus(runState.Status)
	if err = updateBrainRunBindingStatus(ctx, bindingID, nextStatus); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "update brain run binding status failed", err)
	}
	if nextStatus != binding.RunStatus {
		// Build formal RunResult and DeliveryResult DTOs at the adapter boundary
		runResult := runState.ToRunResult(binding.TaskId)
		deliveryResult := runState.ToDeliveryResult(binding.TaskId)
		if err = appendRunEventIndex(
			ctx,
			binding.ProjectId,
			bindingID,
			runEventTypeForStatus(nextStatus),
			eventLevelForRunStatus(nextStatus),
			runEventSummaryForStatus(nextStatus),
			map[string]any{
				"run_id":          runState.RunID,
				"execution_id":    runState.ExecutionID,
				"runtime_status":  runState.Status,
				"mapped_status":   nextStatus,
				"previous_status": binding.RunStatus,
				"run_result":      runResult,
				"delivery_result": deliveryResult,
			},
		); err != nil {
			return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "append run status event failed", err)
		}
	}
	if err = updateProjectTaskRuntimeStatus(ctx, binding.TaskId, nextStatus); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "update task runtime status failed", err)
	}
	if err = appendRunCheckpointForState(ctx, binding, runState, nextStatus); err != nil {
		recordDiagnostic(ctx, "runtime.sync_checkpoint", "warning", "RUNTIME_CHECKPOINT_WRITE_FAILED", "write runtime checkpoint failed", map[string]any{
			"binding_id":    binding.Id,
			"project_id":    binding.ProjectId,
			"task_id":       binding.TaskId,
			"run_id":        binding.BrainRunId,
			"execution_id":  runState.ExecutionID,
			"mapped_status": nextStatus,
			"error":         err.Error(),
		})
	}
	if err = refreshReplayArtifactsForRun(ctx, binding, runState); err != nil {
		recordDiagnostic(ctx, "runtime.sync_replay_index", "warning", "RUNTIME_REPLAY_INDEX_FAILED", "refresh replay artifacts failed", map[string]any{
			"binding_id":    binding.Id,
			"project_id":    binding.ProjectId,
			"task_id":       binding.TaskId,
			"run_id":        binding.BrainRunId,
			"execution_id":  runState.ExecutionID,
			"mapped_status": nextStatus,
			"error":         err.Error(),
		})
	}
	return runState, nil
}

func (s *sRuntime) SyncRunBindingCommand(ctx context.Context, bindingID string) (*runtimev1.SyncRunBindingRes, error) {
	before, err := getBrainRunBindingByID(ctx, bindingID)
	if err != nil {
		recordDiagnostic(ctx, "runtime.sync_run", "warning", runtimeErrorCodeSyncRun, "load brain run binding before sync failed", map[string]any{
			"binding_id": bindingID,
			"error":      err.Error(),
		})
		return nil, wrapRuntimeError(runtimeErrorCodeSyncRun, "load brain run binding before sync failed", err)
	}
	if _, err = s.SyncBrainRunBinding(ctx, bindingID); err != nil {
		recordDiagnostic(ctx, "runtime.sync_run", "warning", runtimeErrorCodeSyncRun, "sync brain run binding failed", map[string]any{
			"binding_id": bindingID,
			"project_id": before.ProjectId,
			"task_id":    before.TaskId,
			"run_id":     before.BrainRunId,
			"error":      err.Error(),
		})
		return nil, err
	}
	bindingView, err := s.GetRunBindingView(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	after := &bindingView.RunBinding
	if auditErr := insertAuditLog(ctx, before.ProjectId, "runtime.run.synced", "user:local_operator", "Brain run synced", map[string]any{"binding_id": bindingID}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	handleAutomaticRunTerminalActions(ctx, before, after)
	return &runtimev1.SyncRunBindingRes{
		RunBinding: bindingView.RunBinding,
	}, nil
}

func handleAutomaticRunTerminalActions(ctx context.Context, before *entity.BrainRunBindings, after *runtimev1.RunBindingView) {
	if before == nil || after == nil {
		return
	}
	if before.RunStatus == after.RunStatus {
		return
	}
	if isTerminalBrainRunStatus(before.RunStatus) {
		return
	}
	if after.RunStatus != "run_succeeded" && after.RunStatus != "run_failed" {
		return
	}
	if err := maybeAutoAdjudicateAcceptanceRun(ctx, before.ProjectId, before.TaskId, before.Id, before.BrainRunId, after.RunStatus); err != nil {
		handleWorkerFailure(
			ctx,
			"runtime_auto_adjudication",
			before.ProjectId,
			"WORKER_AUTO_ADJUDICATION",
			"automatic acceptance adjudication failed",
			map[string]any{
				"binding_id": before.Id,
				"task_id":    before.TaskId,
				"run_id":     before.BrainRunId,
				"run_status": after.RunStatus,
				"error":      err.Error(),
			},
		)
	}
}

func maybeAutoAdjudicateAcceptanceRun(ctx context.Context, projectID string, taskID string, bindingID string, runID string, nextStatus string) error {
	projectID = strings.TrimSpace(projectID)
	taskID = strings.TrimSpace(taskID)
	if projectID == "" || taskID == "" {
		return nil
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	run, err := getLatestAcceptanceRunForTask(ctx, db, projectID, taskID)
	if err != nil || run == nil {
		return err
	}
	if isTerminalAcceptanceRunStatus(run.Status) {
		return nil
	}

	result, err := Acceptance().AdjudicateAcceptanceRunByID(ctx, projectID, run.Id)
	if err != nil {
		return err
	}
	return appendRunEventIndex(
		ctx,
		projectID,
		bindingID,
		"acceptance.auto_adjudicated",
		"info",
		"acceptance adjudication completed automatically after terminal brain run",
		map[string]any{
			"acceptance_run_id": run.Id,
			"task_id":           taskID,
			"run_id":            runID,
			"run_status":        nextStatus,
			"final_status":      strings.TrimSpace(result.FinalStatus),
		},
	)
}

func (s *sRuntime) ResumeBrainRun(ctx context.Context, runID string) (*BrainRunResumeResult, error) {
	resumed, err := runtimeResumeRun(ctx, runID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeResumeRun, "resume brain run failed", err)
	}
	return &BrainRunResumeResult{
		RunID:      resumed.RunID,
		StoreRunID: resumed.StoreRunID,
		ResumedAs:  resumed.ResumedAs,
		Status:     mapBrainRunStatus(resumed.State),
		Turns:      resumed.Turns,
		Reply:      resumed.Reply,
	}, nil
}

func (s *sRuntime) ResumeRunBindingCommand(ctx context.Context, bindingID string) (*runtimev1.ResumeRunBindingRes, error) {
	if bindingID == "" {
		return nil, gerror.New("binding id is required")
	}

	binding, err := getBrainRunBindingByID(ctx, bindingID)
	if err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeResumeRun, "load brain run binding failed", err)
	}
	resumed, err := s.ResumeBrainRun(ctx, binding.BrainRunId)
	if err != nil {
		return nil, err
	}

	if err = updateBrainRunBindingStatus(ctx, bindingID, resumed.Status); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeResumeRun, "update resumed binding status failed", err)
	}
	if err = appendRunEventIndex(
		ctx,
		binding.ProjectId,
		bindingID,
		"run.resumed",
		"info",
		"brain run resumed via runtime adapter",
		map[string]any{
			"run_id":          resumed.RunID,
			"store_run_id":    resumed.StoreRunID,
			"resumed_as":      resumed.ResumedAs,
			"mapped_status":   resumed.Status,
			"previous_status": binding.RunStatus,
			"turns":           resumed.Turns,
			"reply":           resumed.Reply,
		},
	); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeResumeRun, "append resume event failed", err)
	}
	if resumed.Status != binding.RunStatus {
		if err = appendRunEventIndex(
			ctx,
			binding.ProjectId,
			bindingID,
			runEventTypeForStatus(resumed.Status),
			eventLevelForRunStatus(resumed.Status),
			runEventSummaryForStatus(resumed.Status),
			map[string]any{
				"run_id":          resumed.RunID,
				"store_run_id":    resumed.StoreRunID,
				"resumed_as":      resumed.ResumedAs,
				"mapped_status":   resumed.Status,
				"previous_status": binding.RunStatus,
				"turns":           resumed.Turns,
			},
		); err != nil {
			return nil, wrapRuntimeError(runtimeErrorCodeResumeRun, "append resumed status event failed", err)
		}
	}
	if err = updateProjectTaskRuntimeStatus(ctx, binding.TaskId, resumed.Status); err != nil {
		return nil, wrapRuntimeError(runtimeErrorCodeResumeRun, "update task resumed status failed", err)
	}

	bindingView, err := s.GetRunBindingView(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	if auditErr := insertAuditLog(ctx, binding.ProjectId, "runtime.run.resumed", "user:local_operator", "Brain run resumed", map[string]any{"binding_id": bindingID}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	return &runtimev1.ResumeRunBindingRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: bindingID,
		NextAction: "refresh_project_workspace",
		RunBinding: bindingView.RunBinding,
	}, nil
}
