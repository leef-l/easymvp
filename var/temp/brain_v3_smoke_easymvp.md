.\dev_docker.bat clean
```

> 注意：Docker 模式下 `brainServeBaseURL` 默认为 `http://host.docker.internal:7701`，确保 brain-v3 serve 在宿主机上运行且监听 `0.0.0.0:7701`（或 Docker Desktop 已启用 host gateway）。

**docker-compose.dev.yml 关键服务：**

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| easymvp-core | easymvp-core | 8000 | EasyMVP Core HTTP |
| easymvp-desktop | easymvp-desktop | 3000 | Desktop Dev Server |
| easymvp-brain | easymvp-brain | 7801 | easymvp-brain sidecar JSON-RPC |

---

## 5. EasyMVP Core 连接 brain-v3 的配置

### 5.1 配置位置总览（在项目根目录 `C:\Users\Public\project\easymvp` 内）

| 配置项 | 文件位置 | 修改方式 |
|--------|---------|---------|
| `brainServeBaseURL` | `docker/dev/config.yaml` → `easymvp.brainServeBaseURL` | 改 YAML |
| `easymvpBrain.*` | `docker/dev/config.yaml` → `easymvp.easymvpBrain.*` | 改 YAML |
| 环境变量模板 | `
---
 clean
```

> 注意：Docker 模式下 `brainServeBaseURL` 默认为 `http://host.docker.internal:7701`，确保 brain-v3 serve 在宿主机上运行且监听 `0.0.0.0:7701`（或 Docker Desktop 已启用 host gateway）。

**docker-compose.dev.yml 关键服务：**

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| easymvp-core | easymvp-core | 8000 | EasyMVP Core HTTP |
| easymvp-desktop | easymvp-desktop | 3000 | Desktop Dev Server |
| easymvp-brain | easymvp-brain | 7801 | easymvp-brain sidecar JSON-RPC |

---

## 5. EasyMVP Core 连接 brain-v3 的配置

### 5.1 配置位置总览（在项目根目录 `C:\Users\Public\project\easymvp` 内）

| 配置项 | 文件位置 | 修改方式 |
|--------|---------|---------|
| `brainServeBaseURL` | `docker/dev/config.yaml` → `easymvp.brainServeBaseURL` | 改 YAML |
| `easymvpBrain.*` | `docker/dev/config.yaml` → `easymvp.easymvpBrain.*` | 改 YAML |
| 环境变量模板 | `apps/core/.env.e
---
：Docker 模式下 `brainServeBaseURL` 默认为 `http://host.docker.internal:7701`，确保 brain-v3 serve 在宿主机上运行且监听 `0.0.0.0:7701`（或 Docker Desktop 已启用 host gateway）。

**docker-compose.dev.yml 关键服务：**

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| easymvp-core | easymvp-core | 8000 | EasyMVP Core HTTP |
| easymvp-desktop | easymvp-desktop | 3000 | Desktop Dev Server |
| easymvp-brain | easymvp-brain | 7801 | easymvp-brain sidecar JSON-RPC |

---

## 5. EasyMVP Core 连接 brain-v3 的配置

### 5.1 配置位置总览（在项目根目录 `C:\Users\Public\project\easymvp` 内）

| 配置项 | 文件位置 | 修改方式 |
|--------|---------|---------|
| `brainServeBaseURL` | `docker/dev/config.yaml` → `easymvp.brainServeBaseURL` | 改 YAML |
| `easymvpBrain.*` | `docker/dev/config.yaml` → `easymvp.easymvpBrain.*` | 改 YAML |
| 环境变量模板 | `apps/core/.env.example`
---
 模式下 `brainServeBaseURL` 默认为 `http://host.docker.internal:7701`，确保 brain-v3 serve 在宿主机上运行且监听 `0.0.0.0:7701`（或 Docker Desktop 已启用 host gateway）。

**docker-compose.dev.yml 关键服务：**

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| easymvp-core | easymvp-core | 8000 | EasyMVP Core HTTP |
| easymvp-desktop | easymvp-desktop | 3000 | Desktop Dev Server |
| easymvp-brain | easymvp-brain | 7801 | easymvp-brain sidecar JSON-RPC |

---

## 5. EasyMVP Core 连接 brain-v3 的配置

