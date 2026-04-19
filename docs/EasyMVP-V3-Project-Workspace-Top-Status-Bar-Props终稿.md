# EasyMVP V3 Project Workspace Top Status Bar Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Project-Workspace-Top-Status-Bar组件规范](./EasyMVP-V3-Project-Workspace-Top-Status-Bar组件规范.md)
> 目标：定义 `Top Status Bar` 最终 props。

## 1. Props

```ts
type TopStatusBarProps = {
  snapshot: ProjectSnapshot
  onOpenPlan?: () => void
  onOpenRun?: () => void
  onOpenAcceptance?: () => void
}
```

## 2. `ProjectSnapshot`

```ts
type ProjectSnapshot = {
  project_id: string
  name: string
  project_category: string
  current_stage: string
  current_run_status?: string
  production_status: string
  progress_percent: number
  final_status_hint?: string
}
```

## 3. 口径补充

按当前钱学森总纲，`ProjectSnapshot` 后续建议逐步补齐或等价承载以下字段：

1. `manual_checkpoint_required`
2. `has_runtime_escalation`
3. `decision`
4. `completed`

原因：

- 顶部状态条不能只展示 readiness 摘要
- 它至少要有能力提示人工检查点、升级对象和最终完成裁决
