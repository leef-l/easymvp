package engine

import (
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
)

func TestBlueprintsToTaskRecordsConvertsDependsOnIDsToNames(t *testing.T) {
	t.Parallel()

	blueprints := gdb.Result{
		gdb.Record{
			"id":                       gvar.New(int64(101)),
			"name":                     gvar.New("frontend_init"),
			"description":              gvar.New("初始化前端"),
			"role_type":                gvar.New("implementer"),
			"role_level":               gvar.New("lite"),
			"batch_no":                 gvar.New(1),
			"affected_resources":       gvar.New(`["frontend/package.json"]`),
			"depends_on_blueprint_ids": gvar.New(`[]`),
		},
		gdb.Record{
			"id":                       gvar.New(int64(102)),
			"name":                     gvar.New("game_state"),
			"description":              gvar.New("管理游戏状态"),
			"role_type":                gvar.New("implementer"),
			"role_level":               gvar.New("pro"),
			"batch_no":                 gvar.New(2),
			"affected_resources":       gvar.New(`["frontend/src/hooks/useGameState.js"]`),
			"depends_on_blueprint_ids": gvar.New(`[101]`),
		},
	}

	tasks := blueprintsToTaskRecords(blueprints)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 task records, got %d", len(tasks))
	}
	if got := tasks[1]["depends_on"].String(); got != `["frontend_init"]` {
		t.Fatalf("unexpected depends_on payload: %s", got)
	}
}
