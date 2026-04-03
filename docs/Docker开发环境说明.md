# Docker 开发环境说明

本文说明 EasyMVP 当前开发环境 Docker 目录、启动方式、容器命名规则与常见操作。

## 目录说明

当前 Docker 相关目录按职责拆分：

- `docker/build/`：构建镜像用的 Dockerfile
- `docker/dev/`：开发环境 compose、启动脚本、开发环境变量
- `docker/mysql/`：MySQL 初始化与升级 SQL
- `docker/nginx/`：生产环境 Nginx 配置
- `docker/prod/`：生产环境 compose 与启动脚本

开发环境主要使用以下文件：

- `docker/dev/docker-compose.cn.yml`：中国网络环境推荐使用
- `docker/dev/docker-compose.yml`：通用开发 compose
- `docker/dev/.env`：开发环境变量源文件
- `docker/dev/build.ps1`：Windows 下预打包开发镜像
- `docker/dev/compose.ps1`：Windows 下推荐启动入口
- `docker/dev/start-go-app.sh`：后端容器启动脚本
- `docker/dev/docker-web-start.sh`：前端容器启动脚本

## 容器命名规则

开发环境 Compose 项目名统一为 `easymvp`，容器名如下：

- `easymvp-mysql`
- `easymvp-system`
- `easymvp-ai`
- `easymvp-mvp`
- `easymvp-web`

这样查看容器、日志、网络时名字会更统一。

## 环境变量说明

开发环境以 `docker/dev/.env` 为准。

启动前，`docker/dev/compose.ps1` 会先把：

```powershell
docker/dev/.env
```

同步到：

```powershell
admin-go/.env
```

这样后端本地配置与 Docker 开发配置保持一致。

## 推荐启动方式

### Windows

在仓库根目录执行：

```powershell
.\docker\dev\compose.ps1
```

默认等价于：

```powershell
docker build -f docker/build/Dockerfile.admin-go.dev admin-go -t easymvp-admin-go-dev:latest
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml up -d
```

如果要先预打包开发镜像，执行：

```powershell
.\docker\dev\build.ps1
```

默认等价于：

```powershell
docker build -f docker/build/Dockerfile.admin-go.dev admin-go -t easymvp-admin-go-dev:latest
docker build -f docker/build/Dockerfile.web.dev vue-vben-admin -t easymvp-web-dev:latest
```

如果需要启用前端服务：

```powershell
.\docker\dev\compose.ps1 --profile frontend up -d
```

如果需要重建镜像：

```powershell
.\docker\dev\build.ps1
.\docker\dev\compose.ps1 --profile frontend up -d
```

### Linux / macOS

在仓库根目录执行：

```bash
cp docker/dev/.env admin-go/.env
docker build -f docker/build/Dockerfile.admin-go.dev admin-go -t easymvp-admin-go-dev:latest
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml up -d
```

如果要先预打包开发镜像：

```bash
cp docker/dev/.env admin-go/.env
docker build -f docker/build/Dockerfile.admin-go.dev admin-go -t easymvp-admin-go-dev:latest
docker build -f docker/build/Dockerfile.web.dev vue-vben-admin -t easymvp-web-dev:latest
```

启用前端：

```bash
cp docker/dev/.env admin-go/.env
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml --profile frontend up -d
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

## `.yml` 和 `.sh` 的区别

两类文件职责不同：

- `docker-compose*.yml`：定义有哪些容器、端口、挂载、依赖关系、环境变量
- `*.sh`：定义容器启动后实际执行什么命令

当前对应关系：

- `start-go-app.sh`：`system` / `ai` / `mvp` 容器内部启动 Go 服务
- `docker-web-start.sh`：`web` 容器内部启动前端开发服务

也就是说，`.yml` 负责“编排容器”，`.sh` 负责“容器里跑什么”。

## 常用命令

查看容器：

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | Select-String "easymvp|NAMES"
```

查看某个服务日志：

```powershell
docker logs -f easymvp-system
docker logs -f easymvp-ai
docker logs -f easymvp-mvp
docker logs -f easymvp-web
docker logs -f easymvp-mysql
```

停止开发环境：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml --profile frontend down
```

重启开发环境：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml --profile frontend up -d
```

只重启某个服务：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml restart system
```

## 常见问题

### 1. 前端没启动

`web` 服务放在 `frontend` profile 下，启动时要带：

```powershell
--profile frontend
```

否则只会启动后端和数据库，不会启动前端容器。

### 2. 修改了 `docker/dev/.env`，但后端没生效

因为 `admin-go/.env` 也会被使用。建议始终通过：

```powershell
.\docker\dev\compose.ps1
```

启动，让脚本先同步一次环境变量。

### 3. 容器名字还是旧的

先执行：

```powershell
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml --profile frontend down --remove-orphans
```

再重新执行启动命令，旧容器就会被清掉。

### 4. 数据库端口冲突

默认 MySQL 映射为 `41001`。如果本机已占用，可修改 `docker/dev/.env` 中对应端口后重新启动。

