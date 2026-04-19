package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type HomeViewReq struct {
	g.Meta `path:"/api/v3/workspace/home-view" tags:"Workspace" method:"get" summary:"Workspace home view"`
}

type HomeViewRes struct {
	Overview         HomeOverview        `json:"overview"`
	Summary          HomeSummary         `json:"summary"`
	ActiveProjects   []ProjectCard       `json:"active_projects"`
	NeedAttention    []NeedAttentionItem `json:"need_attention"`
	RecentActivity   []LiveActivityItem  `json:"recent_activity"`
	ReleaseReadiness []ReleaseReadiness  `json:"release_readiness"`
}
