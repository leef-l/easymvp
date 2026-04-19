# EasyMVP V3 Workspace 首页 Stage Overview 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页聚合接口Schema设计](./EasyMVP-V3-Workspace首页聚合接口Schema设计.md)
> 目标：定义首页 `Stage Overview` 阶段分布带的结构、状态聚合和筛选行为。

## 1. 设计结论

`Stage Overview` 是首页的全局分布带，不是导航菜单。

它用来帮助用户快速判断：

1. 项目大都卡在哪个阶段
2. 哪个阶段阻塞最多

## 2. 组件结构

建议每个阶段桶展示：

1. `stage_name`
2. `project_count`
3. `blocked_count`
4. `active_count`

## 3. 阶段

固定为：

1. `Design`
2. `Review`
3. `Compile`
4. `Execute`
5. `Acceptance`
6. `Complete`

## 4. 交互

点击阶段桶时：

1. 筛选 `Running Projects`
2. 同步筛选 `Recent Activity`

## 5. 视觉规则

建议：

1. 横向带状布局
2. 阶段名为主
3. 数量为辅
4. 有阻塞时显示次级告警点

## 6. 不该怎么做

不应该：

1. 做成表格
2. 只显示总数不显示阶段名
3. 点击后没有联动

## 7. 后续细分专题

1. 阶段分布带视觉稿
2. 筛选联动规则

