# EasyMVP V3 创建项目弹层交互状态图

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建项目弹层线框图设计](./EasyMVP-V3-创建项目弹层线框图设计.md)
> 关联文档：[EasyMVP-V3-创建初始化事件流接口设计](./EasyMVP-V3-创建初始化事件流接口设计.md)
> 目标：定义 `Create Project` 弹层从草稿到创建成功/失败的状态切换。

## 1. 状态图

```text
drafting
  -> validating
  -> ready_to_create
  -> creating
  -> initializing_workspace
  -> binding_profiles
  -> bootstrapping_plan
  -> created
  -> create_failed
```

## 2. 关键交互

1. 输入变化进入 `validating`
2. 校验通过进入 `ready_to_create`
3. 点击主按钮进入创建链路
4. 失败进入 `create_failed`
5. 重试回到失败步骤或 `ready_to_create`

## 3. 不该怎么做

不应该：

1. 状态跳转不可追踪
2. 失败后直接清空所有输入

