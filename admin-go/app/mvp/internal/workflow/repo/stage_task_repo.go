package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// StageTaskRepo 阶段任务仓储。
type StageTaskRepo struct{}

func NewStageTaskRepo() *StageTaskRepo { return &StageTaskRepo{} }

func (r *StageTaskRepo) table() string { return "mvp_stage_task" }

// Create 创建阶段任务。
func (r *StageTaskRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByStageRunAndTaskType 查询阶段下指定类型任务。
func (r *StageTaskRepo) GetByStageRunAndTaskType(ctx context.Context, stageRunID int64, taskType string, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("task_type", taskType).
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByStageRun 查询阶段任务列表。
func (r *StageTaskRepo) ListByStageRun(ctx context.Context, stageRunID int64) (gdb.Result, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		WhereNull("deleted_at").
		OrderAsc("created_at").
		All()
}

// UpdateFields 按 ID 更新阶段任务字段。
func (r *StageTaskRepo) UpdateFields(ctx context.Context, stageTaskID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", stageTaskID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// UpdateByStageRunsStatuses 按阶段实例集合和状态集合批量更新阶段任务。
func (r *StageTaskRepo) UpdateByStageRunsStatuses(ctx context.Context, stageRunIDs []int64, statuses []string, data g.Map) error {
	if len(stageRunIDs) == 0 {
		return nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("stage_run_id", stageRunIDs).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	_, err := model.Data(data).Update()
	return err
}

// SoftDeleteByStageRuns 软删除给定阶段实例下的阶段任务。
func (r *StageTaskRepo) SoftDeleteByStageRuns(ctx context.Context, stageRunIDs []int64, deletedAt *gtime.Time) error {
	if len(stageRunIDs) == 0 {
		return nil
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("stage_run_id", stageRunIDs).
		WhereNull("deleted_at").
		Data(g.Map{
			"deleted_at": deletedAt,
			"updated_at": deletedAt,
		}).
		Update()
	return err
}
