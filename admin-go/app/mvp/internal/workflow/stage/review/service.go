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
	FailStageOnly(ctx context.Context, stageRunID int64, reason string)
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

type reviewConclusion struct {
	passed            bool
	summary           string
	blockedByWarnings bool
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
	stageRun, err := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).WhereNull("deleted_at").One()
	if err != nil || stageRun.IsEmpty() {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := stageRun["workflow_run_id"].Int64()

	// 查 workflow_run 获取 project_id
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).WhereNull("deleted_at").One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}
	projectID := wfRun["project_id"].Int64()

	// 获取项目信息
	project, err := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).WhereNull("deleted_at").One()
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
	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": "running", "started_at": now, "updated_at": now}); stErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] precheck stage_task 标记 running 失败: id=%d err=%v", stageTaskID, stErr)
	}

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

	// 预加载角色配置（缺失的自动从默认预设补齐）
	availableRoles := make(map[string]bool)
	roleMap, _ := repo.GetProjectRolesMap(ctx, projectID)
	for key := range roleMap {
		availableRoles[key] = true
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

		// 描述质量（降为 warning，不阻塞审核）
		if len([]rune(desc)) < 5 {
			s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "warning", "short_desc", "precheck", name,
				fmt.Sprintf("蓝图描述较短（%d字），建议至少5字的有效描述", len([]rune(desc))))
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
			if umErr := json.Unmarshal([]byte(depJSON), &depIDs); umErr != nil {
				g.Log().Warningf(ctx, "[ReviewPrecheck] 解析蓝图依赖失败: bp=%d err=%v", bp["id"].Int64(), umErr)
			}
		}

		// 资源冲突（降为 warning，不阻塞审核；调度器有资源锁保护）
		if _, ok := batchResources[batchNo]; !ok {
			batchResources[batchNo] = make(map[string]string)
		}
		for _, res := range resources {
			if existingTask, conflict := batchResources[batchNo][res]; conflict {
				s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, bpID, "warning", "resource_conflict", "precheck", name,
					fmt.Sprintf("资源冲突: 同批次(%d)中 [%s] 和 [%s] 都修改 %s（调度器会自动串行化）", batchNo, existingTask, name, res))
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
	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": status, "completed_at": time.Now(), "updated_at": time.Now()}); stErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] precheck stage_task 完成更新失败: id=%d err=%v", stageTaskID, stErr)
	}

	g.Log().Infof(ctx, "[ReviewStage] Precheck done hasErrors=%v blueprintCount=%d", result.HasErrors, len(blueprints))
	return result, nil
}

// runAuditorReview 审计员 AI 审核。
func (s *Service) runAuditorReview(ctx context.Context, stageRunID, workflowRunID, planVersionID, projectID int64, blueprints gdb.Result) (bool, error) {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypeAuditorReview, "auditor")
	now := time.Now()
	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": "running", "started_at": now, "updated_at": now}); stErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] auditor stage_task 标记 running 失败: id=%d err=%v", stageTaskID, stErr)
	}

	// 复用旧引擎的审计员审核逻辑
	result, err := engine.RunAuditorReviewForBlueprints(ctx, projectID, blueprints)

	taskStatus := "completed"
	var outputJSON []byte
	var errMsg string

	if err != nil {
		taskStatus = "failed"
		errMsg = err.Error()
	} else {
		var mErr error
		outputJSON, mErr = json.Marshal(result)
		if mErr != nil {
			g.Log().Warningf(ctx, "[ReviewStage] auditor result 序列化失败: %v", mErr)
		}
		if !result.Approved {
			// 审计员未通过，记录问题
			if len(result.Issues) > 0 {
				for _, issue := range result.Issues {
					s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, 0, issue.Severity, "auditor_issue", "auditor", issue.TaskName, issue.Message)
				}
			} else {
				// AI 返回 approved=false 但 issues 为空，用 suggestions 构造 issue
				msg := result.Suggestions
				if msg == "" {
					msg = "审计员认为方案需要改进，但未提供具体问题描述"
				}
				s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, 0, "error", "auditor_reject", "auditor", "", msg)
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
	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).Update(updateData); stErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] auditor stage_task 完成更新失败: id=%d err=%v", stageTaskID, stErr)
	}

	if err != nil {
		return false, err
	}
	return result.Approved, nil
}

