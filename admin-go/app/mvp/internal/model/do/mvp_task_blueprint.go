// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskBlueprint is the golang structure of table mvp_task_blueprint for DAO operations like Where/Data.
type MvpTaskBlueprint struct {
	g.Meta                `orm:"table:mvp_task_blueprint, do:true"`
	Id                    any         // 雪花ID
	PlanVersionId         any         // 所属计划版本ID
	ParentBlueprintId     any         // 父蓝图ID(支持层级)
	Name                  any         // 任务名称
	Description           any         // 任务描述
	RoleType              any         // 角色类型: architect/implementer/auditor/coordinator
	RoleLevel             any         // 角色等级: lite/pro/max
	BatchNo               any         // 批次号
	Sort                  any         // 排序
	AffectedResources     any         // 影响资源列表(JSON)
	DependsOnBlueprintIds any         // 依赖蓝图ID列表(JSON)
	BlueprintStatus       any         // 蓝图状态: draft/confirmed/superseded
	CreatedAt             *gtime.Time // 创建时间
	UpdatedAt             *gtime.Time // 更新时间
	DeletedAt             *gtime.Time // 软删除时间
	CreatedBy             any         // 创建人ID
	DeptId                any         // 部门ID
}
