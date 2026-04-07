package rolepreset

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
	service.RegisterRolePreset(New())
}

func New() *sRolePreset {
	return &sRolePreset{}
}

type sRolePreset struct{}

// Create 创建角色预设模板
func (s *sRolePreset) Create(ctx context.Context, in *model.RolePresetCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpRolePreset.Ctx(ctx).Data(g.Map{
		dao.MvpRolePreset.Columns().Id:        id,
		dao.MvpRolePreset.Columns().ProjectCategory: in.ProjectCategory,
		dao.MvpRolePreset.Columns().RoleType: in.RoleType,
		dao.MvpRolePreset.Columns().RoleLevel: in.RoleLevel,
		dao.MvpRolePreset.Columns().ModelId: in.ModelID,
		dao.MvpRolePreset.Columns().SystemPrompt:  in.SystemPrompt,
		dao.MvpRolePreset.Columns().ExecutionMode: in.ExecutionMode,
		dao.MvpRolePreset.Columns().Status:        in.Status,
		dao.MvpRolePreset.Columns().Sort:          in.Sort,
		dao.MvpRolePreset.Columns().CreatedBy:     middleware.GetUserID(ctx),
		dao.MvpRolePreset.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.MvpRolePreset.Columns().CreatedAt: gtime.Now(),
		dao.MvpRolePreset.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新角色预设模板
func (s *sRolePreset) Update(ctx context.Context, in *model.RolePresetUpdateInput) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().DeletedAt, nil), in.ID, dao.MvpRolePreset.Columns().Id, dao.MvpRolePreset.Columns().CreatedBy); err != nil {
		return err
	}
	data := g.Map{
		dao.MvpRolePreset.Columns().ProjectCategory: in.ProjectCategory,
		dao.MvpRolePreset.Columns().RoleType: in.RoleType,
		dao.MvpRolePreset.Columns().RoleLevel: in.RoleLevel,
		dao.MvpRolePreset.Columns().ModelId: in.ModelID,
		dao.MvpRolePreset.Columns().SystemPrompt:  in.SystemPrompt,
		dao.MvpRolePreset.Columns().ExecutionMode: in.ExecutionMode,
		dao.MvpRolePreset.Columns().Status: in.Status,
		dao.MvpRolePreset.Columns().Sort: in.Sort,
		dao.MvpRolePreset.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除角色预设模板
func (s *sRolePreset) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().DeletedAt, nil), id, dao.MvpRolePreset.Columns().Id, dao.MvpRolePreset.Columns().CreatedBy); err != nil {
		return err
	}
	_, err := dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().Id, id).Data(g.Map{
		dao.MvpRolePreset.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除角色预设模板
func (s *sRolePreset) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	m := dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().DeletedAt, nil).WhereIn(dao.MvpRolePreset.Columns().Id, ids)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpRolePreset.Columns().CreatedBy, dao.MvpRolePreset.Columns().DeptId)
	_, err := m.Data(g.Map{
		dao.MvpRolePreset.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取角色预设模板详情
func (s *sRolePreset) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.RolePresetDetailOutput, err error) {
	out = &model.RolePresetDetailOutput{}
	err = dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().Id, id).Where(dao.MvpRolePreset.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, fmt.Errorf("记录不存在")
	}
	// 填充模型名称
	if out.ModelID > 0 {
		modelRecord, _ := g.DB().Model("ai_model").Fields("name").Where("id", out.ModelID).Where("deleted_at IS NULL").One()
		if !modelRecord.IsEmpty() {
			out.ModelName = modelRecord["name"].String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sRolePreset) applyListFilter(ctx context.Context, in *model.RolePresetListInput) *gdb.Model {
	m := dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().DeletedAt, nil)
	if in.Status != nil {
		m = m.Where(dao.MvpRolePreset.Columns().Status, *in.Status)
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpRolePreset.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpRolePreset.Columns().CreatedAt, in.EndTime)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpRolePreset.Columns().CreatedBy, dao.MvpRolePreset.Columns().DeptId)
	return m
}

// List 获取角色预设模板列表
func (s *sRolePreset) List(ctx context.Context, in *model.RolePresetListInput) (list []*model.RolePresetListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序（白名单防 SQL 注入）
	allowedOrderBy := map[string]bool{"id": true, "status": true, "role_type": true, "sort": true, "created_at": true, "updated_at": true}
	if in.OrderBy != "" && allowedOrderBy[in.OrderBy] {
		if in.OrderDir == "desc" {
			m = m.OrderDesc(in.OrderBy)
		} else {
			m = m.OrderAsc(in.OrderBy)
		}
	} else {
		m = m.OrderAsc(dao.MvpRolePreset.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}

// fillRefFields 填充关联字段（模型名称）
func (s *sRolePreset) fillRefFields(ctx context.Context, list []*model.RolePresetListOutput) {
	if len(list) == 0 {
		return
	}
	// 收集 model_id
	modelIDs := make([]int64, 0, len(list))
	for _, item := range list {
		if item.ModelID > 0 {
			modelIDs = append(modelIDs, int64(item.ModelID))
		}
	}
	if len(modelIDs) == 0 {
		return
	}
	// 查模型名称
	models, _ := g.DB().Model("ai_model").
		Fields("id, name").
		WhereIn("id", modelIDs).
		Where("deleted_at IS NULL").
		All()
	modelMap := make(map[int64]string)
	for _, m := range models {
		modelMap[m["id"].Int64()] = m["name"].String()
	}
	for _, item := range list {
		if name, ok := modelMap[int64(item.ModelID)]; ok {
			item.ModelName = name
		}
	}
}

// Export 导出角色预设模板（不分页）
func (s *sRolePreset) Export(ctx context.Context, in *model.RolePresetListInput) (list []*model.RolePresetListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpRolePreset.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}



// BatchUpdate 批量编辑角色预设模板
func (s *sRolePreset) BatchUpdate(ctx context.Context, in *model.RolePresetBatchUpdateInput) error {
	data := g.Map{
		dao.MvpRolePreset.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpRolePreset.Columns().Status] = *in.Status
	}
	m := dao.MvpRolePreset.Ctx(ctx).Where(dao.MvpRolePreset.Columns().DeletedAt, nil).WhereIn(dao.MvpRolePreset.Columns().Id, in.IDs)
	m = middleware.ApplyDataScope(ctx, m, dao.MvpRolePreset.Columns().CreatedBy, dao.MvpRolePreset.Columns().DeptId)
	_, err := m.Data(data).Update()
	return err
}


// Import 导入角色预设模板
func (s *sRolePreset) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.MvpRolePreset.Columns().Id:        id,
			dao.MvpRolePreset.Columns().CreatedBy: middleware.GetUserID(ctx),
			dao.MvpRolePreset.Columns().DeptId:    middleware.GetDeptID(ctx),
			dao.MvpRolePreset.Columns().CreatedAt: gtime.Now(),
			dao.MvpRolePreset.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpRolePreset.Columns().RoleType] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpRolePreset.Columns().RoleLevel] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpRolePreset.Columns().ModelId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpRolePreset.Columns().SystemPrompt] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpRolePreset.Columns().Status] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpRolePreset.Columns().Sort] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpRolePreset.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

