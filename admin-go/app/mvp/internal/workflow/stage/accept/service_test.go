package accept

import (
	"context"
	"testing"
)

func TestManualCompleteBypassContext(t *testing.T) {
	t.Parallel()

	if HasManualCompleteBypass(nil) {
		t.Fatal("nil context should not bypass")
	}
	if HasManualCompleteBypass(context.Background()) {
		t.Fatal("background context should not bypass")
	}

	ctx := WithManualCompleteBypass(context.Background())
	if !HasManualCompleteBypass(ctx) {
		t.Fatal("expected manual bypass flag")
	}
}
