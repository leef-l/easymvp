package fsm

// TaskStatus 领域任务状态（强类型枚举）。
type TaskStatus string

const (
	TaskPending       TaskStatus = "pending"
	TaskRunning       TaskStatus = "running"
	TaskCompleted     TaskStatus = "completed"
	TaskFailed        TaskStatus = "failed"
	TaskEscalated     TaskStatus = "escalated"
	TaskAuditing      TaskStatus = "auditing"
	TaskBugFound      TaskStatus = "bug_found"
	TaskBugDispatched TaskStatus = "bug_dispatched"
)

// String 返回状态的字符串值。
func (s TaskStatus) String() string { return string(s) }

// ParseTaskStatus 将字符串解析为 TaskStatus。
func ParseTaskStatus(s string) TaskStatus {
	switch TaskStatus(s) {
	case TaskPending, TaskRunning, TaskCompleted, TaskFailed,
		TaskEscalated, TaskAuditing, TaskBugFound, TaskBugDispatched:
		return TaskStatus(s)
	}
	return ""
}

// taskTransitions 合法的任务状态迁移矩阵。
var taskTransitions = map[TaskStatus][]TaskStatus{
	TaskPending:       {TaskRunning},
	TaskRunning:       {TaskCompleted, TaskFailed},
	TaskCompleted:     {TaskAuditing},
	TaskFailed:        {TaskPending, TaskEscalated},
	TaskEscalated:     {},
	TaskAuditing:      {TaskBugFound, TaskPending}, // 审计无 bug 时回 pending（重置为已验收），有 bug 进 bug_found
	TaskBugFound:      {TaskBugDispatched},
	TaskBugDispatched: {TaskRunning},
}

// TaskInitial 返回任务初始状态。
func TaskInitial() TaskStatus { return TaskPending }

// IsValidTaskTransition 检查任务状态迁移是否合法。
func IsValidTaskTransition(from, to TaskStatus) bool {
	targets, ok := taskTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// TaskTargets 返回给定任务状态的合法目标状态列表。
func TaskTargets(from TaskStatus) []TaskStatus {
	targets := taskTransitions[from]
	if targets == nil {
		return []TaskStatus{}
	}
	result := make([]TaskStatus, len(targets))
	copy(result, targets)
	return result
}
