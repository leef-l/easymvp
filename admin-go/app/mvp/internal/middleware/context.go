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
	DataScopeAll           = 1 // 全部数据
	DataScopeDeptAndBelow  = 2 // 本部门及以下
	DataScopeDeptOnly      = 3 // 仅本部门
	DataScopePersonal      = 4 // 仅本人
	DataScopeCustom        = 5 // 自定义部门
)

// resolveUserDataScope 查询用户最大权限的 data_scope。
// 一个用户可能有多个角色，取最宽松的（数值最小的）。
func resolveUserDataScope(ctx context.Context, userID int64) (int, []int64) {
	// 查用户关联的所有角色
	roles, err := g.DB().Model("system_user_role AS ur").
		LeftJoin("system_role AS r", "r.id = ur.role_id").
		Fields("r.id, r.data_scope").
		Where("ur.user_id", userID).
		Where("r.status", 1).
		Where("r.deleted_at IS NULL").
		All()
	if err != nil || len(roles) == 0 {
		return DataScopePersonal, nil // 默认仅本人
	}

	bestScope := DataScopeCustom + 1 // 比最大值大，作为初始值
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

	// 如果最宽松的是自定义，则查询自定义部门列表
	var customDeptIDs []int64
	if bestScope == DataScopeCustom && len(customRoleIDs) > 0 {
		depts, _ := g.DB().Model("system_role_dept").
			Fields("DISTINCT dept_id").
			WhereIn("role_id", customRoleIDs).
			All()
		for _, d := range depts {
			customDeptIDs = append(customDeptIDs, d["dept_id"].Int64())
		}
	}

	return bestScope, customDeptIDs
}

// getChildDeptIDs 获取某部门及其所有子部门 ID（递归）。
func getChildDeptIDs(ctx context.Context, parentDeptID int64) []int64 {
	result := []int64{parentDeptID}
	children, err := g.DB().Model("system_dept").
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

// ApplyDataScope 数据权限过滤（五级 DataScope）。
// columns 参数指定需要过滤的列名（如 "created_by", "dept_id" 或带表别名 "p.created_by", "p.dept_id"）。
// 必须同时传入 created_by 列和 dept_id 列才能完整过滤。
func ApplyDataScope(ctx context.Context, m *gdb.Model, columns ...string) *gdb.Model {
	claims := GetClaims(ctx)
	if claims == nil {
		return m.Where("1 = 0")
	}

	// 超级管理员（UserID=1）不限制数据范围
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
		// 全部数据，不过滤
		return m

	case DataScopeDeptAndBelow:
		// 本部门及以下
		if deptIDCol != "" {
			deptIDs := getChildDeptIDs(ctx, claims.DeptID)
			return m.WhereIn(deptIDCol, deptIDs)
		}
		// 无 dept_id 列，退化为本人
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}

	case DataScopeDeptOnly:
		// 仅本部门
		if deptIDCol != "" {
			return m.Where(deptIDCol, claims.DeptID)
		}
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}

	case DataScopeCustom:
		// 自定义部门
		if deptIDCol != "" && len(customDeptIDs) > 0 {
			return m.WhereIn(deptIDCol, customDeptIDs)
		}
		// 无自定义部门配置，退化为本人
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}

	default:
		// DataScopePersonal 或未知值 → 仅本人
		if createdByCol != "" {
			return m.Where(createdByCol, claims.UserID)
		}
	}

	// 兜底：无可用列则禁止查询
	return m.Where("1 = 0")
}

// CheckOwnership 校验单条记录的数据归属（用于 Detail/Update/Delete）。
// 返回 nil 表示有权限，非 nil 表示无权限。
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

// CheckProjectAccess 项目作用域校验（替代 CheckOwnership 用于项目级操作）。
// 支持三级访问：owner（创建人）、dept_member（同部门/子部门/自定义部门）、admin（超管）。
func CheckProjectAccess(ctx context.Context, projectID int64) error {
	claims := GetClaims(ctx)
	if claims == nil {
		return fmt.Errorf("未登录")
	}
	// 超管不限制
	if claims.IsAdmin {
		return nil
	}

	// 查项目的 created_by 和 dept_id
	project, err := g.DB().Model("mvp_project").
		Fields("created_by, dept_id").
		Where("id", projectID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		return err
	}
	if project.IsEmpty() {
		return fmt.Errorf("项目不存在")
	}

	projectCreatedBy := project["created_by"].Int64()
	projectDeptID := project["dept_id"].Int64()

	// 1. Owner 直接通过
	if projectCreatedBy == claims.UserID {
		return nil
	}

	// 2. 按角色数据权限判断
	dataScope, customDeptIDs := resolveUserDataScope(ctx, claims.UserID)

	switch dataScope {
	case DataScopeAll:
		return nil

	case DataScopeDeptAndBelow:
		deptIDs := getChildDeptIDs(ctx, claims.DeptID)
		for _, d := range deptIDs {
			if d == projectDeptID {
				return nil
			}
		}

	case DataScopeDeptOnly:
		if projectDeptID == claims.DeptID {
			return nil
		}

	case DataScopeCustom:
		for _, d := range customDeptIDs {
			if d == projectDeptID {
				return nil
			}
		}
	}

	return fmt.Errorf("无权访问该项目")
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
