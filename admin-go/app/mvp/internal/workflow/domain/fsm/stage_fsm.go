package fsm

// StageStatus 阶段运行状态（强类型枚举）。
type StageStatus string

const (
	StagePending   StageStatus = "pending"
	StageRunning   StageStatus = "running"
	StageCompleted StageStatus = "completed"
	StageFailed    StageStatus = "failed"
	StageSkipped   StageStatus = "skipped"
)

// String 返回状态的字符串值。
func (s StageStatus) String() string { return string(s) }

// ParseStageStatus 将字符串解析为 StageStatus。
func ParseStageStatus(s string) StageStatus {
	switch StageStatus(s) {
	case StagePending, StageRunning, StageCompleted, StageFailed, StageSkipped:
		return StageStatus(s)
	}
	return ""
}

// stageTransitions 合法的阶段状态迁移矩阵。
var stageTransitions = map[StageStatus][]StageStatus{
	StagePending:   {StageRunning, StageSkipped},
	StageRunning:   {StageCompleted, StageFailed},
	StageCompleted: {},
	StageFailed:    {StagePending},
	StageSkipped:   {},
}

// StageInitial 返回阶段初始状态。
func StageInitial() StageStatus { return StagePending }

// IsValidStageTransition 检查阶段状态迁移是否合法。
func IsValidStageTransition(from, to StageStatus) bool {
	targets, ok := stageTransitions[from]
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

// StageTargets 返回给定阶段状态的合法目标状态列表。
func StageTargets(from StageStatus) []StageStatus {
	targets := stageTransitions[from]
	if targets == nil {
		return []StageStatus{}
	}
	result := make([]StageStatus, len(targets))
	copy(result, targets)
	return result
}
