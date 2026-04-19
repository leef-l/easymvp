# EasyMVP V3 Evidence 文件命名与引用规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 关联文档：[EasyMVP-V3-本地目录与项目工作区规范](./EasyMVP-V3-本地目录与项目工作区规范.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-Evidence索引表结构设计](./EasyMVP-V3-Evidence索引表结构设计.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 目标：定义 V3 中 `Evidence` 的文件命名、目录落点、数据库索引字段、引用关系、状态与页面展示边界。

## 1. 设计结论

V3 的 `Evidence` 不能只是一个 `artifact_uri` 字段。

正式设计应定为：

1. Evidence 是结构化对象
2. 文件本体落本地目录
3. 数据库保存索引、引用、状态和校验信息
4. 文件命名由系统生成，不依赖人工命名
5. 页面只读 Evidence 索引，不直接扫目录

一句话：

> Evidence 必须既是“文件”，也是“可裁决的结构化证据对象”。

## 2. 为什么要单独规范

如果没有独立规范，后面会出现：

1. 截图、录屏、报告各自命名
2. 同一个 journey 的证据无法稳定关联
3. 页面只能展示文件列表，不能展示证据语义
4. 回放、验收、工作台无法共用同一证据对象
5. 清理文件后数据库引用失效但没人知道

所以 `Evidence` 必须被视为正式领域对象，而不是附件。

## 3. Evidence 的正式定义

每条 Evidence 至少应同时包含两部分：

### 3.1 文件部分

表示文件本体：

1. 截图
2. 录屏
3. 导出物
4. 报告
5. 日志片段

### 3.2 索引部分

表示可供系统理解和裁决的元信息：

1. 来自哪个项目
2. 来自哪个 run
3. 对应哪个 surface
4. 对应哪个 journey
5. 对应哪种 evidence type
6. 当前是否有效

## 4. Evidence 类型建议

建议第一版至少支持：

1. `browser_screenshot`
2. `browser_trace`
3. `runtime_log`
4. `runtime_screenshot`
5. `runtime_video`
6. `build_artifact`
7. `ci_result`
8. `verification_report`
9. `export_output`
10. `manual_review_note`

## 5. 文件落点规则

默认按目录类型落点：

### 5.1 截图类

落：

```text
projects/{project_id}/evidence/screenshots/
```

### 5.2 视频类

落：

```text
projects/{project_id}/evidence/videos/
```

### 5.3 导出物类

落：

```text
projects/{project_id}/evidence/exports/
```

### 5.4 报告类

落：

```text
projects/{project_id}/evidence/reports/
```

### 5.5 其他附件

落：

```text
projects/{project_id}/evidence/attachments/
```

## 6. 文件命名规范

### 6.1 命名原则

必须满足：

1. 稳定
2. 可追溯
3. 易排序
4. 不依赖中文标题
5. 不依赖人工重命名

### 6.2 推荐模板

建议：

```text
{ts}_{evidence_type}_{evidence_id}_{surface}_{journey_or_scope}.{ext}
```

示例：

1. `20260419T112311Z_browser_screenshot_ev_1001_user_frontend_signup.png`
2. `20260419T112659Z_runtime_video_ev_1002_game_runtime_level-loop.mp4`
3. `20260419T113005Z_verification_report_ev_1003_api_backend_order-submit.json`

### 6.3 必备规则

1. 时间戳在前，保证目录天然按时间排序
2. 文件名只使用 ASCII 小写、数字、下划线、短横线
3. `evidence_id` 必须进入文件名
4. `surface` 必须进入文件名
5. `journey` 可做裁剪后的短标识，不直接拼超长原文

## 7. 数据库索引字段

Evidence 元数据建议至少包含：

1. `evidence_id`
2. `project_id`
3. `acceptance_run_id`
4. `run_id`
5. `domain_task_id`
6. `surface`
7. `journey_id`
8. `evidence_type`
9. `source_brain`
10. `status`
11. `file_path`
12. `file_ext`
13. `file_size`
14. `sha256`
15. `summary`
16. `generated_at`
17. `validated_at`
18. `source_object_kind`
19. `source_object_id`

其中：

1. `source_brain` 仅表示归一化后的证据来源归属
2. 不用于暴露原始工具名、原始 payload 或底层执行细节

## 8. 引用关系规范

