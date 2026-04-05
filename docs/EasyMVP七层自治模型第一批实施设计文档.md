# EasyMVP 七层自治模型第一批实施设计文档

> 基线来源：[EasyMVP自治系统终极基线：七层自治模型.md](/www/wwwroot/project/easymvp/docs/EasyMVP自治系统终极基线：七层自治模型.md)
> 实施范围：Phase A
> 目标：先把 `L6 目标层 + L5 态势感知 + Operator 角色 + DecisionMeta` 做成可灰度、可观测、可回滚的地基层。

---

## 1. 文档定位

这不是新的理念稿，而是七层自治模型的第一批实施稿。

本批次只解决 4 件事：

1. 系统能形成统一 `Situation` 快照
2. 系统能在 `DecisionCenter` 前做目标层准入
3. 系统能有独立的 `Operator` 角色承担恢复分析
4. 系统能把 `DecisionMeta` 作为后续跨层决策统一元数据

本批次**不做**：

- Planner 六大策略函数
- 五契约包与契约匹配
- Learner / Observer / Assessor / Tuner
- 全面替换现有自治链路

一句话：

`先搭地基，不改主业务行为。`

---

## 2. 当前代码现状

按当前仓库代码实态：

### 2.1 已存在

1. `L3.5` 自治中台第一、二批已落地  
已有：
- `decision_center.go`
- `policy_engine.go`
- `risk_gate.go`
- `action_dispatcher.go`
- `decision_action / human_checkpoint` 等数据模型

2. `Phase A` SQL 草稿已存在  
见：
- [20260410_l4_phase_a.sql](/www/wwwroot/project/easymvp/admin-go/docker/mysql/upgrade/20260410_l4_phase_a.sql)

3. 项目分类配置化已存在  
`Operator` 预设可以按 `category_code` 正式落表，不再依赖旧中文分类。

### 2.2 尚未实现

1. `autonomy/situation.go`
2. `autonomy/sensor.go`
3. `autonomy/objective.go`
4. `autonomy/decision_meta.go`
5. `repo/situation_snapshot_repo.go`
6. `consts/task.go` 中 `RoleTypeOperator`
7. `decision_center.go` 前置 `AdmissionControl`
8. `registry.go` 中七层 Phase A 初始化

**实施完成状态（2026-04-06 更新）：**

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase A（L6 目标层 + 态势感知） | Sensor/Situation/Objective/AdmissionControl | ✅ 代码完成 |
| Phase B（L5 策略层） | Planner + 6 策略函数 + Actuator 效果跟踪 | ✅ 代码完成 |
| Phase C（L5 补全） | 8/8 ActionDispatcher 回调 + Parameters 注入 | ✅ 代码完成 |
| Phase D（L7 元认知层） | MetaObserver + Learner + MetaAssessor + MetaTuner | ✅ 代码完成 |

七层自治模型 L1~L7 全部代码完成，14 个灰度开关当前全部关闭，等待真实环境逐级验证。

---

## 3. 第一批边界

### 3.1 本批目标

本批上线后，应达到：

1. `Sensor` 能按 workflow_run 生成态势快照
2. `Situation` 能给 `DecisionCenter` 和后续 Planner 提供统一输入
3. `AdmissionControl` 能根据项目目标约束拒绝或降级动作
4. `Operator` 角色在数据层和常量层可用
5. 所有新增逻辑都可通过灰度开关关闭，关闭后行为与当前 `L3.5` 一致

### 3.2 本批不追求

1. 不改动 Accept / Rework 主链语义
2. 不引入新的自动动作类型
3. 不做前端控制台重构
4. 不做学习与自动调参

---

## 4. 数据库设计

### 4.1 直接采用的升级文件

直接采用现有：

- [20260410_l4_phase_a.sql](/www/wwwroot/project/easymvp/admin-go/docker/mysql/upgrade/20260410_l4_phase_a.sql)

包含：

1. `mvp_situation_snapshot`
2. `mvp_project.objective_json`
3. `mvp_workflow_run.tokens_consumed`
4. `mvp_workflow_run.replan_count`
5. `workflow.autonomy.*` Phase A 配置项
6. `operator` 角色预设初始化

### 4.2 本批直接使用的表

#### `mvp_situation_snapshot`

用途：

- 保存 `Sensor.Perceive()` 输出的统一态势快照
- 给 `DecisionCenter / AdmissionControl / 后续 Planner` 提供调试与审计依据

约束：

- 只追加，不更新
- 查询必须按 `workflow_run_id` 或 `project_id` 作用域
- 数据权限继承项目归属

#### `mvp_project.objective_json`

用途：

- 保存项目级目标约束

本批最小字段建议：

