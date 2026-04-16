// Package complete 管理完成阶段：总结、归档、指标计算。
package complete

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/repo"
)

// CompletionSummary 项目完成总结。
type CompletionSummary struct {
	WorkflowRunID   int64             `json:"workflow_run_id"`
	ProjectID       int64             `json:"project_id"`
	TotalTasks      int               `json:"total_tasks"`
	CompletedTasks  int               `json:"completed_tasks"`
	FailedTasks     int               `json:"failed_tasks"`
	EscalatedTasks  int               `json:"escalated_tasks"`
	SkippedTasks    int               `json:"skipped_tasks"`
	SuccessRate     float64           `json:"success_rate"`
	TotalDuration   string            `json:"total_duration"`
	AvgTaskDuration string            `json:"avg_task_duration"`
	StageDurations  map[string]string `json:"stage_durations"`
	ReworkRounds    int               `json:"rework_rounds"`
	HandoffCount    int               `json:"handoff_count"`
	StartedAt       string            `json:"started_at"`
	FinishedAt      string            `json:"finished_at"`
}

// Service 完成阶段服务。
// 事件发布由 StageService.completeWorkflow() 统一负责，此处不持有 publisher。
type Service struct{}

// NewService 创建完成阶段服务。
func NewService() *Service {
	return &Service{}
}

func normalizeDBUTCGTime(value *gtime.Time) *gtime.Time {
	if value == nil || value.IsZero() {
		return nil
	}

	raw := strings.TrimSpace(value.String())
	switch raw {
	case "", "0000-00-00 00:00:00":
		return nil
	}

	for _, layout := range []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999",
		time.DateTime,
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if parsed, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			return gtime.NewFromTime(parsed.In(time.Local))
		}
	}

	return value
}

func applyTaskMetricCounts(summary *CompletionSummary, statusCount map[string]int, skippedCount int) {
	for status, cnt := range statusCount {
		switch status {
		case "completed":
			summary.CompletedTasks = cnt
		case "failed":
			summary.FailedTasks = cnt
		case "escalated":
			summary.EscalatedTasks = cnt
		case "skipped":
			summary.SkippedTasks = cnt
		}
		summary.TotalTasks += cnt
	}
	if skippedCount > summary.SkippedTasks {
		summary.SkippedTasks = skippedCount
	}
	if summary.TotalTasks > 0 {
		summary.SuccessRate = math.Round(float64(summary.CompletedTasks)/float64(summary.TotalTasks)*10000) / 10000
	}
}

func applyTaskMetricRows(summary *CompletionSummary, rows gdb.Result, skippedCount int) {
	statusCount := make(map[string]int, len(rows))
	for _, row := range rows {
		statusCount[row["status"].String()] = row["cnt"].Int()
	}
	applyTaskMetricCounts(summary, statusCount, skippedCount)
}

// Finalize 执行项目完成流程：收集指标 → 生成总结 → 写入 output_ref。
func (s *Service) Finalize(ctx context.Context, stageRunID, workflowRunID int64) error {
	g.Log().Infof(ctx, "[CompleteStage] Finalize stageRunID=%d workflowRunID=%d", stageRunID, workflowRunID)

	// 查 project_id
	workflowRun, err := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "project_id")
	if err != nil {
		return fmt.Errorf("查询 workflow_run 失败: %w", err)
	}

	summary := &CompletionSummary{
		WorkflowRunID: workflowRunID,
		ProjectID:     g.NewVar(workflowRun["project_id"]).Int64(),
	}

	// 1. 收集任务指标
	if err := s.collectTaskMetrics(ctx, workflowRunID, summary); err != nil {
		g.Log().Warningf(ctx, "[CompleteStage] collectTaskMetrics 失败: %v", err)
	}

	// 2. 收集阶段耗时
	if err := s.collectStageMetrics(ctx, workflowRunID, summary); err != nil {
		g.Log().Warningf(ctx, "[CompleteStage] collectStageMetrics 失败: %v", err)
	}

	// 3. 收集返工统计
	if err := s.collectReworkMetrics(ctx, workflowRunID, summary); err != nil {
		g.Log().Warningf(ctx, "[CompleteStage] collectReworkMetrics 失败: %v", err)
	}

	// 4. 收集工作流耗时
	s.collectWorkflowDuration(ctx, workflowRunID, summary)

	// 5. 写入 output_ref
	summaryJSON, marshalErr := json.Marshal(summary)
	if marshalErr != nil {
		g.Log().Warningf(ctx, "[CompleteStage] summary 序列化失败: %v", marshalErr)
		summaryJSON = []byte("{}")
	}
	if err := repo.NewStageRunRepo().UpdateFields(ctx, stageRunID, g.Map{"output_ref": string(summaryJSON)}); err != nil {
		g.Log().Warningf(ctx, "[CompleteStage] 写入 output_ref 失败: %v", err)
	}

	g.Log().Infof(ctx, "[CompleteStage] 总结已生成: tasks=%d/%d success=%.1f%% rework=%d duration=%s",
		summary.CompletedTasks, summary.TotalTasks, summary.SuccessRate*100, summary.ReworkRounds, summary.TotalDuration)

	return nil
}

