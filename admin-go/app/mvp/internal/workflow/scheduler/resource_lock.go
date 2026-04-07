package scheduler

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// ResourceLockManager 资源锁管理。
type ResourceLockManager struct{}

// NewResourceLockManager 创建资源锁管理器。
func NewResourceLockManager() *ResourceLockManager { return &ResourceLockManager{} }

// AcquireLocks 为任务获取资源锁。
func (m *ResourceLockManager) AcquireLocks(ctx context.Context, workflowRunID, taskID int64, resources []string) error {
	for _, res := range resources {
		_, err := g.DB().Model("mvp_task_resource_lock").Ctx(ctx).Insert(g.Map{
			"id":              snowflake.Generate(),
			"workflow_run_id": workflowRunID,
			"task_id":         taskID,
			"resource_path":   res,
			"lock_status":     "held",
			"locked_at":       time.Now(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ReleaseLocks 释放任务的所有资源锁。
func (m *ResourceLockManager) ReleaseLocks(ctx context.Context, taskID int64) error {
	_, err := g.DB().Model("mvp_task_resource_lock").Ctx(ctx).
		Where("task_id", taskID).
		Where("lock_status", "held").
		Data(g.Map{"lock_status": "released", "released_at": time.Now()}).
		Update()
	return err
}
