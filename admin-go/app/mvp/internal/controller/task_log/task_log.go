package tasklog

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var TaskLog = cTaskLog{}

type cTaskLog struct{}

// Create 创建任务日志表
func (c *cTaskLog) Create(ctx context.Context, req *v1.TaskLogCreateReq) (res *v1.TaskLogCreateRes, err error) {
	err = service.TaskLog().Create(ctx, &model.TaskLogCreateInput{
		TaskID: req.TaskID,
		Action: req.Action,
		FromStatus: req.FromStatus,
		ToStatus: req.ToStatus,
		Message: req.Message,
		Operator: req.Operator,
	})
	return
}

// Update 更新任务日志表
func (c *cTaskLog) Update(ctx context.Context, req *v1.TaskLogUpdateReq) (res *v1.TaskLogUpdateRes, err error) {
	err = service.TaskLog().Update(ctx, &model.TaskLogUpdateInput{
		ID: req.ID,
		TaskID: req.TaskID,
		Action: req.Action,
		FromStatus: req.FromStatus,
		ToStatus: req.ToStatus,
		Message: req.Message,
		Operator: req.Operator,
	})
	return
}

// Delete 删除任务日志表
func (c *cTaskLog) Delete(ctx context.Context, req *v1.TaskLogDeleteReq) (res *v1.TaskLogDeleteRes, err error) {
	err = service.TaskLog().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除任务日志表
func (c *cTaskLog) BatchDelete(ctx context.Context, req *v1.TaskLogBatchDeleteReq) (res *v1.TaskLogBatchDeleteRes, err error) {
	err = service.TaskLog().BatchDelete(ctx, req.IDs)
	return
}

// Detail 获取任务日志表详情
func (c *cTaskLog) Detail(ctx context.Context, req *v1.TaskLogDetailReq) (res *v1.TaskLogDetailRes, err error) {
	res = &v1.TaskLogDetailRes{}
	res.TaskLogDetailOutput, err = service.TaskLog().Detail(ctx, req.ID)
	return
}

// List 获取任务日志表列表
func (c *cTaskLog) List(ctx context.Context, req *v1.TaskLogListReq) (res *v1.TaskLogListRes, err error) {
	res = &v1.TaskLogListRes{}
	res.List, res.Total, err = service.TaskLog().List(ctx, &model.TaskLogListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	return
}
// Export 导出任务日志表
func (c *cTaskLog) Export(ctx context.Context, req *v1.TaskLogExportReq) (res *v1.TaskLogExportRes, err error) {
	list, err := service.TaskLog().Export(ctx, &model.TaskLogListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="task_log.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 使用 encoding/csv 正确转义，防止 CSV 注入和格式损坏
	w := csv.NewWriter(r.Response.Writer)
	// 表头
	_ = w.Write([]string{"任务名称", "动作", "原状态", "新状态", "日志内容", "操作者", "创建时间"})
	// 数据行
	for _, item := range list {
		_ = w.Write([]string{
			fmt.Sprintf("%v", item.TaskName),
			fmt.Sprintf("%v", item.Action),
			fmt.Sprintf("%v", item.FromStatus),
			fmt.Sprintf("%v", item.ToStatus),
			fmt.Sprintf("%v", item.Message),
			fmt.Sprintf("%v", item.Operator),
			fmt.Sprintf("%v", item.CreatedAt),
		})
	}
	w.Flush()
	return
}

// Import 导入任务日志表
func (c *cTaskLog) Import(ctx context.Context, req *v1.TaskLogImportReq) (res *v1.TaskLogImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.TaskLog().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.TaskLogImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载任务日志表导入模板
func (c *cTaskLog) ImportTemplate(ctx context.Context, req *v1.TaskLogImportTemplateReq) (res *v1.TaskLogImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="task_log_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("任务ID,动作,原状态,新状态,日志内容,操作者")
	return
}


