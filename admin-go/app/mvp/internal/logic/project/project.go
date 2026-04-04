package project

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/dao"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterProject(New())
}

func New() *sProject {
	return &sProject{}
}

type sProject struct{}

// Create 创建MVP项目表
func (s *sProject) Create(ctx context.Context, in *model.ProjectCreateInput) error {
	workDir, _, err := engine.EnsureWorkDir(in.WorkDir)
	if err != nil {
		return err
	}

	id := snowflake.Generate()
	_, err = dao.MvpProject.Ctx(ctx).Data(g.Map{
		dao.MvpProject.Columns().Id:               id,
		dao.MvpProject.Columns().Name:             in.Name,
		dao.MvpProject.Columns().ProjectCategory:  in.ProjectCategory,
		dao.MvpProject.Columns().Description:      in.Description,
		dao.MvpProject.Columns().Status:           in.Status,
		dao.MvpProject.Columns().PauseReason:      in.PauseReason,
		dao.MvpProject.Columns().GlobalContext:    in.GlobalContext,
		dao.MvpProject.Columns().ArchitectModelId: in.ArchitectModelID,
		dao.MvpProject.Columns().WorkDir:          workDir,
		dao.MvpProject.Columns().CreatedBy:        middleware.GetUserID(ctx),
		dao.MvpProject.Columns().DeptId:           middleware.GetDeptID(ctx),
		dao.MvpProject.Columns().CreatedAt:        gtime.Now(),
		dao.MvpProject.Columns().UpdatedAt:        gtime.Now(),
	}).Insert()
	return err
}

// Update 更新MVP项目表
func (s *sProject) Update(ctx context.Context, in *model.ProjectUpdateInput) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().DeletedAt, nil), in.ID, dao.MvpProject.Columns().Id, dao.MvpProject.Columns().CreatedBy); err != nil {
		return err
	}

	workDir, _, err := engine.EnsureWorkDir(in.WorkDir)
	if err != nil {
		return err
	}
	data := g.Map{
		dao.MvpProject.Columns().Name:             in.Name,
		dao.MvpProject.Columns().ProjectCategory:  in.ProjectCategory,
		dao.MvpProject.Columns().Description:      in.Description,
		dao.MvpProject.Columns().PauseReason:      in.PauseReason,
		dao.MvpProject.Columns().GlobalContext:    in.GlobalContext,
		dao.MvpProject.Columns().ArchitectModelId: in.ArchitectModelID,
		dao.MvpProject.Columns().WorkDir:          workDir,
		dao.MvpProject.Columns().UpdatedAt:        gtime.Now(),
	}
	_, err = dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除MVP项目表
func (s *sProject) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().DeletedAt, nil), id, dao.MvpProject.Columns().Id, dao.MvpProject.Columns().CreatedBy); err != nil {
		return err
	}
	_, err := dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().Id, id).Data(g.Map{
		dao.MvpProject.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除MVP项目表
func (s *sProject) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	m := dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().DeletedAt, nil).WhereIn(dao.MvpProject.Columns().Id, ids)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProject.Columns().CreatedBy, dao.MvpProject.Columns().DeptId)
	_, err := m.Data(g.Map{
		dao.MvpProject.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取MVP项目表详情
func (s *sProject) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProjectDetailOutput, err error) {
	// 权限校验：只能查看自己创建的项目
	if err = middleware.CheckOwnership(ctx,
		dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().DeletedAt, nil),
		id, dao.MvpProject.Columns().Id, dao.MvpProject.Columns().CreatedBy); err != nil {
		return nil, err
	}

	out = &model.ProjectDetailOutput{}
	err = dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().Id, id).Where(dao.MvpProject.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, fmt.Errorf("记录不存在")
	}
	// 查询架构师AI模型关联显示
	if out.ArchitectModelID != 0 {
		val, err := g.DB().Ctx(ctx).Model("ai_model").Where("id", out.ArchitectModelID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.ArchitectModelName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sProject) applyListFilter(ctx context.Context, in *model.ProjectListInput) *gdb.Model {
	m := dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().DeletedAt, nil)
	if in.Name != "" {
		m = m.WhereLike(dao.MvpProject.Columns().Name, "%"+in.Name+"%")
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpProject.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpProject.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProject.Columns().CreatedBy, dao.MvpProject.Columns().DeptId)
	return m
}

// List 获取MVP项目表列表
func (s *sProject) List(ctx context.Context, in *model.ProjectListInput) (list []*model.ProjectListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序
	if in.OrderBy != "" {
		safeOrderBy := middleware.ValidateOrderBy(in.OrderBy, []string{"id", "name", "status", "created_at", "updated_at"})
		if safeOrderBy != "" {
			if in.OrderDir == "desc" {
				m = m.OrderDesc(safeOrderBy)
			} else {
				m = m.OrderAsc(safeOrderBy)
			}
		}
	} else {
		m = m.OrderAsc(dao.MvpProject.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}

// Export 导出MVP项目表（不分页）
func (s *sProject) Export(ctx context.Context, in *model.ProjectListInput) (list []*model.ProjectListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpProject.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}

// fillRefFields 批量填充关联显示字段
func (s *sProject) fillRefFields(ctx context.Context, list []*model.ProjectListOutput) {
	idSet := make(map[int64]struct{})
	for _, item := range list {
		if item.ArchitectModelID != 0 {
			idSet[int64(item.ArchitectModelID)] = struct{}{}
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
				if val, ok := refMap[int64(item.ArchitectModelID)]; ok {
					item.ArchitectModelName = val
				}
			}
		}
	}
}

// BatchUpdate 批量编辑MVP项目表
func (s *sProject) BatchUpdate(ctx context.Context, in *model.ProjectBatchUpdateInput) error {
	data := g.Map{
		dao.MvpProject.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpProject.Columns().Status] = *in.Status
	}
	m := dao.MvpProject.Ctx(ctx).Where(dao.MvpProject.Columns().DeletedAt, nil).WhereIn(dao.MvpProject.Columns().Id, in.IDs)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProject.Columns().CreatedBy, dao.MvpProject.Columns().DeptId)
	_, err := m.Data(data).Update()
	return err
}

// Import 导入MVP项目表
func (s *sProject) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.MvpProject.Columns().Id:        id,
			dao.MvpProject.Columns().CreatedAt: gtime.Now(),
			dao.MvpProject.Columns().UpdatedAt: gtime.Now(),
			dao.MvpProject.Columns().CreatedBy: middleware.GetUserID(ctx),
			dao.MvpProject.Columns().DeptId:    middleware.GetDeptID(ctx),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpProject.Columns().Name] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProject.Columns().Description] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProject.Columns().Status] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProject.Columns().PauseReason] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProject.Columns().GlobalContext] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProject.Columns().ArchitectModelId] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpProject.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}
