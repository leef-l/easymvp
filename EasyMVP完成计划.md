# EasyMVP 完成计划 — Agent Teams 任务拆分

> **创建时间**: 2026-04-29  
> **状态**: 进行中  
> **负责**: Agent Teams (Opus/Sonnet/Haiku 混合调度)  
> **更新规则**: 每完成一项任务立即更新本文档状态

---

## 0. 当前状况总结

### 已完成

- Go 后端 331 个文件，13 个服务模块，60+ HTTP 端点
- React 前端 36 个文件，10 个页面模块
- Electron 壳层完整（Main + Preload + CoreManager）
- SQLite 19 个 migration，41 张表
- 钱学森总纲设计文档体系完整
- MACCS 七阶段闭环工作流规范已定义

### 当前阻塞

1. **编译错误**: `worker_task_scheduler.go:294` — `StartBrainRunCommand` 无 `Timeout` 字段
2. **MACCS 闭环未落地**: 七阶段工作流（需求→方案→审核→执行→验收→交付→复盘）停留在文档，代码层面尚未对接
3. **前端 vue-vben-admin**: 废弃目录，仅剩 pnpm 缓存，需清理
4. **端到端未打通**: EasyMVP ↔ brain-v3 ↔ easymvp-brain 调用链未验证

---

## 1. 任务分组

### 组 A: 编译修复 & 基础健康 (P0, 阻塞一切)

| # | 任务 | 模型 | 状态 | 完成时间 |
|---|------|------|------|---------|
| A-01 | 修复编译错误：`StartBrainRunCommand` 添加 `Timeout` 字段 | Sonnet | ✅ 完成 | 2026-04-29 |
| A-02 | `go build ./...` 编译通过 | Sonnet | ✅ 完成 | 2026-04-29 |
| A-03 | `go vet ./...` 静态检查通过 | Sonnet | ✅ 完成 | 2026-04-29 |
| A-04 | 清理废弃的 `vue-vben-admin` 目录 | Haiku | ✅ 完成 | 2026-04-29 |
| A-05 | 清理项目根目录临时文件 (`_tmp_*`, `_fix_*`, `check_db.go`, `run_migrations.go`, `test_sse.go`, `test_req.json`, `core_*.log`) | Haiku | ✅ 完成 | 2026-04-29 |

### 组 B: MACCS 七阶段闭环 — 后端实现 (P0, 核心)

> 对应文档: `MACCS-EasyMVP-闭环工作流规范.md`

| # | 任务 | 模型 | 状态 | 完成时间 |
|---|------|------|------|---------|
| B-01 | **Phase 1 需求理解**: 新增 `RequirementDoc` 模型 + `requirements` 表 + migration | Opus | ✅ 完成 | 2026-04-29 |
| B-02 | **Phase 1 需求理解**: 新增 `RequirementService` — 解析/生成结构化需求文档 | Opus | ✅ 完成 | 2026-04-29 |
| B-03 | **Phase 1 需求理解**: 新增 `RequirementController` + API 路由 (`POST /requirements`, `POST /requirements/confirm`) | Sonnet | ✅ 完成 | 2026-04-29 |
| B-04 | **Phase 2 方案设计**: 新增 `SolutionDesign` 模型 + `solution_designs` 表 + migration | Opus | ✅ 完成 | 2026-04-29 |
| B-05 | **Phase 2 方案设计**: 新增 `DesignService` — Central Brain 调用生成方案 | Opus | ✅ 完成 | 2026-04-29 |
| B-06 | **Phase 2 方案设计**: 新增 `DesignController` + API 路由 (`GET /design`, `POST /design/confirm`) | Sonnet | ✅ 完成 | 2026-04-29 |
| B-07 | **Phase 3 方案审核**: 新增 `DesignReviewReport` 模型 + `design_reviews` 表 | Opus | ✅ 完成 | 2026-04-29 |
| B-08 | **Phase 3 方案审核**: 实现 `ReviewLoopController` — 自动审核→修正→重审闭环 (最多5轮) | Opus | ✅ 完成 | 2026-04-29 |
| B-09 | **Phase 3 方案审核**: 审核收敛条件判断 + 人工介入触发 | Sonnet | ✅ 完成 | 2026-04-29 |
| B-10 | **Phase 4 任务执行**: 升级 `TaskSchedulerWorker` — 支持拓扑分层并行调度 | Opus | ✅ 完成 | 2026-04-29 |
| B-11 | **Phase 4 任务执行**: 实现动态调整 — 停滞检测 + 并发冲突仲裁 | Sonnet | ✅ 完成 | 2026-04-29 |
| B-12 | **Phase 4 任务执行**: 实现 SSE 实时进度推送 (`GET /projects/{id}/progress-stream`) | Sonnet | ✅ 完成 | 2026-04-29 |
| B-13 | **Phase 5 验收测试**: 新增 `AcceptanceCriteria` 模型 + 按项目类型默认标准 | Opus | ✅ 完成 | 2026-04-29 |
| B-14 | **Phase 5 验收测试**: 实现多层验收 (单元/集成/E2E/安全/性能) | Opus | ✅ 完成 | 2026-04-29 |
| B-15 | **Phase 5 验收测试**: 验收失败处理 — 生成修复任务→重验收闭环 | Sonnet | ✅ 完成 | 2026-04-29 |
| B-16 | **Phase 6 项目交付**: 新增 `ProjectDelivery` 模型 + 交付物打包逻辑 | Opus | ✅ 完成 | 2026-04-29 |
| B-17 | **Phase 6 项目交付**: 用户验收→确认/修改/拒绝 流程 | Opus | ✅ 完成 | 2026-04-29 |
| B-18 | **Phase 7 复盘学习**: 新增 `ProjectRetrospective` 模型 + 自动复盘数据写入 | Opus | ✅ 完成 | 2026-04-29 |
| B-19 | **Phase 7 复盘学习**: 经验复用 — 相似项目推荐 + 预算调整 | Opus | ✅ 完成 | 2026-04-29 |

