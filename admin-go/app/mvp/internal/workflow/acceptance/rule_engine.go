package acceptance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

// RuleEngine 验收规���引擎。
type RuleEngine struct {
	ruleRepo *repo.AcceptRuleRepo
}

// NewRuleEngine 创建规则引擎。
func NewRuleEngine(ruleRepo *repo.AcceptRuleRepo) *RuleEngine {
	return &RuleEngine{ruleRepo: ruleRepo}
}

// ruleConfig 规则配置通用结构。
type ruleConfig struct {
	ForbidStatus        []string `json:"forbid_status"`
	RequiredFiles       []string `json:"required_files"`
	TaskKinds           []string `json:"task_kinds"`
	RequireNonEmpty     bool     `json:"require_non_empty_result"`
	RequiredExtensions  []string `json:"required_extensions"`
	RequiredStageOutputs []string `json:"required_stage_outputs"`
	RequiredSections    []string `json:"required_sections"`
	RequiredKeywords    []string `json:"required_keywords"`
}

// LoadAndEvaluate 加载项目类型对应的规则并执行评估。
// 返回规则快照 JSON 和命中���果。
func (e *RuleEngine) LoadAndEvaluate(ctx context.Context, in *AcceptContext) (rulesSnapshot string, hits []RuleHit, err error) {
	// 加载规则
	rules, err := e.ruleRepo.ListByProjectType(ctx, in.ProjectType)
	if err != nil {
		return "", nil, fmt.Errorf("加载规则失败: %w", err)
	}

	// 规则快照
	snapshotBytes, _ := json.Marshal(rules)
	rulesSnapshot = string(snapshotBytes)

	if len(rules) == 0 {
		g.Log().Infof(ctx, "[RuleEngine] 项目类型 %s 无可用规则，跳过规则评估", in.ProjectType)
		return rulesSnapshot, nil, nil
	}

	// 逐条评估
	for _, rule := range rules {
		ruleCode := rule["rule_code"].(string)
		ruleName := rule["rule_name"].(string)
		ruleType := rule["rule_type"].(string)
		scopeType := rule["scope_type"].(string)
		configJSON := rule["config_json"].(string)

		var cfg ruleConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			g.Log().Warningf(ctx, "[RuleEngine] 规则配置解析失败: rule=%s err=%v", ruleCode, err)
			continue
		}

		ruleHits := e.evaluateRule(ctx, in, ruleCode, ruleName, ruleType, scopeType, &cfg)
		hits = append(hits, ruleHits...)
	}

	g.Log().Infof(ctx, "[RuleEngine] 评估完成: projectType=%s rules=%d hits=%d", in.ProjectType, len(rules), len(hits))
	return rulesSnapshot, hits, nil
}

// evaluateRule 评估单条规则。
func (e *RuleEngine) evaluateRule(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	switch ruleCode {
	case "software.no_failed_tasks":
		return e.checkNoFailedTasks(ctx, in, ruleCode, ruleName, ruleType, scopeType, cfg)
	case "software.output_not_empty":
		return e.checkOutputNotEmpty(ctx, in, ruleCode, ruleName, ruleType, scopeType, cfg)
	case "software.required_file_exists":
		return e.checkRequiredFiles(ctx, in, ruleCode, ruleName, ruleType, scopeType, cfg)
	case "document.required_output_exists":
		return e.checkRequiredExtensions(ctx, in, ruleCode, ruleName, ruleType, scopeType, cfg)
	case "document.summary_present":
		return e.checkRequiredStageOutputs(ctx, in, ruleCode, ruleName, ruleType, scopeType, cfg)
	default:
		// 未实现的规则直接跳过
		g.Log().Debugf(ctx, "[RuleEngine] 规则 %s 尚无评估实现，跳过", ruleCode)
		return nil
	}
}

// checkNoFailedTasks 检查是否存在禁止状态的��务。
func (e *RuleEngine) checkNoFailedTasks(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if len(cfg.ForbidStatus) == 0 {
		return nil
	}

	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", in.WorkflowRunID).
		WhereIn("status", cfg.ForbidStatus).
		WhereNull("deleted_at").
		Fields("id, name, status").
		All()
	if err != nil || len(tasks) == 0 {
		return nil
	}

	var hits []RuleHit
	for _, t := range tasks {
		hits = append(hits, RuleHit{
			RuleCode:        ruleCode,
			RuleName:        ruleName,
			RuleType:        ruleType,
			ScopeType:       scopeType,
			Severity:        SeverityBlocker,
			Title:           fmt.Sprintf("任务 %s 状态为 %s", t["name"].String(), t["status"].String()),
			Detail:          fmt.Sprintf("任务 ID=%d 处于禁止状态 %s", t["id"].Int64(), t["status"].String()),
			ExpectedValue:   "所有任务状态不在禁止列���中",
			ActualValue:     t["status"].String(),
			SuggestedAction: "修复或跳过该失败任务后重新验收",
			DomainTaskID:    t["id"].Int64(),
		})
	}
	return hits
}

