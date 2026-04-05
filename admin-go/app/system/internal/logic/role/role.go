package role

import (
	"context"
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
	service.RegisterRole(New())
}

func New() *sRole {
	return &sRole{}
}

type sRole struct{}

// Create 创建角色
func (s *sRole) Create(ctx context.Context, in *model.RoleCreateInput) error {
	s.normalizeRoleInput(in)
	if err := s.validateRoleInput(in.Title, in.DataScope, in.Status, in.IsAdmin, in.Sort); err != nil {
		return err
	}
	if err := s.ensureRoleTitleUnique(ctx, in.Title, 0); err != nil {
		return err
	}
	if err := s.ensureParentRoleValid(ctx, in.ParentID, 0); err != nil {
		return err
	}
	id := snowflake.Generate()
	_, err := dao.Role.Ctx(ctx).Data(g.Map{
		dao.Role.Columns().Id:        id,
		dao.Role.Columns().ParentId:  in.ParentID,
		dao.Role.Columns().Title:     in.Title,
		dao.Role.Columns().DataScope: in.DataScope,
		"default_ai_engine":          in.DefaultAiEngine,
		dao.Role.Columns().Sort:      in.Sort,
		dao.Role.Columns().Status:    in.Status,
		dao.Role.Columns().IsAdmin:   in.IsAdmin,
		dao.Role.Columns().CreatedAt: gtime.Now(),
		dao.Role.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新角色
func (s *sRole) Update(ctx context.Context, in *model.RoleUpdateInput) error {
	s.normalizeRoleInput(in)
	currentRole, err := s.getRoleDetail(ctx, in.ID)
	if err != nil {
		return err
	}
	if err := s.validateRoleInput(in.Title, in.DataScope, in.Status, in.IsAdmin, in.Sort); err != nil {
		return err
	}
	if err := s.ensureRoleTitleUnique(ctx, in.Title, int64(in.ID)); err != nil {
		return err
	}
	if err := s.ensureParentRoleValid(ctx, in.ParentID, int64(in.ID)); err != nil {
		return err
	}
	if err := s.ensureAdminRoleEditable(currentRole, in); err != nil {
		return err
	}
	data := g.Map{
		dao.Role.Columns().ParentId:  in.ParentID,
		dao.Role.Columns().Title:     in.Title,
		dao.Role.Columns().DataScope: in.DataScope,
		"default_ai_engine":          in.DefaultAiEngine,
		dao.Role.Columns().Sort:      in.Sort,
		dao.Role.Columns().Status:    in.Status,
		dao.Role.Columns().IsAdmin:   in.IsAdmin,
		dao.Role.Columns().UpdatedAt: gtime.Now(),
	}
	_, err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除角色
func (s *sRole) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	roleDetail, err := s.getRoleDetail(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ensureSuperAdminRoleMutable(roleDetail.IsAdmin, "delete"); err != nil {
		return err
	}
	if err := s.ensureRoleDeletable(ctx, id); err != nil {
		return err
	}
	_, err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, id).Data(g.Map{
		dao.Role.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取角色详情
func (s *sRole) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.RoleDetailOutput, err error) {
	out, err = s.getRoleDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	// 填充上级角色名称
	if out.ParentID != 0 {
		out.RoleTitle = support.LoadTitle(ctx, "system_role", int64(out.ParentID))
	}
	return
}

func (s *sRole) getRoleDetail(ctx context.Context, id snowflake.JsonInt64) (out *model.RoleDetailOutput, err error) {
	out = &model.RoleDetailOutput{}
	err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, id).Where(dao.Role.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, gerror.New("角色不存在")
	}
	return out, nil
}

// List 获取角色列表
func (s *sRole) List(ctx context.Context, in *model.RoleListInput) (list []*model.RoleListOutput, total int, err error) {
	in.Title = strings.TrimSpace(in.Title)
	m := dao.Role.Ctx(ctx).Where(dao.Role.Columns().DeletedAt, nil)
	if in.Title != "" {
		m = m.WhereLike(dao.Role.Columns().Title, "%"+in.Title+"%")
	}
	if in.DataScope > 0 {
		m = m.Where(dao.Role.Columns().DataScope, in.DataScope)
	}
	if in.Status > 0 {
		m = m.Where(dao.Role.Columns().Status, in.Status)
	}
	total, err = m.Count()
	if err != nil {
		return
	}
	err = m.Page(in.PageNum, in.PageSize).OrderAsc(dao.Role.Columns().Sort).OrderAsc(dao.Role.Columns().Id).Scan(&list)
	if err != nil {
		return
	}
	// 批量填充上级角色名称
	parentIDs := make([]int64, 0, len(list))
	for _, item := range list {
		if item.ParentID != 0 {
			parentIDs = append(parentIDs, int64(item.ParentID))
		}
	}
	if len(parentIDs) > 0 {
		titleMap := support.LoadTitleMap(ctx, "system_role", parentIDs)
		for _, item := range list {
			if item.ParentID != 0 {
				item.RoleTitle = titleMap[int64(item.ParentID)]
			}
		}
	}
	return
}

// Tree 获取角色树形结构
func (s *sRole) Tree(ctx context.Context) (tree []*model.RoleTreeOutput, err error) {
	var list []*model.RoleTreeOutput
	err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().DeletedAt, nil).OrderAsc(dao.Role.Columns().Sort).Scan(&list)
	if err != nil {
		return
	}

	nodeMap := make(map[int64]*model.RoleTreeOutput, len(list))
	for _, item := range list {
		item.Children = make([]*model.RoleTreeOutput, 0)
		nodeMap[int64(item.ID)] = item
	}

	tree = make([]*model.RoleTreeOutput, 0)
	for _, item := range list {
		if int64(item.ParentID) == 0 {
			tree = append(tree, item)
		} else if parent, ok := nodeMap[int64(item.ParentID)]; ok {
			parent.Children = append(parent.Children, item)
		}
	}
	return
}

// GrantMenu 角色授权菜单（先删后插）
func (s *sRole) GrantMenu(ctx context.Context, in *model.RoleGrantMenuInput) error {
	roleDetail, err := s.getRoleDetail(ctx, in.ID)
	if err != nil {
		return err
	}
	if err := s.ensureSuperAdminRoleMutable(roleDetail.IsAdmin, "grant-menu"); err != nil {
		return err
	}
	menuIDs := uniqueJSONInt64(in.MenuIDs)
	if err := s.ensureMenuIDsExist(ctx, menuIDs); err != nil {
		return err
	}
	return dao.Role.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.RoleMenu.Table()).Ctx(ctx).
			Where(dao.RoleMenu.Columns().RoleId, in.ID).
			Delete()
		if err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		data := make([]g.Map, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			data = append(data, g.Map{
				dao.RoleMenu.Columns().RoleId: in.ID,
				dao.RoleMenu.Columns().MenuId: menuID,
			})
		}
		_, err = tx.Model(dao.RoleMenu.Table()).Ctx(ctx).Data(data).Insert()
		return err
	})
}

