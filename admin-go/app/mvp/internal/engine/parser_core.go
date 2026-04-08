package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/model/do"
	"easymvp/utility/snowflake"
)

var (
	dryParseCodeBlockRe = regexp.MustCompile("(?s)```(?:json)?\\s*\\n?[\\{\\[][\\s\\S]*?[\\}\\]]\\s*```")
	dryParseInnerRe     = regexp.MustCompile("(?s)```(?:json)?\\s*\\n?([\\s\\S]*?)\\s*```")
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

	// 获取项目分类用于分类感知校验
	projectCategory := ""
	project, _ := g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).Fields("project_category").One()
	if !project.IsEmpty() {
		projectCategory = project["project_category"].String()
	}

	tasks := p.normalizeTasks(ctx, plan.Tasks, projectCategory)
	if len(tasks) == 0 {
		return 0, nil // 回复中没有任务清单，正常情况（还在讨论需求）
	}

	// 2. 查询项目的创建人和部门（任务继承项目的 created_by 和 dept_id）
	projectInfo, err := g.DB().Ctx(ctx).Model("mvp_project").
		Where("id", projectID).
		Fields("created_by, dept_id").
		One()
	if err != nil {
		return 0, fmt.Errorf("查询项目信息失败: %w", err)
	}
	createdBy := projectInfo["created_by"].Int64()
	deptID := projectInfo["dept_id"].Int64()

	// 3. 在事务中执行：清理旧数据 + 创建新任务（保证原子性）
	nameToID := make(map[string]int64, len(tasks))
	createdTasks := make([]ArchitectTask, 0, len(tasks))

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 3a. 获取所有旧任务ID
		oldTaskIDs, txErr := tx.Model("mvp_task").
			Where("project_id", projectID).
			Where("deleted_at IS NULL").
			Fields("id").
			Array()
		if txErr != nil {
			return fmt.Errorf("查询旧任务失败: %w", txErr)
		}
		if len(oldTaskIDs) > 0 {
			// 3b. 软删除任务日志
			_, _ = tx.Model("mvp_task_log").
				WhereIn("task_id", oldTaskIDs).
				Where("deleted_at IS NULL").
				Data(do.MvpTaskLog{
					DeletedAt: gtime.Now(),
					UpdatedAt: gtime.Now(),
				}).
				Update()
			// 3c. 删除任务依赖关系
			_, _ = tx.Model("mvp_task_dependency").
				WhereIn("task_id", oldTaskIDs).
				Delete()
			_, _ = tx.Model("mvp_task_dependency").
				WhereIn("depends_on_id", oldTaskIDs).
				Delete()
		}
		// 3d. 软删除所有任务
		_, txErr = tx.Model("mvp_task").
			Where("project_id", projectID).
			Where("deleted_at IS NULL").
			Data(do.MvpTask{
				DeletedAt: gtime.Now(),
				UpdatedAt: gtime.Now(),
			}).
			Update()
		if txErr != nil {
			return fmt.Errorf("清理旧任务失败: %w", txErr)
		}

		// 4. 第一遍：创建所有任务，建立 name→id 映射
		for i, t := range tasks {
			taskID := int64(snowflake.Generate())
			roleType := t.RoleType
			if roleType == "" {
				roleType = "implementer"
			}
			sort := t.Sort
			if sort == 0 {
				sort = i + 1
			}
			if t.BatchNo < 1 {
				t.BatchNo = 1
			}
			resourcesJSON := "[]"
			if len(t.AffectedResources) > 0 {
				b, _ := json.Marshal(t.AffectedResources)
				resourcesJSON = string(b)
			}
			dependsJSON := "[]"
			if len(t.DependsOn) > 0 {
				b, _ := json.Marshal(t.DependsOn)
				dependsJSON = string(b)
			}
			// 确定 task_kind：根据 roleType 判断
			taskKind := consts.TaskKindImplement
			if roleType == consts.RoleTypeArchitect {
				taskKind = ""
			}
			if roleType == consts.RoleTypeAuditor {
				taskKind = consts.TaskKindAudit
			}
			if roleType == consts.RoleTypeImplementer || roleType == "" {
				taskKind = consts.TaskKindImplement
			}
			_, txErr = tx.Model("mvp_task").Data(do.MvpTask{
				Id:                taskID,
				ProjectId:         projectID,
				ParentId:          0,
				Name:              t.Name,
				Description:       t.Description,
				RoleType:          roleType,
				RoleLevel:         t.RoleLevel,
				TaskKind:          taskKind,
				RootTaskId:        taskID,
				Status:            "draft",
				BatchNo:           t.BatchNo,
				Sort:              sort,
				AffectedResources: resourcesJSON,
				DependsOn:         dependsJSON,
				CreatedBy:         createdBy,
				DeptId:            deptID,
				CreatedAt:         gtime.Now(),
				UpdatedAt:         gtime.Now(),
			}).Insert()
			if txErr != nil {
				g.Log().Errorf(ctx, "创建任务失败 [%s]: %v", t.Name, txErr)
				continue
			}
			nameToID[t.Name] = taskID
			createdTasks = append(createdTasks, t)
		}

		// 5. 第二遍：建立父子关系
		for _, t := range createdTasks {
			if t.ParentName == "" {
				continue
			}
			parentID, ok := nameToID[t.ParentName]
			if !ok {
				continue
			}
			taskID, ok := nameToID[t.Name]
			if !ok {
				continue
			}
			_, _ = tx.Model("mvp_task").Where("id", taskID).Data(do.MvpTask{
				ParentId: parentID,
			}).Update()
		}

		// 6. 第三遍：建立依赖关系（mvp_task_dependency）
		for _, t := range createdTasks {
			if len(t.DependsOn) == 0 {
				continue
			}
			taskID, ok := nameToID[t.Name]
			if !ok {
				continue
			}
			for _, depName := range t.DependsOn {
				depID, ok := nameToID[depName]
				if !ok {
					g.Log().Warningf(ctx, "任务 [%s] 依赖 [%s] 未找到，跳过", t.Name, depName)
					continue
				}
				tx.Model("mvp_task_dependency").Insert(g.Map{
					"task_id":       taskID,
					"depends_on_id": depID,
				})
			}
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("任务创建事务失败: %w", err)
	}

	g.Log().Infof(ctx, "[TaskParser] 项目 %d 解析创建 %d 个任务（draft）", projectID, len(createdTasks))
	return len(createdTasks), nil
}

