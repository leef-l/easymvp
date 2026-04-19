package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

type RepairDraftReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/repair-draft" tags:"Plan" method:"get" summary:"Repair draft view"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type CreateRepairDraftReq struct {
	g.Meta                `path:"/api/v3/projects/{id}/repair-draft" tags:"Plan" method:"post" summary:"Create repair draft"`
	ProjectID             string                       `json:"id" in:"path" v:"required"`
	FailedTaskContextJSON string                       `json:"failed_task_context_json" v:"required"`
	FailureReasonJSON     string                       `json:"failure_reason_json" v:"required"`
	OriginalContractsJSON string                       `json:"original_contracts_json" v:"required"`
	RuntimeSummaryJSON    string                       `json:"runtime_summary_json" v:"required"`
	ArtifactRefs          []braincontracts.ArtifactRef `json:"artifact_refs"`
	CreatedBy             string                       `json:"created_by"`
}

type CreateRepairDraftRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}

type RepairDraftRes struct {
	RepairDraft RepairDraftDetailView `json:"repair_draft"`
}

type RepairDraftDetailView struct {
	ID                    string   `json:"id"`
	Status                string   `json:"status"`
	ReasoningSummary      string   `json:"reasoning_summary"`
	ReplacedConstraints   []string `json:"replaced_constraints,omitempty"`
	FailedTaskContextJSON string   `json:"failed_task_context_json,omitempty"`
	FailureReasonJSON     string   `json:"failure_reason_json,omitempty"`
	OriginalContractsJSON string   `json:"original_contracts_json,omitempty"`
	RuntimeSummaryJSON    string   `json:"runtime_summary_json,omitempty"`
	RepairPlanJSON        string   `json:"repair_plan_json,omitempty"`
	CreatedBy             string   `json:"created_by,omitempty"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}
