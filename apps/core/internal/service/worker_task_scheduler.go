package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/leef-l/easymvp/apps/core/internal/events"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// taskSchedulerWorker listens for PlanCompiled events and automatically launches
// brain runs for all domain tasks, driving the compiled plan → execution pipeline.
type taskSchedulerWorker struct{}

func newTaskSchedulerWorker() backgroundWorker {
	return &taskSchedulerWorker{}
}

func (w *taskSchedulerWorker) Name() string {
	return "task_scheduler_worker"
}

func (w *taskSchedulerWorker) Interval() time.Duration {
	// Periodic scan for compiled projects that haven't been scheduled yet
	// (e.g. after a restart where the event-driven goroutine died).
	return 30 * time.Second
}

func (w *taskSchedulerWorker) RunOnce(ctx context.Context) error {
	// Scan for compiled projects with queued tasks but no brain run bindings.
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "[task_scheduler] RunOnce open db failed: %v", err)
		return nil
	}
	defer closeFn()

	rows, err := db.QueryContext(ctx, `
		SELECT p.id FROM projects p
		WHERE p.status IN ('compiled', 'executing', 'execution_ready')
		AND EXISTS (
			SELECT 1 FROM domain_tasks dt
			WHERE dt.project_id = p.id
			AND dt.status IN ('queued', 'running')
		)
		AND (
			-- No bindings at all (never scheduled)
			NOT EXISTS (
				SELECT 1 FROM brain_run_bindings brb
				WHERE brb.project_id = p.id
			)
			OR
			-- Has running tasks but all bindings are terminal (stale after restart)
			NOT EXISTS (
				SELECT 1 FROM brain_run_bindings brb
				WHERE brb.project_id = p.id
				AND brb.run_status IN ('run_active', 'run_pending')
			)
		)
	`)
	if err != nil {
		g.Log().Warningf(ctx, "[task_scheduler] RunOnce scan failed: %v", err)
		return nil
	}
	defer rows.Close()

	var projectIDs []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			projectIDs = append(projectIDs, pid)
		}
	}

	for _, pid := range projectIDs {
		g.Log().Infof(ctx, "[task_scheduler] RunOne: scheduling missed project=%s", pid)
		go w.scheduleProjectTasks(context.Background(), pid)
	}
	return nil
}

// SubscribeToEvents registers the task scheduler on the global event bus.
func (w *taskSchedulerWorker) SubscribeToEvents() {
	events.Bus().Subscribe(events.PlanReviewApproved, w.handlePlanReviewApproved)
	events.Bus().Subscribe(events.PlanReviewRejected, w.handlePlanReviewRejected)
}

