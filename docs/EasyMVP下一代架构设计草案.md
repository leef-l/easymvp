# EasyMVP 下一代架构设计草案（激进版）

> 文档定位：这是一份**零包袱**的架构重设方案，不是治理方案。目标是回答"如果现在从零开始，应该怎么设计"，并给出从当前系统到新系统的可执行迁移路径。
>
> 适用前提：用户痛点是"项目流程总是卡，反复调整修复"。根因不是某个 bug，而是**编排层、状态层、配置层、数据层的结构性问题**同时存在。任何单点补丁都会在别处冒出新问题——这正是"反复修"的由来。
>
> 写作日期：2026-04-10
> 分支：pss_claude
> 作者视角：假设现有代码不存在，只继承"做什么"（业务意图），不继承"怎么做"（实现形态）。

---

## 一、现状诊断：为什么总卡

以下结论全部基于已完成的代码走查，不是推测。每条都带可定位的证据。

### 1.1 编排层：控制流散落在 6 个地方

当前工作流从 `CreateProject` 到 `Accept` 经过这些地方：
`workflow_lifecycle.go` → `transition.go` → `registry.go` Init 巨构 → `verification/service.go` 1929 行大杂烩 → `decision_center.go` 731 行 switch → `executor_dispatch.go` 448 行 → `chat_engine.go` 267 行。

这里面真正描述"状态怎么流转"的只有 `transition.go`（124 行，干净），其余 5 个文件都在做"顺便还要 xxx"的事。结果是：改一个状态流转，要改 4 个文件；加一种执行器分支，要在 3 个 switch 里都加一遍；一个字段含义变了，7 处都得同步。

**这是"反复修"的第一来源。**

### 1.2 配置层：双套并行，36 条默认预设全坏

`mvp_role_preset` 种子数据里存在两套平行数据集：

- **Set 1（V2 英文 category_code）**：9 类 × 4 角色 = 36 条，全部 `model_id=0`、`system_prompt=''`、`default_executor` 是占位值。
- **Set 2（Legacy 中文 project_category）**：有真实 `model_id` 和完整 `system_prompt`。

`role_resolver.go` 的 fallback 链会优先找 V2 英文集 → 命中空 preset → 回落到 Legacy 中文集 → 命中后又被 `category_code` 映射层再翻译一次。每次解析 5~20ms，零缓存，错误被静默吞掉。

这意味着——**不是 experience_reviewer 缺了**，**是全部 5 个角色的默认预设都缺**。`docs/EasyMVP体验评审师接入方案与计划.md` 那份方案书的前提就是错的，按它执行完依然是坏的。

**这是"反复修"的第二来源：每次你以为修好了一个角色，其实只是把 Legacy 集拖回来当兜底，V2 集还是空的。**

### 1.3 状态层：三套状态机互相打架

- `WorkflowRun`：9 个状态
- `StageRun`：5 个状态
- `DomainTask`：8 个状态（pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched）

三者没有形式化的映射关系。`autonomy/decision_center.go` 用一个巨型 switch 在这三个状态空间之间手动翻译。`verification/service.go:1929` 行里又重复做了一遍。

**后果**：DomainTask 走到 `bug_dispatched`，但 StageRun 还停在 `executing`，WorkflowRun 还认为是 `running`，然后 UI 看到"明明有 bug 却显示进行中"。用户的"卡"很多时候就是这种状态漂移。

### 1.4 数据层：55 张表零外键零 CHECK

- 55 张表**全部**没有外键约束
- 55 张表**全部**没有 CHECK 约束
- 34 个 JSON 字段散落在各表（schema 不受约束）
- `sys_delete_queue` 缺 `created_by` / `dept_id`（违反铁律 13）
- `mvp_task`（9 索引）和 `mvp_domain_task` 并行存在（Legacy 未清）
- `mvp_accept_issue` 过索引（8 个），`mvp_review_issue` 欠索引（2 个）
- `NULL` 与 `NOT NULL DEFAULT 0` 混用，`ai_*` 表系列和 `mvp_*` 表系列规则不一致

