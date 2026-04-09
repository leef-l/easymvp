# EasyMVP全面分析与优化路线图

> 更新日期：2026-04-09
>
> 用途：用于管理汇报、产品与研发对齐、90 天执行落地参考。

## 1. 执行摘要

EasyMVP 当前最准确的定位，不是“AI 编码助手”，而是“AI 软件交付编排平台”。

从仓库实现、联调结果和前端工作台形态看，系统已经具备一条完整主链：

- 需求对话
- 任务拆解
- 审核
- 执行
- 验收
- 返工
- 人工接管
- 飞书 / Telegram 协作

这说明 EasyMVP 已经跨过“能不能做”的阶段，进入“能不能稳定交付、能不能规模化治理”的阶段。

因此，下一阶段的核心优化方向不应继续停留在“再接一个模型、再加几个页面”，而应升级为：

```text
从 AI 工作流系统
升级为 AI 软件交付控制台
```

具体表现为：

```text
任务完成
-> diff / patch / PR
-> CI / 规则校验 / 人工复核
-> 自动修复 / 返工
-> 最终交付闭环
```

这是未来 90 天最值得投入的主方向。

## 2. 当前项目判断

### 2.1 当前定位

根据当前代码和文档，EasyMVP 的核心能力不是 IDE 内即时补全，而是围绕项目级交付进行编排和治理。

可以直接支撑该判断的内部依据：

- [EasyMVP架构设计文档](EasyMVP架构设计文档.md)
- [WorkflowRun阶段化工作流引擎重构架构设计文档](WorkflowRun阶段化工作流引擎重构架构设计文档.md)
- [执行器接入架构设计文档](执行器接入架构设计文档.md)
- [GitWorktree任务级环境隔离设计文档](GitWorktree任务级环境隔离设计文档.md)
- [EasyMVP研发执行版](EasyMVP研发执行版.md)
- [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md)

### 2.2 已经形成的核心能力

当前系统已经形成以下基础能力。

1. Workflow V2 主链编排

- 设计、审核、执行、验收、返工、完成已形成明确阶段链。
- 工作流装配中心位于 [registry.go](../admin-go/app/mvp/internal/workflow/orchestrator/registry.go)。

2. 统一执行器抽象

- 当前已注册 `aider`、`chat`、`openhands`、`claude_code`、`codex_cli`、`gemini_cli`、`auto`。
- 执行器抽象已与调度器主逻辑解耦。

3. 任务级工作区隔离

- 写仓执行器已接入 `git worktree`。
- 任务目录规则为 `{work_dir}/.mvp-worktrees/task-{task_id}`。

4. 人工接管与返工能力

- 已支持审核驳回、验收驳回、运行中重置、完成态重开、人工放行。
- 这类能力对企业场景非常关键。

5. 协作控制面

- 前端已不是单页 demo，而是分出了审核、执行、验收、返工、自治、时间线、飞书、Telegram 等页面。

### 2.3 已验证的关键事实

截至 2026 年 4 月 8 日晚间回归，Workflow V2 主链已经能在纯后端环境完整收口，链路为：

```text
create-project
-> chat/send
-> parse-tasks
-> confirm-plan
-> review
-> execute
-> 人工 update-domain-task / retry / running 中接管
-> accept manual_review
-> 人工 accept-approve
-> complete
```

依据见 [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md)。

这说明：

- 主架构方向是成立的。
- 编排骨架已经具备可用性。
- 下一步瓶颈不在“有没有链路”，而在“链路是否足够稳定、足够可控、足够可审计”。

最新的验证口径与收尾状态，统一以 [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md) 为准，不再单独维护一次性联调现场记录。

## 3. 当前核心问题

### 3.1 产品层问题：结果形态不够完整

当前系统更偏“任务完成态”，而不是“交付完成态”。

也就是说，系统已经会推进任务，但还没有把结果稳定沉淀为：

- 可审查的改动摘要
- 可回放的 patch / diff
- 可落地的 PR
- 可量化的 CI / 验收证据
- 可追踪的交付决策

这会导致产品在演示上已经足够，但在企业落地中仍然偏重人工接管。

### 3.2 架构层问题：控制面耦合偏高

当前阶段推进、自治决策、watchdog、恢复逻辑、执行器注册都集中在编排装配层中。

典型体现：

- [registry.go](../admin-go/app/mvp/internal/workflow/orchestrator/registry.go) 同时承担执行器注册、阶段回调、自治接入、恢复回调、watchdog 接线。
- [workflow.go](../admin-go/app/mvp/internal/controller/chat/workflow.go) 已经达到超大文件规模。

