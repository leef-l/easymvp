# EasyMVP 架构设计文档

> 更新日期：2026-04-08

本文描述当前仓库里的真实实现，而不是早期蓝图。当前主线以 `Workflow V2` 为准；API 中仍保留少量 `legacy` / `engineVersion` 兼容字段，但文档和前端主流程都以 V2 为基线。

## 1. 系统组成

EasyMVP 当前由四个主要部分组成：

| 组成 | 目录 | 职责 |
|------|------|------|
| `system` | `admin-go/app/system` | 用户、角色、菜单、部门、鉴权、RBAC |
| `ai` | `admin-go/app/ai` | AI 供应商、套餐、模型、执行引擎、手工任务 |
| `mvp` | `admin-go/app/mvp` | 项目、对话、Workflow V2、执行器、协作、自治 |
| `web` | `vue-vben-admin` | 管理端前端、工作流仪表盘、配置和协作页面 |

## 2. 仓库结构

```text
easymvp/
├── admin-go/
│   ├── app/
│   │   ├── system/
│   │   ├── ai/
│   │   └── mvp/
│   ├── utility/               # provider/jwt/response/snowflake 等公共能力
│   └── manifest/sql/          # 基线 schema 与增量迁移
├── vue-vben-admin/            # 前端
├── docker/                    # dev/prod compose 与 Dockerfile
└── docs/                      # 当前文档
```

`mvp` 服务内部又分成三层：

- `api` / `controller` / `logic`：HTTP 接口、CRUD 和页面数据接口
- `internal/workflow`：Workflow V2 的运行实体、阶段服务、调度器、执行器、验收、自治
- `internal/collab` / `internal/workspace` / `internal/engine`：协作接入、工作区隔离、对话与模型解析等支撑能力

## 3. MVP 核心数据模型

当前 Workflow V2 的核心表已经落在基线 schema 中：

| 实体 | 作用 |
|------|------|
| `mvp_project` | 项目业务容器、工作目录、分类、展示状态 |
| `mvp_workflow_run` | 一次正式工作流运行的权威状态 |
| `mvp_stage_run` | 设计/审核/执行/验收/返工/完成等阶段实例 |
| `mvp_plan_version` | 架构师输出的一版计划 |
| `mvp_task_blueprint` | 审核前的任务蓝图 |
| `mvp_domain_task` | 审核通过后实例化的执行期任务 |
| `mvp_review_issue` | 审核阶段问题 |
| `mvp_accept_run` / `mvp_accept_issue` / `mvp_accept_evidence` | 验收轮次、问题、证据 |
| `mvp_task_workspace` | 写仓任务的隔离工作区 |
| `mvp_workflow_event` | 工作流事件流与时间线数据 |
| `mvp_decision_action` / `mvp_human_checkpoint` | 自治决策与人工检查点 |
| `mvp_user_collab_binding` | 飞书 / Telegram 用户绑定 |

这套模型对应的基线 SQL 在：

- `admin-go/manifest/sql/mysql/000001_baseline_schema.up.sql`
- `admin-go/docker/mysql/init.sql`

## 4. 后端核心模块

### 4.1 工作流编排

`admin-go/app/mvp/internal/workflow/orchestrator/registry.go` 是当前工作流装配中心，负责：

- 初始化 `WorkflowService`、`StageService`
- 初始化执行器注册表
- 注册审核、执行、验收、返工、完成阶段之间的回调
- 初始化自治决策中台、事件总线、协作通知器

### 4.2 阶段服务

当前阶段服务分别位于：

- `stage/review`
- `stage/execute`
- `stage/accept`
- `stage/rework`
- `stage/complete`

职责分工：

- `review`：预检、审计员审核、协调员优化、问题落库
- `execute`：将蓝图实例化为 `domain_task` 并启动调度
- `accept`：证据收集、规则评估、裁决归并、推进完成或返工
- `rework`：围绕失败任务补建分析/修复链路
- `complete`：工作流闭环、汇总和最终完成态

### 4.3 执行器与工作区

