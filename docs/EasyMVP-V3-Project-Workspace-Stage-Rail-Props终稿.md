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

## 3. 口径补充

按当前钱学森总纲，`stage` 建议优先收口到：

1. `reviewing`
2. `executing`
3. `accepting`
4. `reworking`
5. `completed`

不再建议继续在单项目 `Stage Rail` 中把 `Compile` 暴露成独立主阶段。
