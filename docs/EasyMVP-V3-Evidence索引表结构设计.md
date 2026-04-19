# EasyMVP V3 Evidence 索引表结构设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-Evidence卡片查询接口设计](./EasyMVP-V3-Evidence%E5%8D%A1%E7%89%87%E6%9F%A5%E8%AF%A2%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md)
> 目标：把 Evidence 从文件级规范进一步落成稳定数据库表结构，支撑 Acceptance、Workspace、Replay、Export 的统一索引与查询。

## 1. 设计结论

V3 的 Evidence 不能只散落在 `AcceptanceRun` 的 JSON 字段里。

正式做法应为：

1. 独立 Evidence 索引表
2. Evidence 与 `AcceptanceRun` 强关联
3. Evidence 与 `run_id / domain_task_id / journey_id` 可回链
4. 文件本体仍落文件系统
5. 表只存结构化索引与校验信息

一句话：

> Evidence 的文件在磁盘，Evidence 的真相在索引表。

按当前钱学森总纲，这里的 `AcceptanceRun` 更适合被理解为旧验收主线中的历史关联对象，而不是唯一中心对象。

当前更准确的理解应是：

1. Evidence 可以关联 `AcceptanceRun`
2. 但它还必须能服务 `VerificationResult`
3. 并且要能支撑 `CompletionVerdict / RuntimeEscalation / FaultSummary` 的解释链路

## 2. 要解决的问题

本专题主要解决：

1. 证据文件无法稳定检索
2. Coverage 统计缺少正式事实表
3. 页面无法高效筛选不同类型 Evidence
4. replay 与 Acceptance 之间缺少可靠关联
5. 清理附件后页面缺少可降级状态

还应补一条：

6. Evidence 无法稳定对齐“合同要求了什么、实际补到了什么、还缺什么”

## 3. 表设计原则

建议遵守：

1. 一个 Evidence 一条主索引记录
2. 高频查询字段单独列出
3. 文件相关字段不混进业务 JSON
4. 可用于 Acceptance 页面直接查询
5. 可用于后续导出和恢复

## 4. 核心表建议

建议第一版至少定义两张表：

1. `workflow_evidence_index`
2. `workflow_evidence_links`

其中：

1. `workflow_evidence_index` 存证据主记录
2. `workflow_evidence_links` 存一条证据与多个对象的补充关联

## 5. `workflow_evidence_index`

### 5.1 语义

表示一条正式 Evidence 的主索引。

### 5.2 建议核心列

1. `id`
2. `evidence_id`
3. `project_id`
4. `acceptance_run_id`
5. `run_id`
6. `domain_task_id`
7. `compiled_task_id`
8. `surface`
9. `journey_id`
10. `evidence_type`
11. `source_brain`
12. `status`
13. `title`
14. `summary`
15. `file_path`
16. `file_ext`
17. `mime_type`
18. `file_size`
19. `sha256`
20. `trace_id`
21. `span_id`
22. `event_id`
23. `source_object_kind`
24. `source_object_id`
25. `generated_at`
26. `validated_at`
27. `created_at`
28. `updated_at`

如果后续继续增强，建议优先补充或等价承载：

29. `verification_run_id`
30. `verification_check_id`
31. `contract_requirement_key`

### 5.3 列说明

#### 业务关联列

1. `project_id`
2. `acceptance_run_id`
3. `run_id`
4. `domain_task_id`
5. `compiled_task_id`
6. `surface`
7. `journey_id`

#### 文件索引列

1. `file_path`
2. `file_ext`
3. `mime_type`
4. `file_size`
5. `sha256`

#### 可观测关联列

1. `trace_id`
2. `span_id`

补充边界：

1. `source_brain` 只是索引层保存的归一化 provenance 字段
2. 页面接口仍应通过聚合层返回，不直接把索引行当成 UI DTO
3. `event_id`

#### 页面展示列

1. `title`
2. `summary`
3. `status`
4. `generated_at`
5. `validated_at`

## 6. `workflow_evidence_links`

### 6.1 语义

用于承载一个 Evidence 与多个对象的补充关系，避免主表被可变关联结构污染。

### 6.2 建议核心列

1. `id`
2. `project_id`
3. `evidence_id`
4. `link_kind`
5. `target_object_kind`
6. `target_object_id`
7. `created_at`

### 6.3 适用场景

例如一条 Evidence 可能同时关联：

1. 一个 `journey`
2. 一个 `run event`
3. 一个 `manual review action`

