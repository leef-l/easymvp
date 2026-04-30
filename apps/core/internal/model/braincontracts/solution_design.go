package braincontracts

// SolutionDesignInput is the input contract for the solution_design brain call.
type SolutionDesignInput struct {
	ProjectID          string `json:"project_id"`
	GoalSummary        string `json:"goal_summary"`
	RequirementID      string `json:"requirement_id"`
	RequirementDocJSON string `json:"requirement_doc_json"`
	Instruction        string `json:"instruction"`
}

// SolutionDesignResult is the typed result returned by the solution_design brain contract.
type SolutionDesignResult struct {
	Architecture   string `json:"architecture"`
	ModulesJSON    string `json:"modules_json"`
	DataModelsJSON string `json:"data_models_json"`
	PagesJSON      string `json:"pages_json"`
	TaskDraftsJSON string `json:"task_drafts_json"`
	Summary        string `json:"summary"`
}
