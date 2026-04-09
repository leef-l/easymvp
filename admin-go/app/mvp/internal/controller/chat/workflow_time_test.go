package chat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
)

func TestNormalizeDBUTCGTimeConvertsUTCToLocal(t *testing.T) {
	t.Parallel()

	raw := gtime.NewFromTime(time.Date(2026, 4, 8, 15, 24, 44, 0, time.Local))
	got := normalizeDBUTCGTime(raw)
	if got == nil {
		t.Fatal("expected normalized time")
	}

	want := time.Date(2026, 4, 8, 15, 24, 44, 0, time.UTC).In(time.Local)
	if got.TimestampMilli() != want.UnixMilli() {
		t.Fatalf("normalized timestamp=%d want=%d", got.TimestampMilli(), want.UnixMilli())
	}
}

func TestNormalizeDBUTCGTimeHandlesNilAndZero(t *testing.T) {
	t.Parallel()

	if got := normalizeDBUTCGTime(nil); got != nil {
		t.Fatalf("normalizeDBUTCGTime(nil) = %v, want nil", got)
	}
	if got := normalizeDBUTCGTime(gtime.NewFromTime(time.Time{})); got != nil {
		t.Fatalf("normalizeDBUTCGTime(zero) = %v, want nil", got)
	}
}

func TestIsRecentGTimeUsesNormalizedDBTime(t *testing.T) {
	t.Parallel()

	recentUTC := time.Now().UTC().Add(-2 * time.Minute).Format(time.DateTime)
	value := gtime.New(recentUTC)
	normalized := normalizeDBUTCGTime(value)
	if normalized == nil {
		t.Fatal("expected normalized time")
	}
	if !isRecentGTime(normalized, 3*time.Minute) {
		t.Fatalf("expected %s to be considered recent after normalization", recentUTC)
	}
}

func TestIsRecentGTimeClampsFutureTimestamp(t *testing.T) {
	t.Parallel()

	future := gtime.NewFromTime(time.Now().Add(5 * time.Minute))
	if !isRecentGTime(future, time.Second) {
		t.Fatal("future timestamp should be clamped to zero delta")
	}
}

func TestBuildDomainTaskItemNormalizesTimes(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"id":                 gvar.New(1),
		"name":               gvar.New("task"),
		"description":        gvar.New("desc"),
		"status":             gvar.New("completed"),
		"role_type":          gvar.New("implementer"),
		"role_level":         gvar.New("lite"),
		"batch_no":           gvar.New(1),
		"sort":               gvar.New(1),
		"execution_mode":     gvar.New("auto"),
		"affected_resources": gvar.New(`["README.md"]`),
		"started_at":         gvar.New("2026-04-08 15:27:23"),
		"completed_at":       gvar.New("2026-04-08 15:27:30"),
		"result":             gvar.New("ok"),
		"retry_count":        gvar.New(0),
	}

	item := buildDomainTaskItem(record, taskWorkspaceMeta{
		Status:         "completed",
		CleanupStatus:  "pending",
		DeliveryMode:   "patch",
		DeliveryStatus: "ready",
		SyncStrategy:   "auto_apply",
		SyncStatus:     "applied",
		RiskLevel:      "low",
		PatchRef:       "/tmp/task.patch",
		DiffSummary:    " README.md | 2 +-",
	})
	if item.StartedAt == nil || item.CompletedAt == nil {
		t.Fatal("expected normalized times")
	}

	wantStart := time.Date(2026, 4, 8, 15, 27, 23, 0, time.UTC).In(time.Local).UnixMilli()
	wantFinish := time.Date(2026, 4, 8, 15, 27, 30, 0, time.UTC).In(time.Local).UnixMilli()
	if item.StartedAt.TimestampMilli() != wantStart {
		t.Fatalf("startedAt=%d want=%d", item.StartedAt.TimestampMilli(), wantStart)
	}
	if item.CompletedAt.TimestampMilli() != wantFinish {
		t.Fatalf("completedAt=%d want=%d", item.CompletedAt.TimestampMilli(), wantFinish)
	}

	data, err := json.Marshal(item.AffectedResources)
	if err != nil {
		t.Fatalf("marshal resources: %v", err)
	}
	if string(data) != `["README.md"]` {
		t.Fatalf("resources=%s", string(data))
	}
	if item.DeliveryMode != "patch" || item.SyncStatus != "applied" || item.PatchRef != "/tmp/task.patch" {
		t.Fatalf("unexpected workspace metadata: %+v", item)
	}
}

func TestBuildStageHistoryItemNormalizesTimes(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"id":            gvar.New(11),
		"stage_type":    gvar.New("design"),
		"stage_no":      gvar.New(1),
		"status":        gvar.New("running"),
		"started_at":    gvar.New("2026-04-09 07:09:24"),
		"finished_at":   gvar.New("2026-04-09 07:19:24"),
		"error_message": gvar.New(""),
	}

	item := buildStageHistoryItem(record)
	if item.StartedAt == nil || item.FinishedAt == nil {
		t.Fatalf("expected normalized stage times: %+v", item)
	}

	wantStart := time.Date(2026, 4, 9, 7, 9, 24, 0, time.UTC).In(time.Local).UnixMilli()
	wantFinish := time.Date(2026, 4, 9, 7, 19, 24, 0, time.UTC).In(time.Local).UnixMilli()
	if item.StartedAt.TimestampMilli() != wantStart {
		t.Fatalf("startedAt=%d want=%d", item.StartedAt.TimestampMilli(), wantStart)
	}
	if item.FinishedAt.TimestampMilli() != wantFinish {
		t.Fatalf("finishedAt=%d want=%d", item.FinishedAt.TimestampMilli(), wantFinish)
	}
}