这些补充关系适合放在 links 表，而不是把主表做得过度宽和过度稀疏。

## 7. 状态设计

主表 `status` 建议使用：

1. `collected`
2. `validated`
3. `rejected`
4. `artifact_missing`
5. `artifact_pruned`

这组状态应与 [EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md) 保持一致。

## 8. 唯一约束建议

### 8.1 主记录唯一性

建议：

1. `uniq(evidence_id)`

### 8.2 文件哈希辅助约束

不建议仅靠 `sha256` 做全局唯一，因为：

1. 同一文件可能被不同 acceptance run 合法引用
2. 同一文件可能被复制到不同 evidence 语义

但建议索引：

1. `idx(project_id, sha256)`

## 9. 索引建议

### 9.1 Acceptance 页面常用查询

建议索引：

1. `idx(acceptance_run_id, status, generated_at desc)`
2. `idx(project_id, surface, journey_id)`
3. `idx(project_id, evidence_type, status)`

### 9.2 Workspace / 最近活动查询

建议索引：

1. `idx(project_id, generated_at desc)`
2. `idx(run_id, generated_at desc)`

### 9.3 Replay / Audit 查询

建议索引：

1. `idx(event_id)`
2. `idx(trace_id, span_id)`

## 10. Acceptance 查询视角

为了让 `Acceptance` 页高效展示，主表至少应支持以下查询：

1. 某次 `acceptance_run_id` 的全部 validated evidence
2. 某个 `surface + journey_id` 的证据列表
3. 某个 `evidence_type` 的最近证据
4. 某个 `blocking_gap` 对应缺了哪些 evidence type

按当前总纲，还应补充以下查询视角：

1. 某个 `verification_run_id` 对应的全部证据
2. 某个 `contract_requirement_key` 当前是否已被覆盖
3. 某个 `RuntimeEscalation` 关联了哪些证据

## 11. 与 Coverage 的关系

Coverage 不应直接从文件目录推断，而应基于 `workflow_evidence_index` 聚合。

例如：

1. `validated` 的 blocking evidence 可计入已覆盖
2. `collected` 但未验证的 evidence 计入 `partial`
3. `artifact_missing / artifact_pruned` 不应计入完整覆盖

## 12. 与文件清理的关系

如果物理文件被清理：

1. 不删除主索引记录
2. 更新 `status = artifact_pruned`
3. 保留 `file_path`、`sha256`、`generated_at`
4. 页面仍能展示摘要卡，但预览不可用

如果文件意外丢失：

1. 标记 `status = artifact_missing`
2. 进入 diagnostics 或 warning

## 13. 与导出包的关系

在项目导出时，Evidence 索引表应支持：

1. 只导出 `validated` 且关键的 evidence
2. 导出时按 `evidence_type / surface / journey` 分组
3. 保留 evidence 索引清单，便于离线阅读

## 14. 演进与兼容约束

考虑未来从 `SQLite` 迁移到 `Postgres`，建议：

1. 高价值查询字段单独列出
2. 不把核心条件全塞进 JSON
3. 主键策略保持一致
4. 文件路径与 storage root 解耦

## 15. 建议首批 migration 落点

建议在 migration 中加入：

1. `workflow_evidence_index`
2. `workflow_evidence_links`
3. 对 `acceptance_run_id / project_id / run_id / journey_id / evidence_type` 的关键索引

## 16. 不该怎么做

不建议：

1. 把 Evidence 全塞进 `AcceptanceRun.payload_json`
2. 让页面直接扫目录拼 evidence cards
3. 把 `sha256`、`file_size` 这类校验信息丢掉
4. 只记录 `artifact_uri` 不记录 `surface / journey / run_id`
5. 清理文件时直接删主索引记录
6. 把 Evidence 设计成只能围着 `AcceptanceRun` 转

## 17. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. `Acceptance` 页的证据卡片必须能从主索引表直接查询
2. replay 文档若把日志或回放材料升级为 Evidence，必须写入本表
3. 目录清理文档必须复用 `artifact_missing / artifact_pruned`
4. 导出包文档必须基于本表筛选 Evidence

## 18. 后续细分专题

本专题后续继续拆：

1. Coverage 聚合查询设计
2. [EasyMVP-V3-Evidence卡片查询接口设计](./EasyMVP-V3-Evidence%E5%8D%A1%E7%89%87%E6%9F%A5%E8%AF%A2%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md)
3. Evidence 导出清单 schema
4. Evidence 索引修复与重建设计
