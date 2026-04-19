# EasyMVP V3 单机版导航与页面跳转规则

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 关联文档：[EasyMVP-V3-工作台全页面设计规范](./EasyMVP-V3-工作台全页面设计规范.md)
> 目标：定义单机版 V3 顶层导航、页面跳转入口和跨页面深链规则。

## 1. 顶层导航

固定为：

1. `Workspace`
2. `Plan`
3. `Acceptance`
4. `Settings`

## 2. 主跳转规则

1. 首页项目卡进入 `Project Workspace`
2. `current_stage` 跳向对应页面
3. `production_readiness` 跳向 `Acceptance`
4. `Open Plan` 跳向 `Plan`

## 3. 深链规则

建议所有高价值对象支持深链：

1. `project_id`
2. `run_id`
3. `acceptance_run_id`
4. `evidence_id`
5. `audit_record_id`

## 4. 不该怎么做

不应该：

1. 依赖多层菜单树
2. 同一个对象在不同页面没有稳定入口

