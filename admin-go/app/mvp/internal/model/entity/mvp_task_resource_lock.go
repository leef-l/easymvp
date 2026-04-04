// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskResourceLock is the golang structure for table mvp_task_resource_lock.
type MvpTaskResourceLock struct {
	Id            uint64      `orm:"id"              description:"雪花ID"`                      // 雪花ID
	WorkflowRunId uint64      `orm:"workflow_run_id" description:"所属工作流运行ID"`                 // 所属工作流运行ID
	TaskId        uint64      `orm:"task_id"         description:"持锁任务ID"`                    // 持锁任务ID
	ResourcePath  string      `orm:"resource_path"   description:"资源路径"`                      // 资源路径
	LockStatus    string      `orm:"lock_status"     description:"锁状态: held/released/leaked"` // 锁状态: held/released/leaked
	LockedAt      *gtime.Time `orm:"locked_at"       description:"加锁时间"`                      // 加锁时间
	ReleasedAt    *gtime.Time `orm:"released_at"     description:"释放时间"`                      // 释放时间
}
