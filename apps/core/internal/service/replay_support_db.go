package service

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
)

const (
	replaySummaryEntryPointLimit = 8
	replayTimelineDefaultLimit   = 50
	replayTimelineMaxLimit       = 200
	replayRawDefaultLimit        = 8192
	replayRawMaxLimit            = 65536
	logSegmentDefaultLimit       = 50
	logSegmentMaxLimit           = 200
)

func (s *sReplay) GetReplaySummary(ctx context.Context, projectID string, runID string) (*replayv1.GetReplaySummaryRes, error) {
	projectID = strings.TrimSpace(projectID)
	runID = strings.TrimSpace(runID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if runID == "" {
		return nil, gerror.New("run id is required")
	}

	binding, err := getBrainRunBindingByRunID(ctx, projectID, runID)
	if err != nil {
		return nil, err
	}
	eventCount, err := countRunEventsByBindingID(ctx, binding.ID)
	if err != nil {
		return nil, err
	}
	replayCount, err := countReplayIndexByRunID(ctx, projectID, runID)
	if err != nil {
		return nil, err
	}
	logSegmentCount, err := countLogSegmentsByRunID(ctx, projectID, runID)
	if err != nil {
		return nil, err
	}
	artifactSummary, err := summarizeReplayArtifactStatus(ctx, projectID, runID)
	if err != nil {
		return nil, err
	}
	entryPoints, err := listReplayEntryPoints(ctx, projectID, runID, replaySummaryEntryPointLimit)
	if err != nil {
		return nil, err
	}
	missingArtifacts, err := listRunArtifactIssues(ctx, projectID, runID, 8)
	if err != nil {
		return nil, err
	}
	latestEvent, err := getLatestRunEvent(ctx, binding.ID)
	if err != nil {
		return nil, err
	}
	latestCheckpoint, err := getLatestRunCheckpoint(ctx, binding.ID)
	if err != nil {
		return nil, err
	}

	return &replayv1.GetReplaySummaryRes{
		RunID:                 runID,
		ProjectID:             projectID,
		BrainKind:             binding.BrainKind,
		Status:                binding.RunStatus,
		StartedAt:             binding.StartedAt,
		EndedAt:               binding.FinishedAt,
		EventCount:            eventCount,
		ReplayCount:           replayCount,
		LogSegmentCount:       logSegmentCount,
		ArtifactStatusSummary: artifactSummary,
		MissingArtifacts:      missingArtifacts,
		LatestEvent:           latestEvent,
		LatestCheckpoint:      latestCheckpoint,
		DiagnosticHints:       buildReplayDiagnosticHints(artifactSummary, missingArtifacts, latestEvent, latestCheckpoint),
		EntryPoints:           entryPoints,
	}, nil
}

func (s *sReplay) GetReplayTimeline(ctx context.Context, projectID string, runID string, limit int, replayType string) (*replayv1.GetReplayTimelineRes, error) {
	projectID = strings.TrimSpace(projectID)
	runID = strings.TrimSpace(runID)
	replayType = strings.TrimSpace(replayType)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if runID == "" {
		return nil, gerror.New("run id is required")
	}
	if limit <= 0 {
		limit = replayTimelineDefaultLimit
	}
	if limit > replayTimelineMaxLimit {
		limit = replayTimelineMaxLimit
	}

	items, err := listReplayTimelineRows(ctx, projectID, runID, replayType, limit)
	if err != nil {
		return nil, err
	}

	timeline := make([]replayv1.ReplayTimelineItem, 0, len(items))
	for _, item := range items {
		issue := buildReplayArtifactIssueFromReplay(item)
		timeline = append(timeline, replayv1.ReplayTimelineItem{
			ReplayID:         item.ReplayID,
			DomainTaskID:     item.DomainTaskID,
			CompiledTaskID:   item.CompiledTaskID,
			SeqNo:            item.SeqNo,
			ReplayType:       item.ReplayKind,
			Title:            item.Title,
			Summary:          item.Summary,
			Status:           item.Status,
			PreviewAvailable: item.Status == "available" && item.FilePath != "",
			RawTarget:        item.FilePath,
			ArtifactIssue:    issue,
			DiagnosticHints:  buildItemDiagnosticHints(issue),
			CreatedAt:        item.CreatedAt,
		})
	}

	return &replayv1.GetReplayTimelineRes{
		RunID:       runID,
		ProjectID:   projectID,
		Items:       timeline,
		RefreshHint: "refresh_replay_timeline",
	}, nil
}

func (s *sReplay) GetReplayDetail(ctx context.Context, projectID string, runID string, replayID string) (*replayv1.GetReplayDetailRes, error) {
	item, err := getReplayIndexItem(ctx, projectID, runID, replayID)
	if err != nil {
		return nil, err
	}
	logSegments, err := listLogSegmentRows(ctx, projectID, runID, 10)
	if err != nil {
		return nil, err
	}
	relatedEvent, err := getRunEventByEventID(ctx, item.EventID)
	if err != nil {
		return nil, err
	}
	issue := buildReplayArtifactIssueFromReplay(*item)

	return &replayv1.GetReplayDetailRes{
		ReplayID:         item.ReplayID,
		ReplayKind:       item.ReplayKind,
		Title:            item.Title,
		Summary:          item.Summary,
		DomainTaskID:     item.DomainTaskID,
		CompiledTaskID:   item.CompiledTaskID,
		SourceObjectKind: item.SourceObjectKind,
		SourceObjectID:   item.SourceObjectID,
		EventID:          item.EventID,
		TraceID:          item.TraceID,
		SpanID:           item.SpanID,
		Status:           item.Status,
		RawPreview: replayv1.ReplayRawPreview{
			MimeType:         item.MimeType,
			PreviewAvailable: item.Status == "available" && item.FilePath != "",
			RawTarget:        item.FilePath,
		},
		ArtifactIssue:      issue,
		RelatedEvent:       relatedEvent,
		DiagnosticHints:    buildItemDiagnosticHints(issue),
		RelatedLogSegments: mapLogSegments(logSegments),
	}, nil
}

func (s *sReplay) GetReplayRaw(ctx context.Context, projectID string, runID string, replayID string, limit int) (*replayv1.GetReplayRawRes, error) {
	item, err := getReplayIndexItem(ctx, projectID, runID, replayID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = replayRawDefaultLimit
	}
	if limit > replayRawMaxLimit {
		limit = replayRawMaxLimit
	}
	content, truncated, err := loadRawFileSnippet(item.FilePath, item.Status, limit)
	if err != nil {
		return nil, err
	}
	return &replayv1.GetReplayRawRes{
		ReplayID:  replayID,
		Status:    item.Status,
		MimeType:  item.MimeType,
		Encoding:  "utf-8",
		Content:   content,
		Truncated: truncated,
	}, nil
}

func (s *sReplay) ListLogSegments(ctx context.Context, projectID string, runID string, limit int) (*replayv1.ListLogSegmentsRes, error) {
	if limit <= 0 {
		limit = logSegmentDefaultLimit
	}
	if limit > logSegmentMaxLimit {
		limit = logSegmentMaxLimit
	}
	rows, err := listLogSegmentRows(ctx, projectID, runID, limit)
	if err != nil {
		return nil, err
	}
	return &replayv1.ListLogSegmentsRes{
		RunID:       runID,
		ProjectID:   projectID,
		Segments:    mapLogSegments(rows),
		RefreshHint: "refresh_log_segments",
	}, nil
}

func (s *sReplay) GetLogSegmentRaw(ctx context.Context, projectID string, runID string, segmentID string, limit int) (*replayv1.GetLogSegmentRawRes, error) {
	item, err := getLogSegmentItem(ctx, projectID, runID, segmentID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = replayRawDefaultLimit
	}
	if limit > replayRawMaxLimit {
		limit = replayRawMaxLimit
	}
	content, truncated, err := loadRawFileSnippet(item.FilePath, item.Status, limit)
	if err != nil {
		return nil, err
	}
	return &replayv1.GetLogSegmentRawRes{
		SegmentID:  item.SegmentID,
		StreamKind: item.StreamKind,
		Status:     item.Status,
		Content:    content,
		Truncated:  truncated,
	}, nil
}

type replayIndexRow struct {
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

type logSegmentRow struct {
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

type runEventRow struct {
	ID         string
	SequenceNo int
	EventType  string
	EventLevel string
	Summary    string
	CreatedAt  string
}

type runCheckpointRow struct {
	ID             string
	CheckpointType string
	CreatedAt      string
}

func getBrainRunBindingByRunID(ctx context.Context, projectID string, runID string) (*brainRunBindingViewRow, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(
		ctx,
		`SELECT id, project_id, task_id, brain_kind, brain_run_id, run_status, started_at, COALESCE(finished_at, ''), COALESCE(last_sync_at, ''), created_at, updated_at
FROM brain_run_bindings
WHERE project_id = ? AND brain_run_id = ?
LIMIT 1`,
		projectID,
		runID,
	)

	var binding brainRunBindingViewRow
	if err = row.Scan(
		&binding.ID,
		&binding.ProjectID,
		&binding.TaskID,
		&binding.BrainKind,
		&binding.RunID,
		&binding.RunStatus,
		&binding.StartedAt,
		&binding.FinishedAt,
		&binding.LastSyncAt,
		&binding.CreatedAt,
		&binding.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.New("brain run binding not found")
		}
		return nil, gerror.Wrap(err, "query brain run binding by run id failed")
	}
	return &binding, nil
}

type brainRunBindingViewRow struct {
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

func countRunEventsByBindingID(ctx context.Context, bindingID string) (int, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return 0, err
	}
	defer closeFn()

	var count int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(1) FROM run_event_index WHERE run_binding_id = ?`, bindingID).Scan(&count); err != nil {
		return 0, gerror.Wrap(err, "count run events failed")
	}
	return count, nil
}

func countReplayIndexByRunID(ctx context.Context, projectID string, runID string) (int, error) {
	return countByQuery(ctx, `SELECT COUNT(1) FROM workflow_replay_index WHERE project_id = ? AND run_id = ?`, projectID, runID)
}

func countLogSegmentsByRunID(ctx context.Context, projectID string, runID string) (int, error) {
	return countByQuery(ctx, `SELECT COUNT(1) FROM workflow_run_log_segments WHERE project_id = ? AND run_id = ?`, projectID, runID)
}

func countByQuery(ctx context.Context, query string, args ...any) (int, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return 0, err
	}
	defer closeFn()

	var count int
	if err = db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, gerror.Wrap(err, "count query failed")
	}
	return count, nil
}

func getLatestRunEvent(ctx context.Context, bindingID string) (*replayv1.ReplayEventRef, error) {
	if strings.TrimSpace(bindingID) == "" {
		return nil, nil
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT id, sequence_no, event_type, COALESCE(event_level, ''), summary, created_at
FROM run_event_index
WHERE run_binding_id = ?
ORDER BY sequence_no DESC, created_at DESC
LIMIT 1`, bindingID)
	return scanRunEventRef(row, "query latest run event failed")
}

func getRunEventByEventID(ctx context.Context, eventID string) (*replayv1.ReplayEventRef, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return nil, nil
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT id, sequence_no, event_type, COALESCE(event_level, ''), summary, created_at
FROM run_event_index
WHERE id = ?
LIMIT 1`, eventID)
	return scanRunEventRef(row, "query run event by id failed")
}

func scanRunEventRef(row *sql.Row, wrapMessage string) (*replayv1.ReplayEventRef, error) {
	var item runEventRow
	if err := row.Scan(&item.ID, &item.SequenceNo, &item.EventType, &item.EventLevel, &item.Summary, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, wrapMessage)
	}
	return &replayv1.ReplayEventRef{
		EventID:    item.ID,
		SequenceNo: item.SequenceNo,
		EventType:  item.EventType,
		EventLevel: item.EventLevel,
		Summary:    item.Summary,
		CreatedAt:  item.CreatedAt,
	}, nil
}

func getLatestRunCheckpoint(ctx context.Context, bindingID string) (*replayv1.ReplayCheckpointRef, error) {
	if strings.TrimSpace(bindingID) == "" {
		return nil, nil
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT id, checkpoint_type, created_at
FROM run_checkpoints
WHERE run_binding_id = ?
ORDER BY created_at DESC, id DESC
LIMIT 1`, bindingID)

	var item runCheckpointRow
	if err = row.Scan(&item.ID, &item.CheckpointType, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest run checkpoint failed")
	}
	return &replayv1.ReplayCheckpointRef{
		CheckpointID:   item.ID,
		CheckpointType: item.CheckpointType,
		CreatedAt:      item.CreatedAt,
	}, nil
}

