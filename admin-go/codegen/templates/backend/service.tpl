package service

import (
	"context"
	"easymvp/app/{{.AppName}}/internal/model"
	"easymvp/utility/snowflake"
{{- if .HasImport}}
	"github.com/gogf/gf/v2/net/ghttp"
{{- end}}
)

type I{{.ModelName}} interface {
	Create(ctx context.Context, in *model.{{.ModelName}}CreateInput) error
	Update(ctx context.Context, in *model.{{.ModelName}}UpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.{{.ModelName}}DetailOutput, err error)
	List(ctx context.Context, in *model.{{.ModelName}}ListInput) (list []*model.{{.ModelName}}ListOutput, total int, err error)
	Export(ctx context.Context, in *model.{{.ModelName}}ListInput) (list []*model.{{.ModelName}}ListOutput, err error)
{{- if .HasParentID}}
	Tree(ctx context.Context, in *model.{{.ModelName}}TreeInput) (tree []*model.{{.ModelName}}TreeOutput, err error)
{{- end}}
{{- if .HasBatchEdit}}
	BatchUpdate(ctx context.Context, in *model.{{.ModelName}}BatchUpdateInput) error
{{- end}}
{{- if .HasImport}}
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
{{- end}}
}

var local{{.ModelName}} I{{.ModelName}}

func {{.ModelName}}() I{{.ModelName}} {
	return local{{.ModelName}}
}

func Register{{.ModelName}}(i I{{.ModelName}}) {
	local{{.ModelName}} = i
}
