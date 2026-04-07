// Package plan 计划版本领域服务。
package plan

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

// ReviewTrigger 审核触发回调，由 orchestrator 注册。
type ReviewTrigger func(ctx context.Context, projectID, planVersionID int64) error

// PlanVersionService 计划版本服务。
type PlanVersionService struct {
	planRepo       *repo.PlanVersionRepo
	blueprintRepo  *repo.BlueprintRepo
	reviewTrigger  ReviewTrigger
}

// NewPlanVersionService 创建计划版本服务。
func NewPlanVersionService(pr *repo.PlanVersionRepo, br *repo.BlueprintRepo) *PlanVersionService {
	return &PlanVersionService{planRepo: pr, blueprintRepo: br}
}

// SetReviewTrigger 注册审核触发回调。
func (s *PlanVersionService) SetReviewTrigger(fn ReviewTrigger) {
	s.reviewTrigger = fn
}

// CreateFromArchitectReply 从架构师解析结果创建新的计划版本和任务蓝图。
// tasks 来自 TaskParser.ExtractAndNormalize。
// 返回 planVersionID 和蓝图数量。
// 整个操作在同一事务中完成，保证旧版本废弃与新版本创建的原子性。
func (s *PlanVersionService) CreateFromArchitectReply(
	ctx context.Context,
	projectID, workflowRunID, conversationID, messageID int64,
	tasks []engine.ArchitectTask,
) (int64, int, error) {
	if len(tasks) == 0 {
		return 0, 0, fmt.Errorf("没有可创建的任务蓝图")
	}

	// 预生成所有 ID（事务外分配，减少事务持有时间）
	pvID := int64(snowflake.Generate())
	blueprintIDs := make([]int64, len(tasks))
	nameToID := make(map[string]int64, len(tasks))
	for i, task := range tasks {
		bpID := int64(snowflake.Generate())
		blueprintIDs[i] = bpID
		nameToID[task.Name] = bpID
	}

	now := time.Now()
	var versionNo int

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 获取下一个 version_no
		var err error
		versionNo, err = s.planRepo.NextVersionNo(ctx, projectID)
		if err != nil {
			return fmt.Errorf("获取 version_no 失败: %w", err)
		}

		// 2. 将之前的活跃版本标记为 superseded
		if err := s.supersedePreviousVersionsInTx(ctx, projectID, 0); err != nil {
			return fmt.Errorf("supersede 旧版本失败: %w", err)
		}

		// 3. 获取项目归属字段
		scope := repo.GetProjectScopeByProject(ctx, projectID)

		// 4. 创建 plan_version
		_, err = tx.Model("mvp_plan_version").Ctx(ctx).Insert(g.Map{
			"id":                     pvID,
			"project_id":             projectID,
			"workflow_run_id":        workflowRunID,
			"version_no":             versionNo,
			"source_conversation_id": conversationID,
			"source_message_id":      messageID,
			"status":                 consts.PlanVersionStatusDraft,
			"review_status":          consts.PlanReviewStatusPending,
			"summary":                fmt.Sprintf("第 %d 版方案，共 %d 个任务蓝图", versionNo, len(tasks)),
			"created_by":            scope.CreatedBy,
			"dept_id":               scope.DeptID,
			"created_at":             now,
			"updated_at":             now,
		})
		if err != nil {
			return fmt.Errorf("创建 plan_version 失败: %w", err)
		}

		// 4. 创建蓝图
		for i, task := range tasks {
			affectedJSON, _ := json.Marshal(task.AffectedResources)
			_, err = tx.Model("mvp_task_blueprint").Ctx(ctx).Insert(g.Map{
				"id":                 blueprintIDs[i],
				"plan_version_id":    pvID,
				"name":               task.Name,
				"description":        task.Description,
				"role_type":          defaultRoleType(task.RoleType),
				"role_level":         defaultRoleLevel(task.RoleLevel),
				"batch_no":           task.BatchNo,
				"sort":               i + 1,
				"affected_resources": string(affectedJSON),
				"blueprint_status":   consts.BlueprintStatusDraft,
				"created_by":        scope.CreatedBy,
				"dept_id":           scope.DeptID,
				"created_at":         now,
				"updated_at":         now,
			})
			if err != nil {
				return fmt.Errorf("创建蓝图(%s)失败: %w", task.Name, err)
			}
		}

		// 5. 回写依赖关系
		for i, task := range tasks {
			if len(task.DependsOn) == 0 {
				continue
			}
			depIDs := make([]int64, 0, len(task.DependsOn))
			for _, depName := range task.DependsOn {
				if depID, ok := nameToID[depName]; ok {
					depIDs = append(depIDs, depID)
				}
			}
			if len(depIDs) > 0 {
				depJSON, _ := json.Marshal(depIDs)
				if _, err := tx.Model("mvp_task_blueprint").Ctx(ctx).
					Where("id", blueprintIDs[i]).
					Update(g.Map{"depends_on_blueprint_ids": string(depJSON)}); err != nil {
					return fmt.Errorf("回写蓝图依赖失败: %w", err)
				}
			}
		}

		// 6. 更新 workflow_run 的 active_plan_version_id（CAS 校验）
		if workflowRunID > 0 {
			wfResult, err := tx.Model("mvp_workflow_run").Ctx(ctx).
				Where("id", workflowRunID).
				Update(g.Map{"active_plan_version_id": pvID, "updated_at": now})
			if err != nil {
				return fmt.Errorf("更新 active_plan_version_id 失败: %w", err)
			}
			if rows, _ := wfResult.RowsAffected(); rows == 0 {
				return fmt.Errorf("workflow_run(%d) 不存在，无法关联 plan_version", workflowRunID)
			}
		}

		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	g.Log().Infof(ctx, "[PlanVersionService] 创建 plan_version=%d (v%d), %d 个蓝图, project=%d",
		pvID, versionNo, len(tasks), projectID)
	return pvID, len(tasks), nil
}

