# Workflow V2 主链后端联调问题记录

> 记录日期：2026-04-08
>
> 范围：仅后端联调；不启前端；不使用 Docker；在当前服务器环境直接启动 `admin-go/app/mvp`。

## 1. 联调环境

- 服务目录：`/www/wwwroot/project/easymvp/admin-go`
- HTTP 端口：`9002`
- MySQL：`127.0.0.1:3306 / easymvp`
- Redis：`127.0.0.1:6379`
- 本机 Redis 密码：`123456`
- 服务器资源基线：
  - CPU：`2` 核
  - 内存：约 `3.6GB`
- 启动方式：

```bash
REDIS_PASS=123456 GOMAXPROCS=1 go run ./app/mvp/main.go
```

## 2. 联调样本

本轮后端主链联调使用新建项目：

- `projectID`: `317303721941798912`
- `conversationID`: `317303721996324864`
- `workflowRunID`: `317303722017296384`

目标链路：

```text
create-project -> chat/send -> parse-tasks -> confirm-plan -> review -> execute
```

## 3. 已确认问题

### 3.1 Redis 本机密码与默认配置不一致

现象：

- 本机 Redis 需要密码 `123456`
- `admin-go/app/mvp/manifest/config/config.yaml` 默认只配置了 `address`
- 若直接本机启动服务而不注入 `REDIS_PASS`，删除 worker 等 Redis 相关链路会处于“服务可起，但运行期隐性失败”的状态

影响：

- 不一定阻塞项目创建
- 会影响删除队列、后台异步链路稳定性

### 3.2 新建项目后无法直接 `confirm-plan`

现象：

- 新建项目后直接调用 `/api/mvp/workflow/confirm-plan`
- 返回：`没有待确认的方案版本`

原因：

- 新建项目只会创建 `workflow_run` 和架构师对话
- 不会自动生成 `plan_version`
- 必须先通过架构师输出任务清单，再调用 `/api/mvp/workflow/parse-tasks`

影响：

- 纯后端或人工联调时，如果不了解这条前置链路，会误判主链不可用

### 3.3 审核阶段 `coordinator_optimize` 非阻断失败

现象：

- `review-status` 显示审核整体通过
- 但 `coordinator_optimize` 子任务失败
- 错误信息：`解析协调员优化结果失败: 无法从 AI 回复中提取有效 JSON`

影响：

- 不阻断 `review -> execute`
- 但会污染阶段状态与日志，增加误报和后续排查成本

### 3.4 执行阶段首任务在 root 环境下失败

现象：

- 执行阶段首个领域任务进入 `failed`
- 返回错误：

```text
Claude Code 执行失败: exit status 1
--dangerously-skip-permissions cannot be used with root/sudo privileges for security reasons
```

原因：

- `claude_code` 默认命令硬编码了 `--dangerously-skip-permissions`
- 当前服务器进程以 root 身份运行

影响：

- `auto` 模式优先选择 `claude_code`
- 因此整个执行主链会稳定卡在第一批任务

### 3.5 腾讯云 Coding Anthropic Base URL 在 Claude Code 场景下不能带 `/v1`

通过最小 CLI 验证，结论如下：

- `ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic/v1`
  - `claude -p --model tc-code-latest`
  - 结果：模型不可用
- `ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic`
  - `claude -p --model tc-code-latest`
  - 结果：成功返回 `OK`
- 同样地，`kimi-k2.5` 在去掉 `/v1` 后也可正常返回

说明：

- 当前 `ai_provider.base_url` 和 `ai_engine_config.base_url` 都是 `.../coding/anthropic/v1`
- 这对后端聊天 Provider 不一定有问题
- 但对 `Claude Code CLI` 是错误接法

影响：

- 即使修复了 root 参数问题
- 仍会因为 Base URL 规范错误导致 Claude Code 调不通

### 3.6 `ai_provider.provider_type` 单值设计无法表达多协议供应商

现状：

- `ai_provider.provider_type` 是单个字符串
- 当前腾讯云 Coding 记录使用：`tencent_coding`

问题：

- 腾讯云 Coding 同时具备：
  - Anthropic 风格接口：适配 `Claude Code` / `Aider`
  - OpenAI 兼容接口：适配 `Codex` 等执行器
- 单值字段无法表达“一个供应商支持多个接口类型”

影响：

- 执行器协议选择只能依赖硬编码或 URL 猜测
- 数据层无法稳定表达真实能力边界

### 3.7 `supported_protocols` 只加在管理面还不够，运行时原先完全没读取

现象：

