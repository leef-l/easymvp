# EasyMVP 角色体系演进：契约驱动能力编排设计方案

> 从"固定角色协作"到"契约驱动能力编排"的终极架构

## 1. 为什么"角色层+能力层+策略层"还不够

### 1.1 三层模型的局限

"角色层+能力层+策略层"是经典的能力编排思路，但它仍然是**供给侧驱动**的：
- 先定义"我有什么能力"
- 再看"任务需要哪个能力"
- 最后"把任务路由到能力"

问题在于：
1. **能力爆炸**：随着业务场景增多，能力列表会无限膨胀（代码修改、测试、安全审查、数据库变更、运维、文档...）
2. **组合困境**：一个任务可能需要多个能力的组合，穷举组合不现实
3. **粒度失配**："需求分析"是一个能力还是"理解用户意图+识别约束+拆分子问题"三个能力？
4. **静态注册**：能力列表是预定义的，新能力必须写代码注册

### 1.2 更好的方向：需求侧驱动

不要问"我有什么能力"，要问**"任务需要什么契约"**。

```
传统思路：  角色 → 能力 → 任务
我们的思路：任务 → 契约 → 最优满足者
```

---

## 2. 核心模型：契约驱动能力编排

### 2.1 三个基本概念

| 概念 | 定义 | 类比 |
|------|------|------|
| **TaskContract（任务契约）** | 任务对执行者的需求描述 | 招标书 |
| **CapabilityProfile（能力画像）** | 执行者能提供的能力集合 | 投标方资质 |
| **MatchResult（匹配结果）** | 契约和画像的匹配评估 | 评标结果 |

### 2.2 TaskContract（任务契约）

每个任务不再指定"谁来做"，而是描述"需要什么"：

```go
// TaskContract 任务契约 —— 描述任务对执行者的需求。
// 由架构师在任务拆分时生成，也可由系统自动推断。
type TaskContract struct {
    // ===== 能力需求 =====
    RequiredCapabilities  []string   // 必须能力: ["code_edit", "file_create"]
    PreferredCapabilities []string   // 优选能力: ["multi_file_edit", "test_run"]
    
    // ===== 质量需求 =====
    AccuracyRequirement   string     // "high" / "medium" / "low"
    CreativityRequirement string     // "high" / "medium" / "low"
    
    // ===== 上下文需求 =====
    ContextScope          string     // "local"(仅前置任务) / "batch"(批次级) / "global"(全项目)
    RequiresFileAccess    bool       // 是否需要读写文件
    RequiresSandbox       bool       // 是否需要沙箱环境
    RequiresHumanReview   bool       // 是否需要人工审核
    
    // ===== 资源约束 =====
    MaxTokenBudget        int64      // Token 预算上限（0=不限）
    MaxDurationSeconds    int        // 时间预算上限
    CostSensitivity       string     // "high"(省钱优先) / "low"(效果优先)
    
    // ===== 安全约束 =====  
    RequiresIsolation     bool       // 是否需要隔离执行
    SensitivityLevel      string     // "public" / "internal" / "confidential"
    
    // ===== 领域标签 =====
    DomainTags            []string   // ["frontend", "database", "security", "devops"]
}
```

**关键设计：契约是声明式的，不指定实现方式。**

"我需要能编辑多个文件、需要高精度、需要全局上下文" —— 至于用 Aider 还是 Claude Code，由系统决定。

### 2.3 CapabilityProfile（能力画像）

每个"执行实例"（角色+执行器+模型的组合）有一个能力画像：

```go
// CapabilityProfile 能力画像 —— 描述一个执行实例能提供什么。
// 不是静态注册的，而是由角色、执行器、模型三者组合推导出来的。
type CapabilityProfile struct {
    // 身份
    RoleType       string   // architect / implementer / auditor / operator / coordinator
    ExecutionMode  string   // chat / aider / claude_code / openhands / codex_cli / gemini_cli
    ModelID        int64    // AI 模型
    RoleLevel      string   // lite / pro / max
    
    // 提供的能力
    Capabilities   []string // ["code_edit", "multi_file_edit", "file_create", "test_run", ...]
    
    // 能力强度评分（0-1）
    CapabilityScores map[string]float64 // {"code_edit": 0.95, "test_run": 0.3}
    
    // 质量特征
    AccuracyScore    float64 // 精度评分 0-1
    CreativityScore  float64 // 创意评分 0-1
    SpeedScore       float64 // 速度评分 0-1
    
    // 资源特征
    AvgTokenCost     int64   // 平均 Token 消耗
    AvgDurationSec   int     // 平均耗时
    CostPerTask      float64 // 每任务成本（归一化）
    
    // 上下文特征
    MaxContextScope  string  // 最大可访问的上下文范围
    SupportsFile     bool    // 支持文件读写
    SupportsSandbox  bool    // 支持沙箱
    
    // 历史表现（由 Learner 填充）
    HistoricalSuccessRate float64 // 历史成功率
    SampleCount           int     // 样本数量
}
```

