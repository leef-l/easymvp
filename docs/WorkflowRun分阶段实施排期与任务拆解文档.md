# WorkflowRun 分阶段实施排期与任务拆解文档

> 文档定位：WorkflowRun 架构升级项目管理文档
>
> 目标：将整体重构拆分为可执行阶段、里程碑、任务包、依赖关系和交付标准。

---

## 一、实施策略

### 1.1 总体策略

采用：

- 架构激进
- 发布渐进
- 新旧双轨
- 阶段验收

### 1.2 实施原则

1. 每一阶段必须可独立上线。
2. 每一阶段必须保留回滚路径。
3. 不允许"一次性切全量项目"。
4. 先打地基，再切主链。

---

## 二、总体阶段划分

建议分为 7 个里程碑：

1. M1：基础设施与数据模型
2. M2：设计阶段版本化
3. M3：审核阶段任务化
4. M4：执行阶段迁移
5. M5：返工链重构与事件总线
6. M5.5：默认组织模型收敛
7. M6：前端控制台与旧链退役

---

## 三、M1：基础设施与数据模型

### 3.1 目标

完成新架构地基，不切业务主链。

### 3.2 任务拆解

#### 后端

1. 新建表：
   - workflow_run
   - stage_run
   - plan_version
   - task_blueprint
   - review_issue
   - stage_task
   - workflow_event
2. 生成 DAO / Entity / DO
3. 建立 `workflow/` 目录骨架
4. 建立 runtime manager
5. 建立 event publisher

#### 前端

1. 项目列表增加 `engineVersion`
2. 预留 Workflow Dashboard 路由

### 3.3 验收标准

1. 新表全部建成
2. 新后端模块可编译
3. 项目创建可选择 `workflow_v2`

### 3.4 实际进度：✅ 基本完成（2026-04-04）

- 10 张新表已建成并执行
- DAO/Entity/DO 已生成
- `workflow/` 目录骨架（orchestrator/runtime/event/repo/domain/stage/scheduler/compat）已完成
- runtime manager、event publisher 已实现
- 前端 engineVersion 选择已落地

---

## 四、M2：设计阶段版本化

### 4.1 目标

让新项目的架构师拆分结果进入 `plan_version + task_blueprint`，不再直接生成旧任务。

### 4.2 任务拆解

#### 后端

1. PlanVersionService
2. BlueprintService
3. 从 architect reply 生成 plan version
4. blueprint parser
5. 版本 supersede 规则

#### 前端

1. Plan Design 页面
2. 版本列表
3. 蓝图表
4. 版本 diff 面板

### 4.3 验收标准

1. 新项目每次拆分都生成新版本
2. 旧版本可查询
3. 不再覆盖旧任务

### 4.4 实际进度：✅ 已完成（2026-04-04）

后端已落地：
- PlanVersionService（创建、提交审核、通过、驳回、版本号递增）
- BlueprintCreator 回调模式（避免循环依赖）
- 从 architect reply 生成 plan_version + task_blueprint
- controller V2 分支（ConfirmPlan/ParseTasks 走蓝图路径）

前端已落地：
- 项目创建表单 engineVersion 选择

未完成（可后置到 M6）：
- Plan Design 独立页面
- 版本 diff 面板
- 版本 supersede 全流程可视化

---

## 五、M3：审核阶段任务化

### 5.1 目标

将 `reviewing` 完整迁移到 `stage_run(review) + stage_task + review_issue`。

### 5.2 任务拆解

#### 后端

1. ReviewStageService
2. Precheck stage task
3. Auditor review stage task
4. Coordinator optimize stage task
5. Review summary stage task
6. ReviewIssueService
7. 审核结论状态机

#### 前端

1. Review Workspace
2. Issue Table
3. Stage Task Card
4. 审核结论展示

### 5.3 验收标准

1. 审核问题独立落库
2. 审核过程可追踪
3. 审核拒绝可回退设计阶段
4. 审核通过不再依赖旧 `draft task`

### 5.4 实际进度：✅ 已完成（2026-04-04）

后端已落地：
- ReviewStageService 完整审核流程（precheck → auditor → coordinator → summary）
- ReviewIssue 独立落库
- 审核结论状态机（approve → 推进 execute / reject → 回退 design）
- 手动审批/驳回 API
- DesignRollbackFn 回调模式

前端已落地：
- `workflow/review.vue` 审核工作台（状态、问题列表、手动审批/驳回）
- Review API（reviewStatus/reviewIssues/manualApprove/manualReject）

---

## 六、M4：执行阶段迁移

