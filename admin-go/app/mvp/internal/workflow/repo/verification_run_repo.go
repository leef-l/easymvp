package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// VerificationRunRepo 验证运行仓储。
type VerificationRunRepo struct{}

func NewVerificationRunRepo() *VerificationRunRepo { return &VerificationRunRepo{} }

func (r *VerificationRunRepo) table() string { return "mvp_verification_run" }

// Create 创建验证运行记录。
func (r *VerificationRunRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *VerificationRunRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetLatestByWorkflow 获取最近一次验证运行。
func (r *VerificationRunRepo) GetLatestByWorkflow(ctx context.Context, workflowRunID int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("verification_round").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetNextRound 获取下一轮验证轮次。
func (r *VerificationRunRepo) GetNextRound(ctx context.Context, workflowRunID int64) (int, error) {
	maxRound, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Max("verification_round")
	if err != nil {
		return 1, err
	}
	return int(maxRound) + 1, nil
}

// CountRunningByWorkflow 统计运行中的验证任务。
func (r *VerificationRunRepo) CountRunningByWorkflow(ctx context.Context, workflowRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", "running").
		WhereNull("deleted_at").
		Count()
}

// UpdateStatus CAS 更新状态。
func (r *VerificationRunRepo) UpdateStatus(ctx context.Context, id int64, from, to string, extra g.Map) (int64, error) {
	data := g.Map{"status": to, "updated_at": gtime.Now()}
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
