// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TaskDependencies is the golang structure for table task_dependencies.
type TaskDependencies struct {
	TaskId          string `json:"taskId"          orm:"task_id"            ` //
	DependsOnTaskId string `json:"dependsOnTaskId" orm:"depends_on_task_id" ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"         ` //
}
