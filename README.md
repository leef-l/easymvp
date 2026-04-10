# EasyMVP

> 更新日期：2026-04-09

EasyMVP 是一个 AI 协作开发平台仓库。当前主线代码由 `system`、`ai`、`mvp` 三个 Go 服务和 `vue-vben-admin` 前端组成，核心能力包括 Workflow V2、统一执行器抽象、任务级 `git worktree` 隔离、飞书/Telegram 协作，以及 AI 模型与执行引擎配置管理。

## 仓库结构

- `admin-go/app/system`：系统管理、用户、角色、菜单、RBAC
- `admin-go/app/ai`：供应商、套餐、模型、执行引擎、手工 AI 任务
- `admin-go/app/mvp`：项目、对话、Workflow V2、执行器、协作、自治
- `vue-vben-admin`：管理端前端
- `docker/`：开发和生产环境脚本
- `docs/`：当前仍与实现保持一致的文档

## 文档入口

- [docs/README.md](docs/README.md)：文档索引
- [docs/EasyMVP使用文档.md](docs/EasyMVP使用文档.md)：项目创建、审核、执行、验收、协作使用流程
- [docs/AI应用使用指南.md](docs/AI应用使用指南.md)：AI 供应商、模型、引擎和任务配置
- [docs/EasyMVP架构设计文档.md](docs/EasyMVP架构设计文档.md)：当前仓库结构和模块边界
- [docs/EasyMVP研发执行版.md](docs/EasyMVP研发执行版.md)：当前研发拆解、阶段任务与依赖关系
- [docs/EasyMVP项目收尾计划与进度.md](docs/EasyMVP项目收尾计划与进度.md)：最新收尾进度、验证结果与剩余风险
- [docs/Docker开发环境说明.md](docs/Docker开发环境说明.md)：开发环境启动方式和 compose 差异

2026-04-09 已继续清理仓库中明显过期的迁移方案、阶段性设计稿和一次性联调记录；如需追溯旧方案，请直接查看 `git` 历史。

## 本地开发

Windows 推荐入口：

```powershell
.\docker\dev\compose.ps1
```

这条命令默认会：

- 同步 `docker/dev/.env` 到 `admin-go/.env`
- 构建 `easymvp-admin-go-dev:latest` 和 `easymvp-web-dev:latest`
- 使用 `docker/dev/docker-compose.cn.yml` 启动 `mysql`、`system`、`ai`、`mvp`、`web`

如果还需要 `OpenHands` runtime：

```powershell
.\docker\dev\compose.ps1 --profile ai-runtime up -d
```

如需“Docker 基础设施 + 本地热重启”模式：

```powershell
.\docker\dev\compose.ps1 --local
```

如需手动启用带 `redis` 的通用开发 compose：

```bash
docker compose --project-name easymvp --env-file docker/dev/.env -f docker/dev/docker-compose.yml up -d
```

默认本地端口：

- MySQL：`41001`
- system：`41002`
- ai：`41003`
- mvp：`41004`
- web：`41005`

更多说明见 [docs/Docker开发环境说明.md](docs/Docker开发环境说明.md)。

## 受限资源验证

当前服务器可以执行 `web-antd` 的全量类型检查，但需要受控资源窗口，不能继续沿用 `512MB / 768MB / 1024MB / 1280MB` 这类过低堆上限。

推荐入口：

```bash
bash scripts/web-antd-typecheck-safe.sh
```

这条脚本会：

- 检查 `MemAvailable`
- 使用文件锁避免并发重复执行
- 以低优先级运行
- 默认使用 `1536MB` Node 堆执行 `vue-tsc --noEmit`

当前 EasyMVP 主链还新增了一条硬约束：`category resolver / verification / acceptance / accept stage` 这些编排链路不得直接访问 DB，新增查询必须先下沉到 repo。

可选环境变量：

- `EASYMVP_TYPECHECK_HEAP_MB`
- `EASYMVP_TYPECHECK_MIN_AVAILABLE_MB`
