## 15. 上层产品接入边界

> **v8 新增** · 2026-04-18
> 
> 本节定义 Brain v3 作为运行时底座被上层产品（如 EasyMVP、IDE 插件、CI/CD 系统等）接入时的职责边界。

### 15.1 核心原则

Brain v3 是 **runtime source of truth**，上层产品是 **domain/product source of truth**。两者通过 `run_id` / `execution_id` 做关联，不共库、不共 UI、不混合生命周期。

```
┌────────────────────────────────────────────┐
│  上层产品（EasyMVP / IDE / CI）              │
│  ├─ 自有数据库（project/plan/acceptance）    │
│  ├─ 自有前端（工作台/审计/验收）              │
│  ├─ 领域投影层                               │
│  │   └─ 消费 runtime 事件 → 投影为领域对象    │
│  └─ 关联键：run_id / execution_id            │
│                                              │
│       引用，不复制 ↕                          │
│                                              │
│  Brain v3（运行时底座）                       │
│  ├─ brain.db（run/execution/event/artifact）  │
│  ├─ brain serve API                          │
│  ├─ EventBus + SSE + WebSocket               │
│  └─ replay / audit 原始产物                   │
└────────────────────────────────────────────┘
```

### 15.2 Brain v3 负责

| 领域 | 具体内容 |
|------|---------|
| **TaskExecution 生命周期** | 12 状态机、Mode×Lifecycle×Restart |
| **runtime 事件源** | EventBus、SSE、WebSocket 推送 |
| **execution/runs API** | `/v1/runs`、`/v1/executions` |
| **原始产物** | logs、replay、audit trail、artifact |
| **多脑调度** | Orchestrator、TaskScheduler、L0-L3 学习 |
| **运行时策略** | tool policy、sandbox、file policy、approval |
| **runtime observability** | Dashboard、brain/provider/lease 状态 |

### 15.3 上层产品负责

| 领域 | 具体内容 |
|------|---------|
| **领域对象** | project、plan、task、acceptance 等业务实体 |
| **工作流状态机** | 方案审核 → 编译 → 执行 → 验收的业务流程 |
| **领域投影** | 把 runtime 事件映射为项目进度、风险、待处理 |
| **UI 聚合** | 工作台、ActionInbox、WorkspaceSnapshot |
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

上层产品**不应**重新发明：
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
Brain v3 EventBus → 上层产品后端（消费+投影）→ 上层产品前端
```

Brain v3 Dashboard 是 runtime control plane，上层产品 UI 是 product workflow cockpit，关注点不同。

**4. 原始产物复用，领域索引自建**

- raw replay/log/audit artifact → 由 Brain v3 产出和存储
- 项目级 replay 索引、领域关联、decision trail → 由上层产品维护

上层产品不复制底层全量 runtime 数据，只做必要的索引和投影。

### 15.5 接入 API 契约

上层产品通过以下 API 与 Brain v3 交互：

| API | 用途 | 数据方向 |
|-----|------|---------|
| `POST /v1/runs` | 创建 run | 产品 → runtime |
| `GET /v1/runs/{id}` | 查询 run 状态 | 产品 ← runtime |
| `POST /v1/runs/{id}/cancel` | 取消 run | 产品 → runtime |
| `GET /v1/dashboard/events` | SSE 事件流 | 产品 ← runtime（实时） |
| `GET /v1/dashboard/ws` | WebSocket 事件流 | 产品 ← runtime（实时） |
| `GET /v1/dashboard/overview` | 全局状态 | 产品 ← runtime |

上层产品后端订阅事件流，将 runtime 事件投影为自己的领域对象后，再通过自己的 API 提供给前端。

### 15.6 数据生命周期

```
runtime 事件产生
  ↓
Brain v3 持久化到 brain.db（audit_logs / run_events）
  ↓
上层产品后端通过 API/SSE/WS 消费事件
  ↓
投影为领域对象写入 product.db（binding / projection / aggregate）
  ↓
上层产品前端从 product.db 聚合展示
```

