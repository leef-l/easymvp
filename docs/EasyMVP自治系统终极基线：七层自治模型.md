# EasyMVP 自治系统终极基线：七层自治模型

> 合并自：L4 自治内核 + 契约驱动能力编排 + 六层自治架构 + 外部评审反馈
> 本文档是实施基线，所有后续迭代以此为准。

---

## 0. 文档定位

本文档融合了以下三份设计方案和两轮外部评审反馈，形成**唯一实施基线**：

| 来源 | 贡献 |
|------|------|
| L4 自治内核方案 | Sensor→Planner→Actuator→Learner 四组件、态势模型、策略函数 |
| 契约驱动能力编排方案 | TaskContract、CapabilityProfile、ContractMatcher、能力词表 |
| 六层自治架构方案 | 元认知层、目标层、Build/Operate 双主线、能力合同化 |
| 外部评审第一轮 | 角色降级为治理视图、能力图谱化、Operator 一等公民、LLM 硬边界 |
| 外部评审第二轮 | OutcomeContract、能力认证三层、stage-scoped routing、副作用合同、Evidence 层 |

**前三份设计文档在本文档发布后归档，不再作为实施依据。**

---

## 1. 七层架构总览

```
┌─────────────────────────────────────────────────────────────┐
│ L7  元认知层 Meta-Cognition                                  │
│     Observer · Assessor · Tuner                              │
│     "系统做得怎么样？参数对不对？该怎么改进？"                    │
├─────────────────────────────────────────────────────────────┤
│ L6  目标层 Objective                                         │
│     目标 · 预算 · 时限 · 风险容忍度 · SLA · 自治边界           │
│     "完成什么？不能碰什么？花多少？什么时候停？"                  │
├─────────────────────────────────────────────────────────────┤
│ L5  策略层 Policy                                            │
│     准入控制 · 风险闸门 · 决策分级 · 人工节点                   │
│     "允不允许做？自动还是人工？做完怎么验？失败怎么收？"           │
├─────────────────────────────────────────────────────────────┤
│ L4  契约层 Contract                                          │
│     TaskContract · OutcomeContract · AcceptanceContract       │
│     "任务需要什么？成功产出什么？怎么证明合格？"                  │
├─────────────────────────────────────────────────────────────┤
│ L3  能力层 Capability                                        │
│     画像(声明+认证+观测) · 匹配器 · 降级路径                   │
│     "谁能满足契约？代价多大？失败了怎么降级？"                   │
├─────────────────────────────────────────────────────────────┤
│ L2  执行层 Execution                                         │
│     执行器市场 · 模型 · 工具 · 沙箱 · 资源                    │
│     "用什么工具？跑在哪？消耗多少？健康吗？"                     │
├─────────────────────────────────────────────────────────────┤
│ L1  证据与治理层 Evidence & Governance                        │
│     结构化证据 · 角色视图 · 审计 · 人工接管                     │
│     "怎么证明做了？谁能看？谁能管？人怎么介入？"                  │
└─────────────────────────────────────────────────────────────┘
```

**层间规则：**
- 每层只和相邻层交互，不跨层调用
- 上层约束下层，下层向上层反馈
- 每层独立灰度开关，可独立回滚

---

## 2. 各层详细定义

### 2.1 L7 元认知层（Meta-Cognition）

**职责：** 观测系统自身行为，评估效果，校准参数。不做业务决策，做"关于决策的决策"。

**三要素：**

| 组件 | 职责 | 输入 | 输出 |
|------|------|------|------|
| Observer（观测器） | 记录每次决策的输入/输出/结果/人工覆盖 | 决策事件流 | ObservationRecord |
| Assessor（评估器） | 定期评估策略准确率、闸门误报率、匹配精度 | 观测记录集 | AssessmentResult + Drift[] |
| Tuner（校准器） | 生成参数校准建议 | 偏差报告 | TuneRecommendation[] |

**核心数据模型：**

