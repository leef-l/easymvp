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
| implementer（实现者） | 代码编写与修改 | **多执行器**（见下表） | lite/pro/max |
| auditor（审计员） | 代码审查、Bug 检测 | ChatStream 对话 | pro |
| coordinator（协调者） | 进度跟踪、冲突解决 | ChatStream 对话 | pro |

**支持的执行器**（`ai_engine_config` 表配置）：

| engine_code | 执行器 | 工作空间 | 环境变量 | 说明 |
|-------------|--------|----------|----------|------|
| `aider` | Aider CLI | worktree | — | 默认代码编辑执行器 |
| `openhands` | OpenHands | worktree | — | Docker 沙箱执行，command_template 驱动 |
| `claude_code` | Claude Code CLI | worktree | ANTHROPIC_API_KEY | `claude -p --output-format json` |
| `codex_cli` | OpenAI Codex CLI | worktree | OPENAI_API_KEY | `codex --approval-mode full-auto` |
| `gemini_cli` | Google Gemini CLI | worktree | GEMINI_API_KEY | `gemini -p` |
| `chat` | ChatStream | 不需要 | — | 对话模式（非代码编辑角色） |

**角色预设与项目分类**：每个项目有 `projectCategory`（中文展示名如"软件开发"）和 `categoryCode`（英文编码如"software_dev"）。`mvp_role_preset` 按英文 `categoryCode` 存储默认角色模板（5种角色×9个分类=45条）。创建项目时，用户可选择预设写入 `mvp_project_role`；未选择时不写入，运行时通过 `presetutil` 包从三维提示词库（分类-角色-等级）动态生成内存虚拟记录回退。所有用到 AI 模型的地方都通过项目角色配置获取。

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
│   │   │       ├── engine/            # ★ 核心引擎（手写，详见下方）
│   │   │       └── workflow/          # ★ V2 工作流引擎（手写）
│   │   │           ├── orchestrator/  # 编排层（stage_service/workflow_service/recovery）
│   │   │           ├── stage/         # 各阶段实现（review/execute/rework/accept/complete）
│   │   │           ├── domain/plan/   # 计划版本领域服务
│   │   │           ├── executor/      # 执行器（aider/openhands/claude_code/codex_cli/gemini_cli/auto）
│   │   │           ├── repo/          # 仓库层（project_role_repo/role_preset_repo/plan_version_repo）
│   │   │           └── presetutil/    # 角色预设工具（三维提示词库：分类×角色×等级）
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
└── docs/                              # 设计文档（25 篇）
    ├── EasyMVP架构设计文档.md
    ├── EasyMVP对接OpenHands与Aider引擎设计实现文档.md  ← 引擎开发必读
    ├── WorkflowRun阶段化工作流引擎重构架构设计文档.md   ← V2 工作流必读
    └── ...
```

---

## 四、双引擎架构

### engine 包（admin-go/app/mvp/internal/engine/）

legacy + V2 共用的基础设施：

| 文件 | 职责 | 关键函数 |
|------|------|----------|
| `chat_engine.go` | 对话引擎 + SSE 流式 | `SendMessage()` `runAICall()` `tryParseArchitectTasks()` |
| `role_resolver.go` | 统一角色/模型解析 | `ResolveProjectRole()` `ResolveProjectModelInfo()` `ResolveProjectExecutionMode()` |
| `executor_dispatch.go` | 执行器分发 | `DispatchTask()` — 根据 execution_mode 选择执行器 |
| `executor_bridge.go` | V2 执行器桥接 | `ExecuteDomainTask()` — 连接 engine 和 workflow executor |
| `parser_extractor.go` | 从 AI 回复提取任务 | `ExtractAndNormalize()` |
| `context_compressor.go` | 上下文压缩 | `CompressTaskContext()` `CompressBatchContext()` |
| `review_precheck.go` | 审核预检 + AI 审核 | `RunReview()` `RunAuditorReviewForBlueprints()` |
| `config.go` | 引擎配置管理 | `GetConfig()` — DB→YAML→默认值三级回退 |
| `sse_hub.go` | SSE 消息推送中心 | `Publish()` `Subscribe()` `Done()` |
| `workflow_lifecycle.go` | 项目创建 | `CreateProject()` |

### workflow 包（admin-go/app/mvp/internal/workflow/）— V2 工作流引擎

| 子包 | 职责 |
|------|------|
| `orchestrator/` | 编排层：WorkflowService（创建/暂停/恢复）、StageService（阶段生命周期）、RecoverActiveWorkflows（启动恢复） |
| `stage/review/` | 审核阶段：precheck → auditor AI → coordinator 优化 → concludeReview |
| `stage/execute/` | 执行阶段：蓝图实例化为领域任务 + 调度 |
| `stage/rework/` | 返工阶段 |
| `stage/accept/` | 验收阶段 |
| `stage/complete/` | 收尾阶段 |
| `domain/plan/` | PlanVersionService：蓝图创建 + 提交审核 |
| `executor/` | CLI 执行器（aider/openhands/claude_code/codex_cli/gemini_cli/auto） |
| `repo/` | 仓库层：project_role_repo（角色配置+预设回退）、role_preset_repo、plan_version_repo |
| `presetutil/` | 三维提示词库（9分类×5角色×3等级），BuildRoleSystemPrompt() |

### V2 工作流生命周期（WorkflowRun 六阶段）

```
designing ──→ reviewing ──→ executing ──→ accepting ──→ completed
    ↑         ↓(驳回)         ↓(失败)       ↓(驳回)
    └─────────┘           reworking ←──────┘
                              ↓
                          (返工完成 → executing 或 accepting)
