package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/utility/snowflake"
)

// BugLoop Bug 闭环处理
// 审计员发现 bug → bug_found → 架构师分析 → bug_dispatched → 实施员修复 → running

// ReportBug 审计员报告 bug
// auditorTaskID: 审计任务 ID
// bugDescription: bug 描述
func (s *Scheduler) ReportBug(ctx context.Context, projectID int64, auditorTaskID int64, bugDescription string) error {
	// 1. 更新审计任务状态
	rows, err := updateTaskStatus(ctx, auditorTaskID, "auditing", "bug_found", nil)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("审计任务状态已变更，无法标记 bug_found")
	}

	logTaskAction(auditorTaskID, "bug_found", "auditing", "bug_found", bugDescription, "auditor")

	// 2. 找到审计任务依赖的实施员任务
	dep, err := g.DB().Ctx(ctx).Model("mvp_task_dependency").
		Where("task_id", auditorTaskID).
		One()
	if err != nil || dep.IsEmpty() {
		return fmt.Errorf("找不到审计任务关联的实施员任务")
	}
	implTaskID := dep["depends_on_id"].Int64()

	// 3. 更新实施员任务状态为 bug_found
	_, err = updateTaskStatus(ctx, implTaskID, "completed", "bug_found", g.Map{
		"error_message": bugDescription,
	})
	if err != nil {
		return err
	}

	logTaskAction(implTaskID, "bug_found", "completed", "bug_found", "审计员发现bug: "+bugDescription, "auditor")

	// 4. 创建架构师分析任务（使用项目级 context）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[BugLoop] createBugAnalysisTask panic: project=%d implTask=%d err=%v", projectID, implTaskID, r)
			}
		}()
		s.createBugAnalysisTask(s.getProjectContext(projectID), projectID, implTaskID, auditorTaskID, bugDescription)
	}()

	return nil
}

// createBugAnalysisTask 创建架构师 bug 分析任务
func (s *Scheduler) createBugAnalysisTask(ctx context.Context, projectID int64, implTaskID int64, auditorTaskID int64, bugDescription string) {
	// 暂停保护：项目已取消则不再创建派生任务
	select {
	case <-ctx.Done():
		return
	default:
	}

	// 获取原实施任务信息
	implTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", implTaskID).WhereNull("deleted_at").One()
	if err != nil || implTask.IsEmpty() {
		g.Log().Errorf(ctx, "查询实施任务失败: %v", err)
		return
	}

	// 链路字段：root_task_id 继承自实施任务，兼容旧数据
	rootTaskID := implTask["root_task_id"].Int64()
	if rootTaskID == 0 {
		rootTaskID = implTaskID
	}

	// 创建架构师分析任务
	analysisTaskID := int64(snowflake.Generate())
	_, err = g.DB().Ctx(ctx).Model("mvp_task").Insert(g.Map{
		"id":             analysisTaskID,
		"project_id":     projectID,
		"parent_id":      implTask["parent_id"].Int64(),
		"name":           fmt.Sprintf("Bug分析: %s", implTask["name"].String()),
		"description":    fmt.Sprintf("审计员在任务「%s」中发现以下问题，请分析原因并给出修复方案：\n\n%s\n\n原任务结果：\n%s", implTask["name"].String(), bugDescription, implTask["result"].String()),
		"role_type":      "architect",
		"task_kind":      consts.TaskKindBugAnalysis,
		"source_task_id": auditorTaskID,
		"root_task_id":   rootTaskID,
		"status":         "pending",
		"batch_no":       0, // 高优先级，立即可调度
		"created_by":     0,
		"dept_id":        0,
		"created_at":     gtime.Now(),
		"updated_at":     gtime.Now(),
	})
	if err != nil {
		g.Log().Errorf(ctx, "创建架构师分析任务失败: %v", err)
		return
	}

	logTaskAction(analysisTaskID, "created", "", "pending", "系统创建Bug分析任务", "system")

	// 触发调度（使用项目级 context）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] scheduleOnce panic: project=%d err=%v", projectID, r)
			}
		}()
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
	}()
}

