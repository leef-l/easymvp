# EasyMVP V3 AgentTeam Frontend 计划

> 更新时间：2026-04-24
> 上游文档：[EasyMVP-V3工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
> 关联文档：[EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md)
> 关联文档：[EasyMVP-V3-核心API-DTO与TypeScript类型终稿](./EasyMVP-V3-核心API-DTO与TypeScript类型终稿.md)
> 目标：为 V3 前端实施提供可执行、可恢复、可并行的开机计划，确保中断后可按计划 ID 继续推进。
>
> 当前完成度以 [EasyMVP-V3-AgentTeam状态板](./EasyMVP-V3-AgentTeam状态板.md) 为准。前端剩余重点是 diagnostics / recovery / acceptance / repair 更细粒度联动、replay/evidence 聚合展示与 packaged smoke 证明链。

## 1. 计划使用规则

这份文档是前端 AgentTeam 的正式执行入口。

每个计划项都必须包含：

1. `计划ID`
2. `名称`
3. `优先级`
4. `依赖`
5. `是否可并行`
6. `对应文档位置`
7. `完成定义`
8. `状态`

状态字段只允许使用：

1. `pending`
2. `in_progress`
3. `blocked`
4. `done`

默认状态一律为 `pending`。

## 2. AgentTeam 分工建议

前端实施默认遵守两条边界：

1. 页面与 hooks 只消费 EasyMVP 聚合后的快照、详情和事件 DTO
2. 不直接渲染 `brain-v3` 原始工具结果、原始 `content[]` 或未归一化 payload

前端实施建议至少拆成 4 类 agent，并按模块并行推进：

1. `frontend-arch-agent`
   负责应用壳、路由、状态层、API client、事件流基础设施。
2. `workspace-agent`
   负责 `Workspace Home`、`Project Workspace`、创建项目流。
3. `plan-acceptance-agent`
   负责 `Plan`、`Acceptance`、Evidence、Release Gate。
4. `support-ui-agent`
   负责 `Replay`、审计、恢复模式、通用 loading/error/empty/recovery、设计统一收口。

## 3. 执行顺序总览

优先级分为：

1. `P0`：不开工会阻塞其它前端计划
2. `P1`：核心主流程页面
3. `P2`：辅助页面与支撑能力
4. `P3`：统一收口、联调和 polish

推荐推进节奏：

1. 先完成 `P0`
2. 再按页面簇并行推进 `P1`
3. 然后补齐 `P2`
4. 最后统一做 `P3`


## 4. 前端开机计划

### FE-001

- 计划ID：`FE-001`
- 名称：前端应用壳与目录骨架初始化
- 优先级：`P0`
- 依赖：无
- 是否可并行：否
- 对应文档位置：
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md) / `## 2. 推荐目录终稿`
  - [EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md) / `## 4. 顶层导航规范`
- 完成定义：
  - 建立 `apps/desktop` 前端目录主骨架
  - 落好 `app / modules / shared` 分层
  - 建立顶层导航与基础路由占位页面
  - 页面入口至少包含 `Workspace / Plan / Execution / Acceptance / Settings`
- 状态：`done`

### FE-002

- 计划ID：`FE-002`
- 名称：共享类型、API client 与错误封装初始化
- 优先级：`P0`
- 依赖：`FE-001`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-核心API-DTO与TypeScript类型终稿](./EasyMVP-V3-核心API-DTO与TypeScript类型终稿.md) / `## 2. 通用类型`
  - [EasyMVP-V3-本地API-DTO与错误返回设计](./EasyMVP-V3-本地API-DTO与错误返回设计.md)
  - [EasyMVP-V3-页面Loading-Empty-Error-Recovery状态终稿](./EasyMVP-V3-页面Loading-Empty-Error-Recovery状态终稿.md)
- 完成定义：
  - 建立 `shared/contracts` 类型目录
  - 建立 Query / Command / Error envelope 类型
  - 建立统一 API client、错误映射、请求元信息处理
  - 建立页面级加载、错误、恢复的通用状态协议
- 状态：`done`

### FE-003

- 计划ID：`FE-003`
- 名称：实时事件流与页面刷新机制初始化
- 优先级：`P0`
- 依赖：`FE-001`, `FE-002`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-实时事件流推送机制设计.md)
  - [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md) / `## 7. Hooks 终稿建议`
