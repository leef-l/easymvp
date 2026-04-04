# EasyMVP 项目指令

> AI 协作开发平台 —— 用多角色 AI 团队自动完成软件项目的需求分析、任务拆解、代码实现、质量审计全流程。

## 一、系统全景

EasyMVP 是一个 **AI 驱动的项目管理与代码自动化引擎**。核心流程：

```
用户创建项目 → 架构师AI对话拆任务 → 确认方案 → 调度器批次执行 → 实现/审计/修复闭环 → 项目完成
```

**四大 AI 角色**：

| 角色 | 职责 | 执行方式 | 等级 |
|------|------|----------|------|
| architect（架构师） | 需求分析、方案设计、任务拆解、故障分析 | ChatStream 对话 | max |
| implementer（实现者） | 代码编写与修改 | **Aider CLI** 真实编辑代码 | lite/pro/max |
| auditor（审计员） | 代码审查、Bug 检测 | ChatStream 对话 | pro |
| coordinator（协调者） | 进度跟踪、冲突解决 | ChatStream 对话 | pro |

**角色预设与项目分类**：每个项目有 `projectCategory`（如"软件开发"），创建时按分类从 `mvp_role_preset` 加载对应的角色模板（含 AI 模型、系统提示词），写入 `mvp_project_role`。所有用到 AI 模型的地方都通过项目角色配置获取。

---

## 二、技术栈

### 后端
- **语言**: Go 1.25
- **框架**: GoFrame v2.10（MonoRepo 多应用架构）
- **数据库**: MySQL 8.0（本机 127.0.0.1:3306）
- **认证**: JWT（`utility/jwt`）
- **ID策略**: Snowflake 雪花ID（`utility/snowflake`，类型 `snowflake.JsonInt64`）
- **实时推送**: SSE（Server-Sent Events）

### 管理端前端
- **框架**: Vue 3 + Vben Admin v5.7（`vue-vben-admin/apps/web-antd`）
- **UI库**: Ant Design Vue
- **表格**: VxeTable（`useVbenVxeGrid`）
- **表单**: VbenForm（支持自定义组件）
- **构建**: Vite + pnpm

### 代码生成器
- **位置**: `admin-go/codegen/`
- **模板**: Go `text/template`（`codegen/templates/`）

---

## 三、项目目录结构

```
easymvp/
├── admin-go/                          # 后端 Go MonoRepo
│   ├── app/
│   │   ├── system/                    # 系统管理（用户、角色、部门、菜单、RBAC）
│   │   ├── ai/                        # AI 配置（供应商、套餐、模型）
│   │   ├── mvp/                       # 核心业务（项目、任务、对话、消息、引擎）
│   │   │   ├── api/mvp/v1/            # API 定义（请求/响应结构体）
│   │   │   │   ├── chat.go            # 对话 API（发送消息、SSE 流、历史）
│   │   │   │   ├── workflow.go        # 工作流 API（手写，核心业务）
│   │   │   │   └── *.go              # CRUD API（codegen 生成）
│   │   │   └── internal/
│   │   │       ├── controller/        # HTTP 控制器（按实体分目录）
│   │   │       │   └── chat/
│   │   │       │       ├── chat.go    # 对话控制器（手写）
│   │   │       │       └── workflow.go # 工作流控制器（手写）
│   │   │       ├── logic/             # 业务逻辑层（codegen 生成的 CRUD）
│   │   │       ├── model/             # 数据模型
│   │   │       │   ├── entity/        # 数据库实体（gf gen dao 生成）
│   │   │       │   └── dto/           # 输入输出 DTO
│   │   │       ├── dao/               # 数据访问对象（gf gen dao 生成）
│   │   │       └── engine/            # ★ 核心引擎（手写，详见下方）
│   │   ├── app/                       # 应用入口配置
│   │   ├── svc-template/              # 新服务模板
│   │   └── job-template/              # 定时任务模板
│   ├── codegen/                       # 代码生成器
│   │   ├── templates/                 # 生成模板（修 bug 先改这里）
│   │   └── sql/                       # 数据库初始化 SQL
│   ├── utility/                       # 公共工具（jwt/snowflake/response/oplog）
│   └── deploy/                        # 部署配置
│
├── vue-vben-admin/                    # 管理端前端
│   └── apps/web-antd/src/
│       ├── api/mvp/                   # API 调用层（按模块）
│       │   ├── workflow/index.ts      # 工作流 API（手写）
│       │   └── */                     # CRUD API（codegen 生成）
│       ├── views/mvp/                 # 页面
│       │   ├── chat/                  # 实时对话页（SSE 流式）
│       │   ├── project/               # 项目管理（创建/编辑/状态面板）
│       │   ├── task/                  # 任务列表与树形视图
│       │   ├── dashboard/             # 项目仪表盘
│       │   ├── consts.ts              # 常量（角色类型/等级/分类选项）
│       │   └── */                     # 其他 CRUD 页面
│       └── router/                    # 路由（动态加载）
│
└── docs/                              # 设计文档（9 篇）
    ├── EasyMVP架构设计文档.md
    ├── EasyMVP对接OpenHands与Aider引擎设计实现文档.md  ← 引擎开发必读
    ├── EasyMVP对接Aider实现文档.md
    └── ...
```

