# EasyMVP V3 技术栈与选型基线

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 关联文档：[EasyMVP-V3-Electron进程模型与IPC边界设计](./EasyMVP-V3-Electron进程模型与IPC边界设计.md)
> 目标：把 V3 单机版的正式技术栈定成可执行基线，并明确为什么核心实现应与 `brain-v3` 保持 Go 技术主线一致。

## 1. 设计结论

EasyMVP V3 单机版的推荐正式技术栈定为：

1. 桌面壳：`Electron`
2. 前端：`React + TypeScript + Vite`
3. 本地核心服务：`Go + GoFrame v2`
4. 本地数据库：`SQLite`
5. 本地文件存储：`Local Files`
6. 与 `brain-v3` 通讯：`brain serve HTTP API`
7. UI 与本地核心服务通讯：`localhost HTTP API`
8. Electron 原生能力桥：`IPC`

一句话：

> V3 首版建议采用 `Electron + React + GoFrame v2 + SQLite`，做成单用户、本地优先、可视化、实时驱动的项目工作台。

## 2. 为什么核心服务要用 Go

### 2.1 与 `brain-v3` 保持技术主线一致

`brain-v3` 已经是外部运行时底座。

如果 V3 本地核心服务也用 Go，会带来直接收益：

1. runtime adapter 的心智一致
2. HTTP / run / replay / log 处理风格一致
3. 后续若需要复用公共数据结构或客户端库，迁移成本更低
4. 文档与实现更容易长期稳定

### 2.2 Go 更适合本地工作流内核

V3 本地核心不是简单的 BFF，而是：

1. workflow orchestrator
2. event aggregation core
3. runtime adapter
4. worker scheduler
5. evidence / replay / acceptance index pipeline

这类能力更接近一个轻量本地工作流引擎。

Go 在这类场景的优势更明确：

1. 并发模型自然
2. 单二进制交付简单
3. 本地常驻进程资源占用更稳
4. worker / sync / queue / polling 实现更直接
5. 与 SQLite、本地文件系统、HTTP client/server 组合成熟

### 2.3 为什么不把所有业务放进 Electron Main

不建议让 Electron Main 承担业务内核。

因为那会导致：

1. UI 壳与业务引擎耦合
2. 进程职责不清
3. 长任务与窗口生命周期缠在一起
4. 后续替换桌面壳时核心能力不可复用

V3 正确做法是：

1. Electron 管桌面外壳
2. React 管可视化交互
3. Go 管本地业务内核

## 3. 正式技术栈清单

### 3.1 Desktop Shell

正式建议：

1. `Electron`

负责：

1. 窗口生命周期
2. 原生菜单与托盘
3. preload 安全桥接
4. 启动与守护本地 Go 服务
5. 桌面原生能力调用

### 3.2 Frontend

正式建议：

1. `React`
2. `TypeScript`
3. `Vite`

可选建议：

1. 路由：`TanStack Router`
2. 查询缓存：`TanStack Query`
3. 轻状态：`Zustand`
4. 图形可视化：`React Flow` 或 `visx`

前端负责：

1. 工作台页面
2. 计划页
3. 验收页
4. 设置页
5. 实时事件流可视化

### 3.3 Local Core Service

正式建议：

1. `Go 1.24+`
2. `GoFrame v2`

建议职责：

1. `workflow_orchestrator`
2. `plan_system`
3. `acceptance_system`
4. `runtime_adapter`
5. `view_aggregation`
6. `worker_manager`
7. `repository + storage`
8. `local_api`

说明：

1. V3 的主业务内核统一放在 Go 服务里
2. React 不直接承载业务状态机
3. Electron Main 不直接承载主业务规则
4. Go 服务框架正式选 `GoFrame v2`

### 3.4 为什么选 GoFrame v2

当前 V3 的本地核心服务需要同时满足：

1. 本地 HTTP API
2. 分层 service / repository 组织
3. 配置管理
4. 日志与诊断
5. 命令与查询边界清晰

`GoFrame v2` 更适合做这件事，因为它天然提供：

1. 应用结构约束
2. 配置与日志组件
3. 路由与中间件组织
4. 较完整的工程化基础设施

因此 V3 不再按“裸 Go + 自拼所有基础设施”去理解，而是按：

> `Go + GoFrame v2` 作为本地核心服务框架。

### 3.5 Database

正式建议：

1. `SQLite`
2. Go 驱动优先：`modernc.org/sqlite` 或团队确认后的 CGO 方案

选型原则：

1. 单机版默认内嵌数据库
2. 数据模型应支持事务边界
3. migration 必须可版本化

### 3.6 File Storage

正式建议：

1. 本地文件系统目录

存放：

1. evidence
2. replay
3. logs
4. exports
5. diagnostics
6. snapshots

### 3.7 Runtime Integration

正式建议：

1. `brain serve`
2. HTTP Run API

说明：

1. `brain-v3` 继续作为外部运行时底座
2. EasyMVP 不内嵌 `brain-v3` 内核实现
3. Go `runtime_adapter` 负责对接 `brain serve`

## 4. 通讯基线

### 4.1 UI 到核心服务

首版建议：

1. Electron Renderer 通过 `localhost HTTP API` 调 Go 服务
2. Electron IPC 仅承担桌面原生能力桥接

这样做的原因：

1. 页面层可稳定面向 `/api/v3/*`
2. React 端不需要理解 Electron channel 细节
3. 后续如果改成 Tauri 或 Web 容器，核心 API 仍可复用

### 4.2 实时事件

逻辑语义保持：

1. `snapshot + event stream`

首版物理实现建议：

1. Go 提供本地事件订阅接口
2. 前端可先使用轮询增量或流式响应
3. Electron IPC 只用于桌面壳事件，不作为业务事件总线主实现

## 5. 技术栈分层对应

```text
Electron Main
  ├─ app lifecycle
  ├─ window management
  ├─ native bridge
  └─ go core process supervisor

Renderer
  └─ React + TypeScript + Vite
          ↓
     localhost /api/v3/*
          ↓
Go Local Core Service
  ├─ Workflow Orchestrator
  ├─ Plan System
  ├─ Acceptance System
  ├─ Runtime Adapter
  ├─ View Aggregation
  ├─ Worker Manager
  └─ Repositories / Storage
          ↓
   SQLite + Local Files + brain serve
```

## 6. 建议版本基线

建议当前文档基线定为：

1. `Go 1.24+`
2. `GoFrame v2`
3. `Electron 31+`
4. `React 19`
5. `TypeScript 5.x`
6. `Vite 6`
7. `SQLite 3.45+`

这是一份实现基线，不是永久锁死版本。

## 7. 首版不建议的方向

首版不建议：

1. 把主业务规则写在 Electron Main
2. 用纯 Node app service 作为长期主架构
3. 让 Renderer 直接调用 `brain serve`
4. 让 Renderer 直接读写 SQLite
5. 先做 SaaS 型 Web 后台再套壳

## 8. 一句话结论

如果 V3 要做成长期可扩的本地工作台，且又要和 `brain-v3` 保持实现主线一致，那么核心服务就应该选 Go，而不是继续用 Node 作为主业务内核。
