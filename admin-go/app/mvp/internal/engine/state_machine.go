package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// 任务状态机：定义合法的状态转换路径，防止非法跳转
// 所有任务状态变更必须通过 updateTaskStatus() 函数，禁止裸 Update

// validTransitions 合法的任务状态转换表
// key: 当前状态, value: 允许转换到的目标状态集合
var validTransitions = map[string]map[string]bool{
	"draft": {
		"pending": true, // 确认方案时 draft → pending
	},
	"pending": {
		"running": true, // 调度器分发
		"draft":   true, // 回退到草稿（审核不通过）
	},
	"running": {
		"completed": true, // 正常完成
		"failed":    true, // 执行失败
		"escalated": true, // 升级给架构师
		"pending":   true, // 暂停时回退
	},
	"completed": {
		"auditing":  true, // 进入审计
		"bug_found": true, // 审计发现 bug
	},
	"failed": {
		"pending":   true, // 重试
		"completed": true, // 跳过（SkipTask）
		"escalated": true, // 升级
	},
	"auditing": {
		"completed": true, // 审计通过
		"bug_found": true, // 发现 bug
	},
	"bug_found": {
		"bug_dispatched": true, // 分派修复
		"pending":        true, // 直接重做
		"completed":      true, // 跳过
	},
	"bug_dispatched": {
		"pending": true, // 修复方案回写后重新排队
	},
	"escalated": {
		"pending": true, // 架构师给出方案后重新排队
		"failed":  true, // 升级后仍然失败
	},
	"submit_error": {
		"pending": true, // 重试
	},
}

// ValidateTransition 校验状态转换是否合法
// 返回 nil 表示合法，否则返回错误
func ValidateTransition(from, to string) error {
	targets, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("未知的源状态: %s", from)
	}
	if !targets[to] {
		return fmt.Errorf("非法状态转换: %s → %s", from, to)
	}
	return nil
}

// updateTaskStatus 统一的任务状态变更函数
// 内部强制调用 ValidateTransition + DB CAS（乐观锁），防止并发冲突和非法跳转
// extra: 额外需要更新的字段（如 error_message, result 等）
// 返回实际影响的行数和错误
func updateTaskStatus(ctx context.Context, taskID int64, fromStatus, toStatus string, extra g.Map) (int64, error) {
	if err := ValidateTransition(fromStatus, toStatus); err != nil {
		g.Log().Warningf(ctx, "[StateMachine] 非法状态转换被拦截: task=%d, %s → %s", taskID, fromStatus, toStatus)
		return 0, err
	}

	data := g.Map{
		"status":     toStatus,
		"updated_at": gtime.Now(),
	}
	for k, v := range extra {
		data[k] = v
	}

	result, err := g.DB().Model("mvp_task").
		Where("id", taskID).
		Where("status", fromStatus). // CAS: 只有当前状态匹配才更新
		Update(data)
	if err != nil {
		return 0, fmt.Errorf("更新任务状态失败: task=%d, %s→%s, err=%w", taskID, fromStatus, toStatus, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		g.Log().Infof(ctx, "[StateMachine] CAS 未命中: task=%d, 期望状态=%s→%s（可能已被其他路径处理）",
			taskID, fromStatus, toStatus)
	}
	return rows, nil
}
