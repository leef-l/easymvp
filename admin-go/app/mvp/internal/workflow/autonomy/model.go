// Package autonomy 自治项目管理：风险评估、熔断、重规划、汇报、执行器选择。
package autonomy

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
)

// ==================== 风险评估 ====================

// 风险级别常量。
const (
	RiskTransient  = "transient"  // 瞬态错误：网络超时、API 限流
	RiskStructural = "structural" // 结构性错误：方案缺陷、依赖缺失
	RiskFatal      = "fatal"      // 致命错误：环境损坏、权限错误
)

// RiskInput 风险评估输入。
type RiskInput struct {
	WorkflowRunID int64
	TaskID        int64
	TaskName      string
	ErrorMessage  string
	RetryCount    int
	BatchNo       int
	RoleType      string
}

// RiskResult 风险评估结果。
type RiskResult struct {
	Level      string `json:"level"`      // transient / structural / fatal
	Confidence float64 `json:"confidence"` // 0-1
	Reason     string `json:"reason"`
	Action     string `json:"action"` // retry / rework / replan / pause
}

// ==================== 熔断器 ====================

// CircuitBreakResult 熔断检查结果。
type CircuitBreakResult struct {
	ShouldBreak bool   `json:"shouldBreak"`
	Reason      string `json:"reason"`
	Metrics     *BreakMetrics `json:"metrics"`
}

// BreakMetrics 熔断指标。
type BreakMetrics struct {
	ConsecutiveFailures int     `json:"consecutiveFailures"`
	BatchFailureRate    float64 `json:"batchFailureRate"`
	ReworkRounds        int     `json:"reworkRounds"`
	AcceptRounds        int     `json:"acceptRounds"`
}

// ==================== 重规划 ====================

// ReplanAction 重规划动作类型。
const (
	ReplanPartial = "replan_partial" // 部分调整
	ReplanFull    = "replan_full"    // 全量重做
	ReplanAbort   = "abort"          // 建议终止
)

// ReplanInput 重规划输入。
type ReplanInput struct {
	WorkflowRunID int64
	ProjectID     int64
	TriggerSource string // rework_failed / batch_wipeout / accept_failed / manual / circuit_breaker
	BreakReason   string // 熔断原因（仅 circuit_breaker 触发时填充）
	FailedTasks   []FailedTaskInfo
	AcceptIssues  []string
}

// FailedTaskInfo 失败任务信息。
type FailedTaskInfo struct {
	TaskID       int64  `json:"taskId"`
	TaskName     string `json:"taskName"`
	ErrorMessage string `json:"errorMessage"`
	RetryCount   int    `json:"retryCount"`
}

// ReplanRecommendation 重规划建议。
type ReplanRecommendation struct {
	Action          string   `json:"action"` // replan_partial / replan_full / abort
	AffectedTaskIDs []int64  `json:"affectedTaskIds,omitempty"`
	NewPlanSummary  string   `json:"newPlanSummary"`
	Reasoning       string   `json:"reasoning"`
}

// ==================== 汇报 ====================

// ReportType 报告类型。
const (
	ReportStage   = "stage"
	ReportDaily   = "daily"
	ReportWeekly  = "weekly"
	ReportSummary = "summary"
)

// ProjectReport 项目报告。
type ProjectReport struct {
	ID            int64       `json:"id"`
	WorkflowRunID int64       `json:"workflowRunId"`
	ProjectID     int64       `json:"projectId"`
	ReportType    string      `json:"reportType"`
	StageType     string      `json:"stageType,omitempty"`
	Title         string      `json:"title"`
	Content       string      `json:"content"`
	Metrics       string      `json:"metrics,omitempty"`
	CreatedAt     *gtime.Time `json:"createdAt"`
}

// ==================== 自治决策 ====================

// DecisionType 决策类型常量。
const (
	DecisionReplan       = "replan"
	DecisionRiskEscalate = "risk_escalate"
	DecisionEngineSwitch = "engine_switch"
	DecisionCircuitBreak = "circuit_break"
	DecisionReport       = "report"
)

