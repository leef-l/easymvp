package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IProject interface {
	Create(ctx context.Context, in *model.ProjectCreateInput) error
	Update(ctx context.Context, in *model.ProjectUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProjectDetailOutput, err error)
	List(ctx context.Context, in *model.ProjectListInput) (list []*model.ProjectListOutput, total int, err error)
	Export(ctx context.Context, in *model.ProjectListInput) (list []*model.ProjectListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.ProjectBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localProject IProject

func Project() IProject {
	return localProject
}

func RegisterProject(i IProject) {
	localProject = i
}
