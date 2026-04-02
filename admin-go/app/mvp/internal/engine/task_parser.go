package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// TaskParser 任务解析器
// 从架构师 AI 回复中提取 JSON 任务清单，解析后写入 mvp_task（status=draft）
type TaskParser struct{}

var defaultParser = &TaskParser{}

// GetParser 获取全局解析器
func GetParser() *TaskParser {
	return defaultParser
}

// ArchitectTaskPlan 架构师输出的任务规划结构
type ArchitectTaskPlan struct {
	Tasks []ArchitectTask `json:"tasks"`
}

// ArchitectTask 架构师规划的单个任务
type ArchitectTask struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	RoleLevel         string   `json:"role_level"`          // lite/pro/max
	RoleType          string   `json:"role_type,omitempty"` // 默认 implementer
	BatchNo           int      `json:"batch_no"`
	Sort              int      `json:"sort,omitempty"`
	AffectedResources []string `json:"affected_resources"`
	DependsOn         []string `json:"depends_on"`       // 依赖的任务名称列表
	ParentName        string   `json:"parent,omitempty"` // 父任务名称（用于树形结构）
}

// ParseAndCreateTasks 从架构师回复中解析任务清单并写入数据库
// 返回创建的任务数量
func (p *TaskParser) ParseAndCreateTasks(ctx context.Context, projectID int64, aiReply string) (int, error) {
	// 1. 从 AI 回复中提取 JSON
	plan, err := p.extractTaskPlan(aiReply)
	if err != nil {
		return 0, err
	}
	if len(plan.Tasks) == 0 {
		return 0, nil // 回复中没有任务清单，正常情况（还在讨论需求）
	}

	// 2. 查询项目的创建人和部门（任务继承项目的 created_by 和 dept_id）
	projectInfo, err := g.DB().Model("mvp_project").
		Where("id", projectID).
		Fields("created_by, dept_id").
		One()
	if err != nil {
		return 0, fmt.Errorf("查询项目信息失败: %w", err)
	}
	createdBy := projectInfo["created_by"].Int64()
	deptID := projectInfo["dept_id"].Int64()

	// 3. 清理该项目所有已有任务、任务日志和依赖关系（重新拆分以最新方案为准）
	// 2a. 获取所有任务ID，用于清理日志和依赖
	oldTaskIDs, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("deleted_at IS NULL").
		Fields("id").
		Array()
	if err != nil {
		return 0, fmt.Errorf("查询旧任务失败: %w", err)
	}
	if len(oldTaskIDs) > 0 {
		// 2b. 软删除任务日志
		_, _ = g.DB().Model("mvp_task_log").
			WhereIn("task_id", oldTaskIDs).
			Where("deleted_at IS NULL").
			Update(g.Map{
				"deleted_at": gtime.Now(),
				"updated_at": gtime.Now(),
			})
		// 2c. 删除任务依赖关系
		_, _ = g.DB().Model("mvp_task_dependency").
			WhereIn("task_id", oldTaskIDs).
			Delete()
		_, _ = g.DB().Model("mvp_task_dependency").
			WhereIn("depends_on_id", oldTaskIDs).
			Delete()
	}
	// 2d. 软删除所有任务
	_, err = g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("deleted_at IS NULL").
		Update(g.Map{
			"deleted_at": gtime.Now(),
			"updated_at": gtime.Now(),
		})
	if err != nil {
		return 0, fmt.Errorf("清理旧任务失败: %w", err)
	}

	// 3. 第一遍：创建所有任务，建立 name→id 映射
	nameToID := make(map[string]int64, len(plan.Tasks))
	taskIDs := make([]int64, 0, len(plan.Tasks))

	for i, t := range plan.Tasks {
		taskID := int64(snowflake.Generate())
		nameToID[t.Name] = taskID
		taskIDs = append(taskIDs, taskID)

		roleType := t.RoleType
		if roleType == "" {
			roleType = "implementer" // 默认角色
		}

		sort := t.Sort
		if sort == 0 {
			sort = i + 1 // 按输出顺序
		}

		// affected_resources 转 JSON
		resourcesJSON := "[]"
		if len(t.AffectedResources) > 0 {
			b, _ := json.Marshal(t.AffectedResources)
			resourcesJSON = string(b)
		}

		// depends_on 存名称（后续第二遍转为 ID 关系）
		dependsJSON := "[]"
		if len(t.DependsOn) > 0 {
			b, _ := json.Marshal(t.DependsOn)
			dependsJSON = string(b)
		}

		_, err = g.DB().Model("mvp_task").Insert(g.Map{
			"id":                 taskID,
			"project_id":        projectID,
			"parent_id":         0,
			"name":              t.Name,
			"description":       t.Description,
			"role_type":         roleType,
			"role_level":        t.RoleLevel,
			"status":            "draft", // 草稿状态，确认后才变 pending
			"batch_no":          t.BatchNo,
			"sort":              sort,
			"affected_resources": resourcesJSON,
			"depends_on":        dependsJSON,
			"created_by":        createdBy,
			"dept_id":           deptID,
			"created_at":        gtime.Now(),
			"updated_at":        gtime.Now(),
		})
		if err != nil {
			g.Log().Errorf(ctx, "创建任务失败 [%s]: %v", t.Name, err)
			continue
		}
	}

	// 4. 第二遍：建立父子关系
	for _, t := range plan.Tasks {
		if t.ParentName == "" {
			continue
		}
		parentID, ok := nameToID[t.ParentName]
		if !ok {
			continue
		}
		taskID := nameToID[t.Name]
		g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
			"parent_id": parentID,
		})
	}

	// 5. 第三遍：建立依赖关系（mvp_task_dependency）
	for _, t := range plan.Tasks {
		if len(t.DependsOn) == 0 {
			continue
		}
		taskID := nameToID[t.Name]
		for _, depName := range t.DependsOn {
			depID, ok := nameToID[depName]
			if !ok {
				g.Log().Warningf(ctx, "任务 [%s] 依赖 [%s] 未找到，跳过", t.Name, depName)
				continue
			}
			g.DB().Model("mvp_task_dependency").Insert(g.Map{
				"task_id":       taskID,
				"depends_on_id": depID,
			})
		}
	}

	g.Log().Infof(ctx, "[TaskParser] 项目 %d 解析创建 %d 个任务（draft）", projectID, len(taskIDs))
	return len(taskIDs), nil
}

