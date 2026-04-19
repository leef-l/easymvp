# EasyMVP V3 AgentTeam Runtime 计划

> 更新时间：2026-04-19
> 适用范围：V3 运行时 / 存储 / 集成实施
> 关联文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go本地核心服务架构设计.md)
> 目标：基于最新 V3 边界，为运行时、存储、集成三条实施主线给出可执行、可并行、可恢复的开机计划，并明确哪些能力继续归 `brain-v3`，哪些只做投影层。

## 1. 使用规则

这份计划只服务于以下三条实施主线：

1. 本地 GoFrame v2 核心服务
2. SQLite / 本地目录 / migration / 事务 / 索引
3. Electron / 本地 API / IPC / `brain-v3` 集成

每个计划项必须具备：

1. `plan_id`
2. `name`
3. `priority`
4. `depends_on`
5. `parallelizable`
6. `doc_refs`
7. `definition_of_done`
8. `status`

状态字段默认值统一为：

- `pending`

可选后续状态统一为：

- `in_progress`
- `blocked`
- `completed`
- `cancelled`

补充边界：

1. 本计划不负责 `easymvp-brain` 领域合同实现
2. 本计划只负责 `brain-v3` runtime 接入、投影、索引、聚合支撑
3. 任何新任务不得把 EasyMVP 扩展成通用 runtime 底座
4. `brain-v3` 的 `tools/list` / `tools/call`、`completed / failed / unsupported / denied` 状态变化优先在 runtime adapter 层吸收
5. 运行时 lane 不向领域层透传原始工具名、原始 `content[]` 或未归一化 payload

## 2. 优先级定义

- `P0`：不开工会阻塞后续大部分实现
- `P1`：主链路实现必需，但可在 P0 后并行展开
- `P2`：增强项、配套项、恢复项、诊断项

## 3. 并行规则

允许并行的前提：

1. 不共享同一组核心文件
2. 不同时改同一批数据库对象
3. 不同时改同一条主调用链
4. 不破坏前置计划的完成定义

建议按 4 条执行泳道分配 agent：

1. `Runtime Lane`：Go 核心服务、worker、brain 运行时
2. `Storage Lane`：SQLite、migration、索引、事务、目录
3. `Integration Lane`：本地 API、Electron、IPC、事件流
4. `Diagnostics Lane`：错误码、诊断、safe-mode、恢复支持

说明：

1. `domain-brain lane` 不在本计划内，转由 Backend 计划承接
2. 任何 `code / browser / verifier / fault` 能力都只在 runtime lane 归属 `brain-v3`，不回填到 `domain-brain lane`

## 4. 开机顺序总览

推荐执行顺序：

```text
P0-A 基座对齐
  -> P0-B 存储落库
  -> P0-C 运行时主链路
  -> P0-D 集成接通
  -> P1-A 事件与聚合
  -> P1-B 恢复与诊断
  -> P1-C 性能与一致性补齐
  -> P2-A 稳定性增强
```

其中：

- `P0-A` 必须串行完成
- `P0-B` 与 `P0-C` 局部可并行
- `P0-D` 在 `P0-B` 与 `P0-C` 出首版产物后接入
- `P1` 开始后可多 agent 并行推进

## 5. 计划清单

### 5.1 P0-A 基座对齐

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| RT-001 | 固化 GoFrame v2 本地核心服务骨架与模块装配顺序 | P0 | - | no | `EasyMVP-V3-Go本地核心服务架构设计.md#2. 模块总览`；`EasyMVP-V3-Go本地核心服务架构设计.md#5. 服务装配顺序`；`EasyMVP-V3-代码目录结构与模块归属建议.md` | `apps/core` 的模块边界、装配入口、包归属与文档一致；首版 `app/api/orchestrator/runtime/storage/worker/diagnostics` 目录可用；装配顺序可启动且可健康检查 | completed |
| RT-002 | 固化 Electron Main / Renderer / Go Core 三进程边界 | P0 | RT-001 | no | `EasyMVP-V3-Electron进程模型与IPC边界设计.md#2. 进程分工`；`EasyMVP-V3-Electron进程模型与IPC边界设计.md#3. IPC 的正式定位`；`EasyMVP-V3-本地API与IPC适配设计.md#3. 逻辑边界与物理边界` | 桌面壳、渲染层、Go 核心之间的边界被落实为代码目录与调用约束；主业务 API 不经 IPC 直传；本地 API 与桌面桥接职责不混淆 | completed |
| RT-003 | 固化本地配置加载、启动参数与 safe-mode 入口 | P0 | RT-001 | yes | `EasyMVP-V3-本地配置与启动参数设计.md#3. 推荐配置项`；`EasyMVP-V3-本地配置与启动参数设计.md#5. 启动参数建议`；`EasyMVP-V3-本地配置与启动参数设计.md#6. safe-mode` | 启动时能稳定加载默认配置、工作目录、端口、诊断开关；safe-mode 能独立进入；配置缺失或错误时有标准化诊断输出 | in_progress |

