package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

type ArchitectTaskPatch struct {
	BlueprintID       int64    `json:"blueprint_id,omitempty"`
	TaskName          string   `json:"task_name,omitempty"`
	Description       string   `json:"description,omitempty"`
	RoleType          string   `json:"role_type,omitempty"`
	RoleLevel         string   `json:"role_level,omitempty"`
	BatchNo           *int     `json:"batch_no,omitempty"`
	Sort              *int     `json:"sort,omitempty"`
	AffectedResources []string `json:"affected_resources,omitempty"`
	DependsOn         []string `json:"depends_on,omitempty"`
	Reason            string   `json:"reason,omitempty"`
}

type architectTaskPatchEnvelope struct {
	TaskPatches []ArchitectTaskPatch `json:"task_patches"`
}

type BlueprintPatchApplier func(ctx context.Context, projectID, workflowRunID, conversationID, messageID int64, patches []ArchitectTaskPatch) (planVersionID int64, patchedCount int, err error)

var blueprintPatchApplierFn BlueprintPatchApplier

type ArchitectReviewResubmitter func(ctx context.Context, projectID int64) error

var architectReviewResubmitterFn ArchitectReviewResubmitter

const architectAutoContinueToken = "[AUTO_CONTINUE_NEXT]"
const architectFollowUpLimitDefault = 24

func normalizeArchitectFollowUpLimit(limit int) int {
	if limit <= 0 {
		return architectFollowUpLimitDefault
	}
	return limit
}

var loadArchitectFollowUpLimit = func(ctx context.Context) int {
	if ctx == nil {
		ctx = context.Background()
	}
	return normalizeArchitectFollowUpLimit(GetConfigInt(ctx,
		"workflow.architect.follow_up_limit",
		"engine.workflow.architect.followUpLimit",
		architectFollowUpLimitDefault,
	))
}

func RegisterBlueprintPatchApplier(fn BlueprintPatchApplier) {
	blueprintPatchApplierFn = fn
}

func RegisterArchitectReviewResubmitter(fn ArchitectReviewResubmitter) {
	architectReviewResubmitterFn = fn
}

func architectReplyRequestsContinuation(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}
	return strings.Contains(content, architectAutoContinueToken)
}

func isArchitectFollowUpMessage(content string) bool {
	content = strings.TrimSpace(strings.ToLower(content))
	if content == "" {
		return false
	}
	if isReviewRemediationPrompt(content) {
		return false
	}
	for _, kw := range []string{
		"继续", "接着", "下一部分", "下一段", "后一段", "继续发", "继续发送", "go on", "continue", "next",
	} {
		if strings.Contains(content, kw) {
			return true
		}
	}
	return false
}

func isReviewRemediationPrompt(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}
	return strings.Contains(content, "方案审核未通过") ||
		strings.Contains(content, "警告（当前会阻塞执行，必须修复）") ||
		strings.Contains(content, "重新给出完整修订方案") ||
		strings.Contains(content, "task_patches")
}

func isWorkflowApprovalPrompt(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}
	return strings.Contains(content, "方案审核通过") ||
		strings.Contains(content, "项目已进入执行阶段") ||
		strings.Contains(content, "执行阶段启动失败")
}

func shouldParseArchitectReply(ctx context.Context, userContents []string) bool {
	followUpLimit := loadArchitectFollowUpLimit(ctx)
	followUpCount := 0
	for _, content := range userContents {
		if isArchitectFollowUpMessage(content) {
			followUpCount++
			continue
		}
		if followUpCount >= followUpLimit {
			return false
		}
		return !isWorkflowApprovalPrompt(content)
	}
	return true
}

func reviewPromptAllowsAutoContinue(content string) bool {
	return strings.Contains(content, architectAutoContinueToken)
}

type architectReplyPolicy struct {
	allowAutoContinue bool
	allowAutoResubmit bool
}

func shouldApplyArchitectBlueprintMutation(currentStage string, policy architectReplyPolicy) bool {
	currentStage = strings.TrimSpace(strings.ToLower(currentStage))
	if currentStage == "" || currentStage == "design" {
		return true
	}
	return policy.allowAutoResubmit
}

func resolveArchitectReplyPolicy(ctx context.Context, userContents []string) architectReplyPolicy {
	followUpLimit := loadArchitectFollowUpLimit(ctx)
	followUpCount := 0
	for _, content := range userContents {
		if isArchitectFollowUpMessage(content) {
			followUpCount++
			continue
		}
		if followUpCount >= followUpLimit {
			return architectReplyPolicy{}
		}
		if !isReviewRemediationPrompt(content) {
			return architectReplyPolicy{}
		}
		return architectReplyPolicy{
			allowAutoContinue: reviewPromptAllowsAutoContinue(content),
			allowAutoResubmit: true,
		}
	}
	return architectReplyPolicy{}
}

func loadRecentArchitectUserMessages(ctx context.Context, conversationID int64, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 12
	}
	userMsgs, err := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("role", "user").
		Where("status", "completed").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}

	contents := make([]string, 0, len(userMsgs))
	for _, msg := range userMsgs {
		contents = append(contents, msg["content"].String())
	}
	return contents, nil
}

