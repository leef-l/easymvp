package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
	"github.com/gogf/gf/v2/net/ghttp"
)

type IRolePreset interface {
	Create(ctx context.Context, in *model.RolePresetCreateInput) error
	Update(ctx context.Context, in *model.RolePresetUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.RolePresetDetailOutput, err error)
	List(ctx context.Context, in *model.RolePresetListInput) (list []*model.RolePresetListOutput, total int, err error)
	Export(ctx context.Context, in *model.RolePresetListInput) (list []*model.RolePresetListOutput, err error)
	BatchUpdate(ctx context.Context, in *model.RolePresetBatchUpdateInput) error
	Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error)
}

var localRolePreset IRolePreset

func RolePreset() IRolePreset {
	return localRolePreset
}

func RegisterRolePreset(i IRolePreset) {
	localRolePreset = i
}
