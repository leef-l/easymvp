# EasyMVP V3 Project Workspace Live Activity Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Project-Workspace-Live-Activity组件规范](./EasyMVP-V3-Project-Workspace-Live-Activity组件规范.md)
> 目标：定义 `Live Activity` 最终 props。

## 1. Props

```ts
type LiveActivityPanelProps = {
  items: LiveActivityItem[]
  onOpenEvent?: (id: string) => void
}
```

