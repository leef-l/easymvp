package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/activity"
	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// buildAiderTaskPrompt 为 Aider 构建更紧凑的任务指令，避免上下文过大触发 token limit。
func (e *Executor) buildAiderTaskPrompt(task gdb.Record, resources []string) string {
	name := task["name"].String()
	desc := task["description"].String()

	prompt := fmt.Sprintf("## 任务\n%s\n\n## 任务描述\n%s", name, desc)

	if len(resources) > 0 {
		limitedResources := resources
		if len(limitedResources) > 12 {
			limitedResources = limitedResources[:12]
		}
		prompt += "\n\n允许修改的文件或目录："
		for _, resource := range limitedResources {
			if resource == "" {
				continue
			}
			prompt += "\n- " + resource
		}
		if len(resources) > len(limitedResources) {
			prompt += fmt.Sprintf("\n- 其余 %d 个文件暂不优先处理，必要时再扩展", len(resources)-len(limitedResources))
		}
	}

	taskID := task["id"].Int64()
	deps, depsErr := g.DB().Model("mvp_task_dependency d").
		LeftJoin("mvp_task t", "t.id = d.depends_on_id").
		Fields("t.name, t.result").
		Where("d.task_id", taskID).
		Where("t.status", "completed").
		Limit(2).
		All()
	if depsErr != nil {
		g.Log().Warningf(context.Background(), "[ExecutorCLI] 查询依赖任务失败: taskID=%d err=%v", taskID, depsErr)
	}

	if len(deps) > 0 {
		prompt += "\n\n前置结果摘要："
		for _, dep := range deps {
			depName := dep["name"].String()
			depResult := dep["result"].String()
			if len(depResult) > 800 {
				depResult = depResult[:800] + "...(截断)"
			}
			prompt += fmt.Sprintf("\n- %s：%s", depName, depResult)
		}
	}

	prompt += "\n\n执行约束：只允许修改上面列出的文件或目录；不要输出「1. 标题(路径)」或其他 Markdown 章节标题来描述文件；不要把说明标题、编号、括号说明当成文件名；请直接完成最小必要改动。"
	return prompt
}

