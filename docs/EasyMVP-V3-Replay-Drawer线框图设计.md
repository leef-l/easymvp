# EasyMVP V3 Replay Drawer 线框图设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay Drawer页面设计.md)
> 关联文档：[EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
> 目标：把 `Replay Drawer` 推进到线框图级别，明确布局、区域比例和加载顺序。

## 1. 线框结构

```text
┌──────────────────────────────────────────────┐
│ Replay Summary                               │
├──────────────────────────────────────────────┤
│ Timeline                                     │
├──────────────────────────────────────────────┤
│ Event Detail / Raw / Linked Evidence         │
└──────────────────────────────────────────────┘
```

## 2. 加载顺序

建议：

1. 先显示 Summary
2. 再显示 Timeline
3. 再按需加载 Raw 和关联对象

## 3. 不该怎么做

不应该：

1. 打开抽屉后空白太久
2. 直接先渲染大量原始日志