func (w *taskSchedulerWorker) handlePlanReviewApproved(ctx context.Context, evt *events.WorkflowEvent) error {
	projectID := evt.ProjectID
	if projectID == "" {
		return nil
	}

	g.Log().Infof(ctx, "[task_scheduler] PlanReviewApproved for project=%s — auto-compiling and scheduling tasks", projectID)

	// Check if the project already has a compiled plan (e.g. from CreateProject goroutine).
	// If so, skip re-compilation and go straight to scheduling.
	db, closeFn, err := openProjectDatabase(ctx)
	if err == nil {
		project, _ := getProjectByID(ctx, db, projectID)
		if project != nil && project.CurrentCompiledPlanId != "" {
			closeFn()
			g.Log().Infof(ctx, "[task_scheduler] project=%s already compiled (plan=%s), scheduling tasks directly", projectID, project.CurrentCompiledPlanId)
			_ = insertAuditLog(ctx, projectID, "task_scheduler.triggered", "system", "Task scheduler triggered for pre-compiled plan", map[string]any{"compiled_plan_id": project.CurrentCompiledPlanId})
			go w.scheduleProjectTasks(context.Background(), projectID)
			return nil
		}
		closeFn()
	}

	// Auto-compile the plan.
	compiledPlanID, err := compilePlanForProject(ctx, CompilePlanCommand{
		ProjectID:           projectID,
		AutoRedesign:        true,
		MaxRedesignAttempts: 3,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[task_scheduler] auto-compile failed for project=%s: %v", projectID, err)
		return err
	}
	g.Log().Infof(ctx, "[task_scheduler] project=%s compiled plan=%s", projectID, compiledPlanID)

	// Auto-schedule all domain tasks for execution in a background goroutine
	// so the event handler returns immediately.
	go w.scheduleProjectTasks(context.Background(), projectID)
	return nil
}

func (w *taskSchedulerWorker) handlePlanReviewRejected(ctx context.Context, evt *events.WorkflowEvent) error {
	projectID := evt.ProjectID
	if projectID == "" {
		return nil
	}
	g.Log().Infof(ctx, "[task_scheduler] PlanReviewRejected for project=%s — will auto-redesign on next compile attempt", projectID)
	return nil
}

// scheduleProjectTasks loads all domain tasks for a project and executes them
// in DAG order: tasks with no dependencies run in parallel first, then each
// subsequent layer runs only after the previous layer completes.
func (w *taskSchedulerWorker) scheduleProjectTasks(ctx context.Context, projectID string) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "[task_scheduler] open db failed: %v", err)
		return
	}
	defer closeFn()

	tasks, err := listProjectDomainTasks(ctx, db, projectID)
	if err != nil {
		g.Log().Errorf(ctx, "[task_scheduler] list tasks failed: %v", err)
		return
	}
	if len(tasks) == 0 {
		g.Log().Infof(ctx, "[task_scheduler] no domain tasks for project=%s", projectID)
		return
	}

	// Build execution graph from compiled task dependencies.
	graph := buildTaskExecutionGraph(tasks, db)
	layers := graph.topologicalLayers()
	if len(layers) == 0 {
		g.Log().Warningf(ctx, "[task_scheduler] empty execution graph for project=%s", projectID)
		return
	}

	g.Log().Infof(ctx, "[task_scheduler] project=%s executing %d tasks in %d layers",
		projectID, len(tasks), len(layers))

	for layerIdx, layer := range layers {
		g.Log().Infof(ctx, "[task_scheduler] project=%s layer=%d/%d starting %d tasks",
			projectID, layerIdx+1, len(layers), len(layer))

		var wg sync.WaitGroup
		var mu sync.Mutex
		layerErrors := make(map[string]error)
		runIDs := make([]string, 0, len(layer))

		for _, task := range layer {
			// Skip tasks already in a terminal state to support resumption after crashes/restarts.
			status := strings.ToLower(strings.TrimSpace(task.Status))
			if status == "completed" || status == "failed" || status == "cancelled" {
				g.Log().Infof(ctx, "[task_scheduler] project=%s task=%s already terminal (%s), skipping",
					projectID, task.Id, task.Status)
				continue
			}

			wg.Add(1)
			go func(t entity.DomainTasks) {
				defer wg.Done()
				runID, err := w.launchTaskBrainRun(ctx, projectID, t)
				if err != nil {
					mu.Lock()
					layerErrors[t.Id] = err
					mu.Unlock()
					return
				}
				mu.Lock()
				runIDs = append(runIDs, runID)
				mu.Unlock()
			}(task)
		}

		wg.Wait()

		if len(layerErrors) > 0 {
			g.Log().Warningf(ctx, "[task_scheduler] project=%s layer=%d had %d launch failures",
				projectID, layerIdx+1, len(layerErrors))
		}

		// Wait for brain runs in this layer to actually complete before proceeding.
		if len(runIDs) > 0 {
			w.waitForBrainRuns(ctx, projectID, runIDs)
		}

		g.Log().Infof(ctx, "[task_scheduler] project=%s layer=%d/%d completed",
			projectID, layerIdx+1, len(layers))
	}

	g.Log().Infof(ctx, "[task_scheduler] project=%s all task layers executed", projectID)
}