func summarizeReplayArtifactStatus(ctx context.Context, projectID string, runID string) (replayv1.ReplayArtifactSummary, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return replayv1.ReplayArtifactSummary{}, err
	}
	defer closeFn()

	rows, err := db.QueryContext(ctx, `SELECT status, COUNT(1) FROM workflow_replay_index WHERE project_id = ? AND run_id = ? GROUP BY status`, projectID, runID)
	if err != nil {
		return replayv1.ReplayArtifactSummary{}, gerror.Wrap(err, "summarize replay artifact status failed")
	}
	defer rows.Close()

	summary := replayv1.ReplayArtifactSummary{}
	for rows.Next() {
		var (
			status string
			count  int
		)
		if err = rows.Scan(&status, &count); err != nil {
			return replayv1.ReplayArtifactSummary{}, gerror.Wrap(err, "scan replay artifact status failed")
		}
		switch status {
		case "artifact_missing":
			summary.Missing = count
		case "artifact_pruned":
			summary.Pruned = count
		default:
			summary.Available += count
		}
	}
	if err = rows.Err(); err != nil {
		return replayv1.ReplayArtifactSummary{}, gerror.Wrap(err, "iterate replay artifact status failed")
	}
	return summary, nil
}

func listReplayEntryPoints(ctx context.Context, projectID string, runID string, limit int) ([]replayv1.ReplayEntryPointItem, error) {
	rows, err := listReplayTimelineRows(ctx, projectID, runID, "", limit)
	if err != nil {
		return nil, err
	}
	items := make([]replayv1.ReplayEntryPointItem, 0, len(rows))
	for _, item := range rows {
		items = append(items, replayv1.ReplayEntryPointItem{
			DomainTaskID:   item.DomainTaskID,
			CompiledTaskID: item.CompiledTaskID,
			ReplayID:       item.ReplayID,
			ReplayType:     item.ReplayKind,
			Summary:        item.Summary,
			FilePath:       item.FilePath,
			CreatedAt:      item.CreatedAt,
		})
	}
	return items, nil
}