**结果**：任何跨表一致性都靠应用层手写判断。一处漏写，数据就开始发散；一发散，流程就卡在"验证器认为状态不对、实际业务又想继续"的死角里。

### 1.5 质量门：硬规则与 LLM 判官各跑各的

`qualitygate/standard.go` 定义了 6 个 VerificationStandard，其中 3 个是死的（`analysis.default`、`creative.default`、`{family}.default`）。`DetectProjectSignals()` 在一次 accept 里被调 3 次（性能浪费）。`rule_engine.go:96-114` 把硬编码规则和 DB 规则混在一起，**未实现的 rule_code 会被静默跳过**——这是最隐蔽的一类"明明没通过却算通过"的 bug。

`accept_evidence` / `accept_issue` 没有唯一约束，Rerun 时会产生重复行。

### 1.6 执行器层：分发逻辑和会话逻辑粘在一起

`executor_dispatch.go` 448 行同时做了：路由选择 + 聊天模式分支 + 重试 + 心跳 + 日志落盘。`chat_engine.go` 里 `workDir` 是**硬编码**的 `/www/wwwroot/project/easymvp`（line 85、line 157），5 分钟的 magic timeout 也是硬编码。`SendMessage` 与 `SendFeishuMessage` 90% 代码重复。

### 1.7 并发与锁：压缩在锁内调 AI

`context_compressor.go:192-226` 持有 project 级锁期间**发起 AI 调用**。AI 慢一秒，整个 project 就卡一秒。用户看到的"系统假死"大概率来自这里。

### 1.8 事务边界：CreateProject 是 8 步无事务

`workflow_lifecycle.go` 的 `CreateProject` 是 8 步流程，**没有一个包在事务里**。第 5 步失败，前 4 步已经落库，DB 里留下半死不活的 project。下一次查询就会触发各种 nil 判断，然后不同 handler 各自"兜底"，"兜底"不一致又产生新 bug。

---

## 二、设计原则（激进版七条）

1. **只有一套配置**：中文 `project_category` 全部干掉，只留 `category_code`。Legacy 迁移脚本走一次，种子重写。
2. **状态机形式化**：一个 FSM（`looplab/fsm` 或等价手写），一份 YAML/DSL 定义，所有状态转移必须走它。禁止 `UPDATE status = ?` 类裸写。
3. **编排层只编排，不做 IO**：`orchestrator` 包只能调 `domain.Service` 接口，不能直接访问 DB、AI、文件系统。所有副作用走 `Port`（六边形架构）。
4. **锁内零外部调用**：任何持锁的代码路径都禁止调用 AI、HTTP、长时 DB。压缩、评审、判官必须是异步任务。
5. **数据层带约束**：外键、CHECK、唯一约束一次性补齐。JSON 字段降到 10 个以内，剩下的 24 个要么拆列要么拆子表。
6. **规则引擎换 DSL**：`rule_engine.go` 的硬编码 + DB 混合模式改成 CEL（Common Expression Language）。硬规则、LLM 规则、实验规则统一写在 CEL 表达式里，一行配置替代一个函数。
7. **五角色而非四加一**：`experience_reviewer` 从第一天就是正式角色，不做任何"第五号补丁"式的兼容层。五个角色一视同仁，三维预设表也是 5 行而非 4+1 行。

---

## 三、数据层重设计

### 3.1 目标：55 表 → 46 表

| 动作 | 表数变化 | 说明 |
|---|---|---|
| 删 Legacy：`mvp_task`、`mvp_task_*` 系列 | -4 | V2 已完成，Legacy 可以整体下线 |
| 合并：seven-layer autonomy 11 表 → 6 表 | -5 | 去掉 audit/shadow/snapshot 三层冗余 |
| 合并：`accept_evidence` + `review_issue` + `accept_issue` → `quality_finding` | -2 | 三者本质都是"某次判断的产物" |
| 新增：`workflow_transition_log`（审计用） | +1 | 取代散落在 5 个表里的 last_* 字段 |
| 新增：`role_preset_binding`（显式绑定表） | +1 | 取代 fallback 链的隐式推断 |