```json
{
  "deliveryGoal": "",
  "qualityFloor": 0.8,
  "tokenBudget": 0,
  "timeBudgetHours": 0,
  "costBudgetCents": 0,
  "riskTolerance": "balanced",
  "maxAutoRetries": 3,
  "maxAutoReworks": 2,
  "maxAutoReplans": 1,
  "deadlineAt": null,
  "maxStallMinutes": 60,
  "autonomyLevel": "supervised",
  "maxSideEffectLevel": "medium",
  "allowedStateChanges": [],
  "humanMandatoryPoints": []
}
```

#### `mvp_workflow_run.tokens_consumed`

用途：

- 作为目标层预算判断的第一版累计值

第一批实现要求：

- 不要求 100% 精确
- 先按已知执行与自治动作做累加
- 没有 token 信息的动作允许记 `0`

#### `mvp_workflow_run.replan_count`

用途：

- 给 `AdmissionControl` 判断是否超过 `maxAutoReplans`

---

## 5. Go 模块设计

### 5.1 新增文件

#### 1. `admin-go/app/mvp/internal/workflow/autonomy/situation.go`

职责：

- 定义 `Situation`
- 定义：
  - `ProgressMetrics`
  - `HealthMetrics`
  - `ResourceMetrics`
  - `TrendMetrics`
  - `AnomalySignal`

建议核心接口：

```go
type Situation struct { ... }

func (s *Situation) HasCriticalAnomaly() bool
func (s *Situation) SnapshotData() g.Map
```

#### 2. `admin-go/app/mvp/internal/workflow/autonomy/sensor.go`

职责：

- 从现有 `workflow_run / domain_task / stage_run / decision_action` 聚合态势
- 生成 `Situation`
- 检测异常信号
- 可选定时巡检

建议核心接口：

```go
type Sensor struct { ... }

func NewSensor(...) *Sensor
func (s *Sensor) Perceive(ctx context.Context, workflowRunID int64) (*Situation, error)
func (s *Sensor) DetectAnomalies(sit *Situation) []AnomalySignal
func (s *Sensor) RecordSnapshot(ctx context.Context, sit *Situation) error
```

#### 3. `admin-go/app/mvp/internal/workflow/autonomy/objective.go`

职责：

- 定义 `ProjectObjective`
- 解析 `mvp_project.objective_json`
- 执行 `AdmissionControl`

建议核心接口：

```go
type ProjectObjective struct { ... }
type AdmissionResult struct { ... }

type ObjectiveService struct { ... }

func NewObjectiveService(...) *ObjectiveService
func (s *ObjectiveService) Load(ctx context.Context, projectID int64) (*ProjectObjective, error)
func (s *ObjectiveService) Check(ctx context.Context, sit *Situation, obj *ProjectObjective, actionType string) (*AdmissionResult, error)
```

#### 4. `admin-go/app/mvp/internal/workflow/autonomy/decision_meta.go`

职责：

- 定义 `DecisionMeta`
- 定义 `BlastRadiusMinLevel`
- 定义 `Validate`

这是后续 Planner、Matcher、Tuner 的公共元数据结构。

#### 5. `admin-go/app/mvp/internal/workflow/repo/situation_snapshot_repo.go`

职责：

- 写入/查询 `mvp_situation_snapshot`
- 应用数据权限 scope

建议接口：

```go
type SituationSnapshotRepo struct { ... }

func (r *SituationSnapshotRepo) Create(ctx context.Context, data g.Map) (int64, error)
func (r *SituationSnapshotRepo) GetLatestByWorkflowRunID(ctx context.Context, workflowRunID int64) (g.Map, error)
func (r *SituationSnapshotRepo) ListByProjectID(ctx context.Context, projectID int64, limit int) ([]g.Map, error)
```

### 5.2 修改文件

#### 1. `admin-go/app/mvp/internal/consts/task.go`

新增：

```go
const RoleTypeOperator = "operator"
```

#### 2. `admin-go/app/mvp/internal/workflow/autonomy/model.go`

补充：

- 能力词表中的运维类能力常量
- 决策元数据引用
- `operator` 相关枚举

#### 3. `admin-go/app/mvp/internal/workflow/autonomy/decision_center.go`

改动原则：

- 不推翻现有 `L3.5`
- 只在入口前置：
  1. `Sensor.Perceive()`
  2. `ObjectiveService.Load()`
  3. `AdmissionControl.Check()`

调用顺序固定为：

`Perceive -> LoadObjective -> Check -> 原 Decide 逻辑`

准入失败时：

- 记录 denial evidence
- 返回 `Handled=false`
- 允许上层维持原有保守逻辑或进入人工路径

#### 4. `admin-go/app/mvp/internal/workflow/orchestrator/registry.go`

新增初始化：

- `SituationSnapshotRepo`
- `Sensor`
- `ObjectiveService`

并注入到 `DecisionCenter`

---

## 6. 对现有自治中台的接线方式

### 6.1 不允许的做法

1. 不允许直接重写 `Decide()` 全链
2. 不允许在第一批引入 Planner 六策略
3. 不允许在第一批修改旧动作回调语义

### 6.2 正确接线

第一批只在 `DecisionCenter.Decide()` 前增加：

