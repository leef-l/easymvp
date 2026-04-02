package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// Project API

// ProjectCreateReq 创建MVP项目表请求
type ProjectCreateReq struct {
	g.Meta `path:"/project/create" method:"post" tags:"MVP项目表" summary:"创建MVP项目表"`
	Name string `json:"name" v:"required|max-length:200" dc:"项目名称"`
	Description string `json:"description" v:"max-length:65535" dc:"项目简介"`
	Status string `json:"status" v:"max-length:20" dc:"状态"`
	PauseReason string `json:"pauseReason" v:"max-length:65535" dc:"暂停原因"`
	GlobalContext string `json:"globalContext" v:"max-length:65535" dc:"项目全局上下文（架构师需求分析+方案设计的压缩摘要）"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID"  dc:"架构师使用的AI模型ID"`
}

// ProjectCreateRes 创建MVP项目表响应
type ProjectCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectUpdateReq 更新MVP项目表请求
type ProjectUpdateReq struct {
	g.Meta `path:"/project/update" method:"put" tags:"MVP项目表" summary:"更新MVP项目表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP项目表ID"`
	Name string `json:"name" dc:"项目名称"`
	Description string `json:"description" dc:"项目简介"`
	Status string `json:"status" dc:"状态"`
	PauseReason string `json:"pauseReason" dc:"暂停原因"`
	GlobalContext string `json:"globalContext" dc:"项目全局上下文（架构师需求分析+方案设计的压缩摘要）"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID" dc:"架构师使用的AI模型ID"`
}

// ProjectUpdateRes 更新MVP项目表响应
type ProjectUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectDeleteReq 删除MVP项目表请求
type ProjectDeleteReq struct {
	g.Meta `path:"/project/delete" method:"delete" tags:"MVP项目表" summary:"删除MVP项目表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP项目表ID"`
}

// ProjectDeleteRes 删除MVP项目表响应
type ProjectDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectBatchDeleteReq 批量删除MVP项目表请求
type ProjectBatchDeleteReq struct {
	g.Meta `path:"/project/batch-delete" method:"delete" tags:"MVP项目表" summary:"批量删除MVP项目表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP项目表ID列表"`
}

// ProjectBatchDeleteRes 批量删除MVP项目表响应
type ProjectBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectBatchUpdateReq 批量编辑MVP项目表请求
type ProjectBatchUpdateReq struct {
	g.Meta `path:"/project/batch-update" method:"put" tags:"MVP项目表" summary:"批量编辑MVP项目表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"MVP项目表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// ProjectBatchUpdateRes 批量编辑MVP项目表响应
type ProjectBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectDetailReq 获取MVP项目表详情请求
type ProjectDetailReq struct {
	g.Meta `path:"/project/detail" method:"get" tags:"MVP项目表" summary:"获取MVP项目表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"MVP项目表ID"`
}

// ProjectDetailRes 获取MVP项目表详情响应
type ProjectDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ProjectDetailOutput
}

// ProjectListReq 获取MVP项目表列表请求
type ProjectListReq struct {
	g.Meta    `path:"/project/list" method:"get" tags:"MVP项目表" summary:"获取MVP项目表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Name string `json:"name" dc:"项目名称"`
}

// ProjectListRes 获取MVP项目表列表响应
type ProjectListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ProjectListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// ProjectExportReq 导出MVP项目表请求
type ProjectExportReq struct {
	g.Meta    `path:"/project/export" method:"get" tags:"MVP项目表" summary:"导出MVP项目表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Name string `json:"name" dc:"项目名称"`
}

// ProjectExportRes 导出MVP项目表响应
type ProjectExportRes struct {
	g.Meta `mime:"text/csv"`
}



// ProjectImportReq 导入MVP项目表请求
type ProjectImportReq struct {
	g.Meta `path:"/project/import" method:"post" mime:"multipart/form-data" tags:"MVP项目表" summary:"导入MVP项目表"`
}

// ProjectImportRes 导入MVP项目表响应
type ProjectImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ProjectImportTemplateReq 下载MVP项目表导入模板
type ProjectImportTemplateReq struct {
	g.Meta `path:"/project/import-template" method:"get" tags:"MVP项目表" summary:"下载MVP项目表导入模板"`
}

// ProjectImportTemplateRes 下载MVP项目表导入模板响应
type ProjectImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

