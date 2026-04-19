# EasyMVP V3 Workspace 首页多项目卡片组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页线框图设计](./EasyMVP-V3-Workspace首页线框图设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：定义 `Workspace Home` 中 `Running Projects` 区块的项目卡组件结构、字段优先级、状态样式与交互行为。

## 1. 设计结论

首页项目卡不是项目摘要名片。

它必须一眼回答：

1. 这个项目现在在哪个阶段
2. 当前在干什么
3. 有没有卡住
4. 离交付还有多远

所以首页项目卡的本质是：

> 多项目实时驾驶舱中的“单项目状态卡”。

按当前钱学森总纲，这张卡只承担壳层概览，不承担最终裁决解释职责。

## 2. 组件目标

项目卡必须做到：

1. 首屏可扫读
2. 关键风险前置
3. 下一步入口明确
4. 不退化为项目详情缩略版

## 3. 组件结构

建议组件固定分为 5 个区域：

1. 头部信息区
2. 阶段与进度区
3. 当前活动区
4. 风险与待处理区
5. 底部动作区

## 4. 线框结构

```text
┌──────────────────────────────────────────┐
│ Project Name                Category     │
│ Stage Badge                 62%          │
│ Active: compiling task plan             │
│ Blockers: 1   Need Action: Yes          │
│ Completion: blocked / near-ready        │
│ [Open Project] [Acceptance]             │
└──────────────────────────────────────────┘
```

## 5. 字段优先级

卡片上字段优先级建议固定为：

1. `project_name`
2. `current_stage`
3. `active_task_or_run`
4. `blocker_count`
5. `waiting_action`
6. `production_readiness`
7. `project_category`
8. `overall_progress`

如果聚合层已提供更新后的裁决字段，建议优先过渡为：

1. `project_name`
2. `current_stage`
3. `active_task_or_run`
4. `blocker_count`
5. `waiting_action`
6. `decision`
7. `completed`
8. `project_category`

原因：

1. 用户先看项目是谁
2. 再看项目现在在哪
3. 再看项目正在做什么
4. 最后才看更多上下文

## 6. 头部信息区

建议放：

1. `project_name`
2. `project_category`

规则：

1. 项目名最多两行
2. 分类标签使用短标签
3. 不要把项目 ID 暴露在首屏

## 7. 阶段与进度区

建议放：

1. `current_stage`
2. `overall_progress`

### 7.1 阶段表现

建议使用短 badge：

1. `reviewing`
2. `executing`
3. `accepting`
4. `reworking`
5. `completed`

### 7.2 进度表现

建议只显示一个简洁百分比或短条，不做复杂环图。

## 8. 当前活动区

这是卡片的核心行。

建议展示：

1. `Active: <current task or run summary>`

例如：

1. `Active: reviewing plan scope`
2. `Active: browser acceptance run`
3. `Active: preparing project`

必须是自然语言摘要，不是内部事件码。

## 9. 风险与待处理区

建议固定展示两类信号：

1. `blocker_count`
2. `waiting_action`

### 9.1 blocker 表达

建议：

1. `Blockers: 0`
2. `Blockers: 2`

### 9.2 waiting action 表达

建议：

1. `Need Action: No`
2. `Need Action: Yes`

### 9.3 高风险视觉

如果 `blocker_count > 0`，卡片边缘或标题区应出现明显告警态。

## 10. Readiness 区

建议使用短句表现：

1. `Readiness: not started`
2. `Readiness: partial`
3. `Readiness: near-ready`
4. `Readiness: completed`

不要直接暴露一堆内部布尔字段。

如果页面已经接入最新字段，推荐优先展示：

1. `Decision: blocked / rework / manual checkpoint / complete`
2. `Completed: yes / not yet`

## 11. 动作区

建议固定两个主动作：

1. `Open Project`
2. `Acceptance`

可选次动作：

1. `View Blockers`

但不要让卡片底部出现 4-5 个按钮。

## 12. 组件状态

建议至少定义：

1. `running`
2. `blocked`
3. `waiting_action`
4. `near_ready`
5. `completed`

### 12.1 `running`

显示正常活动态。

### 12.2 `blocked`

显示告警边框或告警角标。

### 12.3 `waiting_action`

显示轻提醒，但不一定是告警色。

### 12.4 `near_ready`

显示更积极的 readiness 提示。

### 12.5 `completed`

整体弱化，不应抢占进行中项目的注意力。

## 13. 排序规则

首页项目卡默认排序建议：

1. 有 blocker 的优先
2. 有待处理动作的次优先
3. 正在运行的再次之
4. 即将交付的接着
5. 已完成的最后

## 14. 点击行为

### 14.1 点击卡片主体

进入 `Project Workspace`。

### 14.2 点击 `Acceptance`

进入项目 `Acceptance` 页。

### 14.3 点击 blocker 区

优先打开项目中的 `Action Inbox` 或 blocker 详情。

若当前卡点来自 `verification_conflict / fault_loop_detected / manual_checkpoint`，应优先深链到能解释结构化原因的页面，而不是只打开普通活动流。

## 15. 空态与长文本规则

### 15.1 无当前活动

显示：

1. `Active: waiting for the next step`

### 15.2 项目名过长

允许两行截断。

### 15.3 摘要过长

当前活动摘要最多一行半。

## 16. 视觉规则

建议：

1. 卡片大圆角但克制
2. 轻阴影
3. 状态色集中在 badge 和边缘提示
4. 大留白，不堆字段

不要：

1. 把所有指标做成彩色徽章
2. 把卡片做成小型仪表盘
3. 叠太多细粒度数字

## 17. 不该怎么做

项目卡不应该：

1. 只有项目名和进度条
2. 没有当前活动摘要
3. blocker 信息要点进去才看到
4. 动作区按钮太多
5. 暴露过多内部术语
6. 用 `production passed` 直接替代完成状态

## 18. 后续细分专题

本专题后续继续拆：

1. 多项目卡片视觉稿
2. 多项目卡片状态映射表
3. 多项目卡片数据接口字段约束
