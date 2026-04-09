package acceptance

import (
	"strings"
	"testing"
)

func TestBuildJudgeUserPromptFallbacks(t *testing.T) {
	t.Parallel()

	prompt := buildJudgeUserPrompt(&AcceptContext{
		ProjectType: "software_dev",
		WorkDir:     "/tmp/project",
	}, nil, nil)

	if !strings.Contains(prompt, "项目类型：software_dev") {
		t.Fatalf("prompt missing project type: %s", prompt)
	}
	if !strings.Contains(prompt, "工作目录：/tmp/project") {
		t.Fatalf("prompt missing workdir: %s", prompt)
	}
	if !strings.Contains(prompt, "所有硬规则通过，无命中问题。") {
		t.Fatalf("prompt missing empty-hit fallback: %s", prompt)
	}
	if !strings.Contains(prompt, "无收集到证据。") {
		t.Fatalf("prompt missing empty-evidence fallback: %s", prompt)
	}
}

func TestNewJudge(t *testing.T) {
	t.Parallel()

	if judge := NewJudge(); judge == nil {
		t.Fatal("NewJudge() returned nil")
	}
}

func TestBuildJudgeSystemPrompt(t *testing.T) {
	t.Parallel()

	prompt := buildJudgeSystemPrompt()
	for _, fragment := range []string{
		`"quality_score": 0-100`,
		`"conclusion": "passed|failed|uncertain"`,
		"评分标准",
		"conclusion 判断标准",
	} {
		if !strings.Contains(prompt, fragment) {
			t.Fatalf("system prompt missing %q: %s", fragment, prompt)
		}
	}
}

func TestBuildJudgeUserPromptFormatsHitsAndTruncatesEvidence(t *testing.T) {
	t.Parallel()

	longSummary := strings.Repeat("a", 510)
	prompt := buildJudgeUserPrompt(&AcceptContext{
		ProjectType: "mini_app",
		WorkDir:     "/workspace/demo",
	}, []EvidenceItem{
		{
			EvidenceType: "workspace_file",
			SourceType:   "workspace",
			Summary:      longSummary,
		},
	}, []RuleHit{
		{
			Severity:      SeverityWarn,
			RuleCode:      "software.required_files",
			Title:         "缺少 README",
			ExpectedValue: "README.md",
			ActualValue:   "missing",
		},
	})

	if !strings.Contains(prompt, "[warn] software.required_files: 缺少 README (期望: README.md, 实际: missing)") {
		t.Fatalf("prompt missing formatted rule hit: %s", prompt)
	}
	if !strings.Contains(prompt, "[workspace_file] workspace: "+strings.Repeat("a", 500)+"...") {
		t.Fatalf("prompt missing truncated evidence summary")
	}
	if strings.Contains(prompt, longSummary) {
		t.Fatalf("prompt should truncate long evidence summary")
	}
}

func TestParseJudgeResponseAcceptsJSONAndCodeBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    JudgeResult
	}{
		{
			name:    "raw json",
			content: `{"quality_score": 88, "conclusion": "passed", "summary": "ok", "suggestions": ["a"]}`,
			want: JudgeResult{
				QualityScore: 88,
				Conclusion:   "passed",
				Summary:      "ok",
				Suggestions:  []string{"a"},
			},
		},
		{
			name: "markdown block",
			content: "分析如下：\n```json\n" +
				`{"quality_score": 120, "conclusion": "unknown", "summary": "need review"}` +
				"\n```",
			want: JudgeResult{
				QualityScore: 100,
				Conclusion:   "uncertain",
				Summary:      "need review",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseJudgeResponse(tc.content)
			if err != nil {
				t.Fatalf("parseJudgeResponse() error = %v", err)
			}
			if got.QualityScore != tc.want.QualityScore {
				t.Fatalf("quality_score = %v, want %v", got.QualityScore, tc.want.QualityScore)
			}
			if got.Conclusion != tc.want.Conclusion {
				t.Fatalf("conclusion = %q, want %q", got.Conclusion, tc.want.Conclusion)
			}
			if got.Summary != tc.want.Summary {
				t.Fatalf("summary = %q, want %q", got.Summary, tc.want.Summary)
			}
			if len(got.Suggestions) != len(tc.want.Suggestions) {
				t.Fatalf("suggestions len = %d, want %d", len(got.Suggestions), len(tc.want.Suggestions))
			}
		})
	}
}

func TestParseJudgeResponseRejectsInvalidContent(t *testing.T) {
	t.Parallel()

	if _, err := parseJudgeResponse("not json"); err == nil {
		t.Fatal("parseJudgeResponse() expected error for invalid content")
	}
}

func TestValidateJudgeResultNormalizesRangeAndConclusion(t *testing.T) {
	t.Parallel()

	got := validateJudgeResult(&JudgeResult{
		QualityScore: -12,
		Conclusion:   "maybe",
		Summary:      "fallback",
	})
	if got.QualityScore != 0 {
		t.Fatalf("quality_score = %v, want 0", got.QualityScore)
	}
	if got.Conclusion != "uncertain" {
		t.Fatalf("conclusion = %q, want uncertain", got.Conclusion)
	}

	got = validateJudgeResult(&JudgeResult{
		QualityScore: 999,
		Conclusion:   "failed",
	})
	if got.QualityScore != 100 {
		t.Fatalf("quality_score = %v, want 100", got.QualityScore)
	}
	if got.Conclusion != "failed" {
		t.Fatalf("conclusion = %q, want failed", got.Conclusion)
	}
}
