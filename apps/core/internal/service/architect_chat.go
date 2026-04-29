package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type SendMessageCommand struct {
	ProjectID string
	Content   string
}

type SendMessageResult struct {
	CommandID      string
	MessageID      string
	ArchitectReply string
}

type MessageView struct {
	ID          string
	SenderRole  string
	SenderName  string
	Content     string
	MessageKind string
	CreatedAt   string
}

type ConversationView struct {
	ID          string
	ProjectID   string
	Status      string
	PlanDraftID string
	Messages    []MessageView
}

type ConfirmPlanResult struct {
	CommandID      string
	Accepted       bool
	CompiledPlanID string
	Reason         string
	ReviewID       string
}

type ArchitectChatStreamEvent struct {
	Type        string `json:"type"`
	Content     string `json:"content,omitempty"`
	ToolName    string `json:"tool_name,omitempty"`
	Detail      string `json:"detail,omitempty"`
	Done        bool   `json:"done,omitempty"`
	Error       string `json:"error,omitempty"`
	Result      *braincontracts.ArchitectChatResult `json:"result,omitempty"`
	CommandID   string `json:"command_id,omitempty"`
	MessageID   string `json:"message_id,omitempty"`
}

type IArchitectChat interface {
	CreateConversation(ctx context.Context, projectID string) (string, error)
	SendMessage(ctx context.Context, req SendMessageCommand) (*SendMessageResult, error)
	SendMessageStream(ctx context.Context, req SendMessageCommand, eventChan chan<- ArchitectChatStreamEvent) error
	ListMessages(ctx context.Context, conversationID string) ([]MessageView, error)
	GetConversationByProject(ctx context.Context, projectID string) (*ConversationView, error)
	ConfirmPlan(ctx context.Context, projectID string) (*ConfirmPlanResult, error)
}

var localArchitectChat IArchitectChat = (*sArchitectChat)(nil)

type sArchitectChat struct{}

func ArchitectChat() IArchitectChat {
	if localArchitectChat == nil {
		localArchitectChat = &sArchitectChat{}
	}
	return localArchitectChat
}

func (s *sArchitectChat) CreateConversation(ctx context.Context, projectID string) (string, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", err
	}
	defer closeFn()

	convID := newResourceID("conv")
	now := nowText()
	_, err = db.ExecContext(ctx,
		`INSERT INTO `+dao.Conversations.Table()+` (id, project_id, conversation_kind, status, plan_draft_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		convID, projectID, "architect", "active", "", now, now,
	)
	if err != nil {
		return "", gerror.Wrap(err, "insert conversation failed")
	}
	return convID, nil
}

func (s *sArchitectChat) SendMessage(ctx context.Context, req SendMessageCommand) (*SendMessageResult, error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Content = strings.TrimSpace(req.Content)
	if req.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}
	if req.Content == "" {
		return nil, gerror.New("content is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	conv, err := s.getConversationByProjectID(ctx, db, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		convID, err := s.CreateConversation(ctx, req.ProjectID)
		if err != nil {
			return nil, err
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
		return nil, gerror.Wrap(err, "insert user message failed")
	}

	messages, err := s.listMessagesByConversationID(ctx, db, conv.Id)
	if err != nil {
		return nil, err
	}

	project, err := getProjectByID(ctx, db, req.ProjectID)
	if err != nil {
		return nil, err
	}

	chatInput := braincontracts.ArchitectChatInput{
		ProjectID:   project.Id,
		GoalSummary: project.GoalSummary,
		Messages:    toBrainChatMessages(messages),
		Instruction: "Continue the architect conversation based on project context and message history.",
	}

	_, result, err := EasyMVPBrain().CallArchitectChat(ctx, chatInput)
	if err != nil {
		g.Log().Warningf(ctx, "architect chat brain call failed: %v", err)
		result = s.fallbackArchitectResultFromMessages(messages, project)
		errStr := err.Error()
		if strings.Contains(errStr, "EOF") || strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") {
			result.Reply = "[Brain 响应超时] AI 架构师正在处理中，但响应时间较长。请稍等几秒后重试，或确认 brain-v3 serve 状态。"
		} else {
			result.Reply = "[Brain 调用异常] 您的消息已保存，但 AI 架构师暂时无法生成回复。错误：" + errStr
		}
	}

	replyMsgID := newResourceID("msg")
	_, err = db.ExecContext(ctx,
		`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		replyMsgID, conv.Id, "architect", "Architect", result.Reply, "chat", nowText(),
	)
	if err != nil {
		return nil, gerror.Wrap(err, "insert architect reply failed")
	}

	return &SendMessageResult{
		CommandID:      newResourceID("cmd"),
		MessageID:      replyMsgID,
		ArchitectReply: result.Reply,
	}, nil
}

