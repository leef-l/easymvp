package v1

// RequirementDetail holds the full requirement record returned by GET endpoints.
type RequirementDetail struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"project_id"`
	RawInput           string `json:"raw_input"`
	Status             string `json:"status"`
	RequirementDocJSON string `json:"requirement_doc_json"`
	UserConfirmed      int    `json:"user_confirmed"`
	ConfirmedAt        string `json:"confirmed_at"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
