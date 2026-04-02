package menu

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/system/internal/dao"
	"easymvp/app/system/internal/model"
	"easymvp/app/system/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterMenu(New())
}

func New() *sMenu {
	return &sMenu{}
}

type sMenu struct{}

// Create 创建菜单表
func (s *sMenu) Create(ctx context.Context, in *model.MenuCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.Menu.Ctx(ctx).Data(g.Map{
		dao.Menu.Columns().Id:        id,
		dao.Menu.Columns().ParentId: in.ParentID,
		dao.Menu.Columns().Title: in.Title,
		dao.Menu.Columns().Type: in.Type,
		dao.Menu.Columns().Path: in.Path,
		dao.Menu.Columns().Component: in.Component,
		dao.Menu.Columns().Permission: in.Permission,
		dao.Menu.Columns().Icon: in.Icon,
		dao.Menu.Columns().Sort: in.Sort,
		dao.Menu.Columns().IsShow: in.IsShow,
		dao.Menu.Columns().IsCache: in.IsCache,
		dao.Menu.Columns().LinkUrl: in.LinkURL,
		dao.Menu.Columns().Status: in.Status,
		dao.Menu.Columns().CreatedAt: gtime.Now(),
		dao.Menu.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新菜单表
func (s *sMenu) Update(ctx context.Context, in *model.MenuUpdateInput) error {
	data := g.Map{
		dao.Menu.Columns().ParentId: in.ParentID,
		dao.Menu.Columns().Title: in.Title,
		dao.Menu.Columns().Type: in.Type,
		dao.Menu.Columns().Path: in.Path,
		dao.Menu.Columns().Component: in.Component,
		dao.Menu.Columns().Permission: in.Permission,
		dao.Menu.Columns().Icon: in.Icon,
		dao.Menu.Columns().Sort: in.Sort,
		dao.Menu.Columns().IsShow: in.IsShow,
		dao.Menu.Columns().IsCache: in.IsCache,
		dao.Menu.Columns().LinkUrl: in.LinkURL,
		dao.Menu.Columns().Status: in.Status,
		dao.Menu.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.Menu.Ctx(ctx).Where(dao.Menu.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除菜单表
func (s *sMenu) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.Menu.Ctx(ctx).Where(dao.Menu.Columns().Id, id).Data(g.Map{
		dao.Menu.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取菜单表详情
func (s *sMenu) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.MenuDetailOutput, err error) {
	out = &model.MenuDetailOutput{}
	err = dao.Menu.Ctx(ctx).Where(dao.Menu.Columns().Id, id).Where(dao.Menu.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询上级菜单ID，0 表示顶级菜单关联显示
	if out.ParentID != 0 {
		val, _ := g.DB().Ctx(ctx).Model("system_menu").Where("id", out.ParentID).Where("deleted_at", nil).Value("title")
		out.MenuTitle = val.String()
	}
	return
}

// List 获取菜单表列表
func (s *sMenu) List(ctx context.Context, in *model.MenuListInput) (list []*model.MenuListOutput, total int, err error) {
	m := dao.Menu.Ctx(ctx).Where(dao.Menu.Columns().DeletedAt, nil)
	if in.Type > 0 {
		m = m.Where(dao.Menu.Columns().Type, in.Type)
	}
	if in.IsShow > 0 {
		m = m.Where(dao.Menu.Columns().IsShow, in.IsShow)
	}
	if in.IsCache > 0 {
		m = m.Where(dao.Menu.Columns().IsCache, in.IsCache)
	}
	if in.Status > 0 {
		m = m.Where(dao.Menu.Columns().Status, in.Status)
	}
	total, err = m.Count()
	if err != nil {
		return
	}
	err = m.Page(in.PageNum, in.PageSize).OrderAsc(dao.Menu.Columns().Id).Scan(&list)
	if err != nil {
		return
	}
	// 填充关联显示字段
	for _, item := range list {
		if item.ParentID != 0 {
			val, _ := g.DB().Ctx(ctx).Model("system_menu").Where("id", item.ParentID).Where("deleted_at", nil).Value("title")
			item.MenuTitle = val.String()
		}
	}
	return
}

// Tree 获取菜单表树形结构
func (s *sMenu) Tree(ctx context.Context) (tree []*model.MenuTreeOutput, err error) {
	var list []*model.MenuTreeOutput
	err = dao.Menu.Ctx(ctx).Where(dao.Menu.Columns().DeletedAt, nil).OrderAsc(dao.Menu.Columns().Sort).Scan(&list)
	if err != nil {
		return
	}

	// 使用 map 迭代方式组装树
	nodeMap := make(map[int64]*model.MenuTreeOutput, len(list))
	for _, item := range list {
		item.Children = make([]*model.MenuTreeOutput, 0)
		nodeMap[int64(item.ID)] = item
	}

	tree = make([]*model.MenuTreeOutput, 0)
	for _, item := range list {
		if int64(item.ParentID) == 0 {
			tree = append(tree, item)
		} else if parent, ok := nodeMap[int64(item.ParentID)]; ok {
			parent.Children = append(parent.Children, item)
		}
	}
	return
}

