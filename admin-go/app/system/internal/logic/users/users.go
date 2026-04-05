package users

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/system/internal/consts"
	"easymvp/app/system/internal/dao"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
	"easymvp/app/system/internal/support"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterUsers(New())
}

func New() *sUsers {
	return &sUsers{}
}

type sUsers struct{}

// Create 创建用户
func (s *sUsers) Create(ctx context.Context, in *model.UsersCreateInput) error {
	s.normalizeUserInput(&in.Username, &in.Nickname, &in.Email, &in.Avatar)
	if strings.TrimSpace(in.Password) == "" {
		return gerror.New("密码不能为空")
	}
	if err := s.validateUserStatus(in.Status); err != nil {
		return err
	}
	if err := s.ensureDeptExists(ctx, in.DeptID); err != nil {
		return err
	}
	if err := s.ensureRolesExist(ctx, in.RoleIDs); err != nil {
		return err
	}
	if err := s.ensureUsernameUnique(ctx, in.Username, 0); err != nil {
		return err
	}

	id := snowflake.Generate()
	hashedPassword, err := support.HashPassword(in.Password)
	if err != nil {
		return err
	}
	return dao.Users.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.Users.Table()).Ctx(ctx).Data(g.Map{
			dao.Users.Columns().Id:        id,
			dao.Users.Columns().Username:  in.Username,
			dao.Users.Columns().Password:  hashedPassword,
			dao.Users.Columns().Nickname:  in.Nickname,
			dao.Users.Columns().Email:     in.Email,
			dao.Users.Columns().Avatar:    in.Avatar,
			dao.Users.Columns().Status:    in.Status,
			dao.Users.Columns().DeptId:    in.DeptID,
			dao.Users.Columns().CreatedAt: gtime.Now(),
			dao.Users.Columns().UpdatedAt: gtime.Now(),
		}).Insert()
		if err != nil {
			return err
		}
		return s.replaceUserRoles(ctx, tx, int64(id), in.RoleIDs)
	})
}

// Update 更新用户
func (s *sUsers) Update(ctx context.Context, in *model.UsersUpdateInput) error {
	s.normalizeUserInput(&in.Username, &in.Nickname, &in.Email, &in.Avatar)
	currentUser, err := s.getUserEntity(ctx, in.ID)
	if err != nil {
		return err
	}
	if err := s.validateUserStatus(in.Status); err != nil {
		return err
	}
	if err := s.ensureDeptExists(ctx, in.DeptID); err != nil {
		return err
	}
	if err := s.ensureRolesExist(ctx, in.RoleIDs); err != nil {
		return err
	}
	if err := s.ensureUsernameUnique(ctx, in.Username, int64(in.ID)); err != nil {
		return err
	}

	// 内置管理员不可禁用
	if in.Status == 0 {
		if err := s.ensureBuiltinAdminEditable(ctx, in.ID, "disable"); err != nil {
			return err
		}
	}
	data := g.Map{
		dao.Users.Columns().Username:  in.Username,
		dao.Users.Columns().Nickname:  in.Nickname,
		dao.Users.Columns().Email:     in.Email,
		dao.Users.Columns().Avatar:    in.Avatar,
		dao.Users.Columns().Status:    in.Status,
		dao.Users.Columns().DeptId:    in.DeptID,
		dao.Users.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Password != "" {
		in.Password = strings.TrimSpace(in.Password)
		if support.VerifyPassword(currentUser.Password, in.Password) {
			return gerror.New("新密码不能与当前密码相同")
		}
		hashedPassword, err := support.HashPassword(in.Password)
		if err != nil {
			return err
		}
		data[dao.Users.Columns().Password] = hashedPassword
	}

	return dao.Users.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.Users.Table()).Ctx(ctx).
			Where(dao.Users.Columns().Id, in.ID).
			Where(dao.Users.Columns().DeletedAt, nil).
			Data(data).
			Update()
		if err != nil {
			return err
		}
		return s.replaceUserRoles(ctx, tx, int64(in.ID), in.RoleIDs)
	})
}

// Delete 软删除用户
func (s *sUsers) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	if err := s.ensureBuiltinAdminEditable(ctx, id, "delete"); err != nil {
		return err
	}
	if _, err := s.getUserEntity(ctx, id); err != nil {
		return err
	}
	_, err := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, id).
		Where(dao.Users.Columns().DeletedAt, nil).
		Data(g.Map{
			dao.Users.Columns().DeletedAt: gtime.Now(),
		}).Update()
	return err
}

