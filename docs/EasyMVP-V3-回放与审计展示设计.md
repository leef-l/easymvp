# EasyMVP V3 回放与审计展示设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 目标：定义 V3 中 replay、审计记录和页面展示的关系，避免回放能力只停留在底层 runtime。

## 1. 设计结论

V3 的 replay 不是底层调试工具，而应成为：

1. 工作台可追溯入口
2. 验收证据辅助入口
3. 审计与复盘入口

## 2. 展示入口

建议从以下地方可进入回放：

1. Live Activity 事件
2. run detail drawer
3. Acceptance evidence

## 3. 展示层次

建议分三层：

1. 摘要层
2. 结构化事件层
3. 原始回放层

## 4. 审计内容

建议至少保留：

1. who
2. when
3. what run
4. what action
5. what result

## 5. 后续细分专题

本专题后续继续拆：

1. Replay drawer 设计
2. 审计过滤器设计
3. 验收证据与 replay 关联设计
