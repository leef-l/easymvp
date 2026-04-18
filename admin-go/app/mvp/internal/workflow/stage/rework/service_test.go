package rework

import (
	"sort"
	"strings"
	"testing"
)

func TestClassifyOriginalTaskStatusForRework(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		status         string
		wantAllowReset bool
		wantConverged  bool
	}{
		{name: "failed can reset", status: "failed", wantAllowReset: true},
		{name: "escalated can reset", status: "escalated", wantAllowReset: true},
		{name: "pending already recovered", status: "pending", wantConverged: true},
		{name: "running already recovered", status: "running", wantConverged: true},
		{name: "completed already recovered", status: "completed", wantConverged: true},
		{name: "unknown blocks recovery", status: "bug_found"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAllowReset, gotConverged := classifyOriginalTaskStatusForRework(tc.status)
			if gotAllowReset != tc.wantAllowReset || gotConverged != tc.wantConverged {
				t.Fatalf("classifyOriginalTaskStatusForRework(%q) = (%v, %v), want (%v, %v)",
					tc.status, gotAllowReset, gotConverged, tc.wantAllowReset, tc.wantConverged)
			}
		})
	}
}

func TestParseTaskPatch(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"description":"更新 README","affected_resources":["README.md"],"reason":"补齐说明"}`, "")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "更新 README" || len(patch.AffectedResources) != 1 || patch.AffectedResources[0] != "README.md" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsFencedJSON(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch("修复建议如下：\n```json\n{\"description\":\"修复任务描述\",\"affected_resources\":[\"docs/plan.md\"],\"reason\":\"补齐交付说明\"}\n```", "")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Reason != "补齐交付说明" || patch.AffectedResources[0] != "docs/plan.md" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsEscapedFencedJSON(t *testing.T) {
	t.Parallel()

	content := "根据错误信息分析，这是一个修复方案。\\n\\n```json\\n" +
		"{\\\"task_repair\\\":{\\\"task_name\\\":\\\"backend-goframe-init\\\",\\\"description\\\":\\\"手工创建最小后端骨架\\\",\\\"affected_resources\\\":[\\\"backend/go.mod\\\",\\\"backend/main.go\\\"],\\\"reason\\\":\\\"消息内容被转义后仍应解析成功\\\"}}\\n```"
	patch, err := parseTaskPatch(content, "backend-goframe-init")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "手工创建最小后端骨架" || patch.Reason != "消息内容被转义后仍应解析成功" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsNarrativeTaskRepairSample(t *testing.T) {
	t.Parallel()

	content := "我来分析这个任务失败的具体原因。\n\n## 失败原因分析\n\n这是说明文字。\n\n```json\n" +
		"{\"task_repair\":{\"task_name\":\"Implement GoFrame leaderboard domain\",\"description\":\"Build the backend game models, validation rules, file-backed leaderboard persistence, and deterministic sorting so submitted snake runs are safely stored and returned in rank order. 注意：本任务需在标准执行模式下运行，不依赖特殊e2e执行器配置。\",\"affected_resources\":[\"backend/internal/model/leaderboard.go\",\"backend/internal/service/leaderboard.go\",\"backend/internal/controller/leaderboard.go\",\"backend/internal/dao/leaderboard.go\",\"backend/manifest/config/config.yaml\"],\"reason\":\"失败原因：任务被标记为disabled_for_e2e执行模式，但执行器未注册该模式处理器，导致执行前校验失败。修复：移除对特殊e2e执行模式的依赖，确保使用标准执行模式；同时在description中显式声明执行模式要求，避免执行器模式匹配冲突。原任务代码实现目标不变，仅调整执行元数据兼容性。\"}}\n```"

	patch, err := parseTaskPatch(content, "Implement GoFrame leaderboard domain")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if !strings.Contains(patch.Description, "标准执行模式") {
		t.Fatalf("unexpected patch description: %+v", patch)
	}
	if len(patch.AffectedResources) != 5 {
		t.Fatalf("unexpected affected resources: %+v", patch.AffectedResources)
	}
}

func TestParseTaskPatchSupportsTaskRepairEnvelope(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"task_repair":{"task_name":"cli-root-init","description":"收缩根目录初始化范围","affected_resources":["package.json",".gitignore"],"reason":"pnpm-workspace.yaml 越界"}}`, "cli-root-init")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "收缩根目录初始化范围" || len(patch.AffectedResources) != 2 {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchSupportsMatchingTaskPatchEnvelope(t *testing.T) {
	t.Parallel()

	patch, err := parseTaskPatch(`{"task_patches":[{"task_name":"backend-init","description":"忽略","reason":"无关"},{"task_name":"cli-root-init","description":"仅初始化根目录脚本","affected_resources":["package.json","scripts/dev.js"],"reason":"不能越界到 workspace 配置"}]}`, "cli-root-init")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if patch.Description != "仅初始化根目录脚本" || patch.Reason != "不能越界到 workspace 配置" {
		t.Fatalf("unexpected patch: %+v", patch)
	}
}

func TestParseTaskPatchRejectsMismatchedMultiTaskPatchEnvelope(t *testing.T) {
	t.Parallel()

	if _, err := parseTaskPatch(`{"task_patches":[{"task_name":"backend-init","description":"忽略","reason":"无关"},{"task_name":"frontend-init","description":"忽略","reason":"无关"}]}`, "cli-root-init"); err == nil {
		t.Fatal("expected mismatched task_patches to fail")
	}
}

func TestParseTaskPatchRejectsInvalidContent(t *testing.T) {
	t.Parallel()

	if _, err := parseTaskPatch("not json", ""); err == nil {
		t.Fatal("expected invalid content to fail")
	}
}

func TestParseTaskPatchSupportsPlanTasksEnvelope(t *testing.T) {
	t.Parallel()

	content := "分析任务失败原因：\n\n```json\n" +
		"{\"plan_meta\":{\"plan_id\":\"snake-repair\",\"declared_total\":3},\"tasks\":[" +
		"{\"name\":\"Implement Frontend Canvas Renderer Core (MVP)\",\"description\":\"创建 Canvas 渲染器基类，实现 60fps requestAnimationFrame 循环。\",\"affected_resources\":[\"frontend/src/renderers/canvas-renderer-core.ts\"]}," +
		"{\"name\":\"Implement Cyberpunk Visual Theme\",\"description\":\"扩展主题配置对象和霓虹配色系统。\",\"affected_resources\":[\"frontend/src/renderers/canvas-renderer-core.ts\"]}," +
		"{\"name\":\"Implement Dynamic VFX System\",\"description\":\"补充 pulse 动画、死亡粒子和性能降级逻辑。\",\"affected_resources\":[\"frontend/src/renderers/canvas-renderer-core.ts\",\"frontend/src/utils/perf-monitor.ts\"]}" +
		"]}\n```"

	patch, err := parseTaskPatch(content, "Implement Frontend Canvas Renderer and VFX")
	if err != nil {
		t.Fatalf("parseTaskPatch() error = %v", err)
	}
	if !strings.Contains(patch.Description, "步骤1：Implement Frontend Canvas Renderer Core (MVP)") {
		t.Fatalf("unexpected patch description: %+v", patch)
	}
	if len(patch.AffectedResources) != 2 {
		t.Fatalf("unexpected affected resources: %+v", patch.AffectedResources)
	}
	if !strings.Contains(patch.Reason, "同批次子任务") {
		t.Fatalf("unexpected patch reason: %+v", patch)
	}
}

