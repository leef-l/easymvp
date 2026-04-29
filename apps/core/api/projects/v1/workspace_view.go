package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
)

type ProjectWorkspaceViewReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/workspace-view" tags:"Projects" method:"get" summary:"Project workspace view"`
	Id string `json:"id" in:"path" v:"required"`
}

type ProjectWorkspaceViewRes struct {
	Overview             WorkspaceOverview                   `json:"overview"`
	ProjectSnapshot      ProjectSnapshot                     `json:"project_snapshot"`
	StageProgress        []StageProgressItem                 `json:"stage_progress"`
	LiveActivity         []LiveActivityItem                  `json:"live_activity"`
	ActionInbox          []ActionInboxItem                   `json:"action_inbox"`
	AcceptanceCoverage   AcceptanceCoverage                  `json:"acceptance_coverage"`
	WorkspaceExplanation WorkspaceExplanation                `json:"workspace_explanation"`
	VerificationResult   acceptancev1.VerificationResultView `json:"verification_result"`
	CompletionVerdict    acceptancev1.CompletionVerdictView  `json:"completion_verdict"`
	RuntimeEscalation    acceptancev1.RuntimeEscalationView  `json:"runtime_escalation"`
	FaultSummary         acceptancev1.FaultSummaryView       `json:"fault_summary"`
	RepairPlanDraft      acceptancev1.RepairPlanDraftSummary `json:"repair_plan_draft"`
	HealthMetrics        ProjectHealthMetrics                `json:"health_metrics"`
}

// ProjectHealthMetrics provides near-term run statistics for project health.
type ProjectHealthMetrics struct {
	RecentRunCount     int     `json:"recent_run_count"`
	SuccessRate        float64 `json:"success_rate"`
	AvgLatencyMs       int     `json:"avg_latency_ms"`
	TopFailureMode     string  `json:"top_failure_mode,omitempty"`
	TopFailureCount    int     `json:"top_failure_count"`
	LastUpdated        string  `json:"last_updated,omitempty"`
}
