# EasyMVP V3 首次进入推荐动作生成规则

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建后首次进入Project-Workspace引导态设计](./EasyMVP-V3-创建后首次进入Project-Workspace引导态设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 目标：定义首次进入工作台时右侧推荐动作如何生成、排序和降级。

## 1. 推荐动作池

建议首批：

1. `Open Plan`
2. `Confirm project goal`
3. `Check workspace path`
4. `Review initialization events`

这些动作现在应被理解为：

1. 服务于早期 `reviewing` 引导态
2. 用来帮助用户理解“项目已启动，正在准备首轮 review/compile”
3. 不是把用户直接推入成熟执行态的快捷入口

## 2. 排序

建议：

1. 若 `plan_ready=true`，`Open Plan` 第一
2. 若路径有 warning，`Check workspace path` 上升
3. 若初始化失败，恢复类动作置顶

补充规则：

4. 若当前仍处于早期 `reviewing` 引导态，不应优先推荐执行态或验收态动作

## 3. 输出模型

建议：

1. `label`
2. `target`
3. `primary`
4. `blocking`

## 4. 不该怎么做

不应该：

1. 推荐动作超过 3 个主项
2. 全部都是一样的固定文案
