package autonomy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/workflow/repo"
)

// MetaAssessor 元认知评估器。
//
// 职责：定期扫描观测记录，计算系统指标，检测参数偏差。
// 只读不改参数——发现 Drift 后交给 Tuner 处理。
type MetaAssessor struct {
	observer          *MetaObserver
	learner           *Learner
	observationRepo   *repo.ObservationRecordRepo
	actionOutcomeRepo *repo.ActionOutcomeRepo
	assessmentRepo    *repo.AssessmentResultRepo
}

// NewMetaAssessor 创建评估器。
func NewMetaAssessor(observer *MetaObserver, learner *Learner) *MetaAssessor {
	return &MetaAssessor{
		observer:          observer,
		learner:           learner,
		observationRepo:   repo.NewObservationRecordRepo(),
		actionOutcomeRepo: repo.NewActionOutcomeRepo(),
		assessmentRepo:    repo.NewAssessmentResultRepo(),
	}
}

// AssessmentResult 评估结果。
type AssessmentResult struct {
	ID                int64       `json:"id"`
	ProjectID         int64       `json:"projectId"`
	PeriodStart       *gtime.Time `json:"periodStart"`
	PeriodEnd         *gtime.Time `json:"periodEnd"`
	SampleCount       int         `json:"sampleCount"`
	PolicyAccuracy    float64     `json:"policyAccuracy"`    // 策略准确率
	GateFalsePositive float64     `json:"gateFalsePositive"` // 闸门误报率
	GateFalseNegative float64     `json:"gateFalseNegative"` // 闸门漏报率
	HumanOverrideRate float64     `json:"humanOverrideRate"` // 人工干预率
	MatchAccuracy     float64     `json:"matchAccuracy"`     // 匹配准确率
	CostEfficiency    float64     `json:"costEfficiency"`    // 成本效率
	Drifts            []Drift     `json:"drifts"`            // 参数偏差
	Summary           string      `json:"summary"`           // 评估摘要
}

// Drift 参数偏差。
type Drift struct {
	Parameter    string  `json:"parameter"`
	CurrentValue float64 `json:"currentValue"`
	OptimalValue float64 `json:"optimalValue"`
	Confidence   float64 `json:"confidence"`
	Evidence     string  `json:"evidence"`
}

// Assess 执行一次评估，覆盖指定周期。
//
// projectID=0 表示全局评估。
func (a *MetaAssessor) Assess(ctx context.Context, projectID int64, periodStart, periodEnd *gtime.Time) (*AssessmentResult, error) {
	// 1. 统计观测结果分布
	outcomeCounts, err := a.observer.CountByOutcome(ctx, projectID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("统计观测结果失败: %w", err)
	}

	totalSamples := 0
	for _, c := range outcomeCounts {
		totalSamples += c
	}

	if totalSamples == 0 {
		return &AssessmentResult{
			ProjectID:   projectID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Summary:     "周期内无观测记录",
		}, nil
	}

	// 2. 计算核心指标
	successCount := outcomeCounts["success"]
	failureCount := outcomeCounts["failure"]

	policyAccuracy := float64(successCount) / float64(totalSamples)

	// 3. 人工干预率（从 Learner EMA 获取）
	humanOverrideRate := 0.0
	if rec := a.learner.Get(ctx, "system.human_override_rate", projectID); rec != nil {
		humanOverrideRate = rec.EMAValue
	}

	// 4. 闸门误报/漏报率（从观测记录中的闸门相关决策计算）
	gateFP, gateFN := a.calcGateMetrics(ctx, projectID, periodStart, periodEnd)

	// 5. 成本效率（从 action_outcome 效果得分计算）
	costEfficiency := a.calcCostEfficiency(ctx, projectID, periodStart, periodEnd)

	// 6. 检测参数偏差
	drifts := a.detectDrifts(ctx, projectID, policyAccuracy, humanOverrideRate, gateFP)

	// 7. 生成摘要
	summary := fmt.Sprintf(
		"周期 %s ~ %s：共 %d 条观测，成功 %d / 失败 %d / 中性 %d / 待定 %d。策略准确率 %.1f%%，人工干预率 %.1f%%，闸门误报率 %.1f%%",
		periodStart.Format("Y-m-d"), periodEnd.Format("Y-m-d"),
		totalSamples, successCount, failureCount, outcomeCounts["neutral"], outcomeCounts["pending"],
		policyAccuracy*100, humanOverrideRate*100, gateFP*100,
	)
	if len(drifts) > 0 {
		summary += fmt.Sprintf("，检测到 %d 个参数偏差", len(drifts))
	}

	result := &AssessmentResult{
		ProjectID:         projectID,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		SampleCount:       totalSamples,
		PolicyAccuracy:    policyAccuracy,
		GateFalsePositive: gateFP,
		GateFalseNegative: gateFN,
		HumanOverrideRate: humanOverrideRate,
		MatchAccuracy:     policyAccuracy, // 初期等同策略准确率
		CostEfficiency:    costEfficiency,
		Drifts:            drifts,
		Summary:           summary,
	}

	// 8. 持久化评估结果
	if err := a.save(ctx, result); err != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] 保存评估结果失败: %v", err)
	}

	return result, nil
}

