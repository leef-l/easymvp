# EasyMVP 架构设计文档

> 版本：v2.0 | 更新日期：2026-04-02

---

## 1. 系统概述

EasyMVP 是一个 **AI 协作开发平台**，核心理念是让多个 AI 角色协同完成软件项目的全流程开发。

用户只需描述需求，系统自动完成：**需求分析 → 方案设计 → 任务拆分 → 代码编写 → 代码审计 → Bug 修复**。

### 核心架构

```
用户 ──→ 架构师 AI（需求分析/方案设计/任务拆分）
              │
              ▼
         任务调度器（依赖检查 + 资源冲突检测 + 并发控制）
              │
     ┌────────┼────────────┐
     ▼        ▼            ▼
 实施员 AI  实施员 AI   实施员 AI   ← 通过 Aider 真实编辑代码
     │        │            │
     ▼        ▼            ▼
 审计员 AI  审计员 AI   审计员 AI   ← 代码审查，发现 Bug 回报
     │
     └──→ Bug 闭环（架构师分析 → 实施员修复 → 审计员复查）
```

---

## 2. 技术栈

| 层级 | 技术选型 |
|------|---------|
| **后端框架** | Go 1.25 + GoFrame v2.10（MonoRepo 多应用架构） |
| **前端框架** | Vue 3 + Vben Admin v5.7 + Ant Design Vue + VxeTable |
| **数据库** | MySQL 8.0 |
| **ID 策略** | Snowflake 雪花 ID（`snowflake.JsonInt64`） |
| **认证** | JWT（自定义实现） |
| **AI 调用** | 统一 Provider 接口（支持 Anthropic / OpenAI 兼容协议） |
| **代码编辑** | Aider CLI（AI 代码编辑工具） |
| **实时推送** | Server-Sent Events (SSE) |
| **代码生成** | 自研 Codegen（Go text/template） |

---

## 3. 项目目录结构

```
easymvp/
├── admin-go/                              # 后端（Go MonoRepo）
│   ├── app/
│   │   ├── mvp/                           # MVP 核心模块
│   │   │   ├── api/mvp/v1/               # API 定义层
│   │   │   │   ├── chat.go               # 对话 API（手写）
│   │   │   │   ├── workflow.go           # 工作流 API（手写）
│   │   │   │   ├── project.go            # 项目 CRUD（codegen）
│   │   │   │   ├── task.go               # 任务 CRUD（codegen）
│   │   │   │   ├── conversation.go       # 对话 CRUD（codegen）
│   │   │   │   ├── message.go            # 消息 CRUD（codegen）
│   │   │   │   ├── project_role.go       # 项目角色 CRUD（codegen）
│   │   │   │   └── role_preset.go        # 角色预设 CRUD（codegen）
│   │   │   └── internal/
│   │   │       ├── engine/                # 核心执行引擎（全部手写）
│   │   │       │   ├── chat_engine.go     # 对话引擎（SSE 流式推送）
│   │   │       │   ├── scheduler.go       # 任务调度器（依赖+资源冲突）
│   │   │       │   ├── executor.go        # 任务执行器（ChatStream/Aider）
│   │   │       │   ├── aider_runner.go    # Aider CLI 封装
│   │   │       │   ├── task_parser.go     # 任务解析器（JSON 提取）
│   │   │       │   ├── context_compressor.go  # 上下文压缩器（三层策略）
│   │   │       │   ├── sse_hub.go         # SSE 推送中心
│   │   │       │   ├── watchdog.go        # 任务看门狗（心跳+自动重启）
│   │   │       │   ├── bugloop.go         # Bug 闭环处理
│   │   │       │   └── workflow.go        # 项目流程编排
│   │   │       ├── controller/            # HTTP 控制器
│   │   │       ├── logic/                 # 业务逻辑（codegen 生成）
│   │   │       ├── service/               # 服务接口
│   │   │       ├── model/                 # DTO 模型
│   │   │       ├── dao/                   # 数据访问层
│   │   │       └── middleware/            # 中间件
│   │   ├── ai/                            # AI 配置管理
│   │   │   └── (供应商/套餐/模型 CRUD)
│   │   └── system/                        # 系统管理
│   │       └── (用户/角色/部门/菜单 RBAC)
│   ├── utility/
│   │   ├── provider/                      # AI Provider 统一接口
│   │   │   ├── provider.go               # 接口定义
│   │   │   ├── factory.go                # 工厂模式（缓存实例）
│   │   │   ├── anthropic.go              # Anthropic 协议实现
│   │   │   └── openai.go                 # OpenAI 兼容协议实现
│   │   ├── snowflake/                     # 雪花 ID 生成器
│   │   ├── jwt/                           # JWT 令牌
│   │   └── response/                      # 统一响应格式
│   ├── codegen/                           # 代码生成器
│   └── deploy/                            # 部署配置
│
└── vue-vben-admin/                        # 管理端前端
    └── apps/web-antd/src/
        ├── views/mvp/                     # MVP 页面
        │   ├── chat/                      # 对话界面
        │   │   ├── index.vue              # 主页面（SSE 实时对话）
        │   │   └── components/            # ChatInput / ChatMessage
        │   ├── project/                   # 项目管理
        │   │   ├── index.vue              # 列表页（状态/进度/操作）
        │   │   └── modules/               # form / pause-modal / status-panel
        │   ├── task/                       # 任务列表
        │   ├── conversation/              # 对话历史
        │   ├── message/                   # 消息管理
        │   ├── role_preset/               # 角色预设模板
        │   └── project_role/              # 项目角色配置
        └── api/mvp/                       # API 调用层
            ├── project/                   # 项目 API
            ├── workflow/                  # 工作流 API
            └── ...
```