- 完成定义：
  - 建立事件流订阅入口
  - 建立 Workspace 和 Project 级别的刷新策略
  - 落地 `useWorkspaceEvents`、`useProjectEvents` 的基础实现
  - 能支撑项目卡片、活动流、验收覆盖等实时刷新
  - 显式处理 `unsupported / denied` 对应的 warning / blocker / attention 语义
- 状态：`done`

### FE-004

- 计划ID：`FE-004`
- 名称：设计系统与通用 UI 基础件初始化
- 优先级：`P0`
- 依赖：`FE-001`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md) / `## 7. 页面布局规范`
  - [EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md) / `## 8. 视觉规范`
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md)
- 完成定义：
  - 建立颜色、间距、圆角、阴影、排版变量
  - 建立通用卡片、状态徽标、按钮、空态、错误态组件
  - 建立页面标题区、区块标题区、抽屉容器等共享 UI
- 状态：`done`

### FE-005

- 计划ID：`FE-005`
- 名称：Workspace Home 页面主结构实施
- 优先级：`P1`
- 依赖：`FE-001`, `FE-002`, `FE-003`, `FE-004`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
  - [EasyMVP-V3-Workspace首页线框图设计](./EasyMVP-V3-Workspace首页线框图设计.md)
  - [EasyMVP-V3-Workspace首页聚合接口Schema设计](./EasyMVP-V3-Workspace首页聚合接口Schema设计.md)
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md) / `## 3. Workspace Home 组件树`
- 完成定义：
  - 完成 `WorkspaceHomePage`
  - 接入 `useWorkspaceHomeView`
  - 首屏能展示多项目总览、待处理事项、近期活动、发布准备度
  - 符合“首屏只回答当前状态/问题/动作”的规范
  - 不直接消费 `brain-v3` 原始运行时对象或原始工具 payload
- 状态：`done`

### FE-006

- 计划ID：`FE-006`
- 名称：Workspace Home 卡片组件簇实施
- 优先级：`P1`
- 依赖：`FE-005`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Workspace首页Stage-Overview组件规范](./EasyMVP-V3-Workspace首页Stage-Overview组件规范.md)
  - [EasyMVP-V3-Workspace首页Stage-Overview卡Props终稿](./EasyMVP-V3-Workspace首页Stage-Overview卡Props终稿.md)
  - [EasyMVP-V3-Workspace首页Need-Attention卡组件规范](./EasyMVP-V3-Workspace首页Need-Attention卡组件规范.md)
  - [EasyMVP-V3-Workspace首页Need-Attention卡Props终稿](./EasyMVP-V3-Workspace首页Need-Attention卡Props终稿.md)
  - [EasyMVP-V3-Workspace首页Recent-Activity组件规范](./EasyMVP-V3-Workspace首页Recent-Activity组件规范.md)
  - [EasyMVP-V3-Workspace首页Recent-Activity卡Props终稿](./EasyMVP-V3-Workspace首页Recent-Activity卡Props终稿.md)
  - [EasyMVP-V3-Workspace首页Release-Readiness卡组件规范](./EasyMVP-V3-Workspace首页Release-Readiness卡组件规范.md)
  - [EasyMVP-V3-Workspace首页Release-Readiness卡Props终稿](./EasyMVP-V3-Workspace首页Release-Readiness卡Props终稿.md)
  - [EasyMVP-V3-Workspace首页多项目卡片组件规范](./EasyMVP-V3-Workspace首页多项目卡片组件规范.md)
  - [EasyMVP-V3-Workspace首页Project卡Props终稿](./EasyMVP-V3-Workspace首页Project卡Props终稿.md)
- 完成定义：
  - `StageOverviewCard / NeedAttentionCard / RecentActivityCard / ReleaseReadinessCard / ProjectCard` 全部落地
  - props、交互事件、loading/empty/error 状态全部对齐文档
  - 项目卡片可跳转项目页
- 状态：`done`

### FE-007

- 计划ID：`FE-007`
- 名称：创建项目流与初始化时间线实施
- 优先级：`P1`
- 依赖：`FE-005`, `FE-004`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
  - [EasyMVP-V3-创建项目弹层线框图设计](./EasyMVP-V3-创建项目弹层线框图设计.md)
  - [EasyMVP-V3-创建项目弹层交互状态图](./EasyMVP-V3-创建项目弹层交互状态图.md)
  - [EasyMVP-V3-路径选择与仓库检测组件规范](./EasyMVP-V3-路径选择与仓库检测组件规范.md)
  - [EasyMVP-V3-创建初始化态时间线组件规范](./EasyMVP-V3-创建初始化态时间线组件规范.md)
  - [EasyMVP-V3-创建失败恢复与回滚策略设计](./EasyMVP-V3-创建失败恢复与回滚策略设计.md)
