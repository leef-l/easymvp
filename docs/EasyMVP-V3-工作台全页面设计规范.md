# EasyMVP V3 工作台全页面设计规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 关联文档：[EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
> 关联文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 目标：统一 V3 工作台所有页面的设计原则、信息架构、视觉规则和交互规则，确保所有页面都做到“简单、大气、明了，一看就懂，一看就会用”。

## 1. 设计结论

V3 工作台不能继续采用后台式页面设计。

所有页面必须满足 4 个标准：

1. 简单
2. 大气
3. 明了
4. 易上手

这 4 个标准不是视觉偏好，而是正式设计约束。

## 2. 总体页面目标

V3 工作台不是“管理系统”，而是“单用户项目工作台”。

所以页面的核心任务不是展示尽可能多的信息，而是帮助用户快速完成这 4 类行为：

1. 看清当前状态
2. 理解系统决策
3. 处理当前问题
4. 推动项目进入下一步

## 3. 全局设计原则

### 3.1 一页只回答一个核心问题

每个页面必须能用一句话定义：

1. `Workspace`：现在发生了什么
2. `Plan`：系统为什么这么决策
3. `Execution`：现在执行到哪了
4. `Acceptance`：现在能不能交付
5. `Settings`：系统怎么配置

任何页面如果同时试图回答多个大问题，就会失焦。

### 3.2 首屏只放最高价值信息

页面首屏只允许展示：

1. 当前状态
2. 当前问题
3. 当前动作

不允许把低频配置、边缘统计和系统内部字段堆在首屏。

### 3.3 图优先于表

默认优先展示：

1. 阶段流
2. 时间线
3. 覆盖矩阵
4. 状态卡片
5. 证据卡片

表格只能作为次级视图出现。

### 3.4 默认自动推导，用户只确认关键项

用户不应该在主流程里配置：

1. brain
2. role
3. acceptance profile
4. verification mode

这些应由系统推导，用户只在必要时修正或确认。

按当前钱学森总纲，这里的验证语义建议进一步收口为：

1. 系统推导 `brain_kind`
2. 系统推导 `preferred_verification_channel`
3. 系统推导 `manual_review_required`

### 3.5 术语必须面向用户理解

页面文案优先使用用户能直接理解的表达，而不是系统内部术语原样暴露。

例如：

1. `Waiting for review` 优于 `review_pending`
2. `Needs your decision` 优于 `manual_action_required`
3. `Ready for release` 优于 `production_passed=true`

但要补一条边界：

- 面向用户的自然语言不等于可以偷换业务语义，`run success / acceptance passed / completed` 仍必须区分

内部字段可以在详情或开发文档里保留，但不能主导界面语言。

## 4. 顶层导航规范

V3 顶层导航建议固定为 5 页：

1. `Workspace`
2. `Plan`
3. `Execution`
4. `Acceptance`
5. `Settings`

补充说明：

- `Workspace` 内部可以继续分为 `Workspace Home` 与 `Project Workspace`
- 当前主入口优先级仍是单项目闭环驾驶舱，不是多项目首页

### 4.1 导航原则

1. 不超过 5 个主入口
2. 不使用深层菜单树
3. 默认从 `Workspace` 进入

### 4.2 不该出现的导航设计

不应该出现：

1. 多层左侧菜单
2. 大量“管理页”入口
3. 按数据库对象命名的导航

## 5. 页面清单与核心问题

### 5.1 Workspace

核心问题：

1. 现在有哪些项目在跑
2. 哪些项目卡住了
3. 哪些项目等我处理

### 5.1.1 Workspace 二层结构

`Workspace` 必须拆成：

1. `Workspace Home`
2. `Project Workspace`

其中：

1. `Workspace Home` 是多项目壳层总览
2. `Project Workspace` 是单项目实时驾驶舱
3. `Create Project` 是 `Workspace Home` 的主入口动作

并且要明确：

- 真正承载闭环对象解释的是 `Project Workspace`
- `Workspace Home` 只做入口与分流，不承担最终裁决解释职责

### 5.2 Plan

核心问题：

1. 原始计划是什么
2. review 为什么拦
3. compile 为什么改

### 5.3 Execution

核心问题：

1. 哪些任务正在执行
2. 哪些 run 失败了
3. 哪些任务需要人工介入

### 5.4 Acceptance

核心问题：

1. 当前覆盖了哪些 surface/journey
2. 还缺哪些证据
3. 现在能不能真正 `completed`

补充说明：

- `Acceptance` 页应同时展示 `VerificationResult` 与 `CompletionVerdict`
- 不能继续只用 `production_passed` 作为页面主锚点

