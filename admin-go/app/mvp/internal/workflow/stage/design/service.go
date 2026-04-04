// Package design 管理设计阶段：架构师对话产物 → plan_version → task_blueprint。
package design

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// Service 设计阶段服务。
type Service struct{}

// NewService 创建设计阶段服务。
func NewService() *Service { return &Service{} }

// CreatePlanFromArchitectReply 从架构师回复生成计划版本和任务蓝图。
func (s *Service) CreatePlanFromArchitectReply(ctx context.Context, workflowRunID int64, conversationID int64, messageID int64, content string) (int64, error) {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[DesignStage] CreatePlanFromArchitectReply workflowRunID=%d", workflowRunID)
	return 0, nil
}