func TestBuildReviewStageTaskNormalizesTimes(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"id":           gvar.New(21),
		"task_type":    gvar.New("auditor_review"),
		"role_type":    gvar.New("auditor"),
		"status":       gvar.New("completed"),
		"started_at":   gvar.New("2026-04-09 07:27:14"),
		"completed_at": gvar.New("2026-04-09 07:27:20"),
	}

	item := buildReviewStageTask(record)
	if item.StartedAt == nil || item.CompletedAt == nil {
		t.Fatalf("expected normalized review stage task times: %+v", item)
	}

	wantStart := time.Date(2026, 4, 9, 7, 27, 14, 0, time.UTC).In(time.Local).UnixMilli()
	wantFinish := time.Date(2026, 4, 9, 7, 27, 20, 0, time.UTC).In(time.Local).UnixMilli()
	if item.StartedAt.TimestampMilli() != wantStart {
		t.Fatalf("startedAt=%d want=%d", item.StartedAt.TimestampMilli(), wantStart)
	}
	if item.CompletedAt.TimestampMilli() != wantFinish {
		t.Fatalf("completedAt=%d want=%d", item.CompletedAt.TimestampMilli(), wantFinish)
	}
}

func TestBuildReviewIssueItemNormalizesTime(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"id":          gvar.New(31),
		"severity":    gvar.New("warning"),
		"issue_code":  gvar.New("short_desc"),
		"source_role": gvar.New("precheck"),
		"task_name":   gvar.New("README"),
		"message":     gvar.New("desc too short"),
		"suggestion":  gvar.New("expand description"),
		"status":      gvar.New("open"),
		"created_at":  gvar.New("2026-04-09 07:27:21"),
	}

	item := buildReviewIssueItem(record)
	if item.CreatedAt == nil {
		t.Fatalf("expected normalized review issue time: %+v", item)
	}

	want := time.Date(2026, 4, 9, 7, 27, 21, 0, time.UTC).In(time.Local).UnixMilli()
	if item.CreatedAt.TimestampMilli() != want {
		t.Fatalf("createdAt=%d want=%d", item.CreatedAt.TimestampMilli(), want)
	}
}

func TestFormatTimelineLabelForVerificationEvent(t *testing.T) {
	t.Parallel()

	payload := `{"reason":"飞书触发"}`
	if got := formatTimelineLabel("verification.started", payload); got != "验证已启动：飞书触发" {
		t.Fatalf("formatTimelineLabel() = %q", got)
	}
}

func TestBuildDomainTaskItemHandlesInvalidAffectedResourcesAndFailedFallback(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"id":                 gvar.New(2),
		"name":               gvar.New("broken task"),
		"status":             gvar.New("failed"),
		"role_type":          gvar.New("implementer"),
		"role_level":         gvar.New("lite"),
		"batch_no":           gvar.New(2),
		"sort":               gvar.New(1),
		"execution_mode":     gvar.New("aider"),
		"affected_resources": gvar.New(`{broken`),
		"result":             gvar.New("executor failed"),
		"retry_count":        gvar.New(1),
	}

	item := buildDomainTaskItem(record, taskWorkspaceMeta{})
	if len(item.AffectedResources) != 0 {
		t.Fatalf("expected empty resources on invalid json, got %+v", item.AffectedResources)
	}
	if item.ErrorMessage != "executor failed" {
		t.Fatalf("unexpected fallback error message: %s", item.ErrorMessage)
	}
}

func TestDomainTaskErrorMessagePrefersExplicitError(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"status":        gvar.New("failed"),
		"error_message": gvar.New("explicit error"),
		"result":        gvar.New("fallback result"),
	}
	if got := domainTaskErrorMessage(record); got != "explicit error" {
		t.Fatalf("domainTaskErrorMessage() = %q, want %q", got, "explicit error")
	}
}

func TestFormatTimelineLabel(t *testing.T) {
	t.Parallel()

	if got := formatTimelineLabel("stage.started", `{"stage_type":"execute"}`); got != "执行阶段已启动" {
		t.Fatalf("unexpected execute stage label: %s", got)
	}
	if got := formatTimelineLabel("stage.started", `{"stage_type":"accept"}`); got != "验收阶段已启动" {
		t.Fatalf("unexpected accept stage label: %s", got)
	}
	if got := formatTimelineLabel("workflow.force_stage", `{"stage_type":"rework","reason":"人工接管"}`); got != "工作流已人工切换到返工阶段：人工接管" {
		t.Fatalf("unexpected force stage label: %s", got)
	}
	if got := formatTimelineLabel("unknown.event", ""); got != "unknown.event" {
		t.Fatalf("unexpected fallback label: %s", got)
	}
}
