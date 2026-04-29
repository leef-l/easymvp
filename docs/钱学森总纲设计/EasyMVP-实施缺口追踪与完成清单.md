# EasyMVP 实施缺口追踪与完成清单

> 创建时间：2026-04-27
> 追踪范围：所有文档要求但代码未实现的缺口
> 更新规则：每完成一项，本文档和对应原文同步更新

---

## 缺口总览

| 编号 | 优先级 | 缺口描述 | 原文位置 | 代码位置 | 状态 |
|---|---|---|---|---|---|
| P0-01 | P0 | `BrainContractEnvelope` 缺少 `NormalizedStatus` 字段 | `合同JSON-Schema终稿.md` §1.1, §2 | `apps/core/internal/model/braincontracts/common.go` | ✅ 已完成 2026-04-27 |
| P0-02 | P0 | 运行时 DTO 族缺失：`VerificationResult`、`FaultSummary`、`RuntimeEscalation`、`EvidenceSummary` | `IO合同及升级规则.md` §5, §7, §9 | `apps/core/internal/model/braincontracts/` | ✅ 已完成 2026-04-27 |
| P0-03 | P0 | `RepairPlanDraft` DTO 字段不完整 | `输入输出契约.md` §6.2 | `apps/core/internal/model/braincontracts/repair_design.go` | ✅ 已完成 2026-04-27 |
| P0-04 | P0 | `brain.json` manifest capabilities 不完整 | `专精大脑接入计划.md` §10 | `brain/brains/easymvp/brain.json` | ✅ 已完成 2026-04-27 |
| P0-05 | P0 | `architect_chat` 合同扩散，需收敛为内部辅助 | `输入输出契约.md` §2 | `brain/brains/easymvp/` | ✅ 已完成 2026-04-27 |
| P0-11 | P0 | 统一错误域与状态枚举 | `IO合同及升级规则.md` §8 | `apps/core/internal/model/braincontracts/contract_status.go` | ✅ 已完成 2026-04-27 |
| P1-06 | P1 | 验证合同字段未结构化：`VerificationContract` DTO | `Verification-Contract统一设计.md` §5.1 | `apps/core/internal/model/braincontracts/verification_contract.go` | ✅ 已完成 2026-04-27 |
| P1-08 | P1 | 故障回路未硬约束：`fault → repair_design → reworking` | `专项实施清单.md` P0-03 | `apps/core/internal/service/acceptance_support_completion.go` | ✅ 已完成 2026-04-27 |
| P1-09 | P1 | `completion_adjudication` 未成为 completed 前强制裁决 | `专项实施清单.md` P0-02 | `apps/core/internal/service/acceptance_support_completion.go` | ✅ 已完成 2026-04-27 |
| P1-07 | P1 | 四基础专精大脑未接入 EasyMVP | `总方案.md` §4.1.1 | `apps/core/internal/service/` | ✅ 已完成 2026-04-27 |
| P1-10 | P1 | 页面未显示 contract gap / escalation reason | `专项实施清单.md` P1-01/02 | `apps/desktop/` | ✅ 已完成 2026-04-27 |
| P1-11 | P1 | 升级规则与状态枚举部分（高级场景） | `IO合同及升级规则.md` §8 | `apps/core/internal/model/braincontracts/` | ✅ 已完成 2026-04-27 |
| P2-12 | P2 | 高配验证环境 `high_spec_remote` 未接入 | `Verification-Contract统一设计.md` §4.2 | `apps/core/internal/service/` | ✅ 已完成 2026-04-27 |
| P2-13 | P2 | 6 阶段主导矩阵未代码化 | `四基础专精大脑阶段调用矩阵.md` §3 | `apps/core/internal/service/` | ✅ 已完成 2026-04-27 |

---

## P0-01：`BrainContractEnvelope` 缺少 `NormalizedStatus`

### 原文位置
- `docs/EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md` §1.1, §2

### 文档要求
```go
type BrainContractEnvelope struct {
    SchemaVersion    int              `json:"schema_version"`
    ResultKind       string           `json:"result_kind"`
    ResultVersion    int              `json:"result_version"`
    SourceRefs       []BrainSourceRef `json:"source_refs"`
    DecisionSummary  string           `json:"decision_summary"`
    TraceID          string           `json:"trace_id,omitempty"`
    DeploymentMode   string           `json:"deployment_mode,omitempty"`
    BrainEndpoint    string           `json:"brain_endpoint,omitempty"`
    NormalizedStatus string           `json:"normalized_status,omitempty"`  // <-- 缺失
    ResultJSON       json.RawMessage  `json:"result_json"`
}
```

`normalized_status` enum: `["success", "failure", "unsupported_or_denied"]`

### 代码位置
- `apps/core/internal/model/braincontracts/common.go`

### 完成标准
- [x] `BrainContractEnvelope` 添加 `NormalizedStatus` 字段
- [x] `easymvp_brain.go` 中的 `executeTypedContract` 在解析 envelope 后对 `NormalizedStatus` 做硬校验
- [x] 如果 `NormalizedStatus == "unsupported_or_denied"`，必须返回明确错误，不能伪装成成功
- [x] 所有 brain handler 返回标准 envelope 格式（含 `normalized_status`）
- [x] 更新原文档 §1.1 标记为"已实现"

