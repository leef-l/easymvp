package replay

import api "github.com/leef-l/easymvp/apps/core/api/replay"

func NewV1() api.IReplayV1 {
	return &ControllerV1{}
}
