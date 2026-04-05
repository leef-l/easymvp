# WorkflowRun 已实现模块数据权限补强清单

## 一、目的

本文档用于审查当前已实现的 WorkflowRun / Workflow V2 模块，判断哪些地方已经为后续数据权限治理留下扩展口，哪些地方仍然缺口明显。

目标不是立即推动重构，而是冻结后续权限治理的补强边界，避免系统继续扩大裸查和弱权限模型。

## 二、结论

当前 WorkflowRun 模块对数据权限只做到了“第一层留口”，还没有做到“系统性可扩展”。

当前已具备：

1. `mvp_project` 有 `created_by / dept_id`
2. `mvp_workflow_run` 有 `created_by / dept_id`
3. Workflow 控制器入口大多通过 `project` 做 owner 校验

当前仍明显缺失：

1. V2 核心业务表缺少统一权限归属策略
2. Repo 层没有统一 scope 入口
3. 控制器仍是“项目创建人独占式”权限模型
4. 运行态 issue / task / version / stage 等对象缺少稳定权限锚点

因此，当前状态只能支撑：

1. 项目创建人访问
2. 管理员全量访问

尚不足以平滑支撑：

1. 部门级数据域
2. 项目成员协作权限
3. 审核人 / 运营人只读权限
4. evidence / issue 的细粒度查看权限

## 三、已留口部分

### 3.1 `mvp_project`

当前已有：

1. `created_by`
2. `dept_id`

这是现有 Workflow 控制器最主要的权限入口。

相关证据：

- [mvp_project.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_project.go)

### 3.2 `mvp_workflow_run`

当前已有：

1. `created_by`
2. `dept_id`

这是 WorkflowRun 主对象级别的权限锚点。

相关证据：

- [mvp_workflow_run.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_workflow_run.go)

### 3.3 控制器入口

当前 Workflow 相关控制器普遍先做项目归属校验。

相关证据：

