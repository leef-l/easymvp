package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// WorkflowObjectiveReq 查询项目目标约束。
type WorkflowObjectiveReq struct {
	g.Meta    `path:"/workflow/objective" method:"get" tags:"自治L4" summary:"查询项目目标约束"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

type WorkflowObjectiveRes struct {
	g.Meta    `mime:"application/json"`
	Objective g.Map `json:"objective"`
}

// WorkflowSaveObjectiveReq 保存项目目标约束。
type WorkflowSaveObjectiveReq struct {
	g.Meta             `path:"/workflow/save-objective" method:"post" tags:"自治L4" summary:"保存项目目标约束"`
	ProjectID          snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	DeliveryGoal       string              `json:"deliveryGoal"`
	QualityFloor       float64             `json:"qualityFloor"`
	TokenBudget        int64               `json:"tokenBudget"`
	TimeBudgetHours    float64             `json:"timeBudgetHours"`
	CostBudgetCents    int64               `json:"costBudgetCents"`
	RiskTolerance      string              `json:"riskTolerance"`
	MaxAutoRetries     int                 `json:"maxAutoRetries"`
	MaxAutoReworks     int                 `json:"maxAutoReworks"`
	MaxAutoReplans     int                 `json:"maxAutoReplans"`
	DeadlineAt         string              `json:"deadlineAt"`
	MaxStallMinutes    int                 `json:"maxStallMinutes"`
	AutonomyLevel      string              `json:"autonomyLevel"`
	MaxSideEffectLevel string              `json:"maxSideEffectLevel"`
}

type WorkflowSaveObjectiveRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowSituationReq 查询当前态势。
type WorkflowSituationReq struct {
	g.Meta        `path:"/workflow/situation" method:"get" tags:"自治L4" summary:"查询工作流当前态势"`
	WorkflowRunID snowflake.JsonInt64 `json:"workflowRunID" v:"required" dc:"工作流运行ID"`
	TaskID        snowflake.JsonInt64 `json:"taskID" dc:"任务ID(可选，传入后返回该任务焦点态势)"`
}

type WorkflowSituationRes struct {
	g.Meta    `mime:"application/json"`
	Situation g.Map `json:"situation"`
}

// WorkflowSituationHistoryReq 查询态势快照历史。
type WorkflowSituationHistoryReq struct {
	g.Meta        `path:"/workflow/situation-history" method:"get" tags:"自治L4" summary:"查询态势快照历史"`
	ProjectID     snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	WorkflowRunID snowflake.JsonInt64 `json:"workflowRunID" dc:"工作流运行ID(可选)"`
	TaskID        snowflake.JsonInt64 `json:"taskID" dc:"任务ID(可选，仅返回该任务焦点快照)"`
	Limit         int                 `json:"limit" dc:"数量限制"`
}

type WorkflowSituationHistoryRes struct {
	g.Meta    `mime:"application/json"`
	Snapshots []g.Map `json:"snapshots"`
}
