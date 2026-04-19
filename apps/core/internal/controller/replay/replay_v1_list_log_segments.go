package replay

import (
	"context"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ListLogSegments(ctx context.Context, req *replayv1.ListLogSegmentsReq) (res *replayv1.ListLogSegmentsRes, err error) {
	return service.Replay().ListLogSegments(ctx, req.ProjectID, req.RunID, req.Limit)
}
