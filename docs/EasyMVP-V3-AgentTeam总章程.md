# EasyMVP V3 Agent Team 总章程

> 更新时间：2026-04-19
> 目标：建立一个长期可持续的多 agent 执行团队，用于推进 EasyMVP V3，从而在中断或切换会话后仍能通过文档持续衔接。

## 1. 团队目标

Agent Team 的唯一目标是：

> 按 V3 文档体系逐步把 EasyMVP V3 从文档设计推进到代码实现，并在每次中断后可从状态板恢复。

## 2. 团队结构

建议固定为四条主线：

1. `backend-core`
2. `frontend-workbench`
3. `runtime-storage`
4. `orchestration-status`

### 2.1 `backend-core`

负责：

1. GoFrame API
2. service / repository
3. orchestrator
4. command / query handlers

### 2.2 `frontend-workbench`

负责：

1. Electron + React 壳
2. 路由
3. 页面组件
4. 状态组件

### 2.3 `runtime-storage`

负责：

1. SQLite migration
2. DAO / entity / DO
3. 文件目录落点
4. `brain-v3` 接入
5. `brain-v3` 结构化工具协议适配与运行时归一化投影

### 2.4 `orchestration-status`

负责：

1. 总计划拆分
2. 状态板更新
3. 中断恢复说明
4. 计划依赖与并行性维护

## 3. 核心规则

1. 每个计划都必须有唯一 `plan_id`
2. 每个计划都必须标注对应文档位置
3. 每个计划都必须标注是否可并行
4. 每完成一个计划就更新状态板
5. 不允许只在会话里口头推进，不落文档
6. `brain-v3` 内置 `code / browser / verifier / fault` 能力统一归 runtime lane，不得回填进 `easymvp-brain`
7. 任何涉及 `brain-v3` 工具结果的实现都必须先经过 EasyMVP 适配层归一化，再进入领域合同或页面对象

## 4. 状态定义

统一状态：

1. `pending`
2. `in_progress`
3. `blocked`
4. `done`

## 5. 文档落点

本团队依赖以下文件：

1. [EasyMVP-V3-AgentTeam总章程](./EasyMVP-V3-AgentTeam总章程.md)
2. [EasyMVP-V3-AgentTeam开机计划总表](./EasyMVP-V3-AgentTeam开机计划总表.md)
3. [EasyMVP-V3-AgentTeam状态板](./EasyMVP-V3-AgentTeam状态板.md)
4. [EasyMVP-V3-当前上下文与重启接续说明](./EasyMVP-V3-当前上下文与重启接续说明.md)

## 6. 执行原则

1. 能并行的任务优先并行
2. 阻塞主链路的任务优先级最高
3. 计划描述必须可执行，不写空话
4. 每个子团队都要保持最小写入范围
5. `tools/list` / `tools/call`、`unsupported` / `denied` 一类协议变化优先在 runtime / adapter 层吸收，不在领域脑重复实现