一条 Evidence 建议至少能回溯到以下对象中的一部分：

1. `project_id`
2. `acceptance_run_id`
3. `run_id`
4. `domain_task_id`
5. `surface`
6. `journey_id`

这样可以支撑：

1. Acceptance 覆盖矩阵
2. 工作台快捷跳转
3. replay 关联
4. 问题定位

## 9. Evidence 状态

建议状态至少分为：

1. `collected`
2. `validated`
3. `rejected`
4. `artifact_missing`
5. `artifact_pruned`

### 9.1 `collected`

文件已生成，已建立索引，但尚未被验收逻辑确认。

### 9.2 `validated`

文件和元信息都通过验收逻辑检查，可作为正式证据使用。

### 9.3 `rejected`

证据已收集，但不满足当前验收要求。

### 9.4 `artifact_missing`

数据库索引存在，但对应物理文件不存在。

### 9.5 `artifact_pruned`

文件曾存在，但因清理策略被移除；索引仍保留。

## 10. 文件生成与索引写入顺序

建议统一流程：

```text
collect evidence
  → write file
  → compute hash/size
  → write evidence index
  → attach to acceptance run
  → emit live event
```

不建议先写索引再异步落文件，否则容易产生大面积 `artifact_missing`。

## 11. 与 Acceptance 的关系

Evidence 必须直接服务 `Acceptance` 页，而不是只服务存档。

一条 Evidence 至少应支持：

1. 出现在 `evidence_cards`
2. 计入 `coverage`
3. 影响 `blocking_gap`
4. 支撑 `final_judgement`

## 12. Evidence 卡片展示约束

页面展示时每张 Evidence 卡至少应展示：

1. `evidence_type`
2. `title`
3. `summary`
4. `surface`
5. `journey`
6. `source_brain`
7. `generated_at`
8. `status`
9. `preview_target`

页面不应展示成裸文件列表。

## 13. 与 Replay / Audit 的关系

以下类型 Evidence 应允许关联 replay：

1. `runtime_log`
2. `runtime_video`
3. `browser_trace`
4. `verification_report`

建议补充：

1. `trace_id`
2. `span_id`
3. `event_id`

这样可以从证据卡直接进入 replay 或审计视图。

## 14. 校验与防篡改

为了避免证据文件漂移，建议至少记录：

1. `sha256`
2. `file_size`
3. `generated_at`
4. `source_brain`

这里的 `source_brain` 仍是 provenance 标识，不应被实现层误用成执行控制开关。

页面预览前或导出前，建议支持：

1. 文件存在性检查
2. hash 一致性检查

## 15. 清理与保留策略

### 15.1 默认不应清理

1. 已 `validated` 的 blocking evidence
2. 当前最新版验收/验证链路仍在引用的 evidence

补充说明：

- 这里的旧表述里可以包含 `AcceptanceRun`
- 但按当前钱学森总纲，Evidence 的保留判断不应只围绕 `AcceptanceRun`
- 还应逐步兼容 `VerificationResult / CompletionVerdict / RuntimeEscalation` 的引用语义

### 15.2 可清理

1. 被 `rejected` 的低价值附件
2. 已归档项目中的非关键视频
3. 可重建的中间截图

### 15.3 清理后的处理

1. 不删除索引
2. 状态改为 `artifact_pruned`
3. 页面上标出“证据索引存在，附件已清理”

## 16. 不该怎么做

不建议：

1. 使用用户原始文件名作为正式 evidence 名
2. 不记录 hash 只记录路径
3. 页面直接扫 `evidence/` 目录组装证据
4. 一个文件同时承担多个 evidence 语义但无结构化索引
5. 清理文件时静默删除数据库记录

## 17. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. `Acceptance` 页的证据卡必须使用本专题字段
2. `ProductionAcceptanceProfile` 中的 required evidence 必须映射到本专题的 `evidence_type`
3. replay 文档必须支持通过 `run_id / trace_id / event_id` 回链到 Evidence
4. 存储清理文档必须复用 `artifact_missing / artifact_pruned` 状态

## 18. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Evidence索引表结构设计](./EasyMVP-V3-Evidence索引表结构设计.md)
2. Evidence 卡片组件规范
3. Evidence 预览与 hash 校验设计
4. 验收导出包中的 Evidence 打包规范
