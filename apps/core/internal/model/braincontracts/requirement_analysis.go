package braincontracts

// RequirementAnalysisInput is the input for the requirement_analysis brain contract.
type RequirementAnalysisInput struct {
	ProjectID   string `json:"project_id"`
	GoalSummary string `json:"goal_summary"`
	RawInput    string `json:"raw_input"`
	Instruction string `json:"instruction"`
}

// RequirementAnalysisResult is the structured output from the requirement_analysis brain contract.
type RequirementAnalysisResult struct {
	// RequirementDoc is the structured requirement document produced by the brain.
	RequirementDoc RequirementDoc `json:"requirement_doc"`
	// Summary is a human-readable summary of the analysis.
	Summary string `json:"summary"`
	// SuggestedNextAction hints what the user should do next (e.g. "confirm_requirement").
	SuggestedNextAction string `json:"suggested_next_action"`
}

// RequirementDoc represents the structured requirement document.
type RequirementDoc struct {
	Title            string               `json:"title"`
	Overview         string               `json:"overview"`
	FunctionalReqs   []RequirementItem    `json:"functional_requirements"`
	NonFunctionalReqs []RequirementItem   `json:"non_functional_requirements"`
	UserStories      []UserStory          `json:"user_stories"`
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria"`
	Constraints      []string             `json:"constraints"`
	Assumptions      []string             `json:"assumptions"`
}

// RequirementItem is a single requirement entry.
type RequirementItem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // must/should/could/wont
}

// UserStory represents a user story in the requirement document.
type UserStory struct {
	ID       string `json:"id"`
	AsA      string `json:"as_a"`
	IWant    string `json:"i_want"`
	SoThat   string `json:"so_that"`
	Priority string `json:"priority"`
}

// AcceptanceCriterion defines a single acceptance criterion.
type AcceptanceCriterion struct {
	ID          string `json:"id"`
	Given       string `json:"given"`
	When        string `json:"when"`
	Then        string `json:"then"`
	Description string `json:"description"`
}
