package v1

import "github.com/gogf/gf/v2/frame/g"

type ReviewItem struct {
	ID              string `json:"id"`
	DesignID        string `json:"design_id"`
	ProjectID       string `json:"project_id"`
	Round           int    `json:"round"`
	Passed          bool   `json:"passed"`
	Score           int    `json:"score"`
	DimensionsJSON  string `json:"dimensions_json,omitempty"`
	IssuesJSON      string `json:"issues_json,omitempty"`
	SuggestionsJSON string `json:"suggestions_json,omitempty"`
	FixTasksJSON    string `json:"fix_tasks_json,omitempty"`
	BrainRunID      string `json:"brain_run_id,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type ListReviewsReq struct {
	g.Meta   `path:"/api/v3/reviews/design/{design_id}" method:"get" tags:"Reviews" summary:"List all review records for a solution design"`
	DesignID string `json:"design_id" in:"path" v:"required"`
}

type ListReviewsRes struct {
	Items []ReviewItem `json:"items"`
	Total int          `json:"total"`
}
