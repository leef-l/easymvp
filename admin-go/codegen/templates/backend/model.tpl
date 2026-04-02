package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// {{.ModelName}} DTO 模型

// {{.ModelName}}CreateInput 创建{{.Comment}}输入
type {{.ModelName}}CreateInput struct {
{{- range .Fields}}
{{- if and (not .IsID) (not .IsHidden)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}UpdateInput 更新{{.Comment}}输入
type {{.ModelName}}UpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
{{- range .Fields}}
{{- if and (not .IsID) (not .IsHidden)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}DetailOutput {{.Comment}}详情输出
type {{.ModelName}}DetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsPassword)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}"`
{{- if .RefFieldName}}
	{{.RefFieldName}} string `json:"{{.RefFieldJSON}}"`
{{- end}}
{{- end}}
{{- end}}
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// {{.ModelName}}ListOutput {{.Comment}}列表输出
type {{.ModelName}}ListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsPassword)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}"`
{{- if .RefFieldName}}
	{{.RefFieldName}} string `json:"{{.RefFieldJSON}}"`
{{- end}}
{{- end}}
{{- end}}
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// {{.ModelName}}ListInput {{.Comment}}列表查询输入
type {{.ModelName}}ListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum)}}
	{{.NameCamel}} *int `json:"{{.NameLower}}"`
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	{{.NameCamel}} string `json:"{{.NameLower}}"`
{{- end}}
{{- end}}
}
{{if .HasParentID}}
// {{.ModelName}}TreeInput {{.Comment}}树形查询输入
type {{.ModelName}}TreeInput struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsParentID) (.IsEnum)}}
	{{.NameCamel}} *int `json:"{{.NameLower}}"`
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	{{.NameCamel}} string `json:"{{.NameLower}}"`
{{- end}}
{{- end}}
}

// {{.ModelName}}TreeOutput {{.Comment}}树形输出
type {{.ModelName}}TreeOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsPassword)}}
	{{.NameCamel}} {{if .IsForeignKey}}snowflake.JsonInt64{{else}}{{.GoType}}{{end}} `json:"{{.NameLower}}"`
{{- if .RefFieldName}}
	{{.RefFieldName}} string `json:"{{.RefFieldJSON}}"`
{{- end}}
{{- end}}
{{- end}}
	Children []*{{.ModelName}}TreeOutput `json:"children"`
}
{{end}}
{{if .HasBatchEdit}}
// {{.ModelName}}BatchUpdateInput 批量编辑{{.Comment}}输入
type {{.ModelName}}BatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}
{{end}}