这意味着：

- 当前还能维护，是因为核心团队仍掌握全局上下文。
- 继续叠加能力后，维护成本、评审成本、回归成本都会快速上升。

### 3.3 运行时问题：协议抽象和执行器路由还不够产品化

联调记录已经明确暴露出多个运行时结构问题：

1. `provider_type` 单值设计无法表达多协议供应商。
2. `base_url` 单字段无法表达不同协议使用不同端点。
3. CLI 与后端 API 共用解析逻辑时，会出现 endpoint 相互污染。
4. 同一供应商既支持 Anthropic 风格协议又支持 OpenAI 兼容协议时，当前配置模型表达不足。

这些问题不只是“适配一个新供应商麻烦”，而是会直接限制后续执行器稳定性和可扩展性。

### 3.4 工程层问题：稳定性主要靠收敛策略，而不是系统性验证

联调文档中已经出现以下信号：

- 调度并发已主动收敛为 `1`
- root 环境下需避开部分执行器
- `coordinator_optimize` 会出现非阻断失败
- `aider` 存在越界生成倾向
- `TaskReset` 清理残留 worktree 时仍有异常场景

这说明当前系统已经能靠工程修补跑通，但还没有建立足够强的“自动回归 + 运行时保护 + 失败归因”体系。

### 3.5 治理层问题：缺少统一轨迹、审计和评测系统

系统已经有时间线、决策动作、人工检查点等基础数据，但还没有形成真正的产品级治理面：

- 一个任务从需求到交付是否可完整回放
- 哪个模型、哪个执行器、哪个提示词导致了问题
- 哪次人工接管改变了结果
- 哪类任务最容易触发返工
- 哪类任务适合自动回写，哪类任务必须人审

如果没有这层治理，系统的企业价值会明显受限。

## 4. 战略判断

EasyMVP 当前最合理的战略方向是：

```text
从“统一执行器平台”
升级为“AI 软件交付控制台”
```

这意味着产品目标要从“让 Agent 跑起来”，转向“让交付过程可控、可审计、可回放、可放权”。

一句话表达：

```text
让 EasyMVP 不只会做任务，
而是会负责把任务安全地交付出去。
```

## 5. 外部前沿与竞品判断

截至 2026 年 4 月 9 日，行业前沿已经出现几个非常清晰的趋势。

### 5.1 单代理正在升级为代理控制台

代表产品：

- OpenAI Codex / Codex App
- Devin
- Claude Code

共同方向：

- 并行代理
- 长周期任务
- worktree / sandbox
- 自动化任务编排
- 轨迹与可观察性

### 5.2 “写代码”正在升级为“PR / Review / Autofix / Merge”

前沿产品已经不再只强调“会改代码”，而是强调：

- 能生成可审查的结果
- 能处理 review comments
- 能回接 CI
- 能自动修复
- 能形成最终交付闭环

这对 EasyMVP 的启发非常直接：

当前最值得建设的，不是再增加一个执行器，而是把验收做成真正的 PR / CI / Review 闭环。

### 5.3 安全与治理正在成为一等能力

无论是 Claude Code 的 sandboxing，还是 GitHub 的 cloud agent，核心趋势都很明确：

- 文件系统边界
- 网络边界
- 权限开关
- 审计日志
- 结果可追溯

EasyMVP 既然走的是企业交付路线，这部分不能长期停留在“命令模板 + 经验控制”。

### 5.4 评测方式正在从静态 benchmark 转向真实交付回放

前沿评测已不再适合只依赖静态、单语言、固定样本的老 benchmark。

对 EasyMVP 更有意义的评测体系应该是：

- 历史项目回放
- 失败案例回放
- 典型项目模板回放
- 多语言、多任务类型对比

## 6. 核心优化方向

### 6.1 方向一：把验收升级成真正的交付闭环

目标不是“验收打分”，而是“交付完成”。

建议输出三类结果形态：

1. `patch`

- 适合低风险、小范围、可自动应用的任务。

2. `PR`

- 适合中高风险、需人审、需挂 CI 的任务。

3. `manual`

- 适合高风险、边界不明、或涉及多资源冲突的任务。

验收阶段应逐步从“规则 + LLM 判断”升级为：

```text
规则校验
+ CI 结果
+ 变更摘要
+ Review 评论
+ 自动修复
+ 最终放行/返工
```

这是整个优化路线中优先级最高的一项。

### 6.2 方向二：明确 worktree 结果的正式沉淀策略

