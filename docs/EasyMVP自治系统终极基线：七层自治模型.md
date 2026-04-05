# EasyMVP 自治系统终极基线：七层自治模型

> 合并自：架构师B《自治宪法操作系统》+ 架构师A《七层自治操作系统》+ 原七层基线
> 本文档是**唯一实施基线**，所有后续迭代以此为准。
> 前序设计文档在本文档发布后归档至 `docs/超前理念文档/`，不再作为实施依据。

---

## 实施现状（2026-04-06 更新）

| 层级 | 模块 | 状态 | 关键组件 |
|------|------|------|----------|
| L1 证据层 | Evidence | ✅ 已上线 | 决策日志、审计记录 |
| L2 执行层 | Execution | ✅ 已上线 | ActionDispatcher（8/8 回调）、6 种执行器 |
| L3 能力层 | Capability | ✅ 已上线 | PolicyEngine / RiskGate / CircuitBreaker |
| L3.5 协作层 | Collaboration | ✅ 已上线 | 飞书审批卡片、少人值守 |
| L4 契约层 | Contract | ✅ 已上线 | DecisionCenter 决策中台 |
| L5 策略层 | Strategy | ✅ 代码完成 | Planner + 6 策略 + Actuator 效果跟踪 |
| L6 目标层 | Objective | ✅ 代码完成 | Sensor 态势感知 + Situation + ObjectiveService |
| L7 元认知层 | Meta-Cognition | ✅ 代码完成 | MetaObserver + Learner + MetaAssessor + MetaTuner |

**执行器：** aider / openhands / claude_code / codex_cli / gemini_cli / chat（6 种，全部注册 + 入库）
**数据库：** 55 张表 | **灰度开关：** 14 个（全部关闭）| **前端：** 11 个 workflow 页面

---

## 0. 三方合并说明

| 来源 | 采纳内容 |
|------|----------|
| 架构师B | Intent→Contract→Evidence→StateChange 主线、IntentContract + EffectContract、自治宪法理念、四个一等公民、四个闭环、六条稳定性铁律 |
| 架构师A | 宪法嵌入类型系统（编译器保证）、DecisionMeta 结构体、四个闭环调用链、态势感知、EMA 学习 + 三重保护、Stage-Scoped Routing、Operator + Build/Operate 双主线、LLM 边界矩阵、10 级灰度 |
| 原七层基线 | 七层架构骨架、完整 Go 结构体、SQL DDL、能力词表、策略函数、评分公式、文件清单、实施路线图 |

**合并原则：理念取架构师B，工程取架构师A，骨架取原基线。冲突时以可编译性为准。**

---

## 1. 核心理念

系统治理的核心不是"谁来做"，而是**"什么状态变化被允许发生"**。

主线：`Intent → Contract → Execute → Evidence → StateChange`

六句话定义：

1. 宪法高于一切——且宪法嵌入类型系统，不靠人遵守
2. 目标层只做减法
3. 策略层先准入后优化，基于统一态势
4. 契约层包含五契约包（Intent + Task + Outcome + Acceptance + Effect）
5. 所有状态推进必须绑定证据
6. 学习机制优先学习边界，不优先学习激进最优

---

## 2. 自治宪法（Constitution）

宪法不是一个运行时层，而是**渗透在所有层里的不可突破约束**。

### 2.1 三个结构性不可能

| 编号 | 不可能 | 实现方式 |
|------|--------|----------|
| 1 | 高层不可能直接执行 | L7/L6/L5 返回 `*Recommendation`，不持有 Executor 引用 |
| 2 | 无证据不可能推进状态 | `TransitionState()` 签名强制要求 `evidence *ExecutionEvidence` 参数 |
| 3 | 学习不可能让系统更激进 | Tuner 输出 `Direction=aggressive` 时，`AutoApplicable` 编译期写死 `false` |

### 2.2 六条稳定性铁律

| # | 铁律 | 保证方式 |
|---|------|----------|
| 1 | 高层永远不能直接改世界 | L7/L6/L5 不持有 Executor 引用（编译器保证） |
| 2 | 目标层永远只能收紧系统 | AdmissionControl 只有 Deny/DowngradeCondition，无 Upgrade 动作 |
| 3 | 元认知层永远只能调参数 | Tuner 输出 TuneRecommendation，无 ActionPlan |
| 4 | 所有状态推进必须绑定证据 | TransitionState 签名强制 evidence 参数 |
| 5 | LLM 不能决定权限和状态 | 状态迁移/权限/闸门求值/CAS/锁 = 规则引擎唯一决策 |
| 6 | 学习只学边界，不学激进 | 保守方向可自动(confidence>0.7)，激进方向必须人工 |

### 2.3 不可调参数（宪法保护）

以下参数 Tuner **永远不能修改**，只能由人工在 UI 或数据库中变更：

| 参数 | 说明 |
|------|------|
| `autonomy_level` | 自治等级（manual/supervised/autonomous） |
| `max_side_effect_level` | 最大允许副作用等级 |
| `human_mandatory_points` | 强制人工节点列表 |
| `decision_level` 的 C 级判定规则 | 什么场景必须人工 |
| `blast_radius_min_level` 映射表 | 影响面 → 最低决策级别 |

---

## 3. 七层架构总览

```
┌─────────────────────────────────────────────────────────────┐
│ L7  元认知层 Meta-Cognition                                  │
│     Observer · Assessor · Tuner                              │
│     输出：TuneRecommendation（不持有任何执行器引用）            │
├─────────────────────────────────────────────────────────────┤
│ L6  目标层 Objective                                         │
│     ProjectObjective · AdmissionControl                      │
│     输出：AdmissionResult（Allowed/Denied + Conditions）      │
├─────────────────────────────────────────────────────────────┤
│ L5  策略层 Policy                                            │
│     Sensor · Planner · PolicyEngine · RiskGate               │
│     6 个策略函数                                              │
│     输出：ActionPlan（推荐动作 + DecisionMeta + 回滚方案）      │
├─────────────────────────────────────────────────────────────┤
│ L4  契约层 Contract                                          │
│     IntentContract · TaskContract · OutcomeContract           │
│     AcceptanceContract · EffectContract                      │
│     输出：ContractBundle（五契约包 + 版本号）                   │
├─────────────────────────────────────────────────────────────┤
│ L3  能力层 Capability                                        │
│     CapabilityProfile（三层认证）· ContractMatcher             │
│     Stage-Scoped Routing                                     │
│     输出：MatchResult（最优匹配 + 评分 + 降级路径）             │
├─────────────────────────────────────────────────────────────┤
│ L2  执行层 Execution                                         │
│     ExecutorRegistry · HealthChecker                         │
│     唯一持有 Executor 引用的层                                 │
│     输出：ExecutionResult + ExecutionEvidence                 │
├─────────────────────────────────────────────────────────────┤
│ L1  证据与治理层 Evidence & Governance                        │
│     EvidenceStore · AuditLog · RoleView · HumanOverride      │
│     输出：所有层的可查询证据底座                                │
└─────────────────────────────────────────────────────────────┘
```

**层间规则：**
- 每层只和相邻层交互，不跨层调用
- 上层约束下层，下层向上层反馈
- 每层独立灰度开关，可独立回滚
- L1 是例外：作为证据底座，所有层可向 L1 写入证据、从 L1 读取证据

---

## 4. 四个一等公民（DecisionMeta）

每一个跨层传递的决策对象必须嵌入：