func TestParseAnalysisResolutionSupportsSplitPlan(t *testing.T) {
	t.Parallel()

	content := `{"plan_meta":{"plan_id":"snake-repair","declared_total":2},"tasks":[` +
		`{"name":"UI Shell","description":"只做主菜单和暂停面板","role_type":"implementer","role_level":"pro","batch_no":3,"affected_resources":["frontend/src/components/panels/MainMenuPanel.tsx"],"depends_on":[]},` +
		`{"name":"Leaderboard Panel","description":"单独做排行榜面板","role_type":"implementer","role_level":"pro","batch_no":3,"affected_resources":["frontend/src/components/panels/LeaderboardPanel.tsx"],"depends_on":["UI Shell"]}` +
		`]}`

	resolution, err := parseAnalysisResolution(content, "Implement Frontend UI Panels and Theme")
	if err != nil {
		t.Fatalf("parseAnalysisResolution() error = %v", err)
	}
	if resolution.SplitPlan == nil {
		t.Fatalf("expected split plan, got %+v", resolution)
	}
	if len(resolution.SplitPlan.Tasks) != 2 {
		t.Fatalf("unexpected split tasks: %+v", resolution.SplitPlan.Tasks)
	}
	if resolution.SplitPlan.Tasks[1].DependsOn[0] != "UI Shell" {
		t.Fatalf("unexpected depends_on: %+v", resolution.SplitPlan.Tasks[1].DependsOn)
	}
	if !strings.Contains(resolution.SplitPlan.Reason, "同批次子任务") {
		t.Fatalf("unexpected split reason: %+v", resolution.SplitPlan.Reason)
	}
}

