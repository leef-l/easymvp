# EasyMVP 对象级字段清单

> 更新时间：2026-04-20  
> 上位文档：[EasyMVP-专项实施清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-专项实施清单.md)  
> 关联文档：[easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md) / [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md) / [EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md)

---

## 1. 文档目的

这份文档把 EasyMVP 专项里最关键的对象，收成字段级基线。

它只回答 4 个问题：

1. 每个对象最少要有哪些字段
2. 哪些字段是推荐补充
3. 哪些做法明确不允许
4. 怎么判断这个对象已经够资格进入实现层

这份文档只服务 EasyMVP，不扩展到 `brain-v3` 原生协议。

---

## 2. 对象范围

本期只固定 6 个对象：

1. `CompiledTask`
2. `VerificationResult`
3. `CompletionVerdict`
4. `FaultSummary`
5. `RepairPlanDraft`
6. `RuntimeEscalation`

原因很直接：

- 这 6 个对象覆盖了编译、验证、完成、故障、返工、升级 6 个闭环节点
- 这 6 个对象一旦做实，EasyMVP 的总纲就不再只是概念

---

## 3. `CompiledTask`

## 3.1 最小必填字段

1. `compiled_task_id`
2. `name`
3. `role_type`
4. `brain_kind`
5. `risk_level`
6. `delivery_contract_json`
7. `verification_contract_json`

## 3.2 推荐字段

1. `task_kind`
2. `stage_hint`
3. `artifact_refs`
4. `manual_review_required`
5. `resolve_trace_json`
6. `depends_on_task_ids`

## 3.3 不允许的做法

1. 只写 `role_type` 不写 `brain_kind`
2. 只有自然语言任务说明，没有交付合同
3. 只有交付要求，没有验证合同
4. 到执行阶段再临时推断该走哪个脑

## 3.4 完成判据

满足以下 3 条才算过关：

1. 编译后任务能直接路由，不需要再猜 `brain_kind`
2. 执行前就知道交付什么、怎么验
3. accepting / reworking / completed 都能回溯到这个对象

---

## 4. `VerificationResult`

## 4.1 最小必填字段

1. `verification_run_id`
2. `contract_id`
3. `task_id`
4. `channel`
5. `executor_brain`
6. `checks`
7. `evidence_refs`
8. `blocking_issue_count`
9. `warning_count`
10. `verdict`
11. `reason_summary`

## 4.2 推荐字段

1. `started_at`
2. `ended_at`
3. `coverage_summary`
4. `missing_evidence`
5. `failed_checks`
6. `raw_ref`

## 4.3 `verdict` 推荐枚举

1. `passed`
2. `passed_with_warning`
3. `failed`
4. `manual_review_required`
5. `channel_unavailable`

## 4.4 不允许的做法

1. 只返回一句“验证通过”
2. 不记录 `channel`
3. 不记录 `executor_brain`
4. 有 evidence 却没有 check 对齐关系

## 4.5 完成判据

满足以下 3 条才算过关：

1. 能对齐 `verification_contract_json`
2. 能解释“缺什么、错什么、通了什么”
3. 能被 acceptance 页面和 diagnostics 页面直接消费

---

## 5. `CompletionVerdict`

## 5.1 最小必填字段

1. `executor_succeeded`
2. `delivery_verified`
3. `acceptance_passed`
4. `completed`
5. `decision`
6. `reason_summary`

## 5.2 推荐字段

1. `manual_release_required`
2. `released_by_human`
3. `blocking_reasons`
4. `advisory_reasons`
5. `next_action`
6. `source_refs`

## 5.3 `decision` 推荐枚举

1. `complete`
2. `rework`
3. `blocked`
4. `manual_checkpoint`

## 5.4 不允许的做法

1. 把 `executor_succeeded = true` 当成 `completed = true`
2. 没区分 `acceptance_passed` 和 `completed`
3. 只返回自由文本，不给结构化决策

## 5.5 完成判据

满足以下 3 条才算过关：

