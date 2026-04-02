// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AiModel is the golang structure for table ai_model.
type AiModel struct {
	Id             uint64      `orm:"id"              description:"雪花ID"`                     // 雪花ID
	PlanId         uint64      `orm:"plan_id"         description:"套餐ID"`                     // 套餐ID
	ProviderId     uint64      `orm:"provider_id"     description:"供应商ID（冗余便于查询）"`            // 供应商ID（冗余便于查询）
	Name           string      `orm:"name"            description:"模型显示名称"`                   // 模型显示名称
	ModelCode      string      `orm:"model_code"      description:"模型代码（API调用用）"`             // 模型代码（API调用用）
	Capability     string      `orm:"capability"      description:"能力：chat/reasoning/coding"` // 能力：chat/reasoning/coding
	MaxTokens      int         `orm:"max_tokens"      description:"最大输出token"`                // 最大输出token
	ContextWindow  int         `orm:"context_window"  description:"上下文窗口大小"`                  // 上下文窗口大小
	SupportsStream int         `orm:"supports_stream" description:"是否支持流式输出:0=否,1=是"`         // 是否支持流式输出:0=否,1=是
	RolePrompt     string      `orm:"role_prompt"     description:"默认角色提示词"`                  // 默认角色提示词
	Status         int         `orm:"status"          description:"状态:0=禁用,1=启用"`             // 状态:0=禁用,1=启用
	Sort           int         `orm:"sort"            description:"排序"`                       // 排序
	CreatedBy      uint64      `orm:"created_by"      description:"创建人ID"`                    // 创建人ID
	DeptId         uint64      `orm:"dept_id"         description:"所属部门ID"`                   // 所属部门ID
	CreatedAt      *gtime.Time `orm:"created_at"      description:"创建时间"`                     // 创建时间
	UpdatedAt      *gtime.Time `orm:"updated_at"      description:"更新时间"`                     // 更新时间
	DeletedAt      *gtime.Time `orm:"deleted_at"      description:"软删除时间"`                    // 软删除时间
}