func TestBuildFallbackPatchFromSplitPlan(t *testing.T) {
	t.Parallel()

	plan := &taskSplitPlan{
		Reason: "原始拆分方案",
		Tasks: []splitTaskSpec{
			{Name: "Task A", Description: "只做 A", AffectedResources: []string{"a.ts"}},
			{Name: "Task B", Description: "只做 B", AffectedResources: []string{"b.ts"}},
		},
	}

	patch := buildFallbackPatchFromSplitPlan(plan, "超过拆分预算")
	if patch == nil {
		t.Fatal("expected fallback patch")
	}
	if patch.Reason != "超过拆分预算" {
		t.Fatalf("unexpected reason: %q", patch.Reason)
	}
	if !strings.Contains(patch.Description, "步骤1：Task A") {
		t.Fatalf("unexpected description: %q", patch.Description)
	}
	if len(patch.AffectedResources) != 2 {
		t.Fatalf("unexpected resources: %+v", patch.AffectedResources)
	}
}

func TestNormalizeReworkSplitLimits(t *testing.T) {
	t.Parallel()

	if got := normalizeReworkSplitDepthLimit(0); got != 1 {
		t.Fatalf("normalizeReworkSplitDepthLimit(0) = %d, want 1", got)
	}
	if got := normalizeReworkSplitDepthLimit(3); got != 3 {
		t.Fatalf("normalizeReworkSplitDepthLimit(3) = %d, want 3", got)
	}
	if got := normalizeReworkSplitTaskBudget(0); got != 2 {
		t.Fatalf("normalizeReworkSplitTaskBudget(0) = %d, want 2", got)
	}
	if got := normalizeReworkSplitTaskBudget(8); got != 8 {
		t.Fatalf("normalizeReworkSplitTaskBudget(8) = %d, want 8", got)
	}
}

func TestFindExpandedTaskResources(t *testing.T) {
	t.Parallel()

	got := findExpandedTaskResources(
		[]string{"frontend/src/a.ts", "frontend/src/b.ts"},
		[]string{"frontend/src/b.ts", "frontend/src/c.ts", "frontend/src/c.ts"},
	)
	if len(got) != 1 || got[0] != "frontend/src/c.ts" {
		t.Fatalf("findExpandedTaskResources() = %+v", got)
	}
}

func TestSanitizeReworkDescriptionRemovesLocalBuildInstructions(t *testing.T) {
	t.Parallel()

	got := sanitizeReworkDescription("保持现有面板逻辑。\n每阶段完成后运行 npm run build 验证。\n补齐主题切换状态。")
	if strings.Contains(strings.ToLower(got), "npm run build") {
		t.Fatalf("sanitizeReworkDescription() still contains local build command: %q", got)
	}
	if !strings.Contains(got, "保持现有面板逻辑") || !strings.Contains(got, "补齐主题切换状态") {
		t.Fatalf("sanitizeReworkDescription() removed too much: %q", got)
	}
}

func TestBuildRootTaskBudgetWhereIncludesRootTaskItself(t *testing.T) {
	t.Parallel()

	expr, args := buildRootTaskBudgetWhere(42)
	if expr != "(id = ? OR root_task_id = ?)" {
		t.Fatalf("unexpected expr: %q", expr)
	}
	if len(args) != 2 {
		t.Fatalf("unexpected args length: %d", len(args))
	}
	if args[0] != int64(42) || args[1] != int64(42) {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestFindNextEscalatedTaskSelectionOrder(t *testing.T) {
	t.Parallel()

	records := []map[string]int64{
		{"id": 3003, "batch_no": 4, "sort": 2},
		{"id": 3001, "batch_no": 3, "sort": 5},
		{"id": 3002, "batch_no": 3, "sort": 1},
	}
	sort.Slice(records, func(i, j int) bool { return records[i]["id"] > records[j]["id"] })

	bestBatch := 0
	bestSort := 0
	bestID := int64(0)
	found := false
	for _, record := range records {
		taskID := record["id"]
		if taskID == 3001 {
			continue
		}
		batchNo := int(record["batch_no"])
		sortNo := int(record["sort"])
		if !found || batchNo < bestBatch || (batchNo == bestBatch && sortNo < bestSort) || (batchNo == bestBatch && sortNo == bestSort && taskID < bestID) {
			bestBatch = batchNo
			bestSort = sortNo
			bestID = taskID
			found = true
		}
	}

	if !found || bestID != 3002 {
		t.Fatalf("expected task 3002, got found=%v id=%d", found, bestID)
	}
}
