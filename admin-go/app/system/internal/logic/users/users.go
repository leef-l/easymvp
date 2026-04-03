package users

import (
	"context"

	"github.com/gogf/gf/v2/crypto/gsha256"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/system/internal/dao"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterUsers(New())
}

func New() *sUsers {
	return &sUsers{}
}

type sUsers struct{}

// Create 创建用户表
func (s *sUsers) Create(ctx context.Context, in *model.UsersCreateInput) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		id := snowflake.Generate()
		hashedPassword := gsha256.Encrypt(in.Password)
		_, err := dao.Users.Ctx(ctx).Data(g.Map{
			dao.Users.Columns().Id:        id,
			dao.Users.Columns().Username: in.Username,
			dao.Users.Columns().Password: hashedPassword,
			dao.Users.Columns().Nickname: in.Nickname,
			dao.Users.Columns().Email: in.Email,
			dao.Users.Columns().Avatar: in.Avatar,
			dao.Users.Columns().Status: in.Status,
			dao.Users.Columns().DeptId: in.DeptID,
			dao.Users.Columns().CreatedAt: gtime.Now(),
			dao.Users.Columns().UpdatedAt: gtime.Now(),
		}).Insert()
		if err != nil {
			return err
		}
		// 写入用户角色关联
		if len(in.RoleIDs) > 0 {
			data := make([]g.Map, 0, len(in.RoleIDs))
			for _, roleID := range in.RoleIDs {
				data = append(data, g.Map{
					dao.UserRole.Columns().UserId: id,
					dao.UserRole.Columns().RoleId: roleID,
				})
			}
			_, err = dao.UserRole.Ctx(ctx).Data(data).Insert()
		}
		return err
	})
}

// Update 更新用户表
func (s *sUsers) Update(ctx context.Context, in *model.UsersUpdateInput) error {
	// 内置管理员不可禁用
	if in.Status == 0 {
		isAdmin, err := s.isBuiltinAdmin(ctx, in.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return gerror.New("内置管理员账号不能禁用")
		}
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		data := g.Map{
			dao.Users.Columns().Username: in.Username,
			dao.Users.Columns().Nickname: in.Nickname,
			dao.Users.Columns().Email: in.Email,
			dao.Users.Columns().Avatar: in.Avatar,
			dao.Users.Columns().Status: in.Status,
			dao.Users.Columns().DeptId: in.DeptID,
			dao.Users.Columns().UpdatedAt: gtime.Now(),
		}
		if in.Password != "" {
			data[dao.Users.Columns().Password] = gsha256.Encrypt(in.Password)
		}
		_, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, in.ID).Data(data).Update()
		if err != nil {
			return err
		}
		// 更新用户角色关联（先删后插，事务保护）
		_, err = dao.UserRole.Ctx(ctx).Where(dao.UserRole.Columns().UserId, in.ID).Delete()
		if err != nil {
			return err
		}
		if len(in.RoleIDs) > 0 {
			roleData := make([]g.Map, 0, len(in.RoleIDs))
			for _, roleID := range in.RoleIDs {
				roleData = append(roleData, g.Map{
					dao.UserRole.Columns().UserId: in.ID,
					dao.UserRole.Columns().RoleId: roleID,
				})
			}
			_, err = dao.UserRole.Ctx(ctx).Data(roleData).Insert()
		}
		return err
	})
}

// isBuiltinAdmin 检查用户是否为内置管理员
func (s *sUsers) isBuiltinAdmin(ctx context.Context, id snowflake.JsonInt64) (bool, error) {
	val, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, id).Where(dao.Users.Columns().DeletedAt, nil).Value(dao.Users.Columns().Username)
	if err != nil {
		return false, err
	}
	return val.String() == "admin", nil
}

// Delete 软删除用户表
func (s *sUsers) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	// 内置管理员不可删除
	isAdmin, err := s.isBuiltinAdmin(ctx, id)
	if err != nil {
		return err
	}
	if isAdmin {
		return gerror.New("内置管理员账号不能删除")
	}
	_, err = dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, id).Data(g.Map{
		dao.Users.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取用户表详情
func (s *sUsers) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.UsersDetailOutput, err error) {
	out = &model.UsersDetailOutput{}
	err = dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, id).Where(dao.Users.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询部门名称
	if out.DeptID != 0 {
		val, _ := g.DB().Ctx(ctx).Model("system_dept").Where("id", out.DeptID).Where("deleted_at", nil).Value("title")
		out.DeptTitle = val.String()
	}
	// 查询用户角色ID列表
	var roles []struct {
		RoleId int64 `json:"roleId"`
	}
	_ = dao.UserRole.Ctx(ctx).Where(dao.UserRole.Columns().UserId, id).Scan(&roles)
	out.RoleIDs = make([]snowflake.JsonInt64, 0, len(roles))
	for _, r := range roles {
		out.RoleIDs = append(out.RoleIDs, snowflake.JsonInt64(r.RoleId))
	}
	return
}

// List 获取用户表列表
func (s *sUsers) List(ctx context.Context, in *model.UsersListInput) (list []*model.UsersListOutput, total int, err error) {
	m := dao.Users.Ctx(ctx).Where(dao.Users.Columns().DeletedAt, nil)
	if in.Status > 0 {
		m = m.Where(dao.Users.Columns().Status, in.Status)
	}
	if in.Username != "" {
		m = m.WhereLike(dao.Users.Columns().Username, "%"+in.Username+"%")
	}
	if in.Nickname != "" {
		m = m.WhereLike(dao.Users.Columns().Nickname, "%"+in.Nickname+"%")
	}
	if in.Email != "" {
		m = m.WhereLike(dao.Users.Columns().Email, "%"+in.Email+"%")
	}
	if in.DeptId > 0 {
		m = m.Where(dao.Users.Columns().DeptId, in.DeptId)
	}
	total, err = m.Count()
	if err != nil {
		return
	}
	err = m.Page(in.PageNum, in.PageSize).OrderAsc(dao.Users.Columns().Id).Scan(&list)
	if err != nil {
		return
	}
	// 填充部门名称
	for _, item := range list {
		if item.DeptID != 0 {
			val, _ := g.DB().Ctx(ctx).Model("system_dept").Where("id", item.DeptID).Where("deleted_at", nil).Value("title")
			item.DeptTitle = val.String()
		}
	}
	// 填充角色名称
	for _, item := range list {
		var userRoles []struct {
			RoleId int64 `json:"roleId"`
		}
		_ = dao.UserRole.Ctx(ctx).Where(dao.UserRole.Columns().UserId, item.ID).Scan(&userRoles)
		item.RoleTitles = make([]string, 0)
		if len(userRoles) > 0 {
			roleIDs := make([]int64, 0, len(userRoles))
			for _, ur := range userRoles {
				roleIDs = append(roleIDs, ur.RoleId)
			}
			var roles []struct {
				Title string `json:"title"`
			}
			_ = g.DB().Ctx(ctx).Model("system_role").Where("id", roleIDs).Where("deleted_at", nil).Scan(&roles)
			for _, r := range roles {
				item.RoleTitles = append(item.RoleTitles, r.Title)
			}
		}
	}
	return
}

// ResetPassword 重置用户密码
func (s *sUsers) ResetPassword(ctx context.Context, in *model.UsersResetPasswordInput) error {
	hashedPassword := gsha256.Encrypt(in.Password)
	_, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, in.ID).Data(g.Map{
		dao.Users.Columns().Password:  hashedPassword,
		dao.Users.Columns().UpdatedAt: gtime.Now(),
	}).Update()
	return err
}