func collectArchitectReplyWindow(ctx context.Context, conversationID int64) (string, error) {
	userMsgs, err := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("role", "user").
		Where("status", "completed").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		Limit(20).
		All()
	if err != nil {
		return "", err
	}

	var boundaryID int64
	for _, msg := range userMsgs {
		if !isArchitectFollowUpMessage(msg["content"].String()) {
			boundaryID = msg["id"].Int64()
			break
		}
	}

	query := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("role", "assistant").
		Where("status", "completed").
		WhereNull("deleted_at")
	if boundaryID > 0 {
		query = query.Where("id >", boundaryID)
	}

	messages, err := query.OrderAsc("created_at").Limit(50).All()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for _, msg := range messages {
		content := strings.TrimSpace(msg["content"].String())
		if content == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n---\n\n")
		}
		builder.WriteString(content)
	}
	return builder.String(), nil
}

func (e *ChatEngine) shouldAutoContinueArchitectReply(ctx context.Context, conversationID int64) bool {
	userContents, err := loadRecentArchitectUserMessages(ctx, conversationID, 12)
	if err != nil || len(userContents) == 0 {
		return false
	}
	return resolveArchitectReplyPolicy(ctx, userContents).allowAutoContinue
}

func (e *ChatEngine) shouldAutoResubmitArchitectReview(ctx context.Context, conversationID int64) bool {
	userContents, err := loadRecentArchitectUserMessages(ctx, conversationID, 12)
	if err != nil || len(userContents) == 0 {
		return false
	}
	return resolveArchitectReplyPolicy(ctx, userContents).allowAutoResubmit
}

func (e *ChatEngine) autoContinueArchitectReply(ctx context.Context, conversationID int64) {
	conv, err := g.DB().Model("mvp_conversation").Ctx(ctx).
		Fields("created_by, dept_id").
		Where("id", conversationID).
		WhereNull("deleted_at").
		One()
	if err != nil || conv.IsEmpty() {
		g.Log().Warningf(ctx, "[ChatEngine] 自动继续前查询对话失败: conversationID=%d err=%v", conversationID, err)
		return
	}
	content := "继续，请继续发送下一段方案或剩余修订内容；如果已经是最后一段，请直接输出最后一段完整 JSON。"
	if _, _, err := e.SendMessage(ctx, conversationID, content, conv["created_by"].Int64(), conv["dept_id"].Int64()); err != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 自动继续架构师回复失败: conversationID=%d err=%v", conversationID, err)
	}
}

func extractArchitectTaskPatches(text string) ([]ArchitectTaskPatch, error) {
	var all []ArchitectTaskPatch
	for _, jsonStr := range extractJSONFromCodeBlocks(text) {
		patches, err := parseArchitectTaskPatchPayload(jsonStr)
		if err == nil && len(patches) > 0 {
			all = append(all, patches...)
		}
	}
	if len(all) > 0 {
		return dedupeArchitectTaskPatches(all), nil
	}

	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		patches, err := parseArchitectTaskPatchPayload(trimmed)
		if err == nil && len(patches) > 0 {
			return dedupeArchitectTaskPatches(patches), nil
		}
	}

	match := jsonCodeBlockRe.FindStringSubmatch(text)
	if len(match) == 2 {
		patches, err := parseArchitectTaskPatchPayload(match[1])
		if err == nil && len(patches) > 0 {
			return dedupeArchitectTaskPatches(patches), nil
		}
	}
	return nil, fmt.Errorf("未解析到有效 task_patches")
}

func parseArchitectTaskPatchPayload(content string) ([]ArchitectTaskPatch, error) {
	cleaned := GetParser().cleanJSON(strings.TrimSpace(content))
	if cleaned == "" {
		return nil, fmt.Errorf("patch 内容为空")
	}

	var envelope architectTaskPatchEnvelope
	if err := json.Unmarshal([]byte(cleaned), &envelope); err == nil && len(envelope.TaskPatches) > 0 {
		return envelope.TaskPatches, nil
	}

	var patch ArchitectTaskPatch
	if err := json.Unmarshal([]byte(cleaned), &patch); err == nil && architectTaskPatchHasTarget(&patch) {
		return []ArchitectTaskPatch{patch}, nil
	}

	var patches []ArchitectTaskPatch
	if err := json.Unmarshal([]byte(cleaned), &patches); err == nil && len(patches) > 0 {
		return patches, nil
	}
	return nil, fmt.Errorf("未识别到 patch payload")
}

func architectTaskPatchHasTarget(patch *ArchitectTaskPatch) bool {
	if patch == nil {
		return false
	}
	return patch.BlueprintID > 0 || strings.TrimSpace(patch.TaskName) != ""
}

func dedupeArchitectTaskPatches(values []ArchitectTaskPatch) []ArchitectTaskPatch {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]int, len(values))
	result := make([]ArchitectTaskPatch, 0, len(values))
	for _, patch := range values {
		if !architectTaskPatchHasTarget(&patch) {
			continue
		}
		key := strings.TrimSpace(patch.TaskName)
		if patch.BlueprintID > 0 {
			key = fmt.Sprintf("id:%d", patch.BlueprintID)
		} else {
			key = "name:" + key
		}
		if idx, ok := seen[key]; ok {
			result[idx] = patch
			continue
		}
		seen[key] = len(result)
		result = append(result, patch)
	}
	return result
}
