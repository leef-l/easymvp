ConcurrencySpec、LeaseScope，v6（2026-04-17）全面细化 8 大核心子系统设计（§13），新增独立规格文档 35 系列，v7（2026-04-17）补充 P0 差距：Brain Capability 标签体系、跨脑通信协议、LeaseManager 独立文档、端到端时序与模块依赖图，v8（2026-04-18）新增 §15 上层产品接入边界（runtime vs domain 职责划分、四条接入铁律）

---

## 0. 版本背景

v1 的架构方向没有错：**Brain-first, manifest-driven, runtime-pluggable**。

但 v1 的排序有问题——过度押注 Manifest/Package/Marketplace 这条"生态先行"路线，而代码实际已经在往"多脑运行时控制面"走。行业对标（OpenClaw 2026.4.15
---
torPrompt` | ✅ 已实现 |
| LLM Proxy 路由 | `sdk/kernel/llm_proxy.go` | ✅ 已实现 |
| 6 个专精大脑 | `brains/data,quant,code,browser,verifier,fault` | ✅ 已实现 |
| Quant WebUI | `brains/quant/webui/` | ✅ 已实现 |
| serve HTTP API | `cmd_serve.go` | ✅ 已实现 |
| doctor 诊断 | `cmd_doctor.go`（8 项检查） | ✅ 已实现 |
| 多任务支持 | `runManager` + `SubtaskRequest` | ⚠️ 部分实现 |

文档应该跟上代码，而不是代码追文档。

---

## 3. v3 的七个核心概念

### 3.1 Brain

`Brain` 
---
 |
|------|--------|------|--------|
| C-1 | Workflow Engine | TaskExecution + Dispatch Policy 已就绪 | 任务图（DAG）：节点是 TaskExecution，边是 Flow Edge（materialized / streaming） |
| C-2 | Background Job | serve HTTP API 已有 | 长驻任务：Data Brain 持续采集、Quant Brain 自动交易、watch/retry/resume |
| C-3 | Manifest v1 解析器 | 33 号文档 schema 已定义 | orchestrator 读取 manifest.json 做脑发现，替代硬编码注册 |
| C-4 | 本地 Brain 管理 | — | `brain install
---
    TaskCount       int                // 执行任务数
    SuccessRate     float64            // 成功率
    AvgLatency      time.Duration      // 平均耗时
    ConfidenceTrend float64            // 置信度趋势（上升/下降）
    DomainMetrics   map[string]float64 // 领域特化指标
}
```

Central 消费这些指标来更新 L1 的 brain 能力画像——**L0 的学习结果自然流入 L1，L1 的评估反过来影响 Central 对该 brain 的委托频率，形成正反馈循环。**

#### 学习循环机制

```text
执行 → 采集 → 评估 → 更新 → 执行
  │       
---
cute(ctx, rpc, req)
}
```

### 7.12 架构总览

```text
┌─────────────────────────────────────────────────────────┐
│  用户层                                                  │
│  chat (串行 REPL)  /  serve (并发 HTTP API)              │
└───────────────┬─────────────────────────────────────────┘
                │
┌───────────────────▼─────────────────────────────────────────┐
│  TaskExecution                     
---
行为**：

| 模式 | 行为 | 说明 |
|------|------|------|
| **chat** | TaskExecution(Interactive+OneShot) → 串行 turn → 结果输出 | 天然串行，不硬塞并发 |
| **serve** | 并发创建多个 TaskExecution → 共享 Brain Pool → LeaseManager 控制并发 | HTTP 层并发，Lease 层隔离 |
| **chat + /background** | TaskExecution(Watch+PushNotify) → 后台独立运行 → EventBus 通知 | 前台交互 + 后台监控 |
| **Data 行情采集** | TaskExecution(Daemon+Always) → daemon scope lease → 持续运行 | 崩溃自动重启 |
---


```text
Step 1: BrainPool 接口 ─┐
Step 2: EventBus 接口 ──┤─→ Step 3: 全局 Pool 注入 serve ─→ Step 4: Orchestrator 瘦身
                        └─→ Step 5: EventBus SSE 端点
Step 6: TaskExecution 模型 ─→ Step 7: HTTP API 统一
Step 8: LeaseManager ─→ Step 10: Dispatch Policy (BatchPlanner)
Step 9: Dashboard API + SPA
Step 11-13: Context Engine / MCP Runtime / 语义审批（Phase B 入口）
```

