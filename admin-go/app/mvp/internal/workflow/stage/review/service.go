// Package review 驱动审核阶段：precheck → auditor → coordinator → summary。
package review

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// StageCompleter 阶段完成回调接口（避免循环依赖 orchestrator）。
type StageCompleter interface {
	CompleteStage(ctx context.Context, stageRunID int64) error
	FailStage(ctx context.Context, stageRunID int64, reason string) error
}

// ExecuteTriggerFn 审核通过后触发执行阶段的回调。
type ExecuteTriggerFn func(ctx context.Context, workflowRunID, planVersionID int64) error

// Service 审核阶段服务。
type Service struct {
	stageCompleter   StageCompleter
	issueRepo        *repo.ReviewIssueRepo
	designRollbackFn DesignRollbackFn
	executeTriggerFn ExecuteTriggerFn
}

// NewService 创建审核阶段服务。
func NewService(sc StageCompleter, ir *repo.ReviewIssueRepo) *Service {
	return &Service{stageCompleter: sc, issueRepo: ir}
}

// stage_task 类型常量
const (
	TaskTypePrecheck            = "precheck"
	TaskTypeAuditorReview       = "auditor_review"
	TaskTypeCoordinatorOptimize = "coordinator_optimize"
	TaskTypeReviewSummary       = "review_summary"
)

// RunReview 执行完整审核流程。
// 步骤：1. 系统预检 → 2. 审计员 AI 审核 → 3. 协调员优化 → 4. 汇总结论。
// 审核问题写入 mvp_review_issue，结论更新到 plan_version。
func (s *Service) RunReview(ctx context.Context, stageRunID int64, planVersionID int64) error {
	g.Log().Infof(ctx, "[ReviewStage] RunReview start stageRunID=%d planVersionID=%d", stageRunID, planVersionID)

	// 查 stage_run 获取 workflow_run_id
	stageRun, err := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).One()
	if err != nil || stageRun.IsEmpty() {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := stageRun["workflow_run_id"].Int64()

	// 查 workflow_run 获取 project_id
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}
	projectID := wfRun["project_id"].Int64()

	// 获取项目信息
	project, err := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("项目(%d) 不存在", projectID)
	}

	// 获取蓝图列表（代替旧的 draft tasks）
	blueprints, err := g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		Where("plan_version_id", planVersionID).
		Where("blueprint_status", consts.BlueprintStatusConfirmed).
		WhereNull("deleted_at").
		OrderAsc("sort").
		All()
	if err != nil {
		return fmt.Errorf("查询蓝图失败: %w", err)
	}
	if len(blueprints) == 0 {
		return fmt.Errorf("plan_version(%d) 没有已确认的蓝图", planVersionID)
	}

	// ======== Step 1: 系统预检 ========
	precheckResult, precheckErr := s.runPrecheck(ctx, stageRunID, workflowRunID, planVersionID, projectID, project, blueprints)
	if precheckErr != nil {
		return precheckErr
	}

	// 如果预检有 error 级别问题，审核不通过
	if precheckResult.HasErrors {
		return s.concludeReview(ctx, stageRunID, planVersionID, projectID, false, "系统预检发现错误")
	}

	// ======== Step 2: 审计员 AI 审核 ========
	auditorPassed, auditorErr := s.runAuditorReview(ctx, stageRunID, workflowRunID, planVersionID, projectID, blueprints)
	if auditorErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] 审计员审核失败/超时，跳过: %v", auditorErr)
		// 审计员失败不阻塞，记录 warning
		s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, 0, "warning", "auditor_skip", "system", "", fmt.Sprintf("审计员审核跳过: %v", auditorErr))
	} else if !auditorPassed {
		return s.concludeReview(ctx, stageRunID, planVersionID, projectID, false, "审计员审核未通过")
	}

	// ======== Step 3: 协调员优化 ========
	_ = s.runCoordinatorOptimize(ctx, stageRunID, workflowRunID, planVersionID, projectID, blueprints)

	// ======== Step 4: 汇总结论 ========
	return s.concludeReview(ctx, stageRunID, planVersionID, projectID, true, "审核通过")
}

// precheckResult 预检结果。
type precheckResult struct {
	HasErrors bool
}

