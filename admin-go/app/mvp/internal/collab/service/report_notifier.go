package service

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/collab"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/workflow/event"
)

// ReportNotifier 报告推送服务，监听阶段/项目完成事件推送报告卡片。
type ReportNotifier struct {
	adapter     collab.MessageAdapter
	bindingRepo *collabRepo.BindingRepo
}

// NewReportNotifier 创建报告推送服务。
func NewReportNotifier(adapter collab.MessageAdapter, bindingRepo *collabRepo.BindingRepo) *ReportNotifier {
	return &ReportNotifier{adapter: adapter, bindingRepo: bindingRepo}
}

// OnStageCompleted 阶段完成时推送阶段报告卡片。
func (n *ReportNotifier) OnStageCompleted(evt event.Event) {
	go func() {
		ctx := context.Background()
		if !n.adapter.IsEnabled(ctx) {
			return
		}

		// 查最新的阶段报告
		report, err := g.DB().Model("mvp_project_report").Ctx(ctx).
			Where("workflow_run_id", evt.WorkflowRunID).
			Where("report_type", "stage").
			WhereNull("deleted_at").
			OrderDesc("created_at").
			One()
		if err != nil || report.IsEmpty() {
			return // 无报告则不推送
		}

		card := &collab.InteractiveCard{
			Title:       "阶段报告",
			ProjectName: n.lookupProjectName(ctx, evt.WorkflowRunID),
			Details:     truncate(report["title"].String(), 200),
			HeaderColor: "blue",
		}

		createdBy := n.lookupProjectOwner(ctx, evt.WorkflowRunID)
		n.sendToOwner(ctx, createdBy, func(openID string) {
			if err := n.adapter.SendCardMessage(ctx, openID, card); err != nil {
				g.Log().Warningf(ctx, "[ReportNotifier] 推送阶段报告失败: err=%v", err)
			}
		})
	}()
}

// OnWorkflowCompleted 项目完成时推送汇总卡片。
func (n *ReportNotifier) OnWorkflowCompleted(evt event.Event) {
	go func() {
		ctx := context.Background()
		if !n.adapter.IsEnabled(ctx) {
			return
		}

		projectName := n.lookupProjectName(ctx, evt.WorkflowRunID)
		card := &collab.InteractiveCard{
			Title:       "项目完成",
			ProjectName: projectName,
			Details:     fmt.Sprintf("项目 [%s] 已完成全部工作流", projectName),
			HeaderColor: "green",
		}

		createdBy := n.lookupProjectOwner(ctx, evt.WorkflowRunID)
		n.sendToOwner(ctx, createdBy, func(openID string) {
			if err := n.adapter.SendCardMessage(ctx, openID, card); err != nil {
				g.Log().Warningf(ctx, "[ReportNotifier] 推送项目完成卡片失败: err=%v", err)
			}
		})
	}()
}

func (n *ReportNotifier) lookupProjectOwner(ctx context.Context, workflowRunID int64) int64 {
	val, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	if val.Int64() == 0 {
		return 0
	}
	createdBy, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", val.Int64()).Value("created_by")
	return createdBy.Int64()
}

func (n *ReportNotifier) lookupProjectName(ctx context.Context, workflowRunID int64) string {
	val, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	if val.Int64() == 0 {
		return "未知项目"
	}
	name, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", val.Int64()).Value("name")
	if name.String() == "" {
		return "未知项目"
	}
	return name.String()
}

func (n *ReportNotifier) sendToOwner(ctx context.Context, createdBy int64, sendFn func(openID string)) {
	if createdBy == 0 {
		return
	}
	platform := string(n.adapter.GetPlatform())
	binding, _ := n.bindingRepo.GetByUserID(ctx, createdBy, platform)
	if binding != nil {
		openID := fmt.Sprintf("%v", binding["platform_user_id"])
		if openID != "" {
			sendFn(openID)
		}
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
