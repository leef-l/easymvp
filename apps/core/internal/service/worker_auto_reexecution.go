package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/leef-l/easymvp/apps/core/internal/events"
)

type autoReexecutionWorker struct {
	interval  time.Duration
	batchSize int
}

func newAutoReexecutionWorker() backgroundWorker {
	cfgInterval := g.Cfg().MustGet(context.Background(), "easymvp.workers.autoReexecutionInterval", "5s").Duration()
	if cfgInterval <= 0 {
		cfgInterval = 5 * time.Second
	}
	batchSize := g.Cfg().MustGet(context.Background(), "easymvp.workers.autoReexecutionBatchSize", 10).Int()
	if batchSize <= 0 {
		batchSize = 10
	}
	return &autoReexecutionWorker{
		interval:  cfgInterval,
		batchSize: batchSize,
	}
}

func (w *autoReexecutionWorker) Name() string {
	return "auto_reexecution_worker"
}

func (w *autoReexecutionWorker) Interval() time.Duration {
	return w.interval
}

// RunOnce polls for repair plan drafts that are ready and have no human checkpoint,
// then automatically triggers re-execution for their associated projects.
func (w *autoReexecutionWorker) RunOnce(ctx context.Context) error {
	records, err := listReadyRepairDraftsWithoutCheckpoint(ctx, w.batchSize)
	if err != nil {
		return err
	}

	for _, draft := range records {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Double-check project status is still 'reworking'.
		project, err := getProjectByID(ctx, nil, draft.ProjectID)
		if err != nil || project == nil || project.Status != "reworking" {
			continue
		}

		// Publish event to trigger re-execution via the event bus.
		events.Publish(ctx, &events.WorkflowEvent{
			ProjectID: draft.ProjectID,
			EventType: events.RepairDraftReady,
			Payload: map[string]interface{}{
				"repair_plan_draft_id": draft.ID,
				"project_id":           draft.ProjectID,
				"repair_plan_json":     draft.RepairPlanJSON,
			},
		})

		g.Log().Infof(ctx, "[worker:%s] published RepairDraftReady for project=%s draft=%s",
			w.Name(), draft.ProjectID, draft.ID)
	}

	return nil
}

// readyRepairDraftRecord is a minimal projection for the auto-reexecution worker.
type readyRepairDraftRecord struct {
	ID            string
	ProjectID     string
	RepairPlanJSON string
}

// listReadyRepairDraftsWithoutCheckpoint returns repair plan drafts with status='ready'
// that have no human_checkpoint_required, ordered by created_at.
func listReadyRepairDraftsWithoutCheckpoint(ctx context.Context, limit int) ([]*readyRepairDraftRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	result, err := g.DB().GetAll(ctx,
		`SELECT id, project_id, repair_plan_json
		 FROM repair_plan_drafts
		 WHERE status = ?
		 ORDER BY created_at ASC
		 LIMIT ?`,
		"ready", limit,
	)
	if err != nil {
		return nil, err
	}

	var out []*readyRepairDraftRecord
	for _, row := range result {
		r := readyRepairDraftRecord{
			ID:             row["id"].String(),
			ProjectID:      row["project_id"].String(),
			RepairPlanJSON: row["repair_plan_json"].String(),
		}
		// Parse human_checkpoint_required from RepairPlanJSON.
		var parsed struct {
			HumanCheckpointRequired bool `json:"human_checkpoint_required"`
		}
		_ = json.Unmarshal([]byte(r.RepairPlanJSON), &parsed)
		if parsed.HumanCheckpointRequired {
			continue
		}
		out = append(out, &r)
	}
	return out, nil
}
