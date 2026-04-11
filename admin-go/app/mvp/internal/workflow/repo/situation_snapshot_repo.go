package repo

import (
	"context"
	"strings"

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

// ListLatestByWorkflowRunIDs 查询工作流集合下各自最新一条态势快照。
func (r *SituationSnapshotRepo) ListLatestByWorkflowRunIDs(ctx context.Context, workflowRunIDs []int64, fields ...string) ([]g.Map, error) {
	if len(workflowRunIDs) == 0 {
		return nil, nil
	}

	type latestSituationSnapshotID struct {
		WorkflowRunID int64 `orm:"workflow_run_id"`
		ID            int64 `orm:"id"`
	}

	var latestIDs []latestSituationSnapshotID
	if err := g.DB().Model("mvp_situation_snapshot").Ctx(ctx).
		WhereIn("workflow_run_id", workflowRunIDs).
		WhereNull("deleted_at").
		Fields("workflow_run_id, MAX(id) AS id").
		Group("workflow_run_id").
		Scan(&latestIDs); err != nil {
		return nil, err
	}
	if len(latestIDs) == 0 {
		return nil, nil
	}

	snapshotIDs := make([]int64, 0, len(latestIDs))
	for _, item := range latestIDs {
		if item.ID > 0 {
			snapshotIDs = append(snapshotIDs, item.ID)
		}
	}
	if len(snapshotIDs) == 0 {
		return nil, nil
	}

	model := g.DB().Model("mvp_situation_snapshot").Ctx(ctx).
		WhereIn("id", snapshotIDs).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}

	records, err := model.All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}