---

## P0-02：运行时 DTO 族缺失

### 原文位置
- `docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md` §5, §7, §9
- `docs/钱学森总纲设计/easymvp-brain-输入输出契约.md` §9

### 文档要求的对象关系
```text
PlanDraft
  -> PlanReviewResult
      -> CompiledPlan
          -> CompiledTask
              -> RunResult
              -> DeliveryResult
              -> VerificationResult      <-- 已补齐
              -> EvidenceSummary         <-- 已补齐
                  -> CompletionVerdict

FailureContext + FaultSummary              <-- 已补齐
  -> RepairPlanDraft
```

### 需要补全的 DTO
1. `VerificationResult` — 验证结果归一化对象 ✅
2. `FaultSummary` — 故障摘要归一化对象 ✅
3. `RuntimeEscalation` — 运行时升级对象 ✅
4. `EvidenceSummary` — 证据汇总对象 ✅

### 代码位置
- `apps/core/internal/model/braincontracts/`

### 完成标准
- [x] 创建 `verification_result.go` 定义 `VerificationResult` DTO
- [x] 创建 `fault_summary.go` 定义 `FaultSummary` DTO
- [x] 创建 `runtime_escalation.go` 定义 `RuntimeEscalation` DTO
- [x] 创建 `evidence_summary.go` 定义 `EvidenceSummary` DTO
- [x] 更新原文档 §9 标记为"已实现"

---

## P0-03：`RepairPlanDraft` DTO 字段不完整

### 原文位置
- `docs/钱学森总纲设计/easymvp-brain-输入输出契约.md` §6.2

### 文档要求字段
- `repair_plan_draft_id` ✅
- `repair_plan_json` ✅
- `repair_reasoning_summary` ✅
- `replaced_constraints` ✅
- `reason_class` ✅ 已补齐
- `repair_strategy` ✅ 已补齐
- `updated_tasks` ✅ 已补齐
- `verification_adjustments` ✅ 已补齐
- `delivery_adjustments` ✅ 已补齐
- `human_checkpoint_required` ✅ 已补齐

### 代码位置
- `apps/core/internal/model/braincontracts/repair_design.go`

### 完成标准
- [x] `RepairDesignResult` 补充缺失字段
- [x] `repair_design` handler 的 LLM prompt 要求生成新字段
- [x] fallback 数据包含新字段
- [x] 更新原文档 §6.2 标记为"已实现"

---

## P0-04：`brain.json` manifest capabilities 不完整

### 原文位置
- `docs/EasyMVP-V3-专精大脑接入计划.md` §10

### 当前 capabilities（5个）
```json
[
  "easymvp.plan_review",
  "easymvp.plan_compile",
  "easymvp.plan_redesign",
  "easymvp.repair_design",
  "easymvp.acceptance_mapping",
  "easymvp.completion_adjudication",
  "easymvp.workspace_explanation"
]
```

### 处理说明
- 移除了 `easymvp.architect_chat`（见 P0-05）
- 补齐了 `easymvp.acceptance_mapping`、`easymvp.completion_adjudication`、`easymvp.workspace_explanation`

### 代码位置
- `brain/brains/easymvp/brain.json`

### 完成标准
- [x] 补全 3 个缺失 capability
- [x] 移除 `easymvp.architect_chat`
- [x] 更新原文档 §10 标记为"已实现"

---

## P0-05：`architect_chat` 合同扩散收敛

### 原文位置
- `docs/钱学森总纲设计/easymvp-brain-输入输出契约.md` §2

### 文档要求
> 第一版收敛为 5 类核心契约：plan_review, plan_compile, repair_design, acceptance_mapping, completion_adjudication

`architect_chat` 不在 5 类核心契约中，但代码中已作为 handler 实现。

### 处理方案
- 保留 `architect_chat` 作为**内部辅助能力**，但：
  1. 从 `brain.json` capabilities 中移除
  2. 在 `handler.go` 中明确标注为内部使用（不纳入正式合同清单）
  3. 文档中明确标注为"内部辅助，不纳入正式合同"

### 代码位置
- `brain/brains/easymvp/handler.go`
- `brain/brains/easymvp/brain.json`

### 完成标准
- [x] `handler.go` 中 `architect_chat` 保留但标注内部使用
- [x] `brain.json` capabilities 移除 `easymvp.architect_chat`
- [x] 更新原文档 §2 标记为"已收敛"

---

## P0-11：统一错误域与状态枚举

### 原文位置
- `docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md` §8

### 文档要求的升级/错误类型
- `unsupported_capability`
- `policy_denied`
- `verification_conflict`
- `environment_unavailable`
- `manual_review_required`
- `fault_loop_detected`

### 代码位置
- `apps/core/internal/model/braincontracts/contract_status.go`

