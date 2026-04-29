# EasyMVP 中央大脑与四专精大脑 I/O 合同及升级规则

> 更新时间：2026-04-20  
> 上位文档：[EasyMVP-四基础专精大脑阶段调用矩阵.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-四基础专精大脑阶段调用矩阵.md)  
> 关联文档：[easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md) / [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md) / [EasyMVP工程铁律.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP工程铁律.md)

---

## 1. 文档目的

这份文档用于把“中央大脑 + 四基础专精大脑”的协作，从原则描述收敛成接口级基线。

它只解决 4 个问题：

1. EasyMVP 给 `central / code / browser / verifier / fault` 的标准输入是什么
2. 每个脑至少要回什么标准输出
3. 哪些失败必须显式升级，不允许被包装成成功
4. 哪些状态由 runtime adapter 吸收，哪些状态进入领域层

这份文档不替代 `brain-v3` 原生协议说明，它只定义 EasyMVP 侧消费和落库的统一口径。

---

## 2. 总体结论

先固定结论：

1. EasyMVP 不直接消费任意原始脑输出，必须先经过 runtime adapter 归一化。
2. `central` 的职责是协调和仲裁，不是业务语义裁决器。
3. `code / browser / verifier / fault` 的结果必须回收成固定类型摘要，不能把各自原始话术透传给领域层。
4. `unsupported`、`denied`、`channel_unavailable`、`manual_review_required` 都必须保留为显式状态，不准伪装成普通失败，更不准伪装成成功。
5. 领域层只消费 6 类对象：
   - `RuntimeTaskIntent`
   - `RunResult`
   - `DeliveryResult`
   - `VerificationResult`
   - `FaultSummary`
   - `RuntimeEscalation`

---

## 3. 分层边界

统一分成 3 层：

### 3.1 编排层

由 EasyMVP 负责：

- Workflow 阶段推进
- 任务持久化
- 合同挂载
- 证据归档
- 人工审批

### 3.2 运行时适配层

由 runtime adapter 负责：

- 调 `brain-v3`
- 吸收原始 `tools/list` / `tools/call`
- 吸收 `completed / failed / unsupported / denied`
- 归一化运行状态、证据引用、故障摘要

### 3.3 领域层

由 `easymvp-brain` 和 Acceptance / Orchestrator 消费：

- 结构化摘要
- 结构化合同
- 结构化裁决

关键边界：

> 原始脑协议停在适配层；  
> 领域层永远只看归一化对象。

---

## 4. 统一输入对象

## 4.1 `RuntimeTaskIntent`

这是 EasyMVP 发给 runtime 层的统一任务意图对象。

建议至少包含：

1. `task_id`
2. `project_id`
3. `workflow_run_id`
4. `stage`
5. `brain_kind`
6. `role_type`
7. `goal`
8. `input_summary`
9. `delivery_contract`
10. `verification_contract`
11. `artifact_refs`
12. `risk_level`
13. `timeout_policy`
14. `manual_review_required`

输入原则：

1. 不传页面展示文案当执行指令
2. 不把原始 DB 对象整包丢给脑
3. 不要求脑理解 EasyMVP 全部业务状态机
4. 输入必须足够让脑知道“做什么、做到什么程度、怎么交付、怎么验证”

## 4.2 `CentralCoordinationIntent`

发给 `central` 的输入不是“请你随便协调”，而应是结构化协调请求。

建议至少包含：

1. `coordination_id`
2. `task_intent`
3. `preferred_brain`
4. `fallback_brains`
5. `delegation_constraints`
6. `workdir_scope`
7. `tool_scope`
8. `risk_policy`
9. `escalation_policy`

---

## 5. 统一输出对象

## 5.1 `RunResult`

适用于所有脑。

建议至少包含：

1. `run_id`
2. `task_id`
3. `brain_kind`
4. `executor_brain`
5. `status`
6. `started_at`
7. `ended_at`
8. `summary`
9. `artifact_refs`
10. `runtime_flags`
11. `raw_event_refs`

`status` 建议统一为：

- `completed`
- `failed`
- `unsupported`
- `denied`
- `cancelled`
- `timeout`

说明：

- `completed` 只表示本次运行完成，不等于业务完成。
- `unsupported` 表示能力或协议不支持。
- `denied` 表示被策略、权限、范围、文件规则拒绝。

## 5.2 `DeliveryResult`

主要用于 `code` 或可生成产物的脑。

建议至少包含：

1. `task_id`
2. `delivery_status`
3. `delivered_artifacts`
4. `changed_resources`
5. `contract_satisfied`
6. `delivery_gaps`

