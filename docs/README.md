# EasyMVP V3 文档目录

> **版本**: v3.1.0  
> **更新日期**: 2026-04-29  
> **状态**: 已清理过期文档，当前索引反映最新状态

---

## 新架构文档（优先阅读）

| 文档 | 说明 | 状态 |
|------|------|------|
| `MACCS-EasyMVP-闭环工作流规范.md` | 七阶段闭环：需求→方案→审核→执行→验收→交付→复盘 | ✅ 最新 |
| `钱学森总纲设计/README.md` | 当前权威总纲入口（工程控制论框架） | ✅ 最新 |

---

## 核心设计文档

### 总纲与架构
- [钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md](钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md) — 顶层架构方案
- [钱学森总纲设计/EasyMVP工程铁律.md](钱学森总纲设计/EasyMVP工程铁律.md) — 工程约束
- [钱学森总纲设计/EasyMVP-三层验证架构说明.md](钱学森总纲设计/EasyMVP-三层验证架构说明.md) — 验证架构
- [EasyMVP-V3-实现架构与模块拆分设计.md](EasyMVP-V3-实现架构与模块拆分设计.md) — 模块拆分

### 后端与运行时
- [EasyMVP-V3-Go本地核心服务架构设计.md](EasyMVP-V3-Go本地核心服务架构设计.md) — Go 核心服务
- [EasyMVP-V3-代码目录结构与模块归属建议.md](EasyMVP-V3-代码目录结构与模块归属建议.md) — 目录结构
- [EasyMVP-V3-本地配置与启动参数设计.md](EasyMVP-V3-本地配置与启动参数设计.md) — 配置设计
- [EasyMVP-V3-SQLite初始化与Migration设计.md](EasyMVP-V3-SQLite初始化与Migration设计.md) — 数据库迁移
- [EasyMVP-V3-独立Migration文件正文终稿.md](EasyMVP-V3-独立Migration文件正文终稿.md) — Migration 正文

### 桌面端与交互
- [EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计.md](EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计.md) — 脚手架
- [EasyMVP-V3-Electron进程模型与IPC边界设计.md](EasyMVP-V3-Electron进程模型与IPC边界设计.md) — 进程模型
- [EasyMVP-V3-工作台全页面设计规范.md](EasyMVP-V3-工作台全页面设计规范.md) — 页面规范
- [EasyMVP-V3-恢复模式与诊断模式页面设计.md](EasyMVP-V3-恢复模式与诊断模式页面设计.md) — 恢复/诊断

### 合同与专精大脑
- [EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md](EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md) — 合同 Schema
- [钱学森总纲设计/easymvp-brain-职责与边界定义.md](钱学森总纲设计/easymvp-brain-职责与边界定义.md) — 职责边界
- [钱学森总纲设计/easymvp-brain-输入输出契约.md](钱学森总纲设计/easymvp-brain-输入输出契约.md) — IO 契约

---

## 已删除文档（内容已归档或合并）

| 原文档 | 处理方式 | 替代/去向 |
|--------|---------|----------|
| `EasyMVP-V3-AgentTeam总章程.md` | ❌ 删除 | AI 自管理文档，非产品文档 |
| `EasyMVP-V3-AgentTeam-Backend计划.md` | ❌ 删除 | 同上 |
| `EasyMVP-V3-AgentTeam-Frontend计划.md` | ❌ 删除 | 同上 |
| `EasyMVP-V3-AgentTeam-Runtime计划.md` | ❌ 删除 | 同上 |
| `EasyMVP-V3-AgentTeam开机计划总表.md` | ❌ 删除 | 同上 |
| `EasyMVP-V3-AgentTeam状态板.md` | ❌ 删除 | 同上 |
| `EasyMVP-V3文档总纲.md` | ❌ 删除 | 被 `钱学森总纲设计/` 取代 |
| `EasyMVP-V3总体架构设计.md` | ❌ 删除 | 被 `钱学森总纲设计/` 取代 |
| `EasyMVP-实施缺口追踪与完成清单.md` | ❌ 删除 | 13 项缺口全部完成，归档至 `钱学森总纲设计/EasyMVP-历史缺口归档.md` |
| `AI应用使用指南.md` | ⚠️ 待废弃 | 旧 AI 管理模块，将被 MACCS 闭环取代 |

---

## 其他入口

- **冒烟测试**: [brain-v3-smoke-manual.md](brain-v3-smoke-manual.md)
- **本地打包**: [本地打包验证手册.md](本地打包验证手册.md)
- **重启接续**: [EasyMVP-V3-当前上下文与重启接续说明.md](EasyMVP-V3-当前上下文与重启接续说明.md)
- **快照脚本**: `../scripts/easymvp-backup-snapshot.sh`
