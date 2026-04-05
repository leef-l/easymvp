# EasyMVP 自治系统终极架构：六层自治模型

> 从"多角色协作"到"目标驱动、自我校准的自治控制系统"

## 0. 本文档的定位

这份方案融合了三个设计思路：

1. **L4 自治内核方案**（Sensor→Planner→Actuator→Learner 四组件）
2. **契约驱动能力编排**（TaskContract → CapabilityProfile → ContractMatcher）
3. **五层自治模型**（目标→策略→能力→执行→治理）

最终形成的是一个**六层自治模型**，比上述任何单一方案都更完整。

---

## 1. 为什么五层还不够

五层模型（目标→策略→能力→执行→治理）解决了"系统应该长什么样"，但有一个盲区：

**谁来维护这五层的参数？**

- 目标层的"风险容忍度"设多少合适？
- 策略层的"连续失败 5 次触发熔断"这个 5 是怎么来的？
- 能力层的"claude_code 对多文件任务成功率 91%"这个数字谁更新？
- 执行层的"最优并发数 12"从哪得出？

如果答案是"人手动配"，那系统的自治程度就被锁死在"人设规则 → 系统执行"，永远是 L3。

真正的 L4 需要系统**自己知道自己的能力边界在哪，自己校准自己的参数**。

这就是第六层：**元认知层（Meta-Cognition Layer）**。

---

## 2. 六层自治模型总览

```
┌─────────────────────────────────────────────────────────┐
│ Layer 6: 元认知层 (Meta-Cognition)                      │
│  系统对自身行为的认知、评估、校准                          │
│  "我做得怎么样？我的参数对不对？我该怎么改进？"             │
└────────────────────────┬────────────────────────────────┘
                         │ 反馈 ↑↓ 校准
┌────────────────────────▼────────────────────────────────┐
│ Layer 5: 目标层 (Objective)                             │
│  项目目标 · 预算 · 时限 · 风险容忍度 · SLA               │
│  "完成什么？不能碰什么？花多少？什么时候停？"               │
└────────────────────────┬────────────────────────────────┘
                         │ 约束 ↓
┌────────────────────────▼────────────────────────────────┐
│ Layer 4: 策略层 (Policy)                                │
│  准入控制 · 风险闸门 · 决策分级 · 人工节点                 │
│  "允不允许做？自动还是人工？做完怎么验？失败怎么收？"        │
└────────────────────────┬────────────────────────────────┘
                         │ 规则 ↓
┌────────────────────────▼────────────────────────────────┐
│ Layer 3: 能力层 (Capability)                            │
│  合同化能力 · 输入/输出契约 · 副作用 · 降级路径            │
│  "需要什么能力？谁能满足？代价多大？失败了怎么降级？"        │
└────────────────────────┬────────────────────────────────┘
                         │ 匹配 ↓
┌────────────────────────▼────────────────────────────────┐
│ Layer 2: 执行层 (Execution)                             │
│  执行器 · 模型 · 工具 · 沙箱 · 资源                      │
│  "用什么工具？跑在哪？消耗多少？"                          │
└────────────────────────┬────────────────────────────────┘
                         │ 结果 ↓
┌────────────────────────▼────────────────────────────────┐
│ Layer 1: 治理视图层 (Governance View)                   │
│  角色 · 权限 · 审计 · UI · 人工接管                      │
│  "谁能看？谁能管？怎么审计？人怎么介入？"                   │
└─────────────────────────────────────────────────────────┘
```

**每一层只和相邻层交互，不跨层直接调用。**

---

## 3. Layer 6：元认知层 — 系统的自我意识

### 3.1 为什么需要这一层

五层模型的每一层都有参数需要设定：

| 层 | 需要设定的参数 | 没有元认知层时 | 有元认知层时 |
|----|---------------|--------------|-------------|
| 目标层 | 风险容忍度、预算分配 | 人拍脑袋 | 从历史项目的成功/失败中推导 |
| 策略层 | 熔断阈值、闸门条件 | 人写规则 | 从误报/漏报率自动校准 |
| 能力层 | 能力评分、匹配权重 | 人配权重 | 从执行结果自动学习 |
| 执行层 | 并发度、超时时间 | 经验值 | 从运行指标自动寻优 |

**元认知层的职责：观察系统自身的行为，评估效果，校准参数。**

它不做业务决策，它做的是**关于决策的决策**：
- "上次那个策略效果好不好？"
- "这个阈值是不是太敏感了？"
- "那个能力评分和实际表现差多少？"

