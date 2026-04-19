# EasyMVP V3 审计查询接口设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-审计过滤器设计](./EasyMVP-V3-审计过滤器设计.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer页面设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-审计列表页面设计](./EasyMVP-V3-%E5%AE%A1%E8%AE%A1%E5%88%97%E8%A1%A8%E9%A1%B5%E9%9D%A2%E8%AE%BE%E8%AE%A1.md)
> 目标：定义 V3 中审计视图的正式查询接口、统一响应模型、过滤参数、分页策略与和 Replay Drawer 的联动接口边界。

## 1. 设计结论

审计查询不应靠前端分别调用：

1. run events
2. replay items
3. log segments
4. evidence links

再自己拼结果。

正确做法是提供统一的审计查询接口，返回标准化的 `AuditRecord` 列表。

一句话：

> 审计接口要输出“可读的审计记录流”，而不是暴露一堆底层表给前端自己拼。

## 2. 为什么必须统一接口

如果不统一，后面会出现：

1. 审计页和 Replay Drawer 使用不同的筛查逻辑
2. 同一个问题在不同页面排序完全不一样
3. 前端承担过多跨表聚合责任
4. 过滤器状态难以复用和保存

## 3. 核心接口建议

建议第一版至少提供：

### 3.1 Audit Records

```text
GET /api/v3/projects/{project_id}/audit-records
```

作用：

1. 返回统一审计记录流
2. 支撑审计视图主列表
3. 支撑审计过滤器

### 3.2 Audit Record Detail

```text
GET /api/v3/projects/{project_id}/audit-records/{record_id}
```

作用：

1. 返回某条记录的详细信息
2. 提供进入 Replay Drawer 或 Evidence 的深链

### 3.3 Saved Filter Presets

如果后续需要，可扩：

```text
GET /api/v3/projects/{project_id}/audit-filter-presets
```

当前可先不做持久化，但接口模型应预留。

## 4. `AuditRecord` 统一响应模型

建议主列表每项至少包含：

1. `record_id`
2. `record_kind`
3. `project_id`
4. `run_id`
5. `event_id`
6. `trace_id`
7. `span_id`
8. `brain_kind`
9. `severity`
10. `summary`
11. `artifact_state`
12. `source_object_kind`
13. `source_object_id`
14. `created_at`
15. `deep_link`
16. `replay_target`

## 5. `record_kind` 建议

建议至少支持：

1. `run_event`
2. `replay_item`
3. `log_segment`
4. `evidence_link`
5. `manual_action`

## 6. 查询参数建议

### 6.1 高频参数

建议支持：

1. `run_id`
2. `event_id`
3. `trace_id`
4. `span_id`
5. `severity`
6. `record_kind`
7. `time_from`
8. `time_to`

### 6.2 中频参数

建议支持：

1. `brain_kind`
2. `status`
3. `artifact_state`
4. `replay_kind`
5. `stream_kind`
6. `domain_task_id`
7. `compiled_task_id`

### 6.3 Evidence 关联参数

建议支持：

1. `evidence_id`
2. `evidence_type`
3. `surface`
4. `journey_id`

## 7. 分页策略

审计结果不应一次性全量返回。

建议：

1. 使用 cursor 分页
2. 默认 `limit` 为适中值
3. 允许 `next_cursor`
4. 支持稳定排序字段

建议返回：

1. `items`
2. `next_cursor`
3. `has_more`
4. `as_of`

## 8. 排序策略

默认排序不应只是时间倒序。

建议支持：

1. `relevance`
2. `created_at_desc`
3. `severity_desc`

### 8.1 默认排序

建议默认：

1. 当存在过滤条件时使用 `relevance`
2. 无过滤条件时使用 `created_at_desc`

## 9. 结果分组建议

接口层可提供可选分组能力，但不应强制。

建议支持：

1. `group_by=run`
2. `group_by=trace`
3. `group_by=record_kind`

默认不分组，只返回平铺记录流。

## 10. 与 Replay Drawer 的联动

每条可回放记录建议包含：

1. `replay_target.run_id`
2. `replay_target.replay_id`
3. `replay_target.event_id`
4. `replay_target.trace_id`
5. `replay_target.span_id`

这样前端无需再推断如何打开 Replay Drawer。

## 11. Detail 接口建议

`GET /audit-records/{record_id}` 建议至少返回：

1. `record_id`
2. `record_kind`
3. `summary`
4. `full_context`
5. `run_id`
6. `event_id`
7. `trace_id`
8. `span_id`
9. `related_records`
10. `replay_target`
11. `evidence_target`

## 12. 降级与异常状态

### 12.1 原始材料已清理

返回：

1. `artifact_state = artifact_pruned`
2. 仍保留记录项
3. 仍可进入 Replay Drawer 摘要层

### 12.2 原始材料缺失

返回：

1. `artifact_state = artifact_missing`
2. 结果项上有 warning
3. 不影响其他记录展示

## 13. 与工作台的关系

工作台里的某些高价值事件可直接跳到审计查询，建议带入：

1. `project_id`
2. `run_id`
3. `event_id`

这样用户可以从工作台快速进入更大范围的上下文审查。

## 14. 不该怎么做

不建议：

1. 前端自己合并 replay/event/evidence 三类接口
2. 审计接口不带 `replay_target`
3. 记录项不返回 `artifact_state`
4. 过滤条件和结果排序规则不稳定

## 15. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. 审计列表页面设计必须以 `AuditRecord` 为唯一结果模型
2. Replay Drawer 若从审计进入，应直接消费 `replay_target`
3. 审计过滤器的所有 UI 条件都必须能映射到这里的查询参数

## 16. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-审计列表页面设计](./EasyMVP-V3-%E5%AE%A1%E8%AE%A1%E5%88%97%E8%A1%A8%E9%A1%B5%E9%9D%A2%E8%AE%BE%E8%AE%A1.md)
2. 审计筛选预设保存设计
3. 审计导出查询设计