```go
// DecisionMeta 决策元数据。
// 嵌入所有跨层传递的决策对象（ActionPlan / MatchResult / TuneRecommendation）。
type DecisionMeta struct {
    Confidence          float64 // 0-1，决策置信度
    EvidenceSufficiency float64 // 0-1，支撑证据充分度
    Reversibility       string  // full / partial / none
    BlastRadius         string  // task / batch / stage / workflow / project
}

// BlastRadius → 最低决策级别
var BlastRadiusMinLevel = map[string]string{
    "task":     "A",
    "batch":    "A",
    "stage":    "B",
    "workflow": "B",
    "project":  "C",
}

// 证据充分度阈值
const (
    EvidenceThresholdAuto   = 0.7 // 自动执行最低证据充分度
    EvidenceThresholdAssist = 0.4 // 辅助决策最低证据充分度
    // 低于 0.4 → 拒绝决策，要求更多证据
)

// Validate 校验决策元数据是否允许自动执行。
func (m *DecisionMeta) Validate(requestedLevel string) *ValidationResult {
    result := &ValidationResult{Allowed: true}

    // 1. 证据不足 → 拒绝
    if m.EvidenceSufficiency < EvidenceThresholdAssist {
        result.Allowed = false
        result.Reason = "evidence_insufficient"
        return result
    }

    // 2. 影响面 → 可能升级决策级别
    minLevel := BlastRadiusMinLevel[m.BlastRadius]
    if levelToInt(requestedLevel) < levelToInt(minLevel) {
        result.UpgradeTo = minLevel
        result.Reason = "blast_radius_requires_upgrade"
    }

    // 3. 低置信度 + 不可逆 → 必须人工
    if m.Confidence < 0.5 && m.Reversibility == "none" {
        result.UpgradeTo = "C"
        result.Reason = "low_confidence_irreversible"
    }

    return result
}
```

---

## 5. 各层详细定义

### 5.1 L7 元认知层（Meta-Cognition）

**职责：** 观测系统自身行为，评估效果，校准参数。不做业务决策，做"关于决策的决策"。

**三要素：**

| 组件 | 输入 | 输出 | 约束 |
|------|------|------|------|
| Observer | 决策事件流 | ObservationRecord | 纯记录，零副作用，异步写 |
| Assessor | 观测记录集 | AssessmentResult + Drift[] | 只读，不改参数 |
| Tuner | 偏差报告 | TuneRecommendation[] | 保守可自动，激进必须人工 |

```go
type ObservationRecord struct {
    DecisionID     int64
    DecisionType   string   // policy_match / contract_match / gate_check / admission
    InputSnapshot  g.Map
    Output         g.Map
    Meta           *DecisionMeta // 四个一等公民
    Outcome        string        // success / failure / neutral（延迟回填）
    EffectScore    float64       // -1 ~ 1（延迟回填）
    HumanOverride  bool
    OverrideReason string
}

type AssessmentResult struct {
    PolicyAccuracy     float64
    GateFalsePositive  float64
    GateFalseNegative  float64
    HumanOverrideRate  float64
    MatchAccuracy      float64
    CostEfficiency     float64
    Drifts             []Drift
}

type Drift struct {
    Parameter    string
    CurrentValue float64
    OptimalValue float64
    Confidence   float64
    Evidence     string
}

type TuneRecommendation struct {
    Parameter      string
    CurrentValue   interface{}
    SuggestedValue interface{}
    Direction      string      // conservative / aggressive
    Reasoning      string
    Confidence     float64
    AutoApplicable bool        // aggressive 时编译期写死 false
    RiskLevel      string      // low / medium / high
}
```

**学习信号强度：**

| 场景 | 强度 | 权重 |
|------|------|------|
| A 级自动执行，人工未干预 | 弱 | 0.3 |
| B 级，人工批准 | 中 | 0.6 |
| B 级，人工驳回 | **强** | 1.0 |
| C 级，人工给出方案 | 学习样本 | 1.0 |

**学习三重保护：**

| 保护 | 规则 | 防什么 |
|------|------|--------|
| 样本量保护 | SampleCount < 10 → 不输出建议 | 小样本过拟合 |
| 置信度衰减 | 7 天无新样本 → Confidence *= 0.9 | 陈旧数据主导 |
| 变更幅度限制 | 单次调整 ≤ 当前值 20% | 剧烈波动 |

```go
type LearningRecord struct {
    ScopeKey      string  // 如 "software_dev:implementer:aider"
    Dimension     string  // 如 "success_rate"
    CurrentValue  float64
    SampleCount   int
    Confidence    float64
    LastUpdatedAt time.Time
}

func UpdateEMA(record *LearningRecord, newSample float64, alpha float64) {
    record.CurrentValue = alpha*newSample + (1-alpha)*record.CurrentValue
    record.SampleCount++
    record.Confidence = min(1.0, float64(record.SampleCount)/30.0)
    record.LastUpdatedAt = time.Now()
}
```

**灰度：**
- `meta_cognition_enabled=0` → Observer 不记录，Assessor/Tuner 不运行
- `meta_auto_tune_enabled=0` → Tuner 只生成建议不自动应用
- `learner_enabled=0` → EMA 不更新

---

### 5.2 L6 目标层（Objective）

**职责：** 定义项目的目标和约束。**只做减法，不做优化，不主动发起动作。**

```go
type ProjectObjective struct {
    // 交付目标
    DeliveryGoal       string
    QualityFloor       float64    // 最低质量标准 0-1

    // 预算约束
    TokenBudget        int64      // Token 总预算（0=不限）
    TimeBudgetHours    float64
    CostBudgetCents    int64

    // 风险约束
    RiskTolerance      string     // conservative / balanced / aggressive
    MaxAutoRetries     int        // 硬限，到了就停
    MaxAutoReworks     int
    MaxAutoReplans     int

    // SLA
    DeadlineAt         *time.Time
    MaxStallMinutes    int

    // 自治边界
    AutonomyLevel      string     // manual / supervised / autonomous

    // 副作用约束
    MaxSideEffectLevel   string   // 允许的最大副作用等级
    AllowedStateChanges  []string // 允许的状态变更列表
    HumanMandatoryPoints []string // 强制人工节点
}
```

**AdmissionControl（准入控制）：**

```go
type AdmissionResult struct {
    Allowed    bool
    DenyReason string
    Conditions []string // 附加条件（如"降级到轻量执行器"）
}
```

| 检查项 | 条件 | 动作 |
|--------|------|------|
| 预算检查 | TokensConsumed > TokenBudget × 90% | 拒绝高成本动作 |
| 时间检查 | Now > DeadlineAt | 拒绝非必要动作 |
| 重试上限 | RetryCount >= MaxAutoRetries | 拒绝自动重试，升级人工 |
| 返工上限 | ReworkRounds >= MaxAutoReworks | 拒绝自动返工 |
| 重规划上限 | ReplanCount >= MaxAutoReplans | 拒绝自动重规划 |
| 副作用上限 | EffectContract.SideEffectLevel > MaxSideEffectLevel | 拒绝或升级人工 |
| 自治级别 | AutonomyLevel = manual | 所有动作需人工确认 |

**目标约束下层示例：**

```
RiskTolerance = "conservative"
  → L5：熔断阈值从 5 降为 3，B 级不自动执行
  → L3：匹配偏向历史表现好的画像
  → L2：并发度从 20 降到 10

TokenBudget 消耗 90%
  → L6：拒绝高成本动作
  → L3：切换到轻量执行器
```

**灰度：** `objective_enabled=0` → AdmissionControl 始终返回 Allowed=true

---

### 5.3 L5 策略层（Policy）

**职责：** 读取统一态势，先判断"允不允许做"，再判断"怎么做更好"。

**原则：先准入，后优化。**

#### 5.3.1 态势感知（Sensor）

所有需要判断当前状态的组件（AdmissionControl / Planner / Operator / Tuner）统一读取 Situation 快照，不各自查数据库。

