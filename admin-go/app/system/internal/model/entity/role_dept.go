// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RoleDept is the golang structure for table role_dept.
type RoleDept struct {
	RoleId uint64 `orm:"role_id" description:"角色ID"` // 角色ID
	DeptId uint64 `orm:"dept_id" description:"部门ID"` // 部门ID
}