### 5.2 P0-B 存储落库

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| ST-001 | 建立 SQLite 初始化、连接策略与 PRAGMA 基线 | P0 | RT-001 | yes | `EasyMVP-V3-SQLite初始化与Migration设计.md#4. 数据库初始化流程`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#2. PRAGMA 基线` | SQLite 文件创建、连接、PRAGMA 设置、启动自检全部落地；失败时返回标准错误域；可在空目录首次启动 | completed |
| ST-002 | 建立 schema_migrations 机制与 migration 执行器 | P0 | ST-001 | no | `EasyMVP-V3-SQLite初始化与Migration设计.md#5. Migration 机制`；`EasyMVP-V3-SQLite初始化与Migration设计.md#6. schema_migrations 表设计`；`EasyMVP-V3-独立Migration文件正文终稿.md` | migration 执行顺序、版本记录、失败中断、升级前备份、首次初始化全部可跑；已有真实 `.sql` 文件可执行 | completed |
| ST-003 | 落首批核心表与索引 | P0 | ST-002 | no | `EasyMVP-V3-数据库Schema总设计.md#2. 表族总览`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#3. 系统表`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#4. 项目表`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#5. 计划表`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#6. 任务与运行时表`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#7. 验收表`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#8. 证据、回放、审计、快照表`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#9. 索引终稿` | 首批 migration 在新库上能完整建表；关键索引存在；基础查询无缺表缺索引问题；表结构与文档主键、外键、JSON 字段约束一致 | completed |
| ST-004 | 落本地数据根目录与项目工作区初始化器 | P0 | RT-003 | yes | `EasyMVP-V3-本地目录与项目工作区规范.md#3. 顶层数据根目录`；`EasyMVP-V3-本地目录与项目工作区规范.md#4. 顶层目录结构`；`EasyMVP-V3-本地目录与项目工作区规范.md#11. 创建项目时的目录初始化` | 顶层数据目录、项目级目录、runs/evidence/replay/diagnostics 初始化可重复执行；路径合法性与可写性校验可用 | completed |
| ST-005 | 落事务边界、幂等键与补偿策略框架 | P0 | ST-003 | yes | `EasyMVP-V3-事务边界与一致性设计.md#2. 必须事务化的动作`；`EasyMVP-V3-事务边界与一致性设计.md#4. 典型事务边界`；`EasyMVP-V3-事务边界与一致性设计.md#5. 补偿原则`；`EasyMVP-V3-事务边界与一致性设计.md#6. 幂等性要求` | 创建项目、编译计划、启动运行、验收裁决四类主动作具备事务边界定义；重复执行安全；失败后可补偿或明确残留状态 | pending |

