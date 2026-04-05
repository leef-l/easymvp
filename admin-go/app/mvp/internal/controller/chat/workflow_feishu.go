package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/collab/adapter"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/utility/snowflake"
)

const feishuCallbackPath = "/api/mvp/collab/feishu/callback"

// FeishuConfig 查询飞书协作配置。
func (c *cWorkflow) FeishuConfig(ctx context.Context, req *v1.WorkflowFeishuConfigReq) (res *v1.WorkflowFeishuConfigRes, err error) {
	return &v1.WorkflowFeishuConfigRes{
		Config: v1.FeishuConfigDTO{
			Enabled:              engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0),
			AppID:                engine.GetConfigString(ctx, "workflow.collab.feishu_app_id", "workflow.collab.feishuAppId", ""),
			AppSecret:            engine.GetConfigString(ctx, "workflow.collab.feishu_app_secret", "workflow.collab.feishuAppSecret", ""),
			VerificationToken:    engine.GetConfigString(ctx, "workflow.collab.feishu_verification_token", "workflow.collab.feishuVerificationToken", ""),
			EncryptKey:           engine.GetConfigString(ctx, "workflow.collab.feishu_encrypt_key", "workflow.collab.feishuEncryptKey", ""),
			DefaultNotifyUserIDs: engine.GetConfigString(ctx, "workflow.collab.feishu_default_notify_user_ids", "workflow.collab.feishuDefaultNotifyUserIds", ""),
			CallbackPath:         feishuCallbackPath,
		},
	}, nil
}

// SaveFeishuConfig 保存飞书协作配置。
func (c *cWorkflow) SaveFeishuConfig(ctx context.Context, req *v1.WorkflowSaveFeishuConfigReq) (res *v1.WorkflowSaveFeishuConfigRes, err error) {
	configs := []struct {
		key         string
		value       string
		configType  string
		description string
	}{
		{"workflow.collab.feishu_enabled", fmt.Sprintf("%d", boolToInt(req.Enabled == 1)), "int", "飞书通知总开关(0关/1开)"},
		{"workflow.collab.feishu_app_id", strings.TrimSpace(req.AppID), "string", "飞书应用 App ID"},
		{"workflow.collab.feishu_app_secret", strings.TrimSpace(req.AppSecret), "string", "飞书应用 App Secret"},
		{"workflow.collab.feishu_verification_token", strings.TrimSpace(req.VerificationToken), "string", "飞书 Verification Token"},
		{"workflow.collab.feishu_encrypt_key", strings.TrimSpace(req.EncryptKey), "string", "飞书事件回调加密 Key(签名验证)"},
		{"workflow.collab.feishu_default_notify_user_ids", strings.TrimSpace(req.DefaultNotifyUserIDs), "string", "降级通知的系统用户ID列表(逗号分隔)"},
	}
	for _, item := range configs {
		if err := saveMvpConfig(ctx, item.key, item.value, item.configType, "collab", item.description); err != nil {
			return nil, err
		}
	}
	return &v1.WorkflowSaveFeishuConfigRes{}, nil
}

// FeishuBindings 查询飞书绑定列表。
func (c *cWorkflow) FeishuBindings(ctx context.Context, req *v1.WorkflowFeishuBindingsReq) (res *v1.WorkflowFeishuBindingsRes, err error) {
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	bindings, err := bindingRepo.List(ctx, "feishu")
	if err != nil {
		return nil, err
	}
	items := make([]v1.FeishuBindingDTO, 0, len(bindings))
	for _, item := range bindings {
		items = append(items, mapToFeishuBindingDTO(item))
	}
	return &v1.WorkflowFeishuBindingsRes{Bindings: items}, nil
}

// BindFeishuUser 绑定飞书用户。
func (c *cWorkflow) BindFeishuUser(ctx context.Context, req *v1.WorkflowBindFeishuUserReq) (res *v1.WorkflowBindFeishuUserRes, err error) {
	currentUserID := middleware.GetUserID(ctx)
	if currentUserID != 1 && int64(req.UserID) != currentUserID {
		return nil, fmt.Errorf("普通用户只能绑定自己的飞书账号")
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	id, err := bindingRepo.Rebind(ctx, g.Map{
		"user_id":          int64(req.UserID),
		"platform":         "feishu",
		"platform_user_id": strings.TrimSpace(req.PlatformUserID),
		"platform_name":    strings.TrimSpace(req.PlatformName),
		"created_by":       currentUserID,
		"dept_id":          middleware.GetDeptID(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowBindFeishuUserRes{ID: snowflake.JsonInt64(id)}, nil
}

// UnbindFeishuUser 解绑飞书用户。
func (c *cWorkflow) UnbindFeishuUser(ctx context.Context, req *v1.WorkflowUnbindFeishuUserReq) (res *v1.WorkflowUnbindFeishuUserRes, err error) {
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByIDScoped(ctx, int64(req.BindingID))
	if err != nil || binding == nil {
		return nil, fmt.Errorf("绑定记录不存在或无权操作")
	}
	if err := bindingRepo.UnbindByID(ctx, int64(req.BindingID)); err != nil {
		return nil, err
	}
	return &v1.WorkflowUnbindFeishuUserRes{}, nil
}

// TestFeishuMessage 发送飞书测试消息。
func (c *cWorkflow) TestFeishuMessage(ctx context.Context, req *v1.WorkflowTestFeishuMessageReq) (res *v1.WorkflowTestFeishuMessageRes, err error) {
	if engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0) != 1 {
		return nil, fmt.Errorf("飞书通知总开关未开启，请先保存并启用飞书配置")
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByIDScoped(ctx, int64(req.BindingID))
	if err != nil || binding == nil {
		return nil, fmt.Errorf("绑定记录不存在或无权操作")
	}

	text := strings.TrimSpace(req.Content)
	if text == "" {
		text = "EasyMVP 飞书联通测试成功。后续审批卡片和阶段报告会通过当前绑定发送。"
	}
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, mapString(binding, "platform_user_id"), text); err != nil {
		return nil, err
	}
	return &v1.WorkflowTestFeishuMessageRes{}, nil
}

func saveMvpConfig(ctx context.Context, key, value, configType, category, description string) error {
	now := gtime.Now()
	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	count, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		_, err = g.DB().Model("mvp_config").Ctx(ctx).
			Where("config_key", key).
			WhereNull("deleted_at").
			Update(g.Map{
				"config_value": value,
				"config_type":  configType,
				"category":     category,
				"description":  description,
				"updated_at":   now,
			})
		return err
	}
	_, err = g.DB().Model("mvp_config").Ctx(ctx).Insert(g.Map{
		"id":           snowflake.Generate(),
		"config_key":   key,
		"config_value": value,
		"config_type":  configType,
		"category":     category,
		"description":  description,
		"created_by":   userID,
		"dept_id":      deptID,
		"created_at":   now,
		"updated_at":   now,
	})
	return err
}

func mapToFeishuBindingDTO(m g.Map) v1.FeishuBindingDTO {
	return v1.FeishuBindingDTO{
		ID:             mapJsonInt64(m, "id"),
		UserID:         mapJsonInt64(m, "user_id"),
		Platform:       mapString(m, "platform"),
		PlatformUserID: mapString(m, "platform_user_id"),
		PlatformName:   mapString(m, "platform_name"),
		CreatedBy:      mapJsonInt64(m, "created_by"),
		DeptID:         mapJsonInt64(m, "dept_id"),
		CreatedAt:      mapGTime(m, "created_at"),
		UpdatedAt:      mapGTime(m, "updated_at"),
	}
}

func boolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}