// runPrecheck 系统预检（零 AI 消耗）。
func (s *Service) runPrecheck(ctx context.Context, stageRunID, workflowRunID, planVersionID, projectID int64, project gdb.Record, blueprints gdb.Result) (*precheckResult, error) {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypePrecheck, "system")
	now := time.Now()

	// 标记开始
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": "running", "started_at": now, "updated_at": now})

	workDir := project["work_dir"].String()
	projectCategory := project["project_category"].String()
	family := engine.GetCategoryFamily(projectCategory)

	result := &precheckResult{}

	// 构建名称集合
	bpNames := make(map[string]bool, len(blueprints))
	nameToBatch := make(map[string]int, len(blueprints))
	for _, bp := range blueprints {
		bpNames[bp["name"].String()] = true
		nameToBatch[bp["name"].String()] = bp["batch_no"].Int()
	}

	// 预加载角色配置
	availableRoles := make(map[string]bool)
	roleConfigs, _ := g.DB().Model("mvp_project_role").Ctx(ctx).
		Where("project_id", projectID).Where("status", 1).WhereNull("deleted_at").
		Fields("role_type, role_level").All()
	for _, rc := range roleConfigs {
		availableRoles[rc["role_type"].String()+"/"+rc["role_level"].String()] = true
	}

	batchResources := make(map[int]map[string]string)

	for _, bp := range blueprints {
		name := bp["name"].String()
		desc := bp["description"].String()
		batchNo := bp["batch_no"].Int()
		bpID := bp["id"].Int64()

		// 名称非空
		if name == "" {
			s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "error", "empty_name", "precheck", name, "蓝图名称为空")
			result.HasErrors = true
			continue
		}

		// 描述质量
		if len([]rune(desc)) < 10 {
			s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "error", "short_desc", "precheck", name,
				fmt.Sprintf("蓝图描述过短（%d字），需要至少10字的有效描述", len([]rune(desc))))
			result.HasErrors = true
		}

		// affected_resources
		var resources []string
		resJSON := bp["affected_resources"].String()
		if resJSON != "" && resJSON != "[]" && resJSON != "null" {
			if err := json.Unmarshal([]byte(resJSON), &resources); err != nil {
				s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "error", "bad_resources", "precheck", name,
					"affected_resources 格式非法: "+err.Error())
				result.HasErrors = true
			}
		}

		// 编码类检查文件存在性
		if family == engine.CategoryFamilyCoding && workDir != "" {
			for _, res := range resources {
				engine.CheckResourceExists(workDir, res, func(severity, msg string) {
					s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, severity, "resource_missing", "precheck", name, msg)
				})
			}
		}

		// depends_on 有效性（蓝图的依赖是 blueprintID 数组，需要转换）
		var depIDs []int64
		depJSON := bp["depends_on_blueprint_ids"].String()
		if depJSON != "" && depJSON != "[]" && depJSON != "null" {
			json.Unmarshal([]byte(depJSON), &depIDs)
		}

		// 资源冲突
		if _, ok := batchResources[batchNo]; !ok {
			batchResources[batchNo] = make(map[string]string)
		}
		for _, res := range resources {
			if existingTask, conflict := batchResources[batchNo][res]; conflict {
				s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "error", "resource_conflict", "precheck", name,
					fmt.Sprintf("资源冲突: 同批次(%d)中 [%s] 和 [%s] 都修改 %s", batchNo, existingTask, name, res))
				result.HasErrors = true
			}
			batchResources[batchNo][res] = name
		}

		// 角色配置检查
		roleType := bp["role_type"].String()
		roleLevel := bp["role_level"].String()
		if roleType != "" && roleLevel != "" && !availableRoles[roleType+"/"+roleLevel] {
			s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "warning", "missing_role", "precheck", name,
				fmt.Sprintf("项目未配置 %s/%s 角色，蓝图可能无法执行", roleType, roleLevel))
		}
	}

	// 完成 stage_task
	status := "completed"
	if result.HasErrors {
		status = "failed"
	}
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": status, "completed_at": time.Now(), "updated_at": time.Now()})

	g.Log().Infof(ctx, "[ReviewStage] Precheck done hasErrors=%v blueprintCount=%d", result.HasErrors, len(blueprints))
	return result, nil
}