执行器抽象在 `internal/workflow/executor`：

- `Registry`：按 `execution_mode` 注册和分发执行器
- 已注册模式：`aider`、`chat`、`openhands`、`claude_code`、`codex_cli`、`gemini_cli`、`auto`
- `Request` / `Result`：统一执行输入输出

写仓执行器与 `internal/workspace` 配合，为任务准备 `git worktree` 隔离目录。

### 4.4 对话与模型解析

`internal/engine` 仍承担当前主线里的几项基础职责：

- 架构师对话与 SSE 流输出
- 模型解析 `ResolveProjectModelInfo`
- 架构师任务拆解回写
- 上下文压缩和部分运行时辅助能力

也就是说，Workflow V2 已经把正式运行实体迁到 `internal/workflow`，但模型解析、对话流和部分兼容逻辑仍留在 `internal/engine`。

### 4.5 自治与协作

当前已落地的外围能力包括：

- `internal/workflow/autonomy`：决策中台、策略、风险闸门、目标层、态势感知、元认知
- `internal/collab`：飞书/Telegram 适配器、通知服务、长连接和回调处理
- `internal/worker`：异步后台任务

## 5. 前端结构

前端位于 `vue-vben-admin/apps/web-antd/src/views/mvp`，当前主要页面包括：

- `dashboard/`：MVP 系统概览与系统检查
- `project/`：项目管理
- `chat/`：与架构师对话
- `workflow/`：仪表盘、审核、执行、验收、返工、时间线、自治、飞书、Telegram
- `config/`：MVP 全局配置
- `project_category/`、`role_preset/`、`project_role/`：分类与角色配置
- `task/`、`task_log/`、`conversation/`、`message/`：任务与消息明细

当前工作流页不是单页 demo，而是已拆成多个独立面板：

- `dashboard.vue`
- `review.vue`
- `execution.vue`
- `accept.vue`
- `rework.vue`
- `autonomy.vue`
- `objective.vue`
- `situation.vue`
- `meta-cognition.vue`
- `timeline.vue`
- `feishu.vue`
- `telegram.vue`

## 6. 运行基础设施

### 6.1 数据与缓存

- 数据库：MySQL 8.0
- Redis：生产 compose 默认启用；开发环境分精简版和全量版两套 compose

### 6.2 AI 协议

AI 供应商通过 `admin-go/utility/provider` 统一抽象，当前覆盖：

- Anthropic 风格协议
- OpenAI 兼容协议

### 6.3 开发与部署

- `docker/dev/docker-compose.cn.yml`：默认开发入口
- `docker/dev/docker-compose.yml`：带 Redis 的全量开发 compose
- `docker/prod/docker-compose.yml`：当前生产 compose

## 7. 当前文档边界

2026-04-08 起，仓库文档只保留与当前代码仍对应的内容：

- 不再在主文档里保留迁移计划、Phase 路线图和旧链说明
- 已删除的历史设计稿如需查看，请从 `git` 历史追溯
- 对尚未落地的能力，文档会明确写成“当前边界”，不再写成“建议新增”

**架构师输出格式：**
```json
{
  "tasks": [
    {
      "name": "任务名称",
      "description": "详细描述",
      "role_level": "max/pro/lite",
      "role_type": "implementer",
      "batch_no": 1,
      "sort": 1,
      "affected_resources": ["src/api/user.go", "src/model/user.go"],
      "depends_on": ["前置任务名"],
      "parent": "父任务名（可选）"
    }
  ]
}
```

**解析策略（三级降级）：**
1. 从 ` ```json ``` ` 代码块提取
2. 查找 `{ "tasks": [...] }` 结构
3. 查找独立 JSON 数组 `[{...}]`

**JSON 清洗（cleanJSON）：**
- 移除 `//` 单行注释（区分字符串内的 URL）
- 移除 `/* */` 多行注释
- 移除尾随逗号 `,]` / `,}`

**三遍写入：**
1. 创建所有任务 → 建立 name→id 映射
2. 更新父子关系（parent_id）
3. 写入依赖关系表（mvp_task_dependency）

