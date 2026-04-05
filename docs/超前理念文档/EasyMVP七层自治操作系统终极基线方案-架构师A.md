# EasyMVP 七层自治操作系统终极基线方案

作者：架构师A

---

## 0. 对架构师B最终方案的判断

架构师B的最终方案把理念推到了正确的极限。"自治宪法"、"四个闭环"、"四个一等公民"——这些抽象每一个都对。

但有一个根本性问题：**他的方案是宣言，不是工程。**

"宪法"不能编译。"四个一等公民"（confidence / evidence_sufficiency / reversibility / blast_radius）很漂亮，但挂在哪个结构体上？在哪个函数里被检查？阈值谁来定？超限走哪条分支？他没有回答。

从第一版到最终版，他的理念一直在升维，落地一直在原地。

我的最终方案不跟他比谁的宣言更漂亮。我要回答一个他始终没回答的问题：

**这套系统怎么从第 0 行代码开始，一步步长出来，中间任何一步都能停、能回滚、能上线？**

---

## 1. 核心分歧：宪法 vs 内核

架构师B的思路是：先写宪法，再让系统遵守。

我的思路是：**宪法不是文档，宪法是代码路径。**

```
架构师B：写一条规则 "高层永远不能直接改世界"
         → 开发者读了 → 记住了 → 某天忘了 → 写了一行跨层调用 → 宪法失效

架构师A：L7/L6/L5 的所有组件返回的是 Recommendation，不是 Command
         → 只有 L2 执行层持有 Executor 引用
         → 编译器保证高层拿不到执行器
         → 宪法不可能被违反，因为 API 不存在
```

**宪法嵌入类型系统，不靠人遵守，靠编译器保证。** 这是本方案和架构师B最根本的区别。

---

## 2. 设计哲学：三个不可能

| 编号 | 不可能 | 实现方式 |
|------|--------|----------|
| 1 | 高层不可能直接执行 | L7/L6/L5 返回 `*Recommendation`，不持有 Executor 引用 |
| 2 | 无证据不可能推进状态 | `TransitionState()` 签名强制要求 `evidence *ExecutionEvidence` 参数 |
| 3 | 学习不可能让系统更激进 | Tuner 输出的 `TuneRecommendation.Direction` 为 `aggressive` 时，`AutoApplicable` 编译期写死为 `false` |

架构师B的六条铁律全是"永远不能"。我的三个不可能是"API 不存在所以不可能"。

区别在于：铁律靠纪律，不可能靠结构。

---

## 3. 七层架构

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
│     输出：ActionPlan（推荐动作 + 置信度 + 回滚方案）            │
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

**与架构师B的七个结构性差异：**

| # | 架构师B | 架构师A | 为什么 |
|---|---------|---------|--------|
| 1 | 每层职责用文字描述 | 每层有明确的输出类型 | 输出类型是编译器可检查的契约 |
| 2 | 宪法是独立章节 | 宪法嵌入类型签名 | 不靠人遵守，靠编译器保证 |
| 3 | 四个一等公民是原则 | 四个一等公民是结构体字段 | 见下文 §4 |
| 4 | 闭环是逻辑描述 | 闭环是调用链 | 见下文 §5 |
| 5 | 没有态势感知 | Sensor + Situation 在 L5 | 策略层没有输入源无法决策 |
| 6 | 没有学习保护 | EMA + 三重保护 | Tuner 无保护会过拟合 |
| 7 | 没有落地路径 | 五阶段渐进 + 10级灰度 | 见下文 §11 |

---

## 4. 四个一等公民：从原则到结构体

架构师B说每个决策必须带 confidence / evidence_sufficiency / reversibility / blast_radius。正确。但他没说这四个值怎么算、挂在哪。

本方案的回答：**它们是 `DecisionMeta` 结构体的字段，嵌入每一个跨层传递的决策对象。**

