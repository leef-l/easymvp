# EasyMVP 文档目录

> 更新日期：2026-04-17

仓库内保留的文档已经按“当前主入口”“使用与运维”“架构与实现”“治理与专项闭环”“专题规划”“专项验证记录”六组整理。

文档使用规则：

- 判断“当前做到哪了、接下来做什么”，以 [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md) 和 [EasyMVP研发执行版](EasyMVP研发执行版.md) 为准。
- 判断“哪些系统治理项已经形成稳定基线”，以 [系统机制类问题彻底根治计划](系统机制类问题彻底根治计划.md)、[Watchdog双保险实施计划](Watchdog双保险实施计划.md) 和 [EasyMVP直连DB收口暂停说明](EasyMVP直连DB收口暂停说明.md) 为准。
- 专项验证目录只保存回放、联调和复测证据，不再承担主计划入口职责。

## 当前状态

- 当前代码主链收尾已完成，当前阶段仍未完成。
- 当前唯一剩余主阻塞是：`web-antd` 虽已形成“当前工作区代码 + 当前计划口径下 full typecheck/build”的 GitHub Actions 权威结论，但该结论仍是失败。
- 2026-04-17 最新 `Web Antd Guard` run 为 `24574466599`。
- 该 run 已真实执行 `full-typecheck / full-build / verify-build / workflow-bundle / entry-bundles`。
- 该 run 当前真实结果为：`full typecheck` 失败、`full build` 失败，`verify-build / workflow-bundle / entry-bundles` 通过。
- 该 run 已成功上传 `web-antd-guard-ci-latest` artifact；当前本地 `.easymvp/ci/latest.json` 已按该 artifact 同步。
- 2026-04-17 最新 `Backend Guard` run 为 `24574470208`，`validate-regression / test-backend / test-codegen / build-services` 已全部通过，并成功上传 `backend-guard-ci-latest` artifact。
- 当前测试与编译的新铁律是：统一只走 GitHub Actions，不再接受本机执行。
- 当前权威验证入口：
  - `.github/workflows/backend-guard.yml`
  - `.github/workflows/deploy.yml`
  - `.github/workflows/web-antd-guard.yml`
- 仓库里保留的 `scripts/web-antd-*-safe.sh` 属于历史受控脚本资产，不再作为人工验证入口。
- 不允许在宿主机、本地开发机或 AI 会话中直接执行 `go test`、`go build`、`pnpm build`、`pnpm exec vite build`、`pnpm exec vue-tsc`、`npm/pnpm test` 一类测试或编译命令。

## 当前主入口

- [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md) — 当前收尾状态、验证约束、剩余风险与下一批次排期主入口
- [EasyMVP研发执行版](EasyMVP研发执行版.md) — 当前研发排期、任务依赖、阶段目标与近期可立即开工项
- [EasyMVP全面分析与优化路线图](EasyMVP全面分析与优化路线图.md) — 面向产品和管理视角的 90 天方向图，保留中期任务池
- [EasyMVP使用文档](EasyMVP使用文档.md) — 面向操作者的创建项目、执行工作流和协作入口说明

## 使用与运维

- [AI应用使用指南](AI应用使用指南.md) — AI 供应商、套餐、模型、执行引擎和手工任务配置
- [Docker开发环境说明](Docker开发环境说明.md) — 开发 compose 差异、启动脚本和本地热重启方式
- [Docker生产环境变量说明](Docker生产环境变量说明.md) — 生产 compose 使用的环境变量和部署注意事项

## 架构与实现

- [EasyMVP架构设计文档](EasyMVP架构设计文档.md) — 仓库结构、服务边界、核心数据模型与整体分层
- [WorkflowRun阶段化工作流引擎重构架构设计文档](WorkflowRun阶段化工作流引擎重构架构设计文档.md) — Workflow V2 的运行实体、阶段链路和调度关系
- [执行器接入架构设计文档](执行器接入架构设计文档.md) — 执行器接口、注册表与运行分发链路
- [EasyMVP对接OpenHands与Aider引擎设计实现文档](EasyMVP对接OpenHands与Aider引擎设计实现文档.md) — AI 模块执行引擎配置与 MVP 工作流关系
- [GitWorktree任务级环境隔离设计文档](GitWorktree任务级环境隔离设计文档.md) — 任务级 `git worktree` 隔离实现、回写策略与边界
- [飞书机器人接入与移动端控制设计文档](飞书机器人接入与移动端控制设计文档.md) — 飞书/Telegram 协作接入与移动控制入口

## 治理与专项闭环

- [EasyMVP工程铁律](EasyMVP工程铁律.md) — 当前研发必须遵守的工程约束与验收铁律
- [系统机制类问题彻底根治计划](系统机制类问题彻底根治计划.md) — 系统机制治理总纲，当前口径为“主体已完成，进入防回退”
- [Watchdog双保险实施计划](Watchdog双保险实施计划.md) — watchdog / event stream / recovery 双保险专项，当前口径为“主干已落地并完成专项验收”
- [EasyMVP直连DB收口暂停说明](EasyMVP直连DB收口暂停说明.md) — “禁止直接 DB、统一 repo 化”治理线的完成基线与恢复顺序说明

## 专题规划

- [EasyMVP体验评审师接入方案与计划](EasyMVP体验评审师接入方案与计划.md) — `experience_reviewer` 接入现状、缺口、实施阶段与交接边界

## 专项验证记录

- [workflow-v2-create-verify/README.md](workflow-v2-create-verify/README.md) — Workflow V2 创建项目与全链路完成的真实验证、问题记录、修复与复测材料
- [workflow-v2-snake-verify/README.md](workflow-v2-snake-verify/README.md) — React + GoFrame v2 贪吃蛇样例项目重建、验证与全链路完成记录
- [workflow-verification-docker/README.md](workflow-verification-docker/README.md) — 验证闭环实现与历史 Docker 方案归档，现行测试/编译只认 GitHub Actions
