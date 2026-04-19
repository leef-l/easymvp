# EasyMVP V3 后台 Worker 与任务调度设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 目标：定义单机版 V3 的 Go Worker 模型、调度方式、失败处理和与主业务层的边界。

## 1. 设计结论

V3 单机版虽然不是分布式系统，但仍需要后台 worker。

否则：

1. run 状态同步会阻塞主流程
2. evidence / replay 索引会拖慢页面
3. 验收刷新会和命令事务竞争
4. 实时工作台快照更新不稳定

## 2. 首批 worker

建议至少有：

1. `run_sync_worker`
2. `evidence_index_worker`
3. `replay_index_worker`
4. `acceptance_refresh_worker`
5. `workspace_snapshot_refresh_worker`

## 3. Go 调度模型

建议由 `worker_manager` 统一管理：

1. goroutine 生命周期
2. 定时轮询任务
3. 队列消费任务
4. 并发上限
5. 退避与重试

原则：

1. 关键状态推进仍由 orchestrator 主导
2. worker 负责异步同步、索引、刷新

## 4. 调度原则

1. UI 触发命令，worker 异步执行长任务
2. worker 只写领域对象或索引对象
3. worker 通过事件让聚合层感知变化
4. 关键业务状态更新要通过 Service 或 orchestrator

## 5. 并发控制建议

建议：

1. 同一 `project_id` 的关键验收刷新避免并发重入
2. 同一 `run_id` 的同步任务串行化
3. 索引任务可低优先级并发
4. 快照刷新允许合并抖动

## 6. 失败处理

若 worker 失败：

1. 记录审计
2. 记录诊断
3. 页面显示 stale / warning
4. 允许后续重试
5. 对可恢复错误做指数退避

## 7. 观察性要求

每个 worker 至少要有：

1. `worker_name`
2. `job_id`
3. `project_id`
4. `started_at`
5. `finished_at`
6. `status`
7. `error_summary`

这样工作台和诊断页才能真正看懂后台发生了什么。

## 8. 不该怎么做

1. worker 直接改 UI
2. worker 绕过 orchestrator 写关键状态
3. 所有长任务都放在 Renderer
4. 把所有异步都塞进 Electron Main
