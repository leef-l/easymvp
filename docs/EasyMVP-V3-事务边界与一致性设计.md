# EasyMVP V3 事务边界与一致性设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Service与Repository接口分层设计](./EasyMVP-V3-Service与Repository接口分层设计.md)
> 关联文档：[EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-数据库Schema总设计.md)
> 关联文档：[EasyMVP-V3-后台Worker与任务调度设计](./EasyMVP-V3-后台Worker与任务调度设计.md)
> 目标：定义 V3 Go 本地核心服务中哪些业务动作必须事务化，哪些动作应采用异步最终一致，避免实现时出现半完成状态。

## 1. 设计结论

V3 的一致性设计应遵守：

1. 关键状态推进使用本地数据库事务
2. 文件落盘与外部运行时调用使用事务外编排
3. 聚合快照与页面事件允许最终一致

一句话：

> 事实对象强一致，视图对象最终一致。

## 2. 必须事务化的动作

以下动作必须在单个事务边界内完成：

1. 创建项目基础记录
2. 写入计划版本链
3. `CompiledTask -> DomainTask` 投影
4. 任务状态关键推进
5. 验收运行创建与最终裁决
6. 人工放行写入

## 3. 适合最终一致的动作

以下动作可以异步完成：

1. Workspace 快照刷新
2. evidence 索引刷新
3. replay 索引刷新
4. 审计视图聚合
5. 首页卡片统计刷新

## 4. 典型事务边界

### 4.1 创建项目

事务内：

1. `projects`
2. `project_profiles`
3. `project_workspaces`
4. 首个 `PlanDraft`

事务外：

1. 目录创建补齐
2. 初始化事件发射
3. 页面跳转

### 4.2 编译计划

事务内：

1. 写 `workflow_plan_review_results`
2. 写 `workflow_compiled_plans`
3. 写 `workflow_compiled_tasks`
4. 投影 `domain_tasks`

事务外：

1. 审计记录扩展字段补齐
2. 工作台快照刷新

### 4.3 启动任务运行

事务内：

1. 更新 `domain_tasks.status`
2. 创建 `brain_run_bindings`
3. 写初始 `audit_logs`

事务外：

1. 实际请求 `brain serve`
2. 启动 run sync worker

说明：

1. 若 `brain serve` 调用失败，需要补偿更新任务状态
2. 不应在数据库事务里直接持有外部 HTTP 调用

### 4.4 完成验收裁决

事务内：

1. 写 `acceptance_runs`
2. 写 `acceptance_issues`
3. 写 `acceptance_judgements`
4. 更新项目 `production_status`

事务外：

1. evidence 详情补索引
2. Coverage 视图缓存刷新

## 5. 补偿原则

凡是“事务内写库 + 事务外调用外部系统”的动作，都需要补偿逻辑。

例如：

1. `brain serve` 创建 run 失败
2. evidence 文件存在但索引失败
3. replay 文件生成成功但 audit 写失败

补偿策略建议：

1. 标记 `pending_repair`
2. 生成审计记录
3. 允许 worker 自动重试
4. 允许用户手动触发恢复

## 6. 幂等性要求

以下动作必须幂等：

1. `CreateProject`
2. `CompilePlan`
3. `StartTaskRun`
4. `StartAcceptance`
5. `ApplyManualDecision`

原因：

1. 桌面端可能重复点击
2. worker 可能重试
3. 异常恢复后可能重新执行

## 7. 一致性优先级

优先级建议：

1. 项目主状态
2. 任务主状态
3. 验收最终裁决
4. Run 绑定关系
5. Evidence / replay 索引
6. 页面缓存

前四类必须严肃保证，后两类允许短暂滞后。

## 8. 后续细分专题

1. 事务 helper 接口设计
2. 命令幂等键设计
3. 补偿任务表设计