// calcGateMetrics 计算闸门误报/漏报率。
func (a *MetaAssessor) calcGateMetrics(ctx context.Context, projectID int64, periodStart, periodEnd *gtime.Time) (falsePositive, falseNegative float64) {
	// 闸门误报：闸门阻断了，但人工批准放行（说明不该拦）
	// 闸门漏报：未被闸门拦截，但最终失败（说明应该拦）
	// 误报：decision_level=C 且 human_override=1 且 override 动作=approve
	fpCount, fpErr := a.observationRepo.CountByFiltersInPeriod(ctx, projectID, periodStart, periodEnd, g.Map{
		"decision_level": "C",
		"human_override": 1,
		"outcome":        "success",
	})
	if fpErr != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] 查询闸门误报数失败: %v", fpErr)
	}

	totalGateBlocked, gbErr := a.observationRepo.CountByFiltersInPeriod(ctx, projectID, periodStart, periodEnd, g.Map{
		"decision_level": "C",
	})
	if gbErr != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] 查询闸门阻断总数失败: %v", gbErr)
	}

	if totalGateBlocked > 0 {
		falsePositive = float64(fpCount) / float64(totalGateBlocked)
	}

	// 漏报：decision_level=A 且 outcome=failure（自动执行但失败了）
	fnCount, fnErr := a.observationRepo.CountByFiltersInPeriod(ctx, projectID, periodStart, periodEnd, g.Map{
		"decision_level": "A",
		"outcome":        "failure",
	})
	if fnErr != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] 查询闸门漏报数失败: %v", fnErr)
	}

	totalAutoExec, aeErr := a.observationRepo.CountByFiltersInPeriod(ctx, projectID, periodStart, periodEnd, g.Map{
		"decision_level": "A",
	})
	if aeErr != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] 查询自动执行总数失败: %v", aeErr)
	}

	if totalAutoExec > 0 {
		falseNegative = float64(fnCount) / float64(totalAutoExec)
	}

	return falsePositive, falseNegative
}

// calcCostEfficiency 计算成本效率。
func (a *MetaAssessor) calcCostEfficiency(ctx context.Context, projectID int64, periodStart, periodEnd *gtime.Time) float64 {
	avgScore, count, err := a.actionOutcomeRepo.GetAverageEffectScore(ctx, projectID, periodStart, periodEnd)
	if err != nil || count == 0 {
		return 0.5 // 无数据时返回中性值
	}
	// 将 [-1,1] 映射到 [0,1]
	return (avgScore + 1) / 2
}

// detectDrifts 检测参数偏差。
func (a *MetaAssessor) detectDrifts(ctx context.Context, projectID int64, policyAccuracy, humanOverrideRate, gateFP float64) []Drift {
	var drifts []Drift

	// 1. 策略准确率偏低（< 0.7）→ 建议提高闸门敏感度
	if policyAccuracy < 0.7 && a.learner.HasEnoughSamples(ctx, "system.human_override_rate", projectID) {
		drifts = append(drifts, Drift{
			Parameter:    "workflow.autonomy.risk_gate_enabled",
			CurrentValue: 1,
			OptimalValue: 1, // 已开启，但可能需要调整规则
			Confidence:   a.learner.CalcConfidence(ctx, "system.human_override_rate", projectID),
			Evidence:     fmt.Sprintf("策略准确率 %.1f%% 低于阈值 70%%", policyAccuracy*100),
		})
	}

	// 2. 人工干预率过高（> 0.5）→ 建议放宽自动执行条件
	if humanOverrideRate > 0.5 {
		conf := a.learner.CalcConfidence(ctx, "system.human_override_rate", projectID)
		if conf > 0.3 {
			drifts = append(drifts, Drift{
				Parameter:    "workflow.autonomy.audit_only",
				CurrentValue: 1,
				OptimalValue: 0,
				Confidence:   conf,
				Evidence:     fmt.Sprintf("人工干预率 %.1f%% 过高，系统可能过于保守", humanOverrideRate*100),
			})
		}
	}

	// 3. 闸门误报率过高（> 0.3）→ 建议降低闸门敏感度
	if gateFP > 0.3 {
		drifts = append(drifts, Drift{
			Parameter:    "gate_sensitivity",
			CurrentValue: gateFP,
			OptimalValue: 0.1,
			Confidence:   0.6,
			Evidence:     fmt.Sprintf("闸门误报率 %.1f%% 过高，正常操作被过度拦截", gateFP*100),
		})
	}

	// 4. 从 Learner 中检测各策略效果偏差
	learningRecords, lrErr := a.learner.ListByProject(ctx, projectID)
	if lrErr != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] 获取学习记录失败: projectID=%d err=%v", projectID, lrErr)
	}
	for _, rec := range learningRecords {
		if rec.SampleCount < a.learner.MinSamples {
			continue
		}
		// 效果指标低于 0.4（映射回 [-1,1] 就是 < -0.2）→ 策略可能有问题
		if rec.EMAValue < 0.4 && rec.MetricKey != "system.human_override_rate" {
			drifts = append(drifts, Drift{
				Parameter:    rec.MetricKey,
				CurrentValue: rec.EMAValue,
				OptimalValue: 0.6,
				Confidence:   a.learner.CalcConfidence(ctx, rec.MetricKey, projectID),
				Evidence:     fmt.Sprintf("指标 %s EMA=%.3f 低于阈值 0.4，样本数 %d", rec.MetricKey, rec.EMAValue, rec.SampleCount),
			})
		}
	}

	return drifts
}