```go
type Situation struct {
    WorkflowRunID int64
    ProjectID     int64
    ProjectFamily string
    CategoryCode  string
    Progress      *ProgressMetrics
    Health        *HealthMetrics
    Resource      *ResourceMetrics
    Trend         *TrendMetrics
    ActiveStage   string
    SnapshotAt    time.Time
}

type ProgressMetrics struct {
    TotalTasks     int
    CompletedTasks int
    RunningTasks   int
    FailedTasks    int
    PendingTasks   int
    CompletionRate float64
    CurrentBatchNo int
    TotalBatches   int
    BatchProgress  float64
}

type HealthMetrics struct {
    ConsecutiveFailures int
    RecentFailureRate   float64
    AvgTaskDuration     int64
    MedianTaskDuration  int64
    P95TaskDuration     int64
    RetryCount          int
    EscalationCount     int
    ReworkRounds        int
    AcceptRounds        int
    StaleTaskCount      int
    HeartbeatMissCount  int
}

type ResourceMetrics struct {
    RunningConcurrency  int
    MaxConcurrency      int
    ResourceUtilization float64
    LockedResourceCount int
    ConflictCount       int
    TokensConsumed      int64
    EstimatedTokensLeft int64
}

type TrendMetrics struct {
    FailureRateTrend    string // rising / stable / falling
    DurationTrend       string
    ThroughputTrend     string
    RecentAvgDuration   int64
    PreviousAvgDuration int64
    RecentFailureRate   float64
    PreviousFailureRate float64
}

type AnomalySignal struct {
    Type       string  // failure_spike / duration_drift / throughput_drop / budget_warning
    Severity   string  // info / warning / critical
    Message    string
    Metrics    g.Map
    Confidence float64
}
```

**Sensor 三种模式：**

| 模式 | 触发 | 说明 |
|------|------|------|
| 事件驱动 | 每次决策前 | `Perceive()` 实时采集 |
| 定时巡检 | 每 60s | `Patrol()` 所有活跃工作流 |
| 异常检测 | 采集后 | `DetectAnomalies()` 提取异常信号 |

#### 5.3.2 策略函数

```go
type Strategy interface {
    Name() string
    Evaluate(ctx context.Context, sit *Situation, trigger string) *ActionPlan
    Applicable(sit *Situation, trigger string) bool
    Priority() int
}

type ActionPlan struct {
    PlanID          string
    Trigger         string
    Situation       *Situation
    Steps           []ActionStep
    Meta            *DecisionMeta // 四个一等公民
    Reasoning       string
    RollbackPlan    *ActionPlan
    ExpectedOutcome string
    DecisionLevel   string
}

type ActionStep struct {
    StepNo     int
    ActionType string
    TargetID   int64
    Parameters map[string]interface{}
    Condition  string
    Timeout    int
}
```

**6 个内置策略：**

| 策略 | 决策域 | 独立开关 |
|------|--------|----------|
| AdaptiveRetryStrategy | 失败时的重试方式 | adaptive_retry_enabled |
| EngineSelectionStrategy | 执行前的执行器选择 | engine_selection_enabled |
| BatchAdjustStrategy | 批次推进时的并发度调整 | batch_adjust_enabled |
| CostGuardStrategy | 动作前的成本检查 | cost_guard_enabled |
| ProactiveReplanStrategy | 趋势恶化时提前重规划 | proactive_replan_enabled |
| QualityGateStrategy | 验收时质量标准动态调整 | quality_gate_enabled |

**Planner 决策流程：**

```
事件触发
  → Sensor.Perceive() → Situation
  → L6 AdmissionControl.Check(situation, objective)
     → Denied → 记录 + 通知人工 → 结束
     → Allowed ↓
  → [快速路径] PolicyEngine.Match(trigger)
     → 命中且 strategy_override=false → 直接使用规则结果
     → 未命中或 strategy_override=true → 进入策略评估
  → [策略评估] 遍历 Applicable 策略，按 Priority 排序
     → 选择 Confidence 最高的 ActionPlan
  → [DecisionMeta 校验] plan.Meta.Validate(plan.DecisionLevel)
     → 影响面/置信度/可逆性不足 → 升级决策级别
  → [风险检查] RiskGate.Check(plan, situation)
  → 输出 ActionPlan
```

**灰度：**
- `patrol_enabled=0` → Sensor 不巡检
- `strategy_enabled=0` → 走原 PolicyEngine 路径，策略函数不参与

---

### 5.4 L4 契约层（Contract）

**职责：** 形式化描述"为什么做"、"需要什么"、"成功产出什么"、"怎么验证"、"允许什么副作用"。

#### 五契约包

```go
// ContractBundle 五契约包。
type ContractBundle struct {
    Intent     *IntentContract
    Task       *TaskContract
    Outcome    *OutcomeContract
    Acceptance *AcceptanceContract
    Effect     *EffectContract
    Version    int // 每次 Escalate 后 +1
}
```

#### IntentContract（意图契约）

```go
type IntentContract struct {
    ObjectiveID     int64
    IntentType      string  // build / fix / refactor / test / deploy / diagnose
    BusinessGoal    string
    Priority        string  // critical / high / medium / low
    RiskClass       string  // high_risk / medium_risk / low_risk
    BudgetClass     string  // high_cost / medium_cost / low_cost
    Urgency         string  // immediate / normal / deferred
}
```

#### TaskContract（任务契约）

```go
type TaskContract struct {
    RequiredCapabilities  []string
    PreferredCapabilities []string
    AccuracyRequirement   string  // high / medium / low
    CreativityRequirement string
    ContextScope          string  // local / batch / global
    RequiresFileAccess    bool
    RequiresSandbox       bool
    RequiresHumanReview   bool
    RequiresIsolation     bool
    MaxTokenBudget        int64
    MaxDurationSeconds    int
    CostSensitivity       string  // high / medium / low
    SensitivityLevel      string  // public / internal / confidential
    DomainTags            []string
}
```

#### OutcomeContract（结果契约）

```go
type OutcomeContract struct {
    RequiredOutputs []RequiredOutput
    SuccessCriteria []string
    FailureCriteria []string
}

type RequiredOutput struct {
    Name      string // 产出名称
    Type      string // file / text / json / metrics
    Schema    g.Map
    Mandatory bool
}
```

#### AcceptanceContract（验收契约）

```go
type AcceptanceContract struct {
    VerifyMethod       string   // rule / llm_judge / human / auto
    RuleSet            []string
    QualityFloor       float64
    EvidenceRequired   []string // diff / test_result / coverage / lint
    ManualReviewPolicy string   // always / on_failure / never
}
```

#### EffectContract（副作用契约）

```go
type EffectContract struct {
    ReadSet         []string
    WriteSet        []string
    ExternalActions []string // api_call / webhook / notification
    SideEffectLevel string   // none / low / medium / high
    Rollbackability string   // full / partial / none
    ApprovalClass   string   // A / B / C
}
```

#### 契约来源优先级

| 优先级 | 来源 |
|--------|------|
| 1 | 架构师显式指定（任务 JSON 中的 contract 字段） |
| 2 | 系统自动推断（InferContract） |
| 3 | 运行时升级（EscalateContract） |

#### 自动推断规则（InferContract）

