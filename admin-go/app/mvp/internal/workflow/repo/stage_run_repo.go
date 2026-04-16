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

// CreateInTx 在事务中创建阶段运行记录。
func (r *StageRunRepo) CreateInTx(ctx context.Context, tx gdb.TX, data g.Map) error {
	_, err := tx.Model(r.table()).Ctx(ctx).Insert(data)
	return err
}

// GetByID 按 ID 查询。
func (r *StageRunRepo) GetByID(ctx context.Context, id int64) (*entity.MvpStageRun, error) {
	var ent entity.MvpStageRun
	err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Scan(&ent)
	return &ent, err
}

// GetByIDMap 按 ID 查询阶段运行记录；可选字段列表用于减少读取范围。
func (r *StageRunRepo) GetByIDMap(ctx context.Context, id int64, fields ...string) (g.Map, error) {
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

// ListByWorkflowMaps 查询工作流下所有阶段；可选字段列表用于减少读取范围。
func (r *StageRunRepo) ListByWorkflowMaps(ctx context.Context, workflowRunID int64, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("stage_no")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByWorkflowStatuses 查询工作流下命中状态集合的阶段列表。
func (r *StageRunRepo) ListByWorkflowStatuses(ctx context.Context, workflowRunID int64, statuses []string, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.OrderAsc("stage_no").All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
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

// UpdateFieldsIfStatus 在状态命中时更新阶段字段。
func (r *StageRunRepo) UpdateFieldsIfStatus(ctx context.Context, stageRunID int64, status string, data g.Map) (int64, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", stageRunID).
		WhereNull("deleted_at")
	if status != "" {
		model = model.Where("status", status)
	}
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// UpdateFields 按 ID 更新阶段字段。
func (r *StageRunRepo) UpdateFields(ctx context.Context, stageRunID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", stageRunID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// UpdateByIDs 批量更新给定阶段实例。
func (r *StageRunRepo) UpdateByIDs(ctx context.Context, stageRunIDs []int64, data g.Map) error {
	if len(stageRunIDs) == 0 {
		return nil
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("id", stageRunIDs).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// GetStatusByID 查询阶段状态。
func (r *StageRunRepo) GetStatusByID(ctx context.Context, stageRunID int64) (string, error) {
	value, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", stageRunID).
		WhereNull("deleted_at").
		Value("status")
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

// ListCompletedStageTypes 查询已完成的阶段类型。
func (r *StageRunRepo) ListCompletedStageTypes(ctx context.Context, workflowRunID int64, stageTypes []string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", "completed").
		WhereNull("deleted_at")
	if len(stageTypes) > 0 {
		model = model.WhereIn("stage_type", stageTypes)
	}
	records, err := model.Fields(strings.Join([]string{"DISTINCT stage_type"}, ",")).All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// GetLatestByWorkflowAndType 查询工作流下最新指定类型的阶段。
func (r *StageRunRepo) GetLatestByWorkflowAndType(ctx context.Context, workflowRunID int64, stageType string, fields ...string) (gdb.Record, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", stageType).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("stage_no").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record, nil
}

// GetLatestByWorkflow 查询工作流下最新阶段。
func (r *StageRunRepo) GetLatestByWorkflow(ctx context.Context, workflowRunID int64, fields ...string) (gdb.Record, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("stage_no").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record, nil
}

// GetLatestByWorkflowTypeStatuses 查询工作流下最新命中状态集合的指定类型阶段。
func (r *StageRunRepo) GetLatestByWorkflowTypeStatuses(ctx context.Context, workflowRunID int64, stageType string, statuses []string, fields ...string) (gdb.Record, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", stageType).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.OrderDesc("stage_no").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record, nil
}

// GetMaxStageNoByWorkflow 查询工作流当前最大 stage_no。
func (r *StageRunRepo) GetMaxStageNoByWorkflow(ctx context.Context, workflowRunID int64) (int, error) {
	val, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Max("stage_no")
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

// GetMaxStageNoByWorkflowInTx 查询事务内工作流当前最大 stage_no。
func (r *StageRunRepo) GetMaxStageNoByWorkflowInTx(ctx context.Context, tx gdb.TX, workflowRunID int64) (int, error) {
	val, err := tx.Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Max("stage_no")
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

// CountByWorkflow 统计工作流下阶段数。
func (r *StageRunRepo) CountByWorkflow(ctx context.Context, workflowRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Count()
}

// CountByWorkflowAndType 统计工作流下指定类型阶段数。
func (r *StageRunRepo) CountByWorkflowAndType(ctx context.Context, workflowRunID int64, stageType string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", stageType).
		WhereNull("deleted_at").
		Count()
}

// UpdateFieldsByWorkflowAndStatusesInTx 在事务中按工作流和状态集合批量更新阶段实例。
func (r *StageRunRepo) UpdateFieldsByWorkflowAndStatusesInTx(ctx context.Context, tx gdb.TX, workflowRunID int64, statuses []string, data g.Map) error {
	model := tx.Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	_, err := model.Data(data).Update()
	return err
}

// SoftDeleteByWorkflow 软删除工作流下所有阶段实例。
func (r *StageRunRepo) SoftDeleteByWorkflow(ctx context.Context, workflowRunID int64, deletedAt *gtime.Time) error {
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
