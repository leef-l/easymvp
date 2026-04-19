# EasyMVP V3 SQLite 初始化与 Migration 设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)
> 关联文档：[EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
> 关联文档：[EasyMVP-V3-brain-serve接口接入与Run生命周期映射](./EasyMVP-V3-brain-serve接口接入与Run生命周期映射.md)
> 目标：定义 EasyMVP V3 单机版在 `SQLite` 上的数据库初始化、schema migration、升级恢复与版本兼容策略。

## 1. 设计结论

V3 单机版的数据库生命周期应统一成一条稳定链路：

1. 初始化数据根目录
2. 打开 `SQLite` 数据库
3. 设置连接级 pragma
4. 读取 `schema_migrations`
5. 顺序执行未应用 migration
6. 执行启动后自检
7. 对外提供 Orchestrator 与聚合层访问

一句话：

> V3 不应靠“启动时偷偷建表”维持数据库，而应采用正式的 migration 机制管理 schema 演进。

## 2. 为什么必须正式做 Migration

如果不做正式 migration，单机版后续会很快失控：

1. 不同版本的数据目录无法兼容
2. 用户升级后数据库结构不可预测
3. 回滚和诊断没有统一依据
4. 表结构演进会和业务逻辑缠死
5. 文档里的表设计无法真正落地

所以 V3 必须把 migration 当成基础设施，而不是补充脚本。

## 3. 适用边界

本专题只覆盖：

1. `SQLite` schema 初始化
2. 结构化表的 migration
3. 启动时版本检测
4. 失败恢复与备份策略

本专题不覆盖：

1. 项目工作区目录初始化细节
2. Evidence 文件命名规则
3. replay 附件目录清理规则
4. 远程数据库迁移

## 4. 数据库初始化流程

推荐首次启动或打开数据目录时，执行如下流程：

```text
resolve data root
  → ensure directories
  → open SQLite connection
  → apply PRAGMA
  → create schema_migrations if absent
  → run pending migrations
  → run bootstrap seed
  → run health check
  → mark db ready
```

### 4.1 目录检查

至少确认：

1. `~/.easymvp/data/`
2. 数据库文件父目录可写
3. `backups/` 目录存在
4. `temp/` 目录存在

### 4.2 数据库连接

打开数据库后建议立即设置：

1. `PRAGMA journal_mode = WAL`
2. `PRAGMA foreign_keys = ON`
3. `PRAGMA busy_timeout = 5000`
4. `PRAGMA synchronous = NORMAL`

说明：

1. `WAL` 支撑单机版读写并存
2. `foreign_keys` 保证引用关系不漂
3. `busy_timeout` 避免短时锁冲突直接失败
4. `synchronous = NORMAL` 在单机版性能与可靠性间较平衡

## 5. Migration 机制

### 5.1 正式原则

V3 的 migration 机制建议遵守：

1. 只增量演进，不依赖运行时隐式补表
2. 每次 schema 变更都必须有独立 migration
3. migration 必须可记录、可诊断、可重试
4. migration 编号必须全局单调递增
5. 启动时只执行“未应用 migration”

### 5.2 推荐命名

建议每个 migration 使用：

```text
{version}_{slug}.sql
```

例如：

1. `0001_init_core_tables.sql`
2. `0002_add_plan_review_indexes.sql`
3. `0003_add_brain_run_checkpoints.sql`

### 5.3 不推荐命名

不建议：

1. 用时间戳但无排序规则
2. 文件名与版本号无关
3. 把多个大改动混进单个超长 migration

## 6. `schema_migrations` 表设计

建议单独维护：

### `schema_migrations`

核心列建议：

1. `version`
2. `name`
3. `checksum`
4. `applied_at`
5. `duration_ms`
6. `status`
7. `error_message`

建议约束：

1. `version` 唯一
2. `status` 只允许 `applied / failed`

作用：

1. 判断数据库当前 schema 版本
2. 定位某次升级是否失败
3. 检查文件和数据库记录是否一致

## 7. 首次初始化策略

### 7.1 首次安装

如果数据库文件不存在：

1. 创建空库
2. 创建 `schema_migrations`
3. 从 `0001` 顺序执行到最新版本
4. 写入系统基础配置
5. 写入默认 schema version

### 7.2 不要直接发“最终完整 SQL 快照”

第一版仍建议保留 migration 序列，而不是只靠一个大 `init.sql`。

原因：

1. 初始化和升级使用同一机制
2. 后续容易诊断版本差异
3. 文档和实现更一致

## 8. 升级策略

### 8.1 启动时升级

应用启动时应读取：

1. 当前代码内置最新 migration version
2. 当前数据库已应用 version

如果数据库版本落后：

1. 自动执行缺失 migration
2. 记录每次 migration 的结果
3. 任一失败即停止升级并进入恢复模式

### 8.2 升级前备份

如果待执行 migration 不为空，建议先备份：

1. `easymvp.db`
2. 必要时 `-wal` 与 `-shm`

备份文件建议落到：

```text
~/.easymvp/backups/db-pre-migrate-{timestamp}.sqlite3
```

### 8.3 失败处理

如果某个 migration 失败：

1. 立即停止后续 migration
2. 将当前 migration 记录为 `failed`
3. 输出清晰错误信息
4. 提示恢复备份或重试
5. 不允许应用继续以“半升级状态”进入正常运行

## 9. Migration 编写约束

### 9.1 必须遵守

1. 单个 migration 目标单一
2. 尽量事务化执行
3. 新增列要考虑旧数据默认值
4. 索引创建要有明确目的
5. 不在 migration 中塞业务逻辑判断

### 9.2 SQLite 特别注意

在 `SQLite` 中修改表结构时要特别克制：

1. 避免频繁依赖复杂 `ALTER TABLE`
2. 需要重建表时，应显式走“新表 -> 数据搬迁 -> 校验 -> 替换”
3. 不要在一次 migration 中做过多 destructive 变更

### 9.3 数据回填

如果 migration 需要回填旧数据：

1. 先加新列或新表
2. 再做结构化回填
3. 回填后做最小一致性校验
4. 最后再切换读路径

## 10. 启动自检

migration 完成后，启动流程还应做最小自检。

至少检查：

1. 核心表是否存在
2. `schema_migrations` 是否达到代码期望版本
3. 外键约束是否启用
4. 关键索引是否存在
5. 数据目录是否可写

如果任一关键项失败，应用应进入：

1. `storage_not_ready`
2. 禁止进入主工作台
3. 只允许展示恢复/诊断信息

## 11. 与业务版本的关系

要区分两类版本：

1. 应用版本
2. schema 版本

### 11.1 应用版本

表示当前 EasyMVP 程序版本。

### 11.2 schema 版本

表示数据库结构版本。

约束：

1. 应用版本升级可能不带 schema 升级
2. schema 升级必须有明确 migration 版本号
3. 页面和聚合层不得直接依赖“应用版本猜测 schema 结构”

## 12. 与核心对象表的关系

第一批 migration 至少应覆盖：

1. `workflow_projects`
2. `workflow_plan_drafts`
3. `workflow_plan_review_results`
4. `workflow_compiled_plans`
5. `workflow_compiled_tasks`
6. `workflow_domain_tasks`
7. `workflow_brain_run_bindings`
8. `workflow_brain_run_events`
9. `workflow_brain_run_checkpoints`
10. `workflow_acceptance_runs`
11. `workflow_action_inbox_items`
12. `workflow_evidence_index`
13. `app_settings`

## 13. 推荐 migration 分批策略

建议不要把所有表放进一个巨型 migration。

推荐首批拆法：

### P0

1. 项目与设置核心表
2. 计划链路核心表
3. 任务链路核心表

### P1

1. Run 绑定与事件表
2. 验收运行表
3. 待处理问题表

### P2

1. 证据索引表
2. 回放与审计辅助表
3. 缓存或诊断辅助表

## 14. 兼容与回滚策略

### 14.1 前向兼容

单机版当前优先保证：

1. 老数据库能升级到新版本
2. 升级路径可重复执行

### 14.2 回滚

不建议对每个 migration 都设计自动“反向 down migration”。

更现实的策略是：

1. 升级前自动备份
2. 升级失败时恢复备份
3. 通过诊断信息引导用户恢复

因为单机版更适合“备份恢复式回滚”，不适合维护复杂双向 migration。

## 15. 不该怎么做

不建议：

1. 启动时发现缺表就临时 `CREATE TABLE`
2. 不记录 migration 版本直接偷偷改 schema
3. 把应用业务初始化和 schema migration 混成一层
4. 在 migration 中扫描全磁盘目录做业务修复
5. 让前端或页面逻辑感知 migration 细节

## 16. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. 新增结构化表必须对应新 migration
2. 存储文档中的表不能只停留在自然语言，必须能落到 migration 序列
3. 计划数据、Run 绑定、验收索引的表演进都应走统一升级机制
4. 页面设计不得假设数据库 schema 会被运行时自动修补

## 17. 后续细分专题

本专题后续继续拆：

1. 首批 migration 清单设计
2. `schema_migrations` 校验与 checksum 规则
3. 启动恢复页与存储诊断设计
4. 数据目录初始化与自检实现规范
