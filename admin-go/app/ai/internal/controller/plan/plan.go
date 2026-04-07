package plan

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/ai/api/ai/v1"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
)

var Plan = cPlan{}

type cPlan struct{}

// Create 创建AI套餐表
func (c *cPlan) Create(ctx context.Context, req *v1.PlanCreateReq) (res *v1.PlanCreateRes, err error) {
	initCount, err := service.Plan().Create(ctx, &model.PlanCreateInput{
		ProviderID: req.ProviderID,
		Name: req.Name,
		Code: req.Code,
		ApiKey: req.ApiKey,
		ApiSecret: req.ApiSecret,
		Status: req.Status,
		Sort: req.Sort,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.PlanCreateRes{InitCount: initCount}
	return
}

// Update 更新AI套餐表
func (c *cPlan) Update(ctx context.Context, req *v1.PlanUpdateReq) (res *v1.PlanUpdateRes, err error) {
	err = service.Plan().Update(ctx, &model.PlanUpdateInput{
		ID: req.ID,
		ProviderID: req.ProviderID,
		Name: req.Name,
		Code: req.Code,
		ApiKey: req.ApiKey,
		ApiSecret: req.ApiSecret,
		Status: req.Status,
		Sort: req.Sort,
	})
	return
}

// Delete 删除AI套餐表
func (c *cPlan) Delete(ctx context.Context, req *v1.PlanDeleteReq) (res *v1.PlanDeleteRes, err error) {
	err = service.Plan().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除AI套餐表
func (c *cPlan) BatchDelete(ctx context.Context, req *v1.PlanBatchDeleteReq) (res *v1.PlanBatchDeleteRes, err error) {
	err = service.Plan().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑AI套餐表
func (c *cPlan) BatchUpdate(ctx context.Context, req *v1.PlanBatchUpdateReq) (res *v1.PlanBatchUpdateRes, err error) {
	err = service.Plan().BatchUpdate(ctx, &model.PlanBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取AI套餐表详情
func (c *cPlan) Detail(ctx context.Context, req *v1.PlanDetailReq) (res *v1.PlanDetailRes, err error) {
	res = &v1.PlanDetailRes{}
	res.PlanDetailOutput, err = service.Plan().Detail(ctx, req.ID)
	return
}

// List 获取AI套餐表列表
func (c *cPlan) List(ctx context.Context, req *v1.PlanListReq) (res *v1.PlanListRes, err error) {
	res = &v1.PlanListRes{}
	res.List, res.Total, err = service.Plan().List(ctx, &model.PlanListInput{
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
// Export 导出AI套餐表
func (c *cPlan) Export(ctx context.Context, req *v1.PlanExportReq) (res *v1.PlanExportRes, err error) {
	list, err := service.Plan().Export(ctx, &model.PlanListInput{
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
	r.Response.Header().Set("Content-Disposition", `attachment; filename="plan.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 表头（不导出 API Key/Secret，防止敏感信息泄露）
	r.Response.Writeln("供应商ID,套餐名称,套餐代码,状态,排序,创建时间")
	// 数据行
	for _, item := range list {
		r.Response.Writefln("%v,%v,%v,%v,%v,%v",
			item.ProviderName,
			item.Name,
			item.Code,
			item.Status,
			item.Sort,
			item.CreatedAt,
		)
	}
	return
}

// Import 导入AI套餐表
func (c *cPlan) Import(ctx context.Context, req *v1.PlanImportReq) (res *v1.PlanImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Plan().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.PlanImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载AI套餐表导入模板
func (c *cPlan) ImportTemplate(ctx context.Context, req *v1.PlanImportTemplateReq) (res *v1.PlanImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="plan_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("供应商ID,套餐名称,套餐代码,API Key,API Secret,状态,排序")
	return
}


