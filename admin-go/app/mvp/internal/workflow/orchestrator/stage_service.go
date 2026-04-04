package orchestrator

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// StageService 阶段编排服务。
type StageService struct {
	workflowSvc *WorkflowService
}

// NewStageService 创建阶段服务。
func NewStageService(wfSvc *WorkflowService) *StageService {
	return &StageService{workflowSvc: wfSvc}
}

// StartStage 启动指定类型的阶段。
func (s *StageService) StartStage(ctx context.Context, workflowRunID int64, stageType string) (int64, error) {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[StageService] StartStage workflowRunID=%d stageType=%s", workflowRunID, stageType)
	return 0, nil
}

// CompleteStage 完成阶段。
func (s *StageService) CompleteStage(ctx context.Context, stageRunID int64) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[StageService] CompleteStage stageRunID=%d", stageRunID)
	return nil
}

// FailStage 标记阶段失败。
func (s *StageService) FailStage(ctx context.Context, stageRunID int64, reason string) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[StageService] FailStage stageRunID=%d reason=%s", stageRunID, reason)
	return nil
}

// TransitionNext 推进到下一阶段。
func (s *StageService) TransitionNext(ctx context.Context, workflowRunID int64) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[StageService] TransitionNext workflowRunID=%d", workflowRunID)
	return nil
}