```

任何阶段都可以 → `paused`（人工暂停），`paused` 可恢复到暂停前的阶段。

### 领域任务（domain_task）生命周期

```
pending → running → completed → (审计) → accepted
            ↓                      ↓
          failed              bug_found → rework
```

### 调度策略

- **批次门控**：同批次任务并行执行，不同批次串行（batch_no=0 为紧急任务可随时插入）
- **资源锁定**：同一文件/模块不允许两个任务同时修改（`affected_resources` 冲突检测）
- **依赖检查**：任务的 `depends_on` 列表中所有任务必须 completed 才能开始
- **自动推进**：当前批次全部完成后自动进入下一批次

---

## 五、数据库设计（55 张表）

### 系统模块（9 表）
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
| `system_role_ai_engine` | 角色-执行引擎授权 |

### AI 模块（7 表）
| 表名 | 用途 |
|------|------|
| `ai_provider` | AI 供应商（openai/anthropic/deepseek...） |
| `ai_plan` | 套餐（含 API Key） |
| `ai_model` | AI 模型配置 |
| `ai_engine` | 执行引擎定义 |
| `ai_engine_config` | 执行引擎配置（5条：aider/openhands/claude_code/codex_cli/gemini_cli） |
| `ai_task` | AI 执行任务 |
| `ai_task_log` | AI 执行任务日志 |

### MVP 核心模块（14 表）
| 表名 | 用途 |
|------|------|
| `mvp_project` | 项目（status/project_category/category_code/work_dir） |
| `mvp_project_category` | 项目分类配置（category_code 英文编码 + display_name 中文展示名） |
| `mvp_project_role` | 项目角色配置（execution_mode 支持 6 种执行器） |
| `mvp_project_report` | 项目汇报 |
| `mvp_role_preset` | 角色预设模板（按英文 category_code 分类，5角色×9分类=45条） |
| `mvp_task` | 任务 |
| `mvp_task_blueprint` | 任务蓝图 |
| `mvp_task_dependency` | 任务依赖关系 |
| `mvp_task_log` | 任务操作日志 |
| `mvp_task_resource_lock` | 任务资源锁 |
| `mvp_task_workspace` | 任务工作空间（Git Worktree 隔离） |
| `mvp_conversation` | 对话 |
| `mvp_message` | 消息 |
| `mvp_message_chunk` | 消息分片（流式输出） |
| `mvp_config` | 引擎配置（38 项灰度开关 + 参数） |

### WorkflowRun 流水线模块（10 表）
| 表名 | 用途 |
|------|------|
| `mvp_workflow_run` | 工作流运行实例 |
| `mvp_workflow_event` | 工作流事件 |
| `mvp_stage_run` | 阶段运行实例 |
| `mvp_stage_task` | 阶段任务 |
| `mvp_domain_task` | 领域任务 |
| `mvp_plan_version` | 计划版本 |
| `mvp_review_issue` | 审核问题 |
| `mvp_handoff_record` | 交接记录 |
| `mvp_accept_run` | 自动验收运行 |
| `mvp_accept_evidence` | 验收证据 |
| `mvp_accept_issue` | 验收问题 |
| `mvp_accept_rule` | 验收规则定义 |

### 七层自治模块（11 表）
| 表名 | 层级 | 用途 |
|------|------|------|
| `mvp_autonomy_decision` | L4 | 自治决策记录 |
| `mvp_decision_action` | L4 | 决策动作记录 |
| `mvp_policy_rule` | L3 | 策略规则 |
| `mvp_risk_gate_rule` | L3 | 风险闸门规则 |
| `mvp_action_outcome` | L5 | 策略效果跟踪 |
| `mvp_human_checkpoint` | L3.5 | 人工介入节点 |
| `mvp_user_collab_binding` | L3.5 | 用户协作平台绑定 |
| `mvp_situation_snapshot` | L6 | 态势快照 |
| `mvp_observation_record` | L7 | 决策观测记录 |
| `mvp_assessment_result` | L7 | 系统评估结果 |
| `mvp_tune_recommendation` | L7 | 调参建议 |
| `mvp_learning_record` | L7 | EMA 学习记录 |

### 数据关系

```
mvp_project_category（9个分类，category_code + display_name）
  └── mvp_role_preset[]    （默认角色预设，按 category_code 分类）

