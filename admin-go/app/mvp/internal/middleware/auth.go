package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"easymvp/utility/jwt"
	"easymvp/utility/response"
)

// Auth JWT 鉴权中间件
func Auth(r *ghttp.Request) {
	tokenStr := r.GetHeader("Authorization")
	if tokenStr == "" {
		response.Unauthorized(r)
		return
	}

	// 支持 Bearer token 格式
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		response.Unauthorized(r)
		return
	}

	claims, err := jwt.ParseToken(tokenStr)
	if err != nil {
		response.Unauthorized(r, "Token无效或已过期")
		return
	}

	// 将用户信息写入 context
	r.SetCtxVar("jwt_user_id", claims.UserID)
	r.SetCtxVar("jwt_username", claims.Username)
	r.SetCtxVar("jwt_dept_id", claims.DeptID)
	r.SetCtxVar("jwt_claims", claims)

	r.Middleware.Next()
}
