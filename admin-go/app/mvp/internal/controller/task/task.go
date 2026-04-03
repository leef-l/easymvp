package task

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var Task = cTask{}

type cTask struct{}

// Create 创建MVP任务表
func (c *cTask) Create(ctx context.Context, req *v1.TaskCreateReq) (res *v1.TaskCreateRes, err error) {
	err = service.Task().Create(ctx, &model.TaskCreateInput{
		ProjectID: req.ProjectID,
		ParentID: req.ParentID,
		Name: req.Name,
		Description: req.Description,
		RoleType: req.RoleType,
		RoleLevel: req.RoleLevel,
		ModelID: req.ModelID,
		Status: req.Status,
		Sort: req.Sort,
		BatchNo: req.BatchNo,
		AffectedResources: req.AffectedResources,
		DependsOn: req.DependsOn,
		Result: req.Result,
		ContextSummary: req.ContextSummary,
		ErrorMessage: req.ErrorMessage,
		StartedAt: req.StartedAt,
		CompletedAt: req.CompletedAt,
	})
	return
}

// Update 更新MVP任务表
func (c *cTask) Update(ctx context.Context, req *v1.TaskUpdateReq) (res *v1.TaskUpdateRes, err error) {
	err = service.Task().Update(ctx, &model.TaskUpdateInput{
		ID: req.ID,
		ProjectID: req.ProjectID,
		ParentID: req.ParentID,
		Name: req.Name,
		Description: req.Description,
		RoleType: req.RoleType,
		RoleLevel: req.RoleLevel,
		ModelID: req.ModelID,
		Status: req.Status,
		Sort: req.Sort,
		BatchNo: req.BatchNo,
		AffectedResources: req.AffectedResources,
		DependsOn: req.DependsOn,
		Result: req.Result,
		ContextSummary: req.ContextSummary,
		ErrorMessage: req.ErrorMessage,
		StartedAt: req.StartedAt,
		CompletedAt: req.CompletedAt,
	})
	return
}

// Delete 删除MVP任务表
func (c *cTask) Delete(ctx context.Context, req *v1.TaskDeleteReq) (res *v1.TaskDeleteRes, err error) {
	err = service.Task().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除MVP任务表
func (c *cTask) BatchDelete(ctx context.Context, req *v1.TaskBatchDeleteReq) (res *v1.TaskBatchDeleteRes, err error) {
	err = service.Task().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑MVP任务表
func (c *cTask) BatchUpdate(ctx context.Context, req *v1.TaskBatchUpdateReq) (res *v1.TaskBatchUpdateRes, err error) {
	err = service.Task().BatchUpdate(ctx, &model.TaskBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取MVP任务表详情
func (c *cTask) Detail(ctx context.Context, req *v1.TaskDetailReq) (res *v1.TaskDetailRes, err error) {
	res = &v1.TaskDetailRes{}
	res.TaskDetailOutput, err = service.Task().Detail(ctx, req.ID)
	return
}

// List 获取MVP任务表列表
func (c *cTask) List(ctx context.Context, req *v1.TaskListReq) (res *v1.TaskListRes, err error) {
	res = &v1.TaskListRes{}
	res.List, res.Total, err = service.Task().List(ctx, &model.TaskListInput{
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
// Export 导出MVP任务表
func (c *cTask) Export(ctx context.Context, req *v1.TaskExportReq) (res *v1.TaskExportRes, err error) {
	list, err := service.Task().Export(ctx, &model.TaskListInput{
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
	r.Response.Header().Set("Content-Disposition", `attachment; filename="task.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 使用 encoding/csv 正确转义，防止 CSV 注入和格式损坏
	w := csv.NewWriter(r.Response.Writer)
	// 表头
	_ = w.Write([]string{"项目名称", "父任务名称(0=顶级)", "任务名称", "任务描述", "角色类型", "角色等级", "使用的AI模型ID", "状态", "排序", "执行批次号", "涉及的资源范围", "依赖的任务ID列表", "任务执行结果", "上下文压缩摘要", "错误信息", "开始时间", "完成时间", "创建时间"})
	// 数据行
	for _, item := range list {
		_ = w.Write([]string{
			fmt.Sprintf("%v", item.ProjectName),
			fmt.Sprintf("%v", item.TaskName),
			fmt.Sprintf("%v", item.Name),
			fmt.Sprintf("%v", item.Description),
			fmt.Sprintf("%v", item.RoleType),
			fmt.Sprintf("%v", item.RoleLevel),
			fmt.Sprintf("%v", item.ModelID),
			fmt.Sprintf("%v", item.Status),
			fmt.Sprintf("%v", item.Sort),
			fmt.Sprintf("%v", item.BatchNo),
			fmt.Sprintf("%v", item.AffectedResources),
			fmt.Sprintf("%v", item.DependsOn),
			fmt.Sprintf("%v", item.Result),
			fmt.Sprintf("%v", item.ContextSummary),
			fmt.Sprintf("%v", item.ErrorMessage),
			fmt.Sprintf("%v", item.StartedAt),
			fmt.Sprintf("%v", item.CompletedAt),
			fmt.Sprintf("%v", item.CreatedAt),
		})
	}
	w.Flush()
	return
}


// Tree 获取MVP任务表树形结构
func (c *cTask) Tree(ctx context.Context, req *v1.TaskTreeReq) (res *v1.TaskTreeRes, err error) {
	res = &v1.TaskTreeRes{}
	res.List, err = service.Task().Tree(ctx, &model.TaskTreeInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Name:      req.Name,
		ProjectID: req.ProjectID,
		Status:    req.Status,
		BatchNo:   req.BatchNo,
		RoleType:  req.RoleType,
	})
	return
}

