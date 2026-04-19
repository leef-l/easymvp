# EasyMVP V3 Project Workspace Stage Rail 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-Project-Workspace线框图设计](./EasyMVP-V3-Project-Workspace线框图设计.md)
> 目标：定义单项目工作台左侧 `Stage Rail` 区域的阶段节点结构、状态样式和交互规则。

## 1. 设计结论

`Stage Rail` 不是导航树。

它是主链路状态带，用来表示项目推进位置与阻塞位置。

## 2. 阶段

固定为：

1. `Design`
2. `Review`
3. `Compile`
4. `Execute`
5. `Acceptance`
6. `Complete`

## 3. 每个阶段节点字段

建议展示：

1. `stage_name`
2. `status`
3. `duration`
4. `active_item_title`
5. `blocker_count`

## 4. 状态

建议固定：

1. `pending`
2. `active`
3. `blocked`
4. `done`

## 5. 点击行为

点击阶段时：

1. `Design / Review / Compile` 打开 `Plan`
2. `Execute` 打开 run/task 详情
3. `Acceptance` 打开 `Acceptance`
4. `Complete` 打开总结抽屉

## 6. 首次进入特殊规则

首次进入时：

1. 只高亮 `Design`
2. 其余阶段弱化为“即将进入”

## 7. 不该怎么做

不应该：

1. 做成普通左侧菜单
2. 不显示 blocker
3. 阶段完成与阻塞没有明显视觉差异

## 8. 后续细分专题

1. Stage Rail 视觉稿
2. 阶段状态到页面跳转映射表