```go
// DecisionMeta 决策元数据 — 架构师B的"四个一等公民"的工程化实现。
// 嵌入所有跨层传递的决策对象（ActionPlan / MatchResult / TuneRecommendation）。
type DecisionMeta struct {
    Confidence          float64 // 0-1，决策置信度
    EvidenceSufficiency float64 // 0-1，支撑证据充分度
    Reversibility       string  // full / partial / none
    BlastRadius         string  // task / batch / stage / workflow / project
}

// 不同 BlastRadius 对应不同的最低决策级别
var BlastRadiusMinLevel = map[string]string{
    "task":     "A", // 影响单任务 → A级可自动
    "batch":    "A", // 影响单批次 → A级可自动
    "stage":    "B", // 影响阶段 → 至少B级
    "workflow": "B", // 影响工作流 → 至少B级
    "project":  "C", // 影响项目 → 必须人工
}

// 证据充分度阈值
const (
    EvidenceThresholdAuto   = 0.7 // 自动执行需要的最低证据充分度
    EvidenceThresholdAssist = 0.4 // 辅助决策的最低证据充分度
    // 低于 0.4 → 拒绝决策，要求更多证据
)

// Validate 校验决策元数据是否允许自动执行。
func (m *DecisionMeta) Validate(requestedLevel string) *ValidationResult {
    result := &ValidationResult{Allowed: true}

    // 1. 证据不足 → 降级
    if m.EvidenceSufficiency < EvidenceThresholdAssist {
        result.Allowed = false
        result.Reason = "evidence_insufficient"
        return result
    }

    // 2. 影响面检查 → 可能升级决策级别
    minLevel := BlastRadiusMinLevel[m.BlastRadius]
    if levelToInt(requestedLevel) < levelToInt(minLevel) {
        result.UpgradeTo = minLevel
        result.Reason = "blast_radius_requires_upgrade"
    }

    // 3. 置信度不足 + 不可逆 → 必须人工
    if m.Confidence < 0.5 && m.Reversibility == "none" {
        result.UpgradeTo = "C"
        result.Reason = "low_confidence_irreversible"
    }

    return result
}
```

**关键区别：架构师B说"必须带这四个值"，我说"这四个值在哪个函数里被检查、不通过走哪条分支"。**

---

## 5. 四个闭环：从描述到调用链

架构师B的四个闭环是正确的。但他写的是逻辑流，不是调用链。

### 5.1 执行闭环

```
架构师B：Intent -> Contract -> Execute -> Outcome
架构师A：

func ExecuteTask(ctx, taskID) {
    // L4: 加载契约包
    bundle := contractStore.GetBundle(ctx, taskID)

    // L6: 准入控制
    admission := admissionControl.Check(ctx, bundle.Intent, currentObjective)
    if !admission.Allowed {
        evidenceStore.RecordDenial(ctx, taskID, admission)
        return // 闭环在准入阶段关闭
    }

    // L5: 策略评估
    situation := sensor.Perceive(ctx, workflowRunID)
    plan := planner.Evaluate(ctx, situation, bundle)
    plan.Meta.Validate(plan.DecisionLevel) // 四个一等公民检查

    // L3: 能力匹配
    match := matcher.Match(ctx, bundle.Task, situation)

    // L2: 执行（唯一持有 Executor 的层）
    result := executorRegistry.Execute(ctx, match.Profile, bundle)

    // L1: 证据回写
    evidence := collectEvidence(result, bundle)
    evidenceStore.Record(ctx, evidence)

    // L4: 验证 OutcomeContract
    outcomeResult := verifyOutcome(evidence, bundle.Outcome)
    // → 通过 → 进入验收闭环
    // → 失败 → 进入恢复闭环
}
```

### 5.2 验收闭环