### 3.2 元认知三要素

```
┌──────────────────────────────────────────┐
│            Meta-Cognition                │
│                                          │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │ Observer  │  │ Assessor │  │ Tuner  │ │
│  │  观测器   │→│  评估器  │→│ 校准器 │ │
│  │          │  │          │  │        │ │
│  │ 采集指标 │  │ 评估效果 │  │ 调整参 │ │
│  │ 记录决策 │  │ 识别偏差 │  │ 数阈值 │ │
│  │ 追踪结果 │  │ 计算信度 │  │ 生成建 │ │
│  │          │  │          │  │ 议    │ │
│  └──────────┘  └──────────┘  └────────┘ │
│       ↑                          │       │
│       └──────────────────────────┘       │
│              反馈环                       │
└──────────────────────────────────────────┘
```

#### Observer（观测器）

不是 Sensor（感知器是 Layer 4 策略层的组件，感知的是"业务态势"）。
Observer 观测的是**系统自身的行为**：

```go
// 观测记录
type ObservationRecord struct {
    // 谁做了什么决策
    DecisionID     int64
    DecisionType   string  // 策略匹配 / 能力匹配 / 闸门判定 / ...
    
    // 决策时的上下文
    InputSnapshot  g.Map   // 决策输入的快照
    
    // 决策结果
    Output         g.Map   // 决策输出
    Confidence     float64 // 决策时的置信度
    
    // 实际效果（延迟回填）
    Outcome        string  // success / failure / neutral
    EffectScore    float64 // -1 ~ 1
    
    // 是否被人工覆盖
    HumanOverride  bool
    OverrideReason string
}
```

**关键设计：每次人工覆盖（approve/reject/手动调整）都是一个高价值的学习信号。**

如果系统决策 A 级自动执行，人工没干预 → 信号弱（可能人没看）。
如果系统决策 B 级，人批准了 → 信号中（人认可）。
如果系统决策 B 级，人驳回了 → **信号强**（系统判断错了）。

#### Assessor（评估器）

定期评估系统各层参数的准确性：

```go
// 评估维度
type AssessmentResult struct {
    // 策略层评估
    PolicyAccuracy     float64 // 策略匹配后动作的成功率
    GateFalsePositive  float64 // 闸门误报率（不该拦的拦了）
    GateFalseNegative  float64 // 闸门漏报率（该拦的没拦）
    HumanOverrideRate  float64 // 人工覆盖率（越低越好）
    
    // 能力层评估
    MatchAccuracy      float64 // 能力匹配后任务的成功率
    CostEfficiency     float64 // 成本效率（成功率/成本）
    
    // 执行层评估
    ConcurrencyOptimal bool    // 当前并发度是否最优
    TimeoutAccuracy    float64 // 超时设定的准确率
    
    // 偏差检测
    Drifts             []Drift // 检测到的参数偏差
}

type Drift struct {
    Parameter    string  // 哪个参数
    CurrentValue float64 // 当前值
    OptimalValue float64 // 建议值
    Confidence   float64 // 建议的置信度
    Evidence     string  // 依据
}
```

#### Tuner（校准器）

**不直接修改参数**，而是生成校准建议：

```go
type TuneRecommendation struct {
    Parameter     string      // 参数路径
    CurrentValue  interface{} // 当前值
    SuggestedValue interface{} // 建议值
    Reasoning     string      // 理由
    Confidence    float64     // 置信度
    AutoApplicable bool       // 是否可自动应用
    RiskLevel     string      // low / medium / high
}
```

**校准的安全边界：**

| 参数类型 | 自动校准 | 需人工确认 |
|----------|---------|-----------|
| 能力评分（成功率/耗时） | ✅ 自动更新 | — |
| 匹配权重（w1-w4） | ✅ 微调（±10%） | 大幅调整（±30%+） |
| 闸门阈值 | ✅ 放宽（降低误报） | 收紧（可能漏报） |
| 并发度 | ✅ 降低（保守） | 提高（可能过载） |
| 决策级别 | — | ✅ 始终需人工确认 |
| 目标层参数 | — | ✅ 始终需人工确认 |

**铁律：元认知层可以自动让系统变得更保守，但让系统变得更激进必须人工确认。**

### 3.3 为什么这比 Learner 更好

L4 方案里的 Learner 是"学习执行器成功率"的单维度学习器。
元认知层是**对整个系统行为的全维度反思**：

