# EasyMVP V3 审计过滤器设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer页面设计.md)
> 关联文档：[EasyMVP-V3-Replay索引表结构设计](./EasyMVP-V3-Replay索引表结构设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-审计查询接口设计](./EasyMVP-V3-%E5%AE%A1%E8%AE%A1%E6%9F%A5%E8%AF%A2%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md)
> 目标：定义 V3 中审计视图的过滤维度、筛选模型、默认视角和与 Replay Drawer 的联动方式，支撑问题定位、运行复盘和证据追溯。

## 1. 设计结论

V3 的审计视图不应只是一个“大列表 + 搜索框”。

正确做法是：

1. 提供结构化过滤器
2. 让过滤器和 `run / event / trace / evidence` 关联
3. 支持一键收窄到可解释的问题上下文
4. 与 Replay Drawer 互相跳转

一句话：

> 审计过滤器的任务不是“找一条日志”，而是“把问题快速收敛到一个可解释的运行上下文”。

## 2. 审计过滤器要解决的问题

本专题主要解决：

1. run 太多时难以定位问题
2. 同一问题分散在 event、trace、evidence、replay 中
3. 用户只能靠关键词碰运气搜索
4. 审计和回放割裂，无法快速上下钻取

## 3. 过滤器设计原则

建议遵守：

1. 先缩小上下文，再看明细
2. 高频过滤项固定可见
3. 低频过滤项折叠到高级筛选
4. 过滤结果默认按“最值得看”排序
5. 任意过滤结果都能进入 Replay Drawer

## 4. 过滤维度

### 4.1 高频过滤维度

建议固定展示：

1. `project_id`
2. `run_id`
3. `status`
4. `event_type`
5. `severity`
6. `trace_id`
7. `time_range`

### 4.2 中频过滤维度

建议放进展开区：

1. `brain_kind`
2. `domain_task_id`
3. `compiled_task_id`
4. `replay_kind`
5. `stream_kind`
6. `artifact_state`

### 4.3 Evidence 关联维度

建议支持：

1. `evidence_id`
2. `evidence_type`
3. `surface`
4. `journey_id`

## 5. 默认视角

不同入口打开审计视图时，默认过滤器应不同。

### 5.1 从 Workspace 打开

默认带入：

1. `project_id`
2. 当前 `run_id` 或当前 `event_id`

### 5.2 从 Evidence 打开

默认带入：

1. `project_id`
2. `evidence_id`
3. 关联的 `trace_id / event_id / run_id`

### 5.3 从 Replay Drawer 打开

默认带入：

1. `run_id`
2. 当前 `replay_id`
3. 当前 `trace_id / span_id`

## 6. 审计结果列表模型

建议结果列表不要直接混排所有原始材料，而是用统一的 `AuditRecord` 视图对象。

### 6.1 建议字段

1. `record_id`
2. `record_kind`
3. `project_id`
4. `run_id`
5. `event_id`
6. `trace_id`
7. `span_id`
8. `brain_kind`
9. `summary`
10. `severity`
11. `artifact_state`
12. `created_at`
13. `deep_link`

### 6.2 `record_kind` 建议

1. `run_event`
2. `replay_item`
3. `log_segment`
4. `evidence_link`
5. `manual_action`

## 7. 排序策略

默认排序不应只是时间倒序。

建议：

1. `severity` 高的优先
2. 当前过滤上下文最相关的优先
3. 最近变化优先

### 7.1 同级排序

同级情况下建议按：

1. `created_at desc`
2. `run_id`
3. `seq_no`

## 8. 与 Replay Drawer 的联动

### 8.1 从审计结果进入 Replay Drawer

如果 `record_kind` 可回放：

1. 点击主卡片或“查看回放”动作
2. 打开 `Replay Drawer`
3. 自动定位到相关 `event_id / replay_id / trace_id`

### 8.2 从 Replay Drawer 进入审计过滤器

允许：

1. 以当前 `run_id`
2. 当前 `event_id`
3. 当前 `trace_id / span_id`

为默认过滤条件打开审计视图。

## 9. 过滤器 UI 结构

建议采用：

1. 顶部固定筛选条
2. 高级筛选折叠区
3. 当前筛选条件 chips 区
4. 结果列表区

### 9.1 顶部固定筛选条

固定展示：

1. 时间范围
2. run
3. severity
4. event type
5. trace

### 9.2 筛选 chips

每个已生效的条件都应显示为可移除 chip，避免用户迷失在多重筛选下。

## 10. 与接口层的关系

审计过滤器不应自己拼多个接口。

建议最终提供统一审计查询接口，例如：

```text
GET /api/v3/projects/{project_id}/audit-records
```

建议参数：

1. `run_id`
2. `event_id`
3. `trace_id`
4. `span_id`
5. `brain_kind`
6. `severity`
7. `record_kind`
8. `evidence_id`
9. `surface`
10. `journey_id`
11. `time_from`
12. `time_to`
13. `cursor`
14. `limit`

## 11. 降级与异常状态

### 11.1 文件已清理

如果关联 replay/log 文件已清理：

1. 记录仍出现在审计结果中
2. `artifact_state = artifact_pruned`
3. 跳转到 Replay Drawer 时仅展示摘要

### 11.2 文件缺失

如果索引存在但文件缺失：

1. 结果项显示 warning
2. 不阻断结果列表其他项展示

## 12. 不该怎么做

不建议：

1. 只有全文搜索没有结构化过滤
2. 把审计过滤器做成纯日志检索
3. 一个过滤结果项无法进入 Replay Drawer
4. 过滤器状态不显式展示给用户

## 13. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. 审计视图必须复用 `run_id / event_id / trace_id / evidence_id` 过滤主线
2. Replay Drawer 与审计视图必须双向可跳转
3. 后续审计查询接口设计应围绕 `AuditRecord` 视图模型展开

## 14. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-审计查询接口设计](./EasyMVP-V3-%E5%AE%A1%E8%AE%A1%E6%9F%A5%E8%AF%A2%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md)
2. 审计列表页面设计
3. 审计 chips 与保存筛选集设计
