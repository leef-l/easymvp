package workspace

import (
	api "github.com/leef-l/easymvp/apps/core/api/workspace"
)

func NewV1() api.IWorkspaceV1 {
	return &ControllerV1{}
}
