package projectrole

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var ProjectRole = cProjectRole{}

type cProjectRole struct{}

// Create 创建项目角色配置表
func (c *cProjectRole) Create(ctx context.Context, req *v1.ProjectRoleCreateReq) (res *v1.ProjectRoleCreateRes, err error) {
	err = service.ProjectRole().Create(ctx, &model.ProjectRoleCreateInput{
		ProjectID:     req.ProjectID,
		ProjectCategory: req.ProjectCategory,
		RoleType:      req.RoleType,
		RoleLevel:     req.RoleLevel,
		ModelID:       req.ModelID,
		SystemPrompt:  req.SystemPrompt,
		ExecutionMode: req.ExecutionMode,
		Status:        req.Status,
	})
	return
}

// Update 更新项目角色配置表
func (c *cProjectRole) Update(ctx context.Context, req *v1.ProjectRoleUpdateReq) (res *v1.ProjectRoleUpdateRes, err error) {
	err = service.ProjectRole().Update(ctx, &model.ProjectRoleUpdateInput{
		ID:            req.ID,
		ProjectID:     req.ProjectID,
		ProjectCategory: req.ProjectCategory,
		RoleType:      req.RoleType,
		RoleLevel:     req.RoleLevel,
		ModelID:       req.ModelID,
		SystemPrompt:  req.SystemPrompt,
		ExecutionMode: req.ExecutionMode,
		Status:        req.Status,
	})
	return
}

// Delete 删除项目角色配置表
func (c *cProjectRole) Delete(ctx context.Context, req *v1.ProjectRoleDeleteReq) (res *v1.ProjectRoleDeleteRes, err error) {
	err = service.ProjectRole().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除项目角色配置表
func (c *cProjectRole) BatchDelete(ctx context.Context, req *v1.ProjectRoleBatchDeleteReq) (res *v1.ProjectRoleBatchDeleteRes, err error) {
	err = service.ProjectRole().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑项目角色配置表
func (c *cProjectRole) BatchUpdate(ctx context.Context, req *v1.ProjectRoleBatchUpdateReq) (res *v1.ProjectRoleBatchUpdateRes, err error) {
	err = service.ProjectRole().BatchUpdate(ctx, &model.ProjectRoleBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取项目角色配置表详情
func (c *cProjectRole) Detail(ctx context.Context, req *v1.ProjectRoleDetailReq) (res *v1.ProjectRoleDetailRes, err error) {
	res = &v1.ProjectRoleDetailRes{}
	res.ProjectRoleDetailOutput, err = service.ProjectRole().Detail(ctx, req.ID)
	return
}

// List 获取项目角色配置表列表
func (c *cProjectRole) List(ctx context.Context, req *v1.ProjectRoleListReq) (res *v1.ProjectRoleListRes, err error) {
	res = &v1.ProjectRoleListRes{}
	res.List, res.Total, err = service.ProjectRole().List(ctx, &model.ProjectRoleListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status: req.Status,
	})
	return
}
// Export 导出项目角色配置表
func (c *cProjectRole) Export(ctx context.Context, req *v1.ProjectRoleExportReq) (res *v1.ProjectRoleExportRes, err error) {
	list, err := service.ProjectRole().Export(ctx, &model.ProjectRoleListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status: req.Status,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="project_role.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 使用 encoding/csv 正确转义，防止 CSV 注入和格式损坏
	w := csv.NewWriter(r.Response.Writer)
	// 表头
	_ = w.Write([]string{"项目名称", "角色类型", "角色等级", "AI模型ID", "系统提示词", "状态", "创建时间"})
	// 数据行
	for _, item := range list {
		_ = w.Write([]string{
			fmt.Sprintf("%v", item.ProjectName),
			fmt.Sprintf("%v", item.RoleType),
			fmt.Sprintf("%v", item.RoleLevel),
			fmt.Sprintf("%v", item.ModelID),
			fmt.Sprintf("%v", item.SystemPrompt),
			fmt.Sprintf("%v", item.Status),
			fmt.Sprintf("%v", item.CreatedAt),
		})
	}
	w.Flush()
	return
}

// Import 导入项目角色配置表
func (c *cProjectRole) Import(ctx context.Context, req *v1.ProjectRoleImportReq) (res *v1.ProjectRoleImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.ProjectRole().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ProjectRoleImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载项目角色配置表导入模板
func (c *cProjectRole) ImportTemplate(ctx context.Context, req *v1.ProjectRoleImportTemplateReq) (res *v1.ProjectRoleImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="project_role_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("项目ID,角色类型,角色等级,AI模型ID,系统提示词,状态")
	return
}


