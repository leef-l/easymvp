package task

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// HandoffService 交接记录服务。
type HandoffService struct{}

// NewHandoffService 创建交接服务。
func NewHandoffService() *HandoffService { return &HandoffService{} }

// RecordHandoff 记录任务间交接。
func (s *HandoffService) RecordHandoff(ctx context.Context, workflowRunID, fromTaskID, toTaskID int64, handoffType, reason string) error {
	return repo.NewHandoffRecordRepo().Create(ctx, g.Map{
		"id":              snowflake.Generate(),
		"workflow_run_id": workflowRunID,
		"from_task_id":    fromTaskID,
		"to_task_id":      toTaskID,
		"handoff_type":    handoffType,
		"reason":          reason,
	})
}