```go
type ObservationRecord struct {
    DecisionID     int64
    DecisionType   string   // policy_match / contract_match / gate_check / admission
    InputSnapshot  g.Map    // 决策输入快照
    Output         g.Map    // 决策输出
    Confidence     float64  // 决策时置信度
    Outcome        string   // success / failure / neutral（延迟回填）
    EffectScore    float64  // -1 ~ 1（延迟回填）
    HumanOverride  bool     // 是否被人工覆盖
    OverrideReason string   // 覆盖原因
}

type AssessmentResult struct {
    PolicyAccuracy     float64 // 策略匹配后动作成功率
    GateFalsePositive  float64 // 闸门误报率
    GateFalseNegative  float64 // 闸门漏报率
    HumanOverrideRate  float64 // 人工覆盖率
    MatchAccuracy      float64 // 能力匹配后任务成功率
    CostEfficiency     float64 // 成本效率
    Drifts             []Drift // 参数偏差
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
    Reasoning      string
    Confidence     float64
    AutoApplicable bool   // 是否可自动应用
    RiskLevel      string // low / medium / high
}
```

**校准安全铁律：**

| 方向 | 规则 |
|------|------|
| 变保守（降并发、收紧闸门、升级决策级别） | ✅ 可自动应用（confidence > 0.7） |
| 变激进（升并发、放宽闸门、降级决策级别） | ❌ 必须人工确认 |
| 目标层参数变更 | ❌ 始终人工确认 |

**学习信号强度：**
- 系统决策 A 级自动执行，人工没干预 → 弱信号
- 系统决策 B 级，人工批准 → 中信号（人认可）
- 系统决策 B 级，人工驳回 → **强信号**（系统判断错了）

---

### 2.2 L6 目标层（Objective）

**职责：** 定义项目的目标和约束，是整个自治系统的方向盘。所有下层决策必须在此边界内运行。

```go
type ProjectObjective struct {
    // 交付目标
    DeliveryGoal       string     // 交付物描述
    QualityFloor       float64    // 最低质量标准 0-1

    // 预算约束
    TokenBudget        int64      // Token 总预算（0=不限）
    TimeBudgetHours    float64    // 时间预算
    CostBudgetCents    int64      // 金钱预算

    // 风险约束
    RiskTolerance      string     // conservative / balanced / aggressive
    MaxAutoRetries     int        // 自动重试上限
    MaxAutoReworks     int        // 自动返工上限
    MaxAutoReplans     int        // 自动重规划上限

    // SLA
    DeadlineAt         *time.Time // 截止时间
    MaxStallMinutes    int        // 最大停滞时间

    // 自治边界
    AutonomyLevel      string     // manual / supervised / autonomous
    // manual: 所有动作人工确认
    // supervised: A 级自动，B/C 级人工（默认）
    // autonomous: A/B 级自动，仅 C 级人工
}
```

**目标如何约束下层（示例）：**

```
RiskTolerance = "conservative"
  → L5 策略层：熔断阈值从 5 降为 3，B 级不自动执行
  → L3 能力层：匹配偏向历史表现好的画像
  → L2 执行层：并发度从 20 降到 10

TokenBudget 消耗 90%
  → L5 策略层：准入控制拒绝新的高成本动作
  → L3 能力层：切换到轻量执行器
```

---

### 2.3 L5 策略层（Policy）

**职责：** 不只是"选谁干"，更重要的是**"允不允许干"**。

**调用链：**

```
动作请求
  → AdmissionControl（准入控制：预算/时限/次数检查）
  → PolicyEngine.Match（规则匹配：决策级别+动作类型）
  → RiskGate.Check（闸门检查：命中则降级）
  → 输出：允许/拒绝/需人工确认
```

```go
type AdmissionResult struct {
    Allowed    bool
    DenyReason string
    Conditions []string // 附加条件
}
```

**准入控制检查项：**

| 检查项 | 条件 | 动作 |
|--------|------|------|
| 预算检查 | TokensConsumed > TokenBudget * 90% | 拒绝高成本动作 |
| 时间检查 | Now > DeadlineAt | 拒绝非必要动作 |
| 重试上限 | RetryCount >= MaxAutoRetries | 拒绝自动重试 |
| 返工上限 | ReworkRounds >= MaxAutoReworks | 拒绝自动返工 |
| 自治级别 | AutonomyLevel = manual | 所有动作需人工确认 |

**与现有代码的关系：**
- `PolicyEngine` — 保持不变，归入 L5
- `RiskGate` — 保持不变，归入 L5
- `DecisionCenter.Decide()` — 升级为 L6→L5→L4→L3→L2 的编排入口
- 新增 `AdmissionControl` — 在 PolicyEngine.Match 之前执行

---

### 2.4 L4 契约层（Contract）