1. 任何 completed 前都能看到 verdict
2. verdict 能明确指向下一步动作
3. 自动完成逻辑不能绕过它

---

## 6. `FaultSummary`

## 6.1 最小必填字段

1. `fault_id`
2. `task_id`
3. `failure_class`
4. `failure_stage`
5. `failure_scope`
6. `probable_causes`
7. `recovery_options`
8. `auto_recoverable`
9. `human_checkpoint_required`
10. `evidence_refs`

## 6.2 推荐字段

1. `retry_count`
2. `last_failure_at`
3. `fault_signature`
4. `related_run_ids`
5. `diagnostic_refs`

## 6.3 不允许的做法

1. 只存原始日志，不做失败归类
2. 不区分是执行失败、验证失败还是环境失败
3. 没有 recovery option 就直接进返工

## 6.4 完成判据

满足以下 3 条才算过关：

1. 失败能被结构化分类
2. 同类故障可识别为同一类
3. `repair_design` 能直接消费它

---

## 7. `RepairPlanDraft`

## 7.1 最小必填字段

1. `repair_plan_id`
2. `reason_class`
3. `repair_strategy`
4. `updated_tasks`
5. `verification_adjustments`
6. `delivery_adjustments`
7. `human_checkpoint_required`

## 7.2 推荐字段

1. `based_on_fault_id`
2. `based_on_contract_refs`
3. `scope_change`
4. `risk_change`
5. `manual_notes`

## 7.3 不允许的做法

1. 把返工写成一句“重试一下”
2. 返工不说明改哪些任务
3. 返工不说明验证要求怎么变化

## 7.4 完成判据

满足以下 3 条才算过关：

1. 能明确局部返工还是整体重编译
2. 能明确新任务边界
3. 能明确返工后的验收要求

---

## 8. `RuntimeEscalation`

## 8.1 最小必填字段

1. `escalation_id`
2. `task_id`
3. `stage`
4. `source_brain`
5. `escalation_type`
6. `severity`
7. `reason_code`
8. `reason_summary`
9. `recommended_action`
10. `evidence_refs`

## 8.2 推荐字段

1. `created_at`
2. `resolved_at`
3. `resolution_status`
4. `resolver_kind`
5. `linked_fault_id`

## 8.3 `escalation_type` 推荐枚举

1. `retry_exhausted`
2. `unsupported_capability`
3. `policy_denied`
4. `verification_conflict`
5. `manual_review_required`
6. `environment_unavailable`
7. `fault_loop_detected`

## 8.4 不允许的做法

1. 升级状态只写进日志，不进结构化对象
2. 页面只能看到“失败”，看不到为什么升级
3. 不记录 `source_brain`

## 8.5 完成判据

满足以下 3 条才算过关：

1. 页面能解释升级原因
2. 编排层能根据升级类型决定下一步
3. diagnostics / acceptance / rework 都能引用它

---

## 9. 对象关系速查

推荐关系如下：

```text
CompiledTask
  -> RunResult / DeliveryResult
  -> VerificationResult
  -> CompletionVerdict

Run failure / Verification failure
  -> FaultSummary
  -> RepairPlanDraft

Runtime exception / policy / channel / conflict
  -> RuntimeEscalation
```

这 3 条关系分别对应：

1. 正常完成链
2. 返工链
3. 升级链

---

## 10. 实现优先级

建议顺序：

1. `CompiledTask`
2. `VerificationResult`
3. `CompletionVerdict`
4. `FaultSummary`
5. `RepairPlanDraft`
6. `RuntimeEscalation`

原因：

1. 没有任务合同，后面的验证和裁决都站不住
2. 没有验证结果，完成裁决只能继续拍脑袋
3. 没有故障和返工对象，闭环依旧是软的

---

## 11. 一句话结论

EasyMVP 这部分后续实现，不应再问“要不要加更多对象”，而应先把这 6 个对象做到字段稳定、状态稳定、页面可读、闭环可追。
