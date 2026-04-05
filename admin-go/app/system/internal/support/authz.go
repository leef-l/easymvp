package support

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/system/internal/dao"
	"easymvp/app/system/internal/model"
	"easymvp/utility/jwt"
	"easymvp/utility/snowflake"
)

// RoleInfo 角色基本信息
type RoleInfo struct {
	ID        int64
	Title     string
	IsAdmin   int
	DataScope int
}

// UserAccessProfile 用户权限配置
type UserAccessProfile struct {
	RoleIDs []int64
	Roles   []RoleInfo
	IsAdmin bool
}

// UserDataScope 用户数据权限范围
type UserDataScope struct {
	All         bool
	IncludeSelf bool
	DeptIDs     []int64
}

// GetCurrentUserID 从 context 获取当前用户 ID
func GetCurrentUserID(ctx context.Context) int64 {
	val := g.RequestFromCtx(ctx).GetCtxVar("jwt_user_id")
	if val.IsNil() {
		return 0
	}
	return val.Int64()
}

// GetCurrentDeptID 从 context 获取当前用户部门 ID
func GetCurrentDeptID(ctx context.Context) int64 {
	val := g.RequestFromCtx(ctx).GetCtxVar("jwt_dept_id")
	if val.IsNil() {
		return 0
	}
	return val.Int64()
}

// GetCurrentIsAdmin 从 context 获取当前用户是否超管
func GetCurrentIsAdmin(ctx context.Context) bool {
	claims := GetCurrentClaims(ctx)
	if claims == nil {
		return false
	}
	return claims.IsAdmin
}

// GetCurrentClaims 从 context 获取 JWT Claims
func GetCurrentClaims(ctx context.Context) *jwt.Claims {
	val := g.RequestFromCtx(ctx).GetCtxVar("jwt_claims")
	if val.IsNil() {
		return nil
	}
	if claims, ok := val.Val().(*jwt.Claims); ok {
		return claims
	}
	return nil
}

// LoadUserRoleIDs 查询用户关联的角色 ID 列表
func LoadUserRoleIDs(ctx context.Context, userID snowflake.JsonInt64) ([]int64, error) {
	var userRoles []struct {
		RoleId int64 `json:"roleId"`
	}
	if err := dao.UserRole.Ctx(ctx).
		Where(dao.UserRole.Columns().UserId, userID).
		Scan(&userRoles); err != nil {
		return nil, err
	}

	roleIDs := make([]int64, 0, len(userRoles))
	seen := make(map[int64]struct{}, len(userRoles))
	for _, userRole := range userRoles {
		if _, ok := seen[userRole.RoleId]; ok {
			continue
		}
		seen[userRole.RoleId] = struct{}{}
		roleIDs = append(roleIDs, userRole.RoleId)
	}
	return roleIDs, nil
}

