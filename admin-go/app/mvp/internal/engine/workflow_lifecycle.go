package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/lifecycle"
)

// Workflow 项目流程编排
// designing → confirmed → running → paused/completed/failed

// ConfirmPlan 确认实施方案，进入方案审核阶段
// 流程：designing → reviewing → (审核通过) → running
//
//	reviewing → (审核不通过) → designing
func (s *Scheduler) ConfirmPlan(ctx context.Context, projectID int64) error {
	// 1. 检查项目状态
	project, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).WhereNull("deleted_at").One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("项目不存在")
	}

	status := project["status"].String()
	if status != "designing" && status != "paused" {
		return fmt.Errorf("当前状态(%s)不允许确认方案", status)
	}

	// 2. 检查是否有 draft 任务
	draftCount := GetParser().GetDraftCount(ctx, projectID)
	if draftCount == 0 {
		return fmt.Errorf("没有待确认的任务，请先让架构师拆分任务")
	}

	// 3. 更新项目状态为 reviewing
	_, err = g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":       "reviewing",
		"pause_reason": nil,
		"updated_at":   gtime.Now(),
	})
	if err != nil {
		return err
	}

	// 4. 异步执行审核流程（不阻塞 API 响应）
	// 使用独立后台 context，避免请求返回后 ctx 被回收导致审核链中断
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Workflow] runReviewAsync panic: project=%d err=%v", projectID, r)
			}
		}()
		reviewCtx, reviewCancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer reviewCancel()
		s.runReviewAsync(reviewCtx, projectID)
	}()

	return nil
}

// runReviewAsync 异步执行方案审核流程
func (s *Scheduler) runReviewAsync(ctx context.Context, projectID int64) {
	g.Log().Infof(ctx, "[Workflow] 项目 %d 进入方案审核阶段", projectID)

	result, err := RunReview(ctx, projectID)
	if err != nil {
		g.Log().Errorf(ctx, "[Workflow] 方案审核执行失败: project=%d, err=%v", projectID, err)
		// 审核执行失败，退回 designing，保留完整错误链
		HandleReviewFailure(ctx, projectID, &ReviewResult{
			Errors: []ReviewIssue{{
				Severity: "error",
				Message:  fmt.Sprintf("审核流程执行异常: %s", FormatErrorChain(err)),
			}},
		})
		return
	}

	if result.Passed {
		g.Log().Infof(ctx, "[Workflow] 项目 %d 方案审核通过，进入执行阶段", projectID)
		if err := HandleReviewSuccess(ctx, projectID, result); err != nil {
			g.Log().Errorf(ctx, "[Workflow] 审核通过后启动失败: project=%d, err=%v", projectID, err)
		}
	} else {
		g.Log().Infof(ctx, "[Workflow] 项目 %d 方案审核未通过（%d errors），退回设计阶段",
			projectID, len(result.Errors))
		HandleReviewFailure(ctx, projectID, result)
	}
}

// Pause 暂停项目（回到设计阶段，可以和架构师沟通），记录暂停原因
func (s *Scheduler) Pause(ctx context.Context, projectID int64, reason string) error {
	_, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":       "paused",
		"pause_reason": reason,
		"updated_at":   gtime.Now(),
	})
	if err != nil {
		return err
	}

	s.PauseProject(projectID)
	return nil
}

// Resume 恢复项目（重新开始调度）
// 如果已有 pending 任务，直接进入 running（跳过审核）
// 如果只有 draft 任务，走审核流程
func (s *Scheduler) Resume(ctx context.Context, projectID int64) error {
	project, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).WhereNull("deleted_at").One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("项目不存在")
	}
	if project["status"].String() != "paused" {
		return fmt.Errorf("当前状态(%s)不允许恢复", project["status"].String())
	}

	// 检查是否有 pending 任务（说明之前已经审核过，直接恢复执行）
	pendingCount, pcErr := g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		WhereNull("deleted_at").
		Count()
	if pcErr != nil {
		return fmt.Errorf("查询 pending 任务数失败: %w", pcErr)
	}
	if pendingCount > 0 {
		// 直接恢复执行
		_, err = g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).Update(g.Map{
			"status":       "running",
			"pause_reason": nil,
			"updated_at":   gtime.Now(),
		})
		if err != nil {
			return err
		}
		s.StartProject(projectID)
		return nil
	}

	// 只有 draft 任务，走审核流程
	return s.ConfirmPlan(ctx, projectID)
}

// CreateProject 转发到 workflow/lifecycle 包。
// Deprecated: 直接调用 lifecycle.CreateProject。
func CreateProject(ctx context.Context, name, projectCategory, description, workDir string, architectModelID int64, userID int64, deptID int64, selectedPresetIDs []int64, engineVersion ...string) (int64, int64, error) {
	return lifecycle.CreateProject(ctx, name, projectCategory, description, workDir, architectModelID, userID, deptID, selectedPresetIDs, engineVersion...)
}
