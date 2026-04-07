package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "easymvp/app/mvp/internal/packed"
	"easymvp/app/mvp/internal/workflow/presetutil"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
)

type syncStats struct {
	Scanned int
	Changed int
	Skipped int
}

func main() {
	var apply bool
	flag.BoolVar(&apply, "apply", false, "write changes to database")
	flag.Parse()

	ctx := gctx.New()

	rolePresetPlan, err := buildRolePresetPlan(ctx)
	if err != nil {
		fatalf("构建角色预设同步计划失败: %v", err)
	}
	projectRolePlan, err := buildProjectRolePlan(ctx)
	if err != nil {
		fatalf("构建项目角色同步计划失败: %v", err)
	}

	printPlan(rolePresetPlan, projectRolePlan, apply)
	if !apply {
		return
	}

	if err := applyPlan(ctx, rolePresetPlan, projectRolePlan); err != nil {
		fatalf("回写数据库失败: %v", err)
	}

	fmt.Println("同步完成。")
}

type rolePresetUpdate struct {
	ID         int64
	OldPrompt  string
	NewPrompt  string
	Category   string
	RoleType   string
	RoleLevel  string
}

type projectRoleUpdate struct {
	ID         int64
	ProjectID  int64
	OldPrompt  string
	NewPrompt  string
	Category   string
	RoleType   string
	RoleLevel  string
}

func buildRolePresetPlan(ctx context.Context) ([]rolePresetUpdate, error) {
	rows, err := g.DB().Model("mvp_role_preset").Ctx(ctx).
		Fields("id, project_category, role_type, role_level, system_prompt").
		Where("status", 1).
		Where("deleted_at IS NULL").
		OrderAsc("id").
		All()
	if err != nil {
		return nil, err
	}

	updates := make([]rolePresetUpdate, 0, len(rows))
	for _, row := range rows {
		category := row["project_category"].String()
		roleType := row["role_type"].String()
		roleLevel := row["role_level"].String()
		newPrompt := strings.TrimSpace(presetutil.BuildDefaultRolePrompt(category, roleType, roleLevel))
		if newPrompt == "" {
			continue
		}
		oldPrompt := strings.TrimSpace(row["system_prompt"].String())
		if oldPrompt == newPrompt {
			continue
		}
		updates = append(updates, rolePresetUpdate{
			ID:        row["id"].Int64(),
			OldPrompt: oldPrompt,
			NewPrompt: newPrompt,
			Category:  category,
			RoleType:  roleType,
			RoleLevel: roleLevel,
		})
	}
	return updates, nil
}

func buildProjectRolePlan(ctx context.Context) ([]projectRoleUpdate, error) {
	rows, err := g.DB().Model("mvp_project_role pr").Ctx(ctx).
		LeftJoin("mvp_project p", "p.id = pr.project_id").
		LeftJoin("ai_model m", "m.id = pr.model_id AND m.deleted_at IS NULL").
		Fields("pr.id, pr.project_id, pr.role_type, pr.role_level, pr.system_prompt, p.name AS project_name, p.description AS project_desc, p.category_code, p.project_category, m.role_prompt").
		Where("pr.status", 1).
		Where("pr.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").
		OrderAsc("pr.id").
		All()
	if err != nil {
		return nil, err
	}

	updates := make([]projectRoleUpdate, 0, len(rows))
	for _, row := range rows {
		category := strings.TrimSpace(row["category_code"].String())
		if category == "" {
			category = strings.TrimSpace(row["project_category"].String())
		}

		roleType := row["role_type"].String()
		roleLevel := row["role_level"].String()
		modelPrompt := row["role_prompt"].String()
		newPrompt := strings.TrimSpace(presetutil.BuildRoleSystemPrompt(category, roleType, roleLevel, "", modelPrompt))
		if roleType == "architect" {
			newPrompt = strings.TrimSpace(presetutil.BuildArchitectSystemPrompt(
				row["project_name"].String(),
				row["project_desc"].String(),
				category,
				newPrompt,
			))
		}
		if newPrompt == "" {
			continue
		}

		oldPrompt := strings.TrimSpace(row["system_prompt"].String())
		if oldPrompt == newPrompt {
			continue
		}
		updates = append(updates, projectRoleUpdate{
			ID:        row["id"].Int64(),
			ProjectID: row["project_id"].Int64(),
			OldPrompt: oldPrompt,
			NewPrompt: newPrompt,
			Category:  category,
			RoleType:  roleType,
			RoleLevel: roleLevel,
		})
	}
	return updates, nil
}

func printPlan(rolePresetPlan []rolePresetUpdate, projectRolePlan []projectRoleUpdate, apply bool) {
	mode := "预览"
	if apply {
		mode = "执行"
	}
	fmt.Printf("[%s] 角色预设待同步: %d 条\n", mode, len(rolePresetPlan))
	for _, item := range sampleRolePresetUpdates(rolePresetPlan, 5) {
		fmt.Printf("  - role_preset id=%d category=%s role=%s/%s\n", item.ID, item.Category, item.RoleType, item.RoleLevel)
	}

	fmt.Printf("[%s] 项目角色待同步: %d 条\n", mode, len(projectRolePlan))
	for _, item := range sampleProjectRoleUpdates(projectRolePlan, 5) {
		fmt.Printf("  - project_role id=%d project=%d category=%s role=%s/%s\n", item.ID, item.ProjectID, item.Category, item.RoleType, item.RoleLevel)
	}

	if !apply {
		fmt.Println("当前为预览模式；如需回写数据库，请追加 --apply。")
	}
}

func sampleRolePresetUpdates(items []rolePresetUpdate, limit int) []rolePresetUpdate {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func sampleProjectRoleUpdates(items []projectRoleUpdate, limit int) []projectRoleUpdate {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func applyPlan(ctx context.Context, rolePresetPlan []rolePresetUpdate, projectRolePlan []projectRoleUpdate) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		now := gtime.Now()

		for _, item := range rolePresetPlan {
			if _, err := tx.Model("mvp_role_preset").Ctx(ctx).
				Where("id", item.ID).
				Update(g.Map{
					"system_prompt": item.NewPrompt,
					"updated_at":    now,
				}); err != nil {
				return fmt.Errorf("更新 role_preset[%d] 失败: %w", item.ID, err)
			}
		}

		for _, item := range projectRolePlan {
			if _, err := tx.Model("mvp_project_role").Ctx(ctx).
				Where("id", item.ID).
				Update(g.Map{
					"system_prompt": item.NewPrompt,
					"updated_at":    now,
				}); err != nil {
				return fmt.Errorf("更新 project_role[%d] 失败: %w", item.ID, err)
			}
		}
		return nil
	})
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
