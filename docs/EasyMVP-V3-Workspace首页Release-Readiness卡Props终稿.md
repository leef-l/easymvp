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

## 3. 口径补充

按当前钱学森总纲，这个 props 结构可以继续保留为旧首页壳层首版字段。

但更准确的后续方向应逐步补齐：

1. `decision`
2. `completed`
3. `manual_release_required`
4. `blocking_reason`

核心边界：

- `production_status` 不应再被实现者误读为“最终完成状态”
- `Release Readiness` 卡只能表达接近收口程度，不能替代 `CompletionVerdict`
