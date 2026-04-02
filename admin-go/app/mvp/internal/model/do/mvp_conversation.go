// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpConversation is the golang structure of table mvp_conversation for DAO operations like Where/Data.
type MvpConversation struct {
	g.Meta    `orm:"table:mvp_conversation, do:true"`
	Id        any         // 雪花ID
	ProjectId any         // 项目ID
	TaskId    any         // 关联任务ID，NULL=项目级对话
	Title     any         // 对话标题
	RoleType  any         // 对话角色类型
	Status    any         // 状态：active/archived
	CreatedBy any         // 创建人ID
	DeptId    any         // 所属部门ID
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 软删除时间
}
