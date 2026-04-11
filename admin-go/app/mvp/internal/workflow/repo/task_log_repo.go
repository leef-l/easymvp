package repo

import (
	"context"

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
