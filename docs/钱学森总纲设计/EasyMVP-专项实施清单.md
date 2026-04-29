# EasyMVP 专项实施清单

> 更新时间：2026-04-20  
> 上位文档：[EasyMVP-钱学森总纲落地缺口与实施顺序.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-钱学森总纲落地缺口与实施顺序.md)  
> 关联文档：[EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md) / [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md) / [easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md)

---

## 1. 文档目的

这份文档不再讨论“为什么这样设计”，只回答：

> EasyMVP 专项接下来具体要做什么，落到哪里，做到什么程度才算完成。

它是 `钱学森总纲设计` 目录下的施工单。

---

## 2. 施工范围

只覆盖 EasyMVP 自己需要落的内容：

1. `easymvp-brain` 领域合同
2. EasyMVP 视角下的四基础专精大脑协作
3. 验证合同、返工闭环、完成裁决
4. EasyMVP 页面和查询侧对这些对象的消费

不在本清单中的内容：

1. `brain-v3` 自身 roadmap
2. 通用 AgentTeam 规划
3. 与 EasyMVP 无关的通用运行时能力扩张

---

## 3. 当前状态判断

当前已经有的：

1. 总纲、边界、契约、矩阵、验证合同、I/O 合同都已成文
2. `plan_review / plan_compile / acceptance_mapping / completion_adjudication / repair_design` 口径已明确
3. `unsupported / denied` 必须显式保留，这条边界已明确

当前还缺的不是“有没有文档”，而是以下 4 类对象是否真的进入 EasyMVP 实现和展示层：

1. `verification_contract_json`
2. `CompletionVerdict`
3. `RepairPlanDraft`
4. `RuntimeEscalation`

---

## 4. P0 清单

## 4.1 P0-01 固化 `CompiledTask` 三个关键字段

目标：

- 让编译后的任务真正携带 EasyMVP 后续闭环所需的最小合同

必须落的字段：

1. `brain_kind`
2. `delivery_contract_json`
3. `verification_contract_json`

完成定义：

1. 这三个字段在 `CompiledTask` 写侧稳定生成
2. 读侧能稳定读出
3. 后续验收、返工、完成裁决不再依赖临时推断

不完成的后果：

- 后面所有 accepting / reworking / completed 都会继续靠补丁逻辑

## 4.2 P0-02 固化 `CompletionVerdict` 为 completed 前最后裁决

> ✅ **已实现**（2026-04-27）：`updateProjectFromAdjudication` 已明确注释为设置 `completed` 的唯一授权路径；运行时自动 adjudication 已集成；`UpdateProject` API 禁止直接修改 status。

目标：

- 禁止把“run 成功”直接判成“完成”

必须保证：

1. `executor_succeeded`
2. `delivery_verified`
3. `acceptance_passed`
4. `completed`

这 4 个状态被明确区分。

完成定义：

1. 任意任务进入最终完成前都必须经过 `completion_adjudication`
2. 页面和查询层展示的是结构化 verdict，而不是一句自然语言总结
3. 自动完成逻辑不能绕过 verdict

## 4.3 P0-03 固化 `fault -> repair_design -> reworking` 唯一路径

> ✅ **已实现**（2026-04-27）：`shouldCreateRepairDraftAfterAdjudication` 已扩展为在 `Decision == "rework"` 或 `FinalStatus == "failed"` 时均触发 repair draft；函数注释明确禁止失败后直接重试。

目标：

- 让返工从“失败后顺手再试一次”变成结构化闭环

必须保证：

1. 失败先归一化为 `FaultSummary`
2. 再进入 `RepairPlanDraft`
3. 再进入 `reworking`

完成定义：

1. 自动重试达到阈值后不再直接重试
2. 同类故障重复出现时，必须有显式返工入口
3. 页面可看到返工原因、返工策略、返工来源证据

## 4.4 P0-04 固化 `verification_contract_json` 进入 accepting 主链

目标：

- 让 accepting 阶段按合同跑，而不是按零散逻辑拼

必须保证：

1. 验收时知道要验什么
2. 知道要收集什么 evidence
3. 知道 blocker 和 warning 的区别

完成定义：

1. accepting 阶段按 `verification_contract_json` 选择验证通道
2. `VerificationResult` 能对齐到合同中的 `required_checks`
3. 缺失证据和失败检查能区分展示

---

## 5. P1 清单

> ✅ **P1-01 / P1-02 页面展示**（2026-04-27）：AcceptancePage / DiagnosticsPage / WorkspacePage 已显示 `runtime_escalation`、`missing_evidence`、`failed_checks`。
> ✅ **P1-03 升级规则代码化**（2026-04-27）：`mapDiagnosticCategoryToEscalationType` 已把诊断分类映射到 `EscalationType` 枚举。

