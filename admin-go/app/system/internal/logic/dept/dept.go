package dept

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
	service.RegisterDept(New())
}

func New() *sDept {
	return &sDept{}
}

type sDept struct{}

// Create 创建部门表
func (s *sDept) Create(ctx context.Context, in *model.DeptCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.Dept.Ctx(ctx).Data(g.Map{
		dao.Dept.Columns().Id:        id,
		dao.Dept.Columns().ParentId: in.ParentID,
		dao.Dept.Columns().Title: in.Title,
		dao.Dept.Columns().Username: in.Username,
		dao.Dept.Columns().Email: in.Email,
		dao.Dept.Columns().Sort: in.Sort,
		dao.Dept.Columns().Status: in.Status,
		dao.Dept.Columns().CreatedAt: gtime.Now(),
		dao.Dept.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新部门表
func (s *sDept) Update(ctx context.Context, in *model.DeptUpdateInput) error {
	data := g.Map{
		dao.Dept.Columns().ParentId: in.ParentID,
		dao.Dept.Columns().Title: in.Title,
		dao.Dept.Columns().Username: in.Username,
		dao.Dept.Columns().Email: in.Email,
		dao.Dept.Columns().Sort: in.Sort,
		dao.Dept.Columns().Status: in.Status,
		dao.Dept.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.Dept.Ctx(ctx).Where(dao.Dept.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除部门表
func (s *sDept) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.Dept.Ctx(ctx).Where(dao.Dept.Columns().Id, id).Data(g.Map{
		dao.Dept.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取部门表详情
func (s *sDept) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.DeptDetailOutput, err error) {
	out = &model.DeptDetailOutput{}
	err = dao.Dept.Ctx(ctx).Where(dao.Dept.Columns().Id, id).Where(dao.Dept.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询上级部门ID，0 表示顶级部门关联显示
	if out.ParentID != 0 {
		val, _ := g.DB().Ctx(ctx).Model("system_dept").Where("id", out.ParentID).Where("deleted_at", nil).Value("title")
		out.DeptTitle = val.String()
	}
	return
}

// List 获取部门表列表
func (s *sDept) List(ctx context.Context, in *model.DeptListInput) (list []*model.DeptListOutput, total int, err error) {
	m := dao.Dept.Ctx(ctx).Where(dao.Dept.Columns().DeletedAt, nil)
	if in.Status > 0 {
		m = m.Where(dao.Dept.Columns().Status, in.Status)
	}
	total, err = m.Count()
	if err != nil {
		return
	}
	err = m.Page(in.PageNum, in.PageSize).OrderAsc(dao.Dept.Columns().Id).Scan(&list)
	if err != nil {
		return
	}
	// 填充关联显示字段
	for _, item := range list {
		if item.ParentID != 0 {
			val, _ := g.DB().Ctx(ctx).Model("system_dept").Where("id", item.ParentID).Where("deleted_at", nil).Value("title")
			item.DeptTitle = val.String()
		}
	}
	return
}

// Tree 获取部门表树形结构
func (s *sDept) Tree(ctx context.Context) (tree []*model.DeptTreeOutput, err error) {
	var list []*model.DeptTreeOutput
	err = dao.Dept.Ctx(ctx).Where(dao.Dept.Columns().DeletedAt, nil).OrderAsc(dao.Dept.Columns().Sort).Scan(&list)
	if err != nil {
		return
	}

	// 使用 map 迭代方式组装树
	nodeMap := make(map[int64]*model.DeptTreeOutput, len(list))
	for _, item := range list {
		item.Children = make([]*model.DeptTreeOutput, 0)
		nodeMap[int64(item.ID)] = item
	}

	tree = make([]*model.DeptTreeOutput, 0)
	for _, item := range list {
		if int64(item.ParentID) == 0 {
			tree = append(tree, item)
		} else if parent, ok := nodeMap[int64(item.ParentID)]; ok {
			parent.Children = append(parent.Children, item)
		}
	}
	return
}

