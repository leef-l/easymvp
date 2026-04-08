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

	item := buildDomainTaskItem(record)
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
}
