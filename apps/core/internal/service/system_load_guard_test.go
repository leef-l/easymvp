package service

import (
	"testing"
)

func TestSystemLoadGuardDefaultsToHealthy(t *testing.T) {
	t.Parallel()
	guard := NewSystemLoadGuard()
	if guard.Status() != LoadGuardHealthy {
		t.Fatalf("expected default status healthy, got %s", guard.Status())
	}
	if !guard.AllowRun() {
		t.Fatal("expected AllowRun=true when healthy")
	}
}

func TestSystemLoadGuardStatusTransitions(t *testing.T) {
	t.Parallel()
	guard := NewSystemLoadGuard()

	// Simulate CPU >= 80 → Stopped
	guard.mu.Lock()
	guard.status = LoadGuardHealthy
	guard.mu.Unlock()

	// Directly manipulate internal state to verify AllowRun behavior
	guard.mu.Lock()
	guard.status = LoadGuardStopped
	guard.mu.Unlock()
	if guard.AllowRun() {
		t.Fatal("expected AllowRun=false when stopped")
	}

	guard.mu.Lock()
	guard.status = LoadGuardThrottled
	guard.mu.Unlock()
	if !guard.AllowRun() {
		t.Fatal("expected AllowRun=true when throttled")
	}
}

func TestSystemLoadGuardLastCPUPercentDefaultsToZero(t *testing.T) {
	t.Parallel()
	guard := NewSystemLoadGuard()
	if guard.LastCPUPercent() != 0 {
		t.Fatalf("expected default LastCPUPercent=0, got %f", guard.LastCPUPercent())
	}
}
