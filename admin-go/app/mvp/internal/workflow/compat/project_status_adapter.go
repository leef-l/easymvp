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
	WorkflowStatus  string `json:"workflow_status,omitempty"`
	CurrentStage    string `json:"current_stage,omitempty"`
	ProgressPercent int    `json:"progress_percent"`
	TotalTasks      int    `json:"total_tasks"`
	CompletedTasks  int    `json:"completed_tasks"`
}

// GetProjectStatus 获取项目聚合状态（兼容新旧引擎）。
func (a *ProjectStatusAdapter) GetProjectStatus(ctx context.Context, projectID int64) (*ProjectStatusDTO, error) {
	// TODO: M6 实现 — 按 engine_version 分别查询并聚合
	g.Log().Infof(ctx, "[ProjectStatusAdapter] GetProjectStatus projectID=%d", projectID)
	return &ProjectStatusDTO{ProjectID: projectID}, nil
}
