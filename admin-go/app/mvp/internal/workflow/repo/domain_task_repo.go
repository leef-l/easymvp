package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// DomainTaskRepo 领域任务仓储。
type DomainTaskRepo struct{}

func NewDomainTaskRepo() *DomainTaskRepo { return &DomainTaskRepo{} }

func (r *DomainTaskRepo) table() string { return "mvp_domain_task" }

// Create 创建领域任务。
func (r *DomainTaskRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *DomainTaskRepo) GetByID(ctx context.Context, id int64) (*entity.MvpDomainTask, error) {
	var ent entity.MvpDomainTask
	err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Scan(&ent)
	return &ent, err
}

// ListByWorkflowAndBatch 按工作流和批次查询。
func (r *DomainTaskRepo) ListByWorkflowAndBatch(ctx context.Context, workflowRunID int64, batchNo int) ([]entity.MvpDomainTask, error) {
	var list []entity.MvpDomainTask
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("batch_no", batchNo).
		WhereNull("deleted_at").
		OrderAsc("sort").
		Scan(&list)
	return list, err
}

// UpdateStatus CAS 更新状态。
func (r *DomainTaskRepo) UpdateStatus(ctx context.Context, id int64, from, to string, extra g.Map) (int64, error) {
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
