package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Conversation API

// ConversationCreateReq 创建MVP对话表请求
type ConversationCreateReq struct {
	g.Meta `path:"/conversation/create" method:"post" tags:"MVP对话表" summary:"创建MVP对话表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	TaskID snowflake.JsonInt64 `json:"taskID"  dc:"关联任务ID，NULL=项目级对话"`
	Title string `json:"title" v:"max-length:200" dc:"对话标题"`
	RoleType string `json:"roleType" v:"required|max-length:20" dc:"对话角色类型"`
	Status string `json:"status" v:"max-length:20" dc:"状态"`
}

// ConversationCreateRes 创建MVP对话表响应
type ConversationCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ConversationUpdateReq 更新MVP对话表请求
type ConversationUpdateReq struct {
	g.Meta `path:"/conversation/update" method:"put" tags:"MVP对话表" summary:"更新MVP对话表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP对话表ID"`
	ProjectID snowflake.JsonInt64 `json:"projectID" dc:"项目ID"`
	TaskID snowflake.JsonInt64 `json:"taskID" dc:"关联任务ID，NULL=项目级对话"`
	Title string `json:"title" dc:"对话标题"`
	RoleType string `json:"roleType" dc:"对话角色类型"`
	Status string `json:"status" dc:"状态"`
}

// ConversationUpdateRes 更新MVP对话表响应
type ConversationUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ConversationDeleteReq 删除MVP对话表请求
type ConversationDeleteReq struct {
	g.Meta `path:"/conversation/delete" method:"delete" tags:"MVP对话表" summary:"删除MVP对话表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP对话表ID"`
}

// ConversationDeleteRes 删除MVP对话表响应
type ConversationDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ConversationBatchDeleteReq 批量删除MVP对话表请求
type ConversationBatchDeleteReq struct {
	g.Meta `path:"/conversation/batch-delete" method:"delete" tags:"MVP对话表" summary:"批量删除MVP对话表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP对话表ID列表"`
}

// ConversationBatchDeleteRes 批量删除MVP对话表响应
type ConversationBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ConversationBatchUpdateReq 批量编辑MVP对话表请求
type ConversationBatchUpdateReq struct {
	g.Meta `path:"/conversation/batch-update" method:"put" tags:"MVP对话表" summary:"批量编辑MVP对话表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP对话表ID列表"`
	Status *string               `json:"status" dc:"状态"`
}

// ConversationBatchUpdateRes 批量编辑MVP对话表响应
type ConversationBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ConversationDetailReq 获取MVP对话表详情请求
type ConversationDetailReq struct {
	g.Meta `path:"/conversation/detail" method:"get" tags:"MVP对话表" summary:"获取MVP对话表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP对话表ID"`
}

// ConversationDetailRes 获取MVP对话表详情响应
type ConversationDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ConversationDetailOutput
}

// ConversationListReq 获取MVP对话表列表请求
type ConversationListReq struct {
	g.Meta    `path:"/conversation/list" method:"get" tags:"MVP对话表" summary:"获取MVP对话表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Title     string `json:"title" dc:"对话标题"`
	ProjectID int64  `json:"projectID" dc:"项目ID"`
	RoleType  string `json:"roleType" dc:"角色类型"`
}

// ConversationListRes 获取MVP对话表列表响应
type ConversationListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ConversationListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// ConversationExportReq 导出MVP对话表请求
type ConversationExportReq struct {
	g.Meta    `path:"/conversation/export" method:"get" tags:"MVP对话表" summary:"导出MVP对话表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Title string `json:"title" dc:"对话标题"`
}

// ConversationExportRes 导出MVP对话表响应
type ConversationExportRes struct {
	g.Meta `mime:"text/csv"`
}



// ConversationImportReq 导入MVP对话表请求
type ConversationImportReq struct {
	g.Meta `path:"/conversation/import" method:"post" mime:"multipart/form-data" tags:"MVP对话表" summary:"导入MVP对话表"`
}

// ConversationImportRes 导入MVP对话表响应
type ConversationImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ConversationImportTemplateReq 下载MVP对话表导入模板
type ConversationImportTemplateReq struct {
	g.Meta `path:"/conversation/import-template" method:"get" tags:"MVP对话表" summary:"下载MVP对话表导入模板"`
}

// ConversationImportTemplateRes 下载MVP对话表导入模板响应
type ConversationImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

