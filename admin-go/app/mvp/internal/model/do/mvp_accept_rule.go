// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptRule is the golang structure of table mvp_accept_rule for DAO operations like Where/Data.
type MvpAcceptRule struct {
	g.Meta      `orm:"table:mvp_accept_rule, do:true"`
	Id          any         // 主键ID
	ProjectType any         // 项目类型模板
	RuleCode    any         // 规则编码
	RuleName    any         // 规则名称
	RuleType    any         // artifact/process/quality
	ScopeType   any         // project/task/file/stage
	ConfigJson  any         // 规则配置
	Enabled     any         // 是否启用
	Priority    any         // 优先级(越小越先执行)
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
	DeletedAt   *gtime.Time // 删除时间
}
