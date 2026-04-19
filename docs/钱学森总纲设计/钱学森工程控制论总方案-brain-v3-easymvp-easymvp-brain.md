# 钱学森工程控制论总方案：brain-v3、EasyMVP、easymvp-brain 的定位与建设路线

> 更新时间：2026-04-20  
> 适用范围：`/www/wwwroot/project/brain-v3`、`/www/wwwroot/project/easymvp`  
> 方法论总纲：钱学森《工程控制论》  
> 参考对象：`brain-v3` 最新核心文档、EasyMVP V3 现有架构文档、Everything Claude Code 的工作流经验

---

## 1. 文档目的

这份文档只回答 4 个问题：

1. 为什么要用《工程控制论》做总纲
2. `brain-v3`、`EasyMVP`、`easymvp-brain` 三者分别是什么
3. 三者之间的边界和协作关系是什么
4. 接下来分别要做什么，优先级如何排

这份文档不讨论单个 API 细节，也不直接替代已有专题文档。它是顶层总方案，用来统一后续设计与执行方向。

---

## 2. 设计结论

先给结论，不绕。

### 2.1 三者不是同一层，不应该互相替代

- `brain-v3` 是 **多脑运行时与控制面**
- `EasyMVP` 是 **业务工作流编排与交付平台**
- `easymvp-brain` 是 **EasyMVP 领域专精大脑**

三者关系不是“谁替代谁”，而是：

```text
钱学森《工程控制论》 = 总纲与控制原则
            ↓
brain-v3 = 通用多脑运行时 / 控制面 / 脑调度内核
            ↓
easymvp-brain = EasyMVP 领域认知与裁决大脑
            ↓
EasyMVP = 业务编排、阶段状态机、证据与验收平台
```

### 2.2 不应该重复造 brain-v3 已经有的脑

`brain-v3` 当前最新核心不是“三脑系统使用指南”那套旧表述，而是：

- **Central Brain**
- 多个 **Specialist Brains**
- Brain Manifest / Capability / Runtime / Policy / Control Plane

对 EasyMVP 真正相关、应该优先利用的基础专精大脑，不是所有脑，而是这 4 个：

1. **代码大脑**（`code`）：代码读写、执行、测试
2. **浏览器大脑**（`browser`）：浏览器操作、提取、截图、页面采证
3. **审核大脑**（当前实现名 `verifier`）：验证、断言、检查、结果审核
4. **故障大脑**（`fault`）：故障诊断、恢复

中央大脑负责协调、审查、委派、仲裁，不直接承担领域业务状态机。

### 2.3 EasyMVP 不该直接把 brain-v3 当“又一个 execution_mode”

错误接法：

- 把 `brain-v3` 塞进 `execution_mode`
- 像跑 `aider`、`codex_cli` 一样把 `brain run` 当 shell 命令执行

正确接法：

- EasyMVP 仍然掌握业务工作流与状态机
- 通过 `brain serve` / Run API 对接 `brain-v3`
- 用 `easymvp-brain` 承担领域判断与合同编译
- 用 `brain-v3` 的中央大脑和 4 个专精大脑完成运行时协作

### 2.4 《工程控制论》要落在结构上，不是当口号

钱学森在这里不是“阅读材料”，而应该变成系统设计总纲，具体体现为：

- 反馈闭环
- 分层控制
- 解耦与最小互扰
- 时滞处理
- 噪声过滤
- 稳定优先于盲目放大
- 约束下求最优
- 自稳定与适应
- 误差显式管理

这些原则必须映射到架构、状态机、验证、风险闸门和脑协作策略中。

---

## 3. 为什么用钱学森《工程控制论》做总纲

EasyMVP 现在面临的问题，本质上不是“少几个 agent”或“少几个 prompt”，而是**系统控制问题**。

典型问题包括：

- 任务执行成功不等于交付成功
- 多执行器、多工作区、多阶段之间互相影响
- 审核、返工、验收不是统一闭环
- 远端验证环境、GitHub Actions、人工审批、浏览器采证都存在时滞
- 模型输出、日志、审计结论存在噪声
- 自动推进过强时会导致不稳定、返工震荡、错误扩散

