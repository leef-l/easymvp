package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/qualitygate"
	"easymvp/app/mvp/internal/workflow/repo"
)

// RuleEngine 验收规���引擎。
type RuleEngine struct {
	ruleRepo             *repo.AcceptRuleRepo
	verificationRunRepo  *repo.VerificationRunRepo
	verificationEvidence *repo.VerificationEvidenceRepo
	taskWorkspaceRepo    *repo.TaskWorkspaceRepo
	domainTaskRepo       *repo.DomainTaskRepo
	stageRunRepo         *repo.StageRunRepo
}

// NewRuleEngine 创建规则引擎。
func NewRuleEngine(ruleRepo *repo.AcceptRuleRepo) *RuleEngine {
	return &RuleEngine{ruleRepo: ruleRepo}
}

// ruleConfig 规则配置通用结构。
type ruleConfig struct {
	ForbidStatus                     []string `json:"forbid_status"`
	RequiredFiles                    []string `json:"required_files"`
	TaskKinds                        []string `json:"task_kinds"`
	RequireNonEmpty                  bool     `json:"require_non_empty_result"`
	RequiredExtensions               []string `json:"required_extensions"`
	RequiredStageOutputs             []string `json:"required_stage_outputs"`
	RequiredSections                 []string `json:"required_sections"`
	RequiredKeywords                 []string `json:"required_keywords"`
	RequireManualReviewDeliveryModes []string `json:"require_manual_review_delivery_modes"`
	RequireManualReviewSyncStatuses  []string `json:"require_manual_review_sync_statuses"`
	MinRiskLevel                     string   `json:"min_risk_level"`
}

// LoadAndEvaluate 加载项目类型对应的规则并执行评估。
// 返回规则快照 JSON 和命中���果。
func (e *RuleEngine) LoadAndEvaluate(ctx context.Context, in *AcceptContext) (rulesSnapshot string, hits []RuleHit, err error) {
	// 加载规则：先按 category_code 精确匹配，无结果则按 family_code 回退
	rules, err := e.ruleRepo.ListByProjectTypeWithFallback(ctx, in.ProjectType, in.FamilyCode)
	if err != nil {
		return "", nil, fmt.Errorf("加载规则失败: %w", err)
	}

	// 规则快照
	snapshotBytes, snapErr := json.Marshal(rules)
	if snapErr != nil {
		snapshotBytes = []byte("[]")
	}
	rulesSnapshot = string(snapshotBytes)

	if len(rules) == 0 {
		g.Log().Infof(ctx, "[RuleEngine] 项目类型 %s 无可用规则，跳过规则评估", in.ProjectType)
		return rulesSnapshot, nil, nil
	}

	// 逐条评估
	for _, rule := range rules {
		ruleCode, _ := rule["rule_code"].(string)
		ruleName, _ := rule["rule_name"].(string)
		ruleType, _ := rule["rule_type"].(string)
		scopeType, _ := rule["scope_type"].(string)
		configJSON, _ := rule["config_json"].(string)
		if ruleCode == "" {
			continue
		}

		var cfg ruleConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			g.Log().Warningf(ctx, "[RuleEngine] 规则配置解析失败: rule=%s err=%v", ruleCode, err)
			continue
		}

		ruleHits := e.evaluateRule(ctx, in, ruleCode, ruleName, ruleType, scopeType, &cfg)
		hits = append(hits, ruleHits...)
	}
	hits = append(hits, e.evaluateVerificationStandard(ctx, in)...)

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
	case "software.delivery_review_required":
		return e.checkDeliveryReviewRequired(ctx, in, ruleCode, ruleName, ruleType, scopeType, cfg)
	default:
		// 未实现的规则产生 warning hit，而非静默跳过（避免"明明没通过却算通过"的隐蔽 bug）
		g.Log().Warningf(ctx, "[RuleEngine] 规则 %s 尚无评估实现，标记为未实现", ruleCode)
		return []RuleHit{{
			RuleCode:        ruleCode,
			RuleName:        ruleName,
			RuleType:        ruleType,
			ScopeType:       scopeType,
			Severity:        SeverityWarn,
			Title:           fmt.Sprintf("规则 %s 尚未实现", ruleCode),
			Detail:          fmt.Sprintf("数据库中配置了规则 %s，但引擎中没有对应的评估逻辑。该规则的检查结果无法保证。", ruleCode),
			ExpectedValue:   "规则已实现并可执行评估",
			ActualValue:     "评估逻辑缺失",
			SuggestedAction: "实现该规则的评估逻辑，或从数据库中移除该规则配置",
		}}
	}
}