// EscalateFailedTask 非 auditor 任务重试耗尽后，创建架构师分析任务
// 与 ReportBug 不同，此方法直接基于失败任务本身创建分析任务，无需 auditor→implementer 关联
func (s *Scheduler) EscalateFailedTask(ctx context.Context, projectID int64, failedTaskID int64, roleType string, errMsg string) {
	// 暂停保护
	select {
	case <-ctx.Done():
		return
	default:
	}

	failedTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", failedTaskID).WhereNull("deleted_at").One()
	if err != nil || failedTask.IsEmpty() {
		g.Log().Errorf(ctx, "[EscalateFailedTask] 查询失败任务 %d 出错: %v", failedTaskID, err)
		return
	}

	// 链路字段：root_task_id 继承自失败任务，兼容旧数据
	rootTaskID := failedTask["root_task_id"].Int64()
	if rootTaskID == 0 {
		rootTaskID = failedTaskID
	}

	analysisTaskID := int64(snowflake.Generate())
	_, err = g.DB().Ctx(ctx).Model("mvp_task").Insert(g.Map{
		"id":         analysisTaskID,
		"project_id": projectID,
		"parent_id":  failedTask["parent_id"].Int64(),
		"name":       fmt.Sprintf("失败分析: %s", failedTask["name"].String()),
		"description": fmt.Sprintf("请分析任务失败原因，并给出可直接回写到原任务的修复方案。\n\n关联任务ID：%d\n角色：%s\n错误信息：\n%s\n\n原任务名称：%s\n原任务描述：\n%s\n\n请严格输出 JSON，格式如下：\n{\"description\":\"修订后的任务描述\",\"affected_resources\":[\"相对路径1\",\"相对路径2\"],\"reason\":\"修订原因\"}",
			failedTaskID, roleType, errMsg, failedTask["name"].String(), failedTask["description"].String()),
		"role_type":      "architect",
		"task_kind":      consts.TaskKindFailureAnalysis,
		"source_task_id": failedTaskID,
		"root_task_id":   rootTaskID,
		"status":         "pending",
		"batch_no":       0, // 高优先级
		"created_by":     0,
		"dept_id":        0,
		"created_at":     gtime.Now(),
		"updated_at":     gtime.Now(),
	})
	if err != nil {
		g.Log().Errorf(ctx, "[EscalateFailedTask] 创建架构师分析任务失败: %v", err)
		return
	}

	logTaskAction(analysisTaskID, "created", "", "pending", "系统创建失败分析任务（升级处理）", "system")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] scheduleOnce panic: project=%d err=%v", projectID, r)
			}
		}()
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
	}()
}

type architectTaskPatch struct {
	Description       string   `json:"description"`
	AffectedResources []string `json:"affected_resources"`
	Reason            string   `json:"reason"`
}

// DispatchBugFix 架构师分析完成后，分派修复任务给实施员
// analysisTaskID: 架构师分析任务 ID（已完成）
// implTaskID: 原实施员任务 ID
func (s *Scheduler) DispatchBugFix(ctx context.Context, projectID int64, analysisTaskID int64, implTaskID int64) error {
	// 暂停保护
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 1. 获取架构师分析结果
	analysisTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", analysisTaskID).WhereNull("deleted_at").One()
	if err != nil || analysisTask.IsEmpty() {
		return fmt.Errorf("架构师分析任务不存在")
	}

	// 2. 获取原实施任务
	implTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", implTaskID).WhereNull("deleted_at").One()
	if err != nil || implTask.IsEmpty() {
		return fmt.Errorf("原实施任务不存在")
	}

	// 3. 先标记 bug_dispatched，再转 pending（两步状态机合规转换）
	_, err = updateTaskStatus(ctx, implTaskID, "bug_found", "bug_dispatched", nil)
	if err != nil {
		return err
	}
	_, err = updateTaskStatus(ctx, implTaskID, "bug_dispatched", "pending", g.Map{
		"description":  fmt.Sprintf("%s\n\n## Bug修复指令\n%s", implTask["description"].String(), analysisTask["result"].String()),
		"result":       nil, // 清空旧结果
		"started_at":   nil,
		"completed_at": nil,
	})
	if err != nil {
		return err
	}

	logTaskAction(implTaskID, "bug_dispatched", "bug_found", "pending", "架构师已分析，分派修复任务", "architect")

	// 4. 触发调度（使用项目级 context）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] scheduleOnce panic: project=%d err=%v", projectID, r)
			}
		}()
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
	}()

	return nil
}

