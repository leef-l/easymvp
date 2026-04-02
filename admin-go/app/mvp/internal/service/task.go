package service

import (
	"context"
	"easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

type ITask interface {
	Create(ctx context.Context, in *model.TaskCreateInput) error
	Update(ctx context.Context, in *model.TaskUpdateInput) error
	Delete(ctx context.Context, id snowflake.JsonInt64) error
	BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error
	Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.TaskDetailOutput, err error)
	List(ctx context.Context, in *model.TaskListInput) (list []*model.TaskListOutput, total int, err error)
	Export(ctx context.Context, in *model.TaskListInput) (list []*model.TaskListOutput, err error)
	Tree(ctx context.Context, in *model.TaskTreeInput) (tree []*model.TaskTreeOutput, err error)
	BatchUpdate(ctx context.Context, in *model.TaskBatchUpdateInput) error
}

var localTask ITask

func Task() ITask {
	return localTask
}

func RegisterTask(i ITask) {
	localTask = i
}
