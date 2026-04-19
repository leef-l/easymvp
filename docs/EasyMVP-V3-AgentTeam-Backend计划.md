# EasyMVP V3 Agent Team Backend 计划

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3文档总纲](./EasyMVP-V3文档总纲.md)
> 架构入口：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 目标：基于现有 V3 文档，给出一份可直接分配给多 Agent 并持续更新状态的后端开机计划。

## 1. 使用说明

本计划只覆盖 V3 后端实施，不覆盖 Electron 壳层和前端页面实现。

本计划用于：

1. 作为后端 Agent Team 的统一任务面板
2. 作为中断恢复后的继续执行入口
3. 作为并行实施时的依赖和边界说明

## 2. 状态字段规则

每个计划项都必须带 `status` 字段，默认值统一为：

`pending`

状态允许值建议固定为：

1. `pending`
2. `in_progress`
3. `blocked`
4. `done`

更新规则：

1. 开始执行前改为 `in_progress`
2. 被依赖阻塞时改为 `blocked`
3. 完成定义全部满足后改为 `done`
4. 不允许删除已存在计划项，只允许更新状态和补充备注

## 3. Agent Team 分工建议

后端建议按 6 类 Agent 分工：

1. `backend-architecture-agent`
   负责模块装配、GoFrame 结构、配置、启动链路
2. `backend-storage-agent`
   负责 SQLite、migration、repository、事务
3. `backend-api-agent`
   负责 controller、DTO、API 路由、错误码映射
4. `backend-runtime-agent`
   负责 `brain-v3` 接入、worker、事件流、诊断
5. `backend-aggregation-agent`
   负责 workspace/project/acceptance 聚合与查询视图
6. `backend-domain-brain-agent`
   负责 `easymvp-brain` 合同 DTO、schema 校验、调用服务和领域结果落库

原则：

1. 命令链路优先由 `backend-api-agent` 和 `backend-storage-agent` 协同
2. 查询与聚合链路优先由 `backend-aggregation-agent` 独立推进
3. `runtime` 和 `worker` 相关任务在核心存储与错误模型稳定后并行推进
4. `easymvp-brain` 任务不得塞进 runtime service，必须独立落在 domain-brain lane
5. `brain-v3` 内置 `code / browser / verifier / fault` 增强后，后端只扩适配与归一化层，不把对应能力回填进 `easymvp-brain`

## 4. 执行优先级原则

后端开机顺序必须遵守：

1. 先基础骨架，再存储，再命令主链路
2. 先稳定事实表和事务边界，再接运行时和事件流
3. 先做可验证的最小链路，再补聚合视图和后台同步
4. 一切实现以文档中已定对象和边界为准，不在实施期临时发明新主模型
5. 优先补 `easymvp-brain` 领域合同层，不再扩张通用 runtime 基座
6. `brain-v3` 新增工具协议优先在 `brain-v3 client / adapter` 层吸收，再向领域层输出稳定 DTO

## 4.1 与 `brain-v3` 内置脑的协作原则

后端在接入 `brain-v3` 最新内置脑能力时，职责分层必须固定为：

1. `brain-v3` 负责内置脑注册、`tools/list`、`tools/call`、权限控制和真实执行结果
2. EasyMVP runtime / adapter 负责把 `output`、`content[]`、`unsupported`、`denied` 归一化成可消费摘要
3. `easymvp-brain` 只消费领域输入，不感知具体内置脑工具名和原始返回结构
4. App 查询层和工作台只消费 EasyMVP 自己的投影视图，不直接暴露内置脑原始 payload

## 5. 计划总表

