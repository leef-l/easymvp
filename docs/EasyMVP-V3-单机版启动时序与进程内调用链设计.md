# EasyMVP V3 单机版启动时序与进程内调用链设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 关联文档：[EasyMVP-V3-Electron进程模型与IPC边界设计](./EasyMVP-V3-Electron进程模型与IPC边界设计.md)
> 目标：把单机版从“模块划分”进一步细化到“桌面壳启动、Go 核心服务启动、页面加载、命令调用、后台同步”的正式时序。

## 1. 设计结论

V3 单机版必须定义正式启动时序。

否则实现时很容易出现：

1. UI 先启动但 Go 核心未就绪
2. worker 启动早于数据库初始化
3. 页面开始拉工作台快照时聚合层还没准备好
4. Renderer 与本地服务的错误状态不一致

## 2. 启动阶段

建议固定为 8 个阶段：

1. `bootstrap_shell`
2. `spawn_local_core`
3. `health_check_core`
4. `init_storage`
5. `init_services`
6. `start_background_workers`
7. `mount_renderer`
8. `load_initial_snapshots`

## 3. 启动时序

```text
App Launch
  → Electron Main bootstrap
  → spawn Go local core process
  → health check /ready
  → Go init storage
  → run migrations
  → init repositories
  → init orchestrator services
  → init runtime adapter
  → start background workers
  → open renderer window
  → renderer load Workspace Home snapshot
```

## 4. 进程级职责时序

### 4.1 Electron Main

启动时：

1. 解析本地配置
2. 选择服务端口或读取固定端口策略
3. 启动 Go 服务进程
4. 轮询健康检查
5. 健康后再创建主窗口

### 4.2 Go Local Core

启动时：

1. 初始化工作目录
2. 初始化 SQLite
3. 执行 migration
4. 装配 repositories
5. 装配 services
6. 初始化 worker manager
7. 暴露 `/healthz` 与 `/api/v3/*`

### 4.3 Renderer

启动时：

1. 加载路由壳
2. 调用 workspace home query
3. 订阅工作台事件流
4. 渲染首页卡片、阶段、待处理、最近活动

## 5. 页面加载调用链

### 5.1 Workspace Home

```text
Renderer mount
  → GET /api/v3/workspace/home-view
  → render summary/cards/activity/readiness
  → connect workspace event stream
  → apply incremental updates
```

### 5.2 Project Workspace

```text
enter project
  → GET /api/v3/projects/{id}/workspace-view
  → render stage/activity/inbox/coverage
  → connect project event stream
  → open drawers on demand
```

### 5.3 Plan

```text
enter plan page
  → GET /api/v3/projects/{id}/plan-view
  → render draft/review/compiled
  → load projection detail on demand
```

### 5.4 Acceptance

```text
enter acceptance page
  → GET /api/v3/projects/{id}/acceptance-view
  → load evidence cards
  → load replay/evidence details on demand
```

## 6. 创建项目调用链

```text
Renderer submit create form
  → desktop.selectDirectory() if needed
  → POST /api/v3/projects
  → ProjectCreationService.create
  → persist project record
  → init workspace structure
  → bind category / role / acceptance profiles
  → bootstrap plan draft
  → emit creation events
  → renderer navigates to project workspace
```

## 7. Run 执行调用链

```text
DomainTask ready
  → TaskRunService.start
  → RuntimeAdapter.createRun
  → persist BrainRunBinding
  → worker manager starts run sync loop
  → normalize runtime events
  → update repositories
  → refresh aggregation snapshots
  → renderer receives incremental update
```

## 8. 验收调用链

```text
Acceptance requested
  → AcceptanceService.start
  → build AcceptanceRun
  → resolve acceptance profile
  → collect evidence/index references
  → compute coverage / issues / judgement
  → emit acceptance events
  → refresh Acceptance view
```

补充说明：

- 这条调用链可以继续表示当前首版实现现实
- 但按当前钱学森总纲，后续验收链路更准确的收口应是：
  - `AcceptanceRun` 作为历史/兼容运行对象
  - `VerificationResult` 作为验证结果对象
  - `CompletionVerdict` 作为最终完成裁决对象
- 因此 “验收调用链” 不应再被误读为“AcceptanceRun 一步定义最终完成”

## 9. 后台 worker 原则

建议首批 worker：

1. `run_sync_worker`
2. `evidence_index_worker`
3. `replay_index_worker`
4. `acceptance_refresh_worker`
5. `workspace_snapshot_refresh_worker`

规则：

1. worker 由 Go 内核管理
2. worker 不直接操作 UI 状态
3. worker 只写领域对象或索引对象
4. 通过事件和聚合层让 UI 感知变化

## 10. 启动与运行失败处理

### 10.1 Go 服务启动失败

若 Go 服务无法启动：

1. Electron 不进入正常工作台
2. 应进入启动错误页或恢复页
3. 展示端口、路径、数据库、可执行文件等诊断信息

### 10.2 migration 失败

若 migration 失败：

1. Go 服务标记为不可用
2. Renderer 只进入恢复模式
3. 不允许继续发业务命令

### 10.3 worker 失败

若后台 worker 失败：

1. 主 UI 不应崩溃
2. 页面表现为 stale 或 warning
3. 记录诊断与审计
4. 允许手动重试或自动退避重试
