# EasyMVP 四基础专精大脑阶段调用矩阵

> 更新时间：2026-04-20  
> 上位文档：[钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md)  
> 关联文档：[easymvp-brain-职责与边界定义.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md) / [easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md) / [EasyMVP-三层验证架构说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-三层验证架构说明.md)

---

## 1. 文档目的

这份文档只回答一个问题：

> 在 EasyMVP 的阶段化工作流里，`central`、`code`、`browser`、`verifier`、`fault`、`easymvp-brain` 分别什么时候主导、什么时候辅助、什么时候升级。

它要解决的不是“每个脑擅长什么”这种泛化描述，而是：

- 到了某个阶段，谁应当先出手
- 其他脑什么时候补位
- 产物和证据应该落到什么对象
- 什么情况必须升级为返工、人工检查或故障链路

---

## 2. 总体原则

先固定 6 条铁律，后面矩阵全部按这 6 条执行：

1. `EasyMVP` 负责业务状态机，`brain-v3` 负责多脑运行时。
2. `central` 负责协调、委派、汇总、仲裁，不吞掉领域判断。
3. `easymvp-brain` 负责方案审核、方案编译、返工设计、验收映射、完成裁决。
4. `code / browser / verifier / fault` 只负责各自能力面，不回填成领域脑。
5. 同一阶段可以多脑协同，但必须有一个主导脑，不能“大家都参与但没人负责”。
6. 执行成功、验证通过、验收通过、业务完成是四个不同层级，不能混判。

---

## 3. 阶段主导矩阵

| 工作流阶段 | 主导层 | 主导脑 | 辅助脑 | 典型触发 | 主要输出 | 升级条件 |
|---|---|---|---|---|---|---|
| `designing` | EasyMVP 编排层 | `easymvp-brain` | `central` | 创建项目、生成初始方案、修改草案 | `PlanDraft`、初始风险提示、初始验收要求草案 | 目标不清、范围失控、分类不明、输入冲突 |
| `reviewing` | 领域判断层 | `easymvp-brain` | `central`、`verifier` | 提交审核、进入编译前检查 | `PlanReviewResult`、blocking/advisory issues、compile 决策 | blocking issue 命中、验证要求缺失、角色解析失败 |
| `executing` | 运行时执行层 | `code` | `central`、`browser`、`verifier` | 已编译任务开始执行 | 代码变更、任务日志、交付物、局部运行结果 | 执行失败、产物缺失、环境不满足、证据缺口 |
| `accepting` | 验证与验收层 | `verifier` / `browser` | `central`、`easymvp-brain` | 执行结束、进入验收 | `VerificationResult`、`Evidence`、验收覆盖结果 | blocking gate 未过、关键 journey 缺证据、结果相互矛盾 |
| `reworking` | 故障与重设计层 | `fault` + `easymvp-brain` | `central`、`code`、`verifier` | 验收失败、执行失败、人工驳回 | `fault_summary`、`RepairPlanDraft`、调整后的合同 | 连续失败、风险升高、需要人工检查点 |
| `completed` | 最终裁决层 | `easymvp-brain` | `central`、`verifier` | 验收收口、人工放行完成 | `CompletionVerdict`、完成理由、归档指针 | 验收只功能通过但未达到生产通过、人工放行缺失 |

说明：

- `accepting` 阶段的主导脑不是固定一个，而是根据任务类型在 `browser` 和 `verifier` 之间切换。
- `reworking` 阶段不是“直接让 code 再改一遍”，而是先由 `fault` 做故障分类，再由 `easymvp-brain` 做结构化返工设计。

---

## 4. 分阶段调用规则

## 4.1 `designing`

目标：

- 把模糊目标转成可审核的 `PlanDraft`
- 在最早阶段引入边界、风险、验收要求

调用顺序建议：

1. EasyMVP 汇总项目目标、分类、上下文
2. 调 `easymvp-brain` 生成或修订 `PlanDraft`
3. 必要时调 `central` 做跨脑资源与能力可行性检查

这一阶段不该发生的事：

- 直接写代码
- 直接跑浏览器验收
- 把工具执行细节当成方案本身

主要输出：

- `PlanDraft`
- `artifact_refs`
- 初始 `acceptance hints`
- 初始风险分类

升级条件：

- `PlanDraft` 范围持续膨胀
- 项目分类无法确定
- 关键交付物说不清
- 方案依赖当前系统不存在的能力

---

## 4.2 `reviewing`

目标：

- 判断计划能否进入编译和执行
- 提前挡住会在执行阶段爆炸的问题

调用顺序建议：

1. EasyMVP 调 `easymvp-brain.plan_review`
2. 如存在规则性或只读一致性检查，补调 `verifier`
3. `central` 汇总 review 结果并形成阶段决策

主要输出：

- `PlanReviewResult`
- `blocking_issues`
- `advisory_issues`
- `compile_allowed`

何时让 `verifier` 介入：

- 方案引用了现有文件、配置、接口或约束
- 需要确认现状和计划描述是否一致
- 需要结构化只读核验，而不是领域裁决

升级条件：

- `compile_allowed = false`
- `resolve_failed`
- 风险升级规则命中
- 验证合同或交付合同缺失

---

## 4.3 `executing`

目标：

- 按编译后的合同落实际变更和产物

调用顺序建议：

