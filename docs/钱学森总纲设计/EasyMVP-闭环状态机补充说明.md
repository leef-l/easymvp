# EasyMVP 闭环状态机补充说明

> 更新时间：2026-04-20  
> 上位文档：[EasyMVP-四基础专精大脑阶段调用矩阵.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-四基础专精大脑阶段调用矩阵.md)  
> 关联文档：[EasyMVP-页面读取与展示清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-页面读取与展示清单.md) / [EasyMVP-对象级字段清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-对象级字段清单.md) / [EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md)

---

## 1. 文档目的

这份文档用于补上 EasyMVP 闭环状态机的控制规则。

它只回答 3 个问题：

1. 哪些状态可以自动推进
2. 哪些状态必须停住等待人工检查点
3. 哪些状态必须升级到返工、阻塞或故障链路

这不是重写工作流引擎文档，而是给 EasyMVP 总纲补一层业务控制规则。

---

## 2. 基本状态口径

当前工作流常见状态：

1. `designing`
2. `reviewing`
3. `executing`
4. `accepting`
5. `reworking`
6. `paused`
7. `completed`
8. `failed`
9. `canceled`

当前关键推进关系：

1. `review -> execute`
2. `accept -> complete`
3. `accept -> rework`
4. `rework -> execute`
5. `rework -> accept`

本补充文档只关注 `reviewing / executing / accepting / reworking / completed` 这条闭环主链。

---

## 3. 总体控制原则

状态机控制统一遵守这 6 条：

1. 没有结构化结果，不推进状态。
2. 有升级对象未收口，不推进到 `completed`。
3. 有人工检查点要求时，不自动越过人工环节。
4. 有验证冲突时，不自动判定通过。
5. 有故障回路时，不继续盲目重试。
6. `completed` 必须由 `CompletionVerdict.completed = true` 驱动，而不是由运行状态偷换。

---

## 4. `reviewing` 阶段规则

## 4.1 自动推进条件

仅当以下条件同时满足时，才允许自动进入 `executing`：

1. `PlanReviewResult.decision` 为 `approved` 或 `approved_with_advisory`
2. `compile_allowed = true`
3. `CompiledTask` 已写出 `brain_kind`
4. `delivery_contract_json` 已存在
5. `verification_contract_json` 已存在

## 4.2 必须停住条件

命中以下任一条件，必须停在 `reviewing`：

1. `compile_allowed = false`
2. 存在 blocking issue
3. 角色解析失败
4. 合同字段缺失

## 4.3 升级条件

命中以下任一条件，必须升级为人工或阻塞：

1. 项目分类不清
2. 计划范围持续膨胀
3. 风险等级过高
4. 关键验证通道无法声明

---

## 5. `executing` 阶段规则

## 5.1 自动推进条件

仅当以下条件同时满足时，才允许自动进入 `accepting`：

1. `RunResult.status = completed`
2. `DeliveryResult.delivery_status` 为 `delivered` 或可接受的 `partially_delivered`
3. `contract_satisfied = true` 或已被显式降级允许
4. 没有未收口的 `RuntimeEscalation`

## 5.2 必须停住条件

命中以下任一条件，必须停在 `executing`：

1. 任务仍在运行
2. 产物尚未生成
3. 交付差距未解释
4. 当前有重试但尚未出最终运行结果

## 5.3 必须升级条件

命中以下任一条件，不能直接进 `accepting`，必须升级：

1. `RunResult.status = unsupported`
2. `RunResult.status = denied`
3. 重试达到阈值
4. 工作区污染或资源冲突
5. 运行完成但交付合同未满足

升级后的默认去向：

1. `unsupported / denied / environment issue` -> `RuntimeEscalation`
2. 重复失败 -> `FaultSummary`
3. 合同不满足 -> `reworking` 或 `manual_checkpoint`

---

## 6. `accepting` 阶段规则

## 6.1 自动推进到 `completed` 的条件

仅当以下条件同时满足时，才允许自动进入 `completed`：

