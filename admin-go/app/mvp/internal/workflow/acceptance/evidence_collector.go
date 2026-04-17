package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/repo"
)

// EvidenceCollector 证据收集器。
type EvidenceCollector struct {
	evidenceRepo *repo.AcceptEvidenceRepo
}

// NewEvidenceCollector 创建证据收集器。
func NewEvidenceCollector(evidenceRepo *repo.AcceptEvidenceRepo) *EvidenceCollector {
	return &EvidenceCollector{evidenceRepo: evidenceRepo}
}

// Collect 收集工作流运行的所有证据并持久化。
func (c *EvidenceCollector) Collect(ctx context.Context, in *AcceptContext) ([]EvidenceItem, error) {
	var items []EvidenceItem

	// 1. 收集 domain_task 结果
	taskItems, err := c.collectTaskOutputs(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集任务输出失败: %v", err)
	} else {
		items = append(items, taskItems...)
	}

	// 2. 收集 stage_run 输出
	stageItems, err := c.collectStageOutputs(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集阶段输出失败: %v", err)
	} else {
		items = append(items, stageItems...)
	}

	// 3. 收集 handoff_record（返工记录）
	handoffItems, err := c.collectHandoffs(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集交接记录失败: %v", err)
	} else {
		items = append(items, handoffItems...)
	}

	// 4. 收集 workspace 交付结果
	workspaceItems, err := c.collectWorkspaceArtifacts(ctx, in.WorkflowRunID)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集 workspace 交付结果失败: %v", err)
	} else {
		items = append(items, workspaceItems...)
	}

	// 5. 收集 CI/构建/静态检查证据
	ciItems, err := c.collectCIArtifacts(ctx, in)
	if err != nil {
		g.Log().Warningf(ctx, "[EvidenceCollector] 收集 CI 证据失败: %v", err)
	} else {
		items = append(items, ciItems...)
	}

	// 6. 持久化证据
	if len(items) > 0 {
		now := gtime.Now()
		var dbItems []g.Map
		for _, item := range items {
			dbItems = append(dbItems, g.Map{
				"accept_run_id": in.AcceptRunID,
				"evidence_type": item.EvidenceType,
				"source_type":   item.SourceType,
				"source_id":     item.SourceID,
				"content_ref":   item.ContentRef,
				"summary":       item.Summary,
				"created_at":    now,
				"updated_at":    now,
			})
		}
		if err := c.evidenceRepo.BatchCreate(ctx, dbItems); err != nil {
			return items, fmt.Errorf("持久化证据失败: %w", err)
		}
	}

	g.Log().Infof(ctx, "[EvidenceCollector] 收集到 %d 条证据, acceptRunID=%d", len(items), in.AcceptRunID)
	return items, nil
}

// collectTaskOutputs 收集领域任务的执行结果。
func (c *EvidenceCollector) collectTaskOutputs(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	tasks, err := repo.NewDomainTaskRepo().ListByWorkflowOrdered(ctx, workflowRunID, "id", "name", "status", "result", "task_kind")
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, t := range tasks {
		summary := fmt.Sprintf("[%s] %s: status=%s",
			g.NewVar(t["task_kind"]).String(),
			g.NewVar(t["name"]).String(),
			g.NewVar(t["status"]).String(),
		)
		items = append(items, EvidenceItem{
			EvidenceType: "task_output",
			SourceType:   "domain_task",
			SourceID:     g.NewVar(t["id"]).Int64(),
			ContentRef:   g.NewVar(t["result"]).String(),
			Summary:      summary,
		})
	}
	return items, nil
}

// collectStageOutputs 收集阶段运行记录。
func (c *EvidenceCollector) collectStageOutputs(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	stages, err := repo.NewStageRunRepo().ListByWorkflowMaps(ctx, workflowRunID, "id", "stage_type", "status", "error_message")
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, s := range stages {
		summary := fmt.Sprintf("stage=%s status=%s", g.NewVar(s["stage_type"]).String(), g.NewVar(s["status"]).String())
		items = append(items, EvidenceItem{
			EvidenceType: "stage_output",
			SourceType:   "stage_run",
			SourceID:     g.NewVar(s["id"]).Int64(),
			Summary:      summary,
		})
	}
	return items, nil
}

