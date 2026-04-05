package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IProjectCategory interface {
	Create(ctx context.Context, in *model.ProjectCategoryCreateInput) error
	Update(ctx context.Context, in *model.ProjectCategoryUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProjectCategoryDetailOutput, err error)
	List(ctx context.Context, in *model.ProjectCategoryListInput) (list []*model.ProjectCategoryListOutput, total int, err error)
	Export(ctx context.Context, in *model.ProjectCategoryListInput) (list []*model.ProjectCategoryListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.ProjectCategoryBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localProjectCategory IProjectCategory

func ProjectCategory() IProjectCategory {
	return localProjectCategory
}

func RegisterProjectCategory(i IProjectCategory) {
	localProjectCategory = i
}
