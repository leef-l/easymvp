# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- V3 桌面工作台 (`apps/desktop`)：Workspace / Plan / Execution / Replay / Acceptance / Audit / Diagnostics / Recovery / Repair / Settings 主链路接线
- V3 本地核心服务 (`apps/core`)：GoFrame v2 骨架，system/healthz 最小入口，真实 migration 文件
- 领域脑 (`easymvp-brain`) 接入：plan_review、plan_compile、acceptance_mapping、completion_adjudication、repair_design、workspace_explanation
- 中文国际化（默认 `zh-CN`，`react-i18next`）
- 前端单元测试框架（Vitest + React Testing Library）
- SQLite WAL 模式
- OpenAPI / Swagger 文档 (`/api.json`, `/swagger`)
- lefthook pre-commit（go vet、go build、frontend typecheck、禁止 main 直推）
- 备份快照脚本 (`scripts/easymvp-backup-snapshot.sh`)
- Docker dev 环境启动脚本 (`dev_docker.bat`)
- CI workflows：v3-guard、core-release、desktop-package

### Changed

- 统一文档口径到 V3 主线，清理旧 V3 对齐进度、残留清单等过渡文档
- `admin-go/`、`vue-vben-admin/` 等非 V3 主线目录不再保留

### Fixed

- CORS 跨域：dev 模式请求改走 Vite proxy
- Docker 编译错误：`easyMVPBrainErrorCodeContractInvalid` 常量缺失
- SettingsPage 语法错误（Babel `Unexpected token`）
- runtime 幂等复用、事件去重与 checkpoint 写入
- 自动 adjudication 按 `acceptance_run_id` 精准裁决，避免跨任务串单
- `unsupported / denied` 显式保留并投影到视图层

## [0.1.0] - 2026-04-25

- V3 初始版本，包含核心服务与桌面工作台主链路。
