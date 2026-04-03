package projectrole

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/dao"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterProjectRole(New())
}

func New() *sProjectRole {
	return &sProjectRole{}
}

type sProjectRole struct{}

// Create 创建项目角色配置表
func (s *sProjectRole) Create(ctx context.Context, in *model.ProjectRoleCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpProjectRole.Ctx(ctx).Data(g.Map{
		dao.MvpProjectRole.Columns().Id:        id,
		dao.MvpProjectRole.Columns().ProjectId: in.ProjectID,
		dao.MvpProjectRole.Columns().RoleType: in.RoleType,
		dao.MvpProjectRole.Columns().RoleLevel: in.RoleLevel,
		dao.MvpProjectRole.Columns().ModelId: in.ModelID,
		dao.MvpProjectRole.Columns().SystemPrompt: in.SystemPrompt,
		dao.MvpProjectRole.Columns().Status: in.Status,
		dao.MvpProjectRole.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.MvpProjectRole.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.MvpProjectRole.Columns().CreatedAt: gtime.Now(),
		dao.MvpProjectRole.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新项目角色配置表
func (s *sProjectRole) Update(ctx context.Context, in *model.ProjectRoleUpdateInput) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().DeletedAt, nil), in.ID, dao.MvpProjectRole.Columns().Id, dao.MvpProjectRole.Columns().CreatedBy); err != nil {
		return err
	}
	data := g.Map{
		dao.MvpProjectRole.Columns().ProjectId: in.ProjectID,
		dao.MvpProjectRole.Columns().RoleType: in.RoleType,
		dao.MvpProjectRole.Columns().RoleLevel: in.RoleLevel,
		dao.MvpProjectRole.Columns().ModelId: in.ModelID,
		dao.MvpProjectRole.Columns().SystemPrompt: in.SystemPrompt,
		dao.MvpProjectRole.Columns().Status: in.Status,
		dao.MvpProjectRole.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除项目角色配置表
func (s *sProjectRole) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().DeletedAt, nil), id, dao.MvpProjectRole.Columns().Id, dao.MvpProjectRole.Columns().CreatedBy); err != nil {
		return err
	}
	_, err := dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().Id, id).Data(g.Map{
		dao.MvpProjectRole.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除项目角色配置表
func (s *sProjectRole) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	m := dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().DeletedAt, nil).WhereIn(dao.MvpProjectRole.Columns().Id, ids)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProjectRole.Columns().CreatedBy, dao.MvpProjectRole.Columns().DeptId)
	_, err := m.Data(g.Map{
		dao.MvpProjectRole.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取项目角色配置表详情
func (s *sProjectRole) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProjectRoleDetailOutput, err error) {
	out = &model.ProjectRoleDetailOutput{}
	err = dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().Id, id).Where(dao.MvpProjectRole.Columns().DeletedAt, nil).Scan(out)
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
	// 查询AI模型关联显示
	if out.ModelID != 0 {
		val, err := g.DB().Ctx(ctx).Model("ai_model").Where("id", out.ModelID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.ModelName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sProjectRole) applyListFilter(ctx context.Context, in *model.ProjectRoleListInput) *gdb.Model {
	m := dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().DeletedAt, nil)
	if in.Status != nil {
		m = m.Where(dao.MvpProjectRole.Columns().Status, *in.Status)
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpProjectRole.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpProjectRole.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProjectRole.Columns().CreatedBy, dao.MvpProjectRole.Columns().DeptId)
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sProjectRole) fillRefFields(ctx context.Context, list []*model.ProjectRoleListOutput) {
	// 填充项目名称
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
	// 填充AI模型名称
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.ModelID != 0 {
				idSet[int64(item.ModelID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("ai_model").
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
					if val, ok := refMap[int64(item.ModelID)]; ok {
						item.ModelName = val
					}
				}
			}
		}
	}
}

// List 获取项目角色配置表列表
func (s *sProjectRole) List(ctx context.Context, in *model.ProjectRoleListInput) (list []*model.ProjectRoleListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序（白名单防 SQL 注入）
	allowedOrderBy := map[string]bool{"id": true, "status": true, "role_type": true, "created_at": true, "updated_at": true}
	if in.OrderBy != "" && allowedOrderBy[in.OrderBy] {
		if in.OrderDir == "desc" {
			m = m.OrderDesc(in.OrderBy)
		} else {
			m = m.OrderAsc(in.OrderBy)
		}
	} else {
		m = m.OrderAsc(dao.MvpProjectRole.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出项目角色配置表（不分页）
func (s *sProjectRole) Export(ctx context.Context, in *model.ProjectRoleListInput) (list []*model.ProjectRoleListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpProjectRole.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}



// BatchUpdate 批量编辑项目角色配置表
func (s *sProjectRole) BatchUpdate(ctx context.Context, in *model.ProjectRoleBatchUpdateInput) error {
	data := g.Map{
		dao.MvpProjectRole.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpProjectRole.Columns().Status] = *in.Status
	}
	m := dao.MvpProjectRole.Ctx(ctx).Where(dao.MvpProjectRole.Columns().DeletedAt, nil).WhereIn(dao.MvpProjectRole.Columns().Id, in.IDs)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProjectRole.Columns().CreatedBy, dao.MvpProjectRole.Columns().DeptId)
	_, err := m.Data(data).Update()
	return err
}


// Import 导入项目角色配置表
func (s *sProjectRole) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
	f, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// 跳过表头
	if _, err = reader.Read(); err != nil {
		return 0, 0, fmt.Errorf("读取CSV表头失败: %w", err)
	}

	for {
		record, readErr := reader.Read()
		if readErr != nil {
			break
		}
		if len(record) == 0 {
			continue
		}
		// 逐行插入
		id := snowflake.Generate()
		data := g.Map{
			dao.MvpProjectRole.Columns().Id:        id,
			dao.MvpProjectRole.Columns().CreatedBy: middleware.GetUserID(ctx),
			dao.MvpProjectRole.Columns().DeptId:    middleware.GetDeptID(ctx),
			dao.MvpProjectRole.Columns().CreatedAt: gtime.Now(),
			dao.MvpProjectRole.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpProjectRole.Columns().ProjectId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectRole.Columns().RoleType] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectRole.Columns().RoleLevel] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectRole.Columns().ModelId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectRole.Columns().SystemPrompt] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectRole.Columns().Status] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpProjectRole.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

