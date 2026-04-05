// Package accept 管理自动验收阶段：证据收集 → 规则评估 → 裁决归并 → 决定走向。
package accept

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/workflow/acceptance"
	"easymvp/app/mvp/internal/workflow/repo"
)

// ErrManualReviewRequired reworkTrigger 无法自动返工时返回此错误，
// 通知 accept service 保持 accept stage running 等待人工介入。
var ErrManualReviewRequired = errors.New("manual review required: no task to rework")

// StageCompleter 阶段操作回调（避免循环依赖）。
type StageCompleter interface {
	CompleteStage(ctx context.Context, stageRunID int64) error
	FailStage(ctx context.Context, stageRunID int64, reason string) error
}

// ReworkTriggerFn 触发返工回调。
type ReworkTriggerFn func(ctx context.Context, workflowRunID int64, acceptRunID int64, issues []acceptance.RuleHit) error

// CompleteTriggerFn 触发完成回调（accept passed → complete）。
type CompleteTriggerFn func(ctx context.Context, workflowRunID int64) error

// Service 验收阶段服务。
type Service struct {
	acceptRunRepo *repo.AcceptRunRepo
	collector     *acceptance.EvidenceCollector
	ruleEngine    *acceptance.RuleEngine
	reducer       *acceptance.DecisionReducer
	stageCompleter StageCompleter
	reworkTrigger  ReworkTriggerFn
	completeTrigger CompleteTriggerFn
}

// NewService 创建验收阶段服务。
func NewService(
	acceptRunRepo *repo.AcceptRunRepo,
	collector *acceptance.EvidenceCollector,
	ruleEngine *acceptance.RuleEngine,
	reducer *acceptance.DecisionReducer,
) *Service {
	return &Service{
		acceptRunRepo: acceptRunRepo,
		collector:     collector,
		ruleEngine:    ruleEngine,
		reducer:       reducer,
	}
}

// SetStageCompleter 注册阶段完成回调。
func (s *Service) SetStageCompleter(sc StageCompleter) { s.stageCompleter = sc }

// SetReworkTrigger 注册返工触发回调。
func (s *Service) SetReworkTrigger(fn ReworkTriggerFn) { s.reworkTrigger = fn }

// SetCompleteTrigger 注册完成触发回调。
func (s *Service) SetCompleteTrigger(fn CompleteTriggerFn) { s.completeTrigger = fn }

