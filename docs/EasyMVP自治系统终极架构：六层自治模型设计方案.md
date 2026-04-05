# EasyMVP 自治系统终极架构：六层自治模型设计方案

> 在契约驱动能力编排方案基础上，补充 Goal Layer + Outcome/Evidence + Meta-Cognition，
> 形成完整的六层自治闭环。

---

## 1. 设计定位

本方案**不替换**现有契约驱动能力编排方案（见 `EasyMVP角色体系演进：契约驱动能力编排设计方案.md`），
而是在其上做结构升级：

| 升级点 | 说明 |
|--------|------|
| 新增 L6 元认知层 | Observer + Assessor + Tuner，系统自我评估与参数校准 |
| 新增 L5 目标层 | 项目级约束（预算/时限/风险容忍度/自治边界） |
| 扩展 L3 契约层 | +OutcomeContract +AcceptanceContract，执行闭环+验收闭环 |
| 保留 L4/L2/L1 | 策略层、能力层、执行层不变 |

**核心理念：目标驱动 → 策略约束 → 契约执行 → 证据回写 → 元认知校准**

---

## 2. 六层架构总览

```
┌──────────────────────────────────────────────────────────┐
│ L6  元认知层 Meta-Cognition                                │
│     Observer · Assessor · Tuner                            │
│     "系统做得怎么样？参数对不对？该怎么改进？"                  │
├──────────────────────────────────────────────────────────┤
│ L5  目标层 Objective                                       │
│     目标 · 预算 · 时限 · 风险容忍度 · SLA · 自治边界         │
│     "完成什么？不能碰什么？花多少？什么时候停？"                │
├──────────────────────────────────────────────────────────┤
│ L4  策略层 Policy（已有）                                   │
│     准入控制 · 风险闸门 · 决策分级 · 人工节点                 │
│     "允不允许做？自动还是人工？做完怎么验？失败怎么收？"         │
├──────────────────────────────────────────────────────────┤
│ L3  契约层 Contract（已有 + Outcome/Acceptance 扩展）        │
│     TaskContract · OutcomeContract · AcceptanceContract     │
│     "任务需要什么？成功产出什么？怎么证明合格？"                │
├──────────────────────────────────────────────────────────┤
│ L2  能力层 Capability（已有）                                │
│     画像(声明+认证+观测) · 匹配器 · 降级路径                 │
│     "谁能满足契约？代价多大？失败了怎么降级？"                 │
├──────────────────────────────────────────────────────────┤
│ L1  执行层 Execution（已有）                                │
│     执行器市场 · 模型 · 工具 · 沙箱 · 资源                  │
│     "用什么工具？跑在哪？消耗多少？健康吗？"                   │
└──────────────────────────────────────────────────────────┘
```

**层间规则：**
- 每层只和相邻层交互，不跨层调用
- 上层约束下层，下层向上层反馈
- 每层独立灰度开关，可独立回滚

---

## 3. L6 元认知层（Meta-Cognition）

### 3.1 职责

观测系统自身行为，评估效果，校准参数。**不做业务决策，做"关于决策的决策"。**

### 3.2 三要素

| 组件 | 职责 | 输入 | 输出 |
|------|------|------|------|
| **Observer**（观测器） | 记录每次决策的输入/输出/结果/人工覆盖 | 决策事件流 | ObservationRecord |
| **Assessor**（评估器） | 定期评估策略准确率、闸门误报率、匹配精度 | 观测记录集 | AssessmentResult + Drift[] |
| **Tuner**（校准器） | 生成参数校准建议 | 偏差报告 | TuneRecommendation[] |

### 3.3 核心数据模型

```go
// ObservationRecord 决策观测记录。
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

// AssessmentResult 系统评估结果。
type AssessmentResult struct {
    PolicyAccuracy     float64 // 策略匹配后动作成功率
    GateFalsePositive  float64 // 闸门误报率（不该拦的拦了）
    GateFalseNegative  float64 // 闸门漏报率（该拦的没拦）
    HumanOverrideRate  float64 // 人工覆盖率（越高说明系统判断越差）
    MatchAccuracy      float64 // 能力匹配后任务成功率
    CostEfficiency     float64 // 成本效率（成功Token / 总Token）
    Drifts             []Drift // 参数偏差列表
}

// Drift 参数偏差。
type Drift struct {
    Parameter    string  // 偏离的参数名
    CurrentValue float64 // 当前值
    OptimalValue float64 // 观测最优值
    Confidence   float64 // 置信度
    Evidence     string  // 证据说明
}

// TuneRecommendation 校准建议。
type TuneRecommendation struct {
    Parameter      string
    CurrentValue   interface{}
    SuggestedValue interface{}
    Reasoning      string
    Confidence     float64
    AutoApplicable bool   // 是否可自动应用（仅保守方向可自动）
    RiskLevel      string // low / medium / high
}
```

