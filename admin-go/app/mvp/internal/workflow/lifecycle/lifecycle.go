// Package lifecycle 提供项目生命周期初始化能力（CreateProject 等）。
// 将 DB 操作封装在事务中，文件系统副作用（EnsureWorkDir）在事务外执行。
package lifecycle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/domain/fsm"
	"easymvp/app/mvp/internal/workflow/presetutil"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// ──────────────────────────────────────────────
// 工作目录相关（复制自 engine/path_validation.go，避免循环引用）
// ──────────────────────────────────────────────

var systemPathBlacklist = []string{
	"/", "/etc", "/bin", "/sbin", "/usr", "/lib", "/lib64",
	"/proc", "/sys", "/dev", "/run", "/tmp", "/var",
	"/boot", "/root", "/home",
}

// ensureWorkDir 校验工作目录，必要时自动创建。
func ensureWorkDir(path string) (string, bool, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || path == "." {
		return "", false, fmt.Errorf("代码工作目录不能为空")
	}
	for _, blocked := range systemPathBlacklist {
		if path == blocked || strings.HasPrefix(path, blocked+string(filepath.Separator)) {
			return "", false, fmt.Errorf("禁止使用系统目录作为工作目录: %s", path)
		}
	}
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return "", false, fmt.Errorf("代码工作目录不是目录: %s", path)
		}
		if realPath, evalErr := filepath.EvalSymlinks(path); evalErr == nil {
			realPath = filepath.Clean(realPath)
			for _, blocked := range systemPathBlacklist {
				if realPath == blocked || strings.HasPrefix(realPath, blocked+string(filepath.Separator)) {
					return "", false, fmt.Errorf("代码工作目录通过符号链接指向系统目录: %s → %s", path, realPath)
				}
			}
			path = realPath
		}
		return path, false, nil
	}
	if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("代码工作目录不可用: %s", path)
	}
	if err = os.MkdirAll(path, 0o755); err != nil {
		return "", false, fmt.Errorf("代码工作目录创建失败: %s", path)
	}
	return path, true, nil
}

// categoryFamilyCoding 用于判断是否为编码类项目。
const categoryFamilyCoding = "coding"

// categoryFamilyFallback 硬编码兜底映射（仅在 DB 不可用时使用）。
var categoryFamilyFallback = map[string]string{
	"software_dev": categoryFamilyCoding, "game_dev": categoryFamilyCoding,
	"software_dev_code": categoryFamilyCoding,
	"软件开发": categoryFamilyCoding, "游戏开发": categoryFamilyCoding,
}

// isCodingCategory 判断分类是否为编码类（编码类不自动生成工作目录）。
func isCodingCategory(projectCategory string) bool {
	if f, ok := categoryFamilyFallback[projectCategory]; ok {
		return f == categoryFamilyCoding
	}
	// 默认视为编码类，避免错误自动生成工作目录
	return true
}

// generateWorkDir 为非编码类项目自动生成工作目录。
func generateWorkDir(projectCategory string, projectID int64) string {
	if isCodingCategory(projectCategory) {
		return "" // 编码类不自动生成
	}
	return fmt.Sprintf("/www/wwwroot/project/easymvp/workspace/non-coding/%d", projectID)
}

// ──────────────────────────────────────────────
// 分类 code 解析（直接查 DB，避免 import engine 包）
// ──────────────────────────────────────────────

func resolveCategoryCode(ctx context.Context, displayName string) string {
	if displayName == "" {
		return ""
	}
	rec, err := g.DB().Ctx(ctx).Model("mvp_project_category").
		Where("display_name", displayName).
		Fields("category_code").
		One()
	if err != nil || rec.IsEmpty() {
		return ""
	}
	return rec["category_code"].String()
}

// ──────────────────────────────────────────────
// CreateProject
// ──────────────────────────────────────────────

