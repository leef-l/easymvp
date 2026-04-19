# EasyMVP V3 Agent Team 状态板

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-AgentTeam开机计划总表](./EasyMVP-V3-AgentTeam开机计划总表.md)
> 目标：作为持续更新的执行状态面板，在中断后可直接恢复当前推进位置。

## 1. 当前总状态

| area | status | summary | last_updated |
|---|---|---|---|
| docs-finalization | done | `easymvp-brain`、AgentTeam 计划/实施、页面聚合、Evidence、事件流、实现架构、重启接续等文档已统一到同一套边界语义 | 2026-04-19 |
| backend-core | in_progress | 基础查询、项目创建、runtime adapter、replay 查询、worker manager、workspace SSE 首版均已可用；`plan compile -> domain_tasks` 桥接、acceptance profile refresh、`acceptance-runs.task_id`、repair API `artifact_refs`、run terminal 自动 adjudication、workspace snapshot 真刷新、manual release 命令入口、audit 日志列表 API、formal replay/log 写侧首版均已落地，当前主要剩更厚的 replay 语义投影与端到端联调 | 2026-04-19 |
| frontend-workbench | in_progress | `apps/desktop` 已完成应用壳、统一 API client、基础 Query/Recovery 协议，以及 Workspace / Plan / Execution / Acceptance / Settings 页面接线；本轮继续补了创建项目入口、Diagnostics 页面、Execution 深链、manual release 按钮、Audit 页面与 Replay 路由别名，并修正 replay summary/timeline/detail 契约消费、diagnostics→replay/audit 跳转与 audit pretty-json 展示，`npm run build` 已通过 | 2026-04-19 |
| runtime-storage | done | migration、dao/entity/do、项目目录初始化、runtime adapter、replay 正式索引链路都已验证通过，当前进入维护与投影增强阶段 | 2026-04-19 |
| domain-brain | in_progress | `BE-019 / P-BR-001` 已完成；`BE-020 / P-BR-002` 已补统一客户端骨架、合同 DTO 的 `json.RawMessage` 编码修正，以及 `plan_review / plan_compile / acceptance_mapping / completion_adjudication / repair_design / workspace_explanation` typed wrapper；`BE-021 / P-BR-003` 已落 `plan compile -> domain_tasks`、acceptance mapping/profile refresh、acceptance start、repair API artifact refs、workspace explain 与自动 adjudication 首版链路 | 2026-04-19 |
| agent-team | in_progress | 总章程、总表、状态板、后端/前端/运行时专项计划已重排，已把新关键路径切到 domain-brain lane | 2026-04-19 |

## 2. 当前已完成

