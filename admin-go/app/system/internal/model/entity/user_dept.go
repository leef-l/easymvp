// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UserDept is the golang structure for table user_dept.
type UserDept struct {
	UserId uint64 `orm:"user_id" description:"用户ID"` // 用户ID
	DeptId uint64 `orm:"dept_id" description:"部门ID"` // 部门ID
}
