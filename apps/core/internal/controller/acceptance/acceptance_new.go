package acceptance

import (
	api "github.com/leef-l/easymvp/apps/core/api/acceptance"
)

func NewV1() api.IAcceptanceV1 {
	return &ControllerV1{}
}