| 计划ID | 名称 | 优先级 | 依赖 | 是否可并行 | 对应文档位置 | 完成定义 | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| BE-001 | 初始化 GoFrame v2 后端工程骨架 | P0 | 无 | 否 | [EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-%E6%8A%80%E6%9C%AF%E6%A0%88%E4%B8%8E%E9%80%89%E5%9E%8B%E5%9F%BA%E7%BA%BF.md) §3-§4；[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go%E6%9C%AC%E5%9C%B0%E6%A0%B8%E5%BF%83%E6%9C%8D%E5%8A%A1%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1.md) §1-§5 | `apps/core` 目录结构、`main.go`、`internal/cmd`、基础 controller/service 装配可运行；`go test ./...` 通过 | pending |
| BE-002 | 建立本地配置与启动参数链路 | P0 | BE-001 | 是 | [EasyMVP-V3-本地配置与启动参数设计](./EasyMVP-V3-%E6%9C%AC%E5%9C%B0%E9%85%8D%E7%BD%AE%E4%B8%8E%E5%90%AF%E5%8A%A8%E5%8F%82%E6%95%B0%E8%AE%BE%E8%AE%A1.md) §1-§6；[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go%E6%9C%AC%E5%9C%B0%E6%A0%B8%E5%BF%83%E6%9C%8D%E5%8A%A1%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1.md) §3.1、§5 | 支持默认配置、本地配置文件、CLI 覆盖；启动时能解析 `data_root`、`db_path`、`brain_serve_base_url`、`safe_mode` | pending |
| BE-003 | 建立 SQLite 连接、PRAGMA 与 migration 执行器 | P0 | BE-001, BE-002 | 否 | [EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-%E6%95%B0%E6%8D%AE%E5%BA%93Schema%E6%80%BB%E8%AE%BE%E8%AE%A1.md) §1-§6；[EasyMVP-V3-完整SQLite建表与索引SQL终稿](./EasyMVP-V3-%E5%AE%8C%E6%95%B4SQLite%E5%BB%BA%E8%A1%A8%E4%B8%8E%E7%B4%A2%E5%BC%95SQL%E7%BB%88%E7%A8%BF.md) §2；[EasyMVP-V3-独立Migration文件正文终稿](./EasyMVP-V3-%E7%8B%AC%E7%AB%8BMigration%E6%96%87%E4%BB%B6%E6%AD%A3%E6%96%87%E7%BB%88%E7%A8%BF.md) §1-§9 | 启动时自动执行 migration；`schema_migrations` 写入成功；失败时有明确错误输出和回滚策略 | pending |
| BE-004 | 落地首批 migration `.sql` 文件并接入清单 | P0 | BE-003 | 是 | [EasyMVP-V3-独立Migration文件正文终稿](./EasyMVP-V3-%E7%8B%AC%E7%AB%8BMigration%E6%96%87%E4%BB%B6%E6%AD%A3%E6%96%87%E7%BB%88%E7%A8%BF.md) §2-§9；[EasyMVP-V3-完整SQLite建表与索引SQL终稿](./EasyMVP-V3-%E5%AE%8C%E6%95%B4SQLite%E5%BB%BA%E8%A1%A8%E4%B8%8E%E7%B4%A2%E5%BC%95SQL%E7%BB%88%E7%A8%BF.md) 全文 | 仓库内生成真实 migration 文件；版本号、文件名、checksum 约束稳定；本地空库初始化成功 | pending |
| BE-005 | 建立基础 repository 层与事务辅助 | P0 | BE-003 | 是 | [EasyMVP-V3-Service与Repository方法清单终稿](./EasyMVP-V3-Service%E4%B8%8ERepository%E6%96%B9%E6%B3%95%E6%B8%85%E5%8D%95%E7%BB%88%E7%A8%BF.md) 全文；[EasyMVP-V3-事务边界与一致性设计](./EasyMVP-V3-%E4%BA%8B%E5%8A%A1%E8%BE%B9%E7%95%8C%E4%B8%8E%E4%B8%80%E8%87%B4%E6%80%A7%E8%AE%BE%E8%AE%A1.md) 全文 | `projects`、`plan`、`task`、`runtime`、`acceptance` 基础 repository 可用；提供统一事务包装器；单元测试覆盖关键写入 | pending |
| BE-006 | 建立错误码、诊断记录与统一错误返回 | P0 | BE-001, BE-003 | 是 | [EasyMVP-V3-错误码与诊断分级设计](./EasyMVP-V3-%E9%94%99%E8%AF%AF%E7%A0%81%E4%B8%8E%E8%AF%8A%E6%96%AD%E5%88%86%E7%BA%A7%E8%AE%BE%E8%AE%A1.md) 全文；[EasyMVP-V3-GoFrame-Handler-DTO逐项终稿](./EasyMVP-V3-GoFrame-Handler-DTO%E9%80%90%E9%A1%B9%E7%BB%88%E7%A8%BF.md) 错误返回相关章节 | 统一 error code、HTTP 映射、诊断记录写库、日志输出全部打通；controller 不再直接返回裸错误 | pending |
| BE-007 | 实现系统基础接口与健康检查 | P1 | BE-002, BE-003, BE-006 | 是 | [EasyMVP-V3-API路由分组与命令查询边界设计](./EasyMVP-V3-API%E8%B7%AF%E7%94%B1%E5%88%86%E7%BB%84%E4%B8%8E%E5%91%BD%E4%BB%A4%E6%9F%A5%E8%AF%A2%E8%BE%B9%E7%95%8C%E8%AE%BE%E8%AE%A1.md) `system` 路由相关章节；[EasyMVP-V3-GoFrame-Handler-DTO逐项终稿](./EasyMVP-V3-GoFrame-Handler-DTO%E9%80%90%E9%A1%B9%E7%BB%88%E7%A8%BF.md) `system` DTO 相关章节 | `/api/v3/system/healthz`、基础诊断查询接口可用；返回配置、数据库、runtime 可达性概览 | pending |
| BE-008 | 实现项目创建主链路命令接口 | P0 | BE-004, BE-005, BE-006 | 否 | [EasyMVP-V3-API路由分组与命令查询边界设计](./EasyMVP-V3-API%E8%B7%AF%E7%94%B1%E5%88%86%E7%BB%84%E4%B8%8E%E5%91%BD%E4%BB%A4%E6%9F%A5%E8%AF%A2%E8%BE%B9%E7%95%8C%E8%AE%BE%E8%AE%A1.md) `projects` 命令章节；[EasyMVP-V3-GoFrame-Handler-DTO逐项终稿](./EasyMVP-V3-GoFrame-Handler-DTO%E9%80%90%E9%A1%B9%E7%BB%88%E7%A8%BF.md) `projects.create` 章节；[EasyMVP-V3-本地目录与项目工作区规范](./EasyMVP-V3-%E6%9C%AC%E5%9C%B0%E7%9B%AE%E5%BD%95%E4%B8%8E%E9%A1%B9%E7%9B%AE%E5%B7%A5%E4%BD%9C%E5%8C%BA%E8%A7%84%E8%8C%83.md) §3-§7 | 项目创建时完成 DB 事实写入、项目目录初始化、默认 profile 绑定、初始审计记录写入；失败可回滚 | pending |
| BE-009 | 实现工作区首页查询聚合接口 | P1 | BE-005, BE-008 | 是 | [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3%E5%B7%A5%E4%BD%9C%E5%8F%B0%E8%A7%86%E5%9B%BE%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%81%9A%E5%90%88%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md) §3-§4、快照接口相关章节；[EasyMVP-V3-API路由分组与命令查询边界设计](./EasyMVP-V3-API%E8%B7%AF%E7%94%B1%E5%88%86%E7%BB%84%E4%B8%8E%E5%91%BD%E4%BB%A4%E6%9F%A5%E8%AF%A2%E8%BE%B9%E7%95%8C%E8%AE%BE%E8%AE%A1.md) `workspace` 查询章节 | `workspace/home-view` 可返回项目卡片、阶段概览、待处理摘要、发布就绪摘要；查询层不直接暴露底表 | pending |
| BE-010 | 实现单项目 Workspace 查询聚合接口 | P1 | BE-005, BE-008 | 是 | [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3%E5%B7%A5%E4%BD%9C%E5%8F%B0%E8%A7%86%E5%9B%BE%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%81%9A%E5%90%88%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md) `ProjectWorkspaceView` 相关章节；[EasyMVP-V3-GoFrame-Handler-DTO逐项终稿](./EasyMVP-V3-GoFrame-Handler-DTO%E9%80%90%E9%A1%B9%E7%BB%88%E7%A8%BF.md) `workspace.project-view` 相关章节 | `project workspace` 快照返回顶部状态、阶段进度、live activity 首屏、action inbox、acceptance 概览 | pending |
| BE-011 | 实现 Plan 查询与计划对象读取接口 | P1 | BE-005, BE-008 | 是 | [EasyMVP-V3方案编译模型设计](./EasyMVP-V3%E6%96%B9%E6%A1%88%E7%BC%96%E8%AF%91%E6%A8%A1%E5%9E%8B%E8%AE%BE%E8%AE%A1.md) 全文；[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3%E5%B7%A5%E4%BD%9C%E5%8F%B0%E8%A7%86%E5%9B%BE%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%81%9A%E5%90%88%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md) `PlanView` 相关章节；[EasyMVP-V3-GoFrame-Handler-DTO逐项终稿](./EasyMVP-V3-GoFrame-Handler-DTO%E9%80%90%E9%A1%B9%E7%BB%88%E7%A8%BF.md) `plan` 相关章节 | 可读取 `PlanDraft`、`PlanReviewResult`、`CompiledPlan` 版本链；支持 diff、task projection、来源引用 | pending |
| BE-012 | 实现 Acceptance 查询聚合接口 | P1 | BE-005, BE-008 | 是 | [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3%E7%94%9F%E4%BA%A7%E7%BA%A7%E5%88%86%E7%B1%BB%E9%AA%8C%E6%94%B6%E4%BD%93%E7%B3%BB%E8%AE%BE%E8%AE%A1.md) 全文；[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3%E5%B7%A5%E4%BD%9C%E5%8F%B0%E8%A7%86%E5%9B%BE%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%81%9A%E5%90%88%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md) `AcceptanceView` 相关章节 | 返回 `surface/journey/evidence` 覆盖、issue、release gate、production status；能支撑验收页首版渲染 | pending |
| BE-013 | 实现 `brain-v3` client 与 Run 绑定模型 | P0 | BE-002, BE-003, BE-005 | 是 | [EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve%E6%8E%A5%E5%8F%A3%E6%8E%A5%E5%85%A5%E4%B8%8ERun%E7%94%9F%E5%91%BD%E5%91%A8%E6%9C%9F%E6%98%A0%E5%B0%84.md) §1-§6；[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go%E6%9C%AC%E5%9C%B0%E6%A0%B8%E5%BF%83%E6%9C%8D%E5%8A%A1%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1.md) `runtime` 章节 | `brain serve` 调用、run 创建、状态查询、取消/恢复、绑定表写入和最小映射都可用；适配 `tools/list` / `tools/call` 结构化协议与高风险工具 `unsupported` / `denied` 状态 | in_progress |
| BE-014 | 实现后台 worker manager 与首批 worker | P1 | BE-005, BE-013 | 是 | [EasyMVP-V3-后台Worker与任务调度设计](./EasyMVP-V3-%E5%90%8E%E5%8F%B0Worker%E4%B8%8E%E4%BB%BB%E5%8A%A1%E8%B0%83%E5%BA%A6%E8%AE%BE%E8%AE%A1.md) 全文；[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go%E6%9C%AC%E5%9C%B0%E6%A0%B8%E5%BF%83%E6%9C%8D%E5%8A%A1%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1.md) `worker` 章节 | `worker_manager` 可启动；`run_sync_worker`、`workspace_snapshot_refresh_worker` 至少落地；失败会写审计和诊断 | done |
| BE-015 | 实现实时事件归一化与 SSE 推送 | P1 | BE-013, BE-014 | 是 | [EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-%E5%AE%9E%E6%97%B6%E4%BA%8B%E4%BB%B6%E6%B5%81%E6%8E%A8%E9%80%81%E6%9C%BA%E5%88%B6%E8%AE%BE%E8%AE%A1.md) 全文；[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3%E5%B7%A5%E4%BD%9C%E5%8F%B0%E8%A7%86%E5%9B%BE%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%81%9A%E5%90%88%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md) `快照、流、详情三分` 章节 | 单条主 SSE 通道可工作；支持事件序号、续传、快照失效提示；前端可订阅 workspace 事件 | done |
| BE-016 | 实现 Action Inbox 与 Live Event 聚合落库 | P1 | BE-010, BE-014, BE-015 | 是 | [EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3%E5%B7%A5%E4%BD%9C%E5%8F%B0%E8%A7%86%E5%9B%BE%E6%A8%A1%E5%9E%8B%E4%B8%8E%E8%81%9A%E5%90%88%E6%8E%A5%E5%8F%A3%E8%AE%BE%E8%AE%A1.md) `ActionInboxItem`、`LiveEvent` 相关章节；[EasyMVP-V3-实时事件流推送机制设计](./EasyMVP-V3-%E5%AE%9E%E6%97%B6%E4%BA%8B%E4%BB%B6%E6%B5%81%E6%8E%A8%E9%80%81%E6%9C%BA%E5%88%B6%E8%AE%BE%E8%AE%A1.md) §2-§5 | 待处理项、事件流、快照刷新之间的聚合边界稳定；支持首页和项目页读取统一 inbox/event 数据 | pending |
| BE-017 | 实现 Acceptance 运行与裁决持久化主链路 | P1 | BE-005, BE-012, BE-013 | 否 | [EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3%E7%94%9F%E4%BA%A7%E7%BA%A7%E5%88%86%E7%B1%BB%E9%AA%8C%E6%94%B6%E4%BD%93%E7%B3%BB%E8%AE%BE%E8%AE%A1.md) §1-§6；[EasyMVP-V3-事务边界与一致性设计](./EasyMVP-V3-%E4%BA%8B%E5%8A%A1%E8%BE%B9%E7%95%8C%E4%B8%8E%E4%B8%80%E8%87%B4%E6%80%A7%E8%AE%BE%E8%AE%A1.md) 验收事务相关章节 | `AcceptanceRun`、`issues`、`judgements`、coverage 写入链路可用；支持 `functional_passed`、`production_passed`、`manual_release_required` | pending |
| BE-018 | 实现审计、诊断与恢复模式后端支撑 | P2 | BE-006, BE-014, BE-015 | 是 | [EasyMVP-V3-错误码与诊断分级设计](./EasyMVP-V3-%E9%94%99%E8%AF%AF%E7%A0%81%E4%B8%8E%E8%AF%8A%E6%96%AD%E5%88%86%E7%BA%A7%E8%AE%BE%E8%AE%A1.md) 全文；[EasyMVP-V3-本地配置与启动参数设计](./EasyMVP-V3-%E6%9C%AC%E5%9C%B0%E9%85%8D%E7%BD%AE%E4%B8%8E%E5%90%AF%E5%8A%A8%E5%8F%82%E6%95%B0%E8%AE%BE%E8%AE%A1.md) `safe-mode` 章节 | 支持恢复模式、诊断记录查询、关键失败事件审计化；中断后能看见未完成原因和恢复建议 | pending |
| BE-019 | 建立 `easymvp-brain` 合同 DTO 与 schema 校验层 | P0 | BE-001, BE-005, BE-006 | 是 | [EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿](./EasyMVP-V3-easymvp-brain-%E5%90%88%E5%90%8CJSON-Schema%E4%B8%8EDTO%E6%98%A0%E5%B0%84%E7%BB%88%E7%A8%BF.md)；[EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计](./EasyMVP-V3-easymvp-brain%E8%81%8C%E8%B4%A3%E8%BE%B9%E7%95%8C%E4%B8%8E%E8%BE%93%E5%85%A5%E8%BE%93%E5%87%BA%E5%90%88%E5%90%8C%E8%AE%BE%E8%AE%A1.md) | 6 类合同 DTO、统一 envelope、schema 校验器、`brain_contract_invalid` 错误包装可用 | done |
| BE-020 | 建立 `easymvp-brain` 调用服务与本地/远程统一客户端 | P0 | BE-019, BE-013 | 否 | [EasyMVP-V3专精大脑接入计划](./EasyMVP-V3%E4%B8%93%E7%B2%BE%E5%A4%A7%E8%84%91%E6%8E%A5%E5%85%A5%E8%AE%A1%E5%88%92.md#7.1-%60easymvp-brain%60-%E7%9A%84%E9%83%A8%E7%BD%B2%E6%A8%A1%E5%9E%8B)；[EasyMVP-V3-easymvp-brain-Manifest与ToolSchema设计](./EasyMVP-V3-easymvp-brain-Manifest%E4%B8%8EToolSchema%E8%AE%BE%E8%AE%A1.md) | 统一领域脑客户端能切换本地 sidecar / 远程服务，返回统一 envelope；不向上层泄漏内置脑工具名、原始 `content[]`、原始工具错误形态 | in_progress |
| BE-021 | 接通 `plan_review` / `plan_compile` / `acceptance_mapping` 三个 P0 领域合同 | P0 | BE-020, BE-008 | 否 | [EasyMVP-V3方案编译模型设计](./EasyMVP-V3%E6%96%B9%E6%A1%88%E7%BC%96%E8%AF%91%E6%A8%A1%E5%9E%8B%E8%AE%BE%E8%AE%A1.md)；[EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿](./EasyMVP-V3-easymvp-brain-%E5%90%88%E5%90%8CJSON-Schema%E4%B8%8EDTO%E6%98%A0%E5%B0%84%E7%BB%88%E7%A8%BF.md) | 可真实生成并落库 `PlanReviewResult`、`CompiledPlan`、`AcceptanceProfile` / `ProductionAcceptanceProfile` | in_progress |
| BE-022 | 接通 `completion_adjudication` / `repair_design` / `workspace_explanation` | P1 | BE-021, BE-016, BE-017 | 是 | [EasyMVP-V3-easymvp-brain-Prompt设计](./EasyMVP-V3-easymvp-brain-Prompt%E8%AE%BE%E8%AE%A1.md)；[EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计](./EasyMVP-V3-easymvp-brain%E8%81%8C%E8%B4%A3%E8%BE%B9%E7%95%8C%E4%B8%8E%E8%BE%93%E5%85%A5%E8%BE%93%E5%87%BA%E5%90%88%E5%90%8C%E8%AE%BE%E8%AE%A1.md) | completion decision、repair draft、workspace explanation 可真实生成并投影到页面；`unsupported` / `denied` 已在领域输入归一化和投影层被显式区分，不伪装成成功态 | pending |

## 6. 推荐开机执行顺序

推荐按下面顺序起跑：

1. BE-001
2. BE-002
3. BE-003
4. 并行启动：BE-004、BE-005、BE-006
5. BE-007、BE-008
6. 并行启动：BE-009、BE-010、BE-011、BE-012
7. BE-013
8. BE-014
9. 并行启动：BE-015、BE-016、BE-017
10. BE-019、BE-020、BE-021
11. BE-022
12. BE-018

备注：2026-04-19 / backend-runtime lane / 当前已收口 health、create、get、cancel、binding sync 与 CLI `resume` 适配、`RUN_001/002/003/004` 错误包装与首批 `run_event_index` 写入；`BE-014` 已完成最小 worker manager、`run_sync_worker`、`workspace_snapshot_refresh_worker` 与失败审计/诊断写库；`BE-015` 已补 `GET /api/v3/workspace/projects/{id}/events` 最小 SSE 主通道，支持命名事件、`id`、keepalive、`Last-Event-ID` 最小续传与 `workspace.snapshot_invalidated` 提示。
备注：2026-04-19 / backend-domain-brain lane / `BE-019` 已完成首版真实代码骨架：新增 `internal/model/braincontracts/` 六类合同 DTO、统一 envelope、错误 DTO，以及 `internal/service/easymvp_brain.go` 校验入口；`go test ./internal/model/braincontracts ./internal/service` 已通过。
备注：2026-04-19 / backend-domain-brain lane / `BE-020` 已进入中后段：新增 `internal/service/easymvp_brain_client.go`，已具备 `easymvpBrain` 配置解析、模式选择、`/rpc` `brain/execute` 调用、summary → envelope 解码；`internal/service/easymvp_brain.go` 已补 `plan_review / plan_compile / acceptance_mapping` typed wrapper，且合同 DTO 已统一改为 `json.RawMessage` 以保持对象字段内联 JSON 语义；当前剩余工作是与业务服务正式接线并完成 sidecar 实际接合。
备注：2026-04-19 / backend-boundary lane / 已按 `brain-v3` 最新内置脑能力补充协作原则：后端新增工作应集中在 `brain-v3 client / adapter` 对 `tools/list` / `tools/call`、`unsupported` / `denied` 的吸收和归一化，不再把 `code / browser / verifier / fault` 能力回填进 `easymvp-brain`。
备注：2026-04-19 / backend-domain-brain lane / `BE-021` 已进入收口段：新增 `POST /api/v3/projects/{id}/plan/compile` API/Controller；`service.Plan().CompilePlan(...)` 已可在写侧自动补 `CreateInitialDraft(...)`、调用 `CallPlanReview / CallPlanCompile`，并写入 `workflow_plan_review_results`、`workflow_compiled_plans`、`workflow_compiled_tasks`；同时已补 migration `0010_add_acceptance_profile_tables.sql`、`service.Acceptance().MapAcceptanceProfiles(...)` 与 `acceptance_profiles / production_acceptance_profiles` 首版持久化。当前已再补 `POST /api/v3/projects/{id}/acceptance-profiles/refresh` 正式命令入口、`plan compile` 后自动 refresh acceptance profiles、`workflow_compiled_tasks -> domain_tasks` 首版桥接、`acceptance-runs` 的可选 `task_id` 与项目状态推进、`repair-draft` 的 `artifact_refs` 透传、acceptance failure 自动 repair 的空 task panic 修复，以及 `Runtime().SyncRunBindingCommand(...)` 首次 terminal 状态下的自动 adjudication 收口；剩余工作集中在页面联动、回放/诊断投影与前端消费链验证。

## 7. 并行执行建议

适合并行的组合：

1. `backend-storage-agent`：BE-003、BE-004、BE-005
2. `backend-api-agent`：BE-006、BE-007、BE-008
3. `backend-aggregation-agent`：BE-009、BE-010、BE-011、BE-012
4. `backend-runtime-agent`：BE-013、BE-014、BE-015、BE-016
5. `backend-domain-brain-agent`：BE-019、BE-020、BE-021、BE-022
6. `backend-architecture-agent`：持续跟进 BE-001、BE-002、BE-018

并行约束：

1. BE-008 之前不得跳过 BE-003 和 BE-005
2. BE-015 之前必须先有 BE-013 和 BE-014
3. BE-017 不应在 acceptance 表和 repository 未稳定前提前开做
4. BE-021 之前必须先有 BE-019 和 BE-020
5. 不允许把 `easymvp-brain` 合同实现塞进 runtime service 里完成

## 8. 中断恢复规则

为避免中断后丢失上下文，每次计划项完成或阻塞时都必须更新本文件对应行的 `状态` 字段。

建议额外补一行简短执行备注，格式如下：

```text
备注：完成时间 / 执行人 / 关键提交或关键文件
```

如果出现中断，恢复顺序如下：

1. 先看哪些计划项仍是 `in_progress`
2. 再看其依赖项是否已经 `done`
3. 优先恢复所有 `P0` 且未完成项
4. 若存在多个可恢复项，优先选择“可独立验证”的项继续推进

## 9. 完成标准

本计划作为后端开机计划，最终完成的判断标准是：

1. 所有 `P0` 计划项状态都为 `done`
2. 至少 `BE-009` 到 `BE-017` 的主链路都已落地
3. 至少 `BE-019` 到 `BE-021` 的领域脑 P0 主链路已落地
4. 后端可以独立提供：
   1. 项目创建
   2. workspace 聚合查询
   3. plan 查询
   4. acceptance 查询
   5. `brain-v3` 运行时对接
   6. `easymvp-brain` 领域合同调用
   7. SSE 实时事件
   8. 审计与诊断支撑

当以上条件满足时，V3 后端才算进入“可联调”状态，而不是仅有骨架状态。
