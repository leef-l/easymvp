package compat

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// ProjectStatusAdapter 将新工作流状态适配为旧项目状态 DTO。
type ProjectStatusAdapter struct{}

// NewProjectStatusAdapter 创建适配器。
func NewProjectStatusAdapter() *ProjectStatusAdapter { return &ProjectStatusAdapter{} }

// ProjectStatusDTO 聚合后的项目状态。
type ProjectStatusDTO struct {
	ProjectID       int64  `json:"project_id"`
	EngineVersion   string `json:"engine_version"`
	Status          string `json:"status"`
	WorkflowStatus  string `json:"workflow_status,omitempty"`
	CurrentStage    string `json:"current_stage,omitempty"`
	ProgressPercent int    `json:"progress_percent"`
	TotalTasks      int    `json:"total_tasks"`
	CompletedTasks  int    `json:"completed_tasks"`
	FailedTasks     int    `json:"failed_tasks"`
	RunningTasks    int    `json:"running_tasks"`
}

// GetProjectStatus 获取项目聚合状态（兼容新旧引擎）。
func (a *ProjectStatusAdapter) GetProjectStatus(ctx context.Context, projectID int64) (*ProjectStatusDTO, error) {
	// 查项目基础信息
	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).WhereNull("deleted_at").One()
	if err != nil || project.IsEmpty() {
		return nil, err
	}

	dto := &ProjectStatusDTO{
		ProjectID:     projectID,
		EngineVersion: project["engine_version"].String(),
		Status:        project["status"].String(),
	}

	if dto.EngineVersion == "" {
		dto.EngineVersion = "legacy"
	}

	switch dto.EngineVersion {
	case "workflow_v2":
		a.fillV2Status(ctx, projectID, dto)
	default:
		a.fillLegacyStatus(ctx, projectID, dto)
	}

	return dto, nil
}

// fillV2Status 填充 workflow_v2 项目状态。
func (a *ProjectStatusAdapter) fillV2Status(ctx context.Context, projectID int64, dto *ProjectStatusDTO) {
	// 查最新的 workflow_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if err != nil || wfRun.IsEmpty() {
		return
	}

	dto.WorkflowStatus = wfRun["status"].String()
	dto.CurrentStage = wfRun["current_stage"].String()

	workflowRunID := wfRun["id"].Int64()

	// 查领域任务统计
	totalVal, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Count()
	dto.TotalTasks = totalVal

	if dto.TotalTasks > 0 {
		completedVal, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			WhereIn("status", g.Slice{"completed", "skipped"}).
			WhereNull("deleted_at").
			Count()
		dto.CompletedTasks = completedVal

		failedVal, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			Where("status", "failed").
			WhereNull("deleted_at").
			Count()
		dto.FailedTasks = failedVal

		runningVal, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			Where("status", "running").
			WhereNull("deleted_at").
			Count()
		dto.RunningTasks = runningVal

		dto.ProgressPercent = dto.CompletedTasks * 100 / dto.TotalTasks
	}
}

// fillLegacyStatus 填充旧引擎项目状态。
func (a *ProjectStatusAdapter) fillLegacyStatus(ctx context.Context, projectID int64, dto *ProjectStatusDTO) {
	totalVal, _ := g.DB().Model("mvp_task").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Count()
	dto.TotalTasks = totalVal

	if dto.TotalTasks > 0 {
		completedVal, _ := g.DB().Model("mvp_task").Ctx(ctx).
			Where("project_id", projectID).
			Where("status", "completed").
			WhereNull("deleted_at").
			Count()
		dto.CompletedTasks = completedVal

		failedVal, _ := g.DB().Model("mvp_task").Ctx(ctx).
			Where("project_id", projectID).
			Where("status", "failed").
			WhereNull("deleted_at").
			Count()
		dto.FailedTasks = failedVal

		runningVal, _ := g.DB().Model("mvp_task").Ctx(ctx).
			Where("project_id", projectID).
			Where("status", "running").
			WhereNull("deleted_at").
			Count()
		dto.RunningTasks = runningVal

		dto.ProgressPercent = dto.CompletedTasks * 100 / dto.TotalTasks
	}
}