type verificationSnapshot struct {
	Present            bool
	Status             string
	Decision           string
	HasBrowserEvidence bool
	Summary            string
}

var resolveRequiredProjectRole = func(ctx context.Context, projectID int64, requirement qualitygate.ProjectRoleRequirement) error {
	_, err := repo.GetProjectRoleByLevel(ctx, projectID, requirement.RoleType, requirement.RoleLevel)
	return err
}

func (e *RuleEngine) evaluateVerificationStandard(ctx context.Context, in *AcceptContext) []RuleHit {
	signals := qualitygate.DetectProjectSignals(in.WorkDir, in.FamilyCode, in.ProjectType)
	standard := qualitygate.ResolveVerificationStandard(signals)
	hits := e.evaluateRequiredProjectRoles(ctx, in, standard)
	if !standard.RequirePassedVerification {
		return hits
	}

	snapshot, err := e.loadLatestVerificationSnapshot(ctx, in.WorkflowRunID)
	if err != nil {
		hits = append(hits, *buildVerificationRuleHit(
			"software.verification_read_failed",
			SeverityError,
			"无法读取最新验证结果",
			err.Error(),
			"已完成且通过的标准化验证",
			"验证结果读取失败",
			"修复验证记录读取异常后重新验收",
		))
		return hits
	}

	if hit := evaluateVerificationSnapshot(standard, snapshot); hit != nil {
		hits = append(hits, *hit)
	}
	return hits
}

func (e *RuleEngine) evaluateRequiredProjectRoles(ctx context.Context, in *AcceptContext, standard qualitygate.VerificationStandard) []RuleHit {
	var hits []RuleHit
	for _, requirement := range standard.RequiredProjectRoles {
		if !requirement.Blocking {
			continue
		}
		if err := resolveRequiredProjectRole(ctx, in.ProjectID, requirement); err != nil {
			hits = append(hits, RuleHit{
				RuleCode:        "software.required_project_role_missing",
				RuleName:        "verification standard required project role",
				RuleType:        "process",
				ScopeType:       "project",
				Severity:        SeverityError,
				Title:           fmt.Sprintf("缺少标准要求的项目角色：%s", requirement.Label()),
				Detail:          fmt.Sprintf("当前项目命中了 %s，但项目无法解析 %s。该角色需要在验收阶段承担%s。", standard.DisplayName, requirement.ExpectedRoleRef(), valueOrDefault(requirement.Purpose, "关键体验评审职责")),
				ExpectedValue:   requirement.ExpectedRoleRef(),
				ActualValue:     valueOrDefault(err.Error(), "未找到可用角色配置"),
				SuggestedAction: fmt.Sprintf("为项目或该分类补齐 %s 角色预设/项目角色后重新验收", requirement.RoleType),
			})
		}
	}
	return hits
}

func (e *RuleEngine) loadLatestVerificationSnapshot(ctx context.Context, workflowRunID int64) (verificationSnapshot, error) {
	var snapshot verificationSnapshot
	runRepo := e.verificationRunRepo
	if runRepo == nil {
		runRepo = repo.NewVerificationRunRepo()
	}
	record, err := runRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil {
		return snapshot, err
	}
	if len(record) == 0 {
		return snapshot, nil
	}

	snapshot = verificationSnapshot{
		Present:  true,
		Status:   g.NewVar(record["status"]).String(),
		Decision: g.NewVar(record["decision"]).String(),
		Summary:  g.NewVar(record["summary"]).String(),
	}
	runID := g.NewVar(record["id"]).Int64()
	hasBrowserEvidence, evidenceErr := e.verificationRunHasCheckKind(ctx, runID, qualitygate.CheckKindBrowser)
	if evidenceErr != nil {
		return snapshot, evidenceErr
	}
	snapshot.HasBrowserEvidence = hasBrowserEvidence
	return snapshot, nil
}