### 5.3 Scheduler（任务调度器）

**文件：** `engine/scheduler.go`

**职责：** 核心调度中枢，管理任务的生命周期。

**调度器状态：**
```go
type Scheduler struct {
    running        map[int64]bool       // 正在执行的任务 ID
    lockedRes      map[string]int64     // 已锁资源 → 占用任务 ID
    maxConcurrency int                  // 最大并发数（默认 20）
    projectCtx     map[int64]CancelFunc // 项目级取消函数
}
```

**调度循环（每 2 秒）：**
```
scheduleOnce(projectID):
  │
  ├─ 1. 检查并发度: len(running) >= 20 → 跳过
  ├─ 2. 查询 pending 任务 (ORDER BY batch_no ASC, sort ASC)
  └─ 3. 遍历每个任务:
       ├─ 跳过已在 running 的
       ├─ checkDependencies(): 所有 depends_on 必须 completed
       ├─ tryLockResources(): affected_resources 不能与其他 running 任务冲突
       ├─ 更新状态 pending → running
       └─ 启动 goroutine: executor.Execute(projectID, taskID)
```

**资源冲突检测原理：**
```
任务A: affected_resources = ["src/user.go", "src/auth.go"]
任务B: affected_resources = ["src/user.go", "src/order.go"]

→ "src/user.go" 冲突，B 必须等 A 完成后才能执行
```

**任务完成回调：**
```
OnTaskCompleted(projectID, taskID):
  ├─ 释放 running[taskID]
  ├─ 释放 lockedRes 中该任务的资源
  ├─ 检查批次是否全部完成 → 触发批次压缩
  ├─ 立即触发 scheduleOnce() → 可能启动依赖任务
  └─ 检查项目是否全部完成 → 标记 completed
```

### 5.4 Executor（任务执行器）

**文件：** `engine/executor.go`

**职责：** 执行单个任务，根据角色类型选择执行方式。

**分发策略：**
```
Execute(projectID, taskID):
  │
  ├─ 查询任务 → 获取 role_type, role_level
  ├─ resolveTaskModel() → 获取模型配置
  │   优先级: task.model_id > project_role(role_type + role_level)
  │
  ├─ role_type == "implementer"
  │   └─ executeWithAider()    ← Aider 真实代码编辑
  │
  └─ 其他角色
      └─ executeWithChat()     ← ChatStream 对话模式
```

**Aider 执行流程（executeWithAider）：**
```
1. 获取项目 work_dir（代码仓库路径）
2. 构建 taskPrompt（任务描述 + 前置任务结果）
3. 解析 affected_resources 为文件列表
4. 创建对话记录 + 保存指令消息
5. AiderRunner.RunTask() → 调用 Aider CLI
6. 保存 Aider 输出为 AI 回复消息
7. 成功 → completed + 压缩上下文 + 创建审计任务
8. 失败 → failTask() → 看门狗自动重试
```

**ChatStream 执行流程：**
```
1. 创建/获取任务对话
2. 保存指令消息
3. 加载对话历史
4. BuildTaskSystemPrompt() 增强 system prompt
5. provider.ChatStream() 流式调用
6. 逐个 chunk 写 message_chunk + SSE 推送
7. 成功 → completed + 压缩上下文 + 创建审计任务（如果是 implementer）
```

### 5.5 AiderRunner（Aider CLI 封装）

**文件：** `engine/aider_runner.go`

**职责：** 封装 Aider 命令行调用，让实施员 AI 能真正编辑代码文件。

**AiderConfig 配置：**
```go
type AiderConfig struct {
    ModelCode    string        // "tc-code-latest" / "glm-5" / "hunyuan-t1"
    APIKey       string
    BaseURL      string        // 不含 /v1
    ProviderType string        // "anthropic" / "openai_compatible"
    SystemPrompt string
    WorkDir      string        // 代码仓库根目录
    Files        []string      // 需要编辑的文件（来自 affected_resources）
    ReadFiles    []string      // 只读参考文件
    Message      string        // 任务指令
    AutoCommit   bool          // 默认 false
    Timeout      time.Duration // 默认 10 分钟
}
```

