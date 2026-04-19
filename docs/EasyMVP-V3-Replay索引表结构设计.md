# EasyMVP V3 Replay 索引表结构设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-replay与log artifact存储规范](./EasyMVP-V3-replay与log artifact存储规范.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
> 关联文档：[EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
> 目标：把 replay 与 log artifact 从文件级规范进一步落成数据库索引表结构，支撑 Replay 页面、审计入口、Run 详情与 Evidence 回链。

## 1. 设计结论

V3 的 replay 不能只靠 `runs/{run_id}/replay/` 目录存在。

正式做法应为：

1. replay 明细文件落本地目录
2. replay 索引记录落数据库
3. log 分片索引落数据库
4. replay 与 `run_id / event_id / trace_id / span_id` 可回链
5. 页面通过索引表按层读取，不直接扫目录

一句话：

> run 目录保存原始材料，Replay 索引表保存“材料之间的可查询关系”。

## 2. 要解决的问题

本专题主要解决：

1. Replay 页面缺少正式事实表
2. 原始回放文件无法稳定筛选和排序
3. 审计入口只能跳到 run，不能跳到具体回放切片
4. log/replay/evidence 之间缺少正式索引桥梁
5. 文件被清理后页面无法优雅降级

## 3. 设计原则

建议遵守：

1. replay 索引和 log 索引分表
2. 高频筛选字段单独列
3. 原始大文本不上主表
4. 与 `workflow_brain_run_events` 保持可关联
5. 与 Evidence 关系通过回链字段实现，不直接耦合成单表

## 4. 核心表建议

建议第一版至少增加：

1. `workflow_replay_index`
2. `workflow_run_log_segments`

其中：

1. `workflow_replay_index` 负责回放切片、结构化步骤、工具调用材料索引
2. `workflow_run_log_segments` 负责原始日志分片索引

## 5. `workflow_replay_index`

### 5.1 语义

表示某个 run 中一段可回放材料的主索引记录。

### 5.2 建议核心列

1. `id`
2. `replay_id`
3. `project_id`
4. `run_id`
5. `domain_task_id`
6. `compiled_task_id`
7. `event_id`
8. `trace_id`
9. `span_id`
10. `replay_kind`
11. `seq_no`
12. `title`
13. `summary`
14. `file_path`
15. `file_ext`
16. `mime_type`
17. `file_size`
18. `sha256`
19. `source_object_kind`
20. `source_object_id`
21. `status`
22. `created_at`
23. `updated_at`

### 5.3 `replay_kind` 建议

建议第一版至少支持：

1. `step_snapshot`
2. `tool_call`
3. `tool_result`
4. `thought_summary`
5. `browser_trace`
6. `runtime_capture`
7. `verification_snapshot`

### 5.4 `status` 建议

建议：

1. `available`
2. `artifact_missing`
3. `artifact_pruned`

## 6. `workflow_run_log_segments`

### 6.1 语义

表示某个 run 的原始日志分片索引。

### 6.2 建议核心列

1. `id`
2. `project_id`
3. `run_id`
4. `segment_id`
5. `stream_kind`
6. `seq_no`
7. `file_path`
8. `file_size`
9. `sha256`
10. `started_at`
11. `ended_at`
12. `status`
13. `created_at`

### 6.3 `stream_kind` 建议

建议：

1. `stdout`
2. `stderr`
3. `system`
4. `tool`

### 6.4 `status` 建议

建议：

1. `available`
2. `artifact_missing`
3. `artifact_pruned`

## 7. 与现有表的关系

### 7.1 与 `workflow_brain_run_bindings`

关系：

1. 一条 run binding 对应多条 replay 索引
2. 一条 run binding 对应多条日志分片索引

### 7.2 与 `workflow_brain_run_events`

关系：

1. 一条高价值 event 可关联零到多条 replay 索引
2. Replay 页面先读事件，再展开对应 replay 切片

### 7.3 与 Evidence

关系：

1. 并非所有 replay 都是 Evidence
2. 只有被验收正式采纳的 replay/log material 才映射为 Evidence
3. Evidence 可通过 `event_id / trace_id / run_id` 回链到 replay

## 8. 唯一约束建议

### 8.1 Replay 主记录

建议：

1. `uniq(replay_id)`

### 8.2 Log 分片

建议：

1. `uniq(run_id, stream_kind, seq_no)`

## 9. 索引建议

### 9.1 Replay 页面查询

建议索引：

1. `idx(run_id, seq_no)`
2. `idx(project_id, created_at desc)`
3. `idx(event_id)`
4. `idx(trace_id, span_id)`

### 9.2 Run 详情查询

建议索引：

1. `idx(run_id, replay_kind, seq_no)`
2. `idx(run_id, stream_kind, seq_no)`

### 9.3 审计查询

建议索引：

1. `idx(project_id, run_id, created_at desc)`
2. `idx(domain_task_id, created_at desc)`

## 10. 页面查询视角

为了支撑三层回放展示，建议查询分层如下：

### 10.1 摘要层

来源：

1. `workflow_brain_run_bindings`
2. `workflow_brain_run_events`

### 10.2 结构化事件层

来源：

1. `workflow_brain_run_events`
2. `workflow_replay_index`

### 10.3 原始材料层

来源：

1. `workflow_replay_index`
2. `workflow_run_log_segments`

## 11. 与文件清理的关系

如果原始 replay 或日志文件被清理：

1. 不删除索引
2. 将对应索引标记为 `artifact_pruned`
3. 页面保留摘要和关联关系
4. 禁止预览原始材料

如果文件异常丢失：

1. 标记为 `artifact_missing`
2. 进入 diagnostics

## 12. 与 migration 的关系

建议在后续 migration 中加入：

1. `workflow_replay_index`
2. `workflow_run_log_segments`
3. 对 `run_id / event_id / trace_id / seq_no` 的关键索引

## 13. 不该怎么做

不建议：

1. 只存目录路径，不存 replay 索引
2. 让 Replay 页面直接扫 `replay/` 目录
3. 把全部原始日志内容塞进数据库
4. 用 `workflow_brain_run_events` 直接替代 replay 索引表
5. 清理原始文件时删除关联索引

## 14. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Replay drawer 页面设计必须以本专题表结构为查询基线
2. Run 详情页若展示日志分片，必须基于 `workflow_run_log_segments`
3. 审计过滤器设计必须复用 `run_id / event_id / trace_id / span_id`
4. Evidence 若回链 replay，必须使用本专题约束的关联字段

## 15. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
2. Replay drawer 页面设计
3. log parser / normalizer 结果落表规则
4. 原始材料按需加载与分页设计
