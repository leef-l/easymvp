package autonomy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	"easymvp/utility/snowflake"
)

// MetaTuner 元认知校准器。
//
// 职责：根据 Assessor 检测到的 Drift 生成调参建议（TuneRecommendation）。
//
// 核心规则：
//   - 保守方向（降低风险）可自动应用（AutoApplicable=true）
//   - 激进方向（放宽限制）编译期固定 AutoApplicable=false
//   - 宪法保护参数永远不可修改
type MetaTuner struct {
	learner *Learner
}

// NewMetaTuner 创建校准器。
func NewMetaTuner(learner *Learner) *MetaTuner {
	return &MetaTuner{learner: learner}
}

// IsAutoTuneEnabled 自动校准开关。
func (t *MetaTuner) IsAutoTuneEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.meta_auto_tune_enabled",
		"workflow.autonomy.metaAutoTuneEnabled", 0) == 1
}

// TuneRecommendation 调参建议。
type TuneRecommendation struct {
	ID             int64       `json:"id"`
	AssessmentID   int64       `json:"assessmentId"`
	ProjectID      int64       `json:"projectId"`
	Parameter      string      `json:"parameter"`
	CurrentValue   string      `json:"currentValue"`
	SuggestedValue string      `json:"suggestedValue"`
	Direction      string      `json:"direction"`      // conservative / aggressive
	Reasoning      string      `json:"reasoning"`
	Confidence     float64     `json:"confidence"`
	AutoApplicable bool        `json:"autoApplicable"` // aggressive 时编译期固定 false
	RiskLevel      string      `json:"riskLevel"`      // low / medium / high
	Status         string      `json:"status"`         // pending / applied / rejected / expired
	AppliedAt      *gtime.Time `json:"appliedAt,omitempty"`
	AppliedBy      int64       `json:"appliedBy,omitempty"`
}

// constitutionProtectedParams 宪法保护参数列表 — Tuner 永远不能修改。
// 这些参数只能由人工在 UI 或数据库中变更。
var constitutionProtectedParams = map[string]bool{
	"workflow.autonomy.default_autonomy_level": true, // 自治级别
	"workflow.autonomy.enabled":                true, // 总开关
	"workflow.autonomy.audit_only":             true, // 审计模式（方向性变更必须人工）
	"workflow.autonomy.meta_cognition_enabled": true, // 元认知开关
	"workflow.autonomy.meta_auto_tune_enabled": true, // 自动校准开关
	"max_side_effect_level":                    true, // 最大副作用等级
	"human_mandatory_points":                   true, // 强制人工节点
}

