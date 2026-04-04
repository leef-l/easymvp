package task

// DomainTaskStatus 领域任务状态常量。
const (
	StatusPending       = "pending"
	StatusRunning       = "running"
	StatusCompleted     = "completed"
	StatusFailed        = "failed"
	StatusEscalated     = "escalated"
	StatusAuditing      = "auditing"
	StatusBugFound      = "bug_found"
	StatusBugDispatched = "bug_dispatched"
)

// validDomainTaskTransitions 合法的领域任务状态迁移。
var validDomainTaskTransitions = map[string][]string{
	StatusPending:       {StatusRunning},
	StatusRunning:       {StatusCompleted, StatusFailed},
	StatusCompleted:     {StatusAuditing},
	StatusFailed:        {StatusPending, StatusEscalated}, // 重试或升级
	StatusEscalated:     {},
	StatusAuditing:      {StatusBugFound, StatusCompleted}, // 审计通过回 completed 或发现 bug
	StatusBugFound:      {StatusBugDispatched},
	StatusBugDispatched: {StatusPending}, // 返工后回到 pending
}

// IsValidTransition 检查领域任务状态迁移是否合法。
func IsValidTransition(from, to string) bool {
	targets, ok := validDomainTaskTransitions[from]
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