func (s *sArchitectChat) ListMessages(ctx context.Context, conversationID string) ([]MessageView, error) {
	if strings.TrimSpace(conversationID) == "" {
		return nil, gerror.New("conversation id is required")
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()
	return s.listMessagesByConversationID(ctx, db, conversationID)
}

func (s *sArchitectChat) GetConversationByProject(ctx context.Context, projectID string) (*ConversationView, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, gerror.New("project id is required")
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	conv, err := s.getConversationByProjectID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return &ConversationView{
			ProjectID: projectID,
			Status:    "missing",
			Messages:  []MessageView{},
		}, nil
	}

	messages, err := s.listMessagesByConversationID(ctx, db, conv.Id)
	if err != nil {
		return nil, err
	}

	return &ConversationView{
		ID:          conv.Id,
		ProjectID:   conv.ProjectId,
		Status:      conv.Status,
		PlanDraftID: conv.PlanDraftId,
		Messages:    messages,
	}, nil
}

func (s *sArchitectChat) ConfirmPlan(ctx context.Context, projectID string) (*ConfirmPlanResult, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	profile, err := getProjectProfileByProjectID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	if err = validateProjectStatusTransition(project.Status, []string{"created", "planning", "plan_draft", "plan_review", "review", "compiled"}); err != nil {
		return nil, err
	}

	conv, err := s.getConversationByProjectID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, gerror.New("architect conversation not found")
	}

	messages, err := s.listMessagesByConversationID(ctx, db, conv.Id)
	if err != nil {
		return nil, err
	}

	chatInput := braincontracts.ArchitectChatInput{
		ProjectID:   project.Id,
		GoalSummary: project.GoalSummary,
		Messages:    toBrainChatMessages(messages),
		Instruction: "根据对话历史生成结构化方案草案",
	}

	_, result, err := EasyMVPBrain().CallArchitectChat(ctx, chatInput)
	if err != nil {
		g.Log().Warningf(ctx, "architect chat confirm plan brain call failed: %v", err)
		result = s.fallbackArchitectResultFromMessages(messages, project)
	}

	draftTasksJSON := mustMarshalJSONString(result.DraftTasks, "[]")
	if draftTasksJSON == "[]" {
		result.DraftTasks = s.fallbackArchitectResultFromMessages(messages, project).DraftTasks
		draftTasksJSON = mustMarshalJSONString(result.DraftTasks, "[]")
	}

	var draft *entity.WorkflowPlanDrafts
	draft, err = getPlanDraftForProject(ctx, *project)
	if err != nil && err != sql.ErrNoRows && !isSchemaMissingError(err) {
		return nil, err
	}
	now := nowText()
	if draft == nil {
		draftID := newResourceID("plan_draft")
		inputRequirements := mustMarshalJSONString(map[string]any{
			"goal_summary":               project.GoalSummary,
			"workspace_root":             project.WorkspaceRoot,
			"repo_root":                  project.RepoRoot,
			"category_profile_version":   profile.CategoryProfileVersion,
			"acceptance_profile_version": profile.AcceptanceProfileVersion,
			"role_profile_version":       profile.RoleProfileVersion,
		}, "{}")
		row := entity.WorkflowPlanDrafts{
			Id:                    draftID,
			ProjectId:             projectID,
			Version:               1,
			SourceKind:            "architect_chat",
			ProjectCategory:       project.ProjectCategory,
			GoalSummary:           project.GoalSummary,
			InputRequirementsJson: inputRequirements,
			DraftTasksJson:        draftTasksJSON,
			Status:                "ready",
			CreatedBy:             "architect_chat",
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, gerror.Wrap(err, "begin draft transaction failed")
		}
		if err = insertPlanDraftRow(ctx, tx, row); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if err = updateProjectCurrentPlanDraft(ctx, tx, projectID, draftID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, gerror.Wrap(err, "commit draft transaction failed")
		}
		draft = &row
	} else {
		_, err = db.ExecContext(ctx,
			`UPDATE `+dao.WorkflowPlanDrafts.Table()+` SET draft_tasks_json = ?, updated_at = ? WHERE id = ?`,
			draftTasksJSON, now, draft.Id,
		)
		if err != nil {
			return nil, gerror.Wrap(err, "update draft tasks failed")
		}
		draft.DraftTasksJson = draftTasksJSON
	}

	review, err := runPlanReview(ctx, project, profile, draft)
	if err != nil {
		return nil, err
	}

	if !planReviewCompileAllowed(review) {
		issues := mustUnmarshalJSONObject(review.IssuesJson)
		blocking := extractIssueSummaries(issues, "blocking_issues")
		advisory := extractIssueSummaries(issues, "advisory_issues")

		reviewContent := buildReviewMessageContent(blocking, advisory)
		reviewMsgID := newResourceID("msg")
		_, err = db.ExecContext(ctx,
			`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			reviewMsgID, conv.Id, "reviewer", "Reviewer", reviewContent, "review_feedback", nowText(),
		)
		if err != nil {
			g.Log().Errorf(ctx, "insert reviewer message failed: %v", err)
		}

		// Auto-generate architect feedback asynchronously so it never blocks the confirm-plan response.
		go func(convID string, reviewContent string) {
			feedbackInput := braincontracts.ArchitectChatInput{
				ProjectID:   project.Id,
				GoalSummary: project.GoalSummary,
				Messages:    toBrainChatMessages(append(messages, MessageView{SenderRole: "reviewer", SenderName: "Reviewer", Content: reviewContent, MessageKind: "review_feedback"})),
				Instruction: "根据审核反馈调整方案草案并回复",
			}
			_, feedbackResult, brainErr := EasyMVPBrain().CallArchitectChat(context.Background(), feedbackInput)
			if brainErr == nil && feedbackResult != nil {
				replyMsgID := newResourceID("msg")
				_, _ = db.ExecContext(context.Background(),
					`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
					replyMsgID, convID, "architect", "Architect", feedbackResult.Reply, "chat", nowText(),
				)
			}
		}(conv.Id, reviewContent)

		return &ConfirmPlanResult{
			CommandID: newResourceID("cmd"),
			Accepted:  false,
			Reason:    "review_rejected",
			ReviewID:  review.Id,
		}, nil
	}

	compiledPlanID, err := runPlanCompile(ctx, project, profile, draft, review)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`UPDATE `+dao.Conversations.Table()+` SET status = ?, plan_draft_id = ?, updated_at = ? WHERE id = ?`,
		"confirmed", draft.Id, nowText(), conv.Id,
	)
	if err != nil {
		g.Log().Errorf(ctx, "update conversation status failed: %v", err)
	}

	systemMsgID := newResourceID("msg")
	_, _ = db.ExecContext(ctx,
		`INSERT INTO `+dao.Messages.Table()+` (id, conversation_id, sender_role, sender_name, content, message_kind, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		systemMsgID, conv.Id, "system", "System", "方案已通过审核并编译完成", "system", nowText(),
	)

	return &ConfirmPlanResult{
		CommandID:      newResourceID("cmd"),
		Accepted:       true,
		CompiledPlanID: compiledPlanID,
	}, nil
}

func (s *sArchitectChat) getConversationByProjectID(ctx context.Context, db *sql.DB, projectID string) (*entity.Conversations, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, conversation_kind, status, plan_draft_id, created_at, updated_at FROM `+dao.Conversations.Table()+` WHERE project_id = ? LIMIT 1`,
		projectID,
	)
	var conv entity.Conversations
	if err := row.Scan(&conv.Id, &conv.ProjectId, &conv.ConversationKind, &conv.Status, &conv.PlanDraftId, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query conversation failed")
	}
	return &conv, nil
}

