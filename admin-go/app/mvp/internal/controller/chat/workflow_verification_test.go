package chat

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestBuildVerificationRepairReason(t *testing.T) {
	t.Parallel()

	issues := []g.Map{
		{
			"severity":         "error",
			"title":            "前端类型检查失败",
			"detail":           "web-antd 全量类型检查返回真实 TS 错误",
			"suggested_action": "修复类型错误后重新执行受控类型检查",
		},
		{
			"severity":         "warn",
			"title":            "缺少浏览器证据",
			"detail":           "未采集到端到端关键路径截图",
			"suggested_action": "补充 Playwright 证据",
		},
	}

	reason := buildVerificationRepairReason(issues, "本轮修复后重新验证")
	for _, fragment := range []string{
		"基于验证问题触发返工：",
		"[error] 前端类型检查失败 - web-antd 全量类型检查返回真实 TS 错误；建议：修复类型错误后重新执行受控类型检查",
		"[warn] 缺少浏览器证据 - 未采集到端到端关键路径截图；建议：补充 Playwright 证据",
		"附加说明：本轮修复后重新验证",
	} {
		if !strings.Contains(reason, fragment) {
			t.Fatalf("reason missing %q: %s", fragment, reason)
		}
	}
}

func TestPickVerificationRepairTaskID(t *testing.T) {
	t.Parallel()

	issues := []g.Map{
		{"domain_task_id": int64(0)},
		{"domain_task_id": int64(2026041101)},
		{"domain_task_id": int64(2026041102)},
	}

	if got := pickVerificationRepairTaskID(issues); got != 2026041101 {
		t.Fatalf("pickVerificationRepairTaskID() = %d, want %d", got, int64(2026041101))
	}
}
