# EasyMVP V3 专精大脑接入计划

> 更新时间：2026-04-19  
> 参考对象：`/www/wwwroot/project/brain-v3`  
> 目标：把 `brain-v3` 作为 EasyMVP V3 的首个执行底座接入，并设计配套的 `easymvp` 专精大脑，支撑 EasyMVP 从“任务执行器编排”升级为“方案编译 + 多脑协作执行”。

> 最新口径说明：
> 这份文档保留为 V3 接入设计背景稿。当前 EasyMVP 最新顶层边界、合同与阶段协作，统一以 [钱学森总纲设计/README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md) 为准。
> 其中与本专题最直接对应的最新文档是：
> [easymvp-brain-职责与边界定义.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md)、
> [easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md)、
> [EasyMVP-四基础专精大脑阶段调用矩阵.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-四基础专精大脑阶段调用矩阵.md)。

## 1. 结论

结论先行：

1. `brain-v3` 不是一个“单执行器”，而是一套完整的 Brain Kernel + sidecar 多脑运行时。
2. EasyMVP V3 不应该把它当成 `aider/claude_code/codex_cli` 一类的 `execution_mode` 替代品。
3. 正确接法是：
   - `brain-v3` 负责通用运行时、工具治理、Run 生命周期、sidecar 协作、审计与回放
   - EasyMVP 负责业务工作流、方案编译、任务合同、阶段状态机
   - 新增一个 `easymvp` 专精大脑，承接 EasyMVP 领域特有的“方案审核 / 方案编译 / 返工重构 / 验收判定”能力

一句话：

> `brain-v3` 是执行内核，`easymvp-brain` 是 EasyMVP 的领域大脑，二者不是替代关系，而是上下层关系。

按当前钱学森总纲，需要把这里的旧表达再收紧：

1. 不再建议围绕 `execution_mode` 建模
2. 应改为围绕“脑路由、验证通道、能力可用性、人工检查点”建模
3. 基础专精大脑统一以 `code / browser / verifier / fault` 加中央协调脑为主口径，`easymvp-brain` 负责领域判断

## 2. 对 `brain-v3` 的分析结论

结合 `brain-v3/README.md`、`sdk/docs/32-v3-Brain架构.md`、`sdk/docs/33-Brain-Manifest规格.md`、`sdk/docs/29-第三方专精大脑开发.md`，当前可确认：

### 2.1 它已经具备可作为执行底座的核心能力

- Kernel / Agent Loop / Tool Registry / Guardrail / Persistence 已完整存在
- 支持 `brain run / status / resume / cancel / logs / replay / serve`
- 支持 sidecar 进程模型与多脑协作
- 支持 `native / mcp-backed / hybrid / remote` runtime 抽象
- 支持 file policy、restricted mode、workdir confinement、tool profile
- 已有 `central / code / browser / verifier / fault / data / quant` 等脑模型
- 具备 Run 生命周期和审计回放能力，适合做 EasyMVP 的任务执行承载层

### 2.2 它的顶层对象是 Brain，不是 Task

`brain-v3` 的设计中心是：

- Brain
- Manifest
- Runtime
- Package
- Capability

而不是：

- workflow
- domain_task
- review issue
- accept run

这意味着它擅长的是：

- 运行一个脑
- 调度一个脑
- 跨脑委托
- 用工具完成任务

它不直接理解 EasyMVP 的：

- design/review/execute/rework/accept 阶段
- plan_version / blueprint / domain_task
- acceptance / verification / delivery gate

所以 EasyMVP 不能把自己的领域语义硬塞进 Brain Kernel；应该用领域大脑对接。

### 2.3 对 EasyMVP 最有价值的不是 CLI，而是 Run API 和 sidecar 模式

`brain-v3` 虽然提供了 `brain run` CLI，但对 EasyMVP 来说更有价值的是：

- `brain serve` 的 HTTP Run API
- sidecar 发现与多脑协作
- `status/logs/replay` 生命周期接口
- file policy / workdir policy / tool scope

结论：

> EasyMVP V3 首接 `brain-v3` 时，应该优先对接服务模式和 Run 生命周期，不应该只把 CLI 当 shell 命令跑。

## 3. 为什么 EasyMVP 需要自己的专精大脑

EasyMVP 当前暴露的问题，本质不是“缺一个更强的代码执行器”，而是缺一个懂 EasyMVP 领域语义的脑。

现有痛点包括：

1. 架构师输出直接落任务，没有中间编译层
2. review / execute / rework / accept 的决策散落在多个服务和补丁里
3. warning、执行器可用性、交付物、验证证据，没有在方案阶段统一处理
4. rework 和初始计划不是同一条机制
5. 任务完成判定混淆了“模型成功”和“交付成功”

这些都不是 `code brain` 或 `browser brain` 能直接解决的。

所以 EasyMVP V3 需要一个单独的领域脑：

