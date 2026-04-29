# EasyMVP + brain-v3 本地端到端冒烟执行手册

> 适用对象：登月者4662  
> 目标：在本地机器一键跑通 EasyMVP Core → brain-v3 → sidecars 全链路  
> 项目根目录：`C:\Users\Public\project\easymvp`

---

## 1. 架构速览

```
EasyMVP Core (HTTP :8000)
    │  brainServeBaseURL: http://127.0.0.1:7701
    ▼
brain-v3 Serve (HTTP :7701)
    │  加载 sidecars 为插件
    ├──► quant sidecar ──► 交易工具 (quant.*)
    └──► data sidecar ──► 数据工具 (data.*)
```

- **brain-v3 serve**：Brain Kernel HTTP 服务，默认监听 `127.0.0.1:7701`
- **EasyMVP Core**：GoFrame 应用，默认监听 `:8000`
- **sidecars**（quant/data）：被 brain-v3 动态加载的插件，不独立暴露端口

---

## 2. 前置条件

| 依赖 | 版本要求 | 验证命令 |
|------|---------|---------|
| Go | ≥ 1.22 | `go version` |
| Node.js + pnpm | ≥ 18 / ≥ 8 | `node -v && pnpm -v` |
| Git | 任意 | `git --version` |
| Windows | 10/11 | — |

**必须设置的环境变量**（brain-v3 需要 LLM Provider）：

```powershell
# PowerShell 示例（brain-v3 运行时需要）
$env:OPENAI_API_KEY = "sk-xxxxxxxxxxxxxxxxxxxxxxxx"
# 或
$env:ANTHROPIC_API_KEY = "sk-ant-xxxxxxxxxxxxxxxxxxxxxxxx"
```

> brain-v3 的 LLM Provider 配置优先从环境变量读取；若未设置，serve 会启动但 execute run 会失败。

---

## 3. brain-v3 配置清单（在 EasyMVP 项目内）

### 3.1 项目内配置位置总览

| 配置类型 | 文件位置 | 说明 |
|---------|---------|------|
| Docker Compose | `docker-compose.dev.yml` | 开发环境容器编排 |
| Docker 配置 | `docker/dev/config.yaml` | 容器内 core 配置（含 brainServeBaseURL） |
| Docker 构建 | `docker/dev/Dockerfile.core` | core 镜像构建脚本 |
| 环境变量示例 | `apps/core/.env.example` | 本地开发环境变量模板 |
| K8s ConfigMap | `apps/core/manifest/deploy/kustomize/base/configmap.yaml` | Kubernetes 部署配置 |
| K8s Dev Overlay | `apps/core/manifest/deploy/kustomize/overlays/develop/configmap.yaml` | 开发环境 K8s 配置 |
| K8s Prod Overlay | `apps/core/manifest/deploy/kustomize/overlays/production/configmap.yaml` | 生产环境 K8s 配置 |
| 启动脚本 | `dev_docker.bat` | Windows Docker 开发环境一键启动 |

### 3.2 brain-v3 服务端配置

| 配置项 | 位置 | 默认值 | 说明 |
|--------|------|--------|------|
| listen address | CLI flag `--listen` | `127.0.0.1:7701` | serve HTTP 监听地址 |
| config path | 自动探测 | `~/.brain/config.yaml` | 用户级配置文件 |
| runtime data dir | 自动探测 | `~/.brain/` | SQLite 数据库、学习数据、hooks |
| sidecars | `.brain/bin/` 目录 | — | quant/data 等 sidecar 可执行文件 |

**配置文件示例**（`~/.brain/config.yaml`，如无则 serve 使用默认）：

```yaml
# ~/.brain/config.yaml — brain-v3 serve 可选配置
server:
  address: "127.0.0.1:7701"

llm:
  provider: openai          # 或 anthropic
  model: gpt-4o
  api_key: ${OPENAI_API_KEY}

brains:
  quant:
    enabled: true
  data:
    enabled: true
```

> 实际上 brain-v3 serve 的大部分配置通过 **环境变量** + **flag** 生效，`config.yaml` 用于高级参数（如 org-policy、learning、hooks）。

### 3.3 sidecar 部署

