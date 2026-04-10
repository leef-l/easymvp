package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// ReviewIssueRepo 审核问题仓储。
type ReviewIssueRepo struct{}

func NewReviewIssueRepo() *ReviewIssueRepo { return &ReviewIssueRepo{} }

func (r *ReviewIssueRepo) table() string { return "mvp_review_issue" }

// BatchCreate 批量创建审核问题。
func (r *ReviewIssueRepo) BatchCreate(ctx context.Context, items []g.Map) error {
	for i := range items {
		items[i]["id"] = snowflake.Generate()
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(items)
	return err
}

// ListByPlanVersion 查询版本下的审核问题。
func (r *ReviewIssueRepo) ListByPlanVersion(ctx context.Context, planVersionID int64) ([]entity.MvpReviewIssue, error) {
	var list []entity.MvpReviewIssue
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at").
		OrderDesc("severity").
		Scan(&list)
	return list, err
}

// CountOpenByStageRunAndSeverity 统计阶段下指定严重级别的 open 问题数。
func (r *ReviewIssueRepo) CountOpenByStageRunAndSeverity(ctx context.Context, stageRunID int64, severity string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("severity", severity).
		Where("status", "open").
		WhereNull("deleted_at").
		Count()
}

// ListByStageRun 查询阶段下的问题列表。
func (r *ReviewIssueRepo) ListByStageRun(ctx context.Context, stageRunID int64, onlyOpen bool, limit int) (gdb.Result, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		WhereNull("deleted_at").
		OrderDesc("severity").
		OrderDesc("created_at")
	if onlyOpen {
		model = model.Where("status", "open")
	}
	if limit > 0 {
		model = model.Limit(limit)
	}
	return model.All()
}

// ListOpenByStageRunAndIDs 查询阶段下指定的 open 问题。
func (r *ReviewIssueRepo) ListOpenByStageRunAndIDs(ctx context.Context, stageRunID int64, issueIDs []int64) (gdb.Result, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("stage_run_id", stageRunID).
		WhereIn("id", issueIDs).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("severity").
		OrderDesc("created_at").
		All()
}

// SoftDeleteByWorkflow 软删除工作流下的审核问题。
func (r *ReviewIssueRepo) SoftDeleteByWorkflow(ctx context.Context, workflowRunID int64, deletedAt *gtime.Time) error {
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