| 维度 | Learner（L4） | Meta-Cognition（L6） |
|------|-------------|---------------------|
| 学什么 | 执行器成功率 | 策略准确率、闸门误报率、匹配精度、人工覆盖率... |
| 从哪学 | 任务完成/失败 | 任务结果 + 人工覆盖 + 偏差检测 + 态势变化 |
| 学了干嘛 | 调匹配权重 | 校准所有五层的参数 |
| 安全边界 | 最小样本量 + 置信度 | 保守自动、激进人工 + 不可逆操作禁止自动 |

---

## 4. Layer 5：目标层 — 系统的方向盘

### 4.1 项目目标模型

```go
// ProjectObjective 项目目标 — 整个自治系统的方向盘。
// 所有下层决策都必须在这个边界内运行。
type ProjectObjective struct {
    // ===== 交付目标 =====
    DeliveryGoal       string     // 交付物描述
    QualityFloor       float64    // 最低质量标准 0-1（低于此值不允许完成）
    
    // ===== 预算约束 =====
    TokenBudget        int64      // Token 总预算（0=不限）
    TimeBudgetHours    float64    // 时间预算（小时，0=不限）
    CostBudgetCents    int64      // 金钱预算（分，0=不限）
    
    // ===== 风险约束 =====
    RiskTolerance      string     // "conservative" / "balanced" / "aggressive"
    MaxAutoRetries     int        // 自动重试上限
    MaxAutoReworks     int        // 自动返工上限
    MaxAutoReplans     int        // 自动重规划上限
    
    // ===== SLA 约束 =====
    DeadlineAt         *time.Time // 截止时间（可选）
    MaxStallMinutes    int        // 最大允许停滞时间
    
    // ===== 自治边界 =====
    AutonomyLevel      string     // "manual" / "supervised" / "autonomous"
    // manual：所有动作需人工确认
    // supervised：A 级自动，B/C 级人工
    // autonomous：A/B 级自动，仅 C 级人工
}
```

### 4.2 目标如何约束下层

```
目标层设定：RiskTolerance = "conservative"
    ↓
策略层响应：
  - 熔断阈值从连续5次失败降为3次
  - B 级动作不自动执行，需人工确认
  - 闸门检查更严格
    ↓
能力层响应：
  - 匹配权重偏向 HistoryScore（选历史表现好的）
  - CostScore 权重降低（不追求省钱，追求稳）
    ↓
执行层响应：
  - 并发度从 20 降到 10
  - 超时时间放宽 50%
```

```
目标层设定：RiskTolerance = "aggressive", TokenBudget = 500000
    ↓
策略层响应：
  - B 级动作可自动执行
  - 闸门阈值放宽
    ↓
能力层响应：
  - 匹配权重偏向 CostScore（省 Token）
  - 优先选轻量执行器
    ↓
执行层响应：
  - 并发度提到 30
  - 超时时间收紧
```

### 4.3 目标不是一次设定的

目标可以在运行中动态调整：

```
项目启动时：
  TokenBudget = 1000000, RiskTolerance = "balanced"
  
运行到 60%，已消耗 800000 Token：
  元认知层检测到：Token 消耗速度超出预期
  ↓
  自动建议调整：CostSensitivity → "high"
  ↓
  能力层响应：切换到更轻量的执行器
  ↓
  人工确认（目标层参数变更始终需要人工确认）
```

---

## 5. Layer 4：策略层 — 系统的交通法规

### 5.1 核心理念

策略层不只是"选谁干"，更重要的是**"允不允许干"**。

每个动作在执行前必须经过策略层的准入控制：

```
动作请求 → 准入控制 → 风险闸门 → 决策分级 → 执行/停驻
```

### 5.2 准入控制（Admission Control）