最终约 46 表。

### 3.2 约束补齐（一次性 migration）

```sql
-- 外键（选关键 20 条）
ALTER TABLE mvp_domain_task
  ADD CONSTRAINT fk_domain_task_project
    FOREIGN KEY (project_id) REFERENCES mvp_project(id) ON DELETE RESTRICT,
  ADD CONSTRAINT fk_domain_task_stage
    FOREIGN KEY (stage_run_id) REFERENCES mvp_stage_run(id) ON DELETE RESTRICT;

-- CHECK 约束（状态白名单）
ALTER TABLE mvp_workflow_run
  ADD CONSTRAINT chk_workflow_status
    CHECK (status IN ('designing','reviewing','executing','accepting','completing','rework','paused','failed','cancelled'));

-- 唯一约束（杀重复）
ALTER TABLE quality_finding
  ADD CONSTRAINT uk_finding_scope
    UNIQUE (workflow_run_id, stage_code, rule_code, target_id);

-- 铁律 13 补齐
ALTER TABLE sys_delete_queue
  ADD COLUMN created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN dept_id    BIGINT UNSIGNED NOT NULL DEFAULT 0;
```

### 3.3 JSON 字段瘦身（34 → 10）

|原字段 | 去向 |
|---|---|
| `metadata_json`（11 处）| 结构化为 3~5 个显式列 |
| `tags_json` | 拆子表 `*_tag`（多对多） |
| `config_snapshot_json` | 保留（版本化快照天然 schemaless） |
| `llm_raw_response_json` | 保留（外部原始输出） |
| `rule_payload_json` | 改 CEL 表达式文本 |

### 3.4 `mvp_role_preset` 重建

只保留一套英文 `category_code`，5 角色 × 9 分类 × 3 级别 = 135 行，全部带真实 `model_id` 和 `system_prompt`。废弃 Legacy 中文集。`role_resolver.go` 的 4 层 fallback 改为**单层精确匹配 + 显式报错**。

---

## 四、领域层重设计

### 4.1 六边形分层

```
┌─────────────────────────────────────────────────┐
│  ui/http/api   ←— 只调 application              │
├─────────────────────────────────────────────────┤
│  application   ←— 用例编排（CreateProject等）   │
├─────────────────────────────────────────────────┤
│  domain        ←— 纯逻辑：FSM、评分、规则       │
├─────────────────────────────────────────────────┤
│  port          ←— 接口：Repo/AI/FS/Lock/Event   │
├─────────────────────────────────────────────────┤
│  adapter       ←— GoFrame/MySQL/Claude/本地FS   │
└─────────────────────────────────────────────────┘
```

**硬约束**：
- `domain` 不引用 `gf`、不引用 `g.DB()`、不引用 `claude_code` 客户端
- `application` 不直接写 SQL，必须通过 `port.ProjectRepo`
- `adapter` 不调用 `domain` 以外的业务逻辑

这一层拆出来之后，1929 行的 `verification/service.go` 会自然降到 3~4 个各自 300 行以内的文件。

### 4.2 engine 包重划

当前 10 个文件的重划：

