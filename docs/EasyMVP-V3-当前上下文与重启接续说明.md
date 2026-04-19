# EasyMVP V3 当前上下文与重启接续说明

> 更新时间：2026-04-19
> 目标：在需要重启 Codex 或切换会话时，保留当前 V3 设计上下文、关键决策和下一步执行入口，避免重新对齐成本。

## 1. 当前项目定位

EasyMVP V3 当前正式定位为：

> 单用户、单机版、本地优先、实时可视化的项目工作台。

不是：

1. SaaS 后台
2. 多租户系统
3. 管理后台

## 2. 当前已确定的核心技术路线

当前正式技术栈已经定为：

1. 桌面壳：`Electron`
2. 前端：`React + TypeScript + Vite`
3. 本地核心服务：`Go + GoFrame v2`
4. 本地数据库：`SQLite`
5. 文件存储：`Local Files`
6. 运行时底座：`brain-v3`
7. 运行时接入方式：`brain serve HTTP API`

## 3. 当前已确定的关键架构结论

### 3.1 进程与职责

1. Electron Main 只负责桌面壳和原生能力
2. React Renderer 只负责页面与可视化
3. GoFrame v2 本地核心服务负责业务内核
4. `brain-v3` 继续作为外部运行时底座
5. `easymvp-brain` 只负责领域裁决与结构化领域结果

### 3.2 通讯边界

1. 业务 API 走本地 GoFrame HTTP 服务
2. Electron IPC 只负责桌面原生能力桥接
3. Renderer 不直接访问 SQLite
4. Renderer 不直接调用 `brain serve`
5. Renderer 和领域层都不直接消费 `brain-v3` 原始工具结果；必须先经过 EasyMVP 归一化

### 3.3 存储边界

1. `SQLite` 存结构化事实数据
2. 本地文件系统存 evidence / replay / logs / exports
3. 视图快照允许缓存，但不能替代事实表

## 4. 当前文档主线已完成范围

当前已完成首版闭环的文档层包括：

1. 总纲与总体架构
2. 方案编译
3. 分类与角色
4. 专精大脑接入
5. 生产级分类验收
6. 实时工作台与页面体系
7. 单机版存储与 SQLite
8. GoFrame v2 本地核心服务实现架构
9. API、DTO、事务、一致性、错误码、配置、恢复模式
10. `easymvp-brain` 边界、Manifest、Prompt、合同 Schema 与 `brain-v3` 协议适配语义

## 5. 当前新增且关键的实现级文档

建议重启后优先继续阅读这些：

1. [EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-技术栈与选型基线.md)
2. [EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go本地核心服务架构设计.md)
3. [EasyMVP-V3-代码目录结构与模块归属建议](./EasyMVP-V3-代码目录结构与模块归属建议.md)
4. [EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-数据库Schema总设计.md)
5. [EasyMVP-V3-首批Migration清单与建表SQL设计](./EasyMVP-V3-首批Migration清单与建表SQL设计.md)
6. [EasyMVP-V3-本地API-DTO与错误返回设计](./EasyMVP-V3-本地API-DTO与错误返回设计.md)
7. [EasyMVP-V3-Go包级接口与依赖关系设计](./EasyMVP-V3-Go包级接口与依赖关系设计.md)
8. [EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计](./EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计.md)
9. [EasyMVP-V3-恢复模式与诊断模式页面设计](./EasyMVP-V3-恢复模式与诊断模式页面设计.md)
10. [EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计](./EasyMVP-V3-easymvp-brain职责边界与输入输出合同设计.md)
11. [EasyMVP-V3-easymvp-brain-Manifest与ToolSchema设计](./EasyMVP-V3-easymvp-brain-Manifest与ToolSchema设计.md)
12. [EasyMVP-V3-easymvp-brain-Prompt设计](./EasyMVP-V3-easymvp-brain-Prompt设计.md)

## 6. 当前工具状态

已完成：

1. 安装 `goframe-v2` skill

安装位置：

1. `/root/.codex/skills/goframe-v2`

注意：

1. 需要重启 Codex 后才能识别这个新 skill

## 7. 重启后的直接下一步

重启后，不需要重新讨论架构方向。

建议直接进入：

1. 使用 `goframe-v2` skill
2. 继续 `apps/core` 中 `completion_adjudication / repair_design / workspace_explanation` 三条领域脑链路
3. 完成 `plan/acceptance` 页面联动与 replay / diagnostics 投影收口
4. 在环境允许时恢复 `apps/desktop` 依赖安装与启动验证

## 8. 重启后的建议指令

重启后可以直接说：

```text
继续 EasyMVP V3，按 docs/EasyMVP-V3-当前上下文与重启接续说明.md 执行，
使用 goframe-v2 skill，
先继续完成 completion_adjudication / repair_design / workspace_explanation，
再收口 plan/acceptance 页面联动。
```

## 9. 一句话总结

即使重启会话，仓库里的 V3 文档和这份接续说明仍然保留，所以不会丢设计结论；丢掉的只是会话缓存，不是项目上下文。
