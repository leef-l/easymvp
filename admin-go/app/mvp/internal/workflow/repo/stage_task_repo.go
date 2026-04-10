package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// StageTaskRepo 阶段任务仓储。
type StageTaskRepo struct{}

func NewStageTaskRepo() *StageTaskRepo { return &StageTaskRepo{} }

func (r *StageTaskRepo) table() string { return "mvp_stage_task" }

// ListByStageRun 查询阶段任务列表。
func (r *StageTaskRepo) ListByStageRun(ctx context.Context, stageRunID int64) (gdb.Result, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		WhereNull("deleted_at").
		OrderAsc("created_at").
		All()
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
