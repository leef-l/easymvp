# 29. 第三方专精大脑开发指南

> 目标：让第三方开发者可以独立实现并发布自己的 Specialist Brain，
> 通过 BrainKernel 的 sidecar 协议被主进程或 central brain 调用。

> **⚠️ 代码对照勘误（2026-04-24）：**
> - §3 import 路径：`github.com/leef-l/brain/sdk/agent` 和 `github.com/leef-l/brain/sdk/sidecar`
> - §7 二进制发现：`brain-<kind>-sidecar` 优先级高
---
# 29. 第三方专精大脑开发指南

> 目标：让第三方开发者可以独立实现并发布自己的 Specialist Brain，
> 通过 BrainKernel 的 sidecar 协议被主进程或 central brain 调用。

> **⚠️ 代码对照勘误（2026-04-24）：**
> - §3 import 路径：`github.com/leef-l/brain/sdk/agent` 和 `github.com/leef-l/brain/sdk/sidecar`
> - §7 二进制发现：`brain-<kind>-sidecar` 优先级高于 `brain-<kind>`（
---
# 29. 第三方专精大脑开发指南

> 目标：让第三方开发者可以独立实现并发布自己的 Specialist Brain，
> 通过 BrainKernel 的 sidecar 协议被主进程或 central brain 调用。

> **⚠️ 代码对照勘误（2026-04-24）：**
> - §3 import 路径：`github.com/leef-l/brain/sdk/agent` 和 `github.com/leef-l/brain/sdk/sidecar`
> - §7 二进制发现：`brain-<kind>-sidecar` 优先级高于 `brain-<kind>`（代码 `cmd/b
---
 - §3 import 路径：`github.com/leef-l/brain/sdk/agent` 和 `github.com/leef-l/brain/sdk/sidecar`
> - §7 二进制发现：`brain-<kind>-sidecar` 优先级高于 `brain-<kind>`（代码 `cmd/brain/helpers.go:55-66`）
> - §10 测试路径：`sdk/kernel/` 和 `sdk/sidecar/`
> - §14 签名命令：keygen/sign 函数已实现但 CLI 路由未注册，install --verify-key 参数未实现，当前只有底层 API 可用
> - 文档中文件链接路径已统一修正为当前仓库相对路径

---

## 1. 总结版

在 v3 术语里，第三方专精大脑的顶层对象是 `Brain`。

最常见的实现方式仍然是：

- `Br
---
话在实现层成立；  
> 但在 v3 顶层模型里，更准确的说法是  
> “第三方 brain 通常是一个以 sidecar 为 runtime 的 Brain Package”。

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
3.
---
；  
> 但在 v3 顶层模型里，更准确的说法是  
> “第三方 brain 通常是一个以 sidecar 为 runtime 的 Brain Package”。

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
3. 至少支持 `brain/
---
 sidecar 二进制**：

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

- `Protocol Version`
---

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

原因很直接：

- `age
---
Handler`
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

原因很直接：

- `agent.Descriptor.Version
---
要怎么管

先区分 4 个概念：

- `Protocol Version`：线缆协议版本，例如 `1.0`
- `Kernel Version`：BrainKernel 行为契约版本，例如 `1.0.0`
- `SDK Version`：SDK 自己的版本
- `Brain Version`：某个专精大脑自己的版本

其中第三方专精大脑最容易搞混的是最后一个。

### 2.1 不需要强制对齐 central brain

**专精大脑版本号不需要强制对齐 central brain。**

原因很直接：

- `agent.Descriptor.Version` 注释里已经写明它是 sidecar 自身版本，**independent of SDK/Kernel**
- `initialize` 响应里的 `brain_version` 也是给诊断和兼容门控用的，不是仓库总版本

相关代码：

- 
---
rsion`：线缆协议版本，例如 `1.0`
- `Kernel Version`：BrainKernel 行为契约版本，例如 `1.0.0`
- `SDK Version`：SDK 自己的版本
- `Brain Version`：某个专精大脑自己的版本

其中第三方专精大脑最容易搞混的是最后一个。

### 2.1 不需要强制对齐 central brain

**专精大脑版本号不需要强制对齐 central brain。**

原因很直接：

- `agent.Descriptor.Version` 注释里已经写明它是 sidecar 自身版本，**independent of SDK/Kernel**
- `initialize` 响应里的 `brain_version` 也是给诊断和兼容门控用的，不是仓库总版本

相关代码：

- [`sdk/agent/agent.go`](../../sdk
---
 `Brain Version`：某个专精大脑自己的版本

其中第三方专精大脑最容易搞混的是最后一个。

### 2.1 不需要强制对齐 central brain

**专精大脑版本号不需要强制对齐 central brain。**

原因很直接：

- `agent.Descriptor.Version` 注释里已经写明它是 sidecar 自身版本，**independent of SDK/Kernel**
- `initialize` 响应里的 `brain_version` 也是给诊断和兼容门控用的，不是仓库总版本

相关代码：

- [`sdk/agent/agent.go`](../../sdk/agent/agent.go:63)
- [`sdk/protocol/lifecycle.go`](../../sdk/protocol/lifecycle.go:60)

### 2.2 真
---
码：

- [`sdk/agent/agent.go`](../../sdk/agent/agent.go:63)
- [`sdk/protocol/lifecycle.go`](../../sdk/protocol/lifecycle.go:60)

### 2.2 真正需要对齐的是这些

- **Protocol** 要兼容：必须能说同一版 sidecar wire protocol
- **Kernel 行为预期** 要兼容：至少知道 host 这一版会怎么调你
- **方法契约** 要兼容：如 `initialize`、`brain/execute`、`tools/call`

### 2.3 内置 sidecar 版本号应该怎么看

规范层面，内置 sidecar 应与当前正式发布口径保持一致；当前主仓库
正式版本当前是 `0.6.0`。

真正应遵循的是：

- `brain versio
---
`、`brain/execute`、`tools/call`

### 2.3 内置 sidecar 版本号应该怎么看

规范层面，内置 sidecar 应与当前正式发布口径保持一致；当前主仓库
正式版本当前是 `0.6.0`。

真正应遵循的是：

- `brain version`、`VERSION.json`、sidecar `Version()` 最好保持同一发布代次
- 第三方不需要跟 central brain 完全同号，但应明确声明“测试过哪版 kernel”
- 版本一致性是发布工程问题，不是协议耦合要求

### 2.4 对第三方的推荐做法

建议第三方专精大脑声明自己的版本矩阵：

```text
brain-image@1.2.0
implements protocol 1.0
tested against kernel 1.0.x
```

也就是：

- `brainVersi
---
### 2.3 内置 sidecar 版本号应该怎么看

