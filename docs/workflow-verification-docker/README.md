# 验证修复与历史 Docker 方案归档

更新日期：2026-04-13

本目录记录 EasyMVP 的 `verification_repair_run` 闭环实现，同时归档早期 Docker-first 验证方案。自 2026-04-13 起，仓库现行铁律已经切换为“测试与编译统一只走 GitHub Actions”；本机、宿主机、AI 会话中的 `go test`、`go build`、`pnpm build`、`pnpm exec vite build`、`pnpm exec vue-tsc`、`npm/pnpm test`、`docker build` 都不再是允许的验收路径。

## 当前口径

- 验证阶段只读取 GitHub Actions workflow run 及 `.easymvp/ci/latest.json`，不再自动执行本机命令、Docker compose 或 Dockerfile 回退。
- `.easymvp/verification.json` 与分类级 `verification_profile_json` 仍可作为配置输入，但 legacy `local / dockerfile / docker_compose / auto` 模式会被统一归一到 `github_actions`，并给出停用告警。
- 正式验收只认 GitHub Actions 的 workflow run、日志与 artifact；脚本资产和历史 Docker 方案只作为归档资料。

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

当前默认策略是 GitHub Actions-first：

1. 优先读取项目根目录 `.easymvp/verification.json`
2. 若项目未配置，则读取项目分类 `verification_profile_json`
3. 分类级 `verification_gate_json` 会作为放行门，约束 runner、最少执行步数和必需检查类型
4. 检测 `.github/workflows/*.yml` / `.github/workflows/*.yaml`
5. 读取 `.easymvp/ci/latest.json`
6. 将 `checks / checkKinds` 映射为 `test / build / browser` 等标准检查项，并据此生成验证结论

现行实现不会在验证阶段直接执行：

- `go test ./...`
- `go build ./...`
- Node 项目的 `lint / test / build / typecheck / bundle`
- `docker compose up -d --build`
- `docker build`

如果 workflow 或 `latest.json` 缺失，系统会直接给出配置/证据缺口，不再回退到本机执行。

## 历史 Docker 方案记录

下述内容保留的是此前 Docker-first 方案的资源限制和排障结论，仅作为历史背景，不代表当前执行入口。

上述历史命令当时会默认注入一组低配服务器保护参数，避免 `npm install`、`pnpm install`、`pnpm build`、`turbo build` 一类命令把宿主机资源打满。默认限制包括：

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

- `web-antd` 专用受控脚本现已固定为 `1 core + 1G memory` 硬限制：
  - [scripts/web-antd-typecheck-safe.sh](../../scripts/web-antd-typecheck-safe.sh)
  - [scripts/web-antd-build-safe.sh](../../scripts/web-antd-build-safe.sh)
  - [scripts/web-antd-verify-build-safe.sh](../../scripts/web-antd-verify-build-safe.sh)
  - [scripts/web-antd-entry-bundle-safe.sh](../../scripts/web-antd-entry-bundle-safe.sh)
  - [scripts/web-antd-workflow-entry-bundles-safe.sh](../../scripts/web-antd-workflow-entry-bundles-safe.sh)
  - [scripts/web-antd-workflow-bundle-safe.sh](../../scripts/web-antd-workflow-bundle-safe.sh)
  - [scripts/web-antd-entry-typecheck-safe.sh](../../scripts/web-antd-entry-typecheck-safe.sh)
  - [scripts/web-antd-workflow-pages-typecheck-safe.sh](../../scripts/web-antd-workflow-pages-typecheck-safe.sh)
- 具体口径是：`systemd-run --scope + AllowedCPUs=0 + CPUQuota=100% + MemoryMax=1G + MemorySwapMax=0 + NODE heap=768MB`
- 在当前这台 `3.6Gi RAM / 2 vCPU` 服务器上，按这条硬限制实跑得到的结果是：
  - `vue-tsc --noEmit` 在 `heap=768MB` 下直接触发 V8 OOM
  - 同样保持 `1G cgroup`，把 heap 提到 `896MB` 后，进程会被 scope 直接终止
  - `vite build --mode production` 在 `heap=768MB` 下会在 `transforming...` 阶段被 `1G` 限制终止，退出码 `143`
- 这意味着：当前 `web-antd` 的 full typecheck/build 不是“还没跑”，而是“已确认无法在 1 core / 1G 限制下完整通过”
- 在同一条硬限制下，`objective / situation / dashboard / execution / review / accept / verification / autonomy` 8 个 `workflow` 页面已通过单入口类型检查
- 当前还额外具备 `workflow` 最小 bundle、单页面 bundle、轻量验证构建这三条静态补齐的构建验证入口；它们都应继续遵守同样的 `1 core / 1G` 硬限制
- 其中轻量验证构建当前还会关闭附加插件，并禁用 `minify / cssMinify / reportCompressedSize / modulePreload / treeshake`，优先降低验证峰值
- 这些拆分验证路径可作为当前 `workflow` 控制台改动的受限验证入口，但它们不能替代 full typecheck/build
- 若后续必须坚持这条资源上限，下一步应继续降低 full typecheck/build 峰值，而不是恢复裸跑

当前 `.easymvp/verification.json` 建议只保留 GitHub Actions 验证模式声明，例如：

- `{"mode":"github_actions"}`
- 与 `.easymvp/ci/latest.json` 对齐所需的兼容元信息
- 不再新增本机 `steps / setup / teardown` 或 `docker_exec` 执行配置

项目分类仍可提供默认验证模板和放行门：

- `verification_profile_json`
  - 作为该分类的默认 profile 输入，在项目未提供 `.easymvp/verification.json` 时生效
  - 推荐只声明 `{"mode":"github_actions"}`
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

当前系统会把首批分类 profile 模板统一归一到 GitHub Actions：

- `coding` 家族默认按 `{"mode":"github_actions"}` 处理
- `analysis / creative` 家族若仍保留 `{"mode":"local"}` 一类旧值，会在验证阶段被标记为停用 legacy 配置并归一到 `github_actions`

## 修复闭环

验证发现问题后，前端或飞书可以把选中的验证问题转成返工。当前实现会：

1. 选取带 `domain_task_id` 的验证问题
2. 把对应领域任务标记为 `failed`
3. 强制启动 `rework` 阶段
4. 复用现有架构师分析与回流机制

这意味着验证结果已经进入正式工作流，而不是停留在一条孤立的告警记录里。

## 当前边界

- 当前验证结果依赖 GitHub Actions workflow 与 `.easymvp/ci/latest.json` 的同步质量
- 如果 `latest.json` 没有 `checks / checkKinds`，系统无法把 CI 结果映射到标准检查类型
- 历史 Docker/profile 字段目前仅保留兼容读取与停用告警，不再触发本机执行

## 建议配置

建议每个需要生产级验证的项目补充：

- `.github/workflows/backend-guard.yml`、`.github/workflows/web-antd-guard.yml`、`.github/workflows/deploy.yml` 一类权威 workflow
- `.easymvp/ci/latest.json`
- `.easymvp/verification.json` 中显式声明 `{"mode":"github_actions"}`
- 项目所属分类的默认 `verification_profile_json / verification_gate_json`
- 明确的 `checks / checkKinds` 输出，覆盖 `test / build / browser` 等标准检查类型
- workflow run 对应的日志与 artifact，作为正式验收依据
