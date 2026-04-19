package projects

import (
	api "github.com/leef-l/easymvp/apps/core/api/projects"
)

func NewV1() api.IProjectsV1 {
	return &ControllerV1{}
}
