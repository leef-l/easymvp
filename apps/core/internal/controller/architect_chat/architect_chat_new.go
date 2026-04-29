package architect_chat

import api "github.com/leef-l/easymvp/apps/core/api/architect_chat"

func NewV1() api.IArchitectChatV1 {
	return &ControllerV1{}
}