mvp_project（project_category=中文, category_code=英文）
  ├── mvp_project_role[]   （角色配置，运行时可回退到 role_preset）
  ├── mvp_workflow_run[]   （V2 流水线实例）
  │   ├── mvp_stage_run[]  （阶段实例：design/review/execute/accept/rework/complete）
  │   │   ├── mvp_domain_task[]  （领域任务）
  │   │   └── mvp_stage_task[]   （阶段子任务：precheck/auditor_review/...）
  │   ├── mvp_plan_version[]     （方案版本：draft→active→superseded）
  │   │   └── mvp_task_blueprint[]（任务蓝图）
  │   ├── mvp_review_issue[]     （审核问题）
  │   └── mvp_workflow_event[]
  ├── mvp_conversation[]   （对话）
  │   └── mvp_message[]    （消息）
  └── mvp_task[]           （legacy 任务，旧引擎使用）

ai_provider → ai_plan → ai_model → mvp_project_role.model_id
ai_engine_config → executor 注册表（6 种执行器）
mvp_autonomy_decision → mvp_decision_action → mvp_observation_record → mvp_learning_record
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
- `mvp_project_category` 定义分类（`category_code` 英文编码 + `display_name` 中文展示名 + `family_code` 分类族：coding/creative/analysis）
- `mvp_role_preset` 按英文 `category_code` 存储默认角色模板（5种角色×9个分类=45条，`is_default=1`）
- `mvp_project` 同时存储 `project_category`（中文）和 `category_code`（英文）
- 创建项目时用户可选预设写入 `mvp_project_role`；未选择时运行时通过 `presetutil` 动态回退
- 所有用到 AI 模型的地方通过 `mvp_project_role` 获取（优先项目配置，回退分类默认预设），不直接引用 `ai_model`

---

## 九、模型分工约定

- **Opus 4.6** — 分析、规划、设计方案、把控全局
- **Sonnet 4.6** — 具体实施、上下文多/复杂的任务
- **Haiku 4.5** — 小任务、关联性不大的独立任务

## 十、开发必读文档

| 文档 | 何时读 |
|------|--------|
| `docs/EasyMVP架构设计文档.md` | 了解系统全局设计 |
| `docs/WorkflowRun阶段化工作流引擎重构架构设计文档.md` | **V2 工作流开发必读** |
| `docs/EasyMVP对接OpenHands与Aider引擎设计实现文档.md` | **接入 AI 引擎前必读** |
| `docs/执行器接入架构设计文档.md` | 新增执行器时 |
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

## 十三、铁律：数据权限必须实现

**所有业务实体必须具备完整的数据权限能力，不允许用 created_by 单点校验代替。**

### 权限模型（五级 DataScope）

| DataScope | 含义 | 过滤逻辑 |
|-----------|------|----------|
| 1 | 全部数据 | 无过滤（超管） |
| 2 | 本部门及以下 | `dept_id IN (本部门 + 所有子部门)` |
| 3 | 本部门 | `dept_id = 当前用户部门` |
| 4 | 仅本人 | `created_by = 当前用户` |
| 5 | 自定义 | `dept_id IN (system_role_dept 关联的部门列表)` |

### 强制规则

1. **所有业务表必须包含 `created_by` 和 `dept_id` 字段**
   - 独立业务对象（project、workflow_run 等）：创建时直接写入
   - 附属对象（stage_run、domain_task、accept_run 等）：从父对象继承
   - 禁止新增业务表时遗漏这两个字段

