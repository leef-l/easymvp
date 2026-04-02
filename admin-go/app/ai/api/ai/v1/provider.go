package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Provider API

// ProviderCreateReq 创建AI供应商表请求
type ProviderCreateReq struct {
	g.Meta `path:"/provider/create" method:"post" tags:"AI供应商表" summary:"创建AI供应商表"`
	Name string `json:"name" v:"required|max-length:100" dc:"供应商名称"`
	Code string `json:"code" v:"required|max-length:50" dc:"供应商代码"`
	ProviderType string `json:"providerType" v:"required|max-length:20" dc:"Provider类型"`
	BaseURL string `json:"baseURL" v:"required|url|max-length:500" dc:"API基础地址"`
	Icon string `json:"icon" v:"max-length:500" dc:"图标URL"`
	Status int `json:"status"  dc:"状态"`
	Sort int `json:"sort"  dc:"排序"`
}

// ProviderCreateRes 创建AI供应商表响应
type ProviderCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ProviderUpdateReq 更新AI供应商表请求
type ProviderUpdateReq struct {
	g.Meta `path:"/provider/update" method:"put" tags:"AI供应商表" summary:"更新AI供应商表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI供应商表ID"`
	Name string `json:"name" dc:"供应商名称"`
	Code string `json:"code" dc:"供应商代码"`
	ProviderType string `json:"providerType" dc:"Provider类型"`
	BaseURL string `json:"baseURL" dc:"API基础地址"`
	Icon string `json:"icon" dc:"图标URL"`
	Status int `json:"status" dc:"状态"`
	Sort int `json:"sort" dc:"排序"`
}

// ProviderUpdateRes 更新AI供应商表响应
type ProviderUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProviderDeleteReq 删除AI供应商表请求
type ProviderDeleteReq struct {
	g.Meta `path:"/provider/delete" method:"delete" tags:"AI供应商表" summary:"删除AI供应商表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI供应商表ID"`
}

// ProviderDeleteRes 删除AI供应商表响应
type ProviderDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProviderBatchDeleteReq 批量删除AI供应商表请求
type ProviderBatchDeleteReq struct {
	g.Meta `path:"/provider/batch-delete" method:"delete" tags:"AI供应商表" summary:"批量删除AI供应商表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"AI供应商表ID列表"`
}

// ProviderBatchDeleteRes 批量删除AI供应商表响应
type ProviderBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProviderBatchUpdateReq 批量编辑AI供应商表请求
type ProviderBatchUpdateReq struct {
	g.Meta `path:"/provider/batch-update" method:"put" tags:"AI供应商表" summary:"批量编辑AI供应商表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"AI供应商表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// ProviderBatchUpdateRes 批量编辑AI供应商表响应
type ProviderBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProviderDetailReq 获取AI供应商表详情请求
type ProviderDetailReq struct {
	g.Meta `path:"/provider/detail" method:"get" tags:"AI供应商表" summary:"获取AI供应商表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI供应商表ID"`
}

// ProviderDetailRes 获取AI供应商表详情响应
type ProviderDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ProviderDetailOutput
}

// ProviderListReq 获取AI供应商表列表请求
type ProviderListReq struct {
	g.Meta    `path:"/provider/list" method:"get" tags:"AI供应商表" summary:"获取AI供应商表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
	Name string `json:"name" dc:"供应商名称"`
}

// ProviderListRes 获取AI供应商表列表响应
type ProviderListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ProviderListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// ProviderExportReq 导出AI供应商表请求
type ProviderExportReq struct {
	g.Meta    `path:"/provider/export" method:"get" tags:"AI供应商表" summary:"导出AI供应商表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
	Name string `json:"name" dc:"供应商名称"`
}

// ProviderExportRes 导出AI供应商表响应
type ProviderExportRes struct {
	g.Meta `mime:"text/csv"`
}



// ProviderImportReq 导入AI供应商表请求
type ProviderImportReq struct {
	g.Meta `path:"/provider/import" method:"post" mime:"multipart/form-data" tags:"AI供应商表" summary:"导入AI供应商表"`
}

// ProviderImportRes 导入AI供应商表响应
type ProviderImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ProviderImportTemplateReq 下载AI供应商表导入模板
type ProviderImportTemplateReq struct {
	g.Meta `path:"/provider/import-template" method:"get" tags:"AI供应商表" summary:"下载AI供应商表导入模板"`
}

// ProviderImportTemplateRes 下载AI供应商表导入模板响应
type ProviderImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

