# Everything Claude Code 提炼与采纳建议

基于仓库 `affaan-m/everything-claude-code` 当前主分支快照整理。

- 仓库地址: <https://github.com/affaan-m/everything-claude-code>
- 本次读取快照提交: `1a50145d39c0fa415311da62e7a018edd4e6d976`
- 本次盘点结果:
  - `48` 个 agents
  - `183` 个 skills
  - `79` 个 commands
  - `89` 个 rules
  - `hooks/` 运行面已内置

这份文档不是全量搬运 ECC，而是针对 `EasyMVP` 当前架构和你的明确要求，提炼出**值得吸收**、**应改造后吸收**、**不建议引入**三类内容。

---

## 一、先说结论

ECC 真正有价值的不是那一堆数量很多的 skill，而是它背后的 6 个工程原则：

1. **技能是主工作流，命令只是入口兼容层**
2. **把固定约束前置到 rules / hooks，不要靠模型记忆**
3. **复杂任务先 plan，再执行，再 review，再 verify**
4. **并行 agent 要有清晰分工、工作区隔离、质量闸门**
5. **上下文是预算，不是免费的**
6. **自动化闭环必须落在可验证产物上，而不是“模型说完成了”**

对 `EasyMVP` 来说，ECC 不该整仓照搬，应该只吸收下面三层：

- **工作流层**: planner / architect / code-reviewer / security-reviewer / code-explorer 这类核心 agent 思路
- **约束层**: development-workflow / code-review / hooks / performance 这类规则
- **治理层**: context-budget / repo-scan / verification-loop / search-first 这类治理型 skill

---

## 二、ECC 最值得我们学的东西

## 1. 技能优先，不靠 slash command 堆功能

ECC 在 `README.md` 和 `the-shortform-guide.md` 里反复强调：

- `skills/` 才是稳定的能力单元
- `commands/` 只是兼容入口
- 真正要沉淀的是“什么时候触发、怎么做、检查什么、输出什么”

这点很适合 `EasyMVP`。

因为 `EasyMVP` 本身就在做多角色 AI 团队协作，如果继续把能力散落在 prompt、角色预设、命令模板、人工记忆里，后面一定会失控。更合理的做法是：

- 角色负责“身份和职责”
- skill 负责“具体工作流”
- rule 负责“硬约束”
- hook 负责“自动拦截和自动补充”

## 2. 先研究，再规划，再编码

ECC 的 `rules/common/development-workflow.md` 很值得吸收。它把流程写得很硬：

1. 先搜索现成方案
2. 再查官方文档
3. 再做 plan
4. 再进入实现
5. 再 review 和 verify

这个顺序对 `EasyMVP` 很关键，因为你这个项目不是普通 CRUD，而是：

- 工作流编排
- 多执行器调度
- Git worktree 隔离
- 计划版本
- 验收证据
- 风险闸门

这种系统如果先写后想，返工成本会很高。

## 3. 把“必须遵守”的东西写进规则和钩子

ECC 的核心经验不是“提示词写长一点”，而是：

- rules 处理长期稳定约束
- hooks 处理执行时强约束

这点和你当前要求高度一致，尤其是你已经明确提出：

- 禁止 `go test`
- 使用 GitHub Actions 验证
- 按优先级推进
- 要闭环，不要做皮毛

这些都不应该继续靠我“记住”，而应该落成项目铁律。

## 4. 并行不是多开几个 agent，而是有边界地并行

ECC 的 `rules/common/agents.md`、`loop-operator.md`、`team-builder` 一类内容强调：

- 并行任务必须独立
- 必须明确角色
- 必须有工作区隔离
- 必须有质量门和回滚路径

这跟 `EasyMVP` 当前的 `workspace / worktree / delivery_policy / risk delivery` 机制是天然契合的。

也就是说，ECC 的多 agent 思想不是外来体系，反而可以加强你现有设计。

## 5. 上下文预算必须治理

ECC 的 `context-budget` 非常有用，因为它点出了一个经常被忽略的问题：

- agent 描述太长
- skills 太多
- MCP 太多
- CLAUDE.md 太长

都会直接拉低执行质量。

