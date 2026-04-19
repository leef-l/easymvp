# EasyMVP 术语与枚举统一表

> 更新时间：2026-04-20  
> 上位文档：[README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md)  
> 关联文档：[EasyMVP-对象级字段清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-对象级字段清单.md) / [EasyMVP-页面读取与展示清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-页面读取与展示清单.md) / [EasyMVP-闭环状态机补充说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-闭环状态机补充说明.md)

---

## 1. 文档目的

这份文档只做统一，不再扩展设计。

它用于固定：

1. 术语到底怎么叫
2. 枚举到底有哪些
3. 页面和实现层该用哪个词
4. 哪些相近词不再混用

后续 `钱学森总纲设计` 目录里的文档，如出现冲突，以本表为准。

---

## 2. 核心对象统一名

以下对象名固定，不再换说法：

| 统一名 | 含义 | 不再推荐的混用说法 |
|---|---|---|
| `CompiledTask` | 编译后的正式执行任务 | compiled item、final task |
| `VerificationResult` | 某次验证执行的结构化结果 | verify result、验收结果 |
| `CompletionVerdict` | 最终完成裁决对象 | final result、completion result |
| `FaultSummary` | 故障归一化摘要 | failure summary、diagnostic summary |
| `RepairPlanDraft` | 返工草稿 | repair draft result、rework draft |
| `RuntimeEscalation` | 运行时升级对象 | escalation event、runtime alert |

---

## 3. 状态层级统一口径

以下 4 个状态层级必须严格区分：

| 统一字段 | 含义 | 不等于 |
|---|---|---|
| `executor_succeeded` | 执行器这次跑完了 | 交付完成、验收通过、业务完成 |
| `delivery_verified` | 交付合同达成 | 验收通过、业务完成 |
| `acceptance_passed` | 验收合同达成 | 业务完成 |
| `completed` | 业务闭环完成 | run success |

固定规则：

1. `executor_succeeded = true` 不代表 `completed = true`
2. `acceptance_passed = true` 也不一定能自动 `completed = true`
3. `completed = true` 必须来自 `CompletionVerdict`

---

## 4. 工作流状态统一名

工作流状态固定为：

1. `designing`
2. `reviewing`
3. `executing`
4. `accepting`
5. `reworking`
6. `paused`
7. `completed`
8. `failed`
9. `canceled`

不再推荐在总纲文档中混用：

1. `running` 代替 `executing`
2. `accept` 代替 `accepting`
3. `rework` 代替 `reworking`

说明：

- 动词缩写可用于内部回调名。
- 面向设计文档和页面口径，优先使用完整阶段名。

---

## 5. 运行结果状态统一名

`RunResult.status` 固定为：

1. `completed`
2. `failed`
3. `unsupported`
4. `denied`
5. `cancelled`
6. `timeout`

统一解释：

| 状态 | 含义 |
|---|---|
| `completed` | 本次运行结束，但不代表业务完成 |
| `failed` | 本次运行失败 |
| `unsupported` | 当前能力或协议不支持 |
| `denied` | 被策略、权限、范围或文件规则拒绝 |
| `cancelled` | 运行被取消 |
| `timeout` | 运行超时 |

不再推荐的混用说法：

1. `success`
2. `run_succeeded`
3. `done`

这些词可做说明文案，但不作为统一枚举。

---

## 6. 交付状态统一名

`DeliveryResult.delivery_status` 固定为：

1. `delivered`
2. `partially_delivered`
3. `not_delivered`

不再推荐混用：

1. `success`
2. `completed`
3. `delivery_done`

---

## 7. 验证结果统一名

`VerificationResult.verdict` 固定为：

1. `passed`
2. `passed_with_warning`
3. `failed`
4. `manual_review_required`
5. `channel_unavailable`

统一解释：

| verdict | 含义 |
|---|---|
| `passed` | 验证通过 |
| `passed_with_warning` | 验证通过但存在可容忍 warning |
| `failed` | 验证不通过 |
| `manual_review_required` | 需要人工检查，不能自动收口 |
| `channel_unavailable` | 验证通道不可用 |

不再推荐混用：

1. `success`
2. `accepted`
3. `verification_passed`

---

## 8. 完成裁决统一名

`CompletionVerdict.decision` 固定为：

1. `complete`
2. `rework`
3. `blocked`
4. `manual_checkpoint`

统一解释：

| decision | 含义 |
|---|---|
| `complete` | 可以判定完成 |
| `rework` | 必须进入返工 |
| `blocked` | 当前阻塞，不能继续自动推进 |
| `manual_checkpoint` | 必须等待人工处理 |

不再推荐混用：

1. `manual_review`
2. `retry`
3. `failed`

这些词可以解释原因，但不作为 `decision` 枚举。

---

## 9. 升级类型统一名

`RuntimeEscalation.escalation_type` 固定为：

1. `retry_exhausted`
2. `unsupported_capability`
3. `policy_denied`
4. `verification_conflict`
5. `manual_review_required`
6. `environment_unavailable`
7. `fault_loop_detected`

统一解释：

| escalation_type | 含义 |
|---|---|
| `retry_exhausted` | 自动重试次数耗尽 |
| `unsupported_capability` | 能力或协议不支持 |
| `policy_denied` | 策略或权限拒绝 |
| `verification_conflict` | 多验证结果冲突 |
| `manual_review_required` | 必须人工检查 |
| `environment_unavailable` | 目标环境或通道不可用 |
| `fault_loop_detected` | 故障回路形成，不允许继续盲重试 |

说明：

- 文档里出现 `unsupported / denied` 时，指的是 `RunResult.status`。
- 页面和诊断层若展示升级原因，优先使用 `unsupported_capability / policy_denied` 这类升级枚举。

---

## 10. 验证通道统一名

`preferred_verification_channel` 与相关通道枚举固定为：

1. `high_spec_remote`
2. `github_actions`
3. `browser_evidence`
4. `manual_review`

统一解释：

| 通道 | 含义 |
|---|---|
| `high_spec_remote` | 长期目标中的高配正式验证环境 |
| `github_actions` | 当前阶段的远端替代验证通道 |
| `browser_evidence` | 浏览器采证验证通道 |
| `manual_review` | 人工裁决或人工放行通道 |

固定规则：

1. `high_spec_remote` 是长期正确口径
2. `github_actions` 是当前替代通道
3. 页面不得把 `github_actions` 展示成长期终局

---

## 11. 页面文案统一口径

页面文案统一采用以下中文口径：

| 英文/对象 | 中文展示建议 |
|---|---|
| `VerificationResult` | 验证结果 |
| `CompletionVerdict` | 完成裁决 |
| `FaultSummary` | 故障摘要 |
| `RepairPlanDraft` | 返工草稿 |
| `RuntimeEscalation` | 升级事项 |
| `manual_checkpoint` | 人工检查点 |
| `high_spec_remote` | 高配验证环境 |
| `github_actions` | GitHub Actions 替代通道 |

不再推荐的页面文案：

1. “运行成功=已完成”
2. “验证通过=可发布”
3. “返工=重试一下”

---

## 12. 推荐阅读映射

如果要查：

1. 总方向：看总方案
2. 对象字段：看对象级字段清单
3. 页面展示：看页面读取与展示清单
4. 自动推进和阻断：看闭环状态机补充说明
5. 验证通道：看 Verification Contract 统一设计

---

## 13. 一句话结论

EasyMVP 这组文档后续要保持稳定，关键不是继续加词，而是：

> 同一个对象只叫一个名字，同一种状态只用一套枚举，同一种结果在页面、文档、实现里都说同一种话。