**能力画像不是手动配置的**，而是从三个维度自动推导：

```
CapabilityProfile = f(RoleType, ExecutionMode, Model) + Learner修正

角色贡献：     architect → {requirement_analysis, task_decomposition, failure_analysis}
执行器贡献：   claude_code → {code_edit, multi_file_edit, global_understanding}
模型贡献：     claude-opus → {high_accuracy: 0.95, high_creativity: 0.8}
学习修正：     历史数据 → {success_rate: 0.91, avg_duration: 72s}
```

### 2.4 ContractMatcher（契约匹配器）

```go
// ContractMatcher 契约匹配器 —— 找到最优满足契约的执行实例。
type ContractMatcher struct {
    profileRegistry *ProfileRegistry  // 所有可用的能力画像
    learner         *Learner          // 历史学习数据（可选）
}

// Match 匹配契约，返回排序后的候选列表。
func (m *ContractMatcher) Match(
    ctx context.Context, 
    contract *TaskContract,
    projectID int64,
) []MatchResult {
    candidates := m.profileRegistry.GetAll(ctx, projectID)
    var results []MatchResult
    
    for _, profile := range candidates {
        result := m.evaluate(contract, profile)
        if result.Eligible {  // 必须能力全覆盖
            results = append(results, result)
        }
    }
    
    // 按综合评分排序
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    
    return results
}

type MatchResult struct {
    Profile         *CapabilityProfile
    Eligible        bool     // 是否满足所有必须能力
    Score           float64  // 综合评分 0-1
    
    // 评分明细（可解释性）
    CapabilityScore float64  // 能力覆盖度
    QualityScore    float64  // 质量匹配度
    CostScore       float64  // 成本效益
    HistoryScore    float64  // 历史表现
    Reasoning       string   // 可读的匹配理由
}
```

**评分公式：**

```
Score = w1 * CapabilityScore 
      + w2 * QualityScore 
      + w3 * CostScore 
      + w4 * HistoryScore

其中：
  CapabilityScore = (必须能力覆盖率 * 0.7 + 优选能力覆盖率 * 0.3)
  QualityScore = 1 - |contract.AccuracyReq - profile.AccuracyScore|
  CostScore = contract.CostSensitivity == "high" 
              ? (1 - profile.CostPerTask / maxCost) 
              : profile.AccuracyScore
  HistoryScore = profile.HistoricalSuccessRate * min(1, profile.SampleCount / 30)

权重根据 CostSensitivity 动态调整：
  省钱优先：w1=0.3, w2=0.2, w3=0.35, w4=0.15
  效果优先：w1=0.3, w2=0.35, w3=0.1, w4=0.25
  均衡：    w1=0.3, w2=0.25, w3=0.2, w4=0.25
```

---

## 3. 角色体系重构

### 3.1 从 4 角色到 5+N 模型

**稳定角色（5 个）：** 代表职责边界，不会频繁变化。

| 角色 | 职责边界 | 核心契约 | 绝对禁区 |
|------|----------|----------|----------|
| **Architect** | 理解需求、设计方案、拆分任务、故障诊断 | 高精度 + 全局上下文 + 需求分析能力 | 不执行代码修改 |
| **Implementer** | 按规格执行代码/内容创作 | 文件修改能力 + 指定精度 | 不做需求判断 |
| **Auditor** | 独立质量评估 | 只读 + 分析能力 | 不修改代码、不做业务决策 |
| **Operator** | 故障恢复、变更风险评估、环境管理 | 系统状态感知 + 恢复策略 | 不做业务设计 |
| **Coordinator** | 调度编排、冲突裁决、进度控制 | 全局视图 + 规则执行 | 不做内容生产、不做质量判断 |

**动态角色（N 个）：** 通过能力画像组合产生，不需要预注册。

```
例如：
  "安全审计员" = Auditor角色 + security_analysis能力 + 安全领域system_prompt
  "数据库专家" = Implementer角色 + database_migration能力 + DBA领域system_prompt  
  "测试工程师" = Implementer角色 + test_execution能力 + 沙箱执行器
  "技术文档员" = Implementer角色 + document_generation能力 + chat执行器
```

动态角色不需要新代码，只需要在 `mvp_role_preset` 中配置新的 `system_prompt` + `capability_tags`。

### 3.2 Operator 角色详细设计

这是当前系统最缺的角色。

