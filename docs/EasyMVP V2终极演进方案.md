# EasyMVP V2 终极演进方案

> 已归档：本文是演进路线与目标态设计，保留作历史背景参考。当前实现请优先以 `docs/README.md` 中“当前有效文档”为准。

> 目标：从"代码完成"到"世界第一的 AI 项目管理引擎"
> 基于 2026-04-06 全面审计结果，覆盖架构、安全、性能、产品化四大维度

---

## 一、当前状态（V1 基线）

| 维度 | 数据 |
|------|------|
| Go 后端代码 | 35,434 行（14 个模块） |
| Vue 前端代码 | 10,989 行（11 个 workflow 页面） |
| 数据库 | 55 张表，47 个配置项 |
| 执行器 | 6 种（aider/openhands/claude_code/codex_cli/gemini_cli/chat） |
| 自治层级 | L1~L7 代码完成，零实战验证 |
| 用户量 | 0 |

---

## 二、审计发现的 30 个问题（按致命度排序）

### S 级（不修必死）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 1 | JWT/数据库密钥明文写死在 config.yaml 和 docker-compose | 全项目 | 源码泄露 = 完全沦陷 |
| 2 | CORS 设置 `Access-Control-Allow-Origin: *` | MVP 服务 | 任意网站可调 API |
| 3 | 无数据库备份策略 | 运维 | 数据丢失不可恢复 |
| 4 | 七层自治零实战验证 | 全项目 | 架构是否成立未知 |