```go
func InferFullContract(blueprint *TaskBlueprint) *ContractBundle {
    bundle := &ContractBundle{Version: 1}

    // Intent
    bundle.Intent = &IntentContract{
        IntentType:  inferIntentType(blueprint),
        Priority:    "medium",
        RiskClass:   "medium_risk",
        BudgetClass: "medium_cost",
    }

    // Task（从 role_type / affected_resources / description 推断）
    bundle.Task = &TaskContract{ContextScope: "local", CostSensitivity: "medium"}
    switch blueprint.RoleType {
    case "implementer":
        bundle.Task.RequiredCapabilities = []string{"code_edit"}
        bundle.Task.RequiresFileAccess = true
    case "auditor":
        bundle.Task.RequiredCapabilities = []string{"code_review"}
    case "architect":
        bundle.Task.RequiredCapabilities = []string{"requirement_analysis"}
        bundle.Task.ContextScope = "global"
    case "operator":
        bundle.Task.RequiredCapabilities = []string{"failure_analysis", "risk_assessment"}
        bundle.Task.ContextScope = "global"
    }
    if len(blueprint.AffectedResources) >= 5 {
        bundle.Task.RequiredCapabilities = append(bundle.Task.RequiredCapabilities, "multi_file_edit")
    }

    // Outcome（从 role_type 推断默认产出）
    bundle.Outcome = inferOutcome(blueprint)

    // Acceptance（从 role_level 推断严格度）
    bundle.Acceptance = inferAcceptance(blueprint)

    // Effect（从 affected_resources 推断副作用）
    bundle.Effect = &EffectContract{
        WriteSet:        blueprint.AffectedResources,
        SideEffectLevel: "medium",
        Rollbackability: "full",
        ApprovalClass:   "A",
    }

    return bundle
}
```

#### 失败后契约升级（EscalateContract）

```go
func EscalateContract(bundle *ContractBundle, failurePattern string) {
    bundle.Version++

    // Intent 升级
    bundle.Intent.RiskClass = escalateRisk(bundle.Intent.RiskClass)

    // Task 升级
    switch failurePattern {
    case PatternCapabilityMismatch:
        bundle.Task.AccuracyRequirement = "high"
        bundle.Task.ContextScope = "global"
        bundle.Task.CostSensitivity = "low"
    case PatternStructural:
        bundle.Task.RequiredCapabilities = appendUnique(
            bundle.Task.RequiredCapabilities, "multi_file_edit")
        bundle.Task.RequiresSandbox = true
    case PatternResourceConflict:
        bundle.Task.RequiresIsolation = true
    }

    // Outcome 升级
    bundle.Outcome.SuccessCriteria = appendUnique(bundle.Outcome.SuccessCriteria, "test_pass")

    // Acceptance 升级
    bundle.Acceptance.QualityFloor = max(bundle.Acceptance.QualityFloor, 0.85)
    if bundle.Acceptance.VerifyMethod == "auto" {
        bundle.Acceptance.VerifyMethod = "rule"
    }
    bundle.Acceptance.EvidenceRequired = appendUnique(
        bundle.Acceptance.EvidenceRequired, "test_result", "coverage")

    // Effect 升级
    if bundle.Effect.ApprovalClass == "A" {
        bundle.Effect.ApprovalClass = "B"
    }
}
```

---

### 5.5 L3 能力层（Capability）

#### 能力画像三层认证

| 层级 | 来源 | 更新频率 |
|------|------|----------|
| Declared（声明能力） | 角色+执行器能力表 | 静态 |
| Certified（认证能力） | 基准测试 | 低频 |
| Observed（观测能力） | 历史执行数据 | 持续 |

```go
type CapabilityProfile struct {
    RoleType      string
    ExecutionMode string
    ModelID       int64
    RoleLevel     string

    DeclaredCaps   []string
    CertifiedCaps  []string
    ObservedScores map[string]float64

    AccuracyScore    float64
    CreativityScore  float64
    SpeedScore       float64

    AvgTokenCost     int64
    AvgDurationSec   int
    CostPerTask      float64

    MaxContextScope string
    SupportsFile    bool
    SupportsSandbox bool

    HistoricalSuccessRate float64
    SampleCount           int
}
```

#### 匹配评分公式

```
Score = w1×CapabilityScore + w2×QualityScore + w3×CostScore + w4×HistoryScore

CapabilityScore = 必须能力覆盖率×0.7 + 优选能力覆盖率×0.3
QualityScore    = 1 - |contract.AccuracyReq - profile.AccuracyScore|
CostScore       = (1 - profile.CostPerTask / maxCost)
HistoryScore    = profile.SuccessRate × min(1, profile.SampleCount / 30)

权重按 CostSensitivity 调整：
  省钱优先：w1=0.30 w2=0.20 w3=0.35 w4=0.15
  效果优先：w1=0.30 w2=0.35 w3=0.10 w4=0.25
  均衡：    w1=0.30 w2=0.25 w3=0.20 w4=0.25
```

`HistoryScore` 中 `min(1, SampleCount/30)` — 样本不足 30 时自动降权，防止小样本过拟合。

```go
type MatchResult struct {
    Profile         *CapabilityProfile
    Eligible        bool
    Score           float64
    CapabilityScore float64
    QualityScore    float64
    CostScore       float64
    HistoryScore    float64
    Meta            *DecisionMeta // 四个一等公民
    Reasoning       string
}
```

#### Stage-Scoped Routing

防止逐任务自由匹配导致运行抖动：

```
1. 批次启动时，聚合该批次所有任务的契约
2. 选定批次主执行链（role_type + execution_mode + model）
3. 个别任务契约与主链不兼容时，才做局部重路由
4. 记录重路由原因 → L7 Observer
```

#### 核心能力词表

```go
// 分析类
CapRequirementAnalysis  = "requirement_analysis"
CapTaskDecomposition    = "task_decomposition"
CapFailureAnalysis      = "failure_analysis"
CapRiskAssessment       = "risk_assessment"
CapCostEstimation       = "cost_estimation"
CapSecurityAnalysis     = "security_analysis"

// 执行类
CapCodeEdit             = "code_edit"
CapMultiFileEdit        = "multi_file_edit"
CapFileCreate           = "file_create"
CapTestExecution        = "test_execution"
CapDatabaseMigration    = "database_migration"
CapDocumentGeneration   = "document_generation"
CapContentCreation      = "content_creation"
CapEnvironmentSetup     = "environment_setup"

// 审查类
CapCodeReview           = "code_review"
CapQualityAssessment    = "quality_assessment"
CapComplianceCheck      = "compliance_check"

// 运维类
CapRollbackPlanning     = "rollback_planning"
CapEnvironmentDiagnosis = "environment_diagnosis"
CapRecoveryExecution    = "recovery_execution"
CapChangeAssessment     = "change_assessment"

// 编排类
CapScheduleOptimization = "schedule_optimization"
CapConflictResolution   = "conflict_resolution"
CapProgressTracking     = "progress_tracking"
```

---

### 5.6 L2 执行层（Execution）

**唯一持有 Executor 引用的层。** 其他层不可能直接执行。

```go
type ExecutorDescriptor struct {
    Name              string
    Version           string
    Capabilities      []string
    CapScores         map[string]float64
    CostPerToken      float64
    AvgLatencyMs      int64
    MaxConcurrent     int
    RequiresSandbox   bool
    SupportsStreaming  bool
    HealthStatus      string // healthy / degraded / offline
}
```

**健康状态机：**
- 最近 10 次成功率 < 30% → `degraded`（降权但不下线）
- 最近 10 次成功率 < 10% → `offline`（不再分配新任务）
- 恢复：连续 3 次成功 → `healthy`

**与现有代码的关系：** `executor.Registry` 保持不变，补充 HealthStatus 字段。6 个执行器不变。

---

### 5.7 L1 证据与治理层（Evidence & Governance）

**职责：** 证明发生了什么、为什么这样推进、谁批准过、如何复盘。

#### ExecutionEvidence

```go
type ExecutionEvidence struct {
    TaskID         int64
    ContractID     int64
    MatchLogID     int64
    ExecutorUsed   string
    ModelUsed      string

    // 执行证据
    StartedAt      time.Time
    CompletedAt    time.Time
    TokensConsumed int64
    FilesModified  []string
    OutputSummary  string

    // 质量证据
    OutcomeVerifyResult string // pass / fail / skip
    AcceptanceResult    string // pass / fail / manual_review

    // 异常证据
    ErrorMessage   string
    FailurePattern string
    RecoveryAction string
}
```

