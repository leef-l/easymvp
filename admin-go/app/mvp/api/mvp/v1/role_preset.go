package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// RolePreset API

// RolePresetCreateReq 创建角色预设模板请求
type RolePresetCreateReq struct {
	g.Meta `path:"/role_preset/create" method:"post" tags:"角色预设模板" summary:"创建角色预设模板"`
	ProjectCategory string `json:"projectCategory" v:"required|max-length:50" dc:"项目分类"`
	RoleType string `json:"roleType" v:"required|max-length:20" dc:"角色类型"`
	RoleLevel string `json:"roleLevel" v:"max-length:10" dc:"角色等级"`
	ModelID snowflake.JsonInt64 `json:"modelID" v:"required" dc:"AI模型ID"`
	SystemPrompt string `json:"systemPrompt" v:"max-length:65535" dc:"默认系统提示词（角色设定）"`
	Status int `json:"status"  dc:"状态"`
	Sort int `json:"sort"  dc:"排序"`
}

// RolePresetCreateRes 创建角色预设模板响应
type RolePresetCreateRes struct {
	g.Meta `mime:"application/json"`
}

// RolePresetUpdateReq 更新角色预设模板请求
type RolePresetUpdateReq struct {
	g.Meta `path:"/role_preset/update" method:"put" tags:"角色预设模板" summary:"更新角色预设模板"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"角色预设模板ID"`
	ProjectCategory string `json:"projectCategory" dc:"项目分类"`
	RoleType string `json:"roleType" dc:"角色类型"`
	RoleLevel string `json:"roleLevel" dc:"角色等级"`
	ModelID snowflake.JsonInt64 `json:"modelID" dc:"AI模型ID"`
	SystemPrompt string `json:"systemPrompt" dc:"默认系统提示词（角色设定）"`
	Status int `json:"status" dc:"状态"`
	Sort int `json:"sort" dc:"排序"`
}

// RolePresetUpdateRes 更新角色预设模板响应
type RolePresetUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// RolePresetDeleteReq 删除角色预设模板请求
type RolePresetDeleteReq struct {
	g.Meta `path:"/role_preset/delete" method:"delete" tags:"角色预设模板" summary:"删除角色预设模板"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"角色预设模板ID"`
}

// RolePresetDeleteRes 删除角色预设模板响应
type RolePresetDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// RolePresetBatchDeleteReq 批量删除角色预设模板请求
type RolePresetBatchDeleteReq struct {
	g.Meta `path:"/role_preset/batch-delete" method:"delete" tags:"角色预设模板" summary:"批量删除角色预设模板"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"角色预设模板ID列表"`
}

// RolePresetBatchDeleteRes 批量删除角色预设模板响应
type RolePresetBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// RolePresetBatchUpdateReq 批量编辑角色预设模板请求
type RolePresetBatchUpdateReq struct {
	g.Meta `path:"/role_preset/batch-update" method:"put" tags:"角色预设模板" summary:"批量编辑角色预设模板"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"角色预设模板ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// RolePresetBatchUpdateRes 批量编辑角色预设模板响应
type RolePresetBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// RolePresetDetailReq 获取角色预设模板详情请求
type RolePresetDetailReq struct {
	g.Meta `path:"/role_preset/detail" method:"get" tags:"角色预设模板" summary:"获取角色预设模板详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"角色预设模板ID"`
}

// RolePresetDetailRes 获取角色预设模板详情响应
type RolePresetDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.RolePresetDetailOutput
}

// RolePresetListReq 获取角色预设模板列表请求
type RolePresetListReq struct {
	g.Meta    `path:"/role_preset/list" method:"get" tags:"角色预设模板" summary:"获取角色预设模板列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
}

// RolePresetListRes 获取角色预设模板列表响应
type RolePresetListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.RolePresetListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// RolePresetExportReq 导出角色预设模板请求
type RolePresetExportReq struct {
	g.Meta    `path:"/role_preset/export" method:"get" tags:"角色预设模板" summary:"导出角色预设模板"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
}

// RolePresetExportRes 导出角色预设模板响应
type RolePresetExportRes struct {
	g.Meta `mime:"text/csv"`
}



// RolePresetImportReq 导入角色预设模板请求
type RolePresetImportReq struct {
	g.Meta `path:"/role_preset/import" method:"post" mime:"multipart/form-data" tags:"角色预设模板" summary:"导入角色预设模板"`
}

// RolePresetImportRes 导入角色预设模板响应
type RolePresetImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// RolePresetImportTemplateReq 下载角色预设模板导入模板
type RolePresetImportTemplateReq struct {
	g.Meta `path:"/role_preset/import-template" method:"get" tags:"角色预设模板" summary:"下载角色预设模板导入模板"`
}

// RolePresetImportTemplateRes 下载角色预设模板导入模板响应
type RolePresetImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

