# EasyMVP Verification Contract 统一设计

> 更新时间：2026-04-20  
> 上位文档：[EasyMVP-三层验证架构说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-三层验证架构说明.md)  
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md](/www/wwwroot/project/easymvp/docs/EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md) / [easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md) / [EasyMVP工程铁律.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP工程铁律.md)

---

## 1. 文档目的

这份文档用于统一 `verification_contract` 的设计口径。

它要解决 5 个混乱点：

1. 什么叫验证，什么叫验收
2. 开发机、高配验证环境、GitHub Actions 各自算什么角色
3. `verification_contract_json` 至少要有哪些字段
4. `VerificationResult`、`Evidence`、`ProductionAcceptanceProfile` 怎么衔接
5. 什么情况下验证通过但仍不能判定完成

---

## 2. 设计结论

先给结论：

1. `Verification Contract` 是任务级验证合同，不等于最终验收结论。
2. 最终目标验证环境是高配验证环境，不是当前低配开发机。
3. GitHub Actions 只是当前服务器资源不足时的远端替代验证通道，不应被表述为长期终局。
4. 开发机可以做开发级轻量检查，但不能作为最终 acceptance 的唯一依据。
5. 验证结果必须结构化落库，并能映射到 `AcceptanceProfile` 和 `ProductionAcceptanceProfile`。

---

## 3. 核心定义

## 3.1 `Verification Contract`

定义：

> 某个 `CompiledTask` 在交付后必须满足的验证目标、验证通道、证据要求和裁决规则。

它回答的是：

- 这个任务要怎么验证
- 最少需要什么证据
- 哪些失败算 blocker
- 哪些 warning 可以容忍
- 优先走哪条验证通道

它不直接回答：

- 整个项目是否已达到生产可交付
- 是否已经人工放行
- 是否已经业务完成

这些属于 `AcceptanceProfile` / `ProductionAcceptanceProfile` / `CompletionVerdict`。

## 3.2 `Verification Result`

定义：

> 某次验证执行的结构化结果，包括通道、证据、检查项、阻塞项和最终 verdict。

## 3.3 `Acceptance Profile`

定义：

> 某类项目在验收层要求覆盖哪些 surface、journey、evidence 和 gate。

关系如下：

```text
CompiledTask
  -> Verification Contract
  -> Verification Run
  -> Verification Result
  -> Evidence
  -> Acceptance Profile / ProductionAcceptanceProfile
  -> Completion Verdict
```

---

## 4. 三层验证环境口径

## 4.1 低配开发机

角色：

- 开发过程中的轻量验证环境
- 本地快速检查和问题定位环境

允许承担：

- 语法/静态检查
- 小范围开发级脚本
- 局部 smoke 验证
- 调试辅助日志

不允许承担：

- 最终生产级验证
- 最终验收唯一依据
- 重负载构建与完整验证闭环

## 4.2 高配验证环境

角色：

- 长期目标中的正式最终验证环境
- 完整构建、完整测试、完整验收证据生成环境

这是未来应当优先命中的：

- `preferred_verification_channel = high_spec_remote`

## 4.3 GitHub Actions

角色：

- 当前阶段的远端替代验证环境
- 由于服务器配置过低而采取的现实折中

当前可以作为：

- `preferred_verification_channel = github_actions`

但文档和系统语义都必须保留这个限定：

> 它是当前替代通道，不是长期最终形态。

---

## 5. 统一对象设计

## 5.1 `verification_contract_json` 建议字段

建议至少包含：

1. `contract_id`
2. `contract_version`
3. `task_id`
4. `verification_scope`
5. `verification_goal`
6. `required_checks`
7. `required_evidence`
8. `preferred_channel`
9. `fallback_channels`
10. `blocking_rules`
11. `warning_tolerance`
12. `manual_review_rules`
13. `evidence_retention_policy`
14. `result_schema_version`

## 5.2 字段解释

### `verification_scope`

用于说明验证范围，例如：

- `task_output_only`
- `module_surface`
- `critical_user_journey`
- `release_readiness`

### `verification_goal`

