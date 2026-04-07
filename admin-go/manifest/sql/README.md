# 数据库迁移管理（golang-migrate）

## 一、概述

本项目使用 [golang-migrate](https://github.com/golang-migrate/migrate) 管理数据库 Schema 变更。

**核心原则：所有数据库结构变更必须通过迁移文件，禁止手动改库不写迁移文件。**

### 目录结构

```
admin-go/
├── manifest/sql/
│   ├── mysql/                    # 迁移文件目录（golang-migrate 管理）
│   │   ├── 000001_baseline_schema.up.sql     # 基线：全量建表
│   │   ├── 000001_baseline_schema.down.sql   # 基线回滚：全量删表
│   │   ├── 000002_add_xxx.up.sql             # 增量迁移
│   │   └── 000002_add_xxx.down.sql           # 增量回滚
│   ├── seed/
│   │   └── mysql_seed.sql        # 种子数据（系统用户/角色/菜单/预设等）
│   └── README.md                 # 本文件
├── docker/mysql/
│   ├── init.sql                  # 完整快照（备份用，非主要升级路径）
│   └── schema.sql                # 纯结构快照
└── hack/
    ├── hack-db.mk                # Makefile targets
    └── db-migrate.sh             # 迁移脚本
```

## 二、常用命令

在 `admin-go/` 目录下执行：

```bash
# 查看当前迁移版本
make db-version

# 执行所有待迁移（升级到最新）
make db-up

# 回滚 1 步
make db-down STEPS=1

# 回滚 N 步
make db-down STEPS=3

# 创建新迁移文件（自动编号）
make db-create NAME=add_user_avatar_field

# 跳转到指定版本
make db-goto VERSION=2

# 强制设置版本（修复 dirty 状态，不执行 SQL）
make db-force VERSION=2

# 首次初始化（迁移 + 种子数据）
make db-bootstrap
```

## 三、日常开发流程

### 新增数据库变更

```bash
# 1. 创建迁移文件
make db-create NAME=add_user_avatar_field

# 2. 编写 SQL
#    生成的文件：
#    manifest/sql/mysql/000003_add_user_avatar_field.up.sql   ← 升级
#    manifest/sql/mysql/000003_add_user_avatar_field.down.sql ← 回滚

# 3. 编写 up.sql（升级 SQL）
# 示例：
#   ALTER TABLE system_users ADD COLUMN avatar VARCHAR(500) DEFAULT '' COMMENT '头像URL';

# 4. 编写 down.sql（回滚 SQL，必须能完整撤销 up.sql 的变更）
# 示例：
#   ALTER TABLE system_users DROP COLUMN avatar;

# 5. 本地执行验证
make db-up

# 6. 验证回滚
make db-down STEPS=1
make db-up

# 7. 同步快照（可选，用于备份和审计）
mysqldump --no-tablespaces -u easymvp -pJKcHFJYXnkrB6BXE -h 127.0.0.1 easymvp \
  --result-file=docker/mysql/init.sql

# 8. 提交代码
git add manifest/sql/mysql/000003_*.sql
git commit -m "migrate: add user avatar field"
git push
```

### Docker 环境自动迁移

Docker 容器启动时，**mvp 服务会自动执行 `migrate up`**：

```
docker compose up
  → MySQL 创建空数据库
  → mvp 容器启动前自动执行：
    1. migrate up（建表 + 加索引 + 所有增量迁移）
    2. dbctl seed（首次建表后导入种子数据）
  → 服务正常启动
```

其他环境只需 `git pull && docker compose up --build`，数据库自动升级。

## 四、编写迁移文件的规范

### 必须遵守

1. **每个变更一对文件**：`.up.sql` 和 `.down.sql` 必须成对出现
2. **down.sql 必须能完整回滚 up.sql**：`make db-up && make db-down STEPS=1` 必须无损
3. **禁止修改已提交的迁移文件**：如果发现错误，创建新的迁移文件来修复
4. **幂等性**：推荐使用 `IF NOT EXISTS`（建表）、`IF EXISTS`（删表/删列）
5. **字符集**：所有表/列必须使用 `utf8mb4`，禁止其他编码

### SQL 编写建议

```sql
-- up.sql 示例：加字段
ALTER TABLE mvp_project ADD COLUMN IF NOT EXISTS priority INT DEFAULT 0 COMMENT '优先级';

-- up.sql 示例：加索引
CREATE INDEX idx_project_priority ON mvp_project (priority);

-- up.sql 示例：建表
CREATE TABLE IF NOT EXISTS mvp_new_table (
  id BIGINT UNSIGNED NOT NULL COMMENT '主键',
  name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '名称',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人',
  dept_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '部门ID',
  created_at DATETIME DEFAULT NULL,
  updated_at DATETIME DEFAULT NULL,
  deleted_at DATETIME DEFAULT NULL,
  PRIMARY KEY (id),
  INDEX idx_datascope (dept_id, created_by),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='新表';
```

### 种子数据 vs 迁移

| 类型 | 放在哪里 | 说明 |
|------|---------|------|
| 建表/加字段/加索引 | `manifest/sql/mysql/` 迁移文件 | DDL 结构变更 |
| 系统初始数据（管理员账号、菜单、角色预设） | `manifest/sql/seed/` 种子文件 | 首次初始化数据 |
| 业务配置默认值 | `manifest/sql/mysql/` 迁移文件中 INSERT | 与结构一起版本化 |

## 五、故障排除

### dirty database

```
error: Dirty database version X. Fix and force version.
```

原因：上次迁移执行到一半失败了。

修复：
```bash
# 检查当前实际数据库状态，确认哪些 SQL 已执行
# 然后强制设置到正确的版本号
make db-force VERSION=X    # X 是最后成功的版本号
make db-up                 # 重新执行待迁移
```

### no change

```
no change
```

正常——数据库已经是最新版本，没有新迁移需要执行。

### 连接失败

```
error: failed to open database
```

检查：
1. MySQL 是否启动：`docker compose ps mysql`
2. 数据库连接信息：`make db-version`（会显示连接 URL）
3. 环境变量：检查 `.env` 中的 `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`

## 六、安装 migrate CLI

```bash
# Go 安装（推荐）
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# macOS
brew install golang-migrate

# Docker 容器中已预装（Dockerfile.admin-go.dev）
```

## 七、环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MIGRATE_DATABASE_URL` | 覆盖数据库连接 URL | 从 GoFrame config 自动解析 |
| `MIGRATE_BIN` | migrate 二进制路径 | `migrate` |
| `MIGRATIONS_DIR` | 迁移文件目录 | `manifest/sql/mysql` |