---

## 四、核心引擎（admin-go/app/mvp/internal/engine/）

这是系统的心脏，约 4,400 行 Go 代码，14 个文件：

| 文件 | 职责 | 关键函数 |
|------|------|----------|
| `workflow.go` | 项目生命周期状态机 | `CreateProject()` `ConfirmPlan()` `Pause()` `Resume()` |
| `scheduler.go` | 事件驱动任务调度器 | `StartProject()` `PauseProject()` `OnTaskCompleted()` `SkipTask()` |
| `executor.go` | 任务执行分发 | `Execute()` — implementer→Aider, 其他→ChatStream |
| `chat_engine.go` | 对话引擎 + SSE 流式 | `SendMessage()` `runAICall()` `tryParseArchitectTasks()` |
| `aider_runner.go` | Aider CLI 封装 | `Run()` `RunTask()` `buildArgs()` `buildEnv()` |
| `task_parser.go` | 从 AI 回复提取任务 JSON | `ParseAndCreateTasks()` `ConfirmDraftTasks()` |
| `context_compressor.go` | 三级上下文压缩 | `CompressTaskContext()` `CompressBatchContext()` `CompressProjectContext()` |
| `watchdog.go` | 心跳检测 + 自动恢复 | `Start()` `checkRunningTasks()` `checkFailedTasks()` |
| `bugloop.go` | Bug 闭环修复流程 | `ReportBug()` `AutoDispatchBugFix()` `EscalateFailedTask()` |
| `sse_hub.go` | SSE 消息推送中心 | `Publish()` `Subscribe()` `Done()` |
| `error_category.go` | 错误分类（planning/execution/policy） | `Categorize()` |
| `config.go` | 引擎配置管理 | `GetConfig()` — DB→YAML→默认值三级回退 |
| `path_validation.go` | 工作目录安全校验 | `ValidateWorkDir()` `EnsureWorkDir()` |
| `file_reader.go` | 安全文件读取（"读取：路径"指令） | `ExpandFileDirectives()` |

### 项目生命周期

```
designing ──→ running ──→ completed
    ↑    confirm    ↓
    │              pause
    │               ↓
    └───── paused ──┘
            resume
```

**注意**：项目只有 `paused` 没有 `failed` 状态。任务失败时项目暂停并记录原因。

### 任务生命周期

```
draft → pending → running → completed
                    ↓           ↓
                  failed    auditing → bug_found → bug_dispatched → pending(重做)
                    ↓
              (watchdog 自动重试 ≤3 次，超限升级到架构师)
```

