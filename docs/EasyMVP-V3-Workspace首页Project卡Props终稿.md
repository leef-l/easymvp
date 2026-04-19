# EasyMVP V3 Workspace 首页 Project 卡 Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页多项目卡片组件规范](./EasyMVP-V3-Workspace首页多项目卡片组件规范.md)
> 目标：定义 Workspace 首页单个项目卡最终 props。

## 1. Props

```ts
type ProjectCardProps = {
  project: ProjectCard
  onOpenProject: (projectId: string) => void
}
```

## 2. `ProjectCard`

```ts
type ProjectCard = {
  project_id: string
  name: string
  project_category: string
  stage: string
  progress_percent: number
  production_status: string
  blocker_count: number
  waiting_action_count: number
  updated_at: string
}
```

## 3. 口径补充

按当前钱学森总纲，这个 props 结构可以继续保留为首页壳层首版现实字段。

但后续建议逐步补齐或等价承载：

1. `decision`
2. `completed`
3. `manual_checkpoint_required`
4. `has_runtime_escalation`

原因：

- `production_status` 不应再被首页卡误读为最终完成状态
- 首页卡至少要有能力提示当前是否存在人工检查点和升级对象