- 完成定义：
  - 创建项目弹层可完整走通
  - 支持路径选择、仓库检测、初始化时间线展示
  - 支持创建失败后的恢复与回滚提示
  - 创建成功后能跳转首次进入引导态
  - 创建页只展示 `role_type / brain_kind` 解析上下文，不把执行面职责错误归给 `easymvp-brain`
- 状态：`done`

### FE-008

- 计划ID：`FE-008`
- 名称：Project Workspace 页面主结构实施
- 优先级：`P1`
- 依赖：`FE-001`, `FE-002`, `FE-003`, `FE-004`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
  - [EasyMVP-V3-Project-Workspace线框图设计](./EasyMVP-V3-Project-Workspace线框图设计.md)
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md) / `## 4. Project Workspace 组件树`
- 完成定义：
  - `ProjectWorkspacePage` 主布局完成
  - 接入 `useProjectWorkspaceView`
  - 页面能展示项目快照、阶段流、活动流、待处理区、验收覆盖区
  - 满足实时驾驶舱定位
- 状态：`done`

### FE-009

- 计划ID：`FE-009`
- 名称：Project Workspace 核心组件簇实施
- 优先级：`P1`
- 依赖：`FE-008`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Project-Workspace-Top-Status-Bar组件规范](./EasyMVP-V3-Project-Workspace-Top-Status-Bar组件规范.md)
  - [EasyMVP-V3-Project-Workspace-Top-Status-Bar-Props终稿](./EasyMVP-V3-Project-Workspace-Top-Status-Bar-Props终稿.md)
  - [EasyMVP-V3-Project-Workspace-Stage-Rail组件规范](./EasyMVP-V3-Project-Workspace-Stage-Rail组件规范.md)
  - [EasyMVP-V3-Project-Workspace-Stage-Rail-Props终稿](./EasyMVP-V3-Project-Workspace-Stage-Rail-Props终稿.md)
  - [EasyMVP-V3-Project-Workspace-Live-Activity组件规范](./EasyMVP-V3-Project-Workspace-Live-Activity组件规范.md)
  - [EasyMVP-V3-Project-Workspace-Live-Activity-Props终稿](./EasyMVP-V3-Project-Workspace-Live-Activity-Props终稿.md)
  - [EasyMVP-V3-Project-Workspace-Action-Inbox组件规范](./EasyMVP-V3-Project-Workspace-Action-Inbox组件规范.md)
  - [EasyMVP-V3-Project-Workspace-Action-Inbox-Props终稿](./EasyMVP-V3-Project-Workspace-Action-Inbox-Props终稿.md)
- 完成定义：
  - `TopStatusBar / StageRail / LiveActivityPanel / ActionInboxPanel` 全部落地
  - 组件事件与跳转对齐文档
  - 支持选阶段、打开操作项、查看实时活动
- 状态：`done`

### FE-010

- 计划ID：`FE-010`
- 名称：首次进入引导态与常规工作台切换实施
- 优先级：`P1`
- 依赖：`FE-007`, `FE-008`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-创建后首次进入Project-Workspace引导态设计](./EasyMVP-V3-创建后首次进入Project-Workspace引导态设计.md)
  - [EasyMVP-V3-首次进入工作台线框图设计](./EasyMVP-V3-首次进入工作台线框图设计.md)
  - [EasyMVP-V3-首次进入引导卡组件规范](./EasyMVP-V3-首次进入引导卡组件规范.md)
  - [EasyMVP-V3-首次进入引导卡-Props终稿](./EasyMVP-V3-首次进入引导卡-Props终稿.md)
  - [EasyMVP-V3-首次进入推荐动作生成规则](./EasyMVP-V3-首次进入推荐动作生成规则.md)
  - [EasyMVP-V3-首次进入与常规工作台切换规则](./EasyMVP-V3-首次进入与常规工作台切换规则.md)
- 完成定义：
  - 创建后首次进入引导态可展示
  - 推荐动作卡与引导卡交互完整
  - 满足引导态向常规工作台的切换规则
- 状态：`done`

### FE-011

