package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// VerificationIssueRepo 验证问题仓储。
type VerificationIssueRepo struct{}

func NewVerificationIssueRepo() *VerificationIssueRepo { return &VerificationIssueRepo{} }

func (r *VerificationIssueRepo) table() string { return "mvp_verification_issue" }

// Create 创建验证问题。
func (r *VerificationIssueRepo) Create(ctx context.Context, item g.Map) error {
	item["id"] = snowflake.Generate()
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(item)
	return err
}

// BatchCreate 批量创建验证问题。
func (r *VerificationIssueRepo) BatchCreate(ctx context.Context, items []g.Map) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i]["id"] = snowflake.Generate()
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(items)
	return err
}

// ListByVerificationRun 查询某次验证的问题。
func (r *VerificationIssueRepo) ListByVerificationRun(ctx context.Context, verificationRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("verification_run_id", verificationRunID).
		WhereNull("deleted_at").
		OrderAsc("severity").
		OrderAsc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}
