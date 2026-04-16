package verification

import (
	"context"
	"encoding/json"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

type issueTaskResolutionInput struct {
	CurrentTaskID int64
	Title         string
	Detail        string
	ResourceRef   string
}

type taskBindingCandidate struct {
	ID                int64
	Name              string
	ParentTaskID      int64
	AffectedResources []string
	Depth             int
}

type issueTaskSignals struct {
	ResourceRef string
	Paths       []string
	Text        string
	IsBuild     bool
	IsTest      bool
	IsLint      bool
	IsBackend   bool
	IsFrontend  bool
}

var issuePathPattern = regexp.MustCompile(`([A-Za-z0-9._-]+(?:/[A-Za-z0-9._-]+)+\.[A-Za-z0-9._-]+|[A-Za-z0-9._-]+\.[A-Za-z0-9._-]+)`)

// ResolveIssueTaskID 根据验证问题内容推断更合适的返工任务。
func ResolveIssueTaskID(ctx context.Context, workflowRunID, currentTaskID int64, title, detail, resourceRef string) (int64, error) {
	if workflowRunID == 0 {
		return currentTaskID, nil
	}

	tasks, err := loadTaskBindingCandidates(ctx, workflowRunID)
	if err != nil {
		return currentTaskID, err
	}
	return resolveIssueTaskIDFromCandidates(issueTaskResolutionInput{
		CurrentTaskID: currentTaskID,
		Title:         title,
		Detail:        detail,
		ResourceRef:   resourceRef,
	}, tasks), nil
}

func loadTaskBindingCandidates(ctx context.Context, workflowRunID int64) ([]taskBindingCandidate, error) {
	records, err := repo.NewDomainTaskRepo().ListByWorkflowOrdered(ctx, workflowRunID, "id", "name", "parent_task_id", "affected_resources")
	if err != nil {
		return nil, err
	}

	candidates := make([]taskBindingCandidate, 0, len(records))
	parentMap := make(map[int64]int64, len(records))
	for _, record := range records {
		var resources []string
		if raw := strings.TrimSpace(g.NewVar(record["affected_resources"]).String()); raw != "" {
			_ = json.Unmarshal([]byte(raw), &resources)
		}
		candidate := taskBindingCandidate{
			ID:                g.NewVar(record["id"]).Int64(),
			Name:              strings.TrimSpace(g.NewVar(record["name"]).String()),
			ParentTaskID:      g.NewVar(record["parent_task_id"]).Int64(),
			AffectedResources: normalizeAffectedResources(resources),
		}
		candidates = append(candidates, candidate)
		parentMap[candidate.ID] = candidate.ParentTaskID
	}
	for i := range candidates {
		candidates[i].Depth = taskDepth(parentMap, candidates[i].ID)
	}
	return candidates, nil
}

func resolveIssueTaskIDFromCandidates(input issueTaskResolutionInput, tasks []taskBindingCandidate) int64 {
	if len(tasks) == 0 {
		return input.CurrentTaskID
	}

	signals := buildIssueTaskSignals(input)
	scores := make(map[int64]int, len(tasks))
	pathMatched := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		score, matchedPath := scoreTaskBindingCandidate(signals, task)
		if score <= 0 {
			continue
		}
		if task.ID == input.CurrentTaskID {
			score += 5
		}
		scores[task.ID] = score
		if matchedPath {
			pathMatched = append(pathMatched, task.ID)
		}
	}

	if ancestorID := deepestCommonAncestor(uniqueInt64s(pathMatched), tasks); ancestorID > 0 {
		bestPathTaskID := pickBestTaskFromSet(uniqueInt64s(pathMatched), scores, tasks, input.CurrentTaskID)
		if bestPathTaskID > 0 &&
			bestPathTaskID != ancestorID &&
			tasksOnSingleLineage(uniqueInt64s(pathMatched), tasks) &&
			isTaskAncestor(ancestorID, bestPathTaskID, tasks) &&
			scores[bestPathTaskID] >= scores[ancestorID]+40 {
			return bestPathTaskID
		}
		return ancestorID
	}
	if len(pathMatched) == 1 {
		return pathMatched[0]
	}

	bestID := pickBestScoredTask(scores, tasks, input.CurrentTaskID)
	if bestID > 0 {
		return bestID
	}
	return input.CurrentTaskID
}

