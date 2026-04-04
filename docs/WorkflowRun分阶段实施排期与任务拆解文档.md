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

建议分为 6 个里程碑：

1. M1：基础设施与数据模型
2. M2：设计阶段版本化
3. M3：审核阶段任务化
4. M4：执行阶段迁移
5. M5：返工链重构与事件总线
6. M6：前端控制台与旧链退役

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

---

## 七、M5：返工链重构与事件总线

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

## 八、M6：前端控制台与旧链退役

### 8.1 目标

新项目全面切到新控制台，旧链退为兼容模式。

### 8.2 任务拆解

#### 后端

1. compat gateway 完成
2. old API adapter 完成
3. legacy 只读维护模式

#### 前端

1. Workflow Dashboard 完整版
2. 项目列表聚合展示
3. 旧入口标记为 legacy

### 8.3 验收标准

1. 新项目默认使用 Workflow V2
2. 旧项目仍可查看
3. 新旧项目混合列表稳定

---

## 九、任务包划分建议

### 9.1 后端任务包

- 包 A：数据模型与 SQL
- 包 B：workflow runtime / orchestrator
- 包 C：plan version / blueprint
- 包 D：review stage
- 包 E：domain task / scheduler
- 包 F：rework / event bus / compat

### 9.2 前端任务包

- 包 G：workflow dashboard
- 包 H：plan design
- 包 I：review workspace
- 包 J：execution console
- 包 K：event stream / timeline

### 9.3 测试任务包

- 包 L：单元测试
- 包 M：集成测试
- 包 N：灰度验证与回归测试

---

## 十、依赖关系

### 10.1 强依赖

1. M1 完成后才能做 M2
2. M2 完成后才能做 M3
3. M3 完成后才能安全做 M4
4. M4 完成后再做 M5
5. M5 完成后才能做 M6

### 10.2 可并行部分

- 前端页面骨架可与后端模型并行
- Dashboard 与 Plan 页面可先行
- 事件流接入可与 Rework 并行推进

---

## 十一、里程碑交付物

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

### M5 交付物

- rework stage
- workflow SSE
- timeline 页面

### M6 交付物

- 兼容层
- 新旧分流上线
- legacy 退役方案

---

## 十二、团队协作建议

### 12.1 推荐分工

1. 后端架构负责人：
   - workflow runtime
   - orchestrator
   - state machine
2. 后端执行负责人：
   - domain task
   - scheduler
   - watchdog
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

## 十三、时间估算建议

如果是 2-3 名核心开发并行推进，建议按以下估算：

- M1：3-5 天
- M2：4-6 天
- M3：5-8 天
- M4：7-10 天
- M5：5-7 天
- M6：4-6 天

整体建议周期：

- 4 到 6 周

如果需要压缩周期，优先保证：

- M1
- M2
- M3
- M4

M5/M6 可稍后补齐。

---

## 十四、上线策略

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

## 十五、结论

实施上最重要的是克制，不要一边重构所有层，一边全量切流量。

正确方式是：

- 架构上一步到位
- 实施上分阶段达成
- 每阶段都可交付、可验证、可回滚

这样才能把一次大重构真正落地。
