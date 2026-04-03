package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/middleware"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterTask(New())
}

func New() *sTask {
	return &sTask{}
}

type sTask struct{}

func (s *sTask) Execute(ctx context.Context, in *model.TaskExecuteInput) (out *model.TaskExecuteOutput, err error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("未获取到当前登录用户")
	}

	engineDetail, err := service.Engine().Detail(ctx, in.EngineCode)
	if err != nil {
		return nil, err
	}
	if engineDetail.Status != 1 || engineDetail.ConfigStatus != 1 {
		return nil, fmt.Errorf("执行引擎未启用或未完成配置")
	}

	roleID, err := s.authorizeEngineUsage(ctx, userID, in.EngineCode)
	if err != nil {
		return nil, err
	}

	taskID := snowflake.Generate()
	now := gtime.Now()
	repoPath, worktreePath, err := s.resolveTaskPaths(engineDetail.WorkspaceRoot, in.RepoPath, in.WorktreePath)
	if err != nil {
		return nil, err
	}
	if err = ensureDirAvailable(repoPath, "仓库路径"); err != nil {
		return nil, err
	}
	if err = ensureDirAvailable(worktreePath, "工作目录"); err != nil {
		return nil, err
	}
	snapshotBytes, _ := json.Marshal(engineDetail)
	requestBytes, _ := json.Marshal(in)

	_, err = g.DB().Ctx(ctx).Model("ai_task").Data(g.Map{
		"id":                     taskID,
		"title":                  in.Title,
		"engine_code":            in.EngineCode,
		"role_id":                roleID,
		"user_id":                userID,
		"project_id":             in.ProjectID,
		"repo_path":              repoPath,
		"worktree_path":          worktreePath,
		"branch_name":            in.BranchName,
		"instruction":            in.Instruction,
		"engine_config_snapshot": string(snapshotBytes),
		"request_payload":        string(requestBytes),
		"status":                 "pending",
		"created_by":             userID,
		"dept_id":                middleware.GetDeptID(ctx),
		"created_at":             now,
		"updated_at":             now,
	}).Insert()
	if err != nil {
		return nil, err
	}

	_ = s.appendTaskLog(ctx, int64(taskID), "system", fmt.Sprintf("任务已创建，等待执行。引擎=%s，工作目录=%s", in.EngineCode, worktreePath))
	s.dispatchTask(int64(taskID))

	return &model.TaskExecuteOutput{
		TaskID: snowflake.JsonInt64(taskID),
		Status: "pending",
	}, nil
}

func (s *sTask) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.TaskDetailOutput, err error) {
	out = &model.TaskDetailOutput{}
	m := g.DB().Ctx(ctx).Model("ai_task").
		Where("id", id).
		Where("deleted_at IS NULL").
		Where("deleted_at IS NULL")
	m = middleware.ApplyDataScope(ctx, m, "created_by", "dept_id")
	err = m.Scan(out)
	if err != nil {
		return
	}
	err = s.enrichTaskDetailActivity(ctx, out)
	return
}

func (s *sTask) List(ctx context.Context, in *model.TaskListInput) (list []*model.TaskListOutput, total int, err error) {
	m := g.DB().Ctx(ctx).Model("ai_task").Where("deleted_at IS NULL")
	if in.EngineCode != "" {
		m = m.Where("engine_code", in.EngineCode)
	}
	if in.Status != "" {
		m = m.Where("status", in.Status)
	}
	m = middleware.ApplyDataScope(ctx, m, "created_by", "dept_id")
	total, err = m.Count()
	if err != nil {
		return
	}
	err = m.Page(in.PageNum, in.PageSize).OrderDesc("created_at").Scan(&list)
	if err != nil {
		return
	}
	err = s.enrichTaskListActivity(ctx, list)
	return
}

func (s *sTask) Logs(ctx context.Context, taskID snowflake.JsonInt64) (list []*model.TaskLogOutput, err error) {
	detail, err := s.Detail(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if detail == nil || detail.ID == 0 {
		return nil, fmt.Errorf("任务不存在")
	}
	err = g.DB().Ctx(ctx).Model("ai_task_log").
		Where("task_id", taskID).
		OrderAsc("seq").
		OrderAsc("created_at").
		Scan(&list)
	return
}

func (s *sTask) Cancel(ctx context.Context, taskID snowflake.JsonInt64) error {
	detail, err := s.Detail(ctx, taskID)
	if err != nil {
		return err
	}
	if detail == nil || detail.ID == 0 {
		return fmt.Errorf("任务不存在")
	}
	now := gtime.Now()
	_, err = g.DB().Ctx(ctx).Model("ai_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		WhereIn("status", []string{"pending", "running"}).
		Data(g.Map{
			"status":      "cancelled",
			"finished_at": now,
			"updated_at":  now,
		}).
		Update()
	if err != nil {
		return err
	}
	if cancel, ok := runningTaskCancels.Load(int64(taskID)); ok {
		cancel.(context.CancelFunc)()
	}
	_ = s.appendTaskLog(context.Background(), int64(taskID), "system", "任务已取消。")
	return nil
}