规范层面，内置 sidecar 应与当前正式发布口径保持一致；当前主仓库
正式版本当前是 `0.6.0`。

真正应遵循的是：

- `brain version`、`VERSION.json`、sidecar `Version()` 最好保持同一发布代次
- 第三方不需要跟 central brain 完全同号，但应明确声明“测试过哪版 kernel”
- 版本一致性是发布工程问题，不是协议耦合要求

### 2.4 对第三方的推荐做法

建议第三方专精大脑声明自己的版本矩阵：

```text
brain-image@1.2.0
implements protocol 1.0
tested against kernel 1.0.x
```

也就是：

- `brainVersion`：你自己的 semver
- `protocolVers
---
rsion()` 最好保持同一发布代次
- 第三方不需要跟 central brain 完全同号，但应明确声明“测试过哪版 kernel”
- 版本一致性是发布工程问题，不是协议耦合要求

### 2.4 对第三方的推荐做法

建议第三方专精大脑声明自己的版本矩阵：

```text
brain-image@1.2.0
implements protocol 1.0
tested against kernel 1.0.x
```

也就是：

- `brainVersion`：你自己的 semver
- `protocolVersion`：你支持哪版线缆协议
- `testedKernel`：你验证过哪些 Kernel 版本

---

## 3. 一个第三方专精大脑长什么样

最小结构：

```text
brain-image/
  manifest.json
  README.md
  go.m
---
求

### 2.4 对第三方的推荐做法

建议第三方专精大脑声明自己的版本矩阵：

```text
brain-image@1.2.0
implements protocol 1.0
tested against kernel 1.0.x
```

也就是：

- `brainVersion`：你自己的 semver
- `protocolVersion`：你支持哪版线缆协议
- `testedKernel`：你验证过哪些 Kernel 版本

---

## 3. 一个第三方专精大脑长什么样

最小结构：

```text
brain-image/
  manifest.json
  README.md
  go.mod
  main.go
```

更推荐的 v3 结构：

```text
brain-image/
  manifest.json
  README.md
  go.mod
  m
---
的推荐做法

建议第三方专精大脑声明自己的版本矩阵：

```text
brain-image@1.2.0
implements protocol 1.0
tested against kernel 1.0.x
```

也就是：

- `brainVersion`：你自己的 semver
- `protocolVersion`：你支持哪版线缆协议
- `testedKernel`：你验证过哪些 Kernel 版本

---

## 3. 一个第三方专精大脑长什么样

最小结构：

```text
brain-image/
  manifest.json
  README.md
  go.mod
  main.go
```

更推荐的 v3 结构：

```text
brain-image/
  manifest.json
  README.md
  go.mod
  main.go
  bin/
 
---
yMVP 领域规则（审核/编译/裁决/返工/验收）。

```
brain-v3 (code/browser/verifier/fault)
         │
         ▼  通过 sidecar 协议调用（brain/execute、tools/call）
   easymvp-brain（独立 sidecar 二进制）
         │
         ▼  反向 RPC 回主进程（llm.complete、tool.invoke、plan.update）
   主进程 Kernel
```

实现要点：

- **Kind**：`agent.Kind("easymvp")`
- **Version**：独立版本号，如 `1.0.0`，不强制对齐 brain-v3 版本
- **入口方法**：`brain/execute` 接收领域任务，内部编排对 code/browser
---
e/browser/verifier/fault)
         │
         ▼  通过 sidecar 协议调用（brain/execute、tools/call）
   easymvp-brain（独立 sidecar 二进制）
         │
         ▼  反向 RPC 回主进程（llm.complete、tool.invoke、plan.update）
   主进程 Kernel
```

实现要点：

- **Kind**：`agent.Kind("easymvp")`
- **Version**：独立版本号，如 `1.0.0`，不强制对齐 brain-v3 版本
- **入口方法**：`brain/execute` 接收领域任务，内部编排对 code/browser/verifier/fault 的调用
- **反向调用**：通过 `RichBrainH