- [workflow.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/controller/chat/workflow.go#L26)

现状说明：

1. 管理员可直接通过
2. 普通用户要求 `project.created_by == userID`

这属于“最基础的 owner 校验”，不是完整数据权限体系。

## 四、必须补强项

以下内容建议列为后续权限治理的强制补强项。

### 4.1 `mvp_review_issue`

问题：

1. 当前表没有 `created_by / dept_id`
2. 当前权限只能通过 `stage_run -> workflow_run -> project` 间接继承

风险：

1. 后续审核问题列表按部门裁剪会变复杂
2. 运营、审计、审核人只读查看难以标准化

相关证据：

- [mvp_review_issue.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_review_issue.go)
- [20260406_workflow_run_core.sql](/www/wwwroot/project/easymvp/docker/mysql/upgrade/20260406_workflow_run_core.sql#L107)

结论：

- 必须补权限归属策略

### 4.2 `mvp_domain_task`

问题：

1. 当前表没有 `created_by / dept_id`
2. 执行控制台直接按 `workflow_run_id` 全量取任务

风险：

1. 后续项目成员、审核员、运营只读查看难以细分
2. Accept / Rework / Dashboard / Execution Console 都会越来越依赖任务级对象

相关证据：

- [mvp_domain_task.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_domain_task.go)
- [20260407_workflow_run_execution.sql](/www/wwwroot/project/easymvp/docker/mysql/upgrade/20260407_workflow_run_execution.sql#L10)
- [workflow.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/controller/chat/workflow.go#L1224)

结论：

- 必须补权限归属策略

### 4.3 `mvp_plan_version`

问题：

1. 当前表没有 `created_by / dept_id`
2. 版本视图后续会是设计阶段的重要查询对象

风险：

1. 后续版本回看、版本 diff、验收关联都缺少直接权限锚点

相关证据：

- [mvp_plan_version.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_plan_version.go)
- [20260406_workflow_run_core.sql](/www/wwwroot/project/easymvp/docker/mysql/upgrade/20260406_workflow_run_core.sql#L59)

结论：

- 必须补权限归属策略

### 4.4 Repo 查询层

问题：

1. 当前 WorkflowRun Repo 基本没有统一的 `scope` 或 `DataScope`
2. 查询大多直接按 `project_id / workflow_run_id` 查

风险：

1. 后续接入部门域时只能分散补 `Where("dept_id", ...)`
2. 查询层难以统一收敛

相关证据：

- [workflow_run_repo.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/repo/workflow_run_repo.go)
- [stage_run_repo.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/repo/stage_run_repo.go)
- [plan_version_repo.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/repo/plan_version_repo.go)
- [review_issue_repo.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/workflow/repo/review_issue_repo.go)

结论：

- 必须补统一查询作用域入口

### 4.5 `checkProjectOwnership()`

问题：

1. 当前只支持管理员或项目创建人
2. 不支持项目成员、部门域、只读角色

相关证据：

- [workflow.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/controller/chat/workflow.go#L26)

结论：

- 必须升级成“项目作用域校验”，不能长期停留在 owner-only

## 五、建议补强项

### 5.1 `mvp_stage_run`

问题：

1. 当前无 `created_by / dept_id`
2. 阶段历史、时间线、Accept/运营台都会使用

相关证据：

- [mvp_stage_run.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_stage_run.go)
- [workflow.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/controller/chat/workflow.go#L1112)

结论：

- 建议明确权限继承策略

### 5.2 `mvp_task_blueprint`

问题：

1. 当前无 `created_by / dept_id`
2. 主要通过 `plan_version_id` 间接挂载

相关证据：

- [mvp_task_blueprint.go](/www/wwwroot/project/easymvp/admin-go/app/mvp/internal/model/entity/mvp_task_blueprint.go)

结论：

- 建议明确权限继承策略

### 5.3 前端 workflow API

问题：

1. 当前前端查询主要只传 `projectID`
2. 没有冻结更细粒度的权限语义

相关证据：

- [index.ts](/www/wwwroot/project/easymvp/vue-vben-admin/apps/web-antd/src/api/mvp/workflow/index.ts)

结论：

- 建议在后续接口文档中冻结作用域语义

## 六、可继承不补项

### 6.1 `mvp_workflow_run`

结论：

- 当前已有 `created_by / dept_id`
- 可继续作为上游权限锚点

### 6.2 `mvp_project`

结论：

- 当前已有 `created_by / dept_id`
- 仍是最稳定的项目级权限根对象

### 6.3 运行附属对象

包括：

1. `mvp_handoff_record`
2. `mvp_workflow_event`
3. `mvp_task_resource_lock`

结论：

- 可以先不单独补归属字段
- 但必须在后续设计里写死“只能继承上游 workflow/project 权限”

## 七、后续冻结原则

### 7.1 新表判断原则

所有新增 WorkflowRun 相关表，先区分两类：

1. 独立业务对象
2. 运行附属对象

规则：

1. 独立业务对象优先带 `created_by / dept_id`
2. 运行附属对象至少要冻结权限继承链

### 7.2 查询设计原则

所有 WorkflowRun Repo 后续统一支持 scope 注入，至少预留：

1. `project_id`
2. `workflow_run_id`
3. `created_by`
4. `dept_id`

### 7.3 issue / evidence 设计原则

所有 issue / evidence / summary 类型对象：

1. 不允许独立裸查
2. 必须绑定上游业务对象
3. 必须明确权限继承来源

## 八、优先级建议

### 8.1 P0

1. 升级 `checkProjectOwnership()` 为项目作用域校验
2. Repo 层引入统一 scope 入口
3. 冻结 `review_issue / domain_task / plan_version` 的权限策略

### 8.2 P1

1. 冻结 `stage_run / task_blueprint` 的权限继承策略
2. 冻结前端 workflow API 的作用域语义

### 8.3 P2

1. 整理运行附属对象的权限继承规则
2. 将权限约束补入后续 Accept / 月 6 文档

## 九、结论

当前 WorkflowRun 已实现模块并不是完全没有数据权限口子，但只开到了第一层：

1. 项目级 owner 校验
2. `project / workflow_run` 归属字段

这不足以支撑后续系统化权限治理。

如果后续要顺利接入：

1. 部门级数据域
2. 项目成员权限
3. 审核/运营只读权限
4. 更细粒度的 issue / task / evidence 权限

则必须将本文档中的 `必须补强项` 纳入后续计划基线。
