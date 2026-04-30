package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type GetDesignReq struct {
	g.Meta `path:"/api/v3/designs/{id}" tags:"Designs" method:"get" summary:"Get solution design by ID"`
	Id     string `json:"id" in:"path" v:"required"`
}

type GetDesignRes struct {
	Design *entity.SolutionDesigns `json:"design"`
}

type GetProjectLatestDesignReq struct {
	g.Meta    `path:"/api/v3/designs/project/{projectId}/latest" tags:"Designs" method:"get" summary:"Get latest solution design for project"`
	ProjectId string `json:"projectId" in:"path" v:"required"`
}

type GetProjectLatestDesignRes struct {
	Design *entity.SolutionDesigns `json:"design"`
}
