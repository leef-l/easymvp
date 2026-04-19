# EasyMVP V3 replay 与 log artifact 存储规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3-本地目录与项目工作区规范](./EasyMVP-V3-本地目录与项目工作区规范.md)
> 关联文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
> 关联文档：[EasyMVP-V3-Replay索引表结构设计](./EasyMVP-V3-Replay索引表结构设计.md)
> 目标：定义 V3 单机版中 run 日志、replay 文件、checkpoint payload、结构化索引与页面读取边界，避免 logs、events、replay 各自成体系。

## 1. 设计结论

V3 中 `logs`、`event_history`、`replay`、`checkpoint` 必须分层存储。

正式原则应定为：

1. 原始日志落文件
2. 高价值事件索引落数据库
3. replay 明细落文件，索引落数据库
4. checkpoint 元数据落数据库，payload 可落文件
5. 页面只消费结构化索引和聚合接口，不直接扫 run 目录

一句话：

> replay 不是“更多日志”，而是围绕 `run_id` 组织的可追溯运行材料集合。

## 2. 为什么必须单独规范

如果不单独定义这一层，后面会出现：

1. 日志文件和事件索引混在一起
2. replay 页面直接读原始日志
3. checkpoint 和 replay 边界不清
4. 日志清理后页面和数据库同时失真
5. 同一个 run 的材料无法稳定归档和恢复

因此 `replay/log artifact` 必须成为正式存储专题。

## 3. 四类对象的边界

### 3.1 原始日志 `logs`

表示运行过程中直接产生的原始输出。

特点：

1. 体量可能大
2. 可读性强
3. 不直接等于页面主时间线

### 3.2 结构化事件 `event_history`

表示从运行中提炼出的结构化变化记录。

特点：

1. 面向工作台、审计、聚合
2. 体量相对小
3. 应优先落库

### 3.3 回放材料 `replay`

表示支持解释“运行时发生了什么”的结构化回放内容。

特点：

1. 比日志更可导航
2. 比事件更细
3. 面向 replay 页面与审计入口

### 3.4 状态快照 `checkpoint`

表示 resume 或状态恢复所需的快照。

特点：

1. 服务恢复
2. 不等于 replay
3. 不应直接当作页面时间线素材

## 4. 目录落点规则

建议统一使用：

```text
projects/{project_id}/runs/{run_id}/
  ├─ logs/
  ├─ checkpoints/
  ├─ artifacts/
  ├─ replay/
  └─ meta.json
```

### 4.1 `logs/`

用于：

1. 标准输出分片
2. 原始日志文件
3. 调试输出

### 4.2 `checkpoints/`

用于：

1. checkpoint payload
2. 状态恢复快照

### 4.3 `artifacts/`

用于：

1. run 直接产物
2. 中间结构化结果

### 4.4 `replay/`

用于：

1. 回放切片
2. 结构化步骤明细
3. 页面可读的原始材料索引文件

## 5. 文件命名建议

### 5.1 logs

建议：

```text
{ts}_log_{run_id}_{stream}_{seq}.log
```

示例：

1. `20260419T120311Z_log_run_88_stdout_0001.log`
2. `20260419T120312Z_log_run_88_stderr_0002.log`

### 5.2 replay

建议：

```text
{ts}_replay_{run_id}_{step_or_kind}_{seq}.json
```

示例：

1. `20260419T120501Z_replay_run_88_step_0001.json`
2. `20260419T120505Z_replay_run_88_tool-call_0002.json`

### 5.3 checkpoint

建议：

```text
{ts}_checkpoint_{run_id}_{checkpoint_id}.json
```

## 6. 数据库存什么

### 6.1 `workflow_brain_run_events`

用于存：

1. 高价值事件索引
2. 页面主时间线
3. ActionInbox 来源

### 6.2 replay 索引

建议后续增加专门索引表，至少存：