### 5.1 配置位置总览（在项目根目录 `C:\Users\Public\project\easymvp` 内）

| 配置项 | 文件位置 | 修改方式 |
|--------|---------|---------|
| `brainServeBaseURL` | `docker/dev/config.yaml` → `easymvp.brainServeBaseURL` | 改 YAML |
| `easymvpBrain.*` | `docker/dev/config.yaml` → `easymvp.easymvpBrain.*` | 改 YAML |
| 环境变量模板 | `apps/core/.env.example` | 复制为 `.env` 并修
---
 Docker 开发配置（`docker/dev/config.yaml`）

```yaml
server:
  address: ":8000"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

logger:
  level: "all"
  stdout: true

easymvp:
  dataRoot: "/app/var"
  dbPath: "/app/var/data/easymvp.db"
  migrationPath: "/app/manifest/migrations"
  brainServeBaseURL: "http://host.docker.internal:7701"   # ← 连接 brain-v3
  safeMode: false
  easymvpBrain:
    mode: "local-sidecar"                                   # ← local-sidecar 或 remote-service
    localEndpoint: "http://host.docker.internal:7801"       # ← easymvp-brain sidecar（如使用）
    remoteEndpoint: ""
    apiKey: ""
    timeout: "30s"
    maxTurns: 6
```

### 5.3 本地裸机运行配置（不通过 Docker）

```powershell
cd C:\Users\Public\project\easymvp\apps\core

# 方式一：CLI flag 覆盖（推荐冒烟测试用）
go run main.go `
  --port 8000 `
  --brain-
---
r:
  address: ":8000"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

logger:
  level: "all"
  stdout: true

easymvp:
  dataRoot: "/app/var"
  dbPath: "/app/var/data/easymvp.db"
  migrationPath: "/app/manifest/migrations"
  brainServeBaseURL: "http://host.docker.internal:7701"   # ← 连接 brain-v3
  safeMode: false
  easymvpBrain:
    mode: "local-sidecar"                                   # ← local-sidecar 或 remote-service
    localEndpoint: "http://host.docker.internal:7801"       # ← easymvp-brain sidecar（如使用）
    remoteEndpoint: ""
    apiKey: ""
    timeout: "30s"
    maxTurns: 6
```

### 5.3 本地裸机运行配置（不通过 Docker）

```powershell
cd C:\Users\Public\project\easymvp\apps\core

# 方式一：CLI flag 覆盖（推荐冒烟测试用）
go run main.go `
  --port 8000 `
  --brain-serve-base-url http://127.0.0.1:7701 `
  --data-root 
---
ger:
  level: "all"
  stdout: true

easymvp:
  dataRoot: "/app/var"
  dbPath: "/app/var/data/easymvp.db"
  migrationPath: "/app/manifest/migrations"
  brainServeBaseURL: "http://host.docker.internal:7701"   # ← 连接 brain-v3
  safeMode: false
  easymvpBrain:
    mode: "local-sidecar"                                   # ← local-sidecar 或 remote-service
    localEndpoint: "http://host.docker.internal:7801"       # ← easymvp-brain sidecar（如使用）
    remoteEndpoint: ""
    apiKey: ""
    timeout: "30s"
    maxTurns: 6
```

### 5.3 本地裸机运行配置（不通过 Docker）

```powershell
cd C:\Users\Public\project\easymvp\apps\core

# 方式一：CLI flag 覆盖（推荐冒烟测试用）
go run main.go `
  --port 8000 `
  --brain-serve-base-url http://127.0.0.1:7701 `
  --data-root ./var `
  --db-path ./var/data/easymvp.db `
  --migration-path ./manif
---
ll"
  stdout: true

easymvp:
  dataRoot: "/app/var"
  dbPath: "/app/var/data/easymvp.db"
  migrationPath: "/app/manifest/migrations"
  brainServeBaseURL: "http://host.docker.internal:7701"   # ← 连接 brain-v3
  safeMode: false
  easymvpBrain:
    mode: "local-sidecar"                                   # ← local-sidecar 或 remote-service
    localEndpoint: "http://host.docker.internal:7801"       # ← easymvp-brain sidecar（如使用）
    remoteEndpoint: ""
    apiKey: ""
    timeout: "30s"
    maxTurns: 6
