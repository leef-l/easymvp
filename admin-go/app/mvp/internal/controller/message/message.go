package message

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var Message = cMessage{}

type cMessage struct{}

// Create 创建MVP消息表
func (c *cMessage) Create(ctx context.Context, req *v1.MessageCreateReq) (res *v1.MessageCreateRes, err error) {
	err = service.Message().Create(ctx, &model.MessageCreateInput{
		ConversationID: req.ConversationID,
		Role: req.Role,
		MessageType: req.MessageType,
		Content: req.Content,
		ModelID: req.ModelID,
		TokenUsage: req.TokenUsage,
		Status: req.Status,
	})
	return
}

// Update 更新MVP消息表
func (c *cMessage) Update(ctx context.Context, req *v1.MessageUpdateReq) (res *v1.MessageUpdateRes, err error) {
	err = service.Message().Update(ctx, &model.MessageUpdateInput{
		ID: req.ID,
		ConversationID: req.ConversationID,
		Role: req.Role,
		MessageType: req.MessageType,
		Content: req.Content,
		ModelID: req.ModelID,
		TokenUsage: req.TokenUsage,
		Status: req.Status,
	})
	return
}

// Delete 删除MVP消息表
func (c *cMessage) Delete(ctx context.Context, req *v1.MessageDeleteReq) (res *v1.MessageDeleteRes, err error) {
	err = service.Message().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除MVP消息表
func (c *cMessage) BatchDelete(ctx context.Context, req *v1.MessageBatchDeleteReq) (res *v1.MessageBatchDeleteRes, err error) {
	err = service.Message().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑MVP消息表
func (c *cMessage) BatchUpdate(ctx context.Context, req *v1.MessageBatchUpdateReq) (res *v1.MessageBatchUpdateRes, err error) {
	err = service.Message().BatchUpdate(ctx, &model.MessageBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取MVP消息表详情
func (c *cMessage) Detail(ctx context.Context, req *v1.MessageDetailReq) (res *v1.MessageDetailRes, err error) {
	res = &v1.MessageDetailRes{}
	res.MessageDetailOutput, err = service.Message().Detail(ctx, req.ID)
	return
}

// List 获取MVP消息表列表
func (c *cMessage) List(ctx context.Context, req *v1.MessageListReq) (res *v1.MessageListRes, err error) {
	res = &v1.MessageListRes{}
	res.List, res.Total, err = service.Message().List(ctx, &model.MessageListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		MessageType: req.MessageType,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	return
}
// Export 导出MVP消息表
func (c *cMessage) Export(ctx context.Context, req *v1.MessageExportReq) (res *v1.MessageExportRes, err error) {
	list, err := service.Message().Export(ctx, &model.MessageListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		MessageType: req.MessageType,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="message.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 使用 encoding/csv 正确转义，防止 CSV 注入和格式损坏
	w := csv.NewWriter(r.Response.Writer)
	// 表头
	_ = w.Write([]string{"对话标题", "消息角色", "消息类型", "消息内容", "使用的AI模型ID", "token消耗", "状态", "创建时间"})
	// 数据行
	for _, item := range list {
		_ = w.Write([]string{
			fmt.Sprintf("%v", item.ConversationTitle),
			fmt.Sprintf("%v", item.Role),
			fmt.Sprintf("%v", item.MessageType),
			fmt.Sprintf("%v", item.Content),
			fmt.Sprintf("%v", item.ModelID),
			fmt.Sprintf("%v", item.TokenUsage),
			fmt.Sprintf("%v", item.Status),
			fmt.Sprintf("%v", item.CreatedAt),
		})
	}
	w.Flush()
	return
}

// Import 导入MVP消息表
func (c *cMessage) Import(ctx context.Context, req *v1.MessageImportReq) (res *v1.MessageImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Message().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.MessageImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载MVP消息表导入模板
func (c *cMessage) ImportTemplate(ctx context.Context, req *v1.MessageImportTemplateReq) (res *v1.MessageImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="message_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("对话ID,消息角色,消息类型,消息内容,使用的AI模型ID,token消耗,状态")
	return
}


