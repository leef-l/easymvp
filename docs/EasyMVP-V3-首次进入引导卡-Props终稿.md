# EasyMVP V3 首次进入引导卡 Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-首次进入引导卡组件规范](./EasyMVP-V3-首次进入引导卡组件规范.md)
> 目标：定义首次进入引导卡最终 props。

## 1. Props

```ts
type FirstEntryGuideCardProps = {
  title: string
  summary: string
  action_label: string
  onAction: () => void
}
```