### 3.4 校准安全铁律

| 方向 | 规则 |
|------|------|
| 变保守（降并发、收紧闸门、升级决策级别） | ✅ 可自动应用（confidence > 0.7） |
| 变激进（升并发、放宽闸门、降级决策级别） | ❌ 必须人工确认 |
| 目标层参数变更 | ❌ 始终人工确认 |

**这是整个元认知层的安全底线：系统可以自动变得更保守，但变得更激进必须人工确认。**

### 3.5 学习信号强度

| 场景 | 信号强度 | 说明 |
|------|----------|------|
| 系统决策 A 级自动执行，人工没干预 | 弱信号 | 可能对也可能人没看 |
| 系统决策 B 级，人工批准 | 中信号 | 人认可了 |
| 系统决策 B 级，人工驳回 | **强信号** | 系统判断错了 |
| 系统决策 C 级，人工给出方案 | 学习样本 | 系统无能力判断的领域 |

### 3.6 运行周期

```
Observer: 每次决策实时记录（零延迟，异步写）
Assessor: 每完成一个阶段评估一次，或每 N 次决策后评估
Tuner:    Assessor 发现 Drift 后触发，生成建议
```

### 3.7 灰度

- `meta_cognition_enabled=0` → Observer 不记录，Assessor/Tuner 不运行
- `meta_auto_tune_enabled=0` → Tuner 只生成建议不自动应用

---

## 4. L5 目标层（Objective）

### 4.1 职责

定义项目的目标和约束。**是整个自治系统的方向盘——所有下层决策必须在此边界内运行。**

### 4.2 ProjectObjective 模型

```go
// ProjectObjective 项目目标约束。
type ProjectObjective struct {
    // 交付目标
    DeliveryGoal       string     // 交付物描述
    QualityFloor       float64    // 最低质量标准（0-1）

    // 预算约束
    TokenBudget        int64      // Token 总预算（0=不限）
    TimeBudgetHours    float64    // 时间预算（小时）
    CostBudgetCents    int64      // 金钱预算（分）

    // 风险约束
    RiskTolerance      string     // conservative / balanced / aggressive
    MaxAutoRetries     int        // 自动重试上限（硬限，到了就停）
    MaxAutoReworks     int        // 自动返工上限
    MaxAutoReplans     int        // 自动重规划上限

    // SLA
    DeadlineAt         *time.Time // 截止时间（nil=不限）
    MaxStallMinutes    int        // 最大停滞时间（超过则告警）

    // 自治边界
    AutonomyLevel      string     // manual / supervised / autonomous
    // manual:     所有动作人工确认
    // supervised: A 级自动，B/C 级人工（默认值）
    // autonomous: A/B 级自动，仅 C 级人工
}
```

### 4.3 AdmissionControl（准入控制）

**目标层是约束而不是优化目标。它不"追求"什么，它划边界。**

所有动作在进入策略层之前，先过 AdmissionControl：

```go
// AdmissionResult 准入控制结果。
type AdmissionResult struct {
    Allowed    bool
    DenyReason string
    Conditions []string // 附加条件（如"降级到轻量执行器"）
}
```

| 检查项 | 条件 | 动作 |
|--------|------|------|
| 预算检查 | TokensConsumed > TokenBudget × 90% | 拒绝高成本动作，允许轻量动作 |
| 时间检查 | Now > DeadlineAt | 拒绝非必要动作 |
| 重试上限 | RetryCount >= MaxAutoRetries | 拒绝自动重试，升级人工 |
| 返工上限 | ReworkRounds >= MaxAutoReworks | 拒绝自动返工，升级人工 |
| 重规划上限 | ReplanCount >= MaxAutoReplans | 拒绝自动重规划 |
| 自治级别 | AutonomyLevel = manual | 所有动作需人工确认 |

### 4.4 目标如何约束下层

```
RiskTolerance = "conservative"
  → L4 策略层：熔断阈值从 5 降为 3，B 级不自动执行
  → L2 能力层：匹配偏向历史表现好的画像
  → L1 执行层：并发度从 20 降到 10

RiskTolerance = "aggressive"
  → L4 策略层：熔断阈值从 5 升为 8，B 级可自动执行
  → L2 能力层：允许尝试新执行器
  → L1 执行层：并发度可到 30

TokenBudget 消耗 90%
  → L4 策略层：准入控制拒绝新的高成本动作
  → L2 能力层：切换到 lite 执行器
  → L6 元认知层：记录预算告警事件
```