// GetMenuIDs 获取角色已授权的菜单ID列表
func (s *sRole) GetMenuIDs(ctx context.Context, roleID snowflake.JsonInt64) ([]snowflake.JsonInt64, error) {
	var list []struct {
		MenuId int64 `json:"menuId"`
	}
	err := dao.RoleMenu.Ctx(ctx).Where(dao.RoleMenu.Columns().RoleId, roleID).Scan(&list)
	if err != nil {
		return nil, err
	}
	ids := make([]snowflake.JsonInt64, 0, len(list))
	for _, item := range list {
		ids = append(ids, snowflake.JsonInt64(item.MenuId))
	}
	return ids, nil
}

// GrantDept 角色授权数据权限
func (s *sRole) GrantDept(ctx context.Context, in *model.RoleGrantDeptInput) error {
	roleDetail, err := s.getRoleDetail(ctx, in.ID)
	if err != nil {
		return err
	}
	if err := s.ensureSuperAdminRoleMutable(roleDetail.IsAdmin, "grant-dept"); err != nil {
		return err
	}
	deptIDs := uniqueJSONInt64(in.DeptIDs)
	if in.DataScope == consts.RoleDataScope自定义 {
		if err := s.ensureDeptIDsExist(ctx, deptIDs); err != nil {
			return err
		}
	}
	return dao.Role.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.Role.Table()).Ctx(ctx).
			Where(dao.Role.Columns().Id, in.ID).
			Data(g.Map{dao.Role.Columns().DataScope: in.DataScope}).
			Update()
		if err != nil {
			return err
		}
		_, err = tx.Model(dao.RoleDept.Table()).Ctx(ctx).
			Where(dao.RoleDept.Columns().RoleId, in.ID).
			Delete()
		if err != nil {
			return err
		}
		if in.DataScope != consts.RoleDataScope自定义 || len(deptIDs) == 0 {
			return nil
		}
		data := make([]g.Map, 0, len(deptIDs))
		for _, deptID := range deptIDs {
			data = append(data, g.Map{
				dao.RoleDept.Columns().RoleId: in.ID,
				dao.RoleDept.Columns().DeptId: deptID,
			})
		}
		_, err = tx.Model(dao.RoleDept.Table()).Ctx(ctx).Data(data).Insert()
		return err
	})
}