### 5.3 P0-C 运行时主链路

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| RN-001 | 落 Runtime Adapter 与 `brain-v3` Run 生命周期映射骨架 | P0 | RT-001 | yes | `EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md#4. Run 生命周期映射`；`EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md#5. 建议接口流程`；`EasyMVP-V3-Go本地核心服务架构设计.md#3.6 runtime` | Go 核心服务可创建、查询、取消、恢复 `brain` run；本地状态映射与远端状态映射有统一结构；失败可回写诊断 | done |
| RN-002 | 落 brain run 绑定、事件、checkpoint 存储链路 | P0 | ST-003,RN-001 | no | `EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md#6. 建议本地表`；`EasyMVP-V3-数据库Schema总设计.md#2.4 runtime`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#6.4 brain_run_bindings`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#6.5 run_checkpoints`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#6.6 run_event_index` | run 与 task / project 绑定关系落库；增量事件与 checkpoint 可写入、去重、查询；页面可稳定读取当前运行态；必要时可把 `tools/call` 结果归一化为稳定 runtime 摘要而不透传原始 payload | pending |
| RN-003 | 落后台 Worker 管理器、调度与失败回传 | P0 | RT-001,RN-001 | yes | `EasyMVP-V3-后台Worker与任务调度设计.md#2. 首批 worker`；`EasyMVP-V3-后台Worker与任务调度设计.md#3. Go 调度模型`；`EasyMVP-V3-后台Worker与任务调度设计.md#6. 失败处理` | 首批 worker 进程内可调度；有并发上限、重试、失败记录、停机回收；失败会产出审计与诊断记录；`unsupported / denied` 会以显式运行时状态回传，不伪装成成功 | completed |
| RN-004 | 落 replay / log / artifact 文件索引绑定与缺失诊断 | P0 | ST-004,RN-002 | yes | `EasyMVP-V3-replay与log artifact存储规范.md`；`EasyMVP-V3-本地目录与项目工作区规范.md#7. Run 目录规范`；`EasyMVP-V3-Evidence文件命名与引用规范.md#10. 文件生成与索引写入顺序` | run 相关日志、checkpoint、artifact、replay 的索引与物理文件一致；文件缺失可诊断；不重复造原始 replay 存储 | pending |

### 5.4 P0-D 集成接通

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| IN-001 | 落本地 Go API 路由分组与健康检查链路 | P0 | RT-001,ST-001 | yes | `EasyMVP-V3-API路由分组与命令查询边界设计.md#3. 路由分组建议`；`EasyMVP-V3-GoFrame-Handler-DTO逐项终稿.md`；`EasyMVP-V3-核心API-DTO与TypeScript类型终稿.md` | `/api/v3/system/healthz` 与首批 query/command 路由可用；路由分组、DTO、错误格式与文档一致 | completed |
| IN-002 | 落 Electron 到 Go Core 的本地 API client 与启动握手 | P0 | RT-002,IN-001 | no | `EasyMVP-V3-Electron进程模型与IPC边界设计.md#2. 进程分工`；`EasyMVP-V3-本地API与IPC适配设计.md#4. 示例映射`；`EasyMVP-V3-单机版启动时序与进程内调用链设计.md` | Electron 能托管或探活 Go 本地服务；renderer 通过统一 client 调用本地 API；启动失败时能进入恢复模式 | in_progress |
| IN-003 | 落桌面原生桥接最小面：文件选择、目录探测、shell 能力 | P0 | RT-002 | yes | `EasyMVP-V3-Electron进程模型与IPC边界设计.md#7. 首版建议的 preload 暴露面`；`EasyMVP-V3-本地API与IPC适配设计.md#7. 首版 client 建议` | preload 只暴露必要原生能力；不承载主业务；桥接 API 有类型定义与错误包装 | completed |
| IN-004 | 接通运行时事件流到工作台聚合入口 | P0 | RN-002,IN-001 | yes | `EasyMVP-V3-实时事件流推送机制设计.md`；`EasyMVP-V3工作台视图模型与聚合接口设计.md`；`EasyMVP-V3-Workspace首页聚合接口Schema设计.md` | runtime 事件可被聚合层消费；工作台和项目页能看到基础实时状态；中断重连后可恢复订阅 | completed |

