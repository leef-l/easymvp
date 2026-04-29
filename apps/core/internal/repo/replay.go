package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type ReplayIndexRow struct {
	ID               string
	ReplayID         string
	ProjectID        string
	RunID            string
	DomainTaskID     string
	CompiledTaskID   string
	EventID          string
	TraceID          string
	SpanID           string
	ReplayKind       string
	SeqNo            int
	Title            string
	Summary          string
	FilePath         string
	FileExt          string
	MimeType         string
	FileSize         int64
	SHA256           string
	SourceObjectKind string
	SourceObjectID   string
	Status           string
	CreatedAt        string
	UpdatedAt        string
}

type LogSegmentRow struct {
	ID         string
	ProjectID  string
	RunID      string
	SegmentID  string
	StreamKind string
	SeqNo      int
	FilePath   string
	FileSize   int64
	SHA256     string
	StartedAt  string
	EndedAt    string
	Status     string
	CreatedAt  string
}

type RunEventRow struct {
	ID         string
	SequenceNo int
	EventType  string
	EventLevel string
	Summary    string
	CreatedAt  string
}

type RunCheckpointRow struct {
	ID             string
	CheckpointType string
	CreatedAt      string
}

type BrainRunBindingViewRow struct {
	ID         string
	ProjectID  string
	TaskID     string
	BrainKind  string
	RunID      string
	RunStatus  string
	StartedAt  string
	FinishedAt string
	LastSyncAt string
	CreatedAt  string
	UpdatedAt  string
}

func GetBrainRunBindingByRunID(ctx context.Context, projectID string, runID string) (*BrainRunBindingViewRow, error) {
	var item entity.BrainRunBindings
	err := dao.BrainRunBindings.Ctx(ctx).
		Where("project_id", projectID).
		Where("brain_run_id", runID).
		Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.New("brain run binding not found")
		}
		return nil, gerror.Wrap(err, "query brain run binding by run id failed")
	}

	return &BrainRunBindingViewRow{
		ID:         item.Id,
		ProjectID:  item.ProjectId,
		TaskID:     item.TaskId,
		BrainKind:  item.BrainKind,
		RunID:      item.BrainRunId,
		RunStatus:  item.RunStatus,
		StartedAt:  item.StartedAt,
		FinishedAt: item.FinishedAt,
		LastSyncAt: item.LastSyncAt,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}, nil
}

func CountRunEventsByBindingID(ctx context.Context, bindingID string) (int, error) {
	count, err := dao.RunEventIndex.Ctx(ctx).Where("run_binding_id", bindingID).Count()
	if err != nil {
		return 0, gerror.Wrap(err, "count run events failed")
	}
	return count, nil
}

func CountReplayIndexByRunID(ctx context.Context, projectID string, runID string) (int, error) {
	count, err := g.DB().GetCount(ctx,
		"SELECT COUNT(1) FROM workflow_replay_index WHERE project_id = ? AND run_id = ?",
		projectID, runID)
	if err != nil {
		return 0, gerror.Wrap(err, "count replay index failed")
	}
	return count, nil
}

func CountLogSegmentsByRunID(ctx context.Context, projectID string, runID string) (int, error) {
	count, err := g.DB().GetCount(ctx,
		"SELECT COUNT(1) FROM workflow_run_log_segments WHERE project_id = ? AND run_id = ?",
		projectID, runID)
	if err != nil {
		return 0, gerror.Wrap(err, "count log segments failed")
	}
	return count, nil
}

func GetLatestRunEvent(ctx context.Context, bindingID string) (*replayv1.ReplayEventRef, error) {
	if strings.TrimSpace(bindingID) == "" {
		return nil, nil
	}

	var item entity.RunEventIndex
	err := dao.RunEventIndex.Ctx(ctx).
		Where("run_binding_id", bindingID).
		Order("sequence_no DESC, created_at DESC").
		Limit(1).
		Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest run event failed")
	}

	return &replayv1.ReplayEventRef{
		EventID:    item.Id,
		SequenceNo: item.SequenceNo,
		EventType:  item.EventType,
		EventLevel: item.EventLevel,
		Summary:    item.Summary,
		CreatedAt:  item.CreatedAt,
	}, nil
}

func GetRunEventByEventID(ctx context.Context, eventID string) (*replayv1.ReplayEventRef, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return nil, nil
	}

	var item entity.RunEventIndex
	err := dao.RunEventIndex.Ctx(ctx).
		Where("id", eventID).
		Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query run event by id failed")
	}

	return &replayv1.ReplayEventRef{
		EventID:    item.Id,
		SequenceNo: item.SequenceNo,
		EventType:  item.EventType,
		EventLevel: item.EventLevel,
		Summary:    item.Summary,
		CreatedAt:  item.CreatedAt,
	}, nil
}

