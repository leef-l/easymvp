package service

import (
	"context"
	"testing"
)

func TestRuntimeCheckHealthReturnsErrorInsteadOfPanickingBeforeInit(t *testing.T) {
	t.Parallel()

	previous := localRuntime
	localRuntime = nil
	defer func() {
		localRuntime = previous
	}()

	err := Runtime().CheckHealth(context.Background())
	if err == nil {
		t.Fatal("expected runtime health check to fail without configured brain service")
	}
}
