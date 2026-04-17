# 2026-04-18 Workflow V2 新建贪吃蛇项目实时回放问题记录

## 1. 记录目的

记录本轮“从新建项目开始，真实推进 React CLI + GoFrame v2 贪吃蛇小游戏”的执行过程里已经暴露出来的问题、自动返工结果和恢复建议，避免会话中断后重复排查。

## 2. 本轮对象

- 项目名：`snake-game-prod-20260418-030627`
- `projectID = 320666565290758144`
- `conversationID = 320666565429170176`
- `workflowRunID = 320666565479501824`
- 工作目录：`/www/wwwroot/project/easymvp/test-workspaces/snake-game-prod-20260418-030627`
- 回放日期：`2026-04-18`

## 3. 当前推进状态

截至当前记录时：

- 工作流阶段：`rework`
- 工作流状态：`failed`
- 总任务数：`21`
- 已完成任务：`10`
- 失败任务：`1`
- 当前无运行任务
- 当前失败原因：`架构师修复方案解析失败: 未解析到有效 JSON patch`

## 4. 已确认问题与处理

### 4.1 运行态接口字段口径错误导致状态查询失效

问题：

- `project-status / execution-status / stage-history` 曾直接返回异常
- 根因是 `workflow_runtime.go` 在无别名查询里仍写了 `wr.id / ss.id` 这类字段

处理：

- 已修复运行态查询字段选择
- 修复提交：`8990019 fix(mvp): repair runtime status field selection`
- 对应部署：`Deploy EasyMVP` run `24582591208`

结果：

- 运行态接口已恢复，可继续作为本轮监控入口

### 4.2 调度并发 `3` 会把首批 `aider` 任务一起压进内存高压

问题：

- 首批 3 个脚手架任务并发执行时，`aider` 进程长时间卡住
- 现场观察到 `mem_cgroup_handle_over_high`

处理：

- 将调度并发配置收敛为 `1`
- 对项目执行 `pause -> resume`
- 将残留 `running` 任务恢复为可继续推进的状态

结果：

- 当前本轮工作流已经切到串行推进
- 后续任务不再三条并发抢内存

### 4.3 文档与 CI 骨架任务曾因资源范围过宽触发越界修改

问题：

- `Init Docs Scripts and CI Scaffold` 首次执行时越界改动了整个 `.github/`

处理：

- 系统自动创建失败分析任务
- 将原任务修订为只允许修改：
  - `docs/README.md`
  - `scripts/.gitkeep`
  - `.github/workflows/ci.yml`

结果：

- 该任务第二次执行成功

### 4.4 前端核心逻辑任务首次触发 token limit，但 retry=1 后收口

问题：

- `Implement Frontend Core Game Logic` 首轮执行命中 `token limit`
- 涉及文件：
  - `frontend/src/logic/useSnakeGame.ts`
  - `frontend/src/logic/types.ts`
  - `frontend/src/logic/constants.ts`
  - `frontend/src/utils/storage.ts`

处理：

- 系统自动走 compact retry
- 在 `retry=1` 这一轮完成并回写成功

结果：

- 前端核心逻辑任务已完成
- 说明 compact retry 对“中等规模前端逻辑任务”有效

### 4.5 后端服务任务连续三轮 token limit，已进入自动返工拆分

问题：

- `Implement Backend Service and API` 原始任务一次要求实现：
  - `model`
  - `dao`
  - `service`
  - `controller`
  - `cmd`
  - 路由/中间件/配置
- 该任务已连续三轮命中 `token limit`

处理：

- 系统已创建失败分析任务：`320674332336459776`
- 返工结论已回写到原任务
- 新范围被收缩成更小的文件级步骤

返工结论：

- “单次同时创建 model、dao、service、controller 四层代码并完成路由注册与中间件配置”超出单次上下文上限
- 应拆成更小的 implementer 顺序步骤

当前收缩后的文件集：

- `internal/model/score.go`
- `internal/dao/score_dao.go`
- `internal/service/score_service.go`
- `internal/controller/score_controller.go`
- `cmd/cmd.go`

