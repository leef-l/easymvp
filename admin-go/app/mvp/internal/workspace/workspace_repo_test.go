package workspace

import (
	"errors"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestIsDeliveryReferenceColumnErr(t *testing.T) {
	t.Parallel()

	if isDeliveryReferenceColumnErr(nil) {
		t.Fatal("nil error should not match delivery reference columns")
	}
	if !isDeliveryReferenceColumnErr(errors.New("Unknown column 'delivery_ref' in 'field list'")) {
		t.Fatal("delivery_ref column error should be detected")
	}
	if !isDeliveryReferenceColumnErr(errors.New("Unknown column 'delivery_title' in 'field list'")) {
		t.Fatal("delivery_title column error should be detected")
	}
	if isDeliveryReferenceColumnErr(errors.New("Unknown column 'risk_level' in 'field list'")) {
		t.Fatal("non delivery reference column error should not match")
	}
}

func TestIsUnknownColumnErrHandlesNil(t *testing.T) {
	t.Parallel()

	if isUnknownColumnErr(nil) {
		t.Fatal("nil error should not be treated as unknown column")
	}
	if !isUnknownColumnErr(errors.New("Error 1054 (42S22): Unknown column 'foo' in 'field list'")) {
		t.Fatal("unknown column error should match")
	}
	if isUnknownColumnErr(errors.New("duplicate key")) {
		t.Fatal("non unknown-column error should not match")
	}
}

func TestFilterWorkspaceDataByError(t *testing.T) {
	t.Parallel()

	data := g.Map{
		"workspace_path":  "/tmp/demo",
		"delivery_mode":   DeliveryModePR,
		"delivery_status": DeliveryStatusReady,
		"sync_strategy":   SyncStrategyManual,
		"sync_status":     SyncStatusPending,
		"risk_level":      RiskLevelHigh,
		"patch_ref":       "patch.diff",
		"delivery_ref":    "pr://1",
		"delivery_title":  "Draft PR",
	}

	partial := filterWorkspaceDataByError(data, errors.New("Unknown column 'delivery_ref' in 'field list'"))
	if _, ok := partial["delivery_ref"]; ok {
		t.Fatalf("partial data should remove delivery_ref: %+v", partial)
	}
	if _, ok := partial["delivery_title"]; ok {
		t.Fatalf("partial data should remove delivery_title: %+v", partial)
	}
	for _, key := range []string{"workspace_path", "delivery_mode", "delivery_status", "sync_strategy", "sync_status", "risk_level", "patch_ref"} {
		if _, ok := partial[key]; !ok {
			t.Fatalf("partial data should keep %s: %+v", key, partial)
		}
	}

	legacy := filterWorkspaceDataByError(data, errors.New("Unknown column 'risk_level' in 'field list'"))
	for _, key := range []string{"delivery_mode", "delivery_status", "sync_strategy", "sync_status", "risk_level", "patch_ref", "delivery_ref", "delivery_title"} {
		if _, ok := legacy[key]; ok {
			t.Fatalf("legacy data should remove %s: %+v", key, legacy)
		}
	}
	if got := legacy["workspace_path"]; got != "/tmp/demo" {
		t.Fatalf("legacy data should keep workspace_path, got %+v", legacy)
	}
}