brain-v3 自动从以下路径加载 sidecars：

```
~/.brain/bin/
├── brain-quant-sidecar.exe   (quant 工具集)
├── brain-data-sidecar.exe    (数据工具集)
└── ...
```

**从 brain 项目构建 sidecars**（可选，如果 `~/.brain/bin/` 已存在可跳过）：

```powershell
# 在 brain 项目根目录
cd C:\Users\Public\project\brain

# 构建 quant sidecar
go build -o ~/.brain/bin/brain-quant-sidecar.exe ./cmd/brain-quant-sidecar

# 构建 data sidecar
go build -o ~/.brain/bin/brain-data-sidecar.exe ./cmd/brain-data-sidecar
```

---

## 4. 启动 brain-v3 的 Exact 步骤

### 4.1 步骤一：确保 sidecars 就位

```powershell
# 检查 sidecars 是否存在
test-path ~/.brain/bin/brain-quant-sidecar.exe
test-path ~/.brain/bin/brain-data-sidecar.exe

# 如果不存在，从 brain 项目 dist 复制或构建
Copy-Item C:\Users\Public\project\brain\dist\brain-quant-sidecar.exe ~/.brain\bin\ -ErrorAction SilentlyContinue
Copy-Item C:\Users\Public\project\brain\dist\brain-data-sidecar.exe ~/.brain\bin\ -ErrorAction SilentlyContinue
```

### 4.2 步骤二：启动 brain serve

```powershell
cd C:\Users\Public\project\brain\dist

# 基础启动（前台运行，调试用）
.\brain.exe serve --listen 127.0.0.1:7701

# 后台启动（PowerShell）
Start-Process -FilePath .\brain.exe -ArgumentList "serve","--listen","127.0.0.1:7701" -WindowStyle Hidden

# 验证启动
Invoke-RestMethod -Uri http://127.0.0.1:7701/health -Method GET
```

预期输出（健康检查）：

```json
{"status":"ok"}
```

> brain serve 默认端口 **7701**。API 前缀无版本号（如 `/v1/runs`）。

### 4.3 步骤三：验证 sidecars 加载

```powershell
# 查看 brain serve 日志输出中是否出现 sidecar 注册信息
# 或使用 dashboard API（如启用）
Invoke-RestMethod -Uri http://127.0.0.1:7701/v1/tools -Method GET 2>$null
```

### 4.4 步骤四（可选）— Docker Compose 一键启动 EasyMVP

项目根目录提供 `docker-compose.dev.yml` 和 `dev_docker.bat`，可直接启动 EasyMVP Core 容器：

```powershell
cd C:\Users\Public\project\easymvp

# 启动 Docker 开发环境（自动构建 core Linux 二进制并启动容器）
.\dev_docker.bat

# 查看日志
.\dev_docker.bat logs

# 停止
.\dev_docker.bat down

# 清理
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
| 环境变量模板 | `apps/core/.env.example` | 复制为 `.env` 并修改 |
| K8s ConfigMap | `apps/core/manifest/deploy/kustomize/base/configmap.yaml` | K8s 部署时生效 |
| K8s Dev Overlay | `apps/core/manifest/deploy/kustomize/overlays/develop/configmap.yaml` | 开发环境 K8s |
| 运行时覆盖 | CLI flag `--brain-serve-base-url` | 命令行 |

### 5.2 Docker 开发配置（`docker/dev/config.yaml`）

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
  --brain-serve-base-url http://127.0.0.1:7701 `
  --data-root ./var `
  --db-path ./var/data/easymvp.db `
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
$env:EASYMVP_BRAIN_API_KEY = ""
$env:EASYMVP_BRAIN_TIMEOUT = "30"
$env:EASYMVP_BRAIN_MAX_TURNS = "6"
```

---

## 6. 一键启动脚本（PowerShell）

创建文件 `smoke-start.ps1`：

```powershell
#Requires -Version 5.1
# EasyMVP + brain-v3 端到端冒烟启动脚本

$ErrorActionPreference = "Stop"

# ─── 路径配置 ───
$BrainDist   = "C:\Users\Public\project\brain\dist"
$EasymvpCore = "C:\Users\Public\project\easymvp\apps\core"
$SmokeVar    = "$PSScriptRoot\smoke-var"

