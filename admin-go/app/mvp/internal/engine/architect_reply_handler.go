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

const architectAutoContinueToken = "[AUTO_CONTINUE_NEXT]"

func RegisterBlueprintPatchApplier(fn BlueprintPatchApplier) {
	blueprintPatchApplierFn = fn
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

func reviewPromptAllowsAutoContinue(content string) bool {
	return strings.Contains(content, architectAutoContinueToken)
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
	userMsgs, err := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("role", "user").
		Where("status", "completed").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		Limit(12).
		All()
	if err != nil || len(userMsgs) == 0 {
		return false
	}

	followUpCount := 0
	for _, msg := range userMsgs {
		content := msg["content"].String()
		if isArchitectFollowUpMessage(content) {
			followUpCount++
			continue
		}
		if followUpCount >= 6 {
			return false
		}
		return isReviewRemediationPrompt(content) && reviewPromptAllowsAutoContinue(content)
	}
	return false
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
