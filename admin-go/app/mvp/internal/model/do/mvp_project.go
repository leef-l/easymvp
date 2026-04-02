// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpProject is the golang structure of table mvp_project for DAO operations like Where/Data.
type MvpProject struct {
	g.Meta           `orm:"table:mvp_project, do:true"`
	Id               any         // 雪花ID
	Name             any         // 项目名称
	Description      any         // 项目简介
	Status           any         // 状态：designing/confirmed/running/paused/completed
	PauseReason      any         // 暂停原因
	GlobalContext    any         // 项目全局上下文（架构师需求分析+方案设计的压缩摘要）
	ArchitectModelId any         // 架构师使用的AI模型ID
	CreatedBy        any         // 创建人ID
	DeptId           any         // 所属部门ID
	CreatedAt        *gtime.Time // 创建时间
	UpdatedAt        *gtime.Time // 更新时间
	DeletedAt        *gtime.Time // 软删除时间
}
