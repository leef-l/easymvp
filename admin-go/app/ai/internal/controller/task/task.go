package task

import (
	"context"

	v1 "easymvp/app/ai/api/ai/v1"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
)

var Task = cTask{}

type cTask struct{}

func (c *cTask) Execute(ctx context.Context, req *v1.TaskExecuteReq) (res *v1.TaskExecuteRes, err error) {
	res = &v1.TaskExecuteRes{}
	res.TaskExecuteOutput, err = service.Task().Execute(ctx, &model.TaskExecuteInput{
		Title:        req.Title,
		EngineCode:   req.EngineCode,
		ProjectID:    req.ProjectID,
		RepoPath:     req.RepoPath,
		WorktreePath: req.WorktreePath,
		BranchName:   req.BranchName,
		Instruction:  req.Instruction,
	})
	return
}

func (c *cTask) Detail(ctx context.Context, req *v1.TaskDetailReq) (res *v1.TaskDetailRes, err error) {
	res = &v1.TaskDetailRes{}
	res.TaskDetailOutput, err = service.Task().Detail(ctx, req.ID)
	return
}

func (c *cTask) List(ctx context.Context, req *v1.TaskListReq) (res *v1.TaskListRes, err error) {
	res = &v1.TaskListRes{}
	res.List, res.Total, err = service.Task().List(ctx, &model.TaskListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		EngineCode: req.EngineCode,
		Status:     req.Status,
	})
	return
}

func (c *cTask) Logs(ctx context.Context, req *v1.TaskLogsReq) (res *v1.TaskLogsRes, err error) {
	res = &v1.TaskLogsRes{}
	res.List, err = service.Task().Logs(ctx, req.TaskID)
	return
}

func (c *cTask) Cancel(ctx context.Context, req *v1.TaskCancelReq) (res *v1.TaskCancelRes, err error) {
	err = service.Task().Cancel(ctx, req.TaskID)
	return
}