func (e *RuleEngine) verificationRunHasCheckKind(ctx context.Context, verificationRunID int64, kind string) (bool, error) {
	evidenceRepo := e.verificationEvidence
	if evidenceRepo == nil {
		evidenceRepo = repo.NewVerificationEvidenceRepo()
	}
	records, err := evidenceRepo.ListByVerificationRun(ctx, verificationRunID)
	if err != nil {
		return false, err
	}
	for _, record := range records {
		if g.NewVar(record["evidence_type"]).String() != "command" {
			continue
		}
		if verificationEvidenceCheckKind(g.NewVar(record["content_ref"]).String(), g.NewVar(record["summary"]).String()) == kind {
			return true, nil
		}
	}
	return false, nil
}

func verificationEvidenceCheckKind(contentRef string, summary string) string {
	type stepEvidence struct {
		Name    string `json:"name"`
		Command string `json:"command"`
		Runner  string `json:"runner"`
	}

	var payload stepEvidence
	if json.Unmarshal([]byte(contentRef), &payload) == nil {
		return qualitygate.InferCheckKind(payload.Name, strings.Fields(payload.Command), payload.Runner)
	}
	return qualitygate.InferCheckKind(summary, strings.Fields(summary), "")
}

func evaluateVerificationSnapshot(standard qualitygate.VerificationStandard, snapshot verificationSnapshot) *RuleHit {
	if !snapshot.Present {
		return buildVerificationRuleHit(
			"software.verification_required",
			SeverityError,
			"缺少标准化验证结果",
			"当前项目所属标准要求验收前必须存在最新验证结果，但未找到任何验证运行记录。",
			"已完成且通过的标准化验证",
			"未找到验证记录",
			"先执行标准化验证，再重新进入验收",
		)
	}
	if snapshot.Status != "completed" {
		return buildVerificationRuleHit(
			"software.verification_not_completed",
			SeverityError,
			"最新验证尚未完成",
			"验收要求消费完成态验证结果，但最新验证运行仍未收口。",
			"最新验证状态为 completed",
			snapshot.Status,
			"等待验证完成或重新发起验证",
		)
	}
	if snapshot.Decision != DecisionPassed {
		return buildVerificationRuleHit(
			"software.verification_not_passed",
			SeverityError,
			"最新验证未通过",
			"验收标准要求最新验证必须通过后才能放行。",
			"验证决策为 passed",
			valueOrDefault(snapshot.Decision, "unknown"),
			"修复验证问题后重新验证并重新验收",
		)
	}
	if standard.RequireBrowserEvidence && !snapshot.HasBrowserEvidence {
		return buildVerificationRuleHit(
			"software.browser_verification_required",
			SeverityError,
			"缺少浏览器级验证证据",
			"当前项目命中了交互式交付标准，但最新验证结果中没有浏览器级关键路径证据。",
			"最新验证包含 browser/e2e 级别证据",
			"未检测到 browser/e2e 证据",
			"补齐 Playwright/Cypress/真机 UI 验证后重新验证",
		)
	}
	return nil
}

