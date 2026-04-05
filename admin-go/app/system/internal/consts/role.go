package consts

// RoleDataScope 数据范围
const (
	RoleDataScope全部        = 1
	RoleDataScope本部门及以下 = 2
	RoleDataScope本部门      = 3
	RoleDataScope仅本人      = 4
	RoleDataScope自定义      = 5
)

// RoleStatus 状态
const (
	RoleStatus关闭 = 0
	RoleStatus开启 = 1
)

// RoleIsAdmin 超级管理员标记
const (
	RoleIsAdmin否 = 0
	RoleIsAdmin是 = 1
)

func IsValidRoleDataScope(value int) bool {
	switch value {
	case RoleDataScope全部, RoleDataScope本部门及以下, RoleDataScope本部门, RoleDataScope仅本人, RoleDataScope自定义:
		return true
	default:
		return false
	}
}

func IsValidRoleStatus(value int) bool {
	return value == RoleStatus关闭 || value == RoleStatus开启
}

func IsValidRoleIsAdmin(value int) bool {
	return value == RoleIsAdmin否 || value == RoleIsAdmin是
}