**职责清单：**

| 场景 | Operator 做什么 | 现在谁做 | 问题 |
|------|-----------------|----------|------|
| 任务失败分析 | 判断失败模式 + 推荐恢复策略 | watchdog 硬编码规则 | 规则太粗，不理解错误语义 |
| 回滚决策 | 评估回滚影响范围 + 制定回滚计划 | 无 | 完全缺失 |
| 环境异常 | 检测环境问题 + 建议修复 | 无 | 完全缺失 |
| 变更风险评估 | 评估代码变更的风险等级 | RiskGate 静态规则 | 不理解变更内容 |
| 自动止损 | 在人工接管前执行保护性动作 | CircuitBreaker 熔断 | 只会暂停，不会智能止损 |
| 恢复路径选择 | 从多种恢复方案中选择最优 | AdaptiveRetryStrategy | 策略级别，缺少语义理解 |

**Operator 的契约模板：**

```go
var OperatorContract = &TaskContract{
    RequiredCapabilities:  []string{"failure_analysis", "risk_assessment"},
    PreferredCapabilities: []string{"rollback_planning", "environment_diagnosis"},
    AccuracyRequirement:   "high",
    ContextScope:          "global",      // 需要看到全局态势
    RequiresFileAccess:    true,          // 需要读日志和配置
    RequiresSandbox:       false,         // 不需要沙箱
    CostSensitivity:       "low",         // 效果优先
    SensitivityLevel:      "internal",
    DomainTags:            []string{"operations", "recovery"},
}
```

**Operator 的 system_prompt 核心要求：**

```
你是一个运维恢复专家。你的职责是：
1. 分析系统异常的根因（不是猜测，是基于证据推理）
2. 评估每个恢复方案的风险和代价
3. 推荐最优恢复路径
4. 在人工介入前执行保护性止损动作

你的边界：
- 你可以建议回滚、重试、切换执行器、暂停
- 你不能修改业务代码
- 你不能做需求判断
- 你不能替代人工做 C 级决策

你的输出必须是结构化 JSON：
{
  "diagnosis": "根因分析",
  "severity": "critical/high/medium/low",
  "recovery_options": [
    {"action": "...", "risk": "...", "cost": "...", "confidence": 0.8}
  ],
  "recommended_action": "...",
  "auto_stop_loss": "暂停受影响的批次" 或 null
}
```

### 3.3 Coordinator 角色升级

从"调度优化建议器"升级为**控制面代理（Control Plane Agent）**：

| 现在 | 升级后 |
|------|--------|
| 仅在审核阶段优化 batch_no | 持续监控调度状态 |
| 输出是建议（可忽略） | 输出是调度指令（强制执行，受策略约束） |
| 不参与运行时 | 参与运行时冲突裁决和超时接管 |

**Coordinator 新增职责：**

1. **运行时冲突裁决**：两个任务争同一资源时，Coordinator 决定谁先执行
2. **超时接管**：任务超时后，Coordinator 决定是等待、取消还是切换执行器
3. **批次动态调整**：根据运行态势（Sensor 输出）动态调整并发度和批次编排
4. **进度报告**：生成面向人类可读的项目进度摘要

**Coordinator 的硬边界（铁律）：**

```
Coordinator 可以做：
  ✓ 调整任务优先级和批次顺序
  ✓ 调整并发度
  ✓ 裁决资源冲突
  ✓ 触发任务超时
  ✓ 生成进度报告

Coordinator 不能做：
  ✗ 修改任务内容（那是 Architect 的事）
  ✗ 执行代码变更（那是 Implementer 的事）
  ✗ 做质量判断（那是 Auditor 的事）
  ✗ 做故障诊断（那是 Operator 的事）
  ✗ 做 C 级决策（那是人的事）
```

---

## 4. 能力注册表

### 4.1 预定义能力词表

不需要穷举所有能力。定义一个**核心能力词表**，足够覆盖 90% 场景：

