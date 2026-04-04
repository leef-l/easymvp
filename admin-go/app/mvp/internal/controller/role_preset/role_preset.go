package rolepreset

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var RolePreset = cRolePreset{}

type cRolePreset struct{}

// Create 创建角色预设模板
func (c *cRolePreset) Create(ctx context.Context, req *v1.RolePresetCreateReq) (res *v1.RolePresetCreateRes, err error) {
	err = service.RolePreset().Create(ctx, &model.RolePresetCreateInput{
		ProjectCategory: req.ProjectCategory,
		RoleType: req.RoleType,
		RoleLevel: req.RoleLevel,
		ModelID: req.ModelID,
		SystemPrompt: req.SystemPrompt,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Update 更新角色预设模板
func (c *cRolePreset) Update(ctx context.Context, req *v1.RolePresetUpdateReq) (res *v1.RolePresetUpdateRes, err error) {
	err = service.RolePreset().Update(ctx, &model.RolePresetUpdateInput{
		ID: req.ID,
		ProjectCategory: req.ProjectCategory,
		RoleType: req.RoleType,
		RoleLevel: req.RoleLevel,
		ModelID: req.ModelID,
		SystemPrompt: req.SystemPrompt,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Delete 删除角色预设模板
func (c *cRolePreset) Delete(ctx context.Context, req *v1.RolePresetDeleteReq) (res *v1.RolePresetDeleteRes, err error) {
	err = service.RolePreset().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除角色预设模板
func (c *cRolePreset) BatchDelete(ctx context.Context, req *v1.RolePresetBatchDeleteReq) (res *v1.RolePresetBatchDeleteRes, err error) {
	err = service.RolePreset().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑角色预设模板
func (c *cRolePreset) BatchUpdate(ctx context.Context, req *v1.RolePresetBatchUpdateReq) (res *v1.RolePresetBatchUpdateRes, err error) {
	err = service.RolePreset().BatchUpdate(ctx, &model.RolePresetBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取角色预设模板详情
func (c *cRolePreset) Detail(ctx context.Context, req *v1.RolePresetDetailReq) (res *v1.RolePresetDetailRes, err error) {
	res = &v1.RolePresetDetailRes{}
	res.RolePresetDetailOutput, err = service.RolePreset().Detail(ctx, req.ID)
	return
}

// List 获取角色预设模板列表
func (c *cRolePreset) List(ctx context.Context, req *v1.RolePresetListReq) (res *v1.RolePresetListRes, err error) {
	res = &v1.RolePresetListRes{}
	res.List, res.Total, err = service.RolePreset().List(ctx, &model.RolePresetListInput{
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
// Export 导出角色预设模板
func (c *cRolePreset) Export(ctx context.Context, req *v1.RolePresetExportReq) (res *v1.RolePresetExportRes, err error) {
	list, err := service.RolePreset().Export(ctx, &model.RolePresetListInput{
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
	r.Response.Header().Set("Content-Disposition", `attachment; filename="role_preset.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 使用 encoding/csv 正确转义，防止 CSV 注入和格式损坏
	w := csv.NewWriter(r.Response.Writer)
	// 表头
	_ = w.Write([]string{"角色类型", "角色等级", "AI模型ID", "默认系统提示词", "状态", "排序", "创建时间"})
	// 数据行
	for _, item := range list {
		_ = w.Write([]string{
			fmt.Sprintf("%v", item.RoleType),
			fmt.Sprintf("%v", item.RoleLevel),
			fmt.Sprintf("%v", item.ModelID),
			fmt.Sprintf("%v", item.SystemPrompt),
			fmt.Sprintf("%v", item.Status),
			fmt.Sprintf("%v", item.Sort),
			fmt.Sprintf("%v", item.CreatedAt),
		})
	}
	w.Flush()
	return
}

// Import 导入角色预设模板
func (c *cRolePreset) Import(ctx context.Context, req *v1.RolePresetImportReq) (res *v1.RolePresetImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.RolePreset().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.RolePresetImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载角色预设模板导入模板
func (c *cRolePreset) ImportTemplate(ctx context.Context, req *v1.RolePresetImportTemplateReq) (res *v1.RolePresetImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="role_preset_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("角色类型,角色等级,AI模型ID,默认系统提示词,状态,排序")
	return
}


