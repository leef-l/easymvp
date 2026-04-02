// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AiModel is the golang structure of table ai_model for DAO operations like Where/Data.
type AiModel struct {
	g.Meta         `orm:"table:ai_model, do:true"`
	Id             any         // 雪花ID
	PlanId         any         // 套餐ID
	ProviderId     any         // 供应商ID（冗余便于查询）
	Name           any         // 模型显示名称
	ModelCode      any         // 模型代码（API调用用）
	Capability     any         // 能力：chat/reasoning/coding
	MaxTokens      any         // 最大输出token
	ContextWindow  any         // 上下文窗口大小
	SupportsStream any         // 是否支持流式输出:0=否,1=是
	RolePrompt     any         // 默认角色提示词
	Status         any         // 状态:0=禁用,1=启用
	Sort           any         // 排序
	CreatedBy      any         // 创建人ID
	DeptId         any         // 所属部门ID
	CreatedAt      *gtime.Time // 创建时间
	UpdatedAt      *gtime.Time // 更新时间
	DeletedAt      *gtime.Time // 软删除时间
}