```go
// 核心能力词表（可扩展，新增无需改代码）
const (
    // ===== 分析类 =====
    CapRequirementAnalysis = "requirement_analysis"   // 需求理解
    CapTaskDecomposition   = "task_decomposition"     // 任务拆分
    CapFailureAnalysis     = "failure_analysis"        // 故障分析
    CapRiskAssessment      = "risk_assessment"         // 风险评估
    CapCostEstimation      = "cost_estimation"         // 成本估算
    CapSecurityAnalysis    = "security_analysis"       // 安全分析
    
    // ===== 执行类 =====
    CapCodeEdit            = "code_edit"               // 代码编辑（单文件）
    CapMultiFileEdit       = "multi_file_edit"         // 多文件编辑
    CapFileCreate          = "file_create"             // 创建新文件
    CapTestExecution       = "test_execution"          // 执行测试
    CapDatabaseMigration   = "database_migration"      // 数据库变更
    CapDocumentGeneration  = "document_generation"     // 文档生成
    CapContentCreation     = "content_creation"        // 内容创作
    CapEnvironmentSetup    = "environment_setup"       // 环境配置
    
    // ===== 审查类 =====
    CapCodeReview          = "code_review"             // 代码审查
    CapQualityAssessment   = "quality_assessment"      // 质量评估
    CapComplianceCheck     = "compliance_check"        // 合规检查
    
    // ===== 运维类 =====
    CapRollbackPlanning    = "rollback_planning"       // 回滚规划
    CapEnvironmentDiagnosis = "environment_diagnosis"  // 环境诊断
    CapRecoveryExecution   = "recovery_execution"      // 恢复执行
    CapChangeAssessment    = "change_assessment"       // 变更评估
    
    // ===== 编排类 =====
    CapScheduleOptimization = "schedule_optimization"  // 调度优化
    CapConflictResolution  = "conflict_resolution"     // 冲突裁决
    CapProgressTracking    = "progress_tracking"       // 进度追踪
)
```

### 4.2 能力画像自动推导规则

```go
// 角色贡献的能力
var roleCapabilities = map[string][]string{
    "architect":    {"requirement_analysis", "task_decomposition", "failure_analysis", "cost_estimation"},
    "implementer":  {"code_edit", "file_create", "content_creation"},
    "auditor":      {"code_review", "quality_assessment", "security_analysis", "compliance_check"},
    "operator":     {"failure_analysis", "risk_assessment", "rollback_planning", "environment_diagnosis", "recovery_execution", "change_assessment"},
    "coordinator":  {"schedule_optimization", "conflict_resolution", "progress_tracking"},
}

// 执行器贡献的能力
var executorCapabilities = map[string][]string{
    "chat":         {"document_generation", "content_creation"},
    "aider":        {"code_edit", "file_create"},
    "claude_code":  {"code_edit", "multi_file_edit", "file_create", "test_execution"},
    "openhands":    {"code_edit", "multi_file_edit", "file_create", "test_execution", "environment_setup", "database_migration"},
    "codex_cli":    {"code_edit", "file_create", "test_execution"},
    "gemini_cli":   {"code_edit", "file_create", "content_creation"},
}

// 合并：Profile.Capabilities = union(roleCapabilities[role], executorCapabilities[executor])
```

### 4.3 自定义能力标签

除了预定义能力，支持在 `mvp_role_preset` 中配置自定义能力标签：

```sql
ALTER TABLE `mvp_role_preset`
  ADD COLUMN `capability_tags` json DEFAULT NULL 
  COMMENT '自定义能力标签(JSON数组)，覆盖默认推导';

ALTER TABLE `mvp_project_role`
  ADD COLUMN `capability_tags` json DEFAULT NULL 
  COMMENT '自定义能力标签(JSON数组)，覆盖默认推导';
```

示例：
```json
// 一个"安全审计员"预设
{
  "role_type": "auditor",
  "role_level": "pro",
  "system_prompt": "你是安全审计专家...",
  "capability_tags": ["security_analysis", "compliance_check", "vulnerability_detection"]
}
```

---

## 5. 契约生成：谁来写 TaskContract

### 5.1 三个来源

| 来源 | 时机 | 说明 |
|------|------|------|
| **架构师显式指定** | 任务拆分时 | 架构师在拆分任务时输出契约字段 |
| **系统自动推断** | 蓝图实例化时 | 根据任务描述、affected_resources、role_type 自动推断 |
| **运行时动态调整** | 任务失败重试时 | 根据失败模式升级契约需求 |

### 5.2 架构师输出格式扩展

当前架构师输出的任务 JSON：

```json
{
  "name": "实现用户登录模块",
  "description": "...",
  "role_type": "implementer",
  "role_level": "pro",
  "affected_resources": ["src/auth/login.go"],
  "batch_no": 2
}
```

扩展后（向后兼容）：

```json
{
  "name": "实现用户登录模块",
  "description": "...",
  "role_type": "implementer",
  "role_level": "pro",
  "affected_resources": ["src/auth/login.go"],
  "batch_no": 2,
  "contract": {
    "required_capabilities": ["code_edit", "file_create"],
    "preferred_capabilities": ["test_execution"],
    "accuracy_requirement": "high",
    "context_scope": "batch",
    "domain_tags": ["backend", "auth"],
    "cost_sensitivity": "low"
  }
}
```

**向后兼容：** `contract` 字段可选。不提供时系统自动推断。

