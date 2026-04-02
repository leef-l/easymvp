// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskLog is the golang structure of table mvp_task_log for DAO operations like Where/Data.
type MvpTaskLog struct {
	g.Meta     `orm:"table:mvp_task_log, do:true"`
	Id         any         // 自增ID
	TaskId     any         // 任务ID
	Action     any         // 动作：started/completed/failed/bug_found/reassigned
	FromStatus any         // 原状态
	ToStatus   any         // 新状态
	Message    any         // 日志内容
	Operator   any         // 操作者：user/architect/coordinator/system
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
	DeletedAt  *gtime.Time // 软删除时间
}