`EasyMVP` 现在正朝“多角色 + 多分类 + 多等级 + 多执行器”扩展，如果不控制上下文体积，后面 Architect、Auditor、Coordinator 的提示词会持续膨胀。

---

## 三、结合 EasyMVP，建议吸收的 Agents

下面不是“ECC 最强 agents 列表”，而是**对 EasyMVP 最有用**的最小集合。

## A. 第一优先级，建议直接吸收思路

### 1. `planner`

价值：

- 把需求转成分阶段计划
- 明确文件路径、依赖、风险、测试策略
- 避免 agent 直接进入“边写边猜”

适配 `EasyMVP` 的方式：

- 不是直接照抄 ECC planner 文案
- 而是把它改造成 `mvp-plan-builder` 或 `workflow-planner`
- 输出直接对齐 `PlanVersionService`、`task blueprint`、`batch_no`、`depends_on`、`affected_resources`

结论：

- **必须要**

### 2. `architect`

价值：

- 对复杂特性先做架构评审
- 强制写清 trade-off、边界、扩展性、数据流

适配 `EasyMVP` 的方式：

- 聚焦在 `workflow/`、`engine/`、`workspace/`、`acceptance`、`autonomy` 这类核心域
- 输出应该能进入你的设计文档体系，而不只是聊天结果

结论：

- **必须要**

### 3. `code-explorer`

价值：

- 在修改前先沿调用链摸清现有实现
- 非常适合老代码、深层路径和跨层功能

这对 `EasyMVP` 很重要，因为这个仓库有：

- 后端 MonoRepo
- 前端管理端
- workflow V2
- engine legacy / bridge
- workspace 独立域

很多功能不是看一个文件就能理解。

结论：

- **必须要**

### 4. `code-reviewer`

价值：

- 修改后立即 review
- 先看 diff，再看上下文，再给严重级别结论

适配 `EasyMVP` 的方式：

- review 模板应对齐你的项目约束
- 特别强化：工作流状态机、资源锁、依赖门控、幂等性、恢复逻辑、SSE 推送、权限边界

结论：

- **必须要**

### 5. `security-reviewer`

价值：

- 专查输入、认证、数据库、文件系统、外部调用、密钥暴露

这对 `EasyMVP` 是高优先级，因为你项目里有：

- AI provider / API key
- executor command_template
- workspace / git worktree
- callback_url / callback_secret
- JWT

这些都是高风险点。

结论：

- **必须要**

## B. 第二优先级，建议按 EasyMVP 场景裁剪后引入

### 6. `loop-operator`

价值：

- 面向自治循环
- 检查是否停滞、是否重试风暴、是否越花越多、是否无进展

这对 `EasyMVP` 的未来非常适合，尤其是：

- `WorkflowRun` 自动推进
- `rework` 循环
- 七层自治模块
- 风险闸门和人工 checkpoint

但现在不要整套照搬，应聚焦到：

- workflow 卡死
- batch 无进展
- accept/rework 来回抖动
- 恢复后重复失败

结论：

- **建议二阶段引入**

---

## 四、结合 EasyMVP，建议吸收的 Skills

## A. 必吸收

### 1. `search-first`

原因：

- 你多次强调效率
- `EasyMVP` 里很多能力不该重复造轮子
- 比如 GitHub Actions 汇总、PR 检查、任务编排、代码审查模板、工作区管理辅助工具

适配建议：

- 固化成项目规则：先搜仓内已有实现，再搜 GitHub，再查官方文档，再决定是否新写
- 对 Go/Vue/Actions 三条线都适用

### 2. `context-budget`

原因：

- 你项目的 `CLAUDE.md` 已经很长
- 未来还会加角色预设、技能、团队策略
- 如果不治理，AI 执行质量会越来越差

适配建议：

- 定期审计：
  - `CLAUDE.md`
  - 角色预设文本
  - 任务模板
  - 验收规则
  - 项目分类 prompt

### 3. `verification-loop`

原因：

- 你反复要求“闭环”
- `EasyMVP` 本身就有 acceptance / evidence / issue / rule

ECC 的验证闭环思想可以直接转成 EasyMVP 的产品能力：

