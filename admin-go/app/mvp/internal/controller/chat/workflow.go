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