# ─── 确保 LLM Key ───
if (-not $env:OPENAI_API_KEY -and -not $env:ANTHROPIC_API_KEY) {
    Write-Host "[ERROR] 请先设置 OPENAI_API_KEY 或 ANTHROPIC_API_KEY" -ForegroundColor Red
    exit 1
}

# ─── 确保 sidecars ───
$BrainBin = "$env:USERPROFILE\.brain\bin"
if (-not (Test-Path $BrainBin)) { New-Item -ItemType Directory -Path $BrainBin -Force | Out-Null }
@("brain-quant-sidecar.exe","brain-data-sidecar.exe") | ForEach-Object {
    $src = Join-Path $BrainDist $_
    $dst = Join-Path $BrainBin $_
    if (Test-Path $src -and -not (Test-Path $dst)) {
        Copy-Item $src $dst
        Write-Host "[INFO] 复制 sidecar: $_"
    }
}

# ─── 启动 brain serve ───
Write-Host "[INFO] 启动 brain serve on 127.0.0.1:7701 ..."
$brainJob = Start-Job -ScriptBlock {
    param($dist)
    & "$dist\brain.exe" serve --listen 127.0.0.1:7701 2>&1
} -ArgumentList $BrainDist

# 等待 brain serve 就绪
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 1
    try {
        $r = Invoke-RestMethod -Uri http://127.0.0.1:7701/health -Method GET -TimeoutSec 2
        if ($r.status -eq "ok") { Write-Host "[OK] brain serve ready"; break }
    } catch { }
}
if ($i -eq 30) { Write-Host "[ERROR] brain serve 启动超时"; exit 1 }

# ─── 启动 easymvp-core ───
Write-Host "[INFO] 启动 easymvp-core on :8000 ..."
$coreJob = Start-Job -ScriptBlock {
    param($core,$var)
    $db = "$var\data\easymvp.db"
    New-Item -ItemType Directory -Path "$var\data" -Force | Out-Null
    & go run "$core\main.go" `
        --port 8000 `
        --brain-serve-base-url http://127.0.0.1:7701 `
        --data-root $var `
        --db-path $db `
        --migration-path "$core\manifest\migrations" 2>&1
} -ArgumentList $EasymvpCore,$SmokeVar

# 等待 easymvp-core 就绪
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 1
    try {
        $r = Invoke-RestMethod -Uri http://127.0.0.1:8000/api/v3/system/healthz -Method GET -TimeoutSec 2
        if ($r.data.startup.ready) { Write-Host "[OK] easymvp-core ready"; break }
    } catch { }
}
if ($i -eq 30) { Write-Host "[ERROR] easymvp-core 启动超时"; exit 1 }

Write-Host ""
Write-Host "═══════════════════════════════════════════════"
Write-Host "  冒烟环境已就绪"
Write-Host "  brain-v3:  http://127.0.0.1:7701"
Write-Host "  easymvp:   http://127.0.0.1:8000"
Write-Host "═══════════════════════════════════════════════"
Write-Host ""
Write-Host "按 Enter 停止所有服务 ..."
Read-Host

# ─── 清理 ───
Stop-Job $brainJob -ErrorAction SilentlyContinue
Remove-Job $brainJob -ErrorAction SilentlyContinue
Stop-Job $coreJob -ErrorAction SilentlyContinue
Remove-Job $coreJob -ErrorAction SilentlyContinue
Write-Host "[INFO] 服务已停止"
```

**使用方法**：

```powershell
# 设置 LLM Key（必须）
$env:OPENAI_API_KEY = "sk-xxxx"

# 一键启动
.\smoke-start.ps1
```

---

## 7. 最小冒烟验证步骤

### 7.1 Step 1 — 健康检查双端

```powershell
# brain-v3 健康
Invoke-RestMethod http://127.0.0.1:7701/health | ConvertTo-Json

# EasyMVP Core 健康
Invoke-RestMethod http://127.0.0.1:8000/api/v3/system/healthz | ConvertTo-Json
```

预期响应：

```json
// brain-v3
{"status":"ok"}

