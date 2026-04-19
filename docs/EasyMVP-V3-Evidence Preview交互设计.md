# EasyMVP V3 Evidence Preview 交互设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence卡片组件规范.md)
> 关联文档：[EasyMVP-V3-Evidence卡片查询接口设计](./EasyMVP-V3-Evidence卡片查询接口设计.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer页面设计.md)
> 目标：定义 V3 中 Evidence 从卡片进入预览后的交互流程、预览容器形态、不同 preview_kind 的行为、降级状态与和 Replay 的联动方式。

## 1. 设计结论

Evidence Preview 不应只是“打开原文件”。

正确做法是：

1. 为不同 `preview_kind` 提供统一预览容器
2. 默认保留 Acceptance 页面上下文
3. 允许从预览直接进入 Replay Drawer
4. 文件缺失或已清理时优雅降级

一句话：

> Preview 是 Evidence 卡片和 Replay Drawer 之间的中间解释层，不是简单附件查看器。

补充边界：

1. Preview 只消费 EasyMVP 归一化后的 preview DTO
2. 不直接透出 `brain-v3` 原始 payload、原始工具结果或裸本地文件路径

## 2. 打开方式

建议第一版优先使用右侧抽屉或大面板预览，而不是直接新开页面。

原因：

1. 保留 Acceptance 上下文
2. 便于在多张卡片之间快速切换
3. 能与 Replay Drawer 保持交互一致

### 2.1 桌面端

建议：

1. 右侧大抽屉
2. 宽度略小于 Replay Drawer

### 2.2 移动端

建议：

1. 全屏预览层

## 3. 预览容器结构

建议分成四块：

1. 顶部摘要条
2. 中部预览内容区
3. 底部元信息区
4. 底部操作区

推荐结构：

```text
┌─────────────────────────────────────────────┐
│ Preview Header                              │
├─────────────────────────────────────────────┤
│ Preview Content                             │
├─────────────────────────────────────────────┤
│ Evidence Meta                               │
├─────────────────────────────────────────────┤
│ Replay / Details / Download                 │
└─────────────────────────────────────────────┘
```

## 4. 顶部摘要条

建议展示：

1. `title`
2. `evidence_type`
3. `surface`
4. `journey_id`
5. `status`
6. `artifact_state`

## 5. 预览内容区

### 5.1 `image`

行为：

1. 直接展示大图
2. 支持缩放
3. 支持查看原尺寸

### 5.2 `video`

行为：

1. 内嵌播放器
2. 显示首帧
3. 支持播放、暂停、拖动

### 5.3 `json`

行为：

1. 结构化折叠展示
2. 默认不展示超深层节点
3. 支持复制 JSON 片段

### 5.4 `text`

行为：

1. 文本预览区
2. 默认显示前若干屏
3. 支持高亮关键段落

### 5.5 `download_only`

行为：

1. 不尝试预览
2. 显示文件摘要
3. 提供下载入口

## 6. 元信息区

建议展示：

1. `source_brain`
2. `generated_at`
3. `validated_at`
4. `file_ext`
5. `mime_type`
6. `file_size`
7. `sha256`

说明：

1. `source_brain` 仅用于展示证据来源归属
2. `browser / verifier / code` 应理解为 `brain-v3` 内置脑的归一化来源

## 7. 操作区

建议提供最多四个动作：

1. `查看回放`
2. `查看详情`
3. `下载`
4. `复制标识`

### 7.1 `查看回放`

仅在：

1. `replay_target` 或 `replay-links` 存在

时显示。

点击后：

1. 打开 `Replay Drawer`
2. 自动定位相关上下文

### 7.2 `查看详情`

用于进入更完整的 Evidence 详情区。

### 7.3 `下载`

仅在：

1. `download_allowed = true`
2. `artifact_state = available`

时显示为主操作。

### 7.4 `复制标识`

建议复制：

1. `evidence_id`
2. `run_id`
3. `trace_id` 若存在

## 8. 从卡片到预览的流程

建议交互链：

```text
Evidence Card
  → GET evidence/{id}/preview
  → open preview drawer
  → optional: open Replay Drawer
```

要求：

1. 卡片不自己拼 preview 内容
2. Preview 容器只消费 preview 接口返回
3. 如果上游采证命中 `unsupported / denied`，Preview 必须显示显式受限状态，而不是展示为空白成功页

## 9. 从预览到回放的流程

建议交互链：

```text
Preview Drawer
  → GET evidence/{id}/replay-links
  → open Replay Drawer
  → locate target event/trace/replay item
```

## 10. 状态与降级

### 10.1 `artifact_pruned`

表现：

1. 顶部状态条提示“附件已清理”
2. 内容区显示占位说明
3. 若 replay 可用，则保留“查看回放”

### 10.2 `artifact_missing`

表现：

1. 顶部 warning 高亮
2. 内容区显示“文件缺失”
3. 若 replay 可用，仍允许进入回放

### 10.3 `rejected`

表现：

1. 仍允许预览
2. 详情区提示它未被验收采纳

## 11. 大内容处理

### 11.1 大图

建议：

1. 默认适配容器
2. 支持放大

### 11.2 大文本 / 大 JSON

建议：

1. 默认截断
2. 标出 `truncated`
3. 支持继续加载或下载

### 11.3 长视频

建议：

1. 不自动播放
2. 先显示海报帧

## 12. 不该怎么做

不建议：

1. 直接在浏览器新标签页裸开本地文件
2. 不区分 `preview_kind`
3. 文件缺失时整个预览层报错
4. 把回放入口埋得很深

## 13. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Evidence 卡片点击预览时必须优先打开此预览容器
2. Replay Drawer 和 Preview Drawer 的交互应保持一致的右侧抽屉心智
3. 后续 Acceptance 线框图若包含预览层，应复用本专题结构

## 14. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Evidence详情抽屉设计](./EasyMVP-V3-Evidence详情抽屉设计.md)
2. 图片/视频预览细节规范
3. JSON/Text 大内容截断与展开策略
