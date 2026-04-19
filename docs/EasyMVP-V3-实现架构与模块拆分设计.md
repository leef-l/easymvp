# EasyMVP V3 实现架构与模块拆分设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 关联文档：[EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-技术栈与选型基线.md)
> 目标：把 V3 从“顶层理念架构”细化到“单机版可直接落模块的实现架构”，明确 Electron、React、Go 核心服务、SQLite、`brain-v3` 之间的职责与依赖方向。

## 1. 设计结论

对 V3 来说，仅有三层顶层架构还不够。

真正落地实现时，必须进一步细化成一组可编码的实现模块。

建议 V3 单机版正式按 9 个实现模块落地：

1. `electron_shell`
2. `workspace_renderer`
3. `desktop_bridge`
4. `local_api`
5. `view_aggregation`
6. `workflow_orchestrator`
7. `plan_system`
8. `acceptance_system`
9. `runtime_and_storage`

这里的关键变化是：

1. UI 层不再直接背业务
2. Go 本地核心服务成为唯一业务内核
3. Electron 只负责壳与原生桥
4. `brain-v3` 原始工具面只停留在 runtime adapter，不直接穿透到领域脑或页面

## 2. 为什么需要这一层

当前已有文档已经回答了：

1. V3 是什么
2. 主链路是什么
3. 页面长什么样
4. 单机版产品形态是什么

但还没有完全回答：

1. 代码模块怎么切
2. 哪些模块放 Electron
3. 哪些模块放 Go
4. 哪些逻辑属于 Orchestrator，哪些属于聚合层
5. 哪些属于业务 API，哪些只是桌面桥

## 3. 模块全景

推荐关系如下：

```text
Electron Shell
  ├─ Desktop Bridge
  └─ Workspace Renderer
          ↓
       Local API Client
          ↓
Go Local Core Service
  ├─ View Aggregation
  ├─ Workflow Orchestrator
  ├─ Plan System
  ├─ Acceptance System
  └─ Runtime And Storage
          ↓
    SQLite + Local Files + brain-v3
```

## 4. 模块拆分

### 4.1 `electron_shell`

放在 Electron Main。

负责：

1. 窗口级生命周期
2. 启动与关闭 Go 服务
3. 托盘、菜单、窗口状态
4. 崩溃感知与恢复入口
5. 桌面级诊断

不负责：

1. 项目状态推进
2. 计划编译
3. 验收裁决

### 4.2 `workspace_renderer`

放在 React Renderer。

负责：

1. `Workspace Home`
2. `Project Workspace`
3. `Plan`
4. `Acceptance`
5. `Settings`
6. 组件树与交互动画

不负责：

1. 直接访问数据库
2. 直接访问工作区文件系统
3. 承载主业务状态机

### 4.3 `desktop_bridge`

由 `preload + IPC` 提供。

负责：

1. 目录选择
2. 打开文件或目录
3. 系统通知
4. 壳层信息获取
5. 原生快捷动作

原则：

1. 只暴露桌面能力
2. 不暴露业务 Service

### 4.4 `local_api`

由 Go 本地核心服务提供。

负责：

1. `/api/v3/*` 查询接口
2. `/api/v3/*` 命令接口
3. 事件流接口
4. 健康检查与诊断接口

作用：

1. 成为 Renderer 与业务内核的唯一正式边界

### 4.5 `view_aggregation`

放在 Go 内核里。

负责：

1. 领域对象到页面对象投影
2. `WorkspaceView / PlanView / AcceptanceView`
3. `LiveEvent` 归一化
4. `ActionInboxItem` 聚合
5. `AcceptanceCoverage` 聚合

原则：

1. 只读投影
2. 不推进状态机

### 4.6 `workflow_orchestrator`

放在 Go 内核里。

负责：

1. 项目生命周期
2. 计划主链路推进
3. 任务状态推进
4. 验收主链路推进
5. 人工动作落地
6. 关键事务边界

这是 V3 的主业务内核。

### 4.7 `plan_system`

放在 Go 内核里。

负责：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. compiler
5. `CategoryProfile`
6. `RoleResolver`
7. `role_type / brain_kind` 归一化与路由策略

补充边界：

