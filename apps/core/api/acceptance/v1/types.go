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
}

type CoverageItem struct {
	Key            string `json:"key"`
	Kind           string `json:"kind"`
	Name           string `json:"name"`
	CoverageStatus string `json:"coverage_status"`
	EvidenceCount  int    `json:"evidence_count"`
}

type AcceptanceIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Blocking bool   `json:"blocking"`
	Summary  string `json:"summary"`
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
