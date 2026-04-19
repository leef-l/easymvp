package audit

import api "github.com/leef-l/easymvp/apps/core/api/audit"

func NewV1() api.IAuditV1 {
	return &ControllerV1{}
}
