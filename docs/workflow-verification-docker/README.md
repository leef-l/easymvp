# Docker-First 验证修复

更新日期：2026-04-10

本目录记录 EasyMVP 新增的 `verification_repair_run` 最小闭环实现。目标不是把 `workflow completed` 直接等同于“可生产发布”，而是在项目完成后支持人工或飞书触发一次 Docker-first 的项目验证，并把问题转回正式返工链。

## 当前能力

- 新增独立验证运行：`mvp_verification_run`
- 新增验证问题与证据：`mvp_verification_issue`、`mvp_verification_evidence`
- 提供 API：
  - `POST /workflow/verification-start`
  - `GET /workflow/verification-status`
  - `GET /workflow/verification-issues`
  - `GET /workflow/verification-evidence`
  - `POST /workflow/verification-repair`
- 时间线新增事件：
  - `verification.started`
  - `verification.completed`
  - `verification.failed`
  - `verification.repair_requested`
- 飞书机器人新增动作：
  - `verification_start`
  - `verification_status`
  - `verification_repair`

## 验证策略

默认策略是 Docker-first：

1. 优先读取项目根目录 `.easymvp/verification.json`
2. 若项目未配置，则回退到项目分类 `verification_profile_json`
3. 分类级 `verification_gate_json` 会作为放行门，约束 runner、最少执行步数和必需检查类型
4. 若仍未配置 profile，则自动检测 `compose.yaml` / `docker-compose.yml`
5. 再次回退到 `Dockerfile`
6. 最后回退到本机命令启发式验证

自动检测会尝试：

- `docker compose up -d --build`
- `docker compose ps`
- `go test ./...`
- Node 项目的 `lint` / `test` / `build`

上述命令现在会默认注入一组低配服务器保护参数，避免 `npm install`、`pnpm install`、`pnpm build`、`turbo build` 一类命令把宿主机资源打满。默认限制包括：

- `NODE_OPTIONS=--max-old-space-size=1024`
- `npm_config_maxsockets=4`
- `pnpm child/workspace concurrency = 1`
- `GOMAXPROCS=1`
- `GOMEMLIMIT=768MiB`
- `COMPOSE_PARALLEL_LIMIT=1`
- `TURBO_CONCURRENCY=1`
- `prlimit --as=1536MiB`（Linux，非 Node 命令）

其中 `npm` 与 `pnpm` 会按命令类型分别注入参数，避免把只对 `pnpm` 生效的配置塞给 `npm` 造成告警；Linux 上的 `prlimit` 仅用于非 Node 命令，Node 侧继续依赖 `NODE_OPTIONS + npm/pnpm 并发限制`，避免把正常前端构建误杀成 OOM。

这些限制可通过 `mvp_config` 或 `admin-go/app/mvp/manifest/config/config.yaml` 下的 `engine.commandResource.*` 覆盖。

本轮实测补充一个资源侧结论：

- `web-antd` 这类体量的前端应用，完整 `vue-tsc --noEmit` 在 `768MiB` 内存上限下仍可能 OOM
- 在当前这台 `3.6Gi RAM / 2 vCPU` 服务器上，`1024MB` 和 `1280MB` 堆上限也已实测 OOM
- 同一台服务器在低优先级模式下，`1536MB` 堆上限可以跑完整个 `vue-tsc`，并返回真实类型错误列表
- 因此低负载宿主机上的默认静态验证应优先跑定向 `eslint` / 关键路径检查
- 若必须在当前宿主机执行全量类型检查，应通过受控脚本串行执行，例如 [scripts/web-antd-typecheck-safe.sh](../../scripts/web-antd-typecheck-safe.sh)
- 更高内存的验证 worker 或独立 CI 机仍然更适合作为默认生产级验证位，而不是把这类重检查长期压在生产宿主机上

如果项目提供 `.easymvp/verification.json`，可以显式指定：

- Docker compose 文件
- env 文件
- 自定义 setup / steps / teardown
- `docker_exec` 容器内验证命令

项目分类也可以提供默认验证模板和放行门：

- `verification_profile_json`
  - 作为该分类的默认 profile，在项目未提供 `.easymvp/verification.json` 时生效
- `verification_gate_json`
  - 当前支持 `allowedDecisions`、`minExecutedSteps`、`requiredCheckKinds`、`allowedRunnerTypes`
  - 典型用法：`software_dev` 至少要求执行 `test`；`game_dev` 可要求同时覆盖 `test + build`

当前系统已内置首批分类 gate 回填策略：

