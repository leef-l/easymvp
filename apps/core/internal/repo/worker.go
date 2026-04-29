package repo

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
)

func ListPendingRunBindingIDs(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}

	var results []struct {
		Id string `orm:"id"`
	}
	err := dao.BrainRunBindings.Ctx(ctx).
		Where("run_status", []string{"run_pending", "run_active"}).
		Order("updated_at ASC").
		Limit(limit).
		Scan(&results)
	if err != nil {
		return nil, gerror.Wrap(err, "query pending run bindings failed")
	}

	ids := make([]string, 0, len(results))
	for _, r := range results {
		if r.Id != "" {
			ids = append(ids, r.Id)
		}
	}
	return ids, nil
}

func ListProjectsForWorkspaceRefresh(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 12
	}

	var results []struct {
		Id string `orm:"id"`
	}
	err := dao.Projects.Ctx(ctx).
		Where("status <> ?", "completed").
		Order("updated_at DESC").
		Limit(limit).
		Scan(&results)
	if err != nil {
		return nil, gerror.Wrap(err, "query workspace refresh projects failed")
	}

	projectIDs := make([]string, 0, len(results))
	for _, r := range results {
		if r.Id != "" {
			projectIDs = append(projectIDs, r.Id)
		}
	}
	return projectIDs, nil
}

func InsertWorkerAuditLog(ctx context.Context, workerName, projectID, summary string, detail map[string]any) error {
	result, err := dao.AuditLogs.Ctx(ctx).Data(g.Map{
		"id":           newResourceID("audit"),
		"project_id":   projectID,
		"event_type":   "worker.error",
		"actor_kind":   "worker:" + workerName,
		"summary":      summary,
		"payload_json": marshalDiagnosticDetail(detail),
		"created_at":   nowText(),
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "insert worker audit log failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert worker audit log affected unexpected rows")
	}
	return nil
}