### 6.1 目标

将审核通过后的蓝图实例化到 `mvp_domain_task`，新项目执行完全脱离旧 `mvp_task`。

### 6.2 任务拆解

#### 后端

1. 建立 `mvp_domain_task`
2. 建立 `DomainTaskScheduler`
3. 建立新状态机
4. 建立资源锁表
5. 建立 watchdog v2
6. 对接执行器（chat/aider/openhands）

#### 前端

1. Execution Console
2. Batch Board
3. Resource Lock Panel
4. Task Chain Drawer

### 6.3 验收标准

1. 新项目执行阶段不再写旧 `mvp_task`
2. 资源锁可视化
3. 任务状态与执行流实时一致

### 6.4 实际进度：🔶 大部分完成 + CR 收口（2026-04-04）

后端已落地：
- TaskService.InstantiateFromBlueprint（蓝图 → domain_task）
- DomainTaskScheduler（批次门控 + 资源锁 + 并发控制 + 依赖检查）
- ExecuteStageService + 执行器对接（aider/chat 桥接旧 engine）
- review 通过 → execute stage 自动推进（CR-1 收口）
- V2 Pause/Resume 分流到新 WorkflowService + DomainTaskScheduler（CR-2 收口）
- V2 RetryTask/SkipTask 分流到新任务服务（CR-7 收口）
- Resume 后重建 runtime context（CR-6 收口）
- 审核异步链挂到 runtime context（CR-3 收口）
- ProjectStatusAdapter 最小可用实现（P1-1）
- project-status API V2 聚合字段（CR-4 收口）
- RetryTask 检查 RowsAffected，避免 V2 假成功（CR-8 收口）
- execute 启动失败不再静默吞掉，开始收紧“审核成功但执行未启动”的假成功语义（CR-9 收口）

前端已落地：
- Dashboard 接通真实 API（CR-5 收口）

未完成（M5/M6 阶段）：
- Execution Console 独立页面
- Batch Board / Resource Lock Panel / Task Chain Drawer
- watchdog v2（当前复用旧 watchdog 逻辑）
- review → execute 的失败语义仍未完全原子化，当前已从“静默吞掉”提升为“显式报错/日志”，但还不是强一致事务阶段

### 6.5 当前判断：M4 已进入“主链可用收口期”

截至 `1b0b0a5`：

1. WorkflowRun 主链已经不再是“骨架状态”，而是进入了可运行主链收口阶段。
2. Claude 两轮 code review 修正（`07b0662`、`7f3dcdd`）以及后续补尾巴提交（`1b0b0a5`）已经把大部分 P0/P1 主链问题补掉。
3. 当前 M4 剩余工作已从“大块建模”转向“执行控制台、可观测、失败一致性、隔离能力”。

因此，从排期上不再建议继续把 M4 看成“尚未成型阶段”，而应视为：

- 主链已成型
- 工程化能力待补齐

---

## 七、M4.5：Git Worktree 任务级环境隔离

### 7.1 目标

在 WorkflowRun 执行主链已经基本稳定后，为写仓任务引入轻量级任务环境隔离。

目标不是替代 WorkflowRun 主链，而是作为 M4 的执行层增强能力，优先解决：

1. 主工作区污染
2. 多任务写仓互相影响
3. Aider / OpenHands / 未来 CLI 执行器缺少隔离

### 7.2 为什么放在 M4 后、M5 前

原因：

1. 当前 WorkflowRun 主链已经基本闭环，可以承接执行层增强。
2. 如果在 M2/M3 阶段就接入 worktree，会反复返工执行主链。
3. 如果等到所有执行器都接完再做，主工作区污染会持续积累。

因此，最合适的插入点是：

- 在 M4 主链闭环后开始
- 在 M5/M6 大规模前端与返工能力之前落地

### 7.3 范围

第一阶段只覆盖：

1. `workflow_v2`
2. 写仓任务
3. 优先 `aider`

暂不覆盖：

1. legacy 主链
2. `chat`
3. 重型容器沙箱

### 7.4 任务拆解

#### 后端

1. 新增 `mvp_task_workspace`
2. 新增 `workspace/` 模块
3. 实现 `git worktree prepare/cleanup`
4. 在 execute 链路中接入 workspace manager
5. Aider 改为使用 task workspace
6. 任务结束后支持保留/清理 worktree

#### 文档与运维

1. 清理策略
2. 回写策略
3. 容量与磁盘占用监控

### 7.5 验收标准

