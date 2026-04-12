package hook

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

	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/utility/snowflake"
)

// archiveSummary 工作流归档总结。
type archiveSummary struct {
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

// ArchiveHook 执行指标统计和总结生成，将结果写入 mvp_workflow_event。
// 与 complete.Service.Finalize 逻辑相同，但不依赖 stage_run_id。
type ArchiveHook struct{}

// Name 返回 hook 名称。
func (h *ArchiveHook) Name() string { return "archive" }

// Execute 收集任务指标、阶段耗时、返工统计、工作流总耗时，将总结写入 workflow_event。
func (h *ArchiveHook) Execute(ctx context.Context, workflowRunID int64) error {
	g.Log().Infof(ctx, "[Hook:archive] 开始归档 workflowRunID=%d", workflowRunID)

	// 查 project_id
	projectID, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	if err != nil {
		return fmt.Errorf("查询 workflow_run 失败: %w", err)
	}

	summary := &archiveSummary{
		WorkflowRunID: workflowRunID,
		ProjectID:     projectID.Int64(),
	}

	// 1. 收集任务指标
	if err := collectArchiveTaskMetrics(ctx, workflowRunID, summary); err != nil {
		g.Log().Warningf(ctx, "[Hook:archive] collectTaskMetrics 失败: %v", err)
	}

	// 2. 收集阶段耗时
	if err := collectArchiveStageMetrics(ctx, workflowRunID, summary); err != nil {
		g.Log().Warningf(ctx, "[Hook:archive] collectStageMetrics 失败: %v", err)
	}

	// 3. 收集返工统计
	if err := collectArchiveReworkMetrics(ctx, workflowRunID, summary); err != nil {
		g.Log().Warningf(ctx, "[Hook:archive] collectReworkMetrics 失败: %v", err)
	}

	// 4. 收集工作流耗时
	collectArchiveWorkflowDuration(ctx, workflowRunID, summary)

	// 5. 将总结写入 mvp_workflow_event
	summaryJSON, marshalErr := json.Marshal(summary)
	if marshalErr != nil {
		g.Log().Warningf(ctx, "[Hook:archive] summary 序列化失败: %v", marshalErr)
		summaryJSON = []byte("{}")
	}

	eventID := int64(snowflake.Generate())
	evt := event.Event{
		WorkflowRunID: workflowRunID,
		EntityType:    event.EntityWorkflowRun,
		EntityID:      &workflowRunID,
		EventType:     "workflow.archived",
		Payload:       string(summaryJSON),
	}
	evt = evt.EnsureMetadata()
	_, insertErr := g.DB().Model("mvp_workflow_event").Ctx(ctx).Insert(g.Map{
		"id":              eventID,
		"event_id":        evt.EventID,
		"workflow_run_id": evt.WorkflowRunID,
		"entity_type":     evt.EntityType,
		"entity_id":       evt.EntityID,
		"event_type":      evt.EventType,
		"payload":         evt.Payload,
		"created_at":      time.Now(),
	})
	if insertErr != nil {
		g.Log().Warningf(ctx, "[Hook:archive] 写入 workflow_event 失败: %v", insertErr)
	}

	g.Log().Infof(ctx, "[Hook:archive] 归档完成: tasks=%d/%d success=%.1f%% rework=%d duration=%s",
		summary.CompletedTasks, summary.TotalTasks, summary.SuccessRate*100, summary.ReworkRounds, summary.TotalDuration)

	return nil
}

func collectArchiveTaskMetrics(ctx context.Context, workflowRunID int64, summary *archiveSummary) error {
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("status, COUNT(*) as cnt").
		Group("status").
		All()
	if err != nil {
		return err
	}

	skippedCount, skippedErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("result", "skipped").
		WhereNull("deleted_at").
		Count()
	if skippedErr != nil {
		return skippedErr
	}

	applyArchiveTaskMetricRows(summary, tasks, skippedCount)

	// 平均任务耗时
	avgDur, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", "completed").
		Where("started_at IS NOT NULL").
		Where("completed_at IS NOT NULL").
		WhereNull("deleted_at").
		Value("AVG(TIMESTAMPDIFF(SECOND, started_at, completed_at))")
	if err == nil && avgDur.Float64() > 0 {
		summary.AvgTaskDuration = archiveFormatDuration(time.Duration(avgDur.Float64()) * time.Second)
	}

	return nil
}

func applyArchiveTaskMetricRows(summary *archiveSummary, rows gdb.Result, skippedCount int) {
	for _, row := range rows {
		cnt := row["cnt"].Int()
		switch row["status"].String() {
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

func collectArchiveStageMetrics(ctx context.Context, workflowRunID int64, summary *archiveSummary) error {
	stages, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("started_at IS NOT NULL").
		Where("finished_at IS NOT NULL").
		WhereNull("deleted_at").
		Fields("stage_type, SUM(TIMESTAMPDIFF(SECOND, started_at, finished_at)) as dur").
		Group("stage_type").
		All()
	if err != nil {
		return err
	}

	summary.StageDurations = make(map[string]string, len(stages))
	for _, row := range stages {
		dur := time.Duration(row["dur"].Int64()) * time.Second
		summary.StageDurations[row["stage_type"].String()] = archiveFormatDuration(dur)
	}

	return nil
}

func collectArchiveReworkMetrics(ctx context.Context, workflowRunID int64, summary *archiveSummary) error {
	reworkCount, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", "rework").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	summary.ReworkRounds = reworkCount

	handoffCount, err := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Count()
	if err != nil {
		return err
	}
	summary.HandoffCount = handoffCount

	return nil
}

func collectArchiveWorkflowDuration(ctx context.Context, workflowRunID int64, summary *archiveSummary) {
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		Fields("started_at, finished_at").
		One()
	if err != nil || wfRun.IsEmpty() {
		return
	}

	startedAt := normalizeArchiveGTime(wfRun["started_at"].GTime())
	finishedAt := normalizeArchiveGTime(wfRun["finished_at"].GTime())

	if startedAt != nil {
		summary.StartedAt = startedAt.ISO8601()
	}
	if finishedAt != nil {
		summary.FinishedAt = finishedAt.ISO8601()
	}
	if startedAt != nil && finishedAt != nil {
		summary.TotalDuration = archiveFormatDuration(finishedAt.Time.Sub(startedAt.Time))
	}
}

func normalizeArchiveGTime(value *gtime.Time) *gtime.Time {
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

func archiveFormatDuration(d time.Duration) string {
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
