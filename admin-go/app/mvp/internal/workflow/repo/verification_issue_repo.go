package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

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

// ListOpenByVerificationRunAndIDs 查询某次验证下指定的 open 问题。
func (r *VerificationIssueRepo) ListOpenByVerificationRunAndIDs(ctx context.Context, verificationRunID int64, issueIDs []int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("verification_run_id", verificationRunID).
		WhereIn("id", issueIDs).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// UpdateDomainTaskID 回写问题关联任务。
func (r *VerificationIssueRepo) UpdateDomainTaskID(ctx context.Context, verificationRunID, issueID, domainTaskID int64) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", issueID).
		Where("verification_run_id", verificationRunID).
		Data(g.Map{
			"domain_task_id": domainTaskID,
			"updated_at":     gtime.Now(),
		}).
		Update()
	return err
}

// MarkReworkRequested 批量标记问题已进入返工。
func (r *VerificationIssueRepo) MarkReworkRequested(ctx context.Context, verificationRunID int64, issueIDs []int64) error {
	if len(issueIDs) == 0 {
		return nil
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("verification_run_id", verificationRunID).
		WhereIn("id", issueIDs).
		Where("status", "open").
		Data(g.Map{
			"status":     "rework_requested",
			"updated_at": gtime.Now(),
		}).
		Update()
	return err
}