func (s *sArchitectChat) listMessagesByConversationID(ctx context.Context, db *sql.DB, conversationID string) ([]MessageView, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, sender_role, sender_name, content, message_kind, created_at FROM `+dao.Messages.Table()+` WHERE conversation_id = ? ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "query messages failed")
	}
	defer rows.Close()

	var result []MessageView
	for rows.Next() {
		var m MessageView
		if err := rows.Scan(&m.ID, &m.SenderRole, &m.SenderName, &m.Content, &m.MessageKind, &m.CreatedAt); err != nil {
			return nil, gerror.Wrap(err, "scan message failed")
		}
		result = append(result, m)
	}
	return result, nil
}

func toBrainChatMessages(views []MessageView) []braincontracts.ChatMessageItem {
	items := make([]braincontracts.ChatMessageItem, 0, len(views))
	for _, v := range views {
		items = append(items, braincontracts.ChatMessageItem{
			Role:        v.SenderRole,
			Name:        v.SenderName,
			Content:     v.Content,
			MessageKind: v.MessageKind,
		})
	}
	return items
}

func (s *sArchitectChat) fallbackArchitectResultFromMessages(messages []MessageView, project *entity.Projects) *braincontracts.ArchitectChatResult {
	summary := "Assembled from architect chat"
	if len(messages) > 0 {
		var parts []string
		for _, m := range messages {
			if m.SenderRole == "user" {
				parts = append(parts, m.Content)
			}
		}
		if len(parts) > 0 {
			summary = strings.Join(parts, "; ")
			if len(summary) > 200 {
				summary = summary[:200] + "..."
			}
		}
	}
	return &braincontracts.ArchitectChatResult{
		Reply: "I have assembled a draft based on our conversation. Please review and confirm.",
		DraftTasks: []braincontracts.ArchitectTaskItem{
			{
				TaskKey:   "architect_chat_assembled",
				Name:      project.GoalSummary,
				Phase:     "design",
				TaskKind:  "planning",
				Summary:   summary,
				BrainKind: "easymvp-brain",
				RoleType:  "architect",
			},
		},
		SuggestedNextAction: "confirm_plan",
	}
}

func extractIssueSummaries(issues map[string]any, key string) []string {
	raw, ok := issues[key]
	if !ok {
		return nil
	}
	data, _ := json.Marshal(raw)
	var items []struct {
		Code    string `json:"code"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return nil
	}
	var result []string
	for _, item := range items {
		if strings.TrimSpace(item.Summary) != "" {
			result = append(result, item.Summary)
		}
	}
	return result
}

func buildReviewMessageContent(blocking, advisory []string) string {
	var sb strings.Builder
	sb.WriteString("方案审核结果：未通过。\n")
	if len(blocking) > 0 {
		sb.WriteString("\n阻塞问题：\n")
		for i, issue := range blocking {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
	}
	if len(advisory) > 0 {
		sb.WriteString("\n建议问题：\n")
		for i, issue := range advisory {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
	}
	return sb.String()
}
