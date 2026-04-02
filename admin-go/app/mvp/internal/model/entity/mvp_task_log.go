// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskLog is the golang structure for table mvp_task_log.
type MvpTaskLog struct {
	Id         uint64      `orm:"id"          description:"自增ID"`                                             // 自增ID
	TaskId     uint64      `orm:"task_id"     description:"任务ID"`                                             // 任务ID
	Action     string      `orm:"action"      description:"动作：started/completed/failed/bug_found/reassigned"` // 动作：started/completed/failed/bug_found/reassigned
	FromStatus string      `orm:"from_status" description:"原状态"`                                              // 原状态
	ToStatus   string      `orm:"to_status"   description:"新状态"`                                              // 新状态
	Message    string      `orm:"message"     description:"日志内容"`                                             // 日志内容
	Operator   string      `orm:"operator"    description:"操作者：user/architect/coordinator/system"`            // 操作者：user/architect/coordinator/system
	CreatedAt  *gtime.Time `orm:"created_at"  description:"创建时间"`                                             // 创建时间
	UpdatedAt  *gtime.Time `orm:"updated_at"  description:"更新时间"`                                             // 更新时间
	DeletedAt  *gtime.Time `orm:"deleted_at"  description:"软删除时间"`                                            // 软删除时间
}
