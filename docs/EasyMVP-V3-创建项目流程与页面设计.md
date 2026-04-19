# EasyMVP V3 创建项目流程与页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 关联文档：[EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
> 关联文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3方案编译模型设计](./EasyMVP-V3方案编译模型设计.md)
> 关联文档：[EasyMVP-V3分类策略与角色自动解析设计](./EasyMVP-V3分类策略与角色自动解析设计.md)
> 关联文档：[EasyMVP-V3生产级分类验收体系设计](./EasyMVP-V3生产级分类验收体系设计.md)
> 关联文档：[EasyMVP-V3-本地目录与项目工作区规范](./EasyMVP-V3-本地目录与项目工作区规范.md)
> 关联文档：[EasyMVP-V3-创建初始化事件流接口设计](./EasyMVP-V3-创建初始化事件流接口设计.md)
> 关联文档：[EasyMVP-V3-创建项目弹层线框图设计](./EasyMVP-V3-创建项目弹层线框图设计.md)
> 关联文档：[EasyMVP-V3-路径选择与仓库检测组件规范](./EasyMVP-V3-路径选择与仓库检测组件规范.md)
> 目标：定义 V3 单机版从 `New Project` 到项目进入实时工作台的完整创建流程、页面结构、自动推导规则和失败处理方式。

## 1. 设计结论

V3 的创建项目不能做成重表单。

它必须满足 4 个约束：

1. 输入最少
2. 自动推导优先
3. 创建后立即进入工作流
4. 创建后立即可见实时状态

正式定义为：

> 用户只提供目标、项目分类和本地工作区上下文，系统自动补齐默认策略、角色、`brain_kind` 解析上下文和验收框架，然后创建项目并进入实时工作台。

## 2. 这一步为什么重要

创建项目不是单独的表单页问题。

它直接决定：

1. V3 是否真的轻量化
2. 分类策略是否真正成为正式输入
3. 角色自动解析是否真的是默认主路径
4. 实时工作台首页的 `New Project` 是否顺手
5. 项目创建后能否自然接入当前闭环主链

按当前钱学森总纲，这里的创建后主链建议理解为：

`PlanDraft -> PlanReviewResult -> CompiledPlan -> reviewing -> executing -> accepting`

如果这一步做重了，后面的工作台再简洁也没有意义。

## 3. 设计目标

创建项目页面必须同时解决这 5 个问题：

1. 用户快速开始一个项目
2. 系统拿到后续自动推导所需的最小输入
3. 创建后马上形成 `ProjectSnapshot`
4. 创建后马上能进入方案主链路
5. 创建失败时给出明确可恢复动作

## 4. 最小输入模型

V3 创建项目时，用户只应被要求填写以下正式输入：

1. `project_name`
2. `project_category`
3. `goal_prompt`
4. `workspace_source`
5. `repo_path_or_local_path`

其中：

1. `project_name` 是工作台展示名
2. `project_category` 是 V3 正式策略输入
3. `goal_prompt` 是初始目标描述
4. `workspace_source` 表示新建目录、使用已有目录或连接已有仓库
5. `repo_path_or_local_path` 是本地实际路径

## 5. 创建时不该前置暴露的内容

以下内容不应该在默认创建流程里前置暴露：

1. `role_type`
2. `brain_kind`
3. `acceptance_profile`
4. `verification_mode`
5. `risk_level`
6. `delivery_contract`
7. `manual_release_policy`
8. `tool_policy`
9. `provider/model`

这些内容属于系统自动推导结果，或者低频高级设置。

只有在高级模式中，才允许查看“系统准备怎么推导”，但也不建议在创建首屏要求用户强制修改。

## 6. 创建入口

### 6.1 默认入口

默认入口来自 `Workspace Home` 页头部主按钮：

1. `New Project`

### 6.2 次级入口

允许以下次级入口：

1. 首页空态主按钮
2. 最近项目区旁的 `Create Another`
3. 全局快捷键入口

### 6.3 创建交互形态

建议桌面端采用：

1. 居中弹层作为默认形态
2. 当进入高级步骤时扩展为右侧抽屉或全页

原因：

1. 创建项目是高频操作，但不是长期停留页面
2. 弹层能保持工作台上下文
3. 当需要展示目录检查、仓库检测和初始化进度时，再切到更大容器

## 7. 页面流程结论

V3 创建项目建议采用“单页分段 + 明显主流程按钮”的方式，而不是传统多步表单向导。

建议固定为 4 段：

1. 项目类型与目标
2. 本地工作区选择
3. 系统自动推导预览
4. 创建与初始化状态

## 8. 页面结构

### 8.1 顶部标题区

顶部必须明确告诉用户这是开始项目，而不是做系统配置。

建议展示：

1. 页面标题 `Create Project`
2. 一句副标题 `告诉系统你要做什么，其他默认自动准备`

### 8.2 第一段：项目类型与目标

建议字段如下：

1. `project_name`
2. `project_category`
3. `goal_prompt`

交互要求：

1. `project_category` 用大卡片选择，不用下拉长列表
2. `goal_prompt` 用较大的自然语言输入框
3. 根据分类动态提示不同输入示例

