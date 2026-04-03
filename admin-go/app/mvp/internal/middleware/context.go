package middleware

import (
	"context"
	"fmt"
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
		// 未认证请求禁止查询任何数据
		return m.Where("1 = 0")
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

// CheckOwnership 校验单条记录的数据归属（用于 Detail/Update/Delete）
// 返回 nil 表示有权限，非 nil 表示无权限
func CheckOwnership(ctx context.Context, m *gdb.Model, id interface{}, idColumn string, createdByColumn string) error {
	claims := GetClaims(ctx)
	if claims == nil {
		return fmt.Errorf("未登录")
	}
	// 超级管理员不限制
	if claims.UserID == 1 {
		return nil
	}
	count, err := m.Where(idColumn, id).Where(createdByColumn, claims.UserID).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("无权操作该数据")
	}
	return nil
}

// ValidateOrderBy 校验排序字段是否在白名单中
func ValidateOrderBy(orderBy string, allowedColumns []string) string {
	for _, col := range allowedColumns {
		if orderBy == col {
			return orderBy
		}
	}
	return ""
}