func buildVerificationRuleHit(ruleCode string, severity string, title string, detail string, expected string, actual string, action string) *RuleHit {
	return &RuleHit{
		RuleCode:        ruleCode,
		RuleName:        "verification standard required",
		RuleType:        "process",
		ScopeType:       "project",
		Severity:        severity,
		Title:           title,
		Detail:          detail,
		ExpectedValue:   expected,
		ActualValue:     actual,
		SuggestedAction: action,
	}
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (e *RuleEngine) checkDeliveryReviewRequired(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if len(cfg.RequireManualReviewDeliveryModes) == 0 &&
		len(cfg.RequireManualReviewSyncStatuses) == 0 &&
		strings.TrimSpace(cfg.MinRiskLevel) == "" {
		return nil
	}

	taskWorkspaceRepo := e.taskWorkspaceRepo
	if taskWorkspaceRepo == nil {
		taskWorkspaceRepo = repo.NewTaskWorkspaceRepo()
	}
	records, err := taskWorkspaceRepo.ListDeliveriesByWorkflow(ctx, in.WorkflowRunID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unknown column") {
			g.Log().Warningf(ctx, "[RuleEngine] 跳过交付审核规则，workspace delivery 列尚未迁移: %v", err)
			return nil
		}
		g.Log().Warningf(ctx, "[RuleEngine] 查询 workspace 交付结果失败: %v", err)
		return nil
	}

	allowedModes := make(map[string]struct{}, len(cfg.RequireManualReviewDeliveryModes))
	for _, mode := range cfg.RequireManualReviewDeliveryModes {
		mode = strings.ToLower(strings.TrimSpace(mode))
		if mode != "" {
			allowedModes[mode] = struct{}{}
		}
	}
	allowedSyncStatuses := make(map[string]struct{}, len(cfg.RequireManualReviewSyncStatuses))
	for _, status := range cfg.RequireManualReviewSyncStatuses {
		status = strings.ToLower(strings.TrimSpace(status))
		if status != "" {
			allowedSyncStatuses[status] = struct{}{}
		}
	}

	minRiskLevel := strings.ToLower(strings.TrimSpace(cfg.MinRiskLevel))
	var hits []RuleHit
	for _, record := range records {
		taskID := g.NewVar(record["task_id"]).Int64()
		deliveryMode := strings.ToLower(strings.TrimSpace(g.NewVar(record["delivery_mode"]).String()))
		deliveryStatus := strings.ToLower(strings.TrimSpace(g.NewVar(record["delivery_status"]).String()))
		syncStatus := strings.ToLower(strings.TrimSpace(g.NewVar(record["sync_status"]).String()))
		riskLevel := strings.ToLower(strings.TrimSpace(g.NewVar(record["risk_level"]).String()))
		patchRef := strings.TrimSpace(g.NewVar(record["patch_ref"]).String())
		reasons := buildDeliveryReviewReasons(deliveryMode, deliveryStatus, syncStatus, riskLevel, allowedModes, allowedSyncStatuses, minRiskLevel)
		if len(reasons) == 0 {
			continue
		}

		actualValue := fmt.Sprintf("delivery=%s, deliveryStatus=%s, sync=%s, risk=%s", deliveryMode, deliveryStatus, syncStatus, riskLevel)
		hits = append(hits, RuleHit{
			RuleCode:        ruleCode,
			RuleName:        ruleName,
			RuleType:        ruleType,
			ScopeType:       scopeType,
			Severity:        SeverityWarn,
			Title:           fmt.Sprintf("任务 %d 的交付结果需要人工审核", taskID),
			Detail:          fmt.Sprintf("命中条件：%s", strings.Join(reasons, "；")),
			ExpectedValue:   "低风险 patch 自动回写，或人工确认后再放行",
			ActualValue:     actualValue,
			SuggestedAction: "查看 patch/交付证据并人工确认后放行或返工",
			DomainTaskID:    taskID,
			ResourceRef:     patchRef,
		})
	}
	return hits
}

func buildDeliveryReviewReasons(deliveryMode, deliveryStatus, syncStatus, riskLevel string, allowedModes, allowedSyncStatuses map[string]struct{}, minRiskLevel string) []string {
	if deliveryStatus != "ready" {
		return nil
	}

	reasons := make([]string, 0, 3)
	if _, ok := allowedModes[deliveryMode]; ok {
		reasons = append(reasons, "交付形态="+deliveryMode)
	}
	if _, ok := allowedSyncStatuses[syncStatus]; ok {
		reasons = append(reasons, "回写状态="+syncStatus)
	}
	if riskLevelAtLeast(riskLevel, minRiskLevel) {
		reasons = append(reasons, "风险等级="+riskLevel)
	}
	return reasons
}

