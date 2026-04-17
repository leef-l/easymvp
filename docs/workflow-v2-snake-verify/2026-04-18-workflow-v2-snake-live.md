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

## 5. 当前结论

这轮新建贪吃蛇项目的真实回放已经证明三件事：

1. Workflow V2 主链能在真实新项目上推进，不是只停留在设计/审核阶段
2. 任务过大和并发过高仍会在真实执行里稳定暴露出来
3. 自动返工链已经能把一部分“token limit 导致的大任务失败”转成可继续推进的收缩任务，但对前端重渲染/VFX 这类任务，当前不仅收缩粒度不够，连第二轮失败分析的 JSON patch 产出也不稳定

## 6. 恢复建议

如果会话中断，恢复本轮时优先做这几件事：

1. 先查 `project-status` / `execution-status`
2. 不只看当前 `running` 任务，还要同时确认是否已经掉进 `rework failed`
3. 若 `Canvas Renderer and VFX` 仍是失败状态，优先继续细拆成更小的渲染子步骤
4. 若再次出现 token limit，不要直接继续硬重试，优先让失败分析继续把任务拆小
5. 若失败分析任务完成但工作流仍停在 `rework failed`，优先检查分析产物是否真的输出了可解析 JSON patch
