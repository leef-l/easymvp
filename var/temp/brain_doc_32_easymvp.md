 LearningEngine 持久化
  [E-9]  AuditLog 持久化
  [E-12] Dashboard WebSocket

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
> 本节定义 Brain v3 作为运行时底座被上层产品（如 EasyMVP、IDE 插件、CI/CD 系统等）接入时的职责边界。

### 15.1 核心原则

Brain v3 是 **runtime source of truth**，上层产品是 **domain/product source of truth**。两者通过 `run_id` / `execution_id` 做关联，不共库、不共 UI、不混合生命周期。

```
┌────────────────────────────────────────────┐
│  上层产品（EasyMVP / IDE / CI）              │
│  ├─ 自有数据库（project/plan/ac
---
v8 新增** · 2026-04-18
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
│       引用，不复制 ↕