---

## 4. 核心模块设计

### 4.1 四大 AI 角色

| 角色 | 职责 | 执行方式 | 等级 |
|------|------|---------|------|
| **Architect（架构师）** | 需求分析、方案设计、任务拆分、Bug 分析 | ChatStream 对话 | max |
| **Implementer（实施员）** | 代码编写、文件修改、配置环境 | **Aider 代码编辑** | lite/pro/max |
| **Auditor（审计员）** | 代码审查、质量检测、Bug 发现 | ChatStream 对话 | lite/pro/max |
| **Coordinator（协调员）** | 进度协调、冲突解决、总结报告 | ChatStream 对话 | lite |

### 4.2 角色等级（Role Level）

| 等级 | 模型定位 | 上下文访问 | 适用场景 |
|------|---------|-----------|---------|
| **max** | 最强模型 | 全局摘要（~3000字） | 核心模块、复杂逻辑 |
| **pro** | 中等模型 | 依赖任务摘要 | 常规功能开发 |
| **lite** | 轻量模型 | 依赖任务摘要 | 简单任务、配置修改 |

### 4.3 项目状态机

```
                    用户创建项目
                         │
                         ▼
                    ┌──────────┐
                    │ designing │ ← 与架构师对话，拆分任务
                    └────┬─────┘
                         │ 用户确认方案
                         ▼
                    ┌──────────┐
              ┌────▶│ running  │ ← 任务并发执行中
              │     └────┬─────┘
              │          │
              │     ┌────┴────┐
              │     │         │
              │     ▼         ▼
              │ ┌────────┐ ┌───────────┐
              │ │ paused │ │ completed │
              │ └────┬───┘ └───────────┘
              │      │
              └──────┘ 用户恢复
```

### 4.4 任务状态机

```
draft ──→ pending ──→ running ──→ completed
                        │              │
                        │         (审计发现bug)
                        │              │
                        ▼              ▼
                      failed      bug_found
                        │              │
                   (自动重试)    (架构师分析)
                        │              │
                        ▼              ▼
                     pending       pending (追加修复指令)
```

---

## 5. 引擎模块详细设计

### 5.1 ChatEngine（对话引擎）

**文件：** `engine/chat_engine.go`

**职责：** 管理用户与 AI 的实时对话，支持 SSE 流式推送。

**核心流程：**

```
SendMessage(conversationID, content, userID)
  │
  ├─ 1. 查询对话 → 获取 projectID、roleType
  ├─ 2. resolveModel() → 三表联查获取模型配置
  │     mvp_project_role → ai_model → ai_plan + ai_provider
  │     返回: ModelInfo{ModelCode, ProviderType, BaseURL, APIKey, SystemPrompt}
  ├─ 3. 保存用户消息 (status=completed)
  ├─ 4. 创建 AI 回复消息 (status=streaming)
  └─ 5. goroutine → runAICall()
         ├─ 加载对话历史 (completed 消息，排除当前 streaming)
         ├─ provider.ChatStream() 流式调用
         ├─ 逐个 chunk:
         │   ├─ 写入 mvp_message_chunk (chunk_index, content)
         │   └─ SSEHub.Publish(replyID, chunkJSON)
         ├─ 更新消息 status=completed, content=fullContent
         ├─ SSEHub.Done(replyID)
         └─ tryParseArchitectTasks() → 异步解析任务清单
```

**ModelInfo 结构：**
```go
type ModelInfo struct {
    ModelID      int64
    ModelCode    string   // 如 "tc-code-latest", "glm-5"
    ProviderType string   // "anthropic" / "openai_compatible"
    BaseURL      string   // "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1"
    APIKey       string
    APISecret    string
    SystemPrompt string   // 角色提示词
    MaxTokens    int
}
```

### 5.2 TaskParser（任务解析器）

**文件：** `engine/task_parser.go`

**职责：** 从架构师 AI 回复中提取 JSON 任务清单，创建 draft 任务。

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
| content | longtext | 消息内容 |
| model_id | bigint FK | AI 模型（仅 assistant） |
| token_usage | json | Token 用量统计 |
| status | varchar(20) | completed/streaming/failed |

#### mvp_role_preset（角色预设模板）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 雪花 ID |
| role_type | varchar(20) | 角色类型 |
| role_level | varchar(10) | 角色等级 |
| model_id | bigint FK | 关联模型 |
| system_prompt | text | 系统提示词 |
| status | tinyint(1) | 启用状态 |
| sort | int | 排序 |

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
| **角色预设模板** | 新项目一键复用配置，降低使用门槛 |
| **Codegen + 手写分离** | CRUD 自动生成，核心逻辑手写，互不干扰 |