- 即使给 `ai_provider` 增加了多协议字段
- `chat`、`review`、`accept`、`autonomy`、`Aider`、`ai/task runtime` 等运行时链路如果仍只看 `provider_type`
- 那么多协议配置在执行时等于失效

影响：

- 表结构看似已支持多协议
- 但实际执行仍然是旧逻辑
- 会造成“后台配置能选，运行时不生效”的假象

### 3.8 Provider 导出/导入链路对多协议不完整

现象：

- `provider` CSV 导出原实现中：
  - 列头与数据行占位数量不一致
  - `supported_protocols` 会按 Go 切片默认格式输出，而不是可回导的协议列表

影响：

- 导出文件内容不稳定
- 再导入时会丢失或污染多协议配置

### 3.9 单个 `base_url` 仍无法完整表达“同一供应商多协议不同端点”

现象：

- 即使支持了 `supported_protocols`
- 仍然只有一个 `base_url`
- 像腾讯云 Coding 这类供应商，不同协议通常不是同一条 URL

影响：

- 当前方案足以先修复 `Claude Code` 主链阻断
- 但若后续要把同一供应商同时稳定供给 `Claude Code`、`Codex CLI`、`Aider`
- 仍需要补充“协议级 URL 配置”或统一的供应商路由规则

### 3.10 设计阶段架构师聊天在腾讯云 Coding Anthropic 流式调用下直接 404

复现时间：

- `2026-04-08 21:13:11 CST`

复现样本：

- `projectID`: `317315397890084864`
- `conversationID`: `317315398066245632`
- `workflowRunID`: `317315398087217152`
- `replyID`: `317315520799969280`

现象：

- 新建项目后调用 `/api/mvp/chat/send`
- 架构师模型 `Auto (tc-code-latest)` 直接失败
- 消息内容写入：

```text
AI 调用失败: anthropic stream error (status 404): 404 page not found
```

当前判断：

- 腾讯云 Coding 在 `Claude Code CLI` 场景下要求去掉 `/v1`
- 但后端 `Anthropic Messages API` 自身仍然需要走 `.../v1/messages`
- 当前共享的 `ResolveBaseURLForProtocol(..., anthropic)` 被同时用于：
  - `Claude Code CLI` 的 `ANTHROPIC_BASE_URL`
  - 后端 `AnthropicProvider` 的 HTTP 请求
- 结果导致后端把 API 地址错误拼成了 `.../anthropic/messages`
  - 正确方向应为 `.../anthropic/v1/messages`

影响：

- 主链在 `create-project -> chat/send` 就会中断
- 即使 `claude_code` 执行器已修复 root 参数问题
- 只要架构师/审核/解析等聊天链路走 Anthropic Provider，仍无法进入后续阶段

## 4. 当前建议修复顺序

### 第一优先级：保证执行主链跑通

1. `claude_code` 在 root 环境下移除 `--dangerously-skip-permissions`
2. `claude_code` 对腾讯云 Coding 的 Anthropic Base URL 自动去掉 `/v1`
3. 用当前联调项目重新执行 `execute` 阶段，确认首批任务能跑通

### 第二优先级：降低审核阶段噪音

1. 处理 `coordinator_optimize` 的 JSON 解析脆弱性
2. 至少保证它失败时不会制造误导性红灯

### 第三优先级：补齐数据层表达能力

1. 为 `ai_provider` 增加多协议能力字段
2. 保留 `provider_type` 作为默认协议，兼容老逻辑
3. 后端逐步改为优先读取多协议配置

### 当前修复中的补充项

1. 让 `supported_protocols` 从数据层真正贯穿到运行时
2. 修复 `provider` 导出/导入在多协议场景下的数据失真
3. 保留单 `base_url` 方案先保证主链可运行，协议级 URL 后续再补
4. 拆分“CLI 基地址”和“后端 API 基地址”的 Anthropic 路由解析，避免互相污染

## 5. 备注

- 当前主链不是完全不可用，而是已经确认存在“执行阶段首阻断 bug + 数据层表达不足”
- 本文档用于后续继续修复、回归和补充测试结果

## 6. 2026-04-08 晚间二次回归结论

本轮继续在纯后端环境下回归，新增样本：

- `projectID`: `317335338903146496`
- `conversationID`: `317335338911535104`
- `workflowRunID`: `317335338924118016`

### 6.1 已落地修复

1. 调度并发统一收敛为 `1`
   - `mvp_config` 中 `scheduler.max_concurrent` 与 `workflow.scheduler.max_concurrency` 已统一改为 `1`
   - 代码层新增兼容读取，避免新旧键并存时出现误读

