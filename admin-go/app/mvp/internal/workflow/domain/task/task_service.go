// Package task 领域任务服务。
package task

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

// TaskService 领域任务服务。
type TaskService struct {
	taskRepo *repo.DomainTaskRepo
}

// NewTaskService 创建领域任务服务。
func NewTaskService(tr *repo.DomainTaskRepo) *TaskService {
	return &TaskService{taskRepo: tr}
}

// InstantiateFromBlueprint 将蓝图实例化为领域任务。
func (s *TaskService) InstantiateFromBlueprint(ctx context.Context, planVersionID int64, stageRunID int64, workflowRunID int64) (int, error) {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[TaskService] InstantiateFromBlueprint planVersionID=%d", planVersionID)
	return 0, nil
}

// Retry 重试失败任务。
func (s *TaskService) Retry(ctx context.Context, taskID int64) error {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[TaskService] Retry taskID=%d", taskID)
	return nil
}

// Skip 跳过任务。
func (s *TaskService) Skip(ctx context.Context, taskID int64) error {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[TaskService] Skip taskID=%d", taskID)
	return nil
}

// Escalate 升级任务。
func (s *TaskService) Escalate(ctx context.Context, taskID int64) error {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[TaskService] Escalate taskID=%d", taskID)
	return nil
}
