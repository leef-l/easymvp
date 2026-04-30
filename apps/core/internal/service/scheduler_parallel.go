package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ---------------------------------------------------------------------------
// Layered Parallel Execution (DAG-based task scheduling)
// ---------------------------------------------------------------------------

// LayerExecutionProgress tracks the execution state of a single topology layer.
type LayerExecutionProgress struct {
	LayerIndex   int    `json:"layerIndex"`
	TotalLayers  int    `json:"totalLayers"`
	TasksInLayer int    `json:"tasksInLayer"`
	Completed    int    `json:"completed"`
	Failed       int    `json:"failed"`
	Status       string `json:"status"` // running / completed / partial_failure
}

// executeLayeredPlan executes tasks organised into topological layers. Within
// each layer all tasks are launched in parallel (each gets its own Brain Run).
// The next layer starts only after every task in the current layer reaches a
// terminal state. A single task failure does NOT block other tasks in the same
// layer, but the failure is recorded and the layer is marked partial_failure.
func executeLayeredPlan(ctx context.Context, projectID string, layers [][]string) error {
	if len(layers) == 0 {
		return nil
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return fmt.Errorf("executeLayeredPlan: open db: %w", err)
	}
	defer closeFn()

	// Build a lookup from task ID -> DomainTasks entity.
	taskMap, err := loadDomainTaskMap(ctx, db, projectID)
	if err != nil {
		return fmt.Errorf("executeLayeredPlan: load tasks: %w", err)
	}

	w := &taskSchedulerWorker{}

	for layerIdx, taskIDs := range layers {
		progress := &LayerExecutionProgress{
			LayerIndex:   layerIdx,
			TotalLayers:  len(layers),
			TasksInLayer: len(taskIDs),
			Status:       "running",
		}

		g.Log().Infof(ctx, "[scheduler_parallel] project=%s layer=%d/%d starting %d tasks",
			projectID, layerIdx+1, len(layers), len(taskIDs))

		var wg sync.WaitGroup
		var mu sync.Mutex
		runIDs := make([]string, 0, len(taskIDs))

		for _, taskID := range taskIDs {
			task, ok := taskMap[taskID]
			if !ok {
				g.Log().Warningf(ctx, "[scheduler_parallel] task=%s not found in domain_tasks, skipping", taskID)
				mu.Lock()
				progress.Failed++
				mu.Unlock()
				continue
			}

			// Skip terminal tasks (support crash-recovery / re-scheduling).
			if isTerminalStatus(task.Status) {
				g.Log().Infof(ctx, "[scheduler_parallel] task=%s already terminal (%s), skipping", taskID, task.Status)
				mu.Lock()
				progress.Completed++
				mu.Unlock()
				continue
			}

			wg.Add(1)
			go func(t entity.DomainTasks) {
				defer wg.Done()
				runID, launchErr := w.launchTaskBrainRun(ctx, projectID, t)
				mu.Lock()
				defer mu.Unlock()
				if launchErr != nil {
					progress.Failed++
					g.Log().Warningf(ctx, "[scheduler_parallel] task=%s launch failed: %v", t.Id, launchErr)
					return
				}
				if runID != "" {
					runIDs = append(runIDs, runID)
				}
			}(task)
		}

		wg.Wait()

		// Wait for all brain runs in this layer to reach terminal state.
		if len(runIDs) > 0 {
			w.waitForBrainRuns(ctx, projectID, runIDs)
		}

		// Tally completed runs (those that were not failures at launch).
		mu.Lock()
		progress.Completed += len(runIDs)
		switch {
		case progress.Failed == 0:
			progress.Status = "completed"
		case progress.Failed < progress.TasksInLayer:
			progress.Status = "partial_failure"
		default:
			progress.Status = "failed"
		}
		mu.Unlock()

		g.Log().Infof(ctx,
			"[scheduler_parallel] project=%s layer=%d/%d finished — status=%s completed=%d failed=%d",
			projectID, layerIdx+1, len(layers), progress.Status, progress.Completed, progress.Failed)

		_ = insertAuditLog(ctx, projectID, "scheduler.layer_completed", "system",
			fmt.Sprintf("Layer %d/%d %s (completed=%d, failed=%d)",
				layerIdx+1, len(layers), progress.Status, progress.Completed, progress.Failed),
			map[string]any{
				"layer_index":   layerIdx,
				"total_layers":  len(layers),
				"tasks_in_layer": progress.TasksInLayer,
				"completed":     progress.Completed,
				"failed":        progress.Failed,
				"status":        progress.Status,
			})
	}

	g.Log().Infof(ctx, "[scheduler_parallel] project=%s all %d layers executed", projectID, len(layers))
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// isTerminalStatus returns true for task statuses that indicate execution is
// finished (successfully or not).
func isTerminalStatus(status string) bool {
	switch status {
	case "completed", "failed", "cancelled":
		return true
	}
	return false
}

// loadDomainTaskMap fetches all domain tasks for a project and returns them
// indexed by task ID for O(1) lookup.
func loadDomainTaskMap(ctx context.Context, db *sql.DB, projectID string) (map[string]entity.DomainTasks, error) {
	tasks, err := listProjectDomainTasks(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]entity.DomainTasks, len(tasks))
	for _, t := range tasks {
		m[t.Id] = t
	}
	return m, nil
}
