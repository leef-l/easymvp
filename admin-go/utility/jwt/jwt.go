package jwt

import (
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
	gojwt.RegisteredClaims
}

var (
	secret       []byte
	memberSecret []byte
	expireTime   time.Duration
)

func init() {
	ctx := gctx.New()
	key, _ := g.Cfg().Get(ctx, "jwt.secret", "easymvp-secret-key")
	secret = []byte(key.String())
	// 会员端独立 secret，未配置时回退到管理端 secret
	mKey, _ := g.Cfg().Get(ctx, "jwt.memberSecret", "")
	if mKey.String() != "" {
		memberSecret = []byte(mKey.String())
	} else {
		memberSecret = secret
	}
	hours, _ := g.Cfg().Get(ctx, "jwt.expire", 24)
	expireTime = time.Duration(hours.Int()) * time.Hour
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID int64, username string, deptID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		DeptID:   deptID,
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
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
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
	_, err := ParseToken(tokenStr)
	if err == nil {
		return true
	}
	_, err = ParseMemberToken(tokenStr)
	return err == nil
}

// ParseMemberToken 解析会员 JWT Token
func ParseMemberToken(tokenStr string) (*MemberClaims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &MemberClaims{}, func(t *gojwt.Token) (interface{}, error) {
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
