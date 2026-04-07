package engine

// review_bridge.go 导出旧审核逻辑的函数，供新 workflow/stage/review 包调用。
// 避免在 review.go 中大量改动，保持旧引擎 legacy 路径不变。

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// RunAuditorReviewForBlueprints 为蓝图列表执行审计员 AI 审核。
// 复用旧 auditorReview 的核心逻辑，输入改为蓝图记录。
func RunAuditorReviewForBlueprints(ctx context.Context, projectID int64, blueprints gdb.Result) (*AuditorReviewResult, error) {
	modelInfo, err := getReviewRoleModel(ctx, projectID, "auditor")
	if err != nil {
		return nil, fmt.Errorf("获取审计员模型失败: %w", err)
	}

	// 获取项目信息，传给审计员做整体理解
	project, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("project_category, name, description").
		Where("id", projectID).WhereNull("deleted_at").One()
	projectCategory := "软件开发"
	projectName := ""
	projectDesc := ""
	if !project.IsEmpty() {
		projectCategory = project["project_category"].String()
		projectName = project["name"].String()
		projectDesc = project["description"].String()
	}

	tasks := blueprintsToTaskRecords(blueprints)
	return doAuditorReview(ctx, modelInfo, tasks, projectCategory, projectName, projectDesc)
}

// RunCoordinatorOptimizeForBlueprints 为蓝图列表执行协调员优化。
func RunCoordinatorOptimizeForBlueprints(ctx context.Context, projectID int64, blueprints gdb.Result) (*CoordinatorOptResult, error) {
	modelInfo, err := getReviewRoleModel(ctx, projectID, "coordinator")
	if err != nil {
		return nil, fmt.Errorf("获取协调员模型失败: %w", err)
	}

	tasks := blueprintsToTaskRecords(blueprints)
	return doCoordinatorOptimize(ctx, modelInfo, tasks)
}

// ApplyCoordinatorOptimizationsToBlueprints 应用协调员优化到蓝图的 batch_no。
func ApplyCoordinatorOptimizationsToBlueprints(ctx context.Context, planVersionID int64, blueprints gdb.Result, opt *CoordinatorOptResult) {
	if opt == nil || len(opt.OptimizedBatches) == 0 {
		return
	}

	for _, bp := range blueprints {
		name := bp["name"].String()
		if batch, ok := opt.OptimizedBatches[name]; ok {
			if batch.BatchNo > 0 && batch.BatchNo != bp["batch_no"].Int() {
				_, _ = g.DB().Model("mvp_task_blueprint").Ctx(ctx).
					Where("id", bp["id"].Int64()).
					Update(g.Map{"batch_no": batch.BatchNo, "updated_at": gtime.Now()})
			}
		}
	}
}

// NotifyProjectArchitectConversation 导出的通知函数。
func NotifyProjectArchitectConversation(ctx context.Context, projectID int64, content string) {
	notifyProjectArchitectConversation(ctx, projectID, content)
}

// CheckResourceExists 检查资源路径是否存在（编码类项目用）。
func CheckResourceExists(workDir, res string, callback func(severity, msg string)) {
	absPath := res
	if !filepath.IsAbs(res) {
		absPath = filepath.Join(workDir, res)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		callback("warning", fmt.Sprintf("文件/目录不存在: %s（可能是新建文件，请确认）", res))
	}
}

// blueprintsToTaskRecords 将蓝图记录转为兼容旧审核函数的 gdb.Result。
// 字段映射：name→name, description→description, role_type→role_type, role_level→role_level,
// batch_no→batch_no, affected_resources→affected_resources, depends_on_blueprint_ids→depends_on
func blueprintsToTaskRecords(blueprints gdb.Result) gdb.Result {
	result := make(gdb.Result, 0, len(blueprints))
	for _, bp := range blueprints {
		record := gdb.Record{
			"name":               bp["name"],
			"description":        bp["description"],
			"role_type":          bp["role_type"],
			"role_level":         bp["role_level"],
			"batch_no":           bp["batch_no"],
			"affected_resources": bp["affected_resources"],
			"depends_on":         bp["depends_on_blueprint_ids"],
			"id":                 bp["id"],
		}
		result = append(result, record)
	}
	return result
}
