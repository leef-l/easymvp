// Package review 审核问题领域服务。
package review

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/app/mvp/internal/workflow/repo"
)

// IssueService 审核问题服务。
type IssueService struct {
	issueRepo *repo.ReviewIssueRepo
}

// NewIssueService 创建审核问题服务。
func NewIssueService(ir *repo.ReviewIssueRepo) *IssueService {
	return &IssueService{issueRepo: ir}
}

// ListByPlanVersion 查询版本下的审核问题。
func (s *IssueService) ListByPlanVersion(ctx context.Context, planVersionID int64) ([]entity.MvpReviewIssue, error) {
	return s.issueRepo.ListByPlanVersion(ctx, planVersionID)
}

// ResolveIssue 解决审核问题。
func (s *IssueService) ResolveIssue(ctx context.Context, issueID int64) error {
	// TODO: M3 实现
	g.Log().Infof(ctx, "[IssueService] ResolveIssue issueID=%d", issueID)
	return nil
}