func buildIssueTaskSignals(input issueTaskResolutionInput) issueTaskSignals {
	resourceRef := normalizeBindingPath(input.ResourceRef)
	if resourceRef == "" {
		resourceRef = "."
	}
	text := strings.ToLower(strings.Join([]string{input.Title, input.Detail, resourceRef}, "\n"))
	return issueTaskSignals{
		ResourceRef: resourceRef,
		Paths:       extractIssuePaths(input.Detail, resourceRef),
		Text:        text,
		IsBuild:     strings.Contains(text, " build") || strings.Contains(text, "构建") || strings.Contains(text, "npm ci") || strings.Contains(text, "vite build"),
		IsTest:      strings.Contains(text, "test") || strings.Contains(text, "测试") || strings.Contains(text, "go test"),
		IsLint:      strings.Contains(text, "lint"),
		IsBackend:   strings.Contains(text, "backend") || strings.Contains(text, "goframe") || strings.Contains(text, "go.sum"),
		IsFrontend:  strings.Contains(text, "frontend") || strings.Contains(text, "vite") || strings.Contains(text, "tsc"),
	}
}

func scoreTaskBindingCandidate(signals issueTaskSignals, task taskBindingCandidate) (int, bool) {
	var (
		score       int
		matchedPath bool
		nameLower   = strings.ToLower(task.Name)
	)

	for _, resource := range task.AffectedResources {
		resource = normalizeBindingPath(resource)
		if resource == "" {
			continue
		}

		for _, refPath := range signals.Paths {
			switch {
			case resource == refPath:
				score += 240
				matchedPath = true
			case strings.HasPrefix(refPath, resource+"/"):
				score += 180
				matchedPath = true
			case strings.HasPrefix(resource, refPath+"/"):
				score += 90
				matchedPath = true
			}
		}

		switch {
		case signals.ResourceRef == ".":
			if !strings.Contains(resource, "/") {
				score += 20
			}
		case resource == signals.ResourceRef:
			score += 80
		case strings.HasPrefix(resource, signals.ResourceRef+"/"):
			score += 55
		case strings.HasPrefix(signals.ResourceRef, resource+"/"):
			score += 25
		}

		if signals.IsBuild {
			if strings.Contains(nameLower, "build") {
				score += 70
			}
			switch {
			case resource == "scripts/build.js":
				score += 80
			case resource == "package.json" || strings.HasSuffix(resource, "/package.json"):
				score += 35
			case resource == "Makefile" || strings.HasSuffix(strings.ToLower(resource), "dockerfile"):
				score += 35
			case strings.HasSuffix(resource, "vite.config.ts") || strings.HasSuffix(resource, "tsconfig.json") || strings.HasSuffix(resource, "go.mod"):
				score += 40
			}
		}
		if signals.IsTest {
			if strings.Contains(nameLower, "test") {
				score += 55
			}
			if strings.HasSuffix(resource, "_test.go") || strings.Contains(resource, "/e2e/") || strings.Contains(resource, "playwright") {
				score += 30
			}
		}
		if signals.IsLint && strings.Contains(nameLower, "lint") {
			score += 40
		}
	}

	if signals.IsBackend && strings.Contains(nameLower, "backend") {
		score += 25
	}
	if signals.IsFrontend && strings.Contains(nameLower, "frontend") {
		score += 25
	}
	if strings.Contains(signals.Text, "websocket") && strings.Contains(nameLower, "websocket") {
		score += 20
	}
	if strings.Contains(signals.Text, "goframe") && strings.Contains(nameLower, "goframe") {
		score += 20
	}
	if strings.Contains(signals.Text, "vite") && strings.Contains(nameLower, "vite") {
		score += 20
	}

	return score, matchedPath
}

func extractIssuePaths(detail, resourceRef string) []string {
	lines := strings.Split(strings.ReplaceAll(detail, "\\", "/"), "\n")
	paths := make([]string, 0, 8)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, match := range issuePathPattern.FindAllString(line, -1) {
			candidate := normalizeBindingPath(strings.Trim(match, "\"'(),"))
			if candidate == "" {
				continue
			}
			if idx := strings.Index(candidate, "/repo/"); idx >= 0 {
				candidate = normalizeBindingPath(candidate[idx+len("/repo/"):])
			}
			if strings.HasPrefix(candidate, "/") || strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
				continue
			}
			if resourceRef != "" && resourceRef != "." && candidate != resourceRef && !strings.HasPrefix(candidate, resourceRef+"/") {
				candidate = normalizeBindingPath(path.Join(resourceRef, candidate))
			}
			if candidate == "" {
				continue
			}
			paths = append(paths, candidate)
		}
	}
	return uniqueStrings(paths)
}