- 执行完成不等于完成
- 必须有验证结果
- 验证失败进入 rework
- 验证证据要落库

适配重点：

- 和 GitHub Actions 结果对接
- 和 `.easymvp/ci/latest.json` 对接
- 和 `mvp_accept_evidence / mvp_accept_issue / mvp_accept_rule` 对接

### 4. `codebase-onboarding`

原因：

- 你现在就在不断让 AI 进入陌生目录、文档和模块
- 需要一个稳定的“先盘点、再进入”的标准流程

适配建议：

- 用它来生成每个应用或子域的入口导读
- 比如：
  - `admin-go/app/mvp/`
  - `admin-go/codegen/`
  - `vue-vben-admin/apps/web-antd/`

## B. 有价值，但建议按需使用

### 5. `repo-scan`

适合：

- 做大盘点
- 识别第三方沉积代码
- 清理垃圾目录
- 接手复杂仓库

对 `EasyMVP` 现阶段不是最急，但在你后面做“整理遗留代码、识别重复模块、瘦身仓库”时很有用。

### 6. `documentation-lookup`

适合：

- GoFrame
- Vue / Vben
- pnpm / vite / GitHub Actions
- OpenAI / Anthropic / Gemini / Claude Code / Codex CLI

它的核心价值不是“多一个 skill”，而是建立规则：

- 遇到框架/API/CLI 细节，优先用官方最新文档，不靠旧记忆。

### 7. `deep-research`

适合：

- 引擎选型
- 第三方集成调研
- 竞品/工作流设计研究

但它属于重型 research，不适合作为默认技能。

---

## 五、结合 EasyMVP，建议吸收的 Rules

ECC 的规则层是最值得借的，因为这部分最容易固化成你项目的铁律。

## 1. `development-workflow`

建议改造成 `EasyMVP 开发铁律`，重点保留：

1. 先搜现有实现
2. 先看文档
3. 先 plan
4. 再实现
5. 改完必须 review
6. 必须有 verification

并加入你自己的强制项：

1. **禁止本地 `go test` 作为最终验证口径**
2. **Go 与前端统一以 GitHub Actions 为最终验证结果**
3. **没有证据，不算完成**
4. **没有影响阅读或行为，不做洁癖式修改**
5. **并行执行必须遵守资源隔离与优先级**

## 2. `code-review`

ECC 这里提供了一个非常清楚的 review 阈值体系。建议在 EasyMVP 里直接改造成审计规则：

- blocker: 安全/数据一致性/状态机破坏
- error: 行为错误/恢复逻辑错误/幂等性错误
- warn: 结构风险/性能隐患/测试缺口
- info: 风格建议

它可以直接映射你现有的：

- `mvp_review_issue.severity`
- `mvp_accept_issue.severity`

## 3. `agents`

ECC 规定复杂任务应自动启用 planner / code-review / architect 等 agent。

对 `EasyMVP` 更进一步的做法是：

- 不是“建议用”
- 而是在 WorkflowRun 中把这些 agent 角色制度化

也就是：

- 设计阶段默认 architect + planner
- 执行阶段默认 implementer
- 评审阶段默认 auditor + code-reviewer
- 安全敏感改动自动附加 security-reviewer

## 4. `performance`

这部分最值得吸收的是两个点：

- 上下文窗口是预算
- 重任务和轻任务的模型路由要分开

这和你 `ai_engine_config` / `project_role` / `execution_mode` 体系天然契合。

未来可以让 `EasyMVP` 支持：

- architect 用 max
- auditor 用 pro
- implementer 用 lite/pro/max
- coordinator 用 pro

现在你已有角色等级，只差把路由策略再做硬一点。

---

## 六、结合 EasyMVP，建议吸收的 Hooks 思想

ECC 的 hook 设计不是“炫技”，而是把容易忘的事自动化。对 `EasyMVP` 真正有价值的是下面几个思路。

## 1. PreToolUse 拦截危险动作

建议在你的执行器或本地协作环境里加硬约束：

- 拦截直接 `git push`
- 拦截危险命令
- 拦截未经隔离的长任务
- 拦截违反项目铁律的验证方式

最关键的一条是：