// GetDeptIDs 获取角色已授权的部门ID列表
func (s *sRole) GetDeptIDs(ctx context.Context, roleID snowflake.JsonInt64) ([]snowflake.JsonInt64, error) {
	var list []struct {
		DeptId int64 `json:"deptId"`
	}
	err := dao.RoleDept.Ctx(ctx).Where(dao.RoleDept.Columns().RoleId, roleID).Scan(&list)
	if err != nil {
		return nil, err
	}
	ids := make([]snowflake.JsonInt64, 0, len(list))
	for _, item := range list {
		ids = append(ids, snowflake.JsonInt64(item.DeptId))
	}
	return ids, nil
}

// GrantAiEngine 角色授权AI执行引擎（先删后插）
func (s *sRole) GrantAiEngine(ctx context.Context, in *model.RoleGrantAiEngineInput) error {
	_, err := g.DB().Ctx(ctx).Model("system_role_ai_engine").Where("role_id", in.ID).Delete()
	if err != nil {
		return err
	}

	if len(in.EngineCodes) == 0 {
		return nil
	}

	data := make([]g.Map, 0, len(in.EngineCodes))
	for _, engineCode := range in.EngineCodes {
		if engineCode == "" {
			continue
		}
		data = append(data, g.Map{
			"role_id":     in.ID,
			"engine_code": engineCode,
		})
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Ctx(ctx).Model("system_role_ai_engine").Data(data).Insert()
	return err
}

// GetAiEngineCodes 获取角色已授权AI执行引擎编码列表
func (s *sRole) GetAiEngineCodes(ctx context.Context, roleID snowflake.JsonInt64) ([]string, error) {
	var list []struct {
		EngineCode string `json:"engineCode"`
	}
	err := g.DB().Ctx(ctx).Model("system_role_ai_engine").Where("role_id", roleID).Scan(&list)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(list))
	for _, item := range list {
		if item.EngineCode != "" {
			codes = append(codes, item.EngineCode)
		}
	}
	return codes, nil
}

// ---- 内部辅助方法 ----

