package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/model/do"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/eventstream"
	"easymvp/app/mvp/internal/workflow/scheduler"
	"easymvp/utility/snowflake"
)

func TestRecoverWorkflowRunReconcilesCompletedExecuteStageWithoutRestart(t *testing.T) {
	engine.ClearConfigCache()

	var (
		ctx           = context.Background()
		workflowRunID = int64(snowflake.Generate())
		projectID     = int64(snowflake.Generate())
		stageRunID    = int64(snowflake.Generate())
		taskIDs       = []int64{int64(snowflake.Generate()), int64(snowflake.Generate())}
	)

	requireWorkflowTestDB(t, ctx)

	insertRecoveryWorkflowFixtures(t, ctx, workflowRunID, projectID, stageRunID, taskIDs)
	t.Cleanup(func() {
		cleanupWorkflowRecoveryFixtures(ctx, workflowRunID, taskIDs)
	})

	originalCreateRuntime := createRecoveredRuntime
	originalPrepareScheduler := prepareRecoveredTaskScheduler
	originalHasUnfinished := hasRecoveredWorkflowUnfinishedTasks
	originalStartScheduler := startRecoveredTaskScheduler
	originalReconcile := reconcileWorkflowProgressFn
	originalTaskScheduler := taskScheduler
	defer func() {
		createRecoveredRuntime = originalCreateRuntime
		prepareRecoveredTaskScheduler = originalPrepareScheduler
		hasRecoveredWorkflowUnfinishedTasks = originalHasUnfinished
		startRecoveredTaskScheduler = originalStartScheduler
		reconcileWorkflowProgressFn = originalReconcile
		taskScheduler = originalTaskScheduler
	}()

	taskScheduler = scheduler.NewDomainTaskScheduler()
	var (
		completedCallback bool
		callbackErr       error
		startCalls        int
	)
	taskScheduler.SetCompletionCallback(func(ctx context.Context, recoveredWorkflowRunID int64) {
		completedCallback = true
		_, callbackErr = g.DB().Model("mvp_workflow_run").Ctx(ctx).
			Where(do.MvpWorkflowRun{Id: recoveredWorkflowRunID}).
			Data(do.MvpWorkflowRun{
				Status:       consts.WorkflowRunStatusAccepting,
				CurrentStage: consts.StageTypeAccept,
			}).
			Update()
	})

	createRecoveredRuntime = func(recoveredWorkflowRunID, recoveredProjectID int64) {
		if recoveredWorkflowRunID != workflowRunID || recoveredProjectID != projectID {
			t.Fatalf("unexpected runtime restore args: workflow=%d project=%d", recoveredWorkflowRunID, recoveredProjectID)
		}
	}
	prepareRecoveredTaskScheduler = func(ctx context.Context, recoveredWorkflowRunID int64, stage string, recoveredStageRunID int64) error {
		if recoveredWorkflowRunID != workflowRunID || stage != consts.StageTypeExecute || recoveredStageRunID != stageRunID {
			t.Fatalf("unexpected scheduler prepare args: workflow=%d stage=%s stageRunID=%d", recoveredWorkflowRunID, stage, recoveredStageRunID)
		}
		return nil
	}
	hasRecoveredWorkflowUnfinishedTasks = func(ctx context.Context, recoveredWorkflowRunID int64) bool {
		return taskScheduler.HasUnfinished(ctx, recoveredWorkflowRunID)
	}
	startRecoveredTaskScheduler = func(ctx context.Context, recoveredWorkflowRunID int64) error {
		startCalls++
		return nil
	}
	reconcileWorkflowProgressFn = func(ctx context.Context, recoveredWorkflowRunID int64) bool {
		return taskScheduler.ReconcileWorkflowProgress(ctx, recoveredWorkflowRunID)
	}

	restarted, failed := recoverWorkflowRun(ctx, recoverableWorkflowRun{
		WorkflowRunID: workflowRunID,
		ProjectID:     projectID,
		Status:        consts.WorkflowRunStatusExecuting,
		Stage:         consts.StageTypeExecute,
		StageRunID:    stageRunID,
	})
	if failed {
		t.Fatal("expected recovery to succeed")
	}
	if restarted {
		t.Fatal("expected scheduler restart to be skipped for fully completed execute stage")
	}
	if startCalls != 0 {
		t.Fatalf("expected no scheduler restart, got %d", startCalls)
	}
	if callbackErr != nil {
		t.Fatalf("completion callback failed: %v", callbackErr)
	}
	if !completedCallback {
		t.Fatal("expected completion callback to be triggered by recovery reconcile")
	}

	record, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where(do.MvpWorkflowRun{Id: workflowRunID}).
		Fields("status, current_stage").
		One()
	if err != nil {
		t.Fatalf("query workflow_run failed: %v", err)
	}
	if got := record["status"].String(); got != consts.WorkflowRunStatusAccepting {
		t.Fatalf("unexpected workflow status after recovery: %s", got)
	}
	if got := record["current_stage"].String(); got != consts.StageTypeAccept {
		t.Fatalf("unexpected workflow stage after recovery: %s", got)
	}
}