// DryParseTaskCount 仅解析AI回复，返回任务数量（不写入数据库）
func (p *TaskParser) DryParseTaskCount(aiReply string) int {
	plan, err := p.extractTaskPlan(aiReply)
	if err != nil || len(plan.Tasks) == 0 {
		return 0
	}
	return len(plan.Tasks)
}

// ConfirmDraftTasks 确认草稿任务：draft → pending
func (p *TaskParser) ConfirmDraftTasks(ctx context.Context, projectID int64) (int, error) {
	result, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "draft").
		Where("deleted_at IS NULL").
		Update(g.Map{
			"status":     "pending",
			"updated_at": gtime.Now(),
		})
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), nil
}

// GetDraftCount 获取项目草稿任务数量
func (p *TaskParser) GetDraftCount(ctx context.Context, projectID int64) int {
	count, _ := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "draft").
		Where("deleted_at IS NULL").
		Count()
	return count
}

// extractTaskPlan 从 AI 回复文本中提取 JSON 任务清单
// 支持多种格式：
//  1. 标准 ```json ... ``` 代码块
//  2. { "tasks": [...] } 直接 JSON
//  3. 混合文本中的 JSON 片段
func (p *TaskParser) extractTaskPlan(text string) (*ArchitectTaskPlan, error) {
	// 策略1：从 ```json 代码块中提取
	codeBlockRe := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(\\{[\\s\\S]*?\\})\\s*```")
	matches := codeBlockRe.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if plan, err := p.tryParseJSON(m[1]); err == nil && len(plan.Tasks) > 0 {
			return plan, nil
		}
	}

	// 策略2：直接查找最外层的 { "tasks": [...] }
	tasksRe := regexp.MustCompile(`(?s)\{\s*"tasks"\s*:\s*\[[\s\S]*\]\s*\}`)
	if m := tasksRe.FindString(text); m != "" {
		if plan, err := p.tryParseJSON(m); err == nil && len(plan.Tasks) > 0 {
			return plan, nil
		}
	}

	// 策略3：查找独立的 JSON 数组 [{ ... }]（直接就是 tasks 数组）
	arrayRe := regexp.MustCompile(`(?s)\[\s*\{[\s\S]*\}\s*\]`)
	if m := arrayRe.FindString(text); m != "" {
		cleaned := p.cleanJSON(m)
		var tasks []ArchitectTask
		if err := json.Unmarshal([]byte(cleaned), &tasks); err == nil && len(tasks) > 0 {
			return &ArchitectTaskPlan{Tasks: tasks}, nil
		}
	}

	// 没找到有效的任务清单
	return &ArchitectTaskPlan{}, nil
}

// tryParseJSON 尝试解析 JSON 为 ArchitectTaskPlan
func (p *TaskParser) tryParseJSON(jsonStr string) (*ArchitectTaskPlan, error) {
	jsonStr = p.cleanJSON(jsonStr)

	var plan ArchitectTaskPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// cleanJSON 清理 AI 输出的非标准 JSON（注释、尾随逗号等）
func (p *TaskParser) cleanJSON(s string) string {
	s = strings.TrimSpace(s)

	// 移除单行注释 // ... （但不破坏字符串内的 URL 如 https://）
	// 策略：逐行处理，只移除不在引号内的 // 注释
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = removeLineComment(line)
	}
	s = strings.Join(lines, "\n")

	// 移除多行注释 /* ... */
	multiCommentRe := regexp.MustCompile(`(?s)/\*.*?\*/`)
	s = multiCommentRe.ReplaceAllString(s, "")

	// 移除尾随逗号（数组或对象最后一个元素后的逗号）
	trailingCommaRe := regexp.MustCompile(`,\s*([\]\}])`)
	s = trailingCommaRe.ReplaceAllString(s, "$1")

	return s
}

// removeLineComment 移除一行中不在引号内的 // 注释
func removeLineComment(line string) string {
	inString := false
	escape := false
	for i, ch := range line {
		if escape {
			escape = false
			continue
		}
		if ch == '\\' && inString {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if !inString && ch == '/' && i+1 < len(line) && line[i+1] == '/' {
			return line[:i]
		}
	}
	return line
}