### 5.5 Settings

核心问题：

1. 系统运行依赖怎么配置
2. 默认策略怎么调整
3. 高级调试项放在哪里

## 6. 首屏信息优先级规范

每个页面首屏都必须按以下优先级排列：

1. 当前状态
2. 当前阻塞
3. 当前动作
4. 补充解释
5. 低频详情

如果页面首屏一打开先看到的是：

1. 统计数字
2. 大表格
3. 系统配置

那就说明页面设计偏离了 V3 的目标。

## 7. 页面布局规范

### 7.1 卡片优先

所有主页面优先使用大卡片布局，而不是表单墙或表格墙。

### 7.2 留白优先

页面要有明确呼吸感，不应堆满。

### 7.3 强主次关系

主信息必须一眼最突出，次信息自然靠后。

### 7.4 固定关键区

以下内容建议固定位置：

1. 页面标题
2. 当前状态
3. 主操作按钮
4. 当前 blocker

## 8. 视觉规范

### 8.1 色彩

建议：

1. 浅色底为主
2. 强调色控制在 1-2 个
3. 状态色固定

状态色建议：

1. 中性灰：未开始/待处理
2. 主色：进行中
3. 绿色：通过/完成
4. 橙色或红色：阻塞/失败

### 8.2 边框和阴影

建议：

1. 少边框
2. 轻阴影
3. 通过留白和层级而不是重装饰区分模块

### 8.3 字体和排版

建议：

1. 标题大而清晰
2. 正文简洁
3. 每块文字尽量短
4. 长解释放抽屉或详情层

## 9. 交互规范

### 9.1 主操作必须显眼

每页只允许一个主按钮最突出。

例如：

1. `Continue`
2. `Review blockers`
3. `Open acceptance`
4. `Release`

但要避免：

- 在未满足完成条件时，用 `Release` 把用户误导成“已经完成”

### 9.2 深层信息用抽屉，不轻易跳新页面

建议优先使用：

1. 抽屉
2. 侧边详情
3. 局部展开

减少频繁页面跳转。

### 9.3 错误和阻塞必须可操作

错误提示不能只告诉用户“失败了”，必须同时给出：

1. 原因摘要
2. 推荐动作
3. 入口按钮

## 10. 页面具体规范

### 10.1 Workspace Home

首屏必须先看到：

1. 进行中的项目
2. 各项目当前阶段
3. 各项目 blocker
4. 各项目待处理事项
5. 快接近交付的项目

### 10.2 Project Workspace

首屏必须先看到：

1. 当前阶段
2. 当前活跃 run
3. 当前 blocker
4. 当前 acceptance progress

### 10.3 Plan

首屏必须先看到：

1. 当前 plan 版本
2. review 决策
3. compile 差异摘要

### 10.4 Execution

首屏必须先看到：

1. 活跃任务
2. 失败任务
3. 当前 run 状态

### 10.5 Acceptance

首屏必须先看到：

1. `functional_passed`
2. `production_passed`
3. 缺失证据
4. 是否需要人工放行

### 10.6 Settings

首屏必须先看到：

1. 本地工作目录
2. brain/provider 配置
3. 调试与高级开关入口

## 11. 新建项目页规范

创建项目必须符合“轻启动”原则。

首屏创建表单只允许包含最小必要字段：

1. 项目名称
2. 项目分类
3. 目标摘要
4. 本地工作目录

不允许一开始就要求用户配置：

1. role
2. brain
3. verifier
4. acceptance 规则

## 12. 页面跳转规范

推荐跳转关系：

1. `Workspace -> Plan`
2. `Workspace -> Execution`
3. `Workspace -> Acceptance`
4. `Plan -> Workspace`
5. `Acceptance -> Workspace`

不推荐：

1. 任意页面跳去独立对象 CRUD 页
2. 通过很多 tab 才找到关键状态

## 13. 可用性验收标准

设计是否合格，至少看以下问题是否能 5 秒内回答：

1. 当前项目在哪个阶段
2. 当前最重要的问题是什么
3. 下一步该点哪里
4. 现在能不能交付

如果不能，页面就还不够清晰。

## 14. 与现有页面文档的关系

本规范是总规范。

以下页面文档都应服从它：

1. [EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
2. [EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
3. [EasyMVP-V3-Plan详细页面设计](./EasyMVP-V3-Plan详细页面设计.md)
4. [EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)

如果具体页面设计和本规范冲突，应优先回到本规范修订。

## 15. 后续细分专题

本专题后续继续拆：

1. 页面级线框图
2. 组件级设计规范
3. 状态色与卡片组件规范
