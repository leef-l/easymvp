package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectScope 项目归属信息。
type ProjectScope struct {
	CreatedBy int64
	DeptID    int64
}

// GetProjectScopeByWorkflowRun 根据 workflow_run_id 获取项目归属信息。
// 结果可缓存在调用方，避免重复查询。
func GetProjectScopeByWorkflowRun(ctx context.Context, workflowRunID int64) *ProjectScope {
	row, err := g.DB().Model("mvp_workflow_run AS wr").Ctx(ctx).
		LeftJoin("mvp_project AS p", "p.id = wr.project_id").
		Fields("p.created_by, p.dept_id").
		Where("wr.id", workflowRunID).
		One()
	if err != nil {
		g.Log().Errorf(ctx, "[ScopeHelper] 查询 workflow_run(%d) 归属信息失败: %v", workflowRunID, err)
		return &ProjectScope{}
	}
	if row.IsEmpty() {
		g.Log().Warningf(ctx, "[ScopeHelper] workflow_run(%d) 未找到关联项目的归属信息", workflowRunID)
		return &ProjectScope{}
	}
	return &ProjectScope{
		CreatedBy: row["created_by"].Int64(),
		DeptID:    row["dept_id"].Int64(),
	}
}

// GetProjectScopeByWorkflowRunInTx 根据 workflow_run_id 在事务中获取项目归属信息。
func GetProjectScopeByWorkflowRunInTx(ctx context.Context, tx gdb.TX, workflowRunID int64) *ProjectScope {
	row, err := tx.Model("mvp_workflow_run AS wr").Ctx(ctx).
		LeftJoin("mvp_project AS p", "p.id = wr.project_id").
		Fields("p.created_by, p.dept_id").
		Where("wr.id", workflowRunID).
		One()
	if err != nil {
		g.Log().Errorf(ctx, "[ScopeHelper] 事务内查询 workflow_run(%d) 归属信息失败: %v", workflowRunID, err)
		return &ProjectScope{}
	}
	if row.IsEmpty() {
		g.Log().Warningf(ctx, "[ScopeHelper] 事务内 workflow_run(%d) 未找到关联项目的归属信息", workflowRunID)
		return &ProjectScope{}
	}
	return &ProjectScope{
		CreatedBy: row["created_by"].Int64(),
		DeptID:    row["dept_id"].Int64(),
	}
}

// GetProjectScopeByProject 根据 project_id 获取项目归属信息。
func GetProjectScopeByProject(ctx context.Context, projectID int64) *ProjectScope {
	row, err := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("created_by, dept_id").
		Where("id", projectID).
		One()
	if err != nil {
		g.Log().Errorf(ctx, "[ScopeHelper] 查询 project(%d) 归属信息失败: %v", projectID, err)
		return &ProjectScope{}
	}
	if row.IsEmpty() {
		g.Log().Warningf(ctx, "[ScopeHelper] project(%d) 不存在", projectID)
		return &ProjectScope{}
	}
	return &ProjectScope{
		CreatedBy: row["created_by"].Int64(),
		DeptID:    row["dept_id"].Int64(),
	}
}

// GetProjectScopeByProjectInTx 根据 project_id 在事务中获取项目归属信息。
func GetProjectScopeByProjectInTx(ctx context.Context, tx gdb.TX, projectID int64) *ProjectScope {
	row, err := tx.Model("mvp_project").Ctx(ctx).
		Fields("created_by, dept_id").
		Where("id", projectID).
		One()
	if err != nil {
		g.Log().Errorf(ctx, "[ScopeHelper] 事务内查询 project(%d) 归属信息失败: %v", projectID, err)
		return &ProjectScope{}
	}
	if row.IsEmpty() {
		g.Log().Warningf(ctx, "[ScopeHelper] 事务内 project(%d) 不存在", projectID)
		return &ProjectScope{}
	}
	return &ProjectScope{
		CreatedBy: row["created_by"].Int64(),
		DeptID:    row["dept_id"].Int64(),
	}
}

// ApplyScopeToMap 将归属字段注入到 g.Map 中（仅当 Map 中没有这些字段时）。
func (s *ProjectScope) ApplyScopeToMap(data g.Map) {
	if _, ok := data["created_by"]; !ok {
		data["created_by"] = s.CreatedBy
	}
	if _, ok := data["dept_id"]; !ok {
		data["dept_id"] = s.DeptID
	}
}