`delivery_status` 建议统一为：

- `delivered`
- `partially_delivered`
- `not_delivered`

## 5.3 `VerificationResult`

由 `browser` 或 `verifier` 主产出。

详细结构沿用 [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md)。

这里额外强调：

1. 必须包含 `channel`
2. 必须包含 `executor_brain`
3. 必须包含 `evidence_refs`
4. 必须包含 `verdict`

## 5.4 `FaultSummary`

由 `fault` 主产出。

建议至少包含：

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

---

## 6. 各脑合同定义

## 6.1 `central`

### 输入

- `CentralCoordinationIntent`
- 可选当前阶段上下文
- 可选候选脑能力信息

### 输出

- `RunResult`
- `delegation_trace`
- `selected_brain`
- 可选下游脑结果聚合摘要
- 必要时 `RuntimeEscalation`

### 负责

- 选脑
- 委派
- 汇总
- 仲裁运行时分歧
- 记录 delegate trace

### 不负责

- 输出 `PlanReviewResult`
- 输出 `CompiledPlan`
- 输出 `CompletionVerdict`
- 直接替代 `fault` 做故障诊断

## 6.2 `code`

### 输入

- `RuntimeTaskIntent`

最小必要输入包括：

- `goal`
- `delivery_contract`
- `artifact_refs`
- 允许的工作目录与工具范围

### 输出

- `RunResult`
- `DeliveryResult`
- 可选局部开发级检查摘要

### 负责

- 代码读写
- 开发级执行
- 产物生成
- 交付差异说明

### 不负责

- 直接宣布验收通过
- 替代 `verifier` 产出最终验证结论
- 替代 `easymvp-brain` 决定返工策略

## 6.3 `browser`

### 输入

- `RuntimeTaskIntent`

最小必要输入包括：

- `goal`
- `verification_contract`
- `artifact_refs`
- 目标页面/路径摘要

### 输出

- `RunResult`
- `VerificationResult`
- `Evidence`

### 负责

- 页面操作
- 关键 journey 验证
- 截图、trace、页面证据采集

### 不负责

- 替代 `verifier` 做纯规则型只读核验
- 替代 `easymvp-brain` 做最终完成语义裁决

## 6.4 `verifier`

### 输入

- `RuntimeTaskIntent`

最小必要输入包括：

- `goal`
- `verification_contract`
- `artifact_refs`
- 结构化检查目标

### 输出

- `RunResult`
- `VerificationResult`

### 负责

- 只读核验
- 结构化断言
- 验证报告归一化

### 不负责

- 替代 `browser` 跑真实页面交互
- 替代 `fault` 分析故障根因

## 6.5 `fault`

### 输入

- `RuntimeTaskIntent`
- 失败上下文
- `RunResult`
- 可选 `VerificationResult`
- 可选 `DeliveryResult`

### 输出

- `RunResult`
- `FaultSummary`

### 负责

- 失败分类
- 故障范围判断
- 恢复方向建议
- 是否适合自动恢复的判断

### 不负责

- 直接修改代码
- 直接生成最终返工任务清单
- 直接输出 `RepairPlanDraft`

---

## 7. 统一升级对象

## 7.1 `RuntimeEscalation`

任何需要从运行时上升到业务编排或人工处理的情况，统一收口为这个对象。

建议至少包含：

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

`escalation_type` 建议至少包含：

- `retry_exhausted`
- `unsupported_capability`
- `policy_denied`
- `verification_conflict`
- `manual_review_required`
- `environment_unavailable`
- `fault_loop_detected`

---

## 8. 升级规则

> ✅ **基础已实现**（2026-04-27）：统一状态枚举 `NormalizedStatus` 和升级类型枚举 `EscalationType` 已定义于 `contract_status.go`，包含 `IsValidNormalizedStatus` / `IsValidEscalationType` 校验函数。高级业务规则绑定待后续迭代。

以下情况一律不能静默吞掉，必须产出 `RuntimeEscalation`。

## 8.1 `unsupported`

触发条件：

- 目标脑或目标通道不支持当前任务能力
- 协议不支持所需动作
- 工具范围无法覆盖合同要求

处理规则：

1. `RunResult.status = unsupported`
2. 产出 `RuntimeEscalation`
3. 由编排层决定改路由、降级还是人工接管

## 8.2 `denied`

触发条件：

- 策略拒绝
- 权限不足
- 文件策略拒绝
- 工作目录或工具范围超界

处理规则：

1. `RunResult.status = denied`
2. 保留拒绝原因
3. 不自动伪装成普通失败重试

