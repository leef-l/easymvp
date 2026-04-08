package jwt

import (
	"os"
	"testing"
)

func resetConfigForTest() {
	configMu.Lock()
	defer configMu.Unlock()
	secret = nil
	memberSecret = nil
	expireTime = 0
}

func TestGenerateAndParseTokenWithEnvFallback(t *testing.T) {
	oldSecret := os.Getenv("JWT_SECRET")
	oldMemberSecret := os.Getenv("JWT_MEMBER_SECRET")
	oldExpire := os.Getenv("JWT_EXPIRE_HOURS")
	defer func() {
		_ = os.Setenv("JWT_SECRET", oldSecret)
		_ = os.Setenv("JWT_MEMBER_SECRET", oldMemberSecret)
		_ = os.Setenv("JWT_EXPIRE_HOURS", oldExpire)
		resetConfigForTest()
	}()

	resetConfigForTest()
	_ = os.Setenv("JWT_SECRET", "unit-test-secret")
	_ = os.Setenv("JWT_MEMBER_SECRET", "unit-test-member-secret")
	_ = os.Setenv("JWT_EXPIRE_HOURS", "1")

	token, err := GenerateToken(123, "tester", 7, true)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 123 || claims.Username != "tester" || claims.DeptID != 7 || !claims.IsAdmin {
		t.Fatalf("claims = %+v", claims)
	}
}
