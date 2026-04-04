// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpMessage is the golang structure of table mvp_message for DAO operations like Where/Data.
type MvpMessage struct {
	g.Meta         `orm:"table:mvp_message, do:true"`
	Id             any         // 雪花ID
	ConversationId any         // 对话ID
	Role           any         // 消息角色：user/assistant/system
	MessageType    any         // 消息类型：chat_user/chat_reply/task_prompt/task_reply/system_notice/poison/general
	Content        any         // 消息内容
	ModelId        any         // 使用的AI模型ID
	TokenUsage     any         // token消耗：{prompt_tokens, completion_tokens}
	Status         any         // 状态：pending/streaming/completed/failed
	CreatedBy      any         // 创建人ID
	DeptId         any         // 所属部门ID
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
	DeletedAt      *gtime.Time // 软删除时间
}
