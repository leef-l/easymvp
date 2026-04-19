# EasyMVP V3 Electron GoFrame 单仓脚手架与开发命令设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-代码目录结构与模块归属建议](./EasyMVP-V3-代码目录结构与模块归属建议.md)
> 关联文档：[EasyMVP-V3-本地配置与启动参数设计](./EasyMVP-V3-本地配置与启动参数设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 目标：定义 V3 推荐的 Electron + GoFrame v2 单仓结构、开发命令和本地协作方式，避免仓库组织混乱。
>
> 历史口径说明：本文里的本地命令只表示开发调试方式，不代表当前测试/编译验收入口。现行工程铁律下，正式验证统一只认 GitHub Actions workflow run、日志、artifact 与 `.easymvp/ci/latest.json`。

## 1. 设计结论

V3 建议采用单仓结构：

1. `apps/desktop` 放 Electron + React
2. `apps/core` 放 `Go + GoFrame v2` 本地服务
3. `docs/` 放设计文档

这样做的收益：

1. 文档、UI、核心服务同仓演进
2. 本地路径与启动命令明确
3. 版本对齐更简单

## 2. 推荐目录

```text
apps/
  desktop/
    package.json
    src/
  core/
    go.mod
    main.go
    internal/
      cmd/
      controller/
      service/
      logic/
      dao/
      model/
docs/
```

## 3. 开发命令建议

### 3.1 UI 开发

建议：

1. `pnpm --dir apps/desktop dev`

### 3.2 Go 服务开发

建议：

1. `go run main.go`
2. 若需要完整测试/编译验收，统一以当前 V3 验证脚本和 `apps/core` / `apps/desktop` 的本地构建测试入口为准

### 3.3 联调开发

建议：

1. Electron 开发模式启动 UI
2. Go 服务独立启动
3. Renderer 指向本地 `127.0.0.1:<port>`

### 3.4 一键开发

后续可补：

1. 根目录 `pnpm dev`
2. 内部并发拉起 desktop 与 core

## 4. 构建命令建议

### 4.1 Go Core

建议：

1. 本地开发阶段可用 `go run main.go` 做最小启动联调
2. 正式构建产物与验收统一走 GitHub Actions，不再以本机 `go build` 作为当前口径

### 4.2 Desktop

建议：

1. UI 日常开发以 `pnpm --dir apps/desktop dev` 为主
2. 正式构建、类型检查与打包统一走 GitHub Actions，不再以本机 `pnpm --dir apps/desktop build/package` 作为当前口径

## 5. 联调约束

必须保证：

1. UI 不直接读取 Go 源码结构
2. Go 不依赖 Electron 进程
3. 双方通过本地 API 和 desktop bridge 交互

## 6. 后续细分专题

1. 根级任务脚本设计
2. CI 命令设计
3. 发布包结构设计
