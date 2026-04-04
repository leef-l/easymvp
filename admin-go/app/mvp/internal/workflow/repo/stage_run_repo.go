package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// StageRunRepo 阶段运行仓储。
type StageRunRepo struct{}

func NewStageRunRepo() *StageRunRepo { return &StageRunRepo{} }

func (r *StageRunRepo) table() string { return "mvp_stage_run" }

// Create 创建阶段运行记录。
func (r *StageRunRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *StageRunRepo) GetByID(ctx context.Context, id int64) (*entity.MvpStageRun, error) {
	var ent entity.MvpStageRun
	err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Scan(&ent)
	return &ent, err
}

// ListByWorkflow 查询工作流下所有阶段。
func (r *StageRunRepo) ListByWorkflow(ctx context.Context, workflowRunID int64) ([]entity.MvpStageRun, error) {
	var list []entity.MvpStageRun
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("stage_no").
		Scan(&list)
	return list, err
}

// UpdateStatus CAS 更新状态。
func (r *StageRunRepo) UpdateStatus(ctx context.Context, id int64, from, to string, extra g.Map) (int64, error) {
	data := g.Map{"status": to}
	for k, v := range extra {
		data[k] = v
	}
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
		Where("status", from).
		WhereNull("deleted_at").
		Data(data).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}
