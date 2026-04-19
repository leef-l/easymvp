# EasyMVP V3 Project Workspace Live Activity 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-实时事件流推送机制设计.md)
> 目标：定义单项目工作台中 `Live Activity` 区块的事件结构、排序、过滤与交互规则。

## 1. 设计结论

`Live Activity` 不是日志窗口。

它是单项目工作台中央主区，用来表达：

1. 系统正在做什么
2. 刚刚发生了什么关键变化
3. 哪些变化需要用户关注

## 2. 组件目标

每条活动项至少要回答：

1. 什么时候发生
2. 发生了什么
3. 来自哪个任务或归一化来源
4. 要不要我介入

## 3. 结构

建议每条活动项固定展示：

1. `created_at`
2. `summary`
3. `source_brain`
4. `source_task`
5. `severity`
6. `deep_link`

说明：

1. `source_brain` 只是聚合层补出的来源归属字段
2. 它不直接暴露 `brain-v3` 工具面，也不表示页面层可以反推执行职责

## 4. 事件类型

首批建议：

1. `plan_created`
2. `review_blocking_found`
3. `plan_compiled`
4. `task_started`
5. `task_succeeded`
6. `task_failed`
7. `run_status_changed`
8. `acceptance_blocker_found`
9. `manual_action_required`
10. `creation.workspace_ready`

## 5. 排序与过滤

默认：

1. 按时间倒序
2. `needs_attention=true` 置顶展示
3. 仅展示高价值事件

## 6. 展开规则

建议：

1. 普通事件单行摘要
2. 失败、blocker、manual action 默认展开一层说明

## 7. 交互

每条事件允许：

1. 打开关联抽屉
2. 跳转关联页面
3. 查看来源对象

## 8. 不该怎么做

不应该：

1. 原样铺日志
2. 没有时间和来源
3. 失败事件与普通事件没有层级差异

## 9. 后续细分专题

1. Live Activity 视觉稿
2. 事件颜色映射