- 名称建议：`easymvp-brain`
- kind 建议：`easymvp`
- 定位：Workflow Domain Brain

它不是通用代码脑，而是：

> 专门理解 EasyMVP 的计划、审核、编译、返工、验收语义的业务大脑。

## 4. V3 顶层分层

建议把 EasyMVP V3 分成三层：

### 4.1 领域层：EasyMVP Orchestrator

保留在 EasyMVP 主仓内，负责：

- 项目 / 工作流 / 阶段状态机
- plan draft / compiled plan / domain task 的持久化
- delivery / acceptance / verification 业务规则
- 人工介入、审批、展示、审计入口

### 4.2 领域脑层：EasyMVP Brain

新增 `easymvp-brain`，负责：

- 方案审核
- 方案编译
- 返工方案重构
- 任务完成合同判定
- 验收证据要求生成

它是 EasyMVP 业务语义的 LLM / reasoning 层封装。

### 4.3 执行底座层：brain-v3

由 `brain-v3` 提供：

- Run 生命周期
- 多脑调度
- sidecar / tool / sandbox / file policy
- 日志 / replay / 观测
- code brain / browser brain / verifier brain 等基础脑

## 5. EasyMVP V3 的推荐架构

```text
EasyMVP Web/API
    ↓
EasyMVP Workflow Orchestrator
    ├─ 项目/阶段/审批/验收状态机
    ├─ PlanDraft / PlanReview / CompiledPlan / DomainTask 持久化
    └─ 调用 brain-v3 Run API
             ↓
        brain-v3 Kernel
             ├─ easymvp-brain        ← EasyMVP 领域脑
             ├─ code-brain           ← 代码实现
             ├─ browser-brain        ← 浏览器执行/采证
             ├─ verifier-brain       ← 只读验证
             └─ central-brain        ← 通用协调
```

关键点：

1. EasyMVP 主系统仍是状态机和业务主控
2. `brain-v3` 只承接脑执行和运行时治理
3. `easymvp-brain` 负责业务认知，不让通用 code brain 承担领域判断

## 6. `easymvp-brain` 的职责定义

`easymvp-brain` 第一版只做 EasyMVP 领域内最关键的 5 件事。

### 6.1 方案审核

输入：`PlanDraft`  
输出：`PlanReviewResult`

关注：

- scope 是否过大
- affected_resources 是否过宽
- 脑能力与验证通道是否可用
- verification/delivery 要求是否缺失
- 风险等级是否过高
- 是否需要 split / drop / override

### 6.2 方案编译

输入：`PlanDraft + PlanReviewResult`  
输出：`CompiledPlan`

负责：

- 把草案任务转换为正式可执行任务
- 为每个任务生成 `delivery_contract`
- 为每个任务生成 `verification_contract`
- 落实脑路由、推荐验证通道与人工检查点
- 在方案阶段完成拆分，不把大任务留到 execute 再炸

### 6.3 返工重构

输入：失败任务上下文 + 失败原因 + 原始任务合同  
输出：`RepairPlanDraft`

要求：

- 不直接回 patch
- 走和初始方案相同的 review / compile 链
- 保证 rework 是结构化重设计，而不是临时文本补救

### 6.4 验收规则映射

输入：项目类型 + 当前产物  
输出：验收/验证要求清单

作用：

- 提前声明这个项目最终需要哪些证据
- 在 design 阶段就把 `.github/workflows/ci.yml`、`.easymvp/ci/latest.json`、浏览器采证等要求下放到任务合同

补充说明：

- 当前可以把 `github_actions` 作为远端替代验证通道写入合同
- 但语义上必须明确它是当前资源受限下的替代方案
- 长期最终通道仍应保留 `high_spec_remote`

### 6.5 任务完成语义裁决

输入：executor 结果 + delivery 结果 + verification 结果  
输出：任务真实完成状态

作用：

- 避免“模型说成功”就直接 completed
- 统一判断 `executor_succeeded / delivery_verified / acceptance_passed / completed`

并且要明确：

- `easymvp-brain` 负责输出结构化裁决
- 最终推进到 `completed` 仍应由 EasyMVP 状态机按裁决对象执行
- 不能把“脑返回 success”偷换成系统完成

## 7. EasyMVP 不该怎么接 `brain-v3`

为了避免旧问题在 V3 重演，以下接法明确不推荐：

### 7.1 不推荐把 `brain-v3` 直接塞进现有 `execution_mode`

错误方式：

- 在旧 `execution_mode` 里新增一个 `brain`
- 然后像 `aider` 一样 `shell_exec brain run --prompt ...`

问题：

- 丢失 Run 生命周期
- 丢失 replay / logs / resume / cancel
- 丢失多脑协作能力
- 退化成一个普通命令行执行器

当前更推荐的口径是：