这些问题都非常适合用《工程控制论》的视角重新整理。

### 3.1 映射关系

| 工程控制论概念 | 在本系统中的映射 |
|---|---|
| 反馈伺服系统 | verification loop / acceptance / rework |
| 不互相影响的控制 | worktree 隔离 / resource lock / 批次门控 |
| 采样伺服系统 | 远端验证环境或 GitHub Actions 的异步反馈 / 阶段检查点 / 轮询同步 |
| 有时滞的线性系统 | CI 延迟 / 人工审批延迟 / 远端 brain run 状态滞后 |
| 平稳随机输入 | 模型波动 / 审计噪声 / 日志不确定性 |
| 非线性系统 | 返工振荡 / 多任务耦合失控 / 异常连锁反应 |
| 满足积分条件的控制设计 | 按验收指标和风险约束设计流程，而非拍脑袋编排 |
| 自动寻优 | 模型路由 / 脑选择 / 资源调度优化 |
| 噪声过滤 | 证据归一化 / 日志提炼 / review issue 分级 |
| 自行镇定与适应环境 | 自动降级 / 恢复 / fallback / 人工接管 |
| 误差控制 | issue / evidence / verdict / 风险闸门 / 审计追踪 |

### 3.2 总纲原则

后续所有设计都应服从这 8 条原则：

1. **闭环优先**
   - 没有验证和纠偏的执行，不算完成。
2. **稳定优先**
   - 不为了速度和并发牺牲流程稳定性与可恢复性。
3. **分层控制**
   - 业务编排、脑调度、工具执行、验收裁决必须分层。
4. **最小互扰**
   - 任务、资源、工作区、脑实例之间尽量解耦。
5. **时滞显式处理**
   - 不假设“结果立即可知”，必须有轮询、同步、超时和降级策略。
6. **噪声显式过滤**
   - 不把原始日志、模型话术直接当最终结论。
7. **约束下求最优**
   - 在上下文、成本、风险、人工预算约束下求最优推进。
8. **误差可记录、可回流、可审计**
   - 错误不是被掩盖，而是进入证据链和返工链。

---

## 4. 三者定位

## 4.1 `brain-v3` 的定位

`brain-v3` 的定位应明确为：

> **多脑运行时 + 脑级控制面 + 通用执行内核**

根据 `sdk/docs/32-v3-Brain架构.md`、`sdk/docs/35-BrainCapability标签与匹配算法.md`、`sdk/docs/35-跨脑通信协议设计.md`，它的核心职责是：

- Brain-first 的顶层对象管理
- Central Brain 协调和委派
- Specialist Brain 的运行、健康、授权和观察
- Capability 匹配和脑选择
- Runtime 管理：native / mcp-backed / hybrid / remote
- Run 生命周期
- Policy / Tool Scope / File Policy / Workdir Confine
- 日志、事件流、Replay、Dashboard、审计

### 4.1.1 对 EasyMVP 最有价值的 brain-v3 组成

对 EasyMVP 来说，当前最核心的不是 `data` / `quant`，而是：

- `central`
- **代码大脑** `code`
- **浏览器大脑** `browser`
- **审核大脑** `verifier`
- **故障大脑** `fault`

其中：

- `central` 负责脑间协调、任务理解、delegate、review 和仲裁
- **代码大脑** 负责代码操作与局部执行
- **浏览器大脑** 负责浏览器级行为验证和页面证据采集
- **审核大脑** 负责只读核验、断言、检查和结果审查
- **故障大脑** 负责异常分析、恢复建议、故障路径判断

### 4.1.2 brain-v3 不负责什么

`brain-v3` 不应该负责：

- EasyMVP 的项目、计划、阶段、任务、验收业务对象
- EasyMVP 的 review / execute / rework / accept 业务状态机
- EasyMVP 的合同生成、证据裁决、业务规则归档

一句话：

