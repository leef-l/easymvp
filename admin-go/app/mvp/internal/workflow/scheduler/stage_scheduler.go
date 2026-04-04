// Package scheduler 新架构任务调度器。
// stage_scheduler 调阶段任务，domain_task_scheduler 调执行任务。
package scheduler

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// StageScheduler 阶段任务调度器。
type StageScheduler struct{}

// NewStageScheduler 创建阶段调度器。
func NewStageScheduler() *StageScheduler { return &StageScheduler{} }

// ScheduleStageTasks 调度指定阶段的子任务。
func (s *StageScheduler) ScheduleStageTasks(ctx context.Context, stageRunID int64) error {
	// TODO: M3 实现
	g.Log().Infof(ctx, "[StageScheduler] ScheduleStageTasks stageRunID=%d", stageRunID)
	return nil
}
