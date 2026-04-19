package replay

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetReplayTimeline(ctx context.Context, req *replayv1.GetReplayTimelineReq) (res *replayv1.GetReplayTimelineRes, err error) {
	return service.Replay().GetReplayTimeline(ctx, req.ProjectID, req.RunID, req.Limit, req.ReplayType)
}
