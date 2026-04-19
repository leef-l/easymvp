package service

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
)

type IReplay interface {
	GetReplaySummary(ctx context.Context, projectID string, runID string) (*replayv1.GetReplaySummaryRes, error)
	GetReplayTimeline(ctx context.Context, projectID string, runID string, limit int, replayType string) (*replayv1.GetReplayTimelineRes, error)
	GetReplayDetail(ctx context.Context, projectID string, runID string, replayID string) (*replayv1.GetReplayDetailRes, error)
	GetReplayRaw(ctx context.Context, projectID string, runID string, replayID string, limit int) (*replayv1.GetReplayRawRes, error)
	ListLogSegments(ctx context.Context, projectID string, runID string, limit int) (*replayv1.ListLogSegmentsRes, error)
	GetLogSegmentRaw(ctx context.Context, projectID string, runID string, segmentID string, limit int) (*replayv1.GetLogSegmentRawRes, error)
}

var localReplay IReplay = (*sReplay)(nil)

type sReplay struct{}

func Replay() IReplay {
	if localReplay == nil {
		localReplay = &sReplay{}
	}
	return localReplay
}
