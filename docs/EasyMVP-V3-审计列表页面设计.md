# EasyMVP V3 审计列表页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-审计查询接口设计](./EasyMVP-V3-审计查询接口设计.md)
> 关联文档：[EasyMVP-V3-审计过滤器设计](./EasyMVP-V3-审计过滤器设计.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer页面设计.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 目标：把审计能力落成可用页面，定义审计列表页的布局、过滤器位置、结果列表模型、交互、排序与和 Replay Drawer 的联动方式。

## 1. 设计结论

审计列表页不应做成通用后台表格页。

它的正确产品形态应该是：

> 一个围绕问题收敛、上下文钻取和回放联动设计的审计工作面板。

一句话：

> 审计列表页的任务不是“展示很多记录”，而是“让用户最快找到值得看的记录并进入回放上下文”。

## 2. 页面定位

审计列表页主要回答：

1. 最近哪些运行最值得关注
2. 哪类记录最可能解释当前问题
3. 当前过滤条件下，哪些记录可以直接进入回放

## 3. 页面布局

建议分成四块：

1. 顶部页面概览条
2. 左侧或顶部固定过滤器区
3. 中部结果列表区
4. 右侧详情区或 Replay 快捷入口区

推荐布局：

```text
┌──────────────────────────────────────────────────────────────┐
│ Audit Summary / Active Filters                              │
├───────────────────────┬──────────────────────────────────────┤
│ Filter Panel          │ Audit Record List                    │
│                       │                                      │
├───────────────────────┴──────────────────────────────────────┤
│ Detail / Replay Quick Preview                              │
└──────────────────────────────────────────────────────────────┘
```

## 4. 顶部概览条

### 4.1 目标

让用户一眼看清当前筛选上下文和结果规模。

### 4.2 建议展示

1. 当前 `project`
2. 当前时间范围
3. 当前结果数量
4. 当前高严重级记录数量
5. 当前 `run_id / trace_id / evidence_id` 等关键过滤条件

### 4.3 Chips 区

所有生效过滤条件必须显示为可移除 chips，避免用户迷失在复杂筛选状态里。

## 5. 过滤器区

### 5.1 固定可见项

建议固定可见：

1. `run_id`
2. `severity`
3. `event_type`
4. `trace_id`
5. `time_range`

### 5.2 高级过滤项

折叠区建议放：

1. `brain_kind`
2. `record_kind`
3. `artifact_state`
4. `evidence_id`
5. `surface`
6. `journey_id`
7. `domain_task_id`
8. `compiled_task_id`

### 5.3 交互规则

建议：

1. 修改筛选后自动刷新列表
2. 支持“清空全部”
3. 支持“仅看异常”
4. 支持“仅看可回放项”

## 6. 结果列表区

### 6.1 数据模型

结果列表只消费 `AuditRecord`。

每项至少展示：

1. `record_kind`
2. `summary`
3. `severity`
4. `brain_kind`
5. `run_id`
6. `event_id / trace_id`
7. `artifact_state`
8. `created_at`
9. `replay_target`

### 6.2 视觉层级

建议：

1. `severity` 用明显但克制的色彩标识
2. `record_kind` 用小标签
3. `artifact_state` 单独显示，不埋进摘要里
4. 当前选中项高亮

### 6.3 排序控件

建议支持：

1. `最相关`
2. `最新优先`
3. `严重级优先`

## 7. 详情区

### 7.1 目标

显示当前选中 `AuditRecord` 的更多上下文，而不要求用户立刻进入 Replay Drawer。

### 7.2 建议内容

1. `record_id`
2. `record_kind`
3. `summary`
4. `run_id`
5. `event_id`
6. `trace_id / span_id`
7. `source_object_kind / source_object_id`
8. `artifact_state`
9. `replay_target`
10. `evidence_target`

### 7.3 快捷动作

允许：

1. 打开 Replay Drawer
2. 跳转 Evidence
3. 复制 `run_id / trace_id / event_id`

## 8. 与 Replay Drawer 的联动

### 8.1 主交互

如果当前记录包含 `replay_target`：

1. 结果项上显示“查看回放”
2. 点击后直接打开 `Replay Drawer`
3. Drawer 自动定位到对应上下文

### 8.2 回链

从 Replay Drawer 返回时：

1. 保持当前筛选条件
2. 保持当前选中记录
3. 保持滚动位置

## 9. 默认状态设计

### 9.1 无过滤的初始态

建议：

1. 默认按 `created_at_desc`
2. 只显示最近一段时间
3. 高严重级记录优先视觉高亮

### 9.2 带上下文打开

如果从 Workspace / Evidence / Replay 打开：

1. 自动带入上下文过滤条件
2. 页面标题旁显示“来自某入口”的上下文说明

### 9.3 无结果态

如果筛选后无结果：

1. 清楚显示“当前条件下无记录”
2. 提供“一键清空部分筛选”的动作

## 10. 状态与降级

### 10.1 `artifact_pruned`

表现：

1. 列表项可见
2. 详情区提示“原始材料已清理”
3. 仍允许打开 Replay Drawer 摘要层

### 10.2 `artifact_missing`

表现：

1. 列表项 warning 高亮
2. 详情区提示“索引存在但文件缺失”

## 11. 不该怎么做

不建议：

1. 做成纯表格后台页
2. 只给全文搜索不给结构化筛选
3. 列表项无法直接进入 Replay Drawer
4. 不显示当前已生效过滤条件
5. 切换 Drawer 后丢失列表上下文

## 12. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. 审计查询接口必须服务本页所需字段
2. Replay Drawer 与审计页必须双向联动
3. 任何新的审计记录类型都应兼容 `AuditRecord` 视图模型

## 13. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-审计列表线框图设计](./EasyMVP-V3-审计列表线框图设计.md)
2. 审计 chips 与预设保存交互设计
3. 审计导出面板设计
