package jwt

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	gojwt "github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 载荷
type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	DeptID   int64  `json:"deptId"`
	IsAdmin  bool   `json:"isAdmin"`
	gojwt.RegisteredClaims
}

var (
	secret       []byte
	memberSecret []byte
	expireTime   time.Duration
	configMu     sync.RWMutex
)

func loadConfig() error {
	configMu.RLock()
	if len(secret) > 0 && len(memberSecret) > 0 && expireTime > 0 {
		configMu.RUnlock()
		return nil
	}
	configMu.RUnlock()

	configMu.Lock()
	defer configMu.Unlock()

	if len(secret) > 0 && len(memberSecret) > 0 && expireTime > 0 {
		return nil
	}

	ctx := gctx.New()
	key := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if key == "" {
		cfgValue, err := g.Cfg().Get(ctx, "jwt.secret")
		if err == nil && !cfgValue.IsEmpty() {
			key = strings.TrimSpace(cfgValue.String())
		}
	}
	if key == "" {
		return fmt.Errorf("jwt.secret 未配置")
	}
	secret = []byte(key)

	memberKey := strings.TrimSpace(os.Getenv("JWT_MEMBER_SECRET"))
	if memberKey == "" {
		mKey, _ := g.Cfg().Get(ctx, "jwt.memberSecret", "")
		memberKey = strings.TrimSpace(mKey.String())
	}
	if memberKey != "" {
		memberSecret = []byte(memberKey)
	} else {
		memberSecret = secret
	}

	hours := 24
	if raw := strings.TrimSpace(os.Getenv("JWT_EXPIRE_HOURS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			hours = parsed
		}
	} else {
		cfgHours, _ := g.Cfg().Get(ctx, "jwt.expire", 24)
		if cfgHours.Int() > 0 {
			hours = cfgHours.Int()
		}
	}
	expireTime = time.Duration(hours) * time.Hour
	return nil
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID int64, username string, deptID int64, isAdmin bool) (string, error) {
	if err := loadConfig(); err != nil {
		return "", err
	}
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		DeptID:   deptID,
		IsAdmin:  isAdmin,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(expireTime)),
			IssuedAt:  gojwt.NewNumericDate(now),
			Issuer:    "easymvp",
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析 JWT Token
func ParseToken(tokenStr string) (*Claims, error) {
	if err := loadConfig(); err != nil {
		return nil, err
	}
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, gojwt.ErrTokenInvalidClaims
}

// MemberClaims C端会员 JWT 载荷
type MemberClaims struct {
	MemberID    int64  `json:"memberId"`
	Phone       string `json:"phone"`
	IsCoach     int    `json:"isCoach"`
	CoachID     int64  `json:"coachId"`
	CurrentRole string `json:"currentRole"` // "member" | "coach"
	gojwt.RegisteredClaims
}

// GenerateMemberToken 生成会员 JWT Token
func GenerateMemberToken(memberID int64, phone string, isCoach int, coachID int64, currentRole string) (string, error) {
	if err := loadConfig(); err != nil {
		return "", err
	}
	now := time.Now()
	claims := MemberClaims{
		MemberID:    memberID,
		Phone:       phone,
		IsCoach:     isCoach,
		CoachID:     coachID,
		CurrentRole: currentRole,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(expireTime)),
			IssuedAt:  gojwt.NewNumericDate(now),
			Issuer:    "easymvp-member",
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(memberSecret)
}

// VerifyAnyToken 只验证 token 签名合法且未过期，不关心是哪种身份
func VerifyAnyToken(tokenStr string) bool {
	if err := loadConfig(); err != nil {
		return false
	}
	_, err := ParseToken(tokenStr)
	if err == nil {
		return true
	}
	_, err = ParseMemberToken(tokenStr)
	return err == nil
}

// ParseMemberToken 解析会员 JWT Token
func ParseMemberToken(tokenStr string) (*MemberClaims, error) {
	if err := loadConfig(); err != nil {
		return nil, err
	}
	token, err := gojwt.ParseWithClaims(tokenStr, &MemberClaims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", t.Header["alg"])
		}
		return memberSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MemberClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, gojwt.ErrTokenInvalidClaims
}