### 4.5 灰度

- `objective_enabled=0` → AdmissionControl 始终返回 Allowed=true，不做约束
- 开启后默认 `supervised` 模式，与当前行为一致

---

## 5. L3 契约层扩展（OutcomeContract + AcceptanceContract）

### 5.1 设计选择：不独立成层

OutcomeContract 和 AcceptanceContract 作为 TaskContract 的扩展，而非独立层。原因：

1. Outcome/Acceptance 与 Task 是 1:1 绑定
2. 放同一层减少跨层调用
3. 契约升级时可同时升级能力需求和验收标准

### 5.2 OutcomeContract（结果契约）

```go
// OutcomeContract 结果契约 — 定义成功后必须产出什么。
type OutcomeContract struct {
    RequiredOutputs []RequiredOutput // 必须产出列表
    SuccessCriteria []string         // 成功判定条件（规则表达式）
    FailureCriteria []string         // 失败判定条件
}

// RequiredOutput 必须产出。
type RequiredOutput struct {
    Name       string // 产出名称（如 "modified_files", "test_result"）
    Type       string // file / text / json / metrics
    Schema     g.Map  // 产出结构 Schema（可选）
    Mandatory  bool   // 是否必须（false=建议产出）
}
```

**推断规则（InferOutcomeContract）：**

| role_type | 默认 RequiredOutputs | 默认 SuccessCriteria |
|-----------|---------------------|---------------------|
| implementer | modified_files(file), diff_summary(text) | compile_pass, no_syntax_error |
| auditor | review_report(json), issues(json) | report_generated |
| architect | plan(json), task_list(json) | tasks_count > 0 |
| operator | diagnosis(json), recovery_plan(json) | root_cause_identified |

### 5.3 AcceptanceContract（验收契约）

```go
// AcceptanceContract 验收契约 — 定义怎么验证产出合格。
type AcceptanceContract struct {
    VerifyMethod     string   // rule / llm_judge / human / auto
    RuleSet          []string // 规则集代码列表
    QualityFloor     float64  // 最低质量分（0-1）
    EvidenceRequired []string // 需要的证据类型（diff/test_result/coverage/lint）
}
```

### 5.4 契约升级联动

失败后 `EscalateContract()` 同时升级三个契约：

```go
func EscalateContract(tc *TaskContract, oc *OutcomeContract, ac *AcceptanceContract, failurePattern string) {
    // 1. 升级 TaskContract（已有逻辑）
    tc.AccuracyRequirement = "high"
    tc.ContextScope = "global"

    // 2. 升级 OutcomeContract
    oc.SuccessCriteria = append(oc.SuccessCriteria, "test_pass")
    oc.RequiredOutputs = append(oc.RequiredOutputs, RequiredOutput{
        Name: "regression_check", Type: "json", Mandatory: true,
    })

    // 3. 升级 AcceptanceContract
    ac.QualityFloor = max(ac.QualityFloor, 0.85)
    if ac.VerifyMethod == "auto" {
        ac.VerifyMethod = "rule" // 从自动升级到规则验证
    }
    ac.EvidenceRequired = appendUnique(ac.EvidenceRequired, "test_result", "coverage")
}
```

### 5.5 执行证据（ExecutionEvidence）

每次任务执行完成后，系统自动收集结构化证据：

```go
// ExecutionEvidence 执行证据。
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
    OutcomeVerifyResult string  // pass / fail / skip
    AcceptanceResult    string  // pass / fail / manual_review

    // 异常证据
    ErrorMessage   string
    FailurePattern string
    RecoveryAction string
}
```

证据回写路径：
```
任务完成 → 收集 ExecutionEvidence
  → 验证 OutcomeContract（必须产出是否齐全）
  → 写入证据记录
  → 触发 AcceptanceContract（规则/LLM/人工验收）
  → 回写验收结果
  → L6 Observer 记录观测
```

---

## 6. 升级后的决策流程

```
事件触发
  → L5 AdmissionControl（准入检查：预算/时限/次数）
     → 拒绝 → 记录 + 通知人工
     → 允许 ↓
  → L4 PolicyEngine.Match（策略匹配 + 风险闸门）
     → 输出动作 + 决策级别
  → L3 ContractMatcher.Match（契约匹配：找最优执行者）
     → 包含 OutcomeContract（定义成功标准）
  → L2 CapabilityProfile（能力画像评分）
  → L1 Executor.Execute（执行）
     → 完成后 → 收集 ExecutionEvidence
     → 验证 OutcomeContract + AcceptanceContract
  → L6 Observer.Record（记录决策观测）
     → Assessor 定期评估 → Tuner 校准建议
```

---

## 7. 与现有代码的集成点