func listRunArtifactIssues(ctx context.Context, projectID string, runID string, limit int) ([]replayv1.ReplayArtifactIssue, error) {
	replayRows, err := listReplayTimelineRows(ctx, projectID, runID, "", replayTimelineMaxLimit)
	if err != nil {
		return nil, err
	}
	logRows, err := listLogSegmentRows(ctx, projectID, runID, logSegmentMaxLimit)
	if err != nil {
		return nil, err
	}

	issues := make([]replayv1.ReplayArtifactIssue, 0, limit)
	for _, row := range replayRows {
		if issue := buildReplayArtifactIssueFromReplay(row); issue != nil {
			issues = append(issues, *issue)
			if len(issues) >= limit {
				return issues, nil
			}
		}
	}
	for _, row := range logRows {
		if issue := buildReplayArtifactIssueFromLog(row); issue != nil {
			issues = append(issues, *issue)
			if len(issues) >= limit {
				return issues, nil
			}
		}
	}
	return issues, nil
}

func listReplayTimelineRows(ctx context.Context, projectID string, runID string, replayKind string, limit int) ([]replayIndexRow, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	query := `SELECT
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
		query += ` AND replay_kind = ?`
		args = append(args, replayKind)
	}
	query += ` ORDER BY seq_no ASC, created_at ASC, id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "query replay timeline rows failed")
	}
	defer rows.Close()

	items := make([]replayIndexRow, 0, limit)
	for rows.Next() {
		var item replayIndexRow
		if err = rows.Scan(
			&item.ID, &item.ReplayID, &item.ProjectID, &item.RunID,
			&item.DomainTaskID, &item.CompiledTaskID,
			&item.EventID, &item.TraceID, &item.SpanID,
			&item.ReplayKind, &item.SeqNo, &item.Title, &item.Summary,
			&item.FilePath, &item.FileExt, &item.MimeType, &item.FileSize, &item.SHA256,
			&item.SourceObjectKind, &item.SourceObjectID, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan replay timeline row failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate replay timeline rows failed")
	}
	return items, nil
}

