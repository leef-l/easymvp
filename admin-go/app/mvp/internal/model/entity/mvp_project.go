// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpProject is the golang structure for table mvp_project.
type MvpProject struct {
	Id                  uint64      `orm:"id"                     description:"雪花ID"`                                               // 雪花ID
	Name                string      `orm:"name"                   description:"项目名称"`                                               // 项目名称
	ProjectCategory     string      `orm:"project_category"       description:"项目分类"`                                               // 项目分类
	CategoryCode        string      `orm:"category_code"          description:"项目分类编码"`                                             // 项目分类编码
	Description         string      `orm:"description"            description:"项目简介"`                                               // 项目简介
	Status              string      `orm:"status"                 description:"项目状态: designing/reviewing/running/paused/completed"` // 项目状态: designing/reviewing/running/paused/completed
	PauseReason         string      `orm:"pause_reason"           description:"暂停原因"`                                               // 暂停原因
	GlobalContext       string      `orm:"global_context"         description:"项目全局上下文（架构师需求分析+方案设计的压缩摘要）"`                         // 项目全局上下文（架构师需求分析+方案设计的压缩摘要）
	ArchitectModelId    uint64      `orm:"architect_model_id"     description:"架构师使用的AI模型ID"`                                       // 架构师使用的AI模型ID
	WorkDir             string      `orm:"work_dir"               description:"项目代码工作目录（Aider执行路径）"`                                // 项目代码工作目录（Aider执行路径）
	ActiveBatchNo       int         `orm:"active_batch_no"        description:"当前活跃批次号（调度器持久化，0=无活跃批次）"`                            // 当前活跃批次号（调度器持久化，0=无活跃批次）
	EngineVersion       string      `orm:"engine_version"         description:"执行引擎版本: legacy/workflow_v2"`                         // 执行引擎版本: legacy/workflow_v2
	ActiveWorkflowRunId uint64      `orm:"active_workflow_run_id" description:"当前活跃工作流运行ID"`                                        // 当前活跃工作流运行ID
	CreatedBy           uint64      `orm:"created_by"             description:"创建人ID"`                                              // 创建人ID
	DeptId              uint64      `orm:"dept_id"                description:"所属部门ID"`                                             // 所属部门ID
	CreatedAt           *gtime.Time `orm:"created_at"             description:"创建时间"`                                               // 创建时间
	UpdatedAt           *gtime.Time `orm:"updated_at"             description:"更新时间"`                                               // 更新时间
	DeletedAt           *gtime.Time `orm:"deleted_at"             description:"软删除时间"`                                              // 软删除时间
}