1. `Situation` 快照采集
2. `Objective` 加载
3. `Admission` 判断
4. `Snapshot` 持久化

之后继续走现有：

- `policy_engine`
- `risk_gate`
- `decision_action`
- `human_checkpoint`
- `action_dispatcher`

这保证：

- 开关关闭时行为完全一致
- 开关开启时也是“前置约束增强”，不是“语义替换”

---

## 7. API 设计

第一批不需要大规模前端控制台，但后端建议补最小 API，方便自检与调试。

### 7.1 目标层接口

#### `GET /api/mvp/workflow/objective`

入参：

- `projectId`

出参：

- 当前项目 `objective_json` 解析结果

#### `POST /api/mvp/workflow/save-objective`

入参：

- `projectId`
- `objective`

用途：

- 后台保存目标约束

### 7.2 态势接口

#### `GET /api/mvp/workflow/situation`

入参：

- `workflowRunId`

出参：

- 最新 `Situation`
- 最近一次快照时间
- 异常信号列表

#### `GET /api/mvp/workflow/situation-history`

入参：

- `projectId`
- `workflowRunId`
- `limit`

出参：

- 最近 N 条快照

### 7.3 系统检查增强

建议在现有 `SystemCheck` 里追加：

- `objective_enabled`
- `patrol_enabled`
- `operator role preset`
- `latest situation snapshot`

但这部分可放到第一批后半段，不阻塞本批核心实现。

---

## 8. 前端建议

第一批不单独起完整自治控制台二改，只做最小页面补充：

1. 在现有 `autonomy.vue` 或 `dashboard.vue` 上增加：
   - 当前 `autonomyLevel`
   - `riskTolerance`
   - 最新 `Situation`
   - 最近异常信号

2. 在项目级设置中增加：
   - 目标约束编辑区

如果时间不够，本批前端可以只做：

- `objective` 配置页
- `situation` 查看弹窗

不要求完整时间序列图表。

---

## 9. 灰度与回滚

### 9.1 本批开关

直接沿用并收敛为：

- `workflow.autonomy.patrol_enabled`
- `workflow.autonomy.objective_enabled`

### 9.2 灰度顺序

1. `patrol_enabled=1`
   - 只采集，不拦截

2. `objective_enabled=1`
   - 开启准入控制
   - 但只拒绝明显超限动作

### 9.3 回滚原则

回滚只需：

1. 关闭 `objective_enabled`
2. 关闭 `patrol_enabled`

不需要：

- 回滚表结构
- 回滚已有 `decision_center` 逻辑

所以本批是标准的“加法式地基层”。

---

## 10. 实施顺序

### 第一批开发顺序

1. 确认并执行 [20260410_l4_phase_a.sql](/www/wwwroot/project/easymvp/admin-go/docker/mysql/upgrade/20260410_l4_phase_a.sql)
2. 新增 `situation.go`
3. 新增 `decision_meta.go`
4. 新增 `objective.go`
5. 新增 `situation_snapshot_repo.go`
6. 新增 `sensor.go`
7. 修改 `consts/task.go`
8. 修改 `autonomy/model.go`
9. 修改 `decision_center.go`
10. 修改 `registry.go`
11. 补最小 API
12. 补最小前端展示

### 推荐联调顺序

1. SQL 生效
2. `Sensor.Perceive()` 手动调通
3. `ObjectiveService.Load()` 手动调通
4. `AdmissionControl.Check()` 用单测覆盖关键分支
5. `DecisionCenter.Decide()` 接前置链
6. 灰度打开 `patrol_enabled`
7. 再打开 `objective_enabled`

---

## 11. 验收标准

第一批完成后，至少满足：

1. 能看到 `mvp_situation_snapshot` 持续写入
2. 项目能保存和读取 `objective_json`
3. `DecisionCenter` 能在超预算/超重试/超返工时返回拒绝或降级
4. `operator` 角色常量、预设、后端识别链路可用
5. 关闭开关后，行为与当前 `L3.5` 完全一致

一句话验收标准：

`系统先学会感知和收紧，再学会更聪明。`

---

## 12. 实施完成记录

全部四批均已完成：

| 批次 | 完成日期 | 关键产出 |
|------|----------|----------|
| Phase A | 2026-04-04 | L6 Sensor/Situation/Objective + AdmissionControl + 3 个灰度开关 |
| Phase B | 2026-04-04 | L5 Planner + 6 策略 + Actuator + 6 个灰度开关 |
| Phase C | 2026-04-05 | 8/8 回调补全 + Parameters 注入 + switch_executor |
| Phase D | 2026-04-05 | L7 MetaObserver/Learner/MetaAssessor/MetaTuner + 4 张表 + 3 个灰度开关 |

数据库：55 张表。灰度开关：14 个，全部关闭。前端：11 个 workflow 页面 + 元认知面板。

也就是从：

`会感知、会拒绝`

升级到：

`会感知、会拒绝、会选择更优动作`

