package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Workflow 项目流程编排
// designing → confirmed → running → paused/completed/failed

// ConfirmPlan 确认实施方案，开始任务调度
func (s *Scheduler) ConfirmPlan(ctx context.Context, projectID int64) error {
	// 1. 检查项目状态
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("项目不存在")
	}

	status := project["status"].String()
	if status != "designing" && status != "paused" {
		return fmt.Errorf("当前状态(%s)不允许确认方案", status)
	}

	// 2. 将 draft 任务确认为 pending
	confirmedCount, err := GetParser().ConfirmDraftTasks(ctx, projectID)
	if err != nil {
		return fmt.Errorf("确认草稿任务失败: %w", err)
	}

	// 3. 检查是否有待执行的任务（draft 已转为 pending + 可能已有 pending）
	count, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		Where("deleted_at IS NULL").
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("没有任务可执行，请先让架构师拆分任务")
	}

	_ = confirmedCount // 已使用

	// 3. 更新项目状态，清空暂停原因
	_, err = g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":       "running",
		"pause_reason": nil,
		"updated_at":   gtime.Now(),
	})
	if err != nil {
		return err
	}

	// 4. 压缩架构师对话为全局上下文
	go GetCompressor().CompressProjectContext(context.Background(), projectID)

	// 5. 启动调度器
	s.StartProject(projectID)

	return nil
}

// Pause 暂停项目（回到设计阶段，可以和架构师沟通），记录暂停原因
func (s *Scheduler) Pause(ctx context.Context, projectID int64, reason string) error {
	_, err := g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":       "paused",
		"pause_reason": reason,
		"updated_at":   gtime.Now(),
	})
	if err != nil {
		return err
	}

	s.PauseProject(projectID)
	return nil
}

// ResumeProject 恢复项目（重新开始调度）
func (s *Scheduler) Resume(ctx context.Context, projectID int64) error {
	return s.ConfirmPlan(ctx, projectID)
}

// CreateProject 创建项目并初始化架构师对话
func CreateProject(ctx context.Context, name, description string, architectModelID int64, userID int64, deptID int64) (int64, int64, error) {
	projectID := int64(snowflake.Generate())

	// 1. 创建项目
	_, err := g.DB().Model("mvp_project").Insert(g.Map{
		"id":                 projectID,
		"name":               name,
		"description":        description,
		"status":             "designing",
		"architect_model_id": architectModelID,
		"created_by":         userID,
		"dept_id":            deptID,
		"created_at":         gtime.Now(),
		"updated_at":         gtime.Now(),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("创建项目失败: %w", err)
	}

	// 2. 创建架构师角色配置
	if architectModelID > 0 {
		_, err = g.DB().Model("mvp_project_role").Insert(g.Map{
			"id":            int64(snowflake.Generate()),
			"project_id":    projectID,
			"role_type":     "architect",
			"model_id":      architectModelID,
			"system_prompt": buildArchitectPrompt(name, description),
			"status":        1,
			"created_by":    userID,
			"dept_id":       deptID,
			"created_at":    gtime.Now(),
			"updated_at":    gtime.Now(),
		})
		if err != nil {
			return 0, 0, fmt.Errorf("创建架构师配置失败: %w", err)
		}
	}

	// 3. 创建架构师对话（项目级对话）
	convID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_conversation").Insert(g.Map{
		"id":         convID,
		"project_id": projectID,
		"title":      "架构师对话",
		"role_type":  "architect",
		"status":     "active",
		"created_by": userID,
		"dept_id":    deptID,
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("创建架构师对话失败: %w", err)
	}

	return projectID, convID, nil
}

// buildArchitectPrompt 构建架构师系统提示词
func buildArchitectPrompt(projectName, projectDesc string) string {
	return fmt.Sprintf(`你是一个资深软件架构师，负责项目「%s」的需求分析和方案设计。

项目简介：
%s

你的职责：
1. 与用户沟通，深入理解需求
2. 设计技术方案，考虑高并发和可扩展性
3. 将项目按功能模块细拆为 80-200 个任务
4. 确保并行任务之间不冲突（不修改同一个文件/模块）
5. 为每个任务标注：任务名称、描述、角色等级(lite/pro/max)、执行批次号、涉及的资源范围、依赖关系

输出任务清单时请使用 JSON 格式，便于系统解析：
{
  "tasks": [
    {
      "name": "任务名称",
      "description": "详细描述",
      "role_level": "max/pro/lite",
      "batch_no": 1,
      "affected_resources": ["file1.go", "file2.go"],
      "depends_on": []
    }
  ]
}`, projectName, projectDesc)
}
