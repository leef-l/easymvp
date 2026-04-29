package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/leef-l/easymvp/apps/core/internal/events"
)

type workflowEventWorker struct {
	interval  time.Duration
	batchSize int
}

func newWorkflowEventWorker() backgroundWorker {
	cfgInterval := g.Cfg().MustGet(context.Background(), "easymvp.workers.workflowEventInterval", "3s").Duration()
	if cfgInterval <= 0 {
		cfgInterval = 3 * time.Second
	}
	batchSize := g.Cfg().MustGet(context.Background(), "easymvp.workers.workflowEventBatchSize", 20).Int()
	if batchSize <= 0 {
		batchSize = 20
	}
	return &workflowEventWorker{
		interval:  cfgInterval,
		batchSize: batchSize,
	}
}

func (w *workflowEventWorker) Name() string {
	return "workflow_event_worker"
}

func (w *workflowEventWorker) Interval() time.Duration {
	return w.interval
}

// RunOnce polls pending workflow_events and processes them.
func (w *workflowEventWorker) RunOnce(ctx context.Context) error {
	evts, err := listPendingWorkflowEvents(ctx, w.batchSize)
	if err != nil {
		return err
	}

	for _, evt := range evts {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := w.processEvent(ctx, evt); err != nil {
			markWorkflowEventFailed(ctx, evt.ID, err.Error())
			g.Log().Warningf(ctx, "[worker:%s] event %s failed: %v", w.Name(), evt.ID, err)
			continue
		}
		markWorkflowEventCompleted(ctx, evt.ID)
	}

	return nil
}

func (w *workflowEventWorker) processEvent(ctx context.Context, evt *events.WorkflowEvent) error {
	switch evt.EventType {
	case events.RepairDraftReady:
		return w.handleRepairDraftReady(ctx, evt)
	case events.PlanReviewRejected:
		// TODO: trigger auto-redesign.
		g.Log().Infof(ctx, "[worker:%s] PlanReviewRejected event received (project=%s)", w.Name(), evt.ProjectID)
		return nil
	case events.PlanReviewApproved:
		// Already handled by taskSchedulerWorker (subscribes via event bus directly).
		return nil
	case events.RunTerminal:
		// Already handled by runSyncWorker + maybeAutoAdjudicateAcceptanceRun.
		return nil
	case events.AcceptanceFailed:
		return w.handleAcceptanceFailed(ctx, evt)
	case events.AcceptancePassed:
		return w.handleAcceptancePassed(ctx, evt)
	case events.BrowserCheckCompleted:
		return w.handleBrowserCheckCompleted(ctx, evt)
	case events.VerifierCheckCompleted:
		return w.handleVerifierCheckCompleted(ctx, evt)
	default:
		g.Log().Debugf(ctx, "[worker:%s] unhandled event type %s", w.Name(), evt.EventType)
		return nil
	}
}

// handleRepairDraftReady extracts updated tasks from the repair plan and
// triggers StartBrainRun for each task.
func (w *workflowEventWorker) handleAcceptanceFailed(ctx context.Context, evt *events.WorkflowEvent) error {
	acceptanceRunID, _ := evt.Payload["acceptance_run_id"].(string)
	g.Log().Infof(ctx, "[worker:%s] AcceptanceFailed project=%s run=%s — repair workflow already triggered by adjudication",
		w.Name(), evt.ProjectID, acceptanceRunID)
	// Repair draft creation is already handled synchronously in adjudicateAcceptanceAggregate.
	// This event handler is available for downstream extensions (notifications, metrics, etc.).
	return nil
}

func (w *workflowEventWorker) handleAcceptancePassed(ctx context.Context, evt *events.WorkflowEvent) error {
	acceptanceRunID, _ := evt.Payload["acceptance_run_id"].(string)
	manualRelease, _ := evt.Payload["manual_release"].(bool)
	if manualRelease {
		g.Log().Infof(ctx, "[worker:%s] AcceptancePassed project=%s run=%s — awaiting manual release",
			w.Name(), evt.ProjectID, acceptanceRunID)
	} else {
		g.Log().Infof(ctx, "[worker:%s] AcceptancePassed project=%s run=%s — auto-completing project",
			w.Name(), evt.ProjectID, acceptanceRunID)
		// Auto-complete project if no manual release required.
		if err := updateProjectStatus(ctx, evt.ProjectID, "completed"); err != nil {
			g.Log().Warningf(ctx, "[worker:%s] auto-complete project %s failed: %v", w.Name(), evt.ProjectID, err)
		}
	}
	return nil
}

func (w *workflowEventWorker) handleBrowserCheckCompleted(ctx context.Context, evt *events.WorkflowEvent) error {
	acceptanceRunID, _ := evt.Payload["acceptance_run_id"].(string)
	browserRunID, _ := evt.Payload["browser_run_id"].(string)
	g.Log().Infof(ctx, "[worker:%s] BrowserCheckCompleted project=%s acceptance=%s browser=%s",
		w.Name(), evt.ProjectID, acceptanceRunID, browserRunID)
	// Trigger async adjudication if both browser and verifier are done.
	return w.maybeTriggerAdjudication(ctx, evt.ProjectID, acceptanceRunID)
}

