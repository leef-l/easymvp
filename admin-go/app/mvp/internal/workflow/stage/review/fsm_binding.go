package review

import (
	"fmt"

	"easymvp/app/mvp/internal/workflow/domain/fsm"
)

// ValidateTransition 校验当前阶段的状态转移是否合法。
// 在阶段入口和出口调用，确保所有状态流转经过 FSM 校验。
func ValidateTransition(from, to fsm.WorkflowStatus) error {
	if !fsm.IsValidWorkflowTransition(from, to) {
		return fmt.Errorf("[review] 非法状态转移: %s → %s", from, to)
	}
	return nil
}

// ExpectedStatus 返回本阶段对应的 workflow 状态。
func ExpectedStatus() fsm.WorkflowStatus {
	return fsm.WorkflowReviewing
}
