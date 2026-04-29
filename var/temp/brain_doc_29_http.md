
第三方专精大脑本质上就是一个**独立 sidecar 二进制**：

- 进程模型：独立进程
- 传输：`stdin/stdout` 上的 Content-Length framed JSON-RPC
- 启动：由 Kernel/Orchestrator 拉起
- 握手：`initialize`
- 执行入口：通常实现 `brain/execute`
- 工具执行：在 sidecar 内本地执行，或通过 reverse RPC 调回主进程

如果你只想最快跑通，最小实现就是：

1. 实现 `sidecar.BrainHandler`
2. 在 `main()` 里调用 `sidecar.Run(handler)`
3. 至少支持 `brain/execute`
4. 正确返回 `Tools()` 和 `Version()`

---

## 2. 版本号要怎么管

先区分 4 个概念：

---
如果你只想最快跑通，最小实现就是：

1. 实现 `sidecar.BrainHandler`
2. 在 `main()` 里调用 `sidecar.Run(handler)`
3. 至少支持 `brain/execute`
4. 正确返回 `Tools()` 和 `Version()`

---

## 2. 版本号要怎么管

先区分 4 个概念：

- `Protocol Version`：线缆协议版本，例如 `1.0`
- `Kernel Version`：BrainKernel 行为契约版本，例如 `1.0.0`
- `SDK Version`：SDK 自己的版本
- `Brain Version`：某个专精大脑自己的版本

其中第三方专精大脑最容易搞混的是最后一个。

### 2.1 不需要强制对齐 central brain

**专精大脑版本号不需要强制对齐 central brain。**
---
}

func main() {
  if err := sidecar.Run(&imageHandler{}); err != nil {
    fmt.Fprintf(os.Stderr, "brain-image: %v\n", err)
    os.Exit(1)
  }
}
```

---

## 4. BrainHandler 接口

sidecar 运行时要求你实现的最小接口在：

- [`sidecar/sidecar.go`](/www/wwwroot/project/brain/sidecar/sidecar.go:22)

你需要提供：

- `Kind()`：专精大脑类型标识，如 `image`
- `Version()`：这个 sidecar 自己的版本号
- `Tools()`：你支持的工具名列表
---

### 5.1 `initialize`

这个不需要你自己手写，`sidecar.Run()` 会帮你注册。

它会自动返回：

- `protocol_version`
- `brain_version`
- `supported_tools`

代码在：

- [`sidecar/sidecar.go`](/www/wwwroot/project/brain/sidecar/sidecar.go:85)

### 5.2 `tools/list`

这个也由 `sidecar.Run()` 自动处理。

- 最低兼容模式：只要实现 `Tools()`，返回工具名列表即可
- 推荐模式：实现 `sidecar.ToolSchemaProvider`，让 `tools/list` 返回完整工具元数据

当前内置 sidecar（包括 `central` / `code` / `browser` / `verifier` / `fault`）的 `tools/list` 会返回：
---

### 5.2 `tools/list`

这个也由 `sidecar.Run()` 自动处理。

- 最低兼容模式：只要实现 `Tools()`，返回工具名列表即可
- 推荐模式：实现 `sidecar.ToolSchemaProvider`，让 `tools/list` 返回完整工具元数据

当前内置 sidecar（包括 `central` / `code` / `browser` / `verifier` / `fault`）的 `tools/list` 会返回：

- `name`
- `description`
- `input_schema`
- `output_schema`

其中 `output_schema` 表示工具成功返回值的 JSON Schema；如果工具结果是动态结构，也允许给出宽松 schema。
这套元数据是加法兼容的：后续 `Brain Manifest / Package / Marketplace` 可以直接复用，不需要改变现有 `tools/call` 语义。

### 5.3 `tools/call`

如果你的 brain 有本地工具，这个要支持。通常做法是：

- 反序列化 `{name, arguments, execution?}`
---
### 9.2 标准版

- 实现本地工具 registry
- 通过 `sidecar.RunAgentLoop(...)` 跑完整工具循环
- 支持 `toolpolicy`
- 返回明确 summary / error / turns

### 9.3 进阶版

- 支持自己的 `brain/plan` 或 `brain/verify`
- 自定义 capability flags
- 支持 direct / hybrid LLMAccess
- 补完整集成测试和跨版本兼容矩阵

---

## 10. 推荐测试清单

至少测这些：

1. `initialize` 握手成功
2. `brain/execute` 正常返回
3. `tools/call` 能正确执行和报错
---

---

## 12. 双模式运行：本地 stdio 与远程网络

v3 的每个专精大脑支持两种运行模式，共用同一个 `BrainHandler` 实现：

### 12.1 本地 sidecar 模式（默认）

```bash
brain-image                     # stdio JSON-RPC，被 Kernel fork 拉起
```

这是默认模式，`sidecar.Run(handler)` 通过 stdin/stdout 通信。

### 12.2 远程网络模式

```bash
brain-image --listen :8080      # HTTP + WebSocket 网络服务
```

加 `--listen` 参数后，`sidecar.ListenAndServe(addr, handler)` 启动 HTTP 服务，暴露三个端点：

---
### 12.1 本地 sidecar 模式（默认）

```bash
brain-image                     # stdio JSON-RPC，被 Kernel fork 拉起
```

这是默认模式，`sidecar.Run(handler)` 通过 stdin/stdout 通信。

### 12.2 远程网络模式

```bash
brain-image --listen :8080      # HTTP + WebSocket 网络服务
```

加 `--listen` 参数后，`sidecar.ListenAndServe(addr, handler)` 启动 HTTP 服务，暴露三个端点：

| 端点 | 方法 | 功能 |
|------|------|------|
| `/health` | GET | 健康检查，返回 `{"status":"ok","kind":"image","version":"1.0.0"}` |
| `/rpc` | POST | 单次 JSON-RPC 请求/响应（无反向调用能力） |
| `/ws` | GET | WebSocket 升级为双向 BidirRPC（完整模式，支持 `llm.complete` 等反向调用） |

### 12.3 第三方 brain 启用双模式
---
brain-image                     # stdio JSON-RPC，被 Kernel fork 拉起
```

