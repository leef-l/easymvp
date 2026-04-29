package entity

// Messages is the golang structure for table messages.
type Messages struct {
	Id             string `json:"id" orm:"id"`
	ConversationId string `json:"conversation_id" orm:"conversation_id"`
	SenderRole     string `json:"sender_role" orm:"sender_role"`
	SenderName     string `json:"sender_name" orm:"sender_name"`
	Content        string `json:"content" orm:"content"`
	MessageKind    string `json:"message_kind" orm:"message_kind"`
	CreatedAt      string `json:"created_at" orm:"created_at"`
}