func (w *taskSchedulerWorker) launchTaskBrainRun(ctx context.Context, projectID string, task entity.DomainTasks) (string, error) {
	if task.BrainKind == "" {
		g.Log().Warningf(ctx, "[task_scheduler] task=%s has no brain_kind, skipping", task.Id)
		return "", nil
	}

	// Build prompt from delivery contract if available.
	// All tasks are routed through the Central Brain, which delegates to specialist brains.
	prompt := task.Name
	permissionMode := ""
	originalBrainKind := task.BrainKind
	projectCategory := ""
	if task.SourceCompiledTaskId != "" {
		db, closeFn, err := openProjectDatabase(ctx)
		var compiledTask *entity.WorkflowCompiledTasks
		if err == nil {
			compiledTask, _ = getCompiledTaskByID(ctx, db, task.SourceCompiledTaskId)
			closeFn()
		}
		if compiledTask != nil {
			var contract struct {
				Goal    string `json:"goal"`
				Summary string `json:"summary"`
			}
			_ = json.Unmarshal([]byte(compiledTask.DeliveryContractJson), &contract)
			if contract.Goal != "" {
				prompt = contract.Goal
			} else if contract.Summary != "" {
				prompt = contract.Summary
			}
			permissionMode = inferTaskPermissionMode(compiledTask)
			if originalBrainKind == "easymvp-brain" && compiledTask.RoleType != "" {
				originalBrainKind = mapRoleTypeToBrainKind(compiledTask.RoleType)
			}
		}
	}
	if permissionMode == "" {
		permissionMode = inferPermissionModeFromBrainKind(originalBrainKind)
	}

	// Resolve project workspace and category for richer context.
	workdir := ""
	if proj, err := getProjectByID(ctx, nil, projectID); err == nil && proj != nil {
		workdir = strings.TrimSpace(proj.WorkspaceRoot)
		projectCategory = strings.ToLower(strings.TrimSpace(proj.ProjectCategory))
	}

	// Build a rich orchestrator prompt with task context and workspace path.
	orchestratorPrompt := fmt.Sprintf(
		"[EasyMVP Task]\nProject: %s (category: %s)\nWorkspace: %s\nTask: %s\nSuggested Specialist: %s\n\n"+
		"Execute this task in the workspace directory above. "+
		"If the workspace is empty and this is a setup task, initialize the project scaffold first. "+
		"If it requires specialist capabilities (code writing, testing, etc.), use central.delegate to delegate. "+
		"After completing the task, report what files were created or modified.",
		projectID, projectCategory, workdir, prompt, originalBrainKind,
	)

	result, err := Runtime().StartBrainRun(ctx, StartBrainRunCommand{
		ProjectID:      projectID,
		TaskID:         task.Id,
		BrainKind:      "central",
		Prompt:         orchestratorPrompt,
		Workdir:        workdir,
		MaxTurns:       50,
		Timeout:        "15m",
		Provider:       "default",
		PermissionMode: permissionMode,
	})
	if err != nil {
		return "", err
	}
	g.Log().Infof(ctx, "[task_scheduler] started brain run task=%s run_id=%s via central→%s permission=%s project=%s",
		task.Id, result.RunID, originalBrainKind, permissionMode, projectID)
	return result.RunID, nil
}

// waitForBrainRuns polls the brain serve API until all runs in the layer reach a
// terminal state (completed, failed, or cancelled).
func (w *taskSchedulerWorker) waitForBrainRuns(ctx context.Context, projectID string, runIDs []string) {
	if len(runIDs) == 0 {
		return
	}
	pollInterval := 10 * time.Second
	maxWait := 10 * time.Minute
	deadline := time.Now().Add(maxWait)

	pending := make(map[string]bool, len(runIDs))
	for _, id := range runIDs {
		if id != "" {
			pending[id] = true
		}
	}

	for len(pending) > 0 && time.Now().Before(deadline) {
		for runID := range pending {
			state, err := Runtime().GetBrainRun(ctx, runID)
			if err != nil {
				g.Log().Warningf(ctx, "[task_scheduler] poll run=%s failed: %v", runID, err)
				continue
			}
			status := strings.ToLower(strings.TrimSpace(state.Status))
			if status == "completed" || status == "failed" || status == "cancelled" {
				g.Log().Infof(ctx, "[task_scheduler] run=%s terminal status=%s project=%s", runID, state.Status, projectID)
				delete(pending, runID)
			}
		}
		if len(pending) > 0 {
			time.Sleep(pollInterval)
		}
	}

	if len(pending) > 0 {
		g.Log().Warningf(ctx, "[task_scheduler] project=%s %d runs did not reach terminal status within %v", projectID, len(pending), maxWait)
	}
}

// mapRoleTypeToBrainKind maps a compiled task role_type to the most appropriate specialist brain.
func mapRoleTypeToBrainKind(roleType string) string {
	switch strings.ToLower(strings.TrimSpace(roleType)) {
	case "developer", "implement", "scaffold", "bootstrap", "code":
		return "code"
	case "test", "verify", "audit":
		return "verifier"
	case "browser", "e2e", "ui-check":
		return "browser"
	case "deploy", "release", "ops":
		return "data"
	case "architect", "plan", "review", "orchestrator":
		return "easymvp-brain"
	default:
		return "central"
	}
}