- **拦截直接把 `go test` 当最终验收结论**

可以允许开发期局部自测，但不能把它记作“验证通过”。

## 2. PostToolUse 自动触发轻量质量门

建议在文件修改后自动触发：

- 变更摘要生成
- 受影响模块识别
- 是否需要补文档
- 是否需要更新验收规则
- 是否需要触发 GitHub Actions 验证

## 3. Stop Hook 做收尾闭环

这和你的系统非常匹配。一次工作结束后，自动检查：

- 是否生成了交付摘要
- 是否记录了验证状态
- 是否产出了证据引用
- 是否还有未解决 blocker

---

## 七、针对 EasyMVP 的落地建议

这里给出最现实的采纳方案，不走大而全。

## 第一批，建议尽快落地

### 1. 新增项目级 skill / agent 最小集合

建议只做这 8 个：

- `easymvp-workflow-planner`
- `easymvp-architect`
- `easymvp-code-explorer`
- `easymvp-code-reviewer`
- `easymvp-security-review`
- `easymvp-verification-loop`
- `easymvp-context-budget`
- `easymvp-search-first`

### 2. 把项目铁律写入 `CLAUDE.md`

至少新增这些硬规则：

- GitHub Actions 是最终验证口径
- 不以本地 `go test` 作为验收结论
- 修改完成必须给出验证状态与证据
- 优先修复影响行为、流程、验证、阅读的错误
- 并行工作必须资源隔离

### 3. 把验证结果接入 acceptance 体系

你现在已有：

- `.github/workflows/backend-guard.yml`
- `.github/workflows/web-antd-guard.yml`
- `.easymvp/ci/latest.json`

这已经说明你的项目不是缺“验证机制”，而是缺统一消费规则。

应补的不是再造验证，而是：

- 读取 GitHub Actions 结果
- 转成 acceptance evidence
- 转成 accept issue
- 失败后自动回流 rework

## 第二批，建议在工作流引擎里产品化

### 4. 把 ECC 的 agent 编排思想写进 WorkflowRun

建议做成明确策略：

- `designing`: architect + planner
- `reviewing`: auditor + code-review/security-review
- `executing`: implementer
- `accepting`: verification-loop
- `reworking`: implementer + targeted reviewer

### 5. 增加卡死与回环监控

借鉴 `loop-operator`，监控：

- 同批次无进展
- 同错误重复出现
- rework 次数过多
- acceptance 和 rework 来回震荡

---

## 八、不建议直接引入的部分

ECC 仓库很大，但很多内容对 `EasyMVP` 当前阶段没有价值，甚至会稀释重点。

## 1. 过多行业型/运营型 skills

例如：

- investor
- billing ops
- seo
- outreach
- social graph
- procurement

这些和你当前主线关系不大，不建议引入。

## 2. commands 全量照搬

ECC 自己都已经在弱化 command，强调 skill 才是主面。对 `EasyMVP` 来说，不应该再引入一堆 slash command 名字，而应该直接沉淀到系统工作流。

## 3. 过重的 MCP 依赖

ECC 很强调 MCP，但也明确提醒工具过多会吃掉上下文预算。

对 `EasyMVP` 来说，应坚持：

- 少量高价值 MCP
- 其余尽量走已有 shell / web / GitHub API / 官方文档

---

## 九、对 EasyMVP 最终建议

如果只保留一句话：

**不要把 Everything Claude Code 当成“技能商店”，而要把它当成“AI 工程团队操作系统”的样板。**

对 `EasyMVP` 真正该拿走的是：

- 工作流分层
- agent 分工
- rule 硬约束
- hook 自动补位
- verification 闭环
- context 治理

而不是把 183 个 skill 搬进来。

---

## 十、建议下一步

最合理的下一步不是继续看更多 ECC 文件，而是直接在 `EasyMVP` 落地第一批最小能力：

1. 把 GitHub Actions 验证优先写进 `CLAUDE.md`
2. 产出 `easymvp-search-first`
3. 产出 `easymvp-verification-loop`
4. 产出 `easymvp-code-reviewer`
5. 产出 `easymvp-workflow-planner`

这样你得到的是可执行资产，不是阅读笔记。
