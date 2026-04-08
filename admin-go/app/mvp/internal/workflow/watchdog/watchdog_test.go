package watchdog

import (
	"testing"
	"time"
)

func TestResolveRefTimePrefersHeartbeat(t *testing.T) {
	raw, refTime, ok := resolveRefTime("2026-04-08 14:56:09", "2026-04-08 14:50:01")
	if !ok {
		t.Fatal("expected heartbeat ref time")
	}
	if raw != "2026-04-08 14:56:09" {
		t.Fatalf("unexpected raw ref: %s", raw)
	}
	if refTime.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", refTime.Location())
	}
	if refTime.Hour() != 14 || refTime.Minute() != 56 || refTime.Second() != 9 {
		t.Fatalf("unexpected parsed time: %v", refTime)
	}
}

func TestResolveRefTimeFallsBackToStartedAt(t *testing.T) {
	raw, refTime, ok := resolveRefTime("", "2026-04-08T14:50:01Z")
	if !ok {
		t.Fatal("expected started_at fallback")
	}
	if raw != "2026-04-08T14:50:01Z" {
		t.Fatalf("unexpected raw ref: %s", raw)
	}
	if !refTime.Equal(time.Date(2026, 4, 8, 14, 50, 1, 0, time.UTC)) {
		t.Fatalf("unexpected parsed time: %v", refTime)
	}
}

func TestNormalizeDBTime(t *testing.T) {
	cases := map[string]string{
		"":                    "",
		"  ":                  "",
		"NULL":                "",
		"null":                "",
		"0000-00-00 00:00:00": "",
		"2026-04-08 14:56:09": "2026-04-08 14:56:09",
	}
	for input, expected := range cases {
		if actual := normalizeDBTime(input); actual != expected {
			t.Fatalf("normalizeDBTime(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestResetRetryCountAlsoClearsStaleTracking(t *testing.T) {
	w := &DomainTaskWatchdog{
		staleCount: map[int64]int{1: 2},
		retryCount: map[int64]int{1: 3},
		lastRef:    map[int64]string{1: "2026-04-08 14:56:09"},
	}
	w.ResetRetryCount(1)
	if _, ok := w.retryCount[1]; ok {
		t.Fatal("retry count should be cleared")
	}
	if _, ok := w.staleCount[1]; ok {
		t.Fatal("stale count should be cleared")
	}
	if _, ok := w.lastRef[1]; ok {
		t.Fatal("last ref should be cleared")
	}
}
