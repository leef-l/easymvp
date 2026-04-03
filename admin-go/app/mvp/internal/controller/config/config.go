package config

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var Config = cConfig{}

type cConfig struct{}

// Create 创建MVP配置表
func (c *cConfig) Create(ctx context.Context, req *v1.ConfigCreateReq) (res *v1.ConfigCreateRes, err error) {
	err = service.Config().Create(ctx, &model.ConfigCreateInput{
		ConfigKey: req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType: req.ConfigType,
		Category: req.Category,
		Description: req.Description,
	})
	return
}

// Update 更新MVP配置表
func (c *cConfig) Update(ctx context.Context, req *v1.ConfigUpdateReq) (res *v1.ConfigUpdateRes, err error) {
	err = service.Config().Update(ctx, &model.ConfigUpdateInput{
		ID: req.ID,
		ConfigKey: req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType: req.ConfigType,
		Category: req.Category,
		Description: req.Description,
	})
	return
}

// Delete 删除MVP配置表
func (c *cConfig) Delete(ctx context.Context, req *v1.ConfigDeleteReq) (res *v1.ConfigDeleteRes, err error) {
	err = service.Config().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除MVP配置表
func (c *cConfig) BatchDelete(ctx context.Context, req *v1.ConfigBatchDeleteReq) (res *v1.ConfigBatchDeleteRes, err error) {
	err = service.Config().BatchDelete(ctx, req.IDs)
	return
}

// Detail 获取MVP配置表详情
func (c *cConfig) Detail(ctx context.Context, req *v1.ConfigDetailReq) (res *v1.ConfigDetailRes, err error) {
	res = &v1.ConfigDetailRes{}
	res.ConfigDetailOutput, err = service.Config().Detail(ctx, req.ID)
	return
}

// List 获取MVP配置表列表
func (c *cConfig) List(ctx context.Context, req *v1.ConfigListReq) (res *v1.ConfigListRes, err error) {
	res = &v1.ConfigListRes{}
	res.List, res.Total, err = service.Config().List(ctx, &model.ConfigListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	return
}
// Export 导出MVP配置表
func (c *cConfig) Export(ctx context.Context, req *v1.ConfigExportReq) (res *v1.ConfigExportRes, err error) {
	list, err := service.Config().Export(ctx, &model.ConfigListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="config.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 表头
	r.Response.Writeln("配置键,配置值,值类型,分类,配置说明,创建时间")
	// 数据行
	for _, item := range list {
		r.Response.Writefln("%v,%v,%v,%v,%v,%v",
			item.ConfigKey,
			 item.ConfigValue,
			 item.ConfigType,
			 item.Category,
			 item.Description,
			item.CreatedAt,
		)
	}
	return
}

// Import 导入MVP配置表
func (c *cConfig) Import(ctx context.Context, req *v1.ConfigImportReq) (res *v1.ConfigImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Config().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ConfigImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载MVP配置表导入模板
func (c *cConfig) ImportTemplate(ctx context.Context, req *v1.ConfigImportTemplateReq) (res *v1.ConfigImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="config_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("配置键,配置值,值类型,分类,配置说明")
	return
}