// supersedePreviousVersionsInTx 事务内版本：将项目之前的活跃/草稿版本标记为 superseded。
// 使用事务传播的 ctx，GoFrame 会自动路由到当前事务。
func (s *PlanVersionService) supersedePreviousVersionsInTx(ctx context.Context, projectID int64, exceptVersionID int64) error {
	return s.doSupersede(ctx, projectID, exceptVersionID)
}

// SupersedePreviousVersions 将项目之前的活跃/草稿版本标记为 superseded（非事务调用入口）。
func (s *PlanVersionService) SupersedePreviousVersions(ctx context.Context, projectID int64, exceptVersionID int64) error {
	return s.doSupersede(ctx, projectID, exceptVersionID)
}

// doSupersede 核心 supersede 逻辑，事务和非事务场景共用。
// GoFrame 的 ctx 事务传播机制会自动路由到当前事务（如有）。
func (s *PlanVersionService) doSupersede(ctx context.Context, projectID int64, exceptVersionID int64) error {
	now := gtime.Now()
	q := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{consts.PlanVersionStatusDraft, consts.PlanVersionStatusActive}).
		WhereNull("deleted_at")
	if exceptVersionID > 0 {
		q = q.WhereNot("id", exceptVersionID)
	}

	oldIDs, err := q.Fields("id").Array()
	if err != nil || len(oldIDs) == 0 {
		return err
	}

	idList := make([]int64, 0, len(oldIDs))
	for _, v := range oldIDs {
		idList = append(idList, v.Int64())
	}

	_, err = g.DB().Model("mvp_plan_version").Ctx(ctx).
		WhereIn("id", idList).
		Update(g.Map{"status": consts.PlanVersionStatusSuperseded, "updated_at": now})
	if err != nil {
		return err
	}

	_, err = g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		WhereIn("plan_version_id", idList).
		WhereNot("blueprint_status", consts.BlueprintStatusSuperseded).
		Update(g.Map{"blueprint_status": consts.BlueprintStatusSuperseded, "updated_at": now})
	return err
}

