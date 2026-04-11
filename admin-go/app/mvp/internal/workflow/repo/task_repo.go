package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// TaskRepo 旧版任务仓储。
type TaskRepo struct{}

func NewTaskRepo() *TaskRepo { return &TaskRepo{} }

func (r *TaskRepo) table() string { return "mvp_task" }

// GetByID 按 ID 查询任务。
func (r *TaskRepo) GetByID(ctx context.Context, taskID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByProject 查询项目任务列表。
func (r *TaskRepo) ListByProject(ctx context.Context, projectID int64, limit int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderAsc("batch_no")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	if limit > 0 {
		model = model.Limit(limit)
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// CountStatusRowsByProject 统计项目任务状态分布。
func (r *TaskRepo) CountStatusRowsByProject(ctx context.Context, projectID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Fields("status, COUNT(*) as count").
		Group("status").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListIDsByProjectStatus 查询项目下指定状态的任务 ID。
func (r *TaskRepo) ListIDsByProjectStatus(ctx context.Context, projectID int64, status string) ([]int64, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("status", status).
		WhereNull("deleted_at").
		Fields("id").
		All()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record["id"].Int64())
	}
	return ids, nil
}