// runCoordinatorOptimize 协调员优化。
func (s *Service) runCoordinatorOptimize(ctx context.Context, stageRunID, workflowRunID, planVersionID, projectID int64, blueprints gdb.Result) error {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypeCoordinatorOptimize, "coordinator")
	now := time.Now()
	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{"status": "running", "started_at": now, "updated_at": now}); stErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] coordinator stage_task 标记 running 失败: id=%d err=%v", stageTaskID, stErr)
	}

	result, err := engine.RunCoordinatorOptimizeForBlueprints(ctx, projectID, blueprints)

	taskStatus := "completed"
	var outputJSON []byte
	var errMsg string

	if err != nil {
		errMsg = err.Error()
		g.Log().Warningf(ctx, "[ReviewStage] 协调员优化失败: %v", err)
		s.createIssue(ctx, workflowRunID, stageRunID, planVersionID, 0, "warning", "coordinator_optimize_skip", "coordinator", "",
			fmt.Sprintf("协调员优化已跳过，不影响审核主链: %v", err))
		outputJSON, _ = json.Marshal(g.Map{
			"skipped": true,
			"reason":  errMsg,
		})
	} else {
		var coErr error
		outputJSON, coErr = json.Marshal(result)
		if coErr != nil {
			g.Log().Warningf(ctx, "[ReviewStage] coordinator result 序列化失败: %v", coErr)
		}
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
	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).Update(updateData); stErr != nil {
		g.Log().Warningf(ctx, "[ReviewStage] coordinator stage_task 完成更新失败: id=%d err=%v", stageTaskID, stErr)
	}
	return err
}