结果：

- 原任务已按更小范围重新启动并最终完成
- 完成时间：`2026-04-18 03:40:04`

### 4.6 前端 Canvas Renderer / VFX 任务在返工缩范围后仍然失败

问题：

- `Implement Frontend Canvas Renderer and VFX` 原始任务涉及：
  - `frontend/src/components/GameCanvas.tsx`
  - `frontend/src/renderers/snakeRenderer.ts`
  - `frontend/src/renderers/foodRenderer.ts`
  - `frontend/src/renderers/effectRenderer.ts`
- 该任务连续三轮命中 `token limit`

处理：

- 系统自动创建失败分析任务：`320676341546487808`
- 首轮返工后把任务收缩到：
  - `frontend/src/renderers/canvas-renderer.ts`
  - `frontend/src/types/renderer.ts`
  - `frontend/src/utils/perf-monitor.ts`

结果：

- 即使收缩到 3 个文件，这个任务仍再次命中 `token limit`
- 当前 `320667615200546822` 已落成正式 `failed`

结论：

- 对前端重渲染/动效任务，只按文件数量缩减还不够
- 后续需要进一步按“基础渲染循环 / 主题系统 / 性能监控 / 动态特效”拆成更小步骤

### 4.7 当前批次已出现“单任务失败但调度继续推进”的行为

问题：

- `320667615200546822` 失败后，系统没有停在该任务
- 调度继续启动了 `320667615200546823`

当前状态：

- `320667615200546822 = failed`
- `320667615200546823 = running`

影响：

- 恢复时如果只看当前 running task，容易忽略同批次里已经落下的 failed task
- 这会让人误判为“当前批次只是顺序继续推进”，实际已经有待补救缺口

### 4.8 第二轮 Canvas 返工分析未产出可解析 JSON patch，整条工作流停在 rework failed

问题：

- `Implement Frontend Canvas Renderer and VFX` 在第一次返工收缩后仍然失败
- 系统继续创建了第二个失败分析任务：`320678104559259648`
- 该分析任务虽然执行完成，但返工阶段日志明确记录：
  - `解析修复方案失败: 未解析到有效 JSON patch`

结果：

- `project-status` 已变成：
  - `status = failed`
  - `workflowStatus = failed`
  - `currentStage = rework`
- 当前无 running task
- 工作流不再继续推进到后续 batch

结论：

- 这次阻塞点已经不只是执行器 `token limit`
- 更深一层的问题是：失败分析任务本身没有稳定产出可机读的 JSON 修复方案
- 返工链在“分析任务 completed 但 patch 不可解析”这个分支上仍然会直接停机

### 4.9 返工解析兼容修复后，工作流已恢复，但新 patch 仍夹带了违反铁律的本机构建要求

问题：

- 针对 `plan_meta + tasks[]` 输出的返工解析兼容已经补进系统
- 重新触发 `rework` 后，第三个分析任务 `320681373486551040` 已经成功被回写
- 工作流已从 `rework failed` 回到 `execute`

恢复结果：

- `project-status` 已恢复为：
  - `status = executing`
  - `workflowStatus = executing`
  - `currentStage = execute`
- `Implement Frontend Canvas Renderer and VFX` 已重新启动
- 当前范围已收缩到单文件：
  - `frontend/src/renderers/canvas-renderer.ts`

新暴露问题：

- 新回写的任务描述里出现了：
  - “每阶段完成后运行 `npm run build` 验证”
- 这与本项目铁律直接冲突：
  - 本机禁止 `pnpm/npm/node` 相关构建验证
  - 构建与测试必须统一走 GitHub Actions

结论：

- 返工解析链已经不再因为 `tasks[]` 输出而停机
- 但返工结果本身还会把不合规的本机构建要求写回任务描述
- 后续还需要给返工 patch 增加“铁律过滤层”，避免把 `npm run build` / `go test` 这类本地验证要求注入执行任务

