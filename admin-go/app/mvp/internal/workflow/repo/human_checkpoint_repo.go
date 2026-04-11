package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// HumanCheckpointRepo 人工介入节点仓储。
type HumanCheckpointRepo struct{}

// NewHumanCheckpointRepo 创建人工节点仓储。
func NewHumanCheckpointRepo() *HumanCheckpointRepo { return &HumanCheckpointRepo{} }

func (r *HumanCheckpointRepo) table() string { return "mvp_human_checkpoint" }

// Create 创建人工节点。
func (r *HumanCheckpointRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *HumanCheckpointRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetByActionID 按决策动作 ID 查询。
func (r *HumanCheckpointRepo) GetByActionID(ctx context.Context, actionID int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("decision_action_id", actionID).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// UpdateHandle 更新处理结果。
func (r *HumanCheckpointRepo) UpdateHandle(ctx context.Context, id int64, data g.Map) error {
	data["updated_at"] = gtime.Now()
	_, err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Update(data)
	return err
}

// ListOpen 查询项目下未处理的节点。
func (r *HumanCheckpointRepo) ListOpen(ctx context.Context, projectID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("created_at").
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

// CountOpenByProject 统计项目下 open 节点数。
func (r *HumanCheckpointRepo) CountOpenByProject(ctx context.Context, projectID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "open").
		WhereNull("deleted_at").
		Count()
}

// ListByWorkflow 按工作流查询。
func (r *HumanCheckpointRepo) ListByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
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

// ListByProjectStatus 查询项目下指定状态的人工检查点。
func (r *HumanCheckpointRepo) ListByProjectStatus(ctx context.Context, projectID int64, status string, limit int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if status != "" {
		model = model.Where("status", status)
	}
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

// UpdateStatusByProject 批量更新项目下指定状态的检查点。
func (r *HumanCheckpointRepo) UpdateStatusByProject(ctx context.Context, projectID int64, fromStatus string, data g.Map) (int64, error) {
	data["updated_at"] = gtime.Now()
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if fromStatus != "" {
		model = model.Where("status", fromStatus)
	}
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}
