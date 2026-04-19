# EasyMVP V3 Plan diff 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
> 关联文档：[EasyMVP-V3-Plan线框图设计](./EasyMVP-V3-Plan线框图设计.md)
> 目标：定义 `Plan` 页中差异解释卡片的结构、类型、文案与来源引用规则。

## 1. 设计结论

`Plan diff` 组件是 V3 解释力的核心组件。

它不是代码 diff，而是：

1. 计划结构变化
2. 角色变化
3. 风险变化
4. 合同增强

的解释卡。

## 2. 类型

建议固定：

1. `task_split`
2. `task_drop`
3. `brain_override`
4. `role_override`
5. `risk_upgrade`
6. `contract_enhanced`

## 3. 卡片结构

建议固定：

1. 变化类型
2. `before`
3. `after`
4. `reason`
5. `source_review_issue_id`

## 4. 线框

```text
┌──────────────────────────────────────────┐
│ task_split                               │
│ before: Build whole backend              │
│ after: API layer + admin actions         │
│ reason: scope too large for one task     │
│ source: review_issue_12                  │
└──────────────────────────────────────────┘
```

## 5. 文案原则

必须让用户直接看懂：

1. 原来是什么
2. 现在变成什么
3. 为什么这么变

## 6. 不该怎么做

不应该：

1. 只显示内部 diff 码值
2. before/after 缺一个
3. 没有原因说明

## 7. 后续细分专题

1. diff 卡视觉稿
2. diff 类型到颜色映射