func TestWorkflowRecoveryConsumerDedupesDuplicateTaskCompletedAcrossPolls(t *testing.T) {
	ctx := context.Background()
	redisClient := mustWorkflowTestRedis(t)

	var (
		now            = time.Now().UnixNano()
		streamName     = fmt.Sprintf("easymvp:test:workflow:recovery:%d", now)
		groupName      = streamName + ":group"
		idempotency    = fmt.Sprintf("wf:%d:task:%d:type:task.completed:attempt:1", now, snowflake.Generate())
		ledgerScope    = "workflow.recovery." + event.EventTaskCompleted
		workflowRunID  = int64(snowflake.Generate())
		entityID       = int64(snowflake.Generate())
		reconcileCalls int
	)

	t.Cleanup(func() {
		_, _ = redisClient.Do(ctx, "XGROUP", "DESTROY", streamName, groupName)
		_, _ = redisClient.Do(ctx, "DEL", streamName)
		_, _ = g.DB().Exec(ctx, "DELETE FROM mvp_workflow_event_ledger WHERE scope=? AND idempotency_key=?", ledgerScope, idempotency)
	})

	if _, err := redisClient.Do(ctx, "XGROUP", "CREATE", streamName, groupName, "0", "MKSTREAM"); err != nil &&
		!strings.Contains(strings.ToUpper(err.Error()), "BUSYGROUP") {
		t.Fatalf("create redis group failed: %v", err)
	}

	firstEvent := event.Event{
		EventID:        fmt.Sprintf("evt-first-%d", now),
		IdempotencyKey: idempotency,
		WorkflowRunID:  workflowRunID,
		EntityType:     event.EntityDomainTask,
		EntityID:       &entityID,
		EventType:      event.EventTaskCompleted,
		Attempt:        1,
		Payload: map[string]interface{}{
			"task_id": entityID,
		},
	}.EnsureMetadata()
	if err := publishRedisWorkflowEvent(ctx, redisClient, streamName, firstEvent); err != nil {
		t.Fatalf("publish first redis event failed: %v", err)
	}
	if _, err := redisClient.Do(ctx,
		"XREADGROUP",
		"GROUP", groupName, "stale-consumer",
		"COUNT", 1,
		"STREAMS", streamName, ">",
	); err != nil {
		t.Fatalf("seed pending message via stale consumer failed: %v", err)
	}

	originalGuard := recoveryEventGuard
	originalReconcile := reconcileWorkflowProgressFn
	defer func() {
		recoveryEventGuard = originalGuard
		reconcileWorkflowProgressFn = originalReconcile
	}()

	recoveryEventGuard = newRecoveryEventGuard(5 * time.Millisecond)
	reconcileWorkflowProgressFn = func(ctx context.Context, recoveredWorkflowRunID int64) bool {
		if recoveredWorkflowRunID != workflowRunID {
			t.Fatalf("unexpected workflowRunID in reconcile: %d", recoveredWorkflowRunID)
		}
		reconcileCalls++
		return true
	}

	consumer := eventstream.NewConsumer(redisClient, eventstream.Config{
		Enabled:            true,
		ConsumerEnabled:    true,
		StreamName:         streamName,
		ConsumerGroup:      groupName,
		ConsumerName:       "workflow-recovery-consumer",
		ReclaimIdleSeconds: 1,
		ReclaimCount:       10,
		ReadCount:          10,
	}, handleWorkflowRecoveryEvent)

	time.Sleep(1100 * time.Millisecond)
	if err := consumer.ProcessOnce(ctx); err != nil {
		t.Fatalf("first consumer poll failed: %v", err)
	}
	if reconcileCalls != 1 {
		t.Fatalf("expected first duplicate candidate to reconcile once, got %d", reconcileCalls)
	}

	time.Sleep(20 * time.Millisecond)
	duplicateEvent := firstEvent
	duplicateEvent.EventID = fmt.Sprintf("evt-duplicate-%d", time.Now().UnixNano())
	duplicateEvent.CreatedAtUnix = time.Now().Unix()
	if err := publishRedisWorkflowEvent(ctx, redisClient, streamName, duplicateEvent); err != nil {
		t.Fatalf("publish duplicate redis event failed: %v", err)
	}

	if err := consumer.ProcessOnce(ctx); err != nil {
		t.Fatalf("second consumer poll failed: %v", err)
	}
	if reconcileCalls != 1 {
		t.Fatalf("expected durable dedupe to suppress second reconcile, got %d", reconcileCalls)
	}

	ledgerRecord, err := g.DB().Model("mvp_workflow_event_ledger").Ctx(ctx).
		Where("scope", ledgerScope).
		Where("idempotency_key", idempotency).
		Fields("status").
		One()
	if err != nil {
		t.Fatalf("query durable ledger failed: %v", err)
	}
	if status := ledgerRecord["status"].String(); status != "handled" {
		t.Fatalf("expected durable ledger status=handled, got %s", status)
	}

	snapshot := consumer.Snapshot(ctx)
	if snapshot.Degraded {
		t.Fatalf("expected healthy consumer snapshot, got %+v", snapshot)
	}
	if !snapshot.PendingKnown || snapshot.Pending != 0 {
		t.Fatalf("expected pending queue drained after duplicate ack, got %+v", snapshot)
	}
}

