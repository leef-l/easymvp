package requirements

import (
	api "github.com/leef-l/easymvp/apps/core/api/requirements"
)

func NewV1() api.IRequirementsV1 {
	return &ControllerV1{}
}
