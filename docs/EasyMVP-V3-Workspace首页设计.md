# EasyMVP V3 Workspace 首页设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页多项目卡片组件规范](./EasyMVP-V3-Workspace首页多项目卡片组件规范.md)
> 关联文档：[EasyMVP-V3-Workspace首页Need-Attention卡组件规范](./EasyMVP-V3-Workspace首页Need-Attention卡组件规范.md)
> 关联文档：[EasyMVP-V3-Workspace首页Recent-Activity组件规范](./EasyMVP-V3-Workspace首页Recent-Activity组件规范.md)
> 关联文档：[EasyMVP-V3-Workspace首页Stage-Overview组件规范](./EasyMVP-V3-Workspace首页Stage-Overview组件规范.md)
> 关联文档：[EasyMVP-V3-Workspace首页Release-Readiness卡组件规范](./EasyMVP-V3-Workspace首页Release-Readiness卡组件规范.md)
> 目标：定义 V3 工作台首页的多项目实时总览设计，让用户一进入系统就能看到所有进行中项目的关键状态。

## 1. 设计结论

`Workspace` 顶层必须是多项目实时总览首页。

它不是单项目驾驶舱，而是：

> 所有进行中项目的实时总览首页。

## 2. 首页核心问题

首页必须在一屏内回答：

1. 现在有哪些项目在跑
2. 哪些项目卡住了
3. 哪些项目等我处理
4. 哪些项目快能交付了

## 3. 页面布局

建议采用：

1. 顶部全局状态条
2. 中部左侧 Running Projects
3. 中部右侧 Need Attention
4. 下部 Stage Overview / Recent Activity / Release Readiness

## 4. 顶部全局状态条

建议展示：

1. `running_project_count`
2. `blocked_project_count`
3. `waiting_action_count`
4. `ready_for_acceptance_count`

## 5. Running Projects 区

### 5.1 目标

展示所有进行中的项目卡片，而不是只看一个项目。

### 5.2 每张项目卡

建议至少显示：

1. `project_name`
2. `project_category`
3. `current_stage`
4. `overall_progress`
5. `active_task_or_run`
6. `blocker_count`
7. `waiting_action`
8. `production_readiness`

### 5.3 操作

每张卡片至少允许：

1. `进入项目`
2. `查看 blocker`
3. `打开 Acceptance`

## 6. Need Attention 区

### 6.1 目标

聚合所有项目中最需要人工处理的事项。

### 6.2 建议类型

1. `review_blocker`
2. `run_failed`
3. `acceptance_blocker`
4. `manual_release_required`

### 6.3 展示字段

1. `project_name`
2. `issue_type`
3. `severity`
4. `summary`
5. `action_target`

## 7. Stage Overview 区

### 7.1 目标

从全局维度看所有项目当前分布。

### 7.2 展示方式

建议按阶段展示计数：

1. `Design`
2. `Review`
3. `Compile`
4. `Execute`
5. `Acceptance`
6. `Complete`

## 8. Recent Activity 区

### 8.1 目标

展示跨项目的最新高价值事件。

### 8.2 事件展示

每条至少包含：

1. `project_name`
2. `event_type`
3. `summary`
4. `created_at`
5. `deep_link`

## 9. Release Readiness 区

### 9.1 目标

帮助用户快速识别快可交付和卡在最后一步的项目。

### 9.2 展示字段

1. `project_name`
2. `functional_passed`
3. `production_passed`
4. `manual_release_required`
5. `released_by_human`

## 10. 首页交互规则

建议：

1. 点击项目卡进入 `Project Workspace`
2. 点击待处理项直接进入对应项目详情或 Acceptance
3. 点击阶段分布项筛选项目列表
4. 点击 `New Project` 打开创建项目流程

## 11. 不该怎么做

不应该：

1. 首页只显示统计数字
2. 首页没有进行中项目卡片
3. 首页必须切页后才知道哪些项目阻塞

## 12. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Workspace首页线框图设计](./EasyMVP-V3-Workspace首页线框图设计.md)
2. [EasyMVP-V3-Workspace首页多项目卡片组件规范](./EasyMVP-V3-Workspace首页多项目卡片组件规范.md)
3. [EasyMVP-V3-Workspace首页聚合接口Schema设计](./EasyMVP-V3-Workspace首页聚合接口Schema设计.md)
4. [EasyMVP-V3-Workspace首页Need-Attention卡组件规范](./EasyMVP-V3-Workspace首页Need-Attention卡组件规范.md)
5. [EasyMVP-V3-Workspace首页Recent-Activity组件规范](./EasyMVP-V3-Workspace首页Recent-Activity组件规范.md)
6. [EasyMVP-V3-Workspace首页Stage-Overview组件规范](./EasyMVP-V3-Workspace首页Stage-Overview组件规范.md)
7. [EasyMVP-V3-Workspace首页Release-Readiness卡组件规范](./EasyMVP-V3-Workspace首页Release-Readiness卡组件规范.md)
8. [EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
