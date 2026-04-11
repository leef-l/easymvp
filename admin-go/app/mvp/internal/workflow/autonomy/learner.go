package autonomy

import (
	"context"
	"math"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/repo"
)

// Learner EMA（指数移动平均）学习器。
//
// 核心职责：从观测记录中提取指标，维护 EMA 平滑值，为 Assessor 和 Tuner 提供趋势数据。
//
// 三重保护：
//  1. 样本量保护：SampleCount < MinSamples → 不输出建议
//  2. 置信度衰减：超过 DecayDays 天无新样本 → Confidence *= DecayRate
//  3. 变更幅度限制：单次调整 ≤ MaxAdjustRate 的当前值
type Learner struct {
	MinSamples    int     // 最小样本数阈值（默认 10）
	DecayDays     int     // 衰减天数（默认 7）
	DecayRate     float64 // 衰减率（默认 0.9）
	MaxAdjustRate float64 // 最大调整比例（默认 0.2 = 20%）
	DefaultAlpha  float64 // EMA 默认平滑系数（默认 0.2）
}

// NewLearner 创建学习器（使用默认参数）。
func NewLearner() *Learner {
	return &Learner{
		MinSamples:    10,
		DecayDays:     7,
		DecayRate:     0.9,
		MaxAdjustRate: 0.2,
		DefaultAlpha:  0.2,
	}
}

// IsEnabled 灰度开关。
func (l *Learner) IsEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.learner_enabled",
		"workflow.autonomy.learnerEnabled", 0) == 1
}

// LearningRecord EMA 学习记录。
type LearningRecord struct {
	ID          int64       `json:"id"`
	MetricKey   string      `json:"metricKey"`
	ProjectID   int64       `json:"projectId"`
	EMAValue    float64     `json:"emaValue"`
	RawValue    float64     `json:"rawValue"`
	SampleCount int         `json:"sampleCount"`
	LastUpdated *gtime.Time `json:"lastUpdated"`
	DecayFactor float64     `json:"decayFactor"`
}

// Update 更新一个指标的 EMA 值。
//
// metricKey 格式："{category}.{name}.{metric}"，例如 "strategy.cost_guard.accuracy"
func (l *Learner) Update(ctx context.Context, metricKey string, projectID int64, rawValue, signalWeight float64) error {
	if signalWeight <= 0 {
		return nil // 无效信号，跳过
	}

	// 读取现有记录
	record, err := repo.NewLearningRecordRepo().GetByMetric(ctx, metricKey, projectID)
	if err != nil {
		g.Log().Warningf(ctx, "[Learner] 读取学习记录失败: key=%s err=%v", metricKey, err)
		return err
	}

	now := gtime.Now()

	if len(record) == 0 {
		// 首次记录
		_, err = repo.NewLearningRecordRepo().Create(ctx, g.Map{
			"metric_key":   metricKey,
			"project_id":   projectID,
			"ema_value":    rawValue,
			"raw_value":    rawValue,
			"sample_count": 1,
			"last_updated": now,
			"decay_factor": l.DefaultAlpha,
			"created_at":   now,
		})
		return err
	}

	// 更新 EMA
	oldEMA := g.NewVar(record["ema_value"]).Float64()
	sampleCount := g.NewVar(record["sample_count"]).Int() + 1
	lastUpdated := g.NewVar(record["last_updated"]).GTime()

	// 信号加权的 EMA：alpha 乘以信号权重
	alpha := l.DefaultAlpha * signalWeight
	if alpha > 1 {
		alpha = 1
	}
	newEMA := alpha*rawValue + (1-alpha)*oldEMA

	// 置信度衰减：长时间无更新时降低 EMA 的可信度（向中性值 0.5 靠拢）
	if lastUpdated != nil {
		daysSinceUpdate := now.Sub(lastUpdated).Hours() / 24
		if daysSinceUpdate > float64(l.DecayDays) {
			decayRounds := int(daysSinceUpdate) / l.DecayDays
			for i := 0; i < decayRounds; i++ {
				newEMA = newEMA*l.DecayRate + 0.5*(1-l.DecayRate) // 向 0.5 衰减
			}
		}
	}

	return repo.NewLearningRecordRepo().UpdateByMetric(ctx, metricKey, projectID, g.Map{
		"ema_value":    newEMA,
		"raw_value":    rawValue,
		"sample_count": sampleCount,
		"last_updated": now,
	})
}

