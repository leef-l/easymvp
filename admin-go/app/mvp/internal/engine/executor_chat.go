package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/provider"
	"easymvp/utility/snowflake"
)

// ensureConversation 确保任务有对应的对话
func (e *Executor) ensureConversation(ctx context.Context, projectID int64, taskID int64, roleType string) (int64, error) {
	// 查找已有的任务对话
	conv, err := g.DB().Ctx(ctx).Model("mvp_conversation").
		Where("project_id", projectID).
		Where("task_id", taskID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return 0, err
	}
	if !conv.IsEmpty() {
		return conv["id"].Int64(), nil
	}

	// 创建新对话
	project, err := g.DB().Ctx(ctx).Model("mvp_project").
		Fields("created_by, dept_id").
		Where("id", projectID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return 0, err
	}
	if project.IsEmpty() {
		return 0, fmt.Errorf("项目不存在")
	}

	convID := int64(snowflake.Generate())
	_, err = g.DB().Ctx(ctx).Model("mvp_conversation").Insert(g.Map{
		"id":         convID,
		"project_id": projectID,
		"task_id":    taskID,
		"title":      "任务对话",
		"role_type":  roleType,
		"status":     "active",
		"created_by": project["created_by"].Int64(),
		"dept_id":    project["dept_id"].Int64(),
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return 0, err
	}
	return convID, nil
}

// loadConversationHistory 加载对话历史
func (e *Executor) loadConversationHistory(ctx context.Context, conversationID int64, excludeID int64) ([]provider.Message, error) {
	records, err := g.DB().Ctx(ctx).Model("mvp_message").
		Where("conversation_id", conversationID).
		Where("deleted_at IS NULL").
		Where("status", "completed").
		Where("(message_type IS NULL OR message_type <> ?)", mvpmodel.MessageTypePoison).
		Where("id != ?", excludeID).
		Order("created_at ASC").
		All()
	if err != nil {
		return nil, err
	}

	messages := make([]provider.Message, 0, len(records))
	for _, r := range records {
		messages = append(messages, provider.Message{
			Role:    provider.Role(r["role"].String()),
			Content: r["content"].String(),
		})
	}
	return messages, nil
}

// buildTaskPrompt 构建任务指令
func (e *Executor) buildTaskPrompt(task gdb.Record) string {
	name := task["name"].String()
	desc := task["description"].String()

	prompt := fmt.Sprintf("任务名称：%s\n任务描述：%s", name, desc)

	// 如果有依赖任务的结果，附加上下文
	taskID := task["id"].Int64()
	deps, _ := g.DB().Ctx(context.Background()).Model("mvp_task_dependency d").
		LeftJoin("mvp_task t", "t.id = d.depends_on_id").
		Fields("t.name, t.result").
		Where("d.task_id", taskID).
		Where("t.status", "completed").
		All()

	if len(deps) > 0 {
		prompt += "\n\n## 前置任务结果（供参考）"
		for _, dep := range deps {
			depName := dep["name"].String()
			depResult := dep["result"].String()
			if len(depResult) > 2000 {
				depResult = depResult[:2000] + "...(截断)"
			}
			prompt += fmt.Sprintf("\n\n### %s\n%s", depName, depResult)
		}
	}

	return prompt
}

// createAuditTask 为实施员任务创建对应的审计任务
// 使用依赖表的唯一索引（uk_dep）做幂等保护，防止并发重复创建
func (e *Executor) createAuditTask(ctx context.Context, projectID int64, implTaskID int64, implTask gdb.Record) {
	// 检查是否已有审计任务（通过依赖关系）
	count, _ := g.DB().Ctx(ctx).Model("mvp_task_dependency").
		Where("depends_on_id", implTaskID).
		Count()
	if count > 0 {
		return
	}

	auditTaskID := int64(snowflake.Generate())

	// 链路字段：root_task_id 继承自实施任务，兼容旧数据
	rootTaskID := implTask["root_task_id"].Int64()
	if rootTaskID == 0 {
		rootTaskID = implTaskID
	}

	if _, err := g.DB().Ctx(ctx).Model("mvp_task").Insert(g.Map{
		"id":             auditTaskID,
		"project_id":     projectID,
		"parent_id":      implTask["parent_id"].Int64(),
		"name":           fmt.Sprintf("审计: %s", implTask["name"].String()),
		"description":    fmt.Sprintf("审计实施员任务「%s」的结果，检查是否正确完成，是否有 bug。", implTask["name"].String()),
		"role_type":      "auditor",
		"role_level":     implTask["role_level"].String(),
		"task_kind":      consts.TaskKindAudit,
		"source_task_id": implTaskID,
		"root_task_id":   rootTaskID,
		"status":         "pending",
		"batch_no":       implTask["batch_no"].Int() + 1,
		"created_by":     0,
		"dept_id":        0,
		"created_at":     gtime.Now(),
		"updated_at":     gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 创建审计任务失败: implTask=%d, err=%v", implTaskID, err)
		return
	}

	// 添加依赖关系（uk_dep 唯一索引保证幂等，重复插入会静默失败）
	if _, err := g.DB().Ctx(ctx).Model("mvp_task_dependency").Insert(g.Map{
		"task_id":       auditTaskID,
		"depends_on_id": implTaskID,
	}); err != nil {
		// 唯一索引冲突说明已被并发创建，清理多余的审计任务
		g.Log().Warningf(ctx, "[Executor] 审计任务依赖已存在（并发重复），回滚: implTask=%d, err=%v", implTaskID, err)
		if _, delErr := g.DB().Ctx(ctx).Model("mvp_task").Where("id", auditTaskID).Delete(); delErr != nil {
			g.Log().Errorf(ctx, "[Executor] 回滚审计任务失败: auditTask=%d err=%v", auditTaskID, delErr)
		}
	}
}
