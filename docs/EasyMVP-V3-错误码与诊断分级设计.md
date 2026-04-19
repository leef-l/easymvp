# EasyMVP V3 错误码与诊断分级设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go本地核心服务架构设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 关联文档：[EasyMVP-V3-审计列表页面设计](./EasyMVP-V3-审计列表页面设计.md)
> 目标：为 V3 的本地 Go 服务、工作台前端和诊断页定义统一错误分级、错误码空间和用户可见恢复提示。

## 1. 设计结论

V3 的错误不能只靠自然语言字符串。

必须统一输出：

1. `error_code`
2. `error_level`
3. `error_scope`
4. `user_message`
5. `debug_message`
6. `recovery_hint`

## 2. 错误分级

建议分为：

1. `info`
2. `warning`
3. `error`
4. `critical`

## 3. 错误域

建议至少包括：

1. `startup`
2. `storage`
3. `project_creation`
4. `plan_compile`
5. `runtime`
6. `acceptance`
7. `worker`
8. `desktop_bridge`

## 4. 错误码建议

### 4.1 启动类

1. `BOOT_001` 数据根目录不可写
2. `BOOT_002` Go 核心服务启动失败
3. `BOOT_003` 本地端口占用

### 4.2 存储类

1. `STO_001` SQLite 打开失败
2. `STO_002` Migration 执行失败
3. `STO_003` 外键校验失败

### 4.3 创建项目类

1. `PRJ_001` 项目路径无效
2. `PRJ_002` 工作区初始化失败
3. `PRJ_003` 项目记录写入失败

### 4.4 计划类

1. `PLAN_001` PlanDraft 缺失
2. `PLAN_002` Review 阻塞未解除
3. `PLAN_003` Compile 输出无有效任务

### 4.5 运行时类

1. `RUN_001` `brain serve` 不可用
2. `RUN_002` 创建 run 失败
3. `RUN_003` run 状态同步失败

### 4.6 验收类

1. `ACC_001` AcceptanceProfile 缺失
2. `ACC_002` 证据不足
3. `ACC_003` 生产级裁决失败

### 4.7 Worker 类

1. `WRK_001` worker 启动失败
2. `WRK_002` 重试次数耗尽
3. `WRK_003` 索引刷新失败

## 5. 前端展示规则

规则建议：

1. `info` 显示轻提示
2. `warning` 显示卡片或横幅
3. `error` 进入问题面板
4. `critical` 阻断页面主流程并进入恢复态

## 6. 审计与诊断要求

每条关键错误至少要落：

1. `error_code`
2. `project_id`
3. `scope`
4. `severity`
5. `summary`
6. `context_json`
7. `created_at`

## 7. 恢复提示原则

每个错误码都应能映射出：

1. 是否可自动恢复
2. 是否可手动重试
3. 是否需要用户决策
4. 是否必须停止主流程

## 8. 后续细分专题

1. HTTP 错误返回结构
2. 前端错误提示组件规范
3. 恢复模式页面设计