```go
// AdmissionCheck 准入检查 — 在任何自治动作执行前调用。
// 返回 Deny 时动作被阻止，返回 Allow 时动作可继续。
type AdmissionResult struct {
    Allowed     bool
    DenyReason  string    // 拒绝原因
    Conditions  []string  // 附加条件（允许但有条件）
}

func (p *PolicyLayer) CheckAdmission(
    ctx context.Context,
    action string,           // 动作类型
    objective *ProjectObjective, // 当前目标约束
    situation *Situation,    // 当前态势
) *AdmissionResult {
    // 1. 预算检查
    if objective.TokenBudget > 0 && situation.Resource.TokensConsumed > objective.TokenBudget*90/100 {
        return &AdmissionResult{Allowed: false, DenyReason: "Token 预算已消耗 90%"}
    }
    
    // 2. 时间检查
    if objective.DeadlineAt != nil && time.Now().After(*objective.DeadlineAt) {
        return &AdmissionResult{Allowed: false, DenyReason: "已超过截止时间"}
    }
    
    // 3. 重试/返工/重规划次数检查
    if action == "retry_task" && situation.Health.RetryCount >= objective.MaxAutoRetries {
        return &AdmissionResult{Allowed: false, DenyReason: "自动重试次数已达上限"}
    }
    
    // 4. 自治级别检查
    // ... 根据 AutonomyLevel 和动作的 DecisionLevel 判断
    
    return &AdmissionResult{Allowed: true}
}
```

### 5.3 策略层和现有 DecisionCenter 的关系

**不是替代，是分层：**

```
现有 DecisionCenter.Decide()
  ├─ PolicyEngine.Match()    → 归入 Layer 4 策略层
  ├─ RiskGate.Check()        → 归入 Layer 4 策略层
  ├─ ActionDispatcher.Execute() → 归入 Layer 2 执行层
  └─ Human Checkpoint        → 归入 Layer 1 治理视图层

新增：
  ├─ AdmissionControl        → Layer 4 策略层（在 Match 之前）
  ├─ ObjectiveCheck          → Layer 5 目标层（在 Admission 之前）
  └─ MetaObservation         → Layer 6 元认知层（在 Execute 之后）
```

**调用链变为：**

```
事件触发
  ↓
Layer 6: Observer.Record(decision_input)
  ↓
Layer 5: ObjectiveCheck(objective, situation) → 超出目标边界？→ 拒绝
  ↓
Layer 4: AdmissionControl(action, objective, situation) → 准入失败？→ 拒绝
  ↓
Layer 4: PolicyEngine.Match(trigger) → 决策级别 + 动作类型
  ↓
Layer 4: RiskGate.Check(request) → 闸门阻断？→ 降级
  ↓
Layer 3: ContractMatcher.Match(contract) → 最优执行者
  ↓
Layer 2: Executor.Execute(action) → 结果
  ↓
Layer 6: Observer.RecordOutcome(result)
  ↓
Layer 6: Assessor.EvaluateIfNeeded() → 偏差检测
  ↓
Layer 6: Tuner.SuggestIfNeeded() → 校准建议
```

---

## 6. Layer 3：能力层 — 合同化能力

### 6.1 能力合同（CapabilityContract）

比 TaskContract 更完整，不只描述"需要什么"，还描述"怎么验收"和"失败了怎么办"：

```go
// CapabilityContract 能力合同 — 能力的完整规格说明。
// 定义了能力的输入输出、副作用、验收标准、降级路径。
type CapabilityContract struct {
    // 身份
    CapabilityCode  string  // 能力标识
    CapabilityName  string  // 可读名称
    
    // 输入契约
    InputSchema     g.Map   // 输入参数 Schema
    RequiredContext string  // 所需上下文范围
    RequiredTools   []string // 所需工具/执行器
    RequiredPerms   []string // 所需权限
    
    // 输出契约
    OutputSchema    g.Map   // 输出结构 Schema
    OutputVerify    string  // 验收方式: "rule" / "llm_judge" / "human" / "auto"
    
    // 副作用声明
    SideEffects     []SideEffect
    Reversible      bool    // 是否可回滚
    
    // 运行约束
    MaxDuration     int     // 最大执行时间(秒)
    MaxTokens       int64   // 最大 Token 消耗
    CostLevel       string  // "low" / "medium" / "high" / "very_high"
    
    // 自治级别
    AutoLevel       string  // "A" / "B" / "C"
    
    // 降级路径
    DegradationPath []DegradationStep
}

type SideEffect struct {
    Type        string  // "file_modify" / "db_change" / "api_call" / "resource_lock"
    Scope       string  // 影响范围
    Severity    string  // "low" / "medium" / "high"
}

type DegradationStep struct {
    Condition    string  // 触发条件: "timeout" / "failure" / "cost_exceeded"
    FallbackTo   string  // 降级到哪个能力
    FallbackParams g.Map // 降级参数
}
```

### 6.2 预定义能力合同示例