当前文档与实现在“是否自动回写主工作区”上存在认知不一致。

因此必须尽快明确：

1. 哪些任务允许自动 patch 回写
2. 哪些任务必须生成 PR
3. 哪些任务只能保留 worktree 结果待人工处理
4. 回写失败、冲突、越界时如何自动降级

建议分三阶段落地：

1. 低风险任务支持受控 patch 回写
2. 中风险任务默认生成 PR
3. 高风险任务只保留结果和证据，不自动落主分支

### 6.3 方向三：重构 provider 与执行器协议路由

建议把当前供应商抽象升级为“协议级 endpoint 模型”。

最少需要明确区分：

- Anthropic API endpoint
- OpenAI-compatible endpoint
- CLI 专用 endpoint
- 默认协议
- 支持协议集合

这样才能保证：

- `chat`、`review`、`accept`、`autonomy`
- `aider`、`claude_code`、`codex_cli`、`gemini_cli`

在同一供应商下都能稳定工作，而不会互相污染。

### 6.4 方向四：拆控制面，降低耦合

建议优先拆分以下模块：

1. `system-check`
2. `review`
3. `execution`
4. `accept`
5. `autonomy`
6. `timeline`

目标不是追求“代码更漂亮”，而是为了：

- 降低维护成本
- 降低回归风险
- 提高新增能力的可控性
- 让测试边界更清晰

### 6.5 方向五：建设可观测、可回放、可评测系统

建议补齐统一轨迹：

```text
对话
-> 计划
-> 任务
-> 执行器
-> worktree / patch / PR
-> CI
-> 验收
-> 人工接管
-> 完成态
```

至少要支持以下问题的追踪：

- 这次失败是谁导致的
- 哪一步触发了返工
- 哪个执行器最稳定
- 哪类任务最适合自动回写
- 哪类任务最容易触发人工放行

### 6.6 方向六：最后再做多代理控制台

EasyMVP 已经有调度器、依赖检查、资源锁、worktree 隔离这些基础。

因此，多代理控制台是“顺势升级”，但前提是先把交付闭环和治理面做稳。

换句话说：

- 现在不是没有多代理基础。
- 现在是没必要在闭环没打透时，过早把复杂度推进到多代理产品化。

## 7. 90 天路线图

### 7.1 第 1 阶段：0-30 天

目标：先做稳主链。

重点工作：

1. 拆分超大控制器与核心控制面
2. 统一 provider 协议路由
3. 明确 worktree 回写策略与运行行为
4. 补 execute / accept / rework / autonomy 的关键集成测试
5. 建立主链回归样例

阶段交付：

- `workflow` 控制面模块拆分完成
- 协议级 endpoint 方案落地
- 回写策略文档和配置开关落地
- 主链回归集可重复执行

### 7.2 第 2 阶段：31-60 天

目标：打通交付闭环。

重点工作：

1. 输出 `diff / patch / PR` 三种结果形态
2. 接入 CI 结果到验收阶段
3. 把验收从“评分中心”升级成“证据中心”
4. 支持 review comment 驱动的自动修复
5. 将高风险任务默认切换到 PR / 人审模式

阶段交付：

- 低风险任务支持自动 patch 回写
- 中高风险任务支持生成 PR
- 验收页可展示测试、构建、证据、问题列表
- review comment 可回流为修复任务

### 7.3 第 3 阶段：61-90 天

目标：做成真正的控制台。

重点工作：

1. 增加任务轨迹与 Agent Trace
2. 增加决策动作、人工接管、审批记录统一视图
3. 增加批量观察、批量调度、冲突可视化
4. 增加运行风险闸门与审计日志
5. 建立历史项目回放和失败案例回放

阶段交付：

- 项目级轨迹面板
- 任务级回放能力
- 决策与人审统一日志
- 风险策略与权限开关
- 内部交付评测面板

## 8. 研发任务拆分建议

