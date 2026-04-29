package evidence

import (
	"context"

	evidencev1 "github.com/leef-l/easymvp/apps/core/api/evidence/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) CollectProjectEvidence(ctx context.Context, req *evidencev1.CollectEvidenceReq) (res *evidencev1.CollectEvidenceRes, err error) {
	return service.Evidence().CollectProjectEvidence(ctx, req.ProjectID)
}
