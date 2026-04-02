package model

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/dao"
	"easymvp/app/ai/internal/middleware"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterModel(New())
}

func New() *sModel {
	return &sModel{}
}

type sModel struct{}

// Create 创建AI模型表
func (s *sModel) Create(ctx context.Context, in *model.ModelCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.AiModel.Ctx(ctx).Data(g.Map{
		dao.AiModel.Columns().Id:        id,
		dao.AiModel.Columns().PlanId: in.PlanID,
		dao.AiModel.Columns().ProviderId: in.ProviderID,
		dao.AiModel.Columns().Name: in.Name,
		dao.AiModel.Columns().ModelCode: in.ModelCode,
		dao.AiModel.Columns().Capability: in.Capability,
		dao.AiModel.Columns().MaxTokens: in.MaxTokens,
		dao.AiModel.Columns().ContextWindow: in.ContextWindow,
		dao.AiModel.Columns().SupportsStream: in.SupportsStream,
		dao.AiModel.Columns().RolePrompt: in.RolePrompt,
		dao.AiModel.Columns().Status: in.Status,
		dao.AiModel.Columns().Sort: in.Sort,
		dao.AiModel.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.AiModel.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.AiModel.Columns().CreatedAt: gtime.Now(),
		dao.AiModel.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新AI模型表
func (s *sModel) Update(ctx context.Context, in *model.ModelUpdateInput) error {
	data := g.Map{
		dao.AiModel.Columns().PlanId: in.PlanID,
		dao.AiModel.Columns().ProviderId: in.ProviderID,
		dao.AiModel.Columns().Name: in.Name,
		dao.AiModel.Columns().ModelCode: in.ModelCode,
		dao.AiModel.Columns().Capability: in.Capability,
		dao.AiModel.Columns().MaxTokens: in.MaxTokens,
		dao.AiModel.Columns().ContextWindow: in.ContextWindow,
		dao.AiModel.Columns().SupportsStream: in.SupportsStream,
		dao.AiModel.Columns().RolePrompt: in.RolePrompt,
		dao.AiModel.Columns().Status: in.Status,
		dao.AiModel.Columns().Sort: in.Sort,
		dao.AiModel.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.AiModel.Ctx(ctx).Where(dao.AiModel.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除AI模型表
func (s *sModel) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.AiModel.Ctx(ctx).Where(dao.AiModel.Columns().Id, id).Data(g.Map{
		dao.AiModel.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除AI模型表
func (s *sModel) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.AiModel.Ctx(ctx).WhereIn(dao.AiModel.Columns().Id, ids).Data(g.Map{
		dao.AiModel.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取AI模型表详情
func (s *sModel) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ModelDetailOutput, err error) {
	out = &model.ModelDetailOutput{}
	err = dao.AiModel.Ctx(ctx).Where(dao.AiModel.Columns().Id, id).Where(dao.AiModel.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询套餐ID关联显示
	if out.PlanID != 0 {
		val, err := g.DB().Ctx(ctx).Model("ai_plan").Where("id", out.PlanID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.PlanName = val.String()
		}
	}
	// 查询供应商ID（冗余便于查询）关联显示
	if out.ProviderID != 0 {
		val, err := g.DB().Ctx(ctx).Model("ai_provider").Where("id", out.ProviderID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.ProviderName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sModel) applyListFilter(ctx context.Context, in *model.ModelListInput) *gdb.Model {
	m := dao.AiModel.Ctx(ctx).Where(dao.AiModel.Columns().DeletedAt, nil)
	if in.SupportsStream != nil {
		m = m.Where(dao.AiModel.Columns().SupportsStream, *in.SupportsStream)
	}
	if in.Status != nil {
		m = m.Where(dao.AiModel.Columns().Status, *in.Status)
	}
	if in.Name != "" {
		m = m.WhereLike(dao.AiModel.Columns().Name, "%"+in.Name+"%")
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.AiModel.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.AiModel.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.AiModel.Columns().CreatedBy, dao.AiModel.Columns().DeptId)
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sModel) fillRefFields(ctx context.Context, list []*model.ModelListOutput) {
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.PlanID != 0 {
				idSet[int64(item.PlanID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("ai_plan").
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
					if val, ok := refMap[int64(item.PlanID)]; ok {
						item.PlanName = val
					}
				}
			}
		}
	}
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.ProviderID != 0 {
				idSet[int64(item.ProviderID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("ai_provider").
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
					if val, ok := refMap[int64(item.ProviderID)]; ok {
						item.ProviderName = val
					}
				}
			}
		}
	}
}

// List 获取AI模型表列表
func (s *sModel) List(ctx context.Context, in *model.ModelListInput) (list []*model.ModelListOutput, total int, err error) {
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
		m = m.OrderAsc(dao.AiModel.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出AI模型表（不分页）
func (s *sModel) Export(ctx context.Context, in *model.ModelListInput) (list []*model.ModelListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.AiModel.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}



// BatchUpdate 批量编辑AI模型表
func (s *sModel) BatchUpdate(ctx context.Context, in *model.ModelBatchUpdateInput) error {
	data := g.Map{
		dao.AiModel.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.AiModel.Columns().Status] = *in.Status
	}
	_, err := dao.AiModel.Ctx(ctx).WhereIn(dao.AiModel.Columns().Id, in.IDs).Data(data).Update()
	return err
}


// Import 导入AI模型表
func (s *sModel) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.AiModel.Columns().Id:        id,
			dao.AiModel.Columns().CreatedAt: gtime.Now(),
			dao.AiModel.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.AiModel.Columns().PlanId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().ProviderId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().Name] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().ModelCode] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().Capability] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().MaxTokens] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().ContextWindow] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().SupportsStream] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().RolePrompt] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().Status] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiModel.Columns().Sort] = record[idx]
		}
		idx++
		if _, insertErr := dao.AiModel.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

