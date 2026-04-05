package autonomy

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// RiskAssessor 风险评估器，根据错误模式判断风险级别。
type RiskAssessor struct{}

// NewRiskAssessor 创建风险评估器。
func NewRiskAssessor() *RiskAssessor { return &RiskAssessor{} }

// Assess 评估单个任务失败的风险级别。
func (a *RiskAssessor) Assess(ctx context.Context, input *RiskInput) *RiskResult {
	errLower := strings.ToLower(input.ErrorMessage)

	// 规则 1：重试次数过多 → 从 transient 升级
	if input.RetryCount >= 3 {
		return &RiskResult{
			Level:      RiskStructural,
			Confidence: 0.8,
			Reason:     "任务已重试 3 次仍失败，判定为结构性问题",
			Action:     "rework",
		}
	}

	// 规则 2：致命错误模式匹配
	fatalPatterns := []string{
		"permission denied", "access denied",
		"disk full", "no space left",
		"git conflict", "merge conflict",
		"repository not found", "work_dir",
		"fatal:", "panic:",
	}
	for _, p := range fatalPatterns {
		if strings.Contains(errLower, p) {
			return &RiskResult{
				Level:      RiskFatal,
				Confidence: 0.9,
				Reason:     "匹配致命错误模式: " + p,
				Action:     "pause",
			}
		}
	}

	// 规则 3：结构性错误模式匹配
	structuralPatterns := []string{
		"import not found", "undefined reference", "type mismatch",
		"module not found", "package not found", "cannot find",
		"compilation failed", "build failed", "syntax error",
		"interface not implemented", "missing method",
		"circular dependency", "incompatible",
	}
	for _, p := range structuralPatterns {
		if strings.Contains(errLower, p) {
			return &RiskResult{
				Level:      RiskStructural,
				Confidence: 0.7,
				Reason:     "匹配结构性错误模式: " + p,
				Action:     "rework",
			}
		}
	}

	// 规则 4：瞬态错误模式匹配
	transientPatterns := []string{
		"timeout", "timed out", "deadline exceeded",
		"rate limit", "too many requests", "429",
		"connection refused", "connection reset",
		"temporary failure", "service unavailable", "503",
		"network", "dns",
	}
	for _, p := range transientPatterns {
		if strings.Contains(errLower, p) {
			return &RiskResult{
				Level:      RiskTransient,
				Confidence: 0.85,
				Reason:     "匹配瞬态错误模式: " + p,
				Action:     "retry",
			}
		}
	}

	// 默认：未知错误视为结构性（保守策略）
	g.Log().Infof(ctx, "[RiskAssessor] 未匹配已知模式，保守判定为 structural: task=%d err=%s",
		input.TaskID, truncate(input.ErrorMessage, 200))

	return &RiskResult{
		Level:      RiskStructural,
		Confidence: 0.4,
		Reason:     "未匹配已知错误模式，保守判定",
		Action:     "rework",
	}
}

// AssessBatch 评估批次级别风险。
func (a *RiskAssessor) AssessBatch(ctx context.Context, workflowRunID int64, batchNo int) *RiskResult {
	// 查批次任务统计
	var total, failed int
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("batch_no", batchNo).
		WhereNull("deleted_at").
		All()
	if err != nil || tasks.IsEmpty() {
		return &RiskResult{Level: RiskTransient, Confidence: 0.3, Reason: "无法获取批次任务", Action: "retry"}
	}

	total = len(tasks)
	for _, t := range tasks {
		s := t["status"].String()
		if s == "failed" || s == "escalated" {
			failed++
		}
	}

	if total == 0 {
		return &RiskResult{Level: RiskTransient, Confidence: 0.3, Reason: "批次无任务", Action: "retry"}
	}

	failRate := float64(failed) / float64(total)

	if failRate >= 0.8 {
		return &RiskResult{
			Level:      RiskFatal,
			Confidence: 0.9,
			Reason:     "批次失败率 ≥80%，方案可能不可行",
			Action:     "replan",
		}
	}
	if failRate >= 0.5 {
		return &RiskResult{
			Level:      RiskStructural,
			Confidence: 0.75,
			Reason:     "批次失败率 ≥50%，存在系统性问题",
			Action:     "rework",
		}
	}

	return &RiskResult{
		Level:      RiskTransient,
		Confidence: 0.6,
		Reason:     "批次失败率正常",
		Action:     "retry",
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
