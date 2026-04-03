package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IConfig interface {
	Create(ctx context.Context, in *model.ConfigCreateInput) error
	Update(ctx context.Context, in *model.ConfigUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ConfigDetailOutput, err error)
	List(ctx context.Context, in *model.ConfigListInput) (list []*model.ConfigListOutput, total int, err error)
	Export(ctx context.Context, in *model.ConfigListInput) (list []*model.ConfigListOutput, err error)
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localConfig IConfig

func Config() IConfig {
	return localConfig
}

func RegisterConfig(i IConfig) {
	localConfig = i
}