用于说明目标，例如：

- 构建成功
- 页面真实可操作
- 接口契约未破坏
- 关键 journey 可完成

### `required_checks`

建议每项至少包含：

- `check_id`
- `kind`
- `blocking`
- `executor_brain`
- `channel_constraint`
- `success_criteria`

`kind` 示例：

- `static_validation`
- `artifact_validation`
- `browser_journey`
- `api_assertion`
- `release_gate_check`

### `required_evidence`

建议每项至少包含：

- `evidence_type`
- `required`
- `blocking`
- `source_brain`
- `accepted_channels`

常见 `evidence_type`：

- `browser_screenshot`
- `browser_trace`
- `ci_result`
- `build_artifact`
- `runtime_log`
- `verification_report`

### `preferred_channel`

建议固定枚举：

- `high_spec_remote`
- `github_actions`
- `browser_evidence`
- `manual_review`

### `fallback_channels`

当首选通道不可用时允许的回退链路。

例如：

```json
[
  "github_actions",
  "manual_review"
]
```

### `blocking_rules`

用于定义一票否决条件，例如：

- 缺少关键证据
- blocking checks 未通过
- release gate 未通过
- 人工放行前置项未满足

### `warning_tolerance`

用于定义可容忍项，例如：

- 非关键 warning 数量上限
- 某些已知低风险波动允许人工确认后继续

### `manual_review_rules`

建议至少能表达：

- 是否必须人工检查
- 哪些风险条件命中后必须人工放行
- 哪些通道结果冲突时必须人工介入

---

## 6. 统一结果设计

## 6.1 `VerificationResult` 建议字段

建议至少包含：

1. `verification_run_id`
2. `contract_id`
3. `channel`
4. `executor_brain`
5. `started_at`
6. `ended_at`
7. `checks`
8. `evidence_refs`
9. `blocking_issue_count`
10. `warning_count`
11. `verdict`
12. `reason_summary`

其中 `verdict` 建议固定为：

- `passed`
- `passed_with_warning`
- `failed`
- `manual_review_required`
- `channel_unavailable`

## 6.2 `checks` 建议字段

每项至少包含：

- `check_id`
- `status`
- `blocking`
- `summary`
- `raw_ref`

`status` 建议固定为：

- `passed`
- `failed`
- `skipped`
- `not_run`

## 6.3 `Evidence` 关联要求

`VerificationResult.evidence_refs` 必须能映射到 `Evidence` 主记录。

每条 `Evidence` 至少应保留：

- `evidence_type`
- `source_brain`
- `artifact_uri`
- `status`
- `generated_at`
- `acceptance_run_id` 或 `verification_run_id`

---

## 7. 与 ProductionAcceptanceProfile 的衔接

这部分必须统一，否则验证和验收会各说各话。

衔接规则：

1. `Verification Contract.required_evidence` 必须能映射到 `ProductionAcceptanceProfile.required_evidence`
2. `Verification Contract.required_checks` 必须能映射到一个或多个 `production_gates`
3. `VerificationResult` 只能证明“这次验证结果如何”，不能直接宣布 `production_passed`
4. `AcceptanceRun` 汇总多个 `VerificationResult` 后，才可推进到最终生产级裁决

建议映射关系：

| Verification Contract 字段 | ProductionAcceptanceProfile 字段 |
|---|---|
| `required_evidence` | `required_evidence` |
| `verification_scope` | `required_surfaces` / `required_journeys` |
| `blocking_rules` | `production_gates` |
| `manual_review_rules` | `manual_release_rules` |

---

## 8. 通道选择规则

## 8.1 默认优先级

在长期目标口径下，默认优先级应为：

1. `high_spec_remote`
2. `github_actions`
3. `browser_evidence`
4. `manual_review`

当前现实环境下，可允许部分任务暂时改为：

1. `github_actions`
2. `browser_evidence`
3. `manual_review`

但必须写进合同，不允许运行时隐式改口径。

## 8.2 通道适用建议

