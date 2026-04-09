package projectcategory

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

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
	service.RegisterProjectCategory(New())
}

func New() *sProjectCategory {
	return &sProjectCategory{}
}

type sProjectCategory struct{}

// Create 创建项目分类配置表
func (s *sProjectCategory) Create(ctx context.Context, in *model.ProjectCategoryCreateInput) error {
	verificationProfileJSON, err := normalizeOptionalJSON(in.VerificationProfileJson, "分类默认验证配置")
	if err != nil {
		return err
	}
	verificationGateJSON, err := normalizeOptionalJSON(in.VerificationGateJson, "分类验证放行规则")
	if err != nil {
		return err
	}
	id := snowflake.Generate()
	_, err = dao.MvpProjectCategory.Ctx(ctx).Data(g.Map{
		dao.MvpProjectCategory.Columns().Id:                      id,
		dao.MvpProjectCategory.Columns().CategoryCode:            in.CategoryCode,
		dao.MvpProjectCategory.Columns().DisplayName:             in.DisplayName,
		dao.MvpProjectCategory.Columns().FamilyCode:              in.FamilyCode,
		dao.MvpProjectCategory.Columns().Description:             in.Description,
		dao.MvpProjectCategory.Columns().VerificationProfileJson: verificationProfileJSON,
		dao.MvpProjectCategory.Columns().VerificationGateJson:    verificationGateJSON,
		dao.MvpProjectCategory.Columns().Status:                  in.Status,
		dao.MvpProjectCategory.Columns().Sort:                    in.Sort,
		dao.MvpProjectCategory.Columns().CreatedBy:               middleware.GetUserID(ctx),
		dao.MvpProjectCategory.Columns().DeptId:                  middleware.GetDeptID(ctx),
		dao.MvpProjectCategory.Columns().CreatedAt:               gtime.Now(),
		dao.MvpProjectCategory.Columns().UpdatedAt:               gtime.Now(),
	}).Insert()
	return err
}

// Update 更新项目分类配置表
func (s *sProjectCategory) Update(ctx context.Context, in *model.ProjectCategoryUpdateInput) error {
	verificationProfileJSON, err := normalizeOptionalJSON(in.VerificationProfileJson, "分类默认验证配置")
	if err != nil {
		return err
	}
	verificationGateJSON, err := normalizeOptionalJSON(in.VerificationGateJson, "分类验证放行规则")
	if err != nil {
		return err
	}
	data := g.Map{
		dao.MvpProjectCategory.Columns().CategoryCode:            in.CategoryCode,
		dao.MvpProjectCategory.Columns().DisplayName:             in.DisplayName,
		dao.MvpProjectCategory.Columns().FamilyCode:              in.FamilyCode,
		dao.MvpProjectCategory.Columns().Description:             in.Description,
		dao.MvpProjectCategory.Columns().VerificationProfileJson: verificationProfileJSON,
		dao.MvpProjectCategory.Columns().VerificationGateJson:    verificationGateJSON,
		dao.MvpProjectCategory.Columns().Status:                  in.Status,
		dao.MvpProjectCategory.Columns().Sort:                    in.Sort,
		dao.MvpProjectCategory.Columns().UpdatedAt:               gtime.Now(),
	}
	_, err = dao.MvpProjectCategory.Ctx(ctx).Where(dao.MvpProjectCategory.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除项目分类配置表
func (s *sProjectCategory) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.MvpProjectCategory.Ctx(ctx).Where(dao.MvpProjectCategory.Columns().Id, id).Data(g.Map{
		dao.MvpProjectCategory.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除项目分类配置表
func (s *sProjectCategory) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.MvpProjectCategory.Ctx(ctx).WhereIn(dao.MvpProjectCategory.Columns().Id, ids).Data(g.Map{
		dao.MvpProjectCategory.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取项目分类配置表详情
func (s *sProjectCategory) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.ProjectCategoryDetailOutput, err error) {
	out = &model.ProjectCategoryDetailOutput{}
	err = dao.MvpProjectCategory.Ctx(ctx).Where(dao.MvpProjectCategory.Columns().Id, id).Where(dao.MvpProjectCategory.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sProjectCategory) applyListFilter(ctx context.Context, in *model.ProjectCategoryListInput) *gdb.Model {
	m := dao.MvpProjectCategory.Ctx(ctx).Where(dao.MvpProjectCategory.Columns().DeletedAt, nil)
	if in.DisplayName != "" {
		m = m.WhereLike(dao.MvpProjectCategory.Columns().DisplayName, "%"+in.DisplayName+"%")
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpProjectCategory.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpProjectCategory.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpProjectCategory.Columns().CreatedBy, dao.MvpProjectCategory.Columns().DeptId)
	return m
}

// List 获取项目分类配置表列表
func (s *sProjectCategory) List(ctx context.Context, in *model.ProjectCategoryListInput) (list []*model.ProjectCategoryListOutput, total int, err error) {
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
		m = m.OrderAsc(dao.MvpProjectCategory.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	return
}

// Export 导出项目分类配置表（不分页）
func (s *sProjectCategory) Export(ctx context.Context, in *model.ProjectCategoryListInput) (list []*model.ProjectCategoryListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpProjectCategory.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	return
}

// BatchUpdate 批量编辑项目分类配置表
func (s *sProjectCategory) BatchUpdate(ctx context.Context, in *model.ProjectCategoryBatchUpdateInput) error {
	data := g.Map{
		dao.MvpProjectCategory.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpProjectCategory.Columns().Status] = *in.Status
	}
	_, err := dao.MvpProjectCategory.Ctx(ctx).WhereIn(dao.MvpProjectCategory.Columns().Id, in.IDs).Data(data).Update()
	return err
}

// Import 导入项目分类配置表
func (s *sProjectCategory) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.MvpProjectCategory.Columns().Id:        id,
			dao.MvpProjectCategory.Columns().CreatedAt: gtime.Now(),
			dao.MvpProjectCategory.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpProjectCategory.Columns().CategoryCode] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectCategory.Columns().DisplayName] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectCategory.Columns().FamilyCode] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectCategory.Columns().Description] = record[idx]
		}
		idx++
		if idx < len(record) {
			if value, jsonErr := normalizeOptionalJSON(record[idx], "分类默认验证配置"); jsonErr == nil {
				data[dao.MvpProjectCategory.Columns().VerificationProfileJson] = value
			} else {
				fail++
				continue
			}
		}
		idx++
		if idx < len(record) {
			if value, jsonErr := normalizeOptionalJSON(record[idx], "分类验证放行规则"); jsonErr == nil {
				data[dao.MvpProjectCategory.Columns().VerificationGateJson] = value
			} else {
				fail++
				continue
			}
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectCategory.Columns().Status] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpProjectCategory.Columns().Sort] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpProjectCategory.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

func normalizeOptionalJSON(raw string, label string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", nil
	}
	if !json.Valid([]byte(text)) {
		return "", fmt.Errorf("%s 不是合法 JSON", label)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(text)); err != nil {
		return "", fmt.Errorf("压缩 %s 失败: %w", label, err)
	}
	return compact.String(), nil
}
