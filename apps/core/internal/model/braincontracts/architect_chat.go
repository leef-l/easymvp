package braincontracts

type ArchitectChatInput struct {
	ProjectID   string            `json:"project_id"`
	GoalSummary string            `json:"goal_summary"`
	Messages    []ChatMessageItem `json:"messages"`
	Instruction string            `json:"instruction"`
}

type ArchitectChatResult struct {
	Reply               string              `json:"reply"`
	DraftTasks          []ArchitectTaskItem `json:"draft_tasks"`
	SuggestedNextAction string              `json:"suggested_next_action"`
}

type ArchitectTaskItem struct {
	TaskKey   string `json:"task_key"`
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	TaskKind  string `json:"task_kind"`
	Summary   string `json:"summary"`
	BrainKind string `json:"brain_kind"`
	RoleType  string `json:"role_type"`
}

type ChatMessageItem struct {
	Role        string `json:"role"`
	Name        string `json:"name"`
	Content     string `json:"content"`
	MessageKind string `json:"message_kind,omitempty"`
}