### 4.10 Canvas 单文件收缩已成功，但 UI 返工 patch 出现逆向扩容

结果：

- `Implement Frontend Canvas Renderer and VFX` 已在单文件模式下成功完成：
  - 文件：`frontend/src/renderers/canvas-renderer.ts`
  - 完成时间：`2026-04-18 04:07:37`

新问题：

- `Implement Frontend UI Panels and Theme` 失败后进入返工
- 新回写的资源范围不是更小，而是从原先 6 个文件扩大成了 11 个文件
- 新增内容包括：
  - `frontend/src/styles/animations.css`
  - `frontend/src/hooks/useTheme.ts`
  - `frontend/src/components/ThemeProvider.tsx`
  - `frontend/src/hooks/usePanelManager.ts`
  - `frontend/src/components/index.ts`

结论：

- 这条返工不是“收缩修复”，而是“任务扩容”
- 对高风险前端任务，这会直接重新抬高 token limit 风险

### 4.11 execute 回流后再次出现双任务并行和状态写回竞争

问题：

- 当前数据库状态显示：
  - `320667615200546823 = running`
  - `320667615200546824 = running`
- 与此前人工收敛到串行推进的目标不一致

日志异常：

- `handleFailure 更新状态失败: task=320667615200546824 ... context canceled`
- `OnTaskFailed 查询 workflow_run_id 失败: task=320667615200546824 ... context canceled`

影响：

- 这说明回流到 execute 后，调度层可能再次出现并发放开或状态竞争
- 如果不盯这个问题，后续任务状态可能再次失真

### 4.12 UI Panels and Theme 返工后仍失败，且返工方向继续失焦

结果：

- `Implement Frontend UI Panels and Theme` 当前已落成 `failed`
- 失败前经历了：
  - 原始 6 文件任务 `token limit`
  - 返工后被扩成 11 文件任务
  - 扩容后再次 `token limit`

结论：

- 当前返工分析没有把任务拆小，反而继续把 UI 任务做成“主题 + 动画 + provider + panels + hooks + index 聚合”
- 这说明 UI 类任务也需要像 Canvas 一样，改成更细的单面板/单样式级拆分

### 4.13 Controls Audio and Assets 是被 watchdog 从异常态拉起，不是正常串行推进

问题：

- `320667615200546824` 先出现：
  - `handleFailure 更新状态失败 ... context canceled`
  - `OnTaskFailed 查询 workflow_run_id 失败 ... context canceled`
- 随后被 `WatchdogV2` 判定卡死，并自动触发 `retry=1/3`

当前状态：

- `320667615200546824 = running`

影响：

- 这个任务的启动不是“前一任务正常收口后顺推”，而是 watchdog 从异常状态里拉起来的
- 说明当前执行链里，正常调度和 watchdog 恢复正在混合生效

### 4.14 Watchdog 熔断轮次统计再次失真

问题：

- 日志已出现：
  - `熔断触发: workflowRun=320666565479501824 reason=返工已达 6 轮（上限 3）`

结论：

- 当前 watchdog 的返工/重试预算统计口径不可信
- 它很可能把不同任务或不同返工链路累计到了同一个项目级计数里
- 这类日志已经不能直接当成“单任务真实返工轮次”

### 4.15 UI 与 Controls 两个任务出现交替抢跑

问题：

- `UI Panels and Theme` 一度失败
- `Controls Audio and Assets` 被 watchdog 拉起后开始运行
- 随后 `UI Panels and Theme` 又被 watchdog 自动重试到 `retry=3/3`
- 同时 `Controls Audio and Assets` 被回退成 `pending`

当前状态：

- `320667615200546823 = running`
- `320667615200546824 = pending`

影响：

- 这说明当前并不是单纯的串行推进
- 而是正常调度、返工回流、watchdog 重试三套机制在交替改写状态

## 5. 当前结论

这轮新建贪吃蛇项目的真实回放已经证明三件事：

