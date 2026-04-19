# EasyMVP V3 Workspace 首页 Need Attention 卡 Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页Need-Attention卡组件规范](./EasyMVP-V3-Workspace首页Need-Attention卡组件规范.md)
> 目标：定义 `Need Attention` 单张卡片实现所需的最终 props、事件和状态约束。

## 1. Props

```ts
type NeedAttentionCardItemProps = {
  item: NeedAttentionItem
  onResolve: (item: NeedAttentionItem) => void
  onOpenProject: (projectId: string) => void
}
```

## 2. `NeedAttentionItem`

```ts
type NeedAttentionItem = {
  id: string
  project_id: string
  project_name: string
  issue_type: string
  severity: "warning" | "error" | "critical"
  summary: string
  recommended_action: string
  blocking: boolean
}
```

## 3. 本地状态

1. hover
2. button pending