// easymvp-core
{
  "data": {
    "status": "ok",
    "startup": { "ready": true, "status": "ok" },
    "runtime_status": "ok"
  }
}
```

### 7.2 Step 2 — 验证 brain-v3 可直接创建 run

```powershell
$body = @{
    prompt   = "Say hello"
    brain    = "central"
    max_turns = 3
} | ConvertTo-Json -Depth 3

$response = Invoke-RestMethod `
    -Uri http://127.0.0.1:7701/v1/runs `
    -Method POST `
    -ContentType "application/json" `
    -Body $body

$response | ConvertTo-Json
$runId = $response.run_id

# 轮询状态
while ($true) {
    $state = Invoke-RestMethod -Uri "http://127.0.0.1:7701/v1/runs/$runId" -Method GET
    Write-Host "Status: $($state.status)"
    if ($state.status -in @("completed","succeeded","failed","error")) { break }
    Start-Sleep -Seconds 2
}
```

### 7.3 Step 3 — 通过 EasyMVP 启动 brain run（完整链路）

```powershell
# 先创建一个项目（通过 easymvp API，或已有项目则跳过）
# 假设已有 project_id 和 task_id

$projectId = "proj_smoke_001"
$taskId    = "task_smoke_001"

# 启动 brain run（通过 EasyMVP Core API）
$body = @{
    project_id = $projectId
    task_id    = $taskId
    brain_kind = "central"
    prompt     = "Evaluate this project: hello-world repo"
    max_turns  = 3
} | ConvertTo-Json -Depth 3

$response = Invoke-RestMethod `
    -Uri http://127.0.0.1:8000/api/v3/runtime/start-run `
    -Method POST `
    -ContentType "application/json" `
    -Body $body

$response | ConvertTo-Json
$bindingId = $response.data.binding_id

# 轮询 binding 状态
for ($i = 0; $i -lt 60; $i++) {
    $binding = Invoke-RestMethod `
        -Uri "http://127.0.0.1:8000/api/v3/runtime/run-bindings/$bindingId" `
        -Method GET
    Write-Host "[$i] Status: $($binding.data.run_status)"
    if ($binding.data.run_status -in @("run_succeeded","run_failed","run_cancelled","run_denied","run_unsupported")) {
        break
    }
    Start-Sleep -Seconds 3
}
```

**冒烟通过标准**：

1. `brain-v3 /health` 返回 `ok`
2. `easymvp-core /api/v3/system/healthz` 返回 `startup.ready = true`
3. 直接调用 `POST /v1/runs` 成功创建 run 并返回 `run_id`
4. 通过 EasyMVP `POST /api/v3/runtime/start-run` 成功创建 binding，且状态最终变为 `run_succeeded` 或 `run_failed`（非 stuck 在 `run_pending`）

---

## 8. 已有脚本/测试清单

### 8.1 EasyMVP 项目内已有脚本（`C:\Users\Public\project\easymvp`）

| 文件 | 用途 | 平台 |
|------|------|------|
| `dev_docker.bat` | Windows Docker 开发环境一键启动/停止/查看日志 | Windows |
| `scripts/verify-core-health.sh` | Linux 下编译 core 并探测 healthz | Linux/macOS |
| `scripts/local-verify-apps-core-desktop.bat` | Windows 本地验证 core + desktop 构建 | Windows |
| `scripts/local-verify-completion-verdict-authority.bat` | 本地验证 completion verdict authority | Windows |
| `scripts/verify-apps-core-query-plans.sh` | 验证 core query plans | Linux/macOS |
| `scripts/verify-apps-core-release.sh` | 验证 core release | Linux/macOS |
| `scripts/easymvp-backup-snapshot.sh` | 备份 snapshot | Linux/macOS |

### 8.2 EasyMVP 项目内已有测试

| 文件 | 测试内容 |
|------|---------|
| `apps/core/internal/service/runtime_test.go` | runtime health check 基础测试 |
| `apps/core/internal/service/runtime_support_client_test.go` | resume 命令解析、env override 测试 |
| `apps/core/internal/service/easymvp_brain_test.go` | EasyMVP brain client 配置解析与执行测试 |
| `apps/core/internal/service/runtime_support_execution_test.go` | runtime 执行测试 |
| `apps/core/internal/service/runtime_support_idempotency_test.go` | idempotency 测试 |
| `apps/core/internal/service/startup_diagnostics_test.go` | 启动诊断测试 |
| `apps/core/internal/service/acceptance_view_logic_test.go` | acceptance view 逻辑测试 |
| `apps/core/internal/service/completion_verdict_control_test.go` | completion verdict 控制测试 |

### 8.3 brain 项目内已有测试（`C:\Users\Public\project\brain`）

| 文件 | 测试内容 |
|------|---------|
| `cmd/brain/cmd_serve_test.go` | serve HTTP 路由、创建 run、状态查询测试 |
| `sdk/sidecar/serve_test.go` | sidecar HTTP serve 测试 |

### 8.4 执行已有测试

```powershell
# EasyMVP core 测试
cd C:\Users\Public\project\easymvp\apps\core
go test ./internal/service/... -run "Runtime|Brain" -v

