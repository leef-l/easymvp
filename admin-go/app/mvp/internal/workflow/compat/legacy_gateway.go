// Package compat 新旧架构兼容层。
// 根据 project.engine_version 决定路由到旧 engine 还是新 workflow。
package compat

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// LegacyGateway 新旧引擎分流网关。
type LegacyGateway struct{}

// NewLegacyGateway 创建兼容网关。
func NewLegacyGateway() *LegacyGateway { return &LegacyGateway{} }

// IsWorkflowV2 判断项目是否使用新工作流引擎。
func (gw *LegacyGateway) IsWorkflowV2(ctx context.Context, projectID int64) (bool, error) {
	val, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Value("engine_version")
	if err != nil {
		return false, err
	}
	return val.String() == "workflow_v2", nil
}
