// Package review 驱动审核阶段：precheck → auditor → coordinator → summary。
package review

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// Service 审核阶段服务。
type Service struct{}

// NewService 创建审核阶段服务。
func NewService() *Service { return &Service{} }

// RunReview 执行完整审核流程。
func (s *Service) RunReview(ctx context.Context, stageRunID int64, planVersionID int64) error {
	// TODO: M3 实现
	g.Log().Infof(ctx, "[ReviewStage] RunReview stageRunID=%d planVersionID=%d", stageRunID, planVersionID)
	return nil
}
