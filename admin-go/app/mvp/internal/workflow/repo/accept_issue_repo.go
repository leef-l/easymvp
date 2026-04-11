package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// AcceptIssueRepo 验收问题仓储。
type AcceptIssueRepo struct{}

func NewAcceptIssueRepo() *AcceptIssueRepo { return &AcceptIssueRepo{} }

func (r *AcceptIssueRepo) table() string { return "mvp_accept_issue" }

// BatchCreate 批量创建验收问题。
func (r *AcceptIssueRepo) BatchCreate(ctx context.Context, items []g.Map) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i]["id"] = snowflake.Generate()
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(items)
	return err
}

// ListByAcceptRun 按验收运行查询问题列表。
func (r *AcceptIssueRepo) ListByAcceptRun(ctx context.Context, acceptRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("accept_run_id", acceptRunID).
		WhereNull("deleted_at").
		OrderAsc("severity").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListOpenByAcceptRunAndIDs 查询某次验收下指定的 open 问题。
func (r *AcceptIssueRepo) ListOpenByAcceptRunAndIDs(ctx context.Context, acceptRunID int64, issueIDs []int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("accept_run_id", acceptRunID).
		WhereIn("id", issueIDs).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("severity").
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByWorkflow 按工作流查询问题列表。
func (r *AcceptIssueRepo) ListByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("severity").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// CountBlockersByAcceptRun 统计某次验收中 blocker 级别问题数。
func (r *AcceptIssueRepo) CountBlockersByAcceptRun(ctx context.Context, acceptRunID int64) (int, error) {
	count, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("accept_run_id", acceptRunID).
		Where("severity", "blocker").
		Where("status", "open").
		WhereNull("deleted_at").
		Count()
	return count, err
}

// CountOpenByAcceptRun 统计某次验收中的 open 问题数。
func (r *AcceptIssueRepo) CountOpenByAcceptRun(ctx context.Context, acceptRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("accept_run_id", acceptRunID).
		Where("status", "open").
		WhereNull("deleted_at").
		Count()
}

// ListByWorkflowAndTask 按工作流和任务查询验收问题列表。
func (r *AcceptIssueRepo) ListByWorkflowAndTask(ctx context.Context, workflowRunID, taskID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("domain_task_id", taskID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// CountGroupBySeverityByWorkflow 按严重级别统计工作流下的验收问题。
func (r *AcceptIssueRepo) CountGroupBySeverityByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("severity, COUNT(*) as cnt").
		Group("severity").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}
