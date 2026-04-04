package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// ProjectRole API

// ProjectRoleCreateReq 创建项目角色配置表请求
type ProjectRoleCreateReq struct {
	g.Meta `path:"/project_role/create" method:"post" tags:"项目角色配置表" summary:"创建项目角色配置表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	ProjectCategory string `json:"projectCategory" v:"required|max-length:50" dc:"项目分类"`
	RoleType string `json:"roleType" v:"required|max-length:20" dc:"角色类型"`
	RoleLevel string `json:"roleLevel" v:"max-length:10" dc:"角色等级"`
	ModelID snowflake.JsonInt64 `json:"modelID" v:"required" dc:"AI模型ID"`
	SystemPrompt string `json:"systemPrompt" v:"max-length:65535" dc:"系统提示词（角色设定）"`
	ExecutionMode string `json:"executionMode" v:"max-length:20" dc:"执行方式: chat/aider/openhands"`
	Status int `json:"status"  dc:"状态"`
}

// ProjectRoleCreateRes 创建项目角色配置表响应
type ProjectRoleCreateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectRoleUpdateReq 更新项目角色配置表请求
type ProjectRoleUpdateReq struct {
	g.Meta `path:"/project_role/update" method:"put" tags:"项目角色配置表" summary:"更新项目角色配置表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"项目角色配置表ID"`
	ProjectID snowflake.JsonInt64 `json:"projectID" dc:"项目ID"`
	ProjectCategory string `json:"projectCategory" dc:"项目分类"`
	RoleType string `json:"roleType" dc:"角色类型"`
	RoleLevel string `json:"roleLevel" dc:"角色等级"`
	ModelID snowflake.JsonInt64 `json:"modelID" dc:"AI模型ID"`
	SystemPrompt string `json:"systemPrompt" dc:"系统提示词（角色设定）"`
	ExecutionMode string `json:"executionMode" dc:"执行方式: chat/aider/openhands"`
	Status int `json:"status" dc:"状态"`
}

// ProjectRoleUpdateRes 更新项目角色配置表响应
type ProjectRoleUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectRoleDeleteReq 删除项目角色配置表请求
type ProjectRoleDeleteReq struct {
	g.Meta `path:"/project_role/delete" method:"delete" tags:"项目角色配置表" summary:"删除项目角色配置表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"项目角色配置表ID"`
}

// ProjectRoleDeleteRes 删除项目角色配置表响应
type ProjectRoleDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectRoleBatchDeleteReq 批量删除项目角色配置表请求
type ProjectRoleBatchDeleteReq struct {
	g.Meta `path:"/project_role/batch-delete" method:"delete" tags:"项目角色配置表" summary:"批量删除项目角色配置表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"项目角色配置表ID列表"`
}

// ProjectRoleBatchDeleteRes 批量删除项目角色配置表响应
type ProjectRoleBatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectRoleBatchUpdateReq 批量编辑项目角色配置表请求
type ProjectRoleBatchUpdateReq struct {
	g.Meta `path:"/project_role/batch-update" method:"put" tags:"项目角色配置表" summary:"批量编辑项目角色配置表"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"项目角色配置表ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// ProjectRoleBatchUpdateRes 批量编辑项目角色配置表响应
type ProjectRoleBatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// ProjectRoleDetailReq 获取项目角色配置表详情请求
type ProjectRoleDetailReq struct {
	g.Meta `path:"/project_role/detail" method:"get" tags:"项目角色配置表" summary:"获取项目角色配置表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"项目角色配置表ID"`
}

// ProjectRoleDetailRes 获取项目角色配置表详情响应
type ProjectRoleDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.ProjectRoleDetailOutput
}

// ProjectRoleListReq 获取项目角色配置表列表请求
type ProjectRoleListReq struct {
	g.Meta    `path:"/project_role/list" method:"get" tags:"项目角色配置表" summary:"获取项目角色配置表列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
}

// ProjectRoleListRes 获取项目角色配置表列表响应
type ProjectRoleListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.ProjectRoleListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// ProjectRoleExportReq 导出项目角色配置表请求
type ProjectRoleExportReq struct {
	g.Meta    `path:"/project_role/export" method:"get" tags:"项目角色配置表" summary:"导出项目角色配置表"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
	Status *int `json:"status" dc:"状态"`
}

// ProjectRoleExportRes 导出项目角色配置表响应
type ProjectRoleExportRes struct {
	g.Meta `mime:"text/csv"`
}



// ProjectRoleImportReq 导入项目角色配置表请求
type ProjectRoleImportReq struct {
	g.Meta `path:"/project_role/import" method:"post" mime:"multipart/form-data" tags:"项目角色配置表" summary:"导入项目角色配置表"`
}

// ProjectRoleImportRes 导入项目角色配置表响应
type ProjectRoleImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// ProjectRoleImportTemplateReq 下载项目角色配置表导入模板
type ProjectRoleImportTemplateReq struct {
	g.Meta `path:"/project_role/import-template" method:"get" tags:"项目角色配置表" summary:"下载项目角色配置表导入模板"`
}

// ProjectRoleImportTemplateRes 下载项目角色配置表导入模板响应
type ProjectRoleImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}