```go
var CodeEditContract = &CapabilityContract{
    CapabilityCode: "code_edit",
    CapabilityName: "代码编辑",
    InputSchema: g.Map{
        "file_paths":   "[]string, 要编辑的文件路径",
        "instructions": "string, 编辑指令",
        "context":      "string, 上下文信息",
    },
    OutputSchema: g.Map{
        "modified_files": "[]string, 实际修改的文件",
        "summary":        "string, 变更摘要",
    },
    OutputVerify: "rule",  // 规则验证：文件是否真的被修改
    RequiredTools: []string{"aider", "claude_code", "codex_cli"},  // 可选执行器
    SideEffects: []SideEffect{
        {Type: "file_modify", Scope: "project_workdir", Severity: "medium"},
    },
    Reversible:     true,    // 可 git revert
    AutoLevel:      "A",     // 可自动执行
    CostLevel:      "medium",
    DegradationPath: []DegradationStep{
        {Condition: "failure", FallbackTo: "multi_file_edit", FallbackParams: g.Map{"escalate": true}},
        {Condition: "cost_exceeded", FallbackTo: "code_edit", FallbackParams: g.Map{"model": "lite"}},
    },
}

var FailureAnalysisContract = &CapabilityContract{
    CapabilityCode: "failure_analysis",
    CapabilityName: "故障分析",
    InputSchema: g.Map{
        "error_message":  "string, 错误信息",
        "task_context":   "string, 任务上下文",
        "system_state":   "map, 系统状态快照",
    },
    OutputSchema: g.Map{
        "diagnosis":       "string, 根因分析",
        "severity":        "string, 严重程度",
        "recovery_options": "[]map, 恢复方案列表",
        "recommended":     "string, 推荐方案",
    },
    OutputVerify:  "llm_judge",  // LLM 验证分析质量
    RequiredTools: []string{"chat"},
    RequiredContext: "global",
    SideEffects:   nil,          // 纯分析，无副作用
    Reversible:    true,         // 分析结果可忽略
    AutoLevel:     "B",          // 需人工确认恢复方案
    CostLevel:     "low",
    DegradationPath: []DegradationStep{
        {Condition: "timeout", FallbackTo: "failure_analysis", FallbackParams: g.Map{"model": "lite", "simplified": true}},
    },
}
```

### 6.3 能力合同的注册和发现

能力合同存储在 `mvp_capability_contract` 表中，支持：
- **预定义合同**：系统内置的标准能力合同
- **自定义合同**：用户为特定项目类型配置的专属能力合同
- **版本管理**：合同可以演进，旧版本自动废弃

```sql
CREATE TABLE IF NOT EXISTS `mvp_capability_contract` (
  `id`                bigint unsigned NOT NULL COMMENT '雪花ID',
  `capability_code`   varchar(64)     NOT NULL COMMENT '能力标识',
  `capability_name`   varchar(128)    NOT NULL COMMENT '能力名称',
  `version`           int             NOT NULL DEFAULT 1 COMMENT '版本号',
  `input_schema`      json            DEFAULT NULL,
  `output_schema`     json            DEFAULT NULL,
  `output_verify`     varchar(16)     DEFAULT 'rule',
  `required_tools`    json            DEFAULT NULL COMMENT '可选执行器列表',
  `required_context`  varchar(16)     DEFAULT 'local',
  `required_perms`    json            DEFAULT NULL,
  `side_effects`      json            DEFAULT NULL,
  `reversible`        tinyint         DEFAULT 1,
  `auto_level`        char(1)         DEFAULT 'B',
  `cost_level`        varchar(16)     DEFAULT 'medium',
  `max_duration`      int             DEFAULT 600,
  `max_tokens`        bigint          DEFAULT 0,
  `degradation_path`  json            DEFAULT NULL,
  `project_family`    varchar(32)     DEFAULT NULL COMMENT '适用项目族(NULL=通用)',
  `enabled`           tinyint         DEFAULT 1,
  `created_by`        bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`           bigint unsigned NOT NULL DEFAULT 0,
  `created_at`        datetime        DEFAULT NULL,
  `updated_at`        datetime        DEFAULT NULL,
  `deleted_at`        datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code_version` (`capability_code`, `version`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
  COMMENT='能力合同';
```

---

## 7. Layer 2：执行层 — 执行器市场

### 7.1 执行器市场化

当前执行器是硬编码注册的。执行器市场化意味着：