### 5.5 P1-A 事件与聚合

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| AG-001 | 落 Workspace 首页聚合查询与缓存快照 | P1 | ST-003,IN-001 | yes | `EasyMVP-V3工作台视图模型与聚合接口设计.md`；`EasyMVP-V3-Workspace首页聚合接口Schema设计.md`；`EasyMVP-V3-数据库Schema总设计.md#2.8 view_cache` | Workspace 首页所需聚合接口可返回项目进度、待处理、验收状态、最近活动；快照策略可读且可刷新 | completed |
| AG-002 | 落 Project Workspace 实时状态聚合 | P1 | RN-002,IN-004 | yes | `EasyMVP-V3实时工作台页面设计.md`；`EasyMVP-V3-Workspace详细页面设计.md`；`EasyMVP-V3-Project-Workspace线框图设计.md` | 单项目工作台能读到阶段条、活动流、Action Inbox、验收摘要；聚合对象不直接暴露底表结构 | completed |
| AG-003 | 落审计 / 回放 / Evidence 查询聚合面 | P1 | ST-003,RN-004 | yes | `EasyMVP-V3-审计查询接口设计.md`；`EasyMVP-V3-Replay查询接口设计.md`；`EasyMVP-V3-Evidence卡片查询接口设计.md` | 审计、回放、证据页都有页面语义 API；查询不要求前端拼底表；分页、过滤、详情读取可用 | pending |

### 5.6 P1-B 恢复与诊断

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| DG-001 | 落错误码、错误域与标准诊断记录模型 | P1 | RT-003,ST-001 | yes | `EasyMVP-V3-错误码与诊断分级设计.md#3. 错误域`；`EasyMVP-V3-错误码与诊断分级设计.md#4. 错误码建议`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#3.3 diagnostic_records` | 错误码、错误域、诊断上下文、恢复建议具备统一结构； API 与 worker 错误都能归一化写入诊断记录 | pending |
| DG-002 | 落启动失败、migration 失败、目录异常的恢复模式链路 | P1 | ST-002,ST-004,DG-001 | no | `EasyMVP-V3-恢复模式与诊断模式页面设计.md`；`EasyMVP-V3-本地配置与启动参数设计.md#6. safe-mode` | 启动失败可进入恢复模式；能展示 migration 失败、目录不可写、核心服务不可用等问题；支持重试和诊断导出 | in_progress |
| DG-003 | 落孤儿文件、缺失索引、artifact 缺失检测任务 | P1 | RN-004,DG-001 | yes | `EasyMVP-V3-本地目录与项目工作区规范.md#14. 恢复与校验`；`EasyMVP-V3-Evidence文件命名与引用规范.md`；`EasyMVP-V3-replay与log artifact存储规范.md` | 可以扫描并识别 orphan files、missing artifact、stale index；结果进入诊断与审计；不自动破坏业务数据 | pending |

### 5.7 P1-C 性能与一致性补齐

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| PF-001 | 落关键查询索引校验与慢查询基线 | P1 | ST-003,AG-001 | yes | `EasyMVP-V3-数据库索引与查询优化设计.md`；`EasyMVP-V3-完整SQLite建表与索引SQL终稿.md#9. 索引终稿` | 首页、项目页、审计、回放、证据等关键查询走到既定索引；形成可回归的 explain / benchmark 基线 | pending |
| PF-002 | 落幂等重试与事件去重 | P1 | ST-005,RN-002 | yes | `EasyMVP-V3-事务边界与一致性设计.md#6. 幂等性要求`；`EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md#8. 错误处理原则` | 重试不会重复建项目、重复建 run、重复写事件；乱序或重复事件有去重规则；恢复后状态稳定 | pending |
| PF-003 | 落工作台快照更新策略与读写隔离约束 | P1 | AG-001,AG-002 | yes | `EasyMVP-V3工作台视图模型与聚合接口设计.md`；`EasyMVP-V3-数据库Schema总设计.md#2.8 view_cache`；`EasyMVP-V3-事务边界与一致性设计.md#7. 一致性优先级` | 聚合快照刷新有策略；页面读接口不过度阻塞主事务；用户看到的数据在可接受延迟范围内一致 | pending |

### 5.8 P2-A 稳定性增强

