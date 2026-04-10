// Package repo 新架构专用仓储层，不与旧 DAO 混写。
package repo

import (
	"context"
	"strings"

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

// GetByIDMap 按 ID 查询工作流运行；可选字段列表用于减少读取范围。
func (r *WorkflowRunRepo) GetByIDMap(ctx context.Context, id int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
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

// GetLatestByProject 查询项目下最近一次工作流运行记录。
func (r *WorkflowRunRepo) GetLatestByProject(ctx context.Context, projectID int64) (gdb.Record, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		OrderDesc("created_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record, nil
}

// UpdateFields 按 ID 更新字段。
func (r *WorkflowRunRepo) UpdateFields(ctx context.Context, id int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// GetLatestByProjectExcludingStatuses 查询项目下最近一条不在给定状态集合内的工作流。
func (r *WorkflowRunRepo) GetLatestByProjectExcludingStatuses(ctx context.Context, projectID int64, excluded []string, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if len(excluded) > 0 {
		model = model.WhereNotIn("status", excluded)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("run_no").OrderDesc("created_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
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

// GetLatestByProjectStatuses 查询项目下最近一条命中状态集合的工作流。
func (r *WorkflowRunRepo) GetLatestByProjectStatuses(ctx context.Context, projectID int64, statuses []string, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("created_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
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
