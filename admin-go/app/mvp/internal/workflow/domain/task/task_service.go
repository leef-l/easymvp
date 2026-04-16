// Package task 领域任务服务。
package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/contract"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// TaskService 领域任务服务。
type TaskService struct {
	taskRepo *repo.DomainTaskRepo
}

// NewTaskService 创建领域任务服务。
func NewTaskService(tr *repo.DomainTaskRepo) *TaskService {
	return &TaskService{taskRepo: tr}
}

// InstantiateFromBlueprint 将蓝图实例化为领域任务。
// 从 plan_version 下的 confirmed 蓝图创建 mvp_domain_task 记录。
// 返回创建的任务数量。
func (s *TaskService) InstantiateFromBlueprint(ctx context.Context, planVersionID int64, stageRunID int64, workflowRunID int64) (int, error) {
	now := time.Now()

	// 1. 查询所有已确认的蓝图
	blueprints, err := repo.NewBlueprintRepo().ListByPlanVersionStatuses(ctx, planVersionID, []string{"confirmed"},
		"id", "name", "description", "role_type", "role_level", "batch_no", "sort", "affected_resources", "depends_on_blueprint_ids")
	if err != nil {
		return 0, fmt.Errorf("查询蓝图失败: %w", err)
	}
	if len(blueprints) == 0 {
		return 0, fmt.Errorf("没有已确认的蓝图")
	}

	// 查项目 ID（从 workflow_run 获取）
	workflowRun, pidErr := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "project_id")
	if pidErr != nil {
		return 0, fmt.Errorf("查询 workflow_run 的 project_id 失败: %w", pidErr)
	}
	projectID := g.NewVar(workflowRun["project_id"]).Int64()
	projectContract, contractErr := contract.Load(ctx, projectID)
	if contractErr != nil {
		g.Log().Warningf(ctx, "[TaskService] 加载项目级硬约束失败: projectID=%d err=%v", projectID, contractErr)
		projectContract = &contract.ProjectContract{}
	}

	// 查项目角色配置（获取 execution_mode 和 model_id），缺失的自动从默认预设补齐
	roleConfigs, rcErr := repo.GetProjectRolesMap(ctx, projectID)
	if rcErr != nil {
		g.Log().Warningf(ctx, "[TaskService] 查询项目角色配置失败: projectID=%d err=%v", projectID, rcErr)
	}

	// 2. 获取项目归属字段（继承到领域任务）
	scope := repo.GetProjectScopeByWorkflowRun(ctx, workflowRunID)

	// 3. 构建 blueprintID → domainTaskID 映射（用于依赖关系转换）
	bpIDToTaskID := make(map[int64]int64, len(blueprints))
	taskInserts := make([]g.Map, 0, len(blueprints))

	for _, bp := range blueprints {
		taskID := int64(snowflake.Generate())
		bpID := g.NewVar(bp["id"]).Int64()
		bpIDToTaskID[bpID] = taskID

		roleType := g.NewVar(bp["role_type"]).String()
		roleLevel := g.NewVar(bp["role_level"]).String()
		roleKey := roleType + "/" + roleLevel

		// 从角色配置获取 execution_mode 和 model_id
		executionMode := "chat"
		var modelID int64
		resolvedExactRole := false
		if rc, ok := roleConfigs[roleKey]; ok {
			resolvedExactRole = true
			if em, ok := rc["execution_mode"].(string); ok && em != "" {
				executionMode = em
			}
			if mid, ok := rc["model_id"].(int64); ok {
				modelID = mid
			}
		}
		if !resolvedExactRole || modelID == 0 {
			roleRecord, roleErr := repo.GetProjectRoleByLevel(ctx, projectID, roleType, roleLevel)
			if roleErr != nil {
				g.Log().Warningf(ctx, "[TaskService] 解析角色回退配置失败: projectID=%d role=%s/%s err=%v",
					projectID, roleType, roleLevel, roleErr)
			} else if roleRecord != nil {
				if resolvedLevel := roleRecord["role_level"].String(); resolvedLevel != "" {
					roleLevel = resolvedLevel
					roleKey = roleType + "/" + roleLevel
				}
				if !resolvedExactRole {
					if em := roleRecord["execution_mode"].String(); em != "" {
						executionMode = em
					}
				}
				if modelID == 0 {
					modelID = roleRecord["model_id"].Int64()
				}
			}
		}

		taskInserts = append(taskInserts, g.Map{
			"id":                 taskID,
			"workflow_run_id":    workflowRunID,
			"stage_run_id":       stageRunID,
			"plan_version_id":    planVersionID,
			"blueprint_id":       bpID,
			"task_kind":          "implement",
			"name":               g.NewVar(bp["name"]).String(),
			"description":        contract.AppendConstraintBlock(g.NewVar(bp["description"]).String(), projectContract),
			"role_type":          roleType,
			"role_level":         roleLevel,
			"execution_mode":     executionMode,
			"status":             StatusPending,
			"model_id":           modelID,
			"batch_no":           g.NewVar(bp["batch_no"]).Int(),
			"sort":               g.NewVar(bp["sort"]).Int(),
			"affected_resources": g.NewVar(bp["affected_resources"]).String(),
			"created_by":         scope.CreatedBy,
			"dept_id":            scope.DeptID,
			"created_at":         now,
			"updated_at":         now,
		})
	}

	// 3. 批量插入领域任务
	if err = s.taskRepo.BatchCreate(ctx, taskInserts); err != nil {
		return 0, fmt.Errorf("批量创建领域任务失败: %w", err)
	}

	// 4. 回写依赖关系：将蓝图的 depends_on_blueprint_ids 转换为 domain_task 之间的依赖
	// 存在 parent_task_id 字段用于简单依赖，复杂依赖用独立查询
	for _, bp := range blueprints {
		var depBpIDs []int64
		depJSON := g.NewVar(bp["depends_on_blueprint_ids"]).String()
		if depJSON == "" || depJSON == "[]" || depJSON == "null" {
			continue
		}
		if err := json.Unmarshal([]byte(depJSON), &depBpIDs); err != nil {
			continue
		}
		if len(depBpIDs) == 0 {
			continue
		}

		// 转换为 domain_task ID
		depTaskIDs := make([]int64, 0, len(depBpIDs))
		for _, depBpID := range depBpIDs {
			if depTaskID, ok := bpIDToTaskID[depBpID]; ok {
				depTaskIDs = append(depTaskIDs, depTaskID)
			}
		}
		if len(depTaskIDs) == 0 {
			continue
		}

		// 完整依赖列表写入 depends_on_task_ids，parent_task_id 保留第一个（向后兼容）
		taskID := bpIDToTaskID[g.NewVar(bp["id"]).Int64()]
		depJSON2, jsonErr := json.Marshal(depTaskIDs)
		if jsonErr != nil {
			g.Log().Errorf(ctx, "[TaskService] 序列化依赖任务ID失败: task=%d err=%v", taskID, jsonErr)
			continue
		}
		if upErr := s.taskRepo.UpdateFields(ctx, taskID, g.Map{
			"parent_task_id":      depTaskIDs[0],
			"depends_on_task_ids": string(depJSON2),
			"updated_at":          now,
		}); upErr != nil {
			g.Log().Errorf(ctx, "[TaskService] 更新任务依赖关系失败: task=%d err=%v", taskID, upErr)
		}
	}

	g.Log().Infof(ctx, "[TaskService] InstantiateFromBlueprint planVersionID=%d created=%d tasks, workflowRunID=%d",
		planVersionID, len(taskInserts), workflowRunID)
	return len(taskInserts), nil
}

