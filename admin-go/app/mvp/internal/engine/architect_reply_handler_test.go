package engine

import (
	"context"
	"testing"
)

func repeatFollowUps(n int) []string {
	values := make([]string, 0, n)
	for i := 0; i < n; i++ {
		values = append(values, "继续")
	}
	return values
}

func TestArchitectReplyRequestsContinuation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "explicit token", content: "这是第一段\n[AUTO_CONTINUE_NEXT]", want: true},
		{name: "plain continue text is not enough", content: "这是第 1/4 段，如需继续请回复继续。", want: false},
		{name: "final patch payload", content: "```json\n{\"task_patches\":[{\"task_name\":\"登录\",\"description\":\"补齐 OAuth 回调说明\"}]}\n```", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := architectReplyRequestsContinuation(tc.content); got != tc.want {
				t.Fatalf("architectReplyRequestsContinuation() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractArchitectTaskPatchesMergesMultipleCodeBlocks(t *testing.T) {
	t.Parallel()

	content := "第一段\n```json\n{\"task_patches\":[{\"task_name\":\"登录\",\"description\":\"更新登录描述\",\"affected_resources\":[\"frontend/login.tsx\"]}]}\n```\n\n第二段\n```json\n{\"task_patches\":[{\"task_name\":\"资料页\",\"description\":\"补齐保存逻辑\",\"depends_on\":[\"登录\"]}]}\n```"
	patches, err := extractArchitectTaskPatches(content)
	if err != nil {
		t.Fatalf("extractArchitectTaskPatches() error = %v", err)
	}
	if len(patches) != 2 {
		t.Fatalf("expected 2 patches, got %d", len(patches))
	}
	if patches[0].TaskName != "登录" || patches[1].TaskName != "资料页" {
		t.Fatalf("unexpected patches: %+v", patches)
	}
	if len(patches[1].DependsOn) != 1 || patches[1].DependsOn[0] != "登录" {
		t.Fatalf("unexpected depends_on: %+v", patches[1])
	}
}

func TestExtractArchitectTaskPatchesSupportsSinglePatchObject(t *testing.T) {
	t.Parallel()

	content := "```json\n{\"task_name\":\"订单中心\",\"description\":\"修正批次顺序\",\"batch_no\":2,\"reason\":\"先完成库存接口\"}\n```"
	patches, err := extractArchitectTaskPatches(content)
	if err != nil {
		t.Fatalf("extractArchitectTaskPatches() error = %v", err)
	}
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].TaskName != "订单中心" || patches[0].BatchNo == nil || *patches[0].BatchNo != 2 {
		t.Fatalf("unexpected patch: %+v", patches[0])
	}
}

func TestBuildContinuationPromptIsFollowUp(t *testing.T) {
	t.Parallel()

	declaredTotal := 3
	chunkTotal := 3
	report := &ArchitectPlanParseReport{
		DeclaredTotal:       &declaredTotal,
		ChunkTotal:          &chunkTotal,
		RawTaskCount:        2,
		NormalizedTaskCount: 2,
		MissingChunkIndexes: []int{2},
	}

	prompt := report.BuildContinuationPrompt()
	if !isArchitectFollowUpMessage(prompt) {
		t.Fatalf("expected continuation prompt to be treated as follow-up, got: %s", prompt)
	}
	if isReviewRemediationPrompt(prompt) {
		t.Fatalf("continuation prompt should not look like review remediation: %s", prompt)
	}
}

func TestResolveArchitectReplyPolicy(t *testing.T) {
	t.Parallel()

	originalLoad := loadArchitectFollowUpLimit
	loadArchitectFollowUpLimit = func(ctx context.Context) int { return architectFollowUpLimitDefault }
	t.Cleanup(func() {
		loadArchitectFollowUpLimit = originalLoad
	})

	tests := []struct {
		name         string
		userContents []string
		wantContinue bool
		wantResubmit bool
	}{
		{
			name: "review remediation prompt enables continue and resubmit",
			userContents: []string{
				"## 方案审核未通过\n警告（当前会阻塞执行，必须修复）\n如果只是修正个别任务，请输出局部修订 JSON：{\"task_patches\": []}\n若你还有后续分段，请在当前消息最后单独追加一行 [AUTO_CONTINUE_NEXT]",
			},
			wantContinue: true,
			wantResubmit: true,
		},
		{
			name: "follow up continue keeps remediation context",
			userContents: []string{
				"继续",
				"## 方案审核未通过\n警告（当前会阻塞执行，必须修复）\n如果只是修正个别任务，请输出局部修订 JSON：{\"task_patches\": []}\n若你还有后续分段，请在当前消息最后单独追加一行 [AUTO_CONTINUE_NEXT]",
			},
			wantContinue: true,
			wantResubmit: true,
		},
		{
			name: "plain manual design request does not auto resubmit",
			userContents: []string{
				"请把任务拆细一点",
			},
			wantContinue: false,
			wantResubmit: false,
		},
		{
			name: "too many chained follow ups disable automation",
			userContents: append(
				repeatFollowUps(architectFollowUpLimitDefault),
				"## 方案审核未通过\n警告（当前会阻塞执行，必须修复）\n如果只是修正个别任务，请输出局部修订 JSON：{\"task_patches\": []}\n若你还有后续分段，请在当前消息最后单独追加一行 [AUTO_CONTINUE_NEXT]",
			),
			wantContinue: false,
			wantResubmit: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveArchitectReplyPolicy(context.Background(), tc.userContents)
			if got.allowAutoContinue != tc.wantContinue || got.allowAutoResubmit != tc.wantResubmit {
				t.Fatalf("resolveArchitectReplyPolicy() = %+v, want continue=%v resubmit=%v", got, tc.wantContinue, tc.wantResubmit)
			}
		})
	}
}

func TestShouldApplyArchitectBlueprintMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		currentStage string
		policy       architectReplyPolicy
		want         bool
	}{
		{
			name:         "design stage accepts normal replies",
			currentStage: "design",
			policy:       architectReplyPolicy{},
			want:         true,
		},
		{
			name:         "blank stage accepts normal replies",
			currentStage: "",
			policy:       architectReplyPolicy{},
			want:         true,
		},
		{
			name:         "review stage blocks stale design replies",
			currentStage: "review",
			policy:       architectReplyPolicy{},
			want:         false,
		},
		{
			name:         "review stage allows remediation replies",
			currentStage: "review",
			policy: architectReplyPolicy{
				allowAutoResubmit: true,
			},
			want: true,
		},
		{
			name:         "execute stage blocks non remediation replies",
			currentStage: "execute",
			policy:       architectReplyPolicy{},
			want:         false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := shouldApplyArchitectBlueprintMutation(tc.currentStage, tc.policy); got != tc.want {
				t.Fatalf("shouldApplyArchitectBlueprintMutation(%q, %+v) = %v, want %v", tc.currentStage, tc.policy, got, tc.want)
			}
		})
	}
}

func TestShouldParseArchitectReply(t *testing.T) {
	t.Parallel()

	originalLoad := loadArchitectFollowUpLimit
	loadArchitectFollowUpLimit = func(ctx context.Context) int { return architectFollowUpLimitDefault }
	t.Cleanup(func() {
		loadArchitectFollowUpLimit = originalLoad
	})

	tests := []struct {
		name         string
		userContents []string
		want         bool
	}{
		{
			name: "normal architect request should parse",
			userContents: []string{
				"贪吃蛇小游戏 react cli + goframe v2",
			},
			want: true,
		},
		{
			name: "review remediation prompt should still parse",
			userContents: []string{
				"## 方案审核未通过\n警告（当前会阻塞执行，必须修复）\n请重新给出完整修订方案",
			},
			want: true,
		},
		{
			name: "workflow approval notice should not parse",
			userContents: []string{
				"## 方案审核通过\n\n错误: 0，警告: 0\n\n项目已进入执行阶段。",
			},
			want: false,
		},
		{
			name: "follow up after approval notice should still not parse",
			userContents: []string{
				"继续",
				"## 方案审核通过\n\n错误: 0，警告: 0\n\n项目已进入执行阶段。",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldParseArchitectReply(context.Background(), tc.userContents); got != tc.want {
				t.Fatalf("shouldParseArchitectReply() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeArchitectFollowUpLimitFallsBackToDefaultWhenInvalid(t *testing.T) {
	t.Parallel()

	if got := normalizeArchitectFollowUpLimit(0); got != architectFollowUpLimitDefault {
		t.Fatalf("normalizeArchitectFollowUpLimit(0) = %d, want %d", got, architectFollowUpLimitDefault)
	}
	if got := normalizeArchitectFollowUpLimit(-1); got != architectFollowUpLimitDefault {
		t.Fatalf("normalizeArchitectFollowUpLimit(-1) = %d, want %d", got, architectFollowUpLimitDefault)
	}
	if got := normalizeArchitectFollowUpLimit(5); got != 5 {
		t.Fatalf("normalizeArchitectFollowUpLimit(5) = %d, want 5", got)
	}
}

func TestShouldParseArchitectReplyUsesConfiguredFollowUpLimit(t *testing.T) {
	t.Parallel()

	originalLoad := loadArchitectFollowUpLimit
	loadArchitectFollowUpLimit = func(ctx context.Context) int { return 2 }
	t.Cleanup(func() {
		loadArchitectFollowUpLimit = originalLoad
	})

	userContents := append(repeatFollowUps(2), "请重新规划方案")
	if got := shouldParseArchitectReply(context.Background(), userContents); got {
		t.Fatalf("expected parse to stop when follow-ups hit configured limit")
	}
}
