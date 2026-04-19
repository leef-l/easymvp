package v1

type AcceptanceRunView struct {
	ID                     string `json:"id"`
	TaskID                 string `json:"task_id,omitempty"`
	ProfileVersion         string `json:"profile_version,omitempty"`
	Status                 string `json:"status"`
	FunctionalStatus       string `json:"functional_status"`
	ProductionStatus       string `json:"production_status"`
	ManualReleaseRequired  bool   `json:"manual_release_required"`
	FinishedAt             string `json:"finished_at,omitempty"`
	LatestJudgementKind    string `json:"latest_judgement_kind,omitempty"`
	LatestJudgementResult  string `json:"latest_judgement_result,omitempty"`
	LatestJudgementSummary string `json:"latest_judgement_summary,omitempty"`
	LatestJudgementAt      string `json:"latest_judgement_at,omitempty"`
	ReleaseGateStatus      string `json:"release_gate_status,omitempty"`
	NextAction             string `json:"next_action,omitempty"`
	BlockingIssueCount     int    `json:"blocking_issue_count,omitempty"`
}

type CoverageItem struct {
	Key            string `json:"key"`
	Kind           string `json:"kind"`
	Name           string `json:"name"`
	CoverageStatus string `json:"coverage_status"`
	EvidenceCount  int    `json:"evidence_count"`
}

type AcceptanceIssue struct {
	ID              string `json:"id"`
	AcceptanceRunID string `json:"acceptance_run_id,omitempty"`
	Severity        string `json:"severity"`
	IssueKind       string `json:"issue_kind,omitempty"`
	Blocking        bool   `json:"blocking"`
	Summary         string `json:"summary"`
	DetailJSON      string `json:"detail_json,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

type EvidenceCard struct {
	ID           string `json:"id"`
	Surface      string `json:"surface"`
	Journey      string `json:"journey,omitempty"`
	EvidenceType string `json:"evidence_type"`
	FilePath     string `json:"file_path"`
	CapturedAt   string `json:"captured_at"`
}

type ReleaseGateView struct {
	Status     string `json:"status"`
	NextAction string `json:"next_action"`
	Summary    string `json:"summary"`
}

type AcceptanceOverview struct {
	ProjectID           string `json:"project_id"`
	CurrentStage        string `json:"current_stage"`
	OverallStatus       string `json:"overall_status"`
	FunctionalStatus    string `json:"functional_status"`
	ProductionStatus    string `json:"production_status"`
	ReleaseGateStatus   string `json:"release_gate_status"`
	NextAction          string `json:"next_action"`
	BlockingIssueCount  int    `json:"blocking_issue_count"`
	CoveredItemCount    int    `json:"covered_item_count"`
	RequiredItemCount   int    `json:"required_item_count"`
	EvidenceCardCount   int    `json:"evidence_card_count"`
	ManualReleaseNeeded bool   `json:"manual_release_required"`
}

type VerificationResultView struct {
	Status                   string   `json:"status"`
	PreferredChannel         string   `json:"preferred_verification_channel,omitempty"`
	RequiredChecks           []string `json:"required_checks,omitempty"`
	RequiredEvidence         []string `json:"required_evidence,omitempty"`
	MissingEvidence          []string `json:"missing_evidence,omitempty"`
	FailedChecks             []string `json:"failed_checks,omitempty"`
	VerificationContractJSON string   `json:"verification_contract_json,omitempty"`
	SourceRunID              string   `json:"source_run_id,omitempty"`
	UpdatedAt                string   `json:"updated_at,omitempty"`
	Decision                 string   `json:"decision,omitempty"`
	Completed                bool     `json:"completed"`
	Summary                  string   `json:"summary,omitempty"`
}

type CompletionVerdictView struct {
	Decision               string `json:"decision"`
	FinalStatus            string `json:"final_status,omitempty"`
	Reason                 string `json:"reason,omitempty"`
	ManualReviewRequired   bool   `json:"manual_review_required"`
	ManualReleaseRequired  bool   `json:"manual_release_required"`
	ManualReleaseCompleted bool   `json:"manual_release_completed"`
	ReleaseReady           bool   `json:"release_ready"`
	BlockerCount           int    `json:"blocker_count"`
	NextAction             string `json:"next_action,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
	Completed              bool   `json:"completed"`
	Summary                string `json:"summary,omitempty"`
}

type RuntimeEscalationView struct {
	Status       string `json:"status"`
	ReasonClass  string `json:"reason_class,omitempty"`
	SourceBrain  string `json:"source_brain,omitempty"`
	SourceTaskID string `json:"source_task_id,omitempty"`
	RunBindingID string `json:"run_binding_id,omitempty"`
	RunStatus    string `json:"run_status,omitempty"`
	Severity     string `json:"severity,omitempty"`
	Action       string `json:"action,omitempty"`
	TaskID       string `json:"task_id,omitempty"`
	RunID        string `json:"run_id,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	Summary      string `json:"summary,omitempty"`
	PolicyDenied bool   `json:"policy_denied"`
}

type FaultSummaryView struct {
	Status             string   `json:"status"`
	BlockingIssueCount int      `json:"blocking_issue_count"`
	AdvisoryIssueCount int      `json:"advisory_issue_count"`
	TopIssue           string   `json:"top_issue,omitempty"`
	FaultLoopDetected  bool     `json:"fault_loop_detected"`
	FaultKind          string   `json:"fault_kind,omitempty"`
	Severity           string   `json:"severity,omitempty"`
	Summary            string   `json:"summary,omitempty"`
	FailedChecks       []string `json:"failed_checks,omitempty"`
	AffectedTasks      []string `json:"affected_tasks,omitempty"`
	UpdatedAt          string   `json:"updated_at,omitempty"`
}

type RepairPlanDraftSummary struct {
	ID                   string   `json:"id,omitempty"`
	Status               string   `json:"status"`
	ReasonClass          string   `json:"reason_class,omitempty"`
	RepairStrategy       string   `json:"repair_strategy,omitempty"`
	ReasoningSummary     string   `json:"reasoning_summary,omitempty"`
	Summary              string   `json:"summary,omitempty"`
	UpdatedTasks         []string `json:"updated_tasks,omitempty"`
	ManualReviewRequired bool     `json:"manual_review_required"`
	UpdatedAt            string   `json:"updated_at,omitempty"`
}