// Run 运行验收流程。
// 主编排：创建 accept_run → 收集证据 → 规则评估 → 裁决 → 决定走向。
func (s *Service) Run(ctx context.Context, workflowRunID, stageRunID int64) error {
	now := gtime.Now()

	// 1. 查项目信息
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).WhereNull("deleted_at").One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}
	projectID := wfRun["project_id"].Int64()

	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("project(%d) 不存在", projectID)
	}
	projectType := project["project_category"].String()
	workDir := project["work_dir"].String()
	createdBy := project["created_by"].Int64()
	deptID := project["dept_id"].Int64()

	// 2. 幂等检查：同一 stageRun 不重复创建 accept_run
	existing, _ := g.DB().Model("mvp_accept_run").Ctx(ctx).
		Where("stage_run_id", stageRunID).
		WhereIn("status", g.Slice{"running"}).
		WhereNull("deleted_at").
		Count()
	if existing > 0 {
		g.Log().Warningf(ctx, "[AcceptStage] stageRun=%d 已有运行中的 accept_run，跳过重复创建", stageRunID)
		return nil
	}

	// 3. 获取验收轮次
	round, err := s.acceptRunRepo.GetNextRound(ctx, workflowRunID)
	if err != nil {
		round = 1
	}

	// 4. 创建 accept_run
	acceptRunID, err := s.acceptRunRepo.Create(ctx, g.Map{
		"workflow_run_id": workflowRunID,
		"stage_run_id":    stageRunID,
		"project_id":      projectID,
		"accept_round":    round,
		"status":          "running",
		"rules_version":   "v1.0.0",
		"created_by":      createdBy,
		"dept_id":         deptID,
		"started_at":      now,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return fmt.Errorf("创建 accept_run 失败: %w", err)
	}

	acceptCtx := &acceptance.AcceptContext{
		WorkflowRunID: workflowRunID,
		ProjectID:     projectID,
		AcceptRunID:   acceptRunID,
		StageRunID:    stageRunID,
		ProjectType:   projectType,
		WorkDir:       workDir,
		CreatedBy:     createdBy,
		DeptID:        deptID,
	}

	g.Log().Infof(ctx, "[AcceptStage] 开始验收: workflowRunID=%d stageRunID=%d acceptRunID=%d round=%d projectType=%s",
		workflowRunID, stageRunID, acceptRunID, round, projectType)

	// 4. 收集证据
	evidence, evidenceErr := s.collector.Collect(ctx, acceptCtx)
	if evidenceErr != nil {
		g.Log().Warningf(ctx, "[AcceptStage] 证据收集部分失败（不阻塞）: %v", evidenceErr)
	}

	// 5. 规则评估
	rulesSnapshot, hits, ruleErr := s.ruleEngine.LoadAndEvaluate(ctx, acceptCtx)
	if ruleErr != nil {
		// 规则引擎异常 → accept 自身失败（降级路径）
		s.failAcceptRun(ctx, acceptRunID, ruleErr.Error())
		if s.stageCompleter != nil {
			_ = s.stageCompleter.FailStage(ctx, stageRunID, "规则引擎异常: "+ruleErr.Error())
		}
		return nil
	}

	// 写入规则快照
	_, _ = g.DB().Model("mvp_accept_run").Ctx(ctx).
		Where("id", acceptRunID).
		Data(g.Map{"rules_snapshot_ref": rulesSnapshot, "updated_at": gtime.Now()}).
		Update()

	// 6. 裁决归并
	decision, reduceErr := s.reducer.Reduce(ctx, acceptCtx, hits, evidence)
	if reduceErr != nil {
		s.failAcceptRun(ctx, acceptRunID, reduceErr.Error())
		if s.stageCompleter != nil {
			_ = s.stageCompleter.FailStage(ctx, stageRunID, "裁决归并异常: "+reduceErr.Error())
		}
		return nil
	}

	// 7. 写入决策
	if err := s.acceptRunRepo.UpdateDecision(ctx, acceptRunID, decision.Decision, decision.Score, decision.Summary); err != nil {
		g.Log().Warningf(ctx, "[AcceptStage] 写入决策失败: %v", err)
	}
	// 更新 accept_run 为 completed
	_, _ = s.acceptRunRepo.UpdateStatus(ctx, acceptRunID, "running", "completed", g.Map{})

	g.Log().Infof(ctx, "[AcceptStage] 验收完成: acceptRunID=%d decision=%s score=%.1f",
		acceptRunID, decision.Decision, decision.Score)

	// 8. 决定走向
	switch decision.Decision {
	case acceptance.DecisionPassed:
		// 完成 accept stage → 推进到 complete
		if s.stageCompleter != nil {
			if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
				g.Log().Errorf(ctx, "[AcceptStage] CompleteStage 失败: %v", err)
				return err
			}
		}
		if s.completeTrigger != nil {
			if err := s.completeTrigger(ctx, workflowRunID); err != nil {
				g.Log().Errorf(ctx, "[AcceptStage] 推进 complete 失败: %v", err)
				return err
			}
		}

	case acceptance.DecisionFailed:
		// 先尝试触发 rework
		if s.reworkTrigger != nil {
			if err := s.reworkTrigger(ctx, workflowRunID, acceptRunID, decision.Issues); err != nil {
				if errors.Is(err, ErrManualReviewRequired) {
					// 无法自动返工 → 保持 accept stage running，等待人工介入
					g.Log().Infof(ctx, "[AcceptStage] 无法自动返工，保持 accept running 等待人工: acceptRunID=%d", acceptRunID)
					return nil
				}
				g.Log().Errorf(ctx, "[AcceptStage] 触发返工失败: %v", err)
				return err
			}
		}
		// rework 已成功触发 → 完成 accept stage
		if s.stageCompleter != nil {
			if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
				g.Log().Errorf(ctx, "[AcceptStage] CompleteStage 失败: %v", err)
				return err
			}
		}

	case acceptance.DecisionManualReview:
		// 保持 accept stage running，等待人工介入
		g.Log().Infof(ctx, "[AcceptStage] 需要人工审核: acceptRunID=%d", acceptRunID)
	}

	return nil
}

// failAcceptRun 标记 accept_run 为失败。
func (s *Service) failAcceptRun(ctx context.Context, acceptRunID int64, reason string) {
	_, _ = s.acceptRunRepo.UpdateStatus(ctx, acceptRunID, "running", "failed", g.Map{
		"summary": reason,
	})
}

// GetLatestIssues 获取最近一次验收的问题列表（供返工输入包使用）。
func (s *Service) GetLatestIssues(ctx context.Context, acceptRunID int64) ([]acceptance.RuleHit, error) {
	issueRepo := repo.NewAcceptIssueRepo()
	records, err := issueRepo.ListByAcceptRun(ctx, acceptRunID)
	if err != nil {
		return nil, err
	}

	var hits []acceptance.RuleHit
	for _, r := range records {
		hits = append(hits, acceptance.RuleHit{
			RuleCode:        r["rule_code"].(string),
			Severity:        r["severity"].(string),
			Title:           r["title"].(string),
			Detail:          fmt.Sprintf("%v", r["detail"]),
			ExpectedValue:   fmt.Sprintf("%v", r["expected_value"]),
			ActualValue:     fmt.Sprintf("%v", r["actual_value"]),
			SuggestedAction: fmt.Sprintf("%v", r["suggested_action"]),
		})
	}
	return hits, nil
}

// BuildReworkInput 构建返工输入包 JSON。
func BuildReworkInput(acceptRunID int64, issues []acceptance.RuleHit) string {
	input := map[string]interface{}{
		"accept_run_id": acceptRunID,
		"issues":        issues,
	}
	data, _ := json.Marshal(input)
	return string(data)
}

