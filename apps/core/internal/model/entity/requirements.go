// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Requirements is the golang structure for table requirements.
type Requirements struct {
	Id               string `json:"id"               orm:"id"                  ` //
	ProjectId        string `json:"projectId"        orm:"project_id"          ` //
	RawInput         string `json:"rawInput"         orm:"raw_input"           ` //
	Status           string `json:"status"           orm:"status"              ` //
	RequirementDocJson string `json:"requirementDocJson" orm:"requirement_doc_json" ` //
	UserConfirmed    int    `json:"userConfirmed"    orm:"user_confirmed"      ` //
	ConfirmedAt      string `json:"confirmedAt"      orm:"confirmed_at"        ` //
	CreatedAt        string `json:"createdAt"        orm:"created_at"          ` //
	UpdatedAt        string `json:"updatedAt"        orm:"updated_at"          ` //
}
