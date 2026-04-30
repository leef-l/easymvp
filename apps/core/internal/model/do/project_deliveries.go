// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectDeliveries is the golang structure of table project_deliveries for DAO operations like Where/Data.
type ProjectDeliveries struct {
	g.Meta          `orm:"table:project_deliveries, do:true"`
	Id              any //
	ProjectId       any //
	Status          any //
	WorkspacePath   any //
	Readme          any //
	ArchitectureDoc any //
	ApiDocs         any //
	DeploymentDoc   any //
	TestReportJson  any //
	StatisticsJson  any //
	UserAccepted    any //
	AcceptedAt      any //
	DeliveredAt     any //
	CreatedAt       any //
	UpdatedAt       any //
}
