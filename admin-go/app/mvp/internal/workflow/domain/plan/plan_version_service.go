// Package plan 计划版本领域服务。
package plan

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

// PlanVersionService 计划版本服务。
type PlanVersionService struct {
	planRepo      *repo.PlanVersionRepo
	blueprintRepo *repo.BlueprintRepo
}

// NewPlanVersionService 创建计划版本服务。
func NewPlanVersionService(pr *repo.PlanVersionRepo, br *repo.BlueprintRepo) *PlanVersionService {
	return &PlanVersionService{planRepo: pr, blueprintRepo: br}
}

// SupersedePreviousVersions 将项目之前的活跃版本标记为 superseded。
func (s *PlanVersionService) SupersedePreviousVersions(ctx context.Context, projectID int64, exceptVersionID int64) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[PlanVersionService] SupersedePreviousVersions projectID=%d except=%d", projectID, exceptVersionID)
	return nil
}

// SubmitForReview 提交计划版本进入审核。
func (s *PlanVersionService) SubmitForReview(ctx context.Context, planVersionID int64) error {
	// TODO: M2 实现
	g.Log().Infof(ctx, "[PlanVersionService] SubmitForReview planVersionID=%d", planVersionID)
	return nil
}

// Approve 通过计划版本。
func (s *PlanVersionService) Approve(ctx context.Context, planVersionID int64) error {
	// TODO: M3 实现
	g.Log().Infof(ctx, "[PlanVersionService] Approve planVersionID=%d", planVersionID)
	return nil
}

// Reject 驳回计划版本。
func (s *PlanVersionService) Reject(ctx context.Context, planVersionID int64) error {
	// TODO: M3 实现
	g.Log().Infof(ctx, "[PlanVersionService] Reject planVersionID=%d", planVersionID)
	return nil
}
