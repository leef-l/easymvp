package acceptance

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/repo"
)

// DecisionReducer 统一裁决归并器。
// 合并硬规则命中结果与 LLM 质量判断，产出统一决策。
type DecisionReducer struct {
	issueRepo *repo.AcceptIssueRepo
	judge     *Judge // 可选，nil 时退化为纯硬规则裁决
}

// NewDecisionReducer 创建裁决归并器。judge 为 nil 时退化为纯硬规则模式。
func NewDecisionReducer(issueRepo *repo.AcceptIssueRepo, judge *Judge) *DecisionReducer {
	return &DecisionReducer{issueRepo: issueRepo, judge: judge}
}

// Reduce 根据规则命中结果和 LLM 评审产出最终裁决。
func (r *DecisionReducer) Reduce(ctx context.Context, in *AcceptContext, hits []RuleHit, evidence []EvidenceItem) (*DecisionResult, error) {
	result := &DecisionResult{
		Decision: DecisionPassed,
		Score:    100.0,
		Issues:   hits,
	}

	// 统计各严重级别
	var blockers, errors, warns, infos int
	for _, h := range hits {
		switch h.Severity {
		case SeverityBlocker:
			blockers++
		case SeverityError:
			errors++
		case SeverityWarn:
			warns++
		case SeverityInfo:
			infos++
		}
	}

	// 裁决规则（设计文档 7.3 节）：
	// 1. 任一 blocker → failed
	// 2. error 数量 > 0 → failed
	// 3. 仅 warn/info → passed（扣分但不阻塞）
	if blockers > 0 {
		result.Decision = DecisionFailed
		result.Score = 0
	} else if errors > 0 {
		result.Decision = DecisionFailed
		result.Score = float64(50 - errors*10)
		if result.Score < 0 {
			result.Score = 0
		}
	} else {
		// passed，按 warn 扣分
		result.Score = 100.0 - float64(warns)*5
		if result.Score < 60 {
			result.Score = 60
		}
	}

	// LLM 融合（仅在 judge 可用且硬规则未直接 failed 时触发）
	var llmSummary string
	if r.judge != nil && result.Decision != DecisionFailed {
		judgeResult, judgeErr := r.judge.Evaluate(ctx, in, evidence, hits)
		if judgeErr != nil {
			g.Log().Warningf(ctx, "[DecisionReducer] LLM Judge 调用异常(降级): %v", judgeErr)
			if result.Decision == DecisionPassed {
				result.Decision = DecisionManualReview
			}
		} else {
			result.Score = result.Score*0.4 + judgeResult.QualityScore*0.6
			llmSummary = judgeResult.Summary

			switch judgeResult.Conclusion {
			case "failed":
				result.Decision = DecisionFailed
			case "uncertain":
				result.Decision = DecisionManualReview
			}
		}
	}

	// 生成摘要
	var parts []string
	if blockers > 0 {
		parts = append(parts, fmt.Sprintf("%d 个阻塞问题", blockers))
	}
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d 个错误", errors))
	}
	if warns > 0 {
		parts = append(parts, fmt.Sprintf("%d 个警告", warns))
	}
	if infos > 0 {
		parts = append(parts, fmt.Sprintf("%d 个提示", infos))
	}
	if len(parts) == 0 {
		result.Summary = "所有验收规则通过"
	} else {
		result.Summary = fmt.Sprintf("验收发现 %s", strings.Join(parts, "、"))
	}
	// 追加 LLM 评审摘要
	if llmSummary != "" {
		result.Summary += "; LLM评审: " + llmSummary
	}

	// 持久化 issue 到数据库
	if len(hits) > 0 {
		if err := r.persistIssues(ctx, in, hits); err != nil {
			g.Log().Warningf(ctx, "[DecisionReducer] 持久化 issue 失败: %v", err)
		}
	}

	g.Log().Infof(ctx, "[DecisionReducer] 裁决完成: decision=%s score=%.1f blockers=%d errors=%d warns=%d",
		result.Decision, result.Score, blockers, errors, warns)
	return result, nil
}

// persistIssues 将规则命中结果持久化为 accept_issue。
func (r *DecisionReducer) persistIssues(ctx context.Context, in *AcceptContext, hits []RuleHit) error {
	now := gtime.Now()
	var items []g.Map
	for _, h := range hits {
		items = append(items, g.Map{
			"accept_run_id":   in.AcceptRunID,
			"workflow_run_id": in.WorkflowRunID,
			"project_id":      in.ProjectID,
			"domain_task_id":  h.DomainTaskID,
			"issue_type":      h.RuleType,
			"rule_code":       h.RuleCode,
			"severity":        h.Severity,
			"title":           h.Title,
			"detail":          h.Detail,
			"expected_value":  h.ExpectedValue,
			"actual_value":    h.ActualValue,
			"suggested_action": h.SuggestedAction,
			"resource_ref":    h.ResourceRef,
			"status":          "open",
			"created_by":      in.CreatedBy,
			"dept_id":         in.DeptID,
			"created_at":      now,
			"updated_at":      now,
		})
	}
	return r.issueRepo.BatchCreate(ctx, items)
}
