# Complete Stage 收尾闭环设计文档

> 工作流最终阶段：指标统计、总结归档、状态收口。

## 一、背景

当 Execute 阶段所有 domain_task 完成（或 escalated/skipped），调度器触发 `CompletionCallback`，
当前 `completeWorkflow()` 直接创建 complete stage_run 并立即标记完成，**没有指标统计和总结生成**。

本次改造目标：在 complete stage 中执行实质性的收尾逻辑。

## 二、职责

| # | 职责 | 说明 |
|---|------|------|
| 1 | 任务指标统计 | 统计 completed/failed/escalated/skipped 数量、成功率 |
| 2 | 耗时计算 | workflow 总耗时、各阶段耗时、平均任务耗时 |
| 3 | 返工统计 | rework 轮次、handoff 记录数 |
| 4 | 生成 JSON 总结 | 写入 `mvp_stage_run.output_ref` |
| 5 | 发布完成事件 | `workflow.completed` 事件到事件总线 |
| 6 | 项目状态收口 | `mvp_project.status` → completed |
| 7 | 清理 runtime | 取消 runtime context |

## 三、数据流

```
scheduler.checkAllDone()
  → CompletionCallback (execute/service.go)
    → stageCompleter.CompleteStage(executeStageRunID)
      → StageService.completeWorkflow(workflowRunID)
        → StartStage("complete")     // 创建 complete stage_run
        → completeStageSvc.Finalize(ctx, completeStageRunID, workflowRunID)
            ├── collectTaskMetrics()      // 查 domain_task 统计
            ├── collectStageMetrics()     // 查 stage_run 各阶段耗时
            ├── collectReworkMetrics()    // 查 handoff_record 统计
            ├── buildSummary()            // 组装 JSON
            └── saveSummary()             // 写 output_ref
        → CompleteStage(completeStageRunID)
        → workflow_run → completed
        → project → completed
        → publisher.Emit(workflow.completed)
        → runtimeMgr.Cancel()
```

## 四、接口设计

### 4.1 Complete Service

```go
type Service struct {
    publisher *event.Publisher  // 事件发布（可选）
}

func (s *Service) Finalize(ctx context.Context, stageRunID, workflowRunID int64) error
```

**入参**：stageRunID（当前 complete stage_run）、workflowRunID
**职责**：收集指标 → 生成总结 → 写入 output_ref

### 4.2 CompletionSummary 结构

```go
type CompletionSummary struct {
    WorkflowRunID   int64              `json:"workflow_run_id"`
    ProjectID       int64              `json:"project_id"`
    // 任务指标
    TotalTasks      int                `json:"total_tasks"`
    CompletedTasks  int                `json:"completed_tasks"`
    FailedTasks     int                `json:"failed_tasks"`
    EscalatedTasks  int                `json:"escalated_tasks"`
    SkippedTasks    int                `json:"skipped_tasks"`
    SuccessRate     float64            `json:"success_rate"`
    // 耗时
    TotalDuration   string             `json:"total_duration"`
    AvgTaskDuration string             `json:"avg_task_duration"`
    StageDurations  map[string]string  `json:"stage_durations"`
    // 返工
    ReworkRounds    int                `json:"rework_rounds"`
    HandoffCount    int                `json:"handoff_count"`
    // 元数据
    StartedAt       string             `json:"started_at"`
    FinishedAt      string             `json:"finished_at"`
}
```

### 4.3 API 端点

```
GET /mvp/workflow/completion-summary?projectID=xxx
```

返回最近一次 workflow_run 的 complete stage_run 的 output_ref（即 CompletionSummary JSON）。

## 五、Registry 注册

```go
// registry.go Init() 中
completeStageSvc = completeStage.NewService(eventPublisher)

// completeWorkflow() 调用
completeStageSvc.Finalize(ctx, completeStageID, workflowRunID)
```

## 六、改动文件清单

| 文件 | 改动 |
|------|------|
| `workflow/stage/complete/service.go` | 重写：Finalize() + GetSummary() |
| `workflow/orchestrator/stage_service.go` | completeWorkflow() 插入 Finalize + 发布事件 |
| `workflow/orchestrator/registry.go` | 注册 completeStageSvc + getter |
| `api/mvp/v1/workflow.go` | 新增 CompletionSummary API 定义 |
| `controller/chat/workflow.go` | 新增 CompletionSummary handler |
| `前端 api/mvp/workflow/index.ts` | 新增 getCompletionSummary() |

## 七、不做的事

- 不做 AI 总结（纯结构化统计，不调 AI 接口）
- 不做归档备份（不额外复制数据）
- 不做异步清理（workspace cleanup 已在 execute 完成回调中处理）