- `coding` 家族默认要求 `passed`，且至少执行 1 个 `test`
- `game_dev` 单独提高到至少执行 2 步，并覆盖 `test + build`
- `analysis / creative` 家族默认允许 `passed / manual_review`

本轮开始，系统不再只按“分类名”粗放验收，而是先解析一层标准化核验标准：

- `family_code`：能力家族，例如 `coding / analysis / creative`
- `project signals`：项目能力信号，例如 `Go 后端 / 前端交互应用 / 浏览器自动化能力`
- `verification standard`：标准化核验档位，例如
  - `coding.backend`
  - `coding.interactive_delivery`
  - `coding.game_client_runtime`
  - `coding.android_native_app`
  - `coding.ios_native_app`

这层标准会被 `review / verification / accept` 三个阶段复用，避免把规则散落在多个阶段各自硬编码。

当前已落地的标准化约束：

- `coding.backend`
  - 默认要求有通过的标准化验证
  - 默认要求覆盖 `test`
- `coding.interactive_delivery`
  - 默认要求有通过的标准化验证
  - 默认要求覆盖 `build + browser`
  - `review` 阶段必须出现浏览器级/端到端验证任务
  - `accept` 阶段必须拿到最新通过的浏览器级验证证据
  - `accept` 阶段必须能解析项目级 `experience_reviewer` 角色，用于体验评审
  - `experience_reviewer` 的展示名、提示词和是否可作裁决角色统一来自 `workflow.role_definitions`
- `coding.game_client_runtime / coding.android_native_app / coding.ios_native_app`
  - 标准层会统一挂载项目级 `experience_reviewer` 角色能力
  - 通过 `reviewProfile` 区分 Web 交互、游戏玩法、Android 原生、iOS 原生体验口径
  - 当前原生端先保留“角色能力口子 + 标准编码”，后续再逐步接入真机/模拟器自动化证据

同时新增一条工程铁律：验证、验收、阶段推进这类业务编排层不得直接访问 DB，新增数据能力必须先抽象 `service / repo interface`，再接入上层流程，避免规则链继续散落到控制器与阶段逻辑里。

本轮已完成的主链收口：

- `CategoryResolver` 改为通过 `ProjectRepo + ProjectCategoryRepo` 解析与回填分类
- `verification.Service` 改为通过 `ProjectRepo / ProjectCategoryRepo / DomainTaskRepo / Verification*Repo` 取项目、分类和验证证据
- `acceptance.RuleEngine` 改为通过 `TaskWorkspaceRepo / DomainTaskRepo / StageRunRepo / Verification*Repo` 执行规则评估
- `stage.accept.Service` 改为通过 `WorkflowRunRepo / ProjectRepo / StageRunRepo / AcceptRunRepo` 驱动验收编排

这样至少 `review / verification / accept` 的标准主链里，新增规则和验证能力时不需要再把 SQL 散落回编排层。

`accept` 阶段现在会按标准自动拉起一轮验证并等待结果，而不是只在界面里暴露一个手动入口。这样编码类项目进入验收时，会先完成标准化验证，再做最终裁决；缺少验证、验证未通过、缺少交互级证据、缺少标准要求的项目级体验评审角色，都会被标准规则直接拦住。

当前系统也已内置首批分类 profile 模板：

- `coding` 家族默认 `{"mode":"auto"}`，继续保留 Docker-first 自动探测
- `analysis / creative` 家族默认 `{"mode":"local"}`，直接走本机验证/人工复核链，避免无意义的 Docker 探测

## 修复闭环

验证发现问题后，前端或飞书可以把选中的验证问题转成返工。当前实现会：

1. 选取带 `domain_task_id` 的验证问题
2. 把对应领域任务标记为 `failed`
3. 强制启动 `rework` 阶段
4. 复用现有架构师分析与回流机制

这意味着验证结果已经进入正式工作流，而不是停留在一条孤立的告警记录里。

## 当前边界

- 自动检测的 Docker 验证仍是“通用启发式”，不是项目定制化编排
- 如果项目没有测试脚本，系统会进入 `manual_review` 倾向，而不是伪造“通过”
- 容器内精确验证建议通过 `.easymvp/verification.json` 明确声明

## 建议配置

建议每个需要生产级验证的项目补充：

- `.easymvp/verification.json`
- 项目所属分类的默认 `verification_profile_json / verification_gate_json`
- 明确的 `docker compose` 启动入口
- 后端测试命令
- 前端测试与构建命令
- 需要时的 `docker_exec` 步骤