func (w *workflowEventWorker) handleVerifierCheckCompleted(ctx context.Context, evt *events.WorkflowEvent) error {
	acceptanceRunID, _ := evt.Payload["acceptance_run_id"].(string)
	verifierRunID, _ := evt.Payload["verifier_run_id"].(string)
	g.Log().Infof(ctx, "[worker:%s] VerifierCheckCompleted project=%s acceptance=%s verifier=%s",
		w.Name(), evt.ProjectID, acceptanceRunID, verifierRunID)
	// Trigger async adjudication if both browser and verifier are done.
	return w.maybeTriggerAdjudication(ctx, evt.ProjectID, acceptanceRunID)
}

// maybeTriggerAdjudication checks if both validation runs are complete and
// triggers acceptance adjudication. It uses a simple best-effort check.
func (w *workflowEventWorker) maybeTriggerAdjudication(ctx context.Context, projectID, acceptanceRunID string) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	var browserRunID, verifierRunID string
	row := db.QueryRowContext(ctx,
		`SELECT COALESCE(browser_run_id,''), COALESCE(verifier_run_id,'') FROM acceptance_runs WHERE id = ?`,
		acceptanceRunID)
	if err := row.Scan(&browserRunID, &verifierRunID); err != nil {
		return err
	}

	// If both are present, trigger adjudication (best-effort).
	if browserRunID != "" && verifierRunID != "" {
		g.Log().Infof(ctx, "[worker:%s] both validations complete for acceptance=%s, triggering adjudication", w.Name(), acceptanceRunID)
		if _, adjErr := adjudicateLatestAcceptanceRun(ctx, projectID); adjErr != nil {
			g.Log().Warningf(ctx, "[worker:%s] adjudication failed for project=%s: %v", w.Name(), projectID, adjErr)
		}
	}
	return nil
}

func (w *workflowEventWorker) handleRepairDraftReady(ctx context.Context, evt *events.WorkflowEvent) error {
	repairPlanJSON, _ := evt.Payload["repair_plan_json"].(string)
	if repairPlanJSON == "" {
		return nil
	}

	var plan struct {
		UpdatedTasks []struct {
			TaskKey   string `json:"task_key"`
			Name      string `json:"name"`
			BrainKind string `json:"brain_kind"`
			Summary   string `json:"summary"`
		} `json:"updated_tasks"`
	}
	if err := json.Unmarshal([]byte(repairPlanJSON), &plan); err != nil {
		return err
	}

	projectID := evt.ProjectID
	for _, task := range plan.UpdatedTasks {
		if task.TaskKey == "" || task.BrainKind == "" {
			continue
		}

		_, err := Runtime().StartBrainRun(ctx, StartBrainRunCommand{
			ProjectID:      projectID,
			TaskID:         task.TaskKey,
			BrainKind:      task.BrainKind,
			Provider:       "default",
			Prompt:         task.Summary,
			PermissionMode: "auto", // 修复任务自动接受低风险操作
		})
		if err != nil {
			g.Log().Warningf(ctx, "[worker:%s] StartBrainRun failed for task=%s brain=%s: %v",
				w.Name(), task.TaskKey, task.BrainKind, err)
			continue
		}
		g.Log().Infof(ctx, "[worker:%s] auto-restarted brain run project=%s task=%s brain=%s",
			w.Name(), projectID, task.TaskKey, task.BrainKind)
	}

	return nil
}

// ---------------------------------------------------------------------------
// DB helpers
// ---------------------------------------------------------------------------

type workflowEventRecord struct {
	ID         string
	ProjectID  string
	EventType  string
	PayloadJSON string
	RetryCount int
	CreatedAt  string
}

func listPendingWorkflowEvents(ctx context.Context, limit int) ([]*events.WorkflowEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	result, err := g.DB().GetAll(ctx,
		`SELECT id, project_id, event_type, payload_json, retry_count, created_at
		 FROM workflow_events
		 WHERE status = ?
		 ORDER BY created_at ASC
		 LIMIT ?`,
		"pending", limit,
	)
	if err != nil {
		return nil, err
	}

	var out []*events.WorkflowEvent
	for _, row := range result {
		r := workflowEventRecord{
			ID:          row["id"].String(),
			ProjectID:   row["project_id"].String(),
			EventType:   row["event_type"].String(),
			PayloadJSON: row["payload_json"].String(),
			RetryCount:  row["retry_count"].Int(),
			CreatedAt:   row["created_at"].String(),
		}
		var payload map[string]interface{}
		_ = json.Unmarshal([]byte(r.PayloadJSON), &payload)

		out = append(out, &events.WorkflowEvent{
			ID:         r.ID,
			ProjectID:  r.ProjectID,
			EventType:  events.WorkflowEventType(r.EventType),
			Payload:    payload,
			Status:     "pending",
			RetryCount: r.RetryCount,
			CreatedAt:  parseTimeOrNow(r.CreatedAt),
		})
	}
	return out, nil
}

func markWorkflowEventCompleted(ctx context.Context, eventID string) {
	_, err := g.DB().Exec(ctx,
		`UPDATE workflow_events SET status = ?, processed_at = ? WHERE id = ?`,
		"completed", time.Now().UTC().Format(time.RFC3339), eventID)
	if err != nil {
		g.Log().Warningf(ctx, "[workflow_event_worker] mark completed failed for %s: %v", eventID, err)
	}
}

func markWorkflowEventFailed(ctx context.Context, eventID string, errMsg string) {
	_, err := g.DB().Exec(ctx,
		`UPDATE workflow_events SET status = ?, error_message = ?, retry_count = retry_count + 1 WHERE id = ?`,
		"failed", errMsg, eventID)
	if err != nil {
		g.Log().Warningf(ctx, "[workflow_event_worker] mark failed failed for %s: %v", eventID, err)
	}
}

func parseTimeOrNow(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}
