package eventstream

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// TransitionLog 记录一次工作流状态转移的审计信息。
type TransitionLog struct {
	ID            int64
	WorkflowRunID int64
	Scope         string // workflow 或 stage
	FromState     string
	ToState       string
	GuardResult   string // JSON 字符串，空表示无守卫结果
	DurationMs    *int64 // 源状态停留时长(ms)，nil 表示未知
	Actor         string // "user:<id>" 或 "system"
	Reason        string
	CreatedAt     time.Time
	CreatedBy     int64
	DeptID        int64
}

// LogTransition 将状态转移记录写入 mvp_workflow_transition_log 表。
// 若 log.ID 为 0，自动用雪花算法生成。
// 若 log.CreatedAt 零值，自动填充当前时间。
// 若 log.Scope 为空，默认写入 "workflow"。
// 若 log.Actor 为空，默认写入 "system"。
func LogTransition(ctx context.Context, tl *TransitionLog) error {
	if tl.ID == 0 {
		tl.ID = int64(snowflake.Generate())
	}
	if tl.CreatedAt.IsZero() {
		tl.CreatedAt = time.Now()
	}
	scope := tl.Scope
	if scope == "" {
		scope = "workflow"
	}
	actor := tl.Actor
	if actor == "" {
		actor = "system"
	}

	data := g.Map{
		"id":              tl.ID,
		"workflow_run_id": tl.WorkflowRunID,
		"scope":           scope,
		"from_state":      tl.FromState,
		"to_state":        tl.ToState,
		"actor":           actor,
		"reason":          tl.Reason,
		"created_at":      tl.CreatedAt,
		"created_by":      tl.CreatedBy,
		"dept_id":         tl.DeptID,
	}
	if tl.GuardResult != "" {
		data["guard_result"] = tl.GuardResult
	}
	if tl.DurationMs != nil {
		data["duration_ms"] = *tl.DurationMs
	}

	_, err := g.DB().Ctx(ctx).Model("mvp_workflow_transition_log").Data(data).Insert()
	return err
}
