package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type GetRetrospectiveReq struct {
	g.Meta `path:"/api/v3/retrospectives/{id}" tags:"Retrospectives" method:"get" summary:"Get retrospective by ID"`
	Id     string `json:"id" in:"path" v:"required"`
}

type GetRetrospectiveRes struct {
	Retrospective *entity.ProjectRetrospectives `json:"retrospective"`
}

type GetProjectRetrospectiveReq struct {
	g.Meta    `path:"/api/v3/retrospectives/project/{projectId}" tags:"Retrospectives" method:"get" summary:"Get retrospective by project ID"`
	ProjectId string `json:"projectId" in:"path" v:"required"`
}

type GetProjectRetrospectiveRes struct {
	Retrospective *entity.ProjectRetrospectives `json:"retrospective"`
}
