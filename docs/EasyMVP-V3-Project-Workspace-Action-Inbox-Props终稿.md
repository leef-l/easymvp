# EasyMVP V3 Project Workspace Action Inbox Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Project-Workspace-Action-Inbox组件规范](./EasyMVP-V3-Project-Workspace-Action-Inbox组件规范.md)
> 目标：定义 `Action Inbox` 最终 props。

## 1. Props

```ts
type ActionInboxPanelProps = {
  items: ActionInboxItem[]
  onAction: (item: ActionInboxItem) => void
}
```

## 2. `ActionInboxItem`

```ts
type ActionInboxItem = {
  id: string
  severity: "warning" | "error" | "critical"
  title: string
  summary: string
  action_label: string
  action_kind: string
}
```

## 3. 口径补充

按当前钱学森总纲，建议后续补充或等价承载以下信息：

1. `blocking`
2. `reason_code`
3. `source_object_kind`
4. `source_object_id`

这样才能避免以下问题被压平：

1. `verification_conflict`
2. `fault_loop_detected`
3. `policy_denied`
4. `manual_review_required`
