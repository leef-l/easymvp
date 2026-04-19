# EasyMVP V3 本地配置与启动参数设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-技术栈与选型基线.md)
> 关联文档：[EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go本地核心服务架构设计.md)
> 关联文档：[EasyMVP-V3-Electron进程模型与IPC边界设计](./EasyMVP-V3-Electron进程模型与IPC边界设计.md)
> 目标：为 V3 单机版定义本地配置文件、启动参数和环境探测规则，避免实现期出现配置散落和运行入口不稳定。

## 1. 设计结论

V3 单机版配置应分三层：

1. 固定内建默认值
2. 本地配置文件
3. 启动参数覆盖

优先级：

`CLI flags > local config > built-in defaults`

## 2. 配置分类

建议至少分：

1. `paths`
2. `server`
3. `runtime`
4. `workspace`
5. `ui`
6. `diagnostics`

## 3. 推荐配置项

### 3.1 `paths`

建议：

1. `data_root`
2. `db_path`
3. `projects_root`
4. `temp_root`
5. `backups_root`

### 3.2 `server`

建议：

1. `host`
2. `port`
3. `read_timeout_ms`
4. `write_timeout_ms`

### 3.3 `runtime`

建议：

1. `brain_serve_base_url`
2. `brain_request_timeout_ms`
3. `brain_sync_interval_ms`

### 3.4 `workspace`

建议：

1. `default_project_root`
2. `default_category`
3. `auto_resume_workers`

### 3.5 `ui`

建议：

1. `window_width`
2. `window_height`
3. `last_route`

### 3.6 `diagnostics`

建议：

1. `log_level`
2. `keep_days`
3. `enable_debug_dump`

## 4. 配置文件建议

建议默认文件：

```text
~/.easymvp/settings/config.json
```

说明：

1. 用户级长期配置放这里
2. 项目级差异配置仍进数据库或项目事实对象

## 5. 启动参数建议

建议支持：

1. `--data-root`
2. `--port`
3. `--brain-serve-base-url`
4. `--log-level`
5. `--safe-mode`

## 6. `safe-mode`

建议保留安全启动模式：

1. 不启动后台 worker
2. 不自动恢复未完成同步
3. 只做数据库与配置检查
4. 进入诊断与恢复视图

## 7. 不建议的做法

不建议：

1. 把项目业务配置全塞进本地 JSON
2. 让 Electron 与 Go 各自维护一套不同配置源
3. 把敏感运行时地址散落在多个文件

## 8. 后续细分专题

1. 配置 schema 校验规则
2. 设置页字段与配置映射
3. 安全模式恢复流程

