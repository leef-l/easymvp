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
	config         *CircuitBreakerConfig
	decisionRepo   *repo.AutonomyDecisionRepo
	domainTaskRepo *repo.DomainTaskRepo
	handoffRepo    *repo.HandoffRecordRepo
	stageRunRepo   *repo.StageRunRepo
	acceptRunRepo  *repo.AcceptRunRepo
}

// NewCircuitBreaker 创建熔断器。
func NewCircuitBreaker(decisionRepo *repo.AutonomyDecisionRepo, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		config:         config,
		decisionRepo:   decisionRepo,
		domainTaskRepo: repo.NewDomainTaskRepo(),
		handoffRepo:    repo.NewHandoffRecordRepo(),
		stageRunRepo:   repo.NewStageRunRepo(),
		acceptRunRepo:  repo.NewAcceptRunRepo(),
	}
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
	tasks, err := cb.domainTaskRepo.ListRecentByWorkflow(ctx, workflowRunID, 20, "status")
	if err != nil || len(tasks) == 0 {
		return 0
	}

	count := 0
	for _, t := range tasks {
		s := mapString(t, "status")
		if s == "failed" || s == "escalated" {
			count++
		} else if s == "completed" {
			break // 遇到成功的就中断
		}
		// running/pending 跳过，不影响计数
	}
	return count
}

// currentBatchFailureRate 计算当前活跃批次的失败率。
func (cb *CircuitBreaker) currentBatchFailureRate(ctx context.Context, workflowRunID int64) float64 {
	batchNo, err := cb.domainTaskRepo.GetLatestBatchNoByWorkflowStatuses(ctx, workflowRunID, []string{"running", "failed", "escalated", "completed"})
	if err != nil || batchNo <= 0 {
		return 0
	}

	tasks, err := cb.domainTaskRepo.ListByWorkflowAndBatch(ctx, workflowRunID, batchNo)
	if err != nil || len(tasks) == 0 {
		return 0
	}

	total := 0
	failed := 0
	for _, t := range tasks {
		s := t.Status
		if s != "completed" && s != "failed" && s != "escalated" {
			continue
		}
		total++
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
	if cb.handoffRepo != nil {
		handoffs, err := cb.handoffRepo.ListByWorkflowAndType(ctx, workflowRunID, "failure_escalation")
		if err == nil {
			if rounds := cb.countMaxChainReworkRounds(ctx, handoffs); rounds > 0 {
				return rounds
			}
		}
	}
	count, err := cb.stageRunRepo.CountByWorkflowAndType(ctx, workflowRunID, "rework")
	if err != nil {
		return 0
	}
	return count
}

func (cb *CircuitBreaker) countMaxChainReworkRounds(ctx context.Context, handoffs []g.Map) int {
	chainCounts := countReworkRoundsByChainKey(handoffs, cb.loadTaskChainKeys(ctx, handoffs))
	maxRounds := 0
	for _, rounds := range chainCounts {
		if rounds > maxRounds {
			maxRounds = rounds
		}
	}
	return maxRounds
}

func (cb *CircuitBreaker) loadTaskChainKeys(ctx context.Context, handoffs []g.Map) map[int64]int64 {
	taskIDs := make([]int64, 0, len(handoffs))
	seen := make(map[int64]struct{}, len(handoffs))
	for _, record := range handoffs {
		taskID := mapInt64(record, "from_task_id")
		if taskID <= 0 {
			continue
		}
		if _, exists := seen[taskID]; exists {
			continue
		}
		seen[taskID] = struct{}{}
		taskIDs = append(taskIDs, taskID)
	}
	if len(taskIDs) == 0 || cb.domainTaskRepo == nil {
		return nil
	}

	records, err := cb.domainTaskRepo.ListByIDs(ctx, taskIDs, "id", "source_task_id", "root_task_id")
	if err != nil {
		return nil
	}

	result := make(map[int64]int64, len(records))
	for _, record := range records {
		taskID := mapInt64(record, "id")
		if taskID <= 0 {
			continue
		}
		rootTaskID := mapInt64(record, "root_task_id")
		if rootTaskID <= 0 {
			rootTaskID = taskID
		}
		result[taskID] = rootTaskID
	}
	return result
}

func countReworkRoundsByChainKey(handoffs []g.Map, chainKeys map[int64]int64) map[int64]int {
	counts := make(map[int64]int, len(handoffs))
	for _, record := range handoffs {
		taskID := mapInt64(record, "from_task_id")
		if taskID <= 0 {
			continue
		}
		chainKey := taskID
		if key, ok := chainKeys[taskID]; ok && key > 0 {
			chainKey = key
		}
		counts[chainKey]++
	}
	return counts
}

// countAcceptRounds 统计验收不通过轮次。
func (cb *CircuitBreaker) countAcceptRounds(ctx context.Context, workflowRunID int64) int {
	count, err := cb.acceptRunRepo.CountByWorkflowDecision(ctx, workflowRunID, "failed")
	if err != nil {
		return 0
	}
	return count
}
