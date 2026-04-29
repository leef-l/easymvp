package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// SendMessageStream calls Brain in SSE streaming mode and forwards events through eventChan.
// The caller is responsible for closing eventChan when done.
func (s *sArchitectChat) SendMessageStream(ctx context.Context, req SendMessageCommand, eventChan chan<- ArchitectChatStreamEvent) error {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Content = strings.TrimSpace(req.Content)
	if req.ProjectID == "" {
		return gerror.New("project id is required")
	}
	if req.Content == "" {
		return gerror.New("content is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	conv, err := s.getConversationByProjectID(ctx, db, req.ProjectID)
	if err != nil {
		return err
	}
	if conv == nil {
		convID, err := s.CreateConversation(ctx, req.ProjectID)
		if err != nil {
			return err
		}
		conv = &entity.Conversations{Id: convID, ProjectId: req.ProjectID, Status: "active"}
	}

	now := nowText()
	msgID := newResourceID("msg")
	_, err = db.ExecContext(ctx,
		`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msgID, conv.Id, "user", "User", req.Content, "chat", now,
	)
	if err != nil {
		return gerror.Wrap(err, "insert user message failed")
	}

	messages, err := s.listMessagesByConversationID(ctx, db, conv.Id)
	if err != nil {
		return err
	}

	project, err := getProjectByID(ctx, db, req.ProjectID)
	if err != nil {
		return err
	}

	chatInput := braincontracts.ArchitectChatInput{
		ProjectID:   project.Id,
		GoalSummary: project.GoalSummary,
		Messages:    toBrainChatMessages(messages),
		Instruction: "Continue the architect conversation based on project context and message history.",
	}

	inputJSON, err := json.Marshal(chatInput)
	if err != nil {
		return gerror.Wrap(err, "marshal chat input failed")
	}

	brainEventCh, err := EasyMVPBrain().ExecuteContractStream(ctx, EasyMVPBrainExecuteCommand{
		ContractKind: "architect_chat",
		Instruction:  "Execute the requested domain contract and return only the final contract envelope JSON.",
		ContextJSON:  inputJSON,
	})
	if err != nil {
		g.Log().Warningf(ctx, "architect chat brain stream call failed: %v", err)
		errStr := err.Error()
		fallback := s.fallbackArchitectResultFromMessages(messages, project)
		if strings.Contains(errStr, "EOF") || strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") {
			fallback.Reply = "[Brain 响应超时] AI 架构师正在处理中，但响应时间较长。请稍等几秒后重试，或确认 brain-v3 serve 状态。"
		} else {
			fallback.Reply = "[Brain 调用异常] 您的消息已保存，但 AI 架构师暂时无法生成回复。错误：" + errStr
		}
		replyMsgID := newResourceID("msg")
		_, _ = db.ExecContext(ctx,
			`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			replyMsgID, conv.Id, "architect", "Architect", fallback.Reply, "chat", nowText(),
		)
		eventChan <- ArchitectChatStreamEvent{
			Type:    "error",
			Content: fallback.Reply,
			Done:    true,
		}
		return nil
	}

	var replyBuilder strings.Builder
	var finalResult *braincontracts.ArchitectChatResult
	for ev := range brainEventCh {
		switch ev.Type {
		case "llm.content_delta", "llm.thinking_delta":
			// Brain-v3 content delta: {"text":"..."} (from LLMProxy)
			// Brain-v3 progress delta: {"message":"..."} (from brain/progress)
			var payload struct {
				Text    string `json:"text"`
				Message string `json:"message"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			text := payload.Text
			if text == "" {
				text = payload.Message
			}
			if text != "" {
				replyBuilder.WriteString(text)
				eventChan <- ArchitectChatStreamEvent{
					Type:    ev.Type,
					Content: text,
				}
			}
		case "llm.message_start":
			var payload struct {
				ID    string `json:"id"`
				Model string `json:"model"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			eventChan <- ArchitectChatStreamEvent{
				Type:    "llm.message_start",
				Content: payload.Model,
			}
		case "llm.message_delta":
			var payload struct {
				StopReason string `json:"stop_reason"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			eventChan <- ArchitectChatStreamEvent{
				Type:    "llm.message_delta",
				Content: payload.StopReason,
			}
		case "llm.message_end":
			eventChan <- ArchitectChatStreamEvent{
				Type: "llm.message_end",
			}
		case "agent.tool_start", "agent.tool_end", "agent.turn":
			var payload struct {
				ToolName string `json:"tool_name"`
				Message  string `json:"message"`
				Detail   string `json:"detail"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			eventChan <- ArchitectChatStreamEvent{
				Type:     ev.Type,
				ToolName: payload.ToolName,
				Content:  payload.Message,
				Detail:   payload.Detail,
			}
		case "execution.done":
			var execResp struct {
				Result  json.RawMessage `json:"result"`
				Summary string          `json:"summary"`
			}
			if err := json.Unmarshal(ev.Data, &execResp); err == nil && len(execResp.Result) > 0 {
				var envelope braincontracts.BrainContractEnvelope
				if err := json.Unmarshal(execResp.Result, &envelope); err == nil {
					var typedResult braincontracts.ArchitectChatResult
					if err := json.Unmarshal(envelope.ResultJSON, &typedResult); err == nil {
						finalResult = &typedResult
					}
				}
			}
			if finalResult == nil && replyBuilder.Len() > 0 {
				finalResult = &braincontracts.ArchitectChatResult{Reply: replyBuilder.String()}
			}
			if finalResult == nil || strings.TrimSpace(finalResult.Reply) == "" {
				finalResult = s.fallbackArchitectResultFromMessages(messages, project)
				if replyBuilder.Len() > 0 {
					finalResult.Reply = replyBuilder.String()
				}
			}
			replyMsgID := newResourceID("msg")
			cmdID := newResourceID("cmd")
			_, _ = db.ExecContext(ctx,
				`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				replyMsgID, conv.Id, "architect", "Architect", finalResult.Reply, "chat", nowText(),
			)
			eventChan <- ArchitectChatStreamEvent{
				Type:      "execution.done",
				Content:   finalResult.Reply,
				Result:    finalResult,
				Done:      true,
				CommandID: cmdID,
				MessageID: replyMsgID,
			}
		case "execution.error", "execution.cancelled":
			var payload struct {
				Error string `json:"error"`
			}
			_ = json.Unmarshal(ev.Data, &payload)
			fallback := s.fallbackArchitectResultFromMessages(messages, project)
			fallback.Reply = "[Brain 调用异常] " + payload.Error
			replyMsgID := newResourceID("msg")
			cmdID := newResourceID("cmd")
			_, _ = db.ExecContext(ctx,
				`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				replyMsgID, conv.Id, "architect", "Architect", fallback.Reply, "chat", nowText(),
			)
			eventChan <- ArchitectChatStreamEvent{
				Type:      "error",
				Content:   fallback.Reply,
				Done:      true,
				CommandID: cmdID,
				MessageID: replyMsgID,
			}
		default:
			// forward unknown events transparently
			eventChan <- ArchitectChatStreamEvent{
				Type:    ev.Type,
				Content: string(ev.Data),
			}
		}
	}
	return nil
}
