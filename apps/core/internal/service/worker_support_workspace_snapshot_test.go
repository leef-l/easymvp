package service

import (
	"testing"
	"time"
)

func TestIsSnapshotFresh(t *testing.T) {
	t.Parallel()

	if !isSnapshotFresh(time.Now().Add(-10 * time.Second).Format(time.RFC3339)) {
		t.Fatal("expected recent snapshot to be fresh")
	}
	if isSnapshotFresh(time.Now().Add(-2 * time.Minute).Format(time.RFC3339)) {
		t.Fatal("expected expired snapshot to be stale")
	}
	if isSnapshotFresh("not-a-time") {
		t.Fatal("expected invalid timestamp to be stale")
	}
}
