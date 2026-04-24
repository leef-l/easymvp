# EasyMVP V3 Agent Team 开机计划总表

> 更新时间：2026-04-24
> 上游文档：[EasyMVP-V3-AgentTeam总章程](./EasyMVP-V3-AgentTeam总章程.md)
> 关联文档：[EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
> 目标：基于最新 `brain-v3` / `easymvp-brain` 边界，重新定义 Agent Team 的启动计划、依赖关系、并行方式和文档锚点。
>
> 当前状态以 [EasyMVP-V3-AgentTeam状态板](./EasyMVP-V3-AgentTeam状态板.md) 为准；本文件保留高层计划和后续优先级，不再承载旧 V3 对齐过渡记录。

## 1. 计划格式

每个计划使用以下字段：

1. `plan_id`
2. `owner_team`
3. `name`
4. `priority`
5. `depends_on`
6. `parallelizable`
7. `doc_anchor`
8. `definition_of_done`
9. `status`

## 2. 最新规划原则

本轮重规划后，Agent Team 按 4 条正式主线推进：

1. `domain-backend`
2. `domain-brain`
3. `runtime-projection`
4. `frontend-workbench`

边界原则：

1. `brain-v3` 是 runtime source of truth
2. EasyMVP App 是 domain/product source of truth
3. `easymvp-brain` 是 domain reasoning source
4. 不再新增与 `brain-v3` 重叠的通用 runtime 底座任务
5. `brain-v3` 内置 `code / browser / verifier / fault` 能力增强后，优先修订协作边界，不把这些能力重新并入 `easymvp-brain`
6. 页面、聚合层与领域脑只消费 EasyMVP 归一化后的 DTO / 事件 / 摘要，不直接消费 `brain-v3` 原始 payload
7. `unsupported / denied` 必须在 runtime / adapter / projection 链路中被显式保留，不能伪装成成功态

## 3. 总计划

| plan_id | owner_team | name | priority | depends_on | parallelizable | doc_anchor | definition_of_done | status |
|---|---|---|---|---|---|---|---|---|
| P-DM-001 | domain-backend | 保持 GoFrame system/workspace/projects API skeleton 稳定可编译 | P0 | 无 | 是 | `docs/EasyMVP-V3-GoFrame-Handler-DTO逐项终稿.md` | `apps/core` API skeleton、system/workspace/projects/replay/runtime 路由可编译且回归通过 | done |
| P-DM-002 | domain-backend | 维持 Project 创建主链路与基础目录初始化 | P0 | P-DM-001 | 否 | `docs/EasyMVP-V3-Service与Repository方法清单终稿.md`；`docs/EasyMVP-V3-本地目录与项目工作区规范.md` | `POST /api/v3/projects`、项目目录初始化、默认 profile 绑定和审计写入可用 | done |
| P-DM-003 | domain-backend | 维持 Workspace / Project / Plan / Acceptance 查询链路 | P0 | P-DM-001,P-DM-002 | 是 | `docs/EasyMVP-V3-核心API-DTO与TypeScript类型终稿.md` | `home-view`、`workspace-view`、`plan-view`、`acceptance-view` 真实查询保持可用 | done |
| P-RT-001 | runtime-projection | 维持 `brain-v3` runtime adapter 与 run binding 主链路 | P0 | P-DM-001 | 是 | `docs/EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md` | health/create/get/cancel/resume、binding 写库、状态同步与 execution-first 适配可用 | done |
| P-RT-002 | runtime-projection | 维持 replay / log segment / artifact 索引查询链路 | P0 | P-RT-001 | 是 | `docs/EasyMVP-V3-replay与log artifact存储规范.md`；`docs/EasyMVP-V3-Replay索引表结构设计.md` | `workflow_replay_index`、`workflow_run_log_segments` 查询与 raw 读取接口可用 | done |
| P-RT-003 | runtime-projection | 维持 worker manager 与 workspace SSE 主通道 | P1 | P-RT-001 | 是 | `docs/EasyMVP-V3-后台Worker与任务调度设计.md`；`docs/EasyMVP-V3-实时事件流推送机制设计.md` | `run_sync_worker`、`workspace_snapshot_refresh_worker`、最小 SSE 主通道可用 | done |
| P-BR-001 | domain-brain | 落 `easymvp-brain` 合同 DTO / Schema 校验骨架 | P0 | P-DM-001 | 是 | `docs/EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md` | `BrainContractEnvelope` 与 6 类合同 DTO、schema 校验逻辑、错误包装骨架可编译 | done |
| P-BR-002 | domain-brain | 落 `easymvp-brain` 调用服务与本地/远程统一客户端 | P0 | P-BR-001,P-RT-001 | 否 | `docs/EasyMVP-V3专精大脑接入计划.md#7. 正式调用链`；`docs/钱学森总纲设计/easymvp-brain-输入输出契约.md` | `internal/service/easymvp_brain.go` 可按统一合同调用本地 sidecar 或远程服务，且三类 P0 合同具备 typed wrapper | done |
| P-BR-003 | domain-brain | 接通 `plan_review` / `plan_compile` / `acceptance_mapping` 三个 P0 合同 | P0 | P-BR-002,P-DM-002 | 否 | `docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md`；`docs/EasyMVP-V3方案编译模型设计.md` | 可生成并落库 `PlanReviewResult`、`CompiledPlan`、`AcceptanceProfile` / `ProductionAcceptanceProfile` | done |
| P-BR-004 | domain-brain | 接通 `completion_adjudication` / `repair_design` | P1 | P-BR-003,P-RT-002 | 是 | `docs/EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md`；`docs/钱学森总纲设计/easymvp-brain-输入输出契约.md` | 可生成 completion decision 与 repair draft，失败进入 `ActionInbox` | done |
| P-BR-005 | domain-brain | 接通 `workspace_explanation` 推荐动作链路 | P1 | P-BR-003,P-DM-003,P-RT-003 | 是 | `docs/钱学森总纲设计/easymvp-brain-输入输出契约.md`；`docs/EasyMVP-V3工作台视图模型与聚合接口设计.md` | 工作台可读取 explanation headline、summary、recommended actions | done |
| P-BR-006 | domain-brain | 按 `brain-v3` 最新内置脑能力修订 `easymvp-brain` 协作边界 | P0 | P-BR-001 | 是 | `docs/EasyMVP-V3专精大脑接入计划.md#2.1-brain-v3-内置脑能力更新对-easymvp-的影响`；`docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md` | 文档明确 `code / browser / verifier / fault` 归属 `brain-v3`，`easymvp-brain` 只消费归一化领域摘要，不吸收执行工具面 | done |
| P-FE-001 | frontend-workbench | 初始化 Electron + React 桌面壳 | P0 | 无 | 是 | `docs/EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计.md` | `apps/desktop` 可启动空壳 | done |
| P-FE-002 | frontend-workbench | 落地 Workspace Home 页面骨架与状态组件 | P0 | P-FE-001,P-DM-003 | 是 | `docs/EasyMVP-V3-页面组件实现终稿与代码骨架规范.md` | Workspace 首页可渲染 loading/empty/error/normal | done |
| P-FE-003 | frontend-workbench | 落地 Project Workspace 页面骨架 | P1 | P-FE-001,P-DM-003,P-RT-003 | 是 | `docs/EasyMVP-V3-页面组件实现终稿与代码骨架规范.md` | 单项目工作台页面可渲染 | done |
| P-FE-004 | frontend-workbench | 接入领域脑 explain / acceptance / plan 真实交互 | P1 | P-BR-003,P-BR-005,P-FE-002,P-FE-003 | 是 | `docs/EasyMVP-V3-Plan详细页面设计.md`；`docs/EasyMVP-V3-Acceptance详细页面设计.md` | 前端能读取 plan/acceptance/explanation 真实结果并稳定展示 | done |
| P-OS-001 | orchestration-status | 维护状态板与中断恢复机制 | P0 | 无 | 是 | `docs/EasyMVP-V3-AgentTeam状态板.md` | 每次计划推进都有状态记录，重启后可直接恢复 | done |

## 4. 当前优先级重排

本轮原优先项已完成：

1. `IN-002`
2. `DG-002 / SX-003`
3. `RN-004 / AG-003`
4. `ST-005 / PF-002`

当前口径：

1. runtime / replay / SSE 主干已进入收口态
2. Electron 托管 Go core 的启动、退出清理、端口隔离、重连与 packaged smoke 证明链已补齐
3. replay/evidence 聚合读侧、恢复诊断与生产级幂等硬化已补齐
4. `brain-v3` 与 `easymvp-brain` 协作边界已锁定，后续不再扩张通用 runtime 底座
5. 下一步只做发布级验证，不新增实现计划

## 5. 并行策略

当前不再建议在低配服务器继续并发实现任务；发布验证按资源情况放到 CI 或高配机串行/分阶段执行：

1. 全量 `go test ./...`
2. `cd apps/desktop && pnpm run package:dir`
3. `cd apps/desktop && pnpm run verify:package`
4. 真实 `brain-v3` / `easymvp-brain` 端到端人工冒烟

发布验证前置动作：

1. `scripts/easymvp-backup-snapshot.sh snapshot <label>`
2. `scripts/easymvp-backup-snapshot.sh verify <snapshot_dir>`

## 6. 专项子计划

优先配合以下专项计划文档：

1. [EasyMVP-V3-AgentTeam-Backend计划](./EasyMVP-V3-AgentTeam-Backend计划.md)
2. [EasyMVP-V3-AgentTeam-Frontend计划](./EasyMVP-V3-AgentTeam-Frontend计划.md)
3. [EasyMVP-V3-AgentTeam-Runtime计划](./EasyMVP-V3-AgentTeam-Runtime计划.md)

## 7. 更新规则

1. 启动计划前，把 `status` 改为 `in_progress`
2. 完成后改为 `done`
3. 遇阻塞改为 `blocked`，并补阻塞原因到状态板
4. 已完成且仍有效的历史计划不删除，只在重规划中重分类
