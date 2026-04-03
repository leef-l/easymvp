package conversation

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
)

var Conversation = cConversation{}

type cConversation struct{}

// Create 创建MVP对话表
func (c *cConversation) Create(ctx context.Context, req *v1.ConversationCreateReq) (res *v1.ConversationCreateRes, err error) {
	err = service.Conversation().Create(ctx, &model.ConversationCreateInput{
		ProjectID: req.ProjectID,
		TaskID: req.TaskID,
		Title: req.Title,
		RoleType: req.RoleType,
		Status: req.Status,
	})
	return
}

// Update 更新MVP对话表
func (c *cConversation) Update(ctx context.Context, req *v1.ConversationUpdateReq) (res *v1.ConversationUpdateRes, err error) {
	err = service.Conversation().Update(ctx, &model.ConversationUpdateInput{
		ID: req.ID,
		ProjectID: req.ProjectID,
		TaskID: req.TaskID,
		Title: req.Title,
		RoleType: req.RoleType,
		Status: req.Status,
	})
	return
}

// Delete 删除MVP对话表
func (c *cConversation) Delete(ctx context.Context, req *v1.ConversationDeleteReq) (res *v1.ConversationDeleteRes, err error) {
	err = service.Conversation().Delete(ctx, req.ID)
	return
}

// BatchDelete 批量删除MVP对话表
func (c *cConversation) BatchDelete(ctx context.Context, req *v1.ConversationBatchDeleteReq) (res *v1.ConversationBatchDeleteRes, err error) {
	err = service.Conversation().BatchDelete(ctx, req.IDs)
	return
}

// BatchUpdate 批量编辑MVP对话表
func (c *cConversation) BatchUpdate(ctx context.Context, req *v1.ConversationBatchUpdateReq) (res *v1.ConversationBatchUpdateRes, err error) {
	err = service.Conversation().BatchUpdate(ctx, &model.ConversationBatchUpdateInput{
		IDs:    req.IDs,
		Status: req.Status,
	})
	return
}

// Detail 获取MVP对话表详情
func (c *cConversation) Detail(ctx context.Context, req *v1.ConversationDetailReq) (res *v1.ConversationDetailRes, err error) {
	res = &v1.ConversationDetailRes{}
	res.ConversationDetailOutput, err = service.Conversation().Detail(ctx, req.ID)
	return
}

// List 获取MVP对话表列表
func (c *cConversation) List(ctx context.Context, req *v1.ConversationListReq) (res *v1.ConversationListRes, err error) {
	res = &v1.ConversationListRes{}
	res.List, res.Total, err = service.Conversation().List(ctx, &model.ConversationListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Title: req.Title,
	})
	return
}
// Export 导出MVP对话表
func (c *cConversation) Export(ctx context.Context, req *v1.ConversationExportReq) (res *v1.ConversationExportRes, err error) {
	list, err := service.Conversation().Export(ctx, &model.ConversationListInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Title: req.Title,
	})
	if err != nil {
		return
	}
	// CSV 导出
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="conversation.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	// 使用 encoding/csv 正确转义，防止 CSV 注入和格式损坏
	w := csv.NewWriter(r.Response.Writer)
	// 表头
	_ = w.Write([]string{"项目名称", "关联任务名称(NULL=项目级对话)", "对话标题", "对话角色类型", "状态", "创建时间"})
	// 数据行
	for _, item := range list {
		_ = w.Write([]string{
			fmt.Sprintf("%v", item.ProjectName),
			fmt.Sprintf("%v", item.TaskName),
			fmt.Sprintf("%v", item.Title),
			fmt.Sprintf("%v", item.RoleType),
			fmt.Sprintf("%v", item.Status),
			fmt.Sprintf("%v", item.CreatedAt),
		})
	}
	w.Flush()
	return
}

// Import 导入MVP对话表
func (c *cConversation) Import(ctx context.Context, req *v1.ConversationImportReq) (res *v1.ConversationImportRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, fmt.Errorf("请上传文件")
	}
	success, fail, err := service.Conversation().Import(ctx, file)
	if err != nil {
		return nil, err
	}
	res = &v1.ConversationImportRes{Success: success, Fail: fail}
	return
}

// ImportTemplate 下载MVP对话表导入模板
func (c *cConversation) ImportTemplate(ctx context.Context, req *v1.ConversationImportTemplateReq) (res *v1.ConversationImportTemplateRes, err error) {
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="conversation_template.csv"`)
	r.Response.Write("\xEF\xBB\xBF") // UTF-8 BOM
	r.Response.Writeln("项目ID,关联任务ID，NULL=项目级对话,对话标题,对话角色类型,状态")
	return
}


