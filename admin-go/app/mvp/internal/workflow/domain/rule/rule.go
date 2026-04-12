// Package rule 提供 CEL 规则引擎骨架（当前为 Go 原生实现，预留 CEL 扩展接口）。
//
// 设计目标：
//   - 替换 engine/review_precheck.go 中的硬编码审核规则（PR-10）
//   - 当 precheck.use_rule_engine=true 时，通过 Registry 驱动规则检查
//   - Rule.Expression 字段预留给未来 CEL 表达式，当前不解析
package rule

import "context"

// Rule 审核规则定义。
type Rule struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Severity    string `yaml:"severity" json:"severity"` // error / warning
	Category    string `yaml:"category" json:"category"` // precheck / accept_guard / autonomy
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Expression  string `yaml:"expression" json:"expression"` // 预留 CEL 表达式，当前未解析
}

// RuleSet 规则集合（用于 YAML 加载）。
type RuleSet struct {
	Rules []Rule `yaml:"rules" json:"rules"`
}

// CheckContext 规则检查上下文（传入 Checker 的数据）。
// 只包含单任务可验证的字段；跨任务规则（资源冲突、batch_no 一致性）由 precheck 层处理。
type CheckContext struct {
	TaskName          string
	TaskDescription   string
	BatchNo           int
	RoleType          string
	RoleLevel         string
	AffectedResources []string
	DependsOn         []string
	WorkDir           string
	ProjectCategory   string
	CategoryFamily    string // coding / creative / analysis
}

// CheckResult 单条规则检查结果。
// Passed=false 表示规则未通过；Passed=true 表示通过（通常不加入结果列表，由 Checker 过滤）。
type CheckResult struct {
	RuleName string
	Severity string // error / warning
	Message  string
	Passed   bool
}

// Checker 规则检查器接口。
// 实现者针对单个 CheckContext 执行一组规则，返回所有未通过的结果。
// 通过的规则不需要包含在返回值中。
type Checker interface {
	Check(ctx context.Context, rctx *CheckContext) []CheckResult
}
