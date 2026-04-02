package service

import (
	"context"
	"easymvp/app/system/internal/model"
	"easymvp/utility/snowflake"
)

type IUsers interface {
	Create(ctx context.Context, in *model.UsersCreateInput) error
	Update(ctx context.Context, in *model.UsersUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.UsersDetailOutput, err error)
	List(ctx context.Context, in *model.UsersListInput) (list []*model.UsersListOutput, total int, err error)
	ResetPassword(ctx context.Context, in *model.UsersResetPasswordInput) error
}

var localUsers IUsers

func Users() IUsers {
	return localUsers
}

func RegisterUsers(i IUsers) {
	localUsers = i
}
