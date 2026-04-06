package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/provider"
)

// ReviewResult 审核结果
type ReviewResult struct {
	Passed    bool          `json:"passed"`
	Errors    []ReviewIssue `json:"errors"`
	Warnings  []ReviewIssue `json:"warnings"`
	AutoFixes []string      `json:"auto_fixes,omitempty"`
}

// ReviewIssue 审核问题
type ReviewIssue struct {
	TaskName string `json:"task_name"`
	Severity string `json:"severity"` // error / warning
	Message  string `json:"message"`
}

// AuditorReviewResult 审计员 AI 审核输出
type AuditorReviewResult struct {
	Approved    bool          `json:"approved"`
	Issues      []ReviewIssue `json:"issues"`
	Suggestions string        `json:"suggestions"`
}

// CoordinatorOptResult 协调员 AI 优化输出
type CoordinatorOptResult struct {
	OptimizedBatches map[string]struct {
		BatchNo int    `json:"batch_no"`
		Reason  string `json:"reason"`
	} `json:"optimized_batches"`
	ParallelismScore  float64  `json:"parallelism_score"`
	EstimatedDuration string   `json:"estimated_duration"`
	Warnings          []string `json:"warnings"`
}

// RunReview 执行方案审核流程（三步）
// 返回审核结果。调用者根据结果决定是进入 running 还是回退 designing
func RunReview(ctx context.Context, projectID int64) (*ReviewResult, error) {
	// 获取项目信息
	project, err := g.DB().Model("mvp_project").Where("id", projectID).One()
	if err != nil {
		return nil, fmt.Errorf("查询项目信息失败: %w", err)
	}
	if project.IsEmpty() {
		return nil, fmt.Errorf("项目 %d 不存在", projectID)
	}
	workDir := project["work_dir"].String()
	projectCategory := project["project_category"].String()

	// 获取所有 draft 任务
	tasks, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "draft").
		Where("deleted_at IS NULL").
		Order("batch_no ASC, sort ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("查询 draft 任务失败: %w", err)
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("没有待审核的 draft 任务")
	}

	// ======== Step 1: 系统预检（零 AI 消耗）========
	result := systemPrecheck(ctx, projectID, tasks, workDir, projectCategory)

	// error 数量 > 0 → 阻止进入下一步
	if len(result.Errors) > 0 {
		result.Passed = false
		return result, nil
	}

	// ======== Step 2: 审计员 AI 审核 ========
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(GetReviewTimeout(ctx))*time.Second)
	defer cancel()

	auditorResult, err := auditorReview(timeoutCtx, projectID, tasks)
	if err != nil {
		// 超时或错误，跳过 AI 审核，仅保留预检结果
		g.Log().Warningf(ctx, "[Review] 审计员审核失败/超时，跳过: project=%d, err=%v", projectID, err)
		result.Warnings = append(result.Warnings, ReviewIssue{
			Severity: "warning",
			Message:  fmt.Sprintf("审计员 AI 审核跳过（%v），仅保留系统预检结果", err),
		})
		result.Passed = true
		return result, nil
	}

	if !auditorResult.Approved {
		result.Passed = false
		result.Errors = append(result.Errors, auditorResult.Issues...)
		return result, nil
	}

	// 附加审计员的 warnings
	for _, issue := range auditorResult.Issues {
		if issue.Severity == "warning" {
			result.Warnings = append(result.Warnings, issue)
		}
	}

	// ======== Step 3: 协调员 AI 优化 ========
	optResult, err := coordinatorOptimize(timeoutCtx, projectID, tasks)
	if err != nil {
		g.Log().Warningf(ctx, "[Review] 协调员优化失败/超时，跳过: project=%d, err=%v", projectID, err)
		result.Warnings = append(result.Warnings, ReviewIssue{
			Severity: "warning",
			Message:  fmt.Sprintf("协调员优化跳过（%v）", err),
		})
	} else {
		// 应用协调员的批次优化建议
		applyCoordinatorOptimizations(ctx, projectID, tasks, optResult)
	}

	result.Passed = true
	return result, nil
}

