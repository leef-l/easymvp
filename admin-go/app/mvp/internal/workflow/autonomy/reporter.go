package autonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/provider"
)

// Reporter 项目自治汇报生成器。
type Reporter struct {
	reportRepo *repo.ProjectReportRepo
}

// NewReporter 创建汇报生成器。
func NewReporter(reportRepo *repo.ProjectReportRepo) *Reporter {
	return &Reporter{reportRepo: reportRepo}
}

// GenerateStageReport 生成阶段报告。在阶段完成时自动调用。
func (r *Reporter) GenerateStageReport(ctx context.Context, workflowRunID int64, stageType string) error {
	// 查 workflow_run 获取 projectID
	wfRun, err := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "project_id")
	if err != nil || len(wfRun) == 0 {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}
	projectID := g.NewVar(wfRun["project_id"]).Int64()

	// 收集指标
	metrics := r.collectStageMetrics(ctx, workflowRunID, stageType)

	// 尝试用 LLM 生成自然语言报告，失败则用结构化模板
	content := r.generateReportContent(ctx, projectID, stageType, metrics)

	stageLabel := stageTypeLabel(stageType)
	title := fmt.Sprintf("%s阶段报告", stageLabel)
	metricsJSON, marshalErr := json.Marshal(metrics)
	if marshalErr != nil {
		g.Log().Warningf(ctx, "[Reporter] metrics 序列化失败: %v", marshalErr)
		metricsJSON = []byte("{}")
	}

	_, err = r.reportRepo.Create(ctx, g.Map{
		"workflow_run_id": workflowRunID,
		"project_id":      projectID,
		"report_type":     ReportStage,
		"stage_type":      stageType,
		"title":           title,
		"content":         content,
		"metrics":         string(metricsJSON),
	})
	if err != nil {
		return fmt.Errorf("保存阶段报告失败: %w", err)
	}

	g.Log().Infof(ctx, "[Reporter] 阶段报告已生成: workflowRun=%d stage=%s", workflowRunID, stageType)
	return nil
}

// GetReports 获取项目报告列表。
func (r *Reporter) GetReports(ctx context.Context, projectID int64, reportType string) ([]g.Map, error) {
	return r.reportRepo.ListByProject(ctx, projectID, reportType)
}

// collectStageMetrics 收集阶段指标。
func (r *Reporter) collectStageMetrics(ctx context.Context, workflowRunID int64, stageType string) string {
	var b strings.Builder

	// 阶段运行信息
	stageRun, err := repo.NewStageRunRepo().GetLatestByWorkflowAndType(ctx, workflowRunID, stageType, "status", "started_at", "finished_at")
	if err == nil && !stageRun.IsEmpty() {
		b.WriteString(fmt.Sprintf("阶段状态: %s\n", stageRun["status"].String()))
		if !stageRun["started_at"].IsEmpty() {
			b.WriteString(fmt.Sprintf("开始时间: %s\n", stageRun["started_at"].String()))
		}
		if !stageRun["finished_at"].IsEmpty() {
			b.WriteString(fmt.Sprintf("结束时间: %s\n", stageRun["finished_at"].String()))
		}
	}

	// 按阶段类型收集特定指标
	switch stageType {
	case "execute":
		r.collectExecuteMetrics(ctx, workflowRunID, &b)
	case "accept":
		r.collectAcceptMetrics(ctx, workflowRunID, &b)
	case "rework":
		r.collectReworkMetrics(ctx, workflowRunID, &b)
	case "review":
		r.collectReviewMetrics(ctx, workflowRunID, &b)
	}

	return b.String()
}

func (r *Reporter) collectExecuteMetrics(ctx context.Context, workflowRunID int64, b *strings.Builder) {
	// 任务统计
	tasks, tasksErr := repo.NewDomainTaskRepo().ListStatusRowsByWorkflow(ctx, workflowRunID)
	if tasksErr != nil {
		g.Log().Warningf(ctx, "[Reporter] 查询任务统计失败: wfRunID=%d err=%v", workflowRunID, tasksErr)
	}

	counts := make(map[string]int)
	for _, t := range tasks {
		counts[g.NewVar(t["status"]).String()] += g.NewVar(t["count"]).Int()
	}
	totalTasks := 0
	for _, count := range counts {
		totalTasks += count
	}
	b.WriteString(fmt.Sprintf("总任务数: %d\n", totalTasks))
	for status, count := range counts {
		b.WriteString(fmt.Sprintf("  %s: %d\n", status, count))
	}
}

