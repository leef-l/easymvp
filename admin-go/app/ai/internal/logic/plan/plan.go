package plan

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
	service.RegisterPlan(New())
}

func New() *sPlan {
	return &sPlan{}
}

type sPlan struct{}

// Create 创建AI套餐表
func (s *sPlan) Create(ctx context.Context, in *model.PlanCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.AiPlan.Ctx(ctx).Data(g.Map{
		dao.AiPlan.Columns().Id:        id,
		dao.AiPlan.Columns().ProviderId: in.ProviderID,
		dao.AiPlan.Columns().Name: in.Name,
		dao.AiPlan.Columns().Code: in.Code,
		dao.AiPlan.Columns().ApiKey: in.ApiKey,
		dao.AiPlan.Columns().ApiSecret: in.ApiSecret,
		dao.AiPlan.Columns().Status: in.Status,
		dao.AiPlan.Columns().Sort: in.Sort,
		dao.AiPlan.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.AiPlan.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.AiPlan.Columns().CreatedAt: gtime.Now(),
		dao.AiPlan.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新AI套餐表
func (s *sPlan) Update(ctx context.Context, in *model.PlanUpdateInput) error {
	data := g.Map{
		dao.AiPlan.Columns().ProviderId: in.ProviderID,
		dao.AiPlan.Columns().Name: in.Name,
		dao.AiPlan.Columns().Code: in.Code,
		dao.AiPlan.Columns().ApiKey: in.ApiKey,
		dao.AiPlan.Columns().ApiSecret: in.ApiSecret,
		dao.AiPlan.Columns().Status: in.Status,
		dao.AiPlan.Columns().Sort: in.Sort,
		dao.AiPlan.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.AiPlan.Ctx(ctx).Where(dao.AiPlan.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除AI套餐表
func (s *sPlan) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.AiPlan.Ctx(ctx).Where(dao.AiPlan.Columns().Id, id).Data(g.Map{
		dao.AiPlan.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除AI套餐表
func (s *sPlan) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.AiPlan.Ctx(ctx).WhereIn(dao.AiPlan.Columns().Id, ids).Data(g.Map{
		dao.AiPlan.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取AI套餐表详情
func (s *sPlan) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.PlanDetailOutput, err error) {
	out = &model.PlanDetailOutput{}
	err = dao.AiPlan.Ctx(ctx).Where(dao.AiPlan.Columns().Id, id).Where(dao.AiPlan.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询供应商ID关联显示
	if out.ProviderID != 0 {
		val, err := g.DB().Ctx(ctx).Model("ai_provider").Where("id", out.ProviderID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.ProviderName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sPlan) applyListFilter(ctx context.Context, in *model.PlanListInput) *gdb.Model {
	m := dao.AiPlan.Ctx(ctx).Where(dao.AiPlan.Columns().DeletedAt, nil)
	if in.Status != nil {
		m = m.Where(dao.AiPlan.Columns().Status, *in.Status)
	}
	if in.Name != "" {
		m = m.WhereLike(dao.AiPlan.Columns().Name, "%"+in.Name+"%")
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.AiPlan.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.AiPlan.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.AiPlan.Columns().CreatedBy, dao.AiPlan.Columns().DeptId)
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sPlan) fillRefFields(ctx context.Context, list []*model.PlanListOutput) {
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

// List 获取AI套餐表列表
func (s *sPlan) List(ctx context.Context, in *model.PlanListInput) (list []*model.PlanListOutput, total int, err error) {
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
		m = m.OrderAsc(dao.AiPlan.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出AI套餐表（不分页）
func (s *sPlan) Export(ctx context.Context, in *model.PlanListInput) (list []*model.PlanListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.AiPlan.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}



// BatchUpdate 批量编辑AI套餐表
func (s *sPlan) BatchUpdate(ctx context.Context, in *model.PlanBatchUpdateInput) error {
	data := g.Map{
		dao.AiPlan.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.AiPlan.Columns().Status] = *in.Status
	}
	_, err := dao.AiPlan.Ctx(ctx).WhereIn(dao.AiPlan.Columns().Id, in.IDs).Data(data).Update()
	return err
}


// Import 导入AI套餐表
func (s *sPlan) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.AiPlan.Columns().Id:        id,
			dao.AiPlan.Columns().CreatedAt: gtime.Now(),
			dao.AiPlan.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.AiPlan.Columns().ProviderId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiPlan.Columns().Name] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiPlan.Columns().Code] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiPlan.Columns().ApiKey] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiPlan.Columns().ApiSecret] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiPlan.Columns().Status] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiPlan.Columns().Sort] = record[idx]
		}
		idx++
		if _, insertErr := dao.AiPlan.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