func insertRecoveryWorkflowFixtures(t *testing.T, ctx context.Context, workflowRunID, projectID, stageRunID int64, taskIDs []int64) {
	t.Helper()

	if _, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).Data(do.MvpWorkflowRun{
		Id:                workflowRunID,
		ProjectId:         projectID,
		RunNo:             1,
		Status:            consts.WorkflowRunStatusExecuting,
		CurrentStage:      consts.StageTypeExecute,
		CurrentStageRunId: stageRunID,
		CreatedBy:         0,
		DeptId:            0,
	}).Insert(); err != nil {
		t.Fatalf("insert workflow_run failed: %v", err)
	}

	taskRows := []do.MvpDomainTask{
		{
			Id:            taskIDs[0],
			WorkflowRunId: workflowRunID,
			StageRunId:    stageRunID,
			TaskKind:      "implement",
			Name:          "recovery-test-task-1",
			Description:   "completed recovery test task 1",
			RoleType:      "implementer",
			RoleLevel:     "max",
			ExecutionMode: "chat",
			Status:        "completed",
			BatchNo:       1,
			Sort:          1,
			CreatedBy:     0,
			DeptId:        0,
		},
		{
			Id:            taskIDs[1],
			WorkflowRunId: workflowRunID,
			StageRunId:    stageRunID,
			TaskKind:      "implement",
			Name:          "recovery-test-task-2",
			Description:   "completed recovery test task 2",
			RoleType:      "implementer",
			RoleLevel:     "max",
			ExecutionMode: "chat",
			Status:        "completed",
			BatchNo:       1,
			Sort:          2,
			CreatedBy:     0,
			DeptId:        0,
		},
	}
	if _, err := g.DB().Model("mvp_domain_task").Ctx(ctx).Data(taskRows).Insert(); err != nil {
		t.Fatalf("insert domain_task fixtures failed: %v", err)
	}
}

func requireWorkflowTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("mysql integration unavailable: %v", err)
	}

	if _, err := g.DB().Exec(ctx, "SELECT 1"); err != nil {
		t.Skipf("mysql integration unavailable: %v", err)
	}
}

func cleanupWorkflowRecoveryFixtures(ctx context.Context, workflowRunID int64, taskIDs []int64) {
	if len(taskIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(taskIDs)), ",")
		args := make([]interface{}, 0, len(taskIDs)+1)
		for _, taskID := range taskIDs {
			args = append(args, taskID)
		}
		_, _ = g.DB().Exec(ctx, "DELETE FROM mvp_domain_task WHERE id IN ("+placeholders+")", args...)
	}
	_, _ = g.DB().Exec(ctx, "DELETE FROM mvp_workflow_run WHERE id=?", workflowRunID)
}

func mustWorkflowTestRedis(t *testing.T) *gredis.Redis {
	t.Helper()

	addr := strings.TrimSpace(os.Getenv("EASYMVP_TEST_REDIS_ADDR"))
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	}
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	pass := strings.TrimSpace(os.Getenv("EASYMVP_TEST_REDIS_PASS"))
	if pass == "" {
		pass = strings.TrimSpace(os.Getenv("REDIS_PASS"))
	}
	if pass == "" {
		t.Skip("set EASYMVP_TEST_REDIS_PASS or REDIS_PASS to run redis integration")
	}

	client, err := gredis.New(&gredis.Config{
		Address: addr,
		Pass:    pass,
		Db:      0,
	})
	if err != nil {
		t.Fatalf("create redis client failed: %v", err)
	}
	if _, err := client.Do(context.Background(), "PING"); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}
	return client
}

func publishRedisWorkflowEvent(ctx context.Context, client *gredis.Redis, streamName string, evt event.Event) error {
	evt = evt.EnsureMetadata()
	payloadJSON, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}

	args := []interface{}{
		streamName,
		"*",
		"event_id", evt.EventID,
		"idempotency_key", evt.IdempotencyKey,
		"event_type", evt.EventType,
		"workflow_run_id", strconv.FormatInt(evt.WorkflowRunID, 10),
		"entity_type", evt.EntityType,
		"attempt", strconv.Itoa(evt.Attempt),
		"created_at", strconv.FormatInt(evt.CreatedAtUnix, 10),
		"payload_json", string(payloadJSON),
	}
	if evt.EntityID != nil {
		args = append(args, "entity_id", strconv.FormatInt(*evt.EntityID, 10))
	}

	_, err = client.Do(ctx, "XADD", args...)
	return err
}
