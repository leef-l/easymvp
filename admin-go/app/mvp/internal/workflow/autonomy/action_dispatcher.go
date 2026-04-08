package autonomy

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/repo"
)

// ActionCallback 动作回调函数签名。
type ActionCallback func(ctx context.Context, req *DecisionRequest) error

// ActionDispatcher 动作执行器：按 ActionType 路由到对应回调，回写执行结果。
type ActionDispatcher struct {
	actionRepo *repo.DecisionActionRepo
	callbacks  map[string]ActionCallback
}

// NewActionDispatcher 创建动作执行器。
func NewActionDispatcher(actionRepo *repo.DecisionActionRepo) *ActionDispatcher {
	return &ActionDispatcher{
		actionRepo: actionRepo,
		callbacks:  make(map[string]ActionCallback),
	}
}

// SetCallback 注册指定动作类型的回调。通过回调注入避免循环依赖。
func (d *ActionDispatcher) SetCallback(actionType string, fn ActionCallback) {
	d.callbacks[actionType] = fn
}

// Execute 执行动作：路由到回调 → 回写失败状态/结果。
// 注意：成功时不写状态，由调用方（DecisionCenter）根据场景写 auto_executed 或保持 approved。
func (d *ActionDispatcher) Execute(ctx context.Context, actionID int64, actionType string, req *DecisionRequest) error {
	fn, ok := d.callbacks[actionType]
	if !ok {
		errMsg := fmt.Sprintf("未注册动作回调: %s", actionType)
		g.Log().Warningf(ctx, "[ActionDispatcher] %s actionID=%d", errMsg, actionID)
		if upErr := d.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusFailed, g.Map{
			"result": errMsg,
		}); upErr != nil {
			g.Log().Warningf(ctx, "[ActionDispatcher] 更新失败状态也失败: actionID=%d err=%v", actionID, upErr)
		}
		return fmt.Errorf("%s", errMsg)
	}

	g.Log().Infof(ctx, "[ActionDispatcher] 开始执行动作: actionID=%d type=%s wfRun=%d",
		actionID, actionType, req.WorkflowRunID)

	err := fn(ctx, req)
	if err != nil {
		g.Log().Warningf(ctx, "[ActionDispatcher] 动作执行失败: actionID=%d type=%s err=%v",
			actionID, actionType, err)
		if upErr := d.actionRepo.UpdateStatus(ctx, actionID, consts.ActionStatusFailed, g.Map{
			"result":       err.Error(),
			"final_action": actionType,
		}); upErr != nil {
			g.Log().Warningf(ctx, "[ActionDispatcher] 更新失败状态也失败: actionID=%d err=%v", actionID, upErr)
		}
		return err
	}

	g.Log().Infof(ctx, "[ActionDispatcher] 动作执行成功: actionID=%d type=%s", actionID, actionType)
	return nil
}
