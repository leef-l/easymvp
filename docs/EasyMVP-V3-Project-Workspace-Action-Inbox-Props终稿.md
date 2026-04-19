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

