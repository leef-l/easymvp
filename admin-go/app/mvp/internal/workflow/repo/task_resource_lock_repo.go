package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskResourceLockRepo 任务资源锁仓储。
type TaskResourceLockRepo struct{}

func NewTaskResourceLockRepo() *TaskResourceLockRepo { return &TaskResourceLockRepo{} }

func (r *TaskResourceLockRepo) table() string { return "mvp_task_resource_lock" }

// ReleaseHeldByTask 释放指定任务仍处于 held 的资源锁。
func (r *TaskResourceLockRepo) ReleaseHeldByTask(ctx context.Context, taskID int64, releasedAt *gtime.Time) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		Where("lock_status", "held").
		Data(g.Map{
			"lock_status": "released",
			"released_at": releasedAt,
		}).
		Update()
	return err
}