1. 已安装 `goframe-v2` skill 到本地磁盘
2. 已初始化 `apps/core` GoFrame v2 骨架
3. 已创建 `system/healthz` 最小入口
4. 已落真实 migration 文件到 `apps/core/manifest/migrations`
5. 已补关键卡片的单独 props 终稿文档
6. 已建立 Agent Team 总章程、总表与状态板
7. 已建立后端、前端、运行时三条专项计划文档
8. `P-BE-001` 已完成，GoFrame API skeleton 可编译
9. `P-RT-001` 已完成，SQLite migration 启动链路已验证
10. `P-RT-002` 已完成，`dao / do / entity` 已生成且 `go test ./...` 通过
11. `P-BE-002` 已完成，`POST /api/v3/projects` 已接真实写库事务链路
12. `P-BE-003` 已完成，`GET /api/v3/workspace/home-view` 已接真实 SQLite 聚合查询
13. `P-BE-004` 已完成，`GET /api/v3/projects/{id}/workspace-view` 已接真实 SQLite 聚合查询
14. `P-BE-005` 已完成，`GET /api/v3/projects/{id}/plan-view` 已接真实或最小派生聚合查询
15. `P-BE-006` 已完成，`GET /api/v3/projects/{id}/acceptance-view` 已接真实或最小派生聚合查询
16. `P-RT-003` 已完成，顶层 `dataRoot/projects` 与项目级正式目录初始化已收口，且保留业务 `workspace_root`
17. `P-BE-013` 已完成 runtime client 收口：`brain-v3` health/create/get/cancel + CLI `resume`、run binding 写库、状态映射、首批 `run_event_index` 写入与 `RUN_001/002/003/004` 错误包装已可用
18. `P-BE-014` 已完成，worker manager 已接入启动链路，`run_sync_worker` 会真实扫描并同步 `brain_run_bindings`，`workspace_snapshot_refresh_worker` 已具备安全轮询、可停止与失败审计/诊断写库能力
19. `P-BE-015` 已完成，`GET /api/v3/workspace/projects/{id}/events` 最小 SSE 主通道已接入，基于 `run_event_index` 推送命名事件，支持 `id`、keepalive、`Last-Event-ID` 最小续传与 `workspace.snapshot_invalidated` 提示
20. 已完成 `easymvp-brain` 重新设计，明确其只做领域脑，并补齐 Manifest、合同、JSON Schema 与远程兼容约束
21. 已完成 `BE-019 / P-BR-001` 首版代码骨架：`internal/model/braincontracts` 与 `internal/service/easymvp_brain.go` 已落地并通过 `go test`
22. 已完成 `BE-020 / P-BR-002` 首版客户端骨架：`internal/service/easymvp_brain_client.go` 与配置项已落地，并通过 `go test ./internal/model/braincontracts ./internal/service`
23. 已补 `easymvp-brain` typed contract wrapper：`plan_review`、`plan_compile`、`acceptance_mapping` 三类调用已在 `internal/service/easymvp_brain.go` 收口；同时修正合同 DTO 的 JSON 内联编码，避免对象字段被错误编码为 base64 字符串
24. 已启动 `BE-021 / P-BR-003`：新增 `POST /api/v3/projects/{id}/plan/compile` API/Controller、`service.Plan().CompilePlan(...)` 写侧命令、项目创建后自动 `CreateInitialDraft(...)`、以及首版 `plan_review -> workflow_plan_review_results` / `plan_compile -> workflow_compiled_plans + workflow_compiled_tasks` 持久化链路；`go test ./internal/service ./internal/controller/... ./api/plan/...` 已通过
25. 已继续推进 `BE-021 / P-BR-003` 的 acceptance 侧：新增 migration `0010_add_acceptance_profile_tables.sql`、`service.Acceptance().MapAcceptanceProfiles(...)`、`acceptance_profiles / production_acceptance_profiles` 首版持久化，以及 acceptance 视图优先读取真实 profile 的 fallback；`go test ./internal/service ./internal/controller/... ./api/acceptance/...` 已通过
26. 已补 `POST /api/v3/projects/{id}/acceptance-runs` 首版正式命令入口：`service.Acceptance().StartAcceptanceRun(...)` 会先确保 `acceptance_mapping` 完成，再写入 `acceptance_runs` 与首批 `acceptance_surface_coverage / acceptance_journey_coverage`；同时项目工作台 acceptance coverage 统计已优先读取真实 profile 表
27. 已完成文档收口：总纲、总体架构、专精大脑接入、职责边界、Manifest、Prompt、合同 Schema、Backend 计划、AgentTeam 总表已统一明确 `brain-v3` 负责内置 `code / browser / verifier / fault` 与 `tools/list` / `tools/call` 协议，EasyMVP 只消费归一化结果，`unsupported / denied` 必须显式保留
28. 已继续完成计划/实施类文档收口：创建流、事件流、Workspace 首页聚合、Evidence 查询/Preview/索引/命名、ProductionAcceptanceProfile、方案编译、工作台视图聚合、Live Activity、前端计划、开机总表、计划数据模型、本地核心服务架构、重启接续说明均已同步到同一套边界语义
29. 已把 acceptance mapping 挂到正式命令入口：新增 `POST /api/v3/projects/{id}/acceptance-profiles/refresh`，并在 `plan compile` 后自动刷新 acceptance profile，避免页面只能依赖 `start acceptance` 隐式补建
30. 已补 acceptance/repair 链路关键缺口：`POST /api/v3/projects/{id}/acceptance-runs` 新增可选 `task_id`、启动时会把 `projects.status` 推进到 `acceptance`；`POST /api/v3/projects/{id}/repair-draft` 已补 `artifact_refs` 透传；自动 repair 草稿已修复 `domainTask == nil` 的 panic 风险
31. 已补回归测试并通过：`go test ./internal/service ./internal/controller/... ./api/acceptance/... ./api/plan/...`
32. 已完成 `plan compile -> domain_tasks` 首版桥接：编译后会把 `workflow_compiled_tasks` 物化为 `domain_tasks`，并按当前 `current_compiled_plan_id` 过滤 project workspace / plan 读取，避免旧计划任务污染当前工作台
33. 已把自动 adjudication 接到公共命令层：`Runtime().SyncRunBindingCommand(...)` 会在 binding 首次进入 `run_succeeded` / `run_failed` 时，按 `project_id + task_id` 对齐最新 acceptance run 并自动触发 adjudication，同时写入 `acceptance.auto_adjudicated` 事件
34. 已把 `workspace_snapshot_refresh_worker` 改成真实快照刷新：项目工作台快照写入 `project_snapshots`，workspace 首页快照写入 `workspace_snapshots`；实时读路径成功后也会顺手 upsert，失败时可回退到最近快照
35. 已补本轮验证并通过：`go test ./internal/service/...` 与 `go test ./internal/controller/... ./api/...`
36. 已完成桌面端一轮主链路接线并通过构建验证：`apps/desktop` 已接入统一 `apiGet/apiPost/apiDelete`、stale-while-revalidate `useQuery`、`QueryPanel` recovery 状态，以及 Workspace SSE 刷新、Plan compile/repair、Acceptance refresh/start/adjudicate、Execution start/sync/resume/cancel 操作；`npm run build` 已通过
37. 已继续推进桌面端联调链路：`acceptance_run` 已补 `task_id / profile_version / finished_at / latest_judgement_*` 视图字段；前端已接入创建项目表单、Diagnostics 页面、Execution query 深链、以及 `QueryPanel` 的统一 retry/secondary recovery 动作；`npm run build` 已再次通过
38. 已补通用诊断写入支撑：新增 `diagnostic_support.go` 统一插入 `diagnostic_records`，worker 写入已复用；`runtime.create_run`、`runtime.sync_run` 与 `easymvp-brain` execute/decode 失败现在会主动写诊断事实，且 `go test ./internal/service/... ./internal/controller/... ./api/...` 已通过
39. 已补 acceptance manual release 闭环：新增 `POST /api/v3/projects/{id}/acceptance-runs/manual-release` 命令入口，事务内会写 `task_manual_gates`、`acceptance_judgements`、`audit_logs`，并把 `acceptance_runs / projects` 推进到已人工放行后的完成态；`AcceptancePage` 已接入审批按钮
40. 已补薄审计链与 replay 入口：新增 `GET /api/v3/projects/{project_id}/audit-logs`、desktop `AuditPage`、`/replay` 路由别名与导航入口；同时修正 workspace 首页 release readiness 缺项计算，避免人工放行后仍被误算为缺项；`go test ./internal/service/... ./internal/controller/... ./api/...` 与 `npm run build` 均已通过
41. 已补 formal replay/log 写侧首版：新增 `replay_support_write.go`，`Runtime().SyncBrainRunBinding(...)` 现在会按 `brain_run_bindings + project_workspaces.runs_root/replay_root` 扫描 run artifact，并幂等刷新 `workflow_replay_index / workflow_run_log_segments`；索引失败会写 `diagnostic_records` 但不阻断主同步链路，`go test ./internal/service/... ./internal/controller/... ./api/...` 已通过
42. 已补桌面端 replay/audit/diagnostics 收口：`ExecutionPage` 已切到后端真实 replay contract，补齐 `artifact_status_summary / entry_points / seq_no / raw_target / source_object_id / event_id / span_id / started_at / ended_at` 展示；`DiagnosticsPage` 已增加 Replay/Audit 跳转；`AuditPage` 已补 pretty-json 与回跳入口；`npm run build` 已通过
43. 已继续收口 `BE-022` 三个关键缺口：runtime 适配层现已显式保留 `unsupported / denied -> run_unsupported / run_denied`，并投影到 run event / workspace live activity / action inbox / workspace explanation fallback；同时自动 adjudication 已改为按命中的 `acceptance_run_id` 精准裁决，避免跨任务 acceptance run 串单；`workspace_explanation` 在上游 runtime `unsupported / denied` 时也会输出显式 capability / policy 文案而不是静默通用 fallback；`go test ./internal/service/... ./internal/controller/... ./api/...` 已通过
44. 已继续补桌面端页面联动与 replay 语义投影：`WorkspacePage` 现已把 `explain_links` 做成可跳转入口，并把 `open_task_review` / `open_acceptance_center` 带 task 维度导航；`AuditPage` 已补 event type / actor / keyword 过滤；`SettingsPage` 已补创建流初始化态与创建成功后的 workspace/plan/execution 快捷入口；后端 `replay_support_write.go` 已新增轻量 metadata 提取，把 replay 索引从纯路径摘要升级为 `title / summary / source_object / event_id / trace_id / span_id` 语义投影；`go test ./internal/service/... ./internal/controller/... ./api/...` 与 `cd apps/desktop && npm run build` 已通过
45. 已补 `BE-018`/recovery 最小闭环：新增 `GET /api/v3/projects/{project_id}/diagnostic-records`，后端会按 `diagnostic_records.detail_json` 投影 `project_id / task_id / run_id / binding_id`，桌面端 `DiagnosticsPage` 已接入真实诊断记录列表、结构化恢复上下文与跳转入口；`go test ./internal/service/... ./internal/controller/... ./api/...` 与 `cd apps/desktop && npm run build` 已通过
46. 已继续补 `BE-022` 与 execution 深链闭环：`plan_support_repair.go` 现已按 `failed_task_context / failure_reason / original_contracts / runtime_summary / artifact_refs` 做 repair draft 幂等复用，相同失败上下文不会重复生成新草稿或重复调领域脑；桌面端 `ExecutionPage` 已接入相关 diagnostics 列表、runtime event payload pretty-json，以及从 replay/detail/diagnostics 跳 acceptance / diagnostics 的深链；`go test ./internal/service/... ./internal/controller/... ./api/...` 与 `cd apps/desktop && npm run build` 已通过
47. 已继续补 `Acceptance -> Repair` 页面联动：`AcceptancePage` 现已接入 related diagnostics、repair context 卡片，并从 issue/diagnostic 卡片直接跳 `Execution / Diagnostics / Repair Draft`；`RepairDraftPage` 现已从 `failed_task_context_json` 解析 task 上下文，并提供回到 `Acceptance / Execution / Diagnostics / Plan` 的真实入口；`cd apps/desktop && npm run build` 已通过

