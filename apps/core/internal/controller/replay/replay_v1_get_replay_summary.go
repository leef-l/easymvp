package replay

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetReplaySummary(ctx context.Context, req *replayv1.GetReplaySummaryReq) (res *replayv1.GetReplaySummaryRes, err error) {
	return service.Replay().GetReplaySummary(ctx, req.ProjectID, req.RunID)
}