// concludeReview 汇总审核结论。
func (s *Service) concludeReview(ctx context.Context, stageRunID, planVersionID, projectID int64, passed bool, summary string) error {
	stageTaskID := s.createStageTask(ctx, stageRunID, TaskTypeReviewSummary, "system")
	now := gtime.Now()

	// 统计问题
	stageRun, srErr := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).WhereNull("deleted_at").One()
	if srErr != nil || stageRun.IsEmpty() {
		g.Log().Errorf(ctx, "[ReviewService] concludeReview 查询 stage_run 失败: stageRun=%d err=%v", stageRunID, srErr)
		return fmt.Errorf("查询 stage_run(%d) 失败: %v", stageRunID, srErr)
	}
	workflowRunID := stageRun["workflow_run_id"].Int64()

	errorCount, ecErr := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).Where("severity", "error").Where("status", "open").Count()
	if ecErr != nil {
		g.Log().Warningf(ctx, "[ReviewService] 查询 error 数失败: stageRun=%d err=%v", stageRunID, ecErr)
	}
	warningCount, wcErr := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).Where("severity", "warning").Where("status", "open").Count()
	if wcErr != nil {
		g.Log().Warningf(ctx, "[ReviewService] 查询 warning 数失败: stageRun=%d err=%v", stageRunID, wcErr)
	}

	conclusion := finalizeReviewConclusion(passed, warningCount, summary)

	outputPayload := g.Map{
		"passed":              conclusion.passed,
		"requested_passed":    passed,
		"summary":             conclusion.summary,
		"error_count":         errorCount,
		"warning_count":       warningCount,
		"blocked_by_warnings": conclusion.blockedByWarnings,
	}
	outputJSON, marshalErr := json.Marshal(outputPayload)
	if marshalErr != nil {
		g.Log().Warningf(ctx, "[ReviewService] outputPayload 序列化失败: stageTask=%d err=%v", stageTaskID, marshalErr)
		outputJSON = []byte("{}")
	}

	if _, stErr := g.DB().Model("mvp_stage_task").Ctx(ctx).Where("id", stageTaskID).
		Update(g.Map{
			"status":         "completed",
			"started_at":     now,
			"completed_at":   now,
			"output_payload": string(outputJSON),
			"updated_at":     now,
		}); stErr != nil {
		g.Log().Errorf(ctx, "[ReviewService] 更新 stage_task 状态失败: stageTask=%d err=%v", stageTaskID, stErr)
	}

	if conclusion.passed {
		// 审核通过：更新 plan_version review_status
		if _, pvErr := g.DB().Model("mvp_plan_version").Ctx(ctx).
			Where("id", planVersionID).
			Update(g.Map{"review_status": consts.PlanReviewStatusApproved, "approved_at": now, "updated_at": now}); pvErr != nil {
			g.Log().Errorf(ctx, "[ReviewStage] 更新 plan_version 审核状态失败: pvID=%d err=%v", planVersionID, pvErr)
		}

		// 完成 review stage
		if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
			g.Log().Errorf(ctx, "[ReviewStage] CompleteStage 失败，回滚 review_status: %v", err)
			if _, rbErr := g.DB().Model("mvp_plan_version").Ctx(ctx).
				Where("id", planVersionID).
				Where("review_status", consts.PlanReviewStatusApproved).
				Update(g.Map{"review_status": consts.PlanReviewStatusPending, "approved_at": nil, "updated_at": gtime.Now()}); rbErr != nil {
				g.Log().Errorf(ctx, "[ReviewStage] 回滚 plan_version 也失败: pv=%d err=%v", planVersionID, rbErr)
			}
			return fmt.Errorf("完成审核阶段失败: %w", err)
		}

		// 推进到 execute stage
		executeStarted := false
		if s.executeTriggerFn != nil {
			if err := s.executeTriggerFn(ctx, workflowRunID, planVersionID); err != nil {
				g.Log().Errorf(ctx, "[ReviewStage] 推进执行阶段失败，回滚审核状态: %v", err)

				// 回滚 plan_version review_status → pending（CAS：仅回滚已通过的）
				if _, rbErr := g.DB().Model("mvp_plan_version").Ctx(ctx).
					Where("id", planVersionID).
					Where("review_status", consts.PlanReviewStatusApproved).
					Update(g.Map{"review_status": consts.PlanReviewStatusPending, "approved_at": nil, "updated_at": gtime.Now()}); rbErr != nil {
					g.Log().Errorf(ctx, "[ReviewStage] 回滚 plan_version 失败: pv=%d err=%v", planVersionID, rbErr)
				}

				// 回滚 review stage: completed → running（CAS 防并发）
				if _, rbErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
					Where("id", stageRunID).
					Where("status", consts.StageStatusCompleted).
					Update(g.Map{"status": consts.StageStatusRunning, "finished_at": nil, "updated_at": gtime.Now()}); rbErr != nil {
					g.Log().Errorf(ctx, "[ReviewStage] 回滚 stage_run 失败: stage=%d err=%v", stageRunID, rbErr)
				}

				// 回滚 workflow_run 状态 → reviewing（CAS：仅回滚 executing/failed）
				wfRollback := g.Map{
					"status":               consts.WorkflowRunStatusReviewing,
					"current_stage":        "review",
					"current_stage_run_id": stageRunID,
					"updated_at":           gtime.Now(),
				}
				if _, rbErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
					Where("id", workflowRunID).
					WhereIn("status", g.Slice{consts.WorkflowRunStatusExecuting, consts.WorkflowRunStatusFailed}).
					Update(wfRollback); rbErr != nil {
					g.Log().Errorf(ctx, "[ReviewStage] 回滚 workflow_run 失败: wfRun=%d err=%v", workflowRunID, rbErr)
				}

				// 同步 mvp_project.status 回 reviewing（CAS：仅回滚 executing 状态的项目）
				if _, rbErr := g.DB().Model("mvp_project").Ctx(ctx).
					Where("id", projectID).
					WhereIn("status", g.Slice{consts.WorkflowRunStatusExecuting, consts.WorkflowRunStatusFailed}).
					Update(g.Map{"status": consts.WorkflowRunStatusReviewing, "pause_reason": nil, "updated_at": gtime.Now()}); rbErr != nil {
					g.Log().Errorf(ctx, "[ReviewStage] 回滚 project 状态失败: project=%d err=%v", projectID, rbErr)
				}

				engine.NotifyProjectArchitectConversation(ctx, projectID,
					fmt.Sprintf("## 方案审核通过但执行启动失败\n\n错误: %d，警告: %d\n\n⚠️ 执行阶段启动失败，已自动回滚到审核状态。请检查日志后重新确认方案。\n\n失败原因: %v", errorCount, warningCount, err))
				return fmt.Errorf("执行阶段启动失败（已回滚审核状态）: %w", err)
			} else {
				executeStarted = true
			}
		}

		// 通知架构师对话
		if executeStarted {
			engine.NotifyProjectArchitectConversation(ctx, projectID,
				fmt.Sprintf("## 方案审核通过\n\n错误: %d，警告: %d\n\n项目已进入执行阶段。", errorCount, warningCount))
		}

		g.Log().Infof(ctx, "[ReviewStage] 审核通过 planVersionID=%d errors=%d warnings=%d", planVersionID, errorCount, warningCount)
	} else {
		// 审核不通过：回退 plan_version + blueprints + workflow + project
		if _, pvErr := g.DB().Model("mvp_plan_version").Ctx(ctx).
			Where("id", planVersionID).
			Update(g.Map{
				"status":        "draft",
				"review_status": consts.PlanReviewStatusRejected,
				"rejected_at":   now,
				"updated_at":    now,
			}); pvErr != nil {
			g.Log().Errorf(ctx, "[ReviewStage] 驳回更新 plan_version 失败: pvID=%d err=%v", planVersionID, pvErr)
		}

		// 蓝图回退 confirmed → draft（让用户可以重新编辑）
		if _, bpErr := g.DB().Model("mvp_task_blueprint").Ctx(ctx).
			Where("plan_version_id", planVersionID).
			Where("blueprint_status", consts.BlueprintStatusConfirmed).
			Update(g.Map{"blueprint_status": consts.BlueprintStatusDraft, "updated_at": now}); bpErr != nil {
			g.Log().Errorf(ctx, "[ReviewStage] 蓝图回退失败: pvID=%d err=%v", planVersionID, bpErr)
		}

		// 标记 review stage 失败（仅标记 stage_run，不级联终止 workflow）
		s.stageCompleter.FailStageOnly(ctx, stageRunID, conclusion.summary)

		// 回退到 design 阶段（StartStage("design") 会同时更新 workflow_run.status 和 project.status）
		designRollbackOK := false
		if s.designRollbackFn != nil {
			if rollbackErr := s.designRollbackFn(ctx, workflowRunID); rollbackErr != nil {
				g.Log().Errorf(ctx, "[ReviewStage] designRollbackFn 失败，执行 fallback: workflowRunID=%d err=%v", workflowRunID, rollbackErr)
			} else {
				designRollbackOK = true
			}
		}

		// fallback：如果 designRollbackFn 失败或未注册，直接更新 workflow_run + project 状态
		if !designRollbackOK {
			if _, wfErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
				Where("id", workflowRunID).
				WhereIn("status", g.Slice{"reviewing", "failed"}).
				Update(g.Map{
					"status":        "designing",
					"current_stage": "design",
					"updated_at":    now,
				}); wfErr != nil {
				g.Log().Errorf(ctx, "[ReviewStage] fallback 回退 workflow_run 失败: wfRun=%d err=%v", workflowRunID, wfErr)
			}
			if _, pErr := g.DB().Model("mvp_project").Ctx(ctx).
				Where("id", projectID).
				Update(g.Map{"status": "designing", "updated_at": now}); pErr != nil {
				g.Log().Errorf(ctx, "[ReviewStage] fallback 回退 project 失败: project=%d err=%v", projectID, pErr)
			}
		}

		// 通知架构师对话
		notifyMsg := s.buildRejectNotification(ctx, stageRunID, conclusion.summary)
		engine.NotifyProjectArchitectConversation(ctx, projectID, notifyMsg)

		g.Log().Infof(ctx, "[ReviewStage] 审核不通过 planVersionID=%d reason=%s designRollbackOK=%v", planVersionID, conclusion.summary, designRollbackOK)
	}

	return nil
}

