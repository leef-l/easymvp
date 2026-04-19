# EasyMVP V3 Workspace 详细页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 `Workspace` 从页面分区设计推进到可落地的模块、状态、交互和实时刷新方案。

## 1. 设计结论

`Workspace` 是 V3 的绝对主入口。

它必须一屏回答 4 个问题：

1. 项目当前处于哪个阶段
2. 系统当前在做什么
3. 当前卡住了哪些问题
4. 离 `production_passed` 还有多远

## 2. 页面布局

建议采用：

1. 顶部固定状态条
2. 中部三栏主区
3. 底部验收覆盖区

布局建议：

```text
┌──────────────────────────────────────────────────────────────┐
│ Top Status Bar                                              │
├───────────────────────┬───────────────────────┬──────────────┤
│ Stage Rail            │ Live Activity         │ Action Inbox │
│                       │                       │              │
├───────────────────────┴───────────────────────┴──────────────┤
│ Acceptance Coverage / Release Readiness                     │
└──────────────────────────────────────────────────────────────┘
```

## 3. 顶部固定状态条

### 3.1 目标

让用户 3 秒内知道项目总体状态。

### 3.2 模块

建议固定展示：

1. `project_name`
2. `project_category`
3. `current_stage`
4. `overall_progress`
5. `current_run_status`
6. `production_readiness`
7. `final_status_hint`

### 3.3 状态表现

建议四态：

1. `idle`
2. `running`
3. `blocked`
4. `ready_to_release`

### 3.4 交互

允许点击：

1. `current_stage` 跳转对应页面
2. `production_readiness` 打开验收详情
3. `current_run_status` 打开活跃 run 抽屉

## 4. 阶段流区域

### 4.1 目标

展示主链路推进情况，而不是对象列表。

### 4.2 阶段定义

固定为：

1. `Design`
2. `Review`
3. `Compile`
4. `Execute`
5. `Acceptance`
6. `Complete`

### 4.3 每阶段卡片字段

建议至少展示：

1. `stage_name`
2. `status`
3. `started_at`
4. `duration_ms`
5. `active_item_title`
6. `blocker_count`
7. `linked_view`

### 4.4 状态颜色

建议：

1. `pending` 使用中性灰
2. `active` 使用高亮主色
3. `blocked` 使用告警色
4. `done` 使用通过色

### 4.5 点击行为

点击阶段卡片时：

1. `Design / Review / Compile` 打开 `Plan`
2. `Execute` 展开任务与 run 明细
3. `Acceptance` 跳转 `Acceptance`
4. `Complete` 展示最终总结抽屉

## 5. Live Activity 区域

### 5.1 目标

把系统的“当前动作”和“刚发生的关键变化”直接显出来。

### 5.2 数据来源

来自：

1. `PlanReviewResult`
2. `CompiledPlan`
3. `BrainRunBinding`
4. `workflow_brain_run_events`
5. `AcceptanceRun`

### 5.3 事件类型

建议至少包括：

1. `plan_created`
2. `review_blocking_found`
3. `plan_compiled`
4. `task_started`
5. `task_succeeded`
6. `task_failed`
7. `run_status_changed`
8. `acceptance_blocker_found`
9. `manual_action_required`

### 5.4 展示规则

实时区只展示高价值事件，不原样展示 logs。

每条事件展示：

1. 时间
2. 事件摘要
3. 来源 brain
4. 来源任务
5. 严重级别
6. 深链按钮

### 5.5 默认筛选

默认只显示：

1. 最近 30 条
2. `severity >= info`
3. `needs_attention = true` 的事件置顶

## 6. Action Inbox 区域

### 6.1 目标

把“你现在需要做什么”从日志和详情里剥离出来。

### 6.2 卡片类型

建议固定：

1. `review_blocker`
2. `risk_confirmation`
3. `acceptance_blocker`
4. `manual_release_required`
5. `run_sync_failed`

### 6.3 卡片字段

建议至少包含：

1. `title`
2. `summary`
3. `severity`
4. `blocking`
5. `recommended_action`
6. `action_button_text`
7. `action_target`

### 6.4 排序规则

建议：

1. `blocking = true` 优先
2. `manual_release_required` 次优先
3. 最近创建时间优先

### 6.5 用户动作

允许：

1. 查看详情
2. 执行推荐动作
3. 暂时忽略非阻塞项

## 7. Acceptance Coverage 区域

### 7.1 目标

展示最终交付距离，而不是把验收藏到另一页才看得到。

### 7.2 展示结构

建议使用矩阵：

1. 行为 `surface`
2. 列为 `journey / evidence`

### 7.3 状态

建议：

1. `pass`
2. `partial`
3. `missing`
4. `blocked`

### 7.4 补充模块

矩阵旁边增加：

1. `functional_passed`
2. `production_passed`
3. `manual_release_required`
4. `released_by_human`

## 8. 默认态 / 运行态 / 阻塞态

### 8.1 默认态

特征：

1. 无活跃 run
2. 尚未进入执行
3. 阶段流停在 `Design / Review / Compile`

主按钮建议：

1. `查看计划`
2. `继续编译`

### 8.2 运行态

特征：

1. 存在活跃 run
2. Live Activity 高频刷新
3. `current_run_status = running`

主按钮建议：

1. `查看活跃任务`
2. `暂停/取消`

### 8.3 阻塞态

特征：

1. 顶部状态条进入 `blocked`
2. `ActionInbox` 至少一项 `blocking = true`
3. 阶段流对应阶段高亮告警

主按钮建议：

1. `处理 blocker`
2. `查看原因`

## 9. 抽屉与二级视图

Workspace 不应该跳很多页，但可以用抽屉展开细节。

建议有：

1. `run detail drawer`
2. `task detail drawer`
3. `blocker detail drawer`
4. `release gate drawer`

## 10. 刷新策略

### 10.1 顶部状态条

建议短轮询或 SSE 更新。

### 10.2 Live Activity

建议增量推送。

### 10.3 Action Inbox

建议高优先级即时刷新。

### 10.4 Coverage

建议在验收状态变化时刷新。

## 11. 不该怎么做

不应该：

1. 把 `Workspace` 做成大对象详情页
2. 让用户切多级 tab 才看到 blocker
3. 把 logs 原样铺满主区
4. 把大量设置前置到主入口

## 12. 后续细分专题

本专题后续继续拆：

1. Workspace 线框图
2. 组件级状态图
3. Mobile/窄屏降级方案