1. `VerificationResult.verdict` 为 `passed` 或允许的 `passed_with_warning`
2. `required_checks` 已满足
3. `required_evidence` 已满足
4. `CompletionVerdict.decision = complete`
5. `CompletionVerdict.completed = true`
6. 若 `manual_release_required = true`，则 `released_by_human = true`

## 6.2 自动推进到 `reworking` 的条件

命中以下任一条件，默认自动进入 `reworking`：

1. `VerificationResult.verdict = failed`
2. 必要证据缺失
3. blocker checks 未过
4. `CompletionVerdict.decision = rework`

## 6.3 必须停住等待人工检查点的条件

命中以下任一条件，必须停住，不能自动 complete，也不能直接重试：

1. `VerificationResult.verdict = manual_review_required`
2. `CompletionVerdict.decision = manual_checkpoint`
3. `manual_release_required = true` 且未放行
4. 存在 `verification_conflict`

## 6.4 禁止的推进方式

以下推进一律禁止：

1. `run success -> completed`
2. `acceptance passed -> completed`，但人工放行未完成
3. 有 contract gap 仍自动完成

---

## 7. `reworking` 阶段规则

## 7.1 自动推进到 `executing` 的条件

仅当以下条件同时满足时，才允许从 `reworking` 回到 `executing`：

1. 已生成 `RepairPlanDraft`
2. `updated_tasks` 已明确
3. `verification_adjustments` 已明确
4. `delivery_adjustments` 已明确
5. `human_checkpoint_required = false`

## 7.2 必须停住条件

命中以下任一条件，必须停在 `reworking`：

1. 返工策略不明确
2. 新任务边界不明确
3. 验证调整不明确
4. 人工检查点尚未完成

## 7.3 必须升级为阻塞的条件

命中以下任一条件，不能继续自动返工：

1. `fault_loop_detected`
2. 同一类故障连续复发
3. 风险等级升级到 `high`
4. 验证通道本身失真或不可用

默认去向：

1. `blocked`
2. `manual_checkpoint`
3. 必要时回到 `reviewing` 做整体重编译

---

## 8. `completed` 阶段规则

## 8.1 进入条件

进入 `completed` 的唯一业务条件是：

1. `CompletionVerdict.completed = true`

辅助条件是：

1. `executor_succeeded = true`
2. `delivery_verified = true`
3. `acceptance_passed = true`
4. 必要人工放行已完成

## 8.2 禁止进入条件

命中以下任一条件，禁止进入 `completed`：

1. 存在未处理 `RuntimeEscalation`
2. 存在 `manual_review_required`
3. 存在 `fault_loop_detected`
4. `production_passed = false`
5. 证据链未闭合

---

## 9. 人工检查点规则

以下情况必须创建人工检查点：

1. `policy_denied`
2. `verification_conflict`
3. `manual_review_required`
4. `manual_release_required = true`
5. `fault_loop_detected`
6. 高风险返工

人工检查点至少要回答：

1. 是继续执行
2. 是转返工
3. 是阻塞
4. 是人工放行完成

---

## 10. 升级优先级规则

当多个异常同时命中时，按以下优先级处理：

1. `policy_denied`
2. `verification_conflict`
3. `manual_review_required`
4. `fault_loop_detected`
5. `environment_unavailable`
6. 一般执行失败

原因：

1. 策略拒绝不是普通错误
2. 验证冲突说明测量端已经不可信
3. 故障回路说明自动闭环开始震荡

---

## 11. 页面最低联动要求

为了让状态机不是黑盒，页面最低要满足：

1. `WorkspacePage` 能看到当前是否停在人工检查点
2. `AcceptancePage` 能看到为什么不能 completed
3. `DiagnosticsPage` 能看到升级原因和推荐动作
4. `RepairDraftPage` 能看到为什么进入 reworking
5. `ExecutionPage` 能看到运行成功为何仍未闭环

---

## 12. 一句话结论

EasyMVP 这条闭环状态机真正要守住的，不是“能不能自动往前跑”，而是：

> 只有在测量可信、合同满足、升级收口、人工条件满足时，系统才允许前进。

否则就必须停住、升级、返工，而不是假装完成。
