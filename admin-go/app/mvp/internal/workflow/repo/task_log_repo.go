package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// TaskLogRepo 任务日志仓储。
type TaskLogRepo struct{}

func NewTaskLogRepo() *TaskLogRepo { return &TaskLogRepo{} }

func (r *TaskLogRepo) table() string { return "mvp_task_log" }

// ListByTask 查询任务日志。
func (r *TaskLogRepo) ListByTask(ctx context.Context, taskID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		OrderAsc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListRecentByWorkflow 查询工作流下最近任务日志，联表附带 task_id。
func (r *TaskLogRepo) ListRecentByWorkflow(ctx context.Context, workflowRunID int64, limit int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()+" tl").Ctx(ctx).
		InnerJoin("mvp_domain_task dt", "dt.id = tl.task_id").
		Where("dt.workflow_run_id", workflowRunID).
		WhereNull("tl.deleted_at").
		WhereNull("dt.deleted_at").
		OrderDesc("tl.created_at")
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