> `brain-v3` 管的是“脑如何运行和协作”，不是“EasyMVP 业务如何流转和裁决”。

---

## 4.2 `EasyMVP` 的定位

`EasyMVP` 的定位应明确为：

> **业务工作流编排平台 + 交付控制平台**

它的核心职责是：

- 项目生命周期管理
- WorkflowRun 六阶段状态机
- PlanVersion / Blueprint / DomainTask 业务对象管理
- 批次门控、依赖、资源锁、工作区隔离
- 人工审批、交互、Dashboard、Action Inbox
- Acceptance / Evidence / Issue / Release Gate
- 与高配验证环境、GitHub Actions、执行器、brain-v3 的接入和归一化

### 4.2.1 EasyMVP 的核心不是“会执行”，而是“会控流程”

EasyMVP 必须负责：

- 哪个阶段允许什么动作
- 哪个任务该由哪个脑或执行器承担
- 什么算完成、什么算失败、什么必须返工
- 什么证据才足以支持 acceptance

一句话：

> EasyMVP 是业务控制面，不是执行器本身。

### 4.2.2 EasyMVP 的验证环境分层

EasyMVP 的验证体系不应被误解成“只认 GitHub Actions”。正确口径是三层：

1. **低配开发机**
   - 负责日常开发、编排、轻量检查、触发远端验证
   - 不承担重编译和最终验收
2. **高配验证环境**
   - 负责最终编译、重测试、重构建、重打包
   - 这是长期的正确验证终点
3. **GitHub Actions**
   - 当前阶段因为本地服务器配置不足，承担远端重验证角色
   - 属于阶段性替代方案，不是永久总纲

结论：

> **最终验证口径是“满足资源条件的高配验证环境”，当前临时实现是 GitHub Actions。**

---

## 4.3 `easymvp-brain` 的定位

`easymvp-brain` 的定位应明确为：

> **EasyMVP 领域专精大脑**

它不是通用代码脑，不是浏览器脑，也不是故障脑的重复品。它专门承担 **EasyMVP 业务语义的认知与裁决**。

### 4.3.1 easymvp-brain 负责什么

它第一版应该只负责 5 类高价值领域任务：

1. **方案审核**
   - 审核 PlanDraft / Blueprint 是否合理
2. **方案编译**
   - 把草案编译成正式可执行的任务合同
3. **返工重构**
   - 把失败任务转成结构化 RepairPlan，而不是临时补丁
4. **验收规则映射**
   - 根据项目类型和产物声明所需证据和验证方式
5. **完成语义裁决**
   - 区分 `run_succeeded`、`delivery_verified`、`accepted`

### 4.3.2 easymvp-brain 不负责什么

它不应该直接承担：

- 文件编辑
- 浏览器操作
- 最终代码实现
- 页面自动化
- 通用故障诊断

这些应该继续交给 `brain-v3` 的专精脑或其他执行器。

一句话：

> `easymvp-brain` 负责“懂 EasyMVP 的领域语义”，不负责“替别人干活”。

---

## 5. 推荐总体分层

建议采用四层分层。

```text
第 0 层：钱学森《工程控制论》总纲
  - 闭环 / 稳定 / 时滞 / 噪声 / 误差 / 自适应

第 1 层：EasyMVP 业务控制层
  - 项目 / 工作流 / 阶段状态机 / 验收 / 证据 / 风险闸门

第 2 层：easymvp-brain 领域认知层
  - 审核 / 编译 / 返工重构 / 验收规则映射 / 完成裁决

第 3 层：brain-v3 运行时层
  - central + code + browser + verifier + fault
  - capability / runtime / policy / replay / dashboard / run api
```

### 5.1 协作关系

```text
EasyMVP Workflow Orchestrator
    ├─ 调用 easymvp-brain 做领域判断
    ├─ 通过 brain-v3 central 发起脑协作运行
    ├─ 调用 code/browser/verifier/fault 执行具体能力
    ├─ 汇总 run status / logs / evidence
    └─ 按业务规则进入 accept / rework / complete
```

---

## 6. 具体职责划分

