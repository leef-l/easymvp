package provider

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/dao"
	"easymvp/app/ai/internal/middleware"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
	providerutil "easymvp/utility/provider"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterProvider(New())
}

func New() *sProvider {
	return &sProvider{}
}

type sProvider struct{}

func normalizeSupportedProtocols(providerType string, supported []string) []string {
	return providerutil.NormalizeProtocols(providerType, supported)
}

func encodeSupportedProtocols(providerType string, supported []string) string {
	list := normalizeSupportedProtocols(providerType, supported)
	if len(list) == 0 {
		return ""
	}
	data, err := json.Marshal(list)
	if err != nil {
		return ""
	}
	return string(data)
}

func decodeSupportedProtocols(raw string, providerType string) []string {
	return providerutil.DecodeSupportedProtocols(raw, providerType)
}

func joinSupportedProtocols(protocols []string) string {
	return strings.Join(protocols, "|")
}

func splitSupportedProtocols(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "|")
}

func buildProviderOutput(record gdb.Record) *model.ProviderListOutput {
	providerType := record[dao.AiProvider.Columns().ProviderType].String()
	return &model.ProviderListOutput{
		ID:                 snowflake.JsonInt64(record[dao.AiProvider.Columns().Id].Int64()),
		Name:               record[dao.AiProvider.Columns().Name].String(),
		Code:               record[dao.AiProvider.Columns().Code].String(),
		ProviderType:       providerType,
		SupportedProtocols: decodeSupportedProtocols(record[dao.AiProvider.Columns().SupportedProtocols].String(), providerType),
		BaseURL:            record[dao.AiProvider.Columns().BaseUrl].String(),
		Icon:               record[dao.AiProvider.Columns().Icon].String(),
		Status:             record[dao.AiProvider.Columns().Status].Int(),
		Sort:               record[dao.AiProvider.Columns().Sort].Int(),
		CreatedAt:          record[dao.AiProvider.Columns().CreatedAt].GTime(),
		UpdatedAt:          record[dao.AiProvider.Columns().UpdatedAt].GTime(),
	}
}

// Create 创建AI供应商表
func (s *sProvider) Create(ctx context.Context, in *model.ProviderCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.AiProvider.Ctx(ctx).Data(g.Map{
		dao.AiProvider.Columns().Id:                 id,
		dao.AiProvider.Columns().Name:               in.Name,
		dao.AiProvider.Columns().Code:               in.Code,
		dao.AiProvider.Columns().ProviderType:       in.ProviderType,
		dao.AiProvider.Columns().SupportedProtocols: encodeSupportedProtocols(in.ProviderType, in.SupportedProtocols),
		dao.AiProvider.Columns().BaseUrl:            in.BaseURL,
		dao.AiProvider.Columns().Icon:               in.Icon,
		dao.AiProvider.Columns().Status:             in.Status,
		dao.AiProvider.Columns().Sort:               in.Sort,
		dao.AiProvider.Columns().CreatedBy:          middleware.GetUserID(ctx),
		dao.AiProvider.Columns().DeptId:             middleware.GetDeptID(ctx),
		dao.AiProvider.Columns().CreatedAt:          gtime.Now(),
		dao.AiProvider.Columns().UpdatedAt:          gtime.Now(),
	}).Insert()
	return err
}

// Update 更新AI供应商表
func (s *sProvider) Update(ctx context.Context, in *model.ProviderUpdateInput) error {
	data := g.Map{
		dao.AiProvider.Columns().Name:               in.Name,
		dao.AiProvider.Columns().Code:               in.Code,
		dao.AiProvider.Columns().ProviderType:       in.ProviderType,
		dao.AiProvider.Columns().SupportedProtocols: encodeSupportedProtocols(in.ProviderType, in.SupportedProtocols),
		dao.AiProvider.Columns().BaseUrl:            in.BaseURL,
		dao.AiProvider.Columns().Icon:               in.Icon,
		dao.AiProvider.Columns().Status:             in.Status,
		dao.AiProvider.Columns().Sort:               in.Sort,
		dao.AiProvider.Columns().UpdatedAt:          gtime.Now(),
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
	record, err := dao.AiProvider.Ctx(ctx).
		Where(dao.AiProvider.Columns().Id, id).
		Where(dao.AiProvider.Columns().DeletedAt, nil).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &model.ProviderDetailOutput{}, nil
	}
	item := buildProviderOutput(record)
	return &model.ProviderDetailOutput{
		ID:                 item.ID,
		Name:               item.Name,
		Code:               item.Code,
		ProviderType:       item.ProviderType,
		SupportedProtocols: item.SupportedProtocols,
		BaseURL:            item.BaseURL,
		Icon:               item.Icon,
		Status:             item.Status,
		Sort:               item.Sort,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}, nil
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
	var records gdb.Result
	records, err = m.Page(in.PageNum, in.PageSize).All()
	if err != nil {
		return
	}
	list = make([]*model.ProviderListOutput, 0, len(records))
	for _, record := range records {
		list = append(list, buildProviderOutput(record))
	}
	return
}

// Export 导出AI供应商表（不分页）
func (s *sProvider) Export(ctx context.Context, in *model.ProviderListInput) (list []*model.ProviderListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	var records gdb.Result
	records, err = m.OrderAsc(dao.AiProvider.Columns().Id).Limit(10000).All()
	if err != nil {
		return nil, err
	}
	list = make([]*model.ProviderListOutput, 0, len(records))
	for _, record := range records {
		list = append(list, buildProviderOutput(record))
	}
	return list, nil
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
			data[dao.AiProvider.Columns().SupportedProtocols] = encodeSupportedProtocols(fmt.Sprintf("%v", data[dao.AiProvider.Columns().ProviderType]), splitSupportedProtocols(record[idx]))
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
