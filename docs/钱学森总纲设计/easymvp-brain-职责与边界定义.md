# easymvp-brain 职责与边界定义

> 更新时间：2026-04-20  
> 上位文档：[钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md)  
> 关联技能：[easymvp-control-doctrine](/www/wwwroot/project/easymvp/skills/easymvp-control-doctrine/SKILL.md) / [easymvp-brain-router](/www/wwwroot/project/easymvp/skills/easymvp-brain-router/SKILL.md)

---

## 1. 文档目的

这份文档只解决一件事：

> **`easymvp-brain` 到底负责什么，不负责什么。**

它用于收口以下混淆：

- `brain-v3` 中央大脑与 `easymvp-brain` 的边界
- 审核大脑、故障大脑与 `easymvp-brain` 的边界
- EasyMVP 业务层与领域脑之间的边界

---

## 2. 一句话定义

`easymvp-brain` 是：

> **EasyMVP 的领域专精大脑，负责对计划、返工、验收和完成语义做结构化业务判断。**

它不是通用执行脑，不是浏览器脑，不是代码脑，也不是故障脑的替代品。

---

## 3. 它解决的核心问题

EasyMVP 当前真正难的，不是“有没有工具执行代码”，而是下面这些业务判断：

1. 一个草案计划是否合理
2. 一个计划是否需要拆分、降级、改路由
3. 哪些证据才足以支撑 acceptance
4. 一个失败应该重试、返工还是人工接管
5. 一个任务是“运行成功”还是“业务完成”

这些问题都不是 `code`、`browser`、`verifier`、`fault` 自己应该单独决定的。

它们属于：

- 领域判断
- 合同编译
- 业务裁决

因此必须由 `easymvp-brain` 承担。

---

## 4. 核心职责

第一版 `easymvp-brain` 只做 5 件事，不能扩散。

## 4.1 方案审核

输入：

- `PlanDraft`
- `Blueprint`
- 分类 profile
- 项目上下文

输出：

- `PlanReviewResult`

它要判断：

- scope 是否过大
- affected resources 是否过宽
- 风险是否过高
- 是否缺失验证要求
- 是否缺失交付要求
- 是否需要 split / drop / override

它不负责直接执行任何任务。

## 4.2 方案编译

输入：

- `PlanDraft`
- `PlanReviewResult`
- category profile
- role context

输出：

- `CompiledPlan`
- `CompiledTask[]`
- `DeliveryContract`
- `VerificationContract`

它要把“可以讨论的方案”编译成“可以执行的合同”。

## 4.3 返工重构

输入：

- 失败上下文
- 原始合同
- 故障摘要
- 当前证据

输出：

- `RepairPlanDraft`

它负责：

- 把失败从“补丁式修修看”拉回结构化返工
- 明确返工后的任务边界、验证目标和交付要求

## 4.4 验收规则映射

输入：

- 项目分类
- 当前阶段产物
- 当前系统支持的验证通道

输出：

- `AcceptanceProfile`
- `ProductionAcceptanceProfile`
- required evidence 清单

它负责声明：

- 什么证据是必须的
- 什么验证是必要的
- 什么 warning 可以容忍
- 什么 blocker 必须中断 acceptance

## 4.5 完成语义裁决

输入：

- `RunResult`
- `DeliveryResult`
- `VerificationResult`
- `EvidenceSummary`

输出：

- `CompletionVerdict`

它负责把这些状态区分开：

- `run_succeeded`
- `delivery_verified`
- `accepted`
- `completed`

避免系统把“执行器说成功”误判成“业务完成”。

---

## 5. 明确不负责的事项

`easymvp-brain` 明确不负责以下事项。

## 5.1 不直接写代码

这些交给：

- 代码大脑 `code`
- 或其他执行器

## 5.2 不直接做浏览器操作

这些交给：

- 浏览器大脑 `browser`

## 5.3 不直接做通用只读验证

这些交给：

- 审核大脑 `verifier`

## 5.4 不直接做通用故障诊断

这些交给：

- 故障大脑 `fault`

## 5.5 不直接掌握业务状态机推进

这些交给：

- EasyMVP Workflow Orchestrator

也就是说：

> `easymvp-brain` 提建议、出合同、做裁决；  
> EasyMVP 负责落状态，brain-v3 负责跑脑。

---

## 6. 与其他层的边界

## 6.1 与 EasyMVP 的边界

EasyMVP 负责：

- 工作流状态机
- 数据持久化
- Action Inbox
- Acceptance / Evidence / Issue 表
- 人工审批与前端展示

`easymvp-brain` 负责：

- 给出领域判断结果
- 产出结构化合同与裁决结果

因此：

- EasyMVP 不能把业务裁决偷懒下放给执行脑
- `easymvp-brain` 也不能直接改状态表

## 6.2 与 `brain-v3 central` 的边界

`central` 负责：

- 理解任务
- 路由脑
- 汇总脑结果
- 仲裁执行链条

`easymvp-brain` 负责：

- EasyMVP 领域规则
- 合同编译
- 业务完成语义

因此：

- `central` 不是 `easymvp-brain`
- `easymvp-brain` 不是统一 orchestrator

## 6.3 与审核大脑的边界

审核大脑负责：

- 只读核验
- 断言
- 结果检查

`easymvp-brain` 负责：

- 这次为什么需要这些检查
- 哪些检查结论足以进入 acceptance
- 哪些 warning 仍可继续

## 6.4 与故障大脑的边界

故障大脑负责：

- 分类失败
- 解释失败
- 给出恢复方向

`easymvp-brain` 负责：

- 把恢复方向转成结构化 RepairPlanDraft
- 决定返工在业务上的进入方式

标准链路：

```text
fault -> easymvp-brain -> RepairPlanDraft
```

---

## 7. 触发条件

应触发 `easymvp-brain` 的场景：

1. 新方案生成后，需要审核
2. review 通过后，需要编译成正式任务
3. acceptance 前，需要声明验收规则
4. run / verify / accept 失败后，需要结构化返工
5. 任务处于“看起来成功但业务是否完成不确定”状态时

不应触发 `easymvp-brain` 的场景：

1. 单纯代码改动执行
2. 单纯页面截图和浏览器交互
3. 单纯只读验证
4. 单纯通用故障分析

---

## 8. 输出要求

`easymvp-brain` 的输出必须满足：

1. 可跨进程序列化
2. 可直接落库
3. 可被工作台解释
4. 可用于 replay 与审计
5. 不依赖本地隐式上下文

标准形式：

- envelope
- `decision_summary`
- `result_json`
- `source_refs`

---

## 9. 建设原则

建设 `easymvp-brain` 时要坚持：

1. **小而硬**
   - 只做领域判断，不扩大到通用执行
2. **只吃归一化输入**
   - 不直接吃原始工具返回
3. **只出结构化输出**
   - 不把自然语言长段落作为主输出
4. **不和四个基础专精大脑抢职责**
   - 代码、浏览器、审核、故障各归其位

---

## 10. 最终定义

如果把这份文档压缩成一句话：

> **`easymvp-brain` 是 EasyMVP 的领域裁决与合同编译脑，不是执行脑、验证脑或故障脑。**
