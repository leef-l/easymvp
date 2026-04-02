package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/{{.AppName}}/internal/model"
	"easymvp/utility/snowflake"
)

// 确保 gtime 被引用
var _ = gtime.New

// {{.ModelName}} API

// {{.ModelName}}CreateReq 创建{{.Comment}}请求
type {{.ModelName}}CreateReq struct {
	g.Meta `path:"/{{.ModuleName}}/create" method:"post" tags:"{{.Comment}}" summary:"创建{{.Comment}}"`
{{- range .Fields}}
{{- if and (not .IsID) (not .IsHidden)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}" {{if .ValidationRules}}v:"{{range $i, $v := .ValidationRules}}{{if $i}}|{{end}}{{$v}}{{end}}"{{end}} dc:"{{.Label}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}CreateRes 创建{{.Comment}}响应
type {{.ModelName}}CreateRes struct {
	g.Meta `mime:"application/json"`
}

// {{.ModelName}}UpdateReq 更新{{.Comment}}请求
type {{.ModelName}}UpdateReq struct {
	g.Meta `path:"/{{.ModuleName}}/update" method:"put" tags:"{{.Comment}}" summary:"更新{{.Comment}}"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"{{.Comment}}ID"`
{{- range .Fields}}
{{- if and (not .IsID) (not .IsHidden)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}UpdateRes 更新{{.Comment}}响应
type {{.ModelName}}UpdateRes struct {
	g.Meta `mime:"application/json"`
}

// {{.ModelName}}DeleteReq 删除{{.Comment}}请求
type {{.ModelName}}DeleteReq struct {
	g.Meta `path:"/{{.ModuleName}}/delete" method:"delete" tags:"{{.Comment}}" summary:"删除{{.Comment}}"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"{{.Comment}}ID"`
}

// {{.ModelName}}DeleteRes 删除{{.Comment}}响应
type {{.ModelName}}DeleteRes struct {
	g.Meta `mime:"application/json"`
}

// {{.ModelName}}BatchDeleteReq 批量删除{{.Comment}}请求
type {{.ModelName}}BatchDeleteReq struct {
	g.Meta `path:"/{{.ModuleName}}/batch-delete" method:"delete" tags:"{{.Comment}}" summary:"批量删除{{.Comment}}"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"{{.Comment}}ID列表"`
}

// {{.ModelName}}BatchDeleteRes 批量删除{{.Comment}}响应
type {{.ModelName}}BatchDeleteRes struct {
	g.Meta `mime:"application/json"`
}
{{if .HasBatchEdit}}
// {{.ModelName}}BatchUpdateReq 批量编辑{{.Comment}}请求
type {{.ModelName}}BatchUpdateReq struct {
	g.Meta `path:"/{{.ModuleName}}/batch-update" method:"put" tags:"{{.Comment}}" summary:"批量编辑{{.Comment}}"`
	IDs    []snowflake.JsonInt64 `json:"ids" v:"required#ID列表不能为空" dc:"{{.Comment}}ID列表"`
	Status *int                  `json:"status" dc:"状态"`
}

// {{.ModelName}}BatchUpdateRes 批量编辑{{.Comment}}响应
type {{.ModelName}}BatchUpdateRes struct {
	g.Meta `mime:"application/json"`
}
{{end}}
// {{.ModelName}}DetailReq 获取{{.Comment}}详情请求
type {{.ModelName}}DetailReq struct {
	g.Meta `path:"/{{.ModuleName}}/detail" method:"get" tags:"{{.Comment}}" summary:"获取{{.Comment}}详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"{{.Comment}}ID"`
}

// {{.ModelName}}DetailRes 获取{{.Comment}}详情响应
type {{.ModelName}}DetailRes struct {
	g.Meta `mime:"application/json"`
	*model.{{.ModelName}}DetailOutput
}

// {{.ModelName}}ListReq 获取{{.Comment}}列表请求
type {{.ModelName}}ListReq struct {
	g.Meta    `path:"/{{.ModuleName}}/list" method:"get" tags:"{{.Comment}}" summary:"获取{{.Comment}}列表"`
	PageNum   int    `json:"pageNum" d:"1" dc:"页码"`
	PageSize  int    `json:"pageSize" d:"10" dc:"每页数量"`
	OrderBy   string `json:"orderBy" dc:"排序字段"`
	OrderDir  string `json:"orderDir" d:"asc" dc:"排序方向:asc/desc"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum)}}
	{{.NameCamel}} *int `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	{{.NameCamel}} string `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}ListRes 获取{{.Comment}}列表响应
type {{.ModelName}}ListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.{{.ModelName}}ListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}
// {{.ModelName}}ExportReq 导出{{.Comment}}请求
type {{.ModelName}}ExportReq struct {
	g.Meta    `path:"/{{.ModuleName}}/export" method:"get" tags:"{{.Comment}}" summary:"导出{{.Comment}}"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum)}}
	{{.NameCamel}} *int `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	{{.NameCamel}} string `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}ExportRes 导出{{.Comment}}响应
type {{.ModelName}}ExportRes struct {
	g.Meta `mime:"text/csv"`
}

{{if .HasParentID}}
// {{.ModelName}}TreeReq 获取{{.Comment}}树形结构请求
type {{.ModelName}}TreeReq struct {
	g.Meta    `path:"/{{.ModuleName}}/tree" method:"get" tags:"{{.Comment}}" summary:"获取{{.Comment}}树形结构"`
	StartTime string `json:"startTime" dc:"开始时间"`
	EndTime   string `json:"endTime" dc:"结束时间"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsParentID) (.IsEnum)}}
	{{.NameCamel}} *int `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	{{.NameCamel}} string `json:"{{.NameLower}}" dc:"{{.Label}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}TreeRes 获取{{.Comment}}树形结构响应
type {{.ModelName}}TreeRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.{{.ModelName}}TreeOutput `json:"list" dc:"树形数据"`
}
{{end}}
{{if .HasImport}}
// {{.ModelName}}ImportReq 导入{{.Comment}}请求
type {{.ModelName}}ImportReq struct {
	g.Meta `path:"/{{.ModuleName}}/import" method:"post" mime:"multipart/form-data" tags:"{{.Comment}}" summary:"导入{{.Comment}}"`
}

// {{.ModelName}}ImportRes 导入{{.Comment}}响应
type {{.ModelName}}ImportRes struct {
	g.Meta  `mime:"application/json"`
	Success int `json:"success" dc:"成功条数"`
	Fail    int `json:"fail" dc:"失败条数"`
}

// {{.ModelName}}ImportTemplateReq 下载{{.Comment}}导入模板
type {{.ModelName}}ImportTemplateReq struct {
	g.Meta `path:"/{{.ModuleName}}/import-template" method:"get" tags:"{{.Comment}}" summary:"下载{{.Comment}}导入模板"`
}

// {{.ModelName}}ImportTemplateRes 下载{{.Comment}}导入模板响应
type {{.ModelName}}ImportTemplateRes struct {
	g.Meta `mime:"text/csv"`
}
{{end}}