func GetLatestRunCheckpoint(ctx context.Context, bindingID string) (*replayv1.ReplayCheckpointRef, error) {
	if strings.TrimSpace(bindingID) == "" {
		return nil, nil
	}

	var item entity.RunCheckpoints
	err := dao.RunCheckpoints.Ctx(ctx).
		Where("run_binding_id", bindingID).
		Order("created_at DESC, id DESC").
		Limit(1).
		Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest run checkpoint failed")
	}

	return &replayv1.ReplayCheckpointRef{
		CheckpointID:   item.Id,
		CheckpointType: item.CheckpointType,
		CreatedAt:      item.CreatedAt,
	}, nil
}

func SummarizeReplayArtifactStatus(ctx context.Context, projectID string, runID string) (replayv1.ReplayArtifactSummary, error) {
	var results []struct {
		Status string
		Count  int
	}
	err := g.DB().GetScan(ctx, &results,
		"SELECT status, COUNT(1) as count FROM workflow_replay_index WHERE project_id = ? AND run_id = ? GROUP BY status",
		projectID, runID)
	if err != nil {
		return replayv1.ReplayArtifactSummary{}, gerror.Wrap(err, "summarize replay artifact status failed")
	}

	summary := replayv1.ReplayArtifactSummary{}
	for _, r := range results {
		switch r.Status {
		case "artifact_missing":
			summary.Missing = r.Count
		case "artifact_pruned":
			summary.Pruned = r.Count
		default:
			summary.Available += r.Count
		}
	}
	return summary, nil
}

func ListReplayTimelineRows(ctx context.Context, projectID string, runID string, replayKind string, limit int) ([]ReplayIndexRow, error) {
	sqlStr := `SELECT
  id, replay_id, project_id, run_id,
  COALESCE(domain_task_id, ''), COALESCE(compiled_task_id, ''),
  COALESCE(event_id, ''), COALESCE(trace_id, ''), COALESCE(span_id, ''),
  replay_kind, seq_no, title, COALESCE(summary, ''),
  file_path, COALESCE(file_ext, ''), COALESCE(mime_type, ''), COALESCE(file_size, 0), COALESCE(sha256, ''),
  COALESCE(source_object_kind, ''), COALESCE(source_object_id, ''), status, created_at, updated_at
FROM workflow_replay_index
WHERE project_id = ? AND run_id = ?`
	args := []any{projectID, runID}
	if replayKind != "" {
		sqlStr += ` AND replay_kind = ?`
		args = append(args, replayKind)
	}
	sqlStr += ` ORDER BY seq_no ASC, created_at ASC, id ASC LIMIT ?`
	args = append(args, limit)

	var items []ReplayIndexRow
	if err := g.DB().GetScan(ctx, &items, sqlStr, args...); err != nil {
		return nil, gerror.Wrap(err, "query replay timeline rows failed")
	}
	return items, nil
}

func GetReplayIndexItem(ctx context.Context, projectID string, runID string, replayID string) (*ReplayIndexRow, error) {
	var item ReplayIndexRow
	err := g.DB().GetScan(ctx, &item,
		`SELECT
  id, replay_id, project_id, run_id,
  COALESCE(domain_task_id, ''), COALESCE(compiled_task_id, ''),
  COALESCE(event_id, ''), COALESCE(trace_id, ''), COALESCE(span_id, ''),
  replay_kind, seq_no, title, COALESCE(summary, ''),
  file_path, COALESCE(file_ext, ''), COALESCE(mime_type, ''), COALESCE(file_size, 0), COALESCE(sha256, ''),
  COALESCE(source_object_kind, ''), COALESCE(source_object_id, ''), status, created_at, updated_at
FROM workflow_replay_index
WHERE project_id = ? AND run_id = ? AND replay_id = ?
LIMIT 1`,
		projectID, runID, replayID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.New("replay item not found")
		}
		return nil, gerror.Wrap(err, "query replay item failed")
	}
	return &item, nil
}

func ListLogSegmentRows(ctx context.Context, projectID string, runID string, limit int) ([]LogSegmentRow, error) {
	var items []LogSegmentRow
	err := g.DB().GetScan(ctx, &items,
		`SELECT
  id, project_id, run_id, segment_id, stream_kind, seq_no, file_path,
  COALESCE(file_size, 0), COALESCE(sha256, ''), COALESCE(started_at, ''), COALESCE(ended_at, ''), status, created_at
FROM workflow_run_log_segments
WHERE project_id = ? AND run_id = ?
ORDER BY seq_no ASC, created_at ASC, id ASC
LIMIT ?`,
		projectID, runID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query log segment rows failed")
	}
	return items, nil
}

func GetLogSegmentItem(ctx context.Context, projectID string, runID string, segmentID string) (*LogSegmentRow, error) {
	var item LogSegmentRow
	err := g.DB().GetScan(ctx, &item,
		`SELECT
  id, project_id, run_id, segment_id, stream_kind, seq_no, file_path,
  COALESCE(file_size, 0), COALESCE(sha256, ''), COALESCE(started_at, ''), COALESCE(ended_at, ''), status, created_at
FROM workflow_run_log_segments
WHERE project_id = ? AND run_id = ? AND segment_id = ?
LIMIT 1`,
		projectID, runID, segmentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.New("log segment not found")
		}
		return nil, gerror.Wrap(err, "query log segment failed")
	}
	return &item, nil
}
