package service

import (
	"context"
	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IModel interface {
	Create(ctx context.Context, in *model.ModelCreateInput) error
	Update(ctx context.Context, in *model.ModelUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ModelDetailOutput, err error)
	List(ctx context.Context, in *model.ModelListInput) (list []*model.ModelListOutput, total int, err error)
	Export(ctx context.Context, in *model.ModelListInput) (list []*model.ModelListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.ModelBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localModel IModel

func Model() IModel {
	return localModel
}

func RegisterModel(i IModel) {
	localModel = i
}