func normalizeAffectedResources(resources []string) []string {
	if len(resources) == 0 {
		return nil
	}
	items := make([]string, 0, len(resources))
	for _, item := range resources {
		if normalized := normalizeBindingPath(item); normalized != "" {
			items = append(items, normalized)
		}
	}
	return uniqueStrings(items)
}

func normalizeBindingPath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.TrimPrefix(value, "./")
	if value == "" {
		return ""
	}
	cleaned := path.Clean(value)
	if cleaned == "." && value != "." {
		return ""
	}
	return cleaned
}

func deepestCommonAncestor(taskIDs []int64, tasks []taskBindingCandidate) int64 {
	if len(taskIDs) <= 1 {
		return 0
	}

	parentMap := make(map[int64]int64, len(tasks))
	depthMap := make(map[int64]int, len(tasks))
	for _, task := range tasks {
		parentMap[task.ID] = task.ParentTaskID
		depthMap[task.ID] = task.Depth
	}

	common := ancestorSet(taskIDs[0], parentMap)
	for _, taskID := range taskIDs[1:] {
		current := ancestorSet(taskID, parentMap)
		for id := range common {
			if _, ok := current[id]; !ok {
				delete(common, id)
			}
		}
	}

	var (
		bestID    int64
		bestDepth = -1
	)
	for id := range common {
		if depthMap[id] > bestDepth {
			bestID = id
			bestDepth = depthMap[id]
		}
	}
	return bestID
}

func isTaskAncestor(ancestorID, taskID int64, tasks []taskBindingCandidate) bool {
	if ancestorID <= 0 || taskID <= 0 {
		return false
	}
	parentMap := make(map[int64]int64, len(tasks))
	for _, task := range tasks {
		parentMap[task.ID] = task.ParentTaskID
	}
	for current := taskID; current > 0; current = parentMap[current] {
		if current == ancestorID {
			return true
		}
	}
	return false
}

func tasksOnSingleLineage(taskIDs []int64, tasks []taskBindingCandidate) bool {
	if len(taskIDs) <= 1 {
		return true
	}

	var (
		deepestID    int64
		deepestDepth = -1
	)
	for _, task := range tasks {
		for _, taskID := range taskIDs {
			if task.ID != taskID {
				continue
			}
			if task.Depth > deepestDepth {
				deepestDepth = task.Depth
				deepestID = task.ID
			}
		}
	}
	if deepestID == 0 {
		return false
	}

	for _, taskID := range taskIDs {
		if taskID == deepestID {
			continue
		}
		if !isTaskAncestor(taskID, deepestID, tasks) {
			return false
		}
	}
	return true
}

func ancestorSet(taskID int64, parentMap map[int64]int64) map[int64]struct{} {
	result := make(map[int64]struct{})
	for current := taskID; current > 0; current = parentMap[current] {
		result[current] = struct{}{}
	}
	return result
}

func taskDepth(parentMap map[int64]int64, taskID int64) int {
	depth := 0
	for current := parentMap[taskID]; current > 0; current = parentMap[current] {
		depth++
	}
	return depth
}

func pickBestScoredTask(scores map[int64]int, tasks []taskBindingCandidate, currentTaskID int64) int64 {
	type scoredTask struct {
		id    int64
		score int
		depth int
	}
	best := scoredTask{}
	for _, task := range tasks {
		score := scores[task.ID]
		if score <= 0 {
			continue
		}
		item := scoredTask{id: task.ID, score: score, depth: task.Depth}
		if item.score > best.score ||
			(item.score == best.score && item.depth > best.depth) ||
			(item.score == best.score && item.depth == best.depth && item.id == currentTaskID && best.id != currentTaskID) ||
			(item.score == best.score && item.depth == best.depth && best.id == 0) {
			best = item
		}
	}
	return best.id
}

func pickBestTaskFromSet(taskIDs []int64, scores map[int64]int, tasks []taskBindingCandidate, currentTaskID int64) int64 {
	allowed := make(map[int64]struct{}, len(taskIDs))
	for _, taskID := range taskIDs {
		allowed[taskID] = struct{}{}
	}

	filteredScores := make(map[int64]int, len(taskIDs))
	for taskID := range allowed {
		if score := scores[taskID]; score > 0 {
			filteredScores[taskID] = score
		}
	}
	return pickBestScoredTask(filteredScores, tasks, currentTaskID)
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func uniqueInt64s(items []int64) []int64 {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(items))
	result := make([]int64, 0, len(items))
	for _, item := range items {
		if item <= 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
