package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// ProjectCategory API

// ProjectCategoryCreateReq 创建项目分类配置表请求
type ProjectCategoryCreateReq struct {
	g.Meta                  `path:"/project_category/create" method:"post" tags:"项目分类配置表" summary:"创建项目分类配置表"`
	CategoryCode            string `json:"categoryCode" v:"required|max-length:64" dc:"稳定分类编码"`
	DisplayName             string `json:"displayName" v:"required|max-length:64" dc:"展示名称"`
	FamilyCode              string `json:"familyCode" v:"required|max-length:32" dc:"能力家族编码"`
	Description             string `json:"description" v:"max-length:255" dc:"分类说明"`
	VerificationProfileJson string `json:"verificationProfileJson" dc:"分类默认验证配置(JSON)"`
	VerificationGateJson    string `json:"verificationGateJson" dc:"分类验证放行规则(JSON)"`
	Status                  int    `json:"status"  dc:"1启用 0停用"`
	Sort                    int    `json:"sort"  dc:"排序"`
}

// ProjectCategoryCreateRes 创建项目分类配置表响应
type ProjectCategoryCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectCategoryUpdateReq 更新项目分类配置表请求
type ProjectCategoryUpdateReq struct {
	g.Meta                  `path:"/project_category/update" method:"put" tags:"项目分类配置表" summary:"更新项目分类配置表"`
	ID                      snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"项目分类配置表ID"`
	CategoryCode            string              `json:"categoryCode" dc:"稳定分类编码"`
	DisplayName             string              `json:"displayName" dc:"展示名称"`
	FamilyCode              string              `json:"familyCode" dc:"能力家族编码"`
	Description             string              `json:"description" dc:"分类说明"`
	VerificationProfileJson string              `json:"verificationProfileJson" dc:"分类默认验证配置(JSON)"`
	VerificationGateJson    string              `json:"verificationGateJson" dc:"分类验证放行规则(JSON)"`
	Status                  int                 `json:"status" dc:"1启用 0停用"`
	Sort                    int                 `json:"sort" dc:"排序"`
}

// ProjectCategoryUpdateRes 更新项目分类配置表响应
type ProjectCategoryUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectCategoryDeleteReq 删除项目分类配置表请求
type ProjectCategoryDeleteReq struct {
	g.Meta `path:"/project_category/delete" method:"delete" tags:"项目分类配置表" summary:"删除项目分类配置表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"项目分类配置表ID"`
}

// ProjectCategoryDeleteRes 删除项目分类配置表响应
type ProjectCategoryDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectCategoryBatchDeleteReq 批量删除项目分类配置表请求
type ProjectCategoryBatchDeleteReq struct {
	g.Meta `path:"/project_category/batch-delete" method:"delete" tags:"项目分类配置表" summary:"批量删除项目分类配置表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"项目分类配置表ID列表"`
}

// ProjectCategoryBatchDeleteRes 批量删除项目分类配置表响应
type ProjectCategoryBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectCategoryBatchUpdateReq 批量编辑项目分类配置表请求
type ProjectCategoryBatchUpdateReq struct {
	g.Meta `path:"/project_category/batch-update" method:"put" tags:"项目分类配置表" summary:"批量编辑项目分类配置表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"项目分类配置表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// ProjectCategoryBatchUpdateRes 批量编辑项目分类配置表响应
type ProjectCategoryBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectCategoryDetailReq 获取项目分类配置表详情请求
type ProjectCategoryDetailReq struct {
	g.Meta `path:"/project_category/detail" method:"get" tags:"项目分类配置表" summary:"获取项目分类配置表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"项目分类配置表ID"`
}

// ProjectCategoryDetailRes 获取项目分类配置表详情响应
type ProjectCategoryDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ProjectCategoryDetailOutput
}

// ProjectCategoryListReq 获取项目分类配置表列表请求
type ProjectCategoryListReq struct {
	g.Meta      `path:"/project_category/list" method:"get" tags:"项目分类配置表" summary:"获取项目分类配置表列表"`
	PageNum     int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize    int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy     string `json:"orderBy" dc:"排序字段"`
	OrderDir    string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime   string `json:"startTime" dc:"开始时间"`
	EndTime     string `json:"endTime" dc:"结束时间"`
	DisplayName string `json:"displayName" dc:"展示名称"`
}

// ProjectCategoryListRes 获取项目分类配置表列表响应
type ProjectCategoryListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ProjectCategoryListOutput `json:"list" dc:"列表数据"`
	Total  int                                `json:"total" dc:"总数"`
}

// ProjectCategoryExportReq 导出项目分类配置表请求
type ProjectCategoryExportReq struct {
	g.Meta      `path:"/project_category/export" method:"get" tags:"项目分类配置表" summary:"导出项目分类配置表"`
	StartTime   string `json:"startTime" dc:"开始时间"`
	EndTime     string `json:"endTime" dc:"结束时间"`
	DisplayName string `json:"displayName" dc:"展示名称"`
}

// ProjectCategoryExportRes 导出项目分类配置表响应
type ProjectCategoryExportRes struct {
	g.Meta `mime:"text/csv"`
}

// ProjectCategoryImportReq 导入项目分类配置表请求
type ProjectCategoryImportReq struct {
	g.Meta `path:"/project_category/import" method:"post" mime:"multipart/form-data" tags:"项目分类配置表" summary:"导入项目分类配置表"`
}

// ProjectCategoryImportRes 导入项目分类配置表响应
type ProjectCategoryImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ProjectCategoryImportTemplateReq 下载项目分类配置表导入模板
type ProjectCategoryImportTemplateReq struct {
	g.Meta `path:"/project_category/import-template" method:"get" tags:"项目分类配置表" summary:"下载项目分类配置表导入模板"`
}

// ProjectCategoryImportTemplateRes 下载项目分类配置表导入模板响应
type ProjectCategoryImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}