func (r *Reporter) collectAcceptMetrics(ctx context.Context, workflowRunID int64, b *strings.Builder) {
	// 验收轮次和结果
	runs, runsErr := repo.NewAcceptRunRepo().ListByWorkflow(ctx, workflowRunID, "accept_round", "status", "decision", "score")
	if runsErr != nil {
		g.Log().Warningf(ctx, "[Reporter] 查询验收轮次失败: wfRunID=%d err=%v", workflowRunID, runsErr)
	}

	b.WriteString(fmt.Sprintf("验收轮次: %d\n", len(runs)))
	for _, run := range runs {
		b.WriteString(fmt.Sprintf("  第%d轮: %s (评分: %s, 决策: %s)\n",
			g.NewVar(run["accept_round"]).Int(), g.NewVar(run["status"]).String(),
			g.NewVar(run["score"]).String(), g.NewVar(run["decision"]).String()))
	}

	// 问题统计
	issues, issErr := repo.NewAcceptIssueRepo().CountGroupBySeverityByWorkflow(ctx, workflowRunID)
	if issErr != nil {
		g.Log().Warningf(ctx, "[Reporter] 查询验收问题失败: wfRunID=%d err=%v", workflowRunID, issErr)
	}
	for _, issue := range issues {
		b.WriteString(fmt.Sprintf("  %s 问题: %d 个\n", g.NewVar(issue["severity"]).String(), g.NewVar(issue["cnt"]).Int()))
	}
}

func (r *Reporter) collectReworkMetrics(ctx context.Context, workflowRunID int64, b *strings.Builder) {
	count, cntErr := repo.NewStageRunRepo().CountByWorkflowAndType(ctx, workflowRunID, "rework")
	if cntErr != nil {
		g.Log().Warningf(ctx, "[Reporter] 查询返工轮次失败: wfRunID=%d err=%v", workflowRunID, cntErr)
	}
	b.WriteString(fmt.Sprintf("返工轮次: %d\n", count))

	handoffs, hoErr := repo.NewHandoffRecordRepo().CountByWorkflow(ctx, workflowRunID)
	if hoErr != nil {
		g.Log().Warningf(ctx, "[Reporter] 查询交接记录失败: wfRunID=%d err=%v", workflowRunID, hoErr)
	}
	b.WriteString(fmt.Sprintf("交接记录: %d\n", handoffs))
}

func (r *Reporter) collectReviewMetrics(ctx context.Context, workflowRunID int64, b *strings.Builder) {
	issues, revIssErr := repo.NewReviewIssueRepo().CountGroupBySeverityByWorkflow(ctx, workflowRunID)
	if revIssErr != nil {
		g.Log().Warningf(ctx, "[Reporter] 查询审核问题失败: wfRunID=%d err=%v", workflowRunID, revIssErr)
	}
	for _, issue := range issues {
		b.WriteString(fmt.Sprintf("  %s: %d 个\n", g.NewVar(issue["severity"]).String(), g.NewVar(issue["cnt"]).Int()))
	}
}

// generateReportContent 生成报告正文（LLM 生成或模板降级）。
func (r *Reporter) generateReportContent(ctx context.Context, projectID int64, stageType string, metrics string) string {
	// 尝试 LLM 生成
	modelInfo, err := resolveProjectModel(ctx, projectID, "architect")
	if err != nil {
		return r.templateReport(stageType, metrics)
	}

	p, err := provider.GetProvider(provider.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
		APIKey:             modelInfo.APIKey,
		APISecret:          modelInfo.APISecret,
	})
	if err != nil {
		return r.templateReport(stageType, metrics)
	}

	systemPrompt := "你是项目管理助手，根据提供的阶段指标数据生成简洁的 Markdown 格式阶段报告。报告应包含：阶段概述、关键指标、问题总结、建议。控制在 500 字以内。"
	userPrompt := fmt.Sprintf("## %s阶段指标\n\n```\n%s```\n\n请生成该阶段的 Markdown 格式报告。", stageTypeLabel(stageType), metrics)

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: userPrompt}},
		MaxTokens:    1000,
		Temperature:  0.5,
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Reporter] LLM 生成报告失败(降级为模板): %v", err)
		return r.templateReport(stageType, metrics)
	}

	return resp.Content
}

// templateReport 模板降级报告。
func (r *Reporter) templateReport(stageType string, metrics string) string {
	return fmt.Sprintf("# %s阶段报告\n\n## 指标\n\n```\n%s```\n\n*（本报告由模板自动生成）*\n", stageTypeLabel(stageType), metrics)
}

func stageTypeLabel(stageType string) string {
	labels := map[string]string{
		"design":   "设计",
		"review":   "审核",
		"execute":  "执行",
		"accept":   "验收",
		"rework":   "返工",
		"complete": "完成",
	}
	if l, ok := labels[stageType]; ok {
		return l
	}
	return stageType
}
