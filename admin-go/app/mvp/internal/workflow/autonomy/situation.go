package autonomy

import "github.com/gogf/gf/v2/os/gtime"

// Situation 统一态势快照。
type Situation struct {
	WorkflowRunID     int64            `json:"workflowRunId"`
	ProjectID         int64            `json:"projectId"`
	ProjectFamily     string           `json:"projectFamily,omitempty"`
	CategoryCode      string           `json:"categoryCode,omitempty"`
	ActiveStage       string           `json:"activeStage"`
	WorkflowStatus    string           `json:"workflowStatus"`
	WorkflowStartedAt *gtime.Time      `json:"workflowStartedAt,omitempty"`
	SnapshotAt        *gtime.Time      `json:"snapshotAt"`
	Progress          *ProgressMetrics `json:"progress"`
	Health            *HealthMetrics   `json:"health"`
	Resource          *ResourceMetrics `json:"resource"`
	Trend             *TrendMetrics    `json:"trend"`
	AnomalySignals    []AnomalySignal  `json:"anomalySignals,omitempty"`
}

type ProgressMetrics struct {
	TotalTasks     int     `json:"totalTasks"`
	CompletedTasks int     `json:"completedTasks"`
	RunningTasks   int     `json:"runningTasks"`
	FailedTasks    int     `json:"failedTasks"`
	PendingTasks   int     `json:"pendingTasks"`
	CompletionRate float64 `json:"completionRate"`
	CurrentBatchNo int     `json:"currentBatchNo"`
	TotalBatches   int     `json:"totalBatches"`
	BatchProgress  float64 `json:"batchProgress"`
}

type HealthMetrics struct {
	ConsecutiveFailures int     `json:"consecutiveFailures"`
	RecentFailureRate   float64 `json:"recentFailureRate"`
	AvgTaskDuration     int64   `json:"avgTaskDuration"`
	MedianTaskDuration  int64   `json:"medianTaskDuration"`
	RetryCount          int     `json:"retryCount"`
	TaskRetryCount      int     `json:"taskRetryCount"`
	EscalationCount     int     `json:"escalationCount"`
	ReworkRounds        int     `json:"reworkRounds"`
	TaskReworkRounds    int     `json:"taskReworkRounds"`
	FocusedTaskID       int64   `json:"focusedTaskId,omitempty"`
	AcceptRounds        int     `json:"acceptRounds"`
	ReplanCount         int     `json:"replanCount"`
	StaleTaskCount      int     `json:"staleTaskCount"`
}

type ResourceMetrics struct {
	RunningConcurrency  int     `json:"runningConcurrency"`
	MaxConcurrency      int     `json:"maxConcurrency"`
	ResourceUtilization float64 `json:"resourceUtilization"`
	LockedResourceCount int     `json:"lockedResourceCount"`
	ConflictCount       int     `json:"conflictCount"`
	TokensConsumed      int64   `json:"tokensConsumed"`
	EstimatedTokensLeft int64   `json:"estimatedTokensLeft"`
}

type TrendMetrics struct {
	FailureRateTrend string `json:"failureRateTrend"`
	DurationTrend    string `json:"durationTrend"`
	ThroughputTrend  string `json:"throughputTrend"`
}

type AnomalySignal struct {
	Type       string  `json:"type"`
	Severity   string  `json:"severity"`
	Message    string  `json:"message"`
	Confidence float64 `json:"confidence"`
}

func (s *Situation) HasCriticalAnomaly() bool {
	for _, signal := range s.AnomalySignals {
		if signal.Severity == "critical" {
			return true
		}
	}
	return false
}