```

### 5.3 本地裸机运行配置（不通过 Docker）

```powershell
cd C:\Users\Public\project\easymvp\apps\core

# 方式一：CLI flag 覆盖（推荐冒烟测试用）
go run main.go `
  --port 8000 `
  --brain-serve-base-url http://127.0.0.1:7701 `
  --data-root ./var `
  --db-path ./var/data/easymvp.db `
  --migration-path ./manifest/migrations

# 方式二：本地 
---
data/easymvp.db `
  --migration-path ./manifest/migrations

# 方式二：本地 config.yaml（同 docker/dev/config.yaml 格式，放 ./manifest/config/config.yaml）
# 方式三：环境变量（复制 apps/core/.env.example 为 .env 并修改）
```

### 5.4 配置项详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `easymvp.brainServeBaseURL` | string | `http://127.0.0.1:7701` | brain-v3 serve HTTP 地址 |
| `easymvp.easymvpBrain.mode` | string | `local-sidecar` | `local-sidecar`（本地 sidecar）或 `remote-service`（远程服务） |
| `easymvp.easymvpBrain.localEndpoint` | string | `http://host.docker.internal:7801` | easymvp-brain sidecar HTTP 地址（JSON-RPC） |
| `easymvp.easymvpBrain.remoteEndpoint` | string | `""` | 远程 brain 服务端点 |
| `easymvp.easymvpBrain.apiKey` | string | `""` | Bearer Token（远程模式使用） |
| `easymvp.easymvpBrain.timeout` | duration | `30s` | brain 调用超时 |
---
  --migration-path ./manifest/migrations

# 方式二：本地 config.yaml（同 docker/dev/config.yaml 格式，放 ./manifest/config/config.yaml）
# 方式三：环境变量（复制 apps/core/.env.example 为 .env 并修改）
```

### 5.4 配置项详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `easymvp.brainServeBaseURL` | string | `http://127.0.0.1:7701` | brain-v3 serve HTTP 地址 |
| `easymvp.easymvpBrain.mode` | string | `local-sidecar` | `local-sidecar`（本地 sidecar）或 `remote-service`（远程服务） |
| `easymvp.easymvpBrain.localEndpoint` | string | `http://host.docker.internal:7801` | easymvp-brain sidecar HTTP 地址（JSON-RPC） |
| `easymvp.easymvpBrain.remoteEndpoint` | string | `""` | 远程 brain 服务端点 |
| `easymvp.easymvpBrain.apiKey` | string | `""` | Bearer Token（远程模式使用） |
| `easymvp.easymvpBrain.timeout` | duration | `30s` | brain 调用超时 |
| `easymvp.easymv
---
变量（复制 apps/core/.env.example 为 .env 并修改）
```

### 5.4 配置项详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `easymvp.brainServeBaseURL` | string | `http://127.0.0.1:7701` | brain-v3 serve HTTP 地址 |
| `easymvp.easymvpBrain.mode` | string | `local-sidecar` | `local-sidecar`（本地 sidecar）或 `remote-service`（远程服务） |
| `easymvp.easymvpBrain.localEndpoint` | string | `http://host.docker.internal:7801` | easymvp-brain sidecar HTTP 地址（JSON-RPC） |
| `easymvp.easymvpBrain.remoteEndpoint` | string | `""` | 远程 brain 服务端点 |
| `easymvp.easymvpBrain.apiKey` | string | `""` | Bearer Token（远程模式使用） |
| `easymvp.easymvpBrain.timeout` | duration | `30s` | brain 调用超时 |
| `easymvp.easymvpBrain.maxTurns` | int | `6` | 单次 brain 执行最大轮数 |

### 5.5 环境变量（`apps/core/.env.example` 中定义）

```powershell
# 用于 brain resu
---
ps/core/.env.example 为 .env 并修改）
```

### 5.4 配置项详解

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `easymvp.brainServeBaseURL` | string | `http://127.0.0.1:7701` | brain-v3 serve HTTP 地址 |
| `easymvp.easymvpBrain.mode` | string | `local-sidecar` | `local-sidecar`（本地 sidecar）或 `remote-service`（远程服务） |
| `easymvp.easymvpBrain.localEndpoint` | string | `http://host.docker.internal:7801` | easymvp-brain sidecar HTTP 地址（JSON-RPC） |
| `easymvp.easymvpBrain.remoteEndpoint` | string | `""` | 远程 brain 服务端点 |
| `easymvp.easymvpBrain.apiKey` | string | `""` | Bearer Token（远程模式使用） |
| `easymvp.easymvpBrain.timeout` | duration | `30s` | brain 调用超时 |
| `easymvp.easymvpBrain.maxTurns` | int | `6` | 单次 brain 执行最大轮数 |

