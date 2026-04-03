package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

type TaskExecuteInput struct {
	Title        string
	EngineCode   string
	ProjectID    snowflake.JsonInt64
	RepoPath     string
	WorktreePath string
	BranchName   string
	Instruction  string
}

type TaskExecuteOutput struct {
	TaskID snowflake.JsonInt64 `json:"taskID"`
	Status string              `json:"status"`
}

type TaskDetailOutput struct {
	ID              snowflake.JsonInt64 `json:"id"`
	Title           string              `json:"title"`
	EngineCode      string              `json:"engineCode"`
	Status          string              `json:"status"`
	RepoPath        string              `json:"repoPath"`
	WorktreePath    string              `json:"worktreePath"`
	BranchName      string              `json:"branchName"`
	Instruction     string              `json:"instruction"`
	ResponseSummary string              `json:"responseSummary"`
	ErrorMessage    string              `json:"errorMessage"`
	StartedAt       *gtime.Time         `json:"startedAt"`
	FinishedAt      *gtime.Time         `json:"finishedAt"`
	CreatedAt       *gtime.Time         `json:"createdAt"`
}

type TaskListInput struct {
	PageNum    int
	PageSize   int
	EngineCode string
	Status     string
}

type TaskListOutput struct {
	ID         snowflake.JsonInt64 `json:"id"`
	Title      string              `json:"title"`
	EngineCode string              `json:"engineCode"`
	Status     string              `json:"status"`
	RepoPath   string              `json:"repoPath"`
	CreatedAt  *gtime.Time         `json:"createdAt"`
	StartedAt  *gtime.Time         `json:"startedAt"`
	FinishedAt *gtime.Time         `json:"finishedAt"`
}

type TaskLogOutput struct {
	ID        snowflake.JsonInt64 `json:"id"`
	TaskID    snowflake.JsonInt64 `json:"taskID"`
	Seq       int                 `json:"seq"`
	LogType   string              `json:"logType"`
	Content   string              `json:"content"`
	CreatedAt *gtime.Time         `json:"createdAt"`
}
