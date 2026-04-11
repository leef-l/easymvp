package autonomy

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/repo"
)

// ProjectObjective 项目目标约束。
type ProjectObjective struct {
	DeliveryGoal         string   `json:"deliveryGoal"`
	QualityFloor         float64  `json:"qualityFloor"`
	TokenBudget          int64    `json:"tokenBudget"`
	TimeBudgetHours      float64  `json:"timeBudgetHours"`
	CostBudgetCents      int64    `json:"costBudgetCents"`
	RiskTolerance        string   `json:"riskTolerance"`
	MaxAutoRetries       int      `json:"maxAutoRetries"`
	MaxAutoReworks       int      `json:"maxAutoReworks"`
	MaxAutoReplans       int      `json:"maxAutoReplans"`
	DeadlineAt           string   `json:"deadlineAt"`
	MaxStallMinutes      int      `json:"maxStallMinutes"`
	AutonomyLevel        string   `json:"autonomyLevel"`
	MaxSideEffectLevel   string   `json:"maxSideEffectLevel"`
	AllowedStateChanges  []string `json:"allowedStateChanges"`
	HumanMandatoryPoints []string `json:"humanMandatoryPoints"`
}

// AdmissionResult 目标层准入结果。
type AdmissionResult struct {
	Allowed         bool          `json:"allowed"`
	DenyReason      string        `json:"denyReason,omitempty"`
	Conditions      []string      `json:"conditions,omitempty"`
	SuggestedAction string        `json:"suggestedAction,omitempty"`
	HumanRequired   bool          `json:"humanRequired"`
	DecisionMeta    *DecisionMeta `json:"decisionMeta,omitempty"`
}

// ObjectiveService 目标层服务。
type ObjectiveService struct{}

func NewObjectiveService() *ObjectiveService { return &ObjectiveService{} }

// Load 加载项目目标约束；未配置时回退到系统默认值。
func (s *ObjectiveService) Load(ctx context.Context, projectID int64) (*ProjectObjective, error) {
	obj := s.defaultObjective(ctx)
	if projectID == 0 {
		return obj, nil
	}
	project, err := repo.NewProjectRepo().GetByID(ctx, projectID, "objective_json")
	if err != nil || len(project) == 0 {
		return obj, err
	}

	var override ProjectObjective
	if err = json.Unmarshal([]byte(gconv.String(project["objective_json"])), &override); err != nil {
		return obj, nil
	}
	s.mergeObjective(obj, &override)
	return obj, nil
}

// Check 执行目标层准入控制。
func (s *ObjectiveService) Check(ctx context.Context, sit *Situation, obj *ProjectObjective, actionType string) (*AdmissionResult, error) {
	res := &AdmissionResult{Allowed: true}
	if obj == nil || sit == nil {
		return res, nil
	}

	if obj.AutonomyLevel == "manual" {
		return s.denied("autonomy_level_manual", "notify_human", true, &DecisionMeta{
			Confidence:          1,
			EvidenceSufficiency: 1,
			Reversibility:       "full",
			BlastRadius:         "workflow",
		}), nil
	}

	if obj.TokenBudget > 0 && sit.Resource != nil && sit.Resource.TokensConsumed >= obj.TokenBudget {
		return s.denied("token_budget_exhausted", "notify_human", true, &DecisionMeta{
			Confidence:          1,
			EvidenceSufficiency: 0.95,
			Reversibility:       "full",
			BlastRadius:         "workflow",
		}), nil
	}

	if obj.MaxAutoRetries > 0 && sit.Health != nil && sit.Health.RetryCount >= obj.MaxAutoRetries && actionType == "task.failed" {
		return s.denied("retry_limit_reached", "notify_human", true, &DecisionMeta{
			Confidence:          0.9,
			EvidenceSufficiency: 0.9,
			Reversibility:       "full",
			BlastRadius:         "stage",
		}), nil
	}

	if obj.MaxAutoReworks > 0 && sit.Health != nil && sit.Health.ReworkRounds >= obj.MaxAutoReworks {
		return s.denied("rework_limit_reached", "notify_human", true, &DecisionMeta{
			Confidence:          0.9,
			EvidenceSufficiency: 0.85,
			Reversibility:       "full",
			BlastRadius:         "workflow",
		}), nil
	}

	if obj.MaxAutoReplans > 0 && sit.Health != nil && sit.Health.ReplanCount >= obj.MaxAutoReplans {
		return s.denied("replan_limit_reached", "notify_human", true, &DecisionMeta{
			Confidence:          0.9,
			EvidenceSufficiency: 0.85,
			Reversibility:       "full",
			BlastRadius:         "workflow",
		}), nil
	}

	if obj.TimeBudgetHours > 0 && sit.WorkflowStartedAt != nil {
		if gtime.Now().Timestamp()-sit.WorkflowStartedAt.Timestamp() >= int64(obj.TimeBudgetHours*3600) {
			return s.denied("time_budget_exhausted", "notify_human", true, &DecisionMeta{
				Confidence:          0.85,
				EvidenceSufficiency: 0.8,
				Reversibility:       "full",
				BlastRadius:         "workflow",
			}), nil
		}
	}

	if obj.DeadlineAt != "" {
		if deadline := gtime.New(obj.DeadlineAt); deadline != nil && !deadline.IsZero() && gtime.Now().After(deadline) {
			return s.denied("deadline_exceeded", "notify_human", true, &DecisionMeta{
				Confidence:          0.95,
				EvidenceSufficiency: 0.8,
				Reversibility:       "full",
				BlastRadius:         "workflow",
			}), nil
		}
	}

	return res, nil
}

