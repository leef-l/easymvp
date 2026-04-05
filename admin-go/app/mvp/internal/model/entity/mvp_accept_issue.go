// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptIssue is the golang structure for table mvp_accept_issue.
type MvpAcceptIssue struct {
	Id              int64       `orm:"id"               description:"主键ID"`                          // 主键ID
	AcceptRunId     int64       `orm:"accept_run_id"    description:"验收运行ID"`                        // 验收运行ID
	WorkflowRunId   int64       `orm:"workflow_run_id"  description:"工作流运行ID"`                       // 工作流运行ID
	ProjectId       int64       `orm:"project_id"       description:"项目ID"`                          // 项目ID
	DomainTaskId    int64       `orm:"domain_task_id"   description:"主关联任务ID"`                       // 主关联任务ID
	IssueType       string      `orm:"issue_type"       description:"artifact/process/quality/risk"` // artifact/process/quality/risk
	RuleCode        string      `orm:"rule_code"        description:"规则编码"`                          // 规则编码
	Severity        string      `orm:"severity"         description:"info/warn/error/blocker"`       // info/warn/error/blocker
	Title           string      `orm:"title"            description:"问题标题"`                          // 问题标题
	Detail          string      `orm:"detail"           description:"问题详情"`                          // 问题详情
	ExpectedValue   string      `orm:"expected_value"   description:"预期值"`                           // 预期值
	ActualValue     string      `orm:"actual_value"     description:"实际值"`                           // 实际值
	SuggestedAction string      `orm:"suggested_action" description:"建议动作"`                          // 建议动作
	ResourceRef     string      `orm:"resource_ref"     description:"关联资源引用(JSON)"`                  // 关联资源引用(JSON)
	Status          string      `orm:"status"           description:"open/resolved/ignored"`         // open/resolved/ignored
	CreatedBy       int64       `orm:"created_by"       description:"创建人"`                           // 创建人
	DeptId          int64       `orm:"dept_id"          description:"部门ID"`                          // 部门ID
	CreatedAt       *gtime.Time `orm:"created_at"       description:"创建时间"`                          // 创建时间
	UpdatedAt       *gtime.Time `orm:"updated_at"       description:"更新时间"`                          // 更新时间
	DeletedAt       *gtime.Time `orm:"deleted_at"       description:"删除时间"`                          // 删除时间
}
