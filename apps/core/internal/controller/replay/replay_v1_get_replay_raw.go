package replay

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetReplayRaw(ctx context.Context, req *replayv1.GetReplayRawReq) (res *replayv1.GetReplayRawRes, err error) {
	return service.Replay().GetReplayRaw(ctx, req.ProjectID, req.RunID, req.ReplayID, req.Limit)
}
