// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpRolePreset is the golang structure of table mvp_role_preset for DAO operations like Where/Data.
type MvpRolePreset struct {
	g.Meta       `orm:"table:mvp_role_preset, do:true"`
	Id              any         // 雪花ID
	ProjectCategory any         // 项目分类
	RoleType        any         // 角色类型：architect/implementer/auditor/coordinator
	RoleLevel    any         // 角色等级：lite/pro/max
	ModelId      any         // AI模型ID
	SystemPrompt  any         // 默认系统提示词（角色设定）
	ExecutionMode any         // 执行方式: chat/aider/openhands
	Status        any         // 状态:0=禁用,1=启用
	Sort         any         // 排序
	CreatedBy    any         // 创建人ID
	DeptId       any         // 所属部门ID
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 软删除时间
}
