// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTask is the golang structure for table mvp_task.
type MvpTask struct {
	Id                uint64      `orm:"id"                 description:"雪花ID"`                                                                               // 雪花ID
	ProjectId         uint64      `orm:"project_id"         description:"项目ID"`                                                                               // 项目ID
	ParentId          uint64      `orm:"parent_id"          description:"父任务ID，0=顶级"`                                                                         // 父任务ID，0=顶级
	Name              string      `orm:"name"               description:"任务名称"`                                                                               // 任务名称
	Description       string      `orm:"description"        description:"任务描述"`                                                                               // 任务描述
	RoleType          string      `orm:"role_type"          description:"角色类型：architect/implementer/auditor/coordinator"`                                     // 角色类型：architect/implementer/auditor/coordinator
	RoleLevel         string      `orm:"role_level"         description:"角色等级：lite/pro/max"`                                                                  // 角色等级：lite/pro/max
	ModelId           uint64      `orm:"model_id"           description:"使用的AI模型ID"`                                                                          // 使用的AI模型ID
	ConversationId    int64       `orm:"conversation_id"    description:"任务对话ID，用于检测任务状态"`                                                                    // 任务对话ID，用于检测任务状态
	Status            string      `orm:"status"             description:"状态：pending/running/auditing/completed/bug_found/bug_dispatched/submit_error/failed"` // 状态：pending/running/auditing/completed/bug_found/bug_dispatched/submit_error/failed
	Sort              int         `orm:"sort"               description:"排序"`                                                                                 // 排序
	BatchNo           int         `orm:"batch_no"           description:"执行批次号，同批次内可并行，批次间串行"`                                                                // 执行批次号，同批次内可并行，批次间串行
	AffectedResources string      `orm:"affected_resources" description:"涉及的资源范围（文件/模块），用于并发冲突检测"`                                                            // 涉及的资源范围（文件/模块），用于并发冲突检测
	DependsOn         string      `orm:"depends_on"         description:"依赖的任务ID列表"`                                                                          // 依赖的任务ID列表
	Result            string      `orm:"result"             description:"任务执行结果"`                                                                             // 任务执行结果
	ContextSummary    string      `orm:"context_summary"    description:"任务完成后的上下文压缩摘要，供后续AI读取"`                                                              // 任务完成后的上下文压缩摘要，供后续AI读取
	ErrorMessage      string      `orm:"error_message"      description:"错误信息"`                                                                               // 错误信息
	StartedAt         *gtime.Time `orm:"started_at"         description:"开始时间"`                                                                               // 开始时间
	CompletedAt       *gtime.Time `orm:"completed_at"       description:"完成时间"`                                                                               // 完成时间
	CreatedBy         uint64      `orm:"created_by"         description:"创建人ID"`                                                                              // 创建人ID
	DeptId            uint64      `orm:"dept_id"            description:"所属部门ID"`                                                                             // 所属部门ID
	CreatedAt         *gtime.Time `orm:"created_at"         description:"创建时间"`                                                                               // 创建时间
	UpdatedAt         *gtime.Time `orm:"updated_at"         description:"更新时间"`                                                                               // 更新时间
	DeletedAt         *gtime.Time `orm:"deleted_at"         description:"软删除时间"`                                                                              // 软删除时间
}
