package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// AcceptRunRepo 验收运行仓储。
type AcceptRunRepo struct{}

func NewAcceptRunRepo() *AcceptRunRepo { return &AcceptRunRepo{} }

func (r *AcceptRunRepo) table() string { return "mvp_accept_run" }

// Create 创建验收运行记录。
func (r *AcceptRunRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *AcceptRunRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetLatestByWorkflow 获取工作流最近一次验收运行。
func (r *AcceptRunRepo) GetLatestByWorkflow(ctx context.Context, workflowRunID int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("accept_round").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetNextRound 获取下一轮验收轮次。
func (r *AcceptRunRepo) GetNextRound(ctx context.Context, workflowRunID int64) (int, error) {
	maxRound, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Max("accept_round")
	if err != nil {
		return 1, err
	}
	return int(maxRound) + 1, nil
}

// CountRunningByStageRun 统计指定 stage_run 下运行中的验收任务。
func (r *AcceptRunRepo) CountRunningByStageRun(ctx context.Context, stageRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("status", "running").
		WhereNull("deleted_at").
		Count()
}

// UpdateStatus CAS 更新状态。
func (r *AcceptRunRepo) UpdateStatus(ctx context.Context, id int64, from, to string, extra g.Map) (int64, error) {
	data := g.Map{"status": to, "updated_at": gtime.Now()}
	for k, v := range extra {
		data[k] = v
	}
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
		Where("status", from).
		WhereNull("deleted_at").
		Data(data).Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// UpdateDecision 写入验收决策。
func (r *AcceptRunRepo) UpdateDecision(ctx context.Context, id int64, decision string, score float64, summary string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		Data(g.Map{
			"decision":    decision,
			"score":       score,
			"summary":     summary,
			"finished_at": gtime.Now(),
			"updated_at":  gtime.Now(),
		}).Update()
	return err
}

// UpdateRulesSnapshot 写入规则快照。
func (r *AcceptRunRepo) UpdateRulesSnapshot(ctx context.Context, id int64, rulesSnapshot string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		Data(g.Map{
			"rules_snapshot_ref": rulesSnapshot,
			"updated_at":         gtime.Now(),
		}).
		Update()
	return err
}

// SoftDeleteByWorkflow 软删除工作流下的验收运行。
func (r *AcceptRunRepo) SoftDeleteByWorkflow(ctx context.Context, workflowRunID int64, deletedAt *gtime.Time) error {
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
