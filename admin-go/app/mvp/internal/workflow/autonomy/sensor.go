package autonomy

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/repo"
)

// Sensor 态势感知器。
type Sensor struct {
	snapshotRepo *repo.SituationSnapshotRepo
}

func NewSensor(snapshotRepo *repo.SituationSnapshotRepo) *Sensor {
	return &Sensor{snapshotRepo: snapshotRepo}
}

// Perceive 读取 workflow 当前态势。
func (s *Sensor) Perceive(ctx context.Context, workflowRunID int64) (*Situation, error) {
	if workflowRunID == 0 {
		return nil, nil
	}
	wf, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Fields("id, project_id, status, current_stage, started_at, tokens_consumed, replan_count").
		Where("id", workflowRunID).
		WhereNull("deleted_at").
		One()
	if err != nil || wf.IsEmpty() {
		return nil, err
	}

	projectID := wf["project_id"].Int64()
	projectRow, projectErr := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("project_category, category_code").
		Where("id", projectID).
		WhereNull("deleted_at").
		One()
	if projectErr != nil {
		g.Log().Warningf(ctx, "[Sensor] 查询项目分类失败: projectID=%d err=%v", projectID, projectErr)
	}
	categoryCode := projectRow["category_code"].String()
	if categoryCode == "" {
		categoryCode = projectRow["project_category"].String()
	}
	projectFamily := ""
	if categoryCode != "" {
		if row, err := repo.NewProjectCategoryRepo().GetByCode(ctx, categoryCode); err == nil && row != nil {
			projectFamily = gconv.String(row["family_code"])
		} else if row, err := repo.NewProjectCategoryRepo().GetByDisplayName(ctx, categoryCode); err == nil && row != nil {
			projectFamily = gconv.String(row["family_code"])
		}
	}

	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Fields("status, batch_no, retry_count, affected_resources, started_at, completed_at, heartbeat_at").
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		All()
	if err != nil {
		return nil, err
	}

	progress := &ProgressMetrics{}
	health := &HealthMetrics{}
	resource := &ResourceMetrics{
		MaxConcurrency: engine.GetConfigInt(ctx, "workflow.scheduler.max_concurrency", "workflow.scheduler.maxConcurrency", 20),
		TokensConsumed: wf["tokens_consumed"].Int64(),
	}
	trend := &TrendMetrics{
		FailureRateTrend: "stable",
		DurationTrend:    "stable",
		ThroughputTrend:  "stable",
	}

	var durations []int64
	now := gtime.Now()
	maxBatch := 0
	currentBatch := 0
	lockedResources := 0
	statusCount := map[string]int{}
	consecutiveFailures := 0

	for _, task := range tasks {
		status := task["status"].String()
		statusCount[status]++
		progress.TotalTasks++
		batchNo := task["batch_no"].Int()
		if batchNo > maxBatch {
			maxBatch = batchNo
		}
		if status == "running" && batchNo > currentBatch {
			currentBatch = batchNo
		}
		health.RetryCount += task["retry_count"].Int()

		if task["affected_resources"].String() != "" && task["affected_resources"].String() != "[]" {
			lockedResources++
		}

		if status == "failed" || status == "escalated" {
			consecutiveFailures++
		}
		if status == "completed" {
			started := task["started_at"].GTime()
			completed := task["completed_at"].GTime()
			if started != nil && completed != nil && !started.IsZero() && !completed.IsZero() {
				durations = append(durations, completed.Timestamp()-started.Timestamp())
			}
		}
		if status == "running" {
			hb := task["heartbeat_at"].GTime()
			if hb == nil || hb.IsZero() || now.Timestamp()-hb.Timestamp() > 300 {
				health.StaleTaskCount++
			}
		}
	}

	progress.CompletedTasks = statusCount["completed"]
	progress.RunningTasks = statusCount["running"]
	progress.FailedTasks = statusCount["failed"] + statusCount["escalated"]
	progress.PendingTasks = statusCount["pending"] + statusCount["draft"]
	if progress.TotalTasks > 0 {
		progress.CompletionRate = float64(progress.CompletedTasks) / float64(progress.TotalTasks)
	}
	progress.TotalBatches = maxBatch
	progress.CurrentBatchNo = currentBatch
	if currentBatch > 0 {
		batchTotal := 0
		batchDone := 0
		for _, task := range tasks {
			if task["batch_no"].Int() != currentBatch {
				continue
			}
			batchTotal++
			if task["status"].String() == "completed" {
				batchDone++
			}
		}
		if batchTotal > 0 {
			progress.BatchProgress = float64(batchDone) / float64(batchTotal)
		}
	}

	resource.RunningConcurrency = progress.RunningTasks
	resource.LockedResourceCount = lockedResources
	if resource.MaxConcurrency > 0 {
		resource.ResourceUtilization = float64(resource.RunningConcurrency) / float64(resource.MaxConcurrency)
	}

	health.ConsecutiveFailures = consecutiveFailures
	health.EscalationCount = statusCount["escalated"]
	health.ReplanCount = wf["replan_count"].Int()
	if progress.CompletedTasks+progress.FailedTasks > 0 {
		health.RecentFailureRate = float64(progress.FailedTasks) / float64(progress.CompletedTasks+progress.FailedTasks)
	}
	if len(durations) > 0 {
		var total int64
		for _, d := range durations {
			total += d
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		health.AvgTaskDuration = total / int64(len(durations))
		health.MedianTaskDuration = durations[len(durations)/2]
	}

	stageRuns, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Fields("stage_type, stage_no").
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderAsc("stage_no").
		All()
	for _, sr := range stageRuns {
		switch sr["stage_type"].String() {
		case "rework":
			health.ReworkRounds++
		case "accept":
			health.AcceptRounds++
		}
	}

	sit := &Situation{
		WorkflowRunID:     workflowRunID,
		ProjectID:         projectID,
		ProjectFamily:     projectFamily,
		CategoryCode:      categoryCode,
		ActiveStage:       wf["current_stage"].String(),
		WorkflowStatus:    wf["status"].String(),
		WorkflowStartedAt: wf["started_at"].GTime(),
		SnapshotAt:        now,
		Progress:          progress,
		Health:            health,
		Resource:          resource,
		Trend:             trend,
	}
	sit.AnomalySignals = s.DetectAnomalies(sit)
	return sit, nil
}

func (s *Sensor) DetectAnomalies(sit *Situation) []AnomalySignal {
	if sit == nil {
		return nil
	}
	signals := make([]AnomalySignal, 0, 4)
	if sit.Health != nil && sit.Progress != nil && sit.Health.RecentFailureRate >= 0.5 && sit.Progress.TotalTasks >= 4 {
		signals = append(signals, AnomalySignal{
			Type:       "failure_spike",
			Severity:   "warning",
			Message:    "近期失败率过高",
			Confidence: 0.85,
		})
	}
	if sit.Health != nil && sit.Health.StaleTaskCount > 0 {
		signals = append(signals, AnomalySignal{
			Type:       "stale_task",
			Severity:   "warning",
			Message:    "存在心跳过期任务",
			Confidence: 0.9,
		})
	}
	if sit.Resource != nil && sit.Resource.ResourceUtilization >= 1 {
		signals = append(signals, AnomalySignal{
			Type:       "resource_saturation",
			Severity:   "warning",
			Message:    "执行并发已满",
			Confidence: 0.95,
		})
	}
	if sit.Progress != nil && sit.Progress.PendingTasks > 0 && sit.Progress.RunningTasks == 0 && sit.ActiveStage == "execute" {
		signals = append(signals, AnomalySignal{
			Type:       "throughput_drop",
			Severity:   "critical",
			Message:    "执行阶段无运行任务但仍有待处理任务",
			Confidence: 0.8,
		})
	}
	return signals
}

func (s *Sensor) RecordSnapshot(ctx context.Context, sit *Situation) error {
	if s == nil || s.snapshotRepo == nil || sit == nil {
		return nil
	}
	snapshotData, _ := json.Marshal(g.Map{
		"workflowStatus": sit.WorkflowStatus,
		"activeStage":    sit.ActiveStage,
		"categoryCode":   sit.CategoryCode,
		"progress":       sit.Progress,
		"health":         sit.Health,
		"resource":       sit.Resource,
		"trend":          sit.Trend,
		"snapshotAt":     sit.SnapshotAt,
	})
	anomalyData, _ := json.Marshal(sit.AnomalySignals)
	scope := repo.GetProjectScopeByProject(ctx, sit.ProjectID)
	_, err := s.snapshotRepo.Create(ctx, g.Map{
		"workflow_run_id": sit.WorkflowRunID,
		"project_id":      sit.ProjectID,
		"snapshot_data":   string(snapshotData),
		"anomaly_signals": string(anomalyData),
		"created_by":      scope.CreatedBy,
		"dept_id":         scope.DeptID,
		"created_at":      gtime.Now(),
	})
	return err
}
