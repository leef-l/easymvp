package support

import (
	"strings"

	"github.com/gogf/gf/v2/crypto/gsha256"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword 验证密码（兼容 bcrypt 和遗留 SHA256 两种格式）
func VerifyPassword(storedPassword, plainPassword string) bool {
	if storedPassword == "" {
		return false
	}

	// bcrypt 格式（以 $2 开头）
	if strings.HasPrefix(storedPassword, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(plainPassword)) == nil
	}

	// 遗留的 SHA256 格式
	return storedPassword == gsha256.Encrypt(plainPassword)
}