### A 级（严重影响可用性）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 5 | Legacy 引擎 (engine/) 和 V2 引擎 (workflow/) 双轨并存 | 后端 | 重复逻辑、维护成本翻倍 |
| 6 | Controller 层直接调 g.DB()（67 处中 6 处） | workflow.go | 违反分层、不可测试 |
| 7 | 执行器直接在 Execute() 中查 project 表 | executor/*.go | 每个执行器重复 project 查询 |
| 8 | DecisionCenter 8 层依赖链无接口隔离 | autonomy/ | 改一处牵全链 |
| 9 | 前端状态管理混乱（Store 定义了但没用） | views/mvp/ | 11 个页面各自轮询 |
| 10 | 前端错误处理不统一（静默吞错 / console.error / 无反馈） | views/mvp/ | 用户无感知 |
| 11 | Go 服务无健康检查端点 | 全后端 | Docker/K8s 无法探活 |
| 12 | 无 Prometheus 指标暴露 | 全后端 | 无可观测性 |
| 13 | CI/CD 无单元测试卡点 | .github/workflows | 坏代码直接上线 |
| 14 | Token 注销无黑名单（注销后 token 仍有效 24h） | auth 模块 | 安全漏洞 |

### B 级（影响工程质量）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 15 | 5 个执行器代码高度重复（每个~100 行模板代码） | executor/*.go | 新增执行器要抄 100 行 |
| 16 | 自治链路 4 处隐蔽断点（Sensor 失败仅 log warning） | autonomy/ | 故障不可感知 |
| 17 | goroutine 泄漏风险（Background context + 无 timeout） | executor/scheduler | 生产稳定性 |
| 18 | SSE 无断线重连、无心跳 | chat/index.vue | 长连接不稳 |
| 19 | 配置双 key 命名不一致（snake_case vs camelCase 混用） | engine/config.go | 人为出错率高 |
| 20 | 前端 workflow API 单文件 997 行 | api/mvp/workflow | 可维护性差 |
| 21 | 11 个页面重复 projectID 获取逻辑 | views/mvp/workflow | 命名不统一 |
| 22 | 非结构化日志（字符串模板拼接） | 全后端 | 日志检索困难 |
| 23 | Dev 环境缺少 Redis | docker-compose.yml | 开发环境不完整 |
| 24 | 无 API 层限流 | 全后端 | 无 QoS 保障 |
| 25 | Repo 层无基类（18 个 Repo 各自硬编码 g.DB()） | workflow/repo/ | 换 DB 需全改 |
| 26 | 无链路追踪（TraceID/RequestID） | 全后端 | 分布式调试困难 |
| 27 | 前端表格列定义/状态颜色映射重复 | views/mvp/workflow | 改颜色要改 7 处 |
| 28 | 路由参数依赖风险（直接访问 URL 无 projectId 时白屏） | 前端路由 | 用户体验差 |
| 29 | 资源锁泄漏风险（任务 panic 时锁不释放） | scheduler | 任务阻塞 |
| 30 | Cleanup goroutine 失败无重试 | executor/*.go | worktree 残留 |

---

## 三、V2 架构蓝图

### 3.1 目标架构（一步到位）

```
                        ┌─────────────────────────────────────┐
                        │          API Gateway (Nginx)         │
                        │  限流 / CORS / JWT 校验 / 路由       │
                        └─────┬──────────┬──────────┬─────────┘
                              │          │          │
                    ┌─────────┴──┐ ���─────┴──────┐ ┌┴──────────┐
                    │  system    │ │    ai      │ │    mvp     │
                    │ (用户权限) │ │ (AI 配置)  │ │ (核心引擎) │
                    └────────────┘ └──────────��─┘ └─────┬──────┘
                                                        │
                         ┌──────────────────────────────┤
                         │                              │
                  ┌──────┴──────┐              ┌────────┴────────┐
                  │  调度中心    │              │   自治决策中心    │
                  │ Scheduler   │              │ DecisionCenter  │
                  │ (任务分发)   │              │ (L3~L7 全链路)  │
                  └──────┬──────┘              └─────────────────┘
                         │
              ┌──────────┼──────────┐
              │ Redis 任务队列       │
              │ (BullMQ 风格)        │
              └──────────┬──────────┘
                         │
          ┌──���───┬───────┼───────┬──────┐
          │      │       │       │      │
       Worker Worker  Worker  Worker  Worker
       aider  claude  codex  gemini openhands
       
              每个 Worker 独立进程
              崩溃不影响调度器
```

### 3.2 核心改造点

#### 改造一：合并双引擎（P0）

```
当前：
  engine/          → Legacy 引擎（scheduler/executor/state_machine）
  workflow/        → V2 引擎（orchestrator/stage/domain）
  compat/          → 薄兼容层

目标：
  workflow/        → 唯一引擎
  engine/          → 删除（迁移剩余有用逻辑后）
  compat/          → 删除

步骤：
  1. 列出 engine/ 中仍被调用的函数
  2. 逐个迁移到 workflow/ 对应模块
  3. 切换所有入口到 V2
  4. 删除 engine/ 和 compat/
```

#### 改造二：执行器 Worker 池化（P0）

```go
// 当前：mvp 进程内直接 fork 子进程
func (e *ClaudeCodeExecutor) Execute(ctx, req) *Result {
    cmd := exec.CommandContext(cmdCtx, "bash", "-c", cmdStr)  // 阻塞 mvp 进程
    cmd.Run()
}

// 目标：Redis 队列 + Worker 独立进程
// mvp 进程只负责入队
func (s *Scheduler) Dispatch(ctx, task) error {
    return s.queue.Enqueue(ctx, &ExecutionJob{
        TaskID:    task.ID,
        Engine:    task.ExecutionMode,
        WorkDir:   task.WorkDir,
        Timeout:   1800,
    })
}

// Worker 进程消费执行（独立二进制）
func (w *Worker) Run() {
    for job := range w.queue.Dequeue() {
        executor := w.registry.Get(job.Engine)
        result := executor.Execute(job)
        w.queue.Ack(job.ID, result)  // 回写结果
    }
}
```

好处：
- 执行器崩溃不影响调度器
- 可按引擎类型独立扩缩容
- Worker 天然支持多机部署

#### 改造三：统一 Repo 基类（P1）

```go
// 当前：18 个 Repo 各自硬编码
type SituationSnapshotRepo struct{}
func (r *SituationSnapshotRepo) Create(ctx, data) {
    g.DB().Model("mvp_situation_snapshot").Ctx(ctx).Insert(data)  // 散落各处
}

// 目标：基类 + 泛型
type BaseRepo[T any] struct {
    table string
}

func (r *BaseRepo[T]) Create(ctx context.Context, data *T) (int64, error) {
    result, err := g.DB().Model(r.table).Ctx(ctx).
        Data(data).
        InsertAndGetId()
    return result, err
}

func (r *BaseRepo[T]) FindByID(ctx context.Context, id int64) (*T, error) { ... }
func (r *BaseRepo[T]) List(ctx context.Context, opts ListOptions) ([]*T, int, error) { ... }
func (r *BaseRepo[T]) SoftDelete(ctx context.Context, id int64) error { ... }

// 使用：
type SituationSnapshotRepo struct {
    BaseRepo[entity.SituationSnapshot]
}
var snapshotRepo = &SituationSnapshotRepo{BaseRepo: BaseRepo[entity.SituationSnapshot]{table: "mvp_situation_snapshot"}}
```

#### 改造四：执行器提取 BaseExecutor（P1）

```go
// 当前：5 个执行器各 100 行重复代码
// 目标：BaseExecutor 封装公共逻辑

type BaseExecutor struct {
    engineCode string
    wsMgr      workspace.Manager
    envKey     string  // ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY
}

func (b *BaseExecutor) Execute(ctx context.Context, req *Request) *Result {
    // 1. 加载引擎配置（公共）
    engineCfg := b.loadConfig(ctx)
    
    // 2. 确定工作目录（公共）
    workDir := b.resolveWorkDir(ctx, req)
    
    // 3. 构建指令（公共）
    instruction := b.buildInstruction(req)
    
    // 4. 构建命令（子类实现）
    cmdStr := b.buildCommand(engineCfg, instruction, workDir, req)
    
    // 5. 执行 + 超时（公共）
    output, err := b.runCommand(ctx, cmdStr, workDir, req, engineCfg)
    
    // 6. Workspace finalize（公共）
    b.finalizeWorkspace(ctx, req, err)
    
    return b.buildResult(output, err)
}

// 子类只需实现：
type ClaudeCodeExecutor struct{ BaseExecutor }
func (e *ClaudeCodeExecutor) buildCommand(cfg, instruction, workDir, req) string {
    return fmt.Sprintf("claude -p --output-format json '%s'", instruction)
}
```

#### 改造五：DecisionCenter 接口隔离（P1）

```go
// 当前：8 层直接依赖，SetPhaseA/B/DDeps 手工注入
// 目标：依赖注入 + 接口隔离

type DecisionCenterConfig struct {
    PolicyEngine   PolicyMatcher       // L3 接口
    RiskGate       RiskChecker         // L3 接口
    Planner        StrategyPlanner     // L5 接口
    Dispatcher     ActionExecutor      // L2 接���
    Observer       ObservationRecorder // L7 接口（可选）
    Learner        LearningEngine      // L7 接口（可选）
    Sensor         SituationSensor     // L6 接口（可选）
    Objective      ObjectiveChecker    // L6 接口（可选）
}

func NewDecisionCenter(cfg DecisionCenterConfig) *DecisionCenter {
    // 编译期保证必要依赖不为 nil
    if cfg.PolicyEngine == nil || cfg.RiskGate == nil || cfg.Dispatcher == nil {
        panic("required dependencies missing")
    }
    return &DecisionCenter{...}
}
```

---

## 四、安全加固方案

### 4.1 密钥管理（S 级，立即修复）

```yaml
# 当前（危险）：
jwt:
  secret: "easymvp-secret-key-2026"  # 明文写死
database:
  link: "mysql:easymvp:JKcHFJYXnkrB6BXE@tcp(127.0.0.1:3306)/easymvp"

# 目标：环境变量注入
jwt:
  secret: "${JWT_SECRET}"
database:
  link: "mysql:${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"
```

部署侧：
- 开发：`.env` 文件（git ignore）
- 生产：GitHub Secrets → CI/CD 注入
- 未来：HashiCorp Vault / 阿里云 KMS

### 4.2 CORS 白名单（S 级）

```go
// 当前：
r.Response.Header().Set("Access-Control-Allow-Origin", "*")

// 目标：
allowedOrigins := []string{
    "https://mvp.yourcompany.com",
    "http://localhost:5666",  // dev only
}
```

### 4.3 Token 黑名单（A 级）

```go
// 注销时：
redis.Set(ctx, "token:blacklist:"+tokenID, "1", remainingTTL)

// 鉴权时：
if redis.Exists(ctx, "token:blacklist:"+tokenID) {
    return ErrTokenRevoked
}
```

### 4.4 API 限流（B 级）

```
Nginx 层：
  limit_req_zone $binary_remote_addr zone=api:10m rate=100r/s;
  
Go 中间件层（精细控制）：
  /mvp/chat/send     → 10 req/min/user（AI 调用贵）
  /mvp/workflow/*     → 60 req/min/user
  /system/auth/login  → 5 req/min/ip（防爆破）
```

---

## 五、可观测性方案

### 5.1 健康检查端点

```go
// 每个服务注册 /health
r.GET("/health", func(r *ghttp.Request) {
    checks := map[string]string{
        "db":    checkDB(),
        "redis": checkRedis(),
    }
    r.Response.WriteJson(g.Map{"status": "ok", "checks": checks})
})
```

### 5.2 Prometheus 指标

```go
// 关键业务指标
mvp_task_total{status="completed|failed|running"}         // 任务数
mvp_task_duration_seconds{engine="aider|claude_code|..."}  // 执行耗时
mvp_decision_total{level="A|B|C", outcome="approve|reject"} // 决策统计
mvp_worker_active{engine="..."}                            // Worker 活跃数
mvp_sse_connections                                        // SSE 连接数
```

### 5.3 结构化日志

```go
// 当前：
g.Log().Infof(ctx, "[ClaudeCodeExecutor] 启动: task=%d workDir=%s", req.TaskID, workDir)

// 目标：
g.Log().Info(ctx, "executor.started",
    "engine", "claude_code",
    "task_id", req.TaskID,
    "work_dir", workDir,
    "trace_id", gtrace.GetTraceID(ctx),
)
```

### 5.4 链路追踪

```
请求入口生成 TraceID → 写入 Context → 所有日志自动携带
→ 调度器传递给 Worker → Worker 日志同样携带
→ 一个 TraceID 串联完整链路
```

---

## 六、前端升级方案

### 6.1 Composable 抽取

```typescript
// composables/useProjectId.ts
export function useProjectId() {
  const route = useRoute();
  const projectId = computed(() => (route.query.projectId as string) || '');
  const hasProject = computed(() => !!projectId.value);
  return { projectId, hasProject };
}

// composables/usePoll.ts
export function usePoll(fn: () => Promise<void>, interval = 5000) {
  const loading = ref(false);
  let timer: ReturnType<typeof setInterval>;
  
  async function execute() {
    loading.value = true;
    try { await fn(); } finally { loading.value = false; }
  }
  
  onMounted(() => { execute(); timer = setInterval(execute, interval); });
  onUnmounted(() => clearInterval(timer));
  
  return { loading, refresh: execute };
}

// composables/useWorkflowSSE.ts（替代轮询）
export function useWorkflowSSE(projectId: Ref<string>) {
  // EventSource + 自动重连 + 心跳检测
}
```

### 6.2 API 分拆

```
api/mvp/workflow/
├── index.ts          → 只导出，不定义（barrel file）
├── project.ts        → 项目操作（create/pause/resume/status）
├── stage.ts          → 阶段操作（review/execution/rework��
├── accept.ts         → 验收操作
├── autonomy.ts       → 自治操作（decisions/checkpoints/policies）
├── meta-cognition.ts → 元认知操作
├── feishu.ts         → 飞书集成
├── situation.ts      → 态势感知
└── types.ts          → 所有类型定义
```

### 6.3 全局错误处理

```typescript
// plugins/error-handler.ts
app.config.errorHandler = (err, instance, info) => {
  notification.error({ message: '操作失败', description: err.message });
  reportToSentry(err);  // 可选：错误上报
};

// API 拦截器
requestClient.interceptors.response.use(null, (error) => {
  if (error.response?.status === 401) router.push('/login');
  if (error.response?.status === 403) notification.warn({ message: '权限不足' });
  if (error.response?.status >= 500) notification.error({ message: '服务异常' });
  return Promise.reject(error);
});
```

---

## 七、执行路线图

### Phase 1：安全加固 + 灰度验证（第 1~2 周）

| 任务 | 优先级 | 预���耗时 |
|------|--------|----------|
| 密钥移到环境变量 | S | 2h |
| CORS 白名单 | S | 1h |
| 数据库备份脚本（日备份 + 7 天保留） | S | 2h |
| 开启灰度：enabled=1 + audit_only=1 + meta_cognition=1 | S | 10min |
| 用 EasyMVP 跑第一个真实项目（自己的鸿蒙 App） | S | 持续 |
| 健康检查端点 /health | A | 2h |
| Dev 环境补 Redis | A | 1h |
| Token 黑名单 | A | 4h |

**Phase 1 验收标准**：系统安全可用，第一个真实项目开始跑，元认知开始积累数据。

### Phase 2：架构重构（第 3~6 周）

| 任务 | 优先级 | 预计耗时 |
|------|--------|----------|
| 合并双引擎（删除 engine/，统一 workflow/） | P0 | 1 周 |
| 执行器 BaseExecutor 提取 | P1 | 2 天 |
| 执行器 Worker 池化（Redis 队列） | P0 | 1 周 |
| 统一 Repo 基类 | P1 | 3 天 |
| DecisionCenter 接口隔离 | P1 | 2 天 |
| Controller 层 g.DB() 清理 | P1 | 2 天 |
| 配置 key 统一为 snake_case | P2 | 1 天 |
| goroutine 超时保护 + 泄漏检测 | P2 | 2 天 |

**Phase 2 验收标准**：单引擎、Worker 隔离、Repo 统一。go vet + go test 全绿。

### Phase 3：前端升级 + 可观测性（第 7~8 周）

| 任务 | 优先级 | 预计耗时 |
|------|--------|----------|
| Composable 抽取（useProjectId/usePoll/useSSE） | A | 2 天 |
| API 分拆（997 行 → 8 个文件） | A | 1 天 |
| 全局错误处理 + 用户反馈 | A | 1 天 |
| Prometheus 指标暴露 | A | 2 天 |
| 结构化日志 + TraceID | B | 2 天 |
| API 限流中间件 | B | 1 天 |
| CI/CD 加单元测试卡点 | B | 1 天 |

**Phase 3 验收标准**：前端代码复用率 >60%，Grafana 可看到业务指标。

### Phase 4：产品化扩展（第 9~16 周）

| 任务 | 优先级 | 预计耗时 |
|------|--------|----------|
| SWE-bench 跑分（证明实力） | A | 2 周 |
| 鸿蒙 App（方案 C 混合开发） | A | 3 周 |
| Windows 桌面 Agent（Computer Use） | B | 3 周 |
| 开源核心调度引擎 | B | 1 周 |
| 企业版多租户 | B | 2 周 |

---

## 八、技术选型决策

| 决策点 | 选择 | 原因 |
|--------|------|------|
| 微服务 vs 单体 | **MonoRepo 多应用 + Worker 池**（不拆微服务） | 一个人维护，拆了是负担 |
| 消息队列 | **Redis Stream**（不用 Kafka/RabbitMQ） | Go + Redis 原生支持，运维零成本 |
| 监控 | **Prometheus + Grafana** | 行业标准，GoFrame 有现成中间件 |
| 链路追踪 | **GoFrame 内置 gtrace**（不引入 Jaeger） | 够用，不增加运维负担 |
| 前端状态 | **Composable + Pinia Store**（不引入新框架） | 现有体系内优化 |
| 桌面端 | **Tauri**（不用 Electron） | Rust 内核更小更快，内存占用 1/10 |
| 鸿蒙端 | **ArkTS 原生 + WebView 混合** | 关键页原生，其余复用 Web |
| 数据�� | **继续 MySQL**（不换 PostgreSQL） | 够用，换了迁移成��高 |

---

## 九、成功指标

### 技术指标

| 指标 | V1 现状 | V2 目标 |
|------|---------|---------|
| 任务成功率 | 未知 | > 85% |
| 首次通过率（无需返工） | 未知 | > 70% |
| 平均任务耗时 | 未知 | < 5 分钟 |
| 执行器切换时间 | 不支持 | < 3 秒 |
| API P99 延迟 | 未知 | < 200ms |
| Worker 崩溃恢复时间 | N/A | < 10 秒 |
| 自治决策准确率 | 未知 | > 80% |
| 元认知 EMA 样本量 | 0 | > 1000 |

### 产品指标

| 指标 | V1 现状 | V2 目标 |
|------|---------|---------|
| 完成的真实项目数 | 0 | > 10 |
| SWE-bench ��名 | 未参与 | 前 5 |
| GitHub Stars | 0 | > 1000 |
| 企业试用数 | 0 | > 3 |
| 鸿蒙 App 用户 | 0 | > 100 |

---

## 十、一句话总结

**V1 证明了我们能设计出七层自治架构；V2 要证明这个架构能在真实世界中工作、而且比所有竞品都好。**

从 V1 到 V2 的核心转变不是加功能，而是：

1. **安全**：从"能���"到"不怕攻击"
2. **隔离**：从"一个进程扛所有"到"Worker 池独立执行"
3. **验证**：从"代码完成"到"数据证明"
4. **产品**：从"给自己看"到"给用户用"