func riskLevelAtLeast(actual, threshold string) bool {
	if threshold == "" {
		return false
	}
	order := map[string]int{
		"low":    1,
		"medium": 2,
		"high":   3,
	}
	return order[actual] > 0 && order[threshold] > 0 && order[actual] >= order[threshold]
}

// checkNoFailedTasks 检查是否存在禁止状态的��务。
func (e *RuleEngine) checkNoFailedTasks(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if len(cfg.ForbidStatus) == 0 {
		return nil
	}

	domainTaskRepo := e.domainTaskRepo
	if domainTaskRepo == nil {
		domainTaskRepo = repo.NewDomainTaskRepo()
	}
	tasks, err := domainTaskRepo.ListByWorkflowAndStatuses(ctx, in.WorkflowRunID, cfg.ForbidStatus, "id", "name", "status")
	if err != nil || len(tasks) == 0 {
		return nil
	}

	var hits []RuleHit
	for _, t := range tasks {
		taskName := g.NewVar(t["name"]).String()
		taskStatus := g.NewVar(t["status"]).String()
		taskID := g.NewVar(t["id"]).Int64()
		hits = append(hits, RuleHit{
			RuleCode:        ruleCode,
			RuleName:        ruleName,
			RuleType:        ruleType,
			ScopeType:       scopeType,
			Severity:        SeverityBlocker,
			Title:           fmt.Sprintf("任务 %s 状态为 %s", taskName, taskStatus),
			Detail:          fmt.Sprintf("任务 ID=%d 处于禁止状态 %s", taskID, taskStatus),
			ExpectedValue:   "所有任务状态不在禁止列���中",
			ActualValue:     taskStatus,
			SuggestedAction: "修复或跳过该失败任务后重新验收",
			DomainTaskID:    taskID,
		})
	}
	return hits
}

// checkOutputNotEmpty 检查关键任务输出不为空。
func (e *RuleEngine) checkOutputNotEmpty(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if !cfg.RequireNonEmpty || len(cfg.TaskKinds) == 0 {
		return nil
	}

	domainTaskRepo := e.domainTaskRepo
	if domainTaskRepo == nil {
		domainTaskRepo = repo.NewDomainTaskRepo()
	}
	tasks, err := domainTaskRepo.ListCompletedByWorkflowAndKinds(ctx, in.WorkflowRunID, cfg.TaskKinds, "id", "name", "result", "task_kind")
	if err != nil {
		return nil
	}

	var hits []RuleHit
	for _, t := range tasks {
		resultText := g.NewVar(t["result"]).String()
		if resultText == "" {
			taskName := g.NewVar(t["name"]).String()
			taskID := g.NewVar(t["id"]).Int64()
			taskKind := g.NewVar(t["task_kind"]).String()
			hits = append(hits, RuleHit{
				RuleCode:        ruleCode,
				RuleName:        ruleName,
				RuleType:        ruleType,
				ScopeType:       scopeType,
				Severity:        SeverityError,
				Title:           fmt.Sprintf("任务 %s 输出为空", taskName),
				Detail:          fmt.Sprintf("任务 ID=%d (kind=%s) 已完成但 result 为空", taskID, taskKind),
				ExpectedValue:   "非空输出",
				ActualValue:     "空",
				SuggestedAction: "检查执行器是否正确写回输出",
				DomainTaskID:    taskID,
			})
		}
	}
	return hits
}

