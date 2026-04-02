package conversation

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
	service.RegisterConversation(New())
}

func New() *sConversation {
	return &sConversation{}
}

type sConversation struct{}

// Create 创建MVP对话表
func (s *sConversation) Create(ctx context.Context, in *model.ConversationCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpConversation.Ctx(ctx).Data(g.Map{
		dao.MvpConversation.Columns().Id:        id,
		dao.MvpConversation.Columns().ProjectId: in.ProjectID,
		dao.MvpConversation.Columns().TaskId: in.TaskID,
		dao.MvpConversation.Columns().Title: in.Title,
		dao.MvpConversation.Columns().RoleType: in.RoleType,
		dao.MvpConversation.Columns().Status: in.Status,
		dao.MvpConversation.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.MvpConversation.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.MvpConversation.Columns().CreatedAt: gtime.Now(),
		dao.MvpConversation.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新MVP对话表
func (s *sConversation) Update(ctx context.Context, in *model.ConversationUpdateInput) error {
	data := g.Map{
		dao.MvpConversation.Columns().ProjectId: in.ProjectID,
		dao.MvpConversation.Columns().TaskId: in.TaskID,
		dao.MvpConversation.Columns().Title: in.Title,
		dao.MvpConversation.Columns().RoleType: in.RoleType,
		dao.MvpConversation.Columns().Status: in.Status,
		dao.MvpConversation.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.MvpConversation.Ctx(ctx).Where(dao.MvpConversation.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除MVP对话表
func (s *sConversation) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.MvpConversation.Ctx(ctx).Where(dao.MvpConversation.Columns().Id, id).Data(g.Map{
		dao.MvpConversation.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除MVP对话表
func (s *sConversation) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.MvpConversation.Ctx(ctx).WhereIn(dao.MvpConversation.Columns().Id, ids).Data(g.Map{
		dao.MvpConversation.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取MVP对话表详情
func (s *sConversation) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ConversationDetailOutput, err error) {
	out = &model.ConversationDetailOutput{}
	err = dao.MvpConversation.Ctx(ctx).Where(dao.MvpConversation.Columns().Id, id).Where(dao.MvpConversation.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询项目ID关联显示
	if out.ProjectID != 0 {
		val, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", out.ProjectID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.ProjectName = val.String()
		}
	}
	// 查询关联任务ID，NULL=项目级对话关联显示
	if out.TaskID != 0 {
		val, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", out.TaskID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.TaskName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sConversation) applyListFilter(ctx context.Context, in *model.ConversationListInput) *gdb.Model {
	m := dao.MvpConversation.Ctx(ctx).Where(dao.MvpConversation.Columns().DeletedAt, nil)
	if in.Title != "" {
		m = m.WhereLike(dao.MvpConversation.Columns().Title, "%"+in.Title+"%")
	}
	if in.ProjectID > 0 {
		m = m.Where(dao.MvpConversation.Columns().ProjectId, in.ProjectID)
	}
	if in.RoleType != "" {
		m = m.Where(dao.MvpConversation.Columns().RoleType, in.RoleType)
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpConversation.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpConversation.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpConversation.Columns().CreatedBy, dao.MvpConversation.Columns().DeptId)
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sConversation) fillRefFields(ctx context.Context, list []*model.ConversationListOutput) {
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
			if item.TaskID != 0 {
				idSet[int64(item.TaskID)] = struct{}{}
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
					if val, ok := refMap[int64(item.TaskID)]; ok {
						item.TaskName = val
					}
				}
			}
		}
	}
}

// List 获取MVP对话表列表
func (s *sConversation) List(ctx context.Context, in *model.ConversationListInput) (list []*model.ConversationListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序
	if in.OrderBy != "" {
		if in.OrderDir == "desc" {
			m = m.OrderDesc(in.OrderBy)
		} else {
			m = m.OrderAsc(in.OrderBy)
		}
	} else {
		m = m.OrderAsc(dao.MvpConversation.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出MVP对话表（不分页）
func (s *sConversation) Export(ctx context.Context, in *model.ConversationListInput) (list []*model.ConversationListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpConversation.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}



// BatchUpdate 批量编辑MVP对话表
func (s *sConversation) BatchUpdate(ctx context.Context, in *model.ConversationBatchUpdateInput) error {
	data := g.Map{
		dao.MvpConversation.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpConversation.Columns().Status] = *in.Status
	}
	_, err := dao.MvpConversation.Ctx(ctx).WhereIn(dao.MvpConversation.Columns().Id, in.IDs).Data(data).Update()
	return err
}


// Import 导入MVP对话表
func (s *sConversation) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.MvpConversation.Columns().Id:        id,
			dao.MvpConversation.Columns().CreatedAt: gtime.Now(),
			dao.MvpConversation.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpConversation.Columns().ProjectId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConversation.Columns().TaskId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConversation.Columns().Title] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConversation.Columns().RoleType] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConversation.Columns().Status] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpConversation.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