**生成的命令示例：**
```bash
ANTHROPIC_API_KEY=sk-xxx \
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic \
aider \
  --model anthropic/glm-5 \
  --no-auto-commits \
  --no-show-model-warnings \
  --yes-always \
  --no-pretty \
  --no-stream \
  --chat-language Chinese \
  --file src/api/user.go \
  --file src/model/user.go \
  --message "## 角色设定\n{systemPrompt}\n\n## 任务指令\n{taskPrompt}"
```

**环境变量映射：**
| Provider 类型 | 环境变量 |
|--------------|---------|
| `anthropic` | `ANTHROPIC_API_KEY` + `ANTHROPIC_BASE_URL` |
| `openai_compatible` | `OPENAI_API_KEY` + `OPENAI_API_BASE` |

### 5.6 ContextCompressor（上下文压缩器）

**文件：** `engine/context_compressor.go`

**职责：** 压缩 AI 对话和任务结果，控制上下文大小。

**三层压缩策略：**

```
第一层：单任务压缩
  < 500 字  → 原文保存（零 token 消耗）
  500~3000 → 规则截取（前 500 字 + 末 200 字）
  > 3000   → 调 AI 压缩为 200 字摘要

第二层：批次合并压缩
  同批次任务全部完成后
  → 收集所有 context_summary
  → 拼接为批次摘要
  → 合并进全局上下文

第三层：渐进式全局摘要（恒定 ~3000 字）
  合并后 < 3000 字 → 直接追加
  合并后 > 3000 字 → 调 AI 重新压缩整个全局上下文
  → 保证 global_context 恒定大小
```

**SystemPrompt 增强：**
```
architect / max 角色:
  basePrompt + 项目信息 + global_context (~3000字)

pro / lite 角色:
  basePrompt + 直接依赖任务的 context_summary
```

### 5.7 SSEHub（流式推送中心）

**文件：** `engine/sse_hub.go`

**职责：** 管理 AI 回复的实时推流到前端。

```go
type SSEHub struct {
    channels map[int64][]chan string  // messageID → 订阅者列表
}

// 前端连接 SSE
Subscribe(messageID) → (channel, unsubscribe)

// AI 产生 chunk 时
Publish(messageID, `{"content":"增量文本","index":1}`)

// AI 回复完成时
Done(messageID)  // 关闭所有订阅者 channel
```

**前端 SSE 端点处理：**
```
GET /chat/sse?messageID=xxx
  ├─ 消息已 completed → 直接返回全部内容
  ├─ 消息在 streaming → 先返回已有 chunks，然后订阅后续
  └─ 消息 failed → 返回错误信息
```

### 5.8 Watchdog（任务看门狗）

**文件：** `engine/watchdog.go`

**职责：** 自动检测卡死任务，自动重启失败任务。

**心跳检测（每 2 分钟）：**
```
遍历所有 running 任务
  ├─ 查询最新 chunk_id
  ├─ 与上次检测比较
  │   ├─ 有变化 → 重置计数
  │   └─ 无变化 → staleCount++
  └─ staleCount >= 3 → 判定卡死 → 标记 failed
```

**自动重启（每 2.5 分钟）：**
```
遍历所有 failed 任务
  ├─ 检查项目状态是否 running
  ├─ retryCount < 3 → 自动重启 (failed → pending)
  └─ retryCount >= 3 → 升级到 Bug 流程
       └─ ReportBug() → 创建架构师分析任务
```

### 5.9 BugLoop（Bug 闭环）

**文件：** `engine/bugloop.go`

**职责：** 审计员发现 Bug 后的闭环处理。

