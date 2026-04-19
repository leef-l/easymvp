# EasyMVP V3 Workspace 首页 Recent Activity 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页聚合接口Schema设计](./EasyMVP-V3-Workspace首页聚合接口Schema设计.md)
> 目标：定义 `Workspace Home` 中跨项目 `Recent Activity` 区块的展示结构、过滤规则和交互行为。

## 1. 设计结论

`Recent Activity` 不是全量活动流。

它只展示跨项目、对用户有价值的高信号事件。

## 2. 组件目标

每条活动要回答：

1. 哪个项目发生了变化
2. 发生了什么
3. 要不要我关注
4. 点哪里继续看

## 3. 字段

建议固定：

1. `created_at`
2. `project_name`
3. `event_type`
4. `summary`
5. `severity`
6. `deep_link`

## 4. 推荐事件

首批建议：

1. `plan_compiled`
2. `task_failed`
3. `acceptance_blocker_found`
4. `manual_action_required`
5. `release.readiness_changed`

## 5. 排序

默认按时间倒序。

高严重级别可做轻度置顶，但不要打乱时间感。

## 6. 交互

每条活动允许：

1. 打开关联项目
2. 打开关联页面
3. 按项目过滤活动流

## 7. 不该怎么做

不应该：

1. 把所有低价值事件都放进来
2. 只显示事件码
3. 没有项目名

## 8. 后续细分专题

1. Recent Activity 视觉稿
2. 事件类型到文案映射表