### 5.3 自动推断规则

```go
func InferContract(task *TaskBlueprint) *TaskContract {
    contract := &TaskContract{
        ContextScope:    "local",
        CostSensitivity: "medium",
    }
    
    // 从 role_type 推断基础需求
    switch task.RoleType {
    case "implementer":
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "code_edit")
        contract.RequiresFileAccess = true
    case "auditor":
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "code_review")
    case "architect":
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "requirement_analysis")
        contract.ContextScope = "global"
    }
    
    // 从 affected_resources 推断
    resourceCount := len(task.AffectedResources)
    if resourceCount >= 5 {
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "multi_file_edit")
    }
    
    // 从描述关键词推断
    desc := strings.ToLower(task.Description)
    if containsAny(desc, "测试", "test", "验证") {
        contract.PreferredCapabilities = append(contract.PreferredCapabilities, "test_execution")
        contract.RequiresSandbox = true
    }
    if containsAny(desc, "数据库", "migration", "schema") {
        contract.RequiredCapabilities = append(contract.RequiredCapabilities, "database_migration")
    }
    if containsAny(desc, "安全", "security", "权限") {
        contract.PreferredCapabilities = append(contract.PreferredCapabilities, "security_analysis")
    }
    
    // 从 role_level 推断质量和上下文
    switch task.RoleLevel {
    case "max":
        contract.AccuracyRequirement = "high"
        contract.ContextScope = "global"
        contract.CostSensitivity = "low"
    case "pro":
        contract.AccuracyRequirement = "medium"
        contract.ContextScope = "batch"
    case "lite":
        contract.AccuracyRequirement = "low"
        contract.CostSensitivity = "high"
    }
    
    return contract
}
```

### 5.4 失败后契约升级

任务失败时，系统自动升级契约：

```go
func EscalateContract(contract *TaskContract, failurePattern string) *TaskContract {
    upgraded := contract.Clone()
    
    switch failurePattern {
    case PatternCapabilityMismatch:
        // 能力不足 → 提升精度需求 + 扩展上下文
        upgraded.AccuracyRequirement = "high"
        upgraded.ContextScope = "global"
        upgraded.CostSensitivity = "low"  // 不再省钱
        
    case PatternStructural:
        // 结构性错误 → 需要多文件能力 + 沙箱
        upgraded.RequiredCapabilities = appendIfMissing(
            upgraded.RequiredCapabilities, "multi_file_edit")
        upgraded.RequiresSandbox = true
        
    case PatternResourceConflict:
        // 资源冲突 → 需要隔离执行
        upgraded.RequiresIsolation = true
    }
    
    return upgraded
}
```

**这就是"执行器自动升级"的本质**：不是硬编码"aider→claude_code→openhands"的升级链，而是升级契约需求，让匹配器自然找到更强的满足者。

---

## 6. LLM 边界铁律

### 6.1 硬规则：LLM 绝对不能做的事

```go
// 这些决策必须由确定性规则执行，LLM 不得参与：
const (
    // 状态机转换
    RuleStateTransition      = "state_transition"       // 任务/项目状态迁移
    // 权限校验
    RulePermissionCheck      = "permission_check"       // 数据权限/操作权限
    // 风险闸门
    RuleRiskGateEvaluation   = "risk_gate_evaluation"   // 闸门表达式求值
    // 资源锁
    RuleResourceLocking      = "resource_locking"       // 资源锁定/释放
    // CAS 操作
    RuleCASOperation         = "cas_operation"          // 乐观锁状态转换
    // 成本阈值
    RuleCostThreshold        = "cost_threshold"         // 成本超限判断
    // 熔断条件
    RuleCircuitBreakCondition = "circuit_break_condition" // 熔断触发条件
)
```

### 6.2 LLM 可以做的事（受约束）

```go
// LLM 可参与但必须受规则约束的决策：
const (
    // 故障诊断（LLM 分析错误，但恢复策略由规则决定）
    LLMAssistedFailureDiagnosis = "failure_diagnosis"
    // 重规划建议（LLM 生成方案，但执行/放弃由策略决定）
    LLMAssistedReplanSuggestion = "replan_suggestion"
    // 质量评估（LLM 作为 Judge，但最终裁决由规则引擎汇总）
    LLMAssistedQualityJudge     = "quality_judge"
    // 任务描述增强（LLM 补充细节，但不改变契约）
    LLMAssistedDescriptionEnrich = "description_enrich"
    // 执行器推荐（LLM 精排候选，但最终选择由匹配器决定）
    LLMAssistedEngineRanking    = "engine_ranking"
)
```

### 6.3 执行边界矩阵

