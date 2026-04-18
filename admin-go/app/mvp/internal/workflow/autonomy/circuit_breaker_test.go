package autonomy

import (
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestCountReworkRoundsByChainKeyUsesPerChainMaximum(t *testing.T) {
	t.Parallel()

	handoffs := []g.Map{
		{"from_task_id": int64(101)},
		{"from_task_id": int64(101)},
		{"from_task_id": int64(202)},
		{"from_task_id": int64(101)},
		{"from_task_id": int64(202)},
	}
	chainKeys := map[int64]int64{
		101: 1,
		202: 2,
	}

	got := countReworkRoundsByChainKey(handoffs, chainKeys)
	if got[1] != 3 || got[2] != 2 {
		t.Fatalf("countReworkRoundsByChainKey() = %+v", got)
	}
}

func TestCountReworkRoundsByChainKeyIgnoresInvalidTaskIDs(t *testing.T) {
	t.Parallel()

	handoffs := []g.Map{
		{"from_task_id": int64(0)},
		{"from_task_id": nil},
		{"from_task_id": int64(303)},
	}

	got := countReworkRoundsByChainKey(handoffs, nil)
	if got[303] != 1 || len(got) != 1 {
		t.Fatalf("countReworkRoundsByChainKey() = %+v", got)
	}
}
