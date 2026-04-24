# EasyMVP

> 更新日期：2026-04-24

EasyMVP 当前只保留 V3 主线。V3 本轮列出的实现与文档收口项已完成；正式发布前仍需在 CI 或高配机执行全量验证，当前权威进度以 [docs/EasyMVP-V3-AgentTeam状态板.md](docs/EasyMVP-V3-AgentTeam状态板.md) 为准。

## 仓库结构

- `apps/core`：V3 本地核心服务
- `apps/desktop`：V3 桌面工作台
- `docs/`：V3 设计与实现文档
- `skills/`：EasyMVP 专项技能
- `scripts/local-verify-apps-core-desktop.bat`：Windows 本地校验入口
- `scripts/easymvp-backup-snapshot.sh`：SQLite / migration 快照、校验与恢复入口

## 当前口径

- 历史旧实现、旧前端、CI/CD、Docker 和专项验证资料已完成清理
- `admin-go/`、`vue-vben-admin/` 等非当前 V3 主线目录不再保留
- 当前开发与验证只围绕 `apps/core` 和 `apps/desktop`
- 本地联调入口以桌面端 + core 服务为准

## 当前状态

- 已完成：文档主线收口、`domain-brain` 主链路、GoFrame core 主 API、桌面端主页面接线、replay/log/evidence 聚合读写侧、诊断与恢复、runtime 幂等、packaged smoke 证明链、备份快照入口
- 资源口径：当前服务器不跑全量打包和重型端到端验证；发布前放到 CI 或高配机执行
- 状态判断：继续推进前先看状态板的“当前待做”，避免把发布验证误判成新的功能缺口

## 本地验证

Windows:

```powershell
scripts\local-verify-apps-core-desktop.bat
```

本地 Docker 启动 core 测试环境:

```powershell
dev_docker.bat
```

- 统一配置文件：`docker/dev/config.yaml`
- 依赖：Docker Desktop、Go 1.24+
- 启动脚本会先在宿主机交叉编译 Linux core，再用 `FROM scratch` 组装镜像，避免 Docker Hub 基础镜像拉取
- API 地址：`http://127.0.0.1:8000`
- 健康检查：`http://127.0.0.1:8000/api/v3/system/healthz`
- 查看日志：`dev_docker.bat logs`
- 停止环境：`dev_docker.bat down`
- 清理容器和数据卷：`dev_docker.bat clean`

手动执行：

```bash
cd apps/core && go test ./...
cd apps/desktop && pnpm run build
cd apps/desktop && pnpm run package:dir
cd apps/desktop && pnpm run verify:package
```

低资源本地快照：

```bash
scripts/easymvp-backup-snapshot.sh snapshot manual
scripts/easymvp-backup-snapshot.sh verify apps/core/var/backups/<snapshot-dir>
```

## CI 与打包

- 主验证：`.github/workflows/v3-guard.yml`
- Core 发布：`.github/workflows/core-release.yml`
- 桌面打包：`.github/workflows/desktop-package.yml`
- Core 本地健康校验：`./scripts/verify-core-health.sh`
- 桌面本地打包：`cd apps/desktop && pnpm run package`
- 桌面打包产物 smoke：`cd apps/desktop && pnpm run verify:package`
- 打包前会自动编译 `apps/core` 并把二进制带入安装包资源目录

## 文档入口

- [docs/README.md](docs/README.md)
- [docs/EasyMVP-V3文档总纲.md](docs/EasyMVP-V3文档总纲.md)
- [docs/EasyMVP-V3总体架构设计.md](docs/EasyMVP-V3总体架构设计.md)
- [docs/EasyMVP-V3-AgentTeam状态板.md](docs/EasyMVP-V3-AgentTeam状态板.md)
