package users

import (
	"context"

	v1 "easymvp/app/system/api/system/v1"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
)

var Users = cUsers{}

type cUsers struct{}

// Create 创建用户表
func (c *cUsers) Create(ctx context.Context, req *v1.UsersCreateReq) (res *v1.UsersCreateRes, err error) {
	err = service.Users().Create(ctx, &model.UsersCreateInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email: req.Email,
		Avatar: req.Avatar,
		Status: req.Status,
		DeptID: req.DeptID,
		RoleIDs: req.RoleIDs,
	})
	return
}

// Update 更新用户表
func (c *cUsers) Update(ctx context.Context, req *v1.UsersUpdateReq) (res *v1.UsersUpdateRes, err error) {
	err = service.Users().Update(ctx, &model.UsersUpdateInput{
		ID: req.ID,
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email: req.Email,
		Avatar: req.Avatar,
		Status: req.Status,
		DeptID: req.DeptID,
		RoleIDs: req.RoleIDs,
	})
	return
}

// Delete 删除用户表
func (c *cUsers) Delete(ctx context.Context, req *v1.UsersDeleteReq) (res *v1.UsersDeleteRes, err error) {
	err = service.Users().Delete(ctx, req.ID)
	return
}

// Detail 获取用户表详情
func (c *cUsers) Detail(ctx context.Context, req *v1.UsersDetailReq) (res *v1.UsersDetailRes, err error) {
	res = &v1.UsersDetailRes{}
	res.UsersDetailOutput, err = service.Users().Detail(ctx, req.ID)
	return
}

// List 获取用户表列表
func (c *cUsers) List(ctx context.Context, req *v1.UsersListReq) (res *v1.UsersListRes, err error) {
	res = &v1.UsersListRes{}
	res.List, res.Total, err = service.Users().List(ctx, &model.UsersListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		DeptId:   req.DeptId,
		Status:   req.Status,
	})
	return
}

// ResetPassword 重置用户密码
func (c *cUsers) ResetPassword(ctx context.Context, req *v1.UsersResetPasswordReq) (res *v1.UsersResetPasswordRes, err error) {
	err = service.Users().ResetPassword(ctx, &model.UsersResetPasswordInput{
		ID:       req.ID,
		Password: req.Password,
	})
	return
}

