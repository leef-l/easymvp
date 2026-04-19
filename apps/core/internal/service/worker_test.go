package service

import "testing"

func TestWorkersStatusIsSafeBeforeStart(t *testing.T) {
	t.Parallel()

	previous := localWorkerManager
	localWorkerManager = nil
	defer func() {
		localWorkerManager = previous
	}()

	status := Workers().Status()
	if status.Started {
		t.Fatalf("unexpected started status before worker start: %#v", status)
	}
	if len(status.Workers) == 0 {
		t.Fatalf("expected worker registry to be initialized, got %#v", status)
	}
}
