package provider

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
	service.RegisterProvider(New())
}

func New() *sProvider {
	return &sProvider{}
}

type sProvider struct{}

// Create 创建AI供应商表
func (s *sProvider) Create(ctx context.Context, in *model.ProviderCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.AiProvider.Ctx(ctx).Data(g.Map{
		dao.AiProvider.Columns().Id:        id,
		dao.AiProvider.Columns().Name: in.Name,
		dao.AiProvider.Columns().Code: in.Code,
		dao.AiProvider.Columns().ProviderType: in.ProviderType,
		dao.AiProvider.Columns().BaseUrl: in.BaseURL,
		dao.AiProvider.Columns().Icon: in.Icon,
		dao.AiProvider.Columns().Status: in.Status,
		dao.AiProvider.Columns().Sort: in.Sort,
		dao.AiProvider.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.AiProvider.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.AiProvider.Columns().CreatedAt: gtime.Now(),
		dao.AiProvider.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新AI供应商表
func (s *sProvider) Update(ctx context.Context, in *model.ProviderUpdateInput) error {
	data := g.Map{
		dao.AiProvider.Columns().Name: in.Name,
		dao.AiProvider.Columns().Code: in.Code,
		dao.AiProvider.Columns().ProviderType: in.ProviderType,
		dao.AiProvider.Columns().BaseUrl: in.BaseURL,
		dao.AiProvider.Columns().Icon: in.Icon,
		dao.AiProvider.Columns().Status: in.Status,
		dao.AiProvider.Columns().Sort: in.Sort,
		dao.AiProvider.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.AiProvider.Ctx(ctx).Where(dao.AiProvider.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除AI供应商表
func (s *sProvider) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.AiProvider.Ctx(ctx).Where(dao.AiProvider.Columns().Id, id).Data(g.Map{
		dao.AiProvider.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除AI供应商表
func (s *sProvider) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.AiProvider.Ctx(ctx).WhereIn(dao.AiProvider.Columns().Id, ids).Data(g.Map{
		dao.AiProvider.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取AI供应商表详情
func (s *sProvider) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProviderDetailOutput, err error) {
	out = &model.ProviderDetailOutput{}
	err = dao.AiProvider.Ctx(ctx).Where(dao.AiProvider.Columns().Id, id).Where(dao.AiProvider.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sProvider) applyListFilter(ctx context.Context, in *model.ProviderListInput) *gdb.Model {
	m := dao.AiProvider.Ctx(ctx).Where(dao.AiProvider.Columns().DeletedAt, nil)
	if in.Status != nil {
		m = m.Where(dao.AiProvider.Columns().Status, *in.Status)
	}
	if in.Name != "" {
		m = m.WhereLike(dao.AiProvider.Columns().Name, "%"+in.Name+"%")
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.AiProvider.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.AiProvider.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.AiProvider.Columns().CreatedBy, dao.AiProvider.Columns().DeptId)
	return m
}

// List 获取AI供应商表列表
func (s *sProvider) List(ctx context.Context, in *model.ProviderListInput) (list []*model.ProviderListOutput, total int, err error) {
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
		m = m.OrderAsc(dao.AiProvider.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	return
}
// Export 导出AI供应商表（不分页）
func (s *sProvider) Export(ctx context.Context, in *model.ProviderListInput) (list []*model.ProviderListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.AiProvider.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	return
}



// BatchUpdate 批量编辑AI供应商表
func (s *sProvider) BatchUpdate(ctx context.Context, in *model.ProviderBatchUpdateInput) error {
	data := g.Map{
		dao.AiProvider.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.AiProvider.Columns().Status] = *in.Status
	}
	_, err := dao.AiProvider.Ctx(ctx).WhereIn(dao.AiProvider.Columns().Id, in.IDs).Data(data).Update()
	return err
}


// Import 导入AI供应商表
func (s *sProvider) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.AiProvider.Columns().Id:        id,
			dao.AiProvider.Columns().CreatedAt: gtime.Now(),
			dao.AiProvider.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.AiProvider.Columns().Name] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiProvider.Columns().Code] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiProvider.Columns().ProviderType] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiProvider.Columns().BaseUrl] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiProvider.Columns().Icon] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiProvider.Columns().Status] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.AiProvider.Columns().Sort] = record[idx]
		}
		idx++
		if _, insertErr := dao.AiProvider.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

