package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/middleware"
	"easymvp/utility/snowflake"
)

// SituationSnapshotRepo 态势快照仓储。
type SituationSnapshotRepo struct{}

func NewSituationSnapshotRepo() *SituationSnapshotRepo { return &SituationSnapshotRepo{} }

func (r *SituationSnapshotRepo) model(ctx context.Context) *gdb.Model {
	return g.DB().Model("mvp_situation_snapshot").Ctx(ctx).WhereNull("deleted_at")
}

func (r *SituationSnapshotRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model("mvp_situation_snapshot").Ctx(ctx).Insert(data)
	return int64(id), err
}

func (r *SituationSnapshotRepo) GetLatestByWorkflowRunID(ctx context.Context, workflowRunID int64) (g.Map, error) {
	record, err := middleware.ApplyDataScope(ctx, r.model(ctx), "created_by", "dept_id").
		Where("workflow_run_id", workflowRunID).
		OrderDesc("created_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

func (r *SituationSnapshotRepo) ListByProjectID(ctx context.Context, projectID int64, limit int) ([]g.Map, error) {
	if limit <= 0 {
		limit = 20
	}
	records, err := middleware.ApplyDataScope(ctx, r.model(ctx), "created_by", "dept_id").
		Where("project_id", projectID).
		OrderDesc("created_at").
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}
