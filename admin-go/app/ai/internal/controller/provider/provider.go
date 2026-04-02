package provider

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/ai/api/ai/v1"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
)

var Provider = cProvider{}

type cProvider struct{}

// Create 创建AI供应商表
func (c *cProvider) Create(ctx context.Context, req *v1.ProviderCreateReq) (res *v1.ProviderCreateRes, err error) {
	err = service.Provider().Create(ctx, &model.ProviderCreateInput{
		Name: req.Name,
		Code: req.Code,
		ProviderType: req.ProviderType,
		BaseURL: req.BaseURL,
		Icon: req.Icon,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Update 更新AI供应商表
func (c *cProvider) Update(ctx context.Context, req *v1.ProviderUpdateReq) (res *v1.ProviderUpdateRes, err error) {
	err = service.Provider().Update(ctx, &model.ProviderUpdateInput{
		ID: req.ID,
		Name: req.Name,
		Code: req.Code,
		ProviderType: req.ProviderType,
		BaseURL: req.BaseURL,
		Icon: req.Icon,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Delete 删除AI供应商表
func (c *cProvider) Delete(ctx context.Context, req *v1.ProviderDeleteReq) (res *v1.ProviderDeleteRes, err error) {
	err = service.Provider().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除AI供应商表
func (c *cProvider) BatchDelete(ctx context.Context, req *v1.ProviderBatchDeleteReq) (res *v1.ProviderBatchDeleteRes, err error) {
	err = service.Provider().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑AI供应商表
func (c *cProvider) BatchUpdate(ctx context.Context, req *v1.ProviderBatchUpdateReq) (res *v1.ProviderBatchUpdateRes, err error) {
	err = service.Provider().BatchUpdate(ctx, &model.ProviderBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取AI供应商表详情
func (c *cProvider) Detail(ctx context.Context, req *v1.ProviderDetailReq) (res *v1.ProviderDetailRes, err error) {
	res = &v1.ProviderDetailRes{}
	res.ProviderDetailOutput, err = service.Provider().Detail(ctx, req.ID)
	return
}

// List 获取AI供应商表列表
func (c *cProvider) List(ctx context.Context, req *v1.ProviderListReq) (res *v1.ProviderListRes, err error) {
	res = &v1.ProviderListRes{}
	res.List, res.Total, err = service.Provider().List(ctx, &model.ProviderListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status: req.Status,
		Name: req.Name,
	})
	return
}
// Export 导出AI供应商表
func (c *cProvider) Export(ctx context.Context, req *v1.ProviderExportReq) (res *v1.ProviderExportRes, err error) {
	list, err := service.Provider().Export(ctx, &model.ProviderListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status: req.Status,
		Name: req.Name,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="provider.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 表头
	r.Response.Writeln("供应商名称,供应商代码,Provider类型,API基础地址,图标URL,状态,排序,创建时间")
	// 数据行
	for _, item := range list {
		r.Response.Writefln("%v,%v,%v,%v,%v,%v,%v,%v",
			item.Name,
			 item.Code,
			 item.ProviderType,
			 item.BaseURL,
			 item.Icon,
			 item.Status,
			 item.Sort,
			item.CreatedAt,
		)
	}
	return
}

// Import 导入AI供应商表
func (c *cProvider) Import(ctx context.Context, req *v1.ProviderImportReq) (res *v1.ProviderImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Provider().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ProviderImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载AI供应商表导入模板
func (c *cProvider) ImportTemplate(ctx context.Context, req *v1.ProviderImportTemplateReq) (res *v1.ProviderImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="provider_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("供应商名称,供应商代码,Provider类型,API基础地址,图标URL,状态,排序")
	return
}