// systemPrecheck 系统预检（零 AI 消耗，毫秒级）
func systemPrecheck(ctx context.Context, projectID int64, tasks gdb.Result, workDir string, projectCategory string) *ReviewResult {
	result := &ReviewResult{Passed: true}
	family := GetCategoryFamily(projectCategory)
	autoFixBatch := GetReviewAutoFixBatch(ctx)

	// 构建任务名集合（用于依赖校验）
	taskNames := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		taskNames[t["name"].String()] = true
	}

	// 构建批次→资源映射（用于冲突检测）
	batchResources := make(map[int]map[string]string) // batchNo → resource → taskName

	// 构建依赖关系 name→batchNo（用于 batch_no 一致性检查）
	nameToBatch := make(map[string]int, len(tasks))
	for _, t := range tasks {
		nameToBatch[t["name"].String()] = t["batch_no"].Int()
	}

	// 预加载项目角色配置（消除 N+1）
	roleConfigs, _ := g.DB().Model("mvp_project_role").
		Where("project_id", projectID).
		Where("status", 1).
		Where("deleted_at IS NULL").
		Fields("role_type, role_level").
		All()
	availableRoles := make(map[string]bool, len(roleConfigs))
	for _, rc := range roleConfigs {
		key := rc["role_type"].String() + "/" + rc["role_level"].String()
		availableRoles[key] = true
	}

	// 收集需要自动修正 batch_no 的任务（批量更新，避免逐条落库）
	type batchFix struct {
		taskID   int64
		name     string
		oldBatch int
		newBatch int
		dep      string
	}
	var batchFixes []batchFix

	for _, t := range tasks {
		name := t["name"].String()
		desc := t["description"].String()
		batchNo := t["batch_no"].Int()

		// 1. 任务名非空
		if strings.TrimSpace(name) == "" {
			result.Errors = append(result.Errors, ReviewIssue{
				TaskName: "(空)", Severity: "error", Message: "任务名称为空",
			})
			continue
		}

		// 2. 任务描述质量
		if utf8.RuneCountInString(strings.TrimSpace(desc)) < 10 {
			result.Errors = append(result.Errors, ReviewIssue{
				TaskName: name, Severity: "error",
				Message: fmt.Sprintf("任务描述过短（%d字），需要至少10字的有效描述", utf8.RuneCountInString(desc)),
			})
		}

		// 3. affected_resources 格式检查
		var resources []string
		resJSON := t["affected_resources"].String()
		if resJSON != "" && resJSON != "[]" && resJSON != "null" {
			if err := json.Unmarshal([]byte(resJSON), &resources); err != nil {
				result.Errors = append(result.Errors, ReviewIssue{
					TaskName: name, Severity: "error",
					Message: "affected_resources 格式非法: " + err.Error(),
				})
			}
		}

		for _, res := range resources {
			// 路径格式合法性（无乱码/特殊字符）
			if containsGarbage(res) {
				result.Errors = append(result.Errors, ReviewIssue{
					TaskName: name, Severity: "error",
					Message: fmt.Sprintf("affected_resources 包含疑似乱码路径: %s", res),
				})
			}
		}

		// 4. 编码类项目：检查文件/目录是否存在
		if family == CategoryFamilyCoding && workDir != "" {
			for _, res := range resources {
				absPath := res
				if !filepath.IsAbs(res) {
					absPath = filepath.Join(workDir, res)
				}
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					result.Warnings = append(result.Warnings, ReviewIssue{
						TaskName: name, Severity: "warning",
						Message: fmt.Sprintf("文件/目录不存在: %s（可能是新建文件，请确认）", res),
					})
				}
			}
		}

		// 5. depends_on 有效性
		var dependsOn []string
		depJSON := t["depends_on"].String()
		if depJSON != "" && depJSON != "[]" && depJSON != "null" {
			json.Unmarshal([]byte(depJSON), &dependsOn)
		}
		for _, dep := range dependsOn {
			if !taskNames[dep] {
				result.Errors = append(result.Errors, ReviewIssue{
					TaskName: name, Severity: "error",
					Message: fmt.Sprintf("depends_on 引用了不存在的任务: %s", dep),
				})
			}
		}

		// 6. batch_no 一致性：有依赖的任务 batch_no > 被依赖任务的 batch_no
		for _, dep := range dependsOn {
			if depBatch, ok := nameToBatch[dep]; ok {
				if batchNo <= depBatch {
					if autoFixBatch {
						newBatch := depBatch + 1
						batchFixes = append(batchFixes, batchFix{
							taskID: t["id"].Int64(), name: name,
							oldBatch: batchNo, newBatch: newBatch, dep: dep,
						})
						result.AutoFixes = append(result.AutoFixes,
							fmt.Sprintf("自动修正 [%s] batch_no: %d → %d（依赖 [%s] batch=%d）",
								name, batchNo, newBatch, dep, depBatch))
						nameToBatch[name] = newBatch
						batchNo = newBatch
					} else {
						result.Warnings = append(result.Warnings, ReviewIssue{
							TaskName: name, Severity: "warning",
							Message: fmt.Sprintf("batch_no(%d) <= 被依赖任务 [%s] 的 batch_no(%d)", batchNo, dep, depBatch),
						})
					}
				}
			}
		}

		// 7. 资源冲突预检：同 batch_no 内不能两个任务改同一文件
		if _, ok := batchResources[batchNo]; !ok {
			batchResources[batchNo] = make(map[string]string)
		}
		for _, res := range resources {
			if existingTask, conflict := batchResources[batchNo][res]; conflict {
				result.Errors = append(result.Errors, ReviewIssue{
					TaskName: name, Severity: "error",
					Message: fmt.Sprintf("资源冲突: 同批次(%d)中 [%s] 和 [%s] 都修改 %s", batchNo, existingTask, name, res),
				})
			}
			batchResources[batchNo][res] = name
		}

		// 8. role_level 覆盖检查（使用预加载数据，无 DB 查询）
		roleType := t["role_type"].String()
		roleLevel := t["role_level"].String()
		if roleType != "" && roleLevel != "" {
			if !availableRoles[roleType+"/"+roleLevel] {
				result.Warnings = append(result.Warnings, ReviewIssue{
					TaskName: name, Severity: "warning",
					Message: fmt.Sprintf("项目未配置 %s/%s 角色，任务可能无法执行", roleType, roleLevel),
				})
			}
		}
	}

	// 批量执行 batch_no 自动修正（事务内）
	if len(batchFixes) > 0 {
		if txErr := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			for _, fix := range batchFixes {
				if _, err := tx.Model("mvp_task").Where("id", fix.taskID).Update(g.Map{
					"batch_no":   fix.newBatch,
					"updated_at": gtime.Now(),
				}); err != nil {
					return fmt.Errorf("自动修正 batch_no 失败: task=%d, err=%w", fix.taskID, err)
				}
			}
			return nil
		}); txErr != nil {
			g.Log().Warningf(ctx, "[Review] batch_no 自动修正事务失败: %v", txErr)
		}
	}

	return result
}

