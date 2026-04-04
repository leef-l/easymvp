package review

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// DecisionService 审核结论服务。
type DecisionService struct{}

// NewDecisionService 创建审核结论服务。
func NewDecisionService() *DecisionService { return &DecisionService{} }

// MakeDecision 根据审核问题汇总得出结论。
func (s *DecisionService) MakeDecision(ctx context.Context, stageRunID int64) (string, error) {
	// TODO: M3 实现 — 返回 "approved"/"rejected"/"needs_revision"
	g.Log().Infof(ctx, "[DecisionService] MakeDecision stageRunID=%d", stageRunID)
	return "", nil
}
