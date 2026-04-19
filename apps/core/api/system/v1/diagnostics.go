package v1

import "github.com/gogf/gf/v2/frame/g"

type ProjectDiagnosticItem struct {
	ID         string `json:"id"`
	Scope      string `json:"scope"`
	Severity   string `json:"severity"`
	ErrorCode  string `json:"error_code"`
	Summary    string `json:"summary"`
	DetailJSON string `json:"detail_json,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	BindingID  string `json:"binding_id,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type ListProjectDiagnosticsReq struct {
	g.Meta    `path:"/api/v3/projects/{project_id}/diagnostic-records" tags:"System" method:"get" summary:"List project diagnostic records"`
	ProjectID string `json:"project_id" in:"path" v:"required"`
	Limit     int    `json:"limit" in:"query"`
}

type ListProjectDiagnosticsRes struct {
	Items       []ProjectDiagnosticItem `json:"items"`
	RefreshHint string                  `json:"refresh_hint"`
}