// save 持久化评估结果。
func (a *MetaAssessor) save(ctx context.Context, result *AssessmentResult) error {
	driftsJSON, marshalErr := json.Marshal(result.Drifts)
	if marshalErr != nil {
		g.Log().Warningf(ctx, "[MetaAssessor] drifts 序列化失败: %v", marshalErr)
		driftsJSON = []byte("[]")
	}

	id, err := a.assessmentRepo.Create(ctx, g.Map{
		"project_id":          result.ProjectID,
		"period_start":        result.PeriodStart,
		"period_end":          result.PeriodEnd,
		"sample_count":        result.SampleCount,
		"policy_accuracy":     result.PolicyAccuracy,
		"gate_false_positive": result.GateFalsePositive,
		"gate_false_negative": result.GateFalseNegative,
		"human_override_rate": result.HumanOverrideRate,
		"match_accuracy":      result.MatchAccuracy,
		"cost_efficiency":     result.CostEfficiency,
		"drifts":              string(driftsJSON),
		"summary":             result.Summary,
		"created_at":          gtime.Now(),
	})
	if err == nil {
		result.ID = id
	}
	return err
}

// GetLatest 获取最新的评估结果。
func (a *MetaAssessor) GetLatest(ctx context.Context, projectID int64) (*AssessmentResult, error) {
	record, err := a.assessmentRepo.GetLatestByProject(ctx, projectID)
	if err != nil || record == nil {
		return nil, err
	}

	result := a.parseAssessmentRecord(ctx, record)
	return &result, nil
}

// ListAssessments 查询评估历史。
func (a *MetaAssessor) ListAssessments(ctx context.Context, projectID int64, limit int) ([]AssessmentResult, error) {
	if limit <= 0 {
		limit = 20
	}
	records, err := a.assessmentRepo.ListByProject(ctx, projectID, limit)
	if err != nil {
		return nil, err
	}

	var results []AssessmentResult
	for _, record := range records {
		results = append(results, a.parseAssessmentRecord(ctx, record))
	}
	return results, nil
}

func (a *MetaAssessor) parseAssessmentRecord(ctx context.Context, record g.Map) AssessmentResult {
	result := AssessmentResult{
		ID:                gconv.Int64(record["id"]),
		ProjectID:         gconv.Int64(record["project_id"]),
		PeriodStart:       g.NewVar(record["period_start"]).GTime(),
		PeriodEnd:         g.NewVar(record["period_end"]).GTime(),
		SampleCount:       gconv.Int(record["sample_count"]),
		PolicyAccuracy:    gconv.Float64(record["policy_accuracy"]),
		GateFalsePositive: gconv.Float64(record["gate_false_positive"]),
		GateFalseNegative: gconv.Float64(record["gate_false_negative"]),
		HumanOverrideRate: gconv.Float64(record["human_override_rate"]),
		MatchAccuracy:     gconv.Float64(record["match_accuracy"]),
		CostEfficiency:    gconv.Float64(record["cost_efficiency"]),
		Summary:           gconv.String(record["summary"]),
	}

	driftsStr := gconv.String(record["drifts"])
	if driftsStr != "" && driftsStr != "null" {
		if unmErr := json.Unmarshal([]byte(driftsStr), &result.Drifts); unmErr != nil {
			g.Log().Warningf(ctx, "[MetaAssessor] drifts JSON 解析失败: %v", unmErr)
		}
	}

	return result
}
