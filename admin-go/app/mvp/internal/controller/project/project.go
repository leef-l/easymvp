package project

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var Project = cProject{}

type cProject struct{}

// Create 创建MVP项目表
func (c *cProject) Create(ctx context.Context, req *v1.ProjectCreateReq) (res *v1.ProjectCreateRes, err error) {
	err = service.Project().Create(ctx, &model.ProjectCreateInput{
		Name: req.Name,
		Description: req.Description,
		Status: req.Status,
		PauseReason: req.PauseReason,
		GlobalContext: req.GlobalContext,
		ArchitectModelID: req.ArchitectModelID,
	})
	return
}

// Update 更新MVP项目表
func (c *cProject) Update(ctx context.Context, req *v1.ProjectUpdateReq) (res *v1.ProjectUpdateRes, err error) {
	err = service.Project().Update(ctx, &model.ProjectUpdateInput{
		ID: req.ID,
		Name: req.Name,
		Description: req.Description,
		Status: req.Status,
		PauseReason: req.PauseReason,
		GlobalContext: req.GlobalContext,
		ArchitectModelID: req.ArchitectModelID,
	})
	return
}

// Delete 删除MVP项目表
func (c *cProject) Delete(ctx context.Context, req *v1.ProjectDeleteReq) (res *v1.ProjectDeleteRes, err error) {
	err = service.Project().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除MVP项目表
func (c *cProject) BatchDelete(ctx context.Context, req *v1.ProjectBatchDeleteReq) (res *v1.ProjectBatchDeleteRes, err error) {
	err = service.Project().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑MVP项目表
func (c *cProject) BatchUpdate(ctx context.Context, req *v1.ProjectBatchUpdateReq) (res *v1.ProjectBatchUpdateRes, err error) {
	err = service.Project().BatchUpdate(ctx, &model.ProjectBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取MVP项目表详情
func (c *cProject) Detail(ctx context.Context, req *v1.ProjectDetailReq) (res *v1.ProjectDetailRes, err error) {
	res = &v1.ProjectDetailRes{}
	res.ProjectDetailOutput, err = service.Project().Detail(ctx, req.ID)
	return
}

// List 获取MVP项目表列表
func (c *cProject) List(ctx context.Context, req *v1.ProjectListReq) (res *v1.ProjectListRes, err error) {
	res = &v1.ProjectListRes{}
	res.List, res.Total, err = service.Project().List(ctx, &model.ProjectListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Name: req.Name,
	})
	return
}
// Export 导出MVP项目表
func (c *cProject) Export(ctx context.Context, req *v1.ProjectExportReq) (res *v1.ProjectExportRes, err error) {
	list, err := service.Project().Export(ctx, &model.ProjectListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Name: req.Name,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="project.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 表头
	r.Response.Writeln("项目名称,项目简介,状态,暂停原因,项目全局上下文,架构师使用的AI模型ID,创建时间")
	// 数据行
	for _, item := range list {
		r.Response.Writefln("%v,%v,%v,%v,%v,%v,%v",
			item.Name,
			 item.Description,
			 item.Status,
			 item.PauseReason,
			 item.GlobalContext,
			 item.ArchitectModelID,
			item.CreatedAt,
		)
	}
	return
}

// Import 导入MVP项目表
func (c *cProject) Import(ctx context.Context, req *v1.ProjectImportReq) (res *v1.ProjectImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Project().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ProjectImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载MVP项目表导入模板
func (c *cProject) ImportTemplate(ctx context.Context, req *v1.ProjectImportTemplateReq) (res *v1.ProjectImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="project_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("项目名称,项目简介,状态,暂停原因,项目全局上下文,架构师使用的AI模型ID")
	return
}


