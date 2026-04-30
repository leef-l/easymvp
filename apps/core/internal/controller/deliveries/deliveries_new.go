package deliveries

import (
	api "github.com/leef-l/easymvp/apps/core/api/deliveries"
)

func NewV1() api.IDeliveriesV1 {
	return &ControllerV1{}
}