// checkRequiredFiles 检查关键文件是否存在（基于 domain_task.affected_resources）。
func (e *RuleEngine) checkRequiredFiles(_ context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if in.WorkDir == "" || len(cfg.RequiredFiles) == 0 {
		return nil
	}

	var hits []RuleHit
	for _, f := range cfg.RequiredFiles {
		fullPath := filepath.Clean(filepath.Join(in.WorkDir, f))
		if !strings.HasPrefix(fullPath, filepath.Clean(in.WorkDir)+string(filepath.Separator)) && fullPath != filepath.Clean(in.WorkDir) {
			continue
		}
		if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
			hits = append(hits, RuleHit{
				RuleCode:        ruleCode,
				RuleName:        ruleName,
				RuleType:        ruleType,
				ScopeType:       scopeType,
				Severity:        SeverityError,
				Title:           fmt.Sprintf("必需文件 %s 不存在", f),
				Detail:          fmt.Sprintf("在工作目录 %s 下未找到文件 %s", in.WorkDir, f),
				ExpectedValue:   fmt.Sprintf("文件 %s 存在", f),
				ActualValue:     "文件不存在",
				SuggestedAction: "确保相关任务已生成该文件",
			})
		} else if statErr != nil {
			hits = append(hits, RuleHit{
				RuleCode:        ruleCode,
				RuleName:        ruleName,
				RuleType:        ruleType,
				ScopeType:       scopeType,
				Severity:        SeverityWarn,
				Title:           fmt.Sprintf("无法访问文件 %s", f),
				Detail:          fmt.Sprintf("文件系统检查异常: %v", statErr),
				ExpectedValue:   fmt.Sprintf("文件 %s 可访问", f),
				ActualValue:     "访问异常",
				SuggestedAction: "人工确认文件是否存在",
			})
		}
		// 文件存在 → 不产生 hit
	}
	return hits
}

// checkRequiredExtensions 检查工作目录下是否存在指定扩展名的文件。
func (e *RuleEngine) checkRequiredExtensions(_ context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if in.WorkDir == "" || len(cfg.RequiredExtensions) == 0 {
		return nil
	}

	var hits []RuleHit
	for _, ext := range cfg.RequiredExtensions {
		found := false
		// 扫描工作目录一级文件（不递归，避免性能问题）
		entries, err := os.ReadDir(in.WorkDir)
		if err != nil {
			hits = append(hits, RuleHit{
				RuleCode: ruleCode, RuleName: ruleName, RuleType: ruleType, ScopeType: scopeType,
				Severity:        SeverityWarn,
				Title:           fmt.Sprintf("无法扫描工作目录检查 %s 文件", ext),
				Detail:          fmt.Sprintf("ReadDir 异常: %v", err),
				SuggestedAction: "人工确认文档产物是否存在",
			})
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ext) {
				found = true
				break
			}
		}
		if !found {
			hits = append(hits, RuleHit{
				RuleCode: ruleCode, RuleName: ruleName, RuleType: ruleType, ScopeType: scopeType,
				Severity:        SeverityError,
				Title:           fmt.Sprintf("未找到 %s 格式的文档产物", ext),
				Detail:          fmt.Sprintf("工作目录 %s 下未找到扩展名为 %s 的文件", in.WorkDir, ext),
				ExpectedValue:   fmt.Sprintf("至少一个 %s 文件", ext),
				ActualValue:     "未找到",
				SuggestedAction: "确保文档生成任务已输出对应格式文件",
			})
		}
	}
	return hits
}

// checkRequiredStageOutputs 检查必需阶段是否已完成。
func (e *RuleEngine) checkRequiredStageOutputs(ctx context.Context, in *AcceptContext, ruleCode, ruleName, ruleType, scopeType string, cfg *ruleConfig) []RuleHit {
	if len(cfg.RequiredStageOutputs) == 0 {
		return nil
	}

	// 批量查询已完成的阶段类型（避免 N+1）
	stageRunRepo := e.stageRunRepo
	if stageRunRepo == nil {
		stageRunRepo = repo.NewStageRunRepo()
	}
	completedStages, csErr := stageRunRepo.ListCompletedStageTypes(ctx, in.WorkflowRunID, cfg.RequiredStageOutputs)
	if csErr != nil {
		g.Log().Warningf(ctx, "[RuleEngine] 查询已完成阶段失败: wfRunID=%d err=%v", in.WorkflowRunID, csErr)
	}
	completedSet := make(map[string]bool, len(completedStages))
	for _, r := range completedStages {
		completedSet[g.NewVar(r["stage_type"]).String()] = true
	}

	var hits []RuleHit
	for _, stageType := range cfg.RequiredStageOutputs {
		if !completedSet[stageType] {
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
