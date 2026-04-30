package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type AcceptanceViewReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/acceptance-view" method:"get" tags:"Acceptance" summary:"Acceptance view"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type AcceptanceViewRes struct {
	Overview           AcceptanceOverview     `json:"overview"`
	AcceptanceRun      AcceptanceRunView      `json:"acceptance_run"`
	CoverageMatrix     []CoverageItem         `json:"coverage_matrix"`
	Issues             []AcceptanceIssue      `json:"issues"`
	EvidenceCards      []EvidenceCard         `json:"evidence_cards"`
	ReleaseGate        ReleaseGateView        `json:"release_gate"`
	VerificationResult VerificationResultView `json:"verification_result"`
	CompletionVerdict  CompletionVerdictView  `json:"completion_verdict"`
	RuntimeEscalation  RuntimeEscalationView  `json:"runtime_escalation"`
	FaultSummary       FaultSummaryView       `json:"fault_summary"`
	RepairPlanDraft    RepairPlanDraftSummary `json:"repair_plan_draft"`
	ContractGap        ContractGapView        `json:"contract_gap"`
}

type StartAcceptanceReq struct {
	g.Meta         `path:"/api/v3/projects/{id}/acceptance-runs" method:"post" tags:"Acceptance" summary:"Start acceptance"`
	ProjectID      string `json:"id" in:"path" v:"required"`
	TaskID         string `json:"task_id"`
	ProfileVersion string `json:"profile_version"`
	Mode           string `json:"mode"`
}

type StartAcceptanceRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}

type AdjudicateAcceptanceReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/acceptance-runs/adjudicate" method:"post" tags:"Acceptance" summary:"Adjudicate latest acceptance run"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type AdjudicateAcceptanceRes struct {
	CommandID   string `json:"command_id"`
	Accepted    bool   `json:"accepted"`
	ResourceID  string `json:"resource_id"`
	NextAction  string `json:"next_action"`
	FinalStatus string `json:"final_status"`
}

type ApplyManualReleaseReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/acceptance-runs/manual-release" method:"post" tags:"Acceptance" summary:"Apply manual release approval"`
	ProjectID string `json:"id" in:"path" v:"required"`
	Comment   string `json:"comment"`
}

type ApplyManualReleaseRes struct {
	CommandID   string `json:"command_id"`
	Accepted    bool   `json:"accepted"`
	ResourceID  string `json:"resource_id"`
	NextAction  string `json:"next_action"`
	FinalStatus string `json:"final_status"`
}

type RefreshAcceptanceProfilesReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/acceptance-profiles/refresh" method:"post" tags:"Acceptance" summary:"Refresh acceptance profiles"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type RefreshAcceptanceProfilesRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
