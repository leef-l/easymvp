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

	// 4. 压缩架构师对话为全局上下文（同步执行，确保上下文就绪后再启动调度）
	if compErr := GetCompressor().CompressProjectContext(context.Background(), projectID); compErr != nil {
		g.Log().Errorf(ctx, "[Workflow] 压缩项目上下文失败（非致命）: project=%d, err=%v", projectID, compErr)
	}

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
// architectModelID 为前端传入的架构师模型，若为 0 则从预设读取
func CreateProject(ctx context.Context, name, projectCategory, description, workDir string, architectModelID int64, userID int64, deptID int64) (int64, int64, error) {
	workDir, _, err := EnsureWorkDir(workDir)
	if err != nil {
		return 0, 0, err
	}

	projectID := int64(snowflake.Generate())

	// 1.5 默认分类
	if projectCategory == "" {
		projectCategory = "软件开发"
	}

	// 1. 按项目分类读取角色预设模板
	presets, err := g.DB().Model("mvp_role_preset").
		Where("status", 1).
		Where("project_category", projectCategory).
		Where("deleted_at IS NULL").
		OrderAsc("sort").
		All()
	if err != nil {
		return 0, 0, fmt.Errorf("读取角色预设失败: %w", err)
	}

	// 找到预设中的架构师模型作为兜底
	if architectModelID == 0 {
		for _, p := range presets {
			if p["role_type"].String() == "architect" && p["model_id"].Int64() > 0 {
				architectModelID = p["model_id"].Int64()
				break
			}
		}
	}

	// 2. 创建项目
	_, err = g.DB().Model("mvp_project").Insert(g.Map{
		"id":                 projectID,
		"name":               name,
		"project_category":   projectCategory,
		"description":        description,
		"status":             "designing",
		"work_dir":           workDir,
		"architect_model_id": architectModelID,
		"created_by":         userID,
		"dept_id":            deptID,
		"created_at":         gtime.Now(),
		"updated_at":         gtime.Now(),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("创建项目失败: %w", err)
	}

	// 3. 批量查询预设关联的模型，获取 role_prompt
	modelIDs := make([]int64, 0, len(presets))
	for _, p := range presets {
		if mid := p["model_id"].Int64(); mid > 0 {
			modelIDs = append(modelIDs, mid)
		}
	}
	if architectModelID > 0 {
		modelIDs = append(modelIDs, architectModelID)
	}
	modelPromptMap := make(map[int64]string)
	if len(modelIDs) > 0 {
		models, _ := g.DB().Model("ai_model").
			Fields("id, role_prompt").
			WhereIn("id", modelIDs).
			Where("deleted_at IS NULL").
			All()
		for _, m := range models {
			modelPromptMap[m["id"].Int64()] = m["role_prompt"].String()
		}
	}

	// 4. 根据预设模板创建项目角色配置
	for _, p := range presets {
		roleType := p["role_type"].String()
		modelID := p["model_id"].Int64()

		// 架构师角色：优先用前端传入的模型
		if roleType == "architect" {
			modelID = architectModelID
		}

		// 系统提示词优先级：模型 role_prompt > 预设 system_prompt
		// 架构师特殊处理：动态拼接项目信息
		systemPrompt := ""
		if roleType == "architect" {
			systemPrompt = buildArchitectPrompt(name, description, modelPromptMap[modelID])
		} else if rp, ok := modelPromptMap[modelID]; ok && rp != "" {
			systemPrompt = rp
		} else {
			systemPrompt = p["system_prompt"].String()
		}

		// 复制执行方式，默认 chat
		executionMode := p["execution_mode"].String()
		if executionMode == "" {
			executionMode = "chat"
		}

		_, err = g.DB().Model("mvp_project_role").Insert(g.Map{
			"id":               int64(snowflake.Generate()),
			"project_id":       projectID,
			"project_category": projectCategory,
			"role_type":        roleType,
			"role_level":       p["role_level"].String(),
			"model_id":        modelID,
			"system_prompt":   systemPrompt,
			"execution_mode":  executionMode,
			"status":          1,
			"created_by":      userID,
			"dept_id":         deptID,
			"created_at":      gtime.Now(),
			"updated_at":      gtime.Now(),
		})
		if err != nil {
			return 0, 0, fmt.Errorf("创建角色配置(%s)失败: %w", roleType, err)
		}
	}

	// 如果预设为空，至少创建架构师角色（兼容无预设场景）
	if len(presets) == 0 && architectModelID > 0 {
		_, err = g.DB().Model("mvp_project_role").Insert(g.Map{
			"id":               int64(snowflake.Generate()),
			"project_id":       projectID,
			"project_category": projectCategory,
			"role_type":        "architect",
			"model_id":      architectModelID,
			"system_prompt": buildArchitectPrompt(name, description, modelPromptMap[architectModelID]),
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

	// 4. 创建架构师对话（项目级对话）
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
// 优先使用模型自带的 role_prompt，拼接项目上下文信息
func buildArchitectPrompt(projectName, projectDesc string, modelRolePrompt string) string {
	projectContext := fmt.Sprintf("\n\n===== 当前项目 =====\n项目名称：%s\n项目简介：%s", projectName, projectDesc)

	if modelRolePrompt != "" {
		return modelRolePrompt + projectContext
	}

	// 兜底：模型没有配 role_prompt 时使用默认提示词
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
