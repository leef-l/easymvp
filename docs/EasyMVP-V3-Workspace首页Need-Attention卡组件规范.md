# EasyMVP V3 Workspace 首页 Need Attention 卡组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页线框图设计](./EasyMVP-V3-Workspace首页线框图设计.md)
> 目标：定义 `Workspace Home` 右侧 `Need Attention` 区块中跨项目待处理卡片的结构、优先级和交互规则。

## 1. 设计结论

`Need Attention` 卡不是 issue 列表项。

它必须是一个：

1. 跨项目统一格式
2. 立即告诉用户要不要处理
3. 一键跳到正确位置

的动作卡。

## 2. 组件目标

每张卡片必须回答：

1. 哪个项目出问题了
2. 这是什么问题
3. 严重不严重
4. 我现在点哪里处理

## 3. 线框

```text
┌──────────────────────────────────┐
│ Project Name                     │
│ review_blocker · high            │
│ Plan review found 2 blocking...  │
│ [Resolve] [Open Project]         │
└──────────────────────────────────┘
```

## 4. 字段

建议固定字段：

1. `project_name`
2. `issue_type`
3. `severity`
4. `summary`
5. `recommended_action`
6. `action_target`

## 5. 类型

建议首批固定：

1. `review_blocker`
2. `run_failed`
3. `acceptance_blocker`
4. `manual_release_required`
5. `path_or_workspace_problem`
6. `verification_conflict`
7. `fault_loop_detected`
8. `policy_denied`

## 6. 排序

默认排序：

1. `blocking = true`
2. `manual_release_required`
3. `verification_conflict / fault_loop_detected`
4. `severity = high`
4. 最近更新时间

## 7. 动作

每张卡最多两个按钮：

1. 主动作 `Resolve`
2. 次动作 `Open Project`

## 8. 文案原则

主文案用自然语言摘要，不要直接暴露内部码值。

## 9. 不该怎么做

不应该：

1. 一卡放很多字段
2. 要点开后才知道是什么问题
3. 一个卡片塞 4 个按钮
4. 把升级类问题隐藏成普通失败

## 10. 后续细分专题

1. Need Attention 卡视觉稿
2. 跨项目动作映射表
