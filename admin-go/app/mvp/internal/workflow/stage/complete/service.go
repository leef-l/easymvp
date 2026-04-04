// Package complete 管理完成阶段：总结、归档、指标计算。
package complete

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// Service 完成阶段服务。
type Service struct{}

// NewService 创建完成阶段服务。
func NewService() *Service { return &Service{} }

// Finalize 执行项目完成流程。
func (s *Service) Finalize(ctx context.Context, stageRunID int64) error {
	// TODO: M5 实现
	g.Log().Infof(ctx, "[CompleteStage] Finalize stageRunID=%d", stageRunID)
	return nil
}
