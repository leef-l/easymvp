# EasyMVP V3 Workspace 首页 Release Readiness 卡 Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页Release-Readiness卡组件规范](./EasyMVP-V3-Workspace首页Release-Readiness卡组件规范.md)
> 目标：定义 `Release Readiness` 卡最终 props。

## 1. Props

```ts
type ReleaseReadinessCardProps = {
  items: ReleaseReadinessItem[]
  onOpenAcceptance?: (projectId: string) => void
}
```

## 2. `ReleaseReadinessItem`

```ts
type ReleaseReadinessItem = {
  project_id: string
  project_name: string
  production_status: string
  missing_count: number
  next_action: string
}
```