// AutoDispatchBugFix 架构师分析任务自动完成后的回调
// 两跳回溯：bug_analysis.source_task_id → audit.source_task_id → implement.id
func (s *Scheduler) AutoDispatchBugFix(ctx context.Context, projectID int64, analysisTaskID int64) {
	// 暂停保护
	select {
	case <-ctx.Done():
		return
	default:
	}

	analysisTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", analysisTaskID).WhereNull("deleted_at").One()
	if err != nil || analysisTask.IsEmpty() {
		return
	}

	// 第一跳：从 bug_analysis 找到审计任务
	auditTaskID := analysisTask["source_task_id"].Int64()
	if auditTaskID == 0 {
		// 旧数据 fallback：查找同父节点下 bug_found 的实施任务
		implTask, err := g.DB().Ctx(ctx).Model("mvp_task").
			Where("project_id", projectID).
			Where("parent_id", analysisTask["parent_id"].Int64()).
			Where("role_type", "implementer").
			Where("status", "bug_found").
			WhereNull("deleted_at").
			One()
		if err != nil || implTask.IsEmpty() {
			return
		}
		s.DispatchBugFix(ctx, projectID, analysisTaskID, implTask["id"].Int64())
		return
	}

	// 第二跳：从审计任务找到原实施任务
	auditTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", auditTaskID).WhereNull("deleted_at").One()
	if err != nil || auditTask.IsEmpty() {
		return
	}

	implTaskID := auditTask["source_task_id"].Int64()
	if implTaskID == 0 {
		// 旧数据 fallback：从 mvp_task_dependency 查找
		dep, depErr := g.DB().Ctx(ctx).Model("mvp_task_dependency").
			Where("task_id", auditTaskID).
			One()
		if depErr != nil {
			g.Log().Warningf(ctx, "[AutoDispatchBugFix] 查询依赖关系失败: taskID=%d err=%v", auditTaskID, depErr)
		}
		if !dep.IsEmpty() {
			implTaskID = dep["depends_on_id"].Int64()
		}
	}

	if implTaskID == 0 {
		g.Log().Warningf(ctx, "[AutoDispatchBugFix] 无法回溯到实施任务: analysis=%d audit=%d", analysisTaskID, auditTaskID)
		return
	}

	s.DispatchBugFix(ctx, projectID, analysisTaskID, implTaskID)
}

func (s *Scheduler) AutoDispatchFailureFix(ctx context.Context, projectID int64, analysisTaskID int64) {
	// 暂停保护
	select {
	case <-ctx.Done():
		return
	default:
	}

	analysisTask, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", analysisTaskID).WhereNull("deleted_at").One()
	if err != nil || analysisTask.IsEmpty() {
		return
	}

	// 优先走显式链路字段，旧数据 fallback 到描述正则
	implTaskID := analysisTask["source_task_id"].Int64()
	if implTaskID == 0 {
		implTaskID = parseLinkedTaskID(analysisTask["description"].String())
	}
	if implTaskID == 0 {
		return
	}

	patch, err := parseArchitectTaskPatch(analysisTask["result"].String())
	if err != nil {
		g.Log().Warningf(ctx, "[AutoDispatchFailureFix] 解析架构师修复方案失败: task=%d err=%v", analysisTaskID, err)
		return
	}

	maxRounds := GetConfigInt(ctx, "failure_handoff.max_rounds", "engine.failureHandoff.maxRounds", 3)
	rounds, rndErr := g.DB().Ctx(ctx).Model("mvp_task_log").
		Where("task_id", implTaskID).
		Where("action", "escalate_to_architect").
		Count()
	if rndErr != nil {
		g.Log().Warningf(ctx, "[AutoDispatchFailureFix] 查询轮次失败: taskID=%d err=%v", implTaskID, rndErr)
	}
	if rounds >= maxRounds {
		pauseReason := fmt.Sprintf("任务 %d 多次在角色协作后仍无法稳定修复，已达到托底上限 %d 次，请人工介入。", implTaskID, maxRounds)
		if err = s.Pause(ctx, projectID, pauseReason); err != nil {
			g.Log().Errorf(ctx, "[AutoDispatchFailureFix] 暂停项目失败: project=%d err=%v", projectID, err)
		}
		notifyProjectArchitectConversation(ctx, projectID, fmt.Sprintf(
			"任务 %d 在 implementer ↔ architect 协作修复中已达到托底上限 %d 次，项目已自动暂停。\n\n请重新审视任务拆分、依赖关系、affected_resources 与修复方案，并决定是重拆任务还是人工介入。\n\n最近一次架构师修复建议：\n%s",
			implTaskID, maxRounds, analysisTask["result"].String(),
		))
		logTaskAction(implTaskID, "handoff_limit_reached", "escalated", "escalated", pauseReason, "system")
		return
	}

	extra := g.Map{
		"result":       nil,
		"started_at":   nil,
		"completed_at": nil,
	}
	if strings.TrimSpace(patch.Description) != "" {
		extra["description"] = patch.Description
	}
	if normalized, _ := normalizePatchResources(patch.AffectedResources); len(normalized) > 0 {
		resourceJSON, _ := json.Marshal(normalized)
		extra["affected_resources"] = string(resourceJSON)
	}

	if _, err = updateTaskStatus(ctx, implTaskID, "escalated", "pending", extra); err != nil {
		g.Log().Errorf(ctx, "[AutoDispatchFailureFix] 回写原任务失败: task=%d err=%v", implTaskID, err)
		return
	}

	logMessage := "架构师已给出修复方案，原任务重新进入 pending"
	if strings.TrimSpace(patch.Reason) != "" {
		logMessage += "；原因：" + strings.TrimSpace(patch.Reason)
	}
	logTaskAction(implTaskID, "architect_revised", "escalated", "pending", logMessage, "architect")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] scheduleOnce panic: project=%d err=%v", projectID, r)
			}
		}()
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
	}()
}

