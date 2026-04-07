package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/presetutil"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// Workflow 项目流程编排
// designing → confirmed → running → paused/completed/failed

// ConfirmPlan 确认实施方案，进入方案审核阶段
// 流程：designing → reviewing → (审核通过) → running
//
//	reviewing → (审核不通过) → designing
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

	// 2. 检查是否有 draft 任务
	draftCount := GetParser().GetDraftCount(ctx, projectID)
	if draftCount == 0 {
		return fmt.Errorf("没有待确认的任务，请先让架构师拆分任务")
	}

	// 3. 更新项目状态为 reviewing
	_, err = g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":       "reviewing",
		"pause_reason": nil,
		"updated_at":   gtime.Now(),
	})
	if err != nil {
		return err
	}

	// 4. 异步执行审核流程（不阻塞 API 响应）
	// 使用独立后台 context，避免请求返回后 ctx 被回收导致审核链中断
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Workflow] runReviewAsync panic: project=%d err=%v", projectID, r)
			}
		}()
		s.runReviewAsync(context.Background(), projectID)
	}()

	return nil
}

// runReviewAsync 异步执行方案审核流程
func (s *Scheduler) runReviewAsync(ctx context.Context, projectID int64) {
	g.Log().Infof(ctx, "[Workflow] 项目 %d 进入方案审核阶段", projectID)

	result, err := RunReview(ctx, projectID)
	if err != nil {
		g.Log().Errorf(ctx, "[Workflow] 方案审核执行失败: project=%d, err=%v", projectID, err)
		// 审核执行失败，退回 designing，保留完整错误链
		HandleReviewFailure(ctx, projectID, &ReviewResult{
			Errors: []ReviewIssue{{
				Severity: "error",
				Message:  fmt.Sprintf("审核流程执行异常: %s", FormatErrorChain(err)),
			}},
		})
		return
	}

	if result.Passed {
		g.Log().Infof(ctx, "[Workflow] 项目 %d 方案审核通过，进入执行阶段", projectID)
		if err := HandleReviewSuccess(ctx, projectID, result); err != nil {
			g.Log().Errorf(ctx, "[Workflow] 审核通过后启动失败: project=%d, err=%v", projectID, err)
		}
	} else {
		g.Log().Infof(ctx, "[Workflow] 项目 %d 方案审核未通过（%d errors），退回设计阶段",
			projectID, len(result.Errors))
		HandleReviewFailure(ctx, projectID, result)
	}
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

// Resume 恢复项目（重新开始调度）
// 如果已有 pending 任务，直接进入 running（跳过审核）
// 如果只有 draft 任务，走审核流程
func (s *Scheduler) Resume(ctx context.Context, projectID int64) error {
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("项目不存在")
	}
	if project["status"].String() != "paused" {
		return fmt.Errorf("当前状态(%s)不允许恢复", project["status"].String())
	}

	// 检查是否有 pending 任务（说明之前已经审核过，直接恢复执行）
	pendingCount, _ := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		Where("deleted_at IS NULL").
		Count()
	if pendingCount > 0 {
		// 直接恢复执行
		_, err = g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
			"status":       "running",
			"pause_reason": nil,
			"updated_at":   gtime.Now(),
		})
		if err != nil {
			return err
		}
		s.StartProject(projectID)
		return nil
	}

	// 只有 draft 任务，走审核流程
	return s.ConfirmPlan(ctx, projectID)
}

