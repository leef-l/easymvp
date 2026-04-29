package v1

import "github.com/gogf/gf/v2/frame/g"

type PlanViewReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/plan-view" tags:"Plan" method:"get" summary:"Plan view"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type PlanViewRes struct {
	Overview       PlanOverview       `json:"overview"`
	Draft          PlanDraftView      `json:"draft"`
	Review         PlanReviewView     `json:"review"`
	Compiled       CompiledPlanView   `json:"compiled"`
	RepairDraft    RepairDraftView    `json:"repair_draft"`
	TaskProjection []CompiledTaskView `json:"task_projection"`
	DiffSummary    DiffSummary        `json:"diff_summary"`
}

type CompilePlanReq struct {
	g.Meta              `path:"/api/v3/projects/{id}/plan/compile" tags:"Plan" method:"post" summary:"Compile plan"`
	ProjectID           string `json:"id" in:"path" v:"required"`
	PlanDraftID         string `json:"plan_draft_id"`
	ForceRecompile      bool   `json:"force_recompile"`
	AutoRedesign        bool   `json:"auto_redesign"`
	MaxRedesignAttempts int    `json:"max_redesign_attempts"`
}

type CompilePlanRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}

type RedesignPlanReq struct {
	g.Meta   `path:"/api/v3/projects/{id}/plan/redesign" tags:"Plan" method:"post" summary:"Redesign plan draft based on review feedback"`
	Id       string `json:"id" in:"path" v:"required"`
	Feedback string `json:"feedback"` // optional user feedback
}

type RedesignPlanRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}

type PlanDraftView struct {
	ID          string `json:"id"`
	Version     int    `json:"version"`
	Status      string `json:"status"`
	GoalSummary string `json:"goal_summary"`
}

type PlanReviewView struct {
	ID                 string `json:"id"`
	ReviewVersion      int    `json:"review_version"`
	Decision           string `json:"decision"`
	BlockingIssueCount int    `json:"blocking_issue_count"`
	AdvisoryIssueCount int    `json:"advisory_issue_count"`
}

type CompiledPlanView struct {
	ID              string `json:"id"`
	CompiledVersion int    `json:"compiled_version"`
	Status          string `json:"status"`
	RiskSummary     string `json:"risk_summary,omitempty"`
}

type RepairDraftView struct {
	ID                  string   `json:"id"`
	Status              string   `json:"status"`
	ReasoningSummary    string   `json:"reasoning_summary"`
	ReplacedConstraints []string `json:"replaced_constraints,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

type CompiledTaskView struct {
	TaskID                string   `json:"task_id"`
	TaskKey               string   `json:"task_key"`
	TaskName              string   `json:"task_name"`
	Phase                 string   `json:"phase"`
	TaskKind              string   `json:"task_kind"`
	RoleType              string   `json:"role_type"`
	BrainKind             string   `json:"brain_kind"`
	RiskLevel             string   `json:"risk_level"`
	Status                string   `json:"status"`
	DeliverySummary       string   `json:"delivery_summary"`
	VerificationSummary   string   `json:"verification_summary"`
	AffectedResources     []string `json:"affected_resources"`
	ManualReviewRequired  bool     `json:"manual_review_required"`
	MappedDomainTaskID    string   `json:"mapped_domain_task_id,omitempty"`
	MappedDomainTaskState string   `json:"mapped_domain_task_status,omitempty"`
}

type DiffSummary struct {
	TotalChanges     int               `json:"total_changes"`
	SplitCount       int               `json:"split_count"`
	OverrideCount    int               `json:"override_count"`
	DropCount        int               `json:"drop_count"`
	UnchangedCount   int               `json:"unchanged_count"`
	ReviewIssueCount int               `json:"review_issue_count"`
	Summary          string            `json:"summary"`
	Items            []DiffSummaryItem `json:"items"`
}

type DiffSummaryItem struct {
	DiffKind            string `json:"diff_kind"`
	BeforeLabel         string `json:"before_label"`
	AfterLabel          string `json:"after_label"`
	Reason              string `json:"reason"`
	SourceReviewIssueID string `json:"source_review_issue_id,omitempty"`
}

type PlanOverview struct {
	ProjectID             string `json:"project_id"`
	DraftStatus           string `json:"draft_status"`
	ReviewDecision        string `json:"review_decision"`
	CompiledStatus        string `json:"compiled_status"`
	RepairDraftStatus     string `json:"repair_draft_status"`
	CurrentStage          string `json:"current_stage"`
	NextAction            string `json:"next_action"`
	TaskCount             int    `json:"task_count"`
	ManualReviewTaskCount int    `json:"manual_review_task_count"`
	BlockingIssueCount    int    `json:"blocking_issue_count"`
	AdvisoryIssueCount    int    `json:"advisory_issue_count"`
	CompiledVersion       int    `json:"compiled_version"`
}
