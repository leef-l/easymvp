package service

import (
	"path/filepath"
	"testing"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
)

func TestBuildArtifactIssueDetectsMissingAndStaleIndex(t *testing.T) {
	t.Parallel()

	missing := buildArtifactIssue("replay", "replay_1", "artifact_missing", "/tmp/missing.json", "tool result")
	if missing == nil || missing.Kind != "missing_artifact" {
		t.Fatalf("expected missing artifact issue, got %#v", missing)
	}

	stalePath := filepath.Join(t.TempDir(), "gone.json")
	stale := buildArtifactIssue("log_segment", "seg_1", "available", stalePath, "stdout")
	if stale == nil || stale.Kind != "stale_index" {
		t.Fatalf("expected stale index issue, got %#v", stale)
	}

	available := buildArtifactIssue("replay", "replay_2", "available", "", "snapshot")
	if available != nil {
		t.Fatalf("did not expect issue for empty available path, got %#v", available)
	}
}

func TestBuildReplayDiagnosticHintsIncludesArtifactAndMarkerHints(t *testing.T) {
	t.Parallel()

	hints := buildReplayDiagnosticHints(
		replayv1.ReplayArtifactSummary{Missing: 1},
		[]replayv1.ReplayArtifactIssue{{Kind: "stale_index"}},
		nil,
		nil,
	)

	codes := make(map[string]struct{}, len(hints))
	for _, hint := range hints {
		codes[hint.Code] = struct{}{}
	}
	for _, code := range []string{"missing_artifact", "stale_index", "missing_runtime_markers"} {
		if _, ok := codes[code]; !ok {
			t.Fatalf("expected diagnostic hint %s in %#v", code, hints)
		}
	}
}