2. **ApplyDataScope 必须实现真实的五级过滤**
   - 读取用户角色的 `data_scope` 值
   - 按上表逻辑拼接 WHERE 条件
   - 禁止只按 `created_by` 过滤就声称"已实现数据权限"

3. **CheckOwnership 必须升级为 CheckProjectAccess**
   - 支持 owner（创建人）、dept_member（同部门）、admin（超管）三级
   - Workflow 控制器层所有接口必须使用项目作用域校验，不允许只校验 created_by

4. **Repo 层必须提供 scope 入口**
   - 查询方法需支持按 `created_by` / `dept_id` 过滤
   - 可通过 context 注入或显式参数传递

### 检查清单（每次新增实体时对照）

- [ ] 表有 `created_by` + `dept_id` 字段？
- [ ] 创建时从 context 或父对象继承权限字段？
- [ ] 列表查询经过 `ApplyDataScope`？
- [ ] 详情/更新/删除经过 `CheckProjectAccess`？

---

## 十四、铁律：数据库变更必须同步初始化 SQL，且不得引入中文乱码

**凡是数据库结构、索引、初始化数据、默认配置发生变更，必须同步到本地初始化 SQL，且必须保证其他电脑用 Docker 首次启动时中文内容不乱码。**

### 强制规则

1. **先改库，必须回写初始化文件**
   - 本地库执行过的 DDL / DML，如果会影响新环境初始化，必须同步回 `admin-go/docker/mysql/init.sql`
   - 纯表结构变更，必须同步回 `admin-go/docker/mysql/schema.sql`
   - 禁止只改本机数据库、不改初始化 SQL

2. **新增索引也属于数据库变更**
   - 索引、唯一约束、默认值、字符集、排序规则、种子数据，都算数据库变更
   - 不能因为“只是加索引”就跳过 `init.sql`

3. **字符集铁律：统一使用 utf8mb4**
   - 禁止把库、表、字段、连接串改成 `utf8`、`latin1` 或其他会导致中文风险的编码
   - MySQL Docker 配置必须保持 `utf8mb4` / `utf8mb4_unicode_ci`
   - 应用连接串必须显式带 `charset=utf8mb4`

4. **改库后必须考虑其他电脑的首次 Docker 启动**
   - 你的本机库能跑不算完成，必须保证别人 `docker compose up` 后也能拿到同样的库结构和初始数据
   - 任何可能导致中文乱码的改动都视为阻断问题，禁止提交

### 最低检查清单

- [ ] 本地数据库变更是否已同步到 `admin-go/docker/mysql/init.sql`？
- [ ] 表结构变更是否已同步到 `admin-go/docker/mysql/schema.sql`？
- [ ] MySQL Docker 字符集是否仍保持 `docker/dev/docker-compose.yml` 的 `utf8mb4` 配置？
- [ ] 应用数据库连接是否仍保持 `docker/dev/start-go-app.sh` 的 `charset=utf8mb4`？

---

## 十五、铁律：设计文档生命周期

**大改动前必须先写设计文档。**

流程：
1. 分析问题 → 写设计文档到 `docs/` 目录
2. 按文档实施改动
3. 编译通过 → git push

**删除策略**：
- **单次性 bug 修复、小功能**：实施完成并 push 后可删除设计文档
- **跨里程碑重构、架构改动、阶段性设计**：**不删除**，保留在 `docs/` 目录作为长期参考
- 判断标准：如果后续开发可能需要回溯该设计的上下文和决策，则保留

---

## 十六、铁律：人工最高权限

**项目默认全自动执行，人工拥有最高权限，可在任何阶段介入。**

### 核心原则

1. **默认自动化**：项目创建后全流程自动运行（设计→审核→执行→验收），人工无需参与
2. **随时可介入**：人工可以在任何阶段、任何状态下介入操作（恢复、暂停、重试、跳过、通过、驳回等）
3. **不能被锁死**：任何系统保护机制（熔断、暂停、审核拒绝等）都不能阻止人工操作
4. **操作入口必须可见**：所有状态下都要在 UI 上暴露对应的人工操作按钮

### 人工可执行的操作

| 阶段 | 可用操作 |
|------|---------|
| 设计中 | 确认方案 |
| 审核中 | 手动通过、手动驳回 |
| 执行中 | 暂停、重试任务、跳过任务、触发重规划 |
| 返工中 | 暂停、重试任务、跳过任务、触发重规划 |
| 已暂停 | 恢复执行、重试全部失败任务、跳过全部失败任务 |
| 验收中 | 手动放行、驳回返工 |
