package autonomy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

// CircuitBreakerConfig 熔断器配置。
type CircuitBreakerConfig struct {
	ConsecutiveFailures int     // 连续失败任务数阈值，默认 5
	BatchFailureRate    float64 // 批次失败率阈值，默认 0.6
	MaxReworkRounds     int     // 同任务最大 rework 轮次，默认 3
	MaxAcceptRounds     int     // 验收最大轮次，默认 3
	MaxReplanCount      int     // 最大重规划次数，默认 2
}

// DefaultCircuitBreakerConfig 默认配置。
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		ConsecutiveFailures: 5,
		BatchFailureRate:    0.6,
		MaxReworkRounds:     3,
		MaxAcceptRounds:     3,
		MaxReplanCount:      2,
	}
}

// CircuitBreaker 熔断器：检测项目级异常，触发保护性暂停。
type CircuitBreaker struct {
	config       *CircuitBreakerConfig
	decisionRepo *repo.AutonomyDecisionRepo
}

// NewCircuitBreaker 创建熔断器。
func NewCircuitBreaker(decisionRepo *repo.AutonomyDecisionRepo, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{config: config, decisionRepo: decisionRepo}
}

// Check 检查项目���否应该触发熔断。
func (cb *CircuitBreaker) Check(ctx context.Context, workflowRunID int64) *CircuitBreakResult {
	metrics := &BreakMetrics{}

	// 检查 1：连续失败任务数
	consecutiveFailures := cb.countConsecutiveFailures(ctx, workflowRunID)
	metrics.ConsecutiveFailures = consecutiveFailures
	if consecutiveFailures >= cb.config.ConsecutiveFailures {
		return &CircuitBreakResult{
			ShouldBreak: true,
			Reason:      fmt.Sprintf("连续 %d 个任务失败（��值 %d）", consecutiveFailures, cb.config.ConsecutiveFailures),
			Metrics:     metrics,
		}
	}

	// 检查 2：当前活跃批次失败率
	batchFailRate := cb.currentBatchFailureRate(ctx, workflowRunID)
	metrics.BatchFailureRate = batchFailRate
	if batchFailRate >= cb.config.BatchFailureRate {
		return &CircuitBreakResult{
			ShouldBreak: true,
			Reason:      fmt.Sprintf("当前批次失败率 %.0f%%（阈值 %.0f%%）", batchFailRate*100, cb.config.BatchFailureRate*100),
			Metrics:     metrics,
		}
	}

	// 检查 3：rework 轮次
	reworkRounds := cb.countReworkRounds(ctx, workflowRunID)
	metrics.ReworkRounds = reworkRounds
	if reworkRounds >= cb.config.MaxReworkRounds {
		return &CircuitBreakResult{
			ShouldBreak: true,
			Reason:      fmt.Sprintf("返工已达 %d 轮（上限 %d）", reworkRounds, cb.config.MaxReworkRounds),
			Metrics:     metrics,
		}
	}

	// 检查 4：accept 轮次
	acceptRounds := cb.countAcceptRounds(ctx, workflowRunID)
	metrics.AcceptRounds = acceptRounds
	if acceptRounds >= cb.config.MaxAcceptRounds {
		return &CircuitBreakResult{
			ShouldBreak: true,
			Reason:      fmt.Sprintf("验收已达 %d 轮且未通过（上限 %d）", acceptRounds, cb.config.MaxAcceptRounds),
			Metrics:     metrics,
		}
	}

	return &CircuitBreakResult{ShouldBreak: false, Metrics: metrics}
}

// RecordBreak 记录熔断决策。熔断暂停始终执行，但决策记录模式跟随全局配置。
func (cb *CircuitBreaker) RecordBreak(ctx context.Context, workflowRunID, projectID int64, result *CircuitBreakResult) (int64, error) {
	metricsJSON, mErr := json.Marshal(result.Metrics)
	if mErr != nil {
		metricsJSON = []byte("{}")
	}
	recommendation, rErr := json.Marshal(map[string]interface{}{
		"action": "pause",
		"reason": result.Reason,
	})
	if rErr != nil {
		recommendation = []byte(`{"action":"pause"}`)
	}

	// 熔断暂停本身始终自动执行（保护性措施），但记录中标注当前模式供前端展示
	mode := GetAutonomyMode(ctx)
	humanAction := ActionApproved
	if mode == ModeSuggest {
		humanAction = ActionPending // suggest 模式下标记为待审核，让用户知道有熔断发生
	}

	return cb.decisionRepo.Create(ctx, g.Map{
		"workflow_run_id": workflowRunID,
		"project_id":      projectID,
		"decision_type":   DecisionCircuitBreak,
		"trigger_source":  "watchdog",
		"trigger_context": string(metricsJSON),
		"recommendation":  string(recommendation),
		"decision_mode":   mode,
		"human_action":    humanAction,
	})
}

// countConsecutiveFailures 统计最近连续失败任务数。
func (cb *CircuitBreaker) countConsecutiveFailures(ctx context.Context, workflowRunID int64) int {
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("updated_at").
		Limit(20).
		Fields("status").
		All()
	if err != nil || tasks.IsEmpty() {
		return 0
	}

	count := 0
	for _, t := range tasks {
		s := t["status"].String()
		if s == "failed" || s == "escalated" {
			count++
		} else if s == "completed" {
			break // 遇到成功的就中断
		}
		// running/pending 跳过��影响计数
	}
	return count
}

// currentBatchFailureRate 计算当前活跃批次的失败率。
func (cb *CircuitBreaker) currentBatchFailureRate(ctx context.Context, workflowRunID int64) float64 {
	// 获取当前活跃批次号
	batchRecord, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereIn("status", g.Slice{"running", "failed", "escalated", "completed"}).
		WhereNull("deleted_at").
		Fields("batch_no").
		Group("batch_no").
		OrderDesc("batch_no").
		One()
	if err != nil || batchRecord.IsEmpty() {
		return 0
	}
	batchNo := batchRecord["batch_no"].Int()

	// 统计该批次的完成和失败
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("batch_no", batchNo).
		WhereIn("status", g.Slice{"completed", "failed", "escalated"}).
		WhereNull("deleted_at").
		Fields("status").
		All()
	if err != nil || tasks.IsEmpty() {
		return 0
	}

	total := len(tasks)
	failed := 0
	for _, t := range tasks {
		s := t["status"].String()
		if s == "failed" || s == "escalated" {
			failed++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(failed) / float64(total)
}

// countReworkRounds 统计 rework 阶段轮次。
func (cb *CircuitBreaker) countReworkRounds(ctx context.Context, workflowRunID int64) int {
	count, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", "rework").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return 0
	}
	return count
}

// countAcceptRounds 统计验收不通过轮次。
func (cb *CircuitBreaker) countAcceptRounds(ctx context.Context, workflowRunID int64) int {
	count, err := g.DB().Model("mvp_accept_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("decision", "failed").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return 0
	}
	return count
}