1. **执行器自描述**：每个执行器声明自己的能力、成本、限制
2. **动态发现**：新执行器上线后自动被能力层发现
3. **竞争择优**：多个执行器都能满足契约时，按评分竞争
4. **淘汰机制**：持续表现差的执行器被降权

```go
// ExecutorDescriptor 执行器自描述
type ExecutorDescriptor struct {
    Name          string            // "aider" / "claude_code" / ...
    Version       string            // 版本号
    Capabilities  []string          // 提供的能力
    CapScores     map[string]float64 // 能力评分
    CostPerToken  float64           // 每 Token 成本
    AvgLatencyMs  int64             // 平均延迟
    MaxConcurrent int               // 最大并发
    RequiresSandbox bool
    SupportsStreaming bool
    HealthStatus  string            // "healthy" / "degraded" / "offline"
}
```

### 7.2 执行器健康检查

```go
// 执行器市场定期健康检查
func (m *ExecutorMarket) HealthCheck(ctx context.Context) {
    for _, exec := range m.executors {
        // 1. 最近 10 次执行的成功率
        recentSuccess := m.getRecentSuccessRate(exec.Name, 10)
        
        // 2. 最近平均延迟
        recentLatency := m.getRecentAvgLatency(exec.Name, 10)
        
        // 3. 判定状态
        if recentSuccess < 0.3 {
            exec.HealthStatus = "degraded"  // 降级
        } else if recentSuccess < 0.1 {
            exec.HealthStatus = "offline"   // 离线（不再分配新任务）
        } else {
            exec.HealthStatus = "healthy"
        }
    }
}
```

---

## 8. Layer 1：治理视图层 — 给人看的界面

### 8.1 角色的新定位

角色从"调度主键"降级为"治理视图"：

| 用途 | 说明 |
|------|------|
| **权限边界** | 角色决定能看到什么数据、能操作什么功能 |
| **审计视角** | "这个动作是以什么角色的身份执行的" |
| **成本归集** | "架构师消耗了多少 Token，实施员消耗了多少" |
| **UI 展示** | "用不同颜色/图标展示不同角色的任务" |
| **人工接管** | "我要以架构师身份介入这个任务" |

### 8.2 角色不再是调度主键

```
现在（L3.5）：
  InstantiateFromBlueprint() {
    roleType := blueprint["role_type"]
    executionMode := projectRole[roleType]["execution_mode"]  // 角色→执行器
    // ... 创建任务
  }

未来（L4+）：
  InstantiateFromBlueprint() {
    contract := InferContract(blueprint)   // 从蓝图推断契约
    matchResult := ContractMatcher.Match(contract)  // 契约→最优匹配
    executionMode := matchResult.Profile.ExecutionMode  // 匹配结果→执行器
    roleType := matchResult.Profile.RoleType  // 匹配结果→角色（治理视图用）
    // ... 创建任务
  }
```

**调度主键从 `role_type` 变成了 `contract`。**

角色仍然存在，但它是匹配结果的一个属性，不是匹配的输入。

---

## 9. Build Line + Operate Line 双主线

### 9.1 架构定位

```
                    ┌──────────────┐
                    │  目标层       │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ↓                         ↓
    ┌─────────────────┐      ┌──────────────────┐
    │   Build Line    │      │   Operate Line   │
    │   研发主线       │      │   运维主线        │
    │                 │      │                  │
    │  设计→审核→执行  │      │  感知→诊断→恢复   │
    │  →验收→完成      │      │  →回滚→报告       │
    │                 │      │                  │
    │  角色：          │      │  角色：           │
    │  Architect      │      │  Operator        │
    │  Implementer    │      │  Coordinator     │
    │  Auditor        │      │                  │
    └────────┬────────┘      └────────┬─────────┘
             │                        │
             └────────────┬───────────┘
                          │
                   ┌──────▼───────┐
                   │  能力层（共享）│
                   └──────────────┘
```

### 9.2 两条主线的协作

```
Build Line 正常运行时：
  Operate Line 在后台巡检（Sensor.Patrol）
  ↓
Build Line 出异常：
  ① Sensor 检测到异常信号
  ② Operate Line 接管：Operator 做故障诊断
  ③ 策略层决定恢复方案
  ④ Operate Line 执行恢复（重试/回滚/切换执行器）
  ⑤ 恢复成功 → Build Line 继续
  ⑥ 恢复失败 → 升级到人工（C 级）
  ↓
Build Line 完成：
  Operate Line 生成报告 + 归档指标
```

