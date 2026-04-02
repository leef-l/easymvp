package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Model API

// ModelCreateReq 创建AI模型表请求
type ModelCreateReq struct {
	g.Meta `path:"/model/create" method:"post" tags:"AI模型表" summary:"创建AI模型表"`
	PlanID snowflake.JsonInt64 `json:"planID" v:"required" dc:"套餐ID"`
	ProviderID snowflake.JsonInt64 `json:"providerID" v:"required" dc:"供应商ID（冗余便于查询）"`
	Name string `json:"name" v:"required|max-length:100" dc:"模型显示名称"`
	ModelCode string `json:"modelCode" v:"required|max-length:100" dc:"模型代码（API调用用）"`
	Capability string `json:"capability" v:"max-length:20" dc:"能力"`
	MaxTokens int `json:"maxTokens"  dc:"最大输出token"`
	ContextWindow int `json:"contextWindow"  dc:"上下文窗口大小"`
	SupportsStream int `json:"supportsStream"  dc:"是否支持流式输出"`
	RolePrompt string `json:"rolePrompt" v:"max-length:65535" dc:"默认角色提示词"`
	Status int `json:"status"  dc:"状态"`
	Sort int `json:"sort"  dc:"排序"`
}

// ModelCreateRes 创建AI模型表响应
type ModelCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ModelUpdateReq 更新AI模型表请求
type ModelUpdateReq struct {
	g.Meta `path:"/model/update" method:"put" tags:"AI模型表" summary:"更新AI模型表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI模型表ID"`
	PlanID snowflake.JsonInt64 `json:"planID" dc:"套餐ID"`
	ProviderID snowflake.JsonInt64 `json:"providerID" dc:"供应商ID（冗余便于查询）"`
	Name string `json:"name" dc:"模型显示名称"`
	ModelCode string `json:"modelCode" dc:"模型代码（API调用用）"`
	Capability string `json:"capability" dc:"能力"`
	MaxTokens int `json:"maxTokens" dc:"最大输出token"`
	ContextWindow int `json:"contextWindow" dc:"上下文窗口大小"`
	SupportsStream int `json:"supportsStream" dc:"是否支持流式输出"`
	RolePrompt string `json:"rolePrompt" dc:"默认角色提示词"`
	Status int `json:"status" dc:"状态"`
	Sort int `json:"sort" dc:"排序"`
}

// ModelUpdateRes 更新AI模型表响应
type ModelUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ModelDeleteReq 删除AI模型表请求
type ModelDeleteReq struct {
	g.Meta `path:"/model/delete" method:"delete" tags:"AI模型表" summary:"删除AI模型表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI模型表ID"`
}

// ModelDeleteRes 删除AI模型表响应
type ModelDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ModelBatchDeleteReq 批量删除AI模型表请求
type ModelBatchDeleteReq struct {
	g.Meta `path:"/model/batch-delete" method:"delete" tags:"AI模型表" summary:"批量删除AI模型表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"AI模型表ID列表"`
}

// ModelBatchDeleteRes 批量删除AI模型表响应
type ModelBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ModelBatchUpdateReq 批量编辑AI模型表请求
type ModelBatchUpdateReq struct {
	g.Meta `path:"/model/batch-update" method:"put" tags:"AI模型表" summary:"批量编辑AI模型表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"AI模型表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// ModelBatchUpdateRes 批量编辑AI模型表响应
type ModelBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ModelDetailReq 获取AI模型表详情请求
type ModelDetailReq struct {
	g.Meta `path:"/model/detail" method:"get" tags:"AI模型表" summary:"获取AI模型表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI模型表ID"`
}

// ModelDetailRes 获取AI模型表详情响应
type ModelDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ModelDetailOutput
}

// ModelListReq 获取AI模型表列表请求
type ModelListReq struct {
	g.Meta    `path:"/model/list" method:"get" tags:"AI模型表" summary:"获取AI模型表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	SupportsStream *int `json:"supportsStream" dc:"是否支持流式输出"`
	Status *int `json:"status" dc:"状态"`
	Name string `json:"name" dc:"模型显示名称"`
}

// ModelListRes 获取AI模型表列表响应
type ModelListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ModelListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// ModelExportReq 导出AI模型表请求
type ModelExportReq struct {
	g.Meta    `path:"/model/export" method:"get" tags:"AI模型表" summary:"导出AI模型表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	SupportsStream *int `json:"supportsStream" dc:"是否支持流式输出"`
	Status *int `json:"status" dc:"状态"`
	Name string `json:"name" dc:"模型显示名称"`
}

// ModelExportRes 导出AI模型表响应
type ModelExportRes struct {
	g.Meta `mime:"text/csv"`
}



// ModelImportReq 导入AI模型表请求
type ModelImportReq struct {
	g.Meta `path:"/model/import" method:"post" mime:"multipart/form-data" tags:"AI模型表" summary:"导入AI模型表"`
}

// ModelImportRes 导入AI模型表响应
type ModelImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ModelImportTemplateReq 下载AI模型表导入模板
type ModelImportTemplateReq struct {
	g.Meta `path:"/model/import-template" method:"get" tags:"AI模型表" summary:"下载AI模型表导入模板"`
}

// ModelImportTemplateRes 下载AI模型表导入模板响应
type ModelImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

