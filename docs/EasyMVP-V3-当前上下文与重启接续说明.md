# EasyMVP V3 当前上下文与重启接续说明

> 更新时间：2026-04-24
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
3. V3 当前剩余不再是已知代码功能待办，而是发布前的 CI / 高配机全量验证和真实 `brain-v3` / `easymvp-brain` 端到端冒烟
4. 最新状态以代码实现和 GitHub Actions CI 状态为准

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

重启后，不需要重新讨论架构方向。

建议直接进入：

1. 使用 `goframe-v2` skill
2. 先查看 [EasyMVP-V3-AgentTeam状态板](./EasyMVP-V3-AgentTeam状态板.md) 的验证口径
3. 在 CI 或高配机运行全量 `go test ./...`
4. 在 CI 或高配机运行 `cd apps/desktop && pnpm run package:dir && pnpm run verify:package`
5. 用真实 `brain-v3` / `easymvp-brain` 服务做一次端到端人工冒烟
6. 若升级或迁移数据，先运行 `scripts/easymvp-backup-snapshot.sh snapshot <label>` 生成升级前快照

## 8. 重启后的建议指令

重启后可以直接说：

```text
继续 EasyMVP V3，按 docs/EasyMVP-V3-当前上下文与重启接续说明.md 执行，
使用 goframe-v2 skill，
不要在低配服务器跑重型构建，
先在 CI 或高配机做全量 Go/desktop package/verify-package 验证，
再接真实 brain-v3/easymvp-brain 做端到端冒烟。
```

## 9. 一句话总结

即使重启会话，仓库里的 V3 文档和这份接续说明仍然保留，所以不会丢设计结论；当前已知代码实现项已收口，恢复后优先做发布级验证而不是继续找功能缺口。
