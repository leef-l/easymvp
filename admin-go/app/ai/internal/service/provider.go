package service

import (
	"context"
	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IProvider interface {
	Create(ctx context.Context, in *model.ProviderCreateInput) error
	Update(ctx context.Context, in *model.ProviderUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProviderDetailOutput, err error)
	List(ctx context.Context, in *model.ProviderListInput) (list []*model.ProviderListOutput, total int, err error)
	Export(ctx context.Context, in *model.ProviderListInput) (list []*model.ProviderListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.ProviderBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localProvider IProvider

func Provider() IProvider {
	return localProvider
}

func RegisterProvider(i IProvider) {
	localProvider = i
}
