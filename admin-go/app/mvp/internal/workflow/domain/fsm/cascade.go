package fsm

// Effect 描述 Task 状态变化对上层 Stage 的级联影响。
type Effect struct {
	// MarkStageHasBug 标记当前阶段存在 Bug，触发审计/返工流程。
	MarkStageHasBug bool
	// EscalateToHuman 升级为人工介入，不再自动重试。
	EscalateToHuman bool
	// Retry 标记该任务应被重试（回到 pending）。
	Retry bool
	// StageCompleted 当前阶段全部任务完成，可推进到下一阶段。
	StageCompleted bool
}

// TaskToStageEffect 定义单条级联规则：某个 Task 状态触发何种 Stage 影响。
type TaskToStageEffect struct {
	// TaskStatus 触发此规则的任务状态。
	TaskStatus TaskStatus
	// StageEffect 根据重试次数和最大重试限制计算实际影响。
	StageEffect func(retryCount, maxRetry int) Effect
}

// CascadeRules 返回所有 Task→Stage 级联规则。
//
// 规则优先级从上到下：
//  1. TaskBugFound  → MarkStageHasBug
//  2. TaskFailed    → 重试次数未耗尽时 Retry，否则 EscalateToHuman
//  3. TaskCompleted → 所有任务完成时 StageCompleted（调用方负责全量检查）
func CascadeRules() []TaskToStageEffect {
	return []TaskToStageEffect{
		{
			TaskStatus: TaskBugFound,
			StageEffect: func(_, _ int) Effect {
				return Effect{MarkStageHasBug: true}
			},
		},
		{
			TaskStatus: TaskFailed,
			StageEffect: func(retryCount, maxRetry int) Effect {
				if retryCount >= maxRetry {
					return Effect{EscalateToHuman: true}
				}
				return Effect{Retry: true}
			},
		},
		{
			TaskStatus: TaskCompleted,
			// 调用方需自行判断阶段内全部任务是否均已 completed 后再使用此 Effect。
			StageEffect: func(_, _ int) Effect {
				return Effect{StageCompleted: true}
			},
		},
	}
}

// ApplyCascade 根据任务状态和重试信息查找并返回对应的级联影响。
// 若无匹配规则，返回零值 Effect。
func ApplyCascade(taskStatus TaskStatus, retryCount, maxRetry int) Effect {
	for _, rule := range CascadeRules() {
		if rule.TaskStatus == taskStatus {
			return rule.StageEffect(retryCount, maxRetry)
		}
	}
	return Effect{}
}
