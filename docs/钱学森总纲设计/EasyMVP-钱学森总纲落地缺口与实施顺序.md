# EasyMVP 钱学森总纲落地缺口与实施顺序

> 更新时间：2026-04-20  
> 上位文档：[钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md)  
> 关联文档：[EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md) / [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md) / [EasyMVP-V3-AgentTeam-Runtime计划.md](/www/wwwroot/project/easymvp/docs/EasyMVP-V3-AgentTeam-Runtime计划.md) / [EasyMVP-V3-AgentTeam状态板.md](/www/wwwroot/project/easymvp/docs/EasyMVP-V3-AgentTeam状态板.md)

---

## 1. 文档目的

这份文档只做一件事：

> 把钱学森总纲下真正还没落地的事项，按优先级、责任泳道、依赖关系重新收口。

它不是再写一次理想架构，而是把“已经有文档口径但还没做完”的部分拉成明确施工单。

---

## 2. 结论先行

当前真正还缺的，不是再新增脑，也不是再写一轮概念文档，而是 5 个闭环缺口：

1. runtime 归一化对象还没有在所有运行时链路完全固化
2. 验证合同还没有在读取侧、事件侧、诊断侧完全贯通
3. 故障回路已经有文档口径，但 `fault -> repair_design -> rework` 的强制收口还不够硬
4. 高配验证环境只是目标口径，当前仍缺正式接入层
5. 桌面端与恢复链路虽然已补第一轮兼容消费，但还没以统一 contract 完整消费这些状态

因此，后续优先级不应再放在“继续补总纲”，而应放在：

1. runtime / adapter
2. diagnostics / recovery
3. acceptance / evidence / replay consumption
4. high-spec verification environment 接入

---

## 3. 已经完成的基线

下面这些已经基本不是主要缺口：

1. 顶层总纲已明确 `brain-v3`、`EasyMVP`、`easymvp-brain` 分层
2. `easymvp-brain` 的职责、输入输出契约已收口
3. 四基础专精大脑阶段调用矩阵已收口
4. `Verification Contract` 统一设计已成文
5. `unsupported / denied` 必须显式保留，这条边界已进入旧计划与状态板
6. `completion_adjudication` / `repair_design` 主链已接通

换句话说：

> 现在的主问题已经不是“想清楚没有”，而是“这些口径有没有彻底进入运行时、页面和恢复机制”。

第二轮收口说明：

1. `apps/core` 已新增 `AcceptanceOverview / WorkspaceOverview / HomeOverview / PlanOverview`
2. `WorkspacePage / AcceptancePage` 已新增第一轮总纲兼容消费
3. `manual_review_required / verification_conflict / fault_loop_detected / policy_denied` 已进入 workspace 页面可见状态
4. 因此当前 P0 的重点，已经从“先把页面接上”转为“把兼容层收口成统一 contract”

---

## 4. 真正未完成的缺口

## 4.1 P0 缺口：runtime 归一化对象全链路固化

目标：

- 把 `RunResult / DeliveryResult / VerificationResult / FaultSummary / RuntimeEscalation` 固化为真正稳定的 adapter DTO

当前问题：

- 已有 `overview` 兼容层，但还没有形成统一对象族
- 一些页面和查询虽然已经接上新字段，仍偏向直接消费各自 view 推导结果，而不是稳定 DTO 语义
- `unsupported / denied` 已保留，但 `channel_unavailable`、`verification_conflict`、`fault_loop_detected` 还未形成同等硬度的统一 persistence 与诊断出口

建议归属：

- `runtime-projection`

直接关联计划：

- `RN-002`
- `DG-001`
- `PF-002`

完成定义：

1. runtime adapter 有固定 DTO
2. 事件层、查询层、页面层用同一套枚举
3. 新状态不会再以字符串散落在不同模块
4. `AcceptanceOverview / WorkspaceOverview / HomeOverview` 不再各自维护半重复语义

## 4.2 P0 缺口：故障回路强制收口

目标：

- 把 `fault -> easymvp-brain.repair_design -> reworking` 变成硬约束，而不是建议链路

当前问题：

- 文档已经明确，但代码层仍可能出现“普通失败直接重试”或“失败后只投 diagnostics”
- 相同故障上下文的幂等收口已有一部分，但还缺完整的故障回路状态定义

建议归属：

- `runtime-projection` + `domain-backend`

直接关联计划：

- `PF-002`
- `DG-001`
- `BE-022` 后续读取侧联调

完成定义：

1. 同类故障重复命中时自动生成 `fault_loop_detected`
2. 自动重试达到阈值后必须进入 repair 链
3. 页面上能看到“为什么进入返工”而不是只看到失败

## 4.3 P0 缺口：验证合同读取侧贯通

