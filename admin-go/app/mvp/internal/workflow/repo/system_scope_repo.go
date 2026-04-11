package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// SystemScopeRepo 系统权限范围相关仓储。
type SystemScopeRepo struct{}

func NewSystemScopeRepo() *SystemScopeRepo { return &SystemScopeRepo{} }

// GetUserDeptID 查询用户部门。
func (r *SystemScopeRepo) GetUserDeptID(ctx context.Context, userID int64) (int64, error) {
	record, err := g.DB().Ctx(ctx).Model("system_users").
		Fields("dept_id").
		Where("id", userID).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return 0, err
	}
	return record["dept_id"].Int64(), nil
}

// ListRoleScopesByUser 查询用户角色权限范围。
func (r *SystemScopeRepo) ListRoleScopesByUser(ctx context.Context, userID int64) ([]g.Map, error) {
	records, err := g.DB().Ctx(ctx).Model("system_user_role AS ur").
		LeftJoin("system_role AS r", "r.id = ur.role_id").
		Fields("r.id, r.is_admin, r.data_scope").
		Where("ur.user_id", userID).
		Where("r.status", 1).
		Where("r.deleted_at IS NULL").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListDeptIDsByRoleIDs 查询角色绑定的部门集合。
func (r *SystemScopeRepo) ListDeptIDsByRoleIDs(ctx context.Context, roleIDs []int64) ([]int64, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	records, err := g.DB().Ctx(ctx).Model("system_role_dept").
		Fields("DISTINCT dept_id").
		WhereIn("role_id", roleIDs).
		All()
	if err != nil {
		return nil, err
	}
	deptIDs := make([]int64, 0, len(records))
	for _, record := range records {
		deptID := record["dept_id"].Int64()
		if deptID > 0 {
			deptIDs = append(deptIDs, deptID)
		}
	}
	return deptIDs, nil
}

// ListChildDeptIDs 查询直属子部门 ID。
func (r *SystemScopeRepo) ListChildDeptIDs(ctx context.Context, parentDeptID int64) ([]int64, error) {
	records, err := g.DB().Ctx(ctx).Model("system_dept").
		Fields("id").
		Where("parent_id", parentDeptID).
		Where("status", 1).
		WhereNull("deleted_at").
		All()
	if err != nil {
		return nil, err
	}
	deptIDs := make([]int64, 0, len(records))
	for _, record := range records {
		deptIDs = append(deptIDs, record["id"].Int64())
	}
	return deptIDs, nil
}
