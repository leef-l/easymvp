// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpProjectRole is the golang structure for table mvp_project_role.
type MvpProjectRole struct {
	Id              uint64      `orm:"id"               description:"雪花ID"`                                                    // 雪花ID
	ProjectId       uint64      `orm:"project_id"       description:"项目ID"`                                                    // 项目ID
	ProjectCategory string      `orm:"project_category" description:"项目分类"`                                                    // 项目分类
	RoleType        string      `orm:"role_type"        description:"角色类型：architect/implementer/auditor/coordinator"`          // 角色类型：architect/implementer/auditor/coordinator
	RoleLevel       string      `orm:"role_level"       description:"角色等级：lite/pro/max"`                                       // 角色等级：lite/pro/max
	ModelId         uint64      `orm:"model_id"         description:"AI模型ID"`                                                  // AI模型ID
	SystemPrompt    string      `orm:"system_prompt"    description:"系统提示词（角色设定）"`                                             // 系统提示词（角色设定）
	ExecutionMode   string      `orm:"execution_mode"   description:"执行方式: chat=对话模式, aider=Aider代码编辑, openhands=OpenHands沙箱"` // 执行方式: chat=对话模式, aider=Aider代码编辑, openhands=OpenHands沙箱
	Status          int         `orm:"status"           description:"状态:0=禁用,1=启用"`                                            // 状态:0=禁用,1=启用
	CreatedBy       uint64      `orm:"created_by"       description:"创建人ID"`                                                   // 创建人ID
	DeptId          uint64      `orm:"dept_id"          description:"所属部门ID"`                                                  // 所属部门ID
	CreatedAt       *gtime.Time `orm:"created_at"       description:"创建时间"`                                                    // 创建时间
	UpdatedAt       *gtime.Time `orm:"updated_at"       description:"更新时间"`                                                    // 更新时间
	DeletedAt       *gtime.Time `orm:"deleted_at"       description:"软删除时间"`                                                   // 软删除时间
}
