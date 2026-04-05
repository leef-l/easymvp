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

// presetModel 预设模型数据结构
type presetModel struct {
	Name          string
	ModelCode     string
	MaxTokens     int
	ContextWindow int
}

// insertedModel 已插入模型的关键信息
type insertedModel struct {
	ID        snowflake.JsonInt64
	Name      string
	ModelCode string
	Sort      int
}

// providerModelPresets 按供应商类型的预设模型列表
var providerModelPresets = map[string][]presetModel{
	"openai": {
		{"GPT-4.1", "gpt-4.1", 32768, 1047576},
		{"GPT-4.1 mini", "gpt-4.1-mini", 32768, 1047576},
		{"GPT-4.1 nano", "gpt-4.1-nano", 32768, 1047576},
		{"o3", "o3", 100000, 200000},
		{"o4-mini", "o4-mini", 100000, 200000},
		{"GPT-4o", "gpt-4o", 16384, 128000},
		{"GPT-4o mini", "gpt-4o-mini", 16384, 128000},
	},
	"anthropic": {
		{"Claude Opus 4.6", "claude-opus-4-6", 32000, 200000},
		{"Claude Sonnet 4.6", "claude-sonnet-4-6", 64000, 200000},
		{"Claude Haiku 4.5", "claude-haiku-4-5-20251001", 8096, 200000},
	},
	"google": {
		{"Gemini 2.5 Pro", "gemini-2.5-pro", 65536, 1048576},
		{"Gemini 2.5 Flash", "gemini-2.5-flash", 65536, 1048576},
		{"Gemini 2.0 Flash", "gemini-2.0-flash", 8192, 1048576},
	},
	"deepseek": {
		{"DeepSeek R2", "deepseek-r2", 32768, 128000},
		{"DeepSeek V3", "deepseek-chat", 8192, 64000},
		{"DeepSeek R1", "deepseek-reasoner", 32768, 64000},
	},
	"qwen": {
		{"Qwen Max", "qwen-max", 8192, 32768},
		{"Qwen Plus", "qwen-plus", 8192, 131072},
		{"Qwen Turbo", "qwen-turbo", 8192, 1000000},
		{"Qwen Long", "qwen-long", 8192, 10000000},
	},
	"zhipu": {
		{"GLM-5", "glm-5", 32768, 131072},
		{"GLM-4-Plus", "glm-4-plus", 8192, 128000},
		{"GLM-4-Flash", "glm-4-flash", 8192, 128000},
	},
	"moonshot": {
		{"Kimi K2.5", "kimi-k2.5", 32768, 131072},
		{"Moonshot v1 128k", "moonshot-v1-128k", 8192, 128000},
		{"Moonshot v1 8k", "moonshot-v1-8k", 8192, 8000},
	},
	"minimax": {
		{"MiniMax M2", "minimax-m2", 40960, 1000000},
		{"MiniMax M2.5", "minimax-m2.5", 40960, 1000000},
		{"abab6.5s", "abab6.5s-chat", 8192, 245760},
	},
	"baidu": {
		{"ERNIE 4.5 Turbo", "ernie-4.5-turbo-128k", 4096, 128000},
		{"ERNIE 4.5", "ernie-4.5-8k", 4096, 8192},
		{"ERNIE Speed", "ernie-speed-128k", 4096, 128000},
	},
	"tencent": {
		{"Hunyuan 2.0 Thinking", "hunyuan-2.0-thinking", 16384, 32768},
		{"Hunyuan T1", "hunyuan-t1", 16384, 131072},
		{"Hunyuan Turbos", "hunyuan-turbos", 8192, 256000},
		{"Hunyuan 2.0 Instruct", "hunyuan-2.0-instruct", 8192, 256000},
	},
	"tencent_coding": {
		{"Auto (tc-code-latest)", "tc-code-latest", 196608, 32768},
		{"Kimi-K2.5", "kimi-k2.5", 262144, 32768},
		{"GLM-5", "glm-5", 202752, 16384},
		{"MiniMax-M2.5", "minimax-m2.5", 200704, 16384},
		{"Hunyuan 2.0 Thinking", "hunyuan-2.0-thinking", 131072, 16384},
		{"Hunyuan T1", "hunyuan-t1", 131072, 16384},
		{"Hunyuan Turbos", "hunyuan-turbos", 131072, 16384},
		{"Hunyuan 2.0 Instruct", "hunyuan-2.0-instruct", 131072, 16384},
	},
	"aliyun_coding": {
		{"Qwen Max", "qwen-max", 8192, 32768},
		{"Qwen Plus", "qwen-plus", 8192, 131072},
		{"Qwen Coder Plus", "qwen-coder-plus", 8192, 131072},
		{"Qwen2.5 Coder 32B", "qwen2.5-coder-32b-instruct", 8192, 131072},
	},
	"baidu_coding": {
		{"ERNIE 4.5 Turbo", "ernie-4.5-turbo-128k", 4096, 128000},
		{"ERNIE X1 Turbo", "ernie-x1-turbo-32k", 4096, 32000},
	},
	"bytedance": {
		{"Doubao 2.5 Pro 32k", "doubao-pro-32k-250115", 4096, 32000},
		{"Doubao 2.5 Pro 128k", "doubao-pro-128k-250115", 4096, 128000},
		{"Doubao Lite 128k", "doubao-lite-128k-240828", 4096, 128000},
	},
	"01ai": {
		{"Yi Large", "yi-large", 16384, 32000},
		{"Yi Medium", "yi-medium", 16384, 16000},
	},
	"mistral": {
		{"Mistral Large", "mistral-large-latest", 8192, 128000},
		{"Mistral Small", "mistral-small-latest", 8192, 128000},
	},
	"groq": {
		{"Llama 3.3 70B", "llama-3.3-70b-versatile", 8192, 128000},
		{"Mixtral 8x7B", "mixtral-8x7b-32768", 8192, 32768},
	},
	"cohere": {
		{"Command R+", "command-r-plus", 4096, 128000},
		{"Command R", "command-r", 4096, 128000},
	},
}

