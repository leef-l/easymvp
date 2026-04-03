package config

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
	service.RegisterConfig(New())
}

func New() *sConfig {
	return &sConfig{}
}

type sConfig struct{}

// Create 创建MVP配置表
func (s *sConfig) Create(ctx context.Context, in *model.ConfigCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpConfig.Ctx(ctx).Data(g.Map{
		dao.MvpConfig.Columns().Id:          id,
		dao.MvpConfig.Columns().ConfigKey:   in.ConfigKey,
		dao.MvpConfig.Columns().ConfigValue: in.ConfigValue,
		dao.MvpConfig.Columns().ConfigType:  in.ConfigType,
		dao.MvpConfig.Columns().Category:    in.Category,
		dao.MvpConfig.Columns().Description: in.Description,
		dao.MvpConfig.Columns().CreatedBy:   middleware.GetUserID(ctx),
		dao.MvpConfig.Columns().DeptId:      middleware.GetDeptID(ctx),
		dao.MvpConfig.Columns().CreatedAt:   gtime.Now(),
		dao.MvpConfig.Columns().UpdatedAt:   gtime.Now(),
	}).Insert()
	return err
}

// Update 更新MVP配置表
func (s *sConfig) Update(ctx context.Context, in *model.ConfigUpdateInput) error {
	data := g.Map{
		dao.MvpConfig.Columns().ConfigKey:   in.ConfigKey,
		dao.MvpConfig.Columns().ConfigValue: in.ConfigValue,
		dao.MvpConfig.Columns().ConfigType:  in.ConfigType,
		dao.MvpConfig.Columns().Category:    in.Category,
		dao.MvpConfig.Columns().Description: in.Description,
		dao.MvpConfig.Columns().UpdatedAt:   gtime.Now(),
	}
	_, err := dao.MvpConfig.Ctx(ctx).Where(dao.MvpConfig.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除MVP配置表
func (s *sConfig) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.MvpConfig.Ctx(ctx).Where(dao.MvpConfig.Columns().Id, id).Data(g.Map{
		dao.MvpConfig.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除MVP配置表
func (s *sConfig) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.MvpConfig.Ctx(ctx).WhereIn(dao.MvpConfig.Columns().Id, ids).Data(g.Map{
		dao.MvpConfig.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取MVP配置表详情
func (s *sConfig) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ConfigDetailOutput, err error) {
	out = &model.ConfigDetailOutput{}
	err = dao.MvpConfig.Ctx(ctx).Where(dao.MvpConfig.Columns().Id, id).Where(dao.MvpConfig.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sConfig) applyListFilter(ctx context.Context, in *model.ConfigListInput) *gdb.Model {
	m := dao.MvpConfig.Ctx(ctx).Where(dao.MvpConfig.Columns().DeletedAt, nil)
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpConfig.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpConfig.Columns().CreatedAt, in.EndTime)
	}
	return m
}

// List 获取MVP配置表列表
func (s *sConfig) List(ctx context.Context, in *model.ConfigListInput) (list []*model.ConfigListOutput, total int, err error) {
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
		m = m.OrderAsc(dao.MvpConfig.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	return
}

// Export 导出MVP配置表（不分页）
func (s *sConfig) Export(ctx context.Context, in *model.ConfigListInput) (list []*model.ConfigListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpConfig.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	return
}

// Import 导入MVP配置表
func (s *sConfig) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.MvpConfig.Columns().Id:        id,
			dao.MvpConfig.Columns().CreatedAt: gtime.Now(),
			dao.MvpConfig.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpConfig.Columns().ConfigKey] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConfig.Columns().ConfigValue] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConfig.Columns().ConfigType] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConfig.Columns().Category] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpConfig.Columns().Description] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpConfig.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}