// ExtractAndNormalize 解析 AI 回复并返回标准化的任务列表（不写入数据库）。
// 供 V2 蓝图写入复用已有的解析和校验逻辑。
// 当正则提取失败时，使用 AI 做二次提取（适用于非标准 JSON 格式，如小说章节等）。
func (p *TaskParser) ExtractAndNormalize(ctx context.Context, aiReply, projectCategory string) ([]ArchitectTask, error) {
	plan, err := p.extractTaskPlan(aiReply)
	if err == nil && len(plan.Tasks) > 0 {
		return p.normalizeTasks(ctx, plan.Tasks, projectCategory), nil
	}

	// 正则提取失败，尝试 AI 二次提取（适用于跨消息 JSON 碎片、非标准格式等）
	if len(aiReply) > 100 {
		aiTasks, aiErr := p.aiExtractTasks(ctx, aiReply, projectCategory)
		if aiErr == nil && len(aiTasks) > 0 {
			g.Log().Infof(ctx, "[TaskParser] 正则提取失败，AI 二次提取成功: %d 个任务", len(aiTasks))
			return p.normalizeTasks(ctx, aiTasks, projectCategory), nil
		}
		if aiErr != nil {
			g.Log().Warningf(ctx, "[TaskParser] AI 二次提取也失败: %v", aiErr)
		}
	}

	return nil, nil
}

// FastExtract 快速正则提取（不走 AI），返回 nil 表示提取失败。
func (p *TaskParser) FastExtract(aiReply string) (*ArchitectTaskPlan, error) {
	return p.extractTaskPlan(aiReply)
}

// NormalizeTasks 导出的标准化方法。
func (p *TaskParser) NormalizeTasks(ctx context.Context, tasks []ArchitectTask, projectCategory string) []ArchitectTask {
	return p.normalizeTasks(ctx, tasks, projectCategory)
}

// AIExtractTasks 导出的 AI 二次提取方法。
func (p *TaskParser) AIExtractTasks(ctx context.Context, aiReply, projectCategory string) ([]ArchitectTask, error) {
	return p.aiExtractTasks(ctx, aiReply, projectCategory)
}

