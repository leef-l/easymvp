package designs

import (
	api "github.com/leef-l/easymvp/apps/core/api/designs"
)

func NewV1() api.IDesignsV1 {
	return &ControllerV1{}
}