2. root 环境下 `auto` 不再误选 `claude_code`
   - `auto` 执行器在 root 环境会跳过 `claude_code`
   - 当前服务器会自动回落到 `aider`

3. `claude_code` 权限拒绝不再误判成功
   - 之前 Claude CLI 仅输出“需要授权”时，任务会被错误记为 `completed`
   - 现已改为识别权限拒绝/假成功输出并正确置为 `failed`

4. `implementer/lite` 角色不再错误退化为 `chat`
   - 角色精确命中失败时，现会回退到同角色类型的项目级默认配置
   - 当前 `software_dev` 项目里的 `lite` 实例任务已正确生成为 `auto`

5. 任务资源锁支持同任务重复执行
   - 之前同一 `task_id + resource_path` 释放后再次执行会撞唯一索引
   - 现已改为可重复获取并复用同一锁记录

6. 任务工作空间支持按 `task_id` 原子复用
   - 之前 `mvp_task_workspace.uk_task` 导致人工重启/自动重试无法再次创建工作空间
   - 现已改为按唯一键复用同一记录，`deleted_at` 可正确回收并重新拉起

7. 服务重启后的执行器恢复链路已补齐
   - 之前进程重启后，`DomainTaskScheduler` 会恢复调度循环但丢失执行器绑定，出现“任务 running 但不真正执行”的假活
   - 现已统一在 `workflow recovery / resume / retry / force execute` 前重绑执行器；`execute` 阶段还会重绑完成回调

8. `mvp_decision_action.result` JSON 列写入已修正
   - 仓储层已增加 JSON 字段统一规范化，避免把裸字符串直接写进 JSON 列
   - 此问题首次暴露在 `retry_limit_reached -> notify_human` 场景

### 6.2 已实际验证通过的主链节点

已通过后端接口确认以下链路可实际推进：

```text
create-project
-> chat/send
-> parse-tasks
-> confirm-plan
-> review
-> execute
-> 人工 retry-task / update-domain-task 重启
-> 服务重启后继续执行
```

其中已拿到的关键事实：

1. `system-check` 已通过，调度并发实际为单并发
2. 首个任务 `317335697453223936` 在修复后已真实完成，不再卡在资源锁或工作空间唯一键
3. 日志已明确出现：
   - `Workspace 创建 worktree`
   - `AiderRunner 启动`
   - `AiderRunner 完成`
   - `Workspace Finalize success=true`
4. 首个任务完成后，第二个任务 `317335697461612544` 已自动进入 `running`

### 6.3 当前仍观察到的非主链阻断项

1. 架构师输出仍存在一定不稳定性
   - 偶发输出伪工具调用或解释文本，需要更强提示词才能稳定得到任务 JSON

2. `aider` 存在“生成过宽”倾向
   - 会主动补较多脚手架文件，超出 `affected_resources` 时系统会正确拦截
   - 这属于任务边界控制问题，不再是调度/恢复类阻断

3. 飞书配置仍持续报错
   - 当前日志中仍有 `app secret invalid` / token 获取失败
   - 不阻断 Workflow V2 主链，但会污染日志并影响协作通知

4. `TaskReset` 清理旧 worktree 时偶发出现：

```text
fatal: not a git repository (or any of the parent directories): .git
```

   - 现阶段不会阻断后续重建 worktree
   - 但说明历史残留目录的清理逻辑仍可继续收敛

### 6.4 本轮继续联调新增确认项

1. `domain-tasks` 接口已补兼容
   - `mvp_domain_task` 真实表中没有 `error_message` 列
   - 列表接口已改为不直接查询该列，失败任务的错误信息从 `result` 兜底返回

2. `aider` 会把自然语言标题/命令误当文件名
   - 实际产生过伪文件：`运行方式：`、`验证方式：`
   - 在 README 修订任务中又出现过把命令和 JSON 片段当文件名的情况
   - 现已做两层收敛：
     - `AiderExecutor` 增加严格任务提示，明确禁止把说明文本、命令、标题当成文件名
     - `worktreeguard` 在 `allowPaths` 为空时改为“任何新增改动都非法”，并增强对 `:`、`：`、带引号 UTF-8 路径的识别
   - 回归结果：
     - 旧问题不再“假成功”
     - 一旦越界，会被守卫直接拦下并把任务置为 `failed`

3. 样例项目产物已人工补齐并完成真实启动验证
   - 生成目录：`test-workspaces/workflow-v2-backend/repo/workflow-v2-backend-retest-223040`
   - 发现问题：
     - `go.mod` 为空文件
     - `README.md` 为空文件
   - 已人工补齐最小可运行内容后，实际执行：

