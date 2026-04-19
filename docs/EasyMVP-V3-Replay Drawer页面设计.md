# EasyMVP V3 Replay Drawer 页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3-Replay索引表结构设计](./EasyMVP-V3-Replay索引表结构设计.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 目标：把 Replay 能力落成可用页面，定义 Replay Drawer 的入口、布局、分层展示、状态、加载策略与交互规则。

## 1. 设计结论

Replay 不应跳出到一个笨重的新页面当日志查看器。

第一版更合适的产品形态是：

> 在 Workspace、Acceptance、Run Detail 等入口上，以 `Replay Drawer` 形式打开结构化回放视图。

这样做的原因：

1. 用户大部分时候是在排查一个具体问题
2. 回放是上下文能力，不是主导航页面
3. Drawer 更适合从事件、证据、run 详情深链进入

## 2. 页面定位

`Replay Drawer` 的任务不是展示所有原始日志。

它必须回答 4 个问题：

1. 这次 run 发生了什么
2. 哪些事件最关键
3. 哪些 replay 切片可以解释问题
4. 如果需要，我如何看到原始材料

## 3. 打开入口

建议至少支持以下入口：

1. `Project Workspace` 的 Live Activity 事件
2. `Project Workspace` 的 active run 状态入口
3. `Acceptance` 页中的 Evidence 卡
4. `Run Detail` 抽屉中的“查看回放”

## 4. 抽屉形式

建议使用右侧大抽屉，而不是居中 modal。

原因：

1. 需要较大阅读空间
2. 需要保留原页面上下文
3. 便于在事件、证据、任务间来回比对

### 4.1 尺寸建议

桌面端建议：

1. 默认宽度 `56% - 68%`
2. 最小宽度不低于可读时间线需求

移动端建议：

1. 全屏抽屉

## 5. 顶层布局

建议分成四块：

1. 顶部摘要条
2. 左侧结构化时间线
3. 中部详情面板
4. 底部或右下原始材料区

推荐结构：

```text
┌──────────────────────────────────────────────────────────────┐
│ Replay Summary Bar                                          │
├───────────────────────┬──────────────────────────────────────┤
│ Timeline              │ Detail Panel                         │
│                       │                                      │
├───────────────────────┴──────────────────────────────────────┤
│ Raw Content / Log Segment Preview                           │
└──────────────────────────────────────────────────────────────┘
```

## 6. 顶部摘要条

### 6.1 目标

让用户在几秒内知道当前看的是哪次 run、处于什么状态、还剩多少原始材料可读。

### 6.2 字段建议

至少展示：

1. `run_id`
2. `brain_kind`
3. `status`
4. `started_at`
5. `ended_at`
6. `event_count`
7. `replay_count`
8. `log_segment_count`
9. `artifact_status_summary`

### 6.3 快捷动作

建议允许：

1. 复制 `run_id`
2. 跳转 run detail
3. 刷新 replay
4. 打开审计视图

## 7. 时间线区域

### 7.1 目标

让用户优先看“关键切片序列”，而不是直接陷入原始日志。

### 7.2 数据来源

优先来自：

1. `replay-timeline`
2. `workflow_brain_run_events`

### 7.3 每条时间线项建议展示

1. `seq_no`
2. `replay_kind`
3. `title`
4. `summary`
5. `event_id`
6. `trace_id`
7. `status`
8. `preview_available`

### 7.4 排序规则

建议：

1. 默认按 `seq_no asc`
2. 支持切到按 `created_at desc`
3. 高风险或失败相关项允许高亮，不改变主排序

### 7.5 筛选

建议支持：

1. `replay_kind`
2. `event_id`
3. `trace_id`
4. `status`

## 8. 详情面板

### 8.1 目标

显示当前选中 replay 项的结构化信息。

### 8.2 展示字段

至少展示：

1. `title`
2. `summary`
3. `source_object_kind`
4. `source_object_id`
5. `event_id`
6. `trace_id`
7. `span_id`
8. `status`
9. `related_log_segments`

### 8.3 交互

允许：

1. 跳转到对应 event
2. 跳转到对应 task
3. 跳转到对应 evidence
4. 展开原始内容

## 9. 原始材料区

### 9.1 目标

在必要时给用户看原始回放或日志，但不让它占据默认焦点。

### 9.2 展示对象

支持：

1. 原始 replay 内容
2. 原始日志分片
3. 材料缺失/已清理提示

### 9.3 加载规则

默认不加载原始内容。

只有当用户：

1. 点击“查看原始回放”
2. 点击“查看日志分片”

时，才调用：

1. `replay-items/{replay_id}/raw`
2. `log-segments/{segment_id}/raw`

### 9.4 大内容处理

建议：

1. 默认截断显示
2. 标出 `truncated = true`
3. 允许继续加载更多

## 10. Evidence 回链模式

如果从 Evidence 卡进入 Replay Drawer：

1. 顶部摘要条仍展示 run 级摘要
2. 时间线默认定位到对应 `event_id` 或 `trace_id`
3. 当前关联的 Evidence 在详情区单独高亮

## 11. Workspace 回链模式

如果从 Workspace Live Activity 进入：

1. 时间线默认定位到对应 event
2. 抽屉打开时自动滚动到相关切片
3. 当前事件上下文保持高亮 1 段时间

## 12. 状态设计

### 12.1 材料完整可用

特征：

1. `artifact_state = available`

表现：

1. 正常展示预览入口

### 12.2 材料已清理

特征：

1. `artifact_state = artifact_pruned`

表现：

1. 仍展示结构化摘要
2. 原始内容区显示“原始材料已清理”

### 12.3 材料缺失

特征：

1. `artifact_state = artifact_missing`

表现：

1. 以 warning 方式提示
2. 不让整页崩掉

## 13. 加载策略

建议分阶段加载：

### 首屏加载

1. `replay-summary`
2. `replay-timeline` 首屏第一页

### 选中项加载

1. `replay-detail`

### 展开原始内容时加载

1. `replay-raw`
2. `log-segment-raw`

这样可以避免抽屉一打开就读取大量原始材料。

## 14. 不该怎么做

不建议：

1. 把 Replay Drawer 做成纯日志窗口
2. 抽屉首屏直接加载所有 raw 内容
3. 文件缺失时整个 Drawer 报错
4. 不保留与 event / trace / evidence 的跳转关系

## 15. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Workspace 中的 replay 入口应默认打开此 Drawer
2. Acceptance Evidence 的 replay 深链应直接定位到相关时间线项
3. 审计过滤器若展开 run 细节，应复用此 Drawer 的时间线和详情结构

## 16. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Replay-Drawer线框图设计](./EasyMVP-V3-Replay-Drawer线框图设计.md)
2. 审计过滤器与 Replay 联动设计
3. 原始材料按需加载交互细节
