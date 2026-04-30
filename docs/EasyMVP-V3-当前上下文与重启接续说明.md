# EasyMVP V3 当前上下文与重启接续说明

> 更新时间：2026-04-29
> 目标：在需要重启 Codex 或切换会话时，保留当前 V3 设计上下文、关键决策和下一步执行入口，避免重新对齐成本。

## 1. 当前项目定位

EasyMVP V3 当前正式定位为：

> 单用户、单机版、本地优先、实时可视化的项目工作台。

不是：

1. SaaS 后台
2. 多租户系统
3. 管理后台

当前完成度：

1. 文档主线和 `domain-brain` 主链路已经完成
2. `apps/core` 与 `apps/desktop` 已完成本轮列出的主链路、恢复诊断、replay/evidence 聚合、runtime 幂等与 desktop packaged smoke 证明链收口
3. **MACCS 七阶段闭环已全部落地**：需求分析 -> 方案设计 -> 设计审核 -> 任务执行 -> 多层验收 -> 交付管理 -> 项目复盘，后端服务层与前端 API 层均已实现
4. MACCS 闭环端到端集成测试已编写（`maccs_integration_test.go`），覆盖完整生命周期数据层验证
5. 拓扑排序调度器和多层验收标准的纯逻辑单元测试已补充
6. V3 当前剩余不再是已知代码功能待办，而是发布前的端到端冒烟测试和真实 `brain-v3` / `easymvp-brain` 集成验证
7. 最新状态以代码实现和 GitHub Actions CI 状态为准

## 2. 当前已确定的核心技术路线

当前正式技术栈已经定为：

1. 桌面壳：`Electron`
2. 前端：`React + TypeScript + Vite`
3. 本地核心服务：`Go + GoFrame v2`
4. 本地数据库：`SQLite`
5. 文件存储：`Local Files`
6. 运行时底座：`brain-v3`
7. 运行时接入方式：`brain serve HTTP API`

## 3. 当前已确定的关键架构结论

### 3.1 进程与职责

1. Electron Main 只负责桌面壳和原生能力
2. React Renderer 只负责页面与可视化
3. GoFrame v2 本地核心服务负责业务内核
4. `brain-v3` 继续作为外部运行时底座
5. `easymvp-brain` 只负责领域裁决与结构化领域结果

### 3.2 通讯边界

1. 业务 API 走本地 GoFrame HTTP 服务
2. Electron IPC 只负责桌面原生能力桥接
3. Renderer 不直接访问 SQLite
4. Renderer 不直接调用 `brain serve`
5. Renderer 和领域层都不直接消费 `brain-v3` 原始工具结果；必须先经过 EasyMVP 归一化

### 3.3 存储边界

1. `SQLite` 存结构化事实数据
2. 本地文件系统存 evidence / replay / logs / exports
3. 视图快照允许缓存，但不能替代事实表

## 4. 当前文档主线已完成范围

当前已完成首版闭环的文档层包括：

1. 总纲与总体架构
2. 方案编译
3. 分类与角色
4. 专精大脑接入
5. 生产级分类验收
6. 实时工作台与页面体系
7. 单机版存储与 SQLite
8. GoFrame v2 本地核心服务实现架构
9. API、DTO、事务、一致性、错误码、配置、恢复模式
10. `easymvp-brain` 边界、合同 Schema 与 `brain-v3` 协议适配语义
11. 旧 V3 对齐进度、旧 V3 残留清单等过渡文档已删除，不再作为入口

## 5. 当前新增且关键的实现级文档

建议重启后优先继续阅读这些：

1. [EasyMVP-V3-技术栈与选型基线](./EasyMVP-V3-技术栈与选型基线.md)
2. [EasyMVP-V3-Go本地核心服务架构设计](./EasyMVP-V3-Go本地核心服务架构设计.md)
3. [EasyMVP-V3-代码目录结构与模块归属建议](./EasyMVP-V3-代码目录结构与模块归属建议.md)
4. [EasyMVP-V3-数据库Schema总设计](./EasyMVP-V3-数据库Schema总设计.md)
5. [EasyMVP-V3-首批Migration清单与建表SQL设计](./EasyMVP-V3-首批Migration清单与建表SQL设计.md)
6. [EasyMVP-V3-本地API-DTO与错误返回设计](./EasyMVP-V3-本地API-DTO与错误返回设计.md)
7. [EasyMVP-V3-Go包级接口与依赖关系设计](./EasyMVP-V3-Go包级接口与依赖关系设计.md)
8. [EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计](./EasyMVP-V3-Electron-Go单仓脚手架与开发命令设计.md)
9. [EasyMVP-V3-恢复模式与诊断模式页面设计](./EasyMVP-V3-恢复模式与诊断模式页面设计.md)
10. [EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿](./EasyMVP-V3-easymvp-brain-合同JSON-Schema与DTO映射终稿.md)
11. [钱学森总纲设计/easymvp-brain-职责与边界定义](./钱学森总纲设计/easymvp-brain-职责与边界定义.md)
12. [钱学森总纲设计/easymvp-brain-输入输出契约](./钱学森总纲设计/easymvp-brain-输入输出契约.md)
13. [EasyMVP-V3-AgentTeam状态板](./EasyMVP-V3-AgentTeam状态板.md)

