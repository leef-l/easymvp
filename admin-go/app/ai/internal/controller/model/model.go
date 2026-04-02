package model

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/ai/api/ai/v1"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
)

var Model = cModel{}

type cModel struct{}

// Create 创建AI模型表
func (c *cModel) Create(ctx context.Context, req *v1.ModelCreateReq) (res *v1.ModelCreateRes, err error) {
	err = service.Model().Create(ctx, &model.ModelCreateInput{
		PlanID: req.PlanID,
		ProviderID: req.ProviderID,
		Name: req.Name,
		ModelCode: req.ModelCode,
		Capability: req.Capability,
		MaxTokens: req.MaxTokens,
		ContextWindow: req.ContextWindow,
		SupportsStream: req.SupportsStream,
		RolePrompt: req.RolePrompt,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Update 更新AI模型表
func (c *cModel) Update(ctx context.Context, req *v1.ModelUpdateReq) (res *v1.ModelUpdateRes, err error) {
	err = service.Model().Update(ctx, &model.ModelUpdateInput{
		ID: req.ID,
		PlanID: req.PlanID,
		ProviderID: req.ProviderID,
		Name: req.Name,
		ModelCode: req.ModelCode,
		Capability: req.Capability,
		MaxTokens: req.MaxTokens,
		ContextWindow: req.ContextWindow,
		SupportsStream: req.SupportsStream,
		RolePrompt: req.RolePrompt,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Delete 删除AI模型表
func (c *cModel) Delete(ctx context.Context, req *v1.ModelDeleteReq) (res *v1.ModelDeleteRes, err error) {
	err = service.Model().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除AI模型表
func (c *cModel) BatchDelete(ctx context.Context, req *v1.ModelBatchDeleteReq) (res *v1.ModelBatchDeleteRes, err error) {
	err = service.Model().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑AI模型表
func (c *cModel) BatchUpdate(ctx context.Context, req *v1.ModelBatchUpdateReq) (res *v1.ModelBatchUpdateRes, err error) {
	err = service.Model().BatchUpdate(ctx, &model.ModelBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取AI模型表详情
func (c *cModel) Detail(ctx context.Context, req *v1.ModelDetailReq) (res *v1.ModelDetailRes, err error) {
	res = &v1.ModelDetailRes{}
	res.ModelDetailOutput, err = service.Model().Detail(ctx, req.ID)
	return
}

// List 获取AI模型表列表
func (c *cModel) List(ctx context.Context, req *v1.ModelListReq) (res *v1.ModelListRes, err error) {
	res = &v1.ModelListRes{}
	res.List, res.Total, err = service.Model().List(ctx, &model.ModelListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		SupportsStream: req.SupportsStream,
		Status: req.Status,
		Name: req.Name,
	})
	return
}
// Export 导出AI模型表
func (c *cModel) Export(ctx context.Context, req *v1.ModelExportReq) (res *v1.ModelExportRes, err error) {
	list, err := service.Model().Export(ctx, &model.ModelListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		SupportsStream: req.SupportsStream,
		Status: req.Status,
		Name: req.Name,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="model.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 表头
	r.Response.Writeln("套餐ID,供应商ID,模型显示名称,模型代码,能力,最大输出token,上下文窗口大小,是否支持流式输出,默认角色提示词,状态,排序,创建时间")
	// 数据行
	for _, item := range list {
		r.Response.Writefln("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v",
			item.PlanName,
			 item.ProviderName,
			 item.Name,
			 item.ModelCode,
			 item.Capability,
			 item.MaxTokens,
			 item.ContextWindow,
			 item.SupportsStream,
			 item.RolePrompt,
			 item.Status,
			 item.Sort,
			item.CreatedAt,
		)
	}
	return
}

// Import 导入AI模型表
func (c *cModel) Import(ctx context.Context, req *v1.ModelImportReq) (res *v1.ModelImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Model().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ModelImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载AI模型表导入模板
func (c *cModel) ImportTemplate(ctx context.Context, req *v1.ModelImportTemplateReq) (res *v1.ModelImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="model_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("套餐ID,供应商ID,模型显示名称,模型代码,能力,最大输出token,上下文窗口大小,是否支持流式输出,默认角色提示词,状态,排序")
	return
}


