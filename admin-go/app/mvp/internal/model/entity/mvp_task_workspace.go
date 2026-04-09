// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTaskWorkspace is the golang structure for table mvp_task_workspace.
type MvpTaskWorkspace struct {
	Id             uint64      `orm:"id"              description:"雪花ID"`                                                 // 雪花ID
	TaskId         uint64      `orm:"task_id"         description:"任务ID(domain_task或mvp_task)"`                           // 任务ID(domain_task或mvp_task)
	WorkflowRunId  uint64      `orm:"workflow_run_id" description:"所属工作流运行ID"`                                            // 所属工作流运行ID
	ProjectId      uint64      `orm:"project_id"      description:"项目ID"`                                                 // 项目ID
	WorkspaceType  string      `orm:"workspace_type"  description:"工作空间类型: git_worktree"`                                 // 工作空间类型: git_worktree
	WorkspacePath  string      `orm:"workspace_path"  description:"工作空间绝对路径"`                                             // 工作空间绝对路径
	BaseRef        string      `orm:"base_ref"        description:"基线引用(commit hash/branch)"`                             // 基线引用(commit hash/branch)
	Status         string      `orm:"status"          description:"状态: creating/ready/running/completed/failed/canceled"` // 状态: creating/ready/running/completed/failed/canceled
	CleanupStatus  string      `orm:"cleanup_status"  description:"清理状态: pending/done/retained/failed"`                   // 清理状态: pending/done/retained/failed
	DeliveryMode   string      `orm:"delivery_mode"   description:"交付结果形态: patch/pr/manual"`                              // 交付结果形态: patch/pr/manual
	DeliveryStatus string      `orm:"delivery_status" description:"交付状态: pending/ready/skipped/failed"`                   // 交付状态: pending/ready/skipped/failed
	SyncStrategy   string      `orm:"sync_strategy"   description:"回写策略: auto_apply/manual"`                              // 回写策略: auto_apply/manual
	SyncStatus     string      `orm:"sync_status"     description:"回写状态: pending/applied/skipped/failed"`                 // 回写状态: pending/applied/skipped/failed
	RiskLevel      string      `orm:"risk_level"      description:"风险等级: low/medium/high"`                                // 风险等级: low/medium/high
	PatchRef       string      `orm:"patch_ref"       description:"patch 产物路径"`                                           // patch 产物路径
	DeliveryRef    string      `orm:"delivery_ref"    description:"交付引用(PR草稿文件/外部链接)"`                                    // 交付引用(PR草稿文件/外部链接)
	DeliveryTitle  string      `orm:"delivery_title"  description:"交付标题"`                                                 // 交付标题
	DiffSummary    string      `orm:"diff_summary"    description:"变更摘要(diff统计)"`                                         // 变更摘要(diff统计)
	ErrorMessage   string      `orm:"error_message"   description:"错误信息"`                                                 // 错误信息
	CreatedAt      *gtime.Time `orm:"created_at"      description:"创建时间"`                                                 // 创建时间
	UpdatedAt      *gtime.Time `orm:"updated_at"      description:"更新时间"`                                                 // 更新时间
	DeletedAt      *gtime.Time `orm:"deleted_at"      description:"软删除时间"`                                                // 软删除时间
}