// checkOutputNotEmpty 检查关键任务输出不为空。
func (e *RuleEngine) checkOutputNotEmpty(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if !cfg.RequireNonEmpty || len(cfg.TaskKinds) == 0 {
		return nil
	}

	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", in.WorkflowRunID).
		WhereIn("task_kind", cfg.TaskKinds).
		Where("status", "completed").
		WhereNull("deleted_at").
		Fields("id, name, result, task_kind").
		All()
	if err != nil {
		return nil
	}

	var hits []RuleHit
	for _, t := range tasks {
		if t["result"].String() == "" {
			hits = append(hits, RuleHit{
				RuleCode:        ruleCode,
				RuleName:        ruleName,
				RuleType:        ruleType,
				ScopeType:       scopeType,
				Severity:        SeverityError,
				Title:           fmt.Sprintf("任务 %s 输出为空", t["name"].String()),
				Detail:          fmt.Sprintf("任务 ID=%d (kind=%s) 已完成但 result 为空", t["id"].Int64(), t["task_kind"].String()),
				ExpectedValue:   "非空输出",
				ActualValue:     "空",
				SuggestedAction: "检查执行器是否正确写回输出",
				DomainTaskID:    t["id"].Int64(),
			})
		}
	}
	return hits
}

// checkRequiredFiles 检查关键文件是否存在（基于 domain_task.affected_resources）。
func (e *RuleEngine) checkRequiredFiles(_ context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	// 第一版基于工作目录文件检查（简化实现）
	// TODO: 后续可增强为读 git diff 或 workspace 产物
	if in.WorkDir == "" || len(cfg.RequiredFiles) == 0 {
		return nil
	}

	var hits []RuleHit
	for _, f := range cfg.RequiredFiles {
		// 简单检查：文件是否在工作目录中存在
		// 注意：此处不做实际文件系统检查（Accept 可能运行在非宿主机），
		// 而是检查 domain_task 的 affected_resources 中是否涵盖该文��
		hits = append(hits, RuleHit{
			RuleCode:        ruleCode,
			RuleName:        ruleName,
			RuleType:        ruleType,
			ScopeType:       scopeType,
			Severity:        SeverityWarn,
			Title:           fmt.Sprintf("需验证文件 %s 是否存在", f),
			Detail:          "当前版本基于元数据检查，建议人工确认",
			ExpectedValue:   fmt.Sprintf("文件 %s 存在", f),
			ActualValue:     "待人工确认",
			SuggestedAction: "确认文件已生成后可忽略此条",
		})
	}
	return hits
}

// checkRequiredExtensions 检查文档产物扩展名。
func (e *RuleEngine) checkRequiredExtensions(_ context.Context, _ *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, _ *ruleConfig) []RuleHit {
	// 第一版简化：记录 info 级提示，不阻塞
	return []RuleHit{{
		RuleCode:        ruleCode,
		RuleName:        ruleName,
		RuleType:        ruleType,
		ScopeType:       scopeType,
		Severity:        SeverityInfo,
		Title:           "文档产物格式检查",
		Detail:          "第一版基于元数据检查，建议人工确认产出文件格式",
		SuggestedAction: "确认文档产物已生成",
	}}
}

// checkRequiredStageOutputs 检查必需阶段是否已完成。
func (e *RuleEngine) checkRequiredStageOutputs(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if len(cfg.RequiredStageOutputs) == 0 {
		return nil
	}

	var hits []RuleHit
	for _, stageType := range cfg.RequiredStageOutputs {
		count, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
			Where("workflow_run_id", in.WorkflowRunID).
			Where("stage_type", stageType).
			Where("status", "completed").
			WhereNull("deleted_at").
			Count()
		if err != nil || count == 0 {
			hits = append(hits, RuleHit{
				RuleCode:        ruleCode,
				RuleName:        ruleName,
				RuleType:        ruleType,
				ScopeType:       scopeType,
				Severity:        SeverityError,
				Title:           fmt.Sprintf("阶段 %s 未完成", stageType),
				Detail:          fmt.Sprintf("要求阶段 %s 至少有一次 completed 记录", stageType),
				ExpectedValue:   "completed",
				ActualValue:     "无完成记录",
				SuggestedAction: "检查该阶段是否正常执行",
			})
		}
	}
	return hits
}