func init() {
	service.RegisterPlan(New())
}

func New() *sPlan {
	return &sPlan{}
}

type sPlan struct{}

// Create 创建AI套餐表，返回自动初始化的模型数量
func (s *sPlan) Create(ctx context.Context, in *model.PlanCreateInput) (int, error) {
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
	if err != nil {
		return 0, err
	}
	// 查询供应商 provider_type
	providerType, ptErr := g.DB().Ctx(ctx).Model("ai_provider").
		Where("id", in.ProviderID).Where("deleted_at", nil).
		Value("provider_type")
	if ptErr != nil {
		g.Log().Warningf(ctx, "initModelsForPlan: 查询供应商类型失败: %v", ptErr)
		return 0, nil
	}
	insertedModels, initErr := s.initModelsForPlan(ctx, id, snowflake.JsonInt64(in.ProviderID), providerType.String())
	if initErr != nil {
		g.Log().Warningf(ctx, "initModelsForPlan: 自动初始化模型失败: %v", initErr)
	}
	s.updateRolePresets(ctx, insertedModels)
	return len(insertedModels), nil
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
	if err != nil {
		return err
	}
	// 查询供应商 provider_type，补全缺失的预设模型
	providerType, ptErr := g.DB().Ctx(ctx).Model("ai_provider").
		Where("id", in.ProviderID).Where("deleted_at", nil).
		Value("provider_type")
	if ptErr != nil {
		g.Log().Warningf(ctx, "Update initModelsForPlan: 查询供应商类型失败: %v", ptErr)
		return nil
	}
	insertedModels, initErr := s.initModelsForPlan(ctx, in.ID, snowflake.JsonInt64(in.ProviderID), providerType.String())
	if initErr != nil {
		g.Log().Warningf(ctx, "Update initModelsForPlan: 自动初始化模型失败: %v", initErr)
	}
	s.updateRolePresets(ctx, insertedModels)
	return nil
}