1. `replay_id`
2. `project_id`
3. `run_id`
4. `event_id`
5. `trace_id`
6. `span_id`
7. `replay_kind`
8. `file_path`
9. `created_at`

### 6.3 checkpoint 索引

已有 `workflow_brain_run_checkpoints`，建议至少存：

1. `checkpoint_id`
2. `run_id`
3. `checkpoint_kind`
4. `resume_supported`
5. `payload_ref`
6. `created_at`

## 7. 页面读取边界

页面不应：

1. 直接扫 `runs/{run_id}/logs/`
2. 直接把原始日志当时间线
3. 直接读 `checkpoint` 文件来解释运行过程

页面应：

1. 先读结构化事件索引
2. 再按需读取 replay 索引
3. 最后才打开原始日志或原始回放文件

## 8. Replay 展示层次与文件关系

### 8.1 摘要层

来源：

1. `workflow_brain_run_bindings`
2. `workflow_brain_run_events`

### 8.2 结构化事件层

来源：

1. `workflow_brain_run_events`
2. replay 索引表

### 8.3 原始回放层

来源：

1. `runs/{run_id}/replay/`
2. `runs/{run_id}/logs/`

## 9. 日志归档与切片

考虑到单次 run 可能很长，建议原始日志支持切片：

1. 按大小切片
2. 按时间切片
3. 按 stream 分片

数据库不存完整日志文本，只存：

1. 索引
2. 偏移
3. 文件路径
4. 起止时间

## 10. 与事件流的关系

原始日志不应直接推送到工作台主界面。

建议流程：

```text
raw log
  → parser / normalizer
  → structured event
  → workflow_brain_run_events
  → LiveEvent / ActionInbox / Replay Index
```

也就是说：

1. 工作台看事件
2. replay 看事件 + 回放材料
3. 原始日志只作为最底层追溯入口

## 11. 与 Evidence 的关系

并不是所有 log/replay 文件都是 Evidence。

只有当某个日志或回放材料被验收正式采纳时，才应映射成 Evidence 索引对象。

例如：

1. `runtime_log` 可从日志文件派生为 Evidence
2. `browser_trace` 可从 replay 文件派生为 Evidence
3. 普通调试日志不应自动进入 Acceptance evidence

## 12. 清理与 retention

### 12.1 默认长期保留

1. `workflow_brain_run_events`
2. 当前 Acceptance 引用的 replay/log 相关索引
3. 当前活跃或最近失败 run 的关键日志

### 12.2 可清理

1. 低价值原始日志分片
2. 可重建的 replay 中间文件
3. 已归档项目的非关键回放材料

### 12.3 清理后的状态

如果原始文件被清理：

1. 事件索引仍保留
2. replay 索引仍保留
3. 状态标记为 `artifact_pruned`
4. 页面上应提示“摘要可见，原始材料已清理”

## 13. 恢复与校验

恢复项目或 run 数据后，建议至少校验：

1. `run_id` 与目录匹配
2. replay 索引指向的文件是否存在
3. checkpoint `payload_ref` 是否可读
4. 日志切片顺序是否连续
5. 事件索引与原始文件是否存在明显断层

## 14. 不该怎么做

不建议：

1. 只存原始日志，不存结构化事件
2. 把 replay 页面做成纯日志查看器
3. 用 checkpoint 代替 replay
4. 清理原始文件时静默删掉事件索引
5. 让页面靠路径规则推断 run 状态

## 15. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. replay 页面设计必须区分摘要层、结构化事件层、原始回放层
2. `brain serve` 接入文档中的 logs/replay/checkpoint 必须复用本专题边界
3. 目录规范中的 `runs/{run_id}` 必须以本专题为 run 文件结构基线
4. Evidence 文档只在“被正式采纳”时把 replay/log material 映射成 Evidence

## 16. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Replay索引表结构设计](./EasyMVP-V3-Replay索引表结构设计.md)
2. log parser / normalizer 规则
3. Replay drawer 页面设计
4. 日志切片与按需加载策略