// LoadRoleInfos 批量加载角色信息
func LoadRoleInfos(ctx context.Context, roleIDs []int64) ([]RoleInfo, error) {
	roleIDs = UniqueNonZeroIDs(roleIDs)
	if len(roleIDs) == 0 {
		return make([]RoleInfo, 0), nil
	}

	var roles []RoleInfo
	err := g.DB().Ctx(ctx).Model("system_role").
		Fields("id,title,is_admin,data_scope").
		Where("id", roleIDs).
		Where("deleted_at", nil).
		Where("status", 1).
		OrderAsc("id").
		Scan(&roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// LoadUserAccessProfile 加载用户完整权限配置（角色列表 + 超管标记）
func LoadUserAccessProfile(ctx context.Context, userID snowflake.JsonInt64) (*UserAccessProfile, error) {
	roleIDs, err := LoadUserRoleIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := &UserAccessProfile{
		RoleIDs: roleIDs,
		Roles:   make([]RoleInfo, 0),
	}
	if len(roleIDs) == 0 {
		return profile, nil
	}
	roles, err := LoadRoleInfos(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	profile.Roles = roles
	for _, role := range roles {
		if role.IsAdmin == 1 {
			profile.IsAdmin = true
			break
		}
	}
	return profile, nil
}

// LoadRoleDeptIDs 加载角色关联的自定义部门 ID 列表
func LoadRoleDeptIDs(ctx context.Context, roleIDs []int64) ([]int64, error) {
	roleIDs = UniqueNonZeroIDs(roleIDs)
	if len(roleIDs) == 0 {
		return make([]int64, 0), nil
	}

	var roleDepts []struct {
		DeptId int64 `json:"deptId"`
	}
	if err := dao.RoleDept.Ctx(ctx).
		WhereIn(dao.RoleDept.Columns().RoleId, roleIDs).
		Scan(&roleDepts); err != nil {
		return nil, err
	}

	deptIDs := make([]int64, 0, len(roleDepts))
	seen := make(map[int64]struct{}, len(roleDepts))
	for _, roleDept := range roleDepts {
		if roleDept.DeptId == 0 {
			continue
		}
		if _, ok := seen[roleDept.DeptId]; ok {
			continue
		}
		seen[roleDept.DeptId] = struct{}{}
		deptIDs = append(deptIDs, roleDept.DeptId)
	}
	return deptIDs, nil
}

// LoadDeptSubtreeIDs 加载部门及所有子部门 ID（BFS 算法）
func LoadDeptSubtreeIDs(ctx context.Context, rootID int64) ([]int64, error) {
	if rootID == 0 {
		return make([]int64, 0), nil
	}

	var depts []struct {
		ID       int64 `json:"id"`
		ParentID int64 `json:"parentId"`
	}
	if err := dao.Dept.Ctx(ctx).
		Fields(dao.Dept.Columns().Id, dao.Dept.Columns().ParentId).
		Where(dao.Dept.Columns().DeletedAt, nil).
		Where(dao.Dept.Columns().Status, 1).
		OrderAsc(dao.Dept.Columns().Id).
		Scan(&depts); err != nil {
		return nil, err
	}

	childrenMap := make(map[int64][]int64, len(depts))
	exists := make(map[int64]struct{}, len(depts))
	for _, dept := range depts {
		exists[dept.ID] = struct{}{}
		childrenMap[dept.ParentID] = append(childrenMap[dept.ParentID], dept.ID)
	}
	if _, ok := exists[rootID]; !ok {
		return make([]int64, 0), nil
	}

	result := make([]int64, 0, len(depts))
	queue := []int64{rootID}
	seen := map[int64]struct{}{rootID: {}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		for _, childID := range childrenMap[current] {
			if _, ok := seen[childID]; ok {
				continue
			}
			seen[childID] = struct{}{}
			queue = append(queue, childID)
		}
	}
	return result, nil
}

// ResolveUserDataScope 解析用户数据权限范围（五级模型）
func ResolveUserDataScope(ctx context.Context, profile *UserAccessProfile, currentDeptID int64) (*UserDataScope, error) {
	scope := &UserDataScope{
		DeptIDs: make([]int64, 0),
	}
	if profile == nil || len(profile.RoleIDs) == 0 {
		scope.IncludeSelf = true
		return scope, nil
	}
	if profile.IsAdmin {
		scope.All = true
		return scope, nil
	}

	deptSet := make(map[int64]struct{})
	roleDeptIDs, err := LoadRoleDeptIDs(ctx, profile.RoleIDs)
	if err != nil {
		return nil, err
	}
	for _, role := range profile.Roles {
		switch role.DataScope {
		case 1: // 全部
			scope.All = true
			return scope, nil
		case 2: // 本部门及以下
			subtreeIDs, err := LoadDeptSubtreeIDs(ctx, currentDeptID)
			if err != nil {
				return nil, err
			}
			for _, deptID := range subtreeIDs {
				deptSet[deptID] = struct{}{}
			}
		case 3: // 本部门
			if currentDeptID > 0 {
				deptSet[currentDeptID] = struct{}{}
			}
		case 4: // 仅本人
			scope.IncludeSelf = true
		case 5: // 自定义
			for _, deptID := range roleDeptIDs {
				deptSet[deptID] = struct{}{}
			}
		}
	}
	for deptID := range deptSet {
		scope.DeptIDs = append(scope.DeptIDs, deptID)
	}
	sort.Slice(scope.DeptIDs, func(i, j int) bool { return scope.DeptIDs[i] < scope.DeptIDs[j] })
	return scope, nil
}

// ResolveVisibleDeptIDs 解析用户可见部门 ID 列表
func ResolveVisibleDeptIDs(ctx context.Context, profile *UserAccessProfile, currentDeptID int64) (deptIDs []int64, all bool, err error) {
	scope, err := ResolveUserDataScope(ctx, profile, currentDeptID)
	if err != nil {
		return nil, false, err
	}
	if scope == nil || scope.All {
		return nil, true, nil
	}

	deptSet := make(map[int64]struct{}, len(scope.DeptIDs)+1)
	for _, deptID := range scope.DeptIDs {
		if deptID == 0 {
			continue
		}
		deptSet[deptID] = struct{}{}
	}
	if scope.IncludeSelf && currentDeptID > 0 {
		deptSet[currentDeptID] = struct{}{}
	}

	deptIDs = make([]int64, 0, len(deptSet))
	for deptID := range deptSet {
		deptIDs = append(deptIDs, deptID)
	}
	sort.Slice(deptIDs, func(i, j int) bool { return deptIDs[i] < deptIDs[j] })
	return deptIDs, false, nil
}

// LoadRoleMenuIDs 加载角色关联的菜单 ID 列表
func LoadRoleMenuIDs(ctx context.Context, roleIDs []int64) ([]int64, error) {
	roleIDs = UniqueNonZeroIDs(roleIDs)
	if len(roleIDs) == 0 {
		return make([]int64, 0), nil
	}

	var roleMenus []struct {
		MenuId int64 `json:"menuId"`
	}
	if err := dao.RoleMenu.Ctx(ctx).
		WhereIn(dao.RoleMenu.Columns().RoleId, roleIDs).
		Scan(&roleMenus); err != nil {
		return nil, err
	}

	menuIDs := make([]int64, 0, len(roleMenus))
	seen := make(map[int64]struct{}, len(roleMenus))
	for _, roleMenu := range roleMenus {
		if _, ok := seen[roleMenu.MenuId]; ok {
			continue
		}
		seen[roleMenu.MenuId] = struct{}{}
		menuIDs = append(menuIDs, roleMenu.MenuId)
	}
	return menuIDs, nil
}

// LoadAllMenuPermissions 加载全部菜单权限标识
func LoadAllMenuPermissions(ctx context.Context) ([]string, error) {
	return loadMenuPermissions(ctx, nil)
}

// LoadMenuPermissions 加载指定菜单的权限标识
func LoadMenuPermissions(ctx context.Context, menuIDs []int64) ([]string, error) {
	if len(menuIDs) == 0 {
		return make([]string, 0), nil
	}
	return loadMenuPermissions(ctx, menuIDs)
}

func loadMenuPermissions(ctx context.Context, menuIDs []int64) ([]string, error) {
	var perms []struct {
		Permission string `json:"permission"`
	}

	query := g.DB().Ctx(ctx).Model("system_menu").
		Fields("permission").
		Where("deleted_at", nil).
		Where("status", 1).
		WhereNot("permission", "")
	if len(menuIDs) > 0 {
		query = query.Where("id", menuIDs)
	}
	if err := query.OrderAsc("id").Scan(&perms); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(perms))
	seen := make(map[string]struct{}, len(perms))
	for _, perm := range perms {
		if perm.Permission == "" {
			continue
		}
		if _, ok := seen[perm.Permission]; ok {
			continue
		}
		seen[perm.Permission] = struct{}{}
		result = append(result, perm.Permission)
	}
	return result, nil
}

// LoadRoleTitleMap 批量加载角色名称映射
func LoadRoleTitleMap(ctx context.Context, roleIDs []int64) map[int64]string {
	return LoadTitleMap(ctx, "system_role", roleIDs)
}

// LoadAllAuthMenus 加载全部菜单（超管用）
func LoadAllAuthMenus(ctx context.Context) ([]*model.AuthMenuOutput, error) {
	return loadAuthMenus(ctx, nil)
}

// LoadAuthMenus 加载指定菜单
func LoadAuthMenus(ctx context.Context, menuIDs []int64) ([]*model.AuthMenuOutput, error) {
	if len(menuIDs) == 0 {
		return make([]*model.AuthMenuOutput, 0), nil
	}
	return loadAuthMenus(ctx, menuIDs)
}

func loadAuthMenus(ctx context.Context, menuIDs []int64) ([]*model.AuthMenuOutput, error) {
	var list []*model.AuthMenuOutput

	query := g.DB().Ctx(ctx).Model("system_menu").
		Where("deleted_at", nil).
		Where("status", 1)
	if len(menuIDs) > 0 {
		query = query.Where("id", menuIDs)
	}
	if err := query.OrderAsc("sort").OrderAsc("id").Scan(&list); err != nil {
		return nil, err
	}
	return list, nil
}

// BuildAuthMenuTree 组装菜单树
func BuildAuthMenuTree(list []*model.AuthMenuOutput) []*model.AuthMenuOutput {
	nodeMap := make(map[int64]*model.AuthMenuOutput, len(list))
	for _, item := range list {
		item.Children = make([]*model.AuthMenuOutput, 0)
		nodeMap[int64(item.ID)] = item
	}
	tree := make([]*model.AuthMenuOutput, 0)
	for _, item := range list {
		if int64(item.ParentID) == 0 {
			tree = append(tree, item)
		} else if parent, ok := nodeMap[int64(item.ParentID)]; ok {
			parent.Children = append(parent.Children, item)
		} else {
			tree = append(tree, item)
		}
	}
	return tree
}