**职责：** 形式化描述"任务需要什么"、"成功产出什么"、"怎么证明合格"。

**三种契约：**

| 契约 | 描述 | 时机 |
|------|------|------|
| TaskContract | 任务对执行者的需求 | 任务创建时 |
| OutcomeContract | 成功后必须产出什么 | 任务创建时 |
| AcceptanceContract | 怎么验证产出合格 | 验收阶段 |

```go
// TaskContract 任务契约 — 描述任务需要什么能力和资源
type TaskContract struct {
    // 能力需求
    RequiredCapabilities  []string   // 必须能力
    PreferredCapabilities []string   // 优选能力

    // 质量需求
    AccuracyRequirement   string     // high / medium / low
    CreativityRequirement string     // high / medium / low

    // 上下文需求
    ContextScope          string     // local / batch / global
    RequiresFileAccess    bool
    RequiresSandbox       bool
    RequiresHumanReview   bool

    // 资源约束
    MaxTokenBudget        int64
    MaxDurationSeconds    int
    CostSensitivity       string     // high / medium / low

    // 安全约束
    RequiresIsolation     bool
    SensitivityLevel      string     // public / internal / confidential

    // 副作用声明
    ReadSet               []string   // 读取的资源列表
    WriteSet              []string   // 写入的资源列表
    SideEffectLevel       string     // none / low / medium / high
    Rollbackable          bool       // 是否可回滚
    ApprovalClass         string     // A / B / C（副作用决定审批级别）

    // 领域标签
    DomainTags            []string
}

// OutcomeContract 结果契约 — 描述成功后必须产出什么
type OutcomeContract struct {
    RequiredOutputs []RequiredOutput // 必须产出列表
    SuccessCriteria []string         // 成功判定条件（规则表达式）
    FailureCriteria []string         // 失败判定条件
}

type RequiredOutput struct {
    Name       string // 产出名称
    Type       string // file / text / json / metrics
    Schema     g.Map  // 产出结构 Schema（可选）
    Mandatory  bool   // 是否必须
}

// AcceptanceContract 验收契约 — 描述怎么证明产出合格
type AcceptanceContract struct {
    VerifyMethod    string   // rule / llm_judge / human / auto
    RuleSet         []string // 规则集代码列表
    QualityFloor    float64  // 最低质量分
    EvidenceRequired []string // 需要的证据类型
}
```

**契约来源优先级：**

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 1 | 架构师显式指定 | 任务 JSON 中的 contract 字段 |
| 2 | 系统自动推断 | InferContract() 从蓝图推断 |
| 3 | 运行时升级 | 失败后 EscalateContract() 升级需求 |

**自动推断规则（InferContract）：**

```go
func InferContract(blueprint *TaskBlueprint) *TaskContract {
    contract := &TaskContract{ContextScope: "local", CostSensitivity: "medium"}

    // 从 role_type 推断
    switch blueprint.RoleType {
    case "implementer":
        contract.RequiredCapabilities = []string{"code_edit"}
        contract.RequiresFileAccess = true
        contract.WriteSet = blueprint.AffectedResources
        contract.SideEffectLevel = "medium"
        contract.Rollbackable = true
    case "auditor":
        contract.RequiredCapabilities = []string{"code_review"}
        contract.SideEffectLevel = "none"
    case "architect":
        contract.RequiredCapabilities = []string{"requirement_analysis"}
        contract.ContextScope = "global"
        contract.SideEffectLevel = "none"
    case "operator":
        contract.RequiredCapabilities = []string{"failure_analysis", "risk_assessment"}
        contract.ContextScope = "global"
    }

    // 从 affected_resources 数量推断
    if len(blueprint.AffectedResources) >= 5 {
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "multi_file_edit")
    }

    // 从描述关键词推断
    desc := strings.ToLower(blueprint.Description)
    if containsAny(desc, "测试", "test") {
        contract.PreferredCapabilities = append(contract.PreferredCapabilities, "test_execution")
        contract.RequiresSandbox = true
    }
    if containsAny(desc, "数据库", "migration", "schema") {
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "database_migration")
    }

    // 从 role_level 推断质量和成本
    switch blueprint.RoleLevel {
    case "max":
        contract.AccuracyRequirement = "high"
        contract.ContextScope = "global"
        contract.CostSensitivity = "low"
    case "lite":
        contract.AccuracyRequirement = "low"
        contract.CostSensitivity = "high"
    default:
        contract.AccuracyRequirement = "medium"
    }

    return contract
}
```