### 7.1 新增文件

| 文件 | 层 | 职责 |
|------|------|------|
| `autonomy/objective.go` | L5 | ProjectObjective 模型 + AdmissionControl |
| `autonomy/meta_observer.go` | L6 | 观测器 |
| `autonomy/meta_assessor.go` | L6 | 评估器 |
| `autonomy/meta_tuner.go` | L6 | 校准器 |

### 7.2 修改文件

| 文件 | 改动 |
|------|------|
| `autonomy/model.go` | 新增 ObservationRecord / AssessmentResult / ExecutionEvidence 等类型 |
| `autonomy/decision_center.go` | Decide() 入口增加 AdmissionControl 前置检查 |
| `autonomy/contract.go`（已规划） | 扩展 OutcomeContract + AcceptanceContract |
| `orchestrator/registry.go` | Init() 初始化 L5/L6 组件 |

### 7.3 不修改的文件

| 文件 | 原因 |
|------|------|
| `autonomy/policy_engine.go` | L4 不变 |
| `autonomy/risk_gate.go` | L4 不变 |
| `autonomy/action_dispatcher.go` | L1 不变 |
| 所有 `stage/` 下文件 | 通过回调接入，不直接修改 |
| 所有 `executor/` 下文件 | L1 不变 |

---

## 8. 数据模型变更

### 8.1 mvp_project 新增字段

```sql
ALTER TABLE `mvp_project`
  ADD COLUMN `objective_json` json DEFAULT NULL
  COMMENT '项目目标约束(JSON): budget/deadline/risk_tolerance/autonomy_level';
```

### 8.2 新增配置项

```sql
INSERT INTO `mvp_config` (`config_key`,`config_value`,`config_type`,`category`,`description`,`created_at`,`updated_at`) VALUES
('workflow.autonomy.objective_enabled','0','int','autonomy','目标层准入控制开关',NOW(),NOW()),
('workflow.autonomy.meta_cognition_enabled','0','int','autonomy','元认知观测开关',NOW(),NOW()),
('workflow.autonomy.meta_auto_tune_enabled','0','int','autonomy','自动校准开关(仅保守方向可自动)',NOW(),NOW()),
('workflow.autonomy.default_autonomy_level','supervised','string','autonomy','默认自治级别',NOW(),NOW()),
('workflow.autonomy.default_risk_tolerance','balanced','string','autonomy','默认风险容忍度',NOW(),NOW())
ON DUPLICATE KEY UPDATE `updated_at`=NOW();
```

---

## 9. 灰度策略（10 级）

| 级别 | 开关状态 | 行为 |
|------|----------|------|
| 0 | 全部关闭 | 与 L3.5 行为完全一致 |
| 1 | patrol_enabled=1 | Sensor 开始采集态势，不影响决策 |
| 2 | objective_enabled=1 | AdmissionControl 生效，超限动作被拒绝 |
| 3 | strategy_enabled=1 | 策略函数参与决策（各策略独立开关） |
| 4 | contract_matching_enabled=1 | 契约匹配仅记日志（shadow mode） |
| 5 | contract_matching_active=1 | 匹配结果接入执行流程 |
| 6 | learner_enabled=1 | 学习器开始 EMA 更新 |
| 7 | meta_cognition_enabled=1 | Observer 开始记录 |
| 8 | meta_auto_tune_enabled=1 | Tuner 可自动应用保守方向建议 |
| 9 | autonomy_level=autonomous | A/B 级自动，仅 C 级人工 |

**每一级都可以独立回滚到上一级，不影响其他层。**

---

## 10. 稳定性保证

1. **每层独立灰度** — 任何一层出问题都可以单独关闭
2. **元认知铁律** — 系统只能自动变保守，变激进必须人工
3. **目标层只做减法** — 只拒绝超限操作，不主动发起操作
4. **向后兼容** — L5/L6 全部关闭时，行为与当前基线完全一致
5. **证据可追溯** — 每次决策都有 ObservationRecord，事后可审计
6. **不引入新依赖** — 全部用 Go 标准库 + GoFrame，无新三方包

---

## 11. 实施顺序建议

| 阶段 | 内容 | 前置 |
|------|------|------|
| Phase A | L5 目标层（Objective + AdmissionControl） | 无 |
| Phase B | L3 扩展（OutcomeContract + AcceptanceContract + Evidence） | Phase A |
| Phase C | L6 元认知层（Observer → Assessor → Tuner） | Phase A+B |
| Phase D | 灰度级别 1-5 逐级打开 | Phase A+B |
| Phase E | 灰度级别 6-9 逐级打开 | Phase C+D |

每个 Phase 独立编译、独立测试、独立上线。
