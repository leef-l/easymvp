// Package plan 计划版本领域服务。
package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
func (s *PlanVersionService) CreateFromArchitectReply(
	ctx context.Context,
	projectID, workflowRunID, conversationID, messageID int64,
	tasks []engine.ArchitectTask,
) (int64, int, error) {
	if len(tasks) == 0 {
		return 0, 0, fmt.Errorf("没有可创建的任务蓝图")
	}

	now := time.Now()

	// 1. 获取下一个 version_no
	versionNo, err := s.planRepo.NextVersionNo(ctx, projectID)
	if err != nil {
		return 0, 0, fmt.Errorf("获取 version_no 失败: %w", err)
	}

	// 2. 将之前的活跃版本标记为 superseded
	if err := s.SupersedePreviousVersions(ctx, projectID, 0); err != nil {
		g.Log().Warningf(ctx, "[PlanVersionService] supersede 旧版本失败: %v", err)
	}

	// 3. 创建 plan_version
	pvID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_plan_version").Ctx(ctx).Insert(g.Map{
		"id":                     pvID,
		"project_id":             projectID,
		"workflow_run_id":        workflowRunID,
		"version_no":             versionNo,
		"source_conversation_id": conversationID,
		"source_message_id":      messageID,
		"status":                 consts.PlanVersionStatusDraft,
		"review_status":          consts.PlanReviewStatusPending,
		"summary":                fmt.Sprintf("第 %d 版方案，共 %d 个任务蓝图", versionNo, len(tasks)),
		"created_at":             now,
		"updated_at":             now,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("创建 plan_version 失败: %w", err)
	}

	// 4. 创建蓝图，构建 name -> blueprintID 映射
	nameToID := make(map[string]int64, len(tasks))
	blueprintIDs := make([]int64, 0, len(tasks))

	for i, task := range tasks {
		bpID := int64(snowflake.Generate())
		nameToID[task.Name] = bpID
		blueprintIDs = append(blueprintIDs, bpID)

		affectedJSON, _ := json.Marshal(task.AffectedResources)

		_, err = g.DB().Model("mvp_task_blueprint").Ctx(ctx).Insert(g.Map{
			"id":                  bpID,
			"plan_version_id":     pvID,
			"name":                task.Name,
			"description":         task.Description,
			"role_type":           defaultRoleType(task.RoleType),
			"role_level":          defaultRoleLevel(task.RoleLevel),
			"batch_no":            task.BatchNo,
			"sort":                i + 1,
			"affected_resources":  string(affectedJSON),
			"blueprint_status":    consts.BlueprintStatusDraft,
			"created_at":          now,
			"updated_at":          now,
		})
		if err != nil {
			return 0, 0, fmt.Errorf("创建蓝图(%s)失败: %w", task.Name, err)
		}
	}

	// 5. 第二遍：回写依赖关系 (depends_on_blueprint_ids)
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
			_, _ = g.DB().Model("mvp_task_blueprint").Ctx(ctx).
				Where("id", blueprintIDs[i]).
				Update(g.Map{"depends_on_blueprint_ids": string(depJSON)})
		}
	}

	// 6. 更新 workflow_run 的 active_plan_version_id
	if workflowRunID > 0 {
		_, _ = g.DB().Model("mvp_workflow_run").Ctx(ctx).
			Where("id", workflowRunID).
			Update(g.Map{"active_plan_version_id": pvID, "updated_at": now})
	}

	g.Log().Infof(ctx, "[PlanVersionService] 创建 plan_version=%d (v%d), %d 个蓝图, project=%d",
		pvID, versionNo, len(tasks), projectID)
	return pvID, len(tasks), nil
}

// SupersedePreviousVersions 将项目之前的活跃/草稿版本标记为 superseded。
func (s *PlanVersionService) SupersedePreviousVersions(ctx context.Context, projectID int64, exceptVersionID int64) error {
	now := gtime.Now()
	q := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{consts.PlanVersionStatusDraft, consts.PlanVersionStatusActive}).
		WhereNull("deleted_at")
	if exceptVersionID > 0 {
		q = q.WhereNot("id", exceptVersionID)
	}

	// 获取要 supersede 的版本 ID 列表
	oldIDs, err := q.Fields("id").Array()
	if err != nil || len(oldIDs) == 0 {
		return err
	}

	// supersede 版本
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

	// supersede 对应的蓝图
	_, err = g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		WhereIn("plan_version_id", idList).
		WhereNot("blueprint_status", consts.BlueprintStatusSuperseded).
		Update(g.Map{"blueprint_status": consts.BlueprintStatusSuperseded, "updated_at": now})
	return err
}

// SubmitForReview 提交当前草稿版本进入审核。
// 流程：plan_version draft→active, blueprints draft→confirmed, project status→reviewing。
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

	// 3. plan_version: draft → active
	_, err = g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("id", pvID).
		Where("status", consts.PlanVersionStatusDraft).
		Update(g.Map{"status": consts.PlanVersionStatusActive, "updated_at": now})
	if err != nil {
		return fmt.Errorf("更新 plan_version 状态失败: %w", err)
	}

	// 4. blueprints: draft → confirmed
	_, err = g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		Where("plan_version_id", pvID).
		Where("blueprint_status", consts.BlueprintStatusDraft).
		Update(g.Map{"blueprint_status": consts.BlueprintStatusConfirmed, "updated_at": now})
	if err != nil {
		return fmt.Errorf("确认蓝图状态失败: %w", err)
	}

	// 5. project status → reviewing
	_, err = g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		Update(g.Map{"status": "reviewing", "pause_reason": nil, "updated_at": now})
	if err != nil {
		return fmt.Errorf("更新项目状态失败: %w", err)
	}

	g.Log().Infof(ctx, "[PlanVersionService] SubmitForReview projectID=%d pvID=%d bpCount=%d", projectID, pvID, bpCount)

	// 触发 review stage（异步，不阻塞 API 响应）
	if s.reviewTrigger != nil {
		go func() {
			if err := s.reviewTrigger(context.Background(), projectID, pvID); err != nil {
				g.Log().Errorf(ctx, "[PlanVersionService] 触发审核失败: projectID=%d pvID=%d err=%v", projectID, pvID, err)
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