这是默认模式，`sidecar.Run(handler)` 通过 stdin/stdout 通信。

### 12.2 远程网络模式

```bash
brain-image --listen :8080      # HTTP + WebSocket 网络服务
```

加 `--listen` 参数后，`sidecar.ListenAndServe(addr, handler)` 启动 HTTP 服务，暴露三个端点：

| 端点 | 方法 | 功能 |
|------|------|------|
| `/health` | GET | 健康检查，返回 `{"status":"ok","kind":"image","version":"1.0.0"}` |
| `/rpc` | POST | 单次 JSON-RPC 请求/响应（无反向调用能力） |
| `/ws` | GET | WebSocket 升级为双向 BidirRPC（完整模式，支持 `llm.complete` 等反向调用） |

### 12.3 第三方 brain 启用双模式

在 `main.go` 中加 3 行即可：

---
### 12.2 远程网络模式

```bash
brain-image --listen :8080      # HTTP + WebSocket 网络服务
```

加 `--listen` 参数后，`sidecar.ListenAndServe(addr, handler)` 启动 HTTP 服务，暴露三个端点：

| 端点 | 方法 | 功能 |
|------|------|------|
| `/health` | GET | 健康检查，返回 `{"status":"ok","kind":"image","version":"1.0.0"}` |
| `/rpc` | POST | 单次 JSON-RPC 请求/响应（无反向调用能力） |
| `/ws` | GET | WebSocket 升级为双向 BidirRPC（完整模式，支持 `llm.complete` 等反向调用） |

### 12.3 第三方 brain 启用双模式

在 `main.go` 中加 3 行即可：

```go
func main() {
    listen := ""
    for i, arg := range os.Args[1:] {
        if arg == "--listen" && i+1 < len(os.Args[1:]) {
---
brain-image --listen :8080      # HTTP + WebSocket 网络服务
```

加 `--listen` 参数后，`sidecar.ListenAndServe(addr, handler)` 启动 HTTP 服务，暴露三个端点：

| 端点 | 方法 | 功能 |
|------|------|------|
| `/health` | GET | 健康检查，返回 `{"status":"ok","kind":"image","version":"1.0.0"}` |
| `/rpc` | POST | 单次 JSON-RPC 请求/响应（无反向调用能力） |
| `/ws` | GET | WebSocket 升级为双向 BidirRPC（完整模式，支持 `llm.complete` 等反向调用） |

### 12.3 第三方 brain 启用双模式

在 `main.go` 中加 3 行即可：

```go
func main() {
    listen := ""
    for i, arg := range os.Args[1:] {
        if arg == "--listen" && i+1 < len(os.Args[1:]) {
            listen = os.Args[i+2]
        }
    }
---

    handler := &imageHandler{}
    if listen != "" {
        sidecar.ListenAndServe(listen, handler)  // 网络模式
    } else {
        sidecar.Run(handler)                      // stdio 模式
    }
}
```

**业务逻辑完全不变**——两种模式共用同一个 `BrainHandler`。

### 12.4 远程 brain 的连接配置

Kernel 端在 `~/.brain/config.json` 中配置远程 brain：

```json
{
  "remote_brains": [
    {
      "kind": "image",
      "endpoint": "https://gpu-server:8080",
      "api_key": "sk-xxx",
---
    if listen != "" {
        sidecar.ListenAndServe(listen, handler)  // 网络模式
    } else {
        sidecar.Run(handler)                      // stdio 模式
    }
}
```

**业务逻辑完全不变**——两种模式共用同一个 `BrainHandler`。

### 12.4 远程 brain 的连接配置

Kernel 端在 `~/.brain/config.json` 中配置远程 brain：

```json
{
  "remote_brains": [
    {
      "kind": "image",
      "endpoint": "https://gpu-server:8080",
      "api_key": "sk-xxx",
      "auto_start": true
    }
---
  "remote_brains": [
    {
      "kind": "image",
      "endpoint": "https://gpu-server:8080",
      "api_key": "sk-xxx",
      "auto_start": true
    }
  ]
}
```

Kernel 启动时自动连接远程 brain，先 HTTP health check，再升级为 WebSocket BidirRPC。

### 12.5 两种远程使用模式

**模式 A：远程推理 + 本地工具**（推荐）

```
你的机器                          GPU 服务器
Kernel ◄── WebSocket ──► brain-image --listen :8080
  │                            │
  │ brain/execute ──────────►  │ LLM 推理
  │                            │ llm.complete ──► Kernel 代理 LLM 调用
---
}
```

Kernel 启动时自动连接远程 brain，先 HTTP health check，再升级为 WebSocket BidirRPC。

### 12.5 两种远程使用模式

**模式 A：远程推理 + 本地工具**（推荐）

```
你的机器                          GPU 服务器
Kernel ◄── WebSocket ──► brain-image --listen :8080
  │                            │
  │ brain/execute ──────────►  │ LLM 推理
  │                            │ llm.complete ──► Kernel 代理 LLM 调用
  │ ◄─────────── tool_use      │ 返回 tool 指令
  │ 本地执行工具                  │
  │ tool_result ──────────►    │ 继续推理
  │ ◄─────────── 最终结果        │
```

**模式 B：远程全栈执行**

---