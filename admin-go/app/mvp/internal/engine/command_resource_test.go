package engine

import (
	"context"
	"testing"
)

func TestGetCommandResourcePolicyReturnsSaneDefaults(t *testing.T) {
	t.Parallel()

	policy := GetCommandResourcePolicy(context.Background())
	if policy.NPMJobs < 1 {
		t.Fatalf("NPMJobs = %d", policy.NPMJobs)
	}
	if policy.PNPMChildConcurrency < 1 {
		t.Fatalf("PNPMChildConcurrency = %d", policy.PNPMChildConcurrency)
	}
	if policy.GoMaxProcs < 1 {
		t.Fatalf("GoMaxProcs = %d", policy.GoMaxProcs)
	}
}
