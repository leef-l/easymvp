# easymvp-brain 输入输出契约

> 更新时间：2026-04-20  
> 上位文档：[easymvp-brain-职责与边界定义.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md)  
> 参考文档：[EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md](/www/wwwroot/project/easymvp/docs/EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md)

---

## 1. 文档目的

这份文档不重写全量 schema，而是给出**当前总方案口径下的最小稳定契约**。

目标是回答：

1. `easymvp-brain` 有哪些核心输入输出
2. 每种输入输出各自服务什么业务动作
3. 这些对象在工作流里如何串起来

---

## 2. 契约总览

第一版收敛为 5 类核心契约：

1. `plan_review`
2. `plan_compile`
3. `repair_design`
4. `acceptance_mapping`
5. `completion_adjudication`

主链路：

```text
PlanDraft
  -> PlanReviewResult
  -> CompiledPlan
  -> DomainTask / Run / Evidence / Verification
  -> CompletionVerdict

FailureContext
  -> RepairPlanDraft
```

---

## 3. 统一输入原则

所有输入都必须满足：

1. 已被 EasyMVP 归一化
2. 可序列化
3. 不直接依赖原始工具输出格式
4. 不要求知道具体内置脑工具名

这意味着：

- `browser`、`verifier`、`fault` 的结果，先由 EasyMVP 归一化
- `easymvp-brain` 只接收领域语义输入

---

## 4. 契约一：`plan_review`

## 4.1 输入

- `plan_draft_id`
- `plan_draft_version`
- `plan_draft_json`
- `project_category`
- `category_profile_version`
- `category_profile_json`
- `project_context_json`
- 可选 `artifact_refs`

## 4.2 输出

- `PlanReviewResult`

最少应包含：

- `review_result_id`
- `review_version`
- `decision`
  - `approved`
  - `approved_with_advisory`
  - `rejected`
- `compile_allowed`
- `blocking_issues`
- `advisory_issues`
- `rewrite_hints`

## 4.3 用途

它用于回答：

- 这个计划能不能进入编译
- 哪些问题必须先修
- 哪些问题只作为 advisory

---

## 5. 契约二：`plan_compile`

## 5.1 输入

- `plan_draft_json`
- `plan_review_result_json`
- `category_profile_json`
- `role_context_json`
- 可选 `artifact_refs`

## 5.2 输出

- `CompiledPlan`
- `CompiledTask[]`

最少每个 `CompiledTask` 必须包含：

- `compiled_task_id`
- `name`
- `role_type`
- `brain_kind`
- `risk_level`
- `delivery_contract`
- `verification_contract`

## 5.3 用途

它用于回答：

- 草案如何变成正式执行计划
- 哪些任务交给哪个脑或执行器
- 每个任务要交什么货
- 每个任务怎么验

---

## 6. 契约三：`repair_design`

## 6.1 输入

- `failure_context_json`
- `original_contract_json`
- `fault_summary_json`
- `evidence_summary_json`
- `current_stage`

其中 `failure_context_json` 应至少概括：

- 哪一步失败
- 当前状态
- 重试历史
- 风险等级

## 6.2 输出

- `RepairPlanDraft`

最少应包含：

- `repair_plan_id`
- `reason_class`
- `repair_strategy`
- `updated_tasks`
- `verification_adjustments`
- `delivery_adjustments`
- `human_checkpoint_required`

## 6.3 用途

它用于回答：

- 这次失败应该怎么返工
- 是局部返工还是重编译
- 需不需要人工介入

---

## 7. 契约四：`acceptance_mapping`

## 7.1 输入

- `project_category`
- `category_profile_json`
- `current_artifact_summary_json`
- `available_verification_channels`

`available_verification_channels` 至少应包含：

- `high_spec_remote`
- `github_actions`
- `browser_evidence`
- `manual_review`

## 7.2 输出

- `AcceptanceProfile`
- `ProductionAcceptanceProfile`

最少应包含：

- `required_evidence`
- `required_checks`
- `warning_tolerance`
- `blocker_rules`
- `preferred_verification_channel`

## 7.3 用途

它用于回答：

- 验收需要什么
- 哪些检查是强制的
- 当前应优先走哪条验证通道

---

## 8. 契约五：`completion_adjudication`

## 8.1 输入

- `run_result_json`
- `delivery_result_json`
- `verification_result_json`
- `evidence_summary_json`
- `acceptance_profile_json`

## 8.2 输出

- `CompletionVerdict`

最少应包含：

- `executor_succeeded`
- `delivery_verified`
- `acceptance_passed`
- `completed`
- `decision`
  - `complete`
  - `rework`
  - `blocked`
  - `manual_checkpoint`
- `reason_summary`

## 8.3 用途

它用于回答：

- 这次到底算不算完成
- 是继续、返工、阻塞还是人工接管

---

## 9. 推荐对象关系

推荐的对象关系如下：

```text
PlanDraft
  -> PlanReviewResult
      -> CompiledPlan
          -> CompiledTask
              -> RunResult
              -> DeliveryResult
              -> VerificationResult
              -> EvidenceSummary
                  -> CompletionVerdict

FailureContext + FaultSummary
  -> RepairPlanDraft
```

---

## 10. 与四个基础专精大脑的关系

### 代码大脑

产出：

- `run_result_json`
- 局部交付结果

### 浏览器大脑

产出：

- 页面证据
- 页面事实摘要

### 审核大脑

产出：

- `verification_result_json`
- 断言结果

### 故障大脑

产出：

- `fault_summary_json`
- 失败分类

### easymvp-brain

消费上述归一化结果后，产出：

- `PlanReviewResult`
- `CompiledPlan`
- `RepairPlanDraft`
- `AcceptanceProfile`
- `CompletionVerdict`

---

## 11. 当前实现优先级

如果按落地顺序，优先级建议如下：

### P0

1. `plan_review`
2. `plan_compile`
3. `completion_adjudication`

### P1

4. `repair_design`
5. `acceptance_mapping`

原因：

- 没有 review / compile，就没有结构化主链
- 没有 completion adjudication，就仍然会把“run success”误判成完成
- repair 与 acceptance mapping 在第一版可以跟进，但不应阻塞主链成形

---

## 12. 最终定义

如果把这份文档压缩成一句话：

> **`easymvp-brain` 以 5 类核心契约承担 EasyMVP 的方案审核、方案编译、返工设计、验收映射和完成裁决。**
