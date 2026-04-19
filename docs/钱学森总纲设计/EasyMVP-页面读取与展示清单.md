# EasyMVP 页面读取与展示清单

> 更新时间：2026-04-20  
> 上位文档：[EasyMVP-对象级字段清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-对象级字段清单.md)  
> 关联文档：[EasyMVP-专项实施清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-专项实施清单.md) / [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md) / [EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md)

---

## 1. 文档目的

这份文档只解决一件事：

> EasyMVP 各页面必须读取哪些对象、展示哪些字段、禁止隐藏哪些状态。

它是对象级字段清单的前端/查询消费版。

---

## 2. 总体展示原则

所有页面统一遵守这 6 条：

1. 页面只消费 EasyMVP 归一化后的对象，不直接渲染原始脑 payload。
2. 页面不能只显示最终状态，必须显示导致该状态的关键原因。
3. `unsupported / denied / channel_unavailable / verification_conflict / manual_review_required / fault_loop_detected` 不允许被隐藏成普通失败。
4. `run success`、`acceptance passed`、`completed` 必须区分展示。
5. 返工页面不能只展示“可返工”，必须展示返工来源和返工策略。
6. GitHub Actions 必须展示为当前替代通道，不得展示成长期最终验证环境。

---

## 3. `WorkspacePage`

## 3.1 必读对象

1. `CompiledTask` 摘要
2. `CompletionVerdict` 摘要
3. `RuntimeEscalation` 摘要
4. 最近 `VerificationResult` 摘要

## 3.2 必展示字段

1. 当前阶段
2. 当前主任务
3. 当前脑路由摘要
4. 当前是否存在 escalation
5. 当前是否存在未收口的 verification / repair / manual checkpoint

## 3.3 必须可见的状态

1. `manual_review_required`
2. `policy_denied`
3. `verification_conflict`
4. `fault_loop_detected`

## 3.4 禁止的做法

1. 只展示“项目正常/异常”
2. 只展示活动流，不展示当前阻塞原因
3. 有 escalation 但页面入口层看不到

---

## 4. `AcceptancePage`

## 4.1 必读对象

1. `verification_contract_json`
2. `VerificationResult`
3. `CompletionVerdict`
4. 相关 `RuntimeEscalation`

## 4.2 必展示字段

1. `required_checks`
2. `required_evidence`
3. `preferred_verification_channel`
4. `blocking_issue_count`
5. `warning_count`
6. `missing_evidence`
7. `failed_checks`
8. `decision`

## 4.3 必须可见的差异信息

1. 合同要求了什么
2. 实际完成了什么
3. 缺了什么
4. 为什么还不能 completed

## 4.4 必须可见的通道信息

1. 当前用了哪个验证通道
2. 是 `high_spec_remote` 还是 `github_actions`
3. 若是 `github_actions`，要明确它是替代通道

## 4.5 禁止的做法

1. 只展示“验收通过/失败”
2. 不展示 contract gap
3. 有 `manual_review_required` 却不显示人工检查点

---

## 5. `DiagnosticsPage`

## 5.1 必读对象

1. `RuntimeEscalation`
2. `FaultSummary`
3. 最近 `VerificationResult`
4. 相关 `RepairPlanDraft`

## 5.2 必展示字段

1. `escalation_type`
2. `source_brain`
3. `reason_code`
4. `reason_summary`
5. `recommended_action`
6. `failure_class`
7. `failure_stage`
8. `recovery_options`

## 5.3 必须可见的分型

1. 能力不支持
2. 策略拒绝
3. 通道不可用
4. 验证冲突
5. 故障回路
6. 人工检查点

## 5.4 禁止的做法

1. 把所有问题都堆成 generic error
2. 只有日志，没有结构化问题类型
3. 有推荐动作却不展示

---

## 6. `Replay / Evidence Detail`

## 6.1 必读对象

1. `VerificationResult`
2. `verification_contract_json`
3. evidence 关联对象
4. 可选 `RuntimeEscalation`

## 6.2 必展示字段

1. 这条 replay / evidence 属于哪个 task
2. 属于哪个 check / evidence requirement
3. `channel`
4. `executor_brain`
5. 这条证据当前状态
6. 是否覆盖了合同中的某项要求

## 6.3 必须可见的解释

1. 这条证据支撑什么结论
2. 这条证据为什么不足
3. 这条证据对应的是 blocker 还是 warning

## 6.4 禁止的做法

1. 只展示 raw 内容，不展示语义归属
2. 用户看不到这条 replay / evidence 为什么重要

---

## 7. `RepairDraftPage`

## 7.1 必读对象

1. `RepairPlanDraft`
2. `FaultSummary`
3. `CompletionVerdict`
4. `verification_contract_json` 的调整项

## 7.2 必展示字段

1. `reason_class`
2. `repair_strategy`
3. `updated_tasks`
4. `verification_adjustments`
5. `delivery_adjustments`
6. `human_checkpoint_required`
7. `based_on_fault_id` 或等价来源引用

## 7.3 必须可见的返工来源

1. 因为什么失败触发返工
2. 原合同哪里出了问题
3. 返工后要改什么
4. 返工后怎么重新验

## 7.4 禁止的做法

1. 只给“生成返工草稿成功”
2. 不展示返工策略
3. 不展示验证调整

---

## 8. `ExecutionPage`

## 8.1 必读对象

1. `CompiledTask`
2. `RunResult`
3. `DeliveryResult`
4. 可选 `RuntimeEscalation`
5. 可选相关 `FaultSummary`

## 8.2 必展示字段

1. `brain_kind`
2. `role_type`
3. `delivery_status`
4. `contract_satisfied`
5. `delivery_gaps`
6. `RunResult.status`
7. 是否已触发 escalation

## 8.3 必须区分的状态

1. 运行完成
2. 交付完成
3. 验证完成
4. 业务完成

## 8.4 禁止的做法

1. 把执行页成功直接表达成项目已完成
2. 隐藏 `denied` 和 `unsupported`

---

## 9. `Plan / Task Detail`

## 9.1 必读对象

1. `CompiledTask`
2. `delivery_contract_json`
3. `verification_contract_json`
4. `resolve_trace_json`

## 9.2 必展示字段

1. 为什么是这个 `brain_kind`
2. 为什么是这个 `role_type`
3. 这个任务交什么
4. 这个任务怎么验
5. 是否要求人工检查

## 9.3 禁止的做法

1. 计划页只展示任务标题，不展示合同
2. 用户看不到脑路由依据

---

## 10. 页面最少要覆盖的隐藏风险

所有关键页面都至少要覆盖这 6 类隐藏风险：

1. `run success but not completed`
2. `verification passed but production not passed`
3. `manual release required but not released`
4. `policy denied`
5. `channel unavailable`
6. `fault loop detected`

如果页面看不到这些风险，就说明展示层还没有真正接上总纲。

---

## 11. 推荐实现顺序

页面侧建议顺序：

1. `AcceptancePage`
2. `DiagnosticsPage`
3. `RepairDraftPage`
4. `Replay / Evidence Detail`
5. `ExecutionPage`
6. `WorkspacePage`
7. `Plan / Task Detail`

原因：

1. acceptance 和 diagnostics 直接决定闭环是否成立
2. repair draft 决定返工链是不是结构化
3. replay/evidence 决定证据是不是可解释

---

## 12. 一句话结论

EasyMVP 页面层后续真正要做的，不是继续堆页面，而是让每个关键页面都能回答这三个问题：

1. 当前系统知道了什么
2. 当前系统还缺什么
3. 下一步应该做什么

回答不了这三个问题的页面，就还没有真正接上钱学森总纲。