// inferTaskPermissionMode 从 compiled task 推断权限模式。
// 优先使用任务声明的 PermissionMode（在 delivery_contract_json 中），
// 其次根据 role_type 推断。
func inferTaskPermissionMode(task *entity.WorkflowCompiledTasks) string {
	if task == nil {
		return ""
	}
	// 1. 尝试从 delivery_contract_json 中读取声明的权限。
	var contract struct {
		PermissionMode string `json:"permission_mode"`
	}
	if err := json.Unmarshal([]byte(task.DeliveryContractJson), &contract); err == nil && contract.PermissionMode != "" {
		return contract.PermissionMode
	}
	// 2. 根据 role_type 推断。
	roleType := strings.ToLower(strings.TrimSpace(task.RoleType))
	switch roleType {
	case "init", "scaffold", "bootstrap":
		return "bypass-permissions" // 骨架初始化需要 shell_exec 最大权限
	case "implement", "develop", "code":
		return "auto" // 业务代码开发自动接受低风险操作
	case "test", "verify", "audit":
		return "restricted" // 测试验证只读或受限
	case "deploy", "release":
		return "accept-edits" // 部署接受编辑但确认高风险
	}
	return ""
}

// inferPermissionModeFromBrainKind 根据 brain kind 推断默认权限。
func inferPermissionModeFromBrainKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "browser", "verifier":
		return "restricted"
	case "fault":
		return "auto"
	case "code", "data":
		return "auto"
	case "central", "easymvp":
		return "default"
	default:
		return "default"
	}
}

// ---------------------------------------------------------------------------
// Task Execution Graph
// ---------------------------------------------------------------------------

type taskExecutionGraph struct {
	tasks       map[string]entity.DomainTasks
	deps        map[string][]string // taskID -> dependsOn taskIDs
	dependents  map[string][]string // taskID -> tasks that depend on it
}

// matchTaskKey checks whether a compiled task key matches a dependency key.
// It tries exact match, then normalizes by stripping a trailing _001 suffix,
// then falls back to substring containment.
func matchTaskKey(compiledKey, depKey string) bool {
	if compiledKey == depKey {
		return true
	}
	normalized := compiledKey
	if idx := strings.LastIndex(normalized, "_"); idx > 0 {
		suffix := normalized[idx+1:]
		if _, err := strconv.Atoi(suffix); err == nil {
			normalized = normalized[:idx]
		}
	}
	if normalized == depKey {
		return true
	}
	return strings.Contains(normalized, depKey)
}

func buildTaskExecutionGraph(tasks []entity.DomainTasks, db *sql.DB) *taskExecutionGraph {
	g := &taskExecutionGraph{
		tasks:      make(map[string]entity.DomainTasks, len(tasks)),
		deps:       make(map[string][]string, len(tasks)),
		dependents: make(map[string][]string, len(tasks)),
	}
	for _, t := range tasks {
		g.tasks[t.Id] = t
	}
	for _, t := range tasks {
		if t.SourceCompiledTaskId == "" {
			continue
		}
		compiledTask, _ := getCompiledTaskByID(context.Background(), db, t.SourceCompiledTaskId)
		if compiledTask == nil || compiledTask.DependsOnTaskKeysJson == "" {
			continue
		}
		var keys []string
		_ = json.Unmarshal([]byte(compiledTask.DependsOnTaskKeysJson), &keys)
		for _, depKey := range keys {
			// Find the domain task corresponding to this compiled task key.
			// The LLM may use draft task keys (e.g. "game_logic") while the
			// compiled task key is name-based (e.g. "implement_game_logic_002").
			// Use fuzzy matching: exact, normalized (strip _001), or substring.
			for _, dt := range tasks {
				if matchTaskKey(dt.SourceTaskKey, depKey) {
					g.deps[t.Id] = append(g.deps[t.Id], dt.Id)
					g.dependents[dt.Id] = append(g.dependents[dt.Id], t.Id)
					break
				}
			}
		}
	}
	return g
}

// topologicalLayers returns tasks grouped by dependency depth (Kahn's algorithm).
// Layer 0 = tasks with no dependencies. Each subsequent layer contains tasks whose
// dependencies are all in previous layers.
func (g *taskExecutionGraph) topologicalLayers() [][]entity.DomainTasks {
	inDegree := make(map[string]int, len(g.tasks))
	for id := range g.tasks {
		inDegree[id] = len(g.deps[id])
	}

	var layers [][]entity.DomainTasks
	visited := make(map[string]bool)

	for len(visited) < len(g.tasks) {
		var layer []entity.DomainTasks
		for id, task := range g.tasks {
			if visited[id] {
				continue
			}
			// Count how many dependencies are still unvisited.
			unresolved := 0
			for _, dep := range g.deps[id] {
				if !visited[dep] {
					unresolved++
				}
			}
			if unresolved == 0 {
				layer = append(layer, task)
			}
		}
		if len(layer) == 0 {
			// Cycle detected or stuck — break to avoid infinite loop.
			break
		}
		for _, t := range layer {
			visited[t.Id] = true
		}
		layers = append(layers, layer)
	}

	return layers
}
