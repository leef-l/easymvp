package evidence

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/evidence/v1"
)

type IEvidenceV1 interface {
	CollectProjectEvidence(ctx context.Context, req *v1.CollectEvidenceReq) (res *v1.CollectEvidenceRes, err error)
}
