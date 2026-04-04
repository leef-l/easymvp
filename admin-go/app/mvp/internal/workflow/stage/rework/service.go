// Package rework 统一承接 bug 修复与失败升级的返工阶段。
package rework

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// Service 返工阶段服务。
type Service struct{}

// NewService 创建返工阶段服务。
func NewService() *Service { return &Service{} }

// HandleRework 处理返工流程。
func (s *Service) HandleRework(ctx context.Context, stageRunID int64) error {
	// TODO: M5 实现
	g.Log().Infof(ctx, "[ReworkStage] HandleRework stageRunID=%d", stageRunID)
	return nil
}
