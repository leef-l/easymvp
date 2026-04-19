# EasyMVP V3 Workspace 首页聚合接口 Schema 设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace首页设计](./EasyMVP-V3-Workspace首页设计.md)
> 关联文档：[EasyMVP-V3-Workspace首页线框图设计](./EasyMVP-V3-Workspace首页线框图设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：定义 Workspace 首页的聚合对象、返回字段、刷新边界和推荐接口结构，作为前后端共同遵守的数据契约。

## 1. 设计结论

`Workspace Home` 不应由前端自行拼装多个对象。

后端必须提供一个正式的首页聚合对象：

1. `WorkspaceHomeView`

前端首页只消费这个对象，不直接拼接：

1. 项目列表
2. 事件流
3. acceptance 状态
4. action inbox

补充边界：

1. 首页对象只能消费 EasyMVP 聚合后的页面语义字段
2. 不直接透出 `brain-v3` 原始 `tools/list` / `tools/call`、原始 `content[]` 或原始工具 payload
3. `unsupported / denied` 必须在聚合层先变成显式首页状态或待处理项语义

## 2. 根对象定义

建议首页根对象固定为：

1. `summary`
2. `running_projects`
3. `need_attention`
4. `stage_overview`
5. `recent_activity`
6. `release_readiness`
7. `quick_actions`

## 3. summary

用于顶部全局状态条。

建议字段：

1. `running_project_count`
2. `blocked_project_count`
3. `waiting_action_count`
4. `ready_for_acceptance_count`
5. `last_updated_at`

## 4. running_projects

用于首页项目卡片区。

建议为数组，每项结构如下：

### 4.1 ProjectHomeCard

1. `project_id`
2. `project_name`
3. `project_category`
4. `current_stage`
5. `overall_progress`
6. `active_task_title`
7. `active_run_id`
8. `current_run_status`
9. `blocker_count`
10. `waiting_action`
11. `production_readiness`
12. `functional_passed`
13. `production_passed`
14. `deep_link`

### 4.2 字段说明

1. `overall_progress` 建议返回 0-100 数值
2. `production_readiness` 建议返回：
   - `not_ready`
   - `in_progress`
   - `near_ready`
   - `ready`
3. `current_run_status` 应来自 EasyMVP 归一化后的运行时状态，而不是 `brain-v3` 原始状态枚举

## 5. need_attention

用于右侧待处理区。

建议为数组，每项结构如下：

### 5.1 HomeActionCard

1. `item_id`
2. `project_id`
3. `project_name`
4. `issue_type`
5. `severity`
6. `blocking`
7. `summary`
8. `recommended_action`
9. `action_button_text`
10. `action_target`
11. `created_at`

### 5.2 issue_type

建议统一枚举：

1. `review_blocker`
2. `run_failed`
3. `acceptance_blocker`
4. `manual_release_required`
5. `run_sync_failed`

## 6. stage_overview

用于阶段分布视图。

建议结构如下：

### 6.1 StageBucket

1. `stage_name`
2. `project_count`
3. `blocked_count`
4. `active_count`
5. `filter_key`

### 6.2 stage_name

建议固定枚举：

1. `Design`
2. `Review`
3. `Compile`
4. `Execute`
5. `Acceptance`
6. `Complete`

## 7. recent_activity

用于跨项目活动流。

建议为数组，每项结构如下：

### 7.1 HomeActivityEvent

1. `event_id`
2. `project_id`
3. `project_name`
4. `event_type`
5. `summary`
6. `severity`
7. `created_at`
8. `source_brain`
9. `source_run_id`
10. `deep_link`

补充说明：

1. `source_brain` 只是归一化后的来源归属字段
2. 不表示页面直接拥有底层执行脑选择能力
3. 不直接暴露 `brain-v3` 原始工具名或原始 payload

### 7.2 推荐 event_type

1. `plan_created`
2. `review_blocking_found`
3. `plan_compiled`
4. `task_started`
5. `task_failed`
6. `acceptance_blocker_found`
7. `manual_action_required`

说明：

1. `recent_activity` 只展示高价值归一化事件
2. 不直接暴露底层 run log 行、原始工具回包或内置脑原始字段

## 8. release_readiness

用于首页右下可交付进度区。

建议为数组，每项结构如下：

### 8.1 ReleaseReadinessCard

1. `project_id`
2. `project_name`
3. `functional_passed`
4. `production_passed`
5. `manual_release_required`
6. `released_by_human`
7. `blocking_reason`
8. `acceptance_link`

## 9. quick_actions

用于首页快捷入口。

建议固定：

1. `new_project`
2. `view_all_blockers`
3. `open_acceptance`
4. `open_settings`

## 10. 推荐接口

建议接口：

1. `GET /api/v3/workspace/home-view`

返回：

```json
{
  "summary": {},
  "running_projects": [],
  "need_attention": [],
  "stage_overview": [],
  "recent_activity": [],
  "release_readiness": [],
  "quick_actions": []
}
```

## 11. 刷新策略

### 11.1 全量刷新

建议首页快照接口支持短轮询。

### 11.2 增量刷新

以下区域推荐增量更新：

1. `recent_activity`
2. `need_attention`
3. `running_projects.current_run_status`

### 11.3 强制即时更新

以下变化发生时应立即刷新：

1. blocker 数变化
2. run 失败
3. manual release 触发
4. production readiness 进入 `near_ready`

## 12. 排序规则

### 12.1 running_projects

建议排序：

1. `blocked` 优先
2. `waiting_action = true` 次优先
3. 最近活跃时间优先

### 12.2 need_attention

建议排序：

1. `blocking = true`
2. `severity`
3. `created_at desc`

### 12.3 release_readiness

建议排序：

1. `production_passed = false` 但 `functional_passed = true`
2. `manual_release_required = true`
3. 最近验收更新时间

## 13. 空态约定

### 13.1 无项目

返回：

1. `running_projects = []`
2. `need_attention = []`
3. `recent_activity = []`
4. `release_readiness = []`

同时 `quick_actions` 必须包含 `new_project`。

### 13.2 有项目但无进行中项目

可允许：

1. `running_projects = []`
2. `recent_activity` 有历史项目数据

## 14. 来源映射

### 14.1 running_projects 来源

来自：

1. `ProjectSnapshot`
2. `BrainRunBinding`
3. `AcceptanceRun`

### 14.2 need_attention 来源

来自：

1. `ActionInboxItem`

### 14.3 recent_activity 来源

来自：

1. `LiveEvent`
2. `workflow_brain_run_events`

### 14.4 release_readiness 来源

来自：

1. `AcceptanceRun`
2. `AcceptanceCoverage`

## 15. 不该怎么做

不应该：

1. 返回过大的嵌套对象
2. 把底层数据库结构原样透传
3. 让前端自己算 `production_readiness`
4. 让前端自己关联 blocker 和 release 状态

## 16. 后续细分专题

本专题后续继续拆：

1. 首页 SSE 增量 payload
2. 首页筛选参数 schema
3. 首页缓存与分页策略
