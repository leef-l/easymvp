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

// ==================== 飞书群菜单 API ====================

// ChatMenuItem 群菜单项（跳转链接类型）。
type ChatMenuItem struct {
	Name    string         `json:"name"`              // 菜单名称
	URL     string         `json:"url"`               // 跳转链接
	Children []ChatMenuItem `json:"children,omitempty"` // 二级菜单（有子菜单时 URL 忽略）
}

// WorkflowCreateChatMenuReq 创建群菜单请求。
type WorkflowCreateChatMenuReq struct {
	g.Meta   `path:"/workflow/feishu-create-chat-menu" method:"post" tags:"飞书协作" summary:"创建飞书群菜单"`
	ChatID   string         `json:"chatId" v:"required" dc:"飞书群 chat_id"`
	MenuItems []ChatMenuItem `json:"menuItems" dc:"菜单项列表（不传则使用默认 EasyMVP 快捷菜单）"`
}

// WorkflowCreateChatMenuRes 创建群菜单响应。
type WorkflowCreateChatMenuRes struct {
	g.Meta  `mime:"application/json"`
	Message string `json:"message"`
}

// WorkflowGetChatMenuReq 获取群菜单请求。
type WorkflowGetChatMenuReq struct {
	g.Meta `path:"/workflow/feishu-get-chat-menu" method:"get" tags:"飞书协作" summary:"获取飞书群菜单"`
	ChatID string `json:"chatId" v:"required" dc:"飞书群 chat_id"`
}

// WorkflowGetChatMenuRes 获取群菜单响应。
type WorkflowGetChatMenuRes struct {
	g.Meta    `mime:"application/json"`
	MenuItems []g.Map `json:"menuItems"`
}

// WorkflowDeleteChatMenuReq 删除群菜单请求。
type WorkflowDeleteChatMenuReq struct {
	g.Meta  `path:"/workflow/feishu-delete-chat-menu" method:"post" tags:"飞书协作" summary:"删除飞书群菜单"`
	ChatID  string   `json:"chatId" v:"required" dc:"飞书群 chat_id"`
	MenuIDs []string `json:"menuIds" v:"required" dc:"要删除的一级菜单ID列表"`
}

// WorkflowDeleteChatMenuRes 删除群菜单响应。
type WorkflowDeleteChatMenuRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== Telegram 协作管理 API ====================

// WorkflowTelegramConfigReq 查询 Telegram 配置。
type WorkflowTelegramConfigReq struct {
	g.Meta `path:"/workflow/telegram-config" method:"get" tags:"Telegram协作" summary:"查询 Telegram Bot 配置"`
}

// TelegramConfigDTO Telegram 配置 DTO。
type TelegramConfigDTO struct {
	Enabled    int    `json:"enabled"`
	BotToken   string `json:"botToken"`
	BotRunning bool   `json:"botRunning"` // Polling 是否在线
}

// WorkflowTelegramConfigRes 查询 Telegram 配置响应。
type WorkflowTelegramConfigRes struct {
	g.Meta `mime:"application/json"`
	Config TelegramConfigDTO `json:"config"`
}

// WorkflowSaveTelegramConfigReq 保存 Telegram 配置。
type WorkflowSaveTelegramConfigReq struct {
	g.Meta   `path:"/workflow/save-telegram-config" method:"post" tags:"Telegram协作" summary:"保存 Telegram Bot 配置"`
	Enabled  int    `json:"enabled" dc:"0=关闭 1=开启"`
	BotToken string `json:"botToken" dc:"Telegram Bot Token"`
}

// WorkflowSaveTelegramConfigRes 保存 Telegram 配置响应。
type WorkflowSaveTelegramConfigRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowTelegramBindingsReq 查询 Telegram 绑定列表。
type WorkflowTelegramBindingsReq struct {
	g.Meta `path:"/workflow/telegram-bindings" method:"get" tags:"Telegram协作" summary:"查询 Telegram 用户绑定列表"`
}

// WorkflowTelegramBindingsRes 查询 Telegram 绑定列表响应。
type WorkflowTelegramBindingsRes struct {
	g.Meta   `mime:"application/json"`
	Bindings []FeishuBindingDTO `json:"bindings"` // 复用同一 DTO
}

// WorkflowBindTelegramUserReq 绑定 Telegram 用户。
type WorkflowBindTelegramUserReq struct {
	g.Meta         `path:"/workflow/bind-telegram-user" method:"post" tags:"Telegram协作" summary:"绑定 Telegram 用户"`
	UserID         snowflake.JsonInt64 `json:"userId" v:"required" dc:"系统用户ID"`
	PlatformUserID string              `json:"platformUserId" v:"required" dc:"Telegram chat_id（数字字符串）"`
	PlatformName   string              `json:"platformName" dc:"Telegram 用户名（@username）"`
}

// WorkflowBindTelegramUserRes 绑定 Telegram 用户响应。
type WorkflowBindTelegramUserRes struct {
	g.Meta `mime:"application/json"`
	ID     snowflake.JsonInt64 `json:"id"`
}

// WorkflowUnbindTelegramUserReq 解绑 Telegram 用户。
type WorkflowUnbindTelegramUserReq struct {
	g.Meta    `path:"/workflow/unbind-telegram-user" method:"post" tags:"Telegram协作" summary:"解绑 Telegram 用户"`
	BindingID snowflake.JsonInt64 `json:"bindingId" v:"required" dc:"绑定记录ID"`
}

// WorkflowUnbindTelegramUserRes 解绑 Telegram 用户响应。
type WorkflowUnbindTelegramUserRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowTestTelegramMessageReq 发送 Telegram 测试消息。
type WorkflowTestTelegramMessageReq struct {
	g.Meta    `path:"/workflow/test-telegram-message" method:"post" tags:"Telegram协作" summary:"发送 Telegram 测试消息"`
	BindingID snowflake.JsonInt64 `json:"bindingId" v:"required" dc:"绑定记录ID"`
	Content   string              `json:"content" dc:"测试消息内容"`
}

// WorkflowTestTelegramMessageRes 发送 Telegram 测试消息响应。
type WorkflowTestTelegramMessageRes struct {
	g.Meta `mime:"application/json"`
}

// TelegramCommandItem Telegram Bot 命令菜单项。
type TelegramCommandItem struct {
	Command     string `json:"command"`     // 命令（不含/，如 help）
	Description string `json:"description"` // 描述（显示在命令菜单里）
}

// WorkflowSetTelegramCommandsReq 设置 Telegram Bot 命令菜单。
type WorkflowSetTelegramCommandsReq struct {
	g.Meta   `path:"/workflow/telegram-set-commands" method:"post" tags:"Telegram协作" summary:"设置 Telegram Bot 命令菜单"`
	Commands []TelegramCommandItem `json:"commands" dc:"命令列表（传空则恢复默认）"`
	UseDefault bool                `json:"useDefault" dc:"true=恢复默认命令菜单"`
}

// WorkflowSetTelegramCommandsRes 设置 Telegram Bot 命令菜单响应。
type WorkflowSetTelegramCommandsRes struct {
	g.Meta   `mime:"application/json"`
	Message  string               `json:"message"`
	Commands []TelegramCommandItem `json:"commands"`
}
