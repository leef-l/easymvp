package notifier

// feishu_notifier.go
// 飞书主动推送服务。负责在引擎各事件点将消息主动推送给绑定了飞书的用户。
// 使用方式：在 engine 层事件回调中调用 GetNotifier().NotifyXxx()

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/collab/adapter"
)

var (
	once     sync.Once
	instance *FeishuNotifier
)

// FeishuNotifier 飞书推送服务单例。
type FeishuNotifier struct{}

// GetNotifier 返回全局单例。
func GetNotifier() *FeishuNotifier {
	once.Do(func() { instance = &FeishuNotifier{} })
	return instance
}

// ─── 对话类推送 ────────────────────────────────────────────────────────────────

// NotifyAIReply 将 AI 回复内容推送到飞书用户。
// 在 runAICall 完成后调用，conversationID 用于查找绑定的飞书 open_id。
func (n *FeishuNotifier) NotifyAIReply(ctx context.Context, conversationID int64, aiContent string) {
	if aiContent == "" {
		return
	}
	openIDs := n.getConversationOpenIDs(ctx, conversationID)
	if len(openIDs) == 0 {
		return
	}

	// 飞书单条消息上限约 4096 字符，超长截断
	text := aiContent
	if len([]rune(text)) > 1000 {
		runes := []rune(text)
		text = string(runes[:1000]) + "\n\n…（内容较长，完整内容请在 EasyMVP 管理端查看）"
	}

	feishu := adapter.NewFeishuAdapter()
	for _, openID := range openIDs {
		if err := feishu.SendTextMessage(ctx, openID, "💬 AI 回复：\n"+text); err != nil {
			g.Log().Warningf(ctx, "[FeishuNotifier] 推送 AI 回复失败 openID=%s: %v", openID, err)
		}
	}
}

// ─── 任务类推送 ────────────────────────────────────────────────────────────────

// NotifyTaskFailed 任务失败时推送给项目创建人。
func (n *FeishuNotifier) NotifyTaskFailed(ctx context.Context, projectID, taskID int64, taskName, errMsg string) {
	openID := n.getProjectOwnerOpenID(ctx, projectID)
	if openID == "" {
		return
	}
	projectName := n.getProjectName(ctx, projectID)
	text := fmt.Sprintf(
		"❌ 任务执行失败\n───────────────\n📁 项目：%s\n📋 任务：%s（ID:%d）\n💬 原因：%s\n\n可回复「重试失败任务 %s」让我重新执行",
		projectName, taskName, taskID, truncate(errMsg, 200), projectName,
	)
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, openID, text); err != nil {
		g.Log().Warningf(ctx, "[FeishuNotifier] 推送任务失败通知失败: %v", err)
	}
}

// NotifyProjectCompleted 项目全部任务完成时推送。
func (n *FeishuNotifier) NotifyProjectCompleted(ctx context.Context, projectID int64) {
	openID := n.getProjectOwnerOpenID(ctx, projectID)
	if openID == "" {
		return
	}
	projectName := n.getProjectName(ctx, projectID)

	// 统计任务数
	total, _ := g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).WhereNull("deleted_at").Count()
	done, _ := g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).Where("status", "completed").WhereNull("deleted_at").Count()

	text := fmt.Sprintf(
		"🎉 项目执行完成！\n───────────────\n📁 项目：%s\n✅ 完成任务：%d / %d\n\n可在 EasyMVP 管理端查看完整报告",
		projectName, done, total,
	)
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, openID, text); err != nil {
		g.Log().Warningf(ctx, "[FeishuNotifier] 推送项目完成通知失败: %v", err)
	}
}

// ─── 审核/验收类推送 ────────────────────────────────────────────────────────────

// NotifyReviewNeeded 需要人工审核时推送。
func (n *FeishuNotifier) NotifyReviewNeeded(ctx context.Context, projectID int64, issueCount int) {
	openID := n.getProjectOwnerOpenID(ctx, projectID)
	if openID == "" {
		return
	}
	projectName := n.getProjectName(ctx, projectID)
	text := fmt.Sprintf(
		"🔍 需要您人工审核\n───────────────\n📁 项目：%s\n⚠️ 问题数：%d 个\n\n回复「通过审核 %s」或「驳回审核 %s」",
		projectName, issueCount, projectName, projectName,
	)
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, openID, text); err != nil {
		g.Log().Warningf(ctx, "[FeishuNotifier] 推送审核通知失败: %v", err)
	}
}

