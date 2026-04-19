package v1

import "github.com/gogf/gf/v2/frame/g"

type AuditLogItem struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	EventType   string `json:"event_type"`
	ActorKind   string `json:"actor_kind"`
	Summary     string `json:"summary"`
	PayloadJSON string `json:"payload_json,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type ListProjectAuditLogsReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/audit-logs" method:"get" tags:"Audit" summary:"List project audit logs"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type ListProjectAuditLogsRes struct {
	Items       []AuditLogItem `json:"items"`
	RefreshHint string         `json:"refresh_hint"`
}
