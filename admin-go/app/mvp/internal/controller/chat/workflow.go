package chat

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/utility/snowflake"
)

var Workflow = cWorkflow{}

type cWorkflow struct{}

// CreateProject 创建项目
func (c *cWorkflow) CreateProject(ctx context.Context, req *v1.WorkflowCreateProjectReq) (res *v1.WorkflowCreateProjectRes, err error) {
	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	projectID, convID, err := engine.CreateProject(ctx, req.Name, req.Description, int64(req.ArchitectModelID), userID, deptID)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowCreateProjectRes{
		ProjectID:      snowflake.JsonInt64(projectID),
		ConversationID: snowflake.JsonInt64(convID),
	}, nil
}

// ConfirmPlan 确认实施方案
func (c *cWorkflow) ConfirmPlan(ctx context.Context, req *v1.WorkflowConfirmPlanReq) (res *v1.WorkflowConfirmPlanRes, err error) {
	err = engine.GetScheduler().ConfirmPlan(ctx, int64(req.ProjectID))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowConfirmPlanRes{}, nil
}

// Pause 暂停项目
func (c *cWorkflow) Pause(ctx context.Context, req *v1.WorkflowPauseReq) (res *v1.WorkflowPauseRes, err error) {
	err = engine.GetScheduler().Pause(ctx, int64(req.ProjectID), req.PauseReason)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowPauseRes{}, nil
}

// Resume 恢复项目
func (c *cWorkflow) Resume(ctx context.Context, req *v1.WorkflowResumeReq) (res *v1.WorkflowResumeRes, err error) {
	err = engine.GetScheduler().Resume(ctx, int64(req.ProjectID))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowResumeRes{}, nil
}

// RetryTask 重新执行失败任务
func (c *cWorkflow) RetryTask(ctx context.Context, req *v1.WorkflowRetryTaskReq) (res *v1.WorkflowRetryTaskRes, err error) {
	engine.GetWatchdog().ResetRetryCount(int64(req.TaskID))
	err = engine.GetScheduler().RetryTask(int64(req.ProjectID), int64(req.TaskID))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowRetryTaskRes{}, nil
}

// ParseTasks 手动解析架构师回复中的任务清单（托底机制）
func (c *cWorkflow) ParseTasks(ctx context.Context, req *v1.WorkflowParseTasksReq) (res *v1.WorkflowParseTasksRes, err error) {
	projectID := int64(req.ProjectID)

	// 查找该项目的架构师对话
	conv, err := g.DB().Model("mvp_conversation").
		Where("project_id", projectID).
		Where("role_type", "architect").
		Where("task_id IS NULL OR task_id = 0").
		Where("deleted_at IS NULL").
		One()
	if err != nil || conv.IsEmpty() {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	// 查找最新的 assistant 消息
	msg, err := g.DB().Model("mvp_message").
		Where("conversation_id", conv["id"].Int64()).
		Where("role", "assistant").
		Where("deleted_at IS NULL").
		OrderDesc("created_at").
		One()
	if err != nil || msg.IsEmpty() {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	aiReply := msg["content"].String()
	count, err := engine.GetParser().ParseAndCreateTasks(ctx, projectID, aiReply)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowParseTasksRes{
		HasTasks:  count > 0,
		TaskCount: count,
	}, nil
}

// RolePresets 获取角色预设列表（前端创建项目时读取默认模型）
func (c *cWorkflow) RolePresets(ctx context.Context, req *v1.WorkflowRolePresetsReq) (res *v1.WorkflowRolePresetsRes, err error) {
	presets, err := g.DB().Model("mvp_role_preset AS p").
		LeftJoin("ai_model AS m", "m.id = p.model_id").
		Fields("p.role_type, p.role_level, p.model_id, m.name AS model_name, p.system_prompt").
		Where("p.status", 1).
		Where("p.deleted_at IS NULL").
		OrderAsc("p.sort").
		All()
	if err != nil {
		return nil, err
	}

	list := make([]v1.RolePresetItem, 0, len(presets))
	for _, p := range presets {
		list = append(list, v1.RolePresetItem{
			RoleType:     p["role_type"].String(),
			RoleLevel:    p["role_level"].String(),
			ModelID:      snowflake.JsonInt64(p["model_id"].Int64()),
			ModelName:    p["model_name"].String(),
			SystemPrompt: p["system_prompt"].String(),
		})
	}

	return &v1.WorkflowRolePresetsRes{List: list}, nil
}

// ProjectStatus 获取项目状态
func (c *cWorkflow) ProjectStatus(ctx context.Context, req *v1.WorkflowProjectStatusReq) (res *v1.WorkflowProjectStatusRes, err error) {
	projectID := int64(req.ProjectID)

	// 查项目状态
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil || project.IsEmpty() {
		return nil, err
	}

	// 统计各状态任务数
	type StatusCount struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	var counts []StatusCount
	g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("deleted_at IS NULL").
		Fields("status, COUNT(*) as count").
		Group("status").
		Scan(&counts)

	statusCounts := make(map[string]int)
	total := 0
	for _, sc := range counts {
		statusCounts[sc.Status] = sc.Count
		total += sc.Count
	}

	return &v1.WorkflowProjectStatusRes{
		Status:       project["status"].String(),
		PauseReason:  project["pause_reason"].String(),
		TotalTasks:   total,
		StatusCounts: statusCounts,
	}, nil
}