| 决策类型 | 规则引擎 | LLM | 人工 | 说明 |
|----------|---------|-----|------|------|
| 状态迁移 | **唯一决策者** | 不参与 | 不参与 | CAS + 状态机 |
| 权限校验 | **唯一决策者** | 不参与 | 不参与 | DataScope 五级 |
| 风险闸门 | **唯一决策者** | 不参与 | 闸门命中时确认 | 表达式求值 |
| 资源调度 | **主决策者** | 辅助（优化建议） | 冲突裁决 | Coordinator |
| 故障诊断 | 分类规则 | **分析诊断** | C级审批 | Operator |
| 质量评估 | 规则引擎基线 | **辅助Judge** | manual_review | Auditor |
| 重规划 | 触发条件 | **方案生成** | B/C级审批 | Architect |
| 执行器选择 | 契约匹配 | 可选精排 | 不参与 | ContractMatcher |
| 成本控制 | **阈值判断** | 不参与 | 超限审批 | CostGuard |
| 任务拆分 | 格式校验 | **核心生成** | 确认方案 | Architect |

---

## 7. 与 L4 自治内核的整合

### 7.1 契约驱动如何接入 Sensor → Planner → Actuator → Learner

```
Sensor.Perceive()
  ↓ Situation（含当前运行的所有任务契约+实际执行者画像）
  
Planner.Plan()
  ├─ [快速路径] PolicyEngine.Match() → 传统规则
  ├─ [契约路径] ContractMatcher.Match() → 为失败任务重新匹配
  └─ [策略路径] 各Strategy函数 → 综合态势决策
  ↓ ActionPlan（含 switch_executor + 新的 CapabilityProfile）

Actuator.Execute()
  ├─ 执行动作
  └─ ScheduleOutcomeEvaluation() → 延迟评估效果
  ↓ ActionResult + ActionOutcome

Learner.Learn()
  ├─ 更新 CapabilityProfile.HistoricalSuccessRate
  ├─ 更新 CapabilityProfile.AvgTokenCost / AvgDurationSec
  └─ 更新 ContractMatcher 的权重参数
```

### 7.2 具体整合点

| L4 组件 | 契约驱动增强 | 说明 |
|---------|-------------|------|
| **Sensor** | 态势中包含契约满足度 | "当前批次有 3 个任务的契约未被最优满足" |
| **EngineSelectionStrategy** | 用 ContractMatcher 替代 EngineSelector | 从"规则推荐"升级为"契约匹配" |
| **AdaptiveRetryStrategy** | 失败后升级契约，重新匹配 | 不是硬编码升级链，是需求驱动的重新匹配 |
| **Learner** | 学习对象从"执行器成功率"扩展为"画像匹配度" | "这类契约用这种画像最有效" |

---

## 8. 数据模型变更

### 8.1 新增表

```sql
-- 任务契约表
CREATE TABLE IF NOT EXISTS `mvp_task_contract` (
  `id`                    bigint unsigned NOT NULL COMMENT '雪花ID',
  `domain_task_id`        bigint unsigned NOT NULL COMMENT '关联领域任务',
  `blueprint_id`          bigint unsigned DEFAULT NULL COMMENT '关联蓝图（可选）',
  `required_capabilities` json            DEFAULT NULL COMMENT '必须能力(JSON数组)',
  `preferred_capabilities` json           DEFAULT NULL COMMENT '优选能力(JSON数组)',
  `accuracy_requirement`  varchar(16)     DEFAULT 'medium' COMMENT 'high/medium/low',
  `creativity_requirement` varchar(16)    DEFAULT 'medium',
  `context_scope`         varchar(16)     DEFAULT 'local' COMMENT 'local/batch/global',
  `requires_file_access`  tinyint         DEFAULT 0,
  `requires_sandbox`      tinyint         DEFAULT 0,
  `requires_isolation`    tinyint         DEFAULT 0,
  `requires_human_review` tinyint         DEFAULT 0,
  `max_token_budget`      bigint          DEFAULT 0,
  `max_duration_seconds`  int             DEFAULT 0,
  `cost_sensitivity`      varchar(16)     DEFAULT 'medium' COMMENT 'high/medium/low',
  `sensitivity_level`     varchar(16)     DEFAULT 'internal',
  `domain_tags`           json            DEFAULT NULL COMMENT '领域标签(JSON数组)',
  `source`                varchar(16)     DEFAULT 'inferred' COMMENT '来源: architect/inferred/escalated',
  `version`               int             NOT NULL DEFAULT 1 COMMENT '契约版本(失败升级时递增)',
  `created_by`            bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`               bigint unsigned NOT NULL DEFAULT 0,
  `created_at`            datetime        DEFAULT NULL,
  `updated_at`            datetime        DEFAULT NULL,
  `deleted_at`            datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_domain_task` (`domain_task_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
  COMMENT='任务契约';

