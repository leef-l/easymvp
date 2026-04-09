// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskWorkspace is the golang structure of table mvp_task_workspace for DAO operations like Where/Data.
type MvpTaskWorkspace struct {
	g.Meta         `orm:"table:mvp_task_workspace, do:true"`
	Id             any         // 雪花ID
	TaskId         any         // 任务ID(domain_task或mvp_task)
	WorkflowRunId  any         // 所属工作流运行ID
	ProjectId      any         // 项目ID
	WorkspaceType  any         // 工作空间类型: git_worktree
	WorkspacePath  any         // 工作空间绝对路径
	BaseRef        any         // 基线引用(commit hash/branch)
	Status         any         // 状态: creating/ready/running/completed/failed/canceled
	CleanupStatus  any         // 清理状态: pending/done/retained/failed
	DeliveryMode   any         // 交付结果形态: patch/pr/manual
	DeliveryStatus any         // 交付状态: pending/ready/skipped/failed
	SyncStrategy   any         // 回写策略: auto_apply/manual
	SyncStatus     any         // 回写状态: pending/applied/skipped/failed
	RiskLevel      any         // 风险等级: low/medium/high
	PatchRef       any         // patch 产物路径
	DeliveryRef    any         // 交付引用(PR草稿文件/外部链接)
	DeliveryTitle  any         // 交付标题
	DiffSummary    any         // 变更摘要(diff统计)
	ErrorMessage   any         // 错误信息
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
	DeletedAt      *gtime.Time // 软删除时间
}