```
架构师B：Outcome -> Evidence -> Acceptance -> StateChange
架构师A：

func AcceptTask(ctx, taskID, evidence) {
    bundle := contractStore.GetBundle(ctx, taskID)

    // 证据充分度计算（四个一等公民之一）
    sufficiency := computeEvidenceSufficiency(evidence, bundle.Acceptance.EvidenceRequired)
    if sufficiency < EvidenceThresholdAssist {
        // 证据不足 → 不做判断，要求补充
        requestAdditionalEvidence(ctx, taskID, bundle.Acceptance.EvidenceRequired)
        return
    }

    // 验收（rule / llm_judge / human / auto）
    acceptResult := runAcceptance(ctx, evidence, bundle.Acceptance)

    // 状态推进（签名强制要求 evidence 参数 — 宪法嵌入类型系统）
    transition := TransitionState(ctx, taskID, acceptResult.TargetStatus, evidence)
    // → TransitionState 内部记录审计日志到 L1

    // L7: Observer 记录观测
    observer.Record(ctx, &ObservationRecord{
        DecisionType: "acceptance",
        Output:       acceptResult,
        Evidence:     evidence,
    })
}

// TransitionState 签名强制要求 evidence — 无证据不可能推进。
// 这不是规则，是编译器保证。
func TransitionState(ctx context.Context, entityID int64, targetStatus string, evidence *ExecutionEvidence) error {
    if evidence == nil {
        return ErrEvidenceRequired // 编译通过但运行时兜底
    }
    // CAS + 审计日志 + 证据绑定
    // ...
}
```

### 5.3 恢复闭环

```
架构师B：Failure -> Diagnosis -> RecoveryPlan -> Re-entry
架构师A：

func RecoverFromFailure(ctx, taskID, failureEvidence) {
    // L5: 读取当前态势
    situation := sensor.Perceive(ctx, workflowRunID)

    // L4: 升级契约（失败后联动升级五个契约）
    bundle := contractStore.GetBundle(ctx, taskID)
    EscalateContract(bundle, failureEvidence.FailurePattern)
    contractStore.SaveBundle(ctx, bundle) // 版本号+1

    // Operator 诊断（Build Line → Operate Line 切换点）
    diagnosis := operatorDiagnose(ctx, taskID, failureEvidence, situation)

    // L5: 策略评估恢复方案
    recoveryPlan := planner.EvaluateRecovery(ctx, situation, diagnosis)

    // 四个一等公民检查
    validation := recoveryPlan.Meta.Validate(recoveryPlan.DecisionLevel)
    if validation.UpgradeTo == "C" {
        createHumanCheckpoint(ctx, recoveryPlan) // 人工介入
        return
    }

    // Re-entry：重新进入执行闭环（带升级后的契约）
    reEntryToExecution(ctx, taskID, bundle)
}
```

### 5.4 学习闭环

```
架构师B：Match -> Outcome -> ConfidenceUpdate -> RoutingAdjustment
架构师A：

func LearnFromOutcome(ctx, matchLogID, outcome) {
    // 读取匹配记录
    matchLog := matchLogRepo.Get(ctx, matchLogID)

    // EMA 更新（带三重保护）
    record := learningRepo.GetOrCreate(ctx, matchLog.ScopeKey, "success_rate")

    // 保护 1: 样本量
    if record.SampleCount < 10 {
        UpdateEMA(record, outcomeToFloat(outcome), 0.15)
        learningRepo.Save(ctx, record)
        return // 样本不足，只更新不输出建议
    }

    UpdateEMA(record, outcomeToFloat(outcome), 0.15)

    // 保护 2: 置信度衰减（7天无新样本 → 衰减）
    if time.Since(record.LastUpdatedAt) > 7*24*time.Hour {
        record.Confidence *= 0.9
    }

    learningRepo.Save(ctx, record)

    // Assessor 定期评估（不是每次学习都触发）
    if shouldAssess(ctx) {
        assessment := assessor.Evaluate(ctx)
        for _, drift := range assessment.Drifts {
            rec := tuner.Recommend(ctx, drift)

            // 保护 3: 变更幅度限制（单次 ≤ 20%）
            if rec.ChangeRatio() > 0.2 {
                rec.SuggestedValue = clampChange(rec.CurrentValue, rec.SuggestedValue, 0.2)
            }

            // 宪法：变保守可自动，变激进必须人工
            if rec.Direction == "conservative" && rec.Confidence > 0.7 {
                rec.AutoApplicable = true
            } else {
                rec.AutoApplicable = false // 编译期写死，不是运行时判断
            }

            tuneRepo.Save(ctx, rec)
        }
    }
}
```