// runAuditorReview 审计员 AI 审核。
func (s *Service) runAuditorReview(ctx context.Context, stageRunID, workflowRunID, planVersionID, projectID int64, blueprints gdb.Result) (bool, error) {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypeAuditorReview, "auditor")
	now := time.Now()
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": "running", "started_at": now, "updated_at": now})

	// 复用旧引擎的审计员审核逻辑
	result, err := engine.RunAuditorReviewForBlueprints(ctx, projectID, blueprints)

	taskStatus := "completed"
	var outputJSON []byte
	var errMsg string

	if err != nil {
		taskStatus = "failed"
		errMsg = err.Error()
	} else {
		outputJSON, _ = json.Marshal(result)
		if !result.Approved {
			// 审计员未通过，记录问题
			for _, issue := range result.Issues {
				s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, 0, issue.Severity, "auditor_issue", "auditor", issue.TaskName, issue.Message)
			}
		}
	}

	// 更新 stage_task
	updateData := g.Map{"status": taskStatus, "completed_at": time.Now(), "updated_at": time.Now()}
	if len(outputJSON) > 0 {
		updateData["output_payload"] = string(outputJSON)
	}
	if errMsg != "" {
		updateData["error_message"] = errMsg
	}
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).Update(updateData)

	if err != nil {
		return false, err
	}
	return result.Approved, nil
}

// runCoordinatorOptimize 协调员优化。
func (s *Service) runCoordinatorOptimize(ctx context.Context, stageRunID, workflowRunID, planVersionID, projectID int64, blueprints gdb.Result) error {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypeCoordinatorOptimize, "coordinator")
	now := time.Now()
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": "running", "started_at": now, "updated_at": now})

	result, err := engine.RunCoordinatorOptimizeForBlueprints(ctx, projectID, blueprints)

	taskStatus := "completed"
	var outputJSON []byte
	var errMsg string

	if err != nil {
		taskStatus = "failed"
		errMsg = err.Error()
		g.Log().Warningf(ctx, "[ReviewStage] 协调员优化失败: %v", err)
	} else {
		outputJSON, _ = json.Marshal(result)
		// 应用优化到蓝图的 batch_no
		engine.ApplyCoordinatorOptimizationsToBlueprints(ctx, planVersionID, blueprints, result)
	}

	updateData := g.Map{"status": taskStatus, "completed_at": time.Now(), "updated_at": time.Now()}
	if len(outputJSON) > 0 {
		updateData["output_payload"] = string(outputJSON)
	}
	if errMsg != "" {
		updateData["error_message"] = errMsg
	}
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).Update(updateData)
	return err
}

// concludeReview 汇总审核结论。
func (s *Service) concludeReview(ctx context.Context, stageRunID, planVersionID, projectID int64, passed bool, summary string) error {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypeReviewSummary, "system")
	now := gtime.Now()

	// 统计问题
	stageRun, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).One()
	workflowRunID := stageRun["workflow_run_id"].Int64()

	errorCount, _ := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).Where("severity", "error").Where("status", "open").Count()
	warningCount, _ := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).Where("severity", "warning").Where("status", "open").Count()

	outputPayload := g.Map{
		"passed":       passed,
		"summary":      summary,
		"error_count":  errorCount,
		"warning_count": warningCount,
	}
	outputJSON, _ := json.Marshal(outputPayload)

	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{
			"status":         "completed",
			"started_at":     now,
			"completed_at":   now,
			"output_payload": string(outputJSON),
			"updated_at":     now,
		})

	if passed {
		// 审核通过：更新 plan_version review_status
		_, _ = g.DB().Model("mvp_plan_version").Ctx(ctx).
			Where("id", planVersionID).
			Update(g.Map{"review_status": consts.PlanReviewStatusApproved, "approved_at": now, "updated_at": now})

		// 完成 review stage
		_ = s.stageCompleter.CompleteStage(ctx, stageRunID)

		// 推进到 execute stage
		if s.executeTriggerFn != nil {
			if err := s.executeTriggerFn(ctx, workflowRunID, planVersionID); err != nil {
				g.Log().Errorf(ctx, "[ReviewStage] 推进执行阶段失败: %v", err)
				// 即使推进失败也不阻塞审核结论，后续可手动触发
			}
		}

		// 通知架构师对话
		engine.NotifyProjectArchitectConversation(ctx, projectID,
			fmt.Sprintf("## 方案审核通过\n\n错误: %d，警告: %d\n\n项目已进入执行阶段。", errorCount, warningCount))

		g.Log().Infof(ctx, "[ReviewStage] 审核通过 planVersionID=%d errors=%d warnings=%d", planVersionID, errorCount, warningCount)
	} else {
		// 审核不通过：回退
		_, _ = g.DB().Model("mvp_plan_version").Ctx(ctx).
			Where("id", planVersionID).
			Update(g.Map{"review_status": consts.PlanReviewStatusRejected, "rejected_at": now, "updated_at": now})

		// 项目状态回退 designing
		_, _ = g.DB().Model("mvp_project").Ctx(ctx).
			Where("id", projectID).
			Update(g.Map{"status": "designing", "updated_at": now})

		// 失败 review stage
		_ = s.stageCompleter.FailStage(ctx, stageRunID, summary)

		// 回退 design stage（通过 orchestrator 回调）
		if s.designRollbackFn != nil {
			_ = s.designRollbackFn(ctx, workflowRunID)
		}

		// 通知架构师对话
		notifyMsg := s.buildRejectNotification(ctx, stageRunID, summary)
		engine.NotifyProjectArchitectConversation(ctx, projectID, notifyMsg)

		g.Log().Infof(ctx, "[ReviewStage] 审核不通过 planVersionID=%d reason=%s", planVersionID, summary)
	}

	return nil
}