func getReplayIndexItem(ctx context.Context, projectID string, runID string, replayID string) (*replayIndexRow, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT
  id, replay_id, project_id, run_id,
  COALESCE(domain_task_id, ''), COALESCE(compiled_task_id, ''),
  COALESCE(event_id, ''), COALESCE(trace_id, ''), COALESCE(span_id, ''),
  replay_kind, seq_no, title, COALESCE(summary, ''),
  file_path, COALESCE(file_ext, ''), COALESCE(mime_type, ''), COALESCE(file_size, 0), COALESCE(sha256, ''),
  COALESCE(source_object_kind, ''), COALESCE(source_object_id, ''), status, created_at, updated_at
FROM workflow_replay_index
WHERE project_id = ? AND run_id = ? AND replay_id = ?
LIMIT 1`, projectID, runID, replayID)

	var item replayIndexRow
	if err = row.Scan(
		&item.ID, &item.ReplayID, &item.ProjectID, &item.RunID,
		&item.DomainTaskID, &item.CompiledTaskID,
		&item.EventID, &item.TraceID, &item.SpanID,
		&item.ReplayKind, &item.SeqNo, &item.Title, &item.Summary,
		&item.FilePath, &item.FileExt, &item.MimeType, &item.FileSize, &item.SHA256,
		&item.SourceObjectKind, &item.SourceObjectID, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.New("replay item not found")
		}
		return nil, gerror.Wrap(err, "query replay item failed")
	}
	return &item, nil
}

func listLogSegmentRows(ctx context.Context, projectID string, runID string, limit int) ([]logSegmentRow, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(ctx, `SELECT
  id, project_id, run_id, segment_id, stream_kind, seq_no, file_path,
  COALESCE(file_size, 0), COALESCE(sha256, ''), COALESCE(started_at, ''), COALESCE(ended_at, ''), status, created_at
FROM workflow_run_log_segments
WHERE project_id = ? AND run_id = ?
ORDER BY seq_no ASC, created_at ASC, id ASC
LIMIT ?`, projectID, runID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query log segment rows failed")
	}
	defer rows.Close()

	items := make([]logSegmentRow, 0, limit)
	for rows.Next() {
		var item logSegmentRow
		if err = rows.Scan(
			&item.ID, &item.ProjectID, &item.RunID, &item.SegmentID, &item.StreamKind, &item.SeqNo, &item.FilePath,
			&item.FileSize, &item.SHA256, &item.StartedAt, &item.EndedAt, &item.Status, &item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan log segment row failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate log segment rows failed")
	}
	return items, nil
}

func getLogSegmentItem(ctx context.Context, projectID string, runID string, segmentID string) (*logSegmentRow, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx, `SELECT
  id, project_id, run_id, segment_id, stream_kind, seq_no, file_path,
  COALESCE(file_size, 0), COALESCE(sha256, ''), COALESCE(started_at, ''), COALESCE(ended_at, ''), status, created_at
FROM workflow_run_log_segments
WHERE project_id = ? AND run_id = ? AND segment_id = ?
LIMIT 1`, projectID, runID, segmentID)

	var item logSegmentRow
	if err = row.Scan(
		&item.ID, &item.ProjectID, &item.RunID, &item.SegmentID, &item.StreamKind, &item.SeqNo, &item.FilePath,
		&item.FileSize, &item.SHA256, &item.StartedAt, &item.EndedAt, &item.Status, &item.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.New("log segment not found")
		}
		return nil, gerror.Wrap(err, "query log segment failed")
	}
	return &item, nil
}

func mapLogSegments(rows []logSegmentRow) []replayv1.LogSegmentListItem {
	items := make([]replayv1.LogSegmentListItem, 0, len(rows))
	for _, item := range rows {
		items = append(items, replayv1.LogSegmentListItem{
			SegmentID:  item.SegmentID,
			StreamKind: item.StreamKind,
			SeqNo:      item.SeqNo,
			StartedAt:  item.StartedAt,
			EndedAt:    item.EndedAt,
			Status:     item.Status,
			Size:       item.FileSize,
			RawTarget:  item.FilePath,
		})
	}
	return items
}

func buildReplayArtifactIssueFromReplay(item replayIndexRow) *replayv1.ReplayArtifactIssue {
	return buildArtifactIssue("replay", item.ReplayID, item.Status, item.FilePath, item.Title)
}

func buildReplayArtifactIssueFromLog(item logSegmentRow) *replayv1.ReplayArtifactIssue {
	return buildArtifactIssue("log_segment", item.SegmentID, item.Status, item.FilePath, item.StreamKind)
}

func buildArtifactIssue(source string, id string, status string, filePath string, summary string) *replayv1.ReplayArtifactIssue {
	status = strings.TrimSpace(status)
	filePath = strings.TrimSpace(filePath)
	switch {
	case status == "artifact_missing":
		return &replayv1.ReplayArtifactIssue{
			Kind:              "missing_artifact",
			Source:            source,
			ID:                id,
			Status:            status,
			FilePath:          filePath,
			Summary:           summary,
			RecommendedAction: "refresh_replay_artifacts_or_restore_raw_file",
		}
	case status == "artifact_pruned":
		return &replayv1.ReplayArtifactIssue{
			Kind:              "pruned_artifact",
			Source:            source,
			ID:                id,
			Status:            status,
			FilePath:          filePath,
			Summary:           summary,
			RecommendedAction: "rerun_or_restore_pruned_artifact",
		}
	case status == "available" && filePath != "" && fileIsMissing(filePath):
		return &replayv1.ReplayArtifactIssue{
			Kind:              "stale_index",
			Source:            source,
			ID:                id,
			Status:            status,
			FilePath:          filePath,
			Summary:           summary,
			RecommendedAction: "refresh_replay_artifact_index",
		}
	default:
		return nil
	}
}

func fileIsMissing(filePath string) bool {
	if strings.TrimSpace(filePath) == "" {
		return false
	}
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}

func buildReplayDiagnosticHints(
	artifactSummary replayv1.ReplayArtifactSummary,
	issues []replayv1.ReplayArtifactIssue,
	latestEvent *replayv1.ReplayEventRef,
	latestCheckpoint *replayv1.ReplayCheckpointRef,
) []replayv1.ReplayDiagnosticHint {
	hints := make([]replayv1.ReplayDiagnosticHint, 0, 4)
	if artifactSummary.Missing > 0 {
		hints = append(hints, replayv1.ReplayDiagnosticHint{
			Code:              "missing_artifact",
			Severity:          "warning",
			Summary:           "Replay index contains missing artifact entries.",
			RecommendedAction: "refresh_replay_artifacts_or_restore_raw_file",
		})
	}
	if artifactSummary.Pruned > 0 {
		hints = append(hints, replayv1.ReplayDiagnosticHint{
			Code:              "pruned_artifact",
			Severity:          "info",
			Summary:           "Some replay artifacts were pruned and raw preview is unavailable.",
			RecommendedAction: "rerun_or_restore_pruned_artifact",
		})
	}
	if hasArtifactIssueKind(issues, "stale_index") {
		hints = append(hints, replayv1.ReplayDiagnosticHint{
			Code:              "stale_index",
			Severity:          "warning",
			Summary:           "Replay index points to a file that no longer exists.",
			RecommendedAction: "refresh_replay_artifact_index",
		})
	}
	if latestEvent == nil && latestCheckpoint == nil {
		hints = append(hints, replayv1.ReplayDiagnosticHint{
			Code:              "missing_runtime_markers",
			Severity:          "info",
			Summary:           "No latest event or checkpoint is available for this run.",
			RecommendedAction: "sync_run_events_and_checkpoints",
		})
	}
	return hints
}

func buildItemDiagnosticHints(issue *replayv1.ReplayArtifactIssue) []replayv1.ReplayDiagnosticHint {
	if issue == nil {
		return nil
	}
	return []replayv1.ReplayDiagnosticHint{
		{
			Code:              issue.Kind,
			Severity:          "warning",
			Summary:           issue.Summary,
			RecommendedAction: issue.RecommendedAction,
		},
	}
}

func hasArtifactIssueKind(issues []replayv1.ReplayArtifactIssue, kind string) bool {
	for _, issue := range issues {
		if issue.Kind == kind {
			return true
		}
	}
	return false
}

func loadRawFileSnippet(filePath string, status string, limit int) (string, bool, error) {
	status = strings.TrimSpace(status)
	if status == "artifact_missing" || status == "artifact_pruned" {
		return "", false, nil
	}
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return "", false, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, gerror.Wrap(err, "open raw file failed")
	}
	defer file.Close()

	buffer := make([]byte, limit+1)
	readN, readErr := file.Read(buffer)
	if readErr != nil && readErr.Error() != "EOF" && readErr != sql.ErrNoRows {
		// os.File returns io.EOF, not sql.ErrNoRows; keep generic handling lightweight.
		if readErr.Error() != "EOF" {
			return "", false, gerror.Wrap(readErr, "read raw file failed")
		}
	}

	truncated := readN > limit
	if truncated {
		readN = limit
	}
	return string(buffer[:readN]), truncated, nil
}
