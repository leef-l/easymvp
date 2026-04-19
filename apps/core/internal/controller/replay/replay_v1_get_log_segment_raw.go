package replay

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetLogSegmentRaw(ctx context.Context, req *replayv1.GetLogSegmentRawReq) (res *replayv1.GetLogSegmentRawRes, err error) {
	return service.Replay().GetLogSegmentRaw(ctx, req.ProjectID, req.RunID, req.SegmentID, req.Limit)
}
