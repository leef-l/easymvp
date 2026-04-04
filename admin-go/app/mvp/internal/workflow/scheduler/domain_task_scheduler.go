package scheduler

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// DomainTaskScheduler 领域任务调度器。
type DomainTaskScheduler struct{}

// NewDomainTaskScheduler 创建领域任务调度器。
func NewDomainTaskScheduler() *DomainTaskScheduler { return &DomainTaskScheduler{} }

// Start 启动调度循环。
func (s *DomainTaskScheduler) Start(ctx context.Context, workflowRunID int64) error {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[DomainTaskScheduler] Start workflowRunID=%d", workflowRunID)
	return nil
}

// OnTaskCompleted 任务完成回调。
func (s *DomainTaskScheduler) OnTaskCompleted(ctx context.Context, taskID int64) error {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[DomainTaskScheduler] OnTaskCompleted taskID=%d", taskID)
	return nil
}