## 3. 当前进行中

1. `BE-020 / P-BR-002` 继续进行中，当前已完成统一客户端骨架与三类 P0 合同 typed wrapper，待继续完成与业务服务的最终收口后关闭
2. `BE-021 / P-BR-003` 进行中，当前已接通 `plan_review / plan_compile / domain_tasks / acceptance_mapping / acceptance profile refresh / acceptance-runs / repair artifact refs / auto adjudication / workspace snapshot refresh / formal replay-log indexing` 首版写侧链路，并补了 replay metadata 投影与 repair draft 幂等，下一步补 acceptance / replay 读取侧更深消费与联调
3. `FE-003 / FE-004 / FE-007 / FE-008 / FE-011 / FE-012 / FE-014 / FE-015 / FE-016 / FE-017 / FE-018 / FE-020` 继续进行中，当前主任务已切到 execution / acceptance / diagnostics / audit 更细粒度联动、更多 recovery 视图，以及 diagnostics -> execution -> acceptance -> repair 的页面级联调验证补齐

文档层当前不再是阻塞项；剩余主工作均转入代码实现与联调。

## 4. 当前待做

1. 完成 `BE-020 / P-BR-002` 的业务服务接线与最终收口
2. 继续完成 `BE-021 / P-BR-003` 中 acceptance / execution / replay 读取侧对更厚 replay/log 语义投影的消费与收口
3. 继续把 `workspace_explanation` 的前端消费链路向 execution / acceptance / diagnostics / repair 深化联调
4. 继续补齐前端 recovery 视图、审计详情结构化展示，以及 acceptance / repair 对诊断上下文的更深消费
5. 继续完成前后端联调与页面级验证

## 5. 阻塞记录

1. `easymvp-brain` 已进入真实代码层并完成 DTO / client / typed wrapper 首版接线；当前阻塞点不在“是否开工”，而在 completion / repair / explain 三条合同的最终业务收口。

## 6. 更新约定

每完成一个计划时，必须同步更新：

1. 本文件的“当前已完成”
2. 本文件的“当前进行中”
3. `EasyMVP-V3-AgentTeam开机计划总表.md` 对应 `status`