// DryParseTaskCount 仅解析AI回复，返回任务数量（不写入数据库）。
// 使用与实际创建相同的提取逻辑（正则 + AI 二次提取），确保 dryRun 和实际创建结果一致。
func (p *TaskParser) DryParseTaskCount(aiReply string) int {
	ctx := context.Background()
	g.Log().Infof(ctx, "[DryParseTaskCount] aiReply length=%d runes", len([]rune(aiReply)))

	// 先用正则快速提取
	plan, err := p.extractTaskPlan(aiReply)
	if err == nil && len(plan.Tasks) > 0 {
		normalized := p.normalizeTasks(ctx, plan.Tasks, "")
		g.Log().Infof(ctx, "[DryParseTaskCount] extractTaskPlan 成功: raw=%d normalized=%d", len(plan.Tasks), len(normalized))
		return len(normalized)
	}
	g.Log().Infof(ctx, "[DryParseTaskCount] extractTaskPlan 失败: err=%v tasks=%d", err, func() int { if plan != nil { return len(plan.Tasks) }; return 0 }())

	// 正则失败，检测是否有 JSON 代码块（意味着实际创建时 AI 二次提取大概率能成功）
	if matches := dryParseCodeBlockRe.FindAllString(aiReply, -1); len(matches) > 0 {
		// 尝试从代码块中解析出实际任务数（而非代码块数）
		total := 0
		for _, m := range matches {
			// 去掉 ``` 标记
			inner := dryParseInnerRe.FindStringSubmatch(m)
			if len(inner) < 2 {
				continue
			}
			if tryPlan, tryErr := p.tryParseJSON(inner[1]); tryErr == nil && len(tryPlan.Tasks) > 0 {
				total += len(tryPlan.Tasks)
			}
		}
		if total > 0 {
			return total
		}
		// 代码块存在但解析不出来，返回 -1 表示"有内容但需要 AI 提取"
		// 前端可根据此值提示用户
		return -1
	}

	// 检测是否有足够长的文本内容（纯文本描述场景，AI 可能能提取）
	if len([]rune(aiReply)) > 200 {
		// 有足够内容但无 JSON 代码块，返回 -1 表示需要 AI 提取
		return -1
	}

	return 0
}

// ConfirmDraftTasks 确认草稿任务：draft → pending
// 全量确认：所有 draft 必须全部成功，任何失败则回滚已转换的任务
func (p *TaskParser) ConfirmDraftTasks(ctx context.Context, projectID int64) (int, error) {
	taskIDs, err := g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "draft").
		Where("deleted_at IS NULL").
		Fields("id").
		Array()
	if err != nil {
		return 0, err
	}
	if len(taskIDs) == 0 {
		return 0, nil
	}

	// 逐条 CAS 转换，记录已成功的 ID 用于失败时回滚
	var confirmedIDs []int64
	for _, idVal := range taskIDs {
		rows, err := updateTaskStatus(ctx, idVal.Int64(), "draft", "pending", nil)
		if err != nil || rows == 0 {
			// 部分失败：回滚已转换的任务
			g.Log().Errorf(ctx, "[ConfirmDraftTasks] task=%d draft→pending 失败，回滚已确认的 %d 个任务: %v",
				idVal.Int64(), len(confirmedIDs), err)
			for _, rollbackID := range confirmedIDs {
				if _, rbErr := updateTaskStatus(ctx, rollbackID, "pending", "draft", nil); rbErr != nil {
					g.Log().Errorf(ctx, "[ConfirmDraftTasks] 回滚 task=%d pending→draft 失败: %v", rollbackID, rbErr)
				}
			}
			errMsg := "CAS 未命中"
			if err != nil {
				errMsg = err.Error()
			}
			return 0, fmt.Errorf("确认任务 %d 失败: %s，已回滚 %d 个已确认任务", idVal.Int64(), errMsg, len(confirmedIDs))
		}
		confirmedIDs = append(confirmedIDs, idVal.Int64())
	}
	return len(confirmedIDs), nil
}

// GetDraftCount 获取项目草稿任务数量
func (p *TaskParser) GetDraftCount(ctx context.Context, projectID int64) int {
	count, _ := g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "draft").
		Where("deleted_at IS NULL").
		Count()
	return count
}