#### 角色的新定位

角色从"调度主键"降级为"治理视图"：

| 用途 | 说明 |
|------|------|
| 权限边界 | 角色决定能看到什么、能操作什么 |
| 审计视角 | "这个动作是以什么角色身份执行的" |
| 成本归集 | 按角色统计 Token 消耗 |
| UI 展示 | 不同颜色/图标 |
| 人工接管 | "我要以架构师身份介入" |

**调度主键变化：**
```go
// 现在（L3.5）：角色 → 执行器
executionMode := projectRole[roleType]["execution_mode"]

// 未来（L4+）：契约 → 匹配 → 执行器
matchResult := ContractMatcher.Match(contract)
executionMode := matchResult.Profile.ExecutionMode
roleType := matchResult.Profile.RoleType // 仅治理视图用
```

#### 五个稳定角色

| 角色 | 职责边界 | 绝对禁区 |
|------|----------|----------|
| **Architect** | 需求理解、方案设计、任务拆分 | 不执行代码修改 |
| **Implementer** | 按规格执行代码/内容创作 | 不做需求判断 |
| **Auditor** | 独立质量评估 | 不修改代码、不做业务决策 |
| **Operator** | 故障恢复、变更风险评估、环境管理 | 不做业务设计 |
| **Coordinator** | 调度编排、冲突裁决、进度控制 | 不做内容生产、不做质量判断 |

---

## 6. 四个闭环

### 6.1 执行闭环

```
Intent → Contract → AdmissionControl → Planner → Match → Execute → Evidence → OutcomeVerify
```

```go
func ExecuteTask(ctx, taskID) {
    bundle := contractStore.GetBundle(ctx, taskID)
    situation := sensor.Perceive(ctx, workflowRunID)

    admission := admissionControl.Check(ctx, bundle.Intent, currentObjective)
    if !admission.Allowed {
        evidenceStore.RecordDenial(ctx, taskID, admission)
        return
    }

    plan := planner.Evaluate(ctx, situation, bundle)
    plan.Meta.Validate(plan.DecisionLevel)

    match := matcher.Match(ctx, bundle.Task, situation)
    result := executorRegistry.Execute(ctx, match.Profile, bundle)

    evidence := collectEvidence(result, bundle)
    evidenceStore.Record(ctx, evidence)

    verifyOutcome(evidence, bundle.Outcome)
}
```

### 6.2 验收闭环

```
Outcome → Evidence → EvidenceSufficiency → Acceptance → TransitionState
```

```go
func AcceptTask(ctx, taskID, evidence) {
    bundle := contractStore.GetBundle(ctx, taskID)

    sufficiency := computeEvidenceSufficiency(evidence, bundle.Acceptance.EvidenceRequired)
    if sufficiency < EvidenceThresholdAssist {
        requestAdditionalEvidence(ctx, taskID)
        return
    }

    acceptResult := runAcceptance(ctx, evidence, bundle.Acceptance)
    TransitionState(ctx, taskID, acceptResult.TargetStatus, evidence)
    observer.Record(ctx, &ObservationRecord{DecisionType: "acceptance", Output: acceptResult})
}
```

### 6.3 恢复闭环

```
Failure → Sensor.Perceive → EscalateContract → Operator.Diagnose → Planner.Recovery → Re-entry
```

```go
func RecoverFromFailure(ctx, taskID, failureEvidence) {
    situation := sensor.Perceive(ctx, workflowRunID)
    bundle := contractStore.GetBundle(ctx, taskID)
    EscalateContract(bundle, failureEvidence.FailurePattern)
    contractStore.SaveBundle(ctx, bundle)

    diagnosis := operatorDiagnose(ctx, taskID, failureEvidence, situation)
    recoveryPlan := planner.EvaluateRecovery(ctx, situation, diagnosis)

    validation := recoveryPlan.Meta.Validate(recoveryPlan.DecisionLevel)
    if validation.UpgradeTo == "C" {
        createHumanCheckpoint(ctx, recoveryPlan)
        return
    }

    reEntryToExecution(ctx, taskID, bundle)
}
```

### 6.4 学习闭环

```
MatchLog → Outcome → EMA Update → Assessor → Tuner → (conservative=auto / aggressive=human)
```

```go
func LearnFromOutcome(ctx, matchLogID, outcome) {
    matchLog := matchLogRepo.Get(ctx, matchLogID)
    record := learningRepo.GetOrCreate(ctx, matchLog.ScopeKey, "success_rate")

    UpdateEMA(record, outcomeToFloat(outcome), 0.15)

    // 置信度衰减
    if time.Since(record.LastUpdatedAt) > 7*24*time.Hour {
        record.Confidence *= 0.9
    }
    learningRepo.Save(ctx, record)

    if record.SampleCount >= 10 && shouldAssess(ctx) {
        assessment := assessor.Evaluate(ctx)
        for _, drift := range assessment.Drifts {
            rec := tuner.Recommend(ctx, drift)
            if rec.ChangeRatio() > 0.2 {
                rec.SuggestedValue = clampChange(rec.CurrentValue, rec.SuggestedValue, 0.2)
            }
            if rec.Direction == "conservative" && rec.Confidence > 0.7 {
                rec.AutoApplicable = true
            } else {
                rec.AutoApplicable = false
            }
            tuneRepo.Save(ctx, rec)
        }
    }
}
```

---

## 7. Build Line + Operate Line 双主线

```
              ┌──────────────┐
              │  L6 目标层    │
              └──────┬───────┘
          ┌──────────┼──────────┐
          ↓                     ↓
┌─────────────────┐   ┌──────────────────┐
│   Build Line    │   │   Operate Line   │
│   研发主线       │   │   运维主线        │
│                 │   │                  │
│ design → review │   │ sense → diagnose │
│ → execute       │   │ → recover        │
│ → accept        │   │ → rollback       │
│ → complete      │   │ → report         │
│                 │   │                  │
│ Architect       │   │ Operator         │
│ Implementer     │   │ Coordinator      │
│ Auditor         │   │                  │
└────────┬────────┘   └────────┬─────────┘
         └─────────┬───────────┘
                   ↓
          ┌────────────────┐
          │ L3 能力层（共享）│
          └────────────────┘
```

**协作规则：**
1. Build Line 正常运行时，Operate Line 在后台巡检（Sensor.Patrol）
2. Build Line 出异常 → Operate Line 接管（Operator 诊断 + Coordinator 调度调整）
3. 恢复成功 → Build Line 继续
4. 恢复失败 → 升级人工（C 级）
5. Build Line 完成 → Operate Line 生成报告归档指标

---

## 8. LLM 边界铁律

### 8.1 LLM 绝对不能做的事

| 决策类型 | 规则引擎 | LLM | 人工 |
|----------|---------|-----|------|
| 状态迁移 | **唯一决策者** | ✗ | ✗ |
| 权限校验 | **唯一决策者** | ✗ | ✗ |
| 风险闸门求值 | **唯一决策者** | ✗ | 命中时确认 |
| 资源锁定/释放 | **唯一决策者** | ✗ | ✗ |
| CAS 状态转换 | **唯一决策者** | ✗ | ✗ |
| 成本阈值判断 | **唯一决策者** | ✗ | 超限审批 |
| 熔断触发条件 | **唯一决策者** | ✗ | ✗ |

### 8.2 LLM 可参与但受约束的决策