| plan_id | name | priority | depends_on | parallelizable | doc_refs | definition_of_done | status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| SX-001 | 落备份、恢复与升级前快照工具链 | P2 | ST-002,ST-004,DG-002 | yes | `EasyMVP-V3-SQLite初始化与Migration设计.md#8. 升级策略`；`EasyMVP-V3-本地目录与项目工作区规范.md#13. 备份与导出边界` | 升级前快照、最小备份单元、恢复校验流程可执行；恢复后数据库与目录一致性有校验步骤 | pending |
| SX-002 | 落运行时诊断导出与问题打包 | P2 | DG-001,RN-003 | yes | `EasyMVP-V3-错误码与诊断分级设计.md#6. 审计与诊断要求`；`EasyMVP-V3-恢复模式与诊断模式页面设计.md` | 一键导出诊断包可包含配置摘要、错误码、日志索引、migration 状态、目录检查结果；敏感信息按规则脱敏 | in_progress |
| SX-003 | 落集成自检清单与开机验收脚本 | P2 | IN-004,DG-002,PF-001 | yes | `EasyMVP-V3-技术栈与选型基线.md`；`EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计.md`；`EasyMVP-V3-Go本地核心服务架构设计.md#8. 可直接开工的第一批 package` | 提供本地环境自检、数据库状态检查、Go API 健康检查、brain-v3 连通性检查、工作目录检查的一次性验收脚本 | pending |

## 6. 推荐多 Agent 分工

### 6.1 Agent Team 结构

建议至少拆成 4 个 agent：

1. `agent-runtime-core`
2. `agent-storage-sqlite`
3. `agent-integration-desktop`
4. `agent-diagnostics-recovery`

补充职责提醒：

1. `agent-runtime-core` 负责吸收 `brain-v3` 协议与状态变化
2. 任何涉及领域合同的修改应转交 backend/domain-brain 计划，不在 runtime 计划内混写

不再新增：

1. `agent-runtime-kernel-clone`
2. `agent-generic-eventbus`
3. `agent-generic-replay-store`

### 6.2 计划归属建议

| agent | 主负责计划 |
| --- | --- |
| `agent-runtime-core` | `RT-001` `RN-001` `RN-002` `RN-003` `RN-004` `PF-002` |
| `agent-runtime-core` | `RT-001` `RN-001` `RN-002` `RN-003` `RN-004` `PF-002` |
| `agent-storage-sqlite` | `ST-001` `ST-002` `ST-003` `ST-004` `ST-005` `PF-001` `SX-001` |
| `agent-integration-desktop` | `RT-002` `IN-001` `IN-002` `IN-003` `IN-004` `AG-001` `AG-002` |
| `agent-diagnostics-recovery` | `RT-003` `DG-001` `DG-002` `DG-003` `AG-003` `SX-002` `SX-003` |

## 7. 更新规则

每次完成一个计划项，至少要同步更新：

1. 本文档中的 `status`
2. 如有新增实现细节，补到对应上游文档
3. 如有偏离设计，必须在对应 `doc_refs` 指向的文档中修订

推荐更新格式：

```text
plan_id: ST-002
status: completed
completed_at: 2026-04-19T18:30:00+08:00
notes: 首批 migration 执行器已落地，支持版本记录、失败中断和升级前备份。
```

## 8. 中断恢复规则

如果中断后需要继续：

1. 先找所有 `in_progress` 和 `blocked`
2. 再看这些计划的 `depends_on` 是否已满足
3. 优先恢复 `P0` 中断项
4. 并行项恢复时，避免多个 agent 同时修改相同模块

恢复顺序建议：

1. `RT-*`
2. `ST-*`
3. `RN-*`
4. `IN-*`
5. `AG-*`
6. `DG-*`
7. `PF-*`
8. `SX-*`

## 9. 一句话结论

这份计划把 V3 的运行时、存储、集成实施切成了可连续推进的任务单元。

已完成说明：

```text
plan_id: ST-004
status: completed
completed_at: 2026-04-19T00:00:00+08:00
notes: 顶层 data root 已补 projects/；创建项目时已把受管目录收口到 dataRoot/projects/{project_id}/ 下，并一次性初始化 meta/runs/evidence/replay/exports/cache/diagnostics；保留 workspace_root 作为业务工作目录；目录初始化可重复执行并带路径边界与可写性校验。
```

真正的开工顺序应是：

先固化本地核心服务与三进程边界，再落 SQLite 与目录，再接通 `brain-v3` 与本地 API，最后补齐聚合、诊断、恢复和性能。
