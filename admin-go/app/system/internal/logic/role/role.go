package role

import (
	"context"

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
	service.RegisterRole(New())
}

func New() *sRole {
	return &sRole{}
}

type sRole struct{}

// Create 创建角色表
func (s *sRole) Create(ctx context.Context, in *model.RoleCreateInput) error {
	// 禁止通过 API 创建超级管理员角色
	in.IsAdmin = 0
	id := snowflake.Generate()
	_, err := dao.Role.Ctx(ctx).Data(g.Map{
		dao.Role.Columns().Id:        id,
		dao.Role.Columns().ParentId:  in.ParentID,
		dao.Role.Columns().Title:     in.Title,
		dao.Role.Columns().DataScope: in.DataScope,
		"default_ai_engine":          in.DefaultAiEngine,
		dao.Role.Columns().Sort:      in.Sort,
		dao.Role.Columns().Status:    in.Status,
		dao.Role.Columns().IsAdmin:   0,
		dao.Role.Columns().CreatedAt: gtime.Now(),
		dao.Role.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新角色表
func (s *sRole) Update(ctx context.Context, in *model.RoleUpdateInput) error {
	// 禁止通过 API 修改 IsAdmin 字段
	data := g.Map{
		dao.Role.Columns().ParentId:  in.ParentID,
		dao.Role.Columns().Title:     in.Title,
		dao.Role.Columns().DataScope: in.DataScope,
		"default_ai_engine":          in.DefaultAiEngine,
		dao.Role.Columns().Sort:      in.Sort,
		dao.Role.Columns().Status:    in.Status,
		dao.Role.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除角色表
func (s *sRole) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	// 检查是否有用户仍关联此角色
	count, err := dao.UserRole.Ctx(ctx).Where(dao.UserRole.Columns().RoleId, id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.Newf("该角色下还有 %d 个用户，请先解除关联", count)
	}
	_, err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, id).Data(g.Map{
		dao.Role.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取角色表详情
func (s *sRole) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.RoleDetailOutput, err error) {
	out = &model.RoleDetailOutput{}
	err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, id).Where(dao.Role.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询上级角色ID，0 表示顶级角色关联显示
	if out.ParentID != 0 {
		val, _ := g.DB().Ctx(ctx).Model("system_role").Where("id", out.ParentID).Where("deleted_at", nil).Value("title")
		out.RoleTitle = val.String()
	}
	return
}

// List 获取角色表列表
func (s *sRole) List(ctx context.Context, in *model.RoleListInput) (list []*model.RoleListOutput, total int, err error) {
	m := dao.Role.Ctx(ctx).Where(dao.Role.Columns().DeletedAt, nil)
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
	err = m.Page(in.PageNum, in.PageSize).OrderAsc(dao.Role.Columns().Id).Scan(&list)
	if err != nil {
		return
	}
	// 填充关联显示字段
	for _, item := range list {
		if item.ParentID != 0 {
			val, _ := g.DB().Ctx(ctx).Model("system_role").Where("id", item.ParentID).Where("deleted_at", nil).Value("title")
			item.RoleTitle = val.String()
		}
	}
	return
}

// Tree 获取角色表树形结构
func (s *sRole) Tree(ctx context.Context) (tree []*model.RoleTreeOutput, err error) {
	var list []*model.RoleTreeOutput
	err = dao.Role.Ctx(ctx).Where(dao.Role.Columns().DeletedAt, nil).OrderAsc(dao.Role.Columns().Sort).Scan(&list)
	if err != nil {
		return
	}

	// 使用 map 迭代方式组装树
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

// GrantMenu 角色授权菜单（先删后插，事务保护）
func (s *sRole) GrantMenu(ctx context.Context, in *model.RoleGrantMenuInput) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 删除旧的关联
		_, err := dao.RoleMenu.Ctx(ctx).Where(dao.RoleMenu.Columns().RoleId, in.ID).Delete()
		if err != nil {
			return err
		}
		// 批量插入新关联
		if len(in.MenuIDs) > 0 {
			data := make([]g.Map, 0, len(in.MenuIDs))
			for _, menuID := range in.MenuIDs {
				data = append(data, g.Map{
					dao.RoleMenu.Columns().RoleId: in.ID,
					dao.RoleMenu.Columns().MenuId: menuID,
				})
			}
			_, err = dao.RoleMenu.Ctx(ctx).Data(data).Insert()
		}
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

// GrantDept 角色授权数据权限（事务保护）
func (s *sRole) GrantDept(ctx context.Context, in *model.RoleGrantDeptInput) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新角色的 data_scope
		_, err := dao.Role.Ctx(ctx).Where(dao.Role.Columns().Id, in.ID).Data(g.Map{
			dao.Role.Columns().DataScope: in.DataScope,
		}).Update()
		if err != nil {
			return err
		}
		// 删除旧的部门关联
		_, err = dao.RoleDept.Ctx(ctx).Where(dao.RoleDept.Columns().RoleId, in.ID).Delete()
		if err != nil {
			return err
		}
		// 自定义数据权限时，插入部门关联
		if in.DataScope == 5 && len(in.DeptIDs) > 0 {
			data := make([]g.Map, 0, len(in.DeptIDs))
			for _, deptID := range in.DeptIDs {
				data = append(data, g.Map{
					dao.RoleDept.Columns().RoleId: in.ID,
					dao.RoleDept.Columns().DeptId: deptID,
				})
			}
			_, err = dao.RoleDept.Ctx(ctx).Data(data).Insert()
		}
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