**失败后契约升级（EscalateContract）：**

```go
func EscalateContract(contract *TaskContract, failurePattern string) *TaskContract {
    upgraded := contract.Clone()
    upgraded.Version++

    switch failurePattern {
    case PatternCapabilityMismatch:
        upgraded.AccuracyRequirement = "high"
        upgraded.ContextScope = "global"
        upgraded.CostSensitivity = "low"
    case PatternStructural:
        upgraded.RequiredCapabilities = appendUnique(upgraded.RequiredCapabilities, "multi_file_edit")
        upgraded.RequiresSandbox = true
    case PatternResourceConflict:
        upgraded.RequiresIsolation = true
    }

    return upgraded
}
```

---

### 2.5 L3 能力层（Capability）

**职责：** 管理能力画像，匹配契约与执行者。

#### 能力画像三层认证

| 层级 | 来源 | 更新频率 | 说明 |
|------|------|----------|------|
| Declared（声明能力） | 角色+执行器能力表 | 静态 | "我声称能做什么" |
| Certified（认证能力） | 基准测试 | 低频（手动触发） | "测试证明我能做什么" |
| Observed（观测能力） | 历史执行数据 | 持续 | "实际表现如何" |

```go
type CapabilityProfile struct {
    // 身份
    RoleType      string
    ExecutionMode string
    ModelID       int64
    RoleLevel     string

    // 三层能力
    DeclaredCaps  []string            // 声明能力
    CertifiedCaps []string            // 认证能力（基准测试通过）
    ObservedScores map[string]float64 // 观测能力评分

    // 质量特征
    AccuracyScore   float64
    CreativityScore float64
    SpeedScore      float64

    // 资源特征
    AvgTokenCost    int64
    AvgDurationSec  int
    CostPerTask     float64

    // 上下文特征
    MaxContextScope string
    SupportsFile    bool
    SupportsSandbox bool

    // 历史表现（Observed 层）
    HistoricalSuccessRate float64
    SampleCount           int
}
```

#### 契约匹配器

```go
type ContractMatcher struct {
    profileRegistry *ProfileRegistry
}

type MatchResult struct {
    Profile         *CapabilityProfile
    Eligible        bool    // 必须能力全覆盖
    Score           float64 // 综合评分 0-1
    CapabilityScore float64
    QualityScore    float64
    CostScore       float64
    HistoryScore    float64
    Reasoning       string  // 可读匹配理由
}
```

**评分公式：**

```
Score = w1×CapabilityScore + w2×QualityScore + w3×CostScore + w4×HistoryScore

CapabilityScore = 必须能力覆盖率×0.7 + 优选能力覆盖率×0.3
QualityScore    = 1 - |contract.AccuracyReq - profile.AccuracyScore|
CostScore       = (1 - profile.CostPerTask / maxCost)
HistoryScore    = profile.SuccessRate × min(1, profile.SampleCount / 30)

权重按 CostSensitivity 动态调整：
  省钱优先：w1=0.30 w2=0.20 w3=0.35 w4=0.15
  效果优先：w1=0.30 w2=0.35 w3=0.10 w4=0.25
  均衡：    w1=0.30 w2=0.25 w3=0.20 w4=0.25
```

#### Stage-Scoped Routing（阶段级路由）

**防止逐任务自由匹配导致运行抖动：**

```
原则：先在阶段/批次级选定主执行链，再允许局部任务重路由。

流程：
  1. 批次启动时，对该批次所有任务的契约做聚合分析
  2. 选定批次主执行链（role_type + execution_mode + model）
  3. 个别任务契约与主链不兼容时，才做局部重路由
  4. 记录重路由原因，供元认知层评估
```

#### 能力合同（CapabilityContract）

