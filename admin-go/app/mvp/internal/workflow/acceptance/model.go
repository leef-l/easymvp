// Package acceptance 提供自动验收的核心领域模型与服务。
package acceptance

// AcceptContext 验收上下文，贯穿整个验收流程。
type AcceptContext struct {
	WorkflowRunID int64
	ProjectID     int64
	AcceptRunID   int64
	StageRunID    int64
	ProjectType   string // 项目分类编码（category_code，如 software_dev）
	FamilyCode    string // 能力家族编码（如 coding/creative/analysis）
	WorkDir       string // 项目工作目录
	CreatedBy     int64  // 继承自 project.created_by
	DeptID        int64  // 继承自 project.dept_id
}

// RuleHit 规则命中结果。
type RuleHit struct {
	RuleCode       string `json:"rule_code"`
	RuleName       string `json:"rule_name"`
	RuleType       string `json:"rule_type"`       // artifact/process/quality
	ScopeType      string `json:"scope_type"`      // project/task/file/stage
	Severity       string `json:"severity"`         // info/warn/error/blocker
	Title          string `json:"title"`
	Detail         string `json:"detail"`
	ExpectedValue  string `json:"expected_value"`
	ActualValue    string `json:"actual_value"`
	SuggestedAction string `json:"suggested_action"`
	DomainTaskID   int64  `json:"domain_task_id,omitempty"`
	ResourceRef    string `json:"resource_ref,omitempty"`
}

// EvidenceItem 证据条目。
type EvidenceItem struct {
	EvidenceType string `json:"evidence_type"` // task_output/file/log/diff/stage_output/handoff/summary
	SourceType   string `json:"source_type"`   // domain_task/stage_run/file/handoff_record/workflow_run
	SourceID     int64  `json:"source_id,omitempty"`
	ContentRef   string `json:"content_ref,omitempty"`
	Summary      string `json:"summary"`
}

// DecisionResult 最终裁决结果。
type DecisionResult struct {
	Decision string  `json:"decision"` // passed/failed/manual_review
	Score    float64 `json:"score"`
	Summary  string  `json:"summary"`
	Issues   []RuleHit `json:"issues,omitempty"`
}

// Decision 常量。
const (
	DecisionPassed       = "passed"
	DecisionFailed       = "failed"
	DecisionManualReview = "manual_review"
)

// Severity 常量。
const (
	SeverityInfo    = "info"
	SeverityWarn    = "warn"
	SeverityError   = "error"
	SeverityBlocker = "blocker"
)
