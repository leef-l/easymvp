// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpReviewIssue is the golang structure for table mvp_review_issue.
type MvpReviewIssue struct {
	Id            uint64      `orm:"id"              description:"雪花ID"`                      // 雪花ID
	WorkflowRunId uint64      `orm:"workflow_run_id" description:"所属工作流运行ID"`                 // 所属工作流运行ID
	StageRunId    uint64      `orm:"stage_run_id"    description:"所属阶段运行ID"`                  // 所属阶段运行ID
	PlanVersionId uint64      `orm:"plan_version_id" description:"所属计划版本ID"`                  // 所属计划版本ID
	BlueprintId   uint64      `orm:"blueprint_id"    description:"关联蓝图ID"`                    // 关联蓝图ID
	Severity      string      `orm:"severity"        description:"严重级别: error/warning/info"`  // 严重级别: error/warning/info
	IssueCode     string      `orm:"issue_code"      description:"问题代码"`                      // 问题代码
	IssueType     string      `orm:"issue_type"      description:"问题类型"`                      // 问题类型
	SourceRole    string      `orm:"source_role"     description:"发现角色"`                      // 发现角色
	TaskName      string      `orm:"task_name"       description:"关联任务名"`                     // 关联任务名
	Message       string      `orm:"message"         description:"问题描述"`                      // 问题描述
	Suggestion    string      `orm:"suggestion"      description:"修复建议"`                      // 修复建议
	Status        string      `orm:"status"          description:"状态: open/resolved/ignored"` // 状态: open/resolved/ignored
	ResolvedAt    *gtime.Time `orm:"resolved_at"     description:"解决时间"`                      // 解决时间
	CreatedAt     *gtime.Time `orm:"created_at"      description:"创建时间"`                      // 创建时间
	UpdatedAt     *gtime.Time `orm:"updated_at"      description:"更新时间"`                      // 更新时间
	DeletedAt     *gtime.Time `orm:"deleted_at"      description:"软删除时间"`                     // 软删除时间
	CreatedBy     int64       `orm:"created_by"      description:"创建人ID"`                     // 创建人ID
	DeptId        int64       `orm:"dept_id"         description:"部门ID"`                      // 部门ID
}
