# WorkflowRun 测试验收与回滚预案文档

> 文档定位：WorkflowRun 重构项目的质量保障文档
>
> 目标：定义从开发测试、联调验收、灰度验证到线上回滚的完整质量体系。

---

## 一、目标

本次重构属于内核级改造，测试目标不是“页面能打开”，而是保证：

1. 新工作流链路完整可跑。
2. 新旧项目可并存。
3. 状态机一致。
4. 暂停/恢复/取消可靠。
5. 灰度失败时可回滚。

---

## 二、测试范围

### 2.1 功能范围

- WorkflowRun 创建
- StageRun 切换
- PlanVersion 生成
- ReviewStage 审核
- DomainTask 执行
- ReworkStage 返工
- Event Stream 推送

### 2.2 非功能范围

- 并发调度
- 锁恢复
- SSE 稳定性
- 兼容老项目
- 回滚路径

### 2.3 当前双轨期说明（截至 2026-04-04）

当前系统处于**新旧引擎双轨运行期**，理解双轨边界是测试的前提：

**旧链路（Legacy）**：
- `internal/engine/` 仍是 legacy 项目的主执行链
- `mvp_task` 表 + `engine.Scheduler` + `engine.Executor` + `engine.Watchdog`
- 所有 `engine_version != 'workflow_v2'` 的项目走此链路

**新链路（Workflow V2）**：
- `internal/workflow/` 是新内核
- `mvp_workflow_run` + `mvp_stage_run` + `mvp_domain_task` + `DomainTaskScheduler`
- 所有 `engine_version = 'workflow_v2'` 的项目走此链路

**控制面分流点**：
- `CreateProject`：按 `engineVersion` 决定是否创建 `workflow_run`
- `ConfirmPlan`：V2 走蓝图路径，Legacy 走旧调度器
- `Pause/Resume`：按 `engine_version` 分流到不同服务
- `RetryTask/SkipTask`：按 `engine_version` 分流到不同任务表
- `ProjectStatus`：V2 返回 `workflowStatus/currentStage/progressPercent` 等聚合字段

**共享部分**：
- `engine.ChatEngine`（对话引擎）两套共用
- `engine.AiderRunner`（Aider 执行器）两套共用
- AI 模型解析（`ResolveModelInfo`）由 `executor_bridge.go` 导出供新链路调用

**测试重点**：
- 新旧项目在同一系统中并行运行时互不干扰
- V2 项目的 Pause/Resume/Retry/Skip 不会误走旧调度器
- Legacy 项目不受新代码路径影响

---

## 三、测试分层

### 3.1 单元测试

重点覆盖：

1. 状态机校验
2. 计划版本 supersede 逻辑
3. 审核决策逻辑
4. blueprint -> domain_task 实例化逻辑
5. handoff 回写逻辑

### 3.2 集成测试

重点覆盖：

1. 新项目完整生命周期
2. 审核拒绝后回到设计阶段
3. 审核通过后实例化任务
4. 执行失败进入返工
5. 返工完成返回执行
6. 完成阶段归档

### 3.3 E2E 测试

重点覆盖：

1. 前端设计页 -> 审核台 -> 执行台完整流
2. SSE 驱动状态刷新
3. 旧项目 UI 不受影响

### 3.4 灰度验证

重点覆盖：

1. 新项目灰度创建
2. 实际 AI 审核超时与降级
3. 暂停恢复的线上真实行为

---

## 四、关键验收场景

### 4.1 场景 A：新项目标准通过链

步骤：

1. 创建 `workflow_v2` 项目
2. 与架构师对话
3. 生成 `plan_version`
4. 提交审核
5. 审核通过
6. 实例化 `domain_task`
7. 执行完成

验收点：

- `workflow_run` 状态正确
- `stage_run` 顺序正确
- `review_issue` 生成正确
- `domain_task` 正常实例化

### 4.2 场景 B：审核拒绝链

步骤：

1. 生成低质量 plan version
2. 提交审核
3. 预检或 auditor 拒绝

验收点：

- 工作流回到设计阶段
- issue 列表完整
- plan_version 状态为 rejected
- 未生成执行任务

### 4.3 场景 C：执行失败返工链

步骤：

1. 审核通过
2. 实施任务执行失败
3. 进入 `rework`
4. failure_analysis 完成
5. 返回 `execute`

验收点：

- rework stage 正常创建
- handoff record 正常写入
- 原任务链可追溯

### 4.4 场景 D：暂停恢复

步骤：

