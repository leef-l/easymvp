package engine

import "fmt"

// 任务状态机：定义合法的状态转换路径，防止非法跳转

// validTransitions 合法的任务状态转换表
// key: 当前状态, value: 允许转换到的目标状态集合
var validTransitions = map[string]map[string]bool{
	"draft": {
		"pending": true, // 确认方案时 draft → pending
	},
	"pending": {
		"running": true, // 调度器分发
		"draft":   true, // 回退到草稿
	},
	"running": {
		"completed": true, // 正常完成
		"failed":    true, // 执行失败
		"escalated": true, // 升级给架构师
		"pending":   true, // 暂停时回退
	},
	"completed": {
		"auditing":  true, // 进入审计
		"bug_found": true, // 审计发现 bug
	},
	"failed": {
		"pending":   true, // 重试
		"completed": true, // 跳过（SkipTask）
		"escalated": true, // 升级
	},
	"auditing": {
		"completed": true, // 审计通过
		"bug_found": true, // 发现 bug
	},
	"bug_found": {
		"bug_dispatched": true, // 分派修复
		"pending":        true, // 直接重做
		"completed":      true, // 跳过
	},
	"bug_dispatched": {
		"pending": true, // 修复方案回写后重新排队
	},
	"escalated": {
		"pending": true, // 架构师给出方案后重新排队
		"failed":  true, // 升级后仍然失败
	},
	"submit_error": {
		"pending": true, // 重试
	},
}

// ValidateTransition 校验状态转换是否合法
// 返回 nil 表示合法，否则返回错误
func ValidateTransition(from, to string) error {
	targets, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("未知的源状态: %s", from)
	}
	if !targets[to] {
		return fmt.Errorf("非法状态转换: %s → %s", from, to)
	}
	return nil
}
