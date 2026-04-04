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
	g.Meta              `orm:"table:mvp_project, do:true"`
	Id                  any         // 雪花ID
	Name                any         // 项目名称
	ProjectCategory     any         // 项目分类
	Description         any         // 项目简介
	Status              any         // 项目状态: designing/reviewing/running/paused/completed
	PauseReason         any         // 暂停原因
	GlobalContext       any         // 项目全局上下文（架构师需求分析+方案设计的压缩摘要）
	ArchitectModelId    any         // 架构师使用的AI模型ID
	WorkDir             any         // 项目代码工作目录（Aider执行路径）
	ActiveBatchNo       any         // 当前活跃批次号（调度器持久化，0=无活跃批次）
	EngineVersion       any         // 执行引擎版本: legacy/workflow_v2
	ActiveWorkflowRunId any         // 当前活跃工作流运行ID
	CreatedBy           any         // 创建人ID
	DeptId              any         // 所属部门ID
	CreatedAt           *gtime.Time // 创建时间
	UpdatedAt           *gtime.Time // 更新时间
	DeletedAt           *gtime.Time // 软删除时间
}
