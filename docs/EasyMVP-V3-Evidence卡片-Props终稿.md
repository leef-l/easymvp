# EasyMVP V3 Evidence 卡片 Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence卡片组件规范.md)
> 目标：定义 Evidence 单卡最终 props。

## 1. Props

```ts
type EvidenceCardProps = {
  card: EvidenceCardView
  onOpenEvidence: (id: string) => void
}
```

## 2. `EvidenceCardView`

```ts
type EvidenceCardView = {
  id: string
  surface: string
  journey?: string
  evidence_type: string
  file_path: string
  captured_at: string
}
```

