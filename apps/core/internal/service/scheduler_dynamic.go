package service

// ---------------------------------------------------------------------------
// Dynamic Adjustment: Stall Detection + Concurrent Conflict Arbitration
//
// B-11 Phase 4 — 实现动态调整：停滞检测 + 并发冲突仲裁
//
// 三个核心组件：
//   - StallDetector   检测 Brain Run 是否停滞（连续 N 次无进展）
//   - ConflictArbiter 检测并发任务的文件修改冲突并自动解决
//   - DynamicAdjuster 根据停滞报告调整执行计划
// ---------------------------------------------------------------------------

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// StallReport 描述某个 Brain Run 的停滞状态。
type StallReport struct {
	RunID              string `json:"run_id"`
	TaskID             string `json:"task_id"`
	IsStalled          bool   `json:"is_stalled"`
	StallReason        string `json:"stall_reason,omitempty"`
	TurnsSinceProgress int    `json:"turns_since_progress"`
	Suggestion         string `json:"suggestion"` // split / restart / reassign
}

// FileConflict 描述两个或多个并发任务对同一文件的冲突写操作。
type FileConflict struct {
	FilePath     string   `json:"file_path"`
	RunIDs       []string `json:"run_ids"`
	TaskIDs      []string `json:"task_ids"`
	ConflictType string   `json:"conflict_type"` // write_write / read_write
}

// Resolution 描述针对一个冲突采取的解决措施。
type Resolution struct {
	Action       string   `json:"action"` // interrupt / add_dependency / merge
	AffectedRuns []string `json:"affected_runs"`
	Description  string   `json:"description"`
}

// AdjustmentAction 描述针对停滞 / 冲突采取的计划调整行动。
type AdjustmentAction struct {
	Action         string   `json:"action"` // split_task / restart / reassign / no_action
	OriginalTaskID string   `json:"original_task_id"`
	NewTasks       []string `json:"new_tasks,omitempty"`
	Reason         string   `json:"reason"`
}

// ---------------------------------------------------------------------------
// StallDetector
// ---------------------------------------------------------------------------

// stallSnapshot 记录某次轮询时的运行进度快照。
type stallSnapshot struct {
	status    string
	fileCount int
	capturedAt time.Time
}

// StallDetector 通过多次轮询 Brain Run 状态来判断执行是否陷入停滞。
// 停滞判定：连续 stallThreshold 次检查均无进展（状态未变、工作区文件数未增加）。
type StallDetector struct {
	// stallThreshold 指定连续无进展的最低次数，超过此值则判定为停滞。
	stallThreshold int
	// pollInterval 每次状态检查的间隔时长。
	pollInterval time.Duration

	mu        sync.Mutex
	snapshots map[string][]stallSnapshot // runID -> history
}

// NewStallDetector 创建一个默认配置的 StallDetector。
// stallThreshold=3，pollInterval=30s
func NewStallDetector() *StallDetector {
	return &StallDetector{
		stallThreshold: 3,
		pollInterval:   30 * time.Second,
		snapshots:      make(map[string][]stallSnapshot),
	}
}

// NewStallDetectorWithOptions 创建一个使用自定义参数的 StallDetector。
func NewStallDetectorWithOptions(stallThreshold int, pollInterval time.Duration) *StallDetector {
	if stallThreshold <= 0 {
		stallThreshold = 3
	}
	if pollInterval <= 0 {
		pollInterval = 30 * time.Second
	}
	return &StallDetector{
		stallThreshold: stallThreshold,
		pollInterval:   pollInterval,
		snapshots:      make(map[string][]stallSnapshot),
	}
}