| 旧文件 | 新归属 | 动作 |
|---|---|---|
| `chat_engine.go` | `adapter/chat/` | 拆成 `claude.go` + `feishu.go`，去重 90% |
| `role_resolver.go` | `domain/preset/` | 单层查询 + LRU 缓存 |
| `executor_dispatch.go` | `application/dispatch/` | 拆成 `router.go` + `retry.go` + `heartbeat.go` |
| `executor_bridge.go` | **删除** | 全局 hook 变量下掉 |
| `parser_extractor.go` | `adapter/llm_parse/` | Claude 正则抽到配置 |
| `context_compressor.go` | `application/compress/` | **锁外异步化** |
| `review_precheck.go` | `domain/precheck/` | 8 规则 → CEL |
| `config.go` | `adapter/config/` | 统一缓存策略 |
| `sse_hub.go` | `adapter/sse/` | **保留**，修预订阅时序 |
| `workflow_lifecycle.go` | `application/lifecycle/` | **全部包事务** |

### 4.3 状态机：一份 YAML

```yaml
# workflow.fsm.yaml
machine: workflow_run
initial: designing
states:
  designing:
    on: { submit_design: reviewing }
  reviewing:
    on:
      approve: executing
      reject: designing
  executing:
    on:
      done: accepting
      bug_found: rework
  rework:
    on: { fix_submitted: executing }
  accepting:
    on:
      passed: completing
      failed: rework
      manual_review: paused
  paused:
    on: { resumed: accepting }
  completing:
    on: { archived: archived }
  archived: { terminal: true }
  failed: { terminal: true }
  cancelled: { terminal: true }
guards:
  approve: "workflow_run.review_score >= 60 && quality_finding.blocker_count == 0"
  passed:  "acceptance.hard_score >= 40 && acceptance.llm_score >= 60 && acceptance.blocker_count == 0"
```

`StageRun` 和 `DomainTask` 也各有一份 YAML。三台 FSM 之间的映射写在 `domain/fsm/cascade.go`，**一个文件管完**，而不是散在 5 个包里。

### 4.4 规则引擎：CEL 替代硬编码

```go
// 旧：硬编码 8 条 precheck 规则
if stage.BatchNo == 0 { stage.BatchNo = 1 } // 静默修正
if len(stage.Deliverables) == 0 { return fail("no deliverables") }
// ... 再 6 条

// 新：precheck.cel.yaml
rules:
  - code: BATCH_NO_REQUIRED
    expr: "stage.batch_no > 0"
    on_fail: block
    message: "batch_no must be > 0"
  - code: DELIVERABLES_REQUIRED
    expr: "size(stage.deliverables) > 0"
    on_fail: block
    message: "at least one deliverable required"
```

静默自动修复彻底禁止。要么通过要么 block，修复必须走人工或显式 workflow。

---

## 五、执行层重设计

### 5.1 Executor 配置化

```yaml
# executors.yaml
executors:
  - name: claude_code
    priority: 100
    needs_workspace: true
    timeout: 5m
    retry: { max: 2, backoff: exponential }
    env_from_preset: [api_key, model]
  - name: aider
    priority: 90
    needs_workspace: true
  # ... 其余
selection_policy: priority  # 或 round_robin / weighted
```

`executor/auto.go` 里的硬编码优先级（`claude_code > aider > codex > gemini > openhands > chat`）全部搬到 YAML。换顺序不改代码。

### 5.2 分发解耦

```
dispatcher  → router        → selects executor
            → pre_hook      → load preset, compose prompt
            → runner        → run executor in isolated workspace
            → post_hook     → parse output, emit events
            → retry_policy  → on failure, backoff or fail
```

每一层都是独立的 interface，独立测试，独立替换。

### 5.3 事件驱动替代轮询

当前 `decision_center.go` 的 731 行 switch 是定时扫 DB。改为：
- 每次状态变更通过 `port.EventBus.Publish` 发事件
- `decision_center` 订阅事件，按 FSM 做下一步判断
- 不再扫表轮询，DB 负载骤降

事件总线第一版可以就用本地 channel + WAL 表（`workflow_event`），不依赖 Kafka / NATS。

---

## 六、观测层

当前几乎零观测。新增：

