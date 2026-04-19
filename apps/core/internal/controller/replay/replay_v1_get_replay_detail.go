package replay

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetReplayDetail(ctx context.Context, req *replayv1.GetReplayDetailReq) (res *replayv1.GetReplayDetailRes, err error) {
	return service.Replay().GetReplayDetail(ctx, req.ProjectID, req.RunID, req.ReplayID)
}
