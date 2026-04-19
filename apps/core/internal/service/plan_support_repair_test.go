package service

import (
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

func TestCanonicalizeJSONObjectStringNormalizesObjectKeyOrder(t *testing.T) {
	t.Parallel()

	left, err := canonicalizeJSONObjectString(`{"b":2,"a":1}`)
	if err != nil {
		t.Fatalf("canonicalize left json failed: %v", err)
	}
	right, err := canonicalizeJSONObjectString(`{"a":1,"b":2}`)
	if err != nil {
		t.Fatalf("canonicalize right json failed: %v", err)
	}
	if left != right {
		t.Fatalf("expected canonical json to match: left=%s right=%s", left, right)
	}
}

func TestBuildRepairArtifactRefsJSONStableForEquivalentRefs(t *testing.T) {
	t.Parallel()

	left := buildRepairArtifactRefsJSON([]braincontracts.ArtifactRef{
		{Kind: "evidence", ID: "ev_1", Path: "/tmp/evidence.png"},
	})
	right := buildRepairArtifactRefsJSON([]braincontracts.ArtifactRef{
		{Kind: "evidence", ID: "ev_1", Path: "/tmp/evidence.png"},
	})
	if left != right {
		t.Fatalf("expected artifact refs json to match: left=%s right=%s", left, right)
	}
}