// initModelsForPlan 根据 provider_type 为套餐批量初始化预设模型（跳过已存在的），返回本次新插入的模型列表
func (s *sPlan) initModelsForPlan(ctx context.Context, planID snowflake.JsonInt64, providerID snowflake.JsonInt64, providerType string) ([]insertedModel, error) {
	presets, ok := providerModelPresets[providerType]
	if !ok || len(presets) == 0 {
		return nil, nil
	}
	// 查询该 plan_id 下已有的 model_code
	existRows, err := dao.AiModel.Ctx(ctx).
		Where(dao.AiModel.Columns().PlanId, planID).
		Where(dao.AiModel.Columns().DeletedAt, nil).
		Fields(dao.AiModel.Columns().ModelCode).
		All()
	if err != nil {
		return nil, fmt.Errorf("查询已有模型失败: %w", err)
	}
	existSet := make(map[string]struct{}, len(existRows))
	for _, row := range existRows {
		existSet[row[dao.AiModel.Columns().ModelCode].String()] = struct{}{}
	}
	// 批量插入尚未存在的预设模型
	now := gtime.Now()
	createdBy := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)
	var result []insertedModel
	for i, preset := range presets {
		if _, exists := existSet[preset.ModelCode]; exists {
			continue
		}
		mid := snowflake.Generate()
		_, insertErr := dao.AiModel.Ctx(ctx).Data(g.Map{
			dao.AiModel.Columns().Id:            mid,
			dao.AiModel.Columns().PlanId:         planID,
			dao.AiModel.Columns().ProviderId:     providerID,
			dao.AiModel.Columns().Name:           preset.Name,
			dao.AiModel.Columns().ModelCode:      preset.ModelCode,
			dao.AiModel.Columns().Capability:     "chat",
			dao.AiModel.Columns().MaxTokens:      preset.MaxTokens,
			dao.AiModel.Columns().ContextWindow:  preset.ContextWindow,
			dao.AiModel.Columns().SupportsStream: 1,
			dao.AiModel.Columns().Status:         1,
			dao.AiModel.Columns().Sort:           i + 1,
			dao.AiModel.Columns().CreatedBy:      createdBy,
			dao.AiModel.Columns().DeptId:         deptID,
			dao.AiModel.Columns().CreatedAt:      now,
			dao.AiModel.Columns().UpdatedAt:      now,
		}).Insert()
		if insertErr != nil {
			g.Log().Warningf(ctx, "initModelsForPlan: 插入模型 %s 失败: %v", preset.ModelCode, insertErr)
			continue
		}
		result = append(result, insertedModel{
			ID:        mid,
			Name:      preset.Name,
			ModelCode: preset.ModelCode,
			Sort:      i + 1,
		})
	}
	return result, nil
}

// updateRolePresets 根据新插入的模型，按角色等级自动更新 mvp_role_preset.model_id（只更新 model_id=0 的未配置预设）
func (s *sPlan) updateRolePresets(ctx context.Context, models []insertedModel) {
	if len(models) == 0 {
		return
	}
	// 按 sort 已经有序（insert 时按 i+1 赋值）
	// max级 → models[0], pro级 → models[1]或[0], lite级 → models[2]或[1]或[0]
	getModel := func(idx int) int64 {
		if idx < len(models) {
			return int64(models[idx].ID)
		}
		return int64(models[len(models)-1].ID)
	}
	maxModelID := getModel(0)
	proModelID := getModel(1)
	liteModelID := getModel(2)

	_, err := g.DB().Ctx(ctx).Model("mvp_role_preset").
		Where("(model_id = 0 OR model_id IS NULL)").
		Where("role_level", "max").
		Where("deleted_at IS NULL").
		Data(g.Map{"model_id": maxModelID, "updated_at": gtime.Now()}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "updateRolePresets max 失败: %v", err)
	}
	_, err = g.DB().Ctx(ctx).Model("mvp_role_preset").
		Where("(model_id = 0 OR model_id IS NULL)").
		Where("role_level", "pro").
		Where("deleted_at IS NULL").
		Data(g.Map{"model_id": proModelID, "updated_at": gtime.Now()}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "updateRolePresets pro 失败: %v", err)
	}
	_, err = g.DB().Ctx(ctx).Model("mvp_role_preset").
		Where("(model_id = 0 OR model_id IS NULL)").
		Where("role_level", "lite").
		Where("deleted_at IS NULL").
		Data(g.Map{"model_id": liteModelID, "updated_at": gtime.Now()}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "updateRolePresets lite 失败: %v", err)
	}
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

