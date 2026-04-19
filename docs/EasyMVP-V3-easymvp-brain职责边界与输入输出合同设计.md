# EasyMVP V3 easymvp-brain 职责边界与输入输出合同设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3专精大脑接入计划](./EasyMVP-V3专精大脑接入计划.md)
> 关联文档：[EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 目标：把 `easymvp-brain` 的职责边界、输入对象、输出对象和不可越界范围正式定义清楚。

## 1. 设计结论

`easymvp-brain` 是领域脑，不是通用执行脑。

它的职责是：

1. 审核计划
2. 编译计划
3. 重构返工方案
4. 映射验收规则
5. 裁决任务完成语义

## 2. 明确不负责的事情

`easymvp-brain` 不负责：

1. 直接做代码实现
2. 直接做浏览器采证
3. 直接替代 verifier
4. 直接控制业务状态机

## 3. 五类核心输入输出合同

### 3.1 plan review

输入：

1. `PlanDraft`
2. `CategoryProfile`

输出：

1. `PlanReviewResult`

### 3.2 plan compile

输入：

1. `PlanDraft`
2. `PlanReviewResult`
3. `CategoryProfile`

输出：

1. `CompiledPlan`
2. `CompiledTask[]`

### 3.3 repair design

输入：

1. 失败任务上下文
2. 失败原因
3. 原始合同

输出：

1. `RepairPlanDraft`

### 3.4 acceptance mapping

输入：

1. `ProjectCategory`
2. `CategoryProfile`
3. 当前项目产物摘要

输出：

1. `ProductionAcceptanceProfile`
2. 验收要求清单

### 3.5 completion adjudication

输入：

1. executor 结果
2. delivery 结果
3. verification 结果

输出：

1. `executor_succeeded`
2. `delivery_verified`
3. `completed`

## 4. 契约要求

所有输出必须满足：

1. 结构化
2. 可版本化
3. 可回溯到输入对象
4. 可被工作台解释

## 5. 工具边界

`easymvp-brain` 可以调用的工具应限于：

1. 读取计划对象
2. 读取分类 profile
3. 写入结构化评审结果
4. 写入结构化编译结果
5. 写入验收映射结果

不应直接拥有广泛文件改写能力。

## 6. 后续细分专题

本专题后续继续拆：

1. 子任务能力矩阵
2. 错误处理合同
3. 输出 JSON schema
