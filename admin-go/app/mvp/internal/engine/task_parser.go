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

	// 2. 清理该项目已有的 draft 任务（架构师可能多次输出方案，以最新为准）
	_, err = g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "draft").
		Update(g.Map{
			"deleted_at": gtime.Now(),
			"updated_at": gtime.Now(),
		})
	if err != nil {
		return 0, fmt.Errorf("清理旧草稿失败: %w", err)
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
		var tasks []ArchitectTask
		if err := json.Unmarshal([]byte(m), &tasks); err == nil && len(tasks) > 0 {
			return &ArchitectTaskPlan{Tasks: tasks}, nil
		}
	}

	// 没找到有效的任务清单
	return &ArchitectTaskPlan{}, nil
}

// tryParseJSON 尝试解析 JSON 为 ArchitectTaskPlan
func (p *TaskParser) tryParseJSON(jsonStr string) (*ArchitectTaskPlan, error) {
	// 清理可能的注释和尾随逗号
	jsonStr = strings.TrimSpace(jsonStr)

	var plan ArchitectTaskPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