func uniqueJSONInt64(ids []snowflake.JsonInt64) []snowflake.JsonInt64 {
	if len(ids) == 0 {
		return make([]snowflake.JsonInt64, 0)
	}
	result := make([]snowflake.JsonInt64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		raw := int64(id)
		if raw == 0 {
			continue
		}
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (s *sRole) ensureParentRoleValid(ctx context.Context, parentID snowflake.JsonInt64, currentID int64) error {
	if parentID == 0 {
		return nil
	}
	if int64(parentID) == currentID && currentID > 0 {
		return gerror.New("上级角色不能是自己")
	}
	visited := map[int64]struct{}{}
	nextParentID := int64(parentID)
	for nextParentID != 0 {
		if nextParentID == currentID && currentID > 0 {
			return gerror.New("上级角色不能设置为自己的下级角色")
		}
		if _, ok := visited[nextParentID]; ok {
			return gerror.New("角色层级存在循环引用")
		}
		visited[nextParentID] = struct{}{}

		record := struct {
			ParentID int64 `json:"parentID"`
		}{}
		err := dao.Role.Ctx(ctx).
			Fields(dao.Role.Columns().ParentId).
			Where(dao.Role.Columns().Id, nextParentID).
			Where(dao.Role.Columns().DeletedAt, nil).
			Scan(&record)
		if err != nil {
			return err
		}
		if record.ParentID == 0 {
			if nextParentID != int64(parentID) {
				return nil
			}
			exists, err := dao.Role.Ctx(ctx).
				Where(dao.Role.Columns().Id, nextParentID).
				Where(dao.Role.Columns().DeletedAt, nil).
				Count()
			if err != nil {
				return err
			}
			if exists == 0 {
				return gerror.New("上级角色不存在")
			}
			return nil
		}
		nextParentID = record.ParentID
	}
	return nil
}

func (s *sRole) ensureRoleTitleUnique(ctx context.Context, title string, excludeID int64) error {
	if title == "" {
		return gerror.New("角色名称不能为空")
	}
	query := dao.Role.Ctx(ctx).
		Where(dao.Role.Columns().Title, title).
		Where(dao.Role.Columns().DeletedAt, nil)
	if excludeID > 0 {
		query = query.WhereNot(dao.Role.Columns().Id, excludeID)
	}
	count, err := query.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("角色名称已存在")
	}
	return nil
}

func (s *sRole) ensureAdminRoleEditable(current *model.RoleDetailOutput, in *model.RoleUpdateInput) error {
	if current.IsAdmin != consts.RoleIsAdmin是 {
		return nil
	}
	if in.IsAdmin != consts.RoleIsAdmin是 {
		return gerror.New("超级管理员角色不能取消超级管理员标记")
	}
	if in.Status == consts.RoleStatus关闭 {
		return gerror.New("超级管理员角色不能禁用")
	}
	return nil
}

func (s *sRole) ensureSuperAdminRoleMutable(isAdmin int, action string) error {
	if isAdmin != consts.RoleIsAdmin是 {
		return nil
	}
	switch action {
	case "delete":
		return gerror.New("超级管理员角色不能删除")
	case "grant-menu":
		return gerror.New("超级管理员角色默认拥有全部菜单权限，无需单独授权")
	case "grant-dept":
		return gerror.New("超级管理员角色默认拥有全部数据权限，无需单独授权")
	default:
		return gerror.New("超级管理员角色不允许执行该操作")
	}
}

func (s *sRole) ensureMenuIDsExist(ctx context.Context, menuIDs []snowflake.JsonInt64) error {
	if len(menuIDs) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		ids = append(ids, int64(menuID))
	}
	count, err := dao.Menu.Ctx(ctx).
		WhereIn(dao.Menu.Columns().Id, ids).
		Where(dao.Menu.Columns().DeletedAt, nil).
		Count()
	if err != nil {
		return err
	}
	if count != len(ids) {
		return gerror.New("关联菜单不存在")
	}
	return nil
}

func (s *sRole) ensureDeptIDsExist(ctx context.Context, deptIDs []snowflake.JsonInt64) error {
	if len(deptIDs) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(deptIDs))
	for _, deptID := range deptIDs {
		ids = append(ids, int64(deptID))
	}
	count, err := dao.Dept.Ctx(ctx).
		WhereIn(dao.Dept.Columns().Id, ids).
		Where(dao.Dept.Columns().DeletedAt, nil).
		Count()
	if err != nil {
		return err
	}
	if count != len(ids) {
		return gerror.New("关联部门不存在")
	}
	return nil
}

func (s *sRole) ensureRoleDeletable(ctx context.Context, id snowflake.JsonInt64) error {
	childCount, err := dao.Role.Ctx(ctx).
		Where(dao.Role.Columns().ParentId, id).
		Where(dao.Role.Columns().DeletedAt, nil).
		Count()
	if err != nil {
		return err
	}
	if childCount > 0 {
		return gerror.New("当前角色下存在子角色，不能删除")
	}

	userRoleCount, err := dao.UserRole.Ctx(ctx).
		Where(dao.UserRole.Columns().RoleId, id).
		Count()
	if err != nil {
		return err
	}
	if userRoleCount > 0 {
		return gerror.New("当前角色已分配给用户，不能删除")
	}
	return nil
}

func (s *sRole) normalizeRoleInput(in interface{}) {
	switch v := in.(type) {
	case *model.RoleCreateInput:
		v.Title = strings.TrimSpace(v.Title)
	case *model.RoleUpdateInput:
		v.Title = strings.TrimSpace(v.Title)
	}
}

func (s *sRole) validateRoleInput(title string, dataScope, status, isAdmin, sort int) error {
	if title == "" {
		return gerror.New("角色名称不能为空")
	}
	if !consts.IsValidRoleDataScope(dataScope) {
		return gerror.New("数据权限范围不合法")
	}
	if !consts.IsValidRoleStatus(status) {
		return gerror.New("角色状态不合法")
	}
	if !consts.IsValidRoleIsAdmin(isAdmin) {
		return gerror.New("管理员标记不合法")
	}
	if sort < 0 {
		return gerror.New("排序值不能小于0")
	}
	return nil
}
