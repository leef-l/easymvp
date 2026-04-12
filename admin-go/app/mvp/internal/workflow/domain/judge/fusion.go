package judge

import (
	"fmt"
	"strings"
)

// Producer 判官来源类型。
type Producer string

const (
	ProducerHardRule           Producer = "hard_rule"
	ProducerLLMJudge           Producer = "llm_judge"
	ProducerExperienceReviewer Producer = "experience_reviewer"
	ProducerHuman              Producer = "human"
)

// JudgeInput 单路判官的输入。
type JudgeInput struct {
	Producer Producer
	Score    float64  // 0-100
	Decision string   // passed/failed/uncertain
	Findings []Finding
}

// Finding 单条发现。
type Finding struct {
	Severity string // blocker/error/warn/info
	Code     string
	Message  string
}

// FusionResult 融合结果。
type FusionResult struct {
	FinalScore    float64
	FinalDecision string // passed/failed/manual_review
	Summary       string
	HasBlocker    bool
}

// DefaultWeights 默认权重配置。
var DefaultWeights = map[Producer]float64{
	ProducerHardRule:           0.30,
	ProducerLLMJudge:           0.30,
	ProducerExperienceReviewer: 0.40,
}

// Fuse 三路融合纯函数。
// 规则：
// 1. 任何一路有 blocker 发现 → 直接 failed（否决权）
// 2. 任何一路 decision=failed → 直接 failed
// 3. 加权计算 FinalScore（按实际提供路重新归一化权重）
// 4. FinalScore >= 60 → passed；否则 → manual_review
// 5. 任何一路 uncertain → manual_review（除非已 failed）
func Fuse(inputs []JudgeInput, weights map[Producer]float64) *FusionResult {
	if weights == nil {
		weights = DefaultWeights
	}

	result := &FusionResult{}

	var hasUncertain bool
	var hasFailed bool
	var summaryParts []string

	// 第一轮：检查否决条件 + 收集 uncertain/failed
	for _, inp := range inputs {
		for _, f := range inp.Findings {
			if strings.EqualFold(f.Severity, "blocker") {
				result.HasBlocker = true
			}
		}
		if strings.EqualFold(inp.Decision, "failed") {
			hasFailed = true
		}
		if strings.EqualFold(inp.Decision, "uncertain") {
			hasUncertain = true
		}
	}

	// 否决权：有 blocker 或任一路 failed → 直接 failed
	if result.HasBlocker || hasFailed {
		result.FinalDecision = "failed"
		if result.HasBlocker {
			summaryParts = append(summaryParts, "blocker findings detected")
		}
		if hasFailed {
			summaryParts = append(summaryParts, "one or more judges returned failed")
		}
		result.Summary = strings.Join(summaryParts, "; ")
		// 仍然计算加权分（供参考）
		result.FinalScore = computeWeightedScore(inputs, weights)
		return result
	}

	// 第二轮：加权计算分数
	result.FinalScore = computeWeightedScore(inputs, weights)

	// 决策：分数 + uncertain
	if result.FinalScore >= 60 && !hasUncertain {
		result.FinalDecision = "passed"
		result.Summary = fmt.Sprintf("all judges passed, weighted score=%.1f", result.FinalScore)
	} else {
		result.FinalDecision = "manual_review"
		if hasUncertain {
			summaryParts = append(summaryParts, "uncertain decision from at least one judge")
		}
		if result.FinalScore < 60 {
			summaryParts = append(summaryParts, fmt.Sprintf("weighted score %.1f below threshold 60", result.FinalScore))
		}
		result.Summary = strings.Join(summaryParts, "; ")
	}

	return result
}

// computeWeightedScore 按实际提供路归一化后计算加权分。
func computeWeightedScore(inputs []JudgeInput, weights map[Producer]float64) float64 {
	var totalWeight float64
	var weightedSum float64

	for _, inp := range inputs {
		w, ok := weights[inp.Producer]
		if !ok || w <= 0 {
			continue
		}
		totalWeight += w
		weightedSum += inp.Score * w
	}

	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}
