# EasyMVP V3 Evidence 卡片查询接口设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Evidence索引表结构设计](./EasyMVP-V3-Evidence索引表结构设计.md)
> 关联文档：[EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 关联文档：[EasyMVP-V3-Replay查询接口设计](./EasyMVP-V3-Replay查询接口设计.md)
> 关联文档：[EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence%E5%8D%A1%E7%89%87%E7%BB%84%E4%BB%B6%E8%A7%84%E8%8C%83.md)
> 关联文档：[EasyMVP-V3-Evidence Preview交互设计](./EasyMVP-V3-Evidence%20Preview%E4%BA%A4%E4%BA%92%E8%AE%BE%E8%AE%A1.md)
> 目标：定义 Acceptance 页和相关页面如何查询 Evidence 卡片、证据详情、预览目标与 Replay 回链，形成稳定的页面接口边界。

## 1. 设计结论

Evidence 不能让前端直接扫 `workflow_evidence_index` 再自己拼卡片。

正确做法是提供页面语义明确的查询接口：

1. Evidence 卡片列表接口
2. Evidence 详情接口
3. Evidence 预览接口
4. Evidence 到 Replay 的回链接口

一句话：

> Evidence 页面的查询对象应是“卡片”和“详情”，不是底层索引行。

补充边界：

1. 接口返回的是 EasyMVP 聚合后的页面字段
2. 不直接暴露 `brain-v3` 原始 payload、原始工具结果或裸文件路径

## 2. 为什么要单独做接口专题

如果没有专门接口，后面很容易退化成：

1. 页面自己按 `evidence_type` 分组
2. 页面自己推断 preview 怎么打开
3. 页面自己拼 replay 跳转逻辑
4. 不同页面对同一 Evidence 展示不一致

所以必须先把 Evidence 页面接口定成正式边界。

## 3. 核心接口建议

建议第一版至少提供：

### 3.1 Evidence Cards

```text
GET /api/v3/projects/{project_id}/acceptance-runs/{acceptance_run_id}/evidence-cards
```

作用：

1. 支撑 Acceptance 页右侧 Evidence 区
2. 支撑按分类分组显示

### 3.2 Evidence Detail

```text
GET /api/v3/projects/{project_id}/evidence/{evidence_id}
```

作用：

1. 打开单条 Evidence 详情
2. 展示完整元信息与文件状态

### 3.3 Evidence Preview

```text
GET /api/v3/projects/{project_id}/evidence/{evidence_id}/preview
```

作用：

1. 返回可直接用于预览的目标信息
2. 文件缺失或已清理时返回降级状态

### 3.4 Evidence Replay Links

```text
GET /api/v3/projects/{project_id}/evidence/{evidence_id}/replay-links
```

作用：

1. 返回与当前 Evidence 相关的 replay 入口
2. 支撑从 Evidence 卡进入 Replay Drawer

## 4. Evidence Card 模型

建议列表接口返回 `EvidenceCard`，每项至少包含：

1. `evidence_id`
2. `evidence_type`
3. `title`
4. `summary`
5. `surface`
6. `journey_id`
7. `source_brain`
8. `status`
9. `artifact_state`
10. `generated_at`
11. `preview_target`
12. `replay_available`

字段边界补充：

1. `source_brain` 只是归一化后的证据来源归属字段，用于 provenance 展示
2. `browser / verifier / code` 应理解为 `brain-v3` 内置脑产生的归一化证据来源
3. `preview_target` 是页面可消费的预览目标描述，不等于底层文件系统路径

## 5. 列表查询参数

### 5.1 高频参数

建议支持：

1. `surface`
2. `journey_id`
3. `evidence_type`
4. `status`
5. `group_by`

### 5.2 分页参数

建议支持：

1. `cursor`
2. `limit`

## 6. 分组策略

由于 Acceptance 页强调“看懂证据类型和覆盖”，建议接口层支持可选分组。

### 6.1 建议 `group_by`

1. `category`
2. `surface`
3. `evidence_type`
4. `journey`

### 6.2 默认分组

建议默认：

1. `category`

对应：

1. `browser`
2. `ci/build`
3. `runtime`
4. `verification`
5. `export`

## 7. Evidence Detail 模型

详情接口建议返回：

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
11. `artifact_state`
12. `title`
13. `summary`
14. `file_ext`
15. `mime_type`
16. `file_size`
17. `sha256`
18. `generated_at`
19. `validated_at`
20. `preview_target`
21. `replay_target`

详情接口同样只返回归一化后的展示与追踪字段，不直接透出原始 run payload。

## 8. Preview 接口建议

Preview 接口不应直接裸返回本地路径。

建议返回：

1. `evidence_id`
2. `artifact_state`
3. `preview_kind`
4. `mime_type`
5. `preview_url_or_content`
6. `download_allowed`

其中：

1. `preview_url_or_content` 是安全可展示的预览结果或受控内容，不应退化成裸本地路径
2. 如果底层产物不可安全直出，应返回降级态而不是透出原始存储位置

### 8.1 `preview_kind` 建议

1. `image`
2. `video`
3. `json`
4. `text`
5. `download_only`

## 9. Replay 回链接口建议

如果 Evidence 存在 replay 关联，建议返回：

1. `replay_target.run_id`
2. `replay_target.replay_id`
3. `replay_target.event_id`
4. `replay_target.trace_id`
5. `replay_target.span_id`

这样页面不需要自己拼回放上下文。

## 10. 降级与异常状态

### 10.1 `artifact_pruned`

表现：

1. 卡片可见
2. 详情可见
3. preview 不可用
4. 页面提示“附件已清理”

### 10.2 `artifact_missing`

表现：

1. 卡片 warning 显示
2. preview 不可用
3. replay 回链若存在仍应可用

### 10.3 上游运行时受限

如果证据采集上游命中 `unsupported / denied`：

1. 应在卡片或详情中保留显式受限语义
2. 不得伪装成“已采集成功但暂无预览”

## 11. 与 Acceptance 页的关系

Acceptance 页的 Evidence 区应直接消费：

1. `evidence-cards`
2. `evidence/{id}`
3. `evidence/{id}/preview`
4. `evidence/{id}/replay-links`

页面不应：

1. 直接扫索引表
2. 自己推断 preview 模式
3. 自己推断 replay 深链

## 12. 不该怎么做

不建议：

1. 只提供 Evidence 索引表查询，不提供页面语义接口
2. preview 接口直接暴露磁盘路径
3. Replay 回链需要前端自己拼 `run_id / trace_id`
4. 卡片接口和详情接口字段风格不统一

## 13. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Acceptance 页设计必须以 `EvidenceCard` 为唯一列表模型
2. Replay Drawer 若从 Evidence 打开，应优先消费 `replay-links`
3. 后续 Evidence 卡片组件规范必须复用这里的字段模型

## 14. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence%E5%8D%A1%E7%89%87%E7%BB%84%E4%BB%B6%E8%A7%84%E8%8C%83.md)
2. Evidence Preview 交互设计
3. Evidence 导出面板设计
