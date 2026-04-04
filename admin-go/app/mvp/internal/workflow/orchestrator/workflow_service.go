// Package orchestrator 驱动工作流阶段切换，是新内核的总协调器。
// 职责：创建/启动/暂停/恢复/取消工作流，驱动阶段前进。
package orchestrator

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/runtime"
)

// WorkflowService 工作流编排服务。
type WorkflowService struct {
	runtimeMgr *runtime.Manager
	publisher  *event.Publisher
}

// NewWorkflowService 创建工作流服务。
func NewWorkflowService(rtMgr *runtime.Manager, pub *event.Publisher) *WorkflowService {
	return &WorkflowService{
		runtimeMgr: rtMgr,
		publisher:  pub,
	}
}

// CreateRun 为项目创建新的工作流运行实例。
func (s *WorkflowService) CreateRun(ctx context.Context, projectID int64) (int64, error) {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[WorkflowService] CreateRun projectID=%d", projectID)
	return 0, nil
}

// StartDesign 启动设计阶段。
func (s *WorkflowService) StartDesign(ctx context.Context, workflowRunID int64) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[WorkflowService] StartDesign workflowRunID=%d", workflowRunID)
	return nil
}

// Pause 暂停工作流。
func (s *WorkflowService) Pause(ctx context.Context, workflowRunID int64, reason string) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[WorkflowService] Pause workflowRunID=%d reason=%s", workflowRunID, reason)
	return nil
}

// Resume 恢复工作流。
func (s *WorkflowService) Resume(ctx context.Context, workflowRunID int64) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[WorkflowService] Resume workflowRunID=%d", workflowRunID)
	return nil
}

// Cancel 取消工作流。
func (s *WorkflowService) Cancel(ctx context.Context, workflowRunID int64, reason string) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[WorkflowService] Cancel workflowRunID=%d reason=%s", workflowRunID, reason)
	return nil
}
