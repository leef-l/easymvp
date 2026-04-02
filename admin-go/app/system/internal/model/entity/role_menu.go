// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RoleMenu is the golang structure for table role_menu.
type RoleMenu struct {
	RoleId uint64 `orm:"role_id" description:"角色ID"` // 角色ID
	MenuId uint64 `orm:"menu_id" description:"菜单ID"` // 菜单ID
}