1. `workflow_v2 + aider` 任务默认不直接在主工作区运行
2. 失败任务不会直接污染主工作区
3. 每个任务都能定位独立 workspace
4. 清理策略可执行

### 7.6 当前实施建议

建议分两步：

#### M4.5-A：地基阶段

1. 表结构
2. workspace manager
3. create / cleanup 原型
4. 先不切默认执行链

#### M4.5-B：Aider 接入阶段

1. 仅 `workflow_v2 + aider` 切入 worktree
2. 先不处理 OpenHands 容器化
3. 跑真实任务验证污染与回写策略

### 7.7 与 OpenHands 的关系

Git worktree 不是 OpenHands 的最终隔离形态。

更准确的路线是：

1. 先用 worktree 解决“目录污染问题”
2. 后续再为 OpenHands 升级到 sandbox/container

因此：

- worktree 是执行层隔离第一版
- OpenHands 容器化是后续增强，不应阻塞当前排期

---

## 八、M5：返工链重构与事件总线

P0 校准已完成：
- WorkflowRun 状态从扁平改为阶段化（designing/reviewing/executing/reworking/paused/completed/failed/canceled）
- FailStage 字段 fail_reason → error_message
- CreateRun → StartDesign 语义闭环

---

### 8.0 排期说明调整

从当前代码进度看，M5 不应再紧接“执行主链未闭环”的旧假设推进。  
正确顺序应改为：

1. 先完成 M4 收口
2. 插入 M4.5 worktree 隔离
3. 再推进 M5 返工链与事件总线

这样可以避免：

1. 返工链建立在不稳定执行环境上
2. 多执行器接入后继续污染主工作区

### 7.1 目标

将 bugloop/failure handoff 升级为正式 `rework stage`，并建立统一事件流。

### 7.2 任务拆解

#### 后端

1. ReworkStageService
2. BugAnalysisTask
3. FailureAnalysisTask
4. HandoffRecord
5. Workflow event stream
6. Conversation event stream

#### 前端

1. Rework Workspace
2. Event Feed
3. Timeline 页面

### 7.3 验收标准

1. 返工链不再依赖名称前缀
2. 事件流可驱动页面实时刷新
3. 返工轮次可统计

---

## 九、M6：前端控制台与旧链退役

## 九、M5.5：默认组织模型收敛

### 9.1 目标

在 M5 主链能力基本闭环后，收敛系统默认组织模型，避免角色继续发散。

默认模型明确为：

1. 1 个主架构师
2. 1 个审核员
3. 1 个协调员

该阶段不扩新角色，只收敛默认模板、默认模型绑定和默认前端呈现。

### 9.2 任务拆解

#### 后端

1. 默认项目角色模板收敛为 `architect / auditor / coordinator`
2. 默认模型绑定与执行方式预设
3. 项目创建时默认角色配置模板
4. rework 阶段继续复用主架构师，不引入第二架构师体系
5. 角色预设与 role preset 文案统一

#### 前端

1. 项目创建页默认角色模板展示
2. Review / Workflow 页面角色文案统一
3. 默认组织模型说明与只读展示
4. 历史 legacy 项目的兼容展示

### 9.3 验收标准

1. 新项目默认组织模型固定为 3 角色
2. 默认模型绑定与默认执行方式可查询
3. 前后端对角色名称、职责、展示文案一致
4. rework / review / design 三阶段不再额外扩默认角色

### 9.4 放置位置说明

M5.5 不插入当前 M5 主线，原因是：

1. 当前主线仍应优先收口 watchdog、rework、complete
2. 角色模型收敛属于组织层/配置层优化，不是当前执行主链阻塞项
3. 放在 M5 后、M6 前最稳，可以为执行器扩展和控制台最终形态提供稳定角色边界

---

## 十、M6：前端控制台与旧链退役
### 10.1 目标

新项目全面切到新控制台，旧链退为兼容模式。

### 10.2 任务拆解

#### 后端

1. compat gateway 完成
2. old API adapter 完成
3. legacy 只读维护模式

#### 前端

1. Workflow Dashboard 完整版
2. 项目列表聚合展示
3. 旧入口标记为 legacy

### 10.3 验收标准

1. 新项目默认使用 Workflow V2
2. 旧项目仍可查看
3. 新旧项目混合列表稳定

---

## 十一、任务包划分建议

### 9.1 后端任务包

- 包 A：数据模型与 SQL
- 包 B：workflow runtime / orchestrator
- 包 C：plan version / blueprint
- 包 D：review stage
- 包 E：domain task / scheduler
- 包 F：task workspace / git worktree isolation
- 包 G：rework / event bus / compat

