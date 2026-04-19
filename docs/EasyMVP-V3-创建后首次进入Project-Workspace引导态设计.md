# EasyMVP V3 创建后首次进入 Project Workspace 引导态设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 关联文档：[EasyMVP-V3-创建初始化事件流接口设计](./EasyMVP-V3-创建初始化事件流接口设计.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：定义项目刚创建完成后第一次进入 `Project Workspace` 时的页面引导态、模块降级规则、首屏提示与下一步动作设计。

## 1. 设计结论

新项目第一次进入 `Project Workspace` 时，页面不能直接表现得像一个“完整运行中的项目”。

它必须进入一个正式的引导态：

1. 明确告诉用户项目刚创建成功
2. 明确告诉用户系统正在准备什么
3. 明确告诉用户下一步应该去哪里
4. 避免展示大量暂时为空的成熟期模块

一句话定义：

> 首次进入 `Project Workspace` 时，页面应优先扮演“项目启动引导页”，而不是“常规运行驾驶舱”。

## 2. 为什么需要单独设计引导态

如果新项目刚创建后直接落到普通工作台，会出现 4 个问题：

1. 很多模块还没有数据
2. 用户不知道项目是否真的开始了
3. 用户不知道下一步应该去 `Plan` 还是继续留在 `Workspace`
4. 空白模块会让产品显得像“坏了”或“还没做完”

所以首次进入必须特殊处理。

## 3. 触发条件

建议以下条件同时满足时进入首次引导态：

1. 项目创建时间处于最近窗口内
2. `current_stage` 位于早期 `reviewing`
3. 尚未产生稳定的执行阶段活动
4. 用户尚未手动关闭首次引导

建议增加正式视图标记：

1. `is_first_entry_guided`

## 4. 引导态目标

首次进入的页面必须回答这 4 个问题：

1. 项目已经创建成功了吗
2. 系统现在正在准备什么
3. 我现在应该先看哪里
4. 什么时候会进入正常工作台状态

## 5. 页面结构结论

首次进入 `Project Workspace` 时，页面仍保留原有四大区域，但优先级和视觉表现发生变化。

建议调整为：

1. 顶部成功与准备状态条
2. 中部启动引导卡
3. 左侧简化阶段流
4. 中央初始化事件流
5. 右侧下一步动作面板
6. 底部验收区降级为“后续将自动准备”

## 6. 顶部状态条设计

### 6.1 目标

先建立确定感。

### 6.2 建议文案

建议顶部主文案使用：

1. `Project created`
2. `EasyMVP is preparing the first review cycle`

### 6.3 建议字段

建议突出：

1. `project_name`
2. `project_category`
3. `current_stage`
4. `creation_status`
5. `initialization_progress`

### 6.4 状态表现

首次引导态建议使用：

1. 主色进行中条
2. 成功创建提示
3. 较弱化的整体进度百分比

因为此时最重要的不是项目总体进度，而是“启动已开始”。

## 7. 启动引导卡

### 7.1 位置

建议放在首屏最显眼区域，位于顶部状态条下方，横跨主内容区。

### 7.2 目标

告诉用户：

1. 系统已经完成什么
2. 接下来马上会做什么
3. 你可以现在做什么

### 7.3 建议内容

建议卡片固定展示：

1. `项目已创建`
2. `工作区已绑定`
3. `分类策略已命中`
4. `首次 review/compile 正在准备`

### 7.4 主按钮

建议只保留一个主按钮：

1. `Open Plan`

次级按钮可以是：

1. `Stay Here`
2. `View Init Events`

## 8. 阶段流降级规则

首次进入时，阶段流不应表现成完整成熟流程。

建议：

1. 仅高亮 `reviewing`
2. `executing / accepting / reworking / completed` 使用弱化态
3. 阶段说明改为“即将进入”

### 8.1 阶段卡提示

例如：

1. `reviewing`：正在准备首轮计划审核与编译
2. `executing`：将在首轮计划收口后进入
3. `accepting`：待执行与验证产出形成后进入

## 9. 初始化事件流设计

### 9.1 目标

把刚创建时的系统动作具体化。

### 9.2 数据来源

直接复用：

1. [EasyMVP-V3-创建初始化事件流接口设计](./EasyMVP-V3-创建初始化事件流接口设计.md)

