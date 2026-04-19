package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	workspacev1 "github.com/leef-l/easymvp/apps/core/api/workspace/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
)

type workspaceSnapshotRefreshWorker struct {
	interval  time.Duration
	batchSize int
}

func newWorkspaceSnapshotRefreshWorker() backgroundWorker {
	return &workspaceSnapshotRefreshWorker{
		interval:  15 * time.Second,
		batchSize: 8,
	}
}

func (w *workspaceSnapshotRefreshWorker) Name() string {
	return "workspace_snapshot_refresh_worker"
}

func (w *workspaceSnapshotRefreshWorker) Interval() time.Duration {
	return w.interval
}

func (w *workspaceSnapshotRefreshWorker) RunOnce(ctx context.Context) error {
	projectIDs, err := listProjectsForWorkspaceRefresh(ctx, w.batchSize)
	if err != nil {
		return err
	}

	for _, projectID := range projectIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err = refreshProjectWorkspaceSnapshot(ctx, projectID); err != nil {
			handleWorkerFailure(
				ctx,
				w.Name(),
				projectID,
				"WORKER_WORKSPACE_REFRESH",
				"workspace snapshot refresh failed",
				map[string]any{
					"project_id": projectID,
					"error":      err.Error(),
				},
			)
			continue
		}
	}
	if err = refreshWorkspaceHomeSnapshot(ctx); err != nil {
		handleWorkerFailure(
			ctx,
			w.Name(),
			"",
			"WORKER_WORKSPACE_HOME_REFRESH",
			"workspace home snapshot refresh failed",
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	if len(projectIDs) > 0 {
		g.Log().Debugf(ctx, "[worker:%s] refreshed %d project snapshots", w.Name(), len(projectIDs))
	}
	return nil
}

func refreshProjectWorkspaceSnapshot(ctx context.Context, projectID string) error {
	view, err := Projects().GetProjectWorkspaceView(ctx, projectID)
	if err != nil {
		return err
	}
	return persistProjectSnapshot(ctx, projectID, view)
}

func refreshWorkspaceHomeSnapshot(ctx context.Context) error {
	view, err := Workspace().GetHomeView(ctx)
	if err != nil {
		return err
	}
	return persistWorkspaceSnapshot(ctx, "home_view", view)
}

func persistProjectSnapshot(ctx context.Context, projectID string, snapshot any) error {
	return persistSnapshot(
		ctx,
		`INSERT INTO `+dao.ProjectSnapshots.Table()+` (project_id, snapshot_json, generated_at)
VALUES (?, ?, ?)
ON CONFLICT(project_id) DO UPDATE SET snapshot_json = excluded.snapshot_json, generated_at = excluded.generated_at`,
		projectID,
		snapshot,
	)
}

func persistWorkspaceSnapshot(ctx context.Context, key string, snapshot any) error {
	return persistSnapshot(
		ctx,
		`INSERT INTO `+dao.WorkspaceSnapshots.Table()+` (key, snapshot_json, generated_at)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET snapshot_json = excluded.snapshot_json, generated_at = excluded.generated_at`,
		key,
		snapshot,
	)
}

func persistSnapshot(ctx context.Context, query string, primaryKey string, snapshot any) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, query, primaryKey, string(payload), nowText())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func loadProjectWorkspaceSnapshot(ctx context.Context, projectID string) (*projectsv1.ProjectWorkspaceViewRes, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT snapshot_json FROM `+dao.ProjectSnapshots.Table()+` WHERE project_id = ? LIMIT 1`, projectID)
	var raw string
	if err = row.Scan(&raw); err != nil {
		return nil, err
	}
	var res projectsv1.ProjectWorkspaceViewRes
	if err = json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func loadWorkspaceHomeSnapshot(ctx context.Context, key string) (*workspacev1.HomeViewRes, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT snapshot_json FROM `+dao.WorkspaceSnapshots.Table()+` WHERE key = ? LIMIT 1`, key)
	var raw string
	if err = row.Scan(&raw); err != nil {
		return nil, err
	}
	var res workspacev1.HomeViewRes
	if err = json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, err
	}
	return &res, nil
}