### 组 C: P0 专项实施清单落地 (P0, 与组 B 并行)

> 对应文档: `钱学森总纲设计/EasyMVP-专项实施清单.md`

| # | 任务 | 模型 | 状态 | 完成时间 |
|---|------|------|------|---------|
| C-01 | P0-01: `CompiledTask` 固化 `brain_kind` + `delivery_contract_json` + `verification_contract_json` 三字段 | Opus | ✅ 已存在 | 2026-04-29 |
| C-02 | P0-04: `verification_contract_json` 进入 accepting 主链 — 验收按合同跑 | Opus | ✅ 完成 | 2026-04-29 |
| C-03 | P1-01: 页面显示 verification contract gap (required checks / evidence / blocker) | Opus | ✅ 完成 | 2026-04-29 |
| C-04 | P1-02: 页面显示 escalation reason (类型/来源脑/建议动作) | Opus | ✅ 完成 | 2026-04-29 |
| C-05 | P2-02: 前后端术语统一 (验证/验收/返工/完成/人工检查点/替代验证通道) | Opus | ✅ 完成 | 2026-04-29 |

### 组 D: 前端 MACCS 闭环页面 (P1, 依赖组 B)

| # | 任务 | 模型 | 状态 | 完成时间 |
|---|------|------|------|---------|
| D-01 | 新增「需求输入」页面 — 用户输入自然语言需求 + 需求确认界面 | Opus | ✅ 完成 | 2026-04-29 |
| D-02 | 新增「方案预览」页面 — 展示 SolutionDesign + 用户确认/修改/拒绝 | Opus | ✅ 完成 | 2026-04-29 |
| D-03 | 新增「审核进度」页面 — 实时展示审核轮次、问题列表、修正进度 | Opus | ✅ 完成 | 2026-04-29 |
| D-04 | 升级「执行监控」页面 — 拓扑分层可视化 + 实时进度 + 动态调整通知 | Opus | ✅ 完成 | 2026-04-29 |
| D-05 | 升级「验收」页面 — 多层验收标准展示 + 通过/失败/缺失明确标识 | Opus | ✅ 完成 | 2026-04-29 |
| D-06 | 新增「交付物」页面 — 展示代码/文档/测试报告/统计 | Opus | ✅ 完成 | 2026-04-29 |
| D-07 | 新增「复盘」页面 — 计划vs实际对比、成功/失败分析 | Opus | ✅ 完成 | 2026-04-29 |
| D-08 | 升级路由 — 整合新页面到 router.tsx + shell.tsx 导航 | Opus | ✅ 完成 | 2026-04-29 |

### 组 E: 端到端验证 & 发布准备 (P2, 依赖 B+D)

