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

