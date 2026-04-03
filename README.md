# EasyMVP

项目说明、架构设计、AI 引擎接入、使用文档和测试报告统一放在 `docs/` 目录。

本地 Docker 开发入口放在 `docker/dev/`，中国网络环境建议直接运行：

```powershell
.\docker\dev\compose.ps1
```

如需先预打包开发镜像，执行：

```powershell
.\docker\dev\build.ps1
```

`compose.ps1` 会先用 `docker/dev/.env` 覆盖 `admin-go/.env`，默认执行：

```powershell
docker build -f docker/build/Dockerfile.admin-go.dev admin-go -t easymvp-admin-go-dev:latest
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.cn.yml up -d
```

默认本地端口：

- MySQL: `41001`
- system: `41002`
- ai: `41003`
- mvp: `41004`
- web: `41005`

更多说明见 `docs/README.md`。
