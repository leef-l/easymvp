// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProjectProfiles is the golang structure for table project_profiles.
type ProjectProfiles struct {
	Id                       string `json:"id"                       orm:"id"                         ` //
	ProjectId                string `json:"projectId"                orm:"project_id"                 ` //
	CategoryProfileVersion   string `json:"categoryProfileVersion"   orm:"category_profile_version"   ` //
	AcceptanceProfileVersion string `json:"acceptanceProfileVersion" orm:"acceptance_profile_version" ` //
	RoleProfileVersion       string `json:"roleProfileVersion"       orm:"role_profile_version"       ` //
	CreatedAt                string `json:"createdAt"                orm:"created_at"                 ` //
	UpdatedAt                string `json:"updatedAt"                orm:"updated_at"                 ` //
}
