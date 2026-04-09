package chat

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
)

func TestBuildAcceptIssueReworkReason(t *testing.T) {
	t.Parallel()

	issues := gdb.Result{
		gdb.Record{
			"rule_code":        gvar.New("accept.diff_required"),
			"title":            gvar.New("缺少 diff"),
			"detail":           gvar.New("验收记录未附带 diff 证据"),
			"suggested_action": gvar.New("补充 patch 或 PR 草稿"),
		},
		gdb.Record{
			"rule_code":        gvar.New("accept.ci_required"),
			"title":            gvar.New("缺少 CI 结果"),
			"detail":           gvar.New("未采集到最近一次 CI 输出"),
			"suggested_action": gvar.New("补充 CI 结果文件"),
		},
	}

	reason := buildAcceptIssueReworkReason(issues, "需要本轮修复后重新验收")
	for _, fragment := range []string{
		"基于验收问题触发返工：",
		"[accept.diff_required] 缺少 diff - 验收记录未附带 diff 证据；建议：补充 patch 或 PR 草稿",
		"[accept.ci_required] 缺少 CI 结果 - 未采集到最近一次 CI 输出；建议：补充 CI 结果文件",
		"附加说明：需要本轮修复后重新验收",
	} {
		if !strings.Contains(reason, fragment) {
			t.Fatalf("reason missing %q: %s", fragment, reason)
		}
	}
}