分类卡建议首批内置：

1. `web_app`
2. `game`
3. `video_editing`
4. `automation_tool`
5. `content_workflow`

### 8.3 第二段：本地工作区选择

建议固定三种模式：

1. `Create New Workspace`
2. `Use Existing Folder`
3. `Use Existing Repository`

每种模式都应展示路径选择器和路径检查结果。

路径检查至少包括：

1. 路径是否存在
2. 是否可读写
3. 是否已被其他项目占用
4. 是否为 git 仓库
5. 是否存在明显的大型无关目录

### 8.4 第三段：系统自动推导预览

这是 V3 区别于普通项目创建页的关键区域。

它不是让用户配置，而是让用户理解系统准备怎么启动。

建议固定展示：

1. 命中的 `CategoryProfile`
2. 预计的计划主链路
3. 默认角色解析策略
4. 默认验收方向
5. 是否需要后续人工放行

建议用 4 张解释卡片表达：

1. `Planning`
2. `Execution`
3. `Acceptance`
4. `Workspace`

每张卡片只展示简短结论，不展示复杂 schema。

### 8.5 第四段：创建与初始化状态

用户点击主按钮后，页面进入初始化态。

此区域建议显示实时进度条和事件流，而不是只显示一个 loading。

初始化至少展示：

1. 项目记录已创建
2. 本地目录已准备
3. 工作区扫描已完成
4. 分类策略已绑定
5. 初始 `PlanDraft` 已创建或待创建
6. 已进入 `Project Workspace`

补充说明：

- 进入 `Project Workspace` 时，项目仍可能处在早期 `reviewing` 引导态
- 这里不应让用户误以为项目已经进入成熟执行态

## 9. 创建流程状态机

建议把创建过程建模为正式状态机：

1. `drafting`
2. `validating`
3. `ready_to_create`
4. `creating`
5. `initializing_workspace`
6. `binding_profiles`
7. `bootstrapping_plan`
8. `created`
9. `create_failed`

这组状态是“创建状态机”，不是项目闭环状态机。

创建完成后，项目层状态建议进入：

1. 早期 `reviewing`
2. 或引导态下的 `reviewing` 准备中

### 9.1 状态说明

1. `drafting`：用户仍在填写最小输入
2. `validating`：系统检查路径、分类和必填项
3. `ready_to_create`：允许点击创建
4. `creating`：写入项目主记录
5. `initializing_workspace`：准备本地目录与索引
6. `binding_profiles`：绑定分类、角色默认策略、验收框架
7. `bootstrapping_plan`：创建初始计划上下文
8. `created`：项目创建完成
9. `create_failed`：创建链路失败，等待恢复

## 10. 创建后生成的正式对象

项目创建成功后，系统至少要产生以下对象：

1. `Project`
2. `ProjectWorkspaceBinding`
3. `ProjectCategory`
4. `CategoryProfileBinding`
5. `ProjectSnapshot`
6. `StageProgress`
7. `ActionInbox` 初始集合
8. 初始 `PlanDraft` 或 `PlanBootstrapIntent`

其中关键点是：

1. 创建项目并不等于直接生成完整 `CompiledPlan`
2. 但必须足够让工作台立刻可展示
3. 创建后用户应立刻看到项目已进入哪个初始阶段

## 11. 系统自动推导规则

### 11.1 分类推导

`project_category` 是唯一必须明确由用户输入的顶层策略项。

系统基于分类自动推导：

1. 默认 `CategoryProfile`
2. 默认 `ProductionAcceptanceProfile`
3. 默认 `surface` 集合
4. 默认 `journey` 基线

### 11.2 角色与 `brain_kind` 解析上下文推导

创建阶段不直接生成最终每个任务的 `role_type / brain_kind`。

但应生成初始解析上下文：

1. `category_default_role_policy`
2. `project_execution_preference`
3. `runtime_capability_snapshot`

这些上下文会在 `PlanReview` 与方案编译阶段继续细化。

这里的“推导”仅指生成后续 `RoleResolver / compiler` 可消费的 hint。

它不表示创建阶段已经决定最终任务级执行归属，也不表示 `easymvp-brain` 在创建链路中接管 `code / browser / verifier / fault` 一类执行能力。

### 11.3 工作区推导

系统根据本地目录情况推导：

1. 是否已有 git 仓库
2. 初始 workspace 根目录
3. 是否需要新建子目录
4. 后续扫描范围
5. 基本文件忽略策略

## 12. 页面交互原则

### 12.1 单屏完成优先

用户最好在一个页面或一个弹层里完成创建，而不是来回切多页。

### 12.2 解释优先于配置

自动推导结果要解释清楚，但不要求用户逐项配置。

### 12.3 主按钮始终明确

主按钮建议使用：

1. `Create Project`

在初始化过程中切换为：

1. `Creating...`

禁止同时放多个不清晰的主操作。

### 12.4 失败可恢复

失败后必须允许：

1. 重新检查路径
2. 修改输入后重试
3. 查看失败原因
4. 清理半完成创建结果

## 13. 校验与失败处理