| # | 任务 | 模型 | 状态 | 完成时间 |
|---|------|------|------|---------|
| E-01 | 编写 MACCS 闭环端到端集成测试 (Go test) | Opus | ✅ 完成 | 2026-04-29 |
| E-02 | 补充核心服务单元测试 (拓扑排序+验收标准) | Opus | ✅ 完成 | 2026-04-29 |
| E-03 | 更新 `docs/README.md` 索引 — 已在索引中，无需修改 | Opus | ✅ 完成 | 2026-04-29 |
| E-04 | 更新 `.env.example` — 已覆盖所需配置，无需修改 | Opus | ✅ 完成 | 2026-04-29 |
| E-05 | 更新 `EasyMVP-V3-当前上下文与重启接续说明.md` — 反映 MACCS 落地进度 | Opus | ✅ 完成 | 2026-04-29 |

### 组 F: easymvp-brain 新 MACCS 合约 Handler (P1, 依赖组 B 合约定义)

> 对应代码: `/www/wwwroot/project/brain-v3/brains/easymvp/`

| # | 任务 | 模型 | 状态 | 完成时间 |
|---|------|------|------|---------|
| F-01 | **requirement_analysis**: 实现 `handleRequirementAnalysis` handler — 解析用户需求生成结构化需求文档 | Opus | ✅ 完成 | 2026-04-29 |
| F-02 | **solution_design**: 实现 `handleSolutionDesign` handler — 根据需求文档生成技术方案 | Opus | ✅ 完成 | 2026-04-29 |
| F-03 | **design_review**: 实现 `handleDesignReview` handler — 多维度审核方案设计 | Opus | ✅ 完成 | 2026-04-29 |
| F-04 | **design_fix**: 实现 `handleDesignFix` handler — 根据审核问题修复方案 | Opus | ✅ 完成 | 2026-04-29 |
| F-05 | 更新 `handler.go` switch 分支 + `brain.json` capabilities | Opus | ✅ 完成 | 2026-04-29 |
| F-06 | 补充 handler 单元测试 (4 个新合约 + 合约名解析) | Opus | ✅ 完成 | 2026-04-29 |
| F-07 | `go build` + `go vet` 编译验证 (brain-v3) | Opus | ✅ 完成 | 2026-04-29 |

---

## 2. 依赖关系

```
组 A (编译修复) ──► 组 B + 组 C 可并行开始
                     │
                     ├──► 组 F (easymvp-brain 合约，依赖 B 的合约定义)
                     ▼
               组 D (前端页面，依赖 B 的 API)
                     │
                     ▼
               组 E (端到端验证，依赖 B + D)
```

**组内并行规则**:
- A-01~A-03 串行（修复→编译→检查）
- A-04, A-05 可与 A-01 并行
- B-01~B-03 串行（模型→服务→控制器），B-04~B-06 串行，两组可并行
- B-07~B-09 串行
- B-10~B-12 串行
- B-13~B-15 串行
- B-16~B-19 可并行
- C-01~C-05 串行（C-01 先行）
- D-01~D-07 按编号顺序，但 D-01/D-02 可并行
- E-01~E-05 可并行

---

## 3. 执行顺序建议

### 第一波: 解除阻塞 (立即)
- A-01, A-02, A-03 → 编译通过
- A-04, A-05 → 清理

### 第二波: 核心闭环 + P0 专项 (并行)
- B-01~B-06 (Phase 1 + Phase 2)
- C-01 (CompiledTask 三字段)

### 第三波: 审核+执行+验收 (依赖第二波)
- B-07~B-15 (Phase 3 + 4 + 5)
- C-02~C-04

### 第四波: 交付+复盘+前端 (依赖第三波)
- B-16~B-19 (Phase 6 + 7)
- D-01~D-08

### 第五波: 验证+收尾
- E-01~E-05
- C-05

---

## 4. 风险与注意事项

1. **禁止运行 pnpm/npm/yarn** — 前端依赖安装由用户手动完成
2. **配置改动三服务同步** — system/ai/mvp 三个 config.yaml
3. **每次代码更新重建全部4个容器** — Docker 全量重建
4. **手写逻辑放独立文件** — 防止 codegen 覆盖
5. **数据库 migration 用 `--no-tablespaces`** — init.sql 只含结构+种子数据

---

## 5. 完成定义

- [x] `go build ./...` 零错误 ✅
- [x] `go vet ./...` 零警告 ✅
- [x] MACCS 七阶段闭环 API 全部可调用 ✅ (19 个端点)
- [x] 前端七阶段页面全部可访问 ✅ (5 个新页面 + 2 个升级页面)
- [x] 端到端集成测试编写完成 ✅ (4 个测试用例)
- [x] P0 专项实施清单 4 项全部落地 ✅
- [x] 文档索引更新 ✅
- [x] easymvp-brain 新 MACCS 合约 Handler 4 个全部实现 ✅
- [x] brain-v3 编译通过 + 14 个测试全部 PASS ✅
