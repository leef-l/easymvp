// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptIssue is the golang structure of table mvp_accept_issue for DAO operations like Where/Data.
type MvpAcceptIssue struct {
	g.Meta          `orm:"table:mvp_accept_issue, do:true"`
	Id              any         // 主键ID
	AcceptRunId     any         // 验收运行ID
	WorkflowRunId   any         // 工作流运行ID
	ProjectId       any         // 项目ID
	DomainTaskId    any         // 主关联任务ID
	IssueType       any         // artifact/process/quality/risk
	RuleCode        any         // 规则编码
	Severity        any         // info/warn/error/blocker
	Title           any         // 问题标题
	Detail          any         // 问题详情
	ExpectedValue   any         // 预期值
	ActualValue     any         // 实际值
	SuggestedAction any         // 建议动作
	ResourceRef     any         // 关联资源引用(JSON)
	Status          any         // open/resolved/ignored
	CreatedBy       any         // 创建人
	DeptId          any         // 部门ID
	CreatedAt       *gtime.Time // 创建时间
	UpdatedAt       *gtime.Time // 更新时间
	DeletedAt       *gtime.Time // 删除时间
}
