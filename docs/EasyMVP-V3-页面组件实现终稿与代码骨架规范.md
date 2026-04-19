# EasyMVP V3 页面组件实现终稿与代码骨架规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 目标：把 V3 页面继续细化到组件树、props 契约、状态归属和代码文件落点的终稿层。

## 1. 设计结论

这份文档不是视觉设计，而是前端实现终稿。

它要回答：

1. 页面拆成哪些组件
2. 每个组件吃什么 props
3. 数据从哪里来
4. 哪些状态放页面，哪些状态放组件

## 2. 推荐目录终稿

```text
apps/desktop/src/renderer/
  app/
  modules/
    workspace/
      pages/
      components/
      hooks/
    plan/
      pages/
      components/
      hooks/
    acceptance/
      pages/
      components/
      hooks/
    settings/
      pages/
      components/
  shared/
    contracts/
    api/
    ui/
```

## 3. Workspace Home 组件树

```text
WorkspaceHomePage
  ├─ WorkspaceTopBar
  ├─ StageOverviewCard
  ├─ NeedAttentionCard
  ├─ RecentActivityCard
  ├─ ReleaseReadinessCard
  └─ ProjectGrid
       └─ ProjectCard
```

### 3.1 `WorkspaceHomePage` Props

页面容器不吃 props，直接调 query hook：

1. `useWorkspaceHomeView()`

### 3.2 `ProjectGrid` Props

```ts
type ProjectGridProps = {
  projects: ProjectCard[]
  onOpenProject: (projectId: string) => void
}
```

### 3.3 `NeedAttentionCard` Props

```ts
type NeedAttentionCardProps = {
  items: NeedAttentionItem[]
  onOpenItem: (id: string) => void
}
```

## 4. Project Workspace 组件树

```text
ProjectWorkspacePage
  ├─ TopStatusBar
  ├─ StageRail
  ├─ LiveActivityPanel
  ├─ ActionInboxPanel
  └─ CoveragePanel
```

### 4.1 `TopStatusBar` Props

```ts
type TopStatusBarProps = {
  snapshot: ProjectSnapshot
}
```

### 4.2 `StageRail` Props

```ts
type StageRailProps = {
  stages: StageProgress[]
  onSelectStage?: (stage: string) => void
}
```

### 4.3 `ActionInboxPanel` Props

```ts
type ActionInboxPanelProps = {
  items: ActionInboxItem[]
  onAction: (item: ActionInboxItem) => void
}
```

## 5. Plan 页面组件树

```text
PlanPage
  ├─ PlanHeader
  ├─ DraftCard
  ├─ ReviewCard
  ├─ CompiledCard
  ├─ PlanDiffPanel
  └─ TaskProjectionList
       └─ TaskProjectionItem
```

### 5.1 `TaskProjectionList` Props

```ts
type TaskProjectionListProps = {
  tasks: CompiledTaskView[]
  onOpenTask: (taskId: string) => void
}
```

## 6. Acceptance 页面组件树

```text
AcceptancePage
  ├─ AcceptanceHeader
  ├─ CoverageMatrix
  ├─ IssuesPanel
  ├─ EvidenceCardList
  └─ ReleaseGatePanel
```

### 6.1 `CoverageMatrix` Props

```ts
type CoverageMatrixProps = {
  items: CoverageMatrixItem[]
  onSelect?: (item: CoverageMatrixItem) => void
}
```

### 6.2 `EvidenceCardList` Props

```ts
type EvidenceCardListProps = {
  cards: EvidenceCardView[]
  onOpenEvidence: (id: string) => void
}
```

## 7. Hooks 终稿建议

### 7.1 Workspace

1. `useWorkspaceHomeView`
2. `useWorkspaceEvents`

### 7.2 Project Workspace

1. `useProjectWorkspaceView`
2. `useProjectEvents`
3. `useActionInboxActions`

### 7.3 Plan

1. `usePlanView`
2. `useCompilePlan`

### 7.4 Acceptance

1. `useAcceptanceView`
2. `useStartAcceptance`
3. `useManualDecision`

## 8. 状态归属规则

### 页面级状态

适合放页面：

1. 当前选中的项目
2. 当前打开的抽屉 id
3. 页面级 filter / tab

### 组件级状态

适合放组件：

1. hover
2. 展开折叠
3. 本地动画控制

### 不应放前端本地状态的内容

1. 项目主状态
2. 任务主状态
3. 验收最终状态

这些必须来自后端 query 或事件流。

## 9. 建议代码骨架

### Workspace Home Page

```tsx
export function WorkspaceHomePage() {
  const { data, isLoading } = useWorkspaceHomeView()

  if (isLoading) return <WorkspaceLoadingState />

  return (
    <WorkspaceHomeLayout>
      <StageOverviewCard summary={data.summary} />
      <NeedAttentionCard items={data.need_attention} onOpenItem={() => {}} />
      <RecentActivityCard items={data.recent_activity} />
      <ReleaseReadinessCard items={data.release_readiness} />
      <ProjectGrid projects={data.active_projects} onOpenProject={() => {}} />
    </WorkspaceHomeLayout>
  )
}
```

### Project Workspace Page

```tsx
export function ProjectWorkspacePage() {
  const { projectId } = useProjectRouteParams()
  const { data, isLoading } = useProjectWorkspaceView(projectId)

  if (isLoading) return <ProjectWorkspaceLoadingState />

  return (
    <ProjectWorkspaceLayout>
      <TopStatusBar snapshot={data.project_snapshot} />
      <StageRail stages={data.stage_progress} />
      <LiveActivityPanel items={data.live_activity} />
      <ActionInboxPanel items={data.action_inbox} onAction={() => {}} />
      <CoveragePanel coverage={data.acceptance_coverage} />
    </ProjectWorkspaceLayout>
  )
}
```

## 10. 实现顺序

建议前端按这个顺序实现：

1. 页面骨架与路由
2. Query hooks
3. Workspace Home
4. Project Workspace
5. Plan
6. Acceptance
7. 抽屉与详情

## 11. 后续细分专题

1. 每个组件的 className / style contract
2. 页面 loading / empty / error state 终稿
3. 事件流更新策略