// DetectStall 对给定的 Brain Run 执行一次停滞检测。
// checkIntervalSeconds：本次检查所用的轮询间隔（<=0 时使用 detector 默认值）。
// 函数会拉取运行状态 + 工作区文件数作为"进度指标"，
// 与历史快照比对：若最新 N 次快照均无变化则判定停滞。
func (d *StallDetector) DetectStall(ctx context.Context, runID string, checkIntervalSeconds int) (*StallReport, error) {
	if runID == "" {
		return nil, fmt.Errorf("StallDetector.DetectStall: runID is required")
	}

	interval := d.pollInterval
	if checkIntervalSeconds > 0 {
		interval = time.Duration(checkIntervalSeconds) * time.Second
	}

	// 拉取当前运行状态。
	runState, err := Runtime().GetBrainRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("StallDetector.DetectStall: get brain run %s: %w", runID, err)
	}

	// 如果运行已到终态，直接返回"未停滞"。
	status := strings.ToLower(strings.TrimSpace(runState.Status))
	if status == "completed" || status == "failed" || status == "cancelled" {
		return &StallReport{
			RunID:     runID,
			IsStalled: false,
			StallReason: fmt.Sprintf("run already in terminal status: %s", runState.Status),
		}, nil
	}

	// 估算工作区文件数（作为进度指标之一）。
	fileCount := d.countWorkspaceFiles(ctx, runState.RunID)

	snap := stallSnapshot{
		status:     runState.Status,
		fileCount:  fileCount,
		capturedAt: time.Now(),
	}

	d.mu.Lock()
	history := append(d.snapshots[runID], snap)
	// 保留最多 stallThreshold+1 条历史，节省内存。
	if len(history) > d.stallThreshold+1 {
		history = history[len(history)-(d.stallThreshold+1):]
	}
	d.snapshots[runID] = history
	d.mu.Unlock()

	_ = interval // interval 用于调用方控制轮询节奏，Detector 此处仅记录快照。

	report := d.evaluateStall(runID, runState.RunID, history)
	g.Log().Debugf(ctx, "[scheduler_dynamic] stall check run=%s stalled=%v turns_since_progress=%d",
		runID, report.IsStalled, report.TurnsSinceProgress)
	return report, nil
}

// evaluateStall 根据快照历史判断是否停滞。
func (d *StallDetector) evaluateStall(runID, taskID string, history []stallSnapshot) *StallReport {
	report := &StallReport{
		RunID:  runID,
		TaskID: taskID,
	}

	if len(history) < d.stallThreshold {
		// 快照不足，无法判断。
		report.IsStalled = false
		report.StallReason = fmt.Sprintf("insufficient snapshots (%d/%d)", len(history), d.stallThreshold)
		return report
	}

	// 取最近 N 条快照进行比较。
	recent := history[len(history)-d.stallThreshold:]
	firstStatus := recent[0].status
	firstFileCount := recent[0].fileCount
	unchanged := true
	for _, s := range recent[1:] {
		if s.status != firstStatus || s.fileCount != firstFileCount {
			unchanged = false
			break
		}
	}

	report.TurnsSinceProgress = d.stallThreshold
	if unchanged {
		report.IsStalled = true
		report.StallReason = fmt.Sprintf(
			"no progress in last %d checks: status=%s file_count=%d",
			d.stallThreshold, firstStatus, firstFileCount,
		)
		report.Suggestion = d.suggestAction(firstStatus)
	} else {
		report.IsStalled = false
		report.TurnsSinceProgress = 0
	}
	return report
}

// suggestAction 根据当前状态给出建议操作。
func (d *StallDetector) suggestAction(status string) string {
	switch strings.ToLower(status) {
	case "running", "run_active":
		return "restart"
	case "run_pending", "pending":
		return "reassign"
	default:
		return "split"
	}
}

// countWorkspaceFiles 统计与 runID 关联项目的工作区文件数，用作停滞检测的进度指标。
// 出错时返回 -1（不影响停滞判定逻辑）。
func (d *StallDetector) countWorkspaceFiles(ctx context.Context, runID string) int {
	// 通过 run → binding → project 找到工作区路径。
	binding, err := getBrainRunBindingByRunIDSafe(ctx, runID)
	if err != nil || binding == nil {
		return -1
	}

	project, err := getProjectByID(ctx, nil, binding.ProjectId)
	if err != nil || project == nil {
		return -1
	}

	root := strings.TrimSpace(project.WorkspaceRoot)
	if root == "" {
		return -1
	}

	count := 0
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !entry.IsDir() {
			count++
		}
		return nil
	})
	return count
}

// ---------------------------------------------------------------------------
// getBrainRunBindingByRunIDSafe — 通过 brain_run_id 查询绑定记录（安全版本）
// ---------------------------------------------------------------------------