```go
type CapabilityContract struct {
    CapabilityCode  string
    CapabilityName  string

    // 输入输出契约
    InputSchema     g.Map
    OutputSchema    g.Map
    OutputVerify    string  // rule / llm_judge / human / auto
    RequiredTools   []string
    RequiredContext string
    RequiredPerms   []string

    // 副作用声明
    SideEffects     []SideEffect
    Reversible      bool

    // 运行约束
    MaxDuration     int
    MaxTokens       int64
    CostLevel       string  // low / medium / high / very_high
    AutoLevel       string  // A / B / C

    // 降级路径
    DegradationPath []DegradationStep
}

type SideEffect struct {
    Type     string // file_modify / db_change / api_call / resource_lock
    Scope    string
    Severity string // low / medium / high
}

type DegradationStep struct {
    Condition      string // timeout / failure / cost_exceeded
    FallbackTo     string // 降级到哪个能力
    FallbackParams g.Map
}
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

### 2.6 L2 执行层（Execution）

**职责：** 执行器市场化管理。

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

**执行器健康检查：**
- 最近 10 次成功率 < 30% → degraded
- 最近 10 次成功率 < 10% → offline（不再分配新任务）
- 恢复：连续 3 次成功 → healthy

**与现有代码的关系：**
- `executor.Registry` — 保持不变，补充 HealthStatus 字段
- 6 个执行器（chat/aider/claude_code/openhands/codex_cli/gemini_cli）不变

---

### 2.7 L1 证据与治理层（Evidence & Governance）

**职责：** 记录结构化证据、提供角色视图、支持审计和人工接管。

#### 角色的新定位

角色从"调度主键"降级为"治理视图"：

| 用途 | 说明 |
|------|------|
| 权限边界 | 角色决定能看到什么、能操作什么 |
| 审计视角 | "这个动作是以什么角色身份执行的" |
| 成本归集 | "架构师消耗多少 Token，实施员消耗多少" |
| UI 展示 | 不同颜色/图标展示不同角色的任务 |
| 人工接管 | "我要以架构师身份介入这个任务" |

**调度主键从 `role_type` 变为 `contract`：**

```go
// 现在（L3.5）
executionMode := projectRole[roleType]["execution_mode"]  // 角色 → 执行器

// 未来（L4+）
matchResult := ContractMatcher.Match(contract)             // 契约 → 最优匹配
executionMode := matchResult.Profile.ExecutionMode          // 匹配 → 执行器
roleType := matchResult.Profile.RoleType                    // 匹配 → 角色（仅治理视图用）
```

#### 结构化证据

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
    OutputVerifyResult string  // pass / fail / skip
    AcceptanceResult   string  // pass / fail / manual_review

    // 异常证据
    ErrorMessage   string
    FailurePattern string
    RecoveryAction string
}
```

#### 5 个稳定角色

| 角色 | 职责边界 | 绝对禁区 |
|------|----------|----------|
| **Architect** | 理解需求、设计方案、拆分任务、故障诊断 | 不执行代码修改 |
| **Implementer** | 按规格执行代码/内容创作 | 不做需求判断 |
| **Auditor** | 独立质量评估 | 不修改代码、不做业务决策 |
| **Operator** | 故障恢复、变更风险评估、环境管理 | 不做业务设计 |
| **Coordinator** | 调度编排、冲突裁决、进度控制 | 不做内容生产、不做质量判断 |

---

## 3. Build Line + Operate Line 双主线

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
         │                     │
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

## 4. LLM 边界铁律

### 4.1 LLM 绝对不能做的事

| 决策类型 | 规则引擎 | LLM | 人工 |
|----------|---------|-----|------|
| 状态迁移 | **唯一决策者** | ✗ | ✗ |
| 权限校验 | **唯一决策者** | ✗ | ✗ |
| 风险闸门求值 | **唯一决策者** | ✗ | 命中时确认 |
| 资源锁定/释放 | **唯一决策者** | ✗ | ✗ |
| CAS 状态转换 | **唯一决策者** | ✗ | ✗ |
| 成本阈值判断 | **唯一决策者** | ✗ | 超限审批 |
| 熔断触发条件 | **唯一决策者** | ✗ | ✗ |

### 4.2 LLM 可参与但受约束的决策

| 决策类型 | 规则引擎 | LLM | 人工 |
|----------|---------|-----|------|
| 故障诊断 | 分类规则 | **分析诊断** | C 级审批 |
| 质量评估 | 规则基线 | **辅助 Judge** | manual_review |
| 重规划 | 触发条件 | **方案生成** | B/C 级审批 |
| 执行器推荐 | 契约匹配 | 可选精排 | ✗ |
| 任务拆分 | 格式校验 | **核心生成** | 确认方案 |
| 调度优化 | 冲突规则 | **优化建议** | 冲突裁决 |

