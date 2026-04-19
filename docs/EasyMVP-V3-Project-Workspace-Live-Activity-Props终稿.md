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

## 2. 口径补充

按当前总纲，`LiveActivityItem` 在实际落地时建议至少能承载以下事件分型：

1. `verification_conflict_found`
2. `runtime_escalation_raised`
3. `fault_loop_detected`

原因：

- 这些事件不能被页面层压平成普通失败
- 组件即使不直接暴露完整对象，也应有能力区分这类高优先级活动
