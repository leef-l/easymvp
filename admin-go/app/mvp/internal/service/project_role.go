package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IProjectRole interface {
	Create(ctx context.Context, in *model.ProjectRoleCreateInput) error
	Update(ctx context.Context, in *model.ProjectRoleUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProjectRoleDetailOutput, err error)
	List(ctx context.Context, in *model.ProjectRoleListInput) (list []*model.ProjectRoleListOutput, total int, err error)
	Export(ctx context.Context, in *model.ProjectRoleListInput) (list []*model.ProjectRoleListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.ProjectRoleBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localProjectRole IProjectRole

func ProjectRole() IProjectRole {
	return localProjectRole
}

func RegisterProjectRole(i IProjectRole) {
	localProjectRole = i
}