1. `plan_system` 只消费并产出标准 `role_type / brain_kind`
2. 它不是“执行脑编排层”，也不直接理解 `brain-v3` 原始工具协议

### 4.8 `acceptance_system`

放在 Go 内核里。

负责：

1. `AcceptanceProfile`
2. `ProductionAcceptanceProfile`
3. `AcceptanceRun`
4. coverage / issue / judgement
5. `manual_release_required`
6. `production_passed` 判定

按当前钱学森总纲，建议把这里理解为“旧验收系统职责表达”。

更准确的收口方向是：

1. `acceptance_system` 继续保留 `AcceptanceRun` 相关现实读写
2. 但验证结果应逐步落到 `VerificationResult`
3. 最终是否 `completed` 应逐步落到 `CompletionVerdict`
4. `production_passed` 不再单独承担最终完成语义

### 4.9 `runtime_and_storage`

放在 Go 内核里。

内部至少再拆成：

1. `runtime_adapter`
2. `worker_manager`
3. `repositories`
4. `sqlite_storage`
5. `file_storage`

分别负责：

1. 调用 `brain serve`
2. 同步 run / logs / replay
3. evidence / replay / diagnostics 索引
4. SQLite 事务与 migration
5. `brain-v3` 原始 `tools/list` / `tools/call`、`content[]`、`completed / failed / unsupported / denied` 到 EasyMVP 归一化状态的适配

边界要求：

1. `runtime_and_storage` 可以理解 `brain-v3` 原始协议
2. `plan_system` 和 `acceptance_system` 只能消费归一化后的 runtime / verification / fault / evidence 摘要
3. `workspace_renderer` 不得直接依赖 `brain-v3` 原始 payload
5. 本地目录与文件落盘

## 5. 依赖方向

V3 必须遵守单向依赖：

```text
Renderer → Local API → Aggregation / Orchestrator → Subsystems → Storage / Runtime
Renderer → Desktop Bridge → Electron Native
```

### 5.1 允许

1. UI 调 `local_api`
2. UI 调 `desktop_bridge`
3. 聚合层读取 orchestrator 和 repository 输出
4. orchestrator 调 plan / acceptance / runtime
5. 子系统调 storage

### 5.2 禁止

1. UI 直接调 storage
2. UI 直接调 `runtime_adapter`
3. UI 直接调 `brain serve`
4. aggregation 修改领域状态
5. storage 承载业务规则
6. Electron Main 直接当 orchestrator

## 6. 服务对象建议

建议至少显式存在以下 Go 服务：

1. `ProjectService`
2. `ProjectCreationService`
3. `PlanService`
4. `CompileService`
5. `TaskRunService`
6. `AcceptanceService`
7. `ManualDecisionService`
8. `WorkspaceViewService`
9. `SystemHealthService`

## 7. Repository 边界建议

建议至少显式区分：

1. `ProjectRepository`
2. `PlanRepository`
3. `TaskRepository`
4. `RunBindingRepository`
5. `AcceptanceRepository`
6. `EvidenceRepository`
7. `ReplayRepository`
8. `AuditRepository`
9. `SettingsRepository`

## 8. 事件边界建议

V3 应显式区分三类事件：

1. 领域事件
2. 运行时事件
3. 视图事件

说明：

1. 领域事件驱动状态推进
2. 运行时事件来自 `brain-v3`
3. 视图事件用于页面增量刷新

## 9. 最关键的架构约束

必须明确：

1. 页面是投影层，不是业务层
2. 聚合层是只读投影层，不是裁决层
3. Orchestrator 是主业务内核
4. `brain-v3` 是运行时底座，不是业务真相来源
5. SQLite 是持久化基础设施，不承载业务策略
6. Electron 是桌面壳，不是业务中心

## 10. 可以直接落的模块顺序

建议实现顺序：

1. `local_api + health check`
2. `sqlite_storage + migrations`
3. `project / plan / acceptance repositories`
4. `workflow_orchestrator`
5. `runtime_adapter`
6. `view_aggregation`
7. `workspace_renderer`
8. `desktop_bridge`
9. `worker_manager`

## 11. 下一层细化

要真正直接开工，还要继续补：

1. 单机版启动时序
2. 代码目录结构
3. Service / Repository 接口
4. API 路由分组
5. Worker 调度
