// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskBlueprint is the golang structure for table mvp_task_blueprint.
type MvpTaskBlueprint struct {
	Id                    uint64      `orm:"id"                       description:"雪花ID"`                                            // 雪花ID
	PlanVersionId         uint64      `orm:"plan_version_id"          description:"所属计划版本ID"`                                        // 所属计划版本ID
	ParentBlueprintId     uint64      `orm:"parent_blueprint_id"      description:"父蓝图ID(支持层级)"`                                     // 父蓝图ID(支持层级)
	Name                  string      `orm:"name"                     description:"任务名称"`                                            // 任务名称
	Description           string      `orm:"description"              description:"任务描述"`                                            // 任务描述
	RoleType              string      `orm:"role_type"                description:"角色类型: architect/implementer/auditor/coordinator"` // 角色类型: architect/implementer/auditor/coordinator
	RoleLevel             string      `orm:"role_level"               description:"角色等级: lite/pro/max"`                              // 角色等级: lite/pro/max
	BatchNo               int         `orm:"batch_no"                 description:"批次号"`                                             // 批次号
	Sort                  int         `orm:"sort"                     description:"排序"`                                              // 排序
	AffectedResources     string      `orm:"affected_resources"       description:"影响资源列表(JSON)"`                                    // 影响资源列表(JSON)
	DependsOnBlueprintIds string      `orm:"depends_on_blueprint_ids" description:"依赖蓝图ID列表(JSON)"`                                  // 依赖蓝图ID列表(JSON)
	BlueprintStatus       string      `orm:"blueprint_status"         description:"蓝图状态: draft/confirmed/superseded"`                // 蓝图状态: draft/confirmed/superseded
	CreatedAt             *gtime.Time `orm:"created_at"               description:"创建时间"`                                            // 创建时间
	UpdatedAt             *gtime.Time `orm:"updated_at"               description:"更新时间"`                                            // 更新时间
	DeletedAt             *gtime.Time `orm:"deleted_at"               description:"软删除时间"`                                           // 软删除时间
	CreatedBy             int64       `orm:"created_by"               description:"创建人ID"`                                           // 创建人ID
	DeptId                int64       `orm:"dept_id"                  description:"部门ID"`                                            // 部门ID
}
