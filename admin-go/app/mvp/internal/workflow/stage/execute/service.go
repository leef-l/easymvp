// Package execute 管理执行阶段：蓝图实例化 → domain_task → 调度执行。
package execute

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// Service 执行阶段服务。
type Service struct{}

// NewService 创建执行阶段服务。
func NewService() *Service { return &Service{} }

// InstantiateAndStart 将审核通过的蓝图实例化为领域任务并启动调度。
func (s *Service) InstantiateAndStart(ctx context.Context, stageRunID int64, planVersionID int64) error {
	// TODO: M4 实现
	g.Log().Infof(ctx, "[ExecuteStage] InstantiateAndStart stageRunID=%d planVersionID=%d", stageRunID, planVersionID)
	return nil
}
