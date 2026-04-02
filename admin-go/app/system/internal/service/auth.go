package service

import (
	"context"

	"easymvp/app/system/internal/model"
	"easymvp/utility/snowflake"
)

type IAuth interface {
	Login(ctx context.Context, in *model.AuthLoginInput) (out *model.AuthLoginOutput, err error)
	Info(ctx context.Context, userID snowflake.JsonInt64) (out *model.AuthInfoOutput, err error)
	ChangePassword(ctx context.Context, in *model.AuthChangePasswordInput) error
	Menus(ctx context.Context, userID snowflake.JsonInt64) ([]*model.AuthMenuOutput, error)
}

var localAuth IAuth

func Auth() IAuth {
	return localAuth
}

func RegisterAuth(i IAuth) {
	localAuth = i
}
