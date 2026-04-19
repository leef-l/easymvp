package v1

type ProjectSnapshot struct {
	ProjectID         string `json:"project_id"`
	Name              string `json:"name"`
	ProjectCategory   string `json:"project_category"`
	CurrentStage      string `json:"current_stage"`
	ProgressPercent   int    `json:"progress_percent"`
	RiskLevel         string `json:"risk_level"`
	ProductionStatus  string `json:"production_status"`
	ManualReleaseNeed bool   `json:"manual_release_required"`
}

type StageProgressItem struct {
	StageKey         string `json:"stage_key"`
	StageName        string `json:"stage_name"`
	Status           string `json:"status"`
	DurationSeconds  int64  `json:"duration_seconds"`
	ActiveItemTitle  string `json:"active_item_title"`
	BlockingIssueCnt int    `json:"blocking_issue_count"`
}

type LiveActivityItem struct {
	EventID        string `json:"event_id"`
	EventType      string `json:"event_type"`
	Title          string `json:"title"`
	SourceBrain    string `json:"source_brain"`
	SourceTaskID   string `json:"source_task_id"`
	OccurredAt     string `json:"occurred_at"`
	RequiresAction bool   `json:"requires_action"`
}

type ActionInboxItem struct {
	ItemID            string `json:"item_id"`
	Title             string `json:"title"`
	Severity          string `json:"severity"`
	IsBlocking        bool   `json:"is_blocking"`
	RecommendedAction string `json:"recommended_action"`
	TargetID          string `json:"target_id"`
}

type AcceptanceCoverage struct {
	Category         string `json:"category"`
	CoveredSurfaces  int    `json:"covered_surfaces"`
	RequiredSurfaces int    `json:"required_surfaces"`
	CoveredJourneys  int    `json:"covered_journeys"`
	RequiredJourneys int    `json:"required_journeys"`
	EvidenceReady    int    `json:"evidence_ready"`
	EvidenceRequired int    `json:"evidence_required"`
	ProductionPassed bool   `json:"production_passed"`
}

type WorkspaceExplanation struct {
	Headline           string                       `json:"headline"`
	Summary            string                       `json:"summary"`
	TopBlockers        []string                     `json:"top_blockers"`
	RecommendedActions []WorkspaceRecommendedAction `json:"recommended_actions"`
	ExplainLinks       []string                     `json:"explain_links,omitempty"`
}

type WorkspaceRecommendedAction struct {
	ActionKey string `json:"action_key"`
	Label     string `json:"label"`
	Reason    string `json:"reason"`
	DeepLink  string `json:"deep_link,omitempty"`
}
