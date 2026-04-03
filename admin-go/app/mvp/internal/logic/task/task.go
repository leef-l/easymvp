package task

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/dao"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterTask(New())
}

func New() *sTask {
	return &sTask{}
}

type sTask struct{}

// Create 创建MVP任务表
func (s *sTask) Create(ctx context.Context, in *model.TaskCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpTask.Ctx(ctx).Data(g.Map{
		dao.MvpTask.Columns().Id:        id,
		dao.MvpTask.Columns().ProjectId: in.ProjectID,
		dao.MvpTask.Columns().ParentId: in.ParentID,
		dao.MvpTask.Columns().Name: in.Name,
		dao.MvpTask.Columns().Description: in.Description,
		dao.MvpTask.Columns().RoleType: in.RoleType,
		dao.MvpTask.Columns().RoleLevel: in.RoleLevel,
		dao.MvpTask.Columns().ModelId: in.ModelID,
		dao.MvpTask.Columns().Status: in.Status,
		dao.MvpTask.Columns().Sort: in.Sort,
		dao.MvpTask.Columns().BatchNo: in.BatchNo,
		dao.MvpTask.Columns().AffectedResources: in.AffectedResources,
		dao.MvpTask.Columns().DependsOn: in.DependsOn,
		dao.MvpTask.Columns().Result: in.Result,
		dao.MvpTask.Columns().ContextSummary: in.ContextSummary,
		dao.MvpTask.Columns().ErrorMessage: in.ErrorMessage,
		dao.MvpTask.Columns().StartedAt: in.StartedAt,
		dao.MvpTask.Columns().CompletedAt: in.CompletedAt,
		dao.MvpTask.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.MvpTask.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.MvpTask.Columns().CreatedAt: gtime.Now(),
		dao.MvpTask.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新MVP任务表
func (s *sTask) Update(ctx context.Context, in *model.TaskUpdateInput) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().DeletedAt, nil), in.ID, dao.MvpTask.Columns().Id, dao.MvpTask.Columns().CreatedBy); err != nil {
		return err
	}
	data := g.Map{
		dao.MvpTask.Columns().ProjectId: in.ProjectID,
		dao.MvpTask.Columns().ParentId: in.ParentID,
		dao.MvpTask.Columns().Name: in.Name,
		dao.MvpTask.Columns().Description: in.Description,
		dao.MvpTask.Columns().RoleType: in.RoleType,
		dao.MvpTask.Columns().RoleLevel: in.RoleLevel,
		dao.MvpTask.Columns().ModelId: in.ModelID,
		dao.MvpTask.Columns().Status: in.Status,
		dao.MvpTask.Columns().Sort: in.Sort,
		dao.MvpTask.Columns().BatchNo: in.BatchNo,
		dao.MvpTask.Columns().AffectedResources: in.AffectedResources,
		dao.MvpTask.Columns().DependsOn: in.DependsOn,
		dao.MvpTask.Columns().Result: in.Result,
		dao.MvpTask.Columns().ContextSummary: in.ContextSummary,
		dao.MvpTask.Columns().ErrorMessage: in.ErrorMessage,
		dao.MvpTask.Columns().StartedAt: in.StartedAt,
		dao.MvpTask.Columns().CompletedAt: in.CompletedAt,
		dao.MvpTask.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除MVP任务表
func (s *sTask) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().DeletedAt, nil), id, dao.MvpTask.Columns().Id, dao.MvpTask.Columns().CreatedBy); err != nil {
		return err
	}
	_, err := dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().Id, id).Data(g.Map{
		dao.MvpTask.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除MVP任务表
func (s *sTask) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	m := middleware.ApplyDataScope(ctx, dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().DeletedAt, nil).WhereIn(dao.MvpTask.Columns().Id, ids), dao.MvpTask.Columns().CreatedBy, dao.MvpTask.Columns().DeptId)
	_, err := m.Data(g.Map{
		dao.MvpTask.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取MVP任务表详情
func (s *sTask) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.TaskDetailOutput, err error) {
	out = &model.TaskDetailOutput{}
	err = dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().Id, id).Where(dao.MvpTask.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, fmt.Errorf("记录不存在")
	}
	// 查询项目ID关联显示
	if out.ProjectID != 0 {
		val, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", out.ProjectID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.ProjectName = val.String()
		}
	}
	// 查询父任务ID，0=顶级关联显示
	if out.ParentID != 0 {
		val, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", out.ParentID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.TaskName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sTask) applyListFilter(ctx context.Context, in *model.TaskListInput) *gdb.Model {
	m := dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().DeletedAt, nil)
	if in.ProjectID > 0 {
		m = m.Where(dao.MvpTask.Columns().ProjectId, in.ProjectID)
	}
	if in.Name != "" {
		m = m.WhereLike(dao.MvpTask.Columns().Name, "%"+in.Name+"%")
	}
	if in.Status != "" {
		m = m.Where(dao.MvpTask.Columns().Status, in.Status)
	}
	if in.BatchNo != nil {
		m = m.Where(dao.MvpTask.Columns().BatchNo, *in.BatchNo)
	}
	if in.RoleType != "" {
		m = m.Where(dao.MvpTask.Columns().RoleType, in.RoleType)
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpTask.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpTask.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpTask.Columns().CreatedBy, dao.MvpTask.Columns().DeptId)
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sTask) fillRefFields(ctx context.Context, list []*model.TaskListOutput) {
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.ProjectID != 0 {
				idSet[int64(item.ProjectID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("mvp_project").
				Fields("id", "name").
				Where("deleted_at", nil).
				WhereIn("id", ids).
				All()
			if err == nil {
				refMap := make(map[int64]string, len(rows))
				for _, row := range rows {
					refMap[row["id"].Int64()] = row["name"].String()
				}
				for _, item := range list {
					if val, ok := refMap[int64(item.ProjectID)]; ok {
						item.ProjectName = val
					}
				}
			}
		}
	}
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.ParentID != 0 {
				idSet[int64(item.ParentID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("mvp_task").
				Fields("id", "name").
				Where("deleted_at", nil).
				WhereIn("id", ids).
				All()
			if err == nil {
				refMap := make(map[int64]string, len(rows))
				for _, row := range rows {
					refMap[row["id"].Int64()] = row["name"].String()
				}
				for _, item := range list {
					if val, ok := refMap[int64(item.ParentID)]; ok {
						item.TaskName = val
					}
				}
			}
		}
	}
}

// List 获取MVP任务表列表
func (s *sTask) List(ctx context.Context, in *model.TaskListInput) (list []*model.TaskListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序
	if in.OrderBy != "" {
		safeOrderBy := middleware.ValidateOrderBy(in.OrderBy, []string{"id", "name", "status", "batch_no", "role_type", "sort", "created_at", "updated_at"})
		if safeOrderBy != "" {
			if in.OrderDir == "desc" {
				m = m.OrderDesc(safeOrderBy)
			} else {
				m = m.OrderAsc(safeOrderBy)
			}
		}
	} else {
		m = m.OrderAsc(dao.MvpTask.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出MVP任务表（不分页）
func (s *sTask) Export(ctx context.Context, in *model.TaskListInput) (list []*model.TaskListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpTask.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}


// Tree 获取MVP任务表树形结构
func (s *sTask) Tree(ctx context.Context, in *model.TaskTreeInput) (tree []*model.TaskTreeOutput, err error) {
	var list []*model.TaskTreeOutput
	m := dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().DeletedAt, nil)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpTask.Columns().CreatedBy, dao.MvpTask.Columns().DeptId)
	if in.ProjectID > 0 {
		m = m.Where(dao.MvpTask.Columns().ProjectId, in.ProjectID)
	}
	if in.Name != "" {
		m = m.WhereLike(dao.MvpTask.Columns().Name, "%"+in.Name+"%")
	}
	if in.Status != "" {
		m = m.Where(dao.MvpTask.Columns().Status, in.Status)
	}
	if in.BatchNo > 0 {
		m = m.Where(dao.MvpTask.Columns().BatchNo, in.BatchNo)
	}
	if in.RoleType != "" {
		m = m.Where(dao.MvpTask.Columns().RoleType, in.RoleType)
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpTask.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpTask.Columns().CreatedAt, in.EndTime)
	}
	err = m.OrderAsc(dao.MvpTask.Columns().Sort).Scan(&list)
	if err != nil {
		return
	}

	// 使用 map 迭代方式组装树
	nodeMap := make(map[int64]*model.TaskTreeOutput, len(list))
	for _, item := range list {
		item.Children = make([]*model.TaskTreeOutput, 0)
		nodeMap[int64(item.ID)] = item
	}

	tree = make([]*model.TaskTreeOutput, 0)
	for _, item := range list {
		if int64(item.ParentID) == 0 {
			tree = append(tree, item)
		} else if parent, ok := nodeMap[int64(item.ParentID)]; ok {
			parent.Children = append(parent.Children, item)
		}
	}
	return
}


// BatchUpdate 批量编辑MVP任务表
func (s *sTask) BatchUpdate(ctx context.Context, in *model.TaskBatchUpdateInput) error {
	data := g.Map{
		dao.MvpTask.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpTask.Columns().Status] = *in.Status
	}
	m := middleware.ApplyDataScope(ctx, dao.MvpTask.Ctx(ctx).Where(dao.MvpTask.Columns().DeletedAt, nil).WhereIn(dao.MvpTask.Columns().Id, in.IDs), dao.MvpTask.Columns().CreatedBy, dao.MvpTask.Columns().DeptId)
	_, err := m.Data(data).Update()
	return err
}


