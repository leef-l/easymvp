package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Message API

// MessageCreateReq 创建MVP消息表请求
type MessageCreateReq struct {
	g.Meta `path:"/message/create" method:"post" tags:"MVP消息表" summary:"创建MVP消息表"`
	ConversationID snowflake.JsonInt64 `json:"conversationID" v:"required" dc:"对话ID"`
	Role string `json:"role" v:"required|max-length:20" dc:"消息角色"`
	Content string `json:"content" v:"required|max-length:4294967295" dc:"消息内容"`
	ModelID snowflake.JsonInt64 `json:"modelID"  dc:"使用的AI模型ID"`
	TokenUsage string `json:"tokenUsage"  dc:"token消耗"`
	Status string `json:"status" v:"max-length:20" dc:"状态"`
}

// MessageCreateRes 创建MVP消息表响应
type MessageCreateRes struct {
	g.Meta `mime:"application/json"`
}

// MessageUpdateReq 更新MVP消息表请求
type MessageUpdateReq struct {
	g.Meta `path:"/message/update" method:"put" tags:"MVP消息表" summary:"更新MVP消息表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP消息表ID"`
	ConversationID snowflake.JsonInt64 `json:"conversationID" dc:"对话ID"`
	Role string `json:"role" dc:"消息角色"`
	Content string `json:"content" dc:"消息内容"`
	ModelID snowflake.JsonInt64 `json:"modelID" dc:"使用的AI模型ID"`
	TokenUsage string `json:"tokenUsage" dc:"token消耗"`
	Status string `json:"status" dc:"状态"`
}

// MessageUpdateRes 更新MVP消息表响应
type MessageUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// MessageDeleteReq 删除MVP消息表请求
type MessageDeleteReq struct {
	g.Meta `path:"/message/delete" method:"delete" tags:"MVP消息表" summary:"删除MVP消息表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP消息表ID"`
}

// MessageDeleteRes 删除MVP消息表响应
type MessageDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// MessageBatchDeleteReq 批量删除MVP消息表请求
type MessageBatchDeleteReq struct {
	g.Meta `path:"/message/batch-delete" method:"delete" tags:"MVP消息表" summary:"批量删除MVP消息表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP消息表ID列表"`
}

// MessageBatchDeleteRes 批量删除MVP消息表响应
type MessageBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// MessageBatchUpdateReq 批量编辑MVP消息表请求
type MessageBatchUpdateReq struct {
	g.Meta `path:"/message/batch-update" method:"put" tags:"MVP消息表" summary:"批量编辑MVP消息表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP消息表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// MessageBatchUpdateRes 批量编辑MVP消息表响应
type MessageBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// MessageDetailReq 获取MVP消息表详情请求
type MessageDetailReq struct {
	g.Meta `path:"/message/detail" method:"get" tags:"MVP消息表" summary:"获取MVP消息表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP消息表ID"`
}

// MessageDetailRes 获取MVP消息表详情响应
type MessageDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.MessageDetailOutput
}

// MessageListReq 获取MVP消息表列表请求
type MessageListReq struct {
	g.Meta    `path:"/message/list" method:"get" tags:"MVP消息表" summary:"获取MVP消息表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
}

// MessageListRes 获取MVP消息表列表响应
type MessageListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.MessageListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// MessageExportReq 导出MVP消息表请求
type MessageExportReq struct {
	g.Meta    `path:"/message/export" method:"get" tags:"MVP消息表" summary:"导出MVP消息表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
}

// MessageExportRes 导出MVP消息表响应
type MessageExportRes struct {
	g.Meta `mime:"text/csv"`
}



// MessageImportReq 导入MVP消息表请求
type MessageImportReq struct {
	g.Meta `path:"/message/import" method:"post" mime:"multipart/form-data" tags:"MVP消息表" summary:"导入MVP消息表"`
}

// MessageImportRes 导入MVP消息表响应
type MessageImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// MessageImportTemplateReq 下载MVP消息表导入模板
type MessageImportTemplateReq struct {
	g.Meta `path:"/message/import-template" method:"get" tags:"MVP消息表" summary:"下载MVP消息表导入模板"`
}

// MessageImportTemplateRes 下载MVP消息表导入模板响应
type MessageImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

