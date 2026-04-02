package dept

import (
	"context"

	v1 "easymvp/app/system/api/system/v1"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
)

var Dept = cDept{}

type cDept struct{}

// Create 创建部门表
func (c *cDept) Create(ctx context.Context, req *v1.DeptCreateReq) (res *v1.DeptCreateRes, err error) {
	err = service.Dept().Create(ctx, &model.DeptCreateInput{
		ParentID: req.ParentID,
		Title: req.Title,
		Username: req.Username,
		Email: req.Email,
		Sort: req.Sort,
		Status: req.Status,
	})
	return
}

// Update 更新部门表
func (c *cDept) Update(ctx context.Context, req *v1.DeptUpdateReq) (res *v1.DeptUpdateRes, err error) {
	err = service.Dept().Update(ctx, &model.DeptUpdateInput{
		ID: req.ID,
		ParentID: req.ParentID,
		Title: req.Title,
		Username: req.Username,
		Email: req.Email,
		Sort: req.Sort,
		Status: req.Status,
	})
	return
}

// Delete 删除部门表
func (c *cDept) Delete(ctx context.Context, req *v1.DeptDeleteReq) (res *v1.DeptDeleteRes, err error) {
	err = service.Dept().Delete(ctx, req.ID)
	return
}

// Detail 获取部门表详情
func (c *cDept) Detail(ctx context.Context, req *v1.DeptDetailReq) (res *v1.DeptDetailRes, err error) {
	res = &v1.DeptDetailRes{}
	res.DeptDetailOutput, err = service.Dept().Detail(ctx, req.ID)
	return
}

// List 获取部门表列表
func (c *cDept) List(ctx context.Context, req *v1.DeptListReq) (res *v1.DeptListRes, err error) {
	res = &v1.DeptListRes{}
	res.List, res.Total, err = service.Dept().List(ctx, &model.DeptListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Status: req.Status,
	})
	return
}

// Tree 获取部门表树形结构
func (c *cDept) Tree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	res = &v1.DeptTreeRes{}
	res.List, err = service.Dept().Tree(ctx)
	return
}

