package chat

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
)

func TestManualApproveModeForPlanVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record gdb.Record
		want   string
	}{
		{
			name: "pending active plan is approvable",
			record: gdb.Record{
				"status":        gvar.New("active"),
				"review_status": gvar.New("pending"),
			},
			want: manualApproveModePending,
		},
		{
			name: "rejected draft plan is approvable",
			record: gdb.Record{
				"status":        gvar.New("draft"),
				"review_status": gvar.New("rejected"),
			},
			want: manualApproveModeRejected,
		},
		{
			name: "approved plan is not approvable again",
			record: gdb.Record{
				"status":        gvar.New("active"),
				"review_status": gvar.New("approved"),
			},
			want: "",
		},
		{
			name: "rejected active plan is not a supported override shape",
			record: gdb.Record{
				"status":        gvar.New("active"),
				"review_status": gvar.New("rejected"),
			},
			want: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := manualApproveModeForPlanVersion(tc.record)
			if got != tc.want {
				t.Fatalf("manualApproveModeForPlanVersion() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildReviewIssueReplanReason(t *testing.T) {
	t.Parallel()

	issues := gdb.Result{
		gdb.Record{
			"severity":   gvar.New("error"),
			"issue_code": gvar.New("missing_blueprint"),
			"message":    gvar.New("缺少任务拆分"),
			"suggestion": gvar.New("补齐执行步骤"),
			"task_name":  gvar.New("用户登录"),
		},
		gdb.Record{
			"severity":   gvar.New("warning"),
			"issue_code": gvar.New("batch_order"),
			"message":    gvar.New("批次依赖顺序不合理"),
			"suggestion": gvar.New("调整批次顺序"),
			"task_name":  gvar.New("订单中心"),
		},
	}

	reason := buildReviewIssueReplanReason(issues, "需要按审核意见重提方案")
	for _, fragment := range []string{
		"基于审核问题发起方案修订：",
		"1. [error/missing_blueprint] 缺少任务拆分；建议：补齐执行步骤；关联蓝图：用户登录",
		"2. [warning/batch_order] 批次依赖顺序不合理；建议：调整批次顺序；关联蓝图：订单中心",
		"附加说明：需要按审核意见重提方案",
	} {
		if !strings.Contains(reason, fragment) {
			t.Fatalf("reason missing %q: %s", fragment, reason)
		}
	}
}

func TestBuildReviewIssueReplanReasonTruncatesAfterFiveItems(t *testing.T) {
	t.Parallel()

	issues := make(gdb.Result, 0, 6)
	for i := 0; i < 6; i++ {
		issues = append(issues, gdb.Record{
			"severity":   gvar.New("warning"),
			"issue_code": gvar.New("issue_code"),
			"message":    gvar.New("issue"),
		})
	}

	reason := buildReviewIssueReplanReason(issues, "")
	if !strings.Contains(reason, "其余 1 条问题请查看审核问题列表。") {
		t.Fatalf("expected truncation notice, got: %s", reason)
	}
}