**Phase A 拆为 7 个可独立合并的 PR**：

| PR 
---
header）→ Phase B 完成后正式标记 deprecated → Phase D 后（≥2 个主版本、≥60 天公示）可删除

### 13.14 Brain Capability 标签与匹配算法

> 详见 → [35-BrainCapability标签与匹配算法](./35-BrainCapability标签与匹配算法.md)

- **标签体系**：四类标签（function / domain / resource / mode），每类均有标准枚举值；分 primary / secondary 两级
- **匹配算法三阶段**：(1) 硬匹配——required capabilities 全部满足方可入围；(2) 软匹配——preferred 按满足比率加权；(3) 负载感知——当前 lease 占用率越低得分越高
- **评分公式**：`Score = 100 + soft_matc
---
-4]  5 脑 L0 BrainLearner
  [E-5]  L1-L3 接入执行路径
  [E-6]  Context Engine LLM 摘要
  [E-7]  Context SharedMessages 持久化
```

---

## 15. 上层产品接入边界

> **v8 新增** · 2026-04-18
> 
> 本节定义 Brain v3 作为运行时底座被上层产品（如 EasyMVP、IDE 插件、CI/CD 系统等）接入时的职责边界。

### 15.1 核心原则

Brain v3 是 **runtime source of truth**，上层产品是 **domain/product source of truth**。两者通过 `run_id` / `execution_id` 做关联，不共库、不共 UI、不混合生命周期。

```
┌────────────────
---
redMessages 持久化
```

---

## 15. 上层产品接入边界

> **v8 新增** · 2026-04-18
> 
> 本节定义 Brain v3 作为运行时底座被上层产品（如 EasyMVP、IDE 插件、CI/CD 系统等）接入时的职责边界。

### 15.1 核心原则

Brain v3 是 **runtime source of truth**，上层产品是 **domain/product source of truth**。两者通过 `run_id` / `execution_id` 做关联，不共库、不共 UI、不混合生命周期。

```
┌────────────────────────────────────────────┐
│  上层产品（EasyMVP / IDE / CI）              │
│  ├─ 自有数据库（project/plan
---
核心原则

Brain v3 是 **runtime source of truth**，上层产品是 **domain/product source of truth**。两者通过 `run_id` / `execution_id` 做关联，不共库、不共 UI、不混合生命周期。

```
┌────────────────────────────────────────────┐
│  上层产品（EasyMVP / IDE / CI）              │
│  ├─ 自有数据库（project/plan/acceptance）    │
│  ├─ 自有前端（工作台/审计/验收）              │
│  ├─ 领域投影层                               │
│  │   └─ 消费 runtime 事件 → 投影为领域对象    │
│  └─ 关联键：
---
shot |
| **领域索引** | replay index、evidence linkage、project projection |

### 15.4 四条接入铁律

**1. 独立数据库，按引用关联**

上层产品不应把业务表建在 `brain.db` 中。Brain v3 的库是 runtime persistence（关注 run/execution/event），上层产品的库是 domain persistence（关注 project/plan/acceptance）。两者迁移频率、回滚策略、兼容约束完全不同。

```
brain.db          ←  brain-v3 独占
product.db        ←  上层产品独占
关联方式：run_id / execution_id 外键引用
```

**2. 不自建底层源，但必须做领域投影**

上层产品**不应**
---
：
- runtime 状态机（用 Brain v3 的 TaskExecution）
- 通用事件总线（用 Brain v3 的 EventBus）
- 通用任务调度器（用 Brain v3 的 TaskScheduler/Orchestrator）
- 底层 run persistence 协议

上层产品**必须**自建：
- run binding（哪个 run 属于哪个项目/任务）
- domain event projection（runtime 事件 → 项目进度）
- workspace aggregate（聚合视图）
- evidence linkage（证据与 run 的关联）

**3. 后端消费 runtime 事件，前端只接产品聚合接口**

上层产品前端不应直接连 Brain v3 的 Dashboard WebSocket。正确做法：

```
Brain v3 Even