// collectHandoffs 收集返工交接记录。
func (c *EvidenceCollector) collectHandoffs(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	records, err := repo.NewHandoffRecordRepo().ListByWorkflow(ctx, workflowRunID, "id", "from_task_id", "to_task_id", "handoff_type", "reason")
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, r := range records {
		summary := fmt.Sprintf("handoff: %s from=%d to=%d",
			g.NewVar(r["handoff_type"]).String(),
			g.NewVar(r["from_task_id"]).Int64(),
			g.NewVar(r["to_task_id"]).Int64(),
		)
		items = append(items, EvidenceItem{
			EvidenceType: "handoff",
			SourceType:   "handoff_record",
			SourceID:     g.NewVar(r["id"]).Int64(),
			ContentRef:   g.NewVar(r["reason"]).String(),
			Summary:      summary,
		})
	}
	return items, nil
}

func (c *EvidenceCollector) collectWorkspaceArtifacts(ctx context.Context, workflowRunID int64) ([]EvidenceItem, error) {
	records, err := queryWorkspaceArtifactRecords(ctx, workflowRunID)
	if err != nil {
		return nil, err
	}

	var items []EvidenceItem
	for _, record := range records {
		patchRef := record["patch_ref"].String()
		deliveryRef := record["delivery_ref"].String()
		diffSummary := record["diff_summary"].String()
		if patchRef == "" && diffSummary == "" && deliveryRef == "" {
			continue
		}

		summary := fmt.Sprintf(
			"workspace task=%d mode=%s delivery=%s sync=%s",
			record["task_id"].Int64(),
			record["delivery_mode"].String(),
			record["delivery_status"].String(),
			record["sync_status"].String(),
		)
		contentRef := patchRef
		if contentRef == "" {
			contentRef = diffSummary
		}

		if contentRef != "" {
			items = append(items, EvidenceItem{
				EvidenceType: "diff",
				SourceType:   "workspace",
				SourceID:     record["id"].Int64(),
				ContentRef:   contentRef,
				Summary:      summary,
			})
		}
		if deliveryRef != "" {
			items = append(items, EvidenceItem{
				EvidenceType: "delivery",
				SourceType:   "workspace",
				SourceID:     record["id"].Int64(),
				ContentRef:   deliveryRef,
				Summary: fmt.Sprintf(
					"delivery task=%d mode=%s title=%s",
					record["task_id"].Int64(),
					record["delivery_mode"].String(),
					record["delivery_title"].String(),
				),
			})
		}
	}

	return items, nil
}

func queryWorkspaceArtifactRecords(ctx context.Context, workflowRunID int64) (gdb.Result, error) {
	return repo.NewTaskWorkspaceRepo().ListArtifactRecordsByWorkflow(ctx, workflowRunID)
}

func isUnknownWorkspaceArtifactColumnErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unknown column")
}

func isDeliveryReferenceArtifactColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "delivery_ref") || strings.Contains(msg, "delivery_title")
}

func (c *EvidenceCollector) collectCIArtifacts(ctx context.Context, in *AcceptContext) ([]EvidenceItem, error) {
	items := make([]EvidenceItem, 0)
	items = append(items, collectCIArtifactFiles(in.WorkDir)...)

	logRecords, err := repo.NewTaskLogRepo().ListRecentByWorkflow(ctx, in.WorkflowRunID, 200, "tl.id", "tl.task_id", "tl.action", "tl.message", "tl.created_at")
	if err != nil {
		return items, err
	}

	for _, record := range logRecords {
		action := g.NewVar(record["action"]).String()
		message := g.NewVar(record["message"]).String()
		if !isCIRelatedLog(action, message) {
			continue
		}

		summary := fmt.Sprintf("[task=%d action=%s] %s", g.NewVar(record["task_id"]).Int64(), action, trimSummary(message, 160))
		items = append(items, EvidenceItem{
			EvidenceType: "ci",
			SourceType:   "task_log",
			SourceID:     g.NewVar(record["id"]).Int64(),
			ContentRef:   trimSummary(message, 500),
			Summary:      summary,
		})
	}

	return items, nil
}