| 优先级 | 方向 | 任务 | 建议周期 | 验收标准 |
|--------|------|------|----------|----------|
| P0 | 交付闭环 | 增加 `diff / patch / PR` 三种结果形态 | 1-2 周 | 每个任务完成后可输出明确结果类型 |
| P0 | 验收升级 | 接入 CI 结果到验收阶段 | 1 周 | `accept` 可展示测试、构建、失败原因 |
| P0 | 回写策略 | 明确 auto patch / PR / manual 三档策略 | 1 周 | 文档、配置、运行行为一致 |
| P0 | 协议抽象 | 重构 provider 为协议级 endpoint 路由 | 1-2 周 | CLI / API 不再互相污染 endpoint |
| P1 | 控制面重构 | 拆分超大 `workflow` 控制器 | 1-2 周 | review / accept / autonomy / system-check 模块独立 |
| P1 | 测试补齐 | execute / accept / rework / autonomy 集成测试 | 2 周 | 主链、返工、人工接管可重复回归 |
| P1 | 可观测性 | 增加任务轨迹、人工介入、决策记录 | 1-2 周 | 单任务可完整回放 |
| P1 | 安全治理 | 增加命令模板白名单、权限策略、审计日志 | 2 周 | 高风险执行具备可配置保护 |
| P2 | 多代理控制台 | 增加批量任务观察与冲突可视化 | 2-3 周 | 多任务执行可统一观察 |
| P2 | 评测体系 | 历史项目回放、失败样本回放、指标面板 | 2 周 | 每次发布可对比真实交付质量 |

## 9. 建议关注的核心指标

建议建立以下五组指标。

### 9.1 交付效率指标

- 主链完成率
- 平均交付时长
- 平均返工轮次
- 从创建项目到第一次可运行结果的时间

### 9.2 稳定性指标

- 任务失败率
- 自动重试恢复成功率
- 服务重启后恢复成功率
- CI 失败后自动修复成功率

### 9.3 自动化指标

- 自动验收通过率
- 自动 patch 回写比例
- PR 自动生成比例
- 人工接管率

### 9.4 治理指标

- 任务轨迹完整率
- 审计可追溯率
- 问题定位平均耗时
- 风险任务拦截率

### 9.5 业务价值指标

- 单项目平均人工节省时长
- 单项目平均需求到交付周期缩短比例
- 可复用项目模板数
- 客户/内部团队复用率

## 10. 资源建议

如果目标是 90 天内做出可对外展示、可内部试点的版本，建议最小配置为：

- 后端 2-3 人
- 前端 1 人
- 测试 / DevOps 1 人
- 产品 / 项目协调 1 人

资源投入优先顺序建议如下：

1. 后端核心链路与回写策略
2. 验收闭环与 CI 集成
3. 测试与可观测
4. 前端控制台与轨迹面板
5. 多代理体验优化

## 11. 当前不建议优先投入的方向

以下方向不是不做，而是不建议放在未来 90 天的最前面。

1. 再接更多模型或更多 CLI 工具
2. 追求高并发而忽略闭环稳定性
3. 过早承诺“全自动交付”
4. 先做复杂自治，再补治理和审计
5. 继续扩页面，但不补结果形态和证据链

## 12. 结论

EasyMVP 已经不是一个早期原型，而是一套已经跑通主链的 AI 软件交付编排系统。

它当前最重要的任务，不是继续证明“Agent 可以写代码”，而是证明：

```text
Agent 产出的结果
可以被稳定地验收、回写、审查、追踪，并最终安全交付。
```

因此，未来阶段最优的产品策略是：

```text
从 AI 工作流平台
升级为 AI 软件交付控制台
```

这是 EasyMVP 与普通 AI 编码助手拉开差异的关键。

## 13. 外部参考

以下资料用于辅助判断行业方向，均截至 2026-04-09 可访问。

- OpenAI Codex: https://openai.com/codex/
- OpenAI Introducing Codex App: https://openai.com/index/introducing-the-codex-app/
- OpenAI on SWE-bench Verified: https://openai.com/index/why-we-no-longer-evaluate-swe-bench-verified/
- Anthropic Claude Code: https://www.anthropic.com/product/claude-code
- Anthropic Claude Code Sandboxing: https://www.anthropic.com/engineering/claude-code-sandboxing
- GitHub Copilot coding agent: https://docs.github.com/en/copilot/concepts/agents/coding-agent/about-coding-agent
- Cognition Devin manage agents: https://cognition.ai/blog/devin-can-now-manage-devins
- Cognition Devin schedule agents: https://cognition.ai/blog/devin-can-now-schedule-devins
- Cognition Devin autofix review comments: https://cognition.ai/blog/closing-the-agent-loop-devin-autofixes-review-comments
- Cognition Agent Trace: https://cognition.ai/blog/agent-trace
- OpenHands runtime architecture: https://docs.openhands.dev/openhands/usage/architecture/runtime
- Factory GA: https://factory.ai/news/factory-is-ga
- Augment IDE agents: https://www.augmentcode.com/product/ide-agents/
- JetBrains Junie: https://www.jetbrains.com/help/ai-assistant/junie-agent.html