// SubmitForReview 提交当前草稿版本进入审核。
// 流程：plan_version draft→active, blueprints draft→confirmed, project status→reviewing, 触发 review stage。
// 状态变更在事务中完成；reviewTrigger 同步调用，失败则回滚全部状态。
func (s *PlanVersionService) SubmitForReview(ctx context.Context, projectID int64) error {
	now := gtime.Now()

	// 1. 找最新的 draft plan_version
	pv, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", consts.PlanVersionStatusDraft).
		WhereNull("deleted_at").
		OrderDesc("version_no").
		One()
	if err != nil || pv.IsEmpty() {
		return fmt.Errorf("没有待确认的方案版本")
	}
	pvID := pv["id"].Int64()

	// 2. 检查蓝图数
	bpCount, _ := g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		Where("plan_version_id", pvID).
		WhereNull("deleted_at").
		Count()
	if bpCount == 0 {
		return fmt.Errorf("方案版本没有任务蓝图")
	}

	// 3. 事务内完成所有状态迁移
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// plan_version: draft → active（CAS）
		pvResult, err := tx.Model("mvp_plan_version").Ctx(ctx).
			Where("id", pvID).
			Where("status", consts.PlanVersionStatusDraft).
			Update(g.Map{"status": consts.PlanVersionStatusActive, "updated_at": now})
		if err != nil {
			return fmt.Errorf("更新 plan_version 状态失败: %w", err)
		}
		if rows, _ := pvResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("plan_version(%d) 已不在 draft 状态，无法提交审核", pvID)
		}

		// blueprints: draft → confirmed
		bpResult, err := tx.Model("mvp_task_blueprint").Ctx(ctx).
			Where("plan_version_id", pvID).
			Where("blueprint_status", consts.BlueprintStatusDraft).
			Update(g.Map{"blueprint_status": consts.BlueprintStatusConfirmed, "updated_at": now})
		if err != nil {
			return fmt.Errorf("确认蓝图状态失败: %w", err)
		}
		if rows, _ := bpResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("plan_version(%d) 下没有 draft 蓝图可确认", pvID)
		}

		// project status → reviewing
		projResult, err := tx.Model("mvp_project").Ctx(ctx).
			Where("id", projectID).
			Update(g.Map{"status": "reviewing", "pause_reason": nil, "updated_at": now})
		if err != nil {
			return fmt.Errorf("更新项目状态失败: %w", err)
		}
		if rows, _ := projResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("项目(%d) 不存在或状态更新失败", projectID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	g.Log().Infof(ctx, "[PlanVersionService] SubmitForReview projectID=%d pvID=%d bpCount=%d", projectID, pvID, bpCount)

	// 4. 同步触发 review stage——失败则回滚业务状态
	if s.reviewTrigger != nil {
		if triggerErr := s.reviewTrigger(ctx, projectID, pvID); triggerErr != nil {
			g.Log().Errorf(ctx, "[PlanVersionService] 触发审核失败，回滚状态: projectID=%d pvID=%d err=%v", projectID, pvID, triggerErr)
			// 回滚：active → draft, confirmed → draft, reviewing → designing
			_, _ = g.DB().Model("mvp_plan_version").Ctx(ctx).
				Where("id", pvID).Where("status", consts.PlanVersionStatusActive).
				Update(g.Map{"status": consts.PlanVersionStatusDraft, "updated_at": gtime.Now()})
			_, _ = g.DB().Model("mvp_task_blueprint").Ctx(ctx).
				Where("plan_version_id", pvID).Where("blueprint_status", consts.BlueprintStatusConfirmed).
				Update(g.Map{"blueprint_status": consts.BlueprintStatusDraft, "updated_at": gtime.Now()})
			_, _ = g.DB().Model("mvp_project").Ctx(ctx).
				Where("id", projectID).Where("status", "reviewing").
				Update(g.Map{"status": "designing", "updated_at": gtime.Now()})
			return fmt.Errorf("提交审核失败: %w", triggerErr)
		}
	}
	return nil
}

