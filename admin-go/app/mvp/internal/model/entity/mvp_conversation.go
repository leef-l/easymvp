// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpConversation is the golang structure for table mvp_conversation.
type MvpConversation struct {
	Id        uint64      `orm:"id"         description:"雪花ID"`               // 雪花ID
	ProjectId uint64      `orm:"project_id" description:"项目ID"`               // 项目ID
	TaskId    uint64      `orm:"task_id"    description:"关联任务ID，NULL=项目级对话"`  // 关联任务ID，NULL=项目级对话
	Title     string      `orm:"title"      description:"对话标题"`               // 对话标题
	RoleType  string      `orm:"role_type"  description:"对话角色类型"`             // 对话角色类型
	Status    string      `orm:"status"     description:"状态：active/archived"` // 状态：active/archived
	CreatedBy uint64      `orm:"created_by" description:"创建人ID"`              // 创建人ID
	DeptId    uint64      `orm:"dept_id"    description:"所属部门ID"`             // 所属部门ID
	CreatedAt *gtime.Time `orm:"created_at" description:"创建时间"`               // 创建时间
	UpdatedAt *gtime.Time `orm:"updated_at" description:"更新时间"`               // 更新时间
	DeletedAt *gtime.Time `orm:"deleted_at" description:"软删除时间"`              // 软删除时间
}
