package retrospectives

import (
	api "github.com/leef-l/easymvp/apps/core/api/retrospectives"
)

func NewV1() api.IRetrospectivesV1 {
	return &ControllerV1{}
}
