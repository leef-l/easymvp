# EasyMVP V3 Project Workspace Top Status Bar 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-Project-Workspace线框图设计](./EasyMVP-V3-Project-Workspace线框图设计.md)
> 目标：定义单项目工作台顶部状态条的字段、状态样式、首次进入特殊表现与交互规则。

## 1. 设计结论

`Top Status Bar` 是单项目工作台的 3 秒总览。

用户进入项目后，必须先从这里知道：

1. 这是什么项目
2. 现在在哪个阶段
3. 当前整体状态如何
4. 离交付还有多远

按当前总纲，这里的“离交付还有多远”不能只用 readiness 文案表达，还必须能看出是否存在人工检查点和升级对象。

## 2. 字段

建议固定：

1. `project_name`
2. `project_category`
3. `current_stage`
4. `overall_progress`
5. `current_run_status`
6. `production_readiness`
7. `final_status_hint`
8. `manual_checkpoint_required`
9. `has_runtime_escalation`

## 3. 状态

建议四态：

1. `idle`
2. `running`
3. `blocked`
4. `ready_to_release`

补充说明：

- 这里的状态条只做顶部概览
- 最终是否 `completed` 仍以 `CompletionVerdict` 为准，不由状态条自行定义

## 4. 首次进入特殊规则

首次进入时：

1. 主文案突出 `Project created`
2. 副文案突出 `preparing first review cycle`
3. 整体进度弱化

## 5. 交互

允许点击：

1. `current_stage`
2. `current_run_status`
3. `production_readiness`

分别跳向：

1. `Plan`
2. 活跃 run 抽屉
3. `Acceptance`

## 6. 不该怎么做

不应该：

1. 只显示项目名
2. 没有当前阶段
3. 所有状态都堆成小 badge

## 7. 后续细分专题

1. Top Status Bar 视觉稿
2. 状态色映射表
