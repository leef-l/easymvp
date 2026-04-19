package system

import (
	"github.com/leef-l/easymvp/apps/core/api/system"
)

func NewV1() system.ISystemV1 {
	return &ControllerV1{}
}
