package menu

import (
	"context"

	v1 "easymvp/app/system/api/system/v1"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
)

var Menu = cMenu{}

type cMenu struct{}

// Create 创建菜单表
func (c *cMenu) Create(ctx context.Context, req *v1.MenuCreateReq) (res *v1.MenuCreateRes, err error) {
	err = service.Menu().Create(ctx, &model.MenuCreateInput{
		ParentID: req.ParentID,
		Title: req.Title,
		Type: req.Type,
		Path: req.Path,
		Component: req.Component,
		Permission: req.Permission,
		Icon: req.Icon,
		Sort: req.Sort,
		IsShow: req.IsShow,
		IsCache: req.IsCache,
		LinkURL: req.LinkURL,
		Status: req.Status,
	})
	return
}

// Update 更新菜单表
func (c *cMenu) Update(ctx context.Context, req *v1.MenuUpdateReq) (res *v1.MenuUpdateRes, err error) {
	err = service.Menu().Update(ctx, &model.MenuUpdateInput{
		ID: req.ID,
		ParentID: req.ParentID,
		Title: req.Title,
		Type: req.Type,
		Path: req.Path,
		Component: req.Component,
		Permission: req.Permission,
		Icon: req.Icon,
		Sort: req.Sort,
		IsShow: req.IsShow,
		IsCache: req.IsCache,
		LinkURL: req.LinkURL,
		Status: req.Status,
	})
	return
}

// Delete 删除菜单表
func (c *cMenu) Delete(ctx context.Context, req *v1.MenuDeleteReq) (res *v1.MenuDeleteRes, err error) {
	err = service.Menu().Delete(ctx, req.ID)
	return
}

// Detail 获取菜单表详情
func (c *cMenu) Detail(ctx context.Context, req *v1.MenuDetailReq) (res *v1.MenuDetailRes, err error) {
	res = &v1.MenuDetailRes{}
	res.MenuDetailOutput, err = service.Menu().Detail(ctx, req.ID)
	return
}

// List 获取菜单表列表
func (c *cMenu) List(ctx context.Context, req *v1.MenuListReq) (res *v1.MenuListRes, err error) {
	res = &v1.MenuListRes{}
	res.List, res.Total, err = service.Menu().List(ctx, &model.MenuListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Type: req.Type,
		IsShow: req.IsShow,
		IsCache: req.IsCache,
		Status: req.Status,
	})
	return
}

// Tree 获取菜单表树形结构
func (c *cMenu) Tree(ctx context.Context, req *v1.MenuTreeReq) (res *v1.MenuTreeRes, err error) {
	res = &v1.MenuTreeRes{}
	res.List, err = service.Menu().Tree(ctx)
	return
}

