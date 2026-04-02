package auth

import (
	"context"

	v1 "easymvp/app/system/api/system/v1"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
	"easymvp/utility/snowflake"
)

var Auth = cAuth{}

type cAuth struct{}

// Login 用户登录
func (c *cAuth) Login(ctx context.Context, req *v1.AuthLoginReq) (res *v1.AuthLoginRes, err error) {
	out, err := service.Auth().Login(ctx, &model.AuthLoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.AuthLoginRes{
		Token:    out.Token,
		UserID:   out.UserID,
		Username: out.Username,
		Nickname: out.Nickname,
		Avatar:   out.Avatar,
	}
	return
}

// Info 获取当前用户信息
func (c *cAuth) Info(ctx context.Context, req *v1.AuthInfoReq) (res *v1.AuthInfoRes, err error) {
	claims := GetClaims(ctx)
	if claims == nil {
		return nil, nil
	}
	out, err := service.Auth().Info(ctx, snowflake.JsonInt64(claims.UserID))
	if err != nil {
		return nil, err
	}
	res = &v1.AuthInfoRes{
		UserID:   out.UserID,
		Username: out.Username,
		Nickname: out.Nickname,
		Email:    out.Email,
		Avatar:   out.Avatar,
		DeptID:   out.DeptID,
		Status:   out.Status,
		Roles:    out.Roles,
		Perms:    out.Perms,
	}
	return
}

// ChangePassword 修改密码
func (c *cAuth) ChangePassword(ctx context.Context, req *v1.AuthChangePasswordReq) (res *v1.AuthChangePasswordRes, err error) {
	claims := GetClaims(ctx)
	if claims == nil {
		return nil, nil
	}
	err = service.Auth().ChangePassword(ctx, &model.AuthChangePasswordInput{
		UserID:      snowflake.JsonInt64(claims.UserID),
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	return
}

// Menus 获取当前用户菜单树
func (c *cAuth) Menus(ctx context.Context, req *v1.AuthMenusReq) (res *v1.AuthMenusRes, err error) {
	claims := GetClaims(ctx)
	if claims == nil {
		return nil, nil
	}
	menus, err := service.Auth().Menus(ctx, snowflake.JsonInt64(claims.UserID))
	if err != nil {
		return nil, err
	}
	res = &v1.AuthMenusRes{Menus: menus}
	return
}