1. **结构化日志**：每一次 FSM transition 一条 log，字段统一（workflow_run_id / from / to / guard_result / duration_ms）
2. **metrics**：Prometheus 暴露 `workflow_stage_duration_seconds{stage=...}`、`quality_finding_total{severity=...}`、`executor_run_seconds{executor=...}`
3. **trace**：关键链路（`CreateProject` → `Dispatch` → `Accept`）打 OpenTelemetry span
4. **audit**：`workflow_transition_log` 表保留所有状态变更，永不物理删

这四条加上以后，"反复修"时至少能**回答"卡在哪"**，而不是像现在这样要临时加 `g.Log().Info` 重新跑一遍。

---

## 七、迁移路径（分四阶段，每阶段可独立上线）

### 阶段 A：数据层重建（2~3 天）
- 新写 `000100_next_gen_schema.up.sql`（新表）和 `.down.sql`
- 种子数据：重建 135 行 role_preset（含 experience_reviewer）
- Legacy 表打 `deprecated_at` 字段但不删
- 灰度：新 Project 走新表，旧 Project 继续走旧表

**验收**：新建项目能完整走完 5 个角色的解析，零 fallback。

### 阶段 B：domain + FSM 层（3~5 天）
- 写 `domain/fsm` 包，三份 YAML
- 写 `domain/preset` 新 role_resolver（带缓存）
- 写 `domain/precheck` 的 CEL 规则加载器
- `application/lifecycle` 包事务

**验收**：一个新项目从 `CreateProject` 到 `Accept` 完整走 FSM，所有状态变更进 `workflow_transition_log`。

### 阶段 C：engine 拆分（4~6 天）
- `chat_engine.go` 拆 `adapter/chat/{claude,feishu}.go`
- `executor_dispatch.go` 拆 `router/retry/heartbeat`
- `context_compressor.go` 异步化（锁外）
- 删 `executor_bridge.go`

**验收**：旧 engine 包可以整包删除，不影响功能。

### 阶段 D：qualitygate 重建（2~3 天）
- 删 3 个死 standard
- `DetectProjectSignals` 加请求级 memoize
- `rule_engine` 全面 CEL 化
- LLM Judge 融合逻辑搬到 `domain/acceptance`

**验收**：一次 accept 只调 `DetectProjectSignals` 一次；未实现的 rule_code 直接报错而不是静默跳过。

---

## 八、新旧对比表

| 维度 | 当前 | 下一代 | 收益 |
|---|---|---|---|
| 角色数 | 4 正式 + 1 补丁 | 5 一等公民 | 无补丁层 |
| 默认预设可用率 | V2 集 0%（36 条全空） | 100%（135 条全配） | 零 fallback 兜底 |
| 状态机 | 3 套隐式互相翻译 | 3 套 YAML + 1 文件映射 | 状态漂移消失 |
| 外键约束 | 0 | 20+ 关键外键 | 跨表一致性由 DB 保证 |
| CHECK 约束 | 0 | 所有状态列带白名单 | 脏数据不再进库 |
| JSON 字段 | 34 | ≤10 | schema 可控 |
| 表数量 | 55 | 46 | 删 Legacy + 合并冗余 |
| 锁内 AI 调用 | 有（压缩器） | 禁止 | 假死消失 |
| CreateProject 事务 | 8 步裸跑 | 单事务 | 半死项目消失 |
| 规则引擎 | 硬编码 + DB 混杂 | CEL 单一来源 | 加规则零改代码 |
| Executor 选择 | 硬编码优先级 | YAML 配置 | 调序不改代码 |
| 编排层代码量 | `verification/service.go` 1929 行 | 3~4 文件各 <300 行 | 可读性 |
| 观测 | 零 | log/metric/trace/audit | 卡哪能看见 |
| "反复修"发生点 | 6 处 | 根因消除 | — |

---

## 九、风险与权衡

### 9.1 承认的风险

