# EasyMVP V3 本地 API 与 IPC 适配设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-技术栈与选型基线.md)
> 关联文档：[EasyMVP-V3-API路由分组与命令查询边界设计](./EasyMVP-V3-API路由分组与命令查询边界设计.md)
> 关联文档：[EasyMVP-V3-Electron进程模型与IPC边界设计](./EasyMVP-V3-Electron进程模型与IPC边界设计.md)
> 目标：定义文档中的 `/api/v3/*` 逻辑接口如何映射到本地 Go 服务与 Electron IPC，保证“文档 API”与“桌面实现”长期一致。

## 1. 设计结论

V3 文档中的 `/api/v3/*` 应被视为正式逻辑接口边界。

在桌面首版实现里，推荐：

1. 业务 API 物理实现为本地 `GoFrame v2` HTTP 服务
2. Renderer 通过统一 API client 调用
3. Electron IPC 只承载桌面原生能力

一句话：

> 业务走本地 HTTP，桌面能力走 IPC。

## 2. 适配原则

建议：

1. 页面层继续面向 query / command / event client
2. client 默认适配到 Go 本地 HTTP 服务
3. 本地文件选择、通知、打开目录等桌面能力走 Electron bridge
4. 文档 API 不暴露 Electron channel 细节

## 3. 逻辑边界与物理边界

### 3.1 逻辑边界

文档中的：

1. `GET /api/v3/workspace/home-view`
2. `POST /api/v3/projects`
3. `GET /api/v3/projects/{id}/workspace-view`
4. `POST /api/v3/projects/{id}/acceptance-runs`

这些都属于正式业务边界。

### 3.2 物理边界

首版物理实现建议：

1. Go 本地服务监听 `127.0.0.1` 回环地址
2. Renderer 直接通过 HTTP client 请求
3. Electron Main 仅负责服务保活与异常感知

IPC 只保留给：

1. `selectDirectory`
2. `openExternal`
3. `revealInFinder`
4. `windowState`
5. `shellDiagnostics`

## 4. 示例映射

例如：

1. `GET /api/v3/workspace/home-view`
   映射为 Go handler：`GET 127.0.0.1:<port>/api/v3/workspace/home-view`
2. `POST /api/v3/projects`
   映射为 Go handler：`POST 127.0.0.1:<port>/api/v3/projects`
3. `GET /api/v3/project-creations/{id}/events`
   映射为 Go 本地事件接口或增量流接口

而：

1. `choose workspace path`
   映射为 `desktop.selectDirectory()`
2. `open evidence folder`
   映射为 `desktop.showItemInFolder(path)`

## 5. 好处

1. 页面不被 Electron 物理细节绑死
2. API client 与文档天然一致
3. 核心服务可脱离桌面壳独立调试
4. 后续如果需要换桌面壳，不需要重写主业务边界

## 6. 不该怎么做

不应该：

1. 让页面直接写 Electron channel
2. 文档一套 API，代码另一套命名
3. 把主业务命令藏进 preload
4. 让 Go API 和页面模型长期失配

## 7. 首版 client 建议

前端建议统一抽象：

1. `queryClient`
2. `commandClient`
3. `eventClient`
4. `desktopClient`

其中：

1. 前三个默认走 Go 本地 API
2. 最后一个走 Electron IPC

## 8. 后续细分专题

1. 本地 HTTP 服务端口与鉴权策略
2. 事件流协议细化
3. API client SDK 设计