## 5.1 P1-01 读取侧显示 verification contract gap

目标：

- 页面不能只告诉用户“失败了”，必须解释“合同要求什么、当前缺什么”

最少需要出现在：

1. acceptance 页面
2. diagnostics 页面
3. replay / evidence 详情页

完成定义：

1. 页面可列出 required checks
2. 页面可列出 required evidence
3. 页面可标识 blocker / warning / missing

## 5.2 P1-02 读取侧显示 escalation reason

目标：

- 把运行时升级原因变成用户能看懂、能处理的对象

最少要覆盖的类型：

1. `unsupported_capability`
2. `policy_denied`
3. `verification_conflict`
4. `environment_unavailable`
5. `manual_review_required`
6. `fault_loop_detected`

完成定义：

1. 页面能显示升级类型
2. 页面能显示来源脑
3. 页面能显示建议动作

## 5.3 P1-03 接通高配验证环境口径

目标：

- 把“高配验证环境是长期目标”从文档变成 EasyMVP 里可配置、可展示、可裁决的通道

最少需要做到：

1. acceptance mapping 可选 `high_spec_remote`
2. verification contract 可声明 `preferred_verification_channel = high_spec_remote`
3. 页面能区分当前是 GitHub Actions 替代通道，还是高配正式通道

完成定义：

1. 通道选择口径进入结构化字段
2. 页面上不再把 GitHub Actions 表达为长期终局

---

## 6. P2 清单

> ✅ **P2-01 阶段矩阵代码化**（2026-04-27）：`stage_matrix.go` 已完成。

## 6.1 P2-01 Completion / Repair / Verification 三条链的统一说明页

目标：

- 让 EasyMVP 内部人员一眼能看懂三条闭环链怎么串

建议内容：

1. 从 `CompiledTask`
2. 到 `RunResult / DeliveryResult / VerificationResult`
3. 到 `CompletionVerdict`
4. 到 `FaultSummary / RepairPlanDraft`

这个页是辅助理解页，不是 P0。

## 6.2 P2-02 页面术语统一

目标：

- 前端和后端、文档和页面，术语全部统一

重点统一的词：

1. 验证
2. 验收
3. 返工
4. 完成
5. 人工检查点
6. 替代验证通道

---

## 7. 对象级施工映射

| 对象 | 该做什么 | 做完怎么判断 |
|---|---|---|
| `CompiledTask` | 补齐脑路由和双合同字段 | 不再靠阶段逻辑反推合同 |
| `VerificationResult` | 与 `verification_contract_json` 对齐 | 可解释每个 check/evidence 的通过与缺失 |
| `CompletionVerdict` | 成为最终完成唯一结构化裁决 | completed 不再被 run success 偷换 |
| `FaultSummary` | 承接失败归一化 | 不再只有日志，没有失败语义 |
| `RepairPlanDraft` | 成为返工正式入口 | rework 不再只是“再来一次” |
| `RuntimeEscalation` | 承接需要升级的运行时状态 | 页面可解释为什么升级、怎么处理 |

---

## 8. 完成顺序

建议严格按这个顺序做：

1. `CompiledTask` 三字段
2. `CompletionVerdict` 强制裁决
3. `FaultSummary -> RepairPlanDraft -> reworking`
4. `verification_contract_json` 进入 accepting
5. 页面显示 contract gap
6. 页面显示 escalation reason
7. 高配验证环境口径接入

原因：

1. 没有任务合同，后面都只是补丁
2. 没有完成裁决，系统就会继续把“执行成功”误判成完成
3. 没有返工正式入口，acceptance / repair 闭环就还是软的

---

## 9. 本目录后续只允许新增的文档类型

为了避免又扩散成大而空，后续如果还要新增文档，只允许这三类：

1. EasyMVP 对象级字段清单
2. EasyMVP 页面读取与展示清单
3. EasyMVP 闭环状态机补充说明

不再新增：

1. 泛泛的总纲重复稿
2. 与 EasyMVP 无直接关系的通用计划
3. 脱离对象和完成定义的概念稿

---

## 10. 一句话结论

EasyMVP 这部分真正要做的，不是继续扩写理论，而是把 6 个对象做实：

1. `CompiledTask`
2. `VerificationResult`
3. `CompletionVerdict`
4. `FaultSummary`
5. `RepairPlanDraft`
6. `RuntimeEscalation`

这 6 个对象做实，钱学森总纲在 EasyMVP 里才算真正落地。