// containsGarbage 检测路径是否包含疑似乱码
func containsGarbage(path string) bool {
	// 检测不可打印字符、连续特殊符号等
	garbageRe := regexp.MustCompile(`[\x00-\x1f]|[^\x00-\x7f]{10,}|[!@#$%^&*()+=\[\]{}|\\;':",<>?]{3,}`)
	return garbageRe.MatchString(path)
}

// auditorReview 审计员 AI 审核（legacy 入口）
func auditorReview(ctx context.Context, projectID int64, tasks gdb.Result) (*AuditorReviewResult, error) {
	modelInfo, err := getReviewRoleModel(ctx, projectID, "auditor")
	if err != nil {
		return nil, fmt.Errorf("获取审计员模型失败: %w", err)
	}
	return doAuditorReview(ctx, modelInfo, tasks)
}

// doAuditorReview 审计员 AI 审核核心逻辑。
func doAuditorReview(ctx context.Context, modelInfo *ModelInfo, tasks gdb.Result) (*AuditorReviewResult, error) {
	taskSummaries := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		taskSummaries = append(taskSummaries, map[string]interface{}{
			"name":               t["name"].String(),
			"description":        truncate(t["description"].String(), 200),
			"role_type":          t["role_type"].String(),
			"role_level":         t["role_level"].String(),
			"batch_no":           t["batch_no"].Int(),
			"affected_resources": t["affected_resources"].String(),
			"depends_on":         t["depends_on"].String(),
		})
	}
	summaryJSON, _ := json.MarshalIndent(taskSummaries, "", "  ")

	prompt := fmt.Sprintf(`请审核以下任务清单的质量。审核维度：
1. 任务描述是否明确可执行
2. 任务粒度是否合理（不能太大也不能太碎）
3. 依赖关系是否合理
4. 是否遗漏了关键模块
5. 角色分配是否合理

任务清单（共 %d 个）：
%s

请严格输出 JSON，格式如下：
{"approved": true/false, "issues": [{"task_name": "xxx", "severity": "error/warning", "message": "问题描述"}], "suggestions": "整体建议"}`, len(tasks), string(summaryJSON))

	// 调用 AI
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.3,
		SystemPrompt: modelInfo.SystemPrompt,
	})
	if err != nil {
		return nil, err
	}

	// 解析审核结果
	var auditorResult AuditorReviewResult
	if err := parseJSONFromAI(resp.Content, &auditorResult); err != nil {
		return nil, fmt.Errorf("解析审计员审核结果失败: %w, 原文: %s", err, truncate(resp.Content, 500))
	}

	return &auditorResult, nil
}

