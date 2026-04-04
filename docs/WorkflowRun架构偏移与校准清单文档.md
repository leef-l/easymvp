# WorkflowRun 架构偏移与校准清单文档

> 文档定位：WorkflowRun 设计方案与当前代码实现之间的校准文档
>
> 目标：明确当前代码与设计文档的偏移点，给出每项偏移的处理决策，避免后续开发在“跟设计”还是“跟现状”之间摇摆。

---

## 一、为什么需要这份文档

当前 WorkflowRun 相关工作已经进入“设计文档已完备、代码已大规模落地”的阶段。

问题在于：

1. 总体设计文档描述的是目标架构。
2. 当前代码实现的是分阶段落地结果。
3. 二者方向一致，但局部语义和阶段完成度已经出现偏移。

如果不显式校准，会出现三种风险：

1. 文档成为理想图，代码成为现实图，二者各走各的。
2. 新开发者不知道应该遵循设计文档还是代码现状。
3. 后续迭代持续在错误语义上叠加，返工成本越来越高。

---

## 二、校准原则

### 2.1 处理策略分类

每个偏移点都必须归类为以下三类之一：

1. `修代码`
2. `修文档`
3. `暂不处理`

### 2.2 决策规则

#### 需要修代码的情况

满足任一条即修代码：

1. 当前实现违背了目标架构的核心语义。
2. 当前实现会对后续阶段造成结构性拖累。
3. 当前实现已经是明确 bug。

#### 需要修文档的情况

满足任一条即修文档：

1. 目标架构没错，但实施阶段描述已与现实进度不符。
2. 页面、模块、排期的落地顺序与文档假设不同。
3. 当前实现是合理渐进方案，文档过于理想化。

#### 暂不处理的情况

满足任一条可暂缓：

1. 不影响主链推进。
2. 属于阶段性骨架保留。
3. 后续阶段自然会覆盖，不值得现在返工。

---

## 三、当前代码实际阶段判断

基于当前代码，不按里程碑名称而按真实落地程度判断：

### 3.1 M1：基础设施与数据模型

状态：`基本完成`

已落地：

- 新表
- DAO/Entity/DO
- `workflow/` 目录骨架
- runtime manager
- workflow service / stage service 骨架
- 前端引擎版本入口

### 3.2 M2：设计阶段版本化

状态：`部分完成`

已落地：

- `plan_version`
- `task_blueprint`
- 部分 repo / service

未完成：

- 完整 design service
- 版本 diff / supersede 全流程
- 前端 plan design 页面

### 3.3 M3：审核阶段任务化

状态：`部分完成`

已落地：

- `review_issue`
- `stage_task`
- review stage service 部分实现
- 前端 `workflow/review.vue`

未完成：

- decision service
- issue service
- stage scheduler
- 审核阶段统一闭环

### 3.4 M4：执行阶段迁移

状态：`已开工，未闭环`

已落地：

- `domain_task`
- `resource_lock`
- 部分 task service / scheduler 骨架

未完成：

- 新项目完全脱离旧 `mvp_task`
- execute stage 全链打通
- execution console 前端

---

## 四、偏移清单

### 偏移 1：WorkflowRun 状态语义被压平

#### 设计文档定义

WorkflowRun 状态应为：

- `designing`
- `reviewing`
- `executing`
- `reworking`
- `paused`
- `completed`
- `failed`
- `canceled`

#### 当前代码现状

当前落地为：

- `pending`
- `running`
- `paused`
- `completed`
- `canceled`

涉及：

- [task.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/consts/task.go)
- [20260406_workflow_run_core.sql](/www/wwwroot/project/easymvp/docker/mysql/upgrade/20260406_workflow_run_core.sql)
- [workflow_service.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/orchestrator/workflow_service.go)
- [consts.ts](/www/wwwroot/project/easymvp/vue-vben-admin/apps/web-antd/src/views/mvp/consts.ts)

#### 风险

1. WorkflowRun 失去阶段化表达力。
2. 以后会重新依赖 `current_stage` 去补语义。
3. 新架构退化回旧“状态 + 附加字段”的模式。

#### 决策

`修代码`

#### 建议动作

1. 统一常量
2. 修 SQL 注释和数据字典
3. 修服务层状态迁移
4. 修前端状态映射

---

### 偏移 2：StageRun 失败字段名不一致

#### 设计文档定义

阶段失败信息字段统一使用：

- `error_message`

#### 当前代码现状

`StageService.FailStage()` 更新的是：

- `fail_reason`

但 SQL 表中定义的是：

- `error_message`

涉及：

- [stage_service.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/orchestrator/stage_service.go)
- [20260406_workflow_run_core.sql](/www/wwwroot/project/easymvp/docker/mysql/upgrade/20260406_workflow_run_core.sql)

#### 风险

1. 阶段失败路径直接报错。
2. M3/M4 所有失败流都不可靠。

#### 决策

`修代码`

#### 建议动作

