# EasyMVP V3 Electron 进程模型与 IPC 边界设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-技术栈与选型基线.md)
> 关联文档：[EasyMVP-V3-本地API与IPC适配设计](./EasyMVP-V3-本地API与IPC适配设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 目标：定义 V3 在 Electron 下的进程职责、Go 核心服务托管方式以及 IPC 的严格边界，避免桌面壳和业务内核混在一起。

## 1. 设计结论

V3 在 Electron 下建议采用三段式结构：

1. `main process`
2. `renderer process`
3. `go local core process`

其中：

1. Electron Main 负责桌面壳
2. Renderer 负责可视化 UI
3. Go 进程负责主业务内核

不要把业务逻辑散到 Renderer，也不要把业务内核塞进 Main。

## 2. 进程分工

### 2.1 Main Process

负责：

1. app lifecycle
2. 窗口管理
3. preload 注入
4. 启动 / 关闭 Go 本地服务
5. 本地端口探活与健康检查
6. 文件选择、系统通知、托盘、菜单等原生能力
7. 桌面级诊断采集

不负责：

1. 项目状态机
2. 计划编译
3. 验收裁决
4. 直接访问业务数据库做业务写入

### 2.2 Renderer Process

负责：

1. React UI
2. 路由
3. 页面交互
4. 查询快照渲染
5. 事件流可视化
6. 本地临时展示状态

不负责：

1. 直接调用 SQLite
2. 直接调用 `brain serve`
3. 承担主业务状态机
4. 本地文件主目录遍历

### 2.3 Go Local Core Process

负责：

1. workflow orchestrator
2. plan system
3. acceptance system
4. runtime adapter
5. worker manager
6. view aggregation
7. repository / storage
8. local HTTP API

说明：

1. 这是 V3 的真正业务内核
2. 所有领域状态推进都应经过这层

## 3. IPC 的正式定位

建议把 IPC 限定为桌面桥接层，而不是主业务 API 总线。

IPC 主要用于：

1. `selectDirectory`
2. `openPath`
3. `showNotification`
4. `windowControl`
5. `desktopDiagnostics`
6. `appVersion`

不建议通过 IPC 承载：

1. `/api/v3/projects`
2. `/api/v3/plan/*`
3. `/api/v3/acceptance/*`
4. 工作台业务事件总线

## 4. 推荐关系

```text
Renderer
  → HTTP client
  → localhost /api/v3/*
Go Local Core
  → SQLite / Local Files / brain serve

Renderer
  → Electron preload bridge
  → IPC
Main Process
  → native desktop capabilities
```

## 5. 严格禁止

不应：

1. Renderer 直接访问 SQLite
2. Renderer 直接访问工作区主目录
3. Renderer 直接调用 `brain serve`
4. Main Process 直接实现业务状态机
5. preload 层承载业务规则

## 6. 为什么这样切

这样切的收益是：

1. 桌面壳可替换
2. 业务内核可测试
3. UI 技术栈和业务内核边界清晰
4. Go 能稳定承接 worker、runtime、aggregation
5. 与 `brain-v3` 的对接路径更自然

## 7. 首版建议的 preload 暴露面

preload 建议只暴露：

1. `desktop.selectDirectory()`
2. `desktop.openPath(path)`
3. `desktop.showItemInFolder(path)`
4. `desktop.getAppInfo()`
5. `desktop.onShellEvent(handler)`

不要暴露：

1. `projectService.*`
2. `acceptanceService.*`
3. `sqlite.*`
4. `brainServe.*`

## 8. 后续细分专题

1. preload API 类型定义
2. Go 服务托管与异常恢复
3. 桌面原生能力白名单策略