// CreateProject 创建项目并初始化架构师对话。
// selectedPresetIDs 为用户显式选择的项目角色预设；为空时不生成项目角色配置，运行时直接回退到分类默认预设。
func CreateProject(ctx context.Context, name, projectCategory, description, workDir string, architectModelID int64, userID int64, deptID int64, selectedPresetIDs []int64, engineVersion ...string) (int64, int64, error) {
	// 1.5 默认分类
	if projectCategory == "" {
		projectCategory = "软件开发"
	}

	projectID := int64(snowflake.Generate())

	// 解析 category_code（通过 CategoryResolver 将中文展示名映射为稳定编码）
	catInfo, _ := GetCategoryResolver().ResolveByDisplayName(ctx, projectCategory)
	categoryCode := ""
	if catInfo != nil {
		categoryCode = catInfo.CategoryCode
	}

	// 非编码类项目如果未指定工作目录，自动生成
	if workDir == "" {
		workDir = GenerateWorkDir(projectCategory, projectID)
	}
	if workDir != "" {
		var err error
		workDir, _, err = EnsureWorkDir(workDir)
		if err != nil {
			return 0, 0, err
		}
	}

	// 1. 读取用户显式选择的项目角色预设；未选择时不在创建阶段生成项目角色配置
	var presets gdb.Result
	var err error
	if len(selectedPresetIDs) > 0 {
		presets, err = repo.ListRolePresets(ctx, repo.RolePresetQuery{
			IDs:             selectedPresetIDs,
			CategoryCode:    categoryCode,
			ProjectCategory: projectCategory,
		})
	}
	if err != nil {
		return 0, 0, fmt.Errorf("读取角色预设失败: %w", err)
	}
	projectArchitectModelID := architectModelID
	if projectArchitectModelID == 0 {
		for _, p := range presets {
			if p["role_type"].String() == "architect" && p["model_id"].Int64() > 0 {
				projectArchitectModelID = p["model_id"].Int64()
				break
			}
		}
	}

	// 解析引擎版本（默认 workflow_v2，仅显式指定 legacy 时走旧链）
	ev := "workflow_v2"
	if len(engineVersion) > 0 && engineVersion[0] == "legacy" {
		ev = "legacy"
	}

	// 2. 创建项目
	_, err = g.DB().Model("mvp_project").Insert(g.Map{
		"id":                 projectID,
		"name":               name,
		"project_category":   projectCategory,
		"category_code":      categoryCode,
		"description":        description,
		"status":             "designing",
		"engine_version":     ev,
		"work_dir":           workDir,
		"architect_model_id": projectArchitectModelID,
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

	// 4. 只为用户显式选择的预设创建项目角色配置
	for _, p := range presets {
		roleType := p["role_type"].String()
		modelID := p["model_id"].Int64()

		// 架构师角色：若用户手动选择了模型，则覆盖预设模型
		if roleType == "architect" {
			modelID = architectModelID
			if modelID == 0 {
				modelID = p["model_id"].Int64()
			}
		}

		systemPrompt := presetutil.BuildRoleSystemPrompt(categoryCode, roleType, p["role_level"].String(), p["system_prompt"].String(), modelPromptMap[modelID])
		if roleType == "architect" {
			systemPrompt = presetutil.BuildArchitectSystemPrompt(name, description, categoryCode, systemPrompt)
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
			"model_id":         modelID,
			"system_prompt":    systemPrompt,
			"execution_mode":   executionMode,
			"status":           1,
			"created_by":       userID,
			"dept_id":          deptID,
			"created_at":       gtime.Now(),
			"updated_at":       gtime.Now(),
		})
		if err != nil {
			return 0, 0, fmt.Errorf("创建角色配置(%s)失败: %w", roleType, err)
		}
	}
	// 仅当用户显式选择了架构师模型时，创建架构师项目角色覆盖
	if len(presets) == 0 && architectModelID > 0 {
		roleLevel := "max"
		executionMode := "chat"
		architectPreset, presetErr := repo.GetRolePreset(ctx, repo.RolePresetQuery{
			CategoryCode:    categoryCode,
			ProjectCategory: projectCategory,
			RoleType:        "architect",
			DefaultOnly:     true,
		})
		if presetErr == nil && architectPreset != nil {
			if architectPreset["role_level"].String() != "" {
				roleLevel = architectPreset["role_level"].String()
			}
			if architectPreset["execution_mode"].String() != "" {
				executionMode = architectPreset["execution_mode"].String()
			}
		}

		systemPrompt := presetutil.BuildRoleSystemPrompt(categoryCode, "architect", roleLevel, "", modelPromptMap[architectModelID])
		systemPrompt = presetutil.BuildArchitectSystemPrompt(name, description, categoryCode, systemPrompt)

		_, err = g.DB().Model("mvp_project_role").Insert(g.Map{
			"id":               int64(snowflake.Generate()),
			"project_id":       projectID,
			"project_category": projectCategory,
			"role_type":        "architect",
			"role_level":       roleLevel,
			"model_id":         architectModelID,
			"system_prompt":    systemPrompt,
			"execution_mode":   executionMode,
			"status":           1,
			"created_by":       userID,
			"dept_id":          deptID,
			"created_at":       gtime.Now(),
			"updated_at":       gtime.Now(),
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