// NotifyAcceptNeeded 需要人工验收时推送。
func (n *FeishuNotifier) NotifyAcceptNeeded(ctx context.Context, projectID int64, passedCount, failedCount int) {
	openID := n.getProjectOwnerOpenID(ctx, projectID)
	if openID == "" {
		return
	}
	projectName := n.getProjectName(ctx, projectID)
	text := fmt.Sprintf(
		"🎯 需要您人工验收\n───────────────\n📁 项目：%s\n✅ 通过：%d  ❌ 失败：%d\n\n回复「验收通过 %s」或「验收驳回 %s」",
		projectName, passedCount, failedCount, projectName, projectName,
	)
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, openID, text); err != nil {
		g.Log().Warningf(ctx, "[FeishuNotifier] 推送验收通知失败: %v", err)
	}
}

// NotifyCheckpointNeeded 自治检查点需要人工介入时推送。
func (n *FeishuNotifier) NotifyCheckpointNeeded(ctx context.Context, projectID int64, checkType, description string) {
	openID := n.getProjectOwnerOpenID(ctx, projectID)
	if openID == "" {
		return
	}
	projectName := n.getProjectName(ctx, projectID)
	text := fmt.Sprintf(
		"🤖 自治检查点需确认\n───────────────\n📁 项目：%s\n📌 类型：%s\n📝 说明：%s\n\n回复「批准检查点 %s」或「拒绝检查点 %s」",
		projectName, checkType, truncate(description, 200), projectName, projectName,
	)
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, openID, text); err != nil {
		g.Log().Warningf(ctx, "[FeishuNotifier] 推送检查点通知失败: %v", err)
	}
}

// ─── 工具函数 ──────────────────────────────────────────────────────────────────

// getConversationOpenIDs 根据 conversationID 找到相关用户的飞书 open_id。
// 优先找对话创建人的绑定，其次找项目创建人的绑定。
func (n *FeishuNotifier) getConversationOpenIDs(ctx context.Context, conversationID int64) []string {
	conv, err := g.DB().Ctx(ctx).Model("mvp_conversation").
		Where("id", conversationID).WhereNull("deleted_at").
		Fields("created_by, project_id").One()
	if err != nil || conv.IsEmpty() {
		return nil
	}

	userID := conv["created_by"].Int64()
	if userID == 0 {
		// 降级：用项目创建人
		proj, _ := g.DB().Ctx(ctx).Model("mvp_project").
			Where("id", conv["project_id"].Int64()).Fields("created_by").One()
		if !proj.IsEmpty() {
			userID = proj["created_by"].Int64()
		}
	}
	if userID == 0 {
		return nil
	}

	openID := n.getUserOpenID(ctx, userID)
	if openID == "" {
		return nil
	}
	return []string{openID}
}

// getProjectOwnerOpenID 获取项目创建人的飞书 open_id。
func (n *FeishuNotifier) getProjectOwnerOpenID(ctx context.Context, projectID int64) string {
	proj, err := g.DB().Ctx(ctx).Model("mvp_project").
		Where("id", projectID).WhereNull("deleted_at").
		Fields("created_by").One()
	if err != nil || proj.IsEmpty() {
		return ""
	}
	return n.getUserOpenID(ctx, proj["created_by"].Int64())
}

// getUserOpenID 根据系统 userID 查飞书 open_id。
func (n *FeishuNotifier) getUserOpenID(ctx context.Context, userID int64) string {
	if userID == 0 {
		return ""
	}
	binding, err := g.DB().Ctx(ctx).Model("mvp_user_collab_binding").
		Where("user_id", userID).
		Where("platform", "feishu").
		WhereNull("deleted_at").
		Fields("platform_user_id").One()
	if err != nil || binding.IsEmpty() {
		return ""
	}
	return binding["platform_user_id"].String()
}

// getProjectName 获取项目名称。
func (n *FeishuNotifier) getProjectName(ctx context.Context, projectID int64) string {
	proj, _ := g.DB().Ctx(ctx).Model("mvp_project").
		Where("id", projectID).Fields("name").One()
	if proj.IsEmpty() {
		return fmt.Sprintf("项目%d", projectID)
	}
	return proj["name"].String()
}

func truncate(s string, maxLen int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}
