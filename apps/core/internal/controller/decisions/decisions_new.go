package decisions

import api "github.com/leef-l/easymvp/apps/core/api/decisions"

func NewV1() api.IDecisionsV1 {
	return &ControllerV1{}
}
