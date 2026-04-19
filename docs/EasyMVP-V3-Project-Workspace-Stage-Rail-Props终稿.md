# EasyMVP V3 Project Workspace Stage Rail Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Project-Workspace-Stage-Rail组件规范](./EasyMVP-V3-Project-Workspace-Stage-Rail组件规范.md)
> 目标：定义 `Stage Rail` 最终 props。

## 1. Props

```ts
type StageRailProps = {
  stages: StageProgressItem[]
  onSelectStage?: (stage: string) => void
}
```

## 2. `StageProgressItem`

```ts
type StageProgressItem = {
  stage: string
  status: string
  progress_percent: number
  blocker_count: number
  started_at?: string
  finished_at?: string
}
```

