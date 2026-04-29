package v1

import "github.com/gogf/gf/v2/frame/g"

type ProjectEventsStreamReq struct {
	g.Meta      `path:"/api/v3/workspace/projects/{id}/events" tags:"Workspace" method:"get" summary:"Workspace project events stream"`
	Id          string `json:"id" in:"path" v:"required"`
	Limit       int    `json:"limit"`
	LastEventID string `json:"last_event_id"`
}

type ProjectEventsStreamRes struct{}
