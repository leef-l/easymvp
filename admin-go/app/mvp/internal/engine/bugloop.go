package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// BugLoop Bug 闭环处理
// 审计员发现 bug → bug_found → 架构师分析 → bug_dispatched → 实施员修复 → running

// ReportBug 审计员报告 bug
// auditorTaskID: 审计任务 ID
// bugDescription: bug 描述
func (s *Scheduler) ReportBug(ctx context.Context, projectID int64, auditorTaskID int64, bugDescription string) error {
	// 1. 更新审计任务状态
	_, err := g.DB().Model("mvp_task").Where("id", auditorTaskID).Update(g.Map{
		"status":     "bug_found",
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return err
	}

	logTaskAction(auditorTaskID, "bug_found", "auditing", "bug_found", bugDescription, "auditor")

	// 2. 找到审计任务依赖的实施员任务
	dep, err := g.DB().Model("mvp_task_dependency").
		Where("task_id", auditorTaskID).
		One()
	if err != nil || dep.IsEmpty() {
		return fmt.Errorf("找不到审计任务关联的实施员任务")
	}
	implTaskID := dep["depends_on_id"].Int64()

	// 3. 更新实施员任务状态为 bug_found
	_, err = g.DB().Model("mvp_task").Where("id", implTaskID).Update(g.Map{
		"status":        "bug_found",
		"error_message": bugDescription,
		"updated_at":    gtime.Now(),
	})
	if err != nil {
		return err
	}

	logTaskAction(implTaskID, "bug_found", "completed", "bug_found", "审计员发现bug: "+bugDescription, "auditor")

	// 4. 创建架构师分析任务
	go s.createBugAnalysisTask(ctx, projectID, implTaskID, auditorTaskID, bugDescription)

	return nil
}

// createBugAnalysisTask 创建架构师 bug 分析任务
func (s *Scheduler) createBugAnalysisTask(ctx context.Context, projectID int64, implTaskID int64, auditorTaskID int64, bugDescription string) {
	// 获取原实施任务信息
	implTask, err := g.DB().Model("mvp_task").Where("id", implTaskID).One()
	if err != nil || implTask.IsEmpty() {
		g.Log().Errorf(ctx, "查询实施任务失败: %v", err)
		return
	}

	// 创建架构师分析任务
	analysisTaskID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_task").Insert(g.Map{
		"id":          analysisTaskID,
		"project_id":  projectID,
		"parent_id":   implTask["parent_id"].Int64(),
		"name":        fmt.Sprintf("Bug分析: %s", implTask["name"].String()),
		"description": fmt.Sprintf("审计员在任务「%s」中发现以下问题，请分析原因并给出修复方案：\n\n%s\n\n原任务结果：\n%s", implTask["name"].String(), bugDescription, implTask["result"].String()),
		"role_type":   "architect",
		"status":      "pending",
		"batch_no":    0, // 高优先级，立即可调度
		"created_by":  0,
		"dept_id":     0,
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	})
	if err != nil {
		g.Log().Errorf(ctx, "创建架构师分析任务失败: %v", err)
		return
	}

	logTaskAction(analysisTaskID, "created", "", "pending", "系统创建Bug分析任务", "system")

	// 触发调度
	go s.scheduleOnce(context.Background(), projectID)
}

// DispatchBugFix 架构师分析完成后，分派修复任务给实施员
// analysisTaskID: 架构师分析任务 ID（已完成）
// implTaskID: 原实施员任务 ID
func (s *Scheduler) DispatchBugFix(ctx context.Context, projectID int64, analysisTaskID int64, implTaskID int64) error {
	// 1. 获取架构师分析结果
	analysisTask, err := g.DB().Model("mvp_task").Where("id", analysisTaskID).One()
	if err != nil || analysisTask.IsEmpty() {
		return fmt.Errorf("架构师分析任务不存在")
	}

	// 2. 获取原实施任务
	implTask, err := g.DB().Model("mvp_task").Where("id", implTaskID).One()
	if err != nil || implTask.IsEmpty() {
		return fmt.Errorf("原实施任务不存在")
	}

	// 3. 更新原实施任务状态为 bug_dispatched → 自动变为 pending（透传修复）
	_, err = g.DB().Model("mvp_task").Where("id", implTaskID).Update(g.Map{
		"status":      "pending",
		"description": fmt.Sprintf("%s\n\n## Bug修复指令\n%s", implTask["description"].String(), analysisTask["result"].String()),
		"result":      nil, // 清空旧结果
		"started_at":  nil,
		"completed_at": nil,
		"updated_at":  gtime.Now(),
	})
	if err != nil {
		return err
	}

	logTaskAction(implTaskID, "bug_dispatched", "bug_found", "pending", "架构师已分析，分派修复任务", "architect")

	// 4. 触发调度
	go s.scheduleOnce(context.Background(), projectID)

	return nil
}

// AutoDispatchBugFix 架构师分析任务自动完成后的回调
// 在 executor.go 中，当角色是 architect 且任务名以 "Bug分析:" 开头时调用
func (s *Scheduler) AutoDispatchBugFix(ctx context.Context, projectID int64, analysisTaskID int64) {
	// 从任务描述中找到原实施任务 ID
	// 通过分析任务的依赖关系或命名规则来关联
	analysisTask, err := g.DB().Model("mvp_task").Where("id", analysisTaskID).One()
	if err != nil || analysisTask.IsEmpty() {
		return
	}

	// 查找同父节点下状态为 bug_found 的实施员任务
	implTask, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("parent_id", analysisTask["parent_id"].Int64()).
		Where("role_type", "implementer").
		Where("status", "bug_found").
		Where("deleted_at IS NULL").
		One()
	if err != nil || implTask.IsEmpty() {
		return
	}

	s.DispatchBugFix(ctx, projectID, analysisTaskID, implTask["id"].Int64())
}