```bash
cd test-workspaces/workflow-v2-backend/repo/workflow-v2-backend-retest-223040
go run main.go
curl http://127.0.0.1:8080/health
```

   - 返回结果：

```json
{"status":"ok"}
```

4. 已通过接口验证“人工最高权限接管”能力
   - 在工作流已进入 `accept` 后，直接调用 `/api/mvp/workflow/update-domain-task`
   - 将已完成的 `write_readme(317335697470001152)` 重置为 `pending`，工作流自动回到 `execute`
   - 后续又在该任务处于 `running` 时再次人工改写任务，并切换执行模式为 `chat`
   - 同时把 `verify_service(317335697478389760)` 改成 `chat + affectedResources=[]`
   - 两个任务均在人工改写后重新完成，执行阶段重新收口并再次进入 `accept`
   - 这说明：
     - 完成态任务可人工重开
     - 运行中任务可人工改写并重置
     - 后续批次会继续自动推进

5. `project-status` 的“实际工作中”判断存在 UTC 漂移，现已修复
   - 现象：
     - 任务刚启动时，`domain_running=1`
     - 但 `activeRunningTasks=0`、`stalledTaskCount=1`
   - 原因：
     - `mvp_domain_task` 中的运行时间按 UTC 写入
     - `project-status` 聚合时按本地时间解释，导致误判为 8 小时前的旧任务
   - 处理：
     - 新增 UTC 数据库时间归一化
     - 补充 `workflow_time_test.go`
   - 热重启后验证：
     - `project-status.lastActiveAt` 已从 `2026-04-08 15:27:23` 修正为本地时间 `2026-04-08 23:27:23`

6. 当前验收轮次状态
   - `acceptRunID`: `317349607774359552`
   - `acceptRound`: `2`
   - 当前结果：`decision=manual_review`，`score=83.2`
   - `project-status` 仍为 `accepting`
   - 说明当前系统把该轮验收视为“需要人工复核后再最终收口”，不是主链已完全自动闭环

### 6.5 人工放行后的最终收口修复

继续联调后，确认了验收阶段最后一个真实阻断：

1. `accept-status` 与 `domain-tasks` 的时间字段存在同类 UTC 漂移
   - 现象：
     - `accept-status.startedAt/finishedAt`
     - `domain-tasks.startedAt/completedAt`
     - `accept-issues.createdAt`
     - `accept-evidence.createdAt`
     都会显示为 UTC 原值，而不是本地时间
   - 处理：
     - 统一复用数据库 UTC 时间归一化逻辑
     - 对上述接口全部做展示层时间修正

2. `ManualApprove` 存在“人工已放行，但 workflow_run 不完成”的二次拦截问题
   - 现象：
     - `accept_run` 已变为 `completed/passed`
     - `accept stage` 也已 `completed`
     - 但 `workflow_run` 仍停留在 `accepting/current_stage=accept`
   - 根因：
     - `registry` 中 `accept -> complete` 回调会先进入自治中台
     - 当前项目命中了目标守卫，`accept.passed` 被转成 `notify_human`
     - `DecisionCenter` 返回 `Handled=true` 后直接吞掉 `completeWorkflow`
     - 当人工再次点击放行时，又会重复命中这层拦截，形成“accept 已通过但 workflow 不收口”的半完成态

3. 已落地修复
   - `ManualApprove` 调用 `completeTrigger` 时会带上“人工已明确放行”的上下文标记
   - `registry` 在识别到该标记后，直接执行 `completeWorkflow`，不再重复触发 `accept.passed` 的自治人审
   - 同时补了幂等保护：
     - 若 `accept stage` 已是 `completed`
     - 允许再次执行人工放行补偿收口，不会因为 `CompleteStage` 命不中而失败

4. 修复后实际验证结果
   - 再次调用：

```bash
POST /api/mvp/workflow/accept-approve
```

   - 返回：`code=0`
   - 随后接口与数据库状态已一致：
     - `project-status.workflowStatus = completed`
     - `project-status.currentStage = complete`
     - `project-status.progressPercent = 100`
     - `accept-status.status = completed`
     - `accept-status.decision = passed`
     - `accept-status.score = 100`
   - 数据库确认：
     - `mvp_workflow_run.status = completed`
     - `mvp_workflow_run.current_stage = complete`
     - 新增 `complete stage_run`
     - `mvp_workflow_run.finished_at` 已写入

5. 本轮最终结论
   - 这条 V2 主链已在纯后端、单并发、无 Docker、无前端环境下完整收口：

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

   - 样例项目已真实启动验证通过
   - 人工可以在完成态、运行态、验收态介入并推进到最终完成