1. `central` 按 `CompiledTask.brain_kind` 路由任务
2. 代码类任务进入 `code`
3. 如执行中需要页面采证，可补调 `browser`
4. 如执行中需要只读检查，可补调 `verifier`

`code` 负责：

- 代码修改
- 脚本执行
- 局部开发级检查
- 交付物生成

`central` 负责：

- 批次门控
- 脑选择与重试策略
- 输出归一化运行摘要

主要输出：

- 代码差异
- `RunResult`
- `DeliveryResult`
- 执行日志与产物引用

升级条件：

- 任务重试超过阈值
- 交付物未生成
- 工作区污染或资源冲突
- 局部成功但合同目标未满足

---

## 4.4 `accepting`

目标：

- 把“执行完成”收成“证据齐全且验收通过”

主导脑选择规则：

1. 页面/交互/截图/用户路径验证优先 `browser`
2. 结构化断言、配置一致性、接口/产物核验优先 `verifier`
3. 跨证据归并和完成语义裁决交给 `easymvp-brain`

调用顺序建议：

1. EasyMVP 根据 `VerificationContract` 选择验证通道
2. `browser` 或 `verifier` 产出归一化 `Evidence` / `VerificationResult`
3. 必要时 `central` 聚合多脑结果
4. EasyMVP 调 `easymvp-brain.completion_adjudication`

主要输出：

- `Evidence`
- `VerificationResult`
- 覆盖率与 blocker/warning 汇总
- `CompletionVerdict`

升级条件：

- 必要证据缺失
- 多个验证通道结果冲突
- `functional_passed = true` 但 `production_passed = false`
- 必须人工放行但没有 `released_by_human`

---

## 4.5 `reworking`

目标：

- 把失败纳入闭环，而不是临时补丁式重试

调用顺序建议：

1. EasyMVP 归一化失败上下文
2. 调 `fault` 生成 `fault_summary`
3. 调 `easymvp-brain.repair_design`
4. 如需要，重新进入 `reviewing -> executing`

`fault` 负责：

- 失败分类
- 故障链路定位
- 恢复建议
- 是否可自动恢复的判断

`easymvp-brain` 负责：

- 把故障摘要改写为 `RepairPlanDraft`
- 判断局部返工还是整体重编译
- 决定是否设置人工检查点

主要输出：

- `fault_summary`
- `RepairPlanDraft`
- `verification_adjustments`
- `delivery_adjustments`

升级条件：

- 同一故障反复出现
- 返工范围扩大到跨批次或跨模块
- 风险等级升级到 `high`
- 验证通道本身失真，无法继续自动闭环

---

## 4.6 `completed`

目标：

- 把最终结果收成可审计、可解释的完成状态

调用顺序建议：

1. EasyMVP 汇总 `RunResult`、`DeliveryResult`、`VerificationResult`、`EvidenceSummary`
2. 调 `easymvp-brain.completion_adjudication`
3. `central` 记录运行时归档与 replay 索引

主要输出：

- `CompletionVerdict`
- 最终完成/返工/阻塞/人工检查点决策
- 证据链归档索引

完成判定最低要求：

1. `executor_succeeded = true` 不是充分条件
2. `delivery_verified = true` 不是充分条件
3. `acceptance_passed = true` 仍需结合人工放行规则
4. 只有 `completed = true` 才算业务闭环完成

---

## 5. 脑级职责边界速查表

| 脑/层 | 应做 | 不应做 |
|---|---|---|
| `central` | 路由、委派、汇总、仲裁、批次门控 | 替代领域脑做业务语义裁决 |
| `easymvp-brain` | 审核、编译、返工设计、验收映射、完成裁决 | 直接写代码、直接跑浏览器、直接做通用故障排查 |
| `code` | 修改代码、执行开发级检查、生成交付物 | 决定业务是否完成 |
| `browser` | 页面操作、截图、trace、真实路径采证 | 代替 verifier 做规则裁决 |
| `verifier` | 只读核验、断言、检查、验证结果归一化 | 代替领域脑决定返工策略 |
| `fault` | 故障分类、恢复建议、失败路径判断 | 直接把修复方案当成最终返工计划 |
| `EasyMVP` | 状态机、持久化、合同挂载、人工审批、证据归档 | 把多脑协作细节写死成单执行器分支 |

---

## 6. 与控制论原则的映射

这份矩阵不是组织图，而是控制结构。

对应关系如下：

1. 闭环优先：`executing -> accepting -> reworking -> executing`
2. 分层控制：EasyMVP / `easymvp-brain` / `brain-v3` / 工具执行明确分层
3. 最小互扰：`code / browser / verifier / fault` 不互相吞职责
4. 时滞显式处理：验收与远端验证不假设即时返回
5. 噪声过滤：原始日志先归一化，再进入领域裁决
6. 稳定优先：先通过 `fault + repair_design` 收口，再决定是否重试

---

## 7. 实施建议

按优先级建议先做这 4 项：

1. 在 `CompiledTask` 中固定写入 `brain_kind`、`role_type`、`verification_contract`
2. 在 Workflow 编排层显式记录每阶段的 `primary_brain` 和 `supporting_brains`
3. 把 `fault -> easymvp-brain -> RepairPlanDraft` 固定成返工唯一路径
4. 把 `completion_adjudication` 作为 `completed` 前最后一道结构化裁决，而不是可选步骤

这 4 项做完，阶段协作关系才算真正从“口头约定”变成“系统基线”。