1. **迁移期并行成本**：阶段 A 灰度期间新旧表并存，需要双写开关。约一周。
2. **CEL 学习曲线**：团队之前写 Go 条件分支，现在要写 CEL 表达式。第一周会慢。
3. **YAML FSM 的边界情况**：某些需要跨状态机联动的场景（如 DomainTask 的 bug_found 同时触发 StageRun 回退），YAML 表达力有限，可能需要在 `cascade.go` 里写少量 Go 胶水。这是接受的折中。
4. **删 Legacy 的窗口**：`mvp_task*` 系列删除前必须确认没有旧代码路径引用。阶段 A 先改名加 `deprecated_` 前缀，观察两周再物理删。

### 9.2 不动的东西（为什么不动）

- **Go 1.25 + GoFrame v2.10** ：框架没有系统性缺陷，换框架收益 < 成本
- **Vue 3 + Vben Admin v5.7** ：前端不是痛点
- **雪花 ID**：全局唯一、趋势递增，没理由换 UUID
- **9 个 category、6 个 stage**：业务划分正确，改了反而丢业务语义
- **3 维预设（category × role × level）**：这是 EasyMVP 的核心抽象，只是当前数据烂，不是模型烂

### 9.3 与《体验评审师接入方案》的关系

那份方案书不是错，而是**范围不够**。它识别对了"experience_reviewer 要接入"这个需求，但没发现整个 V2 预设集都是空的。本方案包含该方案的全部目标，并把根因一起修掉。原方案的 Phase 0-4 可以作为阶段 A 的子任务清单直接复用。

---

## 十、如果只能做一件事

如果时间极度紧张只能做一件事，做**阶段 A + FSM YAML 化**（约 5 天）。

原因：
- 阶段 A 修掉了 36 条空预设（"反复修"根源之一）
- FSM YAML 化修掉了三态漂移（另一根源）
- 这两件事不依赖其他阶段，可独立上线
- 其他阶段都是"更好"，这两件事是"不坏"

其余阶段可以在后续迭代中滚动推进。

---

## 附录 A：关键代码证据索引

为后续评审与追溯，列出本文档每个诊断对应的原始证据位置：

- 1.1 编排散落：`admin-go/app/mvp/internal/workflow/orchestrator/registry.go` 1021 行 Init 巨构；`verification/service.go` 1929 行
- 1.2 双套预设：`admin-go/manifest/sql/seed/mysql_seed.sql` Set 1 36 行 model_id=0 + Set 2 Legacy 中文集
- 1.3 三状态机：`WorkflowRun`/`StageRun`/`DomainTask` 三处定义，映射写在 `decision_center.go` 731 行 switch
- 1.4 零约束：`admin-go/manifest/sql/migrations/000001_baseline_schema.up.sql` 55 表无 FK 无 CHECK
- 1.5 死 standard：`qualitygate/standard.go` 6 standard 中 `analysis.default`/`creative.default`/`{family}.default` 无引用；`rule_engine.go:96-114` 静默跳过
- 1.6 执行器耦合：`executor_dispatch.go` 448 行；`chat_engine.go:85,157` 硬编码 workDir
- 1.7 锁内 AI：`context_compressor.go:192-226`
- 1.8 无事务：`workflow_lifecycle.go` CreateProject 8 步

---

## 附录 B：术语

- **FSM**：Finite State Machine，有限状态机
- **CEL**：Common Expression Language，Google 的嵌入式表达式语言，Go 有官方实现 `cel-go`
- **六边形架构**：Ports and Adapters，Alistair Cockburn 提出的分层模型
- **铁律 13**：本项目 `CLAUDE.md` 规定所有业务表必须含 `created_by` + `dept_id` + 5 级 DataScope
- **铁律 14**：所有 DDL 通过 `golang-migrate`，种子数据通过 `mysql_seed.sql`
- **铁律 17**：服务器 CPU 繁忙 > 80 时停手，< 50 才恢复

---

*文档结束。这是一份草案，欢迎反驳、补充、否决任一条。如果方向认可，下一步是把阶段 A 拆成具体 migration 文件和 seed SQL。*
