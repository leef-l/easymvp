package role

import (
	"context"

	v1 "easymvp/app/system/api/system/v1"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
)

var Role = cRole{}

type cRole struct{}

// Create 创建角色表
func (c *cRole) Create(ctx context.Context, req *v1.RoleCreateReq) (res *v1.RoleCreateRes, err error) {
	err = service.Role().Create(ctx, &model.RoleCreateInput{
		ParentID:        req.ParentID,
		Title:           req.Title,
		DataScope:       req.DataScope,
		DefaultAiEngine: req.DefaultAiEngine,
		Sort:            req.Sort,
		Status:          req.Status,
		IsAdmin:         req.IsAdmin,
	})
	return
}

// Update 更新角色表
func (c *cRole) Update(ctx context.Context, req *v1.RoleUpdateReq) (res *v1.RoleUpdateRes, err error) {
	err = service.Role().Update(ctx, &model.RoleUpdateInput{
		ID:              req.ID,
		ParentID:        req.ParentID,
		Title:           req.Title,
		DataScope:       req.DataScope,
		DefaultAiEngine: req.DefaultAiEngine,
		Sort:            req.Sort,
		Status:          req.Status,
		IsAdmin:         req.IsAdmin,
	})
	return
}

// Delete 删除角色表
func (c *cRole) Delete(ctx context.Context, req *v1.RoleDeleteReq) (res *v1.RoleDeleteRes, err error) {
	err = service.Role().Delete(ctx, req.ID)
	return
}

// Detail 获取角色表详情
func (c *cRole) Detail(ctx context.Context, req *v1.RoleDetailReq) (res *v1.RoleDetailRes, err error) {
	res = &v1.RoleDetailRes{}
	res.RoleDetailOutput, err = service.Role().Detail(ctx, req.ID)
	return
}

// List 获取角色表列表
func (c *cRole) List(ctx context.Context, req *v1.RoleListReq) (res *v1.RoleListRes, err error) {
	res = &v1.RoleListRes{}
	res.List, res.Total, err = service.Role().List(ctx, &model.RoleListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		DataScope: req.DataScope,
		Status:    req.Status,
	})
	return
}

// Tree 获取角色表树形结构
func (c *cRole) Tree(ctx context.Context, req *v1.RoleTreeReq) (res *v1.RoleTreeRes, err error) {
	res = &v1.RoleTreeRes{}
	res.List, err = service.Role().Tree(ctx)
	return
}

// GrantMenu 角色授权菜单
func (c *cRole) GrantMenu(ctx context.Context, req *v1.RoleGrantMenuReq) (res *v1.RoleGrantMenuRes, err error) {
	err = service.Role().GrantMenu(ctx, &model.RoleGrantMenuInput{
		ID:      req.ID,
		MenuIDs: req.MenuIDs,
	})
	return
}

// GetMenuIDs 获取角色已授权菜单ID列表
func (c *cRole) GetMenuIDs(ctx context.Context, req *v1.RoleGetMenuIDsReq) (res *v1.RoleGetMenuIDsRes, err error) {
	res = &v1.RoleGetMenuIDsRes{}
	res.MenuIDs, err = service.Role().GetMenuIDs(ctx, req.ID)
	return
}

// GrantDept 角色授权数据权限
func (c *cRole) GrantDept(ctx context.Context, req *v1.RoleGrantDeptReq) (res *v1.RoleGrantDeptRes, err error) {
	err = service.Role().GrantDept(ctx, &model.RoleGrantDeptInput{
		ID:        req.ID,
		DataScope: req.DataScope,
		DeptIDs:   req.DeptIDs,
	})
	return
}

// GetDeptIDs 获取角色已授权部门ID列表
func (c *cRole) GetDeptIDs(ctx context.Context, req *v1.RoleGetDeptIDsReq) (res *v1.RoleGetDeptIDsRes, err error) {
	res = &v1.RoleGetDeptIDsRes{}
	res.DeptIDs, err = service.Role().GetDeptIDs(ctx, req.ID)
	return
}

// GrantAiEngine 角色授权AI执行引擎
func (c *cRole) GrantAiEngine(ctx context.Context, req *v1.RoleGrantAiEngineReq) (res *v1.RoleGrantAiEngineRes, err error) {
	err = service.Role().GrantAiEngine(ctx, &model.RoleGrantAiEngineInput{
		ID:          req.ID,
		EngineCodes: req.EngineCodes,
	})
	return
}

// GetAiEngineCodes 获取角色已授权AI执行引擎编码列表
func (c *cRole) GetAiEngineCodes(ctx context.Context, req *v1.RoleGetAiEngineCodesReq) (res *v1.RoleGetAiEngineCodesRes, err error) {
	res = &v1.RoleGetAiEngineCodesRes{}
	res.EngineCodes, err = service.Role().GetAiEngineCodes(ctx, req.ID)
	return
}
