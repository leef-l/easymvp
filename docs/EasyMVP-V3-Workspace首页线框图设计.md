# EasyMVP V3 Workspace 首页线框图设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
> 关联文档：[EasyMVP-V3-Workspace首页多项目卡片组件规范](./EasyMVP-V3-Workspace首页多项目卡片组件规范.md)
> 目标：把 Workspace 首页推进到线框图级别，明确页面结构、卡片排序、首屏优先级和视觉层次。

## 1. 设计结论

`Workspace Home` 必须采用“多项目总览 + 待处理聚焦 + 交付进度前置”的首屏结构。

首页不能做成：

1. 统计页
2. 菜单页
3. 纯项目列表页

首页必须做成“工作台首页”。

## 2. 桌面端线框结构

建议桌面端采用 4 段布局：

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Header: Logo / Workspace / Search / New Project / Settings         │
├──────────────────────────────────────────────────────────────────────┤
│ Global Summary Strip                                                │
│ Running | Blocked | Need Action | Ready For Acceptance              │
├───────────────────────────────────────┬──────────────────────────────┤
│ Running Projects                      │ Need Attention              │
│ project cards                         │ action cards                │
│                                       │                              │
├───────────────────────────────────────┴──────────────────────────────┤
│ Stage Overview                                                     │
│ Design | Review | Compile | Execute | Acceptance | Complete        │
├───────────────────────────────────────┬──────────────────────────────┤
│ Recent Activity                       │ Release Readiness           │
│ cross-project live feed               │ near-ready projects         │
└───────────────────────────────────────┴──────────────────────────────┘
```

## 3. 首屏模块优先级

首屏模块排序建议固定为：

1. `Global Summary Strip`
2. `Running Projects`
3. `Need Attention`
4. `Stage Overview`
5. `Recent Activity`
6. `Release Readiness`

排序原则：

1. 先看全局状态
2. 再看正在跑的项目
3. 再看需要处理的问题
4. 最后看补充信息

## 4. Header 设计

### 4.1 左侧

建议包含：

1. 产品标识
2. `Workspace` 标题

### 4.2 中部

建议包含：

1. 搜索框
2. 项目筛选入口

### 4.3 右侧

建议固定：

1. `New Project`
2. `Settings`

### 4.4 Header 原则

1. 不放复杂导航树
2. 不放冗余统计
3. 主按钮 `New Project` 永远显眼

## 5. Global Summary Strip 线框

建议做成 4 张横向 summary 卡：

1. `Running Projects`
2. `Blocked Projects`
3. `Need Action`
4. `Ready For Acceptance`

每张卡只放：

1. 标题
2. 数字
3. 简短说明

不放：

1. 长描述
2. 复杂图表
3. 次级操作

## 6. Running Projects 区线框

### 6.1 布局

建议每行 2-3 张项目卡。

### 6.2 单张项目卡结构

```text
┌──────────────────────────────────────┐
│ Project Name             Category    │
│ Stage Badge             Progress %   │
│ Active: current task/run             │
│ Blockers: n   Need Action: yes/no    │
│ Acceptance: partial / near-ready     │
│ [Open Project] [Acceptance]          │
└──────────────────────────────────────┘
```

### 6.3 卡片层次

最醒目顺序建议：

1. 项目名
2. 当前阶段
3. 当前活跃项
4. 阻塞状态
5. Acceptance 进度

### 6.4 状态表现

建议：

1. `blocked` 项目卡高亮告警
2. `running` 项目卡显示活动状态
3. `near-ready` 项目卡突出可交付感

## 7. Need Attention 区线框

### 7.1 布局

右侧纵向堆叠 action 卡片。

### 7.2 单张 action 卡

```text
┌──────────────────────────────┐
│ Project Name                 │
│ issue type / severity        │
│ short summary                │
│ [Resolve] [Open Project]     │
└──────────────────────────────┘
```

### 7.3 视觉优先级

1. `blocking` 卡片置顶
2. `manual_release_required` 次置顶
3. 普通 warning 最后

## 8. Stage Overview 线框

### 8.1 目标

让用户一眼知道项目都卡在哪个阶段。

### 8.2 布局

建议用横向阶段条，每段带数量：

```text
Design(2) Review(3) Compile(1) Execute(4) Acceptance(2) Complete(5)
```

### 8.3 交互

点击阶段后：

1. 筛选 Running Projects
2. 同步筛选 Recent Activity

## 9. Recent Activity 线框

### 9.1 布局

左下半区使用跨项目活动流。

### 9.2 单条活动结构

```text
time · project · event summary · [Open]
```

### 9.3 展示规则

只展示高价值事件：

1. plan compiled
2. task failed
3. acceptance blocker
4. manual action required

## 10. Release Readiness 线框

### 10.1 布局

右下半区使用小卡片列表。

### 10.2 单张 readiness 卡

```text
┌──────────────────────────────┐
│ Project Name                 │
│ functional: yes/no           │
│ production: yes/no           │
│ manual release: yes/no       │
│ [Open Acceptance]            │
└──────────────────────────────┘
```

### 10.3 目的

首页直接建立“离交付多远”的感觉，而不是只显示项目是否在跑。

## 11. 空态设计

### 11.1 无项目

页面中间只保留：

1. 空态说明
2. `New Project` 主按钮

### 11.2 有项目但无进行中项目

页面应突出：

1. 最近项目
2. 已完成项目
3. 创建新项目按钮

## 12. 响应式降级

### 12.1 平板

建议：

1. `Running Projects` 单列
2. `Need Attention` 下移

### 12.2 手机

建议：

1. 先展示 `Need Attention`
2. 再展示 `Running Projects`
3. 最后展示 `Recent Activity`

## 13. 视觉风格要求

首页必须满足：

1. 卡片大
2. 留白足
3. 文案短
4. 状态清楚

不允许：

1. 大表格
2. 密集字段墙
3. 太多装饰元素

## 14. 后续细分专题

本专题后续继续拆：

1. 首页组件级线框
2. 首页卡片状态图
3. 响应式布局图
