package v1

type HomeSummary struct {
	TotalProjects   int `json:"total_projects"`
	ActiveProjects  int `json:"active_projects"`
	BlockedProjects int `json:"blocked_projects"`
	PendingActions  int `json:"pending_actions"`
}

type HomeOverview struct {
	TotalProjects        int `json:"total_projects"`
	ActiveProjects       int `json:"active_projects"`
	BlockedProjects      int `json:"blocked_projects"`
	PendingActions       int `json:"pending_actions"`
	AttentionItemCount   int `json:"attention_item_count"`
	ReleaseWatchCount    int `json:"release_watch_count"`
	ProductionReadyCount int `json:"production_ready_count"`
}

type ProjectCard struct {
	ProjectID        string `json:"project_id"`
	Name             string `json:"name"`
	ProjectCategory  string `json:"project_category"`
	CurrentStage     string `json:"current_stage"`
	StageStatus      string `json:"stage_status"`
	ProgressPercent  int    `json:"progress_percent"`
	ProductionStatus string `json:"production_status"`
}

type NeedAttentionItem struct {
	ItemID         string `json:"item_id"`
	ProjectID      string `json:"project_id"`
	Title          string `json:"title"`
	Severity       string `json:"severity"`
	IsBlocking     bool   `json:"is_blocking"`
	RecommendedAct string `json:"recommended_action"`
}

type LiveActivityItem struct {
	EventID        string `json:"event_id"`
	ProjectID      string `json:"project_id"`
	EventType      string `json:"event_type"`
	Title          string `json:"title"`
	SourceBrain    string `json:"source_brain"`
	OccurredAt     string `json:"occurred_at"`
	NeedsAttention bool   `json:"needs_attention"`
}

type ReleaseReadiness struct {
	ProjectID        string `json:"project_id"`
	Name             string `json:"name"`
	ProductionStatus string `json:"production_status"`
	MissingItems     int    `json:"missing_items"`
}

type ProjectStreamEvent struct {
	EventID      string      `json:"event_id"`
	ProjectID    string      `json:"project_id"`
	RunBindingID string      `json:"run_binding_id"`
	SequenceNo   int         `json:"sequence_no"`
	EventType    string      `json:"event_type"`
	EventLevel   string      `json:"event_level"`
	Summary      string      `json:"summary"`
	Payload      interface{} `json:"payload,omitempty"`
	CreatedAt    string      `json:"created_at"`
}
