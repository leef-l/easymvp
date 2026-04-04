// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskResourceLock is the golang structure of table mvp_task_resource_lock for DAO operations like Where/Data.
type MvpTaskResourceLock struct {
	g.Meta        `orm:"table:mvp_task_resource_lock, do:true"`
	Id            any         // 雪花ID
	WorkflowRunId any         // 所属工作流运行ID
	TaskId        any         // 持锁任务ID
	ResourcePath  any         // 资源路径
	LockStatus    any         // 锁状态: held/released/leaked
	LockedAt      *gtime.Time // 加锁时间
	ReleasedAt    *gtime.Time // 释放时间
}
