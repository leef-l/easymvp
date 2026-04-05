// Package accept 管理自动验收阶段：证据收集 → 规则评估 → 裁决归并 → 决定走向。
package accept

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/acceptance"
	"easymvp/app/mvp/internal/workflow/repo"
)

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

	// 2. 获取验收轮次
	round, err := s.acceptRunRepo.GetNextRound(ctx, workflowRunID)
	if err != nil {
		round = 1
	}

	// 3. 创建 accept_run
	acceptRunID, err := s.acceptRunRepo.Create(ctx, g.Map{
		"workflow_run_id": workflowRunID,
		"stage_run_id":    stageRunID,
		"project_id":      projectID,
		"accept_round":    round,
		"status":          "running",
		"rules_version":   "v1.0.0",
		"created_by":      0,
		"dept_id":         0,
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
	}

	g.Log().Infof(ctx, "[AcceptStage] 开始验收: workflowRunID=%d stageRunID=%d acceptRunID=%d round=%d projectType=%s",
		workflowRunID, stageRunID, acceptRunID, round, projectType)

	// 4. 收集证据
	_, evidenceErr := s.collector.Collect(ctx, acceptCtx)
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
	decision, reduceErr := s.reducer.Reduce(ctx, acceptCtx, hits)
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
		// 完成 accept stage → 触发 rework
		if s.stageCompleter != nil {
			if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
				g.Log().Errorf(ctx, "[AcceptStage] CompleteStage 失败: %v", err)
				return err
			}
		}
		if s.reworkTrigger != nil {
			if err := s.reworkTrigger(ctx, workflowRunID, acceptRunID, decision.Issues); err != nil {
				g.Log().Errorf(ctx, "[AcceptStage] 触发返工失败: %v", err)
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