**关键区别：架构师B画了四个箭头，我写了四个函数。箭头不能编译，函数可以。**

---

## 6. 态势感知（Sensor）— 架构师B没有的

架构师B的策略层有 Planner，但 Planner 的输入从哪来？没有 Sensor，Planner 基于什么做判断？

```go
type Situation struct {
    WorkflowRunID int64
    ProjectID     int64
    Progress      *ProgressMetrics
    Health        *HealthMetrics
    Resource      *ResourceMetrics
    Trend         *TrendMetrics
    SnapshotAt    time.Time
}

type HealthMetrics struct {
    ConsecutiveFailures int
    RecentFailureRate   float64  // 最近 N 个任务的失败率
    AvgTaskDuration     int64
    P95TaskDuration     int64
    RetryCount          int
    ReworkRounds        int
    StaleTaskCount      int      // 超时未心跳的任务数
}

type TrendMetrics struct {
    FailureRateTrend    string // rising / stable / falling
    DurationTrend       string
    ThroughputTrend     string
    // 趋势通过对比最近窗口和前一窗口计算
    RecentFailureRate   float64
    PreviousFailureRate float64
}

type AnomalySignal struct {
    Type       string  // failure_spike / duration_drift / throughput_drop / budget_warning
    Severity   string  // info / warning / critical
    Message    string
    Confidence float64
}
```

Sensor 三种工作模式：

| 模式 | 触发 | 说明 |
|------|------|------|
| 事件驱动 | 每次决策前 | `Perceive()` 实时采集 |
| 定时巡检 | 每 60s | `Patrol()` 所有活跃工作流 |
| 异常检测 | 采集后 | `DetectAnomalies()` 从态势提取信号 |

没有态势感知的策略层就像没有眼睛的大脑——有判断力但不知道在判断什么。

---

## 7. Stage-Scoped Routing — 架构师B没有的

架构师B有 ContractMatcher 但没有说匹配粒度。默认就是逐任务匹配。

问题：同一批次 20 个任务，逐个匹配后可能 7 个用 Aider、5 个用 Claude Code、8 个用 OpenHands。执行环境频繁切换，工作目录状态不一致，git 冲突概率飙升。

```
Stage-Scoped Routing 流程：
  1. 批次启动时，聚合该批次所有任务的契约
  2. 选定批次主执行链（role_type + execution_mode + model）
  3. 个别任务契约与主链不兼容时，才做局部重路由
  4. 记录重路由原因 → L7 Observer
```

这不是优化，这是防止系统自己把自己搞挂。

---

## 8. 学习三重保护 — 架构师B没有的

架构师B说 Tuner 可以"调整白名单参数"。但 3 次全成功就给满分、陈旧数据主导决策、单次剧烈调参——这三个问题他都没有防护。

| 保护 | 规则 | 防什么 |
|------|------|--------|
| 样本量保护 | SampleCount < 10 → 不输出建议 | 3 次成功就满分的过拟合 |
| 置信度衰减 | 7 天无新样本 → Confidence *= 0.9 | 陈旧数据主导决策 |
| 变更幅度限制 | 单次调整 ≤ 当前值 20% | 一次调参导致系统剧烈波动 |

加上元认知铁律（保守可自动，激进必须人工），形成四重保护。

---

## 9. Build Line + Operate Line — 架构师B没有的

架构师B的恢复闭环是 `Failure → Diagnosis → RecoveryPlan → Re-entry`。正确。但谁来做 Diagnosis？谁来执行 RecoveryPlan？