// DecisionMode 决策模式。
const (
	ModeSuggest = "suggest" // 建议型：需要人工确认
	ModeAuto    = "auto"    // 自动型：直接执行
)

// HumanAction 人工动作。
const (
	ActionPending  = "pending"
	ActionApproved = "approved"
	ActionRejected = "rejected"
)

// AutonomyDecision 自治决策记录。
type AutonomyDecision struct {
	ID             int64       `json:"id"`
	WorkflowRunID  int64       `json:"workflowRunId"`
	ProjectID      int64       `json:"projectId"`
	DecisionType   string      `json:"decisionType"`
	TriggerSource  string      `json:"triggerSource"`
	TriggerContext string      `json:"triggerContext,omitempty"`
	Recommendation string      `json:"recommendation"`
	DecisionMode   string      `json:"decisionMode"`
	HumanAction    string      `json:"humanAction,omitempty"`
	ExecutedAt     *gtime.Time `json:"executedAt,omitempty"`
	Result         string      `json:"result,omitempty"`
	CreatedAt      *gtime.Time `json:"createdAt"`
}

// ==================== 执行器选择 ====================

// EngineRecommendation 执行器推荐结果。
type EngineRecommendation struct {
	EngineType string  `json:"engineType"`
	Confidence float64 `json:"confidence"` // 0-1
	Reason     string  `json:"reason"`
}

// ==================== 自治模式配置 ====================

// GetAutonomyMode 读取自治模式配置：suggest（默认）或 auto。
// 配置键：autonomy.mode，支持三级 fallback（DB mvp_config → config.yaml → 默认值）。
func GetAutonomyMode(ctx context.Context) string {
	mode := engine.GetConfigString(ctx, "autonomy.mode", "engine.autonomy.mode", ModeSuggest)
	if mode != ModeSuggest && mode != ModeAuto {
		return ModeSuggest
	}
	return mode
}

// IsAutoMode 判断当前是否处于全自动模式。
func IsAutoMode(ctx context.Context) bool {
	return GetAutonomyMode(ctx) == ModeAuto
}

// ==================== 自治中台类型（DecisionCenter）====================

// DecisionRequest 统一决策请求。
type DecisionRequest struct {
	WorkflowRunID  int64
	ProjectID      int64
	StageRunID     int64 // 可选
	DomainTaskID   int64 // 可选
	TriggerSource  string
	TriggerContext map[string]interface{}
}

// DecisionResponse 统一决策结果。
type DecisionResponse struct {
	ActionID       int64
	DecisionLevel  string
	ActionType     string
	AutoExecutable bool
	HumanRequired  bool
	Executed       bool
	Handled        bool  // 是否已被中台接管（audit_only=true 时为 false，原逻辑应继续执行）
	Error          error
}

// PolicyMatch 策略匹配结果。
type PolicyMatch struct {
	Rule           *PolicyRule
	DecisionLevel  string
	ActionType     string
	AutoExecutable bool
}

// PolicyRule 策略规则（DB 行映射）。
type PolicyRule struct {
	ID                  int64
	RuleCode            string
	RuleName            string
	DecisionType        string
	DecisionLevel       string
	TriggerSource       string
	ProjectFamily       string
	ProjectCategoryCode string
	ConfigJSON          map[string]interface{}
	Enabled             bool
	Priority            int
}

// GateCheckResult 闸门检查结果。
type GateCheckResult struct {
	Blocked      bool
	BlockedGates []BlockedGate
}

// BlockedGate 被命中的闸门。
type BlockedGate struct {
	GateID         int64
	GateCode       string
	GateType       string
	BlockAction    string
	FallbackAction string
	Reason         string
}

// RiskGateRule 闸门规则（DB 行映射）。
type RiskGateRule struct {
	ID                  int64
	GateCode            string
	GateName            string
	GateType            string
	ProjectFamily       string
	ProjectCategoryCode string
	TriggerExpression   map[string]interface{}
	BlockAction         string
	FallbackAction      string
	Enabled             bool
	Priority            int
}
