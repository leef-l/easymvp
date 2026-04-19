# EasyMVP V3 Coverage Matrix 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-Acceptance线框图设计](./EasyMVP-V3-Acceptance线框图设计.md)
> 目标：定义 `Acceptance` 页面中 `Coverage Matrix` 的结构、状态、交互和文案规则。

## 1. 设计结论

`Coverage Matrix` 是 `Acceptance` 页的主视觉组件。

它必须一眼表达：

1. 哪些 `surface` 已覆盖
2. 哪些 `journey` 仍缺失
3. 哪些证据还不完整
4. 哪些格子已经阻塞生产级通过

## 2. 结构

建议：

1. 行为 `surface`
2. 列为 `journey`
3. 单元格显示覆盖状态
4. 边侧补充 evidence 数量与 blocker 数量

## 3. 状态

建议固定：

1. `pass`
2. `partial`
3. `missing`
4. `blocked`

## 4. 交互

点击单元格时：

1. 打开对应 evidence 列表
2. 高亮相关 blocker
3. 可跳到对应 replay / preview

## 5. 文案

格子内不放长文本，只放：

1. 状态颜色
2. 简短标签
3. evidence 数量

## 6. 不该怎么做

不应该：

1. 只做纯颜色表格
2. 看不到缺的是 journey 还是 evidence
3. 只能看不能点

## 7. 后续细分专题

1. Coverage Matrix 视觉稿
2. Coverage Matrix 响应式折叠规则

