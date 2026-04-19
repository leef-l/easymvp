# EasyMVP V3 Project Workspace 线框图设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-Project-Workspace-Action-Inbox组件规范](./EasyMVP-V3-Project-Workspace-Action-Inbox组件规范.md)
> 关联文档：[EasyMVP-V3-首次进入工作台线框图设计](./EasyMVP-V3-首次进入工作台线框图设计.md)
> 目标：把常规态单项目工作台推进到线框图级别，明确三栏比例、首屏顺序和模块分层。

## 1. 设计结论

常规态 `Project Workspace` 应是一个稳定三栏驾驶舱，而不是 tab 详情页。

按当前钱学森总纲，这个三栏驾驶舱必须优先服务闭环对象可见性，而不是只展示活动和 readiness 摘要。

## 2. 线框结构

```text
┌────────────────────────────────────────────────────────────────────┐
│ Top Status Bar                                                    │
├────────────────────────┬────────────────────────┬──────────────────┤
│ Stage Rail             │ Live Activity          │ Action Inbox     │
│ stages + blockers      │ high value events      │ next actions     │
├────────────────────────┴────────────────────────┴──────────────────┤
│ Verification Coverage / Completion Readiness                      │
└────────────────────────────────────────────────────────────────────┘
```

## 3. 区域优先级

1. Top Status Bar
2. Live Activity
3. Action Inbox
4. Stage Rail
5. Verification / Completion Coverage

## 4. 比例建议

建议：

1. 左栏 `26%`
2. 中栏 `46%`
3. 右栏 `28%`

## 5. 响应式降级

平板下：

1. `Action Inbox` 下移
2. `Live Activity` 保持主区

## 6. 不该怎么做

不应该：

1. 切成很多 tab
2. 让 `Action Inbox` 折叠太深
3. 首屏看不到最新事件
4. 首屏只有 readiness，没有结构化阻塞原因

## 7. 后续细分专题

1. 常规态工作台视觉稿
2. 响应式断点规则
