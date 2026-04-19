# EasyMVP V3 本地目录与项目工作区规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3单机版存储架构设计](./EasyMVP-V3单机版存储架构设计.md)
> 关联文档：[EasyMVP-V3-SQLite初始化与Migration设计](./EasyMVP-V3-SQLite初始化与Migration设计.md)
> 关联文档：[EasyMVP-V3工作台视图模型与聚合接口设计](./EasyMVP-V3工作台视图模型与聚合接口设计.md)
> 关联文档：[EasyMVP-V3-回放与审计展示设计](./EasyMVP-V3-回放与审计展示设计.md)
> 关联文档：[EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
> 关联文档：[EasyMVP-V3-replay与log artifact存储规范](./EasyMVP-V3-replay与log artifact存储规范.md)
> 目标：定义 EasyMVP V3 单机版的数据根目录、项目工作区目录、Run 目录、Evidence 目录、文件命名规则与数据库索引边界。

## 1. 设计结论

V3 单机版必须有一套稳定、本地优先、可恢复、可索引的目录体系。

正式原则应定为：

1. 一个统一数据根目录
2. 一个项目一个稳定目录
3. 一个 `run_id` 一个稳定子目录
4. Evidence、Replay、Export、Diagnostics 明确分层
5. 页面不直接扫目录，数据库只存索引和引用

一句话：

> 目录结构是 V3 单机版的“文件侧状态机边界”，必须和数据库一样正式设计。

## 2. 为什么这份规范必须存在

如果没有目录规范，后面一定会出现这些问题：

1. logs、screenshots、replay 文件到处乱放
2. 同一个项目的数据难以打包、迁移、备份
3. 数据库里的路径缺少稳定语义
4. 页面和调试脚本开始直接扫磁盘拼状态
5. 清理与归档没有统一边界

所以目录结构不能只是“建议”，而应成为正式设计约束。

## 3. 顶层数据根目录

建议统一使用：

```text
~/.easymvp/
```

它是单机版默认数据根目录，承载：

1. 主数据库
2. 项目文件
3. 设置文件
4. 备份
5. 临时文件

## 4. 顶层目录结构

建议如下：

```text
~/.easymvp/
  ├─ data/
  │   ├─ easymvp.db
  │   ├─ easymvp.db-wal
  │   └─ easymvp.db-shm
  ├─ projects/
  │   └─ {project_id}/
  ├─ settings/
  │   └─ config.json
  ├─ backups/
  ├─ temp/
  └─ diagnostics/
```

### 4.1 `data/`

用于：

1. `SQLite` 主库
2. 相关 wal/shm 文件

### 4.2 `projects/`

用于：

1. 项目级工作区
2. run 文件
3. evidence 附件
4. replay 明细
5. 导出产物

### 4.3 `settings/`

用于：

1. 本地配置
2. 用户偏好
3. 默认路径、provider、runtime 选项

### 4.4 `backups/`

用于：

1. 升级前数据库备份
2. 项目级导出备份

### 4.5 `temp/`

用于：

1. 临时中间文件
2. 可随时清理缓存

### 4.6 `diagnostics/`

用于：

1. 启动失败诊断
2. 存储故障诊断
3. 导入/恢复错误报告

## 5. 项目级目录结构

每个项目建议固定使用：

```text
projects/{project_id}/
  ├─ meta/
  ├─ workspace/
  ├─ runs/
  ├─ evidence/
  ├─ replay/
  ├─ exports/
  ├─ cache/
  └─ diagnostics/
```

## 6. 各目录语义

### 6.1 `meta/`

用于：

1. 项目本地补充元信息
2. 只适合放少量辅助文件

不应替代数据库中的正式项目元数据。

### 6.2 `workspace/`

用于：

1. 项目实际工作目录
2. 代码、素材、配置等真实业务文件

说明：

1. 这是项目“被操作对象”的目录
2. 不等于 EasyMVP 自己的运行日志目录

### 6.3 `runs/`

用于：

1. 每次 `run_id` 的日志、checkpoint、artifact

### 6.4 `evidence/`

用于：

1. 验收截图
2. 录屏
3. 报告
4. 导出物引用副本

### 6.5 `replay/`

用于：

1. 回放明细
2. 可重放的结构化输出
3. 审计时引用的原始材料

### 6.6 `exports/`

用于：

1. 用户主动导出的包
2. 发布候选物
3. 可交付产物

### 6.7 `cache/`

用于：

1. 可重建中间产物
2. 缩略图
3. 临时视图缓存

缓存目录内容允许清理，不应作为正式证据来源。

### 6.8 `diagnostics/`

用于：

1. 项目级错误报告
2. 恢复失败诊断
3. 目录校验报告

## 7. Run 目录规范

建议每个 run 使用：

```text
projects/{project_id}/runs/{run_id}/
  ├─ logs/
  ├─ checkpoints/
  ├─ artifacts/
  ├─ replay/
  └─ meta.json
```

### 7.1 `logs/`

用于：

1. 原始日志分片
2. 标准输出快照
3. 调试辅助输出

### 7.2 `checkpoints/`

用于：

1. resume 所需状态快照文件
2. checkpoint payload

### 7.3 `artifacts/`

用于：

1. 本次 run 生成的直接产物
2. 中间结构化结果

### 7.4 `replay/`

用于：

1. 原始回放材料
2. 结构化回放切片

### 7.5 `meta.json`

仅作目录自描述用途，建议包含：

1. `run_id`
2. `project_id`
3. `brain_kind`
4. `created_at`
5. `attempt_no`

数据库仍然是正式事实来源，`meta.json` 只用于离线诊断和恢复辅助。

## 8. Evidence 目录规范

建议使用：

```text
projects/{project_id}/evidence/
  ├─ screenshots/
  ├─ videos/
  ├─ exports/
  ├─ reports/
  └─ attachments/
```

### 8.1 `screenshots/`

用于：

1. 页面截图
2. UI 采证图片
3. 引擎运行截图

### 8.2 `videos/`

用于：

1. 操作录屏
2. 运行态录屏
3. 导出结果视频

### 8.3 `exports/`

用于：

1. 验收涉及的导出文件副本
2. 发布候选物引用副本

### 8.4 `reports/`

用于：

1. 验收报告
2. 结构化摘要导出
3. 手工审阅报告

### 8.5 `attachments/`

用于：

1. 其他证据附件
2. 第三方导入材料

## 9. 文件命名规则

命名原则建议：

1. 稳定
2. 可读
3. 可追溯
4. 不依赖用户手工命名

### 9.1 推荐命名模板

```text
{timestamp}_{object_kind}_{object_id}_{suffix}.{ext}
```

例如：

1. `20260419T103201Z_evidence_ev_1024_homepage.png`
2. `20260419T103355Z_run_run_88_stdout.log`
3. `20260419T103912Z_replay_run_88_step-04.json`

### 9.2 命名要求

1. 时间使用 UTC 或统一时区格式
2. 文件名只用 ASCII
3. 避免空格
4. 不直接使用用户原始标题作为完整文件名

## 10. 路径与数据库的关系

V3 必须明确：

1. 文件系统负责保存附件本体
2. 数据库负责保存索引、引用和状态

### 10.1 数据库存的字段建议

至少应能索引：

1. `file_path`
2. `file_kind`
3. `file_ext`
4. `file_size`
5. `sha256`
6. `project_id`
7. `run_id`
8. `source_object_kind`
9. `source_object_id`
10. `created_at`

### 10.2 页面读取原则

页面不应：

1. 直接扫目录找最新截图
2. 直接靠目录名推断业务状态
3. 跳过数据库直接读取 evidence 列表

页面应：

1. 先读数据库索引
2. 再按引用读取实际文件

## 11. 创建项目时的目录初始化

项目创建成功后，系统应同步初始化：

1. `projects/{project_id}/`
2. `workspace/`
3. `runs/`
4. `evidence/`
5. `replay/`
6. `exports/`
7. `cache/`
8. `diagnostics/`

初始化原则：

1. 可以为空
2. 但目录结构应一次性创建到位
3. 避免运行过程中边写边补目录

## 12. 清理与归档规则

### 12.1 可清理

1. `cache/`
2. 非关键 run 日志明细
3. 临时 replay 切片
4. 非关键 diagnostics

### 12.2 不应直接删

1. 有数据库索引且仍被引用的 evidence
2. 当前活跃 run 目录
3. 当前项目 workspace

### 12.3 清理后的数据库处理

如果物理文件被清理：

1. 数据库索引不应静默删除
2. 应把对象标记为 `artifact_pruned`
3. 页面上应能看到“索引还在，文件已清理”

## 13. 备份与导出边界

### 13.1 最小项目备份单元

建议：

1. `projects/{project_id}/`
2. 数据库中该项目相关记录

### 13.2 导出原则

用户导出项目时，建议导出：

1. 项目元数据摘要
2. 关键 evidence
3. 发布产物
4. 关键 replay 索引

不必默认导出全部缓存和全部调试日志。

## 14. 恢复与校验

目录恢复后应至少执行：

1. 目录存在性检查
2. 数据库索引路径校验
3. 缺失文件计数
4. 孤儿文件检测
5. run 目录与绑定关系校验

### 14.1 孤儿文件

定义：

1. 文件存在
2. 但数据库没有对应索引

处理建议：

1. 不直接删除
2. 进入 `diagnostics`
3. 允许后续人工或工具修复

## 15. 不该怎么做

不建议：

1. 把所有运行文件都堆在一个目录
2. 用用户项目名直接当目录主键
3. 页面扫描目录推断业务状态
4. 目录结构随着不同分类随意变化
5. 把缓存目录内容当正式证据

## 16. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. Evidence 文档必须复用 `evidence/` 目录语义
2. replay 文档必须复用 `runs/{run_id}/replay/` 与项目级 `replay/` 语义
3. 页面设计只能读索引接口，不能直接依赖文件扫描
4. 项目创建流程必须包含目录初始化
5. 备份/恢复设计必须以本规范为目录基线

## 17. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-Evidence文件命名与引用规范](./EasyMVP-V3-Evidence文件命名与引用规范.md)
2. [EasyMVP-V3-replay与log artifact存储规范](./EasyMVP-V3-replay与log artifact存储规范.md)
3. 项目导出与打包结构设计
4. 目录校验与孤儿文件修复设计