---

## 5. 态势感知（Sensor）

**归属：L5 策略层的子组件，为所有层提供运行时数据。**

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
```

**异常信号：**

```go
type AnomalySignal struct {
    Type       string  // failure_spike / duration_drift / throughput_drop / ...
    Severity   string  // info / warning / critical
    Message    string
    Metrics    g.Map
    Confidence float64
}
```

---

## 6. 策略函数（Planner）

**归属：L5 策略层，在 PolicyEngine 基础上增加态势感知能力。**

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
    Reasoning       string
    Confidence      float64
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

**6 个内置策略函数：**

| 策略 | 决策域 | 独立开关 |
|------|--------|----------|
| AdaptiveRetryStrategy | 任务失败时的重试方式 | adaptive_retry_enabled |
| EngineSelectionStrategy | 任务执行前的执行器选择 | engine_selection_enabled |
| BatchAdjustStrategy | 批次推进时的并发度调整 | batch_adjust_enabled |
| CostGuardStrategy | 所有动作前的成本检查 | cost_guard_enabled |
| ProactiveReplanStrategy | 趋势恶化时的提前重规划 | proactive_replan_enabled |
| QualityGateStrategy | 验收决策时的质量标准调整 | quality_gate_enabled |

**Planner 决策流程：**

```
事件触发
  → Sensor.Perceive() → Situation
  → [快速路径] PolicyEngine.Match(trigger)
     → 命中且 strategy_override=false → 直接使用规则结果
     → 未命中或 strategy_override=true → 进入策略评估
  → [策略评估] 遍历 Applicable 的策略函数，按 Priority 排序
     → 选择 Confidence 最高的 ActionPlan
  → [风险检查] RiskGate.Check(plan, situation)
  → 输出 ActionPlan
```

---

## 7. 动作类型总表

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
| switch_executor | 切换任务执行器 | A(同级)/B(升级) |
| adjust_concurrency | 调整并发度 | A(降低)/B(提高) |
| split_batch | 拆分大批次 | B |
| reschedule_task | 延迟重新调度 | A |
| adjust_threshold | 调整闸门阈值 | B(微调)/C(大幅) |

---

## 8. 数据模型变更

### 8.1 新建表（7 个）

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
  `read_set`                json DEFAULT NULL COMMENT '读取资源列表',
  `write_set`               json DEFAULT NULL COMMENT '写入资源列表',
  `side_effect_level`       varchar(16) DEFAULT 'none',
  `rollbackable`            tinyint DEFAULT 1,
  `approval_class`          char(1) DEFAULT 'A',
  `domain_tags`             json DEFAULT NULL,
  `outcome_contract`        json DEFAULT NULL COMMENT '结果契约JSON',
  `acceptance_contract`     json DEFAULT NULL COMMENT '验收契约JSON',
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
) ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='任务契约';

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

### 8.2 修改现有表（6 个）

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

### 8.3 新增配置项

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
-- Learner
('workflow.autonomy.learner_enabled','0','int','autonomy','学习器开关',NOW(),NOW()),
('workflow.autonomy.learning_rate','0.15','string','autonomy','默认学习率',NOW(),NOW()),
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

## 9. 文件清单

### 9.1 新建文件（按实施阶段分组）

#### Phase A（地基）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/situation.go` | 态势模型定义 | 150 |
| `autonomy/sensor.go` | 态势感知 + 异常检测 + 巡检 | 350 |
| `autonomy/objective.go` | 目标模型 + 准入控制 | 200 |
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
| `autonomy/actuator.go` | 动作执行 + 效果跟踪 | 200 |

#### Phase C（契约+能力）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/contract.go` | TaskContract + OutcomeContract + 推断/升级 | 300 |
| `autonomy/capability_profile.go` | 画像模型 + 三层认证 | 200 |
| `autonomy/contract_matcher.go` | 匹配器 + 评分公式 | 300 |
| `autonomy/profile_registry.go` | 画像注册表 | 150 |
| `repo/task_contract_repo.go` | 契约仓储 | 80 |
| `repo/capability_match_log_repo.go` | 匹配日志仓储 | 80 |
| `docker/mysql/upgrade/20260410_l4_phase_c.sql` | DDL: task_contract + match_log + capability_contract | 120 |