// CreateProject 创建项目并初始化架构师对话，将所有 DB 操作包在事务中。
// selectedPresetIDs 为用户显式选择的项目角色预设；为空时不生成项目角色配置，
// 运行时直接回退到分类默认预设。
func CreateProject(
	ctx context.Context,
	name, projectCategory, description, workDir string,
	architectModelID int64,
	userID int64,
	deptID int64,
	selectedPresetIDs []int64,
	engineVersion ...string,
) (int64, int64, error) {
	// 默认分类
	if projectCategory == "" {
		projectCategory = "软件开发"
	}

	projectID := int64(snowflake.Generate())

	// 解析 category_code（通过 DB 直接查询）
	categoryCode := resolveCategoryCode(ctx, projectCategory)

	// ── 文件系统副作用（事务外执行，避免事务回滚后目录残留） ──
	if workDir == "" {
		workDir = generateWorkDir(projectCategory, projectID)
	}
	if workDir != "" {
		var err error
		workDir, _, err = ensureWorkDir(workDir)
		if err != nil {
			return 0, 0, err
		}
	}

	// 引擎版本
	ev := "workflow_v2"
	if len(engineVersion) > 0 && engineVersion[0] == "legacy" {
		ev = "legacy"
	}

	// 事务外读取预设（只读，不需要在事务内）
	var presets gdb.Result
	var presetsErr error
	if len(selectedPresetIDs) > 0 {
		presets, presetsErr = repo.ListRolePresets(ctx, repo.RolePresetQuery{
			IDs:             selectedPresetIDs,
			CategoryCode:    categoryCode,
			ProjectCategory: projectCategory,
		})
		if presetsErr != nil {
			return 0, 0, fmt.Errorf("读取角色预设失败: %w", presetsErr)
		}
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

	// 批量查询模型 role_prompt（事务外，只读）
	modelIDs := make([]int64, 0, len(presets)+1)
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
		models, modErr := g.DB().Ctx(ctx).Model("ai_model").
			Fields("id, role_prompt").
			WhereIn("id", modelIDs).
			WhereNull("deleted_at").
			All()
		if modErr != nil {
			g.Log().Warningf(ctx, "[CreateProject] 批量查询模型失败: err=%v", modErr)
		}
		for _, m := range models {
			modelPromptMap[m["id"].Int64()] = m["role_prompt"].String()
		}
	}

	// 架构师预设（事务外，只读）
	var architectPresetForFallback gdb.Record
	if len(presets) == 0 && architectModelID > 0 {
		ap, apErr := repo.GetRolePreset(ctx, repo.RolePresetQuery{
			CategoryCode:    categoryCode,
			ProjectCategory: projectCategory,
			RoleType:        "architect",
			DefaultOnly:     true,
		})
		if apErr == nil && ap != nil {
			architectPresetForFallback = ap
		}
	}

	// ── 事务：所有写操作 ──
	var convID int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 创建项目
		_, err := tx.Ctx(ctx).Model("mvp_project").Insert(g.Map{
			"id":                 projectID,
			"name":               name,
			"project_category":   projectCategory,
			"category_code":      categoryCode,
			"description":        description,
			"status":             fsm.WorkflowInitial().String(),
			"engine_version":     ev,
			"work_dir":           workDir,
			"architect_model_id": projectArchitectModelID,
			"created_by":         userID,
			"dept_id":            deptID,
			"created_at":         gtime.Now(),
			"updated_at":         gtime.Now(),
		})
		if err != nil {
			return fmt.Errorf("创建项目失败: %w", err)
		}

		// 2. 为用户显式选择的预设创建项目角色配置
		for _, p := range presets {
			roleType := p["role_type"].String()
			modelID := p["model_id"].Int64()

			if roleType == "architect" {
				modelID = architectModelID
				if modelID == 0 {
					modelID = p["model_id"].Int64()
				}
			}

			systemPrompt := presetutil.BuildRoleSystemPrompt(
				categoryCode, roleType,
				p["role_level"].String(),
				p["system_prompt"].String(),
				modelPromptMap[modelID],
			)
			if roleType == "architect" {
				systemPrompt = presetutil.BuildArchitectSystemPrompt(name, description, categoryCode, systemPrompt)
			}

			executionMode := p["execution_mode"].String()
			if executionMode == "" {
				executionMode = "chat"
			}

			_, err = tx.Ctx(ctx).Model("mvp_project_role").Insert(g.Map{
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
				return fmt.Errorf("创建角色配置(%s)失败: %w", roleType, err)
			}
		}

		// 3. 仅当用户未显式选择预设但指定了架构师模型时，创建架构师角色覆盖
		if len(presets) == 0 && architectModelID > 0 {
			roleLevel := "max"
			executionMode := "chat"
			if architectPresetForFallback != nil {
				if architectPresetForFallback["role_level"].String() != "" {
					roleLevel = architectPresetForFallback["role_level"].String()
				}
				if architectPresetForFallback["execution_mode"].String() != "" {
					executionMode = architectPresetForFallback["execution_mode"].String()
				}
			}

			systemPrompt := presetutil.BuildRoleSystemPrompt(
				categoryCode, "architect", roleLevel, "", modelPromptMap[architectModelID],
			)
			systemPrompt = presetutil.BuildArchitectSystemPrompt(name, description, categoryCode, systemPrompt)

			_, err = tx.Ctx(ctx).Model("mvp_project_role").Insert(g.Map{
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
				return fmt.Errorf("创建架构师配置失败: %w", err)
			}
		}

		// 4. 创建架构师对话（项目级对话）
		convID = int64(snowflake.Generate())
		_, err = tx.Ctx(ctx).Model("mvp_conversation").Insert(g.Map{
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
			return fmt.Errorf("创建架构师对话失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	return projectID, convID, nil
}
