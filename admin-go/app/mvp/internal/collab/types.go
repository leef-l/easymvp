// Package collab 协作平台集成层，提供平台无关的消息推送抽象。
package collab

import "context"

// Platform 协作平台标识。
type Platform string

const (
	PlatformFeishu   Platform = "feishu"
	PlatformDingtalk Platform = "dingtalk"
	PlatformWecom    Platform = "wecom"
)

// MessageAdapter 消息适配器接口（平台无关）。
type MessageAdapter interface {
	// SendTextMessage 发送纯文本消息。
	SendTextMessage(ctx context.Context, platformUserID string, text string) error
	// SendCardMessage 发送交互卡片消息。
	SendCardMessage(ctx context.Context, platformUserID string, card *InteractiveCard) error
	// IsEnabled 当前平台是否启用。
	IsEnabled(ctx context.Context) bool
	// Platform 返回平台标识。
	GetPlatform() Platform
}

// InteractiveCard 交互卡片（平台无关的业务描述）。
type InteractiveCard struct {
	Title         string       // 卡片标题
	Level         string       // 决策等级 A/B/C
	ActionType    string       // 动作类型
	ActionID      int64        // 决策动作 ID
	TriggerSource string       // 触发源
	ProjectName   string       // 项目名称
	Details       string       // 详细描述
	Buttons       []CardButton // 交互按钮
	HeaderColor   string       // 头部颜色: blue/orange/red/green
}

// CardButton 卡片交互按钮。
type CardButton struct {
	Label  string // 按钮文字
	Action string // 动作标识: approve / reject
	Style  string // 样式: primary / danger
}
