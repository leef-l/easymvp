# EasyMVP研发执行版

> 更新日期：2026-04-13
>
> 用途：用于研发排期、任务分配、阶段验收与周会跟踪。

## 1. 文档目标

本文件不是管理汇报材料，而是研发执行材料。

目标只有三个：

1. 把 90 天路线图拆成研发可开工的任务
2. 明确每个阶段的依赖、责任、验收标准
3. 确保团队优先解决主链稳定性和交付闭环，而不是继续分散投入

建议配套阅读：

- [EasyMVP全面分析与优化路线图](EasyMVP全面分析与优化路线图.md)
- [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md)
- [EasyMVP架构设计文档](EasyMVP架构设计文档.md)
- [GitWorktree任务级环境隔离设计文档](GitWorktree任务级环境隔离设计文档.md)

## 2. 当前执行原则

未来 90 天所有研发工作，按以下原则执行：

1. 主链优先于外围能力

- 任何新功能，如果不能提升 `create-project -> complete` 主链成功率，优先级都要靠后。

2. 交付闭环优先于执行器扩张

- 暂不把“再接更多模型、更多 CLI”作为一线目标。
- 先把 `diff / patch / PR / CI / accept` 打透。

3. 治理优先于自治增强

- 在没有审计、轨迹、权限和评测前，不继续扩大自动放权范围。

4. 低耦合优先于局部修补

- 新增能力默认模块化落地，不允许继续把复杂流程堆进单一控制器。

5. 回归优先于上线节奏

- 每个阶段都必须有可重复回归样例。

### 2.1 当前落地进度（2026-04-09）

截至当前代码状态，以下能力已完成首轮落地：