func finalizeReviewConclusion(passed bool, warningCount int, summary string) reviewConclusion {
	if !passed {
		return reviewConclusion{
			passed:  false,
			summary: summary,
		}
	}
	if warningCount > 0 {
		return reviewConclusion{
			passed:            false,
			summary:           fmt.Sprintf("审核发现 %d 条警告，需全部修复后才能进入执行阶段", warningCount),
			blockedByWarnings: true,
		}
	}
	return reviewConclusion{
		passed:  true,
		summary: summary,
	}
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
	// 查所有 error 和 warning 级别的 issue
	errors, errQry := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("severity", "error").
		Where("status", "open").
		OrderDesc("created_at").
		All()
	if errQry != nil {
		g.Log().Warningf(ctx, "[ReviewService] 查询 error issues 失败: stageRun=%d err=%v", stageRunID, errQry)
	}

	warnings, warnQry := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("severity", "warning").
		Where("status", "open").
		OrderDesc("created_at").
		All()
	if warnQry != nil {
		g.Log().Warningf(ctx, "[ReviewService] 查询 warning issues 失败: stageRun=%d err=%v", stageRunID, warnQry)
	}

	msg := "## 方案审核未通过\n\n"
	msg += "### 原因\n" + summary + "\n\n"

	if len(errors) > 0 {
		msg += "### 错误（必须修复）\n"
		for i, issue := range errors {
			taskRef := ""
			if tn := issue["task_name"].String(); tn != "" {
				taskRef = fmt.Sprintf("[%s] ", tn)
			}
			msg += fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue["message"].String())
		}
		msg += "\n"
	}

	if len(warnings) > 0 {
		msg += "### 警告（当前会阻塞执行，必须修复）\n"
		for i, issue := range warnings {
			taskRef := ""
			if tn := issue["task_name"].String(); tn != "" {
				taskRef = fmt.Sprintf("[%s] ", tn)
			}
			msg += fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue["message"].String())
		}
		msg += "\n"
	}

	// 查审计员的 suggestions（从 stage_task output_payload 中提取）
	auditorTask, atErr := g.DB().Model("mvp_stage_task").Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("task_type", TaskTypeAuditorReview).
		One()
	if atErr != nil {
		g.Log().Warningf(ctx, "[ReviewService] 查询 auditor stage_task 失败: stageRun=%d err=%v", stageRunID, atErr)
	}
	if !auditorTask.IsEmpty() {
		payload := auditorTask["output_payload"].String()
		if payload != "" {
			var auditorOutput struct {
				Suggestions string `json:"suggestions"`
			}
			if json.Unmarshal([]byte(payload), &auditorOutput) == nil && auditorOutput.Suggestions != "" {
				msg += "### 审计员建议\n" + auditorOutput.Suggestions + "\n\n"
			}
		}
	}

	msg += "请逐条处理以上问题，重新给出修订结果。"
	msg += "\n\n如果是整体重排方案，请直接输出完整 JSON：{\"tasks\": [...]}，系统会自动解析为新的方案版本。"
	msg += "\n如果只是修正个别任务，请输出局部修订 JSON：{\"task_patches\": [{\"task_name\": \"原任务名\", \"description\": \"修订后的描述\", \"affected_resources\": [\"路径\"], \"depends_on\": [\"依赖任务名\"], \"reason\": \"修订原因\"}]}"
	msg += "\n如果内容太长，可以拆成多段发送；每一段都请输出独立合法 JSON。若你还有后续分段，请在当前消息最后单独追加一行 [AUTO_CONTINUE_NEXT]，系统才会自动继续索取下一段；最后一段不要追加该标记。"
	return msg
}

// createStageTask 创建阶段子任务记录。
func (s *Service) createStageTask(ctx context.Context, stageRunID int64, taskType, roleType string) int64 {
	id := int64(snowflake.Generate())
	now := time.Now()
	if _, err := g.DB().Model("mvp_stage_task").Ctx(ctx).Insert(g.Map{
		"id":           id,
		"stage_run_id": stageRunID,
		"task_type":    taskType,
		"role_type":    roleType,
		"status":       "pending",
		"created_at":   now,
		"updated_at":   now,
	}); err != nil {
		g.Log().Errorf(ctx, "[ReviewService] 创建阶段子任务失败: stageRun=%d type=%s err=%v", stageRunID, taskType, err)
	}
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
	if _, err := g.DB().Model("mvp_review_issue").Ctx(ctx).Insert(data); err != nil {
		g.Log().Errorf(ctx, "[ReviewService] 创建审核问题失败: wfRun=%d severity=%s err=%v", workflowRunID, severity, err)
	}
}
