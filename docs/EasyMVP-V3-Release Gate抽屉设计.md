# EasyMVP V3 Release Gate 抽屉设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-Acceptance线框图设计](./EasyMVP-V3-Acceptance%E7%BA%BF%E6%A1%86%E5%9B%BE%E8%AE%BE%E8%AE%A1.md)
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 目标：定义 V3 中 Release Gate 的抽屉形态、打开入口、信息结构、人工放行流程和阻塞原因展示方式。

## 1. 设计结论

Release Gate 不应只是页面底部的一个按钮。

正确做法是：

1. Acceptance 页底部展示 Release Gate 摘要
2. 点击后打开独立 `Release Gate Drawer`
3. 在抽屉内解释为什么现在能/不能交付
4. 在需要时承载人工放行动作

一句话：

> Release Gate 是最终交付决策层，不是普通按钮层。

## 2. 打开入口

建议至少支持：

1. Acceptance 页底部 `View Reason`
2. Acceptance 页底部 `Manual Release`
3. 顶部最终裁决条中的 `manual_release_required`

## 3. 抽屉定位

`Release Gate Drawer` 的任务不是重复页面信息，而是回答：

1. 现在为什么能交付或不能交付
2. 是谁阻塞了交付
3. 若需要人工放行，动作是什么
4. 放行后状态会如何变化

## 4. 抽屉布局

建议分成四块：

1. 顶部状态摘要
2. 阻塞原因区
3. 放行规则区
4. 动作区

推荐结构：

```text
┌──────────────────────────────────────────────────────────────┐
│ Gate Summary                                                 │
├──────────────────────────────────────────────────────────────┤
│ Blocking Reasons                                             │
├──────────────────────────────────────────────────────────────┤
│ Manual Release Rules / Conditions                            │
├──────────────────────────────────────────────────────────────┤
│ Actions                                                      │
└──────────────────────────────────────────────────────────────┘
```

## 5. 顶部状态摘要

建议展示：

1. `can_release`
2. `requires_manual_release`
3. `production_passed`
4. `released_by_human`
5. `current_blocking_reason`

## 6. 阻塞原因区

### 6.1 目标

把“为什么不能交付”从抽象状态翻译成清晰的可读原因。

### 6.2 展示内容

建议至少列出：

1. 当前 blocking gap
2. 当前未满足的 gate
3. 缺失的关键 evidence
4. 当前仍未关闭的 blocking issue

### 6.3 交互

允许：

1. 跳转到对应 Coverage
2. 跳转到缺失 Evidence 分组
3. 跳转到相关 issue 或 replay

## 7. 放行规则区

### 7.1 目标

解释“为什么需要人工放行”，而不是只给一个按钮。

### 7.2 展示内容

建议展示：

1. 触发的 `manual_release_rules`
2. 风险等级
3. 放行前提是否已满足
4. 放行后果说明

## 8. 动作区

### 8.1 允许动作

建议最多三个：

1. `查看阻塞详情`
2. `执行人工放行`
3. `查看放行记录`

### 8.2 `执行人工放行`

仅在以下条件同时满足时高亮可点：

1. `production_passed = true`
2. `manual_release_required = true`
3. `released_by_human = false`
4. 所需确认条件已满足

### 8.3 放行动作确认

建议：

1. 二次确认
2. 要求用户明确知道当前是在放行生产可交付状态

## 9. 典型状态

### 9.1 功能通过但不能交付

抽屉主信息应强调：

1. 当前仍未达到 `production_passed`
2. 缺少哪些证据或 gate

### 9.2 生产通过但需人工放行

抽屉主信息应强调：

1. 已达到 `production_passed`
2. 但仍需人工确认后才算真正完成

### 9.3 已完成人工放行

抽屉主信息应强调：

1. 已放行
2. 放行时间
3. 放行记录入口

## 10. 不该怎么做

不建议：

1. 只在底部放一个按钮不解释原因
2. 把人工放行做成无上下文的危险动作
3. 缺少阻塞原因明细
4. 放行后不保留可见记录入口

## 11. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Acceptance 高保真设计必须为 Release Gate 预留独立抽屉入口
2. 后续人工放行动作记录模型应直接服务此抽屉展示
3. Coverage / Evidence / Replay 的深链必须可从此抽屉进入

## 12. 后续细分专题

本专题后续继续拆：

1. 人工放行记录模型设计
2. Release Gate 高保真视觉设计
3. 放行动作确认弹层设计
