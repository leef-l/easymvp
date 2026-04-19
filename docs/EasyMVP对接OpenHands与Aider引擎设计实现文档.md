# EasyMVP 对接 OpenHands 与 Aider 引擎设计实现文档

> 更新日期：2026-04-08

本文只覆盖当前代码里已经落地的 `Aider` / `OpenHands` 配置与运行路径，以及它们和 MVP 工作流的关系。

## 1. 定位

`Aider` 与 `OpenHands` 现在有两条使用方式：

### 1.1 AI 模块手工任务

适合做：

- 引擎连通性验证
- 仓库路径与命令模板验证
- 一次性手工任务执行

### 1.2 MVP 工作流执行器

适合做：

- `mvp_domain_task` 的正式执行
- 审核通过后的项目任务调度
- 工作流里的写仓任务落地

这两条链路共享引擎配置和模型数据，但不是同一段运行时代码。

## 2. 当前数据模型

AI 模块里与执行引擎相关的核心表包括：

- `ai_engine`
- `ai_engine_config`
- `ai_task`
- `ai_task_log`

其中：

- `ai_engine` 保存引擎定义
- `ai_engine_config` 保存命令模板、超时、默认模型、工作区等
- `ai_task` 保存手工执行任务
- `ai_task_log` 保存运行日志

MVP 工作流侧还会额外使用：

- `mvp_domain_task`
- `mvp_task_workspace`

## 3. Aider 当前运行路径

AI 模块中的 `Aider` 运行链路位于 `admin-go/app/ai/internal/logic/task/runtime.go`，当前行为是：

1. 读取 `ai_engine_config(engine_code=aider)`
2. 解析默认模型与供应商密钥
3. 优先尝试本机 `aider`
4. 缺失时尝试 `uv`
5. 仍不可用时回退 Docker 方式

当前还有两个实现细节值得记录：

- 超时时间来自 `timeout_seconds`
- 命中 token limit 时会做一次精简上下文重试

## 4. OpenHands 当前运行路径

`OpenHands` 现在同时支持两类执行方式：

### 4.1 CLI / uv

优先走本机 `openhands` 或 `uv` 路径，适合已有本地运行环境的场景。

### 4.2 HTTP

如果 CLI 路径不可用，则会退回到 `base_url` 驱动的 HTTP 请求模式。

当前文档更新时确认到的实现特点：

- CLI 与 HTTP 共用 `ai_engine_config`
- 运行前会采集仓库基线信息
- 日志与结果会回写到 `ai_task` / `ai_task_log`

## 5. 与 MVP 工作流的关系

### 5.1 配置共享

MVP 工作流会复用：

- AI 模型
- 供应商配置
- `ai_engine_config`

因此系统概览页会直接检查：

- Aider 引擎配置
- OpenHands 引擎配置
- Aider 可执行环境
- OpenHands 可执行环境

### 5.2 运行入口不同

AI 模块手工任务通过：

- `POST /api/ai/task/execute`

MVP 工作流则通过：

- `mvp_domain_task.execution_mode`
- `executor.Registry`

来决定是否落到 `AiderExecutor` 或 `OpenHandsExecutor`。

## 6. 推荐配置顺序

当前实际可行的配置顺序是：

1. 配置 AI 供应商
2. 配置套餐 / 模型
3. 配置 `Aider` / `OpenHands` 引擎
4. 在 `AI 管理 -> 任务` 做一次手工冒烟验证
5. 再进入 `MVP -> 角色预设` 和项目工作流

## 7. 当前边界

更新本文档时确认到的边界如下：

- AI 模块里的手工任务运行时，当前一等支持的仍然是 `Aider` / `OpenHands`
- `Claude Code` / `Codex CLI` / `Gemini CLI` 已进入 MVP 执行器注册表，但还不是 AI 模块手工任务页里的同级主入口
- 文档不再写“建议新增”或“未来再做”的引擎方案，未落地能力不在此文档中保留