// GenerateRecommendations 根据评估结果生成调参建议。
func (t *MetaTuner) GenerateRecommendations(ctx context.Context, assessment *AssessmentResult) []TuneRecommendation {
	if assessment == nil || len(assessment.Drifts) == 0 {
		return nil
	}

	var recommendations []TuneRecommendation

	for _, drift := range assessment.Drifts {
		// 宪法保护：跳过受保护参数
		if constitutionProtectedParams[drift.Parameter] {
			g.Log().Infof(ctx, "[MetaTuner] 跳过宪法保护参数: %s", drift.Parameter)
			continue
		}

		// 三重保护：样本量不足
		if !t.learner.HasEnoughSamples(ctx, drift.Parameter, assessment.ProjectID) {
			// 尝试用 system.human_override_rate 作为替代指标
			if !t.learner.HasEnoughSamples(ctx, "system.human_override_rate", assessment.ProjectID) {
				continue
			}
		}

		// 三重保护：置信度不足
		if drift.Confidence < 0.3 {
			continue
		}

		// 判断方向
		direction := t.classifyDirection(drift)

		// 三重保护：限幅
		suggestedValue := t.learner.ClampAdjustment(drift.CurrentValue, drift.OptimalValue)

		// 激进方向编译期固定 AutoApplicable=false
		autoApplicable := direction == "conservative"

		riskLevel := "low"
		if direction == "aggressive" {
			riskLevel = "medium"
			if drift.Confidence < 0.5 {
				riskLevel = "high"
			}
		}

		rec := TuneRecommendation{
			AssessmentID:   assessment.ID,
			ProjectID:      assessment.ProjectID,
			Parameter:      drift.Parameter,
			CurrentValue:   fmt.Sprintf("%.6f", drift.CurrentValue),
			SuggestedValue: fmt.Sprintf("%.6f", suggestedValue),
			Direction:      direction,
			Reasoning:      drift.Evidence,
			Confidence:     drift.Confidence,
			AutoApplicable: autoApplicable,
			RiskLevel:      riskLevel,
			Status:         "pending",
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// SaveAndApply 保存建议并自动应用保守方向的建议（如果开关开启）。
func (t *MetaTuner) SaveAndApply(ctx context.Context, recommendations []TuneRecommendation) error {
	autoTuneEnabled := t.IsAutoTuneEnabled(ctx)

	for i := range recommendations {
		rec := &recommendations[i]
		rec.ID = int64(snowflake.Generate())

		// 持久化
		_, err := g.DB().Model("mvp_tune_recommendation").Ctx(ctx).Insert(g.Map{
			"id":              rec.ID,
			"assessment_id":   rec.AssessmentID,
			"project_id":      rec.ProjectID,
			"parameter":       rec.Parameter,
			"current_value":   rec.CurrentValue,
			"suggested_value": rec.SuggestedValue,
			"direction":       rec.Direction,
			"reasoning":       rec.Reasoning,
			"confidence":      rec.Confidence,
			"auto_applicable": boolToInt(rec.AutoApplicable),
			"risk_level":      rec.RiskLevel,
			"status":          "pending",
			"created_at":      gtime.Now(),
		})
		if err != nil {
			g.Log().Warningf(ctx, "[MetaTuner] 保存建议失败: param=%s err=%v", rec.Parameter, err)
			continue
		}

		// 自动应用保守方向（如果开关开启）
		if autoTuneEnabled && rec.AutoApplicable && rec.Direction == "conservative" {
			if err := t.applyRecommendation(ctx, rec); err != nil {
				g.Log().Warningf(ctx, "[MetaTuner] 自动应用失败: param=%s err=%v", rec.Parameter, err)
			} else {
				g.Log().Infof(ctx, "[MetaTuner] 自动应用保守建议: param=%s %s→%s",
					rec.Parameter, rec.CurrentValue, rec.SuggestedValue)
			}
		}
	}
	return nil
}

// ApplyRecommendation 手动应用一条建议。
func (t *MetaTuner) ApplyRecommendation(ctx context.Context, recommendationID, appliedBy int64) error {
	record, err := g.DB().Model("mvp_tune_recommendation").Ctx(ctx).
		Where("id", recommendationID).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return fmt.Errorf("建议不存在: %d", recommendationID)
	}

	if record["status"].String() != "pending" {
		return fmt.Errorf("建议状态不允许应用: %s", record["status"].String())
	}

	// 宪法保护二次校验
	param := record["parameter"].String()
	if constitutionProtectedParams[param] {
		return fmt.Errorf("宪法保护参数不可修改: %s", param)
	}

	rec := &TuneRecommendation{
		ID:             record["id"].Int64(),
		Parameter:      param,
		SuggestedValue: record["suggested_value"].String(),
	}
	if err := t.applyRecommendation(ctx, rec); err != nil {
		return err
	}

	_, err = g.DB().Model("mvp_tune_recommendation").Ctx(ctx).
		Where("id", recommendationID).
		Update(g.Map{
			"status":     "applied",
			"applied_at": gtime.Now(),
			"applied_by": appliedBy,
		})
	return err
}

// RejectRecommendation 驳回一条建议。
func (t *MetaTuner) RejectRecommendation(ctx context.Context, recommendationID int64) error {
	_, err := g.DB().Model("mvp_tune_recommendation").Ctx(ctx).
		Where("id", recommendationID).
		Where("status", "pending").
		Update(g.Map{"status": "rejected"})
	return err
}

// ListPending 查询待处理的建议。
func (t *MetaTuner) ListPending(ctx context.Context, projectID int64) ([]TuneRecommendation, error) {
	m := g.DB().Model("mvp_tune_recommendation").Ctx(ctx).
		Where("status", "pending").
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if projectID > 0 {
		m = m.Where("project_id", projectID)
	}
	records, err := m.Limit(50).All()
	if err != nil {
		return nil, err
	}
	return t.parseRecommendations(records.List()), nil
}

// ListAll 查询所有建议（支持状态过滤）。
func (t *MetaTuner) ListAll(ctx context.Context, projectID int64, status string, limit int) ([]TuneRecommendation, error) {
	if limit <= 0 {
		limit = 50
	}
	m := g.DB().Model("mvp_tune_recommendation").Ctx(ctx).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		Limit(limit)
	if projectID > 0 {
		m = m.Where("project_id", projectID)
	}
	if status != "" {
		m = m.Where("status", status)
	}
	records, err := m.All()
	if err != nil {
		return nil, err
	}
	return t.parseRecommendations(records.List()), nil
}

// applyRecommendation 实际应用配置变更。
func (t *MetaTuner) applyRecommendation(ctx context.Context, rec *TuneRecommendation) error {
	// 二次宪法校验
	if constitutionProtectedParams[rec.Parameter] {
		return fmt.Errorf("宪法保护参数: %s", rec.Parameter)
	}

	// 更新 mvp_config 表
	result, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", rec.Parameter).
		Update(g.Map{
			"config_value": rec.SuggestedValue,
			"updated_at":   gtime.Now(),
		})
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("配置项不存在: %s", rec.Parameter)
	}

	return nil
}

// classifyDirection 判断建议方向。
func (t *MetaTuner) classifyDirection(drift Drift) string {
	// 降低阈值、增加限制 = conservative
	// 提高阈值、放宽限制 = aggressive
	if drift.OptimalValue < drift.CurrentValue {
		return "conservative" // 收紧
	}
	return "aggressive" // 放宽
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (t *MetaTuner) parseRecommendations(records []g.Map) []TuneRecommendation {
	var results []TuneRecommendation
	for _, r := range records {
		rec := TuneRecommendation{
			ID:             mapInt64(r, "id"),
			AssessmentID:   mapInt64(r, "assessment_id"),
			ProjectID:      mapInt64(r, "project_id"),
			Parameter:      mapString(r, "parameter"),
			CurrentValue:   mapString(r, "current_value"),
			SuggestedValue: mapString(r, "suggested_value"),
			Direction:      mapString(r, "direction"),
			Reasoning:      mapString(r, "reasoning"),
			RiskLevel:      mapString(r, "risk_level"),
			Status:         mapString(r, "status"),
		}

		if v, ok := r["confidence"]; ok && v != nil {
			rec.Confidence, _ = v.(json.Number).Float64()
			if rec.Confidence == 0 {
				// fallback
				switch n := v.(type) {
				case float64:
					rec.Confidence = n
				}
			}
		}
		if v, ok := r["auto_applicable"]; ok && v != nil {
			switch n := v.(type) {
			case int64:
				rec.AutoApplicable = n == 1
			case float64:
				rec.AutoApplicable = n == 1
			}
		}
		results = append(results, rec)
	}
	return results
}