// UpdateStatus CAS 更新任务状态。
func (s *TaskService) UpdateStatus(ctx context.Context, taskID int64, from, to string, extra g.Map) (int64, error) {
	return s.taskRepo.UpdateStatus(ctx, taskID, from, to, extra)
}

// Retry 重试失败任务（failed → pending）。
func (s *TaskService) Retry(ctx context.Context, taskID int64) error {
	now := gtime.Now()
	rows, err := s.taskRepo.UpdateStatus(ctx, taskID, StatusFailed, StatusPending, g.Map{
		"retry_count": gdb.Raw("retry_count + 1"),
		"updated_at":  now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("任务(%d) 不在 failed 状态", taskID)
	}
	return nil
}

// Skip 跳过任务（pending/failed → completed，标记为跳过）。
func (s *TaskService) Skip(ctx context.Context, taskID int64) error {
	now := gtime.Now()

	// 尝试从 pending 跳过
	rows, err := s.taskRepo.UpdateStatus(ctx, taskID, StatusPending, StatusCompleted, g.Map{
		"result":       "skipped",
		"completed_at": now,
		"updated_at":   now,
	})
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	// 尝试从 failed 跳过
	rows, err = s.taskRepo.UpdateStatus(ctx, taskID, StatusFailed, StatusCompleted, g.Map{
		"result":       "skipped",
		"completed_at": now,
		"updated_at":   now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("任务(%d) 不在可跳过的状态", taskID)
	}
	return nil
}

// Escalate 升级任务（failed → escalated）。
func (s *TaskService) Escalate(ctx context.Context, taskID int64) error {
	now := gtime.Now()
	rows, err := s.taskRepo.UpdateStatus(ctx, taskID, StatusFailed, StatusEscalated, g.Map{"updated_at": now})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("任务(%d) 不在 failed 状态", taskID)
	}
	return nil
}

// GetPendingByBatch 获取指定批次的 pending 任务。
func (s *TaskService) GetPendingByBatch(ctx context.Context, workflowRunID int64, batchNo int) ([]map[string]interface{}, error) {
	tasks, err := s.taskRepo.ListByWorkflowFiltered(ctx, workflowRunID, StatusPending, batchNo)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, t)
	}
	return result, nil
}