| 通道 | 适合验证什么 | 不适合验证什么 |
|---|---|---|
| `high_spec_remote` | 完整构建、完整测试、重负载验证、正式验收 | 需要即时人工判断的主观体验问题 |
| `github_actions` | 当前阶段远端构建与结构化 CI 验证 | 长期终局环境定位 |
| `browser_evidence` | 页面真实路径、截图、trace、交互可用性 | 纯后端重负载构建 |
| `manual_review` | 争议裁决、人工放行、高风险确认 | 代替所有自动化验证 |

---

## 9. 阶段化使用规则

在工作流阶段里，`Verification Contract` 应这样使用：

1. `reviewing`
   - 检查验证合同是否完整
   - 不完整则阻止进入 compile
2. `executing`
   - 按合同准备未来验证所需产物和证据
3. `accepting`
   - 严格按合同执行验证通道和证据收集
4. `reworking`
   - 根据失败结果调整合同，但不能静默降低标准
5. `completed`
   - 只能在合同要求满足后交给 `completion_adjudication`

---

## 10. 最小 JSON 样例

```json
{
  "contract_id": "vc_task_001",
  "contract_version": "v1",
  "task_id": "task_001",
  "verification_scope": "critical_user_journey",
  "verification_goal": "关键路径可实际完成且证据齐全",
  "required_checks": [
    {
      "check_id": "journey_login_create_publish",
      "kind": "browser_journey",
      "blocking": true,
      "executor_brain": "browser",
      "channel_constraint": "browser_evidence",
      "success_criteria": "journey 完整走通并生成 screenshot 与 trace"
    },
    {
      "check_id": "ci_build",
      "kind": "artifact_validation",
      "blocking": true,
      "executor_brain": "verifier",
      "channel_constraint": "github_actions",
      "success_criteria": "远端构建结果为 success"
    }
  ],
  "required_evidence": [
    {
      "evidence_type": "browser_screenshot",
      "required": true,
      "blocking": true,
      "source_brain": "browser",
      "accepted_channels": [
        "browser_evidence"
      ]
    },
    {
      "evidence_type": "ci_result",
      "required": true,
      "blocking": true,
      "source_brain": "verifier",
      "accepted_channels": [
        "github_actions",
        "high_spec_remote"
      ]
    }
  ],
  "preferred_channel": "github_actions",
  "fallback_channels": [
    "manual_review"
  ],
  "blocking_rules": [
    "missing_required_evidence",
    "blocking_check_failed"
  ],
  "warning_tolerance": {
    "max_non_blocking_warnings": 2
  },
  "manual_review_rules": {
    "required_when_channels_conflict": true,
    "required_when_risk_level_high": true
  },
  "evidence_retention_policy": "keep_until_acceptance_closed",
  "result_schema_version": "v1"
}
```

---

## 11. 与钱学森控制论的对应

这份合同设计本质上是在把“验证”从口头判断改成控制回路的一部分。

对应关系如下：

1. 反馈闭环：`VerificationResult` 是执行后的标准反馈量
2. 时滞处理：远端验证通道天然异步，合同必须声明通道和回退
3. 噪声过滤：原始日志不直接入裁决，先沉淀为结构化 check/evidence
4. 稳定优先：返工只能在失败原因显式化后发生，不能盲目重试
5. 约束求优：在当前资源有限时允许 GitHub Actions 替代，但必须保留长期目标环境

---

## 12. 实施建议

按优先级建议先做这 5 项：

1. 在 `CompiledTask` 落库字段中固定 `verification_contract_json`
2. 在 `acceptance_mapping` 输出中固定 `preferred_verification_channel`
3. 在 `VerificationResult` 结果层保留 `channel`、`executor_brain`、`evidence_refs`
4. 在 `CompletionVerdict` 裁决时强制校验 `manual_release_required` 和 `released_by_human`
5. 在所有文档和界面文案中统一写明：
   - 最终目标环境是高配验证环境
   - GitHub Actions 是当前替代通道
   - 低配开发机不是最终 acceptance 依据

这 5 项完成后，EasyMVP 的验证链路才算真正从“概念性描述”进入“可落库、可执行、可裁决”的状态。
