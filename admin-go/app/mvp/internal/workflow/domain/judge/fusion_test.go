package judge

import (
	"testing"
)

// 场景1：三路全部 passed，高分 → passed
func TestFuse_AllPassed_HighScore(t *testing.T) {
	inputs := []JudgeInput{
		{Producer: ProducerHardRule, Score: 90, Decision: "passed"},
		{Producer: ProducerLLMJudge, Score: 85, Decision: "passed"},
		{Producer: ProducerExperienceReviewer, Score: 80, Decision: "passed"},
	}
	result := Fuse(inputs, DefaultWeights)

	if result.FinalDecision != "passed" {
		t.Errorf("expected passed, got %s", result.FinalDecision)
	}
	if result.HasBlocker {
		t.Error("expected no blocker")
	}
	// 加权分：90*0.3 + 85*0.3 + 80*0.4 = 27+25.5+32 = 84.5
	want := 84.5
	if result.FinalScore < want-0.01 || result.FinalScore > want+0.01 {
		t.Errorf("expected score=%.1f, got %.2f", want, result.FinalScore)
	}
}

// 场景2：hard_rule 有 blocker → failed（不管其他两路分数多高）
func TestFuse_HardRule_Blocker(t *testing.T) {
	inputs := []JudgeInput{
		{
			Producer: ProducerHardRule,
			Score:    40,
			Decision: "failed",
			Findings: []Finding{
				{Severity: "blocker", Code: "NULL_DEREF", Message: "null pointer dereference"},
			},
		},
		{Producer: ProducerLLMJudge, Score: 95, Decision: "passed"},
		{Producer: ProducerExperienceReviewer, Score: 95, Decision: "passed"},
	}
	result := Fuse(inputs, DefaultWeights)

	if result.FinalDecision != "failed" {
		t.Errorf("expected failed, got %s", result.FinalDecision)
	}
	if !result.HasBlocker {
		t.Error("expected HasBlocker=true")
	}
}

// 场景3：llm_judge decision=failed → failed
func TestFuse_LLMJudge_Failed(t *testing.T) {
	inputs := []JudgeInput{
		{Producer: ProducerHardRule, Score: 85, Decision: "passed"},
		{Producer: ProducerLLMJudge, Score: 30, Decision: "failed"},
		{Producer: ProducerExperienceReviewer, Score: 88, Decision: "passed"},
	}
	result := Fuse(inputs, DefaultWeights)

	if result.FinalDecision != "failed" {
		t.Errorf("expected failed, got %s", result.FinalDecision)
	}
	if result.HasBlocker {
		t.Error("expected no blocker")
	}
}

// 场景4：三路 passed 但加权分低于 60 → manual_review
func TestFuse_AllPassed_LowScore(t *testing.T) {
	inputs := []JudgeInput{
		{Producer: ProducerHardRule, Score: 55, Decision: "passed"},
		{Producer: ProducerLLMJudge, Score: 50, Decision: "passed"},
		{Producer: ProducerExperienceReviewer, Score: 45, Decision: "passed"},
	}
	result := Fuse(inputs, DefaultWeights)

	if result.FinalDecision != "manual_review" {
		t.Errorf("expected manual_review, got %s", result.FinalDecision)
	}
	// 加权分：55*0.3 + 50*0.3 + 45*0.4 = 16.5+15+18 = 49.5
	want := 49.5
	if result.FinalScore < want-0.01 || result.FinalScore > want+0.01 {
		t.Errorf("expected score=%.1f, got %.2f", want, result.FinalScore)
	}
}

// 场景5：只有两路有输入（缺 experience_reviewer）→ 按实际提供路重新归一化权重
func TestFuse_TwoInputs_Normalized(t *testing.T) {
	inputs := []JudgeInput{
		{Producer: ProducerHardRule, Score: 80, Decision: "passed"},
		{Producer: ProducerLLMJudge, Score: 70, Decision: "passed"},
		// 缺 experience_reviewer
	}
	result := Fuse(inputs, DefaultWeights)

	if result.FinalDecision != "passed" {
		t.Errorf("expected passed, got %s", result.FinalDecision)
	}
	// 归一化：hard_rule 权重 0.3，llm_judge 权重 0.3，总 0.6
	// 加权分：(80*0.3 + 70*0.3) / 0.6 = (24+21)/0.6 = 45/0.6 = 75
	want := 75.0
	if result.FinalScore < want-0.01 || result.FinalScore > want+0.01 {
		t.Errorf("expected score=%.1f, got %.2f", want, result.FinalScore)
	}
}
