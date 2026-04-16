// Package plan 计划版本领域服务。
package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// ReviewTrigger 审核触发回调，由 orchestrator 注册。
type ReviewTrigger func(ctx context.Context, projectID, planVersionID int64) error

// PlanVersionService 计划版本服务。
type PlanVersionService struct {
	planRepo        *repo.PlanVersionRepo
	blueprintRepo   *repo.BlueprintRepo
	projectRepo     *repo.ProjectRepo
	workflowRunRepo *repo.WorkflowRunRepo
	reviewTrigger   ReviewTrigger
}

// NewPlanVersionService 创建计划版本服务。
func NewPlanVersionService(pr *repo.PlanVersionRepo, br *repo.BlueprintRepo) *PlanVersionService {
	if pr == nil {
		pr = repo.NewPlanVersionRepo()
	}
	if br == nil {
		br = repo.NewBlueprintRepo()
	}
	return &PlanVersionService{
		planRepo:        pr,
		blueprintRepo:   br,
		projectRepo:     repo.NewProjectRepo(),
		workflowRunRepo: repo.NewWorkflowRunRepo(),
	}
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

	resolvedRoleLevels := make(map[string]string, len(tasks))

	now := time.Now()
	var versionNo int

	err := repo.WithTx(ctx, func(ctx context.Context, _ gdb.TX) error {
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
		err = s.planRepo.Insert(ctx, g.Map{
			"id":                     pvID,
			"project_id":             projectID,
			"workflow_run_id":        workflowRunID,
			"version_no":             versionNo,
			"source_conversation_id": conversationID,
			"source_message_id":      messageID,
			"status":                 consts.PlanVersionStatusDraft,
			"review_status":          consts.PlanReviewStatusPending,
			"summary":                fmt.Sprintf("第 %d 版方案，共 %d 个任务蓝图", versionNo, len(tasks)),
			"created_by":             scope.CreatedBy,
			"dept_id":                scope.DeptID,
			"created_at":             now,
			"updated_at":             now,
		})
		if err != nil {
			return fmt.Errorf("创建 plan_version 失败: %w", err)
		}

		// 4. 创建蓝图
		for i, task := range tasks {
			roleType := defaultRoleType(task.RoleType)
			roleLevel := resolveBlueprintRoleLevel(ctx, projectID, roleType, defaultRoleLevel(task.RoleLevel), resolvedRoleLevels)
			affectedJSON, jsonErr := json.Marshal(task.AffectedResources)
			if jsonErr != nil {
				return fmt.Errorf("序列化 affected_resources 失败: task=%s err=%w", task.Name, jsonErr)
			}
			err = s.blueprintRepo.Insert(ctx, g.Map{
				"id":                 blueprintIDs[i],
				"plan_version_id":    pvID,
				"name":               task.Name,
				"description":        task.Description,
				"role_type":          roleType,
				"role_level":         roleLevel,
				"batch_no":           task.BatchNo,
				"sort":               i + 1,
				"affected_resources": string(affectedJSON),
				"blueprint_status":   consts.BlueprintStatusDraft,
				"created_by":         scope.CreatedBy,
				"dept_id":            scope.DeptID,
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
				depJSON, depJSONErr := json.Marshal(depIDs)
				if depJSONErr != nil {
					return fmt.Errorf("序列化依赖ID失败: task=%s err=%w", task.Name, depJSONErr)
				}
				if err := s.blueprintRepo.UpdateFields(ctx, blueprintIDs[i], g.Map{"depends_on_blueprint_ids": string(depJSON)}); err != nil {
					return fmt.Errorf("回写蓝图依赖失败: %w", err)
				}
			}
		}

		// 6. 更新 workflow_run 的 active_plan_version_id（CAS：仅 designing/reviewing 状态可更新）
		if workflowRunID > 0 {
			wfRows, err := s.workflowRunRepo.UpdateFieldsIfStatuses(ctx, workflowRunID, []string{
				consts.WorkflowRunStatusDesigning,
				consts.WorkflowRunStatusReviewing,
			}, g.Map{"active_plan_version_id": pvID, "updated_at": now})
			if err != nil {
				return fmt.Errorf("更新 active_plan_version_id 失败: %w", err)
			}
			if wfRows == 0 {
				// workflow_run 不在 designing/reviewing 状态，跳过关联但不回滚蓝图创建
				g.Log().Warningf(ctx, "[PlanVersionService] workflow_run(%d) 不在可关联状态，跳过 active_plan_version_id 更新", workflowRunID)
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

// ApplyTaskPatchesFromArchitectReply 将架构师返回的 task_patches 回写到当前方案蓝图。
// 适用于 review warning 后仅修订单个或少量蓝图的场景，避免每次都重建整版 plan_version。
func (s *PlanVersionService) ApplyTaskPatchesFromArchitectReply(
	ctx context.Context,
	projectID, workflowRunID, conversationID, messageID int64,
	patches []engine.ArchitectTaskPatch,
) (int64, int, error) {
	if len(patches) == 0 {
		return 0, 0, fmt.Errorf("没有可回写的 task_patches")
	}

	pv, err := s.planRepo.GetLatestByProjectStatuses(ctx, projectID, []string{consts.PlanVersionStatusDraft, consts.PlanVersionStatusActive}, "id")
	if err != nil {
		return 0, 0, fmt.Errorf("查询当前方案版本失败: %w", err)
	}
	if pv == nil {
		return 0, 0, fmt.Errorf("项目 %d 没有可修订的方案版本", projectID)
	}

	pvID := gconv.Int64(pv["id"])
	now := time.Now()
	patchedCount := 0
	resolvedRoleLevels := make(map[string]string, len(patches))

	err = repo.WithTx(ctx, func(ctx context.Context, _ gdb.TX) error {
		blueprints, bpErr := s.blueprintRepo.ListByPlanVersionStatuses(ctx, pvID,
			[]string{consts.BlueprintStatusDraft, consts.BlueprintStatusConfirmed},
			"id", "name", "role_type", "role_level")
		if bpErr != nil {
			return fmt.Errorf("查询蓝图失败: %w", bpErr)
		}
		if len(blueprints) == 0 {
			return fmt.Errorf("方案版本 %d 没有可修订的蓝图", pvID)
		}

		byID := make(map[int64]gdb.Record, len(blueprints))
		byName := make(map[string]gdb.Record, len(blueprints))
		for _, bp := range blueprints {
			record := mapToRecord(bp)
			byID[record["id"].Int64()] = record
			name := strings.TrimSpace(record["name"].String())
			if name != "" {
				byName[name] = record
			}
		}

		for _, patch := range patches {
			target, targetErr := matchBlueprintForPatch(byID, byName, &patch)
			if targetErr != nil {
				return targetErr
			}

			updateData, updateErr := buildBlueprintPatchUpdateData(byName, &patch)
			if updateErr != nil {
				return updateErr
			}
			if len(updateData) == 0 {
				continue
			}
			roleType := strings.TrimSpace(target["role_type"].String())
			if value, ok := updateData["role_type"]; ok {
				roleType = strings.TrimSpace(fmt.Sprint(value))
			}
			if roleType != "" {
				roleLevel := strings.TrimSpace(target["role_level"].String())
				if value, ok := updateData["role_level"]; ok {
					roleLevel = strings.TrimSpace(fmt.Sprint(value))
				}
				updateData["role_level"] = resolveBlueprintRoleLevel(ctx, projectID, roleType, roleLevel, resolvedRoleLevels)
			}
			updateData["updated_at"] = now

			if upErr := s.blueprintRepo.UpdateFields(ctx, target["id"].Int64(), updateData); upErr != nil {
				return fmt.Errorf("更新蓝图 %d 失败: %w", target["id"].Int64(), upErr)
			}
			patchedCount++
		}

		if patchedCount == 0 {
			return fmt.Errorf("task_patches 未产生实际变更")
		}

		if upErr := s.planRepo.UpdateFields(ctx, pvID, g.Map{
			"workflow_run_id":        workflowRunID,
			"source_conversation_id": conversationID,
			"source_message_id":      messageID,
			"updated_at":             now,
		}); upErr != nil {
			return fmt.Errorf("更新方案版本来源失败: %w", upErr)
		}

		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	g.Log().Infof(ctx, "[PlanVersionService] 应用 task_patches: project=%d planVersion=%d patched=%d", projectID, pvID, patchedCount)
	return pvID, patchedCount, nil
}

func matchBlueprintForPatch(byID map[int64]gdb.Record, byName map[string]gdb.Record, patch *engine.ArchitectTaskPatch) (gdb.Record, error) {
	if patch == nil {
		return nil, fmt.Errorf("patch 不能为空")
	}
	if patch.BlueprintID > 0 {
		if bp, ok := byID[patch.BlueprintID]; ok {
			return bp, nil
		}
		return nil, fmt.Errorf("未找到 blueprint_id=%d 对应的蓝图", patch.BlueprintID)
	}

	taskName := strings.TrimSpace(patch.TaskName)
	if taskName == "" {
		return nil, fmt.Errorf("task_patch 缺少 task_name")
	}
	if bp, ok := byName[taskName]; ok {
		return bp, nil
	}
	return nil, fmt.Errorf("未找到任务名为 %q 的蓝图", taskName)
}

func buildBlueprintPatchUpdateData(byName map[string]gdb.Record, patch *engine.ArchitectTaskPatch) (g.Map, error) {
	updateData := g.Map{}
	if patch == nil {
		return updateData, nil
	}
	if desc := strings.TrimSpace(patch.Description); desc != "" {
		updateData["description"] = desc
	}
	if roleType := strings.TrimSpace(patch.RoleType); roleType != "" {
		updateData["role_type"] = roleType
	}
	if roleLevel := strings.TrimSpace(patch.RoleLevel); roleLevel != "" {
		updateData["role_level"] = roleLevel
	}
	if patch.BatchNo != nil && *patch.BatchNo > 0 {
		updateData["batch_no"] = *patch.BatchNo
	}
	if patch.Sort != nil && *patch.Sort > 0 {
		updateData["sort"] = *patch.Sort
	}
	if patch.AffectedResources != nil {
		affectedJSON, err := json.Marshal(patch.AffectedResources)
		if err != nil {
			return nil, fmt.Errorf("序列化 affected_resources 失败: %w", err)
		}
		updateData["affected_resources"] = string(affectedJSON)
	}
	if patch.DependsOn != nil {
		depIDs := make([]int64, 0, len(patch.DependsOn))
		for _, depName := range patch.DependsOn {
			depName = strings.TrimSpace(depName)
			if depName == "" {
				continue
			}
			target, ok := byName[depName]
			if !ok {
				return nil, fmt.Errorf("未找到依赖蓝图 %q", depName)
			}
			depIDs = append(depIDs, target["id"].Int64())
		}
		depJSON, err := json.Marshal(depIDs)
		if err != nil {
			return nil, fmt.Errorf("序列化 depends_on_blueprint_ids 失败: %w", err)
		}
		updateData["depends_on_blueprint_ids"] = string(depJSON)
	}
	return updateData, nil
}

func mapToRecord(data g.Map) gdb.Record {
	record := make(gdb.Record, len(data))
	for key, value := range data {
		record[key] = g.NewVar(value)
	}
	return record
}

func resolveBlueprintRoleLevel(ctx context.Context, projectID int64, roleType string, requestedLevel string, cache map[string]string) string {
	roleType = defaultRoleType(roleType)
	requestedLevel = defaultRoleLevel(requestedLevel)
	cacheKey := roleType + "/" + requestedLevel
	if cache != nil {
		if level, ok := cache[cacheKey]; ok && strings.TrimSpace(level) != "" {
			return level
		}
	}

	roleRecord, err := repo.GetProjectRoleByLevel(ctx, projectID, roleType, requestedLevel)
	if err != nil || roleRecord == nil {
		if cache != nil {
			cache[cacheKey] = requestedLevel
		}
		return requestedLevel
	}
	resolvedLevel := strings.TrimSpace(roleRecord["role_level"].String())
	if resolvedLevel == "" {
		resolvedLevel = requestedLevel
	}
	if cache != nil {
		cache[cacheKey] = resolvedLevel
	}
	return resolvedLevel
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
// 使用单次 UPDATE + 状态 CAS 条件，避免查询-更新间的并发风险。
func (s *PlanVersionService) doSupersede(ctx context.Context, projectID int64, exceptVersionID int64) error {
	now := gtime.Now()

	// 先查出受影响的 ID 列表（供蓝图级联更新使用）
	records, err := s.planRepo.ListByProjectStatuses(ctx, projectID, []string{consts.PlanVersionStatusDraft, consts.PlanVersionStatusActive}, "id")
	if err != nil || len(records) == 0 {
		return err
	}
	idList := make([]int64, 0, len(records))
	for _, record := range records {
		id := record["id"].Int64()
		if exceptVersionID > 0 && id == exceptVersionID {
			continue
		}
		idList = append(idList, id)
	}
	if len(idList) == 0 {
		return nil
	}

	// plan_version 状态更新（CAS：仅更新 draft/active 状态的记录）
	if err = s.planRepo.UpdateByIDs(ctx, idList, g.Map{"status": consts.PlanVersionStatusSuperseded, "updated_at": now}); err != nil {
		return err
	}

	// 级联更新蓝图状态
	return s.blueprintRepo.UpdateByPlanVersionIDs(ctx, idList, g.Map{"blueprint_status": consts.BlueprintStatusSuperseded, "updated_at": now})
}

// SubmitForReview 提交当前草稿版本进入审核。
// 流程：plan_version draft→active, blueprints draft→confirmed, project status→reviewing, 触发 review stage。
// 状态变更在事务中完成；reviewTrigger 同步调用，失败则回滚全部状态。
func (s *PlanVersionService) SubmitForReview(ctx context.Context, projectID int64) error {
	now := gtime.Now()

	// 1. 找最新的 draft plan_version
	pv, err := s.planRepo.GetLatestByProjectStatuses(ctx, projectID, []string{consts.PlanVersionStatusDraft}, "id")
	if err != nil || pv == nil {
		return fmt.Errorf("没有待确认的方案版本")
	}
	pvID := gconv.Int64(pv["id"])

	// 2. 检查蓝图数
	bpCount, bpCountErr := s.blueprintRepo.CountByPlanVersion(ctx, pvID)
	if bpCountErr != nil {
		return fmt.Errorf("查询蓝图数失败: %w", bpCountErr)
	}
	if bpCount == 0 {
		return fmt.Errorf("方案版本没有任务蓝图")
	}

	// 3. 事务内完成所有状态迁移
	err = repo.WithTx(ctx, func(ctx context.Context, _ gdb.TX) error {
		// plan_version: draft → active（CAS）
		pvRows, err := s.planRepo.UpdateFieldsIfStatuses(ctx, pvID, []string{consts.PlanVersionStatusDraft}, g.Map{
			"status":     consts.PlanVersionStatusActive,
			"updated_at": now,
		})
		if err != nil {
			return fmt.Errorf("更新 plan_version 状态失败: %w", err)
		}
		if pvRows == 0 {
			return fmt.Errorf("plan_version(%d) 已不在 draft 状态，无法提交审核", pvID)
		}

		// blueprints: draft → confirmed
		bpRows, err := s.blueprintRepo.UpdateByPlanVersionStatuses(ctx, pvID, []string{consts.BlueprintStatusDraft}, g.Map{
			"blueprint_status": consts.BlueprintStatusConfirmed,
			"updated_at":       now,
		})
		if err != nil {
			return fmt.Errorf("确认蓝图状态失败: %w", err)
		}
		if bpRows == 0 {
			return fmt.Errorf("plan_version(%d) 下没有 draft 蓝图可确认", pvID)
		}

		// project status → reviewing
		if err := s.projectRepo.UpdateFields(ctx, projectID, g.Map{
			"status":       "reviewing",
			"pause_reason": nil,
			"updated_at":   now,
		}); err != nil {
			return fmt.Errorf("更新项目状态失败: %w", err)
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
			if err := s.planRepo.UpdateFields(ctx, pvID, g.Map{"status": consts.PlanVersionStatusDraft, "updated_at": gtime.Now()}); err != nil {
				g.Log().Errorf(ctx, "[PlanVersionService] 回滚 plan_version 失败: pv=%d err=%v", pvID, err)
			}
			if rbErr := s.blueprintRepo.UpdateStatusByPlanVersion(ctx, pvID, consts.BlueprintStatusConfirmed, g.Map{
				"blueprint_status": consts.BlueprintStatusDraft,
				"updated_at":       gtime.Now(),
			}); rbErr != nil {
				g.Log().Errorf(ctx, "[PlanVersionService] 回滚 blueprints 失败: pv=%d err=%v", pvID, rbErr)
			}
			if _, rbErr := s.projectRepo.UpdateFieldsIfStatuses(ctx, projectID, []string{"reviewing"}, g.Map{
				"status":     "designing",
				"updated_at": gtime.Now(),
			}); rbErr != nil {
				g.Log().Errorf(ctx, "[PlanVersionService] 回滚 project 失败: project=%d err=%v", projectID, rbErr)
			}
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
	project, psErr := s.projectRepo.GetByID(ctx, projectID, "status")
	if psErr != nil {
		return fmt.Errorf("查询项目状态失败: %w", psErr)
	}
	if gconv.String(project["status"]) == "reviewing" {
		return fmt.Errorf("方案已在审核中，请勿重复提交")
	}

	// 1. 找最新的 draft plan_version
	pv, err := s.planRepo.GetLatestByProjectStatuses(ctx, projectID, []string{consts.PlanVersionStatusDraft}, "id")
	if err != nil || pv == nil {
		return fmt.Errorf("没有待确认的方案版本")
	}
	pvID := gconv.Int64(pv["id"])

	// 2. 检查蓝图数
	bpCount, bpCountErr := s.blueprintRepo.CountByPlanVersion(ctx, pvID)
	if bpCountErr != nil {
		return fmt.Errorf("查询蓝图数失败: %w", bpCountErr)
	}
	if bpCount == 0 {
		return fmt.Errorf("方案版本没有任务蓝图")
	}

	// 3. 事务内完成所有状态迁移（同步，快速）
	err = repo.WithTx(ctx, func(ctx context.Context, _ gdb.TX) error {
		pvRows, err := s.planRepo.UpdateFieldsIfStatuses(ctx, pvID, []string{consts.PlanVersionStatusDraft}, g.Map{
			"status":     consts.PlanVersionStatusActive,
			"updated_at": now,
		})
		if err != nil {
			return fmt.Errorf("更新 plan_version 状态失败: %w", err)
		}
		if pvRows == 0 {
			return fmt.Errorf("plan_version(%d) 已不在 draft 状态，无法提交审核", pvID)
		}

		bpRows, err := s.blueprintRepo.UpdateByPlanVersionStatuses(ctx, pvID, []string{consts.BlueprintStatusDraft}, g.Map{
			"blueprint_status": consts.BlueprintStatusConfirmed,
			"updated_at":       now,
		})
		if err != nil {
			return fmt.Errorf("确认蓝图状态失败: %w", err)
		}
		if bpRows == 0 {
			return fmt.Errorf("plan_version(%d) 下没有 draft 蓝图可确认", pvID)
		}

		if err := s.projectRepo.UpdateFields(ctx, projectID, g.Map{
			"status":       "reviewing",
			"pause_reason": nil,
			"updated_at":   now,
		}); err != nil {
			return fmt.Errorf("更新项目状态失败: %w", err)
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
				if err := s.planRepo.UpdateFields(bgCtx, pvID, g.Map{"status": consts.PlanVersionStatusDraft, "updated_at": rollbackNow}); err != nil {
					g.Log().Errorf(bgCtx, "[PlanVersionService] 异步回滚 plan_version 失败: pv=%d err=%v", pvID, err)
				}
				if rbErr := s.blueprintRepo.UpdateStatusByPlanVersion(bgCtx, pvID, consts.BlueprintStatusConfirmed, g.Map{
					"blueprint_status": consts.BlueprintStatusDraft,
					"updated_at":       rollbackNow,
				}); rbErr != nil {
					g.Log().Errorf(bgCtx, "[PlanVersionService] 异步回滚 blueprints 失败: pv=%d err=%v", pvID, rbErr)
				}

				// 查找关联的 workflow_run 并回退状态
				wfRun, wfErr := s.workflowRunRepo.GetLatestByProjectStatuses(bgCtx, projectID, []string{"reviewing", "failed"}, "id")
				if wfErr != nil {
					g.Log().Errorf(bgCtx, "[PlanVersionService] 异步回滚查询 workflow_run 失败: project=%d err=%v", projectID, wfErr)
				} else if wfRun != nil {
					if _, rbErr := s.workflowRunRepo.UpdateFieldsIfStatuses(bgCtx, gconv.Int64(wfRun["id"]), []string{"reviewing", "failed"}, g.Map{
						"status":        "designing",
						"current_stage": "design",
						"updated_at":    rollbackNow,
					}); rbErr != nil {
						g.Log().Errorf(bgCtx, "[PlanVersionService] 异步回滚 workflow_run 失败: wfRun=%d err=%v", gconv.Int64(wfRun["id"]), rbErr)
					}
				}

				if _, rbErr := s.projectRepo.UpdateFieldsIfStatuses(bgCtx, projectID, []string{"reviewing", "failed"}, g.Map{
					"status":     "designing",
					"updated_at": rollbackNow,
				}); rbErr != nil {
					g.Log().Errorf(bgCtx, "[PlanVersionService] 异步回滚 project 失败: project=%d err=%v", projectID, rbErr)
				}
			}
		}()
	}
	return nil
}

// Approve 通过计划版本。
func (s *PlanVersionService) Approve(ctx context.Context, planVersionID int64) error {
	now := gtime.Now()
	_, err := s.planRepo.UpdateFieldsIfReviewStatus(ctx, planVersionID, consts.PlanReviewStatusPending, g.Map{
		"review_status": consts.PlanReviewStatusApproved,
		"approved_at":   now,
		"updated_at":    now,
	})
	return err
}

// Reject 驳回计划版本。
func (s *PlanVersionService) Reject(ctx context.Context, planVersionID int64) error {
	now := gtime.Now()
	_, err := s.planRepo.UpdateFieldsIfReviewStatus(ctx, planVersionID, consts.PlanReviewStatusPending, g.Map{
		"review_status": consts.PlanReviewStatusRejected,
		"rejected_at":   now,
		"updated_at":    now,
	})
	return err
}

// GetBlueprintCount 获取版本下的蓝图数量。
func (s *PlanVersionService) GetBlueprintCount(ctx context.Context, planVersionID int64) int {
	count, err := s.blueprintRepo.CountByPlanVersion(ctx, planVersionID)
	if err != nil {
		g.Log().Warningf(ctx, "[PlanVersion] 查询蓝图数失败: pvID=%d err=%v", planVersionID, err)
	}
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
