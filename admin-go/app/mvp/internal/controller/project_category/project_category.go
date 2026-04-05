package projectcategory

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var ProjectCategory = cProjectCategory{}

type cProjectCategory struct{}

// Create 创建项目分类配置表
func (c *cProjectCategory) Create(ctx context.Context, req *v1.ProjectCategoryCreateReq) (res *v1.ProjectCategoryCreateRes, err error) {
	err = service.ProjectCategory().Create(ctx, &model.ProjectCategoryCreateInput{
		CategoryCode: req.CategoryCode,
		DisplayName: req.DisplayName,
		FamilyCode: req.FamilyCode,
		Description: req.Description,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Update 更新项目分类配置表
func (c *cProjectCategory) Update(ctx context.Context, req *v1.ProjectCategoryUpdateReq) (res *v1.ProjectCategoryUpdateRes, err error) {
	err = service.ProjectCategory().Update(ctx, &model.ProjectCategoryUpdateInput{
		ID: req.ID,
		CategoryCode: req.CategoryCode,
		DisplayName: req.DisplayName,
		FamilyCode: req.FamilyCode,
		Description: req.Description,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Delete 删除项目分类配置表
func (c *cProjectCategory) Delete(ctx context.Context, req *v1.ProjectCategoryDeleteReq) (res *v1.ProjectCategoryDeleteRes, err error) {
	err = service.ProjectCategory().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除项目分类配置表
func (c *cProjectCategory) BatchDelete(ctx context.Context, req *v1.ProjectCategoryBatchDeleteReq) (res *v1.ProjectCategoryBatchDeleteRes, err error) {
	err = service.ProjectCategory().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑项目分类配置表
func (c *cProjectCategory) BatchUpdate(ctx context.Context, req *v1.ProjectCategoryBatchUpdateReq) (res *v1.ProjectCategoryBatchUpdateRes, err error) {
	err = service.ProjectCategory().BatchUpdate(ctx, &model.ProjectCategoryBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取项目分类配置表详情
func (c *cProjectCategory) Detail(ctx context.Context, req *v1.ProjectCategoryDetailReq) (res *v1.ProjectCategoryDetailRes, err error) {
	res = &v1.ProjectCategoryDetailRes{}
	res.ProjectCategoryDetailOutput, err = service.ProjectCategory().Detail(ctx, req.ID)
	return
}

// List 获取项目分类配置表列表
func (c *cProjectCategory) List(ctx context.Context, req *v1.ProjectCategoryListReq) (res *v1.ProjectCategoryListRes, err error) {
	res = &v1.ProjectCategoryListRes{}
	res.List, res.Total, err = service.ProjectCategory().List(ctx, &model.ProjectCategoryListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		DisplayName: req.DisplayName,
	})
	return
}
// Export 导出项目分类配置表
func (c *cProjectCategory) Export(ctx context.Context, req *v1.ProjectCategoryExportReq) (res *v1.ProjectCategoryExportRes, err error) {
	list, err := service.ProjectCategory().Export(ctx, &model.ProjectCategoryListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="project_category.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 表头
	r.Response.Writeln("稳定分类编码,展示名称,能力家族编码,分类说明,1启用 0停用,排序,创建时间")
	// 数据行
	for _, item := range list {
		r.Response.Writefln("%v,%v,%v,%v,%v,%v,%v",
			item.CategoryCode,
			 item.DisplayName,
			 item.FamilyCode,
			 item.Description,
			 item.Status,
			 item.Sort,
			item.CreatedAt,
		)
	}
	return
}

// Import 导入项目分类配置表
func (c *cProjectCategory) Import(ctx context.Context, req *v1.ProjectCategoryImportReq) (res *v1.ProjectCategoryImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.ProjectCategory().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ProjectCategoryImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载项目分类配置表导入模板
func (c *cProjectCategory) ImportTemplate(ctx context.Context, req *v1.ProjectCategoryImportTemplateReq) (res *v1.ProjectCategoryImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="project_category_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("稳定分类编码,展示名称,能力家族编码,分类说明,1启用 0停用,排序")
	return
}


