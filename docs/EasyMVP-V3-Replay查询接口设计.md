# EasyMVP V3 Replay 查询接口设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Replay索引表结构设计](./EasyMVP-V3-Replay索引表结构设计.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 目标：定义 V3 中 Replay 页、Run 详情、审计入口对 replay 与日志材料的查询接口边界、响应模型和分层读取策略。

## 1. 设计结论

Replay 不应只有一个“读原始日志”的接口。

正式接口应分成三层：

1. 摘要接口
2. 结构化事件/切片接口
3. 原始材料接口

一句话：

> Replay 查询必须与页面三层展示模型一致，而不是把所有数据塞进一个大接口。

## 2. 为什么要单独做接口专题

如果没有正式接口设计，后面很容易退化成：

1. 页面直接扫目录
2. 页面把日志当 replay
3. 页面首屏拉一大坨原始文件
4. 审计和回放各自造接口

所以必须先把查询边界定清。

## 3. 接口设计原则

建议遵守：

1. 首屏只返回摘要和结构化数据
2. 原始材料按需加载
3. 所有接口都可用 `run_id` 定位
4. 能按 `event_id / trace_id / span_id` 深链
5. 文件不存在时返回降级状态，不直接 500

## 4. 查询层次

### 4.1 摘要层

回答：

1. 这个 run 是什么
2. 当前回放有多少材料
3. 最近发生了哪些关键事件

### 4.2 结构化层

回答：

1. 某个 event 对应哪些 replay 切片
2. 某个 run 的 replay 时间序列长什么样
3. 哪些切片可展开

### 4.3 原始层

回答：

1. 原始 replay 文件内容
2. 原始日志分片内容
3. 某个材料是否已丢失或被清理

## 5. 核心接口建议

建议至少提供以下接口。

### 5.1 Replay Summary

```text
GET /api/v3/projects/{project_id}/runs/{run_id}/replay-summary
```

作用：

1. 支撑 replay 页面首屏
2. 支撑 run detail drawer 顶部信息

建议返回：

1. `run_id`
2. `project_id`
3. `brain_kind`
4. `status`
5. `started_at`
6. `ended_at`
7. `event_count`
8. `replay_count`
9. `log_segment_count`
10. `artifact_status_summary`
11. `entry_points`

### 5.2 Replay Timeline

```text
GET /api/v3/projects/{project_id}/runs/{run_id}/replay-timeline
```

作用：

1. 返回结构化时间线
2. 按 seq 或时间顺序展示 replay 切片

建议参数：

1. `cursor`
2. `limit`
3. `replay_kind`
4. `event_id`
5. `trace_id`

建议返回每项：

1. `replay_id`
2. `seq_no`
3. `replay_kind`
4. `title`
5. `summary`
6. `event_id`
7. `trace_id`
8. `span_id`
9. `status`
10. `preview_available`
11. `raw_target`

### 5.3 Replay Detail

```text
GET /api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}
```

作用：

1. 打开某条 replay 的详情
2. 读取结构化元信息和原始材料引用

建议返回：

1. `replay_id`
2. `replay_kind`
3. `title`
4. `summary`
5. `source_object_kind`
6. `source_object_id`
7. `event_id`
8. `trace_id`
9. `span_id`
10. `status`
11. `raw_preview`
12. `related_log_segments`

### 5.4 Replay Raw Content

```text
GET /api/v3/projects/{project_id}/runs/{run_id}/replay-items/{replay_id}/raw
```

作用：

1. 读取原始回放文件
2. 只在用户展开后调用

建议返回：

1. `replay_id`
2. `status`
3. `mime_type`
4. `encoding`
5. `content`
6. `truncated`

### 5.5 Log Segment List

```text
GET /api/v3/projects/{project_id}/runs/{run_id}/log-segments
```

作用：

1. 查看某个 run 的原始日志分片列表

建议返回每项：

1. `segment_id`
2. `stream_kind`
3. `seq_no`
4. `started_at`
5. `ended_at`
6. `status`
7. `size`
8. `raw_target`

### 5.6 Log Segment Raw

```text
GET /api/v3/projects/{project_id}/runs/{run_id}/log-segments/{segment_id}/raw
```

作用：

1. 按需读取原始日志内容

建议返回：

1. `segment_id`
2. `stream_kind`
3. `status`
4. `content`
5. `truncated`

## 6. 深链查询接口

为了支撑从工作台、Evidence、审计入口直接跳到回放，建议支持以下查询方式：

### 6.1 按 event 查询

```text
GET /api/v3/projects/{project_id}/replay-by-event/{event_id}
```

### 6.2 按 trace 查询

```text
GET /api/v3/projects/{project_id}/replay-by-trace
```

建议参数：

1. `trace_id`
2. `span_id`

### 6.3 按 evidence 回链

```text
GET /api/v3/projects/{project_id}/evidence/{evidence_id}/replay-links
```

## 7. 响应模型建议

### 7.1 通用响应字段

建议所有 Replay 查询接口都统一包含：

1. `as_of`
2. `artifact_state`
3. `last_event_id`
4. `refresh_hint`

### 7.2 `artifact_state` 建议

建议统一为：

1. `available`
2. `artifact_missing`
3. `artifact_pruned`

这样页面可直接按状态降级渲染。

## 8. 分页与按需加载

Replay 时间线和日志分片都不应一次性全量返回。

建议：

1. 时间线支持 cursor 分页
2. 原始内容接口默认可截断
3. 页面滚动到底部再拉下一页
4. 原始日志只在展开时请求

## 9. 与页面的对应关系

### 9.1 Replay 页面

首屏：

1. `replay-summary`
2. `replay-timeline`

展开某项：

1. `replay-detail`
2. 必要时 `replay-raw`

### 9.2 Run Detail Drawer

只需要：

1. `replay-summary`
2. 最近若干 `replay-timeline`

### 9.3 Acceptance Evidence

通过：

1. `evidence -> replay-links`

跳到 replay 详情。

## 10. 错误与降级策略

### 10.1 文件缺失

返回：

1. `artifact_state = artifact_missing`
2. 仍返回结构化摘要

### 10.2 文件已清理

返回：

1. `artifact_state = artifact_pruned`
2. 原始内容字段为空
3. 页面展示“原始材料已清理”

### 10.3 原始内容过大

返回：

1. `truncated = true`
2. 分段加载或下载入口

## 11. 不该怎么做

不建议：

1. 用单个接口返回全部 replay 和全部原始日志
2. 页面自己拼 replay 与 event 的对应关系
3. 原始文件缺失时直接报错中断整个页面
4. 把 replay 查询和工作台主查询混成一个超大接口

## 12. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Replay drawer 页面设计必须复用这里的接口分层
2. 审计过滤器设计必须基于 `event_id / trace_id / span_id`
3. Evidence 卡片若跳 replay，必须走 `replay-links` 接口
4. 工作台视图模型不应直接内嵌原始 replay 内容

## 13. 后续细分专题

本专题后续继续拆：

1. Replay drawer 页面设计
2. 审计过滤器接口设计
3. replay-links 与 Evidence 回链接口细化
4. 原始内容下载与分段加载设计