1. Workflow V2 主链能在真实新项目上推进，不是只停留在设计/审核阶段
2. 任务过大和并发过高仍会在真实执行里稳定暴露出来
3. 自动返工链已经能把一部分“token limit 导致的大任务失败”转成可继续推进的收缩任务；但对前端任务，除了收缩粒度问题，还暴露出“失败分析格式不稳”“返工描述可能违背铁律”“返工后文件集反向扩容”“回流后并发/状态写回竞争”“watchdog 与正常调度交叉恢复”“watchdog 熔断口径失真”几类问题

## 6. 恢复建议

如果会话中断，恢复本轮时优先做这几件事：

1. 先查 `project-status` / `execution-status`
2. 不只看当前 `running` 任务，还要同时确认是否已经掉进 `rework failed`
3. 若 `Canvas Renderer and VFX` 仍是失败状态，优先继续细拆成更小的渲染子步骤
4. 若再次出现 token limit，不要直接继续硬重试，优先让失败分析继续把任务拆小
5. 若失败分析任务完成但工作流仍停在 `rework failed`，优先检查分析产物是否真的输出了可解析 JSON patch
6. 若返工 patch 把 `npm run build`、`go test` 之类本地验证动作写回任务描述，必须先过滤掉再继续执行
7. 若返工 patch 把文件集越改越大，优先拦截这次回写，而不是让更大的任务直接进 execute
8. 若 execute 回流后再次出现双任务同时 running，要立即核对调度并发和状态写回是否失真
9. 若某任务是被 watchdog 从卡死态拉起，要单独记录，不要把它误记成正常批次推进
10. 若 watchdog 日志出现“返工已达 N 轮（上限 3）”这类明显超额数字，必须把它当成统计口径异常，而不是直接采信
11. 即使返工链已经支持把大任务拆成同批次子任务，当前追踪主要仍靠 `source_task_id`；后续最好补显式 `rework_split` 关系记录或查询接口，把“原任务 -> 拆分出的 N 个子任务”展示清楚，否则排障和回放时仍不够直观
12. 当前又暴露出新的“执行阶段空转”状态：`project-status` 显示 `workflowStatus=executing / currentStage=execute / activeBatch=0 / activeRunningTasks=0 / isActuallyWorking=false`，但 `execution-status` 同时显示 `stageStatus=running / activeBatch=4 / pendingTasks=8 / escalatedTasks=1`
13. 这说明 `Implement Frontend UI Panels and Theme` 已经停在 `escalated`，`Controls Audio and Assets` 已完成，但系统既没有自动进入新的 `rework`，也没有按批次继续启动 `batch 4` 的 `Implement Frontend Easter Egg and API Client`
14. 当前卡点已经不是单个 implementer 任务，而是编排器在“存在 escalated 任务 + 无 running task”的情况下没有收口动作，形成了“执行阶段继续挂着，但其实已经不工作”的半停滞态
15. 恢复时必须把这类状态单独识别出来：不要把 `workflowStatus=executing` 误判成系统还在健康推进；需要优先检查为什么 `escalated` 没有触发新的返工或人工介入链路
16. 进一步排查已确认：`823` 的 `task.escalated -> rework` 实际发生过两轮，对应 `stageRun=320682770978312192 / 320684586164031488`，但这些返工都只回写了原任务，没有把 `823` 收口为可继续推进的状态；随后 `824` 又触发了新的 `rework -> execute`，把工作流推进到了新的 execute stage `320685913778688000`
17. 这导致系统形成了一个很具体的不一致：
   - `823` 仍然停在 `escalated`，属于 batch 3
   - `824` 已完成，新的 execute stage 也已启动
   - 调度器的批次门控把 `escalated` 视为“前序未完成”，因此 batch 4 不放行
   - 但 `checkAllDone` / 运行态口径又没有把这类状态收成失败或新的返工，于是 workflow 继续挂在 `executing`
18. 也就是说，这次空转的根因不是“rework 事件没发出去”，而是“旧的 escalated 任务没有被收口，而新的 execute stage 已经继续往后跑”，最终把批次门控和工作流状态机卡成了两套不一致的真相
