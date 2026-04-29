package service

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/repo"
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
	res, err := repo.GetBrainRunBindingByRunID(ctx, projectID, runID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return &brainRunBindingViewRow{
		ID:         res.ID,
		ProjectID:  res.ProjectID,
		TaskID:     res.TaskID,
		BrainKind:  res.BrainKind,
		RunID:      res.RunID,
		RunStatus:  res.RunStatus,
		StartedAt:  res.StartedAt,
		FinishedAt: res.FinishedAt,
		LastSyncAt: res.LastSyncAt,
		CreatedAt:  res.CreatedAt,
		UpdatedAt:  res.UpdatedAt,
	}, nil
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
	return repo.CountRunEventsByBindingID(ctx, bindingID)
}

func countReplayIndexByRunID(ctx context.Context, projectID string, runID string) (int, error) {
	return repo.CountReplayIndexByRunID(ctx, projectID, runID)
}

func countLogSegmentsByRunID(ctx context.Context, projectID string, runID string) (int, error) {
	return repo.CountLogSegmentsByRunID(ctx, projectID, runID)
}

func getLatestRunEvent(ctx context.Context, bindingID string) (*replayv1.ReplayEventRef, error) {
	return repo.GetLatestRunEvent(ctx, bindingID)
}

func getRunEventByEventID(ctx context.Context, eventID string) (*replayv1.ReplayEventRef, error) {
	return repo.GetRunEventByEventID(ctx, eventID)
}

func getLatestRunCheckpoint(ctx context.Context, bindingID string) (*replayv1.ReplayCheckpointRef, error) {
	return repo.GetLatestRunCheckpoint(ctx, bindingID)
}

func summarizeReplayArtifactStatus(ctx context.Context, projectID string, runID string) (replayv1.ReplayArtifactSummary, error) {
	return repo.SummarizeReplayArtifactStatus(ctx, projectID, runID)
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
	rows, err := repo.ListReplayTimelineRows(ctx, projectID, runID, replayKind, limit)
	if err != nil {
		return nil, err
	}
	result := make([]replayIndexRow, len(rows))
	for i, r := range rows {
		result[i] = replayIndexRow{
			ID:               r.ID,
			ReplayID:         r.ReplayID,
			ProjectID:        r.ProjectID,
			RunID:            r.RunID,
			DomainTaskID:     r.DomainTaskID,
			CompiledTaskID:   r.CompiledTaskID,
			EventID:          r.EventID,
			TraceID:          r.TraceID,
			SpanID:           r.SpanID,
			ReplayKind:       r.ReplayKind,
			SeqNo:            r.SeqNo,
			Title:            r.Title,
			Summary:          r.Summary,
			FilePath:         r.FilePath,
			FileExt:          r.FileExt,
			MimeType:         r.MimeType,
			FileSize:         r.FileSize,
			SHA256:           r.SHA256,
			SourceObjectKind: r.SourceObjectKind,
			SourceObjectID:   r.SourceObjectID,
			Status:           r.Status,
			CreatedAt:        r.CreatedAt,
			UpdatedAt:        r.UpdatedAt,
		}
	}
	return result, nil
}

func getReplayIndexItem(ctx context.Context, projectID string, runID string, replayID string) (*replayIndexRow, error) {
	res, err := repo.GetReplayIndexItem(ctx, projectID, runID, replayID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return &replayIndexRow{
		ID:               res.ID,
		ReplayID:         res.ReplayID,
		ProjectID:        res.ProjectID,
		RunID:            res.RunID,
		DomainTaskID:     res.DomainTaskID,
		CompiledTaskID:   res.CompiledTaskID,
		EventID:          res.EventID,
		TraceID:          res.TraceID,
		SpanID:           res.SpanID,
		ReplayKind:       res.ReplayKind,
		SeqNo:            res.SeqNo,
		Title:            res.Title,
		Summary:          res.Summary,
		FilePath:         res.FilePath,
		FileExt:          res.FileExt,
		MimeType:         res.MimeType,
		FileSize:         res.FileSize,
		SHA256:           res.SHA256,
		SourceObjectKind: res.SourceObjectKind,
		SourceObjectID:   res.SourceObjectID,
		Status:           res.Status,
		CreatedAt:        res.CreatedAt,
		UpdatedAt:        res.UpdatedAt,
	}, nil
}

func listLogSegmentRows(ctx context.Context, projectID string, runID string, limit int) ([]logSegmentRow, error) {
	rows, err := repo.ListLogSegmentRows(ctx, projectID, runID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]logSegmentRow, len(rows))
	for i, r := range rows {
		result[i] = logSegmentRow{
			ID:         r.ID,
			ProjectID:  r.ProjectID,
			RunID:      r.RunID,
			SegmentID:  r.SegmentID,
			StreamKind: r.StreamKind,
			SeqNo:      r.SeqNo,
			FilePath:   r.FilePath,
			FileSize:   r.FileSize,
			SHA256:     r.SHA256,
			StartedAt:  r.StartedAt,
			EndedAt:    r.EndedAt,
			Status:     r.Status,
			CreatedAt:  r.CreatedAt,
		}
	}
	return result, nil
}

func getLogSegmentItem(ctx context.Context, projectID string, runID string, segmentID string) (*logSegmentRow, error) {
	res, err := repo.GetLogSegmentItem(ctx, projectID, runID, segmentID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return &logSegmentRow{
		ID:         res.ID,
		ProjectID:  res.ProjectID,
		RunID:      res.RunID,
		SegmentID:  res.SegmentID,
		StreamKind: res.StreamKind,
		SeqNo:      res.SeqNo,
		FilePath:   res.FilePath,
		FileSize:   res.FileSize,
		SHA256:     res.SHA256,
		StartedAt:  res.StartedAt,
		EndedAt:    res.EndedAt,
		Status:     res.Status,
		CreatedAt:  res.CreatedAt,
	}, nil
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
