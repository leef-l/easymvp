// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectProfiles is the golang structure of table project_profiles for DAO operations like Where/Data.
type ProjectProfiles struct {
	g.Meta                   `orm:"table:project_profiles, do:true"`
	Id                       any //
	ProjectId                any //
	CategoryProfileVersion   any //
	AcceptanceProfileVersion any //
	RoleProfileVersion       any //
	CreatedAt                any //
	UpdatedAt                any //
}
