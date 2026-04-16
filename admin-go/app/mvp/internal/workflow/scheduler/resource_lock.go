package scheduler

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/repo"
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
	lockRepo := repo.NewTaskResourceLockRepo()
	return repo.WithTx(ctx, func(ctx context.Context, tx gdb.TX) error {
		existingResources, err := lockRepo.ListHeldByTaskResourcesInTx(ctx, tx, taskID, normalized)
		if err != nil {
			return err
		}

		existingSet := make(map[string]struct{}, len(existingResources))
		for _, resource := range existingResources {
			existingSet[resource] = struct{}{}
		}

		if len(existingResources) > 0 {
			if err := lockRepo.UpdateHeldInTx(ctx, tx, taskID, existingResources, g.Map{
				"workflow_run_id": workflowRunID,
				"lock_status":     "held",
				"locked_at":       now,
				"released_at":     nil,
			}); err != nil {
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
		return lockRepo.BatchCreateInTx(ctx, tx, batch)
	})
}

// ReleaseLocks 释放任务的所有资源锁。
func (m *ResourceLockManager) ReleaseLocks(ctx context.Context, taskID int64) error {
	return repo.NewTaskResourceLockRepo().ReleaseHeldByTask(ctx, taskID, gtime.NewFromTime(time.Now()))
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