**Bug 处理流程：**
```
1. ReportBug(auditorTaskID, bugDescription)
   ├─ 审计任务: auditing → bug_found
   ├─ 实施员任务: completed → bug_found
   └─ 创建架构师分析任务 (status=pending)

2. 架构师 AI 分析 Bug（自动执行）
   └─ 输出修复方案

3. AutoDispatchBugFix(analysisTaskID)
   ├─ 实施员任务: bug_found → pending
   ├─ 追加修复指令到任务描述
   └─ 清空旧结果，触发重新调度

4. 实施员 AI 修复代码（Aider 执行）
   └─ completed

5. 自动创建审计任务 → 重新审查
```

---

## 6. AI Provider 统一接口

**文件：** `utility/provider/`

### 接口定义

```go
type Provider interface {
    Chat(ctx, req *ChatRequest) (*ChatResponse, error)        // 非流式
    ChatStream(ctx, req *ChatRequest, handler) error          // 流式
}

type Config struct {
    ProviderType string  // "anthropic" / "openai_compatible"
    BaseURL      string
    APIKey       string
    APISecret    string
}
```

### 支持的 Provider

| Provider 类型 | 覆盖模型 | API 格式 |
|--------------|---------|---------|
| `anthropic` | Claude Opus/Sonnet/Haiku, 腾讯云 Coding Plan 全系列 | Anthropic Messages API |
| `openai_compatible` | OpenAI, DeepSeek, Qwen, Doubao, GLM, Moonshot, Yi, Ollama | OpenAI Chat Completions API |

### 工厂模式

```go
// 缓存 key: "providerType:baseURL:apiKey"
// 相同配置不重复创建实例
GetProvider(cfg Config) (Provider, error)
```

---

## 7. 数据库设计

### 7.1 ER 关系图

```
ai_provider (供应商)
  └─ ai_plan (套餐/账户)
       └─ ai_model (模型)
            │
            ├─ mvp_role_preset (角色预设模板)
            └─ mvp_project_role (项目角色配置)
                 │
                 └─ mvp_project (项目)
                      ├─ mvp_conversation (对话)
                      │    └─ mvp_message (消息)
                      │         └─ mvp_message_chunk (流式分片)
                      │
                      └─ mvp_task (任务)
                           ├─ mvp_task_dependency (依赖关系)
                           └─ mvp_task_log (执行日志)
```

### 7.2 核心表结构

#### mvp_project（项目）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| name | varchar(200) | 项目名称 |
| project_category | varchar(50) | 项目分类（决定角色预设） |
| description | text | 项目简介 |
| status | varchar(20) | designing/running/paused/completed |
| pause_reason | text | 暂停原因 |
| global_context | text | 全局上下文摘要（~3000 字） |
| architect_model_id | bigint FK | 架构师 AI 模型 |
| work_dir | varchar(500) | 代码工作目录（Aider 执行路径） |
| created_by / dept_id | bigint | 数据隔离 |
| created_at / updated_at / deleted_at | datetime | 时间戳 |

#### mvp_task（任务）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| project_id | bigint FK | 所属项目 |
| parent_id | bigint | 父任务（0=顶层） |
| name | varchar(500) | 任务名称 |
| description | text | 详细描述 |
| role_type | varchar(20) | architect/implementer/auditor/coordinator |
| role_level | varchar(10) | lite/pro/max |
| model_id | bigint FK | 指定模型（可选，覆盖角色配置） |
| conversation_id | bigint FK | 关联对话 |
| status | varchar(20) | draft/pending/running/completed/failed/bug_found |
| batch_no | int | 执行批次 |
| sort | int | 同批次排序 |
| affected_resources | json | 涉及的文件列表 |
| depends_on | json | 依赖的任务名称 |
| result | longtext | 执行结果 |
| context_summary | text | 压缩摘要（<500 字） |
| error_message | text | 错误信息 |
| started_at / completed_at | datetime | 执行时间 |

#### mvp_task_dependency（任务依赖）
| 字段 | 类型 | 说明 |
|------|------|------|
| task_id | bigint | 任务 ID |
| depends_on_id | bigint | 依赖的任务 ID |