### 9.3 Operate Line 的能力合同

| 能力 | AutoLevel | 输入 | 输出 | 副作用 |
|------|-----------|------|------|--------|
| `failure_diagnosis` | B | 错误信息+任务上下文 | 根因分析+恢复方案 | 无 |
| `auto_retry` | A | 失败任务ID | 重试结果 | 任务状态变更 |
| `executor_switch` | A/B | 任务ID+目标执行器 | 切换结果 | 任务配置变更 |
| `rollback_plan` | B | 变更列表 | 回滚计划 | 无 |
| `rollback_execute` | C | 回滚计划 | 回滚结果 | 文件回滚+状态回滚 |
| `concurrency_adjust` | A | 目标并发数 | 调整结果 | 配置变更 |
| `circuit_break` | A | 熔断原因 | 暂停结果 | 项目暂停 |
| `replan_evaluate` | B | 熔断原因+失败任务 | 重规划建议 | 无 |
| `replan_execute` | C | 重规划方案 | 执行结果 | 任务创建/删除/修改 |

---

## 10. 六层如何整合现有代码

### 10.1 现有代码到六层的映射

| 现有组件 | 归入层 | 角色 |
|----------|--------|------|
| `DecisionCenter.Decide()` | Layer 4+5 | 策略层入口 + 目标检查 |
| `PolicyEngine.Match()` | Layer 4 | 规则匹配 |
| `RiskGate.Check()` | Layer 4 | 闸门检查 |
| `ActionDispatcher.Execute()` | Layer 2 | 执行层 |
| `EngineSelector.Recommend()` | Layer 3 | 能力匹配（被 ContractMatcher 替代） |
| `CircuitBreaker.Check()` | Layer 4 | 策略层的熔断子组件 |
| `Replanner.Evaluate()` | Layer 3+4 | 能力合同 replan_evaluate + 策略层 B 级 |
| `RiskAssessor.Assess()` | Layer 3 | 能力合同 risk_assessment |
| `Reporter.GenerateReport()` | Layer 1 | 治理视图层 |
| `Sensor.Perceive()` | Layer 4 | 策略层的感知子组件 |
| `Learner.Learn()` | Layer 6 | 元认知层的 Observer+Tuner |
| `mvp_project_role` | Layer 1 | 治理视图层（角色配置） |
| `mvp_policy_rule` | Layer 4 | 策略层（规则库） |
| `mvp_risk_gate_rule` | Layer 4 | 策略层（闸门库） |
| `mvp_decision_action` | Layer 4+6 | 策略层（记录）+ 元认知层（观测数据） |

### 10.2 不需要改的

以下代码完全不需要修改：

- `scheduler/domain_task_scheduler.go` — 调度器逻辑不变
- `watchdog/watchdog.go` — 已通过回调接入
- `stage/accept/service.go` — 已通过触发器接入
- `stage/rework/service.go` — 已通过触发器接入
- `collab/` — 飞书通知通过事件订阅
- `engine/` — 旧引擎继续为 V1 项目服务

### 10.3 需要改的入口点

只有一个核心入口点需要修改：

**`orchestrator/registry.go` 的 Init() 函数**

```go
// 现有初始化（不变）
policyEngine := autonomy.NewPolicyEngine(...)
riskGate := autonomy.NewRiskGate(...)
actionDispatcher := autonomy.NewActionDispatcher(...)
decisionCenter = autonomy.NewDecisionCenter(...)

// 新增：六层组件初始化
if engine.GetConfigInt(ctx, "workflow.autonomy.six_layer_enabled", ..., 0) == 1 {
    // Layer 6: 元认知
    observer := meta.NewObserver(decisionActionRepo)
    assessor := meta.NewAssessor(observer)
    tuner := meta.NewTuner(assessor)
    
    // Layer 5: 目标层
    objectiveManager := objective.NewManager()
    
    // Layer 3: 能力层
    contractMatcher := capability.NewContractMatcher(profileRegistry, learner)
    
    // 注入到 DecisionCenter
    decisionCenter.SetObjectiveManager(objectiveManager)
    decisionCenter.SetContractMatcher(contractMatcher)
    decisionCenter.SetMetaObserver(observer)
    
    // 启动元认知评估定时任务
    go assessor.StartPeriodicAssessment(ctx, 3600) // 每小时评估一次
}
```

---

## 11. 实施路线图

### 阶段 A：地基（目标层 + Operator 角色）

