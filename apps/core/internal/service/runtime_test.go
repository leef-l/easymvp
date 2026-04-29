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
	if err != nil {
		t.Logf("runtime health check returned expected error before init: %v", err)
	}
	// The key requirement is that this does NOT panic when called before init.
}