// coordinatorOptimize 协调员 AI 优化（legacy 入口）
func coordinatorOptimize(ctx context.Context, projectID int64, tasks gdb.Result) (*CoordinatorOptResult, error) {
	modelInfo, err := getReviewRoleModel(ctx, projectID, "coordinator")
	if err != nil {
		return nil, fmt.Errorf("获取协调员模型失败: %w", err)
	}
	return doCoordinatorOptimize(ctx, modelInfo, tasks)
}

// doCoordinatorOptimize 协调员 AI 优化核心逻辑。
func doCoordinatorOptimize(ctx context.Context, modelInfo *ModelInfo, tasks gdb.Result) (*CoordinatorOptResult, error) {
	taskSummaries := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		taskSummaries = append(taskSummaries, map[string]interface{}{
			"name":               t["name"].String(),
			"batch_no":           t["batch_no"].Int(),
			"affected_resources": t["affected_resources"].String(),
			"depends_on":         t["depends_on"].String(),
			"role_type":          t["role_type"].String(),
		})
	}
	summaryJSON, _ := json.MarshalIndent(taskSummaries, "", "  ")

	prompt := fmt.Sprintf(`请优化以下任务清单的调度计划。优化维度：
1. 资源冲突精细检测（同批次任务不应修改相同文件）
2. 批次顺序优化（最大化并行度）
3. 并行度评估
4. 预估执行时间

任务清单（共 %d 个）：
%s

请严格输出 JSON，格式如下：
{"optimized_batches": {"任务名": {"batch_no": 1, "reason": "调整原因"}}, "parallelism_score": 0.8, "estimated_duration": "约2小时", "warnings": []}`, len(tasks), string(summaryJSON))

	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.3,
		SystemPrompt: modelInfo.SystemPrompt,
	})
	if err != nil {
		return nil, err
	}

	var optResult CoordinatorOptResult
	if err := parseJSONFromAI(resp.Content, &optResult); err != nil {
		return nil, fmt.Errorf("解析协调员优化结果失败: %w", err)
	}

	return &optResult, nil
}

// applyCoordinatorOptimizations 应用协调员的批次优化建议（事务内批量执行）
func applyCoordinatorOptimizations(ctx context.Context, projectID int64, tasks gdb.Result, opt *CoordinatorOptResult) {
	if opt == nil || len(opt.OptimizedBatches) == 0 {
		return
	}

	// 先收集需要更新的任务
	type batchUpdate struct {
		taskID   int64
		newBatch int
	}
	var updates []batchUpdate
	for _, t := range tasks {
		name := t["name"].String()
		if batch, ok := opt.OptimizedBatches[name]; ok {
			if batch.BatchNo > 0 && batch.BatchNo != t["batch_no"].Int() {
				updates = append(updates, batchUpdate{taskID: t["id"].Int64(), newBatch: batch.BatchNo})
			}
		}
	}

	if len(updates) == 0 {
		return
	}

	// 事务内批量更新
	if err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, u := range updates {
			if _, err := tx.Model("mvp_task").Where("id", u.taskID).Update(g.Map{
				"batch_no":   u.newBatch,
				"updated_at": gtime.Now(),
			}); err != nil {
				return fmt.Errorf("更新 task=%d batch_no 失败: %w", u.taskID, err)
			}
		}
		return nil
	}); err != nil {
		g.Log().Errorf(ctx, "[Review] 协调员批次优化事务失败: project=%d, err=%v", projectID, err)
		return
	}

	g.Log().Infof(ctx, "[Review] 协调员优化：应用了 %d 个批次调整", len(updates))
}