func normalizePatchResources(values []string) ([]string, []string) {
	return parseResourcesDetail(func() string {
		raw, _ := json.Marshal(values)
		return string(raw)
	}()).Resources, nil
}

var linkedTaskIDRe = regexp.MustCompile(`关联任务ID：(\d+)`)

func parseLinkedTaskID(desc string) int64 {
	match := linkedTaskIDRe.FindStringSubmatch(desc)
	if len(match) != 2 {
		return 0
	}
	var taskID int64
	fmt.Sscanf(match[1], "%d", &taskID)
	return taskID
}

func parseArchitectTaskPatch(content string) (*architectTaskPatch, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("架构师输出为空")
	}

	var patch architectTaskPatch
	if err := json.Unmarshal([]byte(content), &patch); err == nil {
		return &patch, nil
	}

	match := jsonCodeBlockRe.FindStringSubmatch(content)
	if len(match) == 2 {
		if err := json.Unmarshal([]byte(match[1]), &patch); err == nil {
			return &patch, nil
		}
	}
	return nil, fmt.Errorf("未解析到有效 JSON patch")
}

func notifyProjectArchitectConversation(ctx context.Context, projectID int64, content string) {
	conversationID, userID, deptID, err := ensureProjectArchitectConversation(ctx, projectID)
	if err != nil {
		g.Log().Errorf(ctx, "[notifyProjectArchitectConversation] 获取架构师对话失败: project=%d err=%v", projectID, err)
		return
	}
	if _, _, err = GetEngine().SendMessage(ctx, conversationID, content, userID, deptID); err != nil {
		g.Log().Errorf(ctx, "[notifyProjectArchitectConversation] 发送架构师对话失败: project=%d conversation=%d err=%v", projectID, conversationID, err)
	}
}

func ensureProjectArchitectConversation(ctx context.Context, projectID int64) (int64, int64, int64, error) {
	conv, err := g.DB().Ctx(ctx).Model("mvp_conversation").
		Where("project_id", projectID).
		Where("role_type", "architect").
		Where("task_id IS NULL OR task_id = 0").
		WhereNull("deleted_at").
		One()
	if err != nil {
		return 0, 0, 0, err
	}
	if !conv.IsEmpty() {
		return conv["id"].Int64(), conv["created_by"].Int64(), conv["dept_id"].Int64(), nil
	}

	project, err := g.DB().Ctx(ctx).Model("mvp_project").
		Fields("created_by, dept_id").
		Where("id", projectID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		return 0, 0, 0, err
	}
	if project.IsEmpty() {
		return 0, 0, 0, fmt.Errorf("项目不存在")
	}

	convID := int64(snowflake.Generate())
	_, err = g.DB().Ctx(ctx).Model("mvp_conversation").Insert(g.Map{
		"id":         convID,
		"project_id": projectID,
		"title":      "架构师对话",
		"role_type":  "architect",
		"status":     "active",
		"created_by": project["created_by"].Int64(),
		"dept_id":    project["dept_id"].Int64(),
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return 0, 0, 0, err
	}
	return convID, project["created_by"].Int64(), project["dept_id"].Int64(), nil
}
