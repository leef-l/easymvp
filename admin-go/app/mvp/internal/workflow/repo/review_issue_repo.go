package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

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
