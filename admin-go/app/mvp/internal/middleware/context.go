package middleware

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/jwt"
)

// GetClaims 从 context 中获取 JWT Claims
func GetClaims(ctx context.Context) *jwt.Claims {
	val := g.RequestFromCtx(ctx).GetCtxVar("jwt_claims")
	if val.IsNil() {
		return nil
	}
	if claims, ok := val.Val().(*jwt.Claims); ok {
		return claims
	}
	return nil
}

// GetUserID 从 context 中获取当前用户 ID
func GetUserID(ctx context.Context) int64 {
	claims := GetClaims(ctx)
	if claims == nil {
		return 0
	}
	return claims.UserID
}

// GetDeptID 从 context 中获取当前用户部门 ID
func GetDeptID(ctx context.Context) int64 {
	claims := GetClaims(ctx)
	if claims == nil {
		return 0
	}
	return claims.DeptID
}

// ApplyDataScope 数据权限过滤
func ApplyDataScope(ctx context.Context, m *gdb.Model, columns ...string) *gdb.Model {
	claims := GetClaims(ctx)
	if claims == nil {
		return m
	}
	// 超级管理员（UserID=1）不限制数据范围
	if claims.UserID == 1 {
		return m
	}
	// 按 created_by 和 dept_id 过滤
	for _, col := range columns {
		if col == "created_by" || strings.HasSuffix(col, ".created_by") {
			m = m.Where(col, claims.UserID)
		}
	}
	return m
}