// ManualApprove 人工放行：将最新 accept_run 标记为 passed，推进到 complete。
func (s *Service) ManualApprove(ctx context.Context, projectID int64, reason string) error {
	wfRun, acceptRun, stageRunID, err := s.findActiveAcceptRun(ctx, projectID)
	if err != nil {
		return err
	}
	acceptRunID := gconv.Int64(acceptRun["id"])
	workflowRunID := gconv.Int64(wfRun["id"])

	// CAS: 只有 completed 状态的 accept_run 允许人工操作
	arStatus := gconv.String(acceptRun["status"])
	if arStatus != "completed" && arStatus != "running" {
		return fmt.Errorf("accept_run(%d) 状态为 %s，不允许人工放行", acceptRunID, arStatus)
	}

	// 更新 accept_run 决策
	if err := s.acceptRunRepo.UpdateDecision(ctx, acceptRunID, acceptance.DecisionPassed, 100, "人工放行: "+reason); err != nil {
		return fmt.Errorf("更新决策失败: %w", err)
	}
	if arStatus == "running" {
		_, _ = s.acceptRunRepo.UpdateStatus(ctx, acceptRunID, "running", "completed", g.Map{
			"finished_at": gtime.Now(),
		})
	}

	g.Log().Infof(ctx, "[AcceptStage] 人工放行: acceptRunID=%d project=%d reason=%s", acceptRunID, projectID, reason)

	// 完成 accept stage → 推进到 complete
	if s.stageCompleter != nil {
		if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
			return fmt.Errorf("CompleteStage 失败: %w", err)
		}
	}
	if s.completeTrigger != nil {
		return s.completeTrigger(ctx, workflowRunID)
	}
	return nil
}

// ManualReject 人工驳回：标记为 failed 并触发返工。
func (s *Service) ManualReject(ctx context.Context, projectID int64, reason string) error {
	return s.ManualRework(ctx, projectID, reason)
}

// Rerun 重新验收：创建新一轮 accept_run，重新执行完整验收流程。
func (s *Service) Rerun(ctx context.Context, projectID int64) error {
	wfRun, _, stageRunID, err := s.findActiveAcceptRun(ctx, projectID)
	if err != nil {
		return err
	}
	workflowRunID := gconv.Int64(wfRun["id"])

	g.Log().Infof(ctx, "[AcceptStage] 重新验收: project=%d workflowRun=%d", projectID, workflowRunID)
	return s.Run(ctx, workflowRunID, stageRunID)
}

// ManualRework 驳回并返工：标记 failed 并触发返工链。
func (s *Service) ManualRework(ctx context.Context, projectID int64, reason string) error {
	wfRun, acceptRun, stageRunID, err := s.findActiveAcceptRun(ctx, projectID)
	if err != nil {
		return err
	}
	acceptRunID := gconv.Int64(acceptRun["id"])
	workflowRunID := gconv.Int64(wfRun["id"])

	// CAS: 只有 completed 或 running 状态的 accept_run 允许人工驳回
	arStatus := gconv.String(acceptRun["status"])
	if arStatus != "completed" && arStatus != "running" {
		return fmt.Errorf("accept_run(%d) 状态为 %s，不允许人工驳回", acceptRunID, arStatus)
	}

	// 更新 accept_run 为 failed
	if err := s.acceptRunRepo.UpdateDecision(ctx, acceptRunID, acceptance.DecisionFailed, 0, "人工驳回: "+reason); err != nil {
		return fmt.Errorf("更新决策失败: %w", err)
	}
	if arStatus == "running" {
		_, _ = s.acceptRunRepo.UpdateStatus(ctx, acceptRunID, "running", "completed", g.Map{
			"finished_at": gtime.Now(),
		})
	}

	g.Log().Infof(ctx, "[AcceptStage] 人工驳回并返工: acceptRunID=%d project=%d reason=%s", acceptRunID, projectID, reason)

	// 完成 accept stage → 触发返工
	if s.stageCompleter != nil {
		if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
			return fmt.Errorf("CompleteStage 失败: %w", err)
		}
	}
	if s.reworkTrigger != nil {
		issues, _ := s.GetLatestIssues(ctx, acceptRunID)
		return s.reworkTrigger(ctx, workflowRunID, acceptRunID, issues)
	}
	return nil
}

// findActiveAcceptRun 查找项目当前活跃的验收运行。
func (s *Service) findActiveAcceptRun(ctx context.Context, projectID int64) (wfRun, acceptRun map[string]interface{}, stageRunID int64, err error) {
	// 查活跃 workflow_run
	wfRunRecord, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		One()
	if err != nil || wfRunRecord.IsEmpty() {
		return nil, nil, 0, fmt.Errorf("项目 %d 无活跃的工作流运行", projectID)
	}

	workflowRunID := wfRunRecord["id"].Int64()

	// 查最新 accept_run
	acceptRunRecord, err := s.acceptRunRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(acceptRunRecord) == 0 {
		return nil, nil, 0, fmt.Errorf("工作流 %d 无验收运行记录", workflowRunID)
	}

	stageRunID = gconv.Int64(acceptRunRecord["stage_run_id"])
	return wfRunRecord.Map(), acceptRunRecord, stageRunID, nil
}