## 6. 当前工具状态

已完成：

1. 安装 `goframe-v2` skill

安装位置：

1. `/root/.codex/skills/goframe-v2`

注意：

1. 当前会话已经可以识别 `goframe-v2` skill；修改 GoFrame 相关代码时应继续使用它

## 7. 重启后的直接下一步

重启后，不需要重新讨论架构方向。MACCS 七阶段闭环已全部实现，当前进入端到端冒烟测试阶段。

建议直接进入：

1. 在 CI 或高配机运行 `cd apps/core && go test ./internal/service/...`，确认所有单元测试和集成测试通过
2. 在 CI 或高配机运行全量 `go build ./...` 和 `go vet ./...`
3. 用真实 `brain-v3` / `easymvp-brain` 服务做一次端到端人工冒烟，覆盖完整 MACCS 七阶段流程
4. 在 CI 或高配机运行 `cd apps/desktop && pnpm run package:dir && pnpm run verify:package`
5. 若升级或迁移数据，先运行 `scripts/easymvp-backup-snapshot.sh snapshot <label>` 生成升级前快照

### 7.1 MACCS 七阶段 API 端点概要

| 阶段 | 端点路径 | 方法 | 说明 |
|------|---------|------|------|
| 需求分析 | `/api/v3/requirements/analyze` | POST | 提交原始需求，Brain 生成结构化需求文档 |
| 需求确认 | `/api/v3/requirements/confirm` | POST | 用户确认需求文档 |
| 需求查询 | `/api/v3/requirements/get` | GET | 查询需求详情 |
| 方案设计 | `/api/v3/designs/generate` | POST | 基于需求生成技术方案 |
| 方案确认 | `/api/v3/designs/confirm` | POST | 用户确认方案 |
| 方案拒绝 | `/api/v3/designs/reject` | POST | 用户拒绝方案（可重新生成） |
| 设计审核 | `/api/v3/reviews/start` | POST | 启动单轮设计审核 |
| 审核循环 | `/api/v3/reviews/run-loop` | POST | 自动审核-修复循环 |
| 审核列表 | `/api/v3/reviews/list` | GET | 查询审核记录 |
| 审核介入 | `/api/v3/reviews/intervene` | POST | 人工覆盖/中止/重启审核 |
| 交付准备 | `/api/v3/deliveries/prepare` | POST | 准备交付物 |
| 交付验收 | `/api/v3/deliveries/accept` | POST | 验收通过 |
| 交付拒绝 | `/api/v3/deliveries/reject` | POST | 验收拒绝 |
| 交付查询 | `/api/v3/deliveries/get` | GET | 查询交付详情 |
| 项目复盘 | `/api/v3/retrospectives/generate` | POST | 生成项目复盘报告 |
| 复盘查询 | `/api/v3/retrospectives/get` | GET | 查询复盘报告 |
| 验收启动 | `/api/v3/acceptance/start` | POST | 启动多层验收 |
| 验收视图 | `/api/v3/acceptance/view` | GET | 查看验收状态 |
| 手动放行 | `/api/v3/acceptance/manual-release` | POST | 人工手动放行 |

## 8. 重启后的建议指令

重启后可以直接说：

```text
继续 EasyMVP V3，按 docs/EasyMVP-V3-当前上下文与重启接续说明.md 执行，
MACCS 七阶段闭环已全部实现，
先在 CI 或高配机做全量 go test/go build 验证，
再接真实 brain-v3/easymvp-brain 做七阶段端到端冒烟。
```

## 9. 一句话总结

即使重启会话，仓库里的 V3 文档和这份接续说明仍然保留，所以不会丢设计结论；MACCS 七阶段闭环（需求 -> 方案 -> 审核 -> 执行 -> 验收 -> 交付 -> 复盘）已全部实现并有集成测试覆盖，恢复后优先做端到端冒烟测试。