// collectTaskMetrics 统计 domain_task 各状态数量。
func (s *Service) collectTaskMetrics(ctx context.Context, workflowRunID int64, summary *CompletionSummary) error {
	tasks, err := repo.NewDomainTaskRepo().ListByWorkflowOrdered(ctx, workflowRunID, "status", "result", "started_at", "completed_at")
	if err != nil {
		return err
	}

	statusCount := make(map[string]int)
	var (
		skippedCount int
		totalDur     time.Duration
		durCount     int
	)
	for _, task := range tasks {
		status := g.NewVar(task["status"]).String()
		statusCount[status]++
		if g.NewVar(task["result"]).String() == "skipped" {
			skippedCount++
		}

		startedAt := normalizeDBUTCGTime(g.NewVar(task["started_at"]).GTime())
		completedAt := normalizeDBUTCGTime(g.NewVar(task["completed_at"]).GTime())
		if status == "completed" && startedAt != nil && completedAt != nil {
			totalDur += completedAt.Time.Sub(startedAt.Time)
			durCount++
		}
	}
	applyTaskMetricCounts(summary, statusCount, skippedCount)
	if durCount > 0 {
		summary.AvgTaskDuration = formatDuration(totalDur / time.Duration(durCount))
	}

	return nil
}

// collectStageMetrics 统计各阶段耗时。
func (s *Service) collectStageMetrics(ctx context.Context, workflowRunID int64, summary *CompletionSummary) error {
	stages, err := repo.NewStageRunRepo().ListByWorkflowMaps(ctx, workflowRunID, "stage_type", "started_at", "finished_at")
	if err != nil {
		return err
	}

	stageDurations := make(map[string]time.Duration)
	for _, row := range stages {
		startedAt := normalizeDBUTCGTime(g.NewVar(row["started_at"]).GTime())
		finishedAt := normalizeDBUTCGTime(g.NewVar(row["finished_at"]).GTime())
		if startedAt == nil || finishedAt == nil {
			continue
		}
		stageType := g.NewVar(row["stage_type"]).String()
		stageDurations[stageType] += finishedAt.Time.Sub(startedAt.Time)
	}

	summary.StageDurations = make(map[string]string, len(stageDurations))
	for stageType, duration := range stageDurations {
		summary.StageDurations[stageType] = formatDuration(duration)
	}

	return nil
}

// collectReworkMetrics 统计返工信息。
func (s *Service) collectReworkMetrics(ctx context.Context, workflowRunID int64, summary *CompletionSummary) error {
	stages, err := repo.NewStageRunRepo().ListByWorkflowMaps(ctx, workflowRunID, "stage_type")
	if err != nil {
		return err
	}
	for _, stage := range stages {
		if g.NewVar(stage["stage_type"]).String() == "rework" {
			summary.ReworkRounds++
		}
	}

	handoffCount, err := repo.NewHandoffRecordRepo().CountByWorkflow(ctx, workflowRunID)
	if err != nil {
		return err
	}
	summary.HandoffCount = handoffCount

	return nil
}

// collectWorkflowDuration 收集工作流总耗时。
// 必须在 workflow_run.finished_at 已持久化之后调用，基于 DB 真实值统计。
func (s *Service) collectWorkflowDuration(ctx context.Context, workflowRunID int64, summary *CompletionSummary) {
	wfRun, err := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "started_at", "finished_at")
	if err != nil || len(wfRun) == 0 {
		return
	}

	startedAt := normalizeDBUTCGTime(g.NewVar(wfRun["started_at"]).GTime())
	finishedAt := normalizeDBUTCGTime(g.NewVar(wfRun["finished_at"]).GTime())

	if startedAt != nil {
		summary.StartedAt = startedAt.ISO8601()
	}
	if finishedAt != nil {
		summary.FinishedAt = finishedAt.ISO8601()
	}
	if startedAt != nil && finishedAt != nil {
		summary.TotalDuration = formatDuration(finishedAt.Time.Sub(startedAt.Time))
	}
}

// GetSummary 查询指定项目最近一次完成总结。
func (s *Service) GetSummary(ctx context.Context, projectID int64) (*CompletionSummary, error) {
	wfRun, err := repo.NewWorkflowRunRepo().GetLatestByProjectStatuses(ctx, projectID, []string{"completed"}, "id")
	if err != nil {
		return nil, err
	}
	if len(wfRun) == 0 {
		return nil, fmt.Errorf("项目(%d) 没有已完成的工作流", projectID)
	}

	stageRun, err := repo.NewStageRunRepo().GetLatestByWorkflowAndType(ctx, g.NewVar(wfRun["id"]).Int64(), "complete", "output_ref")
	if err != nil {
		return nil, err
	}
	if stageRun == nil || stageRun["output_ref"].String() == "" {
		return nil, fmt.Errorf("完成总结不存在")
	}

	var summary CompletionSummary
	if err := json.Unmarshal([]byte(stageRun["output_ref"].String()), &summary); err != nil {
		return nil, fmt.Errorf("解析完成总结失败: %w", err)
	}
	return &summary, nil
}

// formatDuration 格式化时长为人类可读。
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-float64(int(d.Minutes()))*60)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) - hours*60
	return fmt.Sprintf("%dh%dm", hours, mins)
}
