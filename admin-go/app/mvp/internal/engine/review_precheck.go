package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/resourcepath"
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

// RunReview 执行方案审核流程（三步）
// 返回审核结果。调用者根据结果决定是进入 running 还是回退 designing
func RunReview(ctx context.Context, projectID int64) (*ReviewResult, error) {
	// 获取项目信息
	project, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).WhereNull("deleted_at").Fields("id, work_dir, project_category").One()
	if err != nil {
		return nil, fmt.Errorf("查询项目信息失败: %w", err)
	}
	if project.IsEmpty() {
		return nil, fmt.Errorf("项目 %d 不存在", projectID)
	}
	workDir := project["work_dir"].String()
	projectCategory := project["project_category"].String()

	// 获取所有 draft 任务
	tasks, err := g.DB().Ctx(ctx).Model("mvp_task").
		Fields("id, name, description, batch_no, role_type, role_level, execution_mode, affected_resources, depends_on").
		Where("project_id", projectID).
		Where("status", "draft").
		WhereNull("deleted_at").
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
//
// TODO(PR-10): 当 precheck.use_rule_engine=true 时，使用 workflow/domain/rule.Registry 替代下方硬编码规则。
// 单任务规则（名称非空、描述质量、资源格式、乱码检测、目录占位）已迁移到 BuiltinChecker；
// 跨任务规则（depends_on 引用有效性、batch_no 一致性、资源冲突、role_level 覆盖）保留在此处。
// 待 CEL 引入后逐步切换，使用 DefaultRegistry.Get("precheck").Check(ctx, &CheckContext{...})。
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

	resolvedRoleLevels := make(map[string]string)

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
		if family == CategoryFamilyCoding {
			for _, resource := range resourcepath.FindCodingDirectoryPlaceholders(resources) {
				result.Errors = append(result.Errors, ReviewIssue{
					TaskName: name, Severity: "error",
					Message: fmt.Sprintf("affected_resources 必须写明确文件路径，不能使用目录占位: %s", resource),
				})
			}
		}

		// 4. 编码类项目允许创建新文件；不存在不再作为审核阻塞项。
		_ = family
		_ = workDir

		// 5. depends_on 有效性
		var dependsOn []string
		depJSON := t["depends_on"].String()
		if depJSON != "" && depJSON != "[]" && depJSON != "null" {
			if jsonErr := json.Unmarshal([]byte(depJSON), &dependsOn); jsonErr != nil {
				result.Warnings = append(result.Warnings, ReviewIssue{
					TaskName: name, Severity: "warning",
					Message: fmt.Sprintf("depends_on JSON 格式异常: %v", jsonErr),
				})
			}
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

		// 8. role_level 覆盖检查：lite 不可用时自动升档到 pro，再没有就升到 max。
		roleType := t["role_type"].String()
		roleLevel := t["role_level"].String()
		if roleType != "" && roleLevel != "" {
			roleKey := roleType + "/" + roleLevel
			resolvedLevel, checked := resolvedRoleLevels[roleKey]
			if !checked {
				roleRecord, roleErr := ResolveProjectRoleByLevel(ctx, projectID, roleType, roleLevel)
				if roleErr != nil || roleRecord == nil {
					resolvedRoleLevels[roleKey] = ""
				} else {
					resolvedRoleLevels[roleKey] = roleRecord["role_level"].String()
				}
				resolvedLevel = resolvedRoleLevels[roleKey]
			}
			if resolvedLevel == "" {
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

var garbageRe = regexp.MustCompile(`[\x00-\x1f]|[^\x00-\x7f]{10,}|[!@#$%^&*()+=\[\]{}|\\;':",<>?]{3,}`)

// containsGarbage 检测路径是否包含疑似乱码
func containsGarbage(path string) bool {
	return garbageRe.MatchString(path)
}

// getProjectCategory 获取项目分类名称
func getProjectCategory(ctx context.Context, projectID int64) string {
	val, err := g.DB().Ctx(ctx).Model("mvp_project").
		Where("id", projectID).WhereNull("deleted_at").
		Value("project_category")
	if err != nil || val.IsEmpty() {
		return "软件开发"
	}
	return val.String()
}
