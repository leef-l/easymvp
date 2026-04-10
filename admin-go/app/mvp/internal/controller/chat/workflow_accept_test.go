package chat

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func TestBuildAcceptIssueReworkReason(t *testing.T) {
	t.Parallel()

	issues := []g.Map{
		{
			"rule_code":        "accept.diff_required",
			"title":            "缺少 diff",
			"detail":           "验收记录未附带 diff 证据",
			"suggested_action": "补充 patch 或 PR 草稿",
		},
		{
			"rule_code":        "accept.ci_required",
			"title":            "缺少 CI 结果",
			"detail":           "未采集到最近一次 CI 输出",
			"suggested_action": "补充 CI 结果文件",
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

func TestNormalizeAcceptRunStatus(t *testing.T) {
	t.Parallel()

	now := gtime.Now()
	cases := []struct {
		name       string
		status     string
		decision   string
		finishedAt *gtime.Time
		want       string
	}{
		{
			name:       "running manual review with finishedAt becomes completed",
			status:     "running",
			decision:   "manual_review",
			finishedAt: now,
			want:       "completed",
		},
		{
			name:       "running passed with finishedAt becomes completed",
			status:     "running",
			decision:   "passed",
			finishedAt: now,
			want:       "completed",
		},
		{
			name:       "running without finishedAt stays running",
			status:     "running",
			decision:   "manual_review",
			finishedAt: nil,
			want:       "running",
		},
		{
			name:       "completed stays completed",
			status:     "completed",
			decision:   "manual_review",
			finishedAt: now,
			want:       "completed",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeAcceptRunStatus(tc.status, tc.decision, tc.finishedAt); got != tc.want {
				t.Fatalf("normalizeAcceptRunStatus(%q, %q, %v) = %q, want %q", tc.status, tc.decision, tc.finishedAt, got, tc.want)
			}
		})
	}
}

func TestAcceptRelatedWorkflowStatuses(t *testing.T) {
	t.Parallel()

	statuses := acceptRelatedWorkflowStatuses()
	index := make(map[string]struct{}, len(statuses))
	for _, status := range statuses {
		index[status.(string)] = struct{}{}
	}

	for _, required := range []string{
		"accepting",
		"executing",
		"reworking",
		"paused",
		"completed",
		"running",
	} {
		if _, ok := index[required]; !ok {
			t.Fatalf("acceptRelatedWorkflowStatuses missing %q: %#v", required, statuses)
		}
	}
}