| 决策类型 | 规则引擎 | LLM | 人工 |
|----------|---------|-----|------|
| 故障诊断 | 分类规则 | **分析诊断** | C 级审批 |
| 质量评估 | 规则基线 | **辅助 Judge** | manual_review |
| 重规划 | 触发条件 | **方案生成** | B/C 级审批 |
| 执行器推荐 | 契约匹配 | 可选精排 | ✗ |
| 任务拆分 | 格式校验 | **核心生成** | 确认方案 |
| 调度优化 | 冲突规则 | **优化建议** | 冲突裁决 |

---

## 9. 动作类型总表

### 现有动作（7 个，不变）

| 动作 | 说明 | 默认级别 |
|------|------|----------|
| retry_task | 重试失败任务 | A |
| trigger_rework | 触发返工 | B |
| rerun_accept | 重新验收 | A |
| pause_workflow | 暂停工作流 | A |
| approve_complete | 批准完成 | A |
| notify_human | 通知人工 | B |
| replan_workflow | 触发重规划 | B |

### 新增动作（5 个）

| 动作 | 说明 | 默认级别 |
|------|------|----------|
| switch_executor | 切换执行器 | A(同级)/B(升级) |
| adjust_concurrency | 调整并发度 | A(降低)/B(提高) |
| split_batch | 拆分大批次 | B |
| reschedule_task | 延迟重调度 | A |
| adjust_threshold | 调整闸门阈值 | B(微调)/C(大幅) |

---

## 10. 数据模型变更

### 10.1 新建表（7 个）

```sql
-- ① 执行器运行统计
CREATE TABLE IF NOT EXISTS `mvp_engine_stats` (
  `id`              bigint unsigned NOT NULL,
  `project_family`  varchar(32) NOT NULL,
  `category_code`   varchar(64) DEFAULT NULL,
  `role_type`       varchar(32) NOT NULL,
  `engine_type`     varchar(32) NOT NULL,
  `total_runs`      int NOT NULL DEFAULT 0,
  `success_runs`    int NOT NULL DEFAULT 0,
  `failure_runs`    int NOT NULL DEFAULT 0,
  `avg_duration_ms` bigint NOT NULL DEFAULT 0,
  `avg_tokens`      bigint NOT NULL DEFAULT 0,
  `success_rate`    decimal(5,4) NOT NULL DEFAULT 0,
  `last_run_at`     datetime DEFAULT NULL,
  `created_at`      datetime DEFAULT NULL,
  `updated_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_scope_engine` (`project_family`,`category_code`,`role_type`,`engine_type`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='执行器运行统计';

-- ② 自治学习记录
CREATE TABLE IF NOT EXISTS `mvp_autonomy_learning` (
  `id`              bigint unsigned NOT NULL,
  `learning_type`   varchar(32) NOT NULL,
  `scope_key`       varchar(128) NOT NULL,
  `dimension`       varchar(64) NOT NULL,
  `parameters`      json DEFAULT NULL,
  `sample_count`    int NOT NULL DEFAULT 0,
  `confidence`      decimal(5,4) NOT NULL DEFAULT 0,
  `last_updated_at` datetime DEFAULT NULL,
  `created_at`      datetime DEFAULT NULL,
  `updated_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_scope_dimension` (`scope_key`,`dimension`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='自治学习记录';

-- ③ 态势快照
CREATE TABLE IF NOT EXISTS `mvp_situation_snapshot` (
  `id`              bigint unsigned NOT NULL,
  `workflow_run_id` bigint unsigned NOT NULL,
  `project_id`      bigint unsigned NOT NULL,
  `snapshot_data`   json NOT NULL,
  `anomaly_signals` json DEFAULT NULL,
  `created_by`      bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`         bigint unsigned NOT NULL DEFAULT 0,
  `created_at`      datetime DEFAULT NULL,
  `deleted_at`      datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workflow_run` (`workflow_run_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='态势快照';

