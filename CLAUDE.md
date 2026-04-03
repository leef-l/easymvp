# EasyMVP 项目指令

## 模型分工约定

- **Opus 4.6** — 分析、规划、设计方案、把控全局
- **Sonnet 4.6** — 具体实施、上下文多/复杂的任务
- **Haiku 4.5** — 小任务、关联性不大的独立任务

## AI 执行引擎开发必读

- **开始 OpenHands / Aider 接入相关开发前，必须先阅读** `docs/EasyMVP对接OpenHands与Aider引擎设计实现文档.md`
- 该文档是当前 AI 执行引擎接入的设计基线，涉及数据库、角色授权、AI 配置、执行任务、前后端页面与安全边界
- 如果实现与文档不一致，优先更新文档后再继续开发

## 技术栈

### 后端
- **语言**: Go 1.25
- **框架**: GoFrame v2.10（MonoRepo 多应用架构）
- **数据库**: MySQL 8.0（本机）
- **认证**: JWT（自定义实现，`utility/jwt`）
- **ID策略**: Snowflake 雪花ID（`utility/snowflake`，类型 `snowflake.JsonInt64`）

### 管理端前端
- **框架**: Vue 3 + Vben Admin v5.7（`vue-vben-admin/apps/web-antd`）
- **UI库**: Ant Design Vue
- **表格**: VxeTable（通过 `useVbenVxeGrid`）
- **表单**: VbenForm（支持自定义组件）
- **构建**: Vite + pnpm

### 代码生成器
- **位置**: `admin-go/codegen/`
- **语言**: Go（独立工具，非 GoFrame 应用）
- **模板**: Go `text/template`（`codegen/templates/`）

---

## 项目目录结构

```
easymvp/
├── admin-go/                     # 后端（Go MonoRepo）
│   ├── app/                      # 应用目录
│   │   ├── system/               # 系统管理（用户、角色、部门、菜单）
│   │   ├── svc-template/         # 新应用模板
│   │   └── job-template/         # 定时任务模板
│   ├── codegen/                  # 代码生成器
│   │   ├── main.go               # CLI 入口
│   │   ├── codegen.yaml          # 数据库连接和输出配置
│   │   ├── parser/               # 表结构解析器
│   │   ├── generator/            # 生成器（backend/frontend/menu）
│   │   ├── templates/            # 模板文件
│   │   └── sql/                  # 数据库初始化SQL
│   ├── utility/                  # 公共工具
│   │   ├── jwt/                  # JWT 工具
│   │   ├── snowflake/            # 雪花ID生成器
│   │   ├── response/             # 统一响应
│   │   └── oplog/                # 操作日志
│   └── deploy/                   # 部署配置
│
└── vue-vben-admin/               # 管理端前端
    └── apps/web-antd/src/
        ├── api/                  # API 调用层
        ├── views/                # 页面
        ├── components/           # 自定义组件
        └── router/               # 路由
```

---

## 代码生成器使用

### 命令格式
```bash
cd admin-go/codegen
go run . -table <表名> [选项]
```

### 常用命令
```bash
# 生成完整代码（后端 + 前端 + 菜单）
go run . -table system_xxx -force -menu

# 只生成后端
go run . -table system_xxx -only backend -force

# 只生成前端
go run . -table system_xxx -only frontend -force

# 预览（不实际生成）
go run . -table system_xxx -dry-run
```

### 表名约定
- 表名格式: `{应用}_{模块}`
- 应用前缀决定生成到哪个 app 目录

---

## 数据库

- **地址**: 127.0.0.1:3306
- **用户名**: easymvp
- **密码**: JKcHFJYXnkrB6BXE
- **数据库**: easymvp

```bash
mysql -u easymvp -pJKcHFJYXnkrB6BXE -h 127.0.0.1 -P 3306 easymvp
```

---

## 重要约定

### 后端约定
- 所有 ID 字段使用 `snowflake.JsonInt64` 类型
- 软删除: `deleted_at` 字段
- 数据隔离: `dept_id` + `created_by` 字段
- 密码加密: SHA256（`gsha256.Encrypt`）

### 代码生成器铁律
- **生成的代码有问题，先修生成器模板再重新生成**，不手写修复
- 生成器模板在 `codegen/templates/` 下

---

## 常用命令

### 后端
```bash
# 编译 system 应用
cd admin-go && go build ./app/system/...

# 生成 DAO
cd admin-go/app/system && gf gen dao
```

### 前端
```bash
cd vue-vben-admin
pnpm install
pnpm dev        # 开发服务
pnpm build      # 生产构建
```