最快见效，不依赖其他层。

1. ProjectObjective 模型 + DB 表
2. 准入控制（AdmissionControl）接入 DecisionCenter
3. Operator 角色上线（预设 + system_prompt + 故障分析能力）
4. 修改 watchdog 的 failure_analysis 任务使用 Operator 角色

**交付：** 项目有预算/时限约束，故障分析有专业角色。

### 阶段 B：能力合同化

5. CapabilityContract 模型 + DB 表
6. 预定义 10 个核心能力合同
7. TaskContract 自动推断
8. ContractMatcher 匹配器（仅记日志，不改变执行器选择）

**交付：** 系统能描述"任务需要什么"和"执行者能提供什么"。

### 阶段 C：匹配器接入执行流程

9. ContractMatcher 正式接入任务实例化
10. 失败后契约升级 + 重新匹配
11. 执行器市场化（自描述 + 健康检查）
12. 架构师契约输出（扩展任务 JSON 格式）

**交付：** 动态执行器选择，失败自动升级。

### 阶段 D：元认知层

13. Observer（记录所有决策+结果+人工覆盖）
14. Assessor（定期评估策略准确率、闸门误报率等）
15. Tuner（生成参数校准建议）
16. 自动校准保守方向，激进方向需人工确认

**交付：** 系统能自我评估、自我优化。

### 阶段 E：Operate Line 完整化

17. Operate Line 完整能力合同（9 个能力）
18. Coordinator 升级为控制面代理
19. Build Line + Operate Line 协作协议
20. 运维报告和人工接管界面

**交付：** 双主线架构完成。

---

## 12. 灰度策略

```
Layer 0：全部关闭                     现有 L3.5 行为
  ↓
Layer 1：six_layer_enabled=0          不初始化六层组件
  ↓
开启 → objective_enabled=1            启用目标层准入控制
  ↓
开启 → contract_matching_enabled=1    启用能力匹配（仅记日志）
  ↓
开启 → contract_matching_active=1     匹配结果接入执行流程
  ↓
开启 → meta_cognition_enabled=1       启用元认知观测
  ↓
开启 → meta_auto_tune_enabled=1       启用自动校准（仅保守方向）
  ↓
开启 → operate_line_enabled=1         启用运维主线
  ↓
全部开启                              完整六层自治
```

**每一层可以独立回滚，不影响其他层。**

---

## 13. 为什么这是"更稳"的方案

### 13.1 和纯 Agent 方案比

| 维度 | 纯多 Agent | 六层自治 |
|------|-----------|---------|
| 决策依据 | LLM 自由发挥 | 规则+闸门+目标约束 |
| 可预测性 | 低（同输入不同输出） | 高（确定性规则优先） |
| 成本控制 | 难（Agent 随意调用 API） | 强（准入控制+预算闸门） |
| 人工介入 | 不知道什么时候该介入 | 明确（A 自动/B 确认/C 必须） |
| 故障恢复 | Agent 可能把恢复变成新故障 | 合同化（副作用声明+降级路径） |
| 可审计性 | 日志不可靠 | 全链路观测记录 |

### 13.2 和传统规则引擎比

| 维度 | 纯规则引擎 | 六层自治 |
|------|-----------|---------|
| 灵活性 | 低（规则写死） | 高（LLM 辅助+学习优化） |
| 适应性 | 无（参数手动调） | 有（元认知层自动校准） |
| 新场景支持 | 改代码 | 配置能力合同+标签 |
| 智能程度 | 低 | 中高（LLM 在边界内辅助） |

### 13.3 核心稳定性设计

1. **规则高于 LLM**：硬决策用规则，软决策 LLM 辅助
2. **保守自动、激进人工**：系统可以自动变保守，变激进必须人工确认
3. **合同化执行**：每个能力的输入/输出/副作用/降级路径都是明确的
4. **全链路可观测**：每个决策都有记录，每个结果都有评估
5. **逐层灰度**：任何一层出问题，关掉它，退回上一层

---

## 14. 一句话总结

**目标驱动 · 策略约束 · 能力合同化 · 执行器市场化 · 角色仅治理视图 · 元认知自校准。**

这不是"更多角色"，而是"更少角色依赖"。
这不是"更多 LLM"，而是"更少 LLM 自由度"。
这不是"更多功能"，而是"更多约束和边界"。

真正稳定的自治系统，不是让 AI 更自由，而是让 AI 在更精确的边界内运行。
