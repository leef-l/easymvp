# EasyMVP V3 Evidence 卡片组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Evidence卡片查询接口设计](./EasyMVP-V3-Evidence卡片查询接口设计.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer页面设计.md)
> 关联文档：[EasyMVP-V3-Evidence Preview交互设计](./EasyMVP-V3-Evidence%20Preview%E4%BA%A4%E4%BA%92%E8%AE%BE%E8%AE%A1.md)
> 目标：定义 V3 中 Evidence 卡片的统一组件结构、状态展示、预览骨架、操作区与和 Replay/Preview 接口的绑定规则。

## 1. 设计结论

Evidence 卡片不应只是“文件名 + 下载按钮”。

正确的卡片组件应同时承载：

1. 证据语义
2. 当前状态
3. 预览入口
4. 回放入口
5. 关联 surface/journey 信息

一句话：

> Evidence 卡片是 Acceptance 页中“抽象裁决”和“具体证据”之间的桥梁。

## 2. 组件定位

Evidence 卡片的目标不是展示文件细节，而是帮助用户快速判断：

1. 这是什么证据
2. 它属于哪个 surface/journey
3. 它是否有效
4. 我能不能预览它
5. 我能不能继续跳到 replay

## 3. 基础卡片结构

建议一张卡片固定分成五块：

1. 顶部元信息条
2. 中部标题与摘要
3. 预览区域
4. 状态与关联信息区
5. 底部操作区

推荐结构：

```text
┌──────────────────────────────────────┐
│ Type / Surface / Status              │
├──────────────────────────────────────┤
│ Title                                │
│ Summary                              │
├──────────────────────────────────────┤
│ Preview Area                         │
├──────────────────────────────────────┤
│ Journey / Brain / Time / Artifact    │
├──────────────────────────────────────┤
│ Preview | Replay | Details           │
└──────────────────────────────────────┘
```

## 4. 顶部元信息条

### 4.1 必显信息

建议固定展示：

1. `evidence_type`
2. `surface`
3. `status`

### 4.2 状态徽标

建议：

1. `validated` 使用通过色
2. `collected` 使用处理中性高亮
3. `rejected` 使用警告色
4. `artifact_missing` 使用错误色
5. `artifact_pruned` 使用灰黄提示色

## 5. 标题与摘要区

### 5.1 标题

优先显示：

1. `title`

若无标题：

1. 使用 `evidence_type + journey_id` 兜底

### 5.2 摘要

显示：

1. `summary`

要求：

1. 限制为短摘要
2. 超长文本截断
3. 不在卡片内展开成长文

## 6. 预览区域

### 6.1 目标

让用户不离开 Acceptance 页就能初步判断证据是否可信。

### 6.2 按 `preview_kind` 切换表现

#### `image`

展示：

1. 缩略图
2. 点击后放大预览

#### `video`

展示：

1. 首帧缩略图
2. 视频时长标签
3. 点击进入预览或独立播放

#### `json`

展示：

1. 结构化摘要片段
2. 不直接展示完整 JSON

#### `text`

展示：

1. 若干行预览
2. 高亮关键行

#### `download_only`

展示：

1. 文件图标
2. “仅支持下载”文案

## 7. 关联信息区

建议展示：

1. `journey_id`
2. `source_brain`
3. `generated_at`
4. `artifact_state`

说明：

1. `source_brain` 仅用于解释证据来源归属
2. 不直接暴露原始运行时 payload 或底层工具细节

### 7.1 `artifact_state`

如果：

1. `available`
2. `artifact_missing`
3. `artifact_pruned`

都应单独显式显示，不应埋在摘要里。

## 8. 底部操作区

建议提供最多三个主操作：

1. `查看预览`
2. `查看回放`
3. `查看详情`

### 8.1 `查看预览`

仅当：

1. `preview_target` 可用
2. `artifact_state = available`

时可点。

### 8.2 `查看回放`

仅当：

1. `replay_available = true`

时显示。

### 8.3 `查看详情`

始终可用，用于打开 Evidence 详情抽屉或页内详情区。

## 9. 从卡片打开 Replay 的规则

点击“查看回放”时：

1. 不要求页面自己拼上下文
2. 直接使用 `replay-links` 或 `replay_target`
3. 打开 `Replay Drawer`
4. 自动定位到相关 `event_id / trace_id / replay_id`

## 10. 卡片尺寸与排列

### 10.1 桌面端

建议：

1. 双列或三列卡片网格
2. 卡片高度近似一致
3. 摘要和预览区做高度约束

### 10.2 移动端

建议：

1. 单列
2. 操作区改为更大触控按钮

## 11. 分类分组样式

Evidence 区按：

1. `browser`
2. `ci/build`
3. `runtime`
4. `verification`
5. `export`

分组时，卡片样式不应完全不同，只在分组标题和小图标上区分。

## 12. 降级与异常状态

### 12.1 `artifact_pruned`

表现：

1. 卡片仍显示
2. 预览区改为说明占位
3. “查看预览”禁用
4. “查看回放”若存在则保留

### 12.2 `artifact_missing`

表现：

1. 顶部状态条 warning 高亮
2. 预览区显示缺失提示
3. 详情区允许查看元信息

### 12.3 `rejected`

表现：

1. 不从列表中直接消失
2. 状态清晰标记
3. 原因在详情里解释

## 13. 不该怎么做

不建议：

1. 只展示文件名不展示 evidence 语义
2. 把回放入口埋进详情层
3. `artifact_pruned` 时直接隐藏整张卡
4. 不显示 `surface / journey`
5. 卡片操作超过三四个导致噪音过大

## 14. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Acceptance 页右侧 Evidence 区必须复用本卡片结构
2. Preview 交互设计必须围绕 `preview_kind` 分层
3. Replay Drawer 的 Evidence 回链入口必须从本卡片主操作触发

## 15. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Evidence Preview交互设计](./EasyMVP-V3-Evidence%20Preview%E4%BA%A4%E4%BA%92%E8%AE%BE%E8%AE%A1.md)
2. [EasyMVP-V3-Evidence详情抽屉设计](./EasyMVP-V3-Evidence详情抽屉设计.md)
3. Acceptance 线框图中的 Evidence 区实现规范
