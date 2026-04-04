// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpReviewIssue is the golang structure of table mvp_review_issue for DAO operations like Where/Data.
type MvpReviewIssue struct {
	g.Meta        `orm:"table:mvp_review_issue, do:true"`
	Id            any         // 雪花ID
	WorkflowRunId any         // 所属工作流运行ID
	StageRunId    any         // 所属阶段运行ID
	PlanVersionId any         // 所属计划版本ID
	BlueprintId   any         // 关联蓝图ID
	Severity      any         // 严重级别: error/warning/info
	IssueCode     any         // 问题代码
	IssueType     any         // 问题类型
	SourceRole    any         // 发现角色
	TaskName      any         // 关联任务名
	Message       any         // 问题描述
	Suggestion    any         // 修复建议
	Status        any         // 状态: open/resolved/ignored
	ResolvedAt    *gtime.Time // 解决时间
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
	DeletedAt     *gtime.Time // 软删除时间
}