func collectCIArtifactFiles(workDir string) []EvidenceItem {
	if strings.TrimSpace(workDir) == "" {
		return nil
	}

	candidates := append(ciLatestJSONCandidates(workDir),
		filepath.Join(workDir, ".easymvp", "ci", "latest.log"),
		filepath.Join(workDir, ".easymvp", "ci", "latest.txt"),
		filepath.Join(workDir, ".gitlab-ci.yml"),
		filepath.Join(workDir, "Jenkinsfile"),
		filepath.Join(workDir, ".circleci", "config.yml"),
	)

	items := make([]EvidenceItem, 0)
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		summary := fmt.Sprintf("检测到 CI 文件: %s", filepath.Base(path))
		if strings.HasSuffix(path, "latest.json") {
			summary = summarizeCIJSON(path)
		}
		items = append(items, EvidenceItem{
			EvidenceType: "ci",
			SourceType:   "project_repo",
			ContentRef:   path,
			Summary:      summary,
		})
	}

	workflowDir := filepath.Join(workDir, ".github", "workflows")
	if entries, err := os.ReadDir(workflowDir); err == nil {
		count := 0
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
				count++
			}
		}
		if count > 0 {
			items = append(items, EvidenceItem{
				EvidenceType: "ci",
				SourceType:   "project_repo",
				ContentRef:   workflowDir,
				Summary:      fmt.Sprintf("检测到 GitHub Actions 工作流 %d 个", count),
			})
		}
	}

	return items
}

func ciLatestJSONCandidates(workDir string) []string {
	cleanRoot := filepath.Clean(strings.TrimSpace(workDir))
	if cleanRoot == "" || cleanRoot == "." {
		cleanRoot = "."
	}

	seen := make(map[string]struct{}, 4)
	paths := make([]string, 0, 4)
	appendPath := func(base string) {
		base = filepath.Clean(strings.TrimSpace(base))
		if base == "" {
			return
		}
		path := filepath.Join(base, ".easymvp", "ci", "latest.json")
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}

	appendPath(cleanRoot)

	if repoRoot := findGitRepoRootForCI(cleanRoot); repoRoot != "" {
		appendPath(repoRoot)
	}

	if mainWorkDir := resolveMainWorkDirForCI(cleanRoot); mainWorkDir != "" {
		appendPath(mainWorkDir)
		if repoRoot := findGitRepoRootForCI(mainWorkDir); repoRoot != "" {
			appendPath(repoRoot)
		}
	}

	return paths
}

func resolveMainWorkDirForCI(path string) string {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "" {
		return ""
	}
	marker := string(filepath.Separator) + ".mvp-worktrees" + string(filepath.Separator)
	idx := strings.Index(cleanPath, marker)
	if idx < 0 {
		return ""
	}
	return cleanPath[:idx]
}

func findGitRepoRootForCI(start string) string {
	current := filepath.Clean(strings.TrimSpace(start))
	if current == "" {
		return ""
	}
	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() || info.Mode().IsRegular() {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func summarizeCIJSON(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return "检测到 CI 结果文件 latest.json"
	}

	var payload map[string]interface{}
	if json.Unmarshal(content, &payload) != nil {
		return "检测到 CI 结果文件 latest.json"
	}

	parts := make([]string, 0, 4)
	for _, key := range []string{"status", "tool", "pipeline", "summary"} {
		if value := strings.TrimSpace(g.NewVar(payload[key]).String()); value != "" {
			parts = append(parts, key+"="+value)
		}
	}
	if len(parts) == 0 {
		return "检测到 CI 结果文件 latest.json"
	}
	return "CI 结果：" + strings.Join(parts, " ")
}

func isCIRelatedLog(action, message string) bool {
	text := strings.ToLower(strings.TrimSpace(action + " " + message))
	if text == "" {
		return false
	}
	keywords := []string{
		"ci", "build", "compile", "test", "lint", "static check",
		"单元测试", "构建", "编译", "静态检查", "验收脚本",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func trimSummary(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}