-- 能力画像快照表（记录每次匹配使用的画像）
CREATE TABLE IF NOT EXISTS `mvp_capability_match_log` (
  `id`                    bigint unsigned NOT NULL COMMENT '雪花ID',
  `domain_task_id`        bigint unsigned NOT NULL,
  `contract_id`           bigint unsigned NOT NULL,
  `contract_version`      int             NOT NULL DEFAULT 1,
  `matched_role_type`     varchar(32)     NOT NULL,
  `matched_engine`        varchar(32)     NOT NULL,
  `matched_model_id`      bigint unsigned DEFAULT NULL,
  `match_score`           decimal(5,4)    NOT NULL,
  `capability_score`      decimal(5,4)    DEFAULT NULL,
  `quality_score`         decimal(5,4)    DEFAULT NULL,
  `cost_score`            decimal(5,4)    DEFAULT NULL,
  `history_score`         decimal(5,4)    DEFAULT NULL,
  `reasoning`             text            DEFAULT NULL,
  `outcome`               varchar(16)     DEFAULT NULL COMMENT 'success/failure/pending',
  `created_by`            bigint unsigned NOT NULL DEFAULT 0,
  `dept_id`               bigint unsigned NOT NULL DEFAULT 0,
  `created_at`            datetime        DEFAULT NULL,
  `deleted_at`            datetime        DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_task` (`domain_task_id`),
  KEY `idx_contract` (`contract_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
  COMMENT='能力匹配日志';
```

### 8.2 修改现有表

```sql
-- mvp_role_preset 新增能力标签
ALTER TABLE `mvp_role_preset`
  ADD COLUMN `capability_tags` json DEFAULT NULL 
  COMMENT '能力标签(JSON数组)' AFTER `system_prompt`;

-- mvp_project_role 新增能力标签
ALTER TABLE `mvp_project_role`
  ADD COLUMN `capability_tags` json DEFAULT NULL 
  COMMENT '能力标签(JSON数组)' AFTER `system_prompt`;

-- mvp_task_blueprint 新增契约字段
ALTER TABLE `mvp_task_blueprint`
  ADD COLUMN `contract_json` json DEFAULT NULL 
  COMMENT '任务契约(JSON)' AFTER `description`;

-- mvp_domain_task 新增契约和匹配字段
ALTER TABLE `mvp_domain_task`
  ADD COLUMN `contract_id` bigint unsigned DEFAULT NULL 
  COMMENT '关联契约ID' AFTER `execution_mode`,
  ADD COLUMN `match_score` decimal(5,4) DEFAULT NULL 
  COMMENT '匹配评分' AFTER `contract_id`;
```

---

## 9. 文件清单

### 9.1 新建文件

| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `autonomy/contract.go` | TaskContract 模型 + 自动推断 + 契约升级 | 250 |
| `autonomy/capability_profile.go` | CapabilityProfile 模型 + 自动推导 | 200 |
| `autonomy/contract_matcher.go` | ContractMatcher 匹配器 + 评分公式 | 300 |
| `autonomy/profile_registry.go` | 能力画像注册表 | 150 |
| `repo/task_contract_repo.go` | 契约仓储 | 80 |
| `repo/capability_match_log_repo.go` | 匹配日志仓储 | 80 |

### 9.2 修改文件

| 文件 | 改动 |
|------|------|
| `autonomy/model.go` | 新增能力词表常量、Operator 角色常量 |
| `autonomy/strategy_engine_selection.go` | 用 ContractMatcher 替代 EngineSelector |
| `autonomy/strategy_adaptive_retry.go` | 失败后升级契约重新匹配 |
| `domain/task/task_service.go` | 实例化时生成契约 + 调用匹配器 |
| `consts/task.go` | 新增 RoleTypeOperator 常量 |
| `orchestrator/registry.go` | 初始化 ContractMatcher + 注册 Operator 动作回调 |

---

## 10. 实施路径

### Phase 0：Operator 角色上线（独立于契约系统）

**最快见效的一步。** 不需要契约系统，只需要：

1. `consts/task.go` 新增 `RoleTypeOperator = "operator"`
2. `mvp_role_preset` 新增 Operator 预设（system_prompt + model_id）
3. 修改 watchdog 的 `SetReplanFn`：失败分析任务改用 Operator 角色
4. 修改 `failure_analysis` 的 prompt：用 Operator 的结构化诊断格式

**灰度：** 新项目才启用 Operator，老项目不影响。

### Phase 1：契约模型 + 自动推断

1. 实现 `contract.go` — TaskContract 模型
2. 实现 `InferContract()` — 从蓝图自动推断契约
3. DDL：`mvp_task_contract` 表
4. 修改 `task_service.go` — 实例化时生成契约并持久化

**灰度：** 纯数据写入，不影响执行器选择。

### Phase 2：能力画像 + 匹配器

1. 实现 `capability_profile.go` — 画像模型 + 自动推导
2. 实现 `profile_registry.go` — 从 mvp_project_role 加载画像
3. 实现 `contract_matcher.go` — 匹配算法
4. DDL：`mvp_capability_match_log` 表

**灰度：** 匹配结果仅记日志，不改变实际执行器选择。

### Phase 3：匹配器接入执行流程

1. 修改 `task_service.go` — 实例化时用匹配器选择执行器
2. 修改 `strategy_engine_selection.go` — 用契约匹配替代规则推荐
3. 修改 `strategy_adaptive_retry.go` — 失败后升级契约

**灰度：** `workflow.autonomy.contract_matching_enabled=0` 关闭时用原逻辑。

### Phase 4：架构师契约输出

1. 修改架构师 system_prompt — 引导输出 contract 字段
2. 修改 task_parser — 解析 contract JSON
3. 修改蓝图表 — 持久化 contract_json

**灰度：** 架构师不输出 contract 时，系统自动推断。

### Phase 5：Learner 反馈闭环

1. 匹配日志写入 outcome（success/failure）
2. Learner 从日志中学习画像评分
3. 画像的 HistoricalSuccessRate 持续更新
4. 匹配器权重自适应调整

---

## 11. 与"三层模型"的对比

| 维度 | 三层模型（角色+能力+策略） | 契约驱动编排 |
|------|--------------------------|-------------|
| **驱动方式** | 供给侧："我有什么能力" | 需求侧："任务需要什么" |
| **新增能力** | 需要在能力层注册新能力类 | 在预设中加 capability_tags |
| **匹配方式** | 按能力名精确路由 | 多维评分 + 最优匹配 |
| **失败恢复** | 切换到另一个能力 | 升级契约需求，自动找到更强满足者 |
| **学习闭环** | 能力级统计 | 契约-画像匹配级统计（更精确） |
| **可解释性** | "用了能力 X" | "因为任务需要 A+B+C，画像 Y 评分最高因为..." |
| **扩展性** | 需要定义新能力类 | 组合现有标签即可表达新能力 |

**契约驱动本质上包含了三层模型**：
- 角色层 = CapabilityProfile 的 RoleType 维度
- 能力层 = CapabilityProfile 的 Capabilities 维度
- 策略层 = ContractMatcher 的评分公式 + Planner 的策略函数

但它多了一个关键的东西：**需求端的形式化描述（TaskContract）**。这让匹配变成了双向的，而不是单向的"任务→能力"查找。

---

## 12. 长期愿景

### 12.1 契约市场

当系统积累了足够的匹配日志后，可以：
- 自动生成"最佳实践契约模板"
- 按项目类型推荐契约配置
- 新项目创建时自动继承同类项目的成功契约

### 12.2 能力热插拔

新执行器上线时：
1. 在 `executorCapabilities` 中注册能力标签
2. 自动生成 CapabilityProfile
3. ContractMatcher 自动将其纳入候选
4. 新执行器的 `HistoricalSuccessRate` 从 0.5 起步，逐步学习

**不需要修改任何业务代码。**

### 12.3 跨项目契约迁移

同一个 `project_family` 下的项目可以共享契约匹配经验：
- A 项目发现"数据库迁移任务用 openhands 成功率 95%"
- B 项目的数据库迁移任务自动获得这个先验知识

### 12.4 人工契约覆盖

管理员可以在后台：
- 查看所有匹配日志
- 锁定某类任务的执行器选择（覆盖匹配器）
- 调整评分权重
- 禁用某个能力画像

---

## 13. 关键设计决策总结

| 决策 | 选择 | 理由 |
|------|------|------|
| 角色是否取消 | **保留稳定角色** | 角色表达职责边界，是人类理解系统的锚点 |
| 能力如何定义 | **标签组合而非类型层次** | 组合优于继承，标签可自由扩展 |
| 匹配算法 | **多维加权评分** | 可解释、可调参、可学习 |
| 契约来源 | **架构师+自动推断+失败升级** | 三级来源，渐进增强 |
| LLM 边界 | **铁律矩阵** | 明确哪些决策规则做、哪些 LLM 辅助、哪些人做 |
| 新角色 Operator | **独立上线** | 不依赖契约系统，最快补齐运维能力短板 |
| 向后兼容 | **全部灰度** | 默认关闭，逐步开启，任何时候可回退 |
