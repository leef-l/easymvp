# EasyMVP V3 Workspace 首页 Release Readiness 卡组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页聚合接口Schema设计](./EasyMVP-V3-Workspace首页聚合接口Schema设计.md)
> 目标：定义首页右下 `Release Readiness` 区块中单项目可交付卡片的结构、状态文案和动作规则。

## 1. 设计结论

`Release Readiness` 卡的任务不是显示更多状态，而是建立“离交付还有多远”的直觉。

## 2. 组件目标

每张卡需要回答：

1. 功能是否通过
2. 生产级是否通过
3. 是否还需要人工放行
4. 点哪里继续处理

## 3. 字段

建议固定：

1. `project_name`
2. `functional_passed`
3. `production_passed`
4. `manual_release_required`
5. `released_by_human`
6. `acceptance_link`

## 4. 文案原则

建议使用：

1. `Functional: passed / not yet`
2. `Production: passed / blocked`
3. `Manual release: required / not required`

不要把一堆内部布尔值直接丢给用户。

## 5. 动作

建议只有一个主动作：

1. `Open Acceptance`

## 6. 排序

默认优先显示：

1. `production_passed = false` 但接近完成
2. `manual_release_required = true`
3. 已完成项目最后

## 7. 不该怎么做

不应该：

1. 卡片信息过多
2. 没有明确动作
3. 看不出生产级与功能级的区别

## 8. 后续细分专题

1. Release Readiness 卡视觉稿
2. readiness 文案映射表