### 调度策略

- **批次门控**：同批次任务并行执行，不同批次串行（batch_no=0 为紧急任务可随时插入）
- **资源锁定**：同一文件/模块不允许两个任务同时修改（`affected_resources` 冲突检测）
- **依赖检查**：任务的 `depends_on` 列表中所有任务必须 completed 才能开始
- **自动推进**：当前批次全部完成后自动进入下一批次

### 三级上下文压缩

1. **单任务压缩**：<500字符保留原文，500-3000规则截断，>3000调AI压缩
2. **批次合并**：批次完成后合并所有任务摘要为批次总结
3. **全局渐进式摘要**：维持 `global_context` ≈3000字符，超限时AI重新压缩

### Bug 闭环流程

```
审计员发现 Bug → ReportBug() → 创建架构师分析任务(batch_no=0)
    → 架构师分析完毕 → AutoDispatchBugFix() → 更新实现任务描述
    → 调度器重新执行实现任务 → 再次审计 → 循环直到通过（上限3轮）
```

---

## 五、数据库设计（19 张表）

### 系统模块（8 表）
| 表名 | 用途 |
|------|------|
| `system_users` | 用户账号 |
| `system_role` | RBAC 角色（支持层级） |
| `system_menu` | 菜单与权限标识 |
| `system_dept` | 部门组织（树形） |
| `system_user_role` | 用户-角色关联 |
| `system_user_dept` | 用户-部门关联 |
| `system_role_menu` | 角色-菜单权限 |
| `system_role_dept` | 角色-部门数据范围 |

### AI 模块（3 表）
| 表名 | 用途 | 关键字段 |
|------|------|----------|
| `ai_provider` | AI 供应商 | code(openai/anthropic/deepseek...), base_url, provider_type |
| `ai_plan` | 套餐（含 API Key） | provider_id, api_key, api_secret |
| `ai_model` | AI 模型配置 | plan_id, model_code, capability(chat/reasoning/coding), context_window, role_prompt |

### MVP 模块（8 表）
| 表名 | 用途 | 关键字段 |
|------|------|----------|
| `mvp_project` | 项目 | status, project_category, global_context, architect_model_id, work_dir |
| `mvp_task` | 任务 | project_id, role_type, role_level, batch_no, affected_resources, depends_on, context_summary |
| `mvp_conversation` | 对话 | project_id, task_id(可空=项目级), role_type |
| `mvp_message` | 消息 | conversation_id, role(user/assistant/system), message_type, token_usage |
| `mvp_project_role` | 项目角色配置 | project_id, project_category, role_type, role_level, model_id, system_prompt |
| `mvp_role_preset` | 角色预设模板 | project_category, role_type, role_level, model_id, system_prompt, sort |
| `mvp_config` | 引擎配置 | config_key, config_value, category(engine/watchdog/scheduler) |
| `mvp_task_log` | 任务操作日志 | task_id, action, from_status, to_status, operator |

### 数据关系

```
mvp_project
  ├── mvp_project_role[]  （项目的 AI 角色配置，从 mvp_role_preset 初始化）
  ├── mvp_conversation[]  （架构师对话 + 各任务对话）
  │   └── mvp_message[]   （用户/AI 消息，支持 SSE 流式）
  └── mvp_task[]          （80-200个任务，分批次执行）
      └── mvp_task_log[]  （状态变更审计日志）

mvp_role_preset → 按 project_category 分组的默认角色模板
ai_provider → ai_plan → ai_model → 被 mvp_project_role.model_id 引用
```

---

## 六、API 端点

### 手写业务 API（核心）

