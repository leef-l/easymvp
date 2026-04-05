// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptRule is the golang structure for table mvp_accept_rule.
type MvpAcceptRule struct {
	Id          int64       `orm:"id"           description:"主键ID"`                     // 主键ID
	ProjectType string      `orm:"project_type" description:"项目类型模板"`                   // 项目类型模板
	RuleCode    string      `orm:"rule_code"    description:"规则编码"`                     // 规则编码
	RuleName    string      `orm:"rule_name"    description:"规则名称"`                     // 规则名称
	RuleType    string      `orm:"rule_type"    description:"artifact/process/quality"` // artifact/process/quality
	ScopeType   string      `orm:"scope_type"   description:"project/task/file/stage"`  // project/task/file/stage
	ConfigJson  string      `orm:"config_json"  description:"规则配置"`                     // 规则配置
	Enabled     int         `orm:"enabled"      description:"是否启用"`                     // 是否启用
	Priority    int         `orm:"priority"     description:"优先级(越小越先执行)"`              // 优先级(越小越先执行)
	CreatedAt   *gtime.Time `orm:"created_at"   description:"创建时间"`                     // 创建时间
	UpdatedAt   *gtime.Time `orm:"updated_at"   description:"更新时间"`                     // 更新时间
	DeletedAt   *gtime.Time `orm:"deleted_at"   description:"删除时间"`                     // 删除时间
}
