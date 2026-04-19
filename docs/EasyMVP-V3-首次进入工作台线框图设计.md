# EasyMVP V3 首次进入工作台线框图设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建后首次进入Project-Workspace引导态设计](./EasyMVP-V3-创建后首次进入Project-Workspace引导态设计.md)
> 关联文档：[EasyMVP-V3-首次进入引导卡组件规范](./EasyMVP-V3-首次进入引导卡组件规范.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 目标：把新项目首次进入 `Project Workspace` 的特殊布局推进到线框图级别。

## 1. 设计结论

首次进入工作台时，应当在常规三栏布局上做一次“启动期重排”。

## 2. 线框结构

```text
┌──────────────────────────────────────────────────────────────────┐
│ Top Status Bar: Project created / preparing first plan          │
├──────────────────────────────────────────────────────────────────┤
│ Startup Guide Card                                              │
├───────────────────────┬───────────────────────┬──────────────────┤
│ Stage Rail            │ Init Activity         │ Next Actions     │
│ Design active         │ creation events       │ Open Plan         │
│ others muted          │ plan bootstrap        │ Check path        │
├───────────────────────┴───────────────────────┴──────────────────┤
│ Acceptance Preparation Summary                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 3. 区域规则

### 3.1 Top Status Bar

突出：

1. 项目已创建
2. 当前准备进度

### 3.2 Startup Guide Card

复用：

1. [EasyMVP-V3-首次进入引导卡组件规范](./EasyMVP-V3-首次进入引导卡组件规范.md)

### 3.3 Stage Rail

只高亮 `Design`。

### 3.4 Init Activity

优先展示创建初始化事件，不混入过多普通运行事件。

### 3.5 Next Actions

优先展示：

1. `Open Plan`
2. `View Init Events`

### 3.6 Acceptance Preparation Summary

不展示完整矩阵，只展示准备态说明。

## 4. 空态避免规则

首次进入页绝不能出现：

1. 大块空白
2. 空表格
3. `No data`

## 5. 后续细分专题

1. 首次进入与常规工作台切换动画
2. 首次进入响应式降级规则
