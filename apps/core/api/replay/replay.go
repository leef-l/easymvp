package replay

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/replay/v1"
)

type IReplayV1 interface {
	GetReplaySummary(ctx context.Context, req *v1.GetReplaySummaryReq) (res *v1.GetReplaySummaryRes, err error)
	GetReplayTimeline(ctx context.Context, req *v1.GetReplayTimelineReq) (res *v1.GetReplayTimelineRes, err error)
	GetReplayDetail(ctx context.Context, req *v1.GetReplayDetailReq) (res *v1.GetReplayDetailRes, err error)
	GetReplayRaw(ctx context.Context, req *v1.GetReplayRawReq) (res *v1.GetReplayRawRes, err error)
	ListLogSegments(ctx context.Context, req *v1.ListLogSegmentsReq) (res *v1.ListLogSegmentsRes, err error)
	GetLogSegmentRaw(ctx context.Context, req *v1.GetLogSegmentRawReq) (res *v1.GetLogSegmentRawRes, err error)
}
