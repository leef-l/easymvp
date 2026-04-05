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

// ConfirmPlan 确认实施方案，进入方案审核阶段
// 流程：designing → reviewing → (审核通过) → running
//       reviewing → (审核不通过) → designing
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
	go s.runReviewAsync(context.Background(), projectID)

	return nil
}

// runReviewAsync 异步执行方案审核流程
func (s *Scheduler) runReviewAsync(ctx context.Context, projectID int64) {
	g.Log().Infof(ctx, "[Workflow] 项目 %d 进入方案审核阶段", projectID)

	result, err := RunReview(ctx, projectID)
	if err != nil {
		g.Log().Errorf(ctx, "[Workflow] 方案审核执行失败: project=%d, err=%v", projectID, err)
		// 审核执行失败，退回 designing
		HandleReviewFailure(ctx, projectID, &ReviewResult{
			Errors: []ReviewIssue{{
				Severity: "error",
				Message:  fmt.Sprintf("审核流程执行异常: %v", err),
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

// CreateProject 创建项目并初始化架构师对话
// architectModelID 为前端传入的架构师模型，若为 0 则从预设读取
func CreateProject(ctx context.Context, name, projectCategory, description, workDir string, architectModelID int64, userID int64, deptID int64, engineVersion ...string) (int64, int64, error) {
	// 1.5 默认分类
	if projectCategory == "" {
		projectCategory = "软件开发"
	}

	projectID := int64(snowflake.Generate())

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

	// 1. 按项目分类读取默认角色预设模板（is_default=1）
	presets, err := g.DB().Model("mvp_role_preset").
		Where("status", 1).
		Where("project_category", projectCategory).
		Where("is_default", 1).
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
		"description":        description,
		"status":             "designing",
		"engine_version":     ev,
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
			systemPrompt = buildArchitectPromptWithCategory(name, description, modelPromptMap[modelID], projectCategory)
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
// projectCategory 用于选择分类族对应的兜底模板
func buildArchitectPrompt(projectName, projectDesc string, modelRolePrompt string) string {
	return buildArchitectPromptWithCategory(projectName, projectDesc, modelRolePrompt, "")
}

// buildArchitectPromptWithCategory 带项目分类的架构师提示词构建
func buildArchitectPromptWithCategory(projectName, projectDesc string, modelRolePrompt string, projectCategory string) string {
	projectContext := fmt.Sprintf("\n\n===== 当前项目 =====\n项目名称：%s\n项目简介：%s", projectName, projectDesc)

	if modelRolePrompt != "" {
		return modelRolePrompt + projectContext
	}

	// 根据分类族选择兜底模板
	family := GetCategoryFamily(projectCategory)
	switch family {
	case CategoryFamilyCreative:
		return buildCreativeArchitectPrompt(projectName, projectDesc)
	case CategoryFamilyAnalysis:
		return buildAnalysisArchitectPrompt(projectName, projectDesc)
	default:
		return buildCodingArchitectPrompt(projectName, projectDesc)
	}
}

func buildCodingArchitectPrompt(projectName, projectDesc string) string {
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

func buildCreativeArchitectPrompt(projectName, projectDesc string) string {
	return fmt.Sprintf(`你是一个资深创意总监，负责项目「%s」的创意策划和内容架构设计。

项目简介：
%s

你的职责：
1. 与用户沟通，深入理解创意需求和风格方向
2. 设计内容架构（世界观、人物关系、故事线、章节/幕结构）
3. 将项目按内容模块拆分为可并行的创作任务
4. 确保不同创作者之间的内容不冲突（不覆盖同一章节/角色/场景）
5. 为每个任务标注：任务名称、描述、角色等级(lite/pro/max)、执行批次号、涉及的资源范围、依赖关系

任务拆分规则：
- 按章节/幕/场景为最小创作单元
- 人物设定、世界观设定为前置任务（批次 1）
- 正文创作按剧情线并行，按时间线串行
- affected_resources 使用相对路径（如 "chapters/ch01.md"）

输出任务清单时请使用 JSON 格式：
{
  "tasks": [
    {
      "name": "任务名称",
      "description": "详细描述（含创作要求和风格指南）",
      "role_level": "max/pro/lite",
      "batch_no": 1,
      "affected_resources": ["chapters/ch01.md"],
      "depends_on": []
    }
  ]
}`, projectName, projectDesc)
}

func buildAnalysisArchitectPrompt(projectName, projectDesc string) string {
	return fmt.Sprintf(`你是一个资深分析架构师，负责项目「%s」的分析方案设计和任务编排。

项目简介：
%s

你的职责：
1. 与用户沟通，明确分析目标、数据源和交付物
2. 设计分析流程（数据采集 → 清洗 → 分析 → 可视化 → 报告）
3. 将项目按分析阶段和数据维度拆分为可执行的任务
4. 确保数据管道的上下游依赖关系正确
5. 为每个任务标注：任务名称、描述、角色等级(lite/pro/max)、执行批次号、涉及的资源范围、依赖关系

任务拆分规则：
- 数据源接入和清洗为前置任务（批次 1）
- 不同维度的分析可以并行
- 汇总报告和可视化依赖所有分析任务
- affected_resources 使用相对路径（如 "reports/summary.md", "data/cleaned.csv"）

输出任务清单时请使用 JSON 格式：
{
  "tasks": [
    {
      "name": "任务名称",
      "description": "详细描述（含分析方法、数据源、预期输出）",
      "role_level": "max/pro/lite",
      "batch_no": 1,
      "affected_resources": ["reports/summary.md"],
      "depends_on": []
    }
  ]
}`, projectName, projectDesc)
}
