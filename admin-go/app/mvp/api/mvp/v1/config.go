package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Config API

// ConfigCreateReq 创建MVP配置表请求
type ConfigCreateReq struct {
	g.Meta `path:"/config/create" method:"post" tags:"MVP配置表" summary:"创建MVP配置表"`
	ConfigKey string `json:"configKey" v:"required|max-length:100" dc:"配置键（唯一）"`
	ConfigValue string `json:"configValue" v:"required|max-length:65535" dc:"配置值"`
	ConfigType string `json:"configType" v:"max-length:20" dc:"值类型"`
	Category string `json:"category" v:"max-length:50" dc:"分类"`
	Description string `json:"description" v:"max-length:255" dc:"配置说明"`
}

// ConfigCreateRes 创建MVP配置表响应
type ConfigCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ConfigUpdateReq 更新MVP配置表请求
type ConfigUpdateReq struct {
	g.Meta `path:"/config/update" method:"put" tags:"MVP配置表" summary:"更新MVP配置表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP配置表ID"`
	ConfigKey string `json:"configKey" dc:"配置键（唯一）"`
	ConfigValue string `json:"configValue" dc:"配置值"`
	ConfigType string `json:"configType" dc:"值类型"`
	Category string `json:"category" dc:"分类"`
	Description string `json:"description" dc:"配置说明"`
}

// ConfigUpdateRes 更新MVP配置表响应
type ConfigUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ConfigDeleteReq 删除MVP配置表请求
type ConfigDeleteReq struct {
	g.Meta `path:"/config/delete" method:"delete" tags:"MVP配置表" summary:"删除MVP配置表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP配置表ID"`
}

// ConfigDeleteRes 删除MVP配置表响应
type ConfigDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ConfigBatchDeleteReq 批量删除MVP配置表请求
type ConfigBatchDeleteReq struct {
	g.Meta `path:"/config/batch-delete" method:"delete" tags:"MVP配置表" summary:"批量删除MVP配置表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP配置表ID列表"`
}

// ConfigBatchDeleteRes 批量删除MVP配置表响应
type ConfigBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ConfigDetailReq 获取MVP配置表详情请求
type ConfigDetailReq struct {
	g.Meta `path:"/config/detail" method:"get" tags:"MVP配置表" summary:"获取MVP配置表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP配置表ID"`
}

// ConfigDetailRes 获取MVP配置表详情响应
type ConfigDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ConfigDetailOutput
}

// ConfigListReq 获取MVP配置表列表请求
type ConfigListReq struct {
	g.Meta    `path:"/config/list" method:"get" tags:"MVP配置表" summary:"获取MVP配置表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
}

// ConfigListRes 获取MVP配置表列表响应
type ConfigListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ConfigListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// ConfigExportReq 导出MVP配置表请求
type ConfigExportReq struct {
	g.Meta    `path:"/config/export" method:"get" tags:"MVP配置表" summary:"导出MVP配置表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
}

// ConfigExportRes 导出MVP配置表响应
type ConfigExportRes struct {
	g.Meta `mime:"text/csv"`
}



// ConfigImportReq 导入MVP配置表请求
type ConfigImportReq struct {
	g.Meta `path:"/config/import" method:"post" mime:"multipart/form-data" tags:"MVP配置表" summary:"导入MVP配置表"`
}

// ConfigImportRes 导入MVP配置表响应
type ConfigImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ConfigImportTemplateReq 下载MVP配置表导入模板
type ConfigImportTemplateReq struct {
	g.Meta `path:"/config/import-template" method:"get" tags:"MVP配置表" summary:"下载MVP配置表导入模板"`
}

// ConfigImportTemplateRes 下载MVP配置表导入模板响应
type ConfigImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

