# EasyMVP

EasyMVP 当前由 `system`、`ai`、`mvp` 三个 Go 服务和 `vue-vben-admin` 前端组成。运行说明、使用文档和设计文档统一放在 `docs/` 目录。

先看 [docs/README.md](docs/README.md)。该索引已经区分“当前有效文档”和“历史设计稿”。

## 本地 Docker 开发

Windows 推荐入口：

```powershell
.\docker\dev\compose.ps1
```

默认会：

- 同步 `docker/dev/.env` 到 `admin-go/.env`
- 构建 `easymvp-admin-go-dev:latest` 和 `easymvp-web-dev:latest`
- 启动 `mysql`、`system`、`ai`、`mvp`、`web`

如果还需要 `OpenHands` runtime：

```powershell
.\docker\dev\compose.ps1 --profile ai-runtime up -d
```

如需单独预打包镜像：

```powershell
.\docker\dev\build.ps1
```

如需“Docker 基础设施 + 本地热重启”模式，可用：

```powershell
.\docker\dev\compose.ps1 --local
```

默认本地端口：

- MySQL: `41001`
- system: `41002`
- ai: `41003`
- mvp: `41004`
- web: `41005`

更多说明见 [docs/Docker开发环境说明.md](docs/Docker开发环境说明.md)。