// DesignRollbackFn 回退到 design 阶段的回调。
type DesignRollbackFn func(ctx context.Context, workflowRunID int64) error

var _ DesignRollbackFn // 类型守卫

// designRollbackFn 回退回调（由 registry 注入）。
// 使用 Service 的字段方式存储。
func (s *Service) SetDesignRollbackFn(fn DesignRollbackFn) {
	s.designRollbackFn = fn
}

// SetExecuteTriggerFn 注册审核通过后的执行阶段触发回调。
func (s *Service) SetExecuteTriggerFn(fn ExecuteTriggerFn) {
	s.executeTriggerFn = fn
}

// buildRejectNotification 构建审核驳回通知消息。
func (s *Service) buildRejectNotification(ctx context.Context, stageRunID int64, summary string) string {
	issues, _ := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("severity", "error").
		Where("status", "open").
		OrderDesc("created_at").
		All()

	msg := "## 方案审核未通过\n\n"
	msg += "### 原因\n" + summary + "\n\n"

	if len(issues) > 0 {
		msg += "### 错误列表\n"
		for i, issue := range issues {
			taskRef := ""
			if tn := issue["task_name"].String(); tn != "" {
				taskRef = fmt.Sprintf("[%s] ", tn)
			}
			msg += fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue["message"].String())
		}
	}

	msg += "\n请修正上述问题后重新确认方案。"
	return msg
}

// createStageTask 创建阶段子任务记录。
func (s *Service) createStageTask(ctx context.Context, stageRunID int64, taskType, roleType string) int64 {
	id := int64(snowflake.Generate())
	now := time.Now()
	_, _ = g.DB().Model("mvp_stage_task").Ctx(ctx).Insert(g.Map{
		"id":           id,
		"stage_run_id": stageRunID,
		"task_type":    taskType,
		"role_type":    roleType,
		"status":       "pending",
		"created_at":   now,
		"updated_at":   now,
	})
	return id
}

// createIssue 创建审核问题记录。
func (s *Service) createIssue(ctx context.Context, workflowRunID, stageRunID, planVersionID, blueprintID int64, severity, issueCode, sourceRole, taskName, message string) {
	now := time.Now()
	data := g.Map{
		"id":              int64(snowflake.Generate()),
		"workflow_run_id": workflowRunID,
		"stage_run_id":    stageRunID,
		"plan_version_id": planVersionID,
		"severity":        severity,
		"issue_code":      issueCode,
		"issue_type":      sourceRole,
		"source_role":     sourceRole,
		"task_name":       taskName,
		"message":         message,
		"status":          "open",
		"created_at":      now,
		"updated_at":      now,
	}
	if blueprintID > 0 {
		data["blueprint_id"] = blueprintID
	}
	_, _ = g.DB().Model("mvp_review_issue").Ctx(ctx).Insert(data)
}
