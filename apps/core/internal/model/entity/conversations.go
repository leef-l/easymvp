package entity

// Conversations is the golang structure for table conversations.
type Conversations struct {
	Id               string `json:"id" orm:"id"`
	ProjectId        string `json:"project_id" orm:"project_id"`
	ConversationKind string `json:"conversation_kind" orm:"conversation_kind"`
	Status           string `json:"status" orm:"status"`
	PlanDraftId      string `json:"plan_draft_id" orm:"plan_draft_id"`
	CreatedAt        string `json:"created_at" orm:"created_at"`
	UpdatedAt        string `json:"updated_at" orm:"updated_at"`
}