目标：

- 让 `verification_contract_json` 不只存在于写侧和文档里，而是进入验收页、诊断页、回放页的真实消费逻辑

当前问题：

- 写侧和部分 adjudication 已有
- 读侧已出现第一轮兼容层，但还没有把“合同要求 vs 实际证据”对比完整展示出来
- `required_checks / required_evidence / preferred_verification_channel / missing_evidence / failed_checks` 仍未稳定固化

建议归属：

- `frontend-workbench` + `runtime-projection`

直接关联计划：

- `AG-003`
- `RN-004`

完成定义：

1. 页面可显示合同要求的 evidence/check 列表
2. 页面可显示哪些是 blocker、哪些是 warning
3. 页面可解释为什么没完成，而不是只给最终状态

## 4.4 P1 缺口：高配验证环境正式接入

目标：

- 把“最终目标验证环境是高配验证环境”从文档原则推进成真实接入能力

当前问题：

- 当前仍主要依赖 GitHub Actions 作为替代通道
- 缺少高配远端验证环境的真实 channel 定义、鉴权、结果回传和证据索引

建议归属：

- `runtime-projection`

直接关联计划：

- 新增专项，当前旧计划中尚未单列

完成定义：

1. `preferred_verification_channel = high_spec_remote` 可真实配置
2. 可以拿到远端验证 run 的结构化结果
3. 结果可落为 `VerificationResult + Evidence`

## 4.5 P1 缺口：恢复模式和诊断模式升级

目标：

- 让恢复链真正消费 runtime 新状态，而不是只展示 generic diagnostics

当前问题：

- 恢复页已有首版，但对策略拒绝、通道不可用、故障回路、人工检查点等分型仍不够细
- 工作台虽然已经能看到四个关键状态位，但 diagnostics/recovery 还没有同步进入同等结构化消费

建议归属：

- `diagnostics-recovery` + `frontend-workbench`

直接关联计划：

- `DG-001`
- `DG-002`
- `SX-002`

完成定义：

1. 恢复页能按 `policy_denied / environment_unavailable / fault_loop_detected / manual_review_required` 分型
2. 每类都有恢复建议和跳转入口
3. 诊断导出中带上结构化 escalation 信息

---

## 5. 责任泳道重排

| 泳道 | 接下来最该做的事 | 不该再做的事 |
|---|---|---|
| `runtime-projection` | 固化 DTO、运行时状态枚举、故障回路、远端验证通道 | 再造一层通用 runtime 内核 |
| `domain-backend` | 把 repair / adjudication / acceptance 读取侧继续接实 | 把 `code / browser / verifier / fault` 回填进领域脑 |
| `frontend-workbench` | 显示 verification contract、evidence gap、escalation reason、rework cause | 只展示最终状态、不解释原因 |
| `diagnostics-recovery` | 强化分型恢复、诊断导出、状态闭环 | 继续停留在 generic error 页面 |

---

## 6. 推荐实施顺序

建议按下面顺序推进，不要乱序：

1. `RN-002 + DG-001`
   - 先固化 runtime DTO 和错误域
2. `PF-002`
   - 再补幂等重试与事件去重，顺手把故障回路收口
3. `RN-004 + AG-003`
   - 再把 replay / evidence / contract 读取侧打通
4. `DG-002 + SX-002`
   - 再强化恢复模式和诊断导出
5. `high_spec_remote` 专项
   - 最后正式接高配验证环境

原因：

1. 不先固定 runtime 对象，后面页面和恢复页会继续漂
2. 不先固定故障回路，返工闭环就仍然会软
3. 不先打通读侧，`Verification Contract` 仍会停在纸面

---

## 7. 建议新增的正式计划项

旧计划里建议新增这 3 项，否则总纲中的关键缺口没有落点：

1. `RT-004`
   - 名称：固化 runtime DTO 与 escalation 枚举
   - 归属：`runtime-projection`
2. `RT-005`
   - 名称：接入 high-spec remote verification channel
   - 归属：`runtime-projection`
3. `FE-021`
   - 名称：验收页/诊断页显示 verification contract gap 与 escalation reason
   - 归属：`frontend-workbench`

---

## 8. 完成标准

钱学森总纲真正落地，不是看文档数量，而是看这 5 条是否都成立：

1. 运行时所有关键状态都能结构化表达
2. 故障会进入显式返工回路，而不是靠人工猜
3. 验证合同能在页面上解释“要求了什么、缺了什么”
4. GitHub Actions 被明确视为当前替代通道，而不是长期终局
5. 恢复模式能解释为什么失败、为什么升级、下一步怎么做

只要这 5 条还没全部满足，钱学森总纲就还没有真正进入系统实现层。