// Detail 获取用户详情
func (s *sUsers) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.UsersDetailOutput, err error) {
	out = &model.UsersDetailOutput{}
	err = dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, id).Where(dao.Users.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, gerror.New("用户不存在")
	}
	// 查询部门名称
	if out.DeptID != 0 {
		out.DeptTitle = support.LoadTitle(ctx, "system_dept", int64(out.DeptID))
	}
	// 查询用户角色ID列表
	var roles []struct {
		RoleId int64 `json:"roleId"`
	}
	_ = dao.UserRole.Ctx(ctx).Where(dao.UserRole.Columns().UserId, id).Scan(&roles)
	out.RoleIDs = make([]snowflake.JsonInt64, 0, len(roles))
	seen := make(map[int64]struct{}, len(roles))
	for _, r := range roles {
		if _, ok := seen[r.RoleId]; ok {
			continue
		}
		seen[r.RoleId] = struct{}{}
		out.RoleIDs = append(out.RoleIDs, snowflake.JsonInt64(r.RoleId))
	}
	sort.Slice(out.RoleIDs, func(i, j int) bool { return out.RoleIDs[i] < out.RoleIDs[j] })
	return
}

// List 获取用户列表（带数据权限过滤）
func (s *sUsers) List(ctx context.Context, in *model.UsersListInput) (list []*model.UsersListOutput, total int, err error) {
	in.Username = strings.TrimSpace(strings.ToLower(in.Username))
	in.Nickname = strings.TrimSpace(in.Nickname)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	m := dao.Users.Ctx(ctx).Where(dao.Users.Columns().DeletedAt, nil)

	// 数据权限过滤
	currentUserID := support.GetCurrentUserID(ctx)
	if currentUserID > 0 {
		profile, profileErr := support.LoadUserAccessProfile(ctx, snowflake.JsonInt64(currentUserID))
		if profileErr != nil {
			return nil, 0, profileErr
		}
		dataScope, scopeErr := support.ResolveUserDataScope(ctx, profile, support.GetCurrentDeptID(ctx))
		if scopeErr != nil {
			return nil, 0, scopeErr
		}
		m = s.applyUserDataScope(m, currentUserID, dataScope)
	}

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
	err = m.Page(in.PageNum, in.PageSize).OrderDesc(dao.Users.Columns().Id).Scan(&list)
	if err != nil {
		return
	}

	// 批量加载部门名称和角色名称
	deptIDs := make([]int64, 0, len(list))
	userIDs := make([]int64, 0, len(list))
	for _, item := range list {
		userIDs = append(userIDs, int64(item.ID))
		if item.DeptID != 0 {
			deptIDs = append(deptIDs, int64(item.DeptID))
		}
		item.RoleTitles = make([]string, 0)
	}

	if len(deptIDs) > 0 {
		deptTitleMap := support.LoadTitleMap(ctx, "system_dept", deptIDs)
		for _, item := range list {
			if item.DeptID != 0 {
				item.DeptTitle = deptTitleMap[int64(item.DeptID)]
			}
		}
	}

	if len(userIDs) > 0 {
		var userRoles []struct {
			UserId int64 `json:"userId"`
			RoleId int64 `json:"roleId"`
		}
		_ = dao.UserRole.Ctx(ctx).
			WhereIn(dao.UserRole.Columns().UserId, userIDs).
			Scan(&userRoles)

		userRoleMap := make(map[int64][]int64)
		roleIDs := make([]int64, 0, len(userRoles))
		for _, userRole := range userRoles {
			userRoleMap[userRole.UserId] = append(userRoleMap[userRole.UserId], userRole.RoleId)
			roleIDs = append(roleIDs, userRole.RoleId)
		}
		roleTitleMap := support.LoadRoleTitleMap(ctx, roleIDs)

		for _, item := range list {
			seenTitles := make(map[string]struct{})
			for _, roleID := range userRoleMap[int64(item.ID)] {
				if title, ok := roleTitleMap[roleID]; ok && title != "" {
					if _, seen := seenTitles[title]; seen {
						continue
					}
					seenTitles[title] = struct{}{}
					item.RoleTitles = append(item.RoleTitles, title)
				}
			}
			sort.Strings(item.RoleTitles)
		}
	}
	return
}

// ResetPassword 重置用户密码
func (s *sUsers) ResetPassword(ctx context.Context, in *model.UsersResetPasswordInput) error {
	in.Password = strings.TrimSpace(in.Password)
	if in.Password == "" {
		return gerror.New("新密码不能为空")
	}
	if err := s.ensureBuiltinAdminEditable(ctx, in.ID, "reset-password"); err != nil {
		return err
	}
	user, err := s.getUserEntity(ctx, in.ID)
	if err != nil {
		return err
	}
	if support.VerifyPassword(user.Password, in.Password) {
		return gerror.New("新密码不能与当前密码相同")
	}
	hashedPassword, err := support.HashPassword(in.Password)
	if err != nil {
		return err
	}
	_, err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, in.ID).
		Where(dao.Users.Columns().DeletedAt, nil).
		Data(g.Map{
			dao.Users.Columns().Password:  hashedPassword,
			dao.Users.Columns().UpdatedAt: gtime.Now(),
		}).Update()
	return err
}

// ---- 内部辅助方法 ----