| 方法 | 路径 | 用途 |
|------|------|------|
| POST | `/mvp/workflow/create-project` | 创建项目（自动按分类初始化角色） |
| POST | `/mvp/workflow/confirm-plan` | 确认方案，启动调度器 |
| POST | `/mvp/workflow/pause` | 暂停项目 |
| POST | `/mvp/workflow/resume` | 恢复执行 |
| POST | `/mvp/workflow/retry-task` | 重试失败任务 |
| POST | `/mvp/workflow/skip-task` | 跳过任务（解除批次阻塞） |
| GET | `/mvp/workflow/project-status` | 项目实时状态 |
| POST | `/mvp/workflow/parse-tasks` | 手动解析架构师回复中的任务 |
| GET | `/mvp/workflow/role-presets` | 获取角色预设（按分类过滤） |
| GET | `/mvp/workflow/system-check` | 系统健康检测 |
| POST | `/mvp/chat/send` | 发送对话消息 |
| GET | `/mvp/chat/sse` | SSE 流式接收 AI 回复 |
| GET | `/mvp/chat/history` | 对话历史 |

### Codegen 生成的 CRUD API
每个实体（project/task/conversation/message/project_role/role_preset/config/task_log）都有标准的：
`Create / Update / Delete / BatchDelete / List / Detail / Export / Import / DownloadTemplate / BatchUpdate`

---

## 七、引擎配置参数

| 配置键 | 默认值 | 说明 |
|--------|--------|------|
| `scheduler.max_concurrent` | 20 | 最大并行任务数 |
| `scheduler.poll_interval` | 2s | 调度轮询间隔 |
| `watchdog.check_interval` | 120s | 心跳检测间隔 |
| `watchdog.max_stale_count` | 3 | 无心跳次数阈值 |
| `watchdog.max_retries` | 3 | 自动重试上限 |
| `runtime.task_timeout_seconds` | 600s | Aider 执行超时 |
| `failure_handoff.max_rounds` | 3 | Bug 修复最大轮次 |

---

## 八、重要约定

### 后端约定
- 所有 ID 字段使用 `snowflake.JsonInt64` 类型
- 软删除：`deleted_at` 字段
- 数据隔离：`dept_id` + `created_by`
- 密码加密：SHA256（`gsha256.Encrypt`）

### 代码分层规则
- **手写业务代码**放独立文件（`controller/chat/workflow.go`、`engine/*.go`），不会被 codegen 覆盖
- **生成代码有问题 → 先修 `codegen/templates/` 模板再重新生成**，不手写修复生成文件

### 项目分类与角色预设
- `mvp_role_preset` 按 `project_category` 存储默认角色模板
- 创建项目时按 `projectCategory` 过滤预设，初始化 `mvp_project_role`
- 所有用到 AI 模型的地方通过 `mvp_project_role` 获取，不直接引用 `ai_model`

---

## 九、模型分工约定

- **Opus 4.6** — 分析、规划、设计方案、把控全局
- **Sonnet 4.6** — 具体实施、上下文多/复杂的任务
- **Haiku 4.5** — 小任务、关联性不大的独立任务

## 十、开发必读文档

| 文档 | 何时读 |
|------|--------|
| `docs/EasyMVP架构设计文档.md` | 了解系统全局设计 |
| `docs/EasyMVP对接OpenHands与Aider引擎设计实现文档.md` | **接入 AI 引擎前必读** |
| `docs/EasyMVP对接Aider实现文档.md` | 修改 Aider 集成时 |
| `docs/EasyMVP使用文档.md` | 了解用户操作流程 |

---

## 十一、数据库连接

```bash
mysql -u easymvp -pJKcHFJYXnkrB6BXE -h 127.0.0.1 -P 3306 easymvp
```

## 十二、常用命令

```bash
# 后端编译
cd admin-go && go build ./app/mvp/...
cd admin-go && go build ./app/system/...
cd admin-go && go build ./app/ai/...

# 生成 DAO
cd admin-go/app/mvp && gf gen dao

# 代码生成器
cd admin-go/codegen && go run . -table mvp_xxx -force -menu

# 前端（需手动执行，AI 环境禁止运行 pnpm/npm/yarn）
cd vue-vben-admin && pnpm dev
```