- 用 `brain_kind` 表达脑路由
- 用 `preferred_verification_channel / fallback_channels` 表达验证通道
- 用 `manual_review_required` 表达人工检查点

### 7.2 不推荐让通用 `code brain` 直接承接 EasyMVP 的 plan review / compile

原因：

- `code brain` 擅长写代码，不擅长理解 EasyMVP 的 workflow contract
- 方案审核、返工重构、验收映射属于领域逻辑，不该由通用代码脑承担

### 7.3 不推荐把 EasyMVP 业务状态机迁进 Brain Kernel

原因：

- Brain Kernel 的抽象中心是 Brain Run，不是业务 workflow
- EasyMVP 的 `design/review/execute/rework/accept` 语义应继续保留在主仓

## 8. 第一阶段推荐接入方式

### 8.1 接入形态

第一阶段建议使用：

- `brain serve`
- EasyMVP 通过 HTTP Run API 发起脑任务
- 不走本地 shell 包装

原因：

- 方便拿到 run_id
- 可以对接 status / logs / replay / cancel / resume
- 后续更容易支持多脑与 Dashboard 对接

### 8.2 首批接入的脑

建议优先接入这 4 类：

1. `easymvp-brain`
   - 方案审核
   - 方案编译
   - rework 设计

2. `code-brain`
   - 正式代码实现任务

3. `verifier-brain`
   - 只读验证与交付核验

4. `browser-brain`
   - 浏览器验收、截图与交互路径采证

这里 `easymvp-brain` 是第一优先级，因为它解决的是根因，不是表层执行。

## 9. EasyMVP V3 建议新增的核心对象

要和 `brain-v3` 协同，EasyMVP 自身建议新增 5 个对象：

### 9.1 `PlanDraft`

架构师的原始方案草案，不直接执行。

### 9.2 `PlanReviewResult`

由 `easymvp-brain` 产出的结构化审核结果。

### 9.3 `CompiledPlan`

通过审核并编译后的正式计划。

### 9.4 `DeliveryContract`

定义任务真正交付完成需要满足的条件，例如：

- required_files
- non_empty_files
- allowed_empty_files
- allowed_scope

### 9.5 `VerificationContract`

定义任务后续必须绑定的验证要求，例如：

- required_check_kinds
- required_evidence
- browser_required
- ci_required

## 10. `easymvp-brain` 的 Manifest 建议

第一版建议如下：

```json
{
  "schema_version": 1,
  "kind": "easymvp",
  "name": "EasyMVP Brain",
  "brain_version": "0.1.0",
  "description": "Workflow domain brain for plan review, compile, repair and acceptance reasoning",
  "capabilities": [
    "workflow.plan.review",
    "workflow.plan.compile",
    "workflow.rework.design",
    "workflow.accept.map",
    "workflow.task.contract"
  ],
  "task_patterns": [
    "review plan",
    "compile plan",
    "repair failed task",
    "acceptance mapping"
  ],
  "runtime": {
    "type": "native"
  },
  "policy": {
    "tool_scope": "delegate.easymvp",
    "approval_mode": "default"
  },
  "compatibility": {
    "protocol": "1.0",
    "tested_kernel": "0.7.x"
  },
  "license": {
    "required": false
  }
}
```

## 11. EasyMVP V3 的第一版分期

### Phase 1：先接底座，不改业务模型

目标：

- EasyMVP 能通过 `brain serve` 调起 Run
- 代码实现任务可委托给 `code-brain`
- 拿到 run_id / status / logs / replay

此阶段先不动主状态机，只替换执行承载层。

### Phase 2：引入 `easymvp-brain`

目标：

- 方案审核前置
- 方案编译前置
- rework 走同一条 reviewer/compiler 链

此阶段开始重构 design/review 主链。

### Phase 3：完成状态机升级

目标：

- 从 `executor success = completed` 升级成
  - `executor_succeeded`
  - `delivery_verified`
  - `completed`
- accept / verification / browser evidence 全部合同化

## 12. 对 EasyMVP V3 的最终判断

如果 EasyMVP V3 要真正解决当前暴露出来的根因，正确路径不是：

- 再换一个更强的代码执行器
- 再堆更多 runtime fallback

而是：

1. 用 `brain-v3` 接管执行底座
2. 用 `easymvp-brain` 接管领域思考
3. 用 `PlanDraft -> PlanReview -> CompiledPlan -> DomainTask` 重构工作流前半段

最终目标不是“接入一个新执行器”，而是：

> 让 EasyMVP 从工作流系统升级成一个以 Brain Runtime 为执行底座、以 EasyMVP Brain 为领域大脑的方案编译系统。

## 13. 下一步建议

按优先级建议下一步只做三件事：

1. [EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
2. 冻结 `easymvp-brain` 第一版 Manifest / Tool Schema / Prompt 设计
3. [EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)

在这三件事定下来之前，不建议继续扩展新的 `execution_mode`。
