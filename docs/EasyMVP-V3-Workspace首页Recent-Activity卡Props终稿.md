# EasyMVP V3 Workspace 首页 Recent Activity 卡 Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页Recent-Activity组件规范](./EasyMVP-V3-Workspace首页Recent-Activity组件规范.md)
> 目标：定义 `Recent Activity` 卡最终 props。

## 1. Props

```ts
type RecentActivityCardProps = {
  items: LiveActivityItem[]
  onOpenProject?: (projectId: string) => void
}
```

## 2. `LiveActivityItem`

```ts
type LiveActivityItem = {
  id: string
  project_id?: string
  project_name?: string
  timestamp: string
  level: "info" | "warning" | "error"
  source_kind: string
  summary: string
}
```

