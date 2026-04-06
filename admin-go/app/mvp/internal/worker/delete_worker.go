package worker

// delete_worker.go
// 基于 Redis List 的异步删除 worker。
//
// 流程：
//   1. 调用方软删除主记录（立即返回）
//   2. EnqueueDelete 将 {entity, ids} JSON 推入 Redis List（LPUSH）
//   3. StartDeleteWorker 后台 goroutine BRPOP 阻塞监听，取到任务后逐步清理关联表
//   4. 每步单独执行，失败重试（最多 deleteMaxRetry 次），超限写 sys_delete_failed 留存

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2" // 注册 gredis 驱动

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

const (
	redisQueueKey   = "easymvp:delete:queue"   // Redis List key
	redisRetryKey   = "easymvp:delete:retry:%d" // 单条重试计数 key（ttl 1h）
	deleteMaxRetry  = 5
	deleteBatchSize = 100 // 单步最多处理的 ID 数
)

// deleteJob 队列中的一条删除任务。
type deleteJob struct {
	JobID  int64   `json:"job_id"`
	Entity string  `json:"entity"`
	IDs    []int64 `json:"ids"`
}

// StartDeleteWorker 启动异步删除 worker（随服务启动调用一次）。
func StartDeleteWorker(ctx context.Context) {
	// 定期清理 sys_delete_queue 中 done/failed 的历史记录（保留7天）
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				res, err := g.DB().Ctx(ctx).Model("sys_delete_queue").
					WhereIn("status", []string{"done", "failed"}).
					Where("updated_at < NOW() - INTERVAL 7 DAY").
					Delete()
				if err == nil {
					n, _ := res.RowsAffected()
					if n > 0 {
						g.Log().Infof(ctx, "[DeleteWorker] 清理过期队列记录 %d 条", n)
					}
				}
			}
		}
	}()

	// 定期扫描软删除残留（兜底：手动删除/异常中断后未入队的记录）
	go func() {
		// 启动5分钟后开始第一次扫描，避免服务启动瞬间抢锁
		time.Sleep(5 * time.Minute)
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				scanSoftDeleted(ctx)
			}
		}
	}()

	go func() {
		// 安全获取 Redis 实例，配置缺失时降级退出
		redis := safeGetRedis(ctx)
		if redis == nil {
			g.Log().Warning(ctx, "[DeleteWorker] Redis 未配置，异步删除 worker 已禁用（软删除记录将保留）")
			return
		}
		g.Log().Info(ctx, "[DeleteWorker] 启动，监听 Redis 队列")
		for {
			select {
			case <-ctx.Done():
				g.Log().Info(ctx, "[DeleteWorker] 停止")
				return
			default:
			}

			// BRPOP 阻塞等待，超时 5 秒重新循环（便于响应 ctx.Done）
			result, err := redis.BRPop(ctx, 5, redisQueueKey)
			if err != nil {
				// 连接异常时短暂等待重试
				g.Log().Warningf(ctx, "[DeleteWorker] BRPOP 错误: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}
			if len(result) < 2 {
				continue // 超时，正常继续
			}

			payload := result[1].String()
			var job deleteJob
			if err := json.Unmarshal([]byte(payload), &job); err != nil {
				g.Log().Warningf(ctx, "[DeleteWorker] 任务解析失败: %v | payload=%s", err, payload)
				continue
			}

			processJob(ctx, redis, &job, payload)
		}
	}()
}

// EnqueueDelete 将删除任务推入 Redis 队列（软删除后调用）。
// Redis 未配置时静默跳过（软删除记录保留，等待兜底扫描或手动清理）。
func EnqueueDelete(ctx context.Context, entity string, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	redis := safeGetRedis(ctx)
	if redis == nil {
		g.Log().Warningf(ctx, "[DeleteWorker] Redis 未配置，跳过入队 entity=%s ids=%v", entity, ids)
		return nil
	}
	job := deleteJob{
		JobID:  int64(snowflake.Generate()),
		Entity: entity,
		IDs:    ids,
	}
	payload, _ := json.Marshal(job)
	_, err := redis.LPush(ctx, redisQueueKey, string(payload))
	if err != nil {
		return fmt.Errorf("推入删除队列失败: %w", err)
	}
	g.Log().Infof(ctx, "[DeleteWorker] 入队 entity=%s ids=%v", entity, ids)
	return nil
}

// processJob 执行一条删除任务，失败时重新入队（带计数限制）。
func processJob(ctx context.Context, redis *gredis.Redis, job *deleteJob, rawPayload string) {
	retryKey := fmt.Sprintf(redisRetryKey, job.JobID)

	var execErr error
	switch job.Entity {
	case "mvp_task":
		execErr = deleteTaskCascade(ctx, job.IDs)
	default:
		execErr = fmt.Errorf("未知实体类型: %s", job.Entity)
	}

	if execErr == nil {
		g.Log().Infof(ctx, "[DeleteWorker] 完成 entity=%s ids=%v", job.Entity, job.IDs)
		redis.Del(ctx, retryKey)
		return
	}

	// 失败：查重试次数
	g.Log().Warningf(ctx, "[DeleteWorker] 执行失败 entity=%s: %v", job.Entity, execErr)
	countVal, _ := redis.Get(ctx, retryKey)
	count := 0
	if !countVal.IsNil() {
		count = countVal.Int()
	}
	count++

	if count >= deleteMaxRetry {
		g.Log().Errorf(ctx, "[DeleteWorker] 超过最大重试次数 entity=%s ids=%v，写入失败表", job.Entity, job.IDs)
		saveFailedJob(ctx, job, execErr.Error())
		redis.Del(ctx, retryKey)
		return
	}

	// 更新重试计数（1小时 TTL）
	redis.Set(ctx, retryKey, count)
	redis.Expire(ctx, retryKey, 3600)

	// 延迟重新入队（指数退避：count * 10s）
	delay := time.Duration(count*10) * time.Second
	go func() {
		time.Sleep(delay)
		redis.LPush(ctx, redisQueueKey, rawPayload)
		g.Log().Infof(ctx, "[DeleteWorker] 第%d次重试入队 entity=%s", count, job.Entity)
	}()
}

// scanSoftDeleted 扫描软删除超过30分钟但实际数据还在的记录，自动入队真删除。
// 通过 Redis SET 做分布式去重锁，避免多实例重复入队。
func scanSoftDeleted(ctx context.Context) {
	redis := safeGetRedis(ctx)
	if redis == nil {
		return // Redis 未配置，跳过扫描
	}

	// 分布式锁：同一时刻只有一个实例在扫描
	lockKey := "easymvp:delete:scan_lock"
	ok, err := redis.SetNX(ctx, lockKey, "1")
	if err != nil || !ok {
		return // 其他实例正在扫描
	}
	defer redis.Del(ctx, lockKey)
	// 锁有效期 5 分钟（扫描正常不超过这个时间）
	redis.Expire(ctx, lockKey, 300)

	type idRow struct{ ID int64 }

	// 扫描 mvp_task：软删除超过30分钟
	var rows []idRow
	err = g.DB().Ctx(ctx).Model("mvp_task").
		WhereNotNull("deleted_at").
		Where("deleted_at < NOW() - INTERVAL 30 MINUTE").
		Fields("id").
		Limit(500). // 每次最多处理500条
		Scan(&rows)
	if err != nil || len(rows) == 0 {
		return
	}

	// 过滤掉已经在队列中等待处理的 ID（避免重复入队）
	// 用 Redis SET 记录已入队的 ID，TTL 1小时
	toEnqueue := make([]int64, 0, len(rows))
	for _, r := range rows {
		inQueueKey := fmt.Sprintf("easymvp:delete:queued:mvp_task:%d", r.ID)
		exists, _ := redis.Exists(ctx, inQueueKey)
		if exists == 0 {
			toEnqueue = append(toEnqueue, r.ID)
			// 标记已入队
			redis.Set(ctx, inQueueKey, "1")
			redis.Expire(ctx, inQueueKey, 3600)
		}
	}

	if len(toEnqueue) == 0 {
		return
	}

	g.Log().Infof(ctx, "[DeleteWorker] 扫描到 %d 条软删除残留 mvp_task，入队清理", len(toEnqueue))

	// 分批入队，每批 deleteBatchSize 条
	for i := 0; i < len(toEnqueue); i += deleteBatchSize {
		end := i + deleteBatchSize
		if end > len(toEnqueue) {
			end = len(toEnqueue)
		}
		_ = EnqueueDelete(ctx, "mvp_task", toEnqueue[i:end])
	}
}

// saveFailedJob 将彻底失败的任务记录到数据库留存。
func saveFailedJob(ctx context.Context, job *deleteJob, errMsg string) {
	idsJSON, _ := json.Marshal(job.IDs)
	_, _ = g.DB().Ctx(ctx).Model("sys_delete_queue").Insert(g.Map{
		"id":          int64(snowflake.Generate()),
		"entity":      job.Entity,
		"ids":         string(idsJSON),
		"status":      "failed",
		"retry_count": deleteMaxRetry,
		"error_msg":   errMsg,
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	})
}

// safeGetRedis 安全获取 Redis 实例，配置不存在时返回 nil 而不是 panic。
func safeGetRedis(ctx context.Context) (r *gredis.Redis) {
	defer func() {
		if rec := recover(); rec != nil {
			g.Log().Warningf(ctx, "[DeleteWorker] Redis 配置未找到，降级跳过: %v", rec)
			r = nil
		}
	}()
	return g.Redis()
}

// ─── mvp_task 级联删除 ────────────────────────────────────────────────────────
// 分批次逐步执行，每步独立，无长事务。

func deleteTaskCascade(ctx context.Context, taskIDs []int64) error {
	for i := 0; i < len(taskIDs); i += deleteBatchSize {
		end := i + deleteBatchSize
		if end > len(taskIDs) {
			end = len(taskIDs)
		}
		if err := deleteTaskBatch(ctx, taskIDs[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func deleteTaskBatch(ctx context.Context, taskIDs []int64) error {
	db := g.DB().Ctx(ctx)

	// 1. 收集关联的 conversation_id
	type idRow struct{ ID int64 }
	var convRows []idRow
	_ = db.Model("mvp_conversation").
		WhereIn("task_id", taskIDs).
		Fields("id").Scan(&convRows)

	convIDs := make([]int64, 0, len(convRows))
	for _, r := range convRows {
		convIDs = append(convIDs, r.ID)
	}

	// 2. 收集 message_id → 删 message_chunk
	if len(convIDs) > 0 {
		var msgRows []idRow
		_ = db.Model("mvp_message").WhereIn("conversation_id", convIDs).Fields("id").Scan(&msgRows)
		msgIDs := make([]int64, 0, len(msgRows))
		for _, r := range msgRows {
			msgIDs = append(msgIDs, r.ID)
		}
		if len(msgIDs) > 0 {
			if _, err := db.Model("mvp_message_chunk").WhereIn("message_id", msgIDs).Delete(); err != nil {
				return fmt.Errorf("删除 message_chunk 失败: %w", err)
			}
		}
		if _, err := db.Model("mvp_message").WhereIn("conversation_id", convIDs).Delete(); err != nil {
			return fmt.Errorf("删除 message 失败: %w", err)
		}
		if _, err := db.Model("mvp_conversation").WhereIn("id", convIDs).Delete(); err != nil {
			return fmt.Errorf("删除 conversation 失败: %w", err)
		}
	}

	// 3. 删各子表
	if _, err := db.Model("mvp_task_log").WhereIn("task_id", taskIDs).Delete(); err != nil {
		return fmt.Errorf("删除 task_log 失败: %w", err)
	}
	if _, err := db.Model("mvp_task_dependency").
		Where("task_id IN (?) OR depends_on_id IN (?)", taskIDs, taskIDs).Delete(); err != nil {
		return fmt.Errorf("删除 task_dependency 失败: %w", err)
	}
	if _, err := db.Model("mvp_task_resource_lock").WhereIn("task_id", taskIDs).Delete(); err != nil {
		return fmt.Errorf("删除 task_resource_lock 失败: %w", err)
	}
	if _, err := db.Model("mvp_task_workspace").WhereIn("task_id", taskIDs).Delete(); err != nil {
		return fmt.Errorf("删除 task_workspace 失败: %w", err)
	}

	// 4. 最后真删主记录
	if _, err := db.Model("mvp_task").WhereIn("id", taskIDs).Delete(); err != nil {
		return fmt.Errorf("删除 task 失败: %w", err)
	}

	return nil
}
