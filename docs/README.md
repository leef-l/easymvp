# EasyMVP 文档目录

> 更新日期：2026-04-11

仓库内现在主要保留仍与当前实现对齐的文档。旧的迁移方案、阶段设计稿、过渡期说明和一次性联调现场记录已从工作树删除；如果需要追溯，请直接查看 `git` 历史。

当前文档按“使用与运维”“核心实现”“产品与规划”“专项验证记录”四组维护；如有新增方案文档，优先补到本索引后再进入主分支。

## 使用与运维

- [EasyMVP使用文档](EasyMVP使用文档.md) — 项目创建、审核、执行、验收与协作入口
- [AI应用使用指南](AI应用使用指南.md) — AI 供应商、套餐、模型、执行引擎和手工任务配置
- [Docker开发环境说明](Docker开发环境说明.md) — 开发 compose 差异、启动脚本和本地热重启方式
- [Docker生产环境变量说明](Docker生产环境变量说明.md) — 生产 compose 使用的环境变量和部署注意事项

## 核心实现

- [EasyMVP架构设计文档](EasyMVP架构设计文档.md) — 仓库结构、服务边界、核心数据模型
- [WorkflowRun阶段化工作流引擎重构架构设计文档](WorkflowRun阶段化工作流引擎重构架构设计文档.md) — 当前 Workflow V2 的运行实体、阶段链路和调度关系
- [执行器接入架构设计文档](执行器接入架构设计文档.md) — 统一执行器接口、注册表与分发链路
- [EasyMVP对接OpenHands与Aider引擎设计实现文档](EasyMVP对接OpenHands与Aider引擎设计实现文档.md) — AI 模块执行引擎配置与 MVP 工作流关系
- [GitWorktree任务级环境隔离设计文档](GitWorktree任务级环境隔离设计文档.md) — 任务级 `git worktree` 隔离的当前实现与边界
- [飞书机器人接入与移动端控制设计文档](飞书机器人接入与移动端控制设计文档.md) — 当前飞书/Telegram 协作接入与通知能力

## 产品与规划

- [EasyMVP全面分析与优化路线图](EasyMVP全面分析与优化路线图.md) — 基于当前实现、联调结果与外部竞品观察形成的优化方向与 90 天路线图
- [EasyMVP研发执行版](EasyMVP研发执行版.md) — 将 90 天路线图拆成研发任务、依赖、负责人建议、阶段验收与周执行节奏
- [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md) — 本轮收尾执行计划、阶段进度、验证结果与剩余风险
- [EasyMVP体验评审师接入方案与计划](EasyMVP体验评审师接入方案与计划.md) — 体验评审师在 Workflow V2 中的接入现状、缺口、实施阶段、验收口径与交接边界

## 专项验证记录

- [workflow-v2-create-verify/README.md](workflow-v2-create-verify/README.md) — Workflow V2 创建项目与全链路完成真实验证、问题记录、修复与复测材料
- [workflow-v2-snake-verify/README.md](workflow-v2-snake-verify/README.md) — React + GoFrame v2 贪吃蛇样例项目重建、代码验证、Workflow V2 全链路完成记录与问题处理材料
- [workflow-verification-docker/README.md](workflow-verification-docker/README.md) — Docker-first 项目验证、问题落库、飞书触发与返工闭环说明