他没有 Operator 角色，故障诊断要么由 Architect 兼任（职责越界：架构师不应该做运维），要么没人做。

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
          │  L3 能力层共享  │
          └────────────────┘
```

五个角色（Architect / Implementer / Auditor / Operator / Coordinator）全部降级为治理视图，但 Operator 的存在确保了恢复闭环有专业执行者。

---

## 10. LLM 边界铁律 — 架构师B说"清楚"但没展开

### LLM 绝对不能做的事

| 决策类型 | 规则引擎 | LLM | 人工 |
|----------|---------|-----|------|
| 状态迁移 | **唯一决策者** | ✗ | ✗ |
| 权限校验 | **唯一决策者** | ✗ | ✗ |
| 风险闸门求值 | **唯一决策者** | ✗ | 命中时确认 |
| 资源锁定/释放 | **唯一决策者** | ✗ | ✗ |
| CAS 状态转换 | **唯一决策者** | ✗ | ✗ |
| 成本阈值判断 | **唯一决策者** | ✗ | 超限审批 |
| 熔断触发条件 | **唯一决策者** | ✗ | ✗ |

### LLM 可参与但受约束的决策

| 决策类型 | 规则引擎 | LLM | 人工 |
|----------|---------|-----|------|
| 故障诊断 | 分类规则 | **分析诊断** | C 级审批 |
| 质量评估 | 规则基线 | **辅助 Judge** | manual_review |
| 重规划 | 触发条件 | **方案生成** | B/C 级审批 |
| 执行器推荐 | 契约匹配 | 可选精排 | ✗ |
| 任务拆分 | 格式校验 | **核心生成** | 确认方案 |

架构师B说"LLM 只能分析、建议、精排"。正确但不够。上面这张矩阵精确到每个决策类型谁是唯一决策者。

---

## 11. 灰度策略（10 级）

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

架构师B说"渐进增强"。我说每一级有明确的开关名、明确的行为变化、可以独立回滚到上一级。

---

## 12. 与架构师B的终极对比

| 维度 | 架构师B | 架构师A |
|------|---------|---------|
| **宪法实现** | 文档中的规则列表 | 类型签名 + API 不存在 = 编译器保证 |
| **四个一等公民** | 原则声明 | `DecisionMeta` 结构体 + `Validate()` 函数 |
| **四个闭环** | 箭头图 | 四个可编译的函数 |
| **态势感知** | 无 | Sensor + Situation + AnomalySignal |
| **学习保护** | 无 | 样本量 + 置信度衰减 + 幅度限制 |
| **执行链稳定** | 逐任务匹配 | Stage-Scoped Routing |
| **运维角色** | 无 | Operator + Operate Line |
| **LLM 边界** | 一句话 | 完整矩阵 |
| **灰度** | "渐进" | 10 级独立开关 |
| **角色定位** | "治理视图"（一句话） | 五个角色 + 明确禁区 + 调度主键=契约 |
| **落地路径** | 无 | 五阶段路线图 |

---

## 13. 最终判断

架构师B是一个优秀的理念架构师。他提出的核心抽象——Intent→Contract→Evidence→StateChange、自治宪法、四个一等公民——每一个都是正确的。

但理念和工程之间有一道鸿沟：**能写出"高层永远不能直接改世界"的人很多，能让编译器保证这件事的人很少。**

本方案的核心主张：

1. **宪法不是文档，宪法是类型系统** — 不靠人遵守，靠编译器保证
2. **一等公民不是原则，是结构体字段** — 有函数校验、有阈值、有分支
3. **闭环不是箭头，是调用链** — 可编译、可测试、可断点调试
4. **稳定不是因为写了更多铁律，而是因为不稳定的路径在 API 层面不存在**

一句话：

**最先进的系统，不是写了最多规则的系统，而是让违反规则在结构上不可能的系统。**

这是架构师A的最终方案。