### 13.1 提交前校验

至少检查：

1. 项目名不能为空
2. 分类不能为空
3. 目标描述不能为空
4. 本地路径不能为空
5. 路径必须可访问

### 13.2 初始化失败类型

建议至少分为：

1. `path_not_accessible`
2. `workspace_conflict`
3. `runtime_not_ready`
4. `profile_binding_failed`
5. `plan_bootstrap_failed`

### 13.3 失败展示

失败时不应该只显示技术报错字符串。

建议固定展示：

1. 失败步骤
2. 失败原因摘要
3. 建议修复动作
4. 是否可重试
5. 是否已回滚部分创建结果

## 14. 创建成功后的跳转规则

### 14.1 默认跳转

创建成功后，默认进入新项目的 `Project Workspace`。

原因：

1. 用户此时最关心项目是否真的开始了
2. `Project Workspace` 能立刻显示初始化事件和当前阶段

### 14.2 首次进入的首屏规则

首次进入项目后，工作台首屏建议固定突出：

1. `当前阶段`
2. `正在准备计划`
3. `最近初始化事件`
4. `下一步建议动作`

### 14.3 二级跳转

如果初始化完成且已生成初始 `PlanDraft`，则允许从工作台直接点击进入 `Plan`。

## 15. 与 Workspace 首页的关系

`Workspace Home` 需要明确为创建项目留出稳定位置。

首页层面至少要支持：

1. Header 主按钮 `New Project`
2. 无项目空态主按钮
3. 刚创建项目后立即出现在 `Running Projects` 或 `Recent Projects`

也就是说：

创建项目不是离开工作台的“后台动作”，而是工作台主链路的一部分。

## 16. 与 V3 主架构的关系

创建项目必须显式对齐 V3 现有主架构，而不是另起一条产品逻辑。

对齐关系如下：

1. 创建页负责收集 `ProjectCategory` 和初始目标
2. 创建成功后产生项目上下文和初始快照
3. 计划主链路从初始上下文进入 `PlanDraft`
4. `PlanDraft` 后续继续进入 `PlanReviewResult`
5. `CompiledPlan` 再推动执行与验收

因此创建页只是主链路入口，不是旁路。

## 17. 页面线框建议

建议桌面端结构如下：

```text
┌───────────────────────────────────────────────────────────────────┐
│ Create Project                                                    │
│ Tell EasyMVP what you want to build                              │
├───────────────────────────────────────────────────────────────────┤
│ 1. Project Type & Goal                                            │
│ [Name] [Category Cards]                                          │
│ [Goal Prompt...................................................]  │
├───────────────────────────────────────────────────────────────────┤
│ 2. Workspace Source                                               │
│ [New Workspace] [Existing Folder] [Existing Repository]          │
│ [Path Picker...................................................]  │
│ [Path Check Result]                                               │
├───────────────────────────────────────────────────────────────────┤
│ 3. System Will Prepare                                            │
│ [Planning Card] [Execution Card] [Acceptance Card] [Workspace]   │
├───────────────────────────────────────────────────────────────────┤
│ Footer                                                            │
│ [Cancel]                                      [Create Project]    │
└───────────────────────────────────────────────────────────────────┘
```

初始化中则切换为：

```text
┌───────────────────────────────────────────────────────────────────┐
│ Creating Project                                                  │
├───────────────────────────────────────────────────────────────────┤
│ Step Progress                                                     │
│ create record -> init workspace -> bind profiles -> bootstrap    │
├───────────────────────────────────────────────────────────────────┤
│ Live Init Events                                                  │
│ time · step · summary                                             │
├───────────────────────────────────────────────────────────────────┤
│ [Open Workspace When Ready]                                       │
└───────────────────────────────────────────────────────────────────┘
```

## 18. 文案原则

创建页文案必须显得直接、可理解、无后台味。

建议：

1. 用 `What do you want to build?`
2. 用 `Where is the project workspace?`
3. 用 `EasyMVP will prepare these defaults`

避免：

1. `配置项目策略`
2. `绑定执行器`
3. `设置验收规则`
4. `配置运行参数`

## 19. 不该怎么做

创建项目页不应该：

1. 一上来展示十几个高级配置项
2. 要用户先选 role 或 brain
3. 先建项目再进入另一个页面继续补完必填项
4. 创建成功后看不到实时初始化过程
5. 创建完成后还要手工再启动计划链路

## 20. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-创建项目弹层线框图设计](./EasyMVP-V3-创建项目弹层线框图设计.md)
2. [EasyMVP-V3-路径选择与仓库检测组件规范](./EasyMVP-V3-路径选择与仓库检测组件规范.md)
3. [EasyMVP-V3-创建初始化事件流接口设计](./EasyMVP-V3-创建初始化事件流接口设计.md)
4. [EasyMVP-V3-创建失败恢复与回滚策略设计](./EasyMVP-V3-创建失败恢复与回滚策略设计.md)
5. [EasyMVP-V3-创建后首次进入Project-Workspace引导态设计](./EasyMVP-V3-创建后首次进入Project-Workspace引导态设计.md)