- 计划ID：`FE-011`
- 名称：Plan 页面主结构与 diff 视图实施
- 优先级：`P1`
- 依赖：`FE-001`, `FE-002`, `FE-004`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
  - [EasyMVP-V3-Plan线框图设计](./EasyMVP-V3-Plan线框图设计.md)
  - [EasyMVP-V3-Plan-diff组件规范](./EasyMVP-V3-Plan-diff组件规范.md)
  - [EasyMVP-V3-Plan任务投影抽屉设计](./EasyMVP-V3-Plan任务投影抽屉设计.md)
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md) / `## 5. Plan 页面组件树`
- 完成定义：
  - `PlanPage` 完成
  - 可展示 `Draft / Review / Compiled / Diff / Task Projection`
  - 任务投影抽屉可打开
  - 页面能解释系统如何编译与修改计划
- 状态：`done`

### FE-012

- 计划ID：`FE-012`
- 名称：Acceptance 页面主结构实施
- 优先级：`P1`
- 依赖：`FE-001`, `FE-002`, `FE-004`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
  - [EasyMVP-V3-Acceptance线框图设计](./EasyMVP-V3-Acceptance线框图设计.md)
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md) / `## 6. Acceptance 页面组件树`
  - [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
- 完成定义：
  - `AcceptancePage` 主结构完成
  - 接入 `useAcceptanceView`
  - 页面能展示 coverage、issues、evidence、release gate
  - 页面定位保持“最终交付裁决页”而不是 issue 列表
- 状态：`done`

### FE-013

- 计划ID：`FE-013`
- 名称：Acceptance 组件簇与证据展示实施
- 优先级：`P1`
- 依赖：`FE-012`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Coverage-Matrix组件规范](./EasyMVP-V3-Coverage-Matrix组件规范.md)
  - [EasyMVP-V3-Coverage-Matrix-Props终稿](./EasyMVP-V3-Coverage-Matrix-Props终稿.md)
  - [EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence卡片组件规范.md)
  - [EasyMVP-V3-Evidence卡片-Props终稿](./EasyMVP-V3-Evidence卡片-Props终稿.md)
  - [EasyMVP-V3-Evidence详情抽屉设计](./EasyMVP-V3-Evidence详情抽屉设计.md)
  - [EasyMVP-V3-Evidence Preview交互设计](./EasyMVP-V3-Evidence Preview交互设计.md)
  - [EasyMVP-V3-Release Gate抽屉设计](./EasyMVP-V3-Release Gate抽屉设计.md)
- 完成定义：
  - `CoverageMatrix / EvidenceCard / EvidenceDrawer / ReleaseGateDrawer` 全部落地
  - Evidence 预览与详情交互完整
  - 覆盖矩阵支持选中与联动
- 状态：`done`

### FE-014

- 计划ID：`FE-014`
- 名称：Execution 占位页与运行态视图收口
- 优先级：`P2`
- 依赖：`FE-003`, `FE-008`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md) / `## 10. Execution 页面设计`
  - [EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md) / `### 5.3 Execution`
- 完成定义：
  - 建立 `ExecutionPage` 基础页
  - 统一承接运行态列表、活跃 run、失败 run、人工介入入口
  - 不与 Project Workspace 的实时区职责冲突
- 状态：`done`

### FE-015

- 计划ID：`FE-015`
- 名称：Replay、审计与诊断辅助页面实施
- 优先级：`P2`
- 依赖：`FE-002`, `FE-004`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer%E9%A1%B5%E9%9D%A2%E8%AE%BE%E8%AE%A1.md)
  - [EasyMVP-V3-Replay-Drawer线框图设计](./EasyMVP-V3-Replay-Drawer线框图设计.md)
  - [EasyMVP-V3-审计列表页面设计](./EasyMVP-V3-审计列表页面设计.md)
  - [EasyMVP-V3-审计列表线框图设计](./EasyMVP-V3-审计列表线框图设计.md)
  - [EasyMVP-V3-审计过滤器设计](./EasyMVP-V3-审计过滤器设计.md)
  - [EasyMVP-V3-恢复模式与诊断模式页面设计](./EasyMVP-V3-恢复模式与诊断模式页面设计.md)
- 完成定义：
  - Replay 抽屉可用
  - 审计列表、过滤器、诊断入口可用
  - 恢复模式页面与诊断模式页面具备基础交互
- 状态：`done`

### FE-016

- 计划ID：`FE-016`
- 名称：导航、页面跳转与深链规则实施
- 优先级：`P2`
- 依赖：`FE-001`, `FE-005`, `FE-008`, `FE-011`, `FE-012`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-单机版导航与页面跳转规则](./EasyMVP-V3-单机版导航与页面跳转规则.md)
  - [EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md) / `## 4. 顶层导航规范`
