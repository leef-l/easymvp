# Docker 开发环境说明

本文以当前代码为准，说明 EasyMVP 开发环境的目录、启动方式、服务矩阵与常见操作。

## 目录说明

- `docker/build/`：开发/生产镜像 Dockerfile
- `docker/dev/`：开发环境 compose、启动脚本、`.env`
- `docker/nginx/`：Nginx 配置
- `docker/prod/`：生产 compose 与启动脚本

开发环境最常用的文件：

- `docker/dev/docker-compose.cn.yml`：当前推荐的开发 compose
- `docker/dev/docker-compose.yml`：通用 compose 变体
- `docker/dev/docker-compose.infra.yml`：`--local` 模式使用的基础设施 compose
- `docker/dev/.env`：开发环境变量源文件
- `docker/dev/compose.ps1`：Windows 推荐入口
- `docker/dev/build.ps1`：预构建开发镜像
- `docker/dev/local-dev.ps1`：Docker 基础设施 + 本地热重启模式
- `docker/dev/start-go-app.sh`：后端容器启动脚本
- `docker/dev/docker-web-start.sh`：前端容器启动脚本

## 当前服务矩阵

默认 `docker-compose.cn.yml` 启动以下服务：

| 服务 | 默认启动 | 说明 |
|------|----------|------|
| `mysql` | 是 | 开发数据库 |
| `system` | 是 | 系统管理服务 |
| `ai` | 是 | AI 配置与手工任务服务 |
| `mvp` | 是 | MVP / Workflow V2 服务 |
| `web` | 是 | `vue-vben-admin` 前端 |
| `openhands-runtime` | 否 | 仅在 `ai-runtime` profile 下启动 |

当前默认开发 compose 里没有 `redis` 服务。

## 环境变量同步

开发环境以 `docker/dev/.env` 为准。

`compose.ps1` 和 `build.ps1` 在执行前都会先把：

```powershell
docker/dev/.env
```

同步到：

```powershell
admin-go/.env
```

这样容器运行和后端本地读取的环境变量保持一致。

## 推荐启动方式

### Windows

仓库根目录执行：

```powershell
.\docker\dev\compose.ps1
```

这条命令会：

1. 同步 `docker/dev/.env` 到 `admin-go/.env`
2. 构建后端开发镜像 `easymvp-admin-go-dev:latest`
3. 构建前端开发镜像 `easymvp-web-dev:latest`
4. 启动 `mysql`、`system`、`ai`、`mvp`、`web`

如果需要 `OpenHands` runtime：

```powershell
.\docker\dev\compose.ps1 --profile ai-runtime up -d
```

如果只想预打包镜像：

```powershell
.\docker\dev\build.ps1
```

如果想使用“Docker 基础设施 + 本地 `gf run`/`pnpm dev` 热重启”模式：

```powershell
.\docker\dev\compose.ps1 --local
```

常用附加参数：

- `.\docker\dev\compose.ps1 --local --no-web`
- `.\docker\dev\compose.ps1 --local --no-infra`
- `.\docker\dev\compose.ps1 --local --stop-infra`

### Linux / macOS

手动执行与默认脚本等价的流程：

```bash
cp docker/dev/.env admin-go/.env
docker build -f docker/build/Dockerfile.admin-go.dev admin-go -t easymvp-admin-go-dev:latest
docker build -f docker/build/Dockerfile.web.dev vue-vben-admin -t easymvp-web-dev:latest
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml up -d
```

如果需要 `OpenHands` runtime：

```bash
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml --profile ai-runtime up -d
```

## 默认端口

- MySQL：`41001`
- system：`41002`
- ai：`41003`
- mvp：`41004`
- web：`41005`

访问地址：

- 前端：`http://127.0.0.1:41005`
- system：`http://127.0.0.1:41002`
- ai：`http://127.0.0.1:41003`
- mvp：`http://127.0.0.1:41004`

## 常用命令

查看容器：

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

查看日志：

```powershell
docker logs -f easymvp-system
docker logs -f easymvp-ai
docker logs -f easymvp-mvp
docker logs -f easymvp-web
docker logs -f easymvp-mysql
```

停止开发环境：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml down
```

重启开发环境：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml up -d
```

启用 `OpenHands` runtime：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml --profile ai-runtime up -d
```

只重启单个服务：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml restart mvp
```

## 常见问题

### 1. 前端没起来

当前默认命令会直接启动 `web`，不需要再带 `frontend` profile。

如果 `web` 没起来，先看：

```powershell
docker logs -f easymvp-web
```

常见原因是前端镜像构建失败、依赖安装失败或本机端口冲突。

### 2. `openhands-runtime` 没启动

`openhands-runtime` 仍然挂在 `ai-runtime` profile 下，需要显式带：

```powershell
.\docker\dev\compose.ps1 --profile ai-runtime up -d
```

### 3. 修改了 `docker/dev/.env`，后端没生效

因为 `admin-go/.env` 也会被读取。建议始终通过 `compose.ps1` / `build.ps1` 启动，让脚本先同步环境变量。

### 4. 旧容器或孤儿容器干扰

可以执行：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml down --remove-orphans
```

然后重新启动。

### 5. 数据库端口冲突

默认 MySQL 映射为 `41001`。如果本机已占用，修改 `docker/dev/.env` 中对应端口后重新启动即可。
