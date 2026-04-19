# EasyMVP V3 Evidence 详情抽屉设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence卡片组件规范.md)
> 关联文档：[EasyMVP-V3-Evidence Preview交互设计](./EasyMVP-V3-Evidence Preview交互设计.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay Drawer页面设计.md)
> 目标：定义从 Evidence 卡片进入的详情抽屉布局、字段和与 Preview/Replay 的联动规则。

## 1. 设计结论

`Evidence` 详情抽屉负责展示“这条证据到底是什么、来自哪里、是否可靠”。

## 2. 结构

建议分四段：

1. 头部摘要
2. 文件与来源信息
3. 预览区
4. replay / validation / links

## 3. 头部字段

建议展示：

1. `evidence_id`
2. `evidence_type`
3. `status`
4. `collected_at`

## 4. 预览联动

抽屉内直接复用 `Evidence Preview` 结构。

## 5. Replay 联动

若存在 replay 链接，展示：

1. `Open Replay`
2. `Open Source Event`

## 6. 不该怎么做

不应该：

1. 详情抽屉里只重复卡片内容
2. 看不到来源和校验状态
3. 不能打开 replay

## 7. 后续细分专题

1. Evidence 详情抽屉线框图
2. Evidence 来源映射规则