// executeWithAider 使用 Aider 执行实施类任务（真实代码编辑）
func (e *Executor) executeWithAider(ctx context.Context, projectID int64, taskID int64, task gdb.Record, modelInfo *ModelInfo) {
	// 1. 查询项目获取工作目录
	project, err := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).WhereNull("deleted_at").Fields("work_dir").One()
	if err != nil || project.IsEmpty() {
		e.failTask(ctx, projectID, taskID, "项目不存在")
		return
	}

	// 工作目录：优先项目配置的 work_dir，兜底用默认
	workDir := project["work_dir"].String()
	if workDir == "" {
		workDir = "/www/wwwroot/project/easymvp"
	}

	// 2. 解析 affected_resources 作为需要编辑的文件
	resourceResult := parseResourcesDetail(task["affected_resources"].String())
	if len(resourceResult.Rejected) > 0 {
		e.escalateImplementerResourceIssue(ctx, projectID, taskID, task, resourceResult)
		return
	}
	resources := resourceResult.Resources

	// 3. 构建更紧凑的 Aider prompt，避免一次性塞入过多上下文
	taskPrompt := e.buildAiderTaskPrompt(task, resources)

	// 4. 创建对话记录（用于前端展示 Aider 过程）
	conversationID, convErr := e.ensureConversation(ctx, projectID, taskID, "implementer")
	if convErr != nil {
		e.handleTaskFailure(ctx, projectID, taskID, "implementer", taskFailureExecution, "创建对话失败: "+convErr.Error())
		return
	}
	if _, upErr := g.DB().Model("mvp_task").Ctx(ctx).Where("id", taskID).Update(g.Map{
		"conversation_id": conversationID,
	}); upErr != nil {
		g.Log().Warningf(ctx, "[Executor] 回写 conversation_id 失败: task=%d err=%v", taskID, upErr)
	}
	activity.TouchTaskActivity(ctx, taskID)
	activity.TouchConversationActivity(ctx, conversationID)

	// 保存指令消息
	userMsgID := int64(snowflake.Generate())
	if _, err := g.DB().Model("mvp_message").Insert(g.Map{
		"id":              userMsgID,
		"conversation_id": conversationID,
		"role":            "user",
		"message_type":    mvpmodel.MessageTypeTaskPrompt,
		"content":         taskPrompt,
		"status":          "completed",
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Warningf(ctx, "[Executor] 保存指令消息失败: task=%d err=%v", taskID, err)
	}

	// 5. 文件快照（执行前）：记录 affected_resources 的 mtime+size，用于检测假成功
	beforeSnap := captureFileSnapshots(workDir, resources)

	// 6. 调用 Aider（先更新心跳，Aider 执行可能较长）
	TouchHeartbeat(ctx, taskID)
	g.Log().Infof(ctx, "[Executor] 任务 %d 使用 Aider 执行: model=%s files=%v", taskID, modelInfo.ModelCode, resources)

	result := GetAiderRunner().RunTask(ctx, projectID, taskID, modelInfo, taskPrompt, workDir, resources, nil)

	// 7. 保存 Aider 输出为 AI 回复消息
	replyID := int64(snowflake.Generate())
	replyStatus := "completed"
	if result.Error != nil {
		replyStatus = "failed"
	}
	if _, err := g.DB().Model("mvp_message").Insert(g.Map{
		"id":              replyID,
		"conversation_id": conversationID,
		"role":            "assistant",
		"message_type":    map[bool]string{true: mvpmodel.MessageTypePoison, false: mvpmodel.MessageTypeTaskReply}[result.Error != nil],
		"content":         result.Output,
		"model_id":        modelInfo.ModelID,
		"status":          replyStatus,
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Warningf(ctx, "[Executor] 保存AI回复消息失败: task=%d err=%v", taskID, err)
	}

	// 8. 判断结果
	if result.Error != nil {
		errMsg := result.FailureHint
		if errMsg == "" {
			errMsg = result.Error.Error()
		}
		category := result.Category
		if category == "" {
			category = taskFailureExecution
		}
		e.handleTaskFailure(ctx, projectID, taskID, task["role_type"].String(), category, fmt.Sprintf("Aider执行失败(code=%d): %s", result.ExitCode, errMsg))
		return
	}

	// 9. 文件变更检测（防假成功）
	afterSnap := captureFileSnapshots(workDir, resources)
	if changedCount := diffSnapshots(beforeSnap, afterSnap); changedCount == 0 && len(resources) > 0 {
		g.Log().Warningf(ctx, "[Executor] Aider 报成功但零文件变更: task=%d, files=%v", taskID, resources)
		// 标记为 suspicious_success，任务仍完成但审计员会重点关注
		updateTaskStatus(ctx, taskID, "running", "completed", g.Map{
			"result":       result.Output + "\n\n⚠️ [系统检测] Aider 报告成功但未检测到文件变更，请审计员重点审核。",
			"completed_at": gtime.Now(),
		})
	} else {
		// 10. 正常更新任务为完成
		updateTaskStatus(ctx, taskID, "running", "completed", g.Map{
			"result":       result.Output,
			"completed_at": gtime.Now(),
		})
	}

	// 11. 压缩上下文（同步执行，确保不丢失）
	compressCtx, compressCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer compressCancel()
	if err := GetCompressor().CompressTaskContext(compressCtx, projectID, taskID); err != nil {
		g.Log().Errorf(ctx, "[Executor] Aider任务压缩上下文失败（非致命）: task=%d, err=%v", taskID, err)
	}

	// 创建审计任务（同步执行，确保一定创建）
	e.createAuditTask(ctx, projectID, taskID, task)

	// 通知调度器
	e.scheduler.OnTaskCompleted(projectID, taskID)
}

// escalateImplementerResourceIssue 资源解析异常时升级到架构师
func (e *Executor) escalateImplementerResourceIssue(ctx context.Context, projectID int64, taskID int64, task gdb.Record, resourceResult resourceParseResult) {
	errMsg := fmt.Sprintf(
		"任务涉及的 affected_resources 存在歧义，无法安全继续执行。可修复资源=%v；无法解析资源=%v；原始值=%s。请架构师重新检查任务拆分和 affected_resources。",
		resourceResult.Resources,
		resourceResult.Rejected,
		task["affected_resources"].String(),
	)

	conversationID, err := e.ensureConversation(ctx, projectID, taskID, "implementer")
	if err == nil {
		replyID := int64(snowflake.Generate())
		if _, insErr := g.DB().Model("mvp_message").Insert(g.Map{
			"id":              replyID,
			"conversation_id": conversationID,
			"role":            "assistant",
			"message_type":    mvpmodel.MessageTypePoison,
			"content":         errMsg,
			"status":          "failed",
			"created_by":      0,
			"dept_id":         0,
			"created_at":      gtime.Now(),
			"updated_at":      gtime.Now(),
		}); insErr != nil {
			g.Log().Warningf(ctx, "[ExecutorCLI] 保存错误消息失败: conv=%d err=%v", conversationID, insErr)
		}
	}

	updateTaskStatus(ctx, taskID, "running", "escalated", g.Map{
		"error_message": errMsg,
	})

	e.scheduler.OnTaskEscalated(projectID, taskID, errMsg)
	roleType := task["role_type"].String()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[ExecutorCLI] EscalateFailedTask panic: project=%d task=%d err=%v", projectID, taskID, r)
			}
		}()
		e.scheduler.EscalateFailedTask(e.scheduler.getProjectContext(projectID), projectID, taskID, roleType, errMsg)
	}()
}

// --- 文件快照：检测 Aider 假成功 ---

type fileSnapshot struct {
	Path    string
	ModTime time.Time
	Size    int64
	Exists  bool
}

// captureFileSnapshots 记录文件列表的 mtime 和 size
func captureFileSnapshots(workDir string, resources []string) []fileSnapshot {
	snaps := make([]fileSnapshot, 0, len(resources))
	for _, res := range resources {
		absPath := res
		if !filepath.IsAbs(res) {
			absPath = filepath.Join(workDir, res)
		}
		info, err := os.Stat(absPath)
		if err != nil {
			snaps = append(snaps, fileSnapshot{Path: absPath, Exists: false})
		} else {
			snaps = append(snaps, fileSnapshot{
				Path:    absPath,
				ModTime: info.ModTime(),
				Size:    info.Size(),
				Exists:  true,
			})
		}
	}
	return snaps
}

// diffSnapshots 对比前后快照，返回有变化的文件数
func diffSnapshots(before, after []fileSnapshot) int {
	if len(before) != len(after) {
		return 1 // 长度不同说明有变化
	}
	changed := 0
	for i := range before {
		b, a := before[i], after[i]
		if b.Exists != a.Exists {
			changed++
		} else if b.Exists && a.Exists {
			if b.ModTime != a.ModTime || b.Size != a.Size {
				changed++
			}
		}
	}
	return changed
}