### 9.2 前端任务包

- 包 H：workflow dashboard
- 包 I：plan design
- 包 J：review workspace
- 包 K：execution console
- 包 L：event stream / timeline

### 9.3 测试任务包

- 包 M：单元测试
- 包 N：集成测试
- 包 O：灰度验证与回归测试

---

## 十二、依赖关系

### 12.1 强依赖

1. M1 完成后才能做 M2
2. M2 完成后才能做 M3
3. M3 完成后才能安全做 M4
4. M4 收口后先做 M4.5
5. M4.5 完成后再做 M5
6. M5 完成后再做 M5.5
7. M5.5 完成后才能做 M6

### 12.2 可并行部分

- 前端页面骨架可与后端模型并行
- Dashboard 与 Plan 页面可先行
- worktree 地基建设可与 M4 后半段收口并行，但不建议在主链未稳前切默认执行链
- 事件流接入可与 Rework 并行推进
- M5.5 的文案、预设和前端只读展示可在 M5 后半段预研，但不建议提前改默认项目模板

---

## 十三、里程碑交付物

### M1 交付物

- SQL 脚本
- 新表 DAO
- runtime 基础模块

### M2 交付物

- plan version API
- blueprint API
- 计划设计页

### M3 交付物

- review stage service
- review issue API
- 审核工作台

### M4 交付物

- domain task API
- execution scheduler
- 执行控制台

### M4.5 交付物

- task workspace 表
- git worktree manager
- aider 隔离执行链
- worktree 清理机制

### M5 交付物

- rework stage
- workflow SSE
- timeline 页面

### M5.5 交付物

- 默认 3 角色模板
- 默认模型绑定与执行方式预设
- 统一角色文案与展示

### M6 交付物

- 兼容层
- 新旧分流上线
- legacy 退役方案

---

## 十四、团队协作建议

### 12.1 推荐分工

1. 后端架构负责人：
   - workflow runtime
   - orchestrator
   - state machine
2. 后端执行负责人：
   - domain task
   - scheduler
   - watchdog
   - task workspace
3. 前端负责人：
   - workflow dashboard
   - review workspace
   - execution console
4. 测试负责人：
   - 场景回归
   - 新旧兼容验证

### 12.2 代码评审重点

1. 状态流转是否合法
2. runtime cancel 是否真正传递
3. 事件是否重复或漏发
4. 新旧项目分流是否正确

---

## 十五、时间估算建议

如果是 2-3 名核心开发并行推进，建议按以下估算：

- M1：3-5 天
- M2：4-6 天
- M3：5-8 天
- M4：7-10 天
- M4.5：4-6 天
- M5：5-7 天
- M5.5：2-3 天
- M6：4-6 天

整体建议周期：

- 4 到 7 周

如果需要压缩周期，优先保证：

- M1
- M2
- M3
- M4
- M4.5

M5/M5.5/M6 可稍后补齐。

---

## 十六、当前剩余工作建议排序（2026-04-04 校准后）

### P0：M4 收口

1. execute 启动失败语义进一步收紧
2. Execution Console 最小版
3. watchdog v2 与 domain_task 对齐

### P1：M4.5 地基

1. `mvp_task_workspace`
2. `workspace/` 模块
3. worktree create/cleanup 原型

### P1：M4.5 Aider 接入

1. `workflow_v2 + aider` 切到 worktree
2. 真实项目验证

### P2：M5

1. rework stage
2. event stream
3. timeline

### P2.5：M5.5

1. 默认 3 角色模板
2. 默认模型绑定
3. 前后端角色文案统一

---

## 十七、上线策略

### 14.1 灰度建议

1. 内部测试项目先切
2. 新建项目小比例切换
3. 观察 3-5 天
4. 扩大到全部新项目
5. 旧项目维持 legacy

### 14.2 上线门槛

满足以下条件才允许扩大灰度：

1. 审核通过率数据正常
2. 执行失败率未明显恶化
3. 暂停恢复稳定
4. 无大规模事件流异常

---

## 十八、结论

实施上最重要的是克制，不要一边重构所有层，一边全量切流量。

正确方式是：

- 架构上一步到位
- 实施上分阶段达成
- 每阶段都可交付、可验证、可回滚

在当前代码进度下，建议正式采用下面的推进顺序：

1. WorkflowRun 主链继续收口到稳定
2. 插入 Git Worktree 任务级隔离
3. 再继续返工链和事件总线
4. 然后收敛默认组织模型
5. 最后再推进执行器扩展和旧链退役

这样才能把一次大重构真正落地。
