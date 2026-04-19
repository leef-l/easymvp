package service

import "testing"

func TestDeriveVerificationCurrentChannel(t *testing.T) {
	t.Parallel()

	if got := deriveVerificationCurrentChannel(false); got != "github_actions" {
		t.Fatalf("unexpected automated channel: got %q", got)
	}
	if got := deriveVerificationCurrentChannel(true); got != "manual_review" {
		t.Fatalf("unexpected manual review channel: got %q", got)
	}
}

func TestBuildVerificationContractJSONIncludesPreferredAndFallbackChannels(t *testing.T) {
	t.Parallel()

	raw := buildVerificationContractJSON(verificationContractParams{
		ProjectCategory:      "web",
		ProfileVersion:       "v3",
		RequiredChecks:       []string{"coverage_matrix"},
		RequiredEvidence:     []string{"ci_result"},
		ManualReviewRequired: true,
		ManualReviewSummary:  "manual checkpoint required",
	})
	if raw == "" || raw == "{}" {
		t.Fatalf("expected non-empty verification contract json, got %q", raw)
	}

	payload := mustUnmarshalJSONObject(raw)
	if payload["preferred_channel"] != "high_spec_remote" {
		t.Fatalf("unexpected preferred channel payload: %#v", payload)
	}
	if payload["current_channel"] != "manual_review" {
		t.Fatalf("unexpected current channel payload: %#v", payload)
	}
}
