# EasyMVP V3 实时工作台页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3文档总纲](./EasyMVP-V3文档总纲.md)
> 关联文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
> 关联文档：[EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
> 关联文档：[EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
> 关联文档：[EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 EasyMVP V3 的产品形态正式定义为“单用户、本地优先、实时可视化项目工作台”，并细化主页面结构、实时数据模型和交互原则。

## 1. 设计结论

EasyMVP V3 不应该继续沿用后台系统思路。

V3 的正式产品形态应定义为：

> 一个面向单个用户的、本地优先的、可实时观察项目进展与待处理问题的 AI 项目工作台。

这意味着：

1. 不是多角色后台
2. 不是配置中心
3. 不是 CRUD 管理系统
4. 而是实时项目驾驶舱

本专题不是新的 V3 主线。

它的定位是：

1. 承接方案编译主线的可视化展示
2. 承接分类与角色主线的自动推导结果展示
3. 承接 `brain-v3` Run 生命周期的实时运行展示
4. 承接生产级验收主线的最终交付展示

## 2. 产品定位

V3 的核心目标不是“管理很多系统对象”，而是帮助单个用户持续回答 4 个问题：

1. 项目现在进行到哪里了
2. 系统现在正在做什么
3. 当前卡住了什么问题
4. 距离生产可交付还差什么

因此 V3 的 UI 设计必须以这 4 个问题为第一原则。

## 3. 上游依赖与输出对象

本专题依赖的上游对象不是新的产品概念，而是 V3 已有对象。

上游依赖至少包括：

1. `ProjectCategory`
2. `CategoryProfile`
3. `PlanDraft`
4. `PlanReviewResult`
5. `CompiledPlan`
6. `DomainTask`
7. `DeliveryContract`
8. `VerificationContract`
9. `ProductionAcceptanceProfile`
10. `AcceptanceRun`
11. `brain-v3` Run / Status / Logs / Replay

本专题主要定义的输出对象包括：

1. `ProjectSnapshot`
2. `StageProgress`
3. `LiveEvent`
4. `ActionInboxItem`
5. `AcceptanceCoverage`

这些对象属于前端聚合视图对象，不替代上游业务对象。

## 4. 为什么不能再做后台

如果继续走后台思路，最终会再次变成：

1. 多菜单
2. 多表单
3. 多配置页
4. 多状态字段堆叠

这种形态会直接破坏 V3 当前最重要的 4 个目标：

1. 轻量化
2. 可视化
3. 方便
4. 简洁 UI

所以 V3 应该明确放弃以下方向：

1. 多租户
2. 管理员后台主入口
3. 大量低频系统配置页
4. 表格式项目详情页
5. 面向治理视角的复杂菜单结构

## 5. 与 V3 主链路的关系

实时工作台不是脱离工作流单独存在的页面层。

它应该直接映射 V3 主链路：

```text
用户需求
  ↓
PlanDraft
  ↓
PlanReviewResult
  ↓
CompiledPlan
  ↓
DomainTask
  ↓
Executor Run / Delivery / Verification
  ↓
Production Acceptance
  ↓
Complete
```

对应到界面层：

1. `PlanDraft / PlanReviewResult / CompiledPlan` 进入 `Plan` 页
2. `DomainTask` 和 `brain-v3 Run` 进入 `Workspace` 的阶段流与事件流
3. `ProductionAcceptanceProfile / AcceptanceRun` 进入 `Acceptance` 页
4. 用户人工确认和放行动作进入 `ActionInbox`

## 6. 页面设计原则

### 6.1 实时优先

工作台不是静态详情页，而是实时状态面板。

用户必须持续看到：

1. 当前 active stage
2. 当前 active run
3. 最新事件
4. 最新 blocker
5. 最新待处理动作

### 6.2 项目推进优先

页面结构必须围绕项目推进，而不是围绕系统配置。

优先展示：

1. 阶段流
2. 任务流
3. 风险
4. 验收覆盖
5. 用户下一步动作

### 6.3 图形优先

主界面优先采用：

1. 阶段条
2. 时间线
3. 覆盖矩阵
4. 风险卡片
5. 待办卡片

表格只作为次级查看方式，不应成为默认主展示。

### 6.4 默认自动推导

系统默认推导：

1. 分类策略
2. 角色
3. execution brain
4. 验收 profile
5. 推荐下一步

用户主要做：

1. 确认
2. 修正
3. 放行
4. 处理 blocker

### 6.5 高频信息前置

高频信息必须在主工作台固定可见：

1. 当前阶段
2. 当前卡点
3. 当前运行项
4. 当前人工待处理
5. 当前 production readiness

低频设置必须后置到设置页或高级抽屉。

## 7. 单机版边界

V3 当前产品形态明确限定为单用户、本地优先。

因此本专题默认不覆盖以下内容：

1. 多租户隔离
2. 组织与成员管理
3. 管理员运营后台
4. 多角色审批流编排
5. SaaS 式数据权限体系

如果未来扩展多人协作，也应作为后续专题处理，而不能回退为 V2 式后台结构。

## 8. 顶层信息架构

V3 顶层导航建议收敛为 4 个主页面：

1. `Workspace`
2. `Plan`
3. `Acceptance`
4. `Settings`

其中：

- `Workspace` 是绝对主入口
- `Plan` 负责解释系统如何从方案走到任务
- `Acceptance` 负责证明项目是否达到生产级
- `Settings` 承担所有低频本地配置

## 9. Workspace 页面设计

`Workspace` 是 V3 的核心页面。

它不应该是“项目详情页”，而应是“实时项目驾驶舱”。

### 9.1 页面目标

用户进入 `Workspace` 后，应在一屏内获得：

1. 项目总状态
2. 当前阶段进度
3. 系统实时活动
4. 当前待处理问题
5. 当前验收覆盖状态

### 9.2 页面分区

建议拆成 5 个主区。

#### A. 顶部总状态条

固定展示：

1. `project_name`
2. `project_category`
3. `current_stage`
4. `overall_progress`
5. `current_run_status`
6. `production_readiness`

这一栏的目标是让用户在 3 秒内知道项目总体状态。

#### B. 阶段流区域

横向展示主阶段：

1. `Design`
2. `Review`
3. `Compile`
4. `Execute`
5. `Acceptance`
6. `Complete`

每个阶段至少显示：

1. `status`
2. `started_at`
3. `duration`
4. `active_item`
5. `blocker_count`

阶段流应支持点击展开该阶段的最近事件和主要问题。

阶段流的状态来源应严格对应业务状态机，而不是前端自己发明阶段。

#### C. 实时事件流区域

这是页面最重要的动态区。

按时间顺序展示：

1. 方案已生成
2. 审核发现问题
3. 方案已编译
4. 任务已启动
5. 任务已完成
6. 任务返工
7. 验收发现 blocker
8. 等待人工决策

每条事件至少带：

1. `event_type`
2. `time`
3. `source_brain`
4. `source_task`
5. `severity`
6. `needs_attention`

实时事件优先映射以下来源：

1. `PlanReviewResult` 生成与变更
2. `CompiledPlan` 生成与任务投影
3. `brain-v3` Run 生命周期变化
4. `DeliveryContract / VerificationContract` 的结果变化
5. `AcceptanceRun` 的 blocker 与裁决变化

#### D. 待处理问题区

这一块必须高亮显示，而不是埋进日志或详情里。

优先展示：

1. 当前 blocker
2. 当前人工审批
3. 当前缺失证据
4. 当前失败任务
5. 当前高风险提醒

每条待处理项都应包含：

1. 问题摘要
2. 严重级别
3. 是否阻塞
4. 推荐动作
5. 快捷操作按钮

`ActionInbox` 的来源建议只接 4 类事件：

1. 审核阻塞
2. 高风险任务确认
3. 验收阻塞
4. 人工放行

#### E. 验收覆盖区

底部区域用来展示当前距离 `production_passed` 还有多远。

推荐用覆盖矩阵展示：

1. `surface`
2. `journey`
3. `evidence`
4. `production_passed`

例如：

- `user_frontend`
- `admin_backend`
- `api_backend`
- `game_runtime`
- `editor_runtime`

每个格子至少显示：

1. `pass`
2. `partial`
3. `missing`

## 10. Plan 页面设计

`Plan` 页不是普通任务列表，而是“系统决策解释页”。

它的核心任务是回答：

1. 为什么这样拆任务
2. 哪些内容在 review 阶段被拦住
3. 哪些内容在 compile 阶段被修改
4. 为什么选择这个角色和这个 brain

### 10.1 主要模块

建议包含：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CompiledPlan`
4. 任务树
5. 编译差异说明

### 10.2 展示方式

推荐使用：

1. 三栏对比
2. 编译差异卡片
3. 任务树
4. 风险标签
5. role / brain 来源说明

## 11. Acceptance 页面设计

`Acceptance` 页是最终交付判断页面。

它必须回答：

1. 当前项目分类要求什么
2. 哪些 surface 已覆盖
3. 哪些 journey 已完成
4. 缺哪些 evidence
5. 是否达到 `production_passed`

### 11.1 主要模块

建议包含：

1. 当前分类验收 profile
2. surface 覆盖图
3. journey 覆盖图
4. evidence 卡片区
5. 最终裁决区

### 11.2 最终裁决区

裁决区必须明确区分：

1. `functional_passed`
2. `production_passed`
3. `manual_release_required`
4. `released_by_human`

用户必须能一眼看到当前停在哪一层。

最终裁决的状态来源必须对齐验收体系文档，不允许前端自定义一套近义状态。

## 12. Settings 页面设计

`Settings` 是低频页面，负责承接所有本地配置。

这里不应承载主流程信息。

建议只包含：

1. 本地工作目录
2. provider / model
3. brain runtime
4. 默认策略
5. 高级调试项

以下内容不应在主流程前置暴露：

1. 角色管理
2. execution mode 细粒度切换
3. tool scope 深度配置
4. 验收 profile 调优

这些都应折叠在高级设置中。

## 13. 核心实时数据模型

为支撑实时工作台，前端应优先围绕以下对象建模。

### 13.1 ProjectSnapshot

用于表示当前项目全局快照。

建议包含：

1. `project_id`
2. `project_name`
3. `project_category`
4. `current_stage`
5. `overall_progress`
6. `active_run_id`
7. `active_task_count`
8. `blocker_count`
9. `waiting_action_count`
10. `acceptance_status`
11. `production_readiness`
12. `manual_release_required`

`ProjectSnapshot` 是聚合快照，不应承担细粒度日志存储责任。

### 13.2 StageProgress

用于表示阶段进展。

建议包含：

1. `stage_name`
2. `status`
3. `started_at`
4. `ended_at`
5. `duration_ms`
6. `active_item`
7. `blockers`

### 13.3 LiveEvent

用于表示实时事件。

建议包含：

1. `event_id`
2. `event_type`
3. `project_id`
4. `source_run_id`
5. `source_task_id`
6. `source_brain`
7. `severity`
8. `summary`
9. `created_at`
10. `needs_attention`
11. `source_object_kind`
12. `source_object_id`

### 13.4 ActionInboxItem

用于表示待处理项。

建议包含：

1. `item_id`
2. `kind`
3. `severity`
4. `blocking`
5. `title`
6. `recommended_action`
7. `action_url`
8. `deadline_hint`
9. `resolved_at`

### 13.5 AcceptanceCoverage

用于表示验收覆盖状态。

建议包含：

1. `project_category`
2. `surface`
3. `journey`
4. `evidence_type`
5. `coverage_status`
6. `last_verified_at`
7. `blocking_gap`

## 14. 关键界面映射关系

为了避免页面和架构脱节，建议固定以下映射关系：

1. `Workspace` 主要展示 `DomainTask + brain-v3 Run + ActionInbox + AcceptanceCoverage`
2. `Plan` 主要展示 `PlanDraft + PlanReviewResult + CompiledPlan`
3. `Acceptance` 主要展示 `ProductionAcceptanceProfile + AcceptanceRun + Evidence`
4. `Settings` 主要展示本地运行时和低频配置

不应出现：

1. 在 `Workspace` 里塞满低频系统设置
2. 在 `Plan` 里展示执行日志为主
3. 在 `Acceptance` 里重新解释计划编译逻辑

## 15. 实时刷新与状态更新原则

V3 工作台应尽量避免“手动刷新才能知道系统状态”。

推荐原则：

1. 顶部项目快照自动刷新
2. 事件流增量刷新
3. 待处理区优先推送
4. 验收覆盖在关键状态变更后即时更新

如果无法全量实时推送，至少保证：

1. 关键状态变化立即刷新
2. 事件流支持短轮询
3. 待处理问题在视觉上始终最醒目

## 16. UI 风格原则

为了满足“轻量化、可视化、方便、简洁 UI”，建议采用以下风格：

1. 浅色背景为主
2. 强层级卡片布局
3. 总色数严格控制
4. 重点状态用有限强调色
5. 减少重边框和大面积表格

页面视觉应更接近：

1. 工作流驾驶舱
2. 实时活动面板
3. 发布准备中心

而不应接近：

1. ERP 后台
2. 企业 CRUD 管理系统
3. 配置中心首页

## 17. 应删减或后置的能力

为了维持单机版的轻量形态，以下能力应删除或后置：

1. 多租户体系
2. 管理员运营页
3. 大量独立配置菜单
4. 角色前置配置流程
5. 执行器前置选择流程
6. 规则前置管理页
7. 大量低价值统计卡

## 18. 与 V3 架构的关系

本专题不是纯 UI 文档，而是 V3 产品形态约束文档。

它会直接影响：

1. 前端信息架构
2. 实时状态聚合接口
3. 事件流模型
4. ActionInbox 设计
5. AcceptanceCoverage 输出模型

因此后续接口和前端实现都应优先围绕本页定义推进，而不是先做后台页再回头裁剪。

## 19. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
2. [EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
3. [EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
4. [EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-实时事件流推送机制设计.md)
5. `ProjectSnapshot / LiveEvent / ActionInboxItem` 接口 schema
6. `AcceptanceCoverage` 可视化组件设计
7. 单机版导航与页面跳转规则