1. review 或 execute 中暂停
2. 检查 runtime cancel
3. 恢复

验收点：

- 无脏任务继续运行
- 恢复后可继续
- 状态不乱跳

### 4.5 场景 E：旧项目兼容

步骤：

1. 打开 legacy 项目
2. 查看任务、聊天、状态

验收点：

- 老功能无回归
- 不误入 Workflow V2 页面

---

## 五、测试用例矩阵

### 5.1 状态机矩阵

必须覆盖：

- workflow status 全转换
- stage status 全转换
- domain_task status 关键转换

### 5.2 引擎版本矩阵

- `legacy`
- `workflow_v2`

### 5.3 执行模式矩阵

- `chat`
- `aider`
- `openhands`（如启用）

### 5.4 项目分类矩阵

- coding
- creative
- analysis

---

## 六、重点自动化测试建议

### 6.1 后端自动化

建议优先写：

1. workflow/stage/task 状态机测试
2. review issue 批量创建测试
3. plan version supersede 测试
4. domain task instantiation 测试
5. runtime pause/resume 测试

### 6.2 前端自动化

建议优先写：

1. workflow dashboard 渲染测试
2. review issue table 交互测试
3. execution console 批次展示测试
4. SSE 事件解析测试

---

## 七、灰度发布验收指标

### 7.1 核心指标

- 新项目创建成功率
- 审核成功率
- 审核超时率
- 任务执行成功率
- 暂停恢复成功率
- SSE 断连恢复成功率

### 7.2 风险指标

- 状态机非法跳转次数
- workflow/stage/task 事件丢失次数
- 资源锁泄漏次数
- 返工链回写失败次数

### 7.3 阈值建议

若出现以下任一情况，暂停扩大灰度：

- 审核阶段失败率显著高于 legacy
- 状态机非法跳转连续出现
- 大量 workflow 进入未知状态
- 旧项目出现回归

---

## 八、回滚策略

### 8.1 回滚原则

1. 先停新流量
2. 再定位问题
3. 不强行把新数据写回旧表

### 8.2 回滚级别

#### 级别 1：前端回滚

适用：

- 新页面 bug
- 事件渲染异常

措施：

- 隐藏 Workflow V2 入口
- 回退到旧页面入口

#### 级别 2：路由分流回滚

适用：

- 新项目主链异常

措施：

- 新建项目强制回到 `legacy`
- 保留新表
- 停止继续写新数据

#### 级别 3：功能关闭

适用：

- workflow runtime 或 scheduler 存在严重一致性问题

措施：

- 关闭 Workflow V2 开关
- 停止新架构调度器

### 8.3 双轨期回滚要点

1. 回滚不需要删除新表——新旧表完全独立，旧链路不读新表
2. V2 项目如需回退到 Legacy，将 `engine_version` 改回 `legacy` 即可，后续操作走旧链路
3. 已在新链路中产生的 `workflow_run/stage_run/domain_task` 数据保留，不需要迁移回旧表
4. 控制面分流基于 `engine_version` 字段，修改该字段即可切换链路

### 8.4 不建议的回滚方式

- 删除新表
- 强制迁回旧任务表
- 直接改动线上历史记录

---

## 九、上线前检查清单

### 9.1 数据库

- 新表存在
- 索引存在
- SQL 幂等验证通过

### 9.2 后端

- 新旧项目分流生效
- runtime registry 正常
- scheduler 正常启动
- event publisher 正常

### 9.3 前端

- 新入口只对 `workflow_v2` 可见
- SSE 能连通
- 页面加载稳定

### 9.4 监控

- 日志关键字已配置
- 错误告警已配置
- 指标看板已准备

---

## 十、上线后观察清单

### 第一天重点观察

- 新建项目是否稳定
- 审核是否能正常完成
- 执行阶段是否能正常调度
- 暂停恢复是否可靠

### 前三天重点观察

- review issue 生成量
- 返工比例
- SSE 异常比例
- legacy 项目是否受影响

---

## 十一、验收结论标准

满足以下条件可视为阶段验收通过：

1. 新项目主链跑通
2. review/execution/rework 正常切换
3. 前端工作流页面可用
4. 旧项目无明显回归
5. 灰度期间无需触发级别 2/3 回滚

---

## 十二、结论

WorkflowRun 重构的成败，不取决于文档是否完整，而取决于：

1. 是否按阶段验收
2. 是否严守灰度和回滚纪律
3. 是否用指标判断上线而不是凭感觉

这份测试与回滚文档的作用，就是把大重构变成可控工程。
