# EasyMVP V3 Plan 详细页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：把 `Plan` 页细化为可解释 V3 方案审核与编译决策的正式页面设计。

## 1. 设计结论

`Plan` 页不是任务管理页。

它的核心任务是解释：

1. 原始计划是什么
2. review 为什么拦
3. compile 为什么改
4. 最终如何落成正式任务

## 2. 页面布局

建议采用四段结构：

1. 顶部计划总览
2. 中部 Draft / Review / Compiled 三栏对照
3. 下部任务投影区
4. 底部差异与来源说明区

## 3. 顶部计划总览

建议展示：

1. `plan_version`
2. `review_decision`
3. `compiled_version`
4. `task_count`
5. `risk_summary`

## 4. 三栏对照区

### 4.1 Draft 栏

展示：

1. 目标摘要
2. 原始任务列表
3. 输入约束

### 4.2 Review 栏

展示：

1. `blocking_issues`
2. `advisory_issues`
3. `split_suggestions`
4. `override_suggestions`

### 4.3 Compiled 栏

展示：

1. 正式任务集合
2. role / brain 结果
3. delivery / verification 合同摘要

## 5. 任务投影区

### 5.1 目标

解释 `CompiledPlan -> CompiledTask -> DomainTask` 的关系。

### 5.2 每项展示

建议包含：

1. `task_name`
2. `task_kind`
3. `role_type`
4. `brain_kind`
5. `risk_level`
6. `delivery_summary`
7. `verification_summary`
8. `source_task_key`

### 5.3 用户价值

用户要能直接看懂：

1. 为什么拆成这些任务
2. 为什么这个任务归这个 brain
3. 哪些任务需要人工 review

## 6. 差异说明区

### 6.1 diff 类型

建议至少展示：

1. `task_split`
2. `task_drop`
3. `brain_override`
4. `role_override`
5. `risk_upgrade`
6. `contract_enhanced`

### 6.2 每条 diff 卡片

建议展示：

1. `before_label`
2. `after_label`
3. `reason`
4. `source_review_issue_id`

## 7. 交互规则

### 7.1 点击 review 问题

打开问题详情抽屉，显示：

1. 问题原因
2. 影响任务
3. 推荐动作

### 7.2 点击 compiled task

打开任务抽屉，显示：

1. 来源 task key
2. 解析链
3. 风险与合同

### 7.3 点击 role / brain

打开解析来源面板，显示：

1. category profile 命中规则
2. project override 是否应用
3. 风险升级是否触发

## 8. 状态与空态

### 8.1 仅有 Draft

显示“等待 review”。

### 8.2 已有 Review 未 Compile

显示“等待 compile”。

### 8.3 已有 CompiledPlan

展示完整三栏与任务投影。

## 9. 不该怎么做

不应该：

1. 只展示一个任务表
2. 把 review 结果埋进长文本
3. 不显示 compile 差异原因

## 10. 后续细分专题

本专题后续继续拆：

1. Plan 线框图
2. diff 组件规范
3. 任务投影抽屉设计
