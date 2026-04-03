// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpMessage is the golang structure for table mvp_message.
type MvpMessage struct {
	Id             uint64      `orm:"id"              description:"雪花ID"`                                       // 雪花ID
	ConversationId uint64      `orm:"conversation_id" description:"对话ID"`                                       // 对话ID
	Role           string      `orm:"role"            description:"消息角色：user/assistant/system"`                 // 消息角色：user/assistant/system
	MessageType    string      `orm:"message_type"    description:"消息类型"`                                      // 消息类型
	Content        string      `orm:"content"         description:"消息内容"`                                       // 消息内容
	ModelId        uint64      `orm:"model_id"        description:"使用的AI模型ID"`                                  // 使用的AI模型ID
	TokenUsage     string      `orm:"token_usage"     description:"token消耗：{prompt_tokens, completion_tokens}"` // token消耗：{prompt_tokens, completion_tokens}
	Status         string      `orm:"status"          description:"状态：pending/streaming/completed/failed"`      // 状态：pending/streaming/completed/failed
	CreatedBy      uint64      `orm:"created_by"      description:"创建人ID"`                                      // 创建人ID
	DeptId         uint64      `orm:"dept_id"         description:"所属部门ID"`                                     // 所属部门ID
	CreatedAt      *gtime.Time `orm:"created_at"      description:"创建时间"`                                       // 创建时间
	UpdatedAt      *gtime.Time `orm:"updated_at"      description:"更新时间"`                                       // 更新时间
	DeletedAt      *gtime.Time `orm:"deleted_at"      description:"软删除时间"`                                      // 软删除时间
}
