package plan

import (
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"

	"easymvp/app/mvp/internal/engine"
)

func TestBuildBlueprintPatchUpdateData(t *testing.T) {
	t.Parallel()

	batchNo := 3
	patch := &engine.ArchitectTaskPatch{
		TaskName:          "资料页",
		Description:       "补齐资料保存逻辑",
		AffectedResources: []string{"frontend/profile.tsx", "backend/profile.go"},
		DependsOn:         []string{"登录"},
		BatchNo:           &batchNo,
	}

	byName := map[string]gdb.Record{
		"登录": {
			"id": gvar.New(int64(1001)),
		},
	}

	updateData, err := buildBlueprintPatchUpdateData(byName, patch)
	if err != nil {
		t.Fatalf("buildBlueprintPatchUpdateData() error = %v", err)
	}
	if updateData["description"] != "补齐资料保存逻辑" {
		t.Fatalf("unexpected description: %+v", updateData)
	}
	if updateData["batch_no"] != 3 {
		t.Fatalf("unexpected batch_no: %+v", updateData)
	}
	if updateData["depends_on_blueprint_ids"] != "[1001]" {
		t.Fatalf("unexpected depends_on_blueprint_ids: %+v", updateData)
	}
	if updateData["affected_resources"] != "[\"frontend/profile.tsx\",\"backend/profile.go\"]" {
		t.Fatalf("unexpected affected_resources: %+v", updateData)
	}
}

func TestBuildBlueprintPatchUpdateDataRejectsUnknownDependency(t *testing.T) {
	t.Parallel()

	_, err := buildBlueprintPatchUpdateData(map[string]gdb.Record{}, &engine.ArchitectTaskPatch{
		TaskName:  "订单中心",
		DependsOn: []string{"支付"},
	})
	if err == nil {
		t.Fatal("expected missing dependency error")
	}
}
