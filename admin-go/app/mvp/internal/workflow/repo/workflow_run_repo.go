// Package repo 新架构专用仓储层，不与旧 DAO 混写。
package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// WorkflowRunRepo 工作流运行仓储。
type WorkflowRunRepo struct{}

func NewWorkflowRunRepo() *WorkflowRunRepo { return &WorkflowRunRepo{} }

func (r *WorkflowRunRepo) table() string { return "mvp_workflow_run" }

// NextRunNo 获取项目下一个 run_no。
func (r *WorkflowRunRepo) NextRunNo(ctx context.Context, projectID int64) (int, error) {
	val, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Max("run_no")
	if err != nil {
		return 0, err
	}
	return int(val) + 1, nil
}

// Create 创建工作流运行记录。
func (r *WorkflowRunRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *WorkflowRunRepo) GetByID(ctx context.Context, id int64) (*entity.MvpWorkflowRun, error) {
	var ent entity.MvpWorkflowRun
	err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Scan(&ent)
	return &ent, err
}

// GetActiveByProject 查询项目的活跃工作流运行。
func (r *WorkflowRunRepo) GetActiveByProject(ctx context.Context, projectID int64) (*entity.MvpWorkflowRun, error) {
	var ent entity.MvpWorkflowRun
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"pending", "running", "paused"}).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		Limit(1).
		Scan(&ent)
	return &ent, err
}

// UpdateStatus CAS 更新状态。
func (r *WorkflowRunRepo) UpdateStatus(ctx context.Context, id int64, from, to string, extra g.Map) (int64, error) {
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

// UpdateInTx 事务内更新。
func (r *WorkflowRunRepo) UpdateInTx(ctx context.Context, tx gdb.TX, id int64, data g.Map) error {
	_, err := tx.Model(r.table()).Ctx(ctx).Where("id", id).Data(data).Update()
	return err
}
