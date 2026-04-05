// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptRun is the golang structure for table mvp_accept_run.
type MvpAcceptRun struct {
	Id               int64       `orm:"id"                 description:"主键ID"`                                      // 主键ID
	WorkflowRunId    int64       `orm:"workflow_run_id"    description:"工作流运行ID"`                                   // 工作流运行ID
	StageRunId       int64       `orm:"stage_run_id"       description:"accept阶段stage_run_id"`                      // accept阶段stage_run_id
	ProjectId        int64       `orm:"project_id"         description:"项目ID"`                                      // 项目ID
	PlanVersionId    int64       `orm:"plan_version_id"    description:"关联方案版本ID"`                                  // 关联方案版本ID
	AcceptRound      int         `orm:"accept_round"       description:"第几轮验收"`                                     // 第几轮验收
	Status           string      `orm:"status"             description:"pending/running/completed/failed/canceled"` // pending/running/completed/failed/canceled
	Decision         string      `orm:"decision"           description:"passed/failed/manual_review"`               // passed/failed/manual_review
	Score            float64     `orm:"score"              description:"验收评分"`                                      // 验收评分
	Summary          string      `orm:"summary"            description:"验收摘要"`                                      // 验收摘要
	RulesVersion     string      `orm:"rules_version"      description:"规则版本号"`                                     // 规则版本号
	RulesSnapshotRef string      `orm:"rules_snapshot_ref" description:"规则快照引用或JSON"`                               // 规则快照引用或JSON
	CreatedBy        int64       `orm:"created_by"         description:"创建人"`                                       // 创建人
	DeptId           int64       `orm:"dept_id"            description:"部门ID"`                                      // 部门ID
	StartedAt        *gtime.Time `orm:"started_at"         description:"开始时间"`                                      // 开始时间
	FinishedAt       *gtime.Time `orm:"finished_at"        description:"结束时间"`                                      // 结束时间
	CreatedAt        *gtime.Time `orm:"created_at"         description:"创建时间"`                                      // 创建时间
	UpdatedAt        *gtime.Time `orm:"updated_at"         description:"更新时间"`                                      // 更新时间
	DeletedAt        *gtime.Time `orm:"deleted_at"         description:"删除时间"`                                      // 删除时间
}