-- ④ 动作效果评估
CREATE TABLE IF NOT EXISTS `mvp_action_outcome` (
  `id`                  bigint unsigned NOT NULL,
  `decision_action_id`  bigint unsigned NOT NULL,
  `before_snapshot_id`  bigint unsigned DEFAULT NULL,
  `after_snapshot_id`   bigint unsigned DEFAULT NULL,
  `effectiveness_score` decimal(5,4) DEFAULT NULL,
  `failure_rate_delta`  decimal(5,4) DEFAULT NULL,
  `duration_delta_ms`   bigint DEFAULT NULL,
  `throughput_delta`    decimal(5,4) DEFAULT NULL,
  `evaluated_at`        datetime DEFAULT NULL,
  `created_by`          bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`             bigint unsigned NOT NULL DEFAULT 0,
  `created_at`          datetime DEFAULT NULL,
  `deleted_at`          datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_action` (`decision_action_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='动作效果评估';

-- ⑤ 任务契约
CREATE TABLE IF NOT EXISTS `mvp_task_contract` (
  `id`                      bigint unsigned NOT NULL,
  `domain_task_id`          bigint unsigned NOT NULL,
  `blueprint_id`            bigint unsigned DEFAULT NULL,
  `intent_contract`         json DEFAULT NULL COMMENT '意图契约JSON',
  `required_capabilities`   json DEFAULT NULL,
  `preferred_capabilities`  json DEFAULT NULL,
  `accuracy_requirement`    varchar(16) DEFAULT 'medium',
  `creativity_requirement`  varchar(16) DEFAULT 'medium',
  `context_scope`           varchar(16) DEFAULT 'local',
  `requires_file_access`    tinyint DEFAULT 0,
  `requires_sandbox`        tinyint DEFAULT 0,
  `requires_isolation`      tinyint DEFAULT 0,
  `requires_human_review`   tinyint DEFAULT 0,
  `max_token_budget`        bigint DEFAULT 0,
  `max_duration_seconds`    int DEFAULT 0,
  `cost_sensitivity`        varchar(16) DEFAULT 'medium',
  `sensitivity_level`       varchar(16) DEFAULT 'internal',
  `outcome_contract`        json DEFAULT NULL COMMENT '结果契约JSON',
  `acceptance_contract`     json DEFAULT NULL COMMENT '验收契约JSON',
  `effect_contract`         json DEFAULT NULL COMMENT '副作用契约JSON',
  `domain_tags`             json DEFAULT NULL,
  `source`                  varchar(16) DEFAULT 'inferred',
  `version`                 int NOT NULL DEFAULT 1,
  `created_by`              bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`                 bigint unsigned NOT NULL DEFAULT 0,
  `created_at`              datetime DEFAULT NULL,
  `updated_at`              datetime DEFAULT NULL,
  `deleted_at`              datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_domain_task` (`domain_task_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='任务契约（五契约包）';

-- ⑥ 能力匹配日志
CREATE TABLE IF NOT EXISTS `mvp_capability_match_log` (
  `id`                bigint unsigned NOT NULL,
  `domain_task_id`    bigint unsigned NOT NULL,
  `contract_id`       bigint unsigned NOT NULL,
  `contract_version`  int NOT NULL DEFAULT 1,
  `matched_role_type` varchar(32) NOT NULL,
  `matched_engine`    varchar(32) NOT NULL,
  `matched_model_id`  bigint unsigned DEFAULT NULL,
  `match_score`       decimal(5,4) NOT NULL,
  `capability_score`  decimal(5,4) DEFAULT NULL,
  `quality_score`     decimal(5,4) DEFAULT NULL,
  `cost_score`        decimal(5,4) DEFAULT NULL,
  `history_score`     decimal(5,4) DEFAULT NULL,
  `decision_meta`     json DEFAULT NULL COMMENT '四个一等公民JSON',
  `reasoning`         text DEFAULT NULL,
  `outcome`           varchar(16) DEFAULT NULL,
  `created_by`        bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`           bigint unsigned NOT NULL DEFAULT 0,
  `created_at`        datetime DEFAULT NULL,
  `deleted_at`        datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_task` (`domain_task_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='能力匹配日志';

-- ⑦ 能力合同
CREATE TABLE IF NOT EXISTS `mvp_capability_contract` (
  `id`               bigint unsigned NOT NULL,
  `capability_code`  varchar(64) NOT NULL,
  `capability_name`  varchar(128) NOT NULL,
  `version`          int NOT NULL DEFAULT 1,
  `input_schema`     json DEFAULT NULL,
  `output_schema`    json DEFAULT NULL,
  `output_verify`    varchar(16) DEFAULT 'rule',
  `required_tools`   json DEFAULT NULL,
  `required_context` varchar(16) DEFAULT 'local',
  `required_perms`   json DEFAULT NULL,
  `side_effects`     json DEFAULT NULL,
  `reversible`       tinyint DEFAULT 1,
  `auto_level`       char(1) DEFAULT 'B',
  `cost_level`       varchar(16) DEFAULT 'medium',
  `max_duration`     int DEFAULT 600,
  `max_tokens`       bigint DEFAULT 0,
  `degradation_path` json DEFAULT NULL,
  `project_family`   varchar(32) DEFAULT NULL,
  `enabled`          tinyint DEFAULT 1,
  `created_by`       bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`          bigint unsigned NOT NULL DEFAULT 0,
  `created_at`       datetime DEFAULT NULL,
  `updated_at`       datetime DEFAULT NULL,
  `deleted_at`       datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code_version` (`capability_code`,`version`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='能力合同';
```

### 10.2 修改现有表（6 个）

```sql
-- mvp_decision_action +4 字段
ALTER TABLE `mvp_decision_action`
  ADD COLUMN `plan_id` varchar(64) DEFAULT NULL AFTER `action_status`,
  ADD COLUMN `situation_snapshot_id` bigint unsigned DEFAULT NULL AFTER `plan_id`,
  ADD COLUMN `confidence` decimal(5,4) DEFAULT NULL AFTER `situation_snapshot_id`,
  ADD COLUMN `reasoning` text DEFAULT NULL AFTER `confidence`;

-- mvp_domain_task +4 字段
ALTER TABLE `mvp_domain_task`
  ADD COLUMN `recommended_engine` varchar(32) DEFAULT NULL AFTER `execution_mode`,
  ADD COLUMN `engine_switch_count` int NOT NULL DEFAULT 0 AFTER `recommended_engine`,
  ADD COLUMN `contract_id` bigint unsigned DEFAULT NULL AFTER `engine_switch_count`,
  ADD COLUMN `match_score` decimal(5,4) DEFAULT NULL AFTER `contract_id`;

-- mvp_policy_rule +1 字段
ALTER TABLE `mvp_policy_rule`
  ADD COLUMN `strategy_override` tinyint NOT NULL DEFAULT 0 AFTER `enabled`;

-- mvp_role_preset +1 字段
ALTER TABLE `mvp_role_preset`
  ADD COLUMN `capability_tags` json DEFAULT NULL AFTER `system_prompt`;

-- mvp_project_role +1 字段
ALTER TABLE `mvp_project_role`
  ADD COLUMN `capability_tags` json DEFAULT NULL AFTER `system_prompt`;

-- mvp_task_blueprint +1 字段
ALTER TABLE `mvp_task_blueprint`
  ADD COLUMN `contract_json` json DEFAULT NULL AFTER `description`;
```

### 10.3 新增配置项

```sql
INSERT INTO `mvp_config` (`config_key`,`config_value`,`config_type`,`category`,`description`,`created_at`,`updated_at`) VALUES
-- Sensor
('workflow.autonomy.patrol_interval','60','int','autonomy','态势巡检间隔(秒)',NOW(),NOW()),
('workflow.autonomy.patrol_enabled','0','int','autonomy','定时巡检开关',NOW(),NOW()),
-- Planner
('workflow.autonomy.strategy_enabled','0','int','autonomy','策略函数总开关',NOW(),NOW()),
('workflow.autonomy.engine_selection_enabled','0','int','autonomy','执行器动态选择',NOW(),NOW()),
('workflow.autonomy.batch_adjust_enabled','0','int','autonomy','批次动态调整',NOW(),NOW()),
('workflow.autonomy.adaptive_retry_enabled','0','int','autonomy','自适应重试',NOW(),NOW()),
('workflow.autonomy.proactive_replan_enabled','0','int','autonomy','主动重规划',NOW(),NOW()),
('workflow.autonomy.cost_guard_enabled','0','int','autonomy','成本守卫',NOW(),NOW()),
('workflow.autonomy.quality_gate_enabled','0','int','autonomy','质量门动态调整',NOW(),NOW()),
-- Learner
('workflow.autonomy.learner_enabled','0','int','autonomy','学习器开关',NOW(),NOW()),
('workflow.autonomy.learning_rate','0.15','string','autonomy','EMA学习率',NOW(),NOW()),
('workflow.autonomy.min_sample_count','10','int','autonomy','最小样本量',NOW(),NOW()),
-- 契约匹配
('workflow.autonomy.contract_matching_enabled','0','int','autonomy','能力匹配开关(仅日志)',NOW(),NOW()),
('workflow.autonomy.contract_matching_active','0','int','autonomy','匹配接入执行流程',NOW(),NOW()),
-- 元认知
('workflow.autonomy.meta_cognition_enabled','0','int','autonomy','元认知观测开关',NOW(),NOW()),
('workflow.autonomy.meta_auto_tune_enabled','0','int','autonomy','自动校准(保守方向)',NOW(),NOW()),
-- 目标层
('workflow.autonomy.objective_enabled','0','int','autonomy','目标层准入控制',NOW(),NOW()),
-- 运维线
('workflow.autonomy.operate_line_enabled','0','int','autonomy','运维主线开关',NOW(),NOW()),
-- 执行器
('workflow.autonomy.engine_escalation_chain','chat,aider,claude_code,openhands','string','autonomy','执行器升级链',NOW(),NOW())
ON DUPLICATE KEY UPDATE `updated_at`=NOW();
```

---

## 11. 文件清单

### 11.1 新建文件（按实施阶段分组）

#### Phase A（地基：目标层 + 感知 + Operator）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/situation.go` | 态势模型定义 | 150 |
| `autonomy/sensor.go` | 态势感知 + 异常检测 + 巡检 | 350 |
| `autonomy/objective.go` | 目标模型 + 准入控制 | 200 |
| `autonomy/decision_meta.go` | DecisionMeta + Validate + BlastRadius | 120 |
| `docker/mysql/upgrade/20260410_l4_phase_a.sql` | DDL: situation_snapshot + 目标层配置 | 80 |

#### Phase B（策略函数）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/planner.go` | 规划器编排 | 250 |
| `autonomy/strategy.go` | Strategy 接口 | 80 |
| `autonomy/strategy_adaptive_retry.go` | 自适应重试 | 200 |
| `autonomy/strategy_engine_selection.go` | 执行器选择 | 200 |
| `autonomy/strategy_batch_adjust.go` | 批次调整 | 180 |
| `autonomy/strategy_cost_guard.go` | 成本守卫 | 120 |
| `autonomy/strategy_proactive_replan.go` | 主动重规划 | 150 |
| `autonomy/strategy_quality_gate.go` | 质量门动态调整 | 150 |
| `autonomy/actuator.go` | 动作执行 + 效果跟踪 | 200 |

#### Phase C（契约 + 能力）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/contract.go` | 五契约包 + 推断 + 升级 | 400 |
| `autonomy/capability_profile.go` | 画像模型 + 三层认证 | 200 |
| `autonomy/contract_matcher.go` | 匹配器 + 评分 + Stage-Scoped Routing | 350 |
| `autonomy/profile_registry.go` | 画像注册表 | 150 |
| `repo/task_contract_repo.go` | 契约仓储 | 80 |
| `repo/capability_match_log_repo.go` | 匹配日志仓储 | 80 |
| `docker/mysql/upgrade/20260410_l4_phase_c.sql` | DDL: task_contract + match_log + capability_contract | 120 |

#### Phase D（学习 + 元认知）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/learner.go` | EMA 学习 + 三重保护 | 300 |
| `autonomy/meta_observer.go` | 观测器 | 200 |
| `autonomy/meta_assessor.go` | 评估器 | 250 |
| `autonomy/meta_tuner.go` | 校准器 | 200 |
| `repo/engine_stats_repo.go` | 执行器统计仓储 | 100 |
| `repo/learning_repo.go` | 学习记录仓储 | 80 |
| `repo/action_outcome_repo.go` | 效果评估仓储 | 80 |
| `repo/situation_snapshot_repo.go` | 态势快照仓储 | 80 |
| `docker/mysql/upgrade/20260410_l4_phase_d.sql` | DDL: engine_stats + learning + action_outcome | 100 |

### 11.2 修改文件

| 文件 | 改动 |
|------|------|
| `autonomy/decision_center.go` | Decide() 升级为 L6→L5→L4→L3→L2 编排链 |
| `autonomy/action_dispatcher.go` | 注册 5 个新动作回调 |
| `autonomy/model.go` | 新增能力词表常量 + 动作类型 + Operator 角色 |
| `orchestrator/registry.go` | Init() 初始化七层组件 |
| `domain/task/task_service.go` | 实例化时生成契约 + 调用匹配器 |
| `stage/execute/service.go` | 查询匹配器推荐执行器 |
| `consts/task.go` | 新增 RoleTypeOperator |

### 11.3 不修改的文件

| 文件 | 原因 |
|------|------|
| `scheduler/domain_task_scheduler.go` | 并发度通过配置动态读取 |
| `watchdog/watchdog.go` | 已通过 SetRetryFn/SetEscalateFn 接入 |
| `stage/accept/service.go` | 已通过 CompleteTrigger/ReworkTrigger 接入 |
| `stage/rework/service.go` | 已通过 AcceptTrigger 接入 |
| `collab/` | 飞书通知通过事件总线订阅 |
| `engine/` | 旧引擎继续为 V1 项目服务 |

---

## 12. 实施路线图

### Phase A：地基（目标层 + 感知 + Operator）

**目标：** 项目有预算/时限约束，系统能持续感知运行态势，故障有专业角色。

| 序号 | 任务 | 说明 |
|------|------|------|
| A1 | Situation 模型 + Sensor 实现 | 态势采集，不影响业务 |
| A2 | DecisionMeta + Validate | 四个一等公民结构体 |
| A3 | ProjectObjective 模型 + AdmissionControl | 准入控制接入 DecisionCenter |
| A4 | Operator 角色上线 | consts + 预设 + system_prompt |
| A5 | DDL + 配置项 | situation_snapshot 表 + 配置 |

**灰度：** `objective_enabled=0` 时准入控制不生效。Sensor 纯采集不影响决策。

### Phase B：策略函数（Planner + 6 个策略）

**目标：** 系统能根据态势智能选择动作。

| 序号 | 任务 | 说明 |
|------|------|------|
| B1 | Strategy 接口 + Planner 编排 | 策略框架 |
| B2 | AdaptiveRetryStrategy | 替换固定重试 |
| B3 | EngineSelectionStrategy | 动态执行器选择 |
| B4 | BatchAdjustStrategy | 并发度动态调整 |
| B5 | CostGuardStrategy | 成本守卫 |
| B6 | ProactiveReplanStrategy | 主动重规划 |
| B7 | QualityGateStrategy | 质量门动态调整 |
| B8 | Actuator + 效果跟踪 | 延迟效果评估 |
| B9 | 修改 decision_center.go | 接入 Planner |

**灰度：** `strategy_enabled=0` 时走原 PolicyEngine 路径。每个策略独立开关。

### Phase C：契约 + 能力匹配

**目标：** 从"角色驱动"升级到"契约驱动"。

| 序号 | 任务 | 说明 |
|------|------|------|
| C1 | 五契约包模型 | IntentContract + TaskContract + OutcomeContract + AcceptanceContract + EffectContract |
| C2 | InferContract + EscalateContract | 自动推断 + 失败升级 |
| C3 | CapabilityProfile（三层认证） | 声明+认证+观测 |
| C4 | ProfileRegistry | 从 project_role 加载画像 |
| C5 | ContractMatcher | 评分匹配（先仅记日志） |
| C6 | Stage-Scoped Routing | 批次级主执行链 |

**灰度：** `contract_matching_enabled=1` 先 shadow mode 记日志。`contract_matching_active=1` 再接入执行流程。

### Phase D：学习 + 元认知

**目标：** 系统能从历史中学习，自动校准保守方向参数。

| 序号 | 任务 | 说明 |
|------|------|------|
| D1 | Learner（EMA + 三重保护） | 学习闭环 |
| D2 | Observer | 决策观测记录 |
| D3 | Assessor | 定期评估 |
| D4 | Tuner（保守可自动，激进必须人工） | 参数校准 |

**灰度：** `learner_enabled=0` 时不学习。`meta_cognition_enabled=0` 时不观测。

### Phase E：全面灰度开放

按 §13 灰度策略逐级打开。

---

## 13. 灰度策略（10 级）

| 级别 | 开关 | 行为 | 可独立回滚 |
|------|------|------|-----------|
| 0 | 全关 | 与 L3.5 完全一致 | — |
| 1 | patrol_enabled=1 | Sensor 采集态势，不影响决策 | ✅ |
| 2 | objective_enabled=1 | AdmissionControl 生效 | ✅ |
| 3 | strategy_enabled=1 | 策略函数参与（各自独立开关） | ✅ |
| 4 | contract_matching_enabled=1 | 契约匹配仅记日志 | ✅ |
| 5 | contract_matching_active=1 | 匹配接入执行流程 | ✅ |
| 6 | learner_enabled=1 | EMA 学习开始 | ✅ |
| 7 | meta_cognition_enabled=1 | Observer 记录 | ✅ |
| 8 | meta_auto_tune_enabled=1 | Tuner 自动应用保守建议 | ✅ |
| 9 | autonomy_level=autonomous | A/B 自动，仅 C 人工 | ✅ |

---

## 14. 核心调用链（全景）

```
事件触发（task.failed / accept.passed / ...）
  │
  ├─→ L5 Sensor.Perceive() → Situation
  │
  ├─→ L6 AdmissionControl.Check(situation, objective)
  │   └─→ Denied → L1 EvidenceStore.RecordDenial() → 结束
  │
  ├─→ L5 PolicyEngine.Match(trigger) + Planner.Evaluate(situation)
  │   └─→ ActionPlan（含 DecisionMeta）
  │
  ├─→ DecisionMeta.Validate()
  │   └─→ 影响面/置信度/证据不足 → 升级决策级别
  │
  ├─→ L5 RiskGate.Check()
  │   └─→ 命中 → 降级/人工
  │
  ├─→ L4 ContractBundle 加载/推断
  │
  ├─→ L3 ContractMatcher.Match() → MatchResult
  │   └─→ Stage-Scoped Routing 检查
  │
  ├─→ L2 Executor.Execute() → ExecutionResult
  │
  ├─→ L1 EvidenceStore.Record(evidence)
  │
  ├─→ L4 OutcomeContract.Verify(evidence)
  │   ├─→ pass → AcceptanceContract → TransitionState(evidence)
  │   └─→ fail → EscalateContract → RecoverFromFailure
  │
  └─→ L7 Observer.Record() → Assessor → Tuner
```
