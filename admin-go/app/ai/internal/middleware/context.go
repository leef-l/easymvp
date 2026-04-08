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

// DataScope 常量（与 system_role.data_scope 一致）
const (
	DataScopeAll          = 1 // 全部数据
	DataScopeDeptAndBelow = 2 // 本部门及以下
	DataScopeDeptOnly     = 3 // 仅本部门
	DataScopePersonal     = 4 // 仅本人
	DataScopeCustom       = 5 // 自定义部门
)

// resolveUserDataScope 查询用户最大权限的 data_scope。
func resolveUserDataScope(ctx context.Context, userID int64) (int, []int64) {
	roles, err := g.DB().Ctx(ctx).Model("system_user_role AS ur").
		LeftJoin("system_role AS r", "r.id = ur.role_id").
		Fields("r.id, r.data_scope").
		Where("ur.user_id", userID).
		Where("r.status", 1).
		Where("r.deleted_at IS NULL").
		All()
	if err != nil || len(roles) == 0 {
		return DataScopePersonal, nil
	}

	bestScope := DataScopeCustom + 1
	var customRoleIDs []int64
	for _, r := range roles {
		scope := r["data_scope"].Int()
		if scope < bestScope {
			bestScope = scope
		}
		if scope == DataScopeCustom {
			customRoleIDs = append(customRoleIDs, r["id"].Int64())
		}
	}

	var customDeptIDs []int64
	if bestScope == DataScopeCustom && len(customRoleIDs) > 0 {
		depts, _ := g.DB().Ctx(ctx).Model("system_role_dept").
			Fields("DISTINCT dept_id").
			WhereIn("role_id", customRoleIDs).
			All()
		for _, d := range depts {
			customDeptIDs = append(customDeptIDs, d["dept_id"].Int64())
		}
	}

	return bestScope, customDeptIDs
}

// getChildDeptIDs 获取某部门及其所有子部门 ID（递归）
func getChildDeptIDs(ctx context.Context, parentDeptID int64) []int64 {
	result := []int64{parentDeptID}
	children, err := g.DB().Ctx(ctx).Model("system_dept").
		Fields("id").
		Where("parent_id", parentDeptID).
		Where("status", 1).
		WhereNull("deleted_at").
		All()
	if err != nil || len(children) == 0 {
		return result
	}
	for _, c := range children {
		childID := c["id"].Int64()
		result = append(result, getChildDeptIDs(ctx, childID)...)
	}
	return result
}

// ApplyDataScope 数据权限过滤（五级 DataScope）— 与 MVP 模块一致
func ApplyDataScope(ctx context.Context, m *gdb.Model, columns ...string) *gdb.Model {
	claims := GetClaims(ctx)
	if claims == nil {
		return m.Where("1 = 0")
	}

	// 超级管理员不限制数据范围
	if claims.IsAdmin {
		return m
	}

	// 解析列名
	var createdByCol, deptIDCol string
	for _, col := range columns {
		if col == "created_by" || strings.HasSuffix(col, ".created_by") {
			createdByCol = col
		}
		if col == "dept_id" || strings.HasSuffix(col, ".dept_id") {
			deptIDCol = col
		}
	}

	// 查询用户数据权限级别
	dataScope, customDeptIDs := resolveUserDataScope(ctx, claims.UserID)

	switch dataScope {
	case DataScopeAll:
		return m

	case DataScopeDeptAndBelow:
		if deptIDCol != "" {
			deptIDs := getChildDeptIDs(ctx, claims.DeptID)
			return m.WhereIn(deptIDCol, deptIDs)
		}
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}

	case DataScopeDeptOnly:
		if deptIDCol != "" {
			return m.Where(deptIDCol, claims.DeptID)
		}
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}

	case DataScopeCustom:
		if deptIDCol != "" && len(customDeptIDs) > 0 {
			return m.WhereIn(deptIDCol, customDeptIDs)
		}
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}

	default:
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}
	}

	return m.Where("1 = 0")
}

// CheckOwnership 校验单条记录的数据归属
func CheckOwnership(ctx context.Context, m *gdb.Model, id interface{}, idColumn string, createdByColumn string) error {
	claims := GetClaims(ctx)
	if claims == nil {
		return fmt.Errorf("未登录")
	}
	if claims.IsAdmin {
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