# brain-v3 serve 测试
cd C:\Users\Public\project\brain
go test ./cmd/brain/... -run "Serve" -v
```

---

## 9. 常见问题速查

| 现象 | 原因 | 解决 |
|------|------|------|
| `brain serve` 启动后 `/health` 不通 | 端口被占用或 listen 地址错 | `netstat -ano \| findstr 7701` 检查占用 |
| `easymvp-core` healthz `runtime_status: degraded` | brainServeBaseURL 连不上 brain-v3 | 确认 `--brain-serve-base-url` 与 brain serve `--listen` 一致 |
| `start-run` 后 stuck in `run_pending` | LLM Key 未设置或 brain sidecar 未加载 | 检查 `$env:OPENAI_API_KEY`，确认 `~/.brain/bin/` 有 sidecars |
| Docker 模式下 `host.docker.internal` 不通 | Windows Docker Desktop 未启用 host gateway | Docker Desktop Settings → Resources → Network → 确认 host.docker.internal 可用 |
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
curl -sf "$CORE_URL/api/v3/system/healthz" | jq '.data.startup.ready'

echo "== 3. create run directly on brain-v3 =="
RUN_BODY='{"prompt":"Say hello","brain":"central","max_turns":2}'
RUN_RESP=$(curl -sf -X POST "$BRAIN_URL/v1/runs" -H "Content-Type: application/json" -d "$RUN_BODY")
echo "$RUN_RESP" | jq .
RUN_ID=$(echo "$RUN_RESP" | jq -r '.run_id')

echo "== 4. poll brain-v3 run status =="
for i in {1..30}; do
    STATE=$(curl -sf "$BRAIN_URL/v1/runs/$RUN_ID")
    STATUS=$(echo "$STATE" | jq -r '.status')
    echo "  [$i] status=$STATUS"
    [[ "$STATUS" == "completed" || "$STATUS" == "succeeded" || "$STATUS" == "failed" ]] && break
    sleep 2
done

echo "== smoke done =="
```

PowerShell 版本 `smoke.ps1`：

```powershell
$brain = "http://127.0.0.1:7701"
$core  = "http://127.0.0.1:8000"

Write-Host "== 1. brain-v3 health =="
(Invoke-RestMethod "$brain/health").status

Write-Host "== 2. easymvp-core healthz =="
(Invoke-RestMethod "$core/api/v3/system/healthz").data.startup.ready

Write-Host "== 3. create run on brain-v3 =="
$body = '{"prompt":"Say hello","brain":"central","max_turns":2}'
$resp = Invoke-RestMethod -Uri "$brain/v1/runs" -Method POST -ContentType "application/json" -Body $body
$resp | ConvertTo-Json
$runId = $resp.run_id

Write-Host "== 4. poll run status =="
for ($i = 1; $i -le 30; $i++) {
    $state = Invoke-RestMethod -Uri "$brain/v1/runs/$runId"
    Write-Host "  [$i] status=$($state.status)"
    if ($state.status -in @("completed","succeeded","failed")) { break }
    Start-Sleep 2
}

Write-Host "== smoke done =="
```

---

*文档生成时间：2026-04-25*  
*覆盖范围：brain-v3 serve (port 7701) ↔ EasyMVP Core (port 8000) ↔ sidecars (quant/data)*