1. 将 `fail_reason` 改为 `error_message`
2. 补失败路径测试

---

### 偏移 3：M1 已接入口，但语义上仍是半初始化

#### 设计文档定义

新建 `workflow_v2` 项目后，应进入正式的 `design stage` 运行状态。

#### 当前代码现状

当前已能：

- 创建 `workflow_run`
- 创建 `design stage_run`

但主链语义仍偏“建实例”，并非完全可用的设计阶段启动。

涉及：

- [controller/chat/workflow.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/controller/chat/workflow.go)
- [workflow_service.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/orchestrator/workflow_service.go)

#### 风险

1. 前端会认为 V2 项目已经可用。
2. 后端实际上只完成部分启动。

#### 决策

`修代码`

#### 建议动作

1. `CreateRun -> StartDesign` 接通
2. 项目创建后确保 workflow 状态进入正式设计态

---

### 偏移 4：Compat 层仍是骨架，文档里描述过于完整

#### 设计文档定义

兼容层负责新旧项目统一聚合展示。

#### 当前代码现状

`LegacyGateway` 已有；
`ProjectStatusAdapter` 仍是 TODO。

涉及：

- [legacy_gateway.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/compat/legacy_gateway.go)
- [project_status_adapter.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/compat/project_status_adapter.go)

#### 风险

1. 项目列表无法稳定聚合新旧状态。
2. 设计文档让人误以为 compat 已闭环。

#### 决策

`修代码 + 修文档`

#### 建议动作

代码：

1. 实现最小可用 `ProjectStatusAdapter`

文档：

1. 在实施排期里明确 compat 仍未完成

---

### 偏移 5：前端页面规划与实际落地顺序偏移

#### 设计文档定义

前端规划包含：

- plan design
- review workspace
- execution console
- rework workspace
- workflow dashboard
- timeline

#### 当前代码现状

目前已出现：

- `workflow/dashboard.vue`
- `workflow/review.vue`
- `workflow/timeline.vue`

未出现：

- plan design
- execution console
- rework workspace

#### 风险

1. 文档中的页面落地顺序与现实不一致。
2. 后续排期容易误判。

#### 决策

`修文档`

#### 建议动作

1. 更新前端实施文档的页面落地顺序
2. 将 `review` 页面提前到已落地项

---

### 偏移 6：M2-M4 的“完成度”文档和现实不一致

#### 设计文档定义

实施文档把 M1-M6 作为目标阶段。

#### 当前代码现状

从代码看：

- M1 基本完成
- M2 部分完成
- M3 部分完成
- M4 已开工但远未完成

#### 风险

1. 团队成员口头上容易误判“已经到 M4”。
2. 排期会失真。

#### 决策

`修文档`

#### 建议动作

1. 新增真实进度说明
2. 明确“阶段定义”和“阶段完成”不是一回事

---

### 偏移 7：旧 engine 仍是主执行链，但总架构文档未突出“双轨期”

#### 设计文档定义

总架构文档更多描述目标态。

#### 当前代码现状

现实是：

- `internal/engine/` 仍是主要可运行主链
- `internal/workflow/` 是新主链建设中

#### 风险

1. 阅读文档的人容易误认为新链已经取代旧链。
2. 实施时低估兼容成本。

#### 决策

`修文档`

#### 建议动作

1. 在实施文档中明确“双轨期”
2. 在测试与回滚文档中增加新旧混跑说明

---

## 五、优先级排序

### P0：必须立刻修代码

1. WorkflowRun 状态语义统一
2. StageRun 失败字段修正
3. `CreateRun -> StartDesign` 主链接通

### P1：必须尽快补齐

1. `ProjectStatusAdapter` 最小实现
2. 前端状态常量同步
3. Dashboard 展示真实 workflow 数据

### P2：修文档，防止认知偏差

1. 更新实施排期文档
2. 更新前端实施文档
3. 在测试回滚文档中加入双轨说明

---

## 六、建议的校准阶段

建议插入一个短阶段：

## `WorkflowRun 校准阶段`

目标：

1. 对齐核心语义
2. 修复实现偏差
3. 更新文档现实进度

该阶段不追求新增功能，只解决“文档和代码不再分叉”。

---

## 七、校准完成标准

满足以下条件，视为校准完成：

1. WorkflowRun 状态语义与设计文档一致
2. StageRun 失败路径可用
3. `workflow_v2` 新建项目能进入正式设计态
4. 项目列表能最小聚合展示新旧状态
5. 实施文档已反映真实完成度

---

## 八、结论

当前 WorkflowRun 方向没有跑偏，但已经出现了“目标架构”和“阶段性实现”之间的自然偏移。

这不是坏事，坏的是不承认偏移。

正确做法不是继续盲目往后冲，而是：

1. 修核心语义偏移
2. 修明显代码错误
3. 更新实施文档
4. 然后继续推进 M2-M4

这样后面的重构才不会边做边返工。
