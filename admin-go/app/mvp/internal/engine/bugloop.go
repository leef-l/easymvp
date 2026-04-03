package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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

// EscalateFailedTask 非 auditor 任务重试耗尽后，创建架构师分析任务
// 与 ReportBug 不同，此方法直接基于失败任务本身创建分析任务，无需 auditor→implementer 关联
func (s *Scheduler) EscalateFailedTask(ctx context.Context, projectID int64, failedTaskID int64, roleType string, errMsg string) {
	failedTask, err := g.DB().Model("mvp_task").Where("id", failedTaskID).One()
	if err != nil || failedTask.IsEmpty() {
		g.Log().Errorf(ctx, "[EscalateFailedTask] 查询失败任务 %d 出错: %v", failedTaskID, err)
		return
	}

	analysisTaskID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_task").Insert(g.Map{
		"id":         analysisTaskID,
		"project_id": projectID,
		"parent_id":  failedTask["parent_id"].Int64(),
		"name":       fmt.Sprintf("失败分析: %s", failedTask["name"].String()),
		"description": fmt.Sprintf("请分析任务失败原因，并给出可直接回写到原任务的修复方案。\n\n关联任务ID：%d\n角色：%s\n错误信息：\n%s\n\n原任务名称：%s\n原任务描述：\n%s\n\n请严格输出 JSON，格式如下：\n{\"description\":\"修订后的任务描述\",\"affected_resources\":[\"相对路径1\",\"相对路径2\"],\"reason\":\"修订原因\"}",
			failedTaskID, roleType, errMsg, failedTask["name"].String(), failedTask["description"].String()),
		"role_type":  "architect",
		"status":     "pending",
		"batch_no":   0, // 高优先级
		"created_by": 0,
		"dept_id":    0,
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})
	if err != nil {
		g.Log().Errorf(ctx, "[EscalateFailedTask] 创建架构师分析任务失败: %v", err)
		return
	}

	logTaskAction(analysisTaskID, "created", "", "pending", "系统创建失败分析任务（升级处理）", "system")
	go s.scheduleOnce(context.Background(), projectID)
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
		"status":       "pending",
		"description":  fmt.Sprintf("%s\n\n## Bug修复指令\n%s", implTask["description"].String(), analysisTask["result"].String()),
		"result":       nil, // 清空旧结果
		"started_at":   nil,
		"completed_at": nil,
		"updated_at":   gtime.Now(),
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

func (s *Scheduler) AutoDispatchFailureFix(ctx context.Context, projectID int64, analysisTaskID int64) {
	analysisTask, err := g.DB().Model("mvp_task").Where("id", analysisTaskID).One()
	if err != nil || analysisTask.IsEmpty() {
		return
	}

	implTaskID := parseLinkedTaskID(analysisTask["description"].String())
	if implTaskID == 0 {
		return
	}

	patch, err := parseArchitectTaskPatch(analysisTask["result"].String())
	if err != nil {
		g.Log().Warningf(ctx, "[AutoDispatchFailureFix] 解析架构师修复方案失败: task=%d err=%v", analysisTaskID, err)
		return
	}

	maxRounds := GetConfigInt(ctx, "failure_handoff.max_rounds", "engine.failureHandoff.maxRounds", 3)
	rounds, _ := g.DB().Model("mvp_task_log").
		Where("task_id", implTaskID).
		Where("action", "escalate_to_architect").
		Count()
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

	data := g.Map{
		"status":       "pending",
		"result":       nil,
		"started_at":   nil,
		"completed_at": nil,
		"updated_at":   gtime.Now(),
	}
	if strings.TrimSpace(patch.Description) != "" {
		data["description"] = patch.Description
	}
	if normalized, _ := normalizePatchResources(patch.AffectedResources); len(normalized) > 0 {
		resourceJSON, _ := json.Marshal(normalized)
		data["affected_resources"] = string(resourceJSON)
	}

	if _, err = g.DB().Model("mvp_task").Where("id", implTaskID).Data(data).Update(); err != nil {
		g.Log().Errorf(ctx, "[AutoDispatchFailureFix] 回写原任务失败: task=%d err=%v", implTaskID, err)
		return
	}

	logMessage := "架构师已给出修复方案，原任务重新进入 pending"
	if strings.TrimSpace(patch.Reason) != "" {
		logMessage += "；原因：" + strings.TrimSpace(patch.Reason)
	}
	logTaskAction(implTaskID, "architect_revised", "escalated", "pending", logMessage, "architect")
	go s.scheduleOnce(context.Background(), projectID)
}

func normalizePatchResources(values []string) ([]string, []string) {
	return parseResourcesDetail(func() string {
		raw, _ := json.Marshal(values)
		return string(raw)
	}()).Resources, nil
}

func parseLinkedTaskID(desc string) int64 {
	re := regexp.MustCompile(`关联任务ID：(\d+)`)
	match := re.FindStringSubmatch(desc)
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

	re := regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")
	match := re.FindStringSubmatch(content)
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
	conv, err := g.DB().Model("mvp_conversation").
		Where("project_id", projectID).
		Where("role_type", "architect").
		Where("task_id IS NULL OR task_id = 0").
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return 0, 0, 0, err
	}
	if !conv.IsEmpty() {
		return conv["id"].Int64(), conv["created_by"].Int64(), conv["dept_id"].Int64(), nil
	}

	project, err := g.DB().Model("mvp_project").
		Fields("created_by, dept_id").
		Where("id", projectID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return 0, 0, 0, err
	}
	if project.IsEmpty() {
		return 0, 0, 0, fmt.Errorf("项目不存在")
	}

	convID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_conversation").Insert(g.Map{
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
