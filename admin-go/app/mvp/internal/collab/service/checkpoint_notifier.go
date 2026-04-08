package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/collab"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/event"
)

// CheckpointNotifier 审批/告警通知服务，监听自治事件并推送到协作平台。
type CheckpointNotifier struct {
	adapter     collab.MessageAdapter
	bindingRepo *collabRepo.BindingRepo
}

// NewCheckpointNotifier 创建通知服务。
func NewCheckpointNotifier(adapter collab.MessageAdapter, bindingRepo *collabRepo.BindingRepo) *CheckpointNotifier {
	return &CheckpointNotifier{adapter: adapter, bindingRepo: bindingRepo}
}

// OnCheckpointOpened 人工节点创建时推送审批卡片。
func (n *CheckpointNotifier) OnCheckpointOpened(evt event.Event) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[CheckpointNotifier] OnCheckpointOpened panic: %v", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if !n.adapter.IsEnabled(ctx) {
			return
		}

		payload, ok := evt.Payload.(g.Map)
		if !ok {
			return
		}

		actionID := toInt64(payload["action_id"])
		if actionID == 0 {
			return
		}

		// 查决策动作详情
		action, err := g.DB().Model("mvp_decision_action").Ctx(ctx).
			Where("id", actionID).WhereNull("deleted_at").One()
		if err != nil || action.IsEmpty() {
			g.Log().Warningf(ctx, "[CheckpointNotifier] 查询决策动作失败: actionID=%d err=%v", actionID, err)
			return
		}

		projectID := action["project_id"].Int64()
		project, projErr := g.DB().Model("mvp_project").Ctx(ctx).
			Where("id", projectID).Fields("name, created_by").One()
		if projErr != nil {
			g.Log().Warningf(ctx, "[CheckpointNotifier] 查询项目失败: projectID=%d err=%v", projectID, projErr)
		}

		projectName := "未知项目"
		var createdBy int64
		if !project.IsEmpty() {
			projectName = project["name"].String()
			createdBy = project["created_by"].Int64()
		}

		level := action["decision_level"].String()
		headerColor := "orange"
		if level == "C" {
			headerColor = "red"
		}

		card := &collab.InteractiveCard{
			Title:         "自治决策待审批",
			Level:         level,
			ActionType:    action["action_type"].String(),
			ActionID:      actionID,
			TriggerSource: action["trigger_source"].String(),
			ProjectName:   projectName,
			HeaderColor:   headerColor,
			Buttons: []collab.CardButton{
				{Label: "批准", Action: "approve", Style: "primary"},
				{Label: "驳回", Action: "reject", Style: "danger"},
			},
		}

		n.sendToProjectOwner(ctx, createdBy, func(openID string) {
			if err := n.adapter.SendCardMessage(ctx, openID, card); err != nil {
				g.Log().Warningf(ctx, "[CheckpointNotifier] 推送审批卡片失败: openID=%s err=%v", openID, err)
			}
		})
	}()
}

// OnActionFailed 动作执行失败时推送告警。
func (n *CheckpointNotifier) OnActionFailed(evt event.Event) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[CheckpointNotifier] OnActionFailed panic: %v", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if !n.adapter.IsEnabled(ctx) {
			return
		}

		payload, ok := evt.Payload.(g.Map)
		if !ok {
			return
		}

		errMsg := fmt.Sprintf("%v", payload["error"])
		text := fmt.Sprintf("[自治告警] 决策动作执行失败\n工作流: %d\n错误: %s", evt.WorkflowRunID, errMsg)

		createdBy := n.lookupProjectOwner(ctx, evt.WorkflowRunID)
		n.sendToProjectOwner(ctx, createdBy, func(openID string) {
			if err := n.adapter.SendTextMessage(ctx, openID, text); err != nil {
				g.Log().Warningf(ctx, "[CheckpointNotifier] 推送失败告警失败: err=%v", err)
			}
		})
	}()
}

// OnGateBlocked 闸门阻断时推送告警。
func (n *CheckpointNotifier) OnGateBlocked(evt event.Event) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[CheckpointNotifier] OnGateBlocked panic: %v", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if !n.adapter.IsEnabled(ctx) {
			return
		}

		payload, ok := evt.Payload.(g.Map)
		if !ok {
			return
		}

		text := fmt.Sprintf("[风险闸门] 动作被阻断\n工作流: %d\n详情: %v", evt.WorkflowRunID, payload)

		createdBy := n.lookupProjectOwner(ctx, evt.WorkflowRunID)
		n.sendToProjectOwner(ctx, createdBy, func(openID string) {
			if err := n.adapter.SendTextMessage(ctx, openID, text); err != nil {
				g.Log().Warningf(ctx, "[CheckpointNotifier] 推送闸门告警失败: err=%v", err)
			}
		})
	}()
}

// sendToProjectOwner 向项目负责人推送消息，降级到默认通知用户。
func (n *CheckpointNotifier) sendToProjectOwner(ctx context.Context, createdBy int64, sendFn func(openID string)) {
	platform := string(n.adapter.GetPlatform())

	// 优先发给项目创建人
	if createdBy > 0 {
		binding, _ := n.bindingRepo.GetByUserID(ctx, createdBy, platform)
		if binding != nil {
			openID := fmt.Sprintf("%v", binding["platform_user_id"])
			if openID != "" {
				sendFn(openID)
				return
			}
		}
	}

	// 降级到默认通知用户
	defaultIDs := engine.GetConfigString(ctx,
		"workflow.collab.feishu_default_notify_user_ids",
		"workflow.collab.feishuDefaultNotifyUserIds", "")
	if defaultIDs == "" {
		g.Log().Warningf(ctx, "[CheckpointNotifier] 无绑定用户且无默认通知用户，跳过推送")
		return
	}

	for _, idStr := range strings.Split(defaultIDs, ",") {
		idStr = strings.TrimSpace(idStr)
		userID, _ := strconv.ParseInt(idStr, 10, 64)
		if userID == 0 {
			continue
		}
		binding, bindErr := n.bindingRepo.GetByUserID(ctx, userID, platform)
		if bindErr != nil {
			g.Log().Warningf(ctx, "[CheckpointNotifier] 查询绑定失败: userID=%d err=%v", userID, bindErr)
			continue
		}
		if binding != nil {
			openID := fmt.Sprintf("%v", binding["platform_user_id"])
			if openID != "" {
				sendFn(openID)
			}
		}
	}
}

// lookupProjectOwner 从工作流ID反查项目创建人。
func (n *CheckpointNotifier) lookupProjectOwner(ctx context.Context, workflowRunID int64) int64 {
	val, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	if err != nil || val.Int64() == 0 {
		return 0
	}
	projectID := val.Int64()
	createdBy, cbErr := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Value("created_by")
	if cbErr != nil {
		g.Log().Warningf(ctx, "[CheckpointNotifier] 查询 created_by 失败: projectID=%d err=%v", projectID, cbErr)
		return 0
	}
	return createdBy.Int64()
}

func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
}