- 完成定义：
  - 顶层导航、子跳转、抽屉跳转、深链打开全部可用
  - Workspace、Plan、Execution、Acceptance、Settings 之间跳转一致
  - 页面间推荐动作跳转与返回路径清晰
- 状态：`done`

### FE-017

- 计划ID：`FE-017`
- 名称：页面状态矩阵统一收口
- 优先级：`P2`
- 依赖：`FE-005`, `FE-008`, `FE-011`, `FE-012`, `FE-015`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-页面Loading-Empty-Error-Recovery状态终稿](./EasyMVP-V3-页面Loading-Empty-Error-Recovery状态终稿.md)
- 完成定义：
  - Workspace Home、Project Workspace、Plan、Acceptance、Replay、Audit 的 loading/empty/error/recovery 状态全部收口
  - 状态文案、恢复动作、重试行为一致
  - 页面不会出现“只有报错字符串没有恢复动作”的情况
- 状态：`done`

### FE-018

- 计划ID：`FE-018`
- 名称：前端联调准备与 Mock/契约校验收口
- 优先级：`P3`
- 依赖：`FE-002`, `FE-003`, `FE-005`, `FE-008`, `FE-011`, `FE-012`, `FE-016`, `FE-017`
- 是否可并行：否
- 对应文档位置：
  - [EasyMVP-V3-核心API-DTO与TypeScript类型终稿](./EasyMVP-V3-核心API-DTO与TypeScript类型终稿.md)
  - [EasyMVP-V3-GoFrame-Handler-DTO逐项终稿](./EasyMVP-V3-GoFrame-Handler-DTO%E9%80%90%E9%A1%B9%E7%BB%88%E7%A8%BF.md)
  - [EasyMVP-V3-本地API与IPC适配设计](./EasyMVP-V3-%E6%9C%AC%E5%9C%B0API%E4%B8%8EIPC%E9%80%82%E9%85%8D%E8%AE%BE%E8%AE%A1.md)
- 完成定义：
  - 前端所有已实现页面都切到统一契约层
  - Mock 数据与真实 DTO 字段保持一致
  - 为后续接 GoFrame 接口联调准备稳定 client 层
- 状态：`done`

### FE-019

- 计划ID：`FE-019`
- 名称：视觉一致性与页面可读性收口
- 优先级：`P3`
- 依赖：`FE-006`, `FE-009`, `FE-010`, `FE-011`, `FE-013`, `FE-015`, `FE-017`
- 是否可并行：是
- 对应文档位置：
  - [EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
  - [EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md)
- 完成定义：
  - 全页面视觉语言统一
  - 卡片层级、状态色、按钮层级、空态文案统一
  - 页面达到“简单、大气、明了、一看就会用”的设计目标
- 状态：`done`

### FE-020

- 计划ID：`FE-020`
- 名称：前端实施完成验收与计划状态更新机制落地
- 优先级：`P3`
- 依赖：`FE-018`, `FE-019`
- 是否可并行：否
- 对应文档位置：
  - [EasyMVP-V3-AgentTeam-Frontend计划](./EasyMVP-V3-AgentTeam-Frontend计划.md)
- 完成定义：
  - 所有计划项状态被持续维护
  - 已完成项更新为 `done`
  - 阻塞项更新为 `blocked` 并补充阻塞原因
  - 该文档成为前端中断恢复后的唯一计划入口
- 状态：`done`

## 5. 并行执行建议

推荐并行分组如下：

1. 基础组
   - `FE-001`
   - `FE-002`
   - `FE-003`
   - `FE-004`
2. Workspace 组
   - `FE-005`
   - `FE-006`
   - `FE-007`
   - `FE-008`
   - `FE-009`
   - `FE-010`
3. Plan / Acceptance 组
   - `FE-011`
   - `FE-012`
   - `FE-013`
4. 支撑与收口组
   - `FE-014`
   - `FE-015`
   - `FE-016`
   - `FE-017`
   - `FE-018`
   - `FE-019`
   - `FE-020`

## 6. 中断恢复规则

中断后恢复时，必须先做这 4 步：

1. 打开本计划文档
2. 查找所有 `in_progress` 和 `blocked` 项
3. 按依赖关系确定先恢复哪一项
4. 完成后立即更新对应状态

不允许口头推进而不更新状态。

## 7. 一句话结论

这份文档的作用不是解释 V3 前端做什么，而是把 V3 前端如何开工、如何并行、如何在中断后继续，固定成可执行计划。
