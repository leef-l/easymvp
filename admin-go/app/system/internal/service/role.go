package service

import (
	"context"
	"easymvp/app/system/internal/model"
	"easymvp/utility/snowflake"
)

type IRole interface {
	Create(ctx context.Context, in *model.RoleCreateInput) error
	Update(ctx context.Context, in *model.RoleUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.RoleDetailOutput, err error)
	List(ctx context.Context, in *model.RoleListInput) (list []*model.RoleListOutput, total int, err error)
	Tree(ctx context.Context) (tree []*model.RoleTreeOutput, err error)
	GrantMenu(ctx context.Context, in *model.RoleGrantMenuInput) error
	GetMenuIDs(ctx context.Context, roleID snowflake.JsonInt64) ([]snowflake.JsonInt64, error)
	GrantDept(ctx context.Context, in *model.RoleGrantDeptInput) error
	GetDeptIDs(ctx context.Context, roleID snowflake.JsonInt64) ([]snowflake.JsonInt64, error)
}

var localRole IRole

func Role() IRole {
	return localRole
}

func RegisterRole(i IRole) {
	localRole = i
}
