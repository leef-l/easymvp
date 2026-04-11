package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

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

// GetByIDMap 按 ID 查询任务记录；可选字段列表用于减少读取范围。
func (r *DomainTaskRepo) GetByIDMap(ctx context.Context, taskID int64, fields ...string) (g.Map, error) {
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

// GetByProjectAndID 查询指定项目下的任务记录。
func (r *DomainTaskRepo) GetByProjectAndID(ctx context.Context, projectID, taskID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()+" t").Ctx(ctx).
		InnerJoin("mvp_workflow_run wf", "wf.id = t.workflow_run_id").
		Where("t.id", taskID).
		Where("wf.project_id", projectID).
		WhereNull("t.deleted_at")
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

// ListByWorkflowOrdered 查询工作流下的任务列表，按 batch_no/sort 升序。
func (r *DomainTaskRepo) ListByWorkflowOrdered(ctx context.Context, workflowRunID int64, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("batch_no").
		OrderAsc("sort")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListRecentByWorkflow 查询工作流下最近更新的任务列表。
func (r *DomainTaskRepo) ListRecentByWorkflow(ctx context.Context, workflowRunID int64, limit int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("updated_at")
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

// ListByWorkflowFiltered 查询工作流下的任务列表，支持状态/批次过滤并按 batch_no/sort 升序。
func (r *DomainTaskRepo) ListByWorkflowFiltered(ctx context.Context, workflowRunID int64, status string, batchNo int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("batch_no").
		OrderAsc("sort")
	if status != "" {
		model = model.Where("status", status)
	}
	if batchNo > 0 {
		model = model.Where("batch_no", batchNo)
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

// ListByWorkflowRunIDs 查询工作流集合下的任务列表。
func (r *DomainTaskRepo) ListByWorkflowRunIDs(ctx context.Context, workflowRunIDs []int64, fields ...string) ([]g.Map, error) {
	if len(workflowRunIDs) == 0 {
		return nil, nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("workflow_run_id", workflowRunIDs).
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

// ListStatRowsByWorkflowRunIDs 查询工作流集合下的任务聚合统计。
func (r *DomainTaskRepo) ListStatRowsByWorkflowRunIDs(ctx context.Context, workflowRunIDs []int64) ([]g.Map, error) {
	if len(workflowRunIDs) == 0 {
		return nil, nil
	}
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("workflow_run_id", workflowRunIDs).
		WhereNull("deleted_at").
		Fields(`
			workflow_run_id,
			COUNT(*) AS total_tasks,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed_tasks,
			SUM(CASE WHEN status IN ('failed', 'escalated') THEN 1 ELSE 0 END) AS failed_tasks,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) AS running_tasks`).
		Group("workflow_run_id").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListStatusRowsByWorkflow 统计工作流下任务状态分布。
func (r *DomainTaskRepo) ListStatusRowsByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("status, COUNT(*) as count").
		Group("status").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
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

// UpdateFields 按 ID 更新任务字段。
func (r *DomainTaskRepo) UpdateFields(ctx context.Context, taskID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// UpdateByIDsStatuses 按任务 ID 集合和状态集合批量更新任务。
func (r *DomainTaskRepo) UpdateByIDsStatuses(ctx context.Context, taskIDs []int64, statuses []string, data g.Map) error {
	if len(taskIDs) == 0 {
		return nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("id", taskIDs).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	_, err := model.Data(data).Update()
	return err
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

// GetLatestBatchNoByWorkflowStatuses 查询指定状态集合下最新批次号。
func (r *DomainTaskRepo) GetLatestBatchNoByWorkflowStatuses(ctx context.Context, workflowRunID int64, statuses []string) (int, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	value, err := model.Max("batch_no")
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

// CountByWorkflow 统计工作流下任务数。
func (r *DomainTaskRepo) CountByWorkflow(ctx context.Context, workflowRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Count()
}

// ResetForRetry 将失败任务重置为 pending 并增加重试次数。
func (r *DomainTaskRepo) ResetForRetry(ctx context.Context, taskID int64) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", taskID).
		WhereIn("status", g.Slice{"failed", "escalated"}).
		WhereNull("deleted_at").
		Data(g.Map{
			"status":           "pending",
			"retry_count":      gdb.Raw("retry_count + 1"),
			"result":           nil,
			"error_message":    nil,
			"started_at":       nil,
			"completed_at":     nil,
			"heartbeat_at":     nil,
			"locked_resources": nil,
			"updated_at":       gtime.Now(),
		}).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// CompleteAsSkipped 将任务标记为 completed/skipped。
func (r *DomainTaskRepo) CompleteAsSkipped(ctx context.Context, taskID int64, completedAt *gtime.Time) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", taskID).
		WhereIn("status", g.Slice{"pending", "failed", "escalated"}).
		WhereNull("deleted_at").
		Data(g.Map{
			"status":       "completed",
			"result":       "skipped",
			"completed_at": completedAt,
			"updated_at":   completedAt,
		}).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// SoftDeleteByWorkflow 软删除工作流下的领域任务。
func (r *DomainTaskRepo) SoftDeleteByWorkflow(ctx context.Context, workflowRunID int64, deletedAt *gtime.Time) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Data(g.Map{
			"deleted_at": deletedAt,
			"updated_at": deletedAt,
		}).
		Update()
	return err
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
