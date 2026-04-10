package repo

import (
	"context"
	"strings"

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

// GetByWorkflowAndID 查询指定工作流下的任务记录。
func (r *DomainTaskRepo) GetByWorkflowAndID(ctx context.Context, workflowRunID, taskID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", taskID).
		Where("workflow_run_id", workflowRunID).
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

// FindLatestByWorkflowAndAffectedResourceLike 查询最新命中 resourceRef 的领域任务。
func (r *DomainTaskRepo) FindLatestByWorkflowAndAffectedResourceLike(ctx context.Context, workflowRunID int64, resourceRef string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("affected_resources LIKE ?", "%"+resourceRef+"%").
		WhereNull("deleted_at").
		OrderDesc("updated_at").
		Fields("id").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByWorkflowAndStatuses 查询工作流下指定状态的任务。
func (r *DomainTaskRepo) ListByWorkflowAndStatuses(ctx context.Context, workflowRunID int64, statuses []string, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListCompletedByWorkflowAndKinds 查询工作流下指定 task_kind 且已完成的任务。
func (r *DomainTaskRepo) ListCompletedByWorkflowAndKinds(ctx context.Context, workflowRunID int64, taskKinds []string, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", "completed").
		WhereNull("deleted_at")
	if len(taskKinds) > 0 {
		model = model.WhereIn("task_kind", taskKinds)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByIDs 查询指定 ID 集合的任务。
func (r *DomainTaskRepo) ListByIDs(ctx context.Context, taskIDs []int64, fields ...string) ([]g.Map, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("id", taskIDs).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// MarkFailedForRework 将任务重置为 failed，供返工阶段消费。
func (r *DomainTaskRepo) MarkFailedForRework(ctx context.Context, taskID int64, reason string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Data(g.Map{
			"status":       "failed",
			"result":       reason,
			"started_at":   nil,
			"completed_at": nil,
		}).
		Update()
	return err
}