func (s *sUsers) applyUserDataScope(m *gdb.Model, currentUserID int64, scope *support.UserDataScope) *gdb.Model {
	if scope == nil || scope.All {
		return m
	}
	if len(scope.DeptIDs) == 0 {
		if scope.IncludeSelf {
			return m.Where(dao.Users.Columns().Id, currentUserID)
		}
		return m.Where(dao.Users.Columns().Id, -1)
	}
	if scope.IncludeSelf {
		return m.Where(
			fmt.Sprintf("(%s = ? OR %s IN (?))", dao.Users.Columns().Id, dao.Users.Columns().DeptId),
			currentUserID,
			scope.DeptIDs,
		)
	}
	return m.WhereIn(dao.Users.Columns().DeptId, scope.DeptIDs)
}

type userEntity struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *sUsers) getUserEntity(ctx context.Context, id snowflake.JsonInt64) (*userEntity, error) {
	out := &userEntity{}
	err := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, id).
		Where(dao.Users.Columns().DeletedAt, nil).
		Scan(out)
	if err != nil {
		return nil, err
	}
	if out.Id == 0 {
		return nil, gerror.New("用户不存在")
	}
	return out, nil
}

func (s *sUsers) replaceUserRoles(ctx context.Context, tx gdb.TX, userID int64, roleIDs []snowflake.JsonInt64) error {
	_, err := tx.Model(dao.UserRole.Table()).Ctx(ctx).
		Where(dao.UserRole.Columns().UserId, userID).
		Delete()
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}

	roleData := make([]g.Map, 0, len(roleIDs))
	seen := make(map[int64]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		id := int64(roleID)
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		roleData = append(roleData, g.Map{
			dao.UserRole.Columns().UserId: userID,
			dao.UserRole.Columns().RoleId: roleID,
		})
	}
	if len(roleData) == 0 {
		return nil
	}
	_, err = tx.Model(dao.UserRole.Table()).Ctx(ctx).Data(roleData).Insert()
	return err
}

// isBuiltinAdmin 检查用户是否为内置管理员（username=admin）
func (s *sUsers) isBuiltinAdmin(ctx context.Context, id snowflake.JsonInt64) (bool, error) {
	val, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, id).Where(dao.Users.Columns().DeletedAt, nil).Value(dao.Users.Columns().Username)
	if err != nil {
		return false, err
	}
	return val.String() == "admin", nil
}

func (s *sUsers) ensureBuiltinAdminEditable(ctx context.Context, id snowflake.JsonInt64, action string) error {
	isAdmin, err := s.isBuiltinAdmin(ctx, id)
	if err != nil {
		return err
	}
	if !isAdmin {
		return nil
	}
	switch action {
	case "delete":
		return gerror.New("内置管理员账号不能删除")
	case "disable":
		return gerror.New("内置管理员账号不能禁用")
	case "reset-password":
		return gerror.New("内置管理员账号不能被重置密码")
	default:
		return gerror.New("内置管理员账号不允许执行该操作")
	}
}

func (s *sUsers) ensureUsernameUnique(ctx context.Context, username string, excludeID int64) error {
	if username == "" {
		return gerror.New("登录用户名不能为空")
	}

	query := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Username, username).
		Where(dao.Users.Columns().DeletedAt, nil)
	if excludeID > 0 {
		query = query.WhereNot(dao.Users.Columns().Id, excludeID)
	}

	count, err := query.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("登录用户名已存在")
	}
	return nil
}

func (s *sUsers) ensureDeptExists(ctx context.Context, deptID snowflake.JsonInt64) error {
	if deptID == 0 {
		return nil
	}
	count, err := dao.Dept.Ctx(ctx).
		Where(dao.Dept.Columns().Id, deptID).
		Where(dao.Dept.Columns().DeletedAt, nil).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("所属部门不存在")
	}
	return nil
}

func (s *sUsers) ensureRolesExist(ctx context.Context, roleIDs []snowflake.JsonInt64) error {
	if len(roleIDs) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID != 0 {
			ids = append(ids, int64(roleID))
		}
	}
	ids = support.UniqueNonZeroIDs(ids)
	if len(ids) == 0 {
		return nil
	}

	count, err := dao.Role.Ctx(ctx).
		WhereIn(dao.Role.Columns().Id, ids).
		Where(dao.Role.Columns().DeletedAt, nil).
		Count()
	if err != nil {
		return err
	}
	if count != len(ids) {
		return gerror.New("关联角色不存在")
	}
	return nil
}

func (s *sUsers) normalizeUserInput(username, nickname, email, avatar *string) {
	*username = strings.TrimSpace(strings.ToLower(*username))
	*nickname = strings.TrimSpace(*nickname)
	*email = strings.TrimSpace(strings.ToLower(*email))
	*avatar = strings.TrimSpace(*avatar)
}

func (s *sUsers) validateUserStatus(status int) error {
	if !consts.IsValidUsersStatus(status) {
		return gerror.New("用户状态不合法")
	}
	return nil
}
