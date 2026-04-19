package plan

import api "github.com/leef-l/easymvp/apps/core/api/plan"

func NewV1() api.IPlanV1 {
	return &ControllerV1{}
}
