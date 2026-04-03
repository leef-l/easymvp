package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
)

// TaskExecuteReq 创建执行任务
type TaskExecuteReq struct {
	g.Meta       `path:"/task/execute" method:"post" tags:"AI执行任务" summary:"创建AI执行任务"`
	Title        string              `json:"title" v:"required#任务标题不能为空"`
	EngineCode   string              `json:"engineCode" v:"required#执行引擎不能为空"`
	ProjectID    snowflake.JsonInt64 `json:"projectID"`
	RepoPath     string              `json:"repoPath" v:"required#仓库路径不能为空"`
	WorktreePath string              `json:"worktreePath"`
	BranchName   string              `json:"branchName"`
	Instruction  string              `json:"instruction" v:"required#任务指令不能为空"`
}

type TaskExecuteRes struct {
	g.Meta `mime:"application/json"`
	*model.TaskExecuteOutput
}

// TaskDetailReq 获取任务详情
type TaskDetailReq struct {
	g.Meta `path:"/task/detail" method:"get" tags:"AI执行任务" summary:"获取AI执行任务详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#任务ID不能为空"`
}

type TaskDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.TaskDetailOutput
}

// TaskListReq 获取任务列表
type TaskListReq struct {
	g.Meta     `path:"/task/list" method:"get" tags:"AI执行任务" summary:"获取AI执行任务列表"`
	PageNum    int    `json:"pageNum" d:"1"`
	PageSize   int    `json:"pageSize" d:"10"`
	EngineCode string `json:"engineCode"`
	Status     string `json:"status"`
}

type TaskListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.TaskListOutput `json:"list"`
	Total  int                     `json:"total"`
}

// TaskLogsReq 获取任务日志
type TaskLogsReq struct {
	g.Meta `path:"/task/logs" method:"get" tags:"AI执行任务" summary:"获取AI执行任务日志"`
	TaskID snowflake.JsonInt64 `json:"taskID" v:"required#任务ID不能为空"`
}

type TaskLogsRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.TaskLogOutput `json:"list"`
}

// TaskCancelReq 取消任务
type TaskCancelReq struct {
	g.Meta `path:"/task/cancel" method:"post" tags:"AI执行任务" summary:"取消AI执行任务"`
	TaskID snowflake.JsonInt64 `json:"taskID" v:"required#任务ID不能为空"`
}

type TaskCancelRes struct {
	g.Meta `mime:"application/json"`
}
