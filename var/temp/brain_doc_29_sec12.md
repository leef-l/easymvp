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
  │ ◄─────────── tool_use      │ 返回 tool 指令
  │ 本地执行工具                  │
  │ tool_result ──────────►    │ 继续推理
  │ ◄─────────── 最终结果        │
```

**模式 B：远程全栈执行**

远程 brain 自带工具执行能力，适合云开发工作站场景。项目文件在远程服务器上。

详细说明参见 [37-远程专精大脑调用说明.md](./37-远程专精大脑调用说明.md)。

---

