package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/provider"
)

// CoordinatorOptResult 协调员 AI 优化输出
type CoordinatorOptResult struct {
	OptimizedBatches map[string]struct {
		BatchNo int    `json:"batch_no"`
		Reason  string `json:"reason"`
	} `json:"optimized_batches"`
	ParallelismScore  float64  `json:"parallelism_score"`
	EstimatedDuration string   `json:"estimated_duration"`
	Warnings          []string `json:"warnings"`
}

// coordinatorOptimize 协调员 AI 优化（legacy 入口）
func coordinatorOptimize(ctx context.Context, projectID int64, tasks gdb.Result) (*CoordinatorOptResult, error) {
	modelInfo, err := getReviewRoleModel(ctx, projectID, "coordinator")
	if err != nil {
		return nil, fmt.Errorf("获取协调员模型失败: %w", err)
	}
	return doCoordinatorOptimize(ctx, modelInfo, tasks)
}

// doCoordinatorOptimize 协调员 AI 优化核心逻辑。
// doCoordinatorOptimize 协调员 AI 优化核心逻辑。
func doCoordinatorOptimize(ctx context.Context, modelInfo *ModelInfo, tasks gdb.Result) (*CoordinatorOptResult, error) {
	taskSummaries := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		taskSummaries = append(taskSummaries, map[string]interface{}{
			"name":               t["name"].String(),
			"batch_no":           t["batch_no"].Int(),
			"affected_resources": t["affected_resources"].String(),
			"depends_on":         t["depends_on"].String(),
			"role_type":          t["role_type"].String(),
		})
	}
	summaryJSON, mErr := json.MarshalIndent(taskSummaries, "", "  ")
	if mErr != nil {
		return nil, fmt.Errorf("序列化任务清单失败: %w", mErr)
	}

	prompt := fmt.Sprintf(`请优化以下任务清单的调度计划。优化维度：
1. 资源冲突精细检测（同批次任务不应修改相同文件）
2. 批次顺序优化（最大化并行度）
3. 并行度评估
4. 预估执行时间

任务清单（共 %d 个）：
%s

请严格输出 JSON，格式如下：
{"optimized_batches": {"任务名": {"batch_no": 1, "reason": "调整原因"}}, "parallelism_score": 0.8, "estimated_duration": "约2小时", "warnings": []}`, len(tasks), string(summaryJSON))

	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.3,
		SystemPrompt: modelInfo.SystemPrompt,
	})
	if err != nil {
		return nil, err
	}

	var optResult CoordinatorOptResult
	if err := parseJSONFromAI(resp.Content, &optResult); err != nil {
		return nil, fmt.Errorf("解析协调员优化结果失败: %w", err)
	}

	return &optResult, nil
}

// applyCoordinatorOptimizations 应用协调员的批次优化建议（事务内批量执行）
func applyCoordinatorOptimizations(ctx context.Context, projectID int64, tasks gdb.Result, opt *CoordinatorOptResult) {
	if opt == nil || len(opt.OptimizedBatches) == 0 {
		return
	}

	// 先收集需要更新的任务
	type batchUpdate struct {
		taskID   int64
		newBatch int
	}
	var updates []batchUpdate
	for _, t := range tasks {
		name := t["name"].String()
		if batch, ok := opt.OptimizedBatches[name]; ok {
			if batch.BatchNo > 0 && batch.BatchNo != t["batch_no"].Int() {
				updates = append(updates, batchUpdate{taskID: t["id"].Int64(), newBatch: batch.BatchNo})
			}
		}
	}

	if len(updates) == 0 {
		return
	}

	// 事务内批量更新
	if err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, u := range updates {
			if _, err := tx.Model("mvp_task").Where("id", u.taskID).Update(g.Map{
				"batch_no":   u.newBatch,
				"updated_at": gtime.Now(),
			}); err != nil {
				return fmt.Errorf("更新 task=%d batch_no 失败: %w", u.taskID, err)
			}
		}
		return nil
	}); err != nil {
		g.Log().Errorf(ctx, "[Review] 协调员批次优化事务失败: project=%d, err=%v", projectID, err)
		return
	}

	g.Log().Infof(ctx, "[Review] 协调员优化：应用了 %d 个批次调整", len(updates))
}
