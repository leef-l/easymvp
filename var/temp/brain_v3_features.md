ine.Compress 签名：实际为 `(ctx, messages, budget int)` 多了 budget 参数
> - §7.6 BrainLearner 无 `Adapt()` 方法；LearningEngine 是 struct 非 interface，方法名：RankBrainsForTask→RankBrains, RecordExecutionChain→RecordSequence, S
---
Status/AutoStart/Shutdown），AutoStart 替代了 WarmUp/Register
> - §13.5 Dashboard 实际 8 个端点（非 11），新增 auth/learning/ws，缺少 brains/:kind 等细粒度端点
> - §15 节号重复：第二个 §15 应为 §16
> **上位规格**：[02-BrainKernel设计.md](./02-BrainKe
---
---------|----------|
| OpenClaw | 单体 + Skills 插件 | 无多脑协作，skills 是工具集合 |
| Hermes Agent | 单 Agent + Learning Loop | 自我改进，无并行脑 |
| Claude Code | 单 Agent + MCP 工具 | 工具扩展，无脑级别委托 |
| **Brain** | **中央大脑 + N 专精大脑**
---

4. 完成

后续字段按需扩展，不阻塞启动。

### 7.6 四层自适应学习体系（Phase B-5）

这是多脑系统和单 agent 系统的**根本差异化**。

Hermes Agent 的 Learning Loop 只能做"单 agent 自我改进"——一个脑从自己的历史里学。我们的多脑架构天然支持四个维度的学习，是单 agent 系统做不到的：

```text
┌─────────────────
---
rtMetrics(ctx context.Context) (BrainMetrics, error)
}

// === L1-L3: Central 侧接口（中央大脑实现） ===

type LearningEngine interface {
    // L1: 协作级学习——消费各 brain 上报的 BrainMetrics，维护能力画像
    IngestBrainMetrics(ctx co
---
or
    GetUserProfile(ctx context.Context, userID string) (UserProfile, error)
}
```

#### 与 Hermes Learning Loop 的对比

| 维度 | Hermes | Brain v3 |
|------|--------|----------|
| 学习范围 | 单 agent 自我改进 | 四层：Brain/
---
量化、数据、代码、浏览器） |
| **可观测** | 单 agent 日志 | 多脑 Dashboard + 事件流 + 跨脑追踪 |
| **学习** | 单 agent 自我改进（Hermes Learning Loop） | 四层学习：Brain 级/协作级/策略级/用户级，越用越好 |
| **商业化** | 插件/技能市场 | 专精大脑生态（免费版/Pro 版/企业版） |

一句话：**别人做的是"
---
，alpha 按任务频率动态（0.1/0.2/0.3）
  - 冷启动：20% 探索预算 + Manifest capabilities 预设评分
  - 持久化：`~/.brain/central/learning/l1/`，30 天滚动窗口淘汰

- **L2 策略级**：
  - ExecutionChain DAG 表示，链路指纹用于快速模式匹配
  - 因果推断四步：样本量检查 → CI 重叠 → Si
---
根因：

| 当前方案 | 问题 |
|---------|------|
| JSON 全量重写（RunStore/PlanStore） | 每次写 O(N)，run 多了严重卡顿 |
| 纯内存（LearningEngine/Context/AuditLog） | 进程重启全丢，L1-L3 学习无效 |
| 无搜索能力 | 无法语义检索历史上下文 |

竞品对标（2026-04）：

| 产品 | 持久化核心
---
ersistence/driver_sqlite.go`） | P0 | 无 |
| E-2 | RunStore/PlanStore 迁移到 SQLite | P0 | E-1 |
| E-3 | LearningEngine 持久化（Save/Load brain_scores） | P0 | E-1 |
| E-4 | 5 脑 L0 BrainLearner 实现 | P0 | E-3 |
| E-5 | 
---
] upgrade/rollback
  [E-11] Package 签名

Sprint E-2（依赖 E-1）：
  [E-2]  RunStore/PlanStore 迁移
  [E-3]  LearningEngine 持久化
  [E-9]  AuditLog 持久化
  [E-12] Dashboard WebSocket

Sprint E-3（依赖 E-2/E-3）：
  [E-4]  5 脑 
---
| 工具扩展，无脑级别委托 |
| **Brain** | **中央大脑 + N 专精大脑** | **进程隔离、工具隔离、跨脑委托、跨脑授权** |

别人有的（Dashboard、Task 管理、Context Engine、MCP 生态），我们可以做。
别人没有的（多脑并行协作、专精大脑生态），是我们的差异化。

### 2.2 代码已经在走这条路

当前 v0.7.0 已经实现的多脑基础：

| 能力 | 实现位置 
---
检查、并发控制、brain 选择 |
| B-3 | 语义审批分级 | SpecialistToolCallAuthorizer 已有 | 5 级审批类，与 Lease 权限集成 |
| B-4 | Context Engine | 三层 Prompt Cache 框架已有 | 可插拔上下文装配层：压缩、摘要、跨 turn 记忆、brain 间上下文共享 |
| B-5 | 自适应工具策略 | tool_profiles/a
---
 L4 | `external-network` | 外部网络请求 | 需确认 + 审计 + 可选人工审批 |

每个工具在注册时声明自己的审批等级，审批策略基于等级而不是工具名。

### 7.4 Context Engine（Phase B-2）

在现有三层 Prompt Cache 基础上，增加统一的上下文装配层：

```text
┌─────────────────────────────────────┐
│ 
---
2）

在现有三层 Prompt Cache 基础上，增加统一的上下文装配层：

```text
┌─────────────────────────────────────┐
│          Context Engine              │
├──────────┬──────────┬───────────────┤
│ Compressor│ Memory  │ Brain-Shared  │
│ 上下
---
TaskExecution 统一模型 + Capability Lease + Dispatch Policy | 多任务已有基础 |
| 静态工具白名单 | 语义审批分级 | 行业趋势 |
| 无 Context Engine | 可插拔上下文装配层 | 直接影响 agent 质量和成本 |
| 无系统级学习能力 | 四层自适应学习体系（L0-L3） | 多脑协作的核心差异化，单 agent 无法复制 |
| Orches
---
 CAS 不丢数据）
- **现有 ringbuf 迁移**：三步走——引入 Frame 泛型单元 → 注册进 PipeRegistry → 清理 quant 直接 import

### 13.7 Context Engine 详细设计

> **详细规格**：[35-Context-Engine详细设计.md](./35-Context-Engine详细设计.md)

**关键决策**：

- **Assemble 流水
---
 8: LeaseManager ─→ Step 10: Dispatch Policy (BatchPlanner)
Step 9: Dashboard API + SPA
Step 11-13: Context Engine / MCP Runtime / 语义审批（Phase B 入口）
```

**Phase A 拆为 7 个可独立合并的 PR**：

| PR | 范围 | 工时 | 前置 |
|----|---
---
1 |
| E-4 | 5 脑 L0 BrainLearner 实现 | P0 | E-3 |
| E-5 | L1-L3 接入 delegate 执行路径 | P0 | E-3 |
| E-6 | Context Engine LLM 摘要路径 | P1 | E-1 |
| E-7 | Context SharedMessages 持久化 | P1 | E-1 |
| E-8 | Streaming Edge 打通到 Wo
---
oard WebSocket

Sprint E-3（依赖 E-2/E-3）：
  [E-4]  5 脑 L0 BrainLearner
  [E-5]  L1-L3 接入执行路径
  [E-6]  Context Engine LLM 摘要
  [E-7]  Context SharedMessages 持久化
```

---

## 15. 上层产品接入边界

> **v8 新增** · 2026-04-18
> 
>
---
ashboard | Quant WebUI 已实现 | 全脑 Dashboard：brain 状态、任务、事件流、provider 认证 |
| A-6 | 事件流 | OTLP + runtimeaudit 分散存在 | 统一 EventBus 聚合层 |

### Phase B：调度层——Parallel Execution + Scheduler（v3.1）

**目标**：再把并发执行做快

|
---
个 major 版本 | 必须 |
| 放宽性修改 | 1 个 major 版本 | 必须 |
| 澄清性修改（不改语义） | 无 | 可选 |

### 10.5 违规检测

5 个审计事件类型：`audit.delegate_rejected`（规则 #1）、`audit.protocol_deviation`（规则 #2）、`audit.invalid_delegate_target`（规则 #3）、
---
ajor 版本 | 必须 |
| 澄清性修改（不改语义） | 无 | 可选 |

### 10.5 违规检测

5 个审计事件类型：`audit.delegate_rejected`（规则 #1）、`audit.protocol_deviation`（规则 #2）、`audit.invalid_delegate_target`（规则 #3）、`audit.policy_violation`（规则 #4）、`
---
 可选 |

### 10.5 违规检测

5 个审计事件类型：`audit.delegate_rejected`（规则 #1）、`audit.protocol_deviation`（规则 #2）、`audit.invalid_delegate_target`（规则 #3）、`audit.policy_violation`（规则 #4）、`audit.manifest_breaking`（规则 #6）。全部
---
delegate_rejected`（规则 #1）、`audit.protocol_deviation`（规则 #2）、`audit.invalid_delegate_target`（规则 #3）、`audit.policy_violation`（规则 #4）、`audit.manifest_breaking`（规则 #6）。全部通过 EventBus 广播，Dashboard 实时展示违规率。

---

---
.protocol_deviation`（规则 #2）、`audit.invalid_delegate_target`（规则 #3）、`audit.policy_violation`（规则 #4）、`audit.manifest_breaking`（规则 #6）。全部通过 EventBus 广播，Dashboard 实时展示违规率。

---

## 11. 与当前仓库的衔接

### 11.1 已有能力映
---
 扩展 `cmd_serve.go` + 复用 `brains/quant/webui/` 模式 | 中等 |
| EventBus 聚合层 | 新增 `sdk/events/`，聚合 runtimeaudit + protocol 事件 | 中等 |

---

## 12. 最终结论

v3 的正式口径：

> **Multi-brain native, runtime-observable, mani
---
tDTO、ProviderDTO、LeaseDTO
- **EventBus 聚合层**：内存环形缓冲（10000 条）+ fan-out 到 WebSocket 订阅者，三个来源汇聚（runtimeaudit + Orchestrator + 持久化存储）
- **向后兼容**：`/v1/runs` 完整保留，响应加 `X-Deprecated-Use` header
- **前端嵌入**：`cmd/br
---
_levels` 豁免
- **接口**：`SemanticApprover` + `ApproverMiddleware` 中间件链（StaticMatrix → ManifestPolicy → Audit → ModeAware）
- **迁移**：三阶段从 SpecialistToolCallAuthorizer 过渡（适配层 → 语义升级 → 静态矩阵下线）

### 13.10 Manifest
---
------|------|
| JSON 全量重写（RunStore/PlanStore） | 每次写 O(N)，run 多了严重卡顿 |
| 纯内存（LearningEngine/Context/AuditLog） | 进程重启全丢，L1-L3 学习无效 |
| 无搜索能力 | 无法语义检索历史上下文 |

竞品对标（2026-04）：

| 产品 | 持久化核心 | 记忆层 | 学习机制 |
|---