// SubmitForReviewAsync 异步提交审核。
// 同步完成状态变更（plan_version/blueprints/project），审核流程在后台 goroutine 中执行。
// 避免 HTTP 请求因 AI 审核耗时过长而超时。
func (s *PlanVersionService) SubmitForReviewAsync(ctx context.Context, projectID int64) error {
	now := gtime.Now()

	// 防重复提交：检查项目是否已在审核中
	projectStatus, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).WhereNull("deleted_at").Value("status")
	if projectStatus.String() == "reviewing" {
		return fmt.Errorf("方案已在审核中，请勿重复提交")
	}

	// 1. 找最新的 draft plan_version
	pv, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", consts.PlanVersionStatusDraft).
		WhereNull("deleted_at").
		OrderDesc("version_no").
		One()
	if err != nil || pv.IsEmpty() {
		return fmt.Errorf("没有待确认的方案版本")
	}
	pvID := pv["id"].Int64()

	// 2. 检查蓝图数
	bpCount, _ := g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		Where("plan_version_id", pvID).
		WhereNull("deleted_at").
		Count()
	if bpCount == 0 {
		return fmt.Errorf("方案版本没有任务蓝图")
	}

	// 3. 事务内完成所有状态迁移（同步，快速）
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		pvResult, err := tx.Model("mvp_plan_version").Ctx(ctx).
			Where("id", pvID).
			Where("status", consts.PlanVersionStatusDraft).
			Update(g.Map{"status": consts.PlanVersionStatusActive, "updated_at": now})
		if err != nil {
			return fmt.Errorf("更新 plan_version 状态失败: %w", err)
		}
		if rows, _ := pvResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("plan_version(%d) 已不在 draft 状态，无法提交审核", pvID)
		}

		bpResult, err := tx.Model("mvp_task_blueprint").Ctx(ctx).
			Where("plan_version_id", pvID).
			Where("blueprint_status", consts.BlueprintStatusDraft).
			Update(g.Map{"blueprint_status": consts.BlueprintStatusConfirmed, "updated_at": now})
		if err != nil {
			return fmt.Errorf("确认蓝图状态失败: %w", err)
		}
		if rows, _ := bpResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("plan_version(%d) 下没有 draft 蓝图可确认", pvID)
		}

		projResult, err := tx.Model("mvp_project").Ctx(ctx).
			Where("id", projectID).
			Update(g.Map{"status": "reviewing", "pause_reason": nil, "updated_at": now})
		if err != nil {
			return fmt.Errorf("更新项目状态失败: %w", err)
		}
		if rows, _ := projResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("项目(%d) 不存在或状态更新失败", projectID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	g.Log().Infof(ctx, "[PlanVersionService] SubmitForReviewAsync projectID=%d pvID=%d bpCount=%d", projectID, pvID, bpCount)

	// 4. 异步触发 review stage
	if s.reviewTrigger != nil {
		go func() {
			bgCtx := context.Background()
			defer func() {
				if r := recover(); r != nil {
					g.Log().Errorf(bgCtx, "[PlanVersionService] 异步审核 panic: projectID=%d pvID=%d err=%v", projectID, pvID, r)
				}
			}()

			if triggerErr := s.reviewTrigger(bgCtx, projectID, pvID); triggerErr != nil {
				g.Log().Errorf(bgCtx, "[PlanVersionService] 异步审核失败，回滚状态: projectID=%d pvID=%d err=%v", projectID, pvID, triggerErr)
				rollbackNow := gtime.Now()
				_, _ = g.DB().Model("mvp_plan_version").Ctx(bgCtx).
					Where("id", pvID).Where("status", consts.PlanVersionStatusActive).
					Update(g.Map{"status": consts.PlanVersionStatusDraft, "updated_at": rollbackNow})
				_, _ = g.DB().Model("mvp_task_blueprint").Ctx(bgCtx).
					Where("plan_version_id", pvID).Where("blueprint_status", consts.BlueprintStatusConfirmed).
					Update(g.Map{"blueprint_status": consts.BlueprintStatusDraft, "updated_at": rollbackNow})

				// 查找关联的 workflow_run 并回退状态
				wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(bgCtx).
					Where("project_id", projectID).
					WhereIn("status", g.Slice{"reviewing", "failed"}).
					WhereNull("deleted_at").
					OrderDesc("id").
					Fields("id").
					One()
				if !wfRun.IsEmpty() {
					_, _ = g.DB().Model("mvp_workflow_run").Ctx(bgCtx).
						Where("id", wfRun["id"].Int64()).
						WhereIn("status", g.Slice{"reviewing", "failed"}).
						Update(g.Map{"status": "designing", "current_stage": "design", "updated_at": rollbackNow})
				}

				_, _ = g.DB().Model("mvp_project").Ctx(bgCtx).
					Where("id", projectID).
					WhereIn("status", g.Slice{"reviewing", "failed"}).
					Update(g.Map{"status": "designing", "updated_at": rollbackNow})
			}
		}()
	}
	return nil
}

// Approve 通过计划版本。
func (s *PlanVersionService) Approve(ctx context.Context, planVersionID int64) error {
	now := gtime.Now()
	_, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("id", planVersionID).
		Where("review_status", consts.PlanReviewStatusPending).
		Update(g.Map{
			"review_status": consts.PlanReviewStatusApproved,
			"approved_at":   now,
			"updated_at":    now,
		})
	return err
}

// Reject 驳回计划版本。
func (s *PlanVersionService) Reject(ctx context.Context, planVersionID int64) error {
	now := gtime.Now()
	_, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("id", planVersionID).
		Where("review_status", consts.PlanReviewStatusPending).
		Update(g.Map{
			"review_status": consts.PlanReviewStatusRejected,
			"rejected_at":   now,
			"updated_at":    now,
		})
	return err
}

// GetBlueprintCount 获取版本下的蓝图数量。
func (s *PlanVersionService) GetBlueprintCount(ctx context.Context, planVersionID int64) int {
	count, _ := g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at").
		Count()
	return count
}

func defaultRoleType(rt string) string {
	if rt == "" {
		return consts.RoleTypeImplementer
	}
	return rt
}

func defaultRoleLevel(rl string) string {
	if rl == "" {
		return consts.RoleLevelPro
	}
	return rl
}