// Get 获取指标的当前 EMA 值。返回 nil 表示无记录。
func (l *Learner) Get(ctx context.Context, metricKey string, projectID int64) *LearningRecord {
	record, err := repo.NewLearningRecordRepo().GetByMetric(ctx, metricKey, projectID)
	if err != nil || len(record) == 0 {
		return nil
	}
	return &LearningRecord{
		ID:          g.NewVar(record["id"]).Int64(),
		MetricKey:   g.NewVar(record["metric_key"]).String(),
		ProjectID:   g.NewVar(record["project_id"]).Int64(),
		EMAValue:    g.NewVar(record["ema_value"]).Float64(),
		RawValue:    g.NewVar(record["raw_value"]).Float64(),
		SampleCount: g.NewVar(record["sample_count"]).Int(),
		LastUpdated: g.NewVar(record["last_updated"]).GTime(),
		DecayFactor: g.NewVar(record["decay_factor"]).Float64(),
	}
}

// HasEnoughSamples 检查样本量是否足够（三重保护之一）。
func (l *Learner) HasEnoughSamples(ctx context.Context, metricKey string, projectID int64) bool {
	rec := l.Get(ctx, metricKey, projectID)
	if rec == nil {
		return false
	}
	return rec.SampleCount >= l.MinSamples
}

// CalcConfidence 计算指标的当前置信度（考虑衰减）。
func (l *Learner) CalcConfidence(ctx context.Context, metricKey string, projectID int64) float64 {
	rec := l.Get(ctx, metricKey, projectID)
	if rec == nil {
		return 0
	}

	// 基础置信度：样本越多越高，上限 0.95
	baseConf := math.Min(float64(rec.SampleCount)/50.0, 0.95)

	// 时间衰减
	if rec.LastUpdated != nil {
		daysSinceUpdate := gtime.Now().Sub(rec.LastUpdated).Hours() / 24
		if daysSinceUpdate > float64(l.DecayDays) {
			decayRounds := int(daysSinceUpdate) / l.DecayDays
			for i := 0; i < decayRounds; i++ {
				baseConf *= l.DecayRate
			}
		}
	}

	return baseConf
}

// ClampAdjustment 限制调整幅度（三重保护之三）。
// 返回经过限幅后的建议值。
func (l *Learner) ClampAdjustment(currentValue, suggestedValue float64) float64 {
	if currentValue == 0 {
		return suggestedValue
	}
	maxDelta := math.Abs(currentValue) * l.MaxAdjustRate
	delta := suggestedValue - currentValue
	if delta > maxDelta {
		return currentValue + maxDelta
	}
	if delta < -maxDelta {
		return currentValue - maxDelta
	}
	return suggestedValue
}

// ListByProject 查询项目的所有学习记录。
func (l *Learner) ListByProject(ctx context.Context, projectID int64) ([]LearningRecord, error) {
	records, err := repo.NewLearningRecordRepo().ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var result []LearningRecord
	for _, r := range records {
		result = append(result, LearningRecord{
			ID:          g.NewVar(r["id"]).Int64(),
			MetricKey:   g.NewVar(r["metric_key"]).String(),
			ProjectID:   g.NewVar(r["project_id"]).Int64(),
			EMAValue:    g.NewVar(r["ema_value"]).Float64(),
			RawValue:    g.NewVar(r["raw_value"]).Float64(),
			SampleCount: g.NewVar(r["sample_count"]).Int(),
			LastUpdated: g.NewVar(r["last_updated"]).GTime(),
			DecayFactor: g.NewVar(r["decay_factor"]).Float64(),
		})
	}
	return result, nil
}

// FeedFromObservation 从观测记录中提取学习信号并更新 EMA。
//
// 这是 Observer → Learner 的桥梁方法，由 DecisionCenter 在观测完成后调用。
func (l *Learner) FeedFromObservation(ctx context.Context, obs *ObservationInput) {
	if obs == nil || obs.Outcome == "pending" {
		return // 未完成的观测不学习
	}

	// 1. 策略准确率：outcome=success → 1.0，failure → 0.0，neutral → 0.5
	var accuracyRaw float64
	switch obs.Outcome {
	case "success":
		accuracyRaw = 1.0
	case "failure":
		accuracyRaw = 0.0
	default:
		accuracyRaw = 0.5
	}

	metricKey := "decision." + obs.DecisionType + ".accuracy"
	_ = l.Update(ctx, metricKey, obs.ProjectID, accuracyRaw, obs.EffectScore)

	// 2. 如果有效果评分，更新效果指标
	if obs.EffectScore != 0 {
		effectKey := "decision." + obs.DecisionType + ".effect"
		normalizedScore := (obs.EffectScore + 1) / 2 // [-1,1] → [0,1]
		_ = l.Update(ctx, effectKey, obs.ProjectID, normalizedScore, 0.5)
	}

	// 3. 人工干预率
	if obs.HumanOverride {
		_ = l.Update(ctx, "system.human_override_rate", obs.ProjectID, 1.0, obs.EffectScore)
	} else {
		_ = l.Update(ctx, "system.human_override_rate", obs.ProjectID, 0.0, 0.3)
	}
}