- `A1`：`workflow` 控制器已按运行时、审核、执行、验收、自治、时间线拆分
- `A2`：provider 已支持协议级路由，CLI / API 解析分离
- `A3`：workspace 已显式支持 `patch / pr / manual` 交付路径、风险分级和回写策略
- `A4`：已补 execute / accept / rework / delivery review 关键路径自动回归
- `A5`：已补样例清单一键校验入口，`ready` 场景结构和口径可自动检查
- `B1`：任务结果已带出 `delivery_mode / sync_status / patch_ref / delivery_ref / delivery_title`
- `B2`：验收证据已纳入 patch、PR 草稿、CI 文件与 CI 相关任务日志
- `B3`：审核问题可转方案修订，验收问题可转正式返工
- `B4`：已建立 low / medium / high 风险交付矩阵，系统检查可对中高风险直写漂移告警
- `C1`：项目级轨迹页已可展示阶段、事件、交付模式、回写状态
- `C2`：任务级执行回放已可展示日志、事件、问题、证据、交接与动作
- `C3`：项目级轨迹页已可展示交付闸门聚合与明细（待人工交付 / PR 草稿 / 待回写 / 高风险任务）
- `C4`：内部评测样例清单已通过 `test-workspaces/regression-manifest.json`、校验接口和控制台“评测样例”面板接入
- `F`：`web-antd` 已建立 GitHub Actions `1C/1G` 守卫基线，`verify-build / workflow-bundle / workflow entry bundles / 174-file full-source bundle shards` 已在 run [`24314526616`](https://github.com/leef-l/easymvp/actions/runs/24314526616) 通过

当前数据库结构变更只允许通过 migration 交付，最新 migration 为：

- `admin-go/manifest/sql/mysql/000006_workspace_delivery_reference.up.sql`
- `admin-go/manifest/sql/mysql/000006_workspace_delivery_reference.down.sql`

截至本轮收尾，之前收敛出的两类剩余项已完成首轮落地：

- `C3` 已补齐为“聚合 + 明细”视图，控制台可以直接查看待人工交付项
- 回归样例已补齐 `readme_refresh / multi_task_dependency / manual_takeover` 三个规格目录

后续若继续推进，重点将从“补主链缺口”转为“扩充样例规模、补 CI 串联、真实运行型回归能力和线上配置治理”。

补充说明：

- 受工具链内存峰值影响，`web-antd` 单次 full `vue-tsc/vite build` 不再作为 `1C/1G` 阶段验收口径
- 当前统一按 GitHub Actions 的全源码分片 guard 收口，该 guard 已覆盖 `vue-vben-admin/apps/web-antd/src/**/*.{vue,ts,tsx}`

## 3. 团队角色建议

按最小可执行配置，建议以职责而不是人名分工。

| 角色 | 建议职责 |
|------|----------|
| 后端 A | Workflow 主链、controller 拆分、accept / rework / orchestrator |
| 后端 B | provider 协议路由、workspace 回写策略、执行器运行时 |
| 前端 | workflow 控制台、验收页、轨迹页、系统检查页 |
| QA / DevOps | 回归脚本、CI 接入、样例项目、指标采集 |
| 产品 / PM | 范围控制、验收口径、风险任务分级、跨角色协调 |

## 4. 90 天阶段目标

### 4.1 阶段一：0-30 天

阶段目标：

```text
把主链做稳，把技术债收口到可控范围
```

阶段完成标志：

- 主链回归集可重复执行
- provider 协议路由不再污染 CLI / API
- worktree 回写策略明确且文档、代码、配置一致
- `workflow` 控制面完成首轮拆分
- 关键集成测试补齐

### 4.2 阶段二：31-60 天

阶段目标：

```text
把执行结果升级成可交付结果
```

阶段完成标志：

- 任务结果可区分 `patch / PR / manual`
- 验收页可展示 CI、证据、问题、决策
- review comment 可回流修复任务
- 中高风险任务默认进入 PR / 人审路径

### 4.3 阶段三：61-90 天

阶段目标：

```text
把系统升级成可治理的交付控制台
```

阶段完成标志：

- 项目级轨迹完整
- 任务级执行回放可用
- 决策动作和人工接管统一可查
- 风险闸门和审计能力上线

## 5. 工作流拆分

本轮研发工作拆成 6 条工作流并行推进。

### 5.1 工作流 A：主链稳定性

目标：

- 降低主链误失败
- 降低恢复链路不一致
- 建立可重复回归样例

涉及模块：

- `admin-go/app/mvp/internal/workflow/orchestrator`
- `admin-go/app/mvp/internal/workflow/stage`
- `admin-go/app/mvp/internal/workflow/scheduler`
- `admin-go/app/mvp/internal/controller/chat`

### 5.2 工作流 B：结果沉淀与回写

目标：

- 定义并实现 `patch / PR / manual` 三种结果形态
- 规范 worktree 结果如何沉淀到主工作区

涉及模块：

- `admin-go/app/mvp/internal/workspace`
- `admin-go/app/mvp/internal/workflow/executor`
- `admin-go/app/mvp/internal/workflow/stage/accept`

### 5.3 工作流 C：运行时协议与执行器路由

目标：

- provider 支持多协议端点
- CLI 与 API 路由彻底解耦

涉及模块：

- `admin-go/app/ai`
- `admin-go/utility/provider`
- `admin-go/app/mvp/internal/engine`
- `admin-go/app/mvp/internal/workflow/executor`

### 5.4 工作流 D：验收闭环

目标：

- 把验收从“评分”升级成“证据 + 决策 + 回修”中心
- 把验收标准从“单一通用规则”升级成“标准注册层 + 项目信号解析”
- 让不同项目类型走不同标准，但在 `review / verification / accept` 三处保持统一口径

当前标准化方向：

- 先按 `family_code` 做一级分层，例如 `coding / analysis / creative`
- 再按项目信号做二级分层，例如：
  - `coding.backend`
  - `coding.interactive_delivery`
  - 后续预留 `android.native_app`、`ios.native_app`、`game.client_runtime`
- 每个标准统一声明：
  - 必需检查类型
  - 是否必须存在通过的标准化验证
  - 是否必须在 review 阶段规划交互级验证任务
  - 是否必须在 accept 阶段拿到浏览器/真机/端到端证据

补充铁律：

- 所有新增数据读写链路禁止在控制器、阶段服务、验收服务里直接碰 DB
- 必须走 `service / repo interface / repo implementation` 的标准分层
- 项目级角色定义统一收口到 `workflow.role_definitions`
- 详细约束见 [EasyMVP工程铁律](./EasyMVP工程铁律.md)

涉及模块：

- `admin-go/app/mvp/internal/workflow/acceptance`
- `admin-go/app/mvp/internal/workflow/stage/accept`
- `vue-vben-admin/apps/web-antd/src/views/mvp/workflow`

### 5.5 工作流 E：控制台与可观测

目标：

- 项目级时间线
- 任务级轨迹
- 人工接管记录
- 决策动作查看

涉及模块：

- `admin-go/app/mvp/internal/controller/chat`
- `admin-go/app/mvp/internal/workflow/autonomy`
- `vue-vben-admin/apps/web-antd/src/views/mvp/workflow`

### 5.6 工作流 F：测试与评测

目标：

- 建立可回放样例
- 建立真实交付指标
- 为每次阶段验收提供数据依据

涉及模块：

- `admin-go/app/mvp/internal/**/*_test.go`
- `test-workspaces/`
- CI 脚本与回归脚本

当前基线（2026-04-13）：

- `.github/workflows/web-antd-guard.yml` 已覆盖 `verify-build`、`workflow-bundle`、`workflow entry bundles` 和 `174` 个 `web-antd` 源文件的 `6` 分片 verify bundle
- 当前通过记录：[`Web Antd Guard #9`](https://github.com/leef-l/easymvp/actions/runs/24314526616)
- `1C/1G` 下单次 full typecheck/build 不再作为本阶段阻塞项

## 6. 阶段一详细执行清单

阶段一是整个计划的起点，必须按依赖顺序推进。

### 6.1 任务 A1：拆分 `workflow` 超大控制器

现状：

- [workflow.go](../admin-go/app/mvp/internal/controller/chat/workflow.go) 已经承载过多职责。

目标：

- 把控制器按业务边界拆为独立文件或子模块。

建议拆分边界：

1. `workflow_review.go`
2. `workflow_execute.go`
3. `workflow_accept.go`
4. `workflow_autonomy.go`
5. `workflow_system_check.go`
6. `workflow_timeline.go`

负责人：

- 后端 A

依赖：

- 无前置依赖，可立即开始

DoD：

- 原接口路由不变
- 主链接口回归通过
- 单文件不再承载多条独立业务线

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `workflow.go` 已从超大控制器拆分为 `workflow_review / workflow_execution / workflow_accept / workflow_autonomy / workflow_system_check / workflow_timeline / workflow_trace / workflow_runtime / workflow_regression`
  - 原接口路径保持不变，共享校验与通用辅助逻辑仍留在主控制器文件
  - 相关控制器纯逻辑测试已覆盖验收问题回流、交付闸门、系统检查与回归样例摘要

### 6.2 任务 A2：统一 provider 协议路由

现状：

- 联调记录已证明 `provider_type / base_url` 无法表达多协议多端点。

目标：

- 支持协议级 endpoint
- 区分 CLI endpoint 与 API endpoint

输出物：

1. 数据结构方案
2. 兼容旧配置的读取逻辑
3. CLI / API 独立解析函数
4. 回归样例

负责人：

- 后端 B

依赖：

- 无前置依赖，可与 A1 并行

DoD：

- `chat`
- `review`
- `accept`
- `autonomy`
- `aider`
- `claude_code`
- `codex_cli`

至少完成一轮协议正确性回归。

当前状态（2026-04-09）：

- 已完成首轮落地：
  - provider 已支持协议级 endpoint 解析，CLI / API 路由不再共用单一 `base_url`
  - 兼容旧配置的读取逻辑已保留，避免线上存量配置直接失效
  - `utility/provider` 相关回归已纳入当前后端测试集

### 6.3 任务 A3：明确 worktree 回写策略

现状：

- 文档与实际行为对“是否回写主工作区”存在认知不一致。

目标：

- 把当前行为改成显式策略，而不是隐式实现。

必须回答的问题：

1. 默认是否自动回写
2. 低风险任务的回写条件是什么
3. 冲突、越界、失败时如何降级
4. 是否先落 patch 再落主工作区

建议策略：

1. 默认输出 `patch`
2. 低风险任务允许自动应用 patch
3. 中高风险任务生成 PR 或保留人工处理

负责人：

- 后端 B
- 产品 / PM 参与口径确认

依赖：

- 与 A2 并行

DoD：

- 文档、配置、代码行为一致
- 风险等级和回写方式可在项目级或任务级表达

当前状态（2026-04-09）：

- 已完成首轮落地：
  - workspace 已显式支持 `patch / pr / manual` 三种交付路径和 `low / medium / high` 风险分级
  - 默认交付矩阵、配置覆写、系统检查与回归测试已统一到同一套口径
  - 回写状态、交付引用和交付摘要已可在执行页、时间线页和验收链路中直接查看

### 6.4 任务 A4：补 execute / accept / rework / delivery review 关键路径自动回归

现状：

- 当前测试文件数量不足以覆盖主链状态复杂度。

目标：

- 补最小必要自动回归，优先覆盖状态机风险点和交付闸门判断。

优先覆盖场景：

1. execute 全任务完成进入 accept
2. accept manual_review 后人工放行收口
3. rework 后回流 execute / accept
4. delivery review 闸门聚合与明细判定
5. running 中人工改写任务

负责人：

- 后端 A
- QA / DevOps 配合样例数据

依赖：

- A1 完成后优先补

DoD：

- 阶段一要求的关键路径均有自动回归

当前状态（2026-04-09）：

- 已完成首轮落地：
  - execute 侧沿用 `executor / workspace / task projection` 既有自动回归
  - accept 侧已补 `decision reducer / manual_review downgrade / accept issue rework reason` 测试
  - rework 侧已补 `parseTaskPatch` 解析测试
  - delivery review 侧已补交付闸门原因与风险排序测试

### 6.5 任务 A5：建立阶段一回归样例集

目标：

- 把现有联调经验沉淀成回归样例和固定输入。

建议样例：

1. 最小 Go 后端服务项目
2. README 修订项目
3. 多任务依赖项目
4. 人工接管回归项目

负责人：

- QA / DevOps
- 后端 A / B 协助

依赖：

- 无前置依赖

DoD：

- 样例至少可一键校验，后续继续扩展为真实运行型回归
- 每轮版本升级后都能先执行样例清单与 ready 结构校验

当前样例清单：

- [test-workspaces/regression-manifest.json](../test-workspaces/regression-manifest.json)

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `regression-manifest.json` 已升级到 `version = 2`，当前 4 个场景均为 `ready`
  - 已补 `readme_refresh / multi_task_dependency / manual_takeover` 三个规格目录
  - 已新增 `go run ./app/mvp/regressioncheck` 与 `bash ./test-workspaces/validate.sh` 一键校验入口，可同时校验 manifest 与风险交付矩阵
  - guard 脚本需在隔离验证环境执行，避免在业务服务器直接运行
  - 回归样例接口与控制台面板已可直接查看 manifest 校验状态和 ready/planned 统计

## 7. 阶段二详细执行清单

### 7.1 任务 B1：输出三种结果形态

目标：

- 每个任务完成后，必须明确属于 `patch / PR / manual` 中哪一类。

负责人：

- 后端 B

DoD：

- 数据库、接口、前端状态可见

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `mvp_task_workspace` 已落库 `delivery_mode / delivery_status / sync_strategy / sync_status / risk_level / delivery_ref / delivery_title`
  - `execution-status / domain-tasks / project-trace / delivery-reviews` 已返回 `patch / pr / manual` 结果形态和回写状态
  - 执行控制台与时间线页已可直接展示交付结果、回写状态与交付引用

### 7.2 任务 B2：验收阶段接入 CI

目标：

- 把测试、构建、静态检查结果纳入验收证据。

负责人：

- 后端 A
- QA / DevOps

DoD：

- 验收页可展示 CI 结果和失败原因

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `EvidenceCollector` 已收集 `.easymvp/ci/latest.*`、GitHub Actions、GitLab/Jenkins/CircleCI 文件与 CI 相关任务日志
  - `/workflow/accept-evidence` 与验收页已可直接展示 CI 证据摘要、来源与引用
  - 已补 CI 证据收集纯逻辑测试，覆盖文件探测、JSON 摘要与日志识别

### 7.3 任务 B3：review comment 回流修复

目标：

- 把 review comment 或验收问题转成正式修复任务。

负责人：

- 后端 A
- 前端

DoD：

- 问题项可一键回流修复链路

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `/workflow/review-issue-replan` 与 `/workflow/accept-issue-rework` 已提供审核问题、验收问题的正式回流入口
  - 审核页与验收页已支持对单条问题发起方案修订或返工
  - 已补验收问题转返工原因拼装测试，确保回流文案与建议动作一致

### 7.4 任务 B4：中高风险任务 PR 化

目标：

- 对越界、跨模块、高影响范围任务，默认不直写主结果。

负责人：

- 后端 B
- 产品 / PM

DoD：

- 风险分级能直接决定结果路径

当前状态（2026-04-09）：

- 已完成首轮落地：
  - low / medium / high 风险已具备默认交付矩阵
  - `workspace delivery policy` 已支持配置覆写，并补齐自动回归
  - 系统检查已可直接提示 medium/high 是否错误退化为自动直写

## 8. 阶段三详细执行清单

### 8.1 任务 C1：项目级轨迹面板

目标：

- 统一展示从需求到交付的整条轨迹。

负责人：

- 前端
- 后端 A

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `/workflow/project-trace` 与时间线页已统一展示阶段、事件、交付模式、回写状态和待处理人工节点
  - 项目级统计已覆盖阶段数、事件数、任务数、返工轮次、待处理人工节点和待决策动作

### 8.2 任务 C2：任务级执行回放

目标：

- 能查看任务执行器、模型、输入、产出、回写、验收、返工全过程。

负责人：

- 后端 A
- 后端 B

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `/workflow/task-replay` 与执行页已可查看任务日志、事件、问题、证据、交接与决策动作
  - 任务回放已带出工作空间交付元数据、验收证据和返工上下文

### 8.3 任务 C3：策略闸门与审计

目标：

- 把高风险执行纳入审计和权限控制。

负责人：

- 后端 B
- QA / DevOps

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `/workflow/delivery-reviews` 已补齐高风险交付、PR 草稿、待回写任务的聚合与明细审计视图
  - 时间线页已可直接查看待人工交付、PR 草稿、高风险和待回写任务统计

### 8.4 任务 C4：内部评测面板

目标：

- 用真实项目回放衡量版本提升，而不是只看主观体验。

负责人：

- QA / DevOps
- 产品 / PM

当前状态（2026-04-09）：

- 已完成首轮落地：
  - `C4`：`/workflow/regression-scenarios` 与“评测样例”面板已接入 manifest 校验结果、ready/planned 统计和样例列表

## 9. 任务依赖图

建议按下面的顺序推进：

```text
A1 控制器拆分
-> A4 集成测试补齐

A2 协议路由
-> A3 回写策略
-> B1 结果形态
-> B4 风险分级 PR 化

A5 回归样例
-> A4 测试回归
-> B2 CI 验收
-> C4 内部评测
```

关键路径只有两条：

1. `控制器拆分 -> 集成测试 -> 主链稳定`
2. `协议路由 -> 回写策略 -> 结果形态 -> 验收闭环`

## 10. 每周执行节奏

建议固定如下节奏。

### 周一

- 确认本周范围
- 确认阻塞项
- 确认关键路径任务 owner

### 周三

- 检查阶段风险
- 决定是否缩范围
- 对主链指标做中期复盘

### 周五

- 跑阶段回归集
- 汇总失败样例
- 更新文档与下周任务

## 11. 验收口径

每个阶段必须满足以下四类验收。

### 11.1 代码验收

- 主链相关改动有对应测试或回归样例
- 不新增更大的单点控制器
- 新增配置具有兼容策略

### 11.2 运行验收

- 能在当前开发环境真实跑通
- 服务重启、人工接管、返工场景可验证

### 11.3 产品验收

- 前端能看到新增状态、证据或轨迹
- 用户路径没有因为功能增强而变得更难理解

### 11.4 文档验收

- 实现边界与文档一致
- 不再出现“文档说一种，代码做另一种”的情况

## 12. 阶段风险清单

### 高风险

1. 控制器拆分过程中出现接口回归
2. provider 路由改造影响现有线上配置
3. worktree 回写策略调整引入结果不一致

### 中风险

1. 测试样例不足导致阶段验收失真
2. 前后端节奏不一致，导致结果形态已落库但控制台不可见
3. CI 接入后验收规则增长过快，影响推进速度

### 低风险

1. 文档滞后
2. 指标采集不完整
3. 样例项目维护成本上升

## 13. 本周可立即开工项

如果按当前状态直接开工，建议第一周只做 5 件事：

1. 建 `workflow` 控制器拆分目录和文件边界
2. 设计 provider 协议级 endpoint 结构
3. 写 worktree 回写策略草案
4. 建最小回归样例集
5. 补一条 `accept manual_review -> accept-approve -> complete` 自动回归

这样做的原因是：

- 这 5 件事都在关键路径上
- 它们不会因为后续方案变化而大幅返工
- 做完后，阶段一的剩余任务会更容易分派和推进

## 14. 结论

研发执行层的核心共识应该只有一句：

```text
未来 90 天不是继续“堆能力”，
而是把 EasyMVP 做成真正可交付、可治理、可回归的系统。
```

因此，执行顺序必须坚定遵守：

```text
主链稳定
-> 回写策略
-> 结果形态
-> 验收闭环
-> 轨迹治理
-> 多代理控制台
```

任何不在这条序列上的需求，都不应该抢占当前关键路径资源。
