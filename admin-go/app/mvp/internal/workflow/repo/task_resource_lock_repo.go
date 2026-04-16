package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskResourceLockRepo 任务资源锁仓储。
type TaskResourceLockRepo struct{}

func NewTaskResourceLockRepo() *TaskResourceLockRepo { return &TaskResourceLockRepo{} }

func (r *TaskResourceLockRepo) table() string { return "mvp_task_resource_lock" }

// ListHeldByTaskResourcesInTx 查询事务内指定任务已存在的资源锁。
func (r *TaskResourceLockRepo) ListHeldByTaskResourcesInTx(ctx context.Context, tx gdb.TX, taskID int64, resources []string) ([]string, error) {
	if len(resources) == 0 {
		return nil, nil
	}
	records, err := tx.Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		WhereIn("resource_path", resources).
		Fields("resource_path").
		All()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		resource := strings.TrimSpace(record["resource_path"].String())
		if resource == "" {
			continue
		}
		if _, ok := seen[resource]; ok {
			continue
		}
		seen[resource] = struct{}{}
		result = append(result, resource)
	}
	return result, nil
}

// UpdateHeldInTx 在事务内回写已存在的资源锁为 held。
func (r *TaskResourceLockRepo) UpdateHeldInTx(ctx context.Context, tx gdb.TX, taskID int64, resources []string, data g.Map) error {
	if len(resources) == 0 {
		return nil
	}
	_, err := tx.Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		WhereIn("resource_path", resources).
		Data(data).
		Update()
	return err
}

// BatchCreateInTx 在事务内批量创建资源锁。
func (r *TaskResourceLockRepo) BatchCreateInTx(ctx context.Context, tx gdb.TX, rows []g.Map) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := tx.Model(r.table()).Ctx(ctx).Insert(rows)
	return err
}

// ReleaseHeldByTask 释放指定任务仍处于 held 的资源锁。
func (r *TaskResourceLockRepo) ReleaseHeldByTask(ctx context.Context, taskID int64, releasedAt *gtime.Time) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		Where("lock_status", "held").
		Data(g.Map{
			"lock_status": "released",
			"released_at": releasedAt,
		}).
		Update()
	return err
}
