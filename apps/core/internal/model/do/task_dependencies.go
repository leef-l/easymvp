// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TaskDependencies is the golang structure of table task_dependencies for DAO operations like Where/Data.
type TaskDependencies struct {
	g.Meta          `orm:"table:task_dependencies, do:true"`
	TaskId          any //
	DependsOnTaskId any //
	CreatedAt       any //
}