#### Phase D（学习+元认知）
| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/learner.go` | EMA 学习 + 保护机制 | 300 |
| `autonomy/meta_observer.go` | 观测器 | 200 |
| `autonomy/meta_assessor.go` | 评估器 | 250 |
| `autonomy/meta_tuner.go` | 校准器 | 200 |
| `repo/engine_stats_repo.go` | 执行器统计仓储 | 100 |
| `repo/learning_repo.go` | 学习记录仓储 | 80 |
| `repo/action_outcome_repo.go` | 效果评估仓储 | 80 |
| `repo/situation_snapshot_repo.go` | 态势快照仓储 | 80 |
| `docker/mysql/upgrade/20260410_l4_phase_d.sql` | DDL: engine_stats + learning + action_outcome | 100 |

### 9.2 修改文件

| 文件 | 改动 |
|------|------|
| `autonomy/decision_center.go` | Decide() 内部升级为 L6→L5→L4→L3→L2 编排链 |
| `autonomy/action_dispatcher.go` | 注册 5 个新动作回调 |
| `autonomy/model.go` | 新增能力词表常量 + 动作类型 + Operator 角色 |
| `orchestrator/registry.go` | Init() 初始化七层组件 |
| `domain/task/task_service.go` | 实例化时生成契约 + 调用匹配器 |
| `stage/execute/service.go` | 查询匹配器推荐执行器 |
| `consts/task.go` | 新增 RoleTypeOperator |

### 9.3 不修改的文件

| 文件 | 原因 |
|------|------|
| `scheduler/domain_task_scheduler.go` | 并发度通过配置动态读取 |
| `watchdog/watchdog.go` | 已通过 SetRetryFn/SetEscalateFn 接入 |
| `stage/accept/service.go` | 已通过 CompleteTrigger/ReworkTrigger 接入 |
| `stage/rework/service.go` | 已通过 AcceptTrigger 接入 |
| `collab/` | 飞书通知通过事件总线订阅 |
| `engine/` | 旧引擎继续为 V1 项目服务 |

---

## 10. 实施路线图

### Phase A：地基（目标层 + 感知 + Operator）

**目标：** 项目有预算/时限约束，系统能持续感知运行态势，故障有专业角色。

| 序号 | 任务 | 说明 |
|------|------|------|
| A1 | Situation 模型 + Sensor 实现 | 态势采集，不影响业务 |
| A2 | ProjectObjective 模型 + AdmissionControl | 准入控制接入 DecisionCenter |
| A3 | Operator 角色上线 | consts + 预设 + system_prompt |
| A4 | 修改 watchdog failure_analysis 使用 Operator | 故障分析专业化 |
| A5 | DDL + 配置项 | situation_snapshot 表 + 配置 |

**灰度：** `objective_enabled=0` 时准入控制不生效。Sensor 纯采集不影响决策。

### Phase B：策略函数（Planner + 6 个策略）

**目标：** 系统能根据态势智能选择动作，而不只是规则匹配。

| 序号 | 任务 | 说明 |
|------|------|------|
| B1 | Strategy 接口 + Planner 编排 | 策略框架 |
| B2 | AdaptiveRetryStrategy | 替换现有固定重试 |
| B3 | EngineSelectionStrategy | 动态执行器选择 |
| B4 | BatchAdjustStrategy | 并发度动态调整 |
| B5 | CostGuardStrategy | 成本守卫 |
| B6 | ProactiveReplanStrategy | 主动重规划 |
| B7 | Actuator + 效果跟踪 | 延迟效果评估 |
| B8 | 修改 decision_center.go | 接入 Planner |

**灰度：** `strategy_enabled=0` 时走原 PolicyEngine 路径。每个策略独立开关。

### Phase C：契约 + 能力匹配

**目标：** 从"角色驱动"升级到"契约驱动"。

| 序号 | 任务 | 说明 |
|------|------|------|
| C1 | TaskContract + OutcomeContract 模型 | 契约定义 |
| C2 | InferContract + EscalateContract | 自动推断 + 失败升级 |
| C3 | CapabilityProfile（三层认证） | 声明+认证+观测 |
| C4 | ProfileRegistry | 从 project_role 加载画像 |
| C5 | ContractMatcher | 评分匹配（先仅记日志） |
| C6 | Stage-Scoped Routing | 批次级主执行链 |
| C7 | 匹配器接入 task_service.go | 替换原 execution_mode 选择 |
| C8 | 架构师契约输出 | 扩展任务 JSON |

**灰度：** `contract_matching_enabled=1` 仅记日志。`contract_matching_active=1` 接入流程。

### Phase D：学习 + 元认知

**目标：** 系统能自我评估、自我校准。

| 序号 | 任务 | 说明 |
|------|------|------|
| D1 | Learner（EMA 学习 + 保护机制） | 从历史数据学习 |
| D2 | EngineStats 统计 | 执行器表现统计 |
| D3 | Observer（观测器） | 记录决策+结果+人工覆盖 |
| D4 | Assessor（评估器） | 定期评估策略准确率 |
| D5 | Tuner（校准器） | 生成参数校准建议 |
| D6 | 自动校准接入 | 保守方向自动，激进方向人工 |

**灰度：** `learner_enabled=0` 不学习。`meta_auto_tune_enabled=0` 不校准。

### Phase E：Operate Line 完整化

**目标：** Build Line + Operate Line 双主线协作。

| 序号 | 任务 | 说明 |
|------|------|------|
| E1 | Operator 完整能力合同（9 个能力） | 运维能力形式化 |
| E2 | Coordinator 升级为控制面代理 | 运行时冲突裁决+超时接管 |
| E3 | Build/Operate 协作协议 | 异常交接+恢复回退 |
| E4 | 运维报告+人工接管界面 | 前端 |

**灰度：** `operate_line_enabled=0` 不启动运维主线。

---

## 11. 灰度策略总览

```
Level 0: 全部关闭                        → 现有 L3.5 行为
  ↓ workflow.autonomy.enabled=1           → 自治中台开启（已有）