#### mvp_conversation（对话）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| project_id | bigint FK | 所属项目 |
| task_id | bigint | 关联任务（0=项目级对话） |
| title | varchar(200) | 对话标题 |
| role_type | varchar(20) | 角色类型 |
| status | varchar(20) | active/closed |

#### mvp_message（消息）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| conversation_id | bigint FK | 所属对话 |
| role | varchar(20) | user/assistant/system |
| message_type | varchar(30) | 消息类型（chat_user/chat_reply/task_prompt/task_reply/system_message/poison） |
| content | longtext | 消息内容 |
| model_id | bigint FK | AI 模型（仅 assistant） |
| token_usage | json | Token 用量统计 |
| status | varchar(20) | completed/streaming/failed |

#### mvp_role_preset（角色预设模板）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| project_category | varchar(50) | 项目分类（如"软件开发"） |
| role_type | varchar(20) | 角色类型 |
| role_level | varchar(10) | 角色等级 |
| model_id | bigint FK | 关联模型 |
| system_prompt | text | 系统提示词 |
| status | tinyint(1) | 启用状态 |
| sort | int | 排序 |

#### mvp_project_role（项目角色配置）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| project_id | bigint FK | 所属项目 |
| project_category | varchar(50) | 项目分类（冗余，便于查询） |
| role_type | varchar(20) | 角色类型 |
| role_level | varchar(10) | 角色等级 |
| model_id | bigint FK | AI 模型 |
| system_prompt | text | 系统提示词 |
| status | tinyint(1) | 启用状态 |

#### AI 配置表

**ai_provider（供应商）：** name, code, provider_type, base_url, icon, status

**ai_plan（套餐）：** provider_id, name, code, api_key, api_secret, status

**ai_model（模型）：** plan_id, provider_id, name, model_code, capability, max_tokens, context_window, supports_stream, role_prompt, status

---

## 8. 性能优化策略

| 优化点 | 策略 | 效果 |
|-------|------|------|
| **并发执行** | 最多 20 个 goroutine 并发，资源冲突检测防止写冲突 | 批次内任务并行，无数据竞争 |
| **上下文压缩** | 三层策略，规则优先，AI 兜底 | 全局摘要恒定 ~3000 字 |
| **Provider 缓存** | Factory 按配置缓存实例 | 避免重复初始化 HTTP Client |
| **SSE 推流** | channel 缓冲 100，已完成直接返回 | 零延迟展示历史消息 |
| **看门狗** | 2 分钟心跳，3 次无进展判定卡死 | 自动恢复卡死任务 |
| **批次压缩** | 批次完成后合并压缩，非逐任务 | 减少 AI 压缩调用次数 |

---

## 9. 安全设计

| 安全点 | 实现 |
|-------|------|
| 数据隔离 | `dept_id` + `created_by` 字段，中间件自动注入和过滤 |
| 软删除 | `deleted_at` 字段，所有查询自动排除已删除数据 |
| 密码加密 | SHA256（`gsha256.Encrypt`） |
| JWT 认证 | 自定义实现，Token 过期刷新 |
| API 权限 | RBAC 菜单权限按钮控制 |
| AI Key 保护 | 存储在 ai_plan 表，不暴露给前端 |

---

## 10. 关键技术决策

| 决策 | 原因 |
|------|------|
| **Aider CLI 而非自建 Tool Use** | Aider 成熟稳定，支持多模型，自带 Git 感知和多文件编辑 |
| **SSE 而非 WebSocket** | SSE 更简单，单向推送足够，自动重连 |
| **资源锁机制** | affected_resources 声明式冲突检测，比文件锁更灵活 |
| **三层上下文压缩** | 规则优先降低 token 消耗，AI 兜底保证质量 |
| **看门狗 + Bug 闭环** | 无人值守运行的核心保障，自动修复 > 人工干预 |
| **角色预设模板 + 项目分类** | 按分类维护不同 AI 配置，新项目一键复用，切换分类自动联动模型 |
| **Codegen + 手写分离** | CRUD 自动生成，核心逻辑手写，互不干扰 |