## 8.3 验证冲突

触发条件：

- `browser` 与 `verifier` 结论不一致
- 多通道验证结果互相矛盾
- 证据齐全但 verdict 不一致

处理规则：

1. 产出 `verification_conflict`
2. 禁止直接进入 `completed`
3. 默认升级到人工检查或 `easymvp-brain.completion_adjudication`

## 8.4 故障回路

触发条件：

- 同类故障重复出现
- 相同恢复动作连续失败
- `fault` 判断恢复路径不稳定

处理规则：

1. 产出 `fault_loop_detected`
2. 强制进入 `fault -> easymvp-brain -> RepairPlanDraft`
3. 禁止继续盲目自动重试

## 8.5 通道不可用

触发条件：

- 高配验证环境不可达
- GitHub Actions 当前不可用
- 浏览器验证通道无法启动

处理规则：

1. `VerificationResult.verdict = channel_unavailable`
2. 检查 `fallback_channels`
3. 若无合法回退，升级到 `environment_unavailable`

## 8.6 人工检查点

触发条件：

- 风险等级高
- 合同明确要求人工审核
- 放行规则要求人工确认
- 自动验证存在争议

处理规则：

1. 产出 `manual_review_required`
2. workflow 不得自动宣告完成

---

## 9. 状态吸收规则

这部分专门防止 EasyMVP 以后又把原始脑协议泄漏到领域层。

## 9.1 停在 runtime adapter 的状态

以下内容只允许出现在 runtime adapter 和审计层：

- 原始 `tools/list`
- 原始 `tools/call`
- 原始 `content[]`
- 供应商特定字段
- 原始提示词细节

## 9.2 允许进入领域层的状态

领域层只应看到：

- `RunResult.status`
- `DeliveryResult.delivery_status`
- `VerificationResult.verdict`
- `FaultSummary.failure_class`
- `RuntimeEscalation.escalation_type`

这几个状态足以让业务继续推进，不需要知道底层工具细节。

---

## 10. 典型链路样例

## 10.1 代码执行成功但验证失败

```text
EasyMVP
  -> RuntimeTaskIntent(brain_kind=code)
  -> code
  -> RunResult(status=completed)
  -> DeliveryResult(delivery_status=delivered)
  -> verifier
  -> VerificationResult(verdict=failed)
  -> RuntimeEscalation(escalation_type=manual_review_required 或 verification_conflict)
  -> fault
  -> FaultSummary
  -> easymvp-brain.repair_design
  -> RepairPlanDraft
```

结论：

- 不能因为 `code` 成功就直接 completed

## 10.2 浏览器采证成功但远端 CI 不可用

```text
browser
  -> VerificationResult(verdict=passed)
github_actions
  -> VerificationResult(verdict=channel_unavailable)
RuntimeEscalation(environment_unavailable)
```

结论：

- 若合同要求 CI 结果为 blocking，则不得完成

## 10.3 策略拒绝

```text
central
  -> code
  -> RunResult(status=denied)
  -> RuntimeEscalation(policy_denied)
```

结论：

- 这不是普通失败，必须显式保留为策略问题

---

## 11. 与控制论原则的对应

这份 I/O 合同是在把多脑协作从“口头调度”变成“可控系统”。

对应关系：

1. 反馈闭环：每个脑都必须输出可回收的结构化反馈量
2. 最小互扰：每个脑只暴露必要摘要，不互相吞职责
3. 时滞显式处理：`channel_unavailable` 与远端验证延迟被显式建模
4. 噪声过滤：原始工具 payload 停在 adapter，不直接污染领域层
5. 稳定优先：`fault_loop_detected` 时禁止继续盲重试

---

## 12. 落地建议

按优先级建议先做这 6 项：

1. 在 runtime adapter 固化 `RunResult`、`DeliveryResult`、`VerificationResult`、`FaultSummary`、`RuntimeEscalation` 五类 DTO
2. 在 `CompiledTask` 固化 `brain_kind`、`delivery_contract_json`、`verification_contract_json`
3. 在运行时事件索引中保留 `raw_event_refs`，但页面和领域层不直接读原始 payload
4. 对 `unsupported / denied / channel_unavailable / manual_review_required` 建立专门错误域和状态枚举
5. 在返工链路中强制经过 `fault -> easymvp-brain.repair_design`
6. 在最终裁决前强制检查所有 `RuntimeEscalation` 是否已收口

做完这 6 项，EasyMVP 的多脑协作才真正具备“可实现、可审计、可闭环”的接口基线。
