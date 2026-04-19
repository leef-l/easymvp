package runtime

import api "github.com/leef-l/easymvp/apps/core/api/runtime"

func NewV1() api.IRuntimeV1 {
	return &ControllerV1{}
}
