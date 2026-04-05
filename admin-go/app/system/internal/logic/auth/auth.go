package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/system/internal/dao"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
	"easymvp/app/system/internal/support"
	"easymvp/utility/jwt"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterAuth(New())
}

func New() *sAuth {
	return &sAuth{}
}

type sAuth struct{}

// Login 用户登录
func (s *sAuth) Login(ctx context.Context, in *model.AuthLoginInput) (out *model.AuthLoginOutput, err error) {
	// 查询用户
	var user struct {
		Id       int64  `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		DeptId   int64  `json:"deptId"`
		Status   int    `json:"status"`
	}

	err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Username, in.Username).
		Where(dao.Users.Columns().DeletedAt, nil).
		Scan(&user)

	if err != nil {
		g.Log().Errorf(ctx, "查询用户失败: %v", err)
		return nil, gerror.New("用户名或密码错误")
	}
	if user.Id == 0 {
		return nil, gerror.New("用户名或密码错误")
	}

	// 校验状态
	if user.Status == 0 {
		return nil, gerror.New("账号已被禁用")
	}

	// 校验密码（兼容 bcrypt 和 SHA256）
	if !support.VerifyPassword(user.Password, in.Password) {
		return nil, gerror.New("用户名或密码错误")
	}

	// 查询用户是否为超管（通过角色的 is_admin 标记）
	isAdmin := false
	profile, _ := support.LoadUserAccessProfile(ctx, snowflake.JsonInt64(user.Id))
	if profile != nil {
		isAdmin = profile.IsAdmin
	}

	// 生成 Token（包含 isAdmin 标记，避免运行时频繁查库）
	token, err := jwt.GenerateToken(user.Id, user.Username, user.DeptId, isAdmin)
	if err != nil {
		return nil, gerror.New("生成Token失败")
	}

	out = &model.AuthLoginOutput{
		Token:    token,
		UserID:   snowflake.JsonInt64(user.Id),
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}
	return
}

// Info 获取当前用户信息
func (s *sAuth) Info(ctx context.Context, userID snowflake.JsonInt64) (out *model.AuthInfoOutput, err error) {
	var user struct {
		Id       int64  `json:"id"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
		DeptId   int64  `json:"deptId"`
		Status   int    `json:"status"`
	}
	err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, userID).
		Where(dao.Users.Columns().DeletedAt, nil).
		Scan(&user)
	if err != nil {
		return nil, err
	}

	out = &model.AuthInfoOutput{
		UserID:   snowflake.JsonInt64(user.Id),
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Avatar:   user.Avatar,
		DeptID:   snowflake.JsonInt64(user.DeptId),
		Status:   user.Status,
		Roles:    make([]string, 0),
		Perms:    make([]string, 0),
	}

	// 加载用户权限配置
	profile, _ := support.LoadUserAccessProfile(ctx, userID)
	if profile == nil || len(profile.RoleIDs) == 0 {
		return out, nil
	}

	// 收集角色名称
	for _, role := range profile.Roles {
		if role.Title != "" {
			out.Roles = append(out.Roles, role.Title)
		}
	}

	// 权限处理
	if profile.IsAdmin {
		// 超级管理员获取所有权限
		out.Perms, _ = support.LoadAllMenuPermissions(ctx)
	} else {
		// 普通用户：查询角色授权的菜单权限
		menuIDs, _ := support.LoadRoleMenuIDs(ctx, profile.RoleIDs)
		out.Perms, _ = support.LoadMenuPermissions(ctx, menuIDs)
	}

	return
}

// ChangePassword 修改密码
func (s *sAuth) ChangePassword(ctx context.Context, in *model.AuthChangePasswordInput) error {
	// 查询当前密码
	password, err := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, in.UserID).
		Value(dao.Users.Columns().Password)
	if err != nil {
		return err
	}

	// 校验旧密码（兼容 bcrypt 和 SHA256）
	if !support.VerifyPassword(password.String(), in.OldPassword) {
		return gerror.New("旧密码错误")
	}

	// 使用 bcrypt 加密新密码
	hashedNew, err := support.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}

	// 更新密码
	_, err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, in.UserID).
		Data(dao.Users.Columns().Password, hashedNew).
		Update()
	return err
}

// Menus 获取当前用户的菜单树（动态路由）
func (s *sAuth) Menus(ctx context.Context, userID snowflake.JsonInt64) ([]*model.AuthMenuOutput, error) {
	// 加载用户权限配置
	profile, err := support.LoadUserAccessProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(profile.RoleIDs) == 0 {
		return make([]*model.AuthMenuOutput, 0), nil
	}

	var list []*model.AuthMenuOutput

	if profile.IsAdmin {
		// 超级管理员获取所有菜单
		list, err = support.LoadAllAuthMenus(ctx)
	} else {
		// 普通用户：获取授权菜单
		menuIDs, _ := support.LoadRoleMenuIDs(ctx, profile.RoleIDs)
		list, err = support.LoadAuthMenus(ctx, menuIDs)
	}
	if err != nil {
		return nil, err
	}

	return support.BuildAuthMenuTree(list), nil
}