### 完成标准
- [x] 新建 `contract_status.go` 定义统一状态枚举和错误域
- [x] 包含 `NormalizedStatus` enum 和 `EscalationType` enum
- [x] 包含 `IsValidNormalizedStatus` 和 `IsValidEscalationType` 校验函数
- [x] 更新原文档 §8 标记为"基础已实现"

---

## P1-06：结构化 `VerificationContract` DTO

### 原文位置
- `docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md` §5.1

### 文档要求字段
```go
type VerificationContract struct {
    ContractID          string   `json:"contract_id"`
    RequiredChecks      []string `json:"required_checks"`
    RequiredEvidence    []string `json:"required_evidence"`
    PreferredChannel    string   `json:"preferred_channel"`
    FallbackChannels    []string `json:"fallback_channels"`
    BlockingRules       []string `json:"blocking_rules"`
    WarningTolerance    int      `json:"warning_tolerance"`
    ManualReviewRules   []string `json:"manual_review_rules"`
}
```

### 当前代码
`CompiledTaskItem.VerificationContract` 是 `json.RawMessage`（无结构）

### 处理方案
- 新建 `VerificationContract` DTO
- 保留 `CompiledTaskItem.VerificationContract` 为 `json.RawMessage` 做向后兼容
- 提供 `ParseVerificationContract` 辅助函数供调用方转换

### 代码位置
- `apps/core/internal/model/braincontracts/verification_contract.go`

### 完成标准
- [x] 新建 `verification_contract.go` 定义 `VerificationContract` DTO
- [x] 提供 `ParseVerificationContract` 转换辅助函数
- [x] `plan_compile` handler 的 LLM prompt 要求生成结构化验证合同字段
- [x] 更新原文档 §5.1 标记为"已实现"

---

## P1-08：故障回路硬约束

### 原文位置
- `docs/钱学森总纲设计/EasyMVP-专项实施清单.md` P0-03

### 文档要求
> `fault → easymvp-brain.repair_design → reworking` 是唯一路径

### 当前问题
`shouldCreateRepairDraftAfterAdjudication` 只在 `FinalStatus == "failed"` 时触发，未覆盖 `Decision == "rework"` 的情况。

### 修改内容
- 扩展 `shouldCreateRepairDraftAfterAdjudication`：当 `Decision == "rework"` 或 `FinalStatus == "failed"` 时都触发 repair draft 创建
- 添加明确注释：失败后直接重试被显式禁止，必须经过 repair design

### 代码位置
- `apps/core/internal/service/acceptance_support_completion.go`

### 完成标准
- [x] `shouldCreateRepairDraftAfterAdjudication` 在 `Decision == "rework"` 时返回 true
- [x] 函数注释明确说明 `fault → repair_design → reworking` 是唯一路径
- [x] 更新原文档 P0-03 标记为"已实现"

---

## P1-09：`completion_adjudication` 强制裁决

### 原文位置
- `docs/钱学森总纲设计/EasyMVP-专项实施清单.md` P0-02

### 文档要求
> 任意任务进入最终完成前都必须经过 `completion_adjudication`

### 当前问题
数据库表 `completion_verdicts` 已有字段，但规则未在代码中显式文档化。

### 修改内容
- 在 `updateProjectFromAdjudication` 添加明确注释：这是设置 `completed` 的唯一授权路径
- 验证 `maybeAutoAdjudicateAcceptanceRun` 已集成 adjudication 到运行时自动流程
- `UpdateProject` API 不允许直接修改 `status`，防止绕过

### 代码位置
- `apps/core/internal/service/acceptance_support_completion.go`
- `apps/core/internal/service/runtime.go`

### 完成标准
- [x] `updateProjectFromAdjudication` 添加注释说明是 `completed` 的唯一授权路径
- [x] 验证运行时自动 adjudication 已集成
- [x] 更新原文档 P0-02 标记为"已实现"

---

## 后续迭代项（P1/P2 中未在本次完成的）

| 编号 | 说明 | 阻塞原因 |
|---|---|---|
| — | LLM JSON 输出约束（`response_format: json_object`） | 需底层 kernel 支持 |
| — | `RunResult`/`DeliveryResult` 独立持久化层 | 需要新表设计，当前通过 `run_event_index` 间接存储 |
| — | 返工振荡计数（`oscillation_control`） | 需结合历史 run 统计，依赖健康度基础 |

**注：原后续迭代项 P1-07 / P1-10 / P1-11 / P2-12 / P2-13 已在 2026-04-27 全部完成。**

---

## 完成记录

| 日期 | 完成项 | 更新人 |
|---|---|---|
| 2026-04-27 | P0-01 / P0-02 / P0-03 / P0-04 / P0-05 / P0-11 / P1-06 / P1-08 / P1-09 | Code Agent |
| 2026-04-27 | P1-07 / P1-10 / P1-11 / P2-12 / P2-13（原后续迭代项全部清零） | Code Agent |
| 2026-04-27 | 补充：项目健康度统计 API（`ProjectHealthMetrics`）、RuntimeEscalation DB 字段扩展、`channel_unavailable` 自动降级 | Code Agent |
