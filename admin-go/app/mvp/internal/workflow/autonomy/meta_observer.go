package autonomy

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/utility/snowflake"
)

// MetaObserver 元认知观测器。
//
// 职责：在 DecisionCenter.Decide() 返回后，异步记录 ObservationRecord。
// 纯记录，零副作用。
type MetaObserver struct{}

// NewMetaObserver 创建观测器。
func NewMetaObserver() *MetaObserver {
	return &MetaObserver{}
}

// IsEnabled 灰度开关。
func (o *MetaObserver) IsEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.meta_cognition_enabled",
		"workflow.autonomy.metaCognitionEnabled", 0) == 1
}

// ObservationInput 观测输入，由 DecisionCenter 组装。
type ObservationInput struct {
	DecisionActionID int64
	WorkflowRunID    int64
	ProjectID        int64
	DecisionType     string // policy_match / strategy:xxx / objective_guard
	TriggerSource    string
	DecisionLevel    string
	ActionType       string
	InputSnapshot    map[string]interface{} // 决策输入上下文
	OutputSnapshot   map[string]interface{} // 决策输出
	Meta             *DecisionMeta
	HumanOverride    bool
	OverrideReason   string
	Outcome          string  // success / failure / neutral / pending
	EffectScore      float64 // -1~1
	CreatedBy        int64
	DeptID           int64
}

// Record 记录一次决策观测（异步写入，不阻塞决策流程）。
func (o *MetaObserver) Record(ctx context.Context, input *ObservationInput) {
	if input == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Warningf(ctx, "[MetaObserver] Record panic recovered: %v", r)
			}
		}()
		o.doRecord(ctx, input)
	}()
}

// doRecord 实际写入逻辑。
func (o *MetaObserver) doRecord(ctx context.Context, input *ObservationInput) {
	id := int64(snowflake.Generate())

	inputJSON, _ := json.Marshal(input.InputSnapshot)
	outputJSON, _ := json.Marshal(input.OutputSnapshot)

	var metaJSON []byte
	if input.Meta != nil {
		metaJSON, _ = json.Marshal(input.Meta)
	}

	// 计算学习信号权重
	weight := o.calcSignalWeight(input.DecisionLevel, input.HumanOverride)

	outcome := input.Outcome
	if outcome == "" {
		outcome = "pending"
	}

	humanOverride := 0
	if input.HumanOverride {
		humanOverride = 1
	}

	_, err := g.DB().Model("mvp_observation_record").Ctx(ctx).Insert(g.Map{
		"id":                  id,
		"decision_action_id":  input.DecisionActionID,
		"workflow_run_id":     input.WorkflowRunID,
		"project_id":          input.ProjectID,
		"decision_type":       input.DecisionType,
		"trigger_source":      input.TriggerSource,
		"decision_level":      input.DecisionLevel,
		"action_type":         input.ActionType,
		"input_snapshot":      string(inputJSON),
		"output_snapshot":     string(outputJSON),
		"meta_snapshot":       string(metaJSON),
		"outcome":             outcome,
		"effect_score":        input.EffectScore,
		"human_override":      humanOverride,
		"override_reason":     input.OverrideReason,
		"signal_weight":       weight,
		"created_by":          input.CreatedBy,
		"dept_id":             input.DeptID,
		"created_at":          gtime.Now(),
	})
	if err != nil {
		g.Log().Warningf(ctx, "[MetaObserver] 观测记录写入失败: %v", err)
	}
}

// UpdateOutcome 回填观测结果（决策执行完成后调用）。
func (o *MetaObserver) UpdateOutcome(ctx context.Context, decisionActionID int64, outcome string, effectScore float64) {
	_, err := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("decision_action_id", decisionActionID).
		Update(g.Map{
			"outcome":      outcome,
			"effect_score": effectScore,
		})
	if err != nil {
		g.Log().Warningf(ctx, "[MetaObserver] 回填观测结果失败: actionID=%d err=%v", decisionActionID, err)
	}
}

// UpdateHumanOverride 记录人工干预事件。
func (o *MetaObserver) UpdateHumanOverride(ctx context.Context, decisionActionID int64, action, reason string) {
	outcome := "neutral"
	if action == consts.HandleActionApprove {
		outcome = "success"
	} else if action == consts.HandleActionReject {
		outcome = "failure" // 被驳回意味着系统决策错误
	}

	// B级被驳回 = 强信号(1.0)
	weight := 0.6
	if action == consts.HandleActionReject {
		weight = 1.0
	}

	_, err := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("decision_action_id", decisionActionID).
		Update(g.Map{
			"human_override":  1,
			"override_reason": reason,
			"outcome":         outcome,
			"signal_weight":   weight,
		})
	if err != nil {
		g.Log().Warningf(ctx, "[MetaObserver] 记录人工干预失败: actionID=%d err=%v", decisionActionID, err)
	}
}

// calcSignalWeight 计算学习信号权重。
//
// 信号强度规则（来自设计文档）：
//   - A 级自动执行，人工未干预 → 弱信号 0.3
//   - B 级，人工批准 → 中信号 0.6
//   - B 级，人工驳回 → 强信号 1.0（在 UpdateHumanOverride 中更新）
//   - C 级，人工给出方案 → 学习样本 1.0
func (o *MetaObserver) calcSignalWeight(level string, humanOverride bool) float64 {
	switch level {
	case consts.DecisionLevelA:
		if humanOverride {
			return 1.0 // A 级却被人工干预 = 系统出错
		}
		return 0.3 // 弱信号
	case consts.DecisionLevelB:
		return 0.6 // 初始中信号，后续被批准/驳回时更新
	case consts.DecisionLevelC:
		return 1.0 // 强信号
	default:
		return 0.3
	}
}

// ListByProject 查询项目的观测记录（供 Assessor 和前端使用）。
func (o *MetaObserver) ListByProject(ctx context.Context, projectID int64, limit int) ([]g.Map, error) {
	if limit <= 0 {
		limit = 100
	}
	records, err := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// CountByOutcome 按结果分组统计（供 Assessor 使用）。
func (o *MetaObserver) CountByOutcome(ctx context.Context, projectID int64, periodStart, periodEnd *gtime.Time) (map[string]int, error) {
	result := map[string]int{
		"success": 0,
		"failure": 0,
		"neutral": 0,
		"pending": 0,
	}
	m := g.DB().Model("mvp_observation_record").Ctx(ctx).
		WhereNull("deleted_at")
	if projectID > 0 {
		m = m.Where("project_id", projectID)
	}
	if periodStart != nil {
		m = m.WhereGTE("created_at", periodStart)
	}
	if periodEnd != nil {
		m = m.WhereLTE("created_at", periodEnd)
	}
	records, err := m.Fields("outcome, COUNT(*) as cnt").Group("outcome").All()
	if err != nil {
		return result, err
	}
	for _, r := range records {
		outcome := r["outcome"].String()
		cnt := r["cnt"].Int()
		result[outcome] = cnt
	}
	return result, nil
}
