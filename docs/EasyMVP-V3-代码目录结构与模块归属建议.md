# EasyMVP V3 代码目录结构与模块归属建议

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-实现架构与模块拆分设计](./EasyMVP-V3-实现架构与模块拆分设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 目标：把实现架构进一步细化成建议代码目录结构，明确 Electron、React、Go 核心服务各自应该放什么，不应该放什么。

## 1. 设计结论

V3 代码结构不应按“页面文件夹 + 杂项工具”散落。

应该按运行边界与架构层分目录。

推荐采用双主体结构：

1. `apps/desktop` 承载 Electron + React
2. `apps/core` 承载 `Go + GoFrame v2` 本地核心服务

## 2. 推荐目录

```text
apps/
  desktop/
    package.json
    src/
      main/
      preload/
      renderer/
        app/
        modules/
        components/
        clients/
        stores/
        routes/
        styles/
  core/
    go.mod
    internal/
      cmd/
      controller/
      service/
      logic/
      dao/
      model/
    migrations/
    manifest/
    resource/
    hack/
    main.go
shared/
  schemas/
  contracts/
docs/
```

## 3. 归属建议

### `apps/desktop/src/main`

放：

1. Electron app shell
2. 窗口管理
3. Go 进程托管
4. 桌面原生事件

### `apps/desktop/src/preload`

放：

1. 桌面桥接 API
2. IPC 安全白名单
3. 类型注入

### `apps/desktop/src/renderer/app`

放：

1. renderer bootstrap
2. 路由入口
3. 全局 providers

### `apps/desktop/src/renderer/modules`

放：

1. `workspace`
2. `plan`
3. `acceptance`
4. `settings`

每个模块放：

1. 页面容器
2. 页面组件
3. 页面局部状态
4. 页面查询封装

### `apps/desktop/src/renderer/clients`

放：

1. 本地 API client
2. event stream client
3. desktop bridge client

### `apps/core/main.go` 与 `internal/cmd`

放：

1. GoFrame 启动入口
2. 配置装配
3. server 启动
4. CLI 初始化

### `apps/core/internal/controller`

放：

1. GoFrame controllers
2. route groups
3. request / response mapping

### `apps/core/internal/service`

放：

1. service interfaces
2. service registration
3. command / query service contracts

### `apps/core/internal/logic`

放：

1. orchestrator logic
2. plan logic
3. acceptance logic
4. runtime logic
5. aggregation logic
6. worker logic

### `apps/core/internal/dao`

放：

1. GoFrame DAO
2. table access
3. query base methods

### `apps/core/internal/model`

放：

1. `entity`
2. `do`
3. request / response models when needed

## 4. `shared/` 的原则

`shared/` 只放：

1. API contracts
2. schema 定义
3. 常量枚举

不放：

1. Go 业务逻辑
2. React 页面逻辑
3. Electron 壳逻辑

## 5. 禁止归属

不应：

1. 在 Renderer 模块里放 SQL
2. 在 storage 层放业务裁决
3. 在 aggregation 层写状态机
4. 在 preload 层放业务 API
5. 在 Electron Main 里写 orchestrator