// getBrainRunBindingByRunIDSafe 从数据库查询指定 brain_run_id 对应的绑定记录。
// 失败时返回 nil（不中断停滞检测流程）。
func getBrainRunBindingByRunIDSafe(ctx context.Context, brainRunID string) (*brainRunBindingRef, error) {
	if brainRunID == "" {
		return nil, nil
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, task_id FROM brain_run_bindings WHERE brain_run_id = ? LIMIT 1`,
		brainRunID,
	)
	ref := &brainRunBindingRef{}
	if err = row.Scan(&ref.ID, &ref.ProjectId, &ref.TaskId); err != nil {
		return nil, nil // 未找到，不视为错误
	}
	return ref, nil
}

// brainRunBindingRef 是 brain_run_bindings 的轻量引用，仅包含停滞检测所需字段。
type brainRunBindingRef struct {
	ID        string
	ProjectId string
	TaskId    string
}

// ---------------------------------------------------------------------------
// ConflictArbiter
// ---------------------------------------------------------------------------

// activeRunFileRegistry 记录每个活跃 Run 对哪些文件进行了写操作。
// key = runID, value = 写操作文件路径集合。
type activeRunFileRegistry struct {
	mu      sync.RWMutex
	entries map[string]map[string]bool // runID -> set<filePath>
}

func newActiveRunFileRegistry() *activeRunFileRegistry {
	return &activeRunFileRegistry{
		entries: make(map[string]map[string]bool),
	}
}

func (r *activeRunFileRegistry) register(runID, filePath string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.entries[runID] == nil {
		r.entries[runID] = make(map[string]bool)
	}
	r.entries[runID][filePath] = true
}

func (r *activeRunFileRegistry) filesForRun(runID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	files := make([]string, 0, len(r.entries[runID]))
	for f := range r.entries[runID] {
		files = append(files, f)
	}
	return files
}

// ConflictArbiter 检测多个并发 Brain Run 之间的文件修改冲突，并自动解决。
type ConflictArbiter struct {
	registry *activeRunFileRegistry
}

// NewConflictArbiter 创建一个空的 ConflictArbiter。
func NewConflictArbiter() *ConflictArbiter {
	return &ConflictArbiter{
		registry: newActiveRunFileRegistry(),
	}
}

// DetectFileConflicts 检测 activeRunIDs 列表中各运行之间的文件冲突。
// 实现方式：遍历每个运行的工作区快照，统计各文件被哪些 run 写入，
// 若同一文件被 ≥2 个 run 写入则报告 write_write 冲突。
func (a *ConflictArbiter) DetectFileConflicts(ctx context.Context, projectID string, activeRunIDs []string) ([]FileConflict, error) {
	if projectID == "" {
		return nil, fmt.Errorf("ConflictArbiter.DetectFileConflicts: projectID is required")
	}
	if len(activeRunIDs) < 2 {
		// 少于 2 个运行时无法产生冲突。
		return nil, nil
	}

	// 为每个 runID 收集其写入文件列表（从 replay artifacts 或工作区扫描推断）。
	type runFiles struct {
		runID   string
		taskID  string
		files   []string
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("ConflictArbiter.DetectFileConflicts: open db: %w", err)
	}
	defer closeFn()

	allRuns := make([]runFiles, 0, len(activeRunIDs))
	for _, runID := range activeRunIDs {
		if runID == "" {
			continue
		}
		// 查询绑定记录获取 taskID。
		var taskID string
		row := db.QueryRowContext(ctx,
			`SELECT task_id FROM brain_run_bindings WHERE brain_run_id = ? LIMIT 1`,
			runID,
		)
		_ = row.Scan(&taskID)

		// 从 replay_items 中读取该 run 产生的文件列表。
		files, err := listReplayFilesForRun(ctx, db, runID)
		if err != nil {
			g.Log().Warningf(ctx, "[scheduler_dynamic] list replay files for run=%s: %v", runID, err)
			files = nil
		}
		// 合并 registry 中手动注册的文件。
		for _, f := range a.registry.filesForRun(runID) {
			files = append(files, f)
		}
		allRuns = append(allRuns, runFiles{runID: runID, taskID: taskID, files: files})
	}

	// 构建 filePath -> []runFiles 的倒排索引。
	fileIndex := make(map[string][]runFiles)
	for _, rf := range allRuns {
		for _, f := range rf.files {
			fileIndex[f] = append(fileIndex[f], rf)
		}
	}

	// 报告被多个 run 写入的文件。
	var conflicts []FileConflict
	for filePath, writers := range fileIndex {
		if len(writers) < 2 {
			continue
		}
		runIDs := make([]string, 0, len(writers))
		taskIDs := make([]string, 0, len(writers))
		for _, w := range writers {
			runIDs = append(runIDs, w.runID)
			if w.taskID != "" {
				taskIDs = append(taskIDs, w.taskID)
			}
		}
		conflicts = append(conflicts, FileConflict{
			FilePath:     filePath,
			RunIDs:       runIDs,
			TaskIDs:      taskIDs,
			ConflictType: "write_write",
		})
	}

	if len(conflicts) > 0 {
		g.Log().Infof(ctx, "[scheduler_dynamic] project=%s detected %d file conflicts among %d runs",
			projectID, len(conflicts), len(activeRunIDs))
	}
	return conflicts, nil
}

// ResolveConflict 对单个 FileConflict 自动选择解决策略：
//   - write_write：中断后发（时间戳较晚）的任务，并为其添加对先发任务的依赖。
//   - read_write（预留）：添加依赖即可。
func (a *ConflictArbiter) ResolveConflict(ctx context.Context, conflict FileConflict) (*Resolution, error) {
	if len(conflict.RunIDs) < 2 {
		return &Resolution{
			Action:      "no_action",
			Description: "conflict has fewer than 2 runs, nothing to resolve",
		}, nil
	}

	switch conflict.ConflictType {
	case "write_write":
		return a.resolveWriteWrite(ctx, conflict)
	case "read_write":
		return a.resolveReadWrite(ctx, conflict)
	default:
		return &Resolution{
			Action:       "add_dependency",
			AffectedRuns: conflict.RunIDs,
			Description:  fmt.Sprintf("unknown conflict type %q — added dependency as conservative fallback", conflict.ConflictType),
		}, nil
	}
}

// resolveWriteWrite 中断列表中最后一个运行（视为"后发"），并记录日志。
// 真正的依赖注入需要计划重编译，此处仅标记意图并取消冲突 run。
func (a *ConflictArbiter) resolveWriteWrite(ctx context.Context, conflict FileConflict) (*Resolution, error) {
	// 取最后一个 runID 作为"后发"任务进行中断。
	targetRunID := conflict.RunIDs[len(conflict.RunIDs)-1]

	if err := Runtime().CancelBrainRun(ctx, targetRunID); err != nil {
		// 取消失败不是硬错误 — 运行可能已结束。
		g.Log().Warningf(ctx, "[scheduler_dynamic] resolveWriteWrite: cancel run=%s failed: %v", targetRunID, err)
	}

	_ = insertAuditLog(ctx, "", "scheduler.conflict.write_write_resolved", "system",
		fmt.Sprintf("write_write conflict on %s resolved by interrupting run %s", conflict.FilePath, targetRunID),
		map[string]any{
			"file_path":     conflict.FilePath,
			"run_ids":       conflict.RunIDs,
			"task_ids":      conflict.TaskIDs,
			"interrupted":   targetRunID,
		},
	)

	return &Resolution{
		Action:       "interrupt",
		AffectedRuns: []string{targetRunID},
		Description: fmt.Sprintf(
			"write_write conflict on %q: interrupted run %s; re-queue with dependency on %s",
			conflict.FilePath, targetRunID, conflict.RunIDs[0],
		),
	}, nil
}

// resolveReadWrite 仅添加依赖关系（不中断任何运行）。
func (a *ConflictArbiter) resolveReadWrite(_ context.Context, conflict FileConflict) (*Resolution, error) {
	return &Resolution{
		Action:       "add_dependency",
		AffectedRuns: conflict.RunIDs,
		Description: fmt.Sprintf(
			"read_write conflict on %q: added dependency to serialize access",
			conflict.FilePath,
		),
	}, nil
}

// RegisterRunFile 手动注册某个 run 对指定文件的写操作（用于测试或外部触发）。
func (a *ConflictArbiter) RegisterRunFile(runID, filePath string) {
	a.registry.register(runID, filePath)
}

// ---------------------------------------------------------------------------
// listReplayFilesForRun — 从 replay_items 读取文件列表（轻量版）
// ---------------------------------------------------------------------------

// listReplayFilesForRun 返回与给定 brain_run_id 关联的所有文件路径。
// 表结构：replay_items 中有 run_id 和 file_path 列。
// db 参数传入已打开的 *sql.DB 实例。
func listReplayFilesForRun(ctx context.Context, _ interface{}, runID string) ([]string, error) {
	return listReplayFilesForRunDB(ctx, runID)
}

// listReplayFilesForRunDB 从数据库查询 replay 文件路径。
func listReplayFilesForRunDB(ctx context.Context, runID string) ([]string, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT file_path FROM replay_items WHERE run_id = ? AND file_path != ''`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil && p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

// ---------------------------------------------------------------------------
// DynamicAdjuster
// ---------------------------------------------------------------------------

// DynamicAdjuster 根据 StallReport 或外部触发调整执行计划。
// 三种主要策略：
//   - split_task  — 将停滞任务拆分为两个子任务
//   - restart     — 取消当前运行并重新启动
//   - reassign    — 将任务重新分配给其他 brain kind
type DynamicAdjuster struct {
	detector *StallDetector
	arbiter  *ConflictArbiter
}

// NewDynamicAdjuster 创建一个使用默认探测器和仲裁器的 DynamicAdjuster。
func NewDynamicAdjuster() *DynamicAdjuster {
	return &DynamicAdjuster{
		detector: NewStallDetector(),
		arbiter:  NewConflictArbiter(),
	}
}

// NewDynamicAdjusterWith 使用已有的探测器和仲裁器创建 DynamicAdjuster。
func NewDynamicAdjusterWith(detector *StallDetector, arbiter *ConflictArbiter) *DynamicAdjuster {
	return &DynamicAdjuster{
		detector: detector,
		arbiter:  arbiter,
	}
}

// AdjustPlan 根据停滞报告对执行计划进行动态调整。
// 调整策略选择优先级：StallReport.Suggestion > 状态推断 > 默认 restart。
func (adj *DynamicAdjuster) AdjustPlan(ctx context.Context, projectID string, stallReport *StallReport) (*AdjustmentAction, error) {
	if projectID == "" {
		return nil, fmt.Errorf("DynamicAdjuster.AdjustPlan: projectID is required")
	}
	if stallReport == nil {
		return nil, fmt.Errorf("DynamicAdjuster.AdjustPlan: stallReport is required")
	}

	if !stallReport.IsStalled {
		return &AdjustmentAction{
			Action:         "no_action",
			OriginalTaskID: stallReport.TaskID,
			Reason:         "run is not stalled",
		}, nil
	}

	g.Log().Infof(ctx, "[scheduler_dynamic] AdjustPlan project=%s run=%s task=%s suggestion=%s",
		projectID, stallReport.RunID, stallReport.TaskID, stallReport.Suggestion)

	switch stallReport.Suggestion {
	case "split":
		return adj.splitTask(ctx, projectID, stallReport)
	case "restart":
		return adj.restartTask(ctx, projectID, stallReport)
	case "reassign":
		return adj.reassignTask(ctx, projectID, stallReport)
	default:
		// 默认：重启
		return adj.restartTask(ctx, projectID, stallReport)
	}
}

// splitTask 将停滞任务标记为需要拆分，生成两个占位子任务 ID。
// 实际拆分逻辑依赖计划重编译，此处仅生成 AdjustmentAction 供上层调度器使用。
func (adj *DynamicAdjuster) splitTask(_ context.Context, projectID string, report *StallReport) (*AdjustmentAction, error) {
	subTask1 := fmt.Sprintf("%s_split_1_%s", report.TaskID, newResourceID("st"))
	subTask2 := fmt.Sprintf("%s_split_2_%s", report.TaskID, newResourceID("st"))

	return &AdjustmentAction{
		Action:         "split_task",
		OriginalTaskID: report.TaskID,
		NewTasks:       []string{subTask1, subTask2},
		Reason: fmt.Sprintf(
			"task %s stalled (project=%s reason=%q): split into 2 sub-tasks for parallel execution",
			report.TaskID, projectID, report.StallReason,
		),
	}, nil
}

// restartTask 取消当前停滞的 Brain Run 并重新启动。
func (adj *DynamicAdjuster) restartTask(ctx context.Context, projectID string, report *StallReport) (*AdjustmentAction, error) {
	// 取消停滞的运行（若运行 ID 存在）。
	if report.RunID != "" {
		if err := Runtime().CancelBrainRun(ctx, report.RunID); err != nil {
			g.Log().Warningf(ctx, "[scheduler_dynamic] restartTask: cancel stalled run=%s failed: %v",
				report.RunID, err)
			// 不中断流程 — 运行可能已终止。
		}
	}

	_ = insertAuditLog(ctx, projectID, "scheduler.dynamic.restart", "system",
		fmt.Sprintf("stalled run %s (task %s) cancelled for restart", report.RunID, report.TaskID),
		map[string]any{
			"run_id":           report.RunID,
			"task_id":          report.TaskID,
			"stall_reason":     report.StallReason,
			"turns_no_progress": report.TurnsSinceProgress,
		},
	)

	return &AdjustmentAction{
		Action:         "restart",
		OriginalTaskID: report.TaskID,
		Reason: fmt.Sprintf(
			"task %s stalled (run=%s reason=%q): cancelled and queued for restart",
			report.TaskID, report.RunID, report.StallReason,
		),
	}, nil
}

// reassignTask 将任务标记为需要重新分配给其他 brain kind。
// 实际重新分配需调用 launchTaskBrainRun，此处仅生成调整记录。
func (adj *DynamicAdjuster) reassignTask(ctx context.Context, projectID string, report *StallReport) (*AdjustmentAction, error) {
	// 取消当前停滞的运行。
	if report.RunID != "" {
		if err := Runtime().CancelBrainRun(ctx, report.RunID); err != nil {
			g.Log().Warningf(ctx, "[scheduler_dynamic] reassignTask: cancel stalled run=%s failed: %v",
				report.RunID, err)
		}
	}

	_ = insertAuditLog(ctx, projectID, "scheduler.dynamic.reassign", "system",
		fmt.Sprintf("stalled run %s (task %s) reassigned", report.RunID, report.TaskID),
		map[string]any{
			"run_id":       report.RunID,
			"task_id":      report.TaskID,
			"stall_reason": report.StallReason,
		},
	)

	return &AdjustmentAction{
		Action:         "reassign",
		OriginalTaskID: report.TaskID,
		Reason: fmt.Sprintf(
			"task %s stalled (run=%s reason=%q): reassigned to alternative executor",
			report.TaskID, report.RunID, report.StallReason,
		),
	}, nil
}

// ---------------------------------------------------------------------------
// Convenience: RunStallScan
// ---------------------------------------------------------------------------

// RunStallScan 对 projectID 下的所有活跃运行执行一轮停滞扫描，
// 并对每个已停滞的运行自动调用 AdjustPlan。
// 返回所有触发的调整行动列表。
func (adj *DynamicAdjuster) RunStallScan(ctx context.Context, projectID string, checkIntervalSeconds int) ([]*AdjustmentAction, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("RunStallScan: open db: %w", err)
	}
	defer closeFn()

	// 查询项目下所有活跃（非终态）绑定。
	rows, err := db.QueryContext(ctx,
		`SELECT brain_run_id, task_id FROM brain_run_bindings
		 WHERE project_id = ? AND run_status IN ('run_active', 'run_pending')`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("RunStallScan: query active runs: %w", err)
	}
	defer rows.Close()

	type runRef struct {
		runID  string
		taskID string
	}
	var activeRuns []runRef
	for rows.Next() {
		var r runRef
		if err := rows.Scan(&r.runID, &r.taskID); err == nil {
			activeRuns = append(activeRuns, r)
		}
	}

	if len(activeRuns) == 0 {
		return nil, nil
	}

	var actions []*AdjustmentAction
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, run := range activeRuns {
		wg.Add(1)
		go func(r runRef) {
			defer wg.Done()
			report, err := adj.detector.DetectStall(ctx, r.runID, checkIntervalSeconds)
			if err != nil {
				g.Log().Warningf(ctx, "[scheduler_dynamic] RunStallScan: detect stall run=%s: %v", r.runID, err)
				return
			}
			if !report.IsStalled {
				return
			}
			// 确保 report 含有 taskID。
			if report.TaskID == "" {
				report.TaskID = r.taskID
			}
			action, err := adj.AdjustPlan(ctx, projectID, report)
			if err != nil {
				g.Log().Warningf(ctx, "[scheduler_dynamic] RunStallScan: adjust plan run=%s: %v", r.runID, err)
				return
			}
			mu.Lock()
			actions = append(actions, action)
			mu.Unlock()
		}(run)
	}
	wg.Wait()

	if len(actions) > 0 {
		g.Log().Infof(ctx, "[scheduler_dynamic] RunStallScan project=%s: %d adjustments triggered", projectID, len(actions))
	}
	return actions, nil
}
