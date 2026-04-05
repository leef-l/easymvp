package consts

// UsersStatus 状态
const (
	UsersStatus关闭 = 0
	UsersStatus开启 = 1
)

func IsValidUsersStatus(value int) bool {
	return value == UsersStatus关闭 || value == UsersStatus开启
}
