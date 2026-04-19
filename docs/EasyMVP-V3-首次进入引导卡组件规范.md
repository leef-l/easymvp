# EasyMVP V3 首次进入引导卡组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建后首次进入Project-Workspace引导态设计](./EasyMVP-V3-创建后首次进入Project-Workspace引导态设计.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 目标：定义新项目首次进入 `Project Workspace` 时顶部启动引导卡的内容结构、文案层级和动作规则。

## 1. 设计结论

首次进入引导卡必须是解释卡，不是欢迎横幅。

它要解决：

1. 项目已经创建成功
2. 系统正在准备什么
3. 用户下一步该点哪里

## 2. 组件结构

建议固定四段：

1. 标题区
2. 完成事项区
3. 当前准备区
4. 动作区

## 3. 线框

```text
┌──────────────────────────────────────────────────────────────┐
│ Your project is ready                                        │
│ EasyMVP is preparing the first review cycle                  │
│ ✓ Project created  ✓ Workspace bound  ✓ Category matched     │
│ Now preparing: first review/compile cycle                    │
│ [Open Plan] [View Init Events]                               │
└──────────────────────────────────────────────────────────────┘
```

## 4. 文案规则

建议：

1. 标题短
2. 副标题解释当前阶段
3. 已完成项用勾选短标签

按当前钱学森总纲，这张卡服务的是早期 `reviewing` 引导态，不应继续把用户心智拉回旧 `Design` 阶段。

## 5. 动作规则

主动作固定：

1. `Open Plan`

次动作最多一个：

1. `View Init Events`

## 6. 状态变化

### 6.1 plan ready

突出 `Open Plan`

### 6.2 plan not ready

主动作改为：

1. `Stay Here`

## 7. 不该怎么做

不应该：

1. 做成长段说明
2. 放 4 个以上按钮
3. 只显示“Welcome”

## 8. 后续细分专题

1. 引导卡视觉稿
2. 引导卡状态映射表
