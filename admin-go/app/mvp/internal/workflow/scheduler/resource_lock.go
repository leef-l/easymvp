package scheduler

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// ResourceLockManager 资源锁管理。
type ResourceLockManager struct{}

// NewResourceLockManager 创建资源锁管理器。
func NewResourceLockManager() *ResourceLockManager { return &ResourceLockManager{} }

// AcquireLocks 为任务获取资源锁。
func (m *ResourceLockManager) AcquireLocks(ctx context.Context, workflowRunID, taskID int64, resources []string) error {
	normalized := normalizeResourcePaths(resources)
	if len(normalized) == 0 {
		return nil
	}
	now := time.Now()
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		existingRecords, err := tx.Model("mvp_task_resource_lock").Ctx(ctx).
			Where("task_id", taskID).
			WhereIn("resource_path", normalized).
			Fields("resource_path").
			All()
		if err != nil {
			return err
		}

		existingSet := make(map[string]struct{}, len(existingRecords))
		existingResources := make([]string, 0, len(existingRecords))
		for _, record := range existingRecords {
			res := strings.TrimSpace(record["resource_path"].String())
			if res == "" {
				continue
			}
			if _, ok := existingSet[res]; ok {
				continue
			}
			existingSet[res] = struct{}{}
			existingResources = append(existingResources, res)
		}

		if len(existingResources) > 0 {
			if _, err := tx.Model("mvp_task_resource_lock").Ctx(ctx).
				Where("task_id", taskID).
				WhereIn("resource_path", existingResources).
				Data(g.Map{
					"workflow_run_id": workflowRunID,
					"lock_status":     "held",
					"locked_at":       now,
					"released_at":     nil,
				}).
				Update(); err != nil {
				return err
			}
		}

		batch := make([]g.Map, 0, len(normalized)-len(existingResources))
		for _, res := range normalized {
			if _, ok := existingSet[res]; ok {
				continue
			}
			batch = append(batch, g.Map{
				"id":              snowflake.Generate(),
				"workflow_run_id": workflowRunID,
				"task_id":         taskID,
				"resource_path":   res,
				"lock_status":     "held",
				"locked_at":       now,
			})
		}
		if len(batch) == 0 {
			return nil
		}
		_, err = tx.Model("mvp_task_resource_lock").Ctx(ctx).Insert(batch)
		return err
	})
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

func normalizeResourcePaths(resources []string) []string {
	if len(resources) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(resources))
	result := make([]string, 0, len(resources))
	for _, res := range resources {
		res = strings.TrimSpace(res)
		if res == "" {
			continue
		}
		if _, ok := seen[res]; ok {
			continue
		}
		seen[res] = struct{}{}
		result = append(result, res)
	}
	return result
}
