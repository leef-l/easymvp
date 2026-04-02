package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Plan API

// PlanCreateReq 创建AI套餐表请求
type PlanCreateReq struct {
	g.Meta `path:"/plan/create" method:"post" tags:"AI套餐表" summary:"创建AI套餐表"`
	ProviderID snowflake.JsonInt64 `json:"providerID" v:"required" dc:"供应商ID"`
	Name string `json:"name" v:"required|max-length:100" dc:"套餐名称"`
	Code string `json:"code" v:"required|max-length:50" dc:"套餐代码"`
	ApiKey string `json:"apiKey" v:"max-length:500" dc:"API Key（加密存储）"`
	ApiSecret string `json:"apiSecret" v:"max-length:500" dc:"API Secret（部分供应商需要）"`
	Status int `json:"status"  dc:"状态"`
	Sort int `json:"sort"  dc:"排序"`
}

// PlanCreateRes 创建AI套餐表响应
type PlanCreateRes struct {
	g.Meta `mime:"application/json"`
}

// PlanUpdateReq 更新AI套餐表请求
type PlanUpdateReq struct {
	g.Meta `path:"/plan/update" method:"put" tags:"AI套餐表" summary:"更新AI套餐表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI套餐表ID"`
	ProviderID snowflake.JsonInt64 `json:"providerID" dc:"供应商ID"`
	Name string `json:"name" dc:"套餐名称"`
	Code string `json:"code" dc:"套餐代码"`
	ApiKey string `json:"apiKey" dc:"API Key（加密存储）"`
	ApiSecret string `json:"apiSecret" dc:"API Secret（部分供应商需要）"`
	Status int `json:"status" dc:"状态"`
	Sort int `json:"sort" dc:"排序"`
}

// PlanUpdateRes 更新AI套餐表响应
type PlanUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// PlanDeleteReq 删除AI套餐表请求
type PlanDeleteReq struct {
	g.Meta `path:"/plan/delete" method:"delete" tags:"AI套餐表" summary:"删除AI套餐表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI套餐表ID"`
}

// PlanDeleteRes 删除AI套餐表响应
type PlanDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// PlanBatchDeleteReq 批量删除AI套餐表请求
type PlanBatchDeleteReq struct {
	g.Meta `path:"/plan/batch-delete" method:"delete" tags:"AI套餐表" summary:"批量删除AI套餐表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"AI套餐表ID列表"`
}

// PlanBatchDeleteRes 批量删除AI套餐表响应
type PlanBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// PlanBatchUpdateReq 批量编辑AI套餐表请求
type PlanBatchUpdateReq struct {
	g.Meta `path:"/plan/batch-update" method:"put" tags:"AI套餐表" summary:"批量编辑AI套餐表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"AI套餐表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// PlanBatchUpdateRes 批量编辑AI套餐表响应
type PlanBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// PlanDetailReq 获取AI套餐表详情请求
type PlanDetailReq struct {
	g.Meta `path:"/plan/detail" method:"get" tags:"AI套餐表" summary:"获取AI套餐表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"AI套餐表ID"`
}

// PlanDetailRes 获取AI套餐表详情响应
type PlanDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.PlanDetailOutput
}

// PlanListReq 获取AI套餐表列表请求
type PlanListReq struct {
	g.Meta    `path:"/plan/list" method:"get" tags:"AI套餐表" summary:"获取AI套餐表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
	Name string `json:"name" dc:"套餐名称"`
}

// PlanListRes 获取AI套餐表列表响应
type PlanListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.PlanListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// PlanExportReq 导出AI套餐表请求
type PlanExportReq struct {
	g.Meta    `path:"/plan/export" method:"get" tags:"AI套餐表" summary:"导出AI套餐表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
	Name string `json:"name" dc:"套餐名称"`
}

// PlanExportRes 导出AI套餐表响应
type PlanExportRes struct {
	g.Meta `mime:"text/csv"`
}



// PlanImportReq 导入AI套餐表请求
type PlanImportReq struct {
	g.Meta `path:"/plan/import" method:"post" mime:"multipart/form-data" tags:"AI套餐表" summary:"导入AI套餐表"`
}

// PlanImportRes 导入AI套餐表响应
type PlanImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// PlanImportTemplateReq 下载AI套餐表导入模板
type PlanImportTemplateReq struct {
	g.Meta `path:"/plan/import-template" method:"get" tags:"AI套餐表" summary:"下载AI套餐表导入模板"`
}

// PlanImportTemplateRes 下载AI套餐表导入模板响应
type PlanImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

