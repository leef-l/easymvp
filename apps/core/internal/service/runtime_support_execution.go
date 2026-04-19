package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const executionRecentBindingLimit = 8
const executionReplayDefaultLimit = 10
const executionLogDefaultLimit = 10

func (s *sRuntime) GetExecutionView(ctx context.Context, projectID string, bindingID string, eventLimit int, replayLimit int, logLimit int) (*runtimev1.ExecutionViewRes, error) {
	projectID = strings.TrimSpace(projectID)
	bindingID = strings.TrimSpace(bindingID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	res := &runtimev1.ExecutionViewRes{
		RuntimeHealth: runtimev1.HealthzRes{
			Status: "unavailable",
		},
		RecentBindings: make([]runtimev1.RunBindingView, 0),
		ReplayTimeline: make([]runtimev1.ExecutionReplayItem, 0),
		LogSegments:    make([]runtimev1.ExecutionLogSegment, 0),
	}

	baseURL, baseURLErr := runtimeBaseURL(ctx)
	if baseURLErr == nil {
		res.RuntimeHealth.BaseURL = baseURL
	}
	health, err := s.Healthz(ctx)
	if err != nil {
		res.RuntimeError = err.Error()
	} else {
		res.RuntimeHealth = *health
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	var selectedBinding *entity.BrainRunBindings
	if bindingID != "" {
		selectedBinding, err = getBrainRunBindingByID(ctx, bindingID)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(selectedBinding.ProjectId) != projectID {
			return nil, gerror.New("binding does not belong to project")
		}
	}

	bindings, err := listProjectBrainRunBindings(ctx, db, projectID, executionRecentBindingLimit)
	if err != nil {
		return nil, err
	}
	if selectedBinding != nil {
		bindings = prioritizeExecutionBinding(bindings, *selectedBinding, executionRecentBindingLimit)
	}
	if selectedBinding == nil && len(bindings) > 0 {
		selectedBinding = &bindings[0]
	}
	for _, item := range bindings {
		binding := item
		res.RecentBindings = append(res.RecentBindings, mapRunBindingView(&binding))
	}
	if selectedBinding == nil {
		return res, nil
	}

	detail, err := s.GetRunBindingDetail(ctx, selectedBinding.Id, normalizeExecutionViewLimit(eventLimit, runtimeRunEventDefaultLimit, runtimeRunEventMaxLimit))
	if err != nil {
		return nil, err
	}
	res.LatestBinding = detail

	if strings.TrimSpace(selectedBinding.BrainRunId) == "" {
		return res, nil
	}

	replaySummary, replayErr := Replay().GetReplaySummary(ctx, projectID, selectedBinding.BrainRunId)
	if replayErr != nil {
		res.ReplayError = replayErr.Error()
		return res, nil
	}
	res.ReplaySummary = &runtimev1.ExecutionReplaySummary{
		RunID:           replaySummary.RunID,
		ProjectID:       replaySummary.ProjectID,
		BrainKind:       replaySummary.BrainKind,
		Status:          replaySummary.Status,
		StartedAt:       replaySummary.StartedAt,
		EndedAt:         replaySummary.EndedAt,
		EventCount:      replaySummary.EventCount,
		ReplayCount:     replaySummary.ReplayCount,
		LogSegmentCount: replaySummary.LogSegmentCount,
		ArtifactReady:   replaySummary.ArtifactStatusSummary.Available,
		ArtifactMissing: replaySummary.ArtifactStatusSummary.Missing,
		ArtifactPruned:  replaySummary.ArtifactStatusSummary.Pruned,
	}

	timeline, replayErr := Replay().GetReplayTimeline(ctx, projectID, selectedBinding.BrainRunId, normalizeExecutionViewLimit(replayLimit, executionReplayDefaultLimit, replayTimelineMaxLimit), "")
	if replayErr == nil {
		for _, item := range timeline.Items {
			res.ReplayTimeline = append(res.ReplayTimeline, runtimev1.ExecutionReplayItem{
				ReplayID:         item.ReplayID,
				ReplayType:       item.ReplayType,
				Title:            item.Title,
				Summary:          item.Summary,
				Status:           item.Status,
				PreviewAvailable: item.PreviewAvailable,
				RawTarget:        item.RawTarget,
				CreatedAt:        item.CreatedAt,
			})
		}
	}

	segments, replayErr := Replay().ListLogSegments(ctx, projectID, selectedBinding.BrainRunId, normalizeExecutionViewLimit(logLimit, executionLogDefaultLimit, logSegmentMaxLimit))
	if replayErr == nil {
		for _, item := range segments.Segments {
			res.LogSegments = append(res.LogSegments, runtimev1.ExecutionLogSegment{
				SegmentID:  item.SegmentID,
				StreamKind: item.StreamKind,
				SeqNo:      item.SeqNo,
				Status:     item.Status,
				Size:       item.Size,
				RawTarget:  item.RawTarget,
			})
		}
	}

	return res, nil
}

func normalizeExecutionViewLimit(limit int, defaultLimit int, maxLimit int) int {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit
}

func prioritizeExecutionBinding(bindings []entity.BrainRunBindings, selected entity.BrainRunBindings, limit int) []entity.BrainRunBindings {
	if limit <= 0 {
		limit = len(bindings)
	}
	result := make([]entity.BrainRunBindings, 0, limit)
	result = append(result, selected)
	for _, item := range bindings {
		if item.Id == selected.Id {
			continue
		}
		result = append(result, item)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func listProjectRunBindingsByProjectID(ctx context.Context, db *sql.DB, projectID string, limit int) ([]runtimev1.RunBindingView, error) {
	items, err := listProjectBrainRunBindings(ctx, db, projectID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]runtimev1.RunBindingView, 0, len(items))
	for _, item := range items {
		binding := item
		result = append(result, mapRunBindingView(&binding))
	}
	return result, nil
}
