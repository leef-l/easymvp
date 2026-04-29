package evidence

import (
	api "github.com/leef-l/easymvp/apps/core/api/evidence"
)

func NewV1() api.IEvidenceV1 {
	return &ControllerV1{}
}