// getReviewRoleModel 获取审核角色的 AI 模型信息
func getReviewRoleModel(ctx context.Context, projectID int64, roleType string) (*ModelInfo, error) {
	role, err := g.DB().Model("mvp_project_role").
		Where("project_id", projectID).
		Where("role_type", roleType).
		Where("status", 1).
		Where("deleted_at IS NULL").
		One()
	if err != nil || role.IsEmpty() {
		return nil, fmt.Errorf("项目未配置 %s 角色", roleType)
	}

	modelID := role["model_id"].Int64()
	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil || model.IsEmpty() {
		return nil, fmt.Errorf("AI模型 %d 不存在", modelID)
	}

	systemPrompt := role["system_prompt"].String()
	if systemPrompt == "" {
		systemPrompt = model["role_prompt"].String()
	}

	return &ModelInfo{
		ModelID:      modelID,
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
		SystemPrompt: systemPrompt,
		MaxTokens:    model["max_tokens"].Int(),
	}, nil
}

// HandleReviewFailure 审核失败时的处理：退回 designing + 通知架构师
func HandleReviewFailure(ctx context.Context, projectID int64, result *ReviewResult) error {
	// 1. 退回项目状态
	_, err := g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":     "designing",
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return err
	}

	// 2. 将 issues 汇总为通知消息
	var msg strings.Builder
	msg.WriteString("## 方案审核未通过\n\n")

	if len(result.Errors) > 0 {
		msg.WriteString("### 错误（必须修复）\n")
		for i, issue := range result.Errors {
			taskRef := ""
			if issue.TaskName != "" {
				taskRef = fmt.Sprintf("[%s] ", issue.TaskName)
			}
			msg.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue.Message))
		}
		msg.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		msg.WriteString("### 警告（建议修复）\n")
		for i, issue := range result.Warnings {
			taskRef := ""
			if issue.TaskName != "" {
				taskRef = fmt.Sprintf("[%s] ", issue.TaskName)
			}
			msg.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue.Message))
		}
		msg.WriteString("\n")
	}

	if len(result.AutoFixes) > 0 {
		msg.WriteString("### 自动修正\n")
		for _, fix := range result.AutoFixes {
			msg.WriteString("- " + fix + "\n")
		}
	}

	msg.WriteString("\n请修正上述问题后重新确认方案。")

	// 3. 在架构师对话中发送通知
	notifyProjectArchitectConversation(ctx, projectID, msg.String())

	return nil
}