func (s *ObjectiveService) denied(reason, action string, humanRequired bool, meta *DecisionMeta) *AdmissionResult {
	return &AdmissionResult{
		Allowed:         false,
		DenyReason:      reason,
		SuggestedAction: action,
		HumanRequired:   humanRequired,
		Conditions:      []string{reason},
		DecisionMeta:    meta,
	}
}

func (s *ObjectiveService) defaultObjective(ctx context.Context) *ProjectObjective {
	return &ProjectObjective{
		QualityFloor:       0.8,
		TokenBudget:        int64(engine.GetConfigInt(ctx, "workflow.autonomy.default_token_budget", "workflow.autonomy.defaultTokenBudget", 0)),
		TimeBudgetHours:    float64(engine.GetConfigInt(ctx, "workflow.autonomy.default_time_budget_hours", "workflow.autonomy.defaultTimeBudgetHours", 0)),
		RiskTolerance:      engine.GetConfigString(ctx, "workflow.autonomy.default_risk_tolerance", "workflow.autonomy.defaultRiskTolerance", "balanced"),
		AutonomyLevel:      engine.GetConfigString(ctx, "workflow.autonomy.default_autonomy_level", "workflow.autonomy.defaultAutonomyLevel", "supervised"),
		MaxAutoRetries:     3,
		MaxAutoReworks:     2,
		MaxAutoReplans:     1,
		MaxStallMinutes:    60,
		MaxSideEffectLevel: "medium",
	}
}

func (s *ObjectiveService) mergeObjective(base, override *ProjectObjective) {
	if override == nil {
		return
	}
	if override.DeliveryGoal != "" {
		base.DeliveryGoal = override.DeliveryGoal
	}
	if override.QualityFloor > 0 {
		base.QualityFloor = override.QualityFloor
	}
	if override.TokenBudget > 0 {
		base.TokenBudget = override.TokenBudget
	}
	if override.TimeBudgetHours > 0 {
		base.TimeBudgetHours = override.TimeBudgetHours
	}
	if override.CostBudgetCents > 0 {
		base.CostBudgetCents = override.CostBudgetCents
	}
	if override.RiskTolerance != "" {
		base.RiskTolerance = override.RiskTolerance
	}
	if override.MaxAutoRetries > 0 {
		base.MaxAutoRetries = override.MaxAutoRetries
	}
	if override.MaxAutoReworks > 0 {
		base.MaxAutoReworks = override.MaxAutoReworks
	}
	if override.MaxAutoReplans > 0 {
		base.MaxAutoReplans = override.MaxAutoReplans
	}
	if override.DeadlineAt != "" {
		base.DeadlineAt = override.DeadlineAt
	}
	if override.MaxStallMinutes > 0 {
		base.MaxStallMinutes = override.MaxStallMinutes
	}
	if override.AutonomyLevel != "" {
		base.AutonomyLevel = override.AutonomyLevel
	}
	if override.MaxSideEffectLevel != "" {
		base.MaxSideEffectLevel = override.MaxSideEffectLevel
	}
	if len(override.AllowedStateChanges) > 0 {
		base.AllowedStateChanges = override.AllowedStateChanges
	}
	if len(override.HumanMandatoryPoints) > 0 {
		base.HumanMandatoryPoints = override.HumanMandatoryPoints
	}
}