### 5.5 环境变量（`apps/core/.env.example` 中定义）

```powershell
# 用于 brain resume CLI 命令解析
$env:
---
-RPC） |
| `easymvp.easymvpBrain.remoteEndpoint` | string | `""` | 远程 brain 服务端点 |
| `easymvp.easymvpBrain.apiKey` | string | `""` | Bearer Token（远程模式使用） |
| `easymvp.easymvpBrain.timeout` | duration | `30s` | brain 调用超时 |
| `easymvp.easymvpBrain.maxTurns` | int | `6` | 单次 brain 执行最大轮数 |

### 5.5 环境变量（`apps/core/.env.example` 中定义）

```powershell
# 用于 brain resume CLI 命令解析
$env:EASYMVP_BRAIN_CMD = "easymvp-brain"           # brain 二进制路径
$env:EASYMVP_BRAIN_ARGS = '["--profile", "default", "--config", "./brain-config.yaml"]'

# Brain serve 连接地址
$env:EASYMVP_BRAIN_SERVE_BASE_URL = "http://127.0.0.1:7701"

# Brain 模式
$env:EASYMVP_BRAIN_MODE = "local"
$env:EASYMVP_BRAIN_LOCAL_ENDPOINT = "http://127.0.0.1:7701"
$env:EASYMVP_BRAIN_REMOTE_ENDPOINT = "https://brain.easymvp.example.com"
$env:EASYMVP_BRAIN_API_KEY 
---
ktop Settings → Resources → Network → 确认 host.docker.internal 可用 |
| sidecar 工具未注册 | brain-quant-sidecar.exe / brain-data-sidecar.exe 不在 `~/.brain/bin/` | 复制或构建 sidecars 到正确位置 |

---

## 10. 端口对照表

| 服务 | 默认端口 | 用途 | 配置键 |
|------|---------|------|--------|
| brain-v3 serve | `7701` | Brain Kernel HTTP API | `--listen` |
| EasyMVP Core | `8000` | EasyMVP REST API | `--port` / `server.address` |
| easymvp-brain sidecar | `7801` | EasyMVP brain JSON-RPC（sidecar 模式） | `easymvpBrain.localEndpoint` |

---

## 附录：Minimal cURL 冒烟脚本（Bash/PowerShell 通用）

保存为 `smoke.sh`（Linux/macOS/Git Bash）：

```bash
#!/usr/bin/env bash
set -euo pipefail

BRAIN_URL="http://127.0.0.1:7701"
CORE_URL="http://127.0.0.1:8000"

echo "== 1. brain-v3 health =="
curl -sf "$BRAIN_URL/health" | jq .

echo "== 2. easymvp-core healthz =="
c
---
 → Network → 确认 host.docker.internal 可用 |
| sidecar 工具未注册 | brain-quant-sidecar.exe / brain-data-sidecar.exe 不在 `~/.brain/bin/` | 复制或构建 sidecars 到正确位置 |

---

## 10. 端口对照表

| 服务 | 默认端口 | 用途 | 配置键 |
|------|---------|------|--------|
| brain-v3 serve | `7701` | Brain Kernel HTTP API | `--listen` |
| EasyMVP Core | `8000` | EasyMVP REST API | `--port` / `server.address` |
| easymvp-brain sidecar | `7801` | EasyMVP brain JSON-RPC（sidecar 模式） | `easymvpBrain.localEndpoint` |

---

## 附录：Minimal cURL 冒烟脚本（Bash/PowerShell 通用）

保存为 `smoke.sh`（Linux/macOS/Git Bash）：

```bash
#!/usr/bin/env bash
set -euo pipefail

BRAIN_URL="http://127.0.0.1:7701"
CORE_URL="http://127.0.0.1:8000"

echo "== 1. brain-v3 health =="
curl -sf "$BRAIN_URL/health" | jq .

echo "== 2. easymvp-core healthz =="
curl -sf "$CORE_U