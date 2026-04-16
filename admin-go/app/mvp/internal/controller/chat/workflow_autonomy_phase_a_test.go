package chat

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/autonomy"
)

func TestBuildSituationSnapshotItemParsesSnapshotData(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(&autonomy.Situation{
		WorkflowRunID:  88,
		ProjectID:      66,
		WorkflowStatus: "executing",
		ActiveStage:    "execute",
		SnapshotAt:     gtime.New("2026-04-12 10:00:00"),
		Health: &autonomy.HealthMetrics{
			RetryCount:       4,
			TaskRetryCount:   1,
			ReworkRounds:     2,
			TaskReworkRounds: 0,
			FocusedTaskID:    123,
		},
	})
	if err != nil {
		t.Fatalf("marshal situation: %v", err)
	}

	item, ok := buildSituationSnapshotItem(g.Map{
		"id":              int64(1),
		"workflow_run_id": int64(88),
		"project_id":      int64(66),
		"snapshot_data":   string(raw),
		"created_at":      gtime.New("2026-04-12 10:00:00"),
	})
	if !ok {
		t.Fatal("expected snapshot item to parse")
	}
	if mapJsonInt64(item, "workflowRunId") != 88 || mapJsonInt64(item, "projectId") != 66 {
		t.Fatalf("unexpected identifiers: %+v", item)
	}
	health := g.NewVar(item["health"]).Map()
	if g.NewVar(health["focusedTaskId"]).Int64() != 123 {
		t.Fatalf("expected focusedTaskId=123, got %+v", health)
	}
}

func TestBuildSituationSnapshotItemRejectsBadJSON(t *testing.T) {
	t.Parallel()

	if _, ok := buildSituationSnapshotItem(g.Map{"snapshot_data": "{bad"}); ok {
		t.Fatal("expected invalid snapshot json to be rejected")
	}
}

func TestMatchSituationSnapshotFilters(t *testing.T) {
	t.Parallel()

	taskItem := g.Map{
		"workflowRunId": int64(88),
		"health": g.Map{
			"focusedTaskId": int64(123),
		},
	}
	workflowItem := g.Map{
		"workflowRunId": int64(88),
		"health":        g.Map{},
	}
	if !matchSituationSnapshotFilters(taskItem, 88, 123) {
		t.Fatal("expected filters to match")
	}
	if matchSituationSnapshotFilters(taskItem, 99, 123) {
		t.Fatal("did not expect workflow filter to match")
	}
	if matchSituationSnapshotFilters(taskItem, 88, 456) {
		t.Fatal("did not expect task filter to match")
	}
	if !matchSituationSnapshotFilters(workflowItem, 88, 0) {
		t.Fatal("expected workflow-level snapshot to match generic history query")
	}
	if matchSituationSnapshotFilters(taskItem, 88, 0) {
		t.Fatal("did not expect task-focused snapshot in generic history query")
	}
}

func TestBuildSituationSnapshotItemFallsBackToCreatedAt(t *testing.T) {
	t.Parallel()

	createdAt := gtime.NewFromTime(time.Date(2026, 4, 12, 2, 0, 0, 0, time.UTC))
	raw, err := json.Marshal(map[string]any{
		"workflowRunId": 88,
		"projectId":     66,
		"health":        map[string]any{},
	})
	if err != nil {
		t.Fatalf("marshal snapshot payload: %v", err)
	}

	item, ok := buildSituationSnapshotItem(g.Map{
		"id":              int64(2),
		"workflow_run_id": int64(88),
		"project_id":      int64(66),
		"snapshot_data":   string(raw),
		"created_at":      createdAt,
	})
	if !ok {
		t.Fatal("expected snapshot item to parse")
	}
	if item["snapshotAt"] == nil {
		t.Fatalf("expected snapshotAt fallback, got %+v", item)
	}
}

func TestLoadSituationHistorySnapshotsPagesUntilEnoughMatches(t *testing.T) {
	t.Parallel()

	var callCount int
	makeRecord := func(id, workflowRunID, projectID, focusedTaskID int64) g.Map {
		payload := map[string]any{
			"workflowRunId": workflowRunID,
			"projectId":     projectID,
			"health": map[string]any{
				"focusedTaskId": focusedTaskID,
			},
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		return g.Map{
			"id":              id,
			"workflow_run_id": workflowRunID,
			"project_id":      projectID,
			"snapshot_data":   string(raw),
			"created_at":      gtime.New("2026-04-12 10:00:00"),
		}
	}

	snapshots, err := loadSituationHistorySnapshots(
		func(offset, limit int) ([]g.Map, error) {
			callCount++
			switch callCount {
			case 1:
				records := make([]g.Map, 0, limit)
				for i := 0; i < limit; i++ {
					focusedTaskID := int64(0)
					if i%2 == 1 {
						focusedTaskID = 456
					}
					records = append(records, makeRecord(int64(i+1), 88, 66, focusedTaskID))
				}
				return records, nil
			case 2:
				return []g.Map{
					makeRecord(99, 88, 66, 123),
				}, nil
			default:
				t.Fatalf("unexpected extra loader call offset=%d limit=%d", offset, limit)
				return nil, nil
			}
		},
		88,
		123,
		1,
	)
	if err != nil {
		t.Fatalf("loadSituationHistorySnapshots error: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected two loader calls, got %d", callCount)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
	if mapJsonInt64(snapshots[0], "id") != 99 {
		t.Fatalf("expected second page match, got %+v", snapshots[0])
	}
}

func TestLoadSituationHistorySnapshotsReturnsLoaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	_, err := loadSituationHistorySnapshots(
		func(offset, limit int) ([]g.Map, error) {
			return nil, expectedErr
		},
		88,
		0,
		20,
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected loader error %v, got %v", expectedErr, err)
	}
}
