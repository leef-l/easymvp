# 下一代 AI 协作开发平台方案

> 代号：**Atlas**（暂定）
> 定位：**不是 EasyMVP 的改进版，而是吸收 EasyMVP 核心思想、融合工业级工作流引擎工程实践、站在 ChatDev/MetaGPT/AutoGen/Temporal/Argo 肩膀上的新一代 AI 协作开发平台**
> 作者视角：假设从零开始，但继承 EasyMVP 已经被项目实战验证过的产品思想
> 写作日期：2026-04-10
> 调研依据：本方案所有关键设计决策均对齐已完成的两份开源同类调研报告
>
> 本文是**独立的完整方案**，可脱离 EasyMVP 现有代码单独阅读。与 EasyMVP 的继承/差异关系集中在第二章说明。

---

## 目录

0. [为什么要重做](#0)
1. [产品愿景与设计哲学](#1)
2. [从 EasyMVP 继承什么、抛弃什么](#2)
3. [顶层架构](#3)
4. [核心概念模型](#4)
5. [工作流引擎：状态机分层设计](#5)
6. [角色体系与三维预设](#6)
7. [验收与归档：双层结构](#7)
8. [七层自治与人工介入](#8)
9. [执行器抽象层](#9)
10. [数据层设计](#10)
11. [六边形架构分层](#11)
12. [可观测性与治理](#12)
13. [从 EasyMVP 迁移的路径与 pss_claude 落地清单](#13)
14. [与 ChatDev/MetaGPT/AutoGen 的差异化定位](#14)
15. [风险、权衡、已知限制](#15)

附录 A：术语表
附录 B：调研证据索引
附录 C：FAQ

---

<a id="0"></a>
## 0. 为什么要重做

EasyMVP 已经把一件难事做对了：**用多 AI 角色自动完成软件项目的设计-审核-执行-验收全流程**。这件事在 ChatDev、MetaGPT、CrewAI、AutoGen 里各有探索，但 EasyMVP 是**少数把五角色、三维预设、六阶段工作流、硬规则+LLM 融合判官、七层自治同时落地到生产**的项目。

但它在工程层有五个结构性缺陷，是"项目总是卡、反复调整修复"的根因：

1. **编排散在 6 个包**，改一个状态流转要改 4 个文件
2. **配置双套并行**（V2 英文集 + Legacy 中文集），V2 集 36 条默认预设 `model_id=0` 全部空壳
3. **三套状态机互相打架**（WorkflowRun 9 态 / StageRun 5 态 / DomainTask 8 态），互译逻辑手写
4. **数据层零外键零 CHECK**，55 表全靠应用层兜底一致性
5. **锁内调 AI**（`context_compressor.go:192-226` 持项目级锁期间发起 LLM 调用）

这些都不是"加个字段"能修的——它们是**设计时就没区分业务语义和工程边界**导致的。

所以这份方案**不是增量治理**，而是**基于 EasyMVP 已验证的产品思想，从零重建工程层**。产品逻辑继承，工程结构重建。下一节详述这条分界线。

---

<a id="1"></a>
## 1. 产品愿景与设计哲学

### 1.1 愿景

**让一个人以产品经理的姿态，指挥一支 AI 团队，完成真实可交付的软件项目。**

不是"帮人写代码"，不是"帮人改 bug"——是**完成项目**。这意味着它必须承担：
- 需求理解与拆解
- 方案评审与风险识别
- 并发执行与资源调度
- 质量验收与体验评审
- 归档、清理、通知

这是 ChatDev 的野心，是 MetaGPT 的野心，EasyMVP 的野心也是这个。本方案继承这个野心。

### 1.2 六条设计哲学

1. **产品抽象不变，工程结构重建**
   EasyMVP 已经验证过"五角色×九分类×三等级"的产品抽象是成立的，本方案直接继承；但底层的包结构、状态机、数据模型全部重建。

2. **状态机是一等公民，不是实现细节**
   Temporal 官方博客原话："workflow 开发者经常把状态机当成实现细节藏起来——这是错误的"。本方案把 FSM 提升到架构中心：所有流转必须走 FSM，禁止裸 `UPDATE status`。

3. **编排归 YAML，状态归代码**
   学自 Argo（YAML CRD）+ Cadence（枚举硬编码）的分层：状态枚举和合法转移矩阵用 Go const + map 硬编码（获得编译期检查），阶段顺序、guard 表达式、hook 注册用 YAML 配置（获得运行时可调）。**不做整体 YAML 化 FSM**——Temporal 的踩坑结论是"JSON/YAML 列步骤的 state machine 是在滥用配置格式表达代码"。

4. **终态之后没有状态，只有 hook**
   学自 Prefect/Airflow/Argo/Temporal 的**完全一致实践**：所有成熟引擎都不把 Archived 作为状态，归档是终态后的 hook 链或异步后台任务。本方案放弃 EasyMVP 原有的 `StageComplete` 独立阶段设计。

5. **多 judge 并行融合，硬规则只管硬**
   学自 ChatDev 的 Tester+Programmer 双 judge 并行 + EasyMVP 已验证的硬规则/LLM 加权融合。体验评审师作为第三路 judge 加入融合，不再是"第五号补丁角色"。

6. **域层零 IO**
   六边形架构硬约束：`domain` 包不允许引用 GoFrame、不允许 `g.DB()`、不允许 LLM 客户端。所有副作用走 `port` 接口。EasyMVP 的 `verification/service.go` 1929 行大杂烩就是没有这条约束的结果。

### 1.3 不做什么（明确的边界）

- 不做"零代码 AI"——目标用户是能读代码的产品经理/开发者
- 不做"完全替代开发者"——人工介入是一等公民（继承 EasyMVP 铁律 16）
- 不做"多租户 SaaS"第一版（先做单部署多用户，SaaS 化是后续话题）
- 不做"训练自己的模型"——调用外部 LLM，聚焦编排与判官
- 不做"可视化拖拽"第一版——YAML 是接入面，UI 是后续话题

---

<a id="2"></a>
## 2. 从 EasyMVP 继承什么、抛弃什么

这一节是本方案与 EasyMVP 的明确分界。**继承的是已被实战验证的产品思想；抛弃的是工程实现的债务。**

### 2.1 必须继承（产品资产）

| 资产 | 证据 | 为什么继承 |
|---|---|---|
| **五角色体系** | `rolecatalog/catalog.go:170-179` 定义的 architect/implementer/auditor/coordinator/experience_reviewer | 这是比 ChatDev（Designer/Coder/Reviewer/Tester）更细的分工，体验评审师是独有优势 |
| **三维预设**（category × role × level） | `presetutil/prompts.go` 的 BuildRoleSystemPrompt 实现 | 业内独有，让同一个角色在不同项目类型下使用不同提示词，避免 prompt 爆炸 |
| **六阶段工作流**（design→review→execute→accept→rework→complete） | `orchestrator/transition.go:59-65` stageOrder | 比 MetaGPT 的 5 角色线性流程更完整，比 ChatDev 的 8 Phase 更聚焦 |
| **硬规则+LLM融合判官**（40%/60% 加权） | `acceptance/decision_reducer.go:97` `result.Score = result.Score*0.4 + judgeResult.QualityScore*0.6` | 业内没有其他项目做融合，这是 EasyMVP 的核心差异化能力 |
| **七层自治** | `workflow/autonomy/` 目录 28 个文件的完整落地 | 业内没有对手，这是长期护城河 |
| **人工最高权限**（铁律 16） | `CLAUDE.md` 第十六节 | 产品哲学正确，与 OpenHands 的 AWAITING_USER_CONFIRMATION 殊途同归 |
| **数据权限五级 DataScope**（铁律 13） | `CLAUDE.md` 第十三节 | 企业级必须，继承 |
| **服务器负载保护**（铁律 17） | `CLAUDE.md` 第十七节 | 运维成熟度，继承 |
| **DDL 全走 migration**（铁律 14） | `CLAUDE.md` 第十四节 | 工程纪律，继承 |
| **多执行器抽象**（aider/openhands/claude_code/codex_cli/gemini_cli/chat） | `workflow/executor/` 包 | 业内最全，比 OpenHands 只支持自身 runtime 更开放 |

**小结**：继承 10 项——全部是**产品思想和工程纪律**，不继承任何一行具体实现。

### 2.2 必须抛弃（工程债务）

| 债务 | 证据 | 为什么抛弃 |
|---|---|---|
| **双套预设数据**（V2 英文 + Legacy 中文） | `mysql_seed.sql` 36 条空壳 + 另一套中文 Legacy | 配置只能一套 |
| **`engine` 包与 `workflow` 包的编排重叠** | `workflow_lifecycle.go` + `orchestrator/registry.go` + `verification/service.go` | 重建为单一编排中心 |
| **三套状态机手工互译** | `decision_center.go` 731 行 switch | 重建为显式级联 FSM |
| **零外键零 CHECK 的数据层** | `000001_baseline_schema.up.sql` 55 表 | 一次性补齐 |
| **锁内调 AI** | `context_compressor.go:192-226` | 架构级禁止：域层零 IO |
| **`CreateProject` 8 步无事务** | `workflow_lifecycle.go` CreateProject | 全部包事务 |
| **`StageComplete` 独立阶段** | `transition.go:64` | 依据调研：7 个成熟引擎都不把 Archived 做成状态 |
| **硬编码 workDir** | `chat_engine.go:85,157` `/www/wwwroot/project/easymvp` | 从配置读 |
| **执行器硬编码优先级** | `executor/auto.go` | 搬 YAML |
| **34 个 JSON 字段** | schema 审计 | 压到 10 个以内，其余拆列或拆子表 |
| **`verification/service.go` 1929 行大杂烩** | 审计 | 拆成 `domain/verification` + `application/accept` + `adapter/judge` |

### 2.3 必须升级（继承但重实现）

| 资产 | EasyMVP 的做法 | 本方案升级 |
|---|---|---|
| **五角色** | 4 常规 + 1 补丁（experience_reviewer 后加） | 5 一等公民，从第一天起权重相同 |
| **三维预设回退** | `role_resolver.go` 4 层 fallback 链 | 单层精确匹配 + 显式报错 + LRU 缓存 |
| **融合判官** | 2 路（硬规则 40% + LLM 60%） | 3 路（硬规则 30% + LLM 通用判官 30% + 体验评审师 40%） |
| **规则引擎** | `rule_engine.go` 硬编码 + DB 混合，未实现规则静默跳过 | CEL DSL 统一，未实现立即 fail |
| **状态机** | 3 套手工互译 | 3 套 FSM + 显式级联映射（一个 `fsm/cascade.go` 文件） |
| **完成阶段** | 独立 StageComplete | 合并入 accepting 终态 + on_completion hook 链（对齐 Prefect/Argo） |
| **事件流转** | `decision_center` 定时扫 DB | 本地事件总线（channel + WAL 表 `workflow_event`） |

### 2.4 不继承的（产品决策分歧）

| 放弃项 | 原因 |
|---|---|
| 中文 `project_category` 字段 | 展示名归 i18n 层，不进业务模型 |
| `mvp_task` 旧表族 | V2 已完成，Legacy 整体下线 |
| 七层自治中的 "L6 态势快照 + L7 观测" 的强耦合 | 观测应该是横切关注点（metrics + log + trace），不应是业务表 |

---

<a id="3"></a>
## 3. 顶层架构

### 3.1 系统分层总图

```
┌──────────────────────────────────────────────────────────────────┐
│                      1. Interface Layer                         │
│   REST API  │  SSE  │  Webhook  │  CLI  │  Feishu Bot           │
├──────────────────────────────────────────────────────────────────┤
│                    2. Application Layer                         │
│   Use Cases: CreateProject / ConfirmPlan / Dispatch / Accept    │
│   Orchestrator: WorkflowCoordinator（编排中心，唯一入口）         │
├──────────────────────────────────────────────────────────────────┤
│                     3. Domain Layer（零 IO）                     │
│   FSM (workflow/stage/task) │ Preset Resolver │ Rule Engine(CEL)│
│   Judge Fusion │ Role Catalog │ Autonomy Policy                 │
├──────────────────────────────────────────────────────────────────┤
│                     4. Port Layer（接口定义）                     │
│   ProjectRepo │ LLMClient │ Executor │ EventBus │ Lock │ FS     │
├──────────────────────────────────────────────────────────────────┤
│                    5. Adapter Layer                             │
│   MySQL(GoFrame) │ Claude/OpenAI/Gemini │ Aider/OpenHands/...   │
│   LocalFS/GitWorktree │ Channel+WAL EventBus │ Redis Lock       │
├──────────────────────────────────────────────────────────────────┤
│                  6. Infrastructure Layer                        │
│   Logger │ Metrics │ Tracing │ Config │ Migration │ Snowflake   │
└──────────────────────────────────────────────────────────────────┘
```

**硬约束**：
- 箭头只能从上向下调用
- 第 3 层不允许 `import` 第 5 层的任何包
- 第 3 层只允许 `import` 第 4 层的 interface
- 第 5 层不允许直接调用第 3 层的业务逻辑，只能实现第 4 层的 interface
- CI 阶段通过 `go-arch-lint` 或 `depguard` 检查违反

### 3.2 包组织

```
atlas/
├── cmd/
│   └── atlas-api/              # 主服务入口
├── internal/
│   ├── interface/              # 第 1 层
│   │   ├── rest/
│   │   ├── sse/
│   │   └── webhook/
│   ├── application/            # 第 2 层
│   │   ├── project/            # 项目用例（含 CreateProject，全事务）
│   │   ├── workflow/           # 工作流用例（Confirm/Pause/Resume/Retry）
│   │   ├── dispatch/           # 调度用例
│   │   ├── accept/             # 验收用例
│   │   └── coordinator.go      # WorkflowCoordinator 编排中心
│   ├── domain/                 # 第 3 层（零 IO）
│   │   ├── fsm/                # 三套 FSM + 级联映射
│   │   ├── role/               # 角色目录
│   │   ├── preset/             # 三维预设解析
│   │   ├── rule/               # CEL 规则引擎
│   │   ├── judge/              # 融合判官
│   │   ├── autonomy/           # 七层自治策略
│   │   └── errors.go           # 领域错误
│   ├── port/                   # 第 4 层
│   │   ├── repo.go             # ProjectRepo / WorkflowRepo / ...
│   │   ├── llm.go              # LLMClient
│   │   ├── executor.go         # Executor
│   │   ├── eventbus.go         # EventBus
│   │   ├── lock.go             # DistributedLock
│   │   └── fs.go               # Workspace
│   ├── adapter/                # 第 5 层
│   │   ├── mysql/              # GoFrame 实现
│   │   ├── llm/                # claude/openai/gemini/deepseek
│   │   ├── executor/           # aider/openhands/claude_code/codex/gemini_cli
│   │   ├── eventbus/           # channel+WAL 实现
│   │   ├── lock/               # redis/inproc
│   │   └── fs/                 # worktree
│   └── infra/                  # 第 6 层
│       ├── log/
│       ├── metric/
│       ├── trace/
│       ├── config/
│       └── snowflake/
├── config/
│   ├── fsm/                    # 三套 FSM 的 YAML 编排
│   │   ├── workflow.yaml
│   │   ├── stage.yaml
│   │   └── task.yaml
│   ├── rules/                  # CEL 规则文件
│   │   ├── precheck.yaml
│   │   └── accept.yaml
│   ├── executors.yaml          # 执行器优先级与重试
│   ├── roles.yaml              # 角色目录
│   └── presets/                # 三维预设（category × role × level）
│       └── *.yaml
└── manifest/
    └── sql/migrations/         # golang-migrate 迁移文件
```

**关键**：`domain` 和 `port` 加起来应当在 3000~5000 行之间；`application` 在 2000~4000 行；`adapter` 可以无限大。如果 `domain` 超过 8000 行，说明往里塞了不该塞的东西。

---

<a id="4"></a>
## 4. 核心概念模型

六个一等公民概念。每个都有明确的字段、不可变规则、生命周期。

### 4.1 Project（项目）

```
Project {
  id             Snowflake
  name           string
  category_code  string        # 'software_dev' / 'data_analysis' / ...
  family_code    string        # 'coding' / 'analysis' / 'creative'（来自 category）
  work_dir       string        # 运行时从 config 拼，不进 DB
  status         ProjectStatus # active / paused / archived / cancelled
  created_by     int64
  dept_id        int64
  created_at     timestamptz
}
```

**关键差异 vs EasyMVP**：
- 删掉中文 `project_category` 字段，展示走 i18n
- `work_dir` 不存 DB，运行时组合 `{base_dir}/{tenant}/{project_id}`

### 4.2 WorkflowRun（工作流运行）

```
WorkflowRun {
  id             Snowflake
  project_id     Snowflake
  status         WorkflowStatus  # 见 5.2 节，9 态
  current_stage  StageCode       # 指向当前活跃的 StageRun
  created_by     int64
  dept_id        int64
  started_at     timestamptz
  ended_at       timestamptz?
  close_reason   CloseReason?    # 对齐 Cadence 的 CloseStatus，见 5.2
}
```

**关键差异**：
- `close_reason` 是新增字段，对齐 Cadence 的 `WorkflowExecutionCloseStatus` 六值枚举
- 一个 Project 可以有多个历史 WorkflowRun（重试、重新立项）

### 4.3 StageRun（阶段运行）

```
StageRun {
  id             Snowflake
  workflow_run_id Snowflake
  stage_code     StageCode    # design / review / execute / accept / rework
  status         StageStatus  # pending / running / completed / failed / skipped
  batch_no       int          # 用于 execute 阶段的批次门控
  started_at     timestamptz
  ended_at       timestamptz?
}
```

**关键差异 vs EasyMVP**：
- **没有 `complete` stage**——completed 是 workflow_run 的终态，归档走 hook 链
- `rework` 不再是独立 stage_code，而是 `execute` stage 的一个 `rework_of` 指针字段（同阶段内的子分支）

### 4.4 DomainTask（领域任务）

```
DomainTask {
  id              Snowflake
  stage_run_id    Snowflake
  role_code       RoleCode        # architect/implementer/auditor/coordinator/experience_reviewer
  execution_kind  ExecutionKind   # ★ code / browser / mobile / game / data / api
  executor_code   ExecutorCode    # code_direct / browser_playwright / mobile_android / aider_fallback ...
  allowed_tools   []string        # 该任务允许的工具子集（在插件默认白名单基础上再收紧）
  status          TaskStatus      # 见 5.4 节
  batch_no        int
  depends_on      []Snowflake
  affected_files  []string        # 仅 code kind 有效；其他 kind 用 allow_domains / allow_packages 等
  allow_domains   []string        # 仅 browser kind 有效
  allow_packages  []string        # 仅 mobile kind 有效
  retry_count     int
  heartbeat_at    timestamptz?
  created_by      int64
  dept_id         int64
}
```

**关键差异**：
- **新增 `execution_kind`**：由 architect 角色在拆任务时决定。不同 kind 的任务必须由对应 kind 的执行器承接，Kernel 在分发时做编译期检查（见 §9.8）。
- **新增 `allowed_tools`**：任务级工具白名单，在插件默认白名单基础上再收紧（例如某次任务只允许 `read_file / write_file / run_tests` 三个工具，连 `delete_file` 都不给）。
- **新增 `allow_domains` / `allow_packages`**：非 code 类任务的资源边界——浏览器类任务声明允许访问的域名，移动类任务声明允许操作的 app 包名。这些字段和 `affected_files` 平行存在，按 `execution_kind` 分别生效。
- 不再区分 `mvp_task`（旧）和 `mvp_domain_task`（新）——只有一种任务
- `affected_files` 作为资源锁粒度直接进入字段，不再依赖 `mvp_task_resource_lock` 副表（读写成本太高）
- `heartbeat_at` 与 watchdog 统一

### 4.5 ProjectRole（项目角色配置）

```
ProjectRole {
  id             Snowflake
  project_id     Snowflake
  role_code      RoleCode
  model_id       Snowflake       # 必须非零，禁止回退到 0
  level          RoleLevel       # lite / pro / max
  executor_code  ExecutorCode
  system_prompt  text            # 从预设复制过来，之后可改
  is_from_preset bool            # 标记是否来自预设（用于后续同步）
  created_by     int64
  dept_id        int64
}
```

**关键差异 vs EasyMVP 的 `mvp_project_role`**：
- **写入时必须非零 model_id**——不存在"运行时回退到 preset"的路径
- **system_prompt 是写死的拷贝**——不再运行时拼接，保证可追溯
- `is_from_preset` 字段用于判断是否需要跟预设同步更新

### 4.6 QualityFinding（质量发现）

**新概念**：统一 `review_issue` + `accept_issue` + `accept_evidence` 三表。

```
QualityFinding {
  id             Snowflake
  workflow_run_id Snowflake
  stage_code     StageCode     # review 或 accept
  producer       FindingProducer # hard_rule / llm_judge / experience_reviewer / human
  severity       Severity      # blocker / error / warn / info
  rule_code      string        # CEL 规则 code 或 'llm_auto' 或 'human_manual'
  target_type    string        # task / file / blueprint / project
  target_id      string
  message        text
  evidence_json  json          # 原始证据（LLM 回复、测试输出、截图 URL）
  resolved       bool
  resolved_by    int64?
  resolved_at    timestamptz?
  created_by     int64
  dept_id        int64
  UNIQUE (workflow_run_id, stage_code, rule_code, target_type, target_id)
}
```

**关键差异**：
- 三表合一，避免 Rerun 时出现重复行（EasyMVP 当前痛点）
- `producer` 字段明确来源于三路 judge 的哪一路
- UNIQUE 约束从 DB 层保证幂等

---

<a id="5"></a>
## 5. 工作流引擎：状态机分层设计

这是本方案的**核心创新**——**三套 FSM 分层硬编码 + YAML 编排 + 级联映射**。

### 5.1 分层原则（再次强调）

| 层 | 形态 | 变更频率 | 谁能改 |
|---|---|---|---|
| 状态枚举 | Go const | 极低（业务语义变更才改） | 核心开发者，需走 code review |
| 合法转移矩阵 | Go `map[State][]State` | 低 | 核心开发者 |
| 转移 guard 表达式 | YAML + CEL | 中 | 业务配置师 |
| 阶段编排顺序 | YAML | 中 | 业务配置师 |
| 规则定义 | YAML + CEL | 高 | 业务配置师 |
| 预设内容 | YAML | 高 | 业务配置师 |

**不能混用**。状态枚举进 YAML 就是给自己挖坑（Temporal 博客原话：["state machine 的 JSON/YAML 化是在滥用配置格式表达代码"](https://temporal.io/blog/temporal-replaces-state-machines-for-distributed-applications)）。

### 5.2 WorkflowRun FSM（顶层工作流）

```go
// domain/fsm/workflow.go
type WorkflowStatus string

const (
    WorkflowDesigning WorkflowStatus = "designing"
    WorkflowReviewing WorkflowStatus = "reviewing"
    WorkflowExecuting WorkflowStatus = "executing"
    WorkflowAccepting WorkflowStatus = "accepting"
    WorkflowReworking WorkflowStatus = "reworking"
    WorkflowPaused    WorkflowStatus = "paused"
    WorkflowCompleted WorkflowStatus = "completed"   // 终态
    WorkflowFailed    WorkflowStatus = "failed"      // 终态
    WorkflowCancelled WorkflowStatus = "cancelled"   // 终态
)

// 对齐 Cadence 的 CloseStatus 0-5
type CloseReason int

const (
    CloseReasonCompleted       CloseReason = 0
    CloseReasonFailed          CloseReason = 1
    CloseReasonCancelled       CloseReason = 2
    CloseReasonTerminated      CloseReason = 3  // 人工终止
    CloseReasonTimedOut        CloseReason = 4
    CloseReasonContinuedAsNew  CloseReason = 5  // 重新立项
)

var workflowTransitions = map[WorkflowStatus][]WorkflowStatus{
    WorkflowDesigning: {WorkflowReviewing, WorkflowPaused, WorkflowCancelled},
    WorkflowReviewing: {WorkflowExecuting, WorkflowDesigning, WorkflowPaused, WorkflowCancelled, WorkflowFailed},
    WorkflowExecuting: {WorkflowAccepting, WorkflowReworking, WorkflowPaused, WorkflowFailed, WorkflowCancelled},
    WorkflowAccepting: {WorkflowCompleted, WorkflowReworking, WorkflowPaused, WorkflowFailed, WorkflowCancelled},
    WorkflowReworking: {WorkflowExecuting, WorkflowAccepting, WorkflowPaused, WorkflowCancelled, WorkflowFailed},
    WorkflowPaused:    {WorkflowDesigning, WorkflowReviewing, WorkflowExecuting, WorkflowAccepting, WorkflowReworking, WorkflowCancelled},
    WorkflowCompleted: {}, // 终态
    WorkflowFailed:    {WorkflowDesigning, WorkflowReworking, WorkflowCancelled}, // 允许从失败恢复
    WorkflowCancelled: {}, // 终态
}
```

**与 EasyMVP 的差异**：
- **删掉了 `completing` 中间态**——`accepting → completed` 一步到位，归档走 hook
- `failed` 允许回到 `designing`，这是 EasyMVP 已经支持的"失败后回到设计"
- 引入 `close_reason` 独立字段，与 `status=completed/failed/cancelled` 正交

### 5.3 StageRun FSM

```go
const (
    StagePending   StageStatus = "pending"
    StageRunning   StageStatus = "running"
    StageCompleted StageStatus = "completed"
    StageFailed    StageStatus = "failed"
    StageSkipped   StageStatus = "skipped"
)

var stageTransitions = map[StageStatus][]StageStatus{
    StagePending:   {StageRunning, StageSkipped},
    StageRunning:   {StageCompleted, StageFailed},
    StageCompleted: {}, StageFailed: {StagePending}, StageSkipped: {},
}
```

**完全继承 EasyMVP**，没有改动——这层已经是对的。

### 5.4 DomainTask FSM

```go
const (
    TaskPending        TaskStatus = "pending"
    TaskRunning        TaskStatus = "running"
    TaskCompleted      TaskStatus = "completed"
    TaskFailed         TaskStatus = "failed"
    TaskEscalated      TaskStatus = "escalated"      // 升级到人工
    TaskAuditing       TaskStatus = "auditing"       // 审计中
    TaskBugFound       TaskStatus = "bug_found"      // 审计发现 bug
    TaskBugDispatched  TaskStatus = "bug_dispatched" // bug 转派修复
)
```

继承 EasyMVP 的 8 态设计，但合法转移矩阵重写为显式 map。

### 5.5 级联映射：`fsm/cascade.go`

**关键**：EasyMVP 的痛点是三套状态机互相翻译的逻辑散在 5 个包里。本方案把所有级联规则集中到**一个文件**。

```go
// domain/fsm/cascade.go

// 当 DomainTask 进入某状态，对 StageRun 和 WorkflowRun 的影响
type TaskToStageEffect struct {
    TaskStatus  TaskStatus
    StageEffect func(stage *StageRun, task *DomainTask) Effect
}

// 三条明确的级联规则（不再散落）
var cascadeRules = []TaskToStageEffect{
    {
        TaskStatus: TaskBugFound,
        StageEffect: func(stage *StageRun, task *DomainTask) Effect {
            // 一个 task 发现 bug → stage 仍 running（其他 task 继续），但必须在 stage 结束时触发 rework
            return Effect{MarkStageHasBug: true}
        },
    },
    {
        TaskStatus: TaskFailed,
        StageEffect: func(stage *StageRun, task *DomainTask) Effect {
            if task.RetryCount >= task.MaxRetry {
                return Effect{EscalateToHuman: true}  // 升级
            }
            return Effect{Retry: true}
        },
    },
    // ...
}
```

**产出**：一个文件、几十条规则，可被直接阅读审计。代替 EasyMVP 当前 `decision_center.go` 的 731 行巨型 switch。

### 5.6 YAML 编排

```yaml
# config/fsm/workflow.yaml
stages:
  - code: design
    entry_guards: []
    exit_guards:
      - expr: "workflow.has_plan_version && workflow.plan_confirmed"
        on_fail: block
    timeout: 24h
    on_timeout: escalate_to_human

  - code: review
    entry_guards: []
    exit_guards:
      - expr: "accept.hard_rule_score >= 40 && accept.llm_score >= 50"
        on_fail: transition_to: rework
    timeout: 2h

  - code: execute
    entry_guards: []
    batching:
      enabled: true
      max_concurrent: 20
      batch_interval: 2s
    exit_guards:
      - expr: "stage.all_tasks_completed"
    timeout: 48h

  - code: accept
    entry_guards:
      - expr: "execute.all_tasks_completed || execute.completion_rate >= 0.8"
    exit_guards:
      - expr: "accept.decision == 'passed'"
        on_pass: transition_to: completed
      - expr: "accept.decision == 'failed'"
        on_pass: transition_to: rework

on_completion_hooks:
  - name: archive_artifacts
    ref: archive.local_fs
    async: true
    on_failure: log_only
  - name: generate_report
    ref: report.coordinator_llm
    async: true
    on_failure: log_only
  - name: cleanup_workspace
    ref: cleanup.worktree
    async: false
    on_failure: log_only
  - name: notify_feishu
    ref: notify.feishu
    async: true
    on_failure: log_only
```

**关键**：YAML 里**没有状态枚举**——状态是 Go const。YAML 只控制"每个阶段的进出条件、超时、批次、hook"。即使 YAML 写错，也不可能出现"未声明的状态"，编译期就挡掉了。

---

<a id="6"></a>
## 6. 角色体系与三维预设

### 6.1 五角色目录（一等公民）

```yaml
# config/roles.yaml
roles:
  - code: architect
    display_name: 架构师
    default_stages: [design]
    default_executor: chat
    preferred_levels: [max, pro]
    acceptance_judge: false
    sort: 10

  - code: implementer
    display_name: 实现者
    default_stages: [execute, rework]
    default_executor: auto    # 自动选 claude_code > aider > ...
    preferred_levels: [pro, max, lite]
    acceptance_judge: false
    sort: 20

  - code: auditor
    display_name: 审计员
    default_stages: [review]
    default_executor: chat
    preferred_levels: [pro, max]
    acceptance_judge: false
    sort: 30

  - code: coordinator
    display_name: 协调者
    default_stages: [review, execute, rework]
    default_executor: chat
    preferred_levels: [pro, max]
    acceptance_judge: false
    sort: 40

  - code: experience_reviewer
    display_name: 体验评审师
    default_stages: [accept]
    default_executor: chat
    preferred_levels: [max, pro]
    acceptance_judge: true       # 这是关键：参与 accept 阶段的融合判官
    acceptance_weight: 0.40      # 三路融合中的权重，见 7.3
    sort: 50
```

**重要差异 vs EasyMVP**：
- `experience_reviewer` 从 role catalog 的**第一行定义**就是一等公民，不再是"加上去的第五个"
- `acceptance_weight` 作为字段显式定义，而不是散落在代码里的 magic number
- 可以通过增加 YAML 条目加新角色（如 `security_reviewer`、`performance_reviewer`），无需改代码

### 6.2 三维预设：`category × role × level`

与 EasyMVP 一样的三维，但**存储方式改为 YAML + DB 分层**：

```yaml
# config/presets/software_dev/architect/max.yaml
category_code: software_dev
role_code: architect
level: max
default_model:
  provider: anthropic
  model: claude-opus-4-6
  temperature: 0.3
  max_tokens: 8192
default_executor: chat
system_prompt: |
  你是一位资深软件架构师，擅长...
  [完整 system prompt]
user_prompt_template: |
  项目名称：{{.project_name}}
  项目分类：{{.category_display}}
  需求描述：{{.requirement}}
  ...
```

**分层读取**：
- 系统启动时把所有 `config/presets/**/*.yaml` 加载进内存（LRU 缓存）
- `mvp_role_preset` 表只存"用户在 UI 里改过的覆盖版"
- 读取优先级：DB 覆盖 > YAML 默认 > 报错（**不再有第 4 层 fallback**）

**解决了什么**：
- EasyMVP 的 36 条空壳记录自然消失——空 YAML 在启动时就会报错
- 预设变更走 git，可回溯
- UI 层依然可以对单个项目做覆盖，写入 DB

### 6.3 角色解析（`domain/preset/resolver.go`）

```go
type Resolver interface {
    // 单层精确匹配，未命中直接报错，不做任何回退
    ResolveRole(ctx, projectID, roleCode, level) (*ResolvedRole, error)
}

type ResolvedRole struct {
    ProjectID     int64
    RoleCode      string
    Level         string
    ModelInfo     ModelInfo    // 非零 model_id，非空 api_key
    ExecutorCode  string       // 非 'auto' 的具体值
    SystemPrompt  string       // 非空
    Source        string       // 'db_override' or 'yaml_default'
}
```

**不变量**（编译期 + 运行时双检查）：
- `ModelInfo.ID != 0`
- `SystemPrompt != ""`
- `ExecutorCode != "" && ExecutorCode != "auto"`（`auto` 在 dispatch 层解析，不在这层）

违反任何一条直接 panic——这是为了让 EasyMVP 当前"36 条空壳"的问题**根本无法存在**。

---

<a id="7"></a>
## 7. 验收与归档：双层结构

这一节是本方案相对 EasyMVP 最显著的结构性改动，**完全基于调研证据**。

### 7.1 结论依据（来自调研报告）

> **所有 7 个成熟工作流引擎**（Temporal/Cadence/Argo/Prefect/Airflow/n8n/Conductor）**都不把 Archived 作为状态**。Completed 就是终态，归档要么是平台底层异步后台任务（Temporal/Cadence/Argo），要么是用户注册的 hook（Prefect/Airflow/Conductor/n8n）。
>
> **所有 5 个 AI agent 平台**中只有 ChatDev 有独立"完成阶段"——因为它要产出真实新 artifact（environment.yml、manual.md）。

EasyMVP 当前的 `StageComplete` 承担"产物打包 + 报告生成 + 资源回收 + 通知"四件事。**后两件在所有成熟系统里都是 hook**。前两件如果只是模板化的，也应该是 hook；只有"需要 LLM 生成新 artifact"才值得做成 stage。

**结论**：本方案**取消 `StageComplete` 独立阶段**，改为 `accepting → completed` 终态后的 hook 链。

### 7.2 Accepting 阶段：三路 judge 融合

```
                ┌─ hard_rules (CEL)         权重 30% ─┐
accepting ──────┼─ llm_judge (通用 LLM)     权重 30% ─┼─→ fused_score
                └─ experience_reviewer     权重 40% ─┘

decision rules:
  blocker_count > 0                           → failed
  fused_score >= 70 && blocker_count == 0     → passed
  50 <= fused_score < 70                      → manual_review (paused)
  fused_score < 50                            → failed (rework)
```

**关键设计**：

1. **硬规则一票否决**：`blocker_count > 0` 直接失败，不参与加权——这是 EasyMVP 已有的正确设计，继承
2. **三路并行**：三个 judge **并行运行**，不串行（对齐 ChatDev 的 Tester+Programmer 并行）
3. **权重可配**：在 `accept.yaml` 里配，不同 category 可以有不同权重
4. **体验评审师权重最高**（40%）：因为这是 Atlas/EasyMVP 区别于 ChatDev（它只有硬规则 tester）、MetaGPT（只有 QA role）、AutoGen（只有 LLM critic）的**独有能力**

### 7.3 融合公式（代码层面）

```go
// domain/judge/fusion.go
type JudgeResult struct {
    Producer FindingProducer
    Score    float64       // 0-100
    Blockers int
    Issues   []QualityFinding
}

type FusionConfig struct {
    Weights map[FindingProducer]float64  // 如 {hard_rule:0.3, llm_judge:0.3, experience_reviewer:0.4}
    PassThreshold    float64  // 70
    ManualThreshold  float64  // 50
}

func Fuse(results []JudgeResult, cfg FusionConfig) Decision {
    totalBlockers := 0
    weighted := 0.0
    allIssues := []QualityFinding{}
    for _, r := range results {
        totalBlockers += r.Blockers
        weighted += r.Score * cfg.Weights[r.Producer]
        allIssues = append(allIssues, r.Issues...)
    }
    switch {
    case totalBlockers > 0:
        return Decision{Result: Failed, Reason: "blocker present"}
    case weighted >= cfg.PassThreshold:
        return Decision{Result: Passed, Score: weighted}
    case weighted >= cfg.ManualThreshold:
        return Decision{Result: ManualReview, Score: weighted}
    default:
        return Decision{Result: Failed, Score: weighted}
    }
}
```

**相对 EasyMVP 当前 `decision_reducer.go:97` 的升级**：
- 从二路（40/60）扩展到 N 路（权重表）
- 融合逻辑纯函数，无 IO，可单元测试
- 新增 judge 只需加 `FindingProducer` 枚举 + 权重表条目，不改融合代码

### 7.4 终态 Hook 链：`on_completion`

**学自**：
- Prefect 的 [State Change Hooks](https://docs.prefect.io/v3/concepts/states)（on_completion / on_failure / on_crashed / on_cancellation）
- Argo 的 [Exit Handlers](https://argo-workflows.readthedocs.io/en/latest/walk-through/exit-handlers/)（onExit template）
- Airflow 的 [Callbacks](https://airflow.apache.org/docs/apache-airflow/stable/administration-and-deployment/logging-monitoring/callbacks.html)（on_success_callback / on_failure_callback）
- Temporal 的 [Archival](https://docs.temporal.io/temporal-service/archival)（终态后的异步归档）

```yaml
# config/fsm/workflow.yaml 的 on_completion_hooks 节
on_completion_hooks:
  - name: archive_artifacts
    ref: archive.local_fs
    async: true                  # 并行执行
    timeout: 5m
    on_failure: log_only         # 失败不影响 workflow 终态
    retries: 3
    config:
      dest: "{{base_dir}}/archive/{{workflow_run_id}}"

  - name: generate_summary_report
    ref: report.coordinator_llm
    async: true
    timeout: 3m
    on_failure: log_only
    config:
      template: coordinator_summary
      target: report.md

  - name: cleanup_workspace
    ref: cleanup.worktree
    async: false                 # 同步执行（串在最后）
    timeout: 30s
    on_failure: log_only
    config:
      keep_snapshot: true

  - name: notify_creator
    ref: notify.feishu
    async: true
    timeout: 10s
    on_failure: log_only
```

**规约**：
- **所有 hook 的失败都不能反向影响 workflow 终态**（对齐 Prefect 官方设计：Completed 永远是 Completed，hook 失败单独记录）
- **hook 有独立的重试与超时**
- **hook 的执行记录写入 `workflow_event` 表**（可观测性）
- **失败的 hook 触发告警**但不阻塞

### 7.5 Failed / Cancelled 终态的 hook 链

类似结构，但 hook 清单不同：

```yaml
on_failure_hooks:
  - name: preserve_workspace   # 保留现场供人工诊断
  - name: notify_creator_with_logs
  - name: escalate_to_human_if_bug_found

on_cancellation_hooks:
  - name: release_locks
  - name: cleanup_partial_work
  - name: notify_creator
```

### 7.6 与 EasyMVP `StageComplete` 的对照

| StageComplete 做的事 | 本方案归属 |
|---|---|
| 产物打包 | `archive_artifacts` hook |
| 报告生成 | `generate_summary_report` hook |
| 资源回收 | `cleanup_workspace` hook |
| 通知 | `notify_creator` hook |
| 状态转 completed | FSM 直接从 accepting 走 |

**净收益**：
- workflow FSM 少一个状态（9→8 实际活跃态）
- 消除"StageComplete 和 completed 终态名重名"的双重身份
- hook 失败不卡主流程
- hook 可以独立重试
- 新增 hook 只需加 YAML 条目

---

<a id="8"></a>
## 8. 七层自治与人工介入

### 8.1 继承 EasyMVP 的核心哲学

七层自治是 EasyMVP 的长期护城河，业内没有对手。本方案**完整继承**，但重实现。

**七层定义**（来自 `CLAUDE.md` 第五节表格）：

| 层 | 名称 | 职责 |
|---|---|---|
| L1 | 执行 | 具体任务执行器 |
| L2 | 调度 | 批次调度与资源分配 |
| L3 | 策略 | policy_rule / risk_gate_rule |
| L3.5 | 人工介入 | human_checkpoint / user_collab_binding |
| L4 | 决策 | autonomy_decision / decision_action |
| L5 | 学习 | action_outcome（策略效果跟踪） |
| L6 | 态势 | situation_snapshot |
| L7 | 观测与学习 | observation_record / assessment / tune_recommendation / learning_record |

### 8.2 本方案的升级

| 维度 | EasyMVP | 本方案 |
|---|---|---|
| L1-L4 | 11 张表，`decision_center.go` 731 行 switch | 6 张表，策略用 CEL 表达式 |
| L3 策略定义 | `policy_rule` 表 + Go 代码硬编码 | 全 CEL（见下） |
| L3.5 人工介入 | `human_checkpoint` 表 + 业务代码判断 | FSM 的 `paused` 状态 + CEL 触发规则 |
| L6-L7 观测 | `situation_snapshot` + `observation_record` 两表 | **下放到基础设施层**——用 Prometheus + Loki + Tempo 替代 |
| L5 学习 | `learning_record` + `action_outcome` | 保留，但简化为单表 `strategy_outcome` |

**关键差异**：L6（态势快照）和 L7（观测记录）在 EasyMVP 里是业务表，本方案把它们**下放到基础设施层**——态势快照就是 metrics 和 trace，观测记录就是 log。这样：
- 业务表减少 4 张
- 观测能力反而提升（Prometheus/Loki/Tempo 比自建表强）
- 保留 L5（`strategy_outcome` 策略效果跟踪）因为这是业务语义，不是通用 metrics

### 8.3 策略引擎：CEL 化

```yaml
# config/rules/autonomy/strategies.yaml
strategies:
  - code: adaptive_retry
    trigger:
      event: task_failed
      condition: "task.retry_count < 3 && task.executor_code != 'chat'"
    action:
      type: retry
      params:
        backoff: exponential
        max_wait: 5m

  - code: escalate_to_human
    trigger:
      event: task_failed
      condition: "task.retry_count >= 3 || task.severity == 'blocker'"
    action:
      type: create_human_checkpoint
      params:
        urgency: high
        notify: creator

  - code: batch_adjust
    trigger:
      event: stage_entered
      condition: "stage.code == 'execute' && system.cpu_load > 0.8"
    action:
      type: reduce_concurrency
      params:
        factor: 0.5
```

EasyMVP 的 28 个策略文件（`strategy_*.go`）全部变成 YAML 条目 + 若干通用 action handler。

### 8.4 人工介入（铁律 16 的工程化）

**铁律 16**："项目默认全自动执行，人工拥有最高权限，可在任何阶段介入"。

本方案的工程化实现：

1. **任何状态都可转 `paused`**（FSM 转移矩阵已经保证）
2. **`paused` 可恢复到任何非终态**（转移矩阵保证）
3. **`failed` 可回 `designing` / `reworking`**（允许从失败恢复）
4. **每个 FSM 状态都有 `allowed_human_actions` YAML 配置**：

```yaml
# config/fsm/workflow.yaml
stages:
  - code: execute
    allowed_human_actions:
      - pause
      - retry_task        # 重试单个失败任务
      - skip_task         # 跳过单个任务
      - force_transition: accepting   # 强制进入验收（跳过剩余 task）
      - force_transition: failed      # 强制失败
      - replan            # 触发重规划
```

5. **人工操作不走 FSM guard**——guard 是给自动流转用的，人工是 override，直接改状态（但仍要过转移矩阵白名单）

### 8.5 OpenHands 的启示

OpenHands 用 `AWAITING_USER_CONFIRMATION → USER_CONFIRMED/USER_REJECTED → FINISHED` 这条链做人工验收。本方案**不采用**这种做法——因为：
- EasyMVP 的 `paused` 已经能表达相同语义
- 再多三个状态会让 FSM 膨胀

但 OpenHands 的核心思想借鉴了：**人工介入不是外挂，是 FSM 的一等公民**。这个思想落在了 5.2 节 `WorkflowPaused` 的转移矩阵里。

---

<a id="9"></a>
## 9. 执行器家族：一个中央大脑 + N 个专精大脑

> 这是本方案相对 EasyMVP **最大的结构性升级**。EasyMVP 的执行器层继承自"包一下外部 agentic CLI"的思路——aider/claude_code/codex_cli/gemini_cli 都是给 Go 代码传一句 prompt，剩下全交给外部 CLI 自己决定改什么、跑什么、怎么跑。**越界的根源在这里**：控制权从第一秒就交出去了。
>
> 下一代执行器不是"Kernel + 插件"这种**被动工具**模型，而是**一个中央大脑 + N 个专精大脑**的**多 Agent 协作**模型。每个大脑都是带自己 LLM 的独立 Agent：中央大脑负责理解总任务、拆成并发子计划、调度专精大脑、按汇报推进；专精大脑各自精通一个领域（代码 / 浏览器 / 移动端 / 游戏 / 数据 / API），在被委派时自己跑 Agent Loop 完成子任务，返回结构化进度与多模态证据。
>
> **两个关键的不被妥协的特性**：
> - **都是内置 AI 的**——LLM 写进每个大脑的接口签名里，不是藏在某个实现细节中。中央大脑有中央大脑的 LLM，每个专精大脑有自己的 LLM，**都**是一等公民
> - **都兼容 OpenAI 和 Claude 系列**——通过自研的薄 LLM Provider 抽象层，任何一个大脑都可以在"Claude Opus / Claude Sonnet / Claude Haiku / GPT-5 / GPT-4.1 / o-series"之间切换，同一次任务里可以混搭（例如中央大脑用 Claude Opus 4.6 做规划，Code 专精大脑用 Claude Sonnet 4.6 写代码，Browser 专精大脑用 GPT-5 看截图）
>
> 本章按这个模型从上到下展开：§9.1 讲为什么换站位，§9.2 给全景图，§9.3~9.8 分别介绍 BrainKernel / CentralBrain / 各个 SpecialistBrain / LLM Provider 抽象 / BrainPlan 持久化协议 / 大脑委派工具，§9.9 是 YAML 配置，§9.10 讲如何作为独立产品发行。

### 9.1 为什么 EasyMVP 的执行器层要被替换（而不是小修）

先诚实对账。EasyMVP 的执行器已经非常用力（`workflow/executor/` 14 文件 + `engine/aider_runner.go` 727 行 + `utility/worktreeguard/guard.go` 514 行），而且已经有相当成熟的防御：`AllowPaths`、`Snapshot.Validate`、`cleanupAiderArtifacts`、`IsSuspiciousPath`、`ResolveRepoRoot`、`PruneGuardNoise`……但仍然拦不住 Aider 改 `.gitignore` / `go.mod` / 新建 CHANGELOG.md 这类越界。

**根因不是防御不厚，是架构站位错了**：

| 维度 | EasyMVP 当前 | 下一代 |
|---|---|---|
| 执行器本质 | 外部 agentic CLI 的 Go wrapper | **一个中央大脑 + N 个专精大脑的多 Agent 协作系统** |
| 控制权 | Prompt 进去，结果出来，中间不可见 | **每个大脑的每次 LLM 调用和每次工具调用都在 Go 代码里过一道** |
| 越界防御 | Post-run diff 审查（事后补救） | **工具白名单物理隔离**（事前不可能） |
| 能跑的环境 | 只能改代码仓库（文件系统） | **代码 / 浏览器 / 移动端 / 游戏 / 数据 / API** 六类，每类一个专精大脑 |
| 验收产出 | 文本 Output | **diff / 截图 / 视频 / 网络日志 / 操控序列** 多模态证据 |
| 扩展方式 | 加一个新 CLI = 写一个新 Go wrapper | 加一个新领域 = 写一个新**专精大脑**（system prompt + 工具集 + LLM 配置） |
| 任务拆分 | 外包给架构师角色 + 外部 CLI | **中央大脑是一个 Agent，自己拆并发子计划**，持久化到 `execution_plan` |
| LLM 角色 | 被外部 CLI 调用（不可见） | **所有大脑的接口签名都强制要求 LLM 配置**，一等公民 |
| LLM 供应商 | 每个 CLI 自己绑定（不可混搭） | **每个大脑单独配 provider/model**，Claude/OpenAI 自由混搭 |

**不替换就根治不了的问题**：
1. `.gitignore` / `go.mod` / `.git/hooks/*` 等越界（Aider 的自由度是设计出来的，不是 bug）
2. 没有**网页/App/游戏的联调验收能力**——EasyMVP 只能"验收代码层质量"，不能"验收最终产品是否真的跑起来了"
3. 没有**截图/录屏/交互回放**等多模态证据——实验评审师（experience_reviewer）拿到的永远是文本，看不到真实页面，判定质量受限
4. 没有**可信的"执行步骤链"**——当前产出是 "Prompt → 成功/失败"的黑盒，出了问题只能看 CLI 的日志去猜

**替换不是推倒 `workflow/executor/`**。EasyMVP 的现有资产里——`Executor/Registry` 接口（`executor.go`）、`resource_targets.go` 的路径解析、`worktreeguard` 的 514 行工具函数——**全部作为下一代 BrainKernel 与各专精大脑的底层工具库保留**。被换的只是"把 LLM 外包给 CLI"和"单层编排"这两层逻辑。

### 9.2 执行器家族全景

下一代执行器由 **1 个中央大脑 + N 个专精大脑 + 1 个 BrainKernel** 三层组成。BrainKernel 不是"又一个大脑"，它是**所有大脑共享的基础设施**——负责 LLM Provider 抽象、BrainPlan 持久化、并发执行、证据收集、审计落盘。

```
                    ┌─────────────────────────────────────────────────┐
                    │            CentralBrain  (中央大脑)              │
                    │  ─────────────────────────────────────────────  │
                    │  输入: 用户任务 + AcceptanceCriteria             │
                    │  职责:                                           │
                    │    1. 分析任务，判断是否可接（可拒绝并写出原因）  │
                    │    2. 拆分成并发子计划（subtask graph）          │
                    │    3. 通过"大脑委派工具"调用专精大脑             │
                    │    4. 收到专精大脑的汇报 → 更新 BrainPlan         │
                    │    5. 按新进度决定下一轮委派谁                   │
                    │    6. 整份计划完成 → 返回最终结果                │
                    │  LLM: 默认 claude-opus-4-6，可换 gpt-5           │
                    │  System Prompt: 来自三维预设库 (central/max)     │
                    │  工具集 (由 BrainKernel 注入):                   │
                    │    • code_brain(subtask_id, spec, context)       │
                    │    • browser_brain(subtask_id, spec, context)    │
                    │    • mobile_brain(subtask_id, spec, context)     │
                    │    • game_brain(subtask_id, spec, context)       │
                    │    • data_brain / api_brain (P3)                 │
                    │    • plan_store.create / update / read           │
                    │    • reject_task(reason_codes, description)      │
                    └────────────────┬────────────────────────────────┘
                                     │ (并发 tool_use，可一次发 3 个)
         ┌───────────────┬───────────┼───────────┬─────────────┐
         │               │           │           │             │
  ┌──────▼──────┐ ┌──────▼──────┐ ┌──▼─────────┐ ┌▼──────────┐ ...
  │ CodeBrain   │ │ BrowserBrain│ │ MobileBrain │ │ GameBrain │
  │ (P0)        │ │ (P0)        │ │ (P1 Android)│ │ (P2 Web)  │
  ├─────────────┤ ├─────────────┤ ├─────────────┤ ├───────────┤
  │ LLM:        │ │ LLM:        │ │ LLM:        │ │ LLM:      │
  │ claude-     │ │ claude-     │ │ claude-     │ │ claude-   │
  │ sonnet-4-6  │ │ sonnet-4-6  │ │ sonnet-4-6  │ │ sonnet-4-6│
  │ (可换 gpt)  │ │ +Vision     │ │ +Vision     │ │ +Vision   │
  │             │ │             │ │             │ │           │
  │ 工具:       │ │ 工具:       │ │ 工具:       │ │ 工具:     │
  │ read_file   │ │ navigate    │ │ start_app   │ │ 复用      │
  │ write_file  │ │ click       │ │ tap/swipe   │ │ Browser   │
  │ apply_patch │ │ fill        │ │ type_text   │ │ + 帧序列  │
  │ run_tests   │ │ screenshot  │ │ screenshot  │ │ +vision   │
  │ git_status  │ │ get_dom     │ │ dump_ui     │ │ _judge    │
  │ ...         │ │ wait_for    │ │ read_logcat │ │           │
  │ report      │ │ ...         │ │ ...         │ │           │
  │             │ │ report      │ │ report      │ │ report    │
  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └─────┬─────┘
         │               │               │              │
         └───────────────┴───────┬───────┴──────────────┘
                                 │
                    ┌────────────▼───────────────────────────┐
                    │       BrainKernel (共享基础设施)         │
                    │  ─────────────────────────────────────  │
                    │  • LLM Provider 抽象 (llm.Provider)     │
                    │    ├─ anthropic adapter (Claude系列)    │
                    │    ├─ openai adapter (GPT/o-series)     │
                    │    └─ 预留: deepseek/gemini/ollama...   │
                    │  • Agent Loop Runner                    │
                    │    (CentralBrain 和 SpecialistBrain     │
                    │     都跑这一份循环，差异在 system       │
                    │     prompt + 工具集)                    │
                    │  • BrainPlan Store                      │
                    │    ├─ DB mode: execution_plan+_subtask │
                    │    └─ CLI mode: ~/.easymvp-exec/runs/  │
                    │  • Artifact Store (证据)                │
                    │    ├─ DB mode: OSS/MinIO + 元数据表     │
                    │    └─ CLI mode: 本地文件系统           │
                    │  • Tool Registry + Guardrail            │
                    │  • Cost Meter (按 provider 分开记账)   │
                    │  • Trace/Audit Logger                   │
                    └─────────────────────────────────────────┘
```

**优先级含义**：
- **P0**（必做，第一版必上）：CodeBrain、BrowserBrain
- **P1**（按需，有 app 项目时上）：MobileBrain-Android
- **P2**（按需，有游戏项目时上）：GameBrain-Web（Canvas/WebGL，复用 BrowserBrain 底盘）
- **P3**（后续扩展）：MobileBrain-iOS（需 Mac 节点）、GameBrain-Native、DataBrain、APIBrain

一次性定义全家族，按优先级 PR 化落地。

**三个核心事实**：

1. **所有大脑共用同一份 Agent Loop 代码**——LLM 循环写在 BrainKernel 里**只写一次**。CentralBrain 和 SpecialistBrain 的差异只在三处：(a) system prompt 不同，(b) 注册的工具集不同，(c) LLM 配置不同。不会出现"中央大脑有一套循环、代码大脑有另一套循环"这种重复
2. **中央大脑是唯一决策者**——专精大脑只回答被问的问题，不自主反向调用其他专精大脑。这是 Claude Code 的 Task(subagent_type=xxx) 模型，避免形成调用环和权威漂移（§9.8 有明确约束）
3. **BrainPlan 是外部真相**——计划不在中央大脑的 LLM 上下文里"记忆"，而是存在 `execution_plan + execution_subtask` 两张表（Atlas 模式）或本地 JSON 文件（CLI 模式）。中央大脑每次被唤醒时，BrainKernel 把最新计划作为上下文的固定一段回注。这保证了：(a) LLM 换供应商不丢进度，(b) 任务中断可以从任意一轮恢复，(c) 人工可以直接 SQL 查到当前所有子任务的状态

### 9.3 Executor Kernel：AI 一等公民的通用内核

Kernel 的核心创新是**把"执行器要调 LLM"写进接口本身**，而不是让每个插件各自写一遍 LLM 循环。

```go
// workflow/executor/kernel/executor.go

// Executor 下一代执行器接口
type Executor interface {
    Kind() ExecutionKind                            // code/browser/mobile/game/data/api
    Capabilities() Capabilities
    Execute(ctx context.Context, req *Request) (*Result, error)
    HealthCheck(ctx context.Context) error
}

// Request 所有执行器的统一请求
type Request struct {
    TaskID         int64
    WorkflowRunID  int64
    RoleCode       RoleCode            // 谁在调（implementer/auditor/experience_reviewer...）
    Intent         string              // 高层意图描述（来自 architect 的任务卡）
    AcceptanceCriteria []string        // 验收标准（来自 plan_version）
    Workspace      *Workspace          // worktree / browser context / device handle
    Guardrail      *GuardrailConfig    // 白名单 / 黑名单 / 步数上限 / 时间上限
    AIConfig       *AIConfig           // ★ 一等公民：LLM 配置
    EvidenceBucket string              // 证据产物上传目录（截图/视频等大件）
}

// AIConfig 执行器自带 AI 大脑
type AIConfig struct {
    Provider     string             // anthropic / openai / local
    Model        string             // claude-sonnet-4-5 / gpt-4o / ...
    Vision       bool               // 是否启用视觉输入（截图直接喂）
    MaxSteps     int                // Agent Loop 最大步数
    Temperature  float32
    SystemPrompt string              // 来自 domain/preset/resolver 的角色系统提示
    ExtraTools   []ToolSchema        // 插件提供的工具集（由 Kernel 合并到 LLM tool-use）
}

// Result 多模态证据
type Result struct {
    Status       ExecutionStatus            // success / partial / failed / blocked_by_guard
    Summary      string                     // LLM 给出的终稿摘要
    StepsTaken   int
    ToolCalls    []ToolCallRecord           // 每一步的工具调用记录（审计用）
    Evidence     []EvidenceRef              // 多模态证据引用（截图 URL/diff/视频/日志）
    TokensUsed   int
    Cost         float64
    Duration     time.Duration
}

// EvidenceRef 多模态证据引用
type EvidenceRef struct {
    Kind     EvidenceKind  // diff / screenshot / video / log / trace / dom / frame_seq
    MimeType string
    URL      string        // 本地 FS 或对象存储
    Caption  string        // LLM 给的说明
    Step     int           // 在哪一步产生
}

// Kernel 通用内核
type Kernel struct {
    toolRegistry *ToolRegistry
    llmClient    llm.Client                 // 统一 LLM 客户端（Anthropic/OpenAI/...）
    guardrail    *Guardrail
    evidenceSink *EvidenceSink
    costMeter    *CostMeter
}

// Run Agent Loop 的主循环——所有插件共享这一份实现
func (k *Kernel) Run(ctx context.Context, req *Request, plugin Plugin) (*Result, error) {
    // 1. 初始化：让插件注册它自己的工具集
    tools := plugin.RegisterTools(k.toolRegistry)
    tools = append(tools, builtinTools()...)    // 通用工具：finish/recall/wait/note

    // 2. 构造 LLM 的 system prompt + tool schemas + 初始 message
    messages := k.buildInitialMessages(req, plugin)

    // 3. Agent Loop
    for step := 0; step < req.Guardrail.MaxSteps; step++ {
        // 超时检查
        if err := ctx.Err(); err != nil { return failed(err) }

        // 3a. 调 LLM，传 tools
        resp, err := k.llmClient.Chat(ctx, llm.ChatRequest{
            Model:    req.AIConfig.Model,
            Messages: messages,
            Tools:    tools,
            Vision:   req.AIConfig.Vision,
        })
        if err != nil { return failed(err) }
        k.costMeter.Add(resp.TokensUsed, resp.Cost)

        // 3b. LLM 要么返回 text（结束），要么返回 tool_call
        if resp.FinishReason == "stop" {
            return k.finalize(req, messages, resp), nil
        }

        // 3c. 对每个 tool_call 做白名单校验 + 执行
        for _, call := range resp.ToolCalls {
            if err := k.guardrail.Check(call); err != nil {
                messages = append(messages, toolError(call, err))
                continue  // LLM 下一步看到 error，会重试
            }
            toolResult, err := k.toolRegistry.Invoke(ctx, req, call)
            k.recordStep(step, call, toolResult, err)
            messages = append(messages, toolReply(call, toolResult, err))
        }
    }
    return failed(errMaxStepsExceeded)
}

// Plugin 感知+动作插件接口
type Plugin interface {
    Kind() ExecutionKind
    RegisterTools(registry *ToolRegistry) []ToolSchema
    PrepareWorkspace(ctx context.Context, req *Request) (*Workspace, error)
    TearDown(ctx context.Context, ws *Workspace) error
    BuildSystemPrompt(req *Request) string
}
```

**关键结构性决策**：

1. **LLM 循环在 Kernel 里写一次，所有插件复用**。插件只负责注册工具和定义"工作空间怎么初始化/清理"。
2. **工具白名单在 Go 代码里，不在配置里**。`toolRegistry.Invoke` 会检查每个 `tool_call` 是否在当前插件注册过、参数是否合法、受 `GuardrailConfig` 约束。**LLM 不可能调用没注册的工具**——它产出 JSON 说要调 `rm -rf /`，Kernel 压根没这个工具，直接报错返回给 LLM。
3. **每一步都记录**。`ToolCallRecord` + `EvidenceRef` 全量落盘，构成完整审计轨迹。
4. **Guardrail 是物理约束，不是事后审查**。和 EasyMVP 的 `Snapshot.Validate` 的区别是：`Snapshot.Validate` 跑完 Aider 后看 git 变更有没有可疑，下一代直接不让 `write_file(".gitignore")` 调用成立。
5. **视觉 LLM 是一等公民**。`AIConfig.Vision = true` 时，截图证据可以直接作为下一步 LLM 输入的 image_url，配合 Claude Sonnet 4.5 / GPT-4o 的视觉能力判页面/判游戏画面。

**Kernel 的工作量**：约 **1500 行 Go**（`kernel.go / tool_registry.go / guardrail.go / evidence_sink.go / cost_meter.go / llm_adapter.go`）。一个人 1 周能写到能跑通 POC 的程度，2 周可以到生产可用。

### 9.4 CodeExecutor（P0）—— 替代 Aider 的受控代码执行器

**定位**：EasyMVP 90% 的任务是"改几个文件 + 跑一下测试"，这些任务应当走 CodeExecutor，而不是 Aider。

**工具集**（在 Go 代码里硬编码，LLM 越不出去）：

| 工具 | 参数 | 白名单约束 | 副作用 |
|---|---|---|---|
| `read_file` | `path` | 必须在 `allow_read_paths` 内 | 无 |
| `list_dir` | `path` | 同上 | 无 |
| `write_file` | `path, content` | **必须在 `allow_write_paths` 内**（核心白名单） | 写 FS |
| `apply_diff` | `path, unified_diff` | 同 write_file | 写 FS |
| `delete_file` | `path` | 必须在 `allow_delete_paths` 内（通常为空） | 删 FS |
| `run_tests` | `package_pattern` | 必须在 `allow_test_paths` 内 | 跑 go test |
| `run_build` | — | `allow_build = true` 才可调 | 跑 go build |
| `git_status` | — | 无副作用 | 无 |
| `git_diff` | `path?` | 无副作用 | 无 |
| `recall_context` | `query` | 从 project context 找相关历史片段 | 无 |
| `finish` | `summary, status` | — | 结束循环 |

**黑名单（永远不允许，即使在白名单里）**：
- `.git/**`、`.github/workflows/**`、`.claude/**`
- `manifest/sql/mysql/**`（迁移文件只能由人工或 architect 角色通过 plan_version 走专用通道改）
- `go.mod` / `go.sum`（依赖变更必须走 architect 角色的"依赖变更"子流程）
- `.gitignore`（除非 architect 在 plan_version 里显式声明）

**AI 接入方式**：
- 默认 `model = claude-sonnet-4-5`（Anthropic SDK tool-use）
- `Vision = false`（代码任务不需要视觉）
- `MaxSteps = 30`（小任务通常 5~10 步就 finish）
- `SystemPrompt` 从 `domain/preset/resolver.go` 拿（来自三维预设库）

**验收证据**：`diff` + `run_tests` 输出 + `git_status` 快照

**和 Aider 的对比**：

| 维度 | Aider | CodeExecutor |
|---|---|---|
| 改 `.gitignore` | 频繁发生，需要 post-run 清理 | **物理不可能**（黑名单） |
| 改 `go.mod` | 可能发生 | **物理不可能**（黑名单） |
| 新建不相关文件 | 可能发生 | **物理不可能**（`write_file` 检查白名单） |
| 删除文件 | 可能发生 | **默认不允许**（`allow_delete_paths = []`） |
| 顺手重构不相关代码 | 经常发生 | 可以发生，但只能发生在白名单内的文件里 |
| 每一步动作可审计 | 只有一坨 stdout | **每次 tool_call 一条记录** |
| Token 预算控制 | 由 Aider 自己管 | 由 Kernel 的 `cost_meter` 管 |
| 失败回滚 | 靠 git snapshot + 手动恢复 | **事务化**：tool_call 失败时 LLM 看到 error 可重试；MaxSteps 超限时自动 `git reset --hard` 到初始 snapshot |

**工作量**：约 **500 行 Go**（`plugin_code.go` + `tools_code.go` + `allowlist_code.go`）。3~5 天。

**关键**：这不是"又一个 Aider"。这是"**只做 Aider 85% 的事，但越界物理不可能**"。剩下 15% 的场景（超大规模重构、需要跨 10 个文件的调试）仍然降级到 Aider/Claude Code，但默认路径不再走它们。

### 9.5 BrowserExecutor（P0）—— 网页联调验收的主力

**定位**：EasyMVP 的核心缺口。软件验收必须能真正打开浏览器点页面，不能只看代码层测试。

**底座**：`github.com/playwright-community/playwright-go`（Microsoft Playwright 的 Go 官方绑定，支持 Chromium/Firefox/WebKit）。

**工具集**：

| 工具 | 参数 | 说明 |
|---|---|---|
| `navigate` | `url` | 导航到页面（URL 必须在 `allow_domains` 白名单） |
| `click` | `selector` / `text` / `x,y` | 点击元素。selector 支持 CSS/XPath/role；text 支持"按文字点击"；xy 支持视觉点击 |
| `fill` | `selector, text` | 填表单 |
| `press_key` | `key` | 键盘事件 |
| `scroll` | `direction, amount` | 滚动 |
| `screenshot` | `full_page?` | 截图，自动上传到 `EvidenceBucket`，**回传的 image URL 直接作为下一步 LLM 的视觉输入** |
| `get_page_text` | `selector?` | 抓 DOM 文本（纯文本 LLM 用） |
| `get_page_tree` | — | 抓 accessibility tree（AX tree），比 DOM 更稳定 |
| `eval_js` | `script` | 执行 JS（黑名单：不能 eval `fetch` 到 `allow_domains` 外） |
| `wait_for` | `selector / url / network_idle` | 等待条件满足 |
| `net_check` | `url_pattern` | 检查浏览器发起过的网络请求，可断言 API 调用是否触发 |
| `record_video_start` / `record_video_stop` | — | 录屏（Playwright 原生支持） |
| `assert_visible` | `selector / text` | 断言元素可见 |
| `finish` | `summary, status` | 结束 |

**Guardrail**：
- `allow_domains`（白名单域名，默认 `localhost` + 项目配置的预览域名）
- `block_domains`（黑名单域名：第三方广告/追踪/统计服务）
- `allow_navigate_external = false`（禁止跳出白名单）
- 视觉工具必须走 `screenshot`，不允许直接下载任意 URL

**AI 接入方式**：
- **默认视觉 LLM**：`model = claude-sonnet-4-5, Vision = true`
- LLM 每次看到的输入是：当前 URL + 最近一次截图（image block）+ 操作历史
- `MaxSteps = 50`（网页任务步数通常更多）
- SystemPrompt 包含验收标准和"只能调本插件提供的工具"的约束

**验收证据**：截图序列（每步一张）+ 完整录屏 + 网络请求日志 + `ToolCallRecord` 序列 + LLM 给出的最终判定

**场景示例**——experience_reviewer 角色跑一个"新用户注册流程验收"：

```
Step 1: navigate("https://preview.example.com/register")
        → screenshot → LLM 看到注册表单
Step 2: LLM 决定 fill("#email", "test@test.com")
        → screenshot → LLM 看到 email 已填
Step 3: LLM 决定 fill("#password", "TestPass123!")
        → screenshot → LLM 看到密码已填
Step 4: LLM 决定 click("button[type=submit]")
        → wait_for(network_idle) → screenshot
        → LLM 看到跳转到 /dashboard，判定"注册流程通过"
Step 5: finish(summary="注册流程端到端验证通过，跳转正确", status=success)
        → 产出：5 张截图 + 完整录屏 + 网络日志 + 最终报告
```

这整个过程**Go 代码从头到尾知道每一步做了什么**，每个动作都在工具白名单内，页面截图直接成为 `QualityFinding.evidence_json` 的证据——这是 EasyMVP 当前**做不到**的事。

**工作量**：约 **1800 行 Go**（`plugin_browser.go` + `tools_browser.go` + `playwright_wrapper.go` + `screenshot_store.go` + `guardrail_browser.go`）。2 周。

**关键风险**：
1. Playwright 需要 headless Chrome 常驻进程，内存开销 ~200MB/context。需要进程池管理，不能每次 new 一个。
2. 视觉 LLM 的 token 消耗——一张 1920x1080 截图约 1000~2000 tokens，MaxSteps=50 意味着最多 100k tokens/次验收。用 Haiku 4.5 跑 screenshot 判断 + Sonnet 4.5 跑决策，可以降成本。
3. 选择器稳定性——推荐 LLM 优先用 `role=button[name=...]` 这类 accessibility selector，而不是 CSS。

**与 Web 游戏的关系**：Canvas/WebGL 游戏跑在浏览器里，**BrowserExecutor 天然支持**——只是 DOM 是空的，全靠视觉 LLM 看截图。工具集额外加 `get_canvas_frame` 即可。所以 GameExecutor-Web（P2）其实是 BrowserExecutor 的一个子模式，共用 Kernel + 大部分工具，只新增约 400 行。

### 9.6 MobileExecutor-Android（P1）—— 移动 app 联调验收

**定位**：当 EasyMVP 接 Android app 项目时必要。

**底座**：`adb` + `uiautomator2`（Google 原生 UI 测试框架）。Go 通过 `os/exec` 调 adb，或直接用 `github.com/electricbubble/gadb`（Go 实现的 adb 客户端）。

**工具集**：

| 工具 | 参数 | 说明 |
|---|---|---|
| `start_app` | `package_name, activity?` | 启动 app |
| `stop_app` | `package_name` | 杀 app |
| `tap` | `x, y` / `resource_id` / `text` | 点击 |
| `long_press` | `x, y, duration_ms` | 长按 |
| `swipe` | `x1, y1, x2, y2, duration_ms` | 滑动 |
| `type_text` | `text` | 输入文字 |
| `press_back` / `press_home` / `press_menu` | — | 系统按键 |
| `screenshot` | — | 截图（存到 EvidenceBucket，喂视觉 LLM） |
| `dump_ui_tree` | — | uiautomator dump xml（AX tree） |
| `read_logcat` | `filter?, since?` | 读 logcat |
| `wait_for` | `text / resource_id / duration_ms` | 等待 |
| `assert_visible` | `text / resource_id` | 断言 |
| `start_video_record` / `stop_video_record` | — | 录屏（`screenrecord`） |
| `install_apk` | `apk_path` | 安装 APK（白名单检查） |
| `finish` | `summary, status` | 结束 |

**Guardrail**：
- 只允许操作 `allow_package_names` 白名单内的 app
- `install_apk` 只允许从 `allow_apk_paths` 路径
- 不允许 `am` 命令注入（除了启动白名单 app）
- 不允许 `su` / root 操作
- 设备只允许操作 `allow_device_serials` 白名单（防止一个任务影响另一个并发任务用的模拟器）

**AI 接入方式**：同 BrowserExecutor，视觉 LLM 为主，dump_ui_tree 作为补充。

**验收证据**：截图序列 + 录屏 + logcat 过滤后的日志 + `ToolCallRecord`

**环境依赖（硬约束）**：
- Android SDK（`sdkmanager` + `emulator` + `platform-tools`）
- 至少一个 Android 模拟器或真机，通过 `adb devices` 可见
- 建议：`ARM64 System Image` + `KVM 加速`，单实例约 2~3 GB 内存 / 2 CPU

**工作量**：约 **2500 行 Go** + 一套环境初始化脚本。3 周。

**iOS 版本（P3）**：同样架构，底座换 `wda (WebDriverAgent)` 或 Appium XCUITest。**必须 macOS 节点 + Xcode**，这是 Apple 的硬约束。不纳入 pss_claude 路线图，作为后续扩展。

### 9.7 GameExecutor（P2/P3）—— 游戏验收

**分两类，难度天差地别**。

#### P2 · Web 游戏（Canvas/WebGL）

**复用 BrowserExecutor**，新增 400 行 Go：
- 工具集补 `get_canvas_frame`（抓 canvas 的 imageData）
- 工具集补 `send_game_input(key/mouse)`（用 `playwright.keyboard.press`/`mouse.click` 精准坐标）
- 新增 `vision_judge` 工具：把最近 N 帧截图作为 image block 喂给视觉 LLM，让它判断"是否到达关卡 X"、"是否出现死亡画面"

**工作量**：1 周，**但前提是 BrowserExecutor 已经就位**。

#### P3 · 原生游戏（Unity / UE / 自研引擎）

**这是真正的难区**，不纳入 pss_claude 路线图。技术路径备忘：

- **感知**：OS 级截屏（Linux: `xwininfo`+`xwd`/`scrot`；Windows: DXGI Duplication API；macOS: `CGDisplayCreateImage`）+ 视觉 LLM
- **动作**：系统级 input 注入（Linux: `xdotool`；Windows: `SendInput`；macOS: `CGEvent`）
- **高级方案**：游戏引擎 SDK 集成（Unity ML-Agents、UE Remote Control API）——需要**游戏开发者在引擎侧配合**，不是执行器单方面能做的
- **判定**：帧序列 → 视觉 LLM → 自然语言判定

**工作量**：1~3 个月，**视觉 LLM 对游戏画面的识别稳定性是最大不确定性**。不建议第一版做。

### 9.8 执行器选择策略与 AI 的关系

执行器不是随意调的。`DomainTask` 新增字段 `execution_kind`（code/browser/mobile/game/data/api），由 architect 角色在拆任务时决定。不同 kind 的任务**必须**由对应 kind 的执行器承接，Kernel 在分发时做编译期检查。

```go
// domain_task 新字段
type DomainTask struct {
    // ... 原有字段
    ExecutionKind ExecutionKind    // code / browser / mobile / game / data / api
    ExecutorHint  string           // 可选，如 "code:direct_apply" / "browser:playwright"
    AllowedTools  []string         // 该任务允许的工具子集（在插件默认白名单基础上再收紧）
}
```

**调度策略**：
1. `ExecutorHint` 指定 → 直接用
2. 未指定 → 按 `ExecutionKind` 找默认执行器（每 kind 一个默认）
3. 默认执行器 HealthCheck 失败 → 降级到兜底（code → aider/claude_code；browser → 无兜底，报错）
4. 降级路径用最严格的 Guardrail 配置（黑名单强化）

**LLM 的身份多重性**：

| 位置 | LLM 做什么 | 谁调用 |
|---|---|---|
| Architect 角色 | 拆任务、决定 `execution_kind` | 现有 ChatStream |
| Auditor 角色 | 审代码/审验收报告 | 现有 ChatStream |
| Experience Reviewer 角色 | 看截图/看录屏做产品评审 | 现有 ChatStream，但 input 包含 BrowserExecutor 产出的多模态证据 |
| CodeExecutor 的 Agent Loop | **自己决定改哪些文件、跑哪个测试** | **Kernel** |
| BrowserExecutor 的 Agent Loop | **自己决定点哪里、等什么** | **Kernel** |
| MobileExecutor 的 Agent Loop | **自己决定在 app 里怎么操作** | **Kernel** |
| Fusion Judge | 融合三路判定 | 纯函数，不调 LLM |

也就是说——**下一代有两种 LLM 调用点**：
1. **角色层 LLM**（设计/审核/评审）——继承自 EasyMVP 的 ChatStream
2. **执行器层 LLM**（Agent Loop）——**新增**，Kernel 内部

这正是用户说的"执行器也要能接入 AI"。这两层的 LLM 调用在 `workflow_transition_log` 里都有记录，都受 `cost_meter` 统一记账。

### 9.9 YAML 配置（仅元数据，不含状态）

```yaml
# config/executors.yaml
kernel:
  default_llm:
    provider: anthropic
    model: claude-sonnet-4-5
    timeout_per_step: 60s
  evidence:
    storage: local_fs       # 或 s3 / oss
    bucket: /data/easymvp/evidence
    retention_days: 90
  cost_meter:
    daily_budget_usd: 100   # 超预算自动熔断

executors:
  - kind: code
    code: code_direct
    plugin: code.DirectApplyPlugin
    default: true
    priority: 100
    llm_override:           # 可覆盖 kernel.default_llm
      model: claude-sonnet-4-5
    guardrail:
      blacklist:
        - .git/**
        - .github/workflows/**
        - .claude/**
        - manifest/sql/mysql/**
        - go.mod
        - go.sum
        - .gitignore
      max_steps: 30
      max_duration: 10m

  - kind: code
    code: aider_fallback
    plugin: code.AiderWrapperPlugin   # 包装老的 aider，作为兜底
    priority: 10                       # 低优先级
    default: false

  - kind: browser
    code: browser_playwright
    plugin: browser.PlaywrightPlugin
    default: true
    priority: 100
    llm_override:
      model: claude-sonnet-4-5
      vision: true                     # 开视觉
    guardrail:
      allow_domains: [localhost, "*.preview.internal"]
      block_domains: ["*.google-analytics.com", "*.facebook.com"]
      max_steps: 50
      max_duration: 15m

  - kind: mobile
    code: mobile_android
    plugin: mobile.AndroidPlugin
    priority: 100
    guardrail:
      allow_packages: []    # 每个任务运行时注入
      allow_devices: ["emulator-5554"]
      max_steps: 50

selection_policy:
  type: kind_match
  health_check:
    enabled: true
    failure_window: 10m
    cooldown: 5m
```

**关键原则**：YAML 里**只有元数据和配置值**，没有任何行为描述。行为在 Go 代码里。加执行器 = 写 Go 插件 + 在 YAML 注册；改行为 = 改 Go 插件；**YAML 不够用绝对不是往 YAML 里塞更多行为**——这条是铁律。

### 9.10 独立产品形态：执行器家族可以脱离 Atlas 单独发行

> **用户原话**："第一，执行器可以作为独立产品使用，就像 Claude Code 一样。"
>
> 这条需求把执行器家族从"Atlas 的子系统"重新定位为"**一等产品**"——Atlas 只是它的一个消费者，就像 Claude Code CLI 可以被任何人在任何终端里直接跑、也可以被 IDE 插件嵌入调用一样。本节规定这件事怎么落到代码结构和依赖边界上。

#### 9.10.1 产品形态：三种发行方式并存

执行器家族从第一天就提供三种发行通道，三者共享同一份 Kernel + Plugin 源码：

| 形态 | 发行物 | 面向用户 | 典型用法 |
|---|---|---|---|
| **CLI 二进制** | `easymvp-exec`（单个静态链接 Go 二进制） | 开发者 / 运维 / CI 管道 | `easymvp-exec run --kind=browser --task "登录 localhost:3000 截图"` |
| **Go SDK** | `github.com/easymvp/executor`（独立 go module） | Go 应用 / Atlas / 第三方 | `import "github.com/easymvp/executor/kernel"` + `kernel.Run(ctx, req, plugin)` |
| **Docker 镜像** | `ghcr.io/easymvp/executor:vX.Y.Z`（预装 playwright/adb/chromium） | Kubernetes / 无 Go 环境 | `docker run --rm easymvp/executor run --kind=browser ...` |

**共同约束**：三种形态使用相同的 `Request/Result/AIConfig` schema，相同的工具白名单语义，相同的 `execution_artifact` 目录结构。这意味着：
- 一条任务在本地 CLI 跑通后，原样丢进 Atlas 也能跑通
- Atlas 产线上遇到的 bug 可以用本地 CLI 一比一复现
- 第三方团队可以只用 SDK 接入自己的编排系统而不承担 Atlas 的 55 表负担

#### 9.10.2 Atlas 与 Executor 的边界：消费者关系，不是父子关系

```
┌────────────────────────────────────────────┐
│ Atlas / EasyMVP 编排层                       │
│  (workflow/orchestrator, workflow/stage)    │
│   ↓ 只调 kernel.Run(ctx, req, plugin)        │
│  ────────────────────────────────────────   │
│  github.com/easymvp/executor  (独立 module) │
│    ├─ kernel/                               │
│    │   (LLM loop + 工具白名单 + 证据)        │
│    ├─ plugin/code/                          │
│    ├─ plugin/browser/                       │
│    ├─ plugin/mobile_android/                │
│    ├─ plugin/game_web/                      │
│    └─ cmd/easymvp-exec/  (CLI 入口)          │
└────────────────────────────────────────────┘
         ↑
  其他消费者（第三方 / IDE 插件 / 竞品编排器）
  用同一个 Go SDK 或 CLI 或 Docker 镜像接入
```

**边界铁律**：
1. **执行器 module 不允许 import Atlas 代码**
   - `workflow/executor/kernel/` 和 `workflow/executor/plugin/*` 必须零依赖于 `workflow/orchestrator`、`workflow/domain/fsm`、`workflow/acceptance`、`engine/*`、任何 `app/mvp/internal/model/entity` 下的 GoFrame 生成代码
   - CI 增加一条静态检查：`go list -deps ./workflow/executor/...` 的结果中不能出现 `easymvp/admin-go/app/mvp/internal/...`
2. **Atlas 作为消费者只 import kernel 包的公开接口**
   - `workflow/orchestrator/dispatch.go` 只能 `import "easymvp/workflow/executor/kernel"`，不能反向访问 plugin 细节
   - 如果需要自定义工具，由 Atlas 实现 `kernel.Tool` 接口注入，不是改 plugin 源码
3. **持久化双轨**
   - Kernel 对外暴露 `ArtifactStore` 接口（`Put(ctx, key, reader) (url string, err error)`）
   - Atlas 内部实现为"OSS/MinIO + `execution_artifact` 表"；独立 CLI 实现为"本地文件系统 + `~/.easymvp-exec/artifacts/<run>/metadata.json`"；不同发行形态只是换一个 `ArtifactStore` 实现
   - `ExecutorRegistry` 同理：Atlas 走 `executor_registry` 表，CLI 走 `executors.yaml` + 内存
   - **没有 mvp_ 前缀的表出现在 kernel/plugin 代码里**
4. **配置双轨**
   - Atlas 从 DB + `manifest/config/executors.yaml` 加载
   - CLI 从 `~/.easymvp-exec/config.yaml` 或命令行 flag 加载
   - 两者走同一个 `config.Load(sources ...Source) (*Config, error)` 入口

#### 9.10.3 代码物理位置与仓库切分策略

**短期（pss_claude 路线图内，PR-13 ~ PR-17）**：
- 源码仍然放在单体仓 `admin-go/app/mvp/internal/workflow/executor/`，但**内部作为独立子 module**——新建 `go.mod`：`module github.com/easymvp/executor`
- 单体仓通过 `replace github.com/easymvp/executor => ./app/mvp/internal/workflow/executor` 使用本地版本
- CI 增加静态依赖检查防止 Atlas 侧代码反向污染

**中期（Executor v1.0 发布时）**：
- 把 `workflow/executor/` 整个子树**物理剥离**到独立仓库 `github.com/easymvp/executor`
- Atlas 改为通过 `go get github.com/easymvp/executor@v1.0.0` 依赖
- 独立仓库拥有自己的 git tag、自己的 CHANGELOG、自己的 release cadence
- 迁移方式：`git subtree split --prefix=admin-go/app/mvp/internal/workflow/executor -b executor-split` → push 到新仓库

**长期（执行器生态建立后）**：
- 独立仓库接受外部贡献，可能拆分为 `easymvp-executor-kernel` / `easymvp-executor-browser` 等多仓库
- Atlas 只是其中一个 reference consumer，类似 Claude Code 之于 Anthropic API

#### 9.10.4 版本与发布

| 维度 | Atlas | Executor |
|---|---|---|
| 版本号 | `atlas-vX.Y.Z`（与数据库迁移编号对齐） | `executor-vX.Y.Z`（独立语义化版本） |
| 发布节奏 | 跟随迁移波次（约 2 周一次） | 跟随插件稳定性（可能更慢） |
| 向后兼容 | Atlas 内部，无公开 API | **严格的 SDK 语义化版本**——`kernel.Run` / `Plugin` / `Tool` 接口的 breaking change 走 major bump |
| 发布通道 | 内部 Docker 镜像 | GitHub Releases + Go module proxy + GHCR 镜像 |

**关键承诺**：Executor 的 `kernel.Run` / `Plugin` / `Tool` / `Request` / `Result` / `AIConfig` 这六个符号的 v1.0 签名一旦发布就**冻结一年**——Atlas 自己对 Executor 的调用站点必须能平滑升级，第三方消费者才敢依赖。

#### 9.10.5 CLI UX 样例（对齐 Claude Code 的直觉）

```bash
# 1. 初始化：生成默认配置
easymvp-exec init
# → 创建 ~/.easymvp-exec/config.yaml，引导填 ANTHROPIC_API_KEY / OPENAI_API_KEY

# 2. 一次性代码任务（像 aider / claude code 一样）
easymvp-exec run \
  --kind=code \
  --work-dir=./myproject \
  --allow-paths="src/**,tests/**" \
  --task="给 /api/ping 接口加限流，5 次/秒，失败返 429"
# → 直接跑 Agent Loop，实时 stream 工具调用到 stderr，最后打印 diff 到 stdout

# 3. 浏览器联调验收（像 playwright codegen 但 AI 驱动）
easymvp-exec run \
  --kind=browser \
  --allow-domains="localhost,preview.internal" \
  --task="访问 localhost:3000 登录页，用 admin/admin 登录，截图确认进入 dashboard"
# → 自动启动 chromium headless，截图保存到 ~/.easymvp-exec/artifacts/<run-id>/

# 4. 接自定义 LLM（像 claude -p）
easymvp-exec run --kind=code --model=claude-sonnet-4-5 --task="..."
easymvp-exec run --kind=code --provider=openai --model=gpt-5 --task="..."

# 5. JSON 输出模式（对接 CI / 被上游调用）
easymvp-exec run --kind=browser --task="..." --output-format=json
# → 输出 {"status":"success","artifacts":[...],"tool_calls":[...],"duration_ms":12340}
```

**对齐 Claude Code 的设计决策**：
- `--output-format=json` 让任何编排器都能把 Executor 当成黑盒子命令行工具调用
- `~/.easymvp-exec/` 目录承载配置 + 产物 + 历史，可以直接 `tar` 带走
- `--task` 参数接受自然语言，Executor 内部的 Agent Loop 负责拆解为工具调用
- 无状态 CLI：每次 `run` 都是独立的 Request，没有隐藏的 session 文件——这让 CLI 在 CI 和 Atlas 里的行为一致

#### 9.10.6 对 §13.5 PR 路线图的影响

PR-13（Kernel）在原有要求上补两条：
- (a) `workflow/executor/` 根目录新建 `go.mod`，成为独立子 module
- (b) 新建 `workflow/executor/cmd/easymvp-exec/main.go` 作为 CLI 入口骨架，支持 `init` / `run` / `version` 三个命令

PR-14 / PR-15 / PR-16 / PR-17 在原有基础上补一条：
- 每个 plugin 的包测试必须包含"只用 SDK + 本地文件系统 ArtifactStore 跑通"的 case，证明脱离 Atlas 也能工作

---

<a id="10"></a>
## 10. 数据层设计

### 10.1 目标

- 表数从 EasyMVP 的 55 张减到 **46 张**
- 外键：**0 → 20+ 关键外键**
- CHECK 约束：**0 → 所有状态列带白名单**
- JSON 字段：**34 → 10 以内**
- 唯一约束：杀重复（quality_finding 首当其冲）
- 所有业务表必须含 `created_by` + `dept_id` NOT NULL DEFAULT 0（铁律 13）

### 10.2 表清单（按模块）

**系统 & AI 模块（16 表，基本继承 EasyMVP）**
- system_users / system_role / system_menu / system_dept / system_user_role / system_user_dept / system_role_menu / system_role_dept / system_role_ai_engine
- ai_provider / ai_plan / ai_model / ai_engine / ai_engine_config / ai_task / ai_task_log

**项目核心模块（6 表）**
- `project` —— 项目（删中文字段，加 family_code）
- `project_category` —— 分类定义（category_code + display_name_i18n_key + family_code）
- `project_role` —— 角色配置（强制非零 model_id）
- `project_role_preset_override` —— **新**：UI 改过的预设覆盖版
- `conversation` —— 对话
- `message` —— 消息（流式用 message_chunk 副表，本版合并到同表 + chunks JSON）

**工作流流水线模块（8 表）**
- `workflow_run` —— 工作流运行
- `workflow_event` —— 事件（含 hook 执行记录）
- `workflow_transition_log` —— **新**：所有 FSM 转移的审计表
- `stage_run` —— 阶段运行
- `stage_task` —— 阶段子任务
- `domain_task` —— 领域任务（合并 `mvp_task_dependency` + `mvp_task_resource_lock` 到 JSON 字段，降索引）
- `plan_version` —— 计划版本
- `task_blueprint` —— 任务蓝图

**质量门模块（3 表）**
- `quality_finding` —— **新**：统一 review_issue + accept_issue + accept_evidence
- `accept_run` —— 验收运行实例
- `accept_rule` —— 验收规则（CEL 源文本）

**执行器家族模块（2 表，新）**
- `execution_artifact` —— **新**：执行器产出的多模态证据（screenshot / video / dom_snapshot / trace / tool_call_log / diff_ref），统一指向 OSS/MinIO，库里只存元数据。被 `quality_finding` 和 accept 流程引用
- `executor_registry` —— **新**：执行器注册表（executor_code / execution_kind / plugin_version / capabilities JSON / default_tool_whitelist / health_status），取代 ai_engine_config 里硬编码的 5 条记录，Kernel 分发时按此表解析

**自治模块（6 表，从 11 表精简）**
- `autonomy_strategy` —— 策略定义（CEL）
- `autonomy_decision` —— 决策记录
- `decision_action` —— 决策动作
- `human_checkpoint` —— 人工介入节点
- `user_collab_binding` —— 用户协作绑定
- `strategy_outcome` —— 策略效果跟踪（L5）

**（L6 situation_snapshot、L7 observation/assessment/tune/learning 全部去掉，走 Prometheus/Loki/Tempo）**

**配置模块（2 表）**
- `config` —— 系统配置（38 项灰度开关）
- `config_history` —— **新**：配置变更审计

**工作空间（1 表）**
- `task_workspace`

**系统队列（1 表）**
- `sys_delete_queue`（补齐 created_by/dept_id）

**合计**：16 + 6 + 8 + 3 + 2 + 6 + 2 + 1 + 1 = **45 张**，比原计划的 46 更激进（新增 executor 家族 2 张后仍未反弹）。

### 10.3 关键约束

```sql
-- 外键（选 20 条关键）
ALTER TABLE workflow_run
  ADD CONSTRAINT fk_workflow_run_project
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE RESTRICT;

ALTER TABLE stage_run
  ADD CONSTRAINT fk_stage_run_workflow
    FOREIGN KEY (workflow_run_id) REFERENCES workflow_run(id) ON DELETE RESTRICT;

ALTER TABLE domain_task
  ADD CONSTRAINT fk_domain_task_stage
    FOREIGN KEY (stage_run_id) REFERENCES stage_run(id) ON DELETE RESTRICT,
  ADD CONSTRAINT fk_domain_task_workflow
    FOREIGN KEY (workflow_run_id) REFERENCES workflow_run(id) ON DELETE RESTRICT;

ALTER TABLE quality_finding
  ADD CONSTRAINT fk_finding_workflow
    FOREIGN KEY (workflow_run_id) REFERENCES workflow_run(id) ON DELETE RESTRICT;

ALTER TABLE execution_artifact
  ADD CONSTRAINT fk_artifact_domain_task
    FOREIGN KEY (domain_task_id) REFERENCES domain_task(id) ON DELETE CASCADE,
  ADD CONSTRAINT fk_artifact_workflow
    FOREIGN KEY (workflow_run_id) REFERENCES workflow_run(id) ON DELETE RESTRICT;

ALTER TABLE execution_artifact
  ADD CONSTRAINT chk_artifact_kind CHECK (artifact_kind IN (
    'screenshot','video','dom_snapshot','frame_seq',
    'trace','tool_call_log','diff_ref','log','other'
  ));

ALTER TABLE executor_registry
  ADD CONSTRAINT chk_execution_kind CHECK (execution_kind IN (
    'code','browser','mobile','game','data','api'
  ));

ALTER TABLE executor_registry
  ADD CONSTRAINT uk_executor_code UNIQUE (executor_code);

ALTER TABLE project_role
  ADD CONSTRAINT fk_project_role_project
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE RESTRICT,
  ADD CONSTRAINT fk_project_role_model
    FOREIGN KEY (model_id) REFERENCES ai_model(id) ON DELETE RESTRICT;

-- CHECK 约束（所有状态列白名单）
ALTER TABLE workflow_run
  ADD CONSTRAINT chk_workflow_status CHECK (status IN (
    'designing','reviewing','executing','accepting','reworking',
    'paused','completed','failed','cancelled'
  ));

ALTER TABLE workflow_run
  ADD CONSTRAINT chk_close_reason CHECK (
    close_reason IS NULL OR close_reason BETWEEN 0 AND 5
  );

-- 唯一约束（杀重复）
ALTER TABLE quality_finding
  ADD CONSTRAINT uk_finding_scope UNIQUE (
    workflow_run_id, stage_code, rule_code, target_type, target_id
  );

ALTER TABLE project_role
  ADD CONSTRAINT uk_project_role UNIQUE (project_id, role_code);

-- NOT NULL 数据权限字段
ALTER TABLE sys_delete_queue
  ADD COLUMN created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN dept_id    BIGINT UNSIGNED NOT NULL DEFAULT 0;
```

### 10.4 JSON 字段瘦身（34 → 10）

| 原 EasyMVP | 本方案 |
|---|---|
| `*.metadata_json`（11 处） | 结构化为 3~5 个显式列 |
| `*.tags_json` | 拆子表 `*_tag`（多对多） |
| `mvp_config.value_json`（配置值） | 保留（配置本质 schemaless） |
| `plan_version.config_snapshot_json` | 保留（版本化快照） |
| `quality_finding.evidence_json` | 保留（外部证据多样） |
| `domain_task.depends_on_json` | 保留（小数组，不拆表） |
| `domain_task.affected_files_json` | 保留（小数组，不拆表） |
| `workflow_event.payload_json` | 保留（事件 payload 本质异构） |
| `llm_call_log.raw_response_json` | 保留（外部原始回复） |
| `strategy_outcome.metrics_json` | 保留（动态指标） |

**合计**：保留 10 个 JSON 字段，其余 24 个全部拆列或拆表。

### 10.5 索引策略

- 每个表默认索引：`(dept_id, created_by)`（配合数据权限五级 DataScope）
- 时间查询索引：`(created_at)` 或 `(status, created_at)`
- 外键列自动索引（MySQL 默认行为）
- **禁止超过 6 个索引的表**（EasyMVP 的 `mvp_task` 9 索引是反面教材）
- 加索引走 migration，每次 migration review 必须审核新增索引

---

<a id="11"></a>
## 11. 六边形架构分层

### 11.1 四条铁律

1. **domain 包不能 import 任何 port 以外的东西**
   - 禁止 `github.com/gogf/gf`
   - 禁止 `g.DB()`、`g.Cfg()`
   - 禁止 `internal/adapter/...`
   - 禁止 `internal/infra/...`（但 `log/slog` 等标准库 ok）

2. **application 包只能通过 port 调用 adapter**
   - application 可以 import domain 和 port
   - 不允许 application 直接 import adapter

3. **adapter 包不能 import application 或 domain 的业务逻辑**
   - adapter 只能 import port 的 interface
   - 以及 domain 里的纯数据结构（`domain/model.go`）

4. **CI 阶段强制检查**
   - 使用 [go-arch-lint](https://github.com/fe3dback/go-arch-lint) 或 [depguard](https://github.com/OpenPeeDeeP/depguard)
   - 违反直接红屏，不许合并

### 11.2 包依赖图

```
interface ──→ application ──→ domain ──→ port
                │                           ↑
                ↓                           │
               port ──→ adapter ────────────┘
                │
                ↓
              infra
```

### 11.3 为什么非要这么严

EasyMVP 的 `verification/service.go` 1929 行就是没有这些规约的结果：它同时做了 domain 逻辑（决策融合）、application 编排（调度 judges）、adapter 职责（直连 DB 和 LLM）。结果是：
- 决策逻辑没法单元测试（依赖真实 DB）
- 单测变集成测试
- 改 DB 表会改业务逻辑
- 换 LLM 提供商要改业务逻辑

**四条铁律的目标**：让域逻辑可以在 5ms 内跑完单元测试。

### 11.4 示例：`domain/judge/fusion.go` 的可测试性

```go
// fusion.go 里没有任何 IO，纯函数
func Fuse(results []JudgeResult, cfg FusionConfig) Decision { ... }

// fusion_test.go
func TestFuse_BlockerAlwaysFails(t *testing.T) {
    results := []JudgeResult{
        {Producer: ExperienceReviewer, Score: 100, Blockers: 1},
    }
    cfg := defaultCfg()
    got := Fuse(results, cfg)
    assert.Equal(t, Failed, got.Result)   // 即使体验师给 100 分，有 blocker 就失败
}

func TestFuse_ExperienceReviewerWeightDominates(t *testing.T) {
    results := []JudgeResult{
        {Producer: HardRule, Score: 100, Blockers: 0},
        {Producer: LLMJudge, Score: 100, Blockers: 0},
        {Producer: ExperienceReviewer, Score: 30, Blockers: 0},
    }
    // 30 * 0.3 + 100 * 0.3 + 100 * 0.4 = 9 + 30 + 40 = 79 → pass
    // 但如果权重是 30 * 0.4 + 100 * 0.3 + 100 * 0.3 = 12 + 30 + 30 = 72 → 也 pass
    // 这里测试"体验师低分但总分还是 pass"的边界
    got := Fuse(results, cfg)
    assert.True(t, got.Score >= 70)
}
```

这样的测试在 EasyMVP 当前根本写不出来——因为融合逻辑和 DB 访问耦合在 `decision_reducer.go` 的同一个函数里。

---

<a id="12"></a>
## 12. 可观测性与治理

### 12.1 四大观测支柱

| 支柱 | 工具 | 覆盖内容 |
|---|---|---|
| **日志** | structured slog → Loki | 所有 FSM transition、hook 执行、执行器调用 |
| **指标** | Prometheus | stage 时延、judge 分数、执行器成功率、成本 |
| **追踪** | OpenTelemetry → Tempo | CreateProject → Dispatch → Accept 全链路 |
| **审计** | `workflow_transition_log` 表 | 所有状态变更（永不物理删） |

### 12.2 关键指标（至少暴露）

```
# stage 时延
atlas_stage_duration_seconds{stage="design|review|execute|accept"}

# judge 分数分布
atlas_judge_score{producer="hard_rule|llm_judge|experience_reviewer"}

# 融合决策分布
atlas_accept_decision_total{result="passed|failed|manual_review"}

# 执行器调用
atlas_executor_run_seconds{executor="claude_code|aider|..."}
atlas_executor_success_total{executor="..."}
atlas_executor_failure_total{executor="...",reason="timeout|error|..."}

# 成本
atlas_llm_tokens_total{model="claude-opus-4-6",role="architect",kind="input|output"}
atlas_llm_cost_usd_total{model="...",role="..."}

# 自治策略
atlas_strategy_triggered_total{strategy="adaptive_retry|escalate|..."}
atlas_strategy_outcome_total{strategy="...",outcome="success|failure"}

# 人工介入
atlas_human_checkpoint_total{urgency="low|medium|high"}
atlas_human_response_seconds{urgency="..."}
```

### 12.3 日志规范

每条结构化日志至少包含：
- `workflow_run_id`
- `stage_code`
- `task_id`（如适用）
- `trace_id`（关联 Tempo）
- `level`
- `event`（枚举：fsm_transition / executor_call / judge_eval / hook_executed / ...）

### 12.4 `workflow_transition_log` 表

```sql
CREATE TABLE workflow_transition_log (
  id             BIGINT UNSIGNED PRIMARY KEY,
  workflow_run_id BIGINT UNSIGNED NOT NULL,
  from_status    VARCHAR(32) NOT NULL,
  to_status      VARCHAR(32) NOT NULL,
  trigger        VARCHAR(64) NOT NULL,  -- 'auto_guard' / 'human_override' / 'timeout' / ...
  trigger_actor  VARCHAR(64),            -- user_id 或 'system'
  guard_result   JSON,                   -- CEL 表达式求值结果
  duration_ms    BIGINT,
  created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX (workflow_run_id, created_at),
  INDEX (trigger, created_at)
);
```

**规约**：
- 所有 FSM transition 必须先写 log 再提交 DB 事务
- 这张表永不物理删（软删也不做）
- 可用于事后审计和回放

### 12.5 治理红线

1. **禁止裸 `UPDATE status = ?`**
   - 所有 status 修改必须走 `domain/fsm.Transition(from, to, trigger)`
   - CI 阶段用正则扫描源码，发现裸 UPDATE 直接拒合并

2. **禁止持锁调外部服务**
   - `context_compressor` 不能在持有项目锁期间发 LLM 请求
   - CI 阶段可以通过静态分析（`lockspector` 类工具）扫描

3. **禁止在 domain 包内引用 time.Now()**
   - 全部走 `port.Clock` 接口，方便测试注入

4. **每次 PR 必须带迁移文件**
   - 如果 PR 改了表结构但没 migration → 拒合并

---

<a id="13"></a>
## 13. 从 EasyMVP 迁移的三种路径

这一节给出三条真实可选的路径，**用户可以明确选择一条**。

### 13.1 路径 A：Greenfield 全新部署（推荐）

**适合**：在另一台机器或另一个数据库里全新搭 Atlas，老 EasyMVP 继续跑历史项目直到自然下线。

**步骤**：
1. 搭新 Atlas 环境，走完整 migration
2. 新项目在 Atlas 创建
3. 老 EasyMVP 保持只读 + 允许继续运行已有项目
4. 6~12 周后老环境退役

**优点**：零迁移风险，新老隔离
**缺点**：两套环境并行运行一段时间

### 13.2 路径 B：原地迁移 + 双写

**适合**：只有一套生产环境，无法新起部署。

**步骤**：
1. 部署 Atlas 到同一数据库（新表 prefix `atlas_`）
2. 写入层做双写（EasyMVP 写旧表 + Atlas 写新表）
3. 读取层先继续走 EasyMVP
4. 数据对账通过后，读取层切到 Atlas
5. 停 EasyMVP 写入
6. 清理旧表

**优点**：无需新环境
**缺点**：双写期间复杂，需要对账

### 13.3 路径 C：单次大切换

**适合**：能接受 2~4 小时停机窗口，且项目不在运行中途。

**步骤**：
1. 维护公告
2. 停 EasyMVP
3. 跑一次性数据迁移脚本（EasyMVP 表 → Atlas 表）
4. 启动 Atlas
5. 冒烟测试
6. 对外服务

**优点**：简单直接
**缺点**：需要停机窗口

### 13.4 选型建议

| 场景 | 推荐路径 |
|---|---|
| 生产数据少，允许重来 | A（Greenfield） |
| 生产数据多，要求平滑 | B（双写） |
| 数据量中等，能接受停机 | C（切换） |

**我的推荐**：**路径 A**。理由：
- EasyMVP 当前处于"收尾期"，历史项目数据不多
- 新老隔离最安全
- 老环境当历史查询使用，成本低
- 团队可以在新环境自由迭代，不怕破坏老数据

### 13.5 路径 B+：pss_claude 原地落地的 PR 级执行清单

> **为什么单列一节**：前面三条路径是"做不做 Greenfield"的战略选择。但用户已经明确要在 `pss_claude` 分支上**原地改造**，不开新仓库。这不是路径 B 的"双写迁移"——更像**结构重整**：保留所有可复用的 EasyMVP 工程资产（V2 已经拆好的 `workflow/stage/` 五阶段包、`workflow/acceptance/` 的融合雏形、`manifest/sql/mysql/` 的 migrate 框架、`watchdog/scheduler/executor` 包、铁律 13/14/16/17），只改**结构性反模式**（三套状态机打架、55 表零约束、engine 包 41 文件缠在一起、编排散在 6 处）。
>
> 本节是第 1~12 章所有设计决策的**具体落地路线**。每个 PR 都能独立编译、独立合并、独立回滚。

#### 13.5.1 当前仓库可复用的工程资产（动手前先认清楚）

这些是 EasyMVP 已经做对、**不能推倒重写**的工程资产。本路线图的每个 PR 都要显式尊重它们：

| 资产 | 位置 | 下一代怎么用 |
|---|---|---|
| 五阶段子包拆分 | `workflow/stage/{accept,execute,review,rework,complete}` | 保留 accept/execute/review/rework 四个，`complete/` 改造为 hook 执行器 |
| 融合判官雏形 | `workflow/acceptance/{decision_reducer,judge,rule_engine,evidence_collector}.go` | 抽成纯函数搬到 `domain/judge/`，不重写 |
| 角色目录 | `workflow/rolecatalog/` | 直接承载 §6.1 五角色 YAML 加载器 |
| 三维预设工具 | `workflow/presetutil/` | 直接承载 §6.3 `ResolveRole` 实现 |
| 领域 plan/task | `workflow/domain/{plan,task}/` | 承载 §4 核心概念的 Go 类型定义 |
| migrate 框架 | `manifest/sql/mysql/000001~000011` + `make db-*` | 继续用，新 PR 从 `000012` 开始编号 |
| Watchdog 双保险 | `workflow/watchdog/` + `engine/watchdog.go` | 迁移过程零动，`heartbeat_at` 字段直接复用 |
| 调度器批次门控 | `workflow/scheduler/` + `engine/scheduler_*.go` | 算法不动，只换数据源（新 `domain_task` 表） |
| 执行器多选 | `workflow/executor/` + 5 条 `ai_engine_config` | 只把 `engine/executor_dispatch.go` 的 447 行 switch 搬进来 |
| Git Worktree 隔离 | `mvp_task_workspace` 表 + 运行时拼路径 | 零动 |
| 铁律 13 数据权限 | 所有业务表 `created_by + dept_id` + 五级 DataScope | 新表从第一天就带齐 |
| 铁律 14 migrate | 所有 DDL 进 `.up/.down.sql`，禁止手动改库 | 本路线图每个 PR 严格遵守 |
| 铁律 16 人工最高权限 | 所有 stage 都支持 pause/resume/retry/skip/manual_pass/reject | FSM YAML 的 `manual_*` 转移必须保留 |
| 铁律 17 负载保护 | CPU>80 停手 / <50 恢复 | 本路线图每个 PR 的 CI 闸门都挂一次 `uptime` 检查 |

**关键原则**：如果一个改动能用"改 5 行"解决，就不写新包；如果新包可以放进已有子包（如 `workflow/acceptance/`），就不新建同级包。

#### 13.5.2 PR 依赖图（17 个 PR，6 个波次）

```
 波次 1（数据层基座，互相独立，可并行）
   PR-01  quality_finding 建表 + 双写开关
   PR-02  workflow_transition_log 建表 + FSM 钩子
   PR-03  role_preset 种子数据修复（135 行）

 波次 2（领域层骨架，依赖波次 1）
   PR-04  domain/fsm 三份 YAML + Go 枚举 + cascade.go     ← 依赖 PR-02
   PR-05  domain/preset resolver（LRU 缓存 + 单层查询）    ← 依赖 PR-03
   PR-06  domain/judge fusion 纯函数化                    ← 依赖 PR-01

 波次 3（编排层接入，依赖波次 2）
   PR-07  application/lifecycle CreateProject 事务化      ← 依赖 PR-04
   PR-08  stage/* 五阶段接 FSM、complete 改 hook 链        ← 依赖 PR-04/PR-06
   PR-09  engine 包瘦身：executor_dispatch 搬家 + 锁外压缩  ← 依赖 PR-07

 波次 4（治理与清理，依赖全部）
   PR-10  CEL 规则引擎替换硬编码 precheck                 ← 依赖 PR-04
   PR-11  autonomy/decision_center 事件驱动化            ← 依赖 PR-02
   PR-12  Legacy 表与 engine 遗留文件物理删除             ← 依赖 PR-07/08/09 全部上线两周

 波次 5（执行器家族一步到位：P0 核心三件套，依赖波次 3）
   PR-13  Executor Kernel（Agent Loop + 工具白名单 + 多模态证据）  ← 依赖 PR-07
   PR-14  CodeExecutor / DirectApply（自研代码执行器 + Aider 兜底）← 依赖 PR-13
   PR-15  BrowserExecutor（Playwright-go + Vision LLM 联调验收）  ← 依赖 PR-13

 波次 6（执行器家族扩展：P1/P2，依赖波次 5）
   PR-16  MobileExecutor-Android（adb/uiautomator2 + Vision LLM） ← 依赖 PR-13 / P1
   PR-17  GameExecutor-Web（复用 BrowserExecutor + 帧序列截取）    ← 依赖 PR-15 / P2
```

**合并节奏**：波次 1 一次性合并，波次 2/3 逐 PR 合，波次 4 等观察期，波次 5 与波次 4 可并行开始（PR-13/14/15 是新增包，不动存量路径）。全部合完后 `engine/` 包剩余文件 ≤ 10 个，`workflow/verification/service.go` 1929 行降到 3~4 个 300 行以内的文件；执行器家族覆盖 **code/browser/mobile/game** 四大 kind，iOS 与原生游戏（UE/Unity）明确推后，不在本路线图内。

#### 13.5.3 PR 清单逐个展开

##### PR-01 · `000012_quality_finding.up.sql`

| 字段 | 说明 |
|---|---|
| 动的迁移文件 | `manifest/sql/mysql/000012_quality_finding.up.sql` + `.down.sql` |
| 动的 Go 代码 | `workflow/acceptance/` 新增 `finding_writer.go`（双写旧表 + 新表） |
| 新增依赖 | 无 |
| 编译闸门 | `go build ./app/mvp/...` |
| 验证方式 | 建一个新项目走到 reviewing → 检查 `quality_finding` 与 `mvp_review_issue` 内容一致 |
| 回滚 | `make db-down STEPS=1` + revert Go 文件 |
| 可独立 merge | 是（双写，读仍走旧表） |

`quality_finding` 建表 DDL 即 §4.6 + §10.3 的字段与约束合集，含 UNIQUE `(workflow_run_id, stage_code, rule_code, target_type, target_id)` 和 `created_by/dept_id` 非空。

##### PR-02 · `000013_workflow_transition_log.up.sql`

| 字段 | 说明 |
|---|---|
| 动的迁移文件 | `000013_workflow_transition_log.up.sql` + `.down.sql` |
| 动的 Go 代码 | `workflow/eventstream/` 新增 `transition_logger.go`；`orchestrator/transition.go` 每次合法转移后 append 一行 |
| 编译闸门 | `go build ./app/mvp/...` |
| 验证 | 走一个完整 workflow，`SELECT * FROM workflow_transition_log WHERE workflow_run_id=?` 看到 design→review→execute→accept→completed 五条记录 |
| 回滚 | `make db-down STEPS=1` + revert，`transition.go` 的钩子用 feature flag 包住可热关 |

表结构即 §12.4 定义：`(id, workflow_run_id, scope, from_state, to_state, guard_result, duration_ms, actor, reason, created_at, created_by, dept_id)`。

##### PR-03 · `000014_role_preset_rebuild.up.sql`

| 字段 | 说明 |
|---|---|
| 动的迁移文件 | `000014_role_preset_rebuild.up.sql`（DELETE 旧 36 条空预设 + INSERT 135 条新预设）+ `.down.sql`（恢复 36 条空预设兜底） |
| 动的 Go 代码 | `workflow/presetutil/` 补齐 5×9×3 的 system_prompt 文案来源；`workflow/repo/role_preset_repo.go` 单层查询不 fallback |
| 编译闸门 | `go build ./app/mvp/...` + 新增单元测试 `presetutil_test.go` 覆盖 45 个组合 |
| 验证 | 新建 9 个不同 category 的项目，5 角色全部非 `model_id=0` 且 `system_prompt != ''` |
| 回滚 | `make db-down STEPS=1` |
| 风险 | 135 条 seed 内容需要人工编写，这是体力活（用户拍板：每 category 的 architect/max 可以互相参考） |

##### PR-04 · `domain/fsm` 三份 YAML + cascade.go

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/domain/fsm/` 子包：`workflow_fsm.go`、`stage_fsm.go`、`task_fsm.go`、`cascade.go`、`loader.go` |
| 新建配置文件 | `manifest/config/fsm/{workflow,stage,task}.yaml` |
| 改的文件 | `workflow/orchestrator/transition.go`（删除硬编码转移矩阵，改为调 `fsm.Next()`）；`workflow/autonomy/decision_center.go` 的 731 行 switch 暂时保留但引用 `cascade.go` 作为 single source of truth |
| 编译闸门 | `go build ./app/mvp/...` + FSM 单元测试覆盖所有合法转移 + 所有非法转移 |
| 验证 | 跑一个完整 workflow，`workflow_transition_log` 里所有 `guard_result` 非空且 `from_state/to_state` 来自 Go const |
| 回滚 | revert，Go const 与 YAML 同一个 PR 里，不存在半截状态 |
| **关键约束** | **Go 枚举 = 真相，YAML 只控制 guard/timeout/hook**——本 PR 必须在 code review 时重点检查有没有偷偷把状态名写进 YAML |

##### PR-05 · `domain/preset` resolver 带 LRU 缓存

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/domain/preset/resolver.go`；删除 `engine/role_resolver.go`（70 行）；`workflow/presetutil/` 的 `BuildRoleSystemPrompt` 改为 resolver 的 adapter |
| 依赖 | `github.com/hashicorp/golang-lru/v2`（已在 go.mod 里确认） |
| 编译闸门 | `go build ./app/mvp/...` + 单元测试覆盖 LRU 命中/未命中/预设空/显式报错 4 个路径 |
| 验证 | bench：`ResolveRole` P99 < 1ms（LRU 命中时） |
| 回滚 | revert |
| 合并节点 | 必须等 PR-03 合完（否则 LRU 缓存到的是空预设） |

##### PR-06 · `domain/judge/fusion.go` 纯函数化

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/domain/judge/fusion.go`（§7.3 的 `Fuse` 函数），从 `workflow/acceptance/decision_reducer.go:97` 的 40/60 硬编码抽出；`decision_reducer.go` 改为调用 `fusion.Fuse(...)` 的薄 adapter |
| 不动的 | `workflow/acceptance/judge.go`（225 行 LLM 调用）、`evidence_collector.go`（412 行证据收集）——这两块是有 IO 的 adapter 层，留在原位 |
| 编译闸门 | `go build ./app/mvp/...` + `fusion_test.go` 覆盖 5 种权重组合 + blocker 否决路径 |
| 验证 | 跑完整 accept 阶段，`quality_finding.producer` 字段包含 hard_rule/llm_judge/experience_reviewer 三路，融合 `Decision` 与预期一致 |
| 回滚 | revert，`decision_reducer.go` 还原 |

##### PR-07 · `application/lifecycle` CreateProject 事务化

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/lifecycle/` 子包承载 `CreateProject` 的 8 步流程；从 `engine/workflow_lifecycle.go`（357 行）剥离；全流程包在 `g.DB().Transaction(ctx, func(ctx, tx) error { ... })` 里 |
| 删的文件 | `engine/workflow_lifecycle.go` |
| 编译闸门 | `go build ./app/mvp/...` |
| 验证 | 在事务中第 5 步 `panic` 测试，确认前 4 步回滚干净（DB 里没有半死项目） |
| 回滚 | revert |
| 合并节点 | 必须等 PR-04 合完（FSM 初始态由 `fsm.Initial(WorkflowFSM)` 决定，不能再硬编码） |

##### PR-08 · `stage/*` 接 FSM + `complete` 改 hook 链

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | `workflow/stage/{accept,execute,review,rework}/` 每包新增 `fsm_binding.go`，所有状态转移走 `fsm.Next()`；**删除** `workflow/stage/complete/` 整个子包 |
| 新建的 | `workflow/hook/` 子包承载 §7.4 的 on_completion hook 链（archive/report/cleanup/notify 四个 hook 实现） |
| 改的配置 | `manifest/config/fsm/workflow.yaml` 的 `on_completion_hooks:` 节 |
| 编译闸门 | `go build ./app/mvp/...` |
| 验证 | 跑一个完整 workflow，stage 从 accept → 直接 → completed（不经过 complete stage），归档/报告/清理/飞书通知四个 hook 异步触发且 `workflow_event` 表有记录 |
| 回滚 | revert |
| 合并节点 | 必须等 PR-04、PR-06 合完 |
| **破坏性提醒** | 删 `stage/complete/` 是**公开破坏**——所有引用它的地方都会编译失败，必须在同一 PR 里一次性改完。回滚需要同时回滚前端"完成阶段"的相关页面文案 |

##### PR-09 · engine 包瘦身

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | (a) `engine/executor_dispatch.go`（447 行）搬到 `workflow/executor/dispatcher.go`，同时按 §5.2 拆成 `router.go + retry.go + heartbeat.go` 三个文件；(b) `engine/context_compressor.go`（505 行）的压缩逻辑改为**锁外异步**——压缩任务投递到 goroutine 池，锁只保护 DB 写，AI 调用在锁外；(c) **删除** `engine/executor_bridge.go`（144 行）的全局 hook 变量 |
| 不动的 | `engine/aider_runner.go`、`engine/scheduler_*.go`、`engine/watchdog.go`、`engine/sse_hub.go`、`engine/parser_*.go`、`engine/review_*.go`——这些功能正确、结构尚可，本波次不碰 |
| 编译闸门 | `go build ./app/mvp/...` + 新增并发测试验证"持锁期间不发起 AI 调用" |
| 验证 | 压测：10 个并发项目同时触发 `CompressTaskContext`，观察 `project_lock` 持锁时间 P99 < 50ms（当前 > 5000ms） |
| 回滚 | revert |
| 合并节点 | 必须等 PR-07 合完（`executor_dispatch` 搬家后 `lifecycle` 的调用路径要同步改） |

##### PR-10 · CEL 规则引擎替换硬编码 precheck

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/domain/rule/cel_runner.go` 承载 `cel-go` 的 AST 预编译 + LRU 缓存；`engine/review_precheck.go`（325 行）的硬编码 8 规则改为从 `manifest/config/rules/precheck.yaml` 加载 |
| 新建配置 | `manifest/config/rules/{precheck,accept_guard,autonomy_strategies}.yaml` |
| 新增依赖 | `github.com/google/cel-go` |
| 编译闸门 | `go build ./app/mvp/...` + CEL 规则加载器单元测试 |
| 验证 | 把原来的 8 条硬编码规则改成 YAML，跑集成测试，行为与改前一致 |
| 回滚 | revert，feature flag `precheck.use_cel=false` 可直接切回硬编码路径 |
| 合并节点 | 必须等 PR-04 合完（规则引擎依赖 `domain/fsm` 的 guard 接口） |

##### PR-11 · `autonomy/decision_center` 事件驱动化

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | `workflow/autonomy/decision_center.go`（731 行）的 switch 改为事件订阅者——从 `workflow/eventstream/` 订阅 `workflow_transition` 事件，按 FSM 做下一步判断；删除当前的定时扫表轮询逻辑 |
| 依赖 | PR-02 的 `workflow_transition_log` 必须能 emit 事件（本 PR 补上 emit 逻辑） |
| 编译闸门 | `go build ./app/mvp/...` |
| 验证 | 压测前后 DB QPS 对比，自治决策相关的 `SELECT` 查询降幅 ≥ 80% |
| 回滚 | revert，feature flag `autonomy.use_event_bus=false` 可切回轮询 |
| 合并节点 | 必须等 PR-02 合完 |

##### PR-12 · Legacy 表与 engine 遗留文件物理删除

| 字段 | 说明 |
|---|---|
| 动的迁移文件 | `000015_drop_legacy_mvp_task.up.sql`（DROP `mvp_task`、`mvp_task_dependency`、`mvp_task_resource_lock`、`mvp_task_log`、`mvp_accept_evidence`、`mvp_review_issue`、`mvp_accept_issue` 等表）+ `.down.sql`（不可逆，仅留警告注释） |
| 动的 Go 代码 | 删除 `engine/` 下所有已无引用的文件；删除 `workflow/verification/service.go`（1929 行）中已被 `stage/*` 与 `domain/judge` 替代的代码路径 |
| 编译闸门 | `go build ./app/mvp/...`，预期大量未使用 import 的 lint 警告需清掉 |
| 验证 | 全量集成测试跑通；`grep -r mvp_task\.`、`grep -r review_issue\.` 零引用 |
| 回滚 | **高风险**：删表不可逆。合并前必须：(a) PR-01~11 已在生产上线 **至少两周** 无故障；(b) 全库 dump 一份快照；(c) 有明确的业务方签字 |
| 合并节点 | 所有前序 PR 合完并观察两周 |

##### PR-13 · Executor Kernel（Agent Loop + 工具白名单 + 多模态证据）

| 字段 | 说明 |
|---|---|
| 动的迁移文件 | `000016_executor_registry.up.sql` + `000017_execution_artifact.up.sql`（含 §10.3 的 FK/CHECK/UNIQUE/数据权限字段） |
| 动的 Go 代码 | 新建 `workflow/executor/kernel/` 子包：`executor.go`（Plugin 接口）、`agent_loop.go`（LLM 主循环）、`tool_registry.go`（工具白名单 + 黑名单双过滤）、`artifact_writer.go`（OSS/MinIO 上传 + DB 元数据写入）、`request.go`（统一 Request/Result/AIConfig）、`dispatcher.go`（按 `execution_kind` 路由到 Plugin） |
| 改的文件 | `workflow/executor/dispatcher.go`（PR-09 搬来的 447 行）调用点改为 `kernel.Run(ctx, req, plugin)`；`workflow/domain/task/` 的 `DomainTask` 加上 §4.4 新增 5 个字段的 Go struct |
| 新增依赖 | `github.com/invopop/jsonschema`（工具 schema 描述）、`github.com/minio/minio-go/v7`（证据上传）、复用已有 `github.com/anthropics/anthropic-sdk-go`（Vision LLM） |
| 编译闸门 | `go build ./app/mvp/...` + kernel 单元测试：(a) 工具白名单在调用前拦截越权；(b) Agent Loop 到达 MaxSteps 正确退出；(c) 多模态 artifact 上传失败时任务整体失败 |
| 验证 | 用一个 NoopPlugin（只返回 "done"）走完整流程，`execution_artifact` 表有 `tool_call_log` 记录，`executor_registry` 表有 `noop` 注册 |
| 回滚 | revert + `make db-down STEPS=2` |
| 合并节点 | 必须等 PR-07 合完（lifecycle 的 CreateProject 事务边界已建立，Kernel 才能在事务外独立运行） |
| **关键约束** | **Kernel 不感知具体 kind**——所有 code/browser/mobile/game 差异都在 Plugin 层；Kernel 只负责 LLM loop、工具白名单、证据写入 |

**独立可发行性**：Kernel 必须从第一天就具备**不依赖 `mvp_*` 表**运行的能力——`executor_registry` / `execution_artifact` 是唯二的持久化表，其余所有状态走 Request/Result；这样未来可以把 `workflow/executor/kernel/` 整个子包抽成独立仓库 `easymvp-executor`（详见 §9.10）。本 PR 的 code review 必须拦截任何 `import "workflow/orchestrator"` 之类的反向依赖。

##### PR-14 · CodeExecutor / DirectApply（自研代码执行器 + Aider 兜底）

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/executor/plugin/code/` 子包：`plugin.go`（实现 `Plugin` 接口）、`tools.go`（read_file / write_file / apply_patch / run_tests 四个工具，每个带 AllowPaths 校验）、`diff_validator.go`（写入前 dry-run，拒绝超出 `domain_task.allow_paths` 的改动）；`workflow/executor/plugin/aider_fallback/` 子包作为兜底 |
| 改的文件 | `engine/aider_runner.go` 保留但只在 `executor_code=aider_fallback` 时被调用；新 `code` kind 默认走自研 |
| 编译闸门 | `go build ./app/mvp/...` + 单测：(a) `apply_patch` 试图写入 `.gitignore` 时被拒；(b) `write_file` 超出 AllowPaths 时工具调用失败；(c) 同一 DomainTask 连续调用 `run_tests` 不被速率限制误杀 |
| 验证 | 用一个真实代码任务（"加一个 /ping 接口"）让 CodeExecutor 完成，`execution_artifact` 有 `diff_ref` + `tool_call_log`；同任务切换到 `aider_fallback` 也能跑完 |
| 回滚 | revert；feature flag `executor.code.use_native=false` 强制走 aider_fallback |
| 合并节点 | 必须等 PR-13 合完 |
| **验收标准** | 跑 100 个真实 code 任务，自研 CodeExecutor 的 `.gitignore` 越界率 = 0（aider 在同样样本上的基线是 7~15%）——这是整个执行器家族存在的理由 |

##### PR-15 · BrowserExecutor（Playwright-go + Vision LLM 联调验收）

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/executor/plugin/browser/` 子包：`plugin.go`、`tools.go`（navigate / click / fill / screenshot / evaluate / wait_for / extract_dom 七个工具，每个带 `allow_domains` 校验）、`vision_bridge.go`（截图 → Vision LLM → 结构化反馈） |
| 新增依赖 | `github.com/playwright-community/playwright-go`（启动时需 `playwright install chromium`，Docker 镜像需预装） |
| 编译闸门 | `go build ./app/mvp/...` + 单测：(a) `navigate` 到白名单外的 domain 被拒；(b) 截图尺寸压缩到 Vision LLM 可接受范围；(c) headless 模式崩溃时 Plugin 返回结构化错误而非 panic |
| 验证 | 用一个 `execution_kind=browser` 的任务："访问 http://localhost:3000 登录页，输入 admin/admin，截图确认看到 dashboard" —— `execution_artifact` 有 3 张截图 + `dom_snapshot` + Vision LLM 的结论 JSON |
| 回滚 | revert |
| 合并节点 | 必须等 PR-13 合完 |
| **独立发布预埋** | Browser Plugin 的 `tools.go` 工具定义文件必须与 Kernel 一样保持对 `mvp_*` 表零依赖——未来单独打包成 `easymvp-browser-executor` 二进制时无需拔数据库 |

##### PR-16 · MobileExecutor-Android（adb/uiautomator2 + Vision LLM）

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/executor/plugin/mobile_android/` 子包：`plugin.go`、`adb.go`（封装 `adb shell` 调用）、`uiautomator.go`（通过 uiautomator2 dump 控件树）、`tools.go`（install_apk / launch_app / tap / swipe / input_text / screenshot / back / home 八个工具） |
| 新增依赖 | 本地 `adb` 命令（Docker 镜像需装 `android-tools-adb`） + Python `uiautomator2` 服务（可选，降级为纯 adb） |
| 编译闸门 | `go build ./app/mvp/...` + 单测：(a) `install_apk` 只接受 `allow_packages` 列表中的包名；(b) 设备离线时返回结构化错误；(c) 截图尺寸与 Vision LLM 兼容 |
| 验证 | 连一台真机或模拟器，用任务："安装 com.example.app，打开后点击登录按钮，截图确认"——`execution_artifact` 含 2 张截图 + `dom_snapshot`（控件树） |
| 回滚 | revert |
| 合并节点 | 必须等 PR-13 合完；**本 PR 是 P1**——不阻塞 pss_claude 主线合并，可以推迟到 code/browser 稳定后再启动 |
| **范围约束** | 仅 Android，iOS 需要 Mac CI + Appium/XCUITest，明确不在 pss_claude 路线图内 |

##### PR-17 · GameExecutor-Web（复用 BrowserExecutor + 帧序列截取）

| 字段 | 说明 |
|---|---|
| 动的 Go 代码 | 新建 `workflow/executor/plugin/game_web/` 子包，**大量复用** `plugin/browser/`——只新增两个工具：`capture_frames`（每 N 毫秒截一帧，组成 `frame_seq` artifact）、`eval_game_state`（调用 game 页面暴露的 `window.__gameState` 读取状态） |
| 新增依赖 | 复用 PR-15 的 playwright-go，无新依赖 |
| 编译闸门 | `go build ./app/mvp/...` + 单测：帧序列写入 OSS 时按 `<workflow_id>/<domain_task_id>/frame_<n>.jpg` 分桶 |
| 验证 | 跑一个 Phaser/Cocos/H5 小游戏的任务："进入游戏首屏，点击开始按钮，运行 5 秒后截取帧序列，Vision LLM 判断是否在加载/跑图/结算"——`execution_artifact` 有 `frame_seq` 元数据 |
| 回滚 | revert |
| 合并节点 | 必须等 PR-15 合完；**本 PR 是 P2**——Web 游戏项目上线时再启动 |
| **范围约束** | 仅 Web/H5 游戏；Unity/UE 原生游戏需要进程内 hook 或 C++ 侧 IPC，不在 pss_claude 路线图内 |

#### 13.5.4 执行顺序小结

```
第 1 周：PR-01 / PR-02 / PR-03（波次 1，三个一起合）
第 2 周：PR-04 → PR-05 → PR-06（波次 2，按顺序合）
第 3 周：PR-07 → PR-08 → PR-09（波次 3，按顺序合）
第 4 周：PR-10 / PR-11（波次 4，并行合）
        同时启动：PR-13（Kernel）、PR-14（CodeExecutor）、PR-15（BrowserExecutor）三人并行开发
第 5 周：PR-13 合并 → PR-14 / PR-15 合并（波次 5 P0 三件套）
观察两周
第 7 周：PR-12（legacy 最终清理，不可逆）
第 8 周+：PR-16（Mobile，P1，按需启动）/ PR-17（Game Web，P2，按需启动）
```

每个 PR 合并前必须通过**铁律 17 闸门**：PR 流水线开头跑 `uptime`，1 分钟负载 > 80 则排队等候，< 50 才放行。

#### 13.5.5 不做的事（同样重要）

- **不换框架**：GoFrame v2.10 保留，不迁 gin/echo
- **不换 ORM**：`g.DB()` 保留，不引入 GORM/ent
- **不换 ID 策略**：Snowflake 保留，不换 UUID
- **不动前端技术栈**：Vue 3 + Vben Admin 保留
- **不拆仓库**：继续单体仓，不做多仓库拆分
- **不引入新消息队列**：事件总线第一版就用本地 channel + `workflow_event` WAL 表，不上 Kafka/NATS/RabbitMQ
- **不引入服务网格**：单体部署，不上 K8s/Istio
- **不动 Git Worktree 隔离机制**：`mvp_task_workspace` 表保留
- **不动 Watchdog 双保险**：`engine/watchdog.go` 430 行一行不碰
- **不动 RBAC 五级数据权限**：所有新表自带 `created_by/dept_id`，继承现有 `ApplyDataScope` 与 `CheckProjectAccess`

这一节是防止 PR 过程中"顺手重写一下"——每一次"顺手"都会把 12 个 PR 变成 24 个。

---

<a id="14"></a>
## 14. 与 ChatDev/MetaGPT/AutoGen 的差异化定位

写完上面所有章节，必须回答一个问题：**做完 Atlas，它比 ChatDev/MetaGPT 强在哪？**

### 14.1 横向对比

| 维度 | ChatDev | MetaGPT | AutoGen | CrewAI | OpenHands | EasyMVP | **Atlas** |
|---|---|---|---|---|---|---|---|
| 角色数 | 4 | 5+ | 可变 | 可变 | 1 | 5 | 5+ 可扩展 |
| 分类×角色×等级 三维预设 | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ 继承 |
| 多执行器支持 | ❌ | ❌ | ✅ | ✅ | ❌（自身 runtime） | ✅ 6 种 | ✅ 6 种 + 配置化 |
| 硬规则+LLM融合判官 | ⚠️ 只有测试 | ⚠️ 只有 QA | ⚠️ 只有 LLM critic | ⚠️ 只有 manager | ❌ | ✅ 2 路 | ✅ 3 路 |
| 体验评审判官 | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ 一等公民 |
| 七层自治 | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ 继承 |
| 显式 FSM 工作流 | ❌ | ⚠️ 隐式 | ❌ | ❌ | ✅（12 态） | ⚠️ 三套打架 | ✅ 三套分层 |
| 人工介入一等公民 | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| 数据权限五级 | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| 六边形架构 | ❌ | ❌ | ❌ | ❌ | ⚠️ | ❌ | ✅ |
| 完成阶段 | ✅ 独立 Phase | ❌ | ❌ | ❌ | ❌ | ⚠️ 独立 stage | ✅ 终态 hook |
| 观测性 | ❌ | ❌ | ❌ | ❌ | ⚠️ | ⚠️ | ✅ 四支柱 |

### 14.2 Atlas 的独有优势（继承自 EasyMVP）

1. **三维预设**：业内唯一
2. **硬规则+LLM融合判官**：业内唯一
3. **七层自治**：业内唯一
4. **五角色 + 体验评审师**：最细分工

### 14.3 Atlas 新增的优势（相对 EasyMVP）

1. **显式 FSM 三层分离**：解决三套状态机打架
2. **终态 hook 链**：对齐工业级引擎实践
3. **六边形架构**：domain 零 IO，可单元测试
4. **配置化执行器**：热加载、健康度路由
5. **四大观测支柱**：替代自建观测表
6. **CEL 规则引擎**：替代硬编码 + DB 混合
7. **数据约束到库层**：外键、CHECK、唯一

### 14.4 Atlas **相对劣势**（需要承认）

1. **不如 CrewAI 社区活跃**：开源生态 0，需要自建
2. **不如 AutoGen 通用**：专注 SDLC，不适合"通用多 agent 对话"场景
3. **不如 ChatDev 学术声誉**：没有论文
4. **不如 OpenHands 单体简单**：多了编排复杂性

---

<a id="15"></a>
## 15. 风险、权衡、已知限制

### 15.1 设计决策的权衡

| 决策 | 收益 | 代价 |
|---|---|---|
| 三套 FSM 分层而非统一 | 各层表达力最强 | 需要 cascade.go 级联映射 |
| YAML 编排 + 代码枚举 | 编译期检查 + 运行时可调 | 两层配置，学习成本 |
| 取消 StageComplete | 对齐成熟引擎 | 失去"完成阶段产物"的概念（需要用 hook 产物替代） |
| CEL 规则引擎 | 统一 DSL，表达力强 | CEL 学习曲线 |
| 六边形严格分层 | 可测试性 | 包数量翻倍 |
| 46 表 + 20+ 外键 | DB 层一致性 | migration 复杂，删数据变难 |
| 放弃 L6/L7 业务表 | 观测更强 | 迁移期间需要搭 Prometheus/Loki/Tempo |
| 三路 judge 并行 | 体验师一等公民 | 验收阶段时延 = max(三路) |

### 15.2 已知限制

1. **YAML FSM 编排的表达力上限**
   - 能表达：顺序、分支、guard、超时、hook、重试
   - 不能表达：循环、变量赋值、复杂条件（这是故意的，学自 Temporal 教训）
   - 需要这些时：在 `application` 层写 Go 代码，不进 YAML

2. **CEL 学习曲线**
   - 团队需要 1~2 周适应 CEL 语法
   - 缓解：第一批规则由核心开发者写，业务配置师改不写

3. **hook 失败不阻断主流程**
   - 好：workflow 终态稳定
   - 坏：如果归档失败被忽视，产物会丢
   - 缓解：hook 失败必须告警，监控面板必看

4. **单体部署 vs 分布式**
   - 第一版是单体（GoFrame monorepo）
   - 本地事件总线用 channel + WAL 表
   - 扩到分布式需要换事件总线（Kafka/NATS）——预留接口

5. **体验评审师权重 40% 是拍脑袋**
   - 需要真实数据校准
   - 预留 `accept.yaml` 里的权重表供后续调整
   - 必要时可按 category 分不同权重（软件开发重 LLM judge，数据分析重体验师）

6. **不支持跨 workflow_run 的状态共享**
   - 每个 workflow_run 是独立的 FSM 实例
   - 多 workflow 协作的场景需要在 application 层实现，不进 FSM

### 15.3 风险清单

| 风险 | 概率 | 影响 | 缓解 |
|---|---|---|---|
| FSM 设计遗漏某状态转移 | 中 | 高 | 完整单元测试覆盖所有合法转移 + 所有非法转移 |
| CEL 表达式执行性能 | 低 | 中 | 预编译 + LRU 缓存 AST |
| 六边形分层导致包爆炸 | 中 | 中 | CI 限制包数量，超阈值提醒 |
| 46 表 migration 出错 | 中 | 高 | 分 5~8 个 migration 文件逐步上线 |
| 多 judge 并行拖慢验收 | 中 | 中 | 监控 accept 总时延，超阈值降级为串行 |
| 配置 hot reload bug | 低 | 高 | SIGHUP 后校验配置，失败回滚 |
| 数据权限五级过滤性能 | 低 | 中 | `(dept_id, created_by)` 复合索引 |
| 归档 hook 失败丢产物 | 中 | 高 | hook 失败必须告警 + 日级巡检脚本 |

### 15.4 需要用户拍板的决策

在开工前需要明确回答这 5 个问题：

1. **迁移路径选 A/B/C？**（推荐 A）
2. **体验评审师权重 40% 是否合适？**（可按 category 分）
3. **是否保留 L6/L7 业务表？**（推荐放弃，走 Prometheus）
4. **是否取消 StageComplete 独立阶段？**（推荐取消，对齐调研）
5. **第一版是否支持多租户？**（推荐不支持，后续再说）

---

## 附录 A：术语表

| 术语 | 定义 |
|---|---|
| **Atlas** | 本方案的暂定代号 |
| **EasyMVP** | 前身项目，本方案继承其产品思想 |
| **FSM** | Finite State Machine，有限状态机 |
| **CEL** | Common Expression Language（google/cel-go） |
| **三维预设** | category × role × level 的角色模板 |
| **融合判官** | hard_rule + llm_judge + experience_reviewer 的加权决策 |
| **六阶段工作流** | design → review → execute → accept → rework（无 complete） |
| **七层自治** | L1 执行 / L2 调度 / L3 策略 / L3.5 人工 / L4 决策 / L5 学习 / L6-L7 下放观测 |
| **六边形架构** | Ports and Adapters，Alistair Cockburn |
| **终态 hook** | workflow 进入终态后触发的异步动作链 |
| **level cascade** | 跨 FSM（workflow/stage/task）的级联映射 |

## 附录 B：调研证据索引

**AI 多 agent 平台：**
- [ChatDev ChatChainConfig.json](https://github.com/OpenBMB/ChatDev/blob/chatdev1.0/CompanyConfig/Default/ChatChainConfig.json) —— 独立 Testing / EnvironmentDoc / Manual Phase
- [MetaGPT software_company.py](https://github.com/geekan/MetaGPT/blob/main/metagpt/software_company.py) —— 5 角色 hire 顺序
- [MetaGPT qa_engineer.py](https://github.com/geekan/MetaGPT/blob/main/metagpt/roles/qa_engineer.py) —— QA 角色消息订阅
- [AutoGen Reflection Pattern](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/design-patterns/reflection.html) —— Coder+Critic pub/sub
- [CrewAI Processes](https://docs.crewai.com/concepts/processes) —— Hierarchical Manager agent
- [OpenHands AgentState](https://github.com/All-Hands-AI/OpenHands/blob/main/openhands/core/schema/agent.py) —— 12 态状态机

**工作流引擎：**
- [Temporal Workflow Execution](https://docs.temporal.io/workflow-execution) —— 6 CloseStatus + 异步 archival
- [Temporal Archival](https://docs.temporal.io/temporal-service/archival) —— 归档是后台任务不是状态
- [Temporal: Beyond State Machines](https://temporal.io/blog/temporal-replaces-state-machines-for-distributed-applications) —— YAML state machine 反模式警告
- [Cadence Workflows](https://cadenceworkflow.io/docs/concepts/workflows) —— CloseStatus 0-5 枚举
- [Argo Workflow Archive](https://argo-workflows.readthedocs.io/en/latest/workflow-archive/) —— pending archive label
- [Argo Exit Handlers](https://argo-workflows.readthedocs.io/en/latest/walk-through/exit-handlers/) —— onExit template
- [Prefect States](https://docs.prefect.io/v3/concepts/states) —— 19 状态 + State Change Hooks
- [Airflow Callbacks](https://airflow.apache.org/docs/apache-airflow/stable/administration-and-deployment/logging-monitoring/callbacks.html) —— on_success_callback
- [Conductor Workflow Definition](https://conductor.netflix.com/documentation/configuration/workflowdef/index.html) —— JSON DSL

**EasyMVP 内部证据：**
- `admin-go/app/mvp/internal/workflow/acceptance/decision_reducer.go:97` —— 40%/60% 融合公式原址
- `admin-go/app/mvp/internal/workflow/rolecatalog/catalog.go:178` —— experience_reviewer 角色定义
- `admin-go/app/mvp/internal/workflow/orchestrator/transition.go:59-65` —— stageOrder 五阶段（含 StageComplete）
- `admin-go/app/mvp/internal/engine/context_compressor.go:192-226` —— 持锁调 AI 反模式
- `admin-go/manifest/sql/seed/mysql_seed.sql` —— 36 条空壳预设

## 附录 C：FAQ

**Q1：为什么不直接改进 EasyMVP，非要新写一个？**
A：EasyMVP 的问题是结构性的（编排散在 6 处、三套状态机打架、数据层零约束），单点修改会被结构反噬。本方案前 20% 的精力改结构，后 80% 的精力做事；否则就是 80% 精力打补丁，20% 做事。

**Q2：体验评审师权重为什么是 40%？**
A：拍脑袋的初始值，但有理由——它是 Atlas 区别于其他平台的独有优势，应该作为主判官。最终需要真实数据校准。不同 category 可以不同权重。

**Q3：为什么取消 StageComplete？**
A：调研 7 个成熟工作流引擎（Temporal/Cadence/Argo/Prefect/Airflow/n8n/Conductor）无一例外都不把 Archived 做成状态。5 个 AI agent 平台中只有 ChatDev 有独立完成阶段，且理由是产出真实 artifact。EasyMVP 的 StageComplete 做的是"打包+报告+回收+通知"——这四件事在所有成熟系统里都是 hook。

**Q4：为什么不用 looplab/fsm 这种现成库？**
A：可以用，但不是核心——核心是"状态枚举硬编码 + 转移矩阵硬编码 + guard 表达式 YAML"这套分层。用 looplab/fsm 只是实现细节。

**Q5：CEL 学不会怎么办？**
A：CEL 语法比 SQL 还简单（没有 JOIN、没有子查询）。团队 1~2 周能上手。核心开发者写第一批规则作为示例，业务配置师改不写。

**Q6：46 表是不是还太多？**
A：可以继续压。但很多表是企业级必须（9 张系统权限表、7 张 AI 供应商配置）。业务核心其实只有 20 张左右。进一步压缩收益低，不值得。

**Q7：为什么不支持 Kubernetes / Argo 集成？**
A：第一版不支持，不是技术问题，是聚焦问题。Atlas 的核心价值是 "AI 角色编排 + 融合判官 + 七层自治"，不是"又一个工作流引擎"。K8s 集成是后续扩展。

**Q8：性能目标是什么？**
A：
- `CreateProject` < 500ms
- `Dispatch` 单批次 < 200ms
- `Accept` 三路 judge 并行 < 30s（max of 三路）
- `workflow_transition` 日志写入 < 10ms
- `ResolveRole` < 1ms（LRU 缓存命中）
- 单实例支持 100 并发 workflow_run

**Q9：可以中途改设计决策吗？**
A：
- 改 YAML 配置：随时（热加载）
- 改 CEL 规则：随时
- 改 FSM 状态枚举：需要 migration + 代码修改，走大版本
- 改六边形分层：需要大重构，避免
- 改融合公式：需要单元测试 + 灰度

**Q10：和 EasyMVP 现有的 `pss_claude` 分支什么关系？**
A：Atlas 可以在 `pss_claude` 分支上实现，也可以新开仓库。推荐新开仓库，原因见 13.4 节（路径 A Greenfield）。如果要在同仓库实现，应该新开 `atlas_main` 分支而不是在 `pss_claude` 上继续改——`pss_claude` 保留为 EasyMVP 的文档/研究分支。

---

*文档结束。*

*这是一份完整方案，不是草案。每一节都可以独立评审、独立反驳。接下来的工作应该是：用户对 15.4 节的 5 个拍板问题给出答案，然后开始阶段 A 的 migration 文件编写。*

*如果这份方案本身还有任何结构性问题需要推翻，欢迎推翻。不推翻就开工。*
