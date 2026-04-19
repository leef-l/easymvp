# EasyMVP

> 更新日期：2026-04-20

EasyMVP 当前只保留 V3 主线。

## 仓库结构

- `apps/core`：V3 本地核心服务
- `apps/desktop`：V3 桌面工作台
- `docs/`：V3 设计与实现文档
- `skills/`：EasyMVP 专项技能
- `scripts/local-verify-apps-core-desktop.bat`：Windows 本地校验入口

## 当前口径

- 已删除旧的 V2 代码、前端、CI/CD、Docker 和专项验证资料
- 当前开发与验证只围绕 `apps/core` 和 `apps/desktop`
- 本地联调入口以桌面端 + core 服务为准

## 本地验证

Windows:

```powershell
scripts\local-verify-apps-core-desktop.bat
```

手动执行：

```bash
cd apps/core && go test ./...
cd apps/desktop && pnpm run build
```

## 文档入口

- [docs/README.md](docs/README.md)
- [docs/EasyMVP-V3文档总纲.md](docs/EasyMVP-V3文档总纲.md)
- [docs/EasyMVP-V3总体架构设计.md](docs/EasyMVP-V3总体架构设计.md)
- [docs/EasyMVP-V3-AgentTeam状态板.md](docs/EasyMVP-V3-AgentTeam状态板.md)
