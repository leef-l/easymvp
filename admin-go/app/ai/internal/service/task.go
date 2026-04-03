package service

import (
	"context"

	"easymvp/app/ai/internal/model"
	"easymvp/utility/snowflake"
)

type ITask interface {
	Execute(ctx context.Context, in *model.TaskExecuteInput) (out *model.TaskExecuteOutput, err error)
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.TaskDetailOutput, err error)
	List(ctx context.Context, in *model.TaskListInput) (list []*model.TaskListOutput, total int, err error)
	Logs(ctx context.Context, taskID snowflake.JsonInt64) (list []*model.TaskLogOutput, err error)
	Cancel(ctx context.Context, taskID snowflake.JsonInt64) error
}

var localTask ITask

func Task() ITask {
	return localTask
}

func RegisterTask(i ITask) {
	localTask = i
}