Level 1: objective_enabled=1              → 启用目标层准入控制
Level 2: strategy_enabled=1              → 策略函数参与决策
  ↓ 各策略独立开关逐个开启
Level 3: contract_matching_enabled=1     → 能力匹配仅记日志
Level 4: contract_matching_active=1      → 匹配结果接入执行流程
Level 5: learner_enabled=1              → 启用学习
Level 6: meta_cognition_enabled=1       → 启用元认知观测
Level 7: meta_auto_tune_enabled=1       → 启用自动校准（保守方向）
Level 8: patrol_enabled=1               → 启用定时巡检
Level 9: operate_line_enabled=1         → 启用运维主线
Level 10: 全部开启                       → 完整七层自治
```

**每一层可以独立回滚到上一层，不影响其他层。**

---

## 12. 核心调用链

```
事件触发 / 巡检异常
  ↓
[L7] Observer.Record(decision_input)              # 记录输入
  ↓
[L6] ObjectiveCheck(objective, situation)          # 目标边界检查
  ↓ 超出边界 → 拒绝
[L5] AdmissionControl(action, objective, situation)# 准入控制
  ↓ 准入失败 → 拒绝
[L5] PolicyEngine.Match(trigger)                   # 规则匹配
  ↓ strategy_override=true 或 未命中
[L5] Planner.Plan(situation, trigger)              # 策略函数评估
  ↓
[L5] RiskGate.Check(plan, situation)               # 风险闸门
  ↓ 命中 → 降级
[L4] 契约校验（contract 满足？）                     # 契约层
  ↓
[L3] ContractMatcher.Match(contract)               # 能力匹配
  ↓
[L2] Executor.Execute(action)                      # 执行
  ↓
[L1] Evidence.Record(execution_result)             # 记录证据
  ↓
[L7] Observer.RecordOutcome(result)                # 记录结果
  ↓
[L7] Assessor.EvaluateIfNeeded()                   # 偏差检测
  ↓
[L7] Tuner.SuggestIfNeeded()                       # 校准建议
```

---

## 13. 前三份文档归档说明

| 文档 | 归档后状态 |
|------|-----------|
| `L4项目级自治能力设计方案.md` | 归档，内容已合并至本文档 Phase A+B+D |
| `EasyMVP角色体系演进：契约驱动能力编排设计方案.md` | 归档，内容已合并至本文档 Phase C + L1/L4 |
| `EasyMVP自治系统终极架构：六层自治模型设计方案.md` | 归档，内容已合并至本文档 L6/L7 + 双主线 |

**本文档发布后，上述三份文档不再作为实施依据，保留在 docs/ 目录供历史参考。**
