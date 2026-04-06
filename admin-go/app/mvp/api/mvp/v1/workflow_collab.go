package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// ==================== 飞书协作管理 API ====================

// WorkflowFeishuConfigReq 查询飞书配置。
type WorkflowFeishuConfigReq struct {
	g.Meta `path:"/workflow/feishu-config" method:"get" tags:"飞书协作" summary:"查询飞书协作配置"`
}

// FeishuConfigDTO 飞书配置 DTO。
type FeishuConfigDTO struct {
	Enabled              int    `json:"enabled"`
	AppID                string `json:"appId"`
	AppSecret            string `json:"appSecret"`
	VerificationToken    string `json:"verificationToken"`
	EncryptKey           string `json:"encryptKey"`
	DefaultNotifyUserIDs string `json:"defaultNotifyUserIds"`
	ConnectionMode       string `json:"connectionMode"` // webhook | websocket
	CallbackPath         string `json:"callbackPath"`
	WSRunning            bool   `json:"wsRunning"` // WebSocket 长连接是否在线
}

// WorkflowFeishuConfigRes 查询飞书配置响应。
type WorkflowFeishuConfigRes struct {
	g.Meta `mime:"application/json"`
	Config FeishuConfigDTO `json:"config"`
}

// WorkflowSaveFeishuConfigReq 保存飞书配置。
type WorkflowSaveFeishuConfigReq struct {
	g.Meta               `path:"/workflow/save-feishu-config" method:"post" tags:"飞书协作" summary:"保存飞书协作配置"`
	Enabled              int    `json:"enabled" dc:"0=关闭 1=开启"`
	AppID                string `json:"appId" dc:"飞书 App ID"`
	AppSecret            string `json:"appSecret" dc:"飞书 App Secret"`
	VerificationToken    string `json:"verificationToken" dc:"飞书 Verification Token"`
	EncryptKey           string `json:"encryptKey" dc:"飞书 Encrypt Key"`
	DefaultNotifyUserIDs string `json:"defaultNotifyUserIds" dc:"默认通知系统用户ID列表(逗号分隔)"`
	ConnectionMode       string `json:"connectionMode" dc:"连接模式：webhook(默认)|websocket"`
}

// WorkflowSaveFeishuConfigRes 保存飞书配置响应。
type WorkflowSaveFeishuConfigRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowFeishuBindingsReq 查询飞书绑定列表。
type WorkflowFeishuBindingsReq struct {
	g.Meta `path:"/workflow/feishu-bindings" method:"get" tags:"飞书协作" summary:"查询飞书用户绑定列表"`
}

// FeishuBindingDTO 飞书绑定 DTO。
type FeishuBindingDTO struct {
	ID             snowflake.JsonInt64 `json:"id"`
	UserID         snowflake.JsonInt64 `json:"userId"`
	Platform       string              `json:"platform"`
	PlatformUserID string              `json:"platformUserId"`
	PlatformName   string              `json:"platformName,omitempty"`
	CreatedBy      snowflake.JsonInt64 `json:"createdBy"`
	DeptID         snowflake.JsonInt64 `json:"deptId"`
	CreatedAt      *gtime.Time         `json:"createdAt,omitempty"`
	UpdatedAt      *gtime.Time         `json:"updatedAt,omitempty"`
}

// WorkflowFeishuBindingsRes 查询飞书绑定列表响应。
type WorkflowFeishuBindingsRes struct {
	g.Meta   `mime:"application/json"`
	Bindings []FeishuBindingDTO `json:"bindings"`
}

// WorkflowBindFeishuUserReq 绑定飞书用户。
type WorkflowBindFeishuUserReq struct {
	g.Meta         `path:"/workflow/bind-feishu-user" method:"post" tags:"飞书协作" summary:"绑定飞书用户"`
	UserID         snowflake.JsonInt64 `json:"userId" v:"required" dc:"系统用户ID"`
	PlatformUserID string              `json:"platformUserId" v:"required" dc:"飞书 open_id"`
	PlatformName   string              `json:"platformName" dc:"飞书显示名"`
}

// WorkflowBindFeishuUserRes 绑定飞书用户响应。
type WorkflowBindFeishuUserRes struct {
	g.Meta `mime:"application/json"`
	ID     snowflake.JsonInt64 `json:"id"`
}

// WorkflowUnbindFeishuUserReq 解绑飞书用户。
type WorkflowUnbindFeishuUserReq struct {
	g.Meta    `path:"/workflow/unbind-feishu-user" method:"post" tags:"飞书协作" summary:"解绑飞书用户"`
	BindingID snowflake.JsonInt64 `json:"bindingId" v:"required" dc:"绑定记录ID"`
}

// WorkflowUnbindFeishuUserRes 解绑飞书用户响应。
type WorkflowUnbindFeishuUserRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowTestFeishuMessageReq 发送飞书测试消息。
type WorkflowTestFeishuMessageReq struct {
	g.Meta    `path:"/workflow/test-feishu-message" method:"post" tags:"飞书协作" summary:"发送飞书测试消息"`
	BindingID snowflake.JsonInt64 `json:"bindingId" v:"required" dc:"绑定记录ID"`
	Content   string              `json:"content" dc:"测试消息内容"`
}

// WorkflowTestFeishuMessageRes 发送飞书测试消息响应。
type WorkflowTestFeishuMessageRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 飞书机器人菜单 API ====================

// BotMenuItem 菜单项（支持最多两级）。
type BotMenuItem struct {
	EventKey string        `json:"eventKey"` // 事件 key（点击后触发，用于识别指令）
	Name     string        `json:"name"`     // 中文显示名
	Children []BotMenuItem `json:"children,omitempty"`
}

// WorkflowGetBotMenuReq 查询当前飞书机器人菜单配置。
type WorkflowGetBotMenuReq struct {
	g.Meta `path:"/workflow/feishu-bot-menu" method:"get" tags:"飞书协作" summary:"查询飞书机器人菜单配置"`
}

// WorkflowGetBotMenuRes 查询飞书机器人菜单配置响应。
type WorkflowGetBotMenuRes struct {
	g.Meta       `mime:"application/json"`
	MenuItems    []BotMenuItem `json:"menuItems"`    // 当前保存的自定义菜单（未设置则返回默认）
	IsDefault    bool          `json:"isDefault"`    // true = 当前使用的是默认菜单
	DefaultItems []BotMenuItem `json:"defaultItems"` // 默认菜单结构（供参考/恢复用）
}

// WorkflowSetBotMenuReq 设置飞书机器人菜单（可自定义，留空则使用默认）。
type WorkflowSetBotMenuReq struct {
	g.Meta    `path:"/workflow/feishu-set-bot-menu" method:"post" tags:"飞书协作" summary:"设置飞书机器人菜单"`
	MenuItems []BotMenuItem `json:"menuItems" dc:"自定义菜单项（不传或传空则使用默认菜单）"`
	UseDefault bool         `json:"useDefault" dc:"true=恢复默认菜单并推送到飞书"`
}

// WorkflowSetBotMenuRes 设置飞书机器人菜单响应。
type WorkflowSetBotMenuRes struct {
	g.Meta  `mime:"application/json"`
	Message string `json:"message"`
}
