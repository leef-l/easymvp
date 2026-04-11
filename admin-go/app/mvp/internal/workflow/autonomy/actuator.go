package autonomy

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/workflow/repo"
)

// Actuator 策略效果跟踪器。
//
// 职责：
//  1. 记录策略执行前的态势快照
//  2. 在延迟窗口后评估策略的实际效果
//  3. 为学习闭环（Phase D）积累数据
type Actuator struct {
	outcomeRepo *repo.ActionOutcomeRepo
}

// NewActuator 创建效果跟踪器。
func NewActuator() *Actuator {
	return &Actuator{
		outcomeRepo: repo.NewActionOutcomeRepo(),
	}
}

// ActionOutcome 策略执行效果记录。
type ActionOutcome struct {
	ID            int64                  `json:"id"`
	ActionID      int64                  `json:"actionId"`
	WorkflowRunID int64                  `json:"workflowRunId"`
	ProjectID     int64                  `json:"projectId"`
	StrategyName  string                 `json:"strategyName"`
	ActionType    string                 `json:"actionType"`
	DecisionLevel string                 `json:"decisionLevel"`
	SitBefore     map[string]interface{} `json:"sitBefore"`   // 执行前态势摘要
	SitAfter      map[string]interface{} `json:"sitAfter"`    // 执行后态势摘要
	Effective     string                 `json:"effective"`   // positive / negative / neutral / unknown
	EffectScore   float64                `json:"effectScore"` // -1 ~ 1
	EvalDelayMs   int64                  `json:"evalDelayMs"` // 评估延迟（毫秒）
	CreatedBy     int64                  `json:"createdBy"`
	DeptID        int64                  `json:"deptId"`
	CreatedAt     *gtime.Time            `json:"createdAt"`
}

// RecordBefore 策略执行前记录态势快照，返回 outcome ID 供后续关联。
func (a *Actuator) RecordBefore(ctx context.Context, actionID, wfRunID, projectID int64, strategyName, actionType, decisionLevel string, sit *Situation, createdBy, deptID int64) int64 {
	sitSummary := a.summarizeSituation(sit)

	id, err := a.outcomeRepo.Create(ctx, g.Map{
		"action_id":       actionID,
		"workflow_run_id": wfRunID,
		"project_id":      projectID,
		"strategy_name":   strategyName,
		"action_type":     actionType,
		"decision_level":  decisionLevel,
		"sit_before":      sitSummary,
		"effective":       "unknown",
		"effect_score":    0,
		"created_by":      createdBy,
		"dept_id":         deptID,
		"created_at":      gtime.Now(),
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Actuator] RecordBefore 失败: %v", err)
		return 0
	}
	return id
}

// EvaluateAfter 策略执行后评估效果。
// 对比执行前后的态势快照，计算效果得分。
func (a *Actuator) EvaluateAfter(ctx context.Context, outcomeID int64, sitAfter *Situation) {
	if outcomeID == 0 {
		return
	}

	// 读取执行前记录
	record, err := a.outcomeRepo.GetByID(ctx, outcomeID, "sit_before", "created_at")
	if err != nil || record == nil {
		g.Log().Warningf(ctx, "[Actuator] EvaluateAfter 未找到记录: id=%d", outcomeID)
		return
	}

	sitAfterSummary := a.summarizeSituation(sitAfter)
	effective, score := a.compareOutcome(record, sitAfter)

	var evalDelay int64
	if ct := g.NewVar(record["created_at"]).GTime(); ct != nil {
		evalDelay = gtime.Now().TimestampMilli() - ct.TimestampMilli()
	}

	err = a.outcomeRepo.UpdateFields(ctx, outcomeID, g.Map{
		"sit_after":     sitAfterSummary,
		"effective":     effective,
		"effect_score":  score,
		"eval_delay_ms": evalDelay,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Actuator] EvaluateAfter 更新失败: %v", err)
	}
}

// compareOutcome 对比执行前后态势，评估策略效果。
func (a *Actuator) compareOutcome(beforeRecord map[string]interface{}, sitAfter *Situation) (string, float64) {
	if sitAfter == nil || sitAfter.Health == nil {
		return "unknown", 0
	}

	// 从 before 记录提取关键指标
	beforeJSON := parseJSONMap(gconv.String(beforeRecord["sit_before"]))
	beforeFailureRate, _ := beforeJSON["failureRate"].(float64)
	beforeConsecutive, _ := beforeJSON["consecutiveFailures"].(float64)

	afterFailureRate := sitAfter.Health.RecentFailureRate
	afterConsecutive := float64(sitAfter.Health.ConsecutiveFailures)

	// 评分逻辑：失败率下降 → positive，上升 → negative
	score := 0.0
	if beforeFailureRate > 0 {
		delta := beforeFailureRate - afterFailureRate
		score += delta * 2 // 失败率每降 10% 得 0.2 分
	}
	if beforeConsecutive > afterConsecutive {
		score += 0.2 // 连续失败减少
	} else if afterConsecutive > beforeConsecutive {
		score -= 0.2
	}

	// 归一化到 [-1, 1]
	if score > 1 {
		score = 1
	} else if score < -1 {
		score = -1
	}

	effective := "neutral"
	if score > 0.1 {
		effective = "positive"
	} else if score < -0.1 {
		effective = "negative"
	}

	return effective, score
}

// summarizeSituation 提取态势关键指标用于存储。
func (a *Actuator) summarizeSituation(sit *Situation) string {
	if sit == nil {
		return "{}"
	}
	m := g.Map{
		"completionRate":      0.0,
		"failureRate":         0.0,
		"consecutiveFailures": 0,
		"retryCount":          0,
		"reworkRounds":        0,
		"tokensConsumed":      int64(0),
	}
	if sit.Progress != nil {
		m["completionRate"] = sit.Progress.CompletionRate
	}
	if sit.Health != nil {
		m["failureRate"] = sit.Health.RecentFailureRate
		m["consecutiveFailures"] = sit.Health.ConsecutiveFailures
		m["retryCount"] = sit.Health.RetryCount
		m["reworkRounds"] = sit.Health.ReworkRounds
	}
	if sit.Resource != nil {
		m["tokensConsumed"] = sit.Resource.TokensConsumed
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