### 9.3 事件显示优先级

首次进入时，`Live Activity` 优先显示：

1. `creation.project_bound`
2. `creation.workspace_ready`
3. `creation.plan_bootstrapped`
4. `creation.completed`

常规 run 事件可以延后。

### 9.4 表现方式

建议事件流前 3 条默认展开说明。

原因：

1. 新用户最需要理解系统现在在干嘛
2. 这比一串简短日志更有解释力

## 10. 右侧下一步动作面板

### 10.1 目标

首次进入时，右侧面板不应主要展示 blocker，而应主要展示推荐下一步。

### 10.2 建议固定项

建议按顺序展示：

1. `Review the first plan`
2. `Confirm project goal`
3. `Check workspace path`

### 10.3 动作形式

每项都建议包含：

1. 简短说明
2. 单一动作按钮
3. 是否阻塞主流程

## 11. Verification 区域降级

首次进入时，验证/验收区不应制造“为什么这里全是空”的感觉。

建议改成准备态摘要卡，而不是完整覆盖矩阵。

建议展示：

1. `Verification requirements will be prepared from project category`
2. `Coverage details will appear after the first review/compile cycle`

这可以避免页面底部出现大片空白。

## 12. 默认跳转策略

### 12.1 创建完成后

默认还是先进入 `Project Workspace`，而不是直接进 `Plan`。

原因：

1. 先让用户确认项目已启动成功
2. 先建立对系统动作的感知
3. 再让用户进入计划解释页

### 12.2 自动引导到 Plan

若满足以下条件，可在页面中突出引导去 `Plan`：

1. 初始 `PlanDraft` 已存在
2. 初始化事件已稳定
3. 当前没有创建失败或路径类问题

## 13. 退出引导态的条件

引导态不应长期存在。

建议以下任一条件满足后退出：

1. 用户主动关闭引导
2. 用户进入过 `Plan` 并返回
3. 项目已进入稳定 `reviewing` 阶段
4. 项目已出现正式执行活动

退出后恢复常规 `Project Workspace` 视图。

## 14. 首次进入视图对象建议

建议在 `ProjectWorkspaceView` 之上增加引导态聚合字段：

1. `is_first_entry_guided`
2. `initialization_progress`
3. `startup_summary`
4. `recommended_next_actions`
5. `plan_ready`

如后续对象口径继续收口，建议补充：

6. `first_verification_prep_summary`

这样前端无需自己猜测是否要展示引导态。

## 15. 建议接口字段

建议在单项目工作台快照接口中补充：

```json
{
  "project_id": "proj_01",
  "is_first_entry_guided": true,
  "initialization_progress": 78,
  "startup_summary": "项目已创建，正在准备初始计划",
  "recommended_next_actions": [
    {
      "label": "Open Plan",
      "target": "/projects/proj_01/plan",
      "primary": true
    }
  ],
  "plan_ready": true
}
```

## 16. 文案原则

首次进入文案必须让用户觉得：

1. 系统已经开始工作
2. 我知道下一步在哪
3. 现在页面是刻意简化，不是缺功能

建议优先用：

1. `Your project is ready`
2. `EasyMVP is preparing the first review cycle`
3. `Open Plan to review how the system will structure the work`

避免使用：

1. `No data`
2. `Empty state`
3. `Not available`
4. `Pending initialization` 这类过于技术化的文案

## 17. 不该怎么做

首次进入引导态不应该：

1. 一打开就是满屏空模块
2. 直接展示完整验收矩阵
3. 让用户自己猜下一步看哪里
4. 把创建初始化事件藏到次级页面
5. 把新项目当成成熟项目直接渲染

## 18. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-首次进入引导卡组件规范](./EasyMVP-V3-首次进入引导卡组件规范.md)
2. [EasyMVP-V3-首次进入工作台线框图设计](./EasyMVP-V3-首次进入工作台线框图设计.md)
3. [EasyMVP-V3-首次进入与常规工作台切换规则](./EasyMVP-V3-首次进入与常规工作台切换规则.md)
4. [EasyMVP-V3-首次进入推荐动作生成规则](./EasyMVP-V3-首次进入推荐动作生成规则.md)