// HandleReviewSuccess 审核通过：draft → pending，项目进入 running
// 事务边界：确认任务 + 追加 warning + 项目状态改 running 在同一事务中
func HandleReviewSuccess(ctx context.Context, projectID int64, result *ReviewResult) error {
	// 1. 确认 draft → pending（自带全量确认或回滚保证）
	confirmedCount, err := GetParser().ConfirmDraftTasks(ctx, projectID)
	if err != nil {
		return fmt.Errorf("确认草稿任务失败: %w", err)
	}
	if confirmedCount == 0 {
		return fmt.Errorf("没有任务可确认")
	}

	// 2. 事务内：追加 warning + 项目状态改 running
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 2a. 附加 warnings 到对应任务的描述中（按任务名聚合）
		warningsByTask := make(map[string][]string)
		for _, w := range result.Warnings {
			if w.TaskName != "" {
				warningsByTask[w.TaskName] = append(warningsByTask[w.TaskName], w.Message)
			}
		}
		for taskName, msgs := range warningsByTask {
			suffix := ""
			for _, msg := range msgs {
				suffix += fmt.Sprintf("\n\n⚠️ 审核警告: %s", msg)
			}
			if _, err := tx.Model("mvp_task").
				Where("project_id", projectID).
				Where("name", taskName).
				Where("status", "pending").
				Where("deleted_at IS NULL").
				Update(g.Map{
					"description": gdb.Raw(fmt.Sprintf("CONCAT(description, '%s')", strings.ReplaceAll(suffix, "'", "''"))),
					"updated_at":  gtime.Now(),
				}); err != nil {
				return fmt.Errorf("附加审核警告失败: task=%s, err=%w", taskName, err)
			}
		}

		// 2b. 更新项目状态为 running
		if _, err := tx.Model("mvp_project").Where("id", projectID).Update(g.Map{
			"status":       "running",
			"pause_reason": nil,
			"updated_at":   gtime.Now(),
		}); err != nil {
			return fmt.Errorf("项目状态更新失败: %w", err)
		}

		return nil
	})
	if err != nil {
		// 事务失败：回滚已确认的任务（pending → draft）
		g.Log().Errorf(ctx, "[Review] 事务失败，回滚 %d 个已确认任务: project=%d, err=%v", confirmedCount, projectID, err)
		rollbackConfirmedTasks(ctx, projectID)
		return fmt.Errorf("审核通过处理失败: %w", err)
	}

	// 4. 压缩架构师对话为全局上下文
	if compErr := GetCompressor().CompressProjectContext(context.Background(), projectID); compErr != nil {
		g.Log().Errorf(ctx, "[Review] 压缩项目上下文失败（非致命）: project=%d, err=%v", projectID, compErr)
	}

	// 5. 启动调度器
	GetScheduler().StartProject(projectID)

	// 6. 如果有 warnings，在架构师对话中通知
	if len(result.Warnings) > 0 || len(result.AutoFixes) > 0 {
		var msg strings.Builder
		msg.WriteString("## 方案审核通过\n\n")
		if len(result.AutoFixes) > 0 {
			msg.WriteString("### 自动修正\n")
			for _, fix := range result.AutoFixes {
				msg.WriteString("- " + fix + "\n")
			}
			msg.WriteString("\n")
		}
		if len(result.Warnings) > 0 {
			msg.WriteString("### 注意事项\n")
			for _, w := range result.Warnings {
				msg.WriteString(fmt.Sprintf("- [%s] %s\n", w.TaskName, w.Message))
			}
		}
		msg.WriteString("\n项目已开始执行。")
		notifyProjectArchitectConversation(ctx, projectID, msg.String())
	}

	return nil
}

// rollbackConfirmedTasks 将项目中已确认的 pending 任务回退为 draft
func rollbackConfirmedTasks(ctx context.Context, projectID int64) {
	taskIDs, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		Where("deleted_at IS NULL").
		Fields("id").
		Array()
	if err != nil {
		g.Log().Errorf(ctx, "[Review] 回滚查询失败: project=%d, err=%v", projectID, err)
		return
	}
	for _, idVal := range taskIDs {
		if _, err := updateTaskStatus(ctx, idVal.Int64(), "pending", "draft", nil); err != nil {
			g.Log().Errorf(ctx, "[Review] 回滚 task=%d pending→draft 失败: %v", idVal.Int64(), err)
		}
	}
}

// --- 辅助函数 ---

// parseJSONFromAI 从 AI 回复中提取 JSON 并解析
func parseJSONFromAI(content string, v interface{}) error {
	content = strings.TrimSpace(content)

	// 尝试直接解析
	if err := json.Unmarshal([]byte(content), v); err == nil {
		return nil
	}

	// 从 ```json 代码块中提取
	re := regexp.MustCompile("(?s)```json\\s*\\n?(\\{[\\s\\S]*?\\})\\s*```")
	if match := re.FindStringSubmatch(content); len(match) == 2 {
		if err := json.Unmarshal([]byte(match[1]), v); err == nil {
			return nil
		}
	}

	// 查找最外层的 { ... }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), v); err == nil {
			return nil
		}
	}

	return fmt.Errorf("无法从 AI 回复中提取有效 JSON")
}

// ModelInfo 已在 executor.go 中定义，此处复用
// 注意：如果 ModelInfo 未导出，需要确认是否在同一个 package 中
var _ = (*ModelInfo)(nil) // 编译期验证 ModelInfo 可访问
