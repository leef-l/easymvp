package entity

// RuntimeEscalations is the golang structure for table runtime_escalations.
type RuntimeEscalations struct {
	Id               string `json:"id" orm:"id"`
	ProjectId        string `json:"project_id" orm:"project_id"`
	AcceptanceRunId  string `json:"acceptance_run_id" orm:"acceptance_run_id"`
	Status           string `json:"status" orm:"status"`
	ReasonClass      string `json:"reason_class" orm:"reason_class"`
	SourceBrain      string `json:"source_brain" orm:"source_brain"`
	SourceTaskID     string `json:"source_task_id" orm:"source_task_id"`
	RunBindingID     string `json:"run_binding_id" orm:"run_binding_id"`
	RunStatus        string `json:"run_status" orm:"run_status"`
	Severity         string `json:"severity" orm:"severity"`
	Action           string `json:"action" orm:"action"`
	TaskID           string `json:"task_id" orm:"task_id"`
	RunID            string `json:"run_id" orm:"run_id"`
	Summary          string `json:"summary" orm:"summary"`
	PolicyDenied     int    `json:"policy_denied" orm:"policy_denied"`
	EvidenceRefsJSON string `json:"evidence_refs_json" orm:"evidence_refs_json"`
	ResolvedAt       string `json:"resolved_at" orm:"resolved_at"`
	ResolutionStatus string `json:"resolution_status" orm:"resolution_status"`
	ResolverKind     string `json:"resolver_kind" orm:"resolver_kind"`
	LinkedFaultID    string `json:"linked_fault_id" orm:"linked_fault_id"`
	CreatedAt        string `json:"created_at" orm:"created_at"`
	UpdatedAt        string `json:"updated_at" orm:"updated_at"`
}
