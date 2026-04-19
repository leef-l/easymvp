# EasyMVP V3 Plan 任务投影抽屉设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
> 关联文档：[EasyMVP-V3-Plan线框图设计](./EasyMVP-V3-Plan线框图设计.md)
> 目标：定义 `Plan` 页中点击任务投影项后打开的详情抽屉结构、字段与交互规则。

## 1. 设计结论

任务投影抽屉用来解释：

1. 一个编译后任务从哪里来
2. 为什么归到这个 role / brain
3. 它的合同和风险是什么

## 2. 抽屉结构

建议分四段：

1. 任务头部摘要
2. 来源映射
3. role / brain 解析
4. delivery / verification / risk

## 3. 线框

```text
┌──────────────────────────────────────────┐
│ Task Name                                │
│ role: architect  brain: easymvp-brain    │
├──────────────────────────────────────────┤
│ Source Mapping                            │
│ source_task_key: draft_task_01            │
│ compiled_from: review split suggestion    │
├──────────────────────────────────────────┤
│ Contracts                                 │
│ delivery: ...                             │
│ verification: ...                         │
│ risk: medium                              │
└──────────────────────────────────────────┘
```

## 4. 关键字段

建议展示：

1. `task_name`
2. `task_kind`
3. `source_task_key`
4. `role_type`
5. `brain_kind`
6. `risk_level`
7. `delivery_summary`
8. `verification_summary`

## 5. 交互

允许：

1. 查看来源 review issue
2. 查看 diff 卡来源
3. 跳转执行阶段对应任务

## 6. 不该怎么做

不应该：

1. 只显示 task id
2. 不解释来源
3. 不显示合同和风险

## 7. 后续细分专题

1. 抽屉视觉稿
2. 来源映射图规范

