# EasyMVP V3 Coverage Matrix Props 终稿

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Coverage-Matrix组件规范](./EasyMVP-V3-Coverage-Matrix组件规范.md)
> 目标：定义 `Coverage Matrix` 最终 props。

## 1. Props

```ts
type CoverageMatrixProps = {
  items: CoverageMatrixItem[]
  onSelect?: (item: CoverageMatrixItem) => void
}
```

## 2. `CoverageMatrixItem`

```ts
type CoverageMatrixItem = {
  key: string
  kind: "surface" | "journey"
  name: string
  coverage_status: "pass" | "partial" | "missing"
  evidence_count: number
}
```

