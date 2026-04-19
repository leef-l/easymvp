# EasyMVP V3 页面 Loading / Empty / Error / Recovery 状态终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-页面组件实现终稿与代码骨架规范](./EasyMVP-V3-页面组件实现终稿与代码骨架规范.md)
> 关联文档：[EasyMVP-V3-恢复模式与诊断模式页面设计](./EasyMVP-V3-恢复模式与诊断模式页面设计.md)
> 目标：为 V3 页面补齐实现最容易遗漏的状态层终稿，包括 loading、empty、error 和 recovery。

## 1. 设计结论

所有主页面都必须有 4 种基础状态：

1. `loading`
2. `empty`
3. `error`
4. `recovery`

不能只写“正常态”。

## 2. Workspace Home

### 2.1 Loading

显示：

1. 顶部骨架条
2. 4 张 summary skeleton card
3. project card skeleton grid

### 2.2 Empty

条件：

1. 无项目

显示：

1. 空态说明
2. `Create Project` 主按钮

### 2.3 Error

条件：

1. 首页 query 失败

显示：

1. 错误摘要
2. `Retry` 按钮
3. `Open Diagnostics` 次按钮

## 3. Project Workspace

### 3.1 Loading

显示：

1. top status skeleton
2. stage rail skeleton
3. activity panel skeleton
4. inbox panel skeleton

### 3.2 Empty

条件：

1. 项目刚创建但尚无计划或运行

显示：

1. 引导说明
2. 推荐下一步

### 3.3 Error

条件：

1. 项目详情 query 失败

显示：

1. 项目级错误卡片
2. `Retry`
3. `Back to Workspace`

### 3.4 Recovery

条件：

1. 项目视图 stale 且核心服务异常

显示：

1. stale 提示
2. `Reconnect`
3. `Open Diagnostics`

## 4. Plan 页面

### 4.1 Empty

条件：

1. 没有 `PlanDraft`

显示：

1. `No plan yet`
2. 引导动作

### 4.2 Error

条件：

1. plan query 失败

显示：

1. 计划加载失败说明
2. `Retry`

## 5. Acceptance 页面

### 5.1 Empty

条件：

1. 尚未启动验收

显示：

1. 当前还未进入验收
2. `Start Acceptance`

### 5.2 Error

条件：

1. acceptance query 失败

显示：

1. 错误卡片
2. `Retry`
3. `Open Diagnostics`

### 5.3 Recovery

条件：

1. `brain serve` 或 evidence 索引异常

显示：

1. 覆盖可能过期提示
2. `Refresh Coverage`
3. `Open Diagnostics`

## 6. 通用组件建议

建议统一提供：

1. `PageLoadingState`
2. `PageEmptyState`
3. `PageErrorState`
4. `PageRecoveryState`

## 7. Props 终稿

```ts
type PageErrorStateProps = {
  title: string
  summary: string
  retryLabel?: string
  onRetry?: () => void
  secondaryLabel?: string
  onSecondary?: () => void
}

type PageEmptyStateProps = {
  title: string
  summary: string
  actionLabel?: string
  onAction?: () => void
}
```

## 8. 实施顺序

1. 先写通用状态组件
2. 再接 Workspace Home
3. 再接 Project Workspace
4. 再接 Plan / Acceptance