## 6.1 `brain-v3` 需要做什么

这里列的是**对 EasyMVP 接入有直接意义**的待做项，而不是 brain-v3 的全量 roadmap。

### P0

1. **稳定 `brain serve` Run API 契约**
   - 让 EasyMVP 能稳定创建、查询、取消、恢复 Run
2. **强化中央大脑 + 4 个基础专精大脑的 capability 路由**
   - central 能稳定选择代码大脑 / 浏览器大脑 / 审核大脑 / 故障大脑
3. **补齐多脑运行时状态可观测性**
   - status / logs / replay / event 必须可被 EasyMVP 消费
4. **workdir 与 file policy 进一步收紧**
   - 保障多租户、任务隔离、工作区收敛

### P1

5. **把 Brain Capability 匹配真正用于 EasyMVP 场景**
   - 例如验证类任务优先命中审核大脑，页面采证优先命中浏览器大脑
6. **强化故障大脑在恢复链中的可用性**
   - 让它可用于失败分类和恢复建议，而不只是单次故障分析
7. **统一 event / log 结构**
   - 便于 EasyMVP 归一化成 Evidence 与 LiveEvent

---

## 6.2 `EasyMVP` 需要做什么

EasyMVP 要做的不是“把脑跑起来”，而是把脑纳入业务闭环。

### P0

1. **建立 BrainRunBinding 体系**
   - 将 `project / domain_task / compiled_task / run_id / brain_kind` 绑定
2. **把 Run 生命周期映射到 DomainTask 生命周期**
   - 明确 `run_succeeded ≠ completed`
3. **统一 Acceptance / Evidence / Issue 的消费口径**
   - Run 输出、高配验证环境结果、GitHub Actions 结果、浏览器采证都要归一化
4. **把高配验证环境作为最终验证口径**
   - 当前阶段由 GitHub Actions 代理承担远端重验证角色
   - 不再把本地低配机上的 `go test` / `go build` 视为最终验收依据
5. **把 rework 从补丁模式提升为结构化回路**
   - 返工必须回到领域判断和合同层

### P1

6. **为不同项目分类配置默认脑协作策略**
   - 哪类项目默认优先代码大脑 / 浏览器大脑 / 审核大脑 / 故障大脑
7. **引入风险闸门与人机接管策略**
   - 高风险任务、中断、重复失败时自动升级人工介入
8. **工作台展示 active runs / live events / evidence timeline**
   - 把控制面能力产品化

---

## 6.3 `easymvp-brain` 需要做什么

它是整个方案里最需要补上的一层。

### P0

1. **方案审核器**
   - 输入：PlanDraft / Blueprint
   - 输出：ReviewResult（scope/risk/resources/verification 缺失项）
2. **方案编译器**
   - 输出：CompiledPlan + DeliveryContract + VerificationContract
3. **返工编译器**
   - 输入：失败上下文 + 原始合同
   - 输出：RepairPlanDraft
4. **完成裁决器**
   - 输入：run result + delivery result + verification result
   - 输出：真实完成语义

### P1

5. **验收规则生成器**
   - 针对项目分类生成 required evidence
6. **多脑任务分派建议器**
   - 给出 central 应优先委派哪个基础专精大脑的建议
7. **领域知识约束器**
   - 保证 EasyMVP 任务合同、阶段规则、风险术语一致

---

## 7. 为什么不是“再造一套审核脑和故障脑”

这点必须说清楚。

因为 `brain-v3` 已经有：

- `central`
- **审核大脑** `verifier`
- **故障大脑** `fault`

如果 EasyMVP 再造一套平行的审核脑、故障脑，会产生 4 个问题：

1. 职责重叠
2. 路由混乱
3. 审计口径不一致
4. 后期维护成本翻倍

所以正确路线是：

- **通用审核 / 验证 / 故障能力**：尽量复用 `brain-v3`
- **EasyMVP 领域语义判断**：沉到 `easymvp-brain`

也就是说：

> **审核大脑** 和 **故障大脑** 解决“怎么验证 / 怎么诊断”，  
> `easymvp-brain` 解决“为什么此时要验证 / 这次失败应进入哪条业务返工路径”。

---

## 8. 推荐的四脑协作方式

对 EasyMVP 当前阶段，推荐以中央大脑协调 4 个基础专精大脑作为核心运行组合：

```text
central
  ├─ code      代码大脑：代码实现 / 命令执行 / 文件操作
  ├─ browser   浏览器大脑：页面验证 / 流程采证 / 截图与提取
  ├─ verifier  审核大脑：只读核验 / 断言 / 验收检查
  └─ fault     故障大脑：故障分类 / 恢复建议 / 失败路径判断
```

### 8.1 典型链路

#### 链路 A：正常执行

1. EasyMVP 触发 `easymvp-brain` 进行方案审核与编译
2. EasyMVP 调用 `brain-v3 central`
3. central 委派 `code`
4. code 完成修改
5. central 或 EasyMVP 调审核大脑
6. 若需要页面侧证据，再调 `browser`
7. EasyMVP 汇总 Evidence，进入 acceptance

#### 链路 B：执行失败

1. code 运行失败或验证失败
2. EasyMVP 触发故障大脑
3. 故障大脑给出失败分类、恢复建议
4. EasyMVP 交给 `easymvp-brain` 生成 RepairPlanDraft
5. 回到 review / compile / execute

---

## 9. 分阶段建设路线

## Phase 1：立总纲、收边界

目标：

- 统一三者定位
- 收口职责重叠
- 明确中央大脑 + 4 个基础专精大脑是 EasyMVP 首期运行组合

要做的事：

1. 固化本方案文档
2. 在 EasyMVP 文档总纲中引用本方案
3. 把“高配验证环境为最终验证口径，GitHub Actions 为当前替代通道”写入项目铁律

## Phase 2：打通 EasyMVP ←→ brain-v3

目标：

- 让 Run 能被业务层绑定、观察、回放、取消、恢复

要做的事：

1. 建 BrainRunBinding 表与事件表
2. 实现 Run API 适配层
3. 将 active runs / live events 接入工作台

## Phase 3：做 easymvp-brain 第一版

目标：

- 让 EasyMVP 拥有自己的领域判断层

要做的事：

1. 方案审核器
2. 方案编译器
3. 返工编译器
4. 完成裁决器

## Phase 4：把 acceptance / rework 真正闭环

目标：

- 执行成功不再等于完成
- 失败进入结构化返工

要做的事：

1. 统一 verification contract
2. 统一 delivery contract
3. 统一 acceptance evidence 消费
4. 统一 rework 入口

## Phase 5：做优化与自适应

目标：

- 让系统具备稳定、降级、重路由和长期优化能力

要做的事：

1. 脑选择优化
2. 风险闸门自适应
3. 失败模式归纳
4. 证据与 issue 质量提升

---

## 10. 最终方案摘要

如果把整份文档压缩成三句话：

1. **钱学森《工程控制论》是总纲，负责定义闭环、稳定、时滞、噪声、误差与自适应原则。**
2. **brain-v3 是多脑运行时，中央大脑协调 code / browser / verifier / fault 四个核心专精脑，为 EasyMVP 提供通用能力与控制面。**
3. **EasyMVP 负责业务工作流与交付闭环，easymvp-brain 负责 EasyMVP 领域判断、方案编译、返工重构与完成裁决。**

再压缩一句：

> **brain-v3 管“脑怎么运行”，EasyMVP 管“业务怎么流转”，easymvp-brain 管“EasyMVP 该如何判断与裁决”。**

---

## 11. 下一步建议

按优先级建议直接继续做这 4 件事：

1. 补一份 `easymvp-brain` 的职责与输入输出契约文档
2. 补一份 EasyMVP ←→ brain-v3 的 Run 绑定数据模型文档
3. 补一份基于钱学森总纲的 `verification / rework / acceptance` 统一规则文档
4. 把这份总方案挂到 `EasyMVP-V3文档总纲.md` 或总入口文档里
