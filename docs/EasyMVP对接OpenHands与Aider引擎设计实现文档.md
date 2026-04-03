# EasyMVP 对接 OpenHands 与 Aider 引擎设计实现文档

## 1. 背景

当前 EasyMVP 需要支持在系统内执行“写代码、改文件、执行命令”等 AI Agent 能力，并允许后台灵活切换使用 `Aider` 或 `OpenHands`。为了兼顾权限、安全、可扩展性与审计能力，采用“角色授权 + 引擎配置中心 + 统一任务执行”的设计。

## 2. 目标

- 同时接入 `Aider` 与 `OpenHands`
- 后台可配置引擎连接参数、启停状态、默认模型
- 角色可配置可用引擎与默认引擎
- 所有执行动作统一走任务中心，保留日志和审计信息
- 为后续扩展第三种引擎保留统一抽象

## 3. 总体设计

### 3.1 分层原则

- `system` 模块负责角色层控制
  - 角色可用引擎
  - 角色默认引擎
- `ai` 模块负责引擎层与执行层
  - 引擎定义
  - 引擎连接配置
  - 执行任务
  - 执行日志

### 3.2 设计结论

角色里只放“权限和默认值”，不放连接地址、密钥、命令模板等细节配置。连接信息统一放在 AI 配置中心。

## 4. 数据库设计

### 4.1 `system_role` 增加默认引擎字段

```sql
ALTER TABLE `system_role`
ADD COLUMN `default_ai_engine` varchar(32) DEFAULT NULL COMMENT '默认AI执行引擎: aider/openhands' AFTER `is_admin`;
```

### 4.2 `ai_engine`

```sql
CREATE TABLE `ai_engine` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `code` varchar(32) NOT NULL COMMENT '引擎编码: aider/openhands',
  `name` varchar(64) NOT NULL COMMENT '引擎名称',
  `description` varchar(255) DEFAULT NULL COMMENT '说明',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1启用 0禁用',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `dept_id` bigint NOT NULL DEFAULT 0 COMMENT '部门ID',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `updated_by` bigint NOT NULL DEFAULT 0 COMMENT '更新人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ai_engine_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI执行引擎定义表';
```

初始化数据：

```sql
INSERT INTO `ai_engine` (`id`, `code`, `name`, `description`, `status`, `sort`, `dept_id`, `created_by`)
VALUES
(100001, 'aider', 'Aider', '本地命令行代码执行引擎', 1, 10, 0, 1),
(100002, 'openhands', 'OpenHands', '远程Agent执行引擎', 1, 20, 0, 1);
```

### 4.3 `ai_engine_config`

```sql
CREATE TABLE `ai_engine_config` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `engine_code` varchar(32) NOT NULL COMMENT '引擎编码',
  `base_url` varchar(255) DEFAULT NULL COMMENT '服务地址',
  `api_key` text COMMENT 'API Key(建议加密存储)',
  `default_model_id` bigint DEFAULT NULL COMMENT '默认模型ID, 关联ai_model.id',
  `timeout_seconds` int NOT NULL DEFAULT 600 COMMENT '超时时间(秒)',
  `max_steps` int NOT NULL DEFAULT 20 COMMENT '最大执行步数',
  `workspace_root` varchar(500) DEFAULT NULL COMMENT '工作区根目录',
  `command_template` varchar(1000) DEFAULT NULL COMMENT '命令模板, 主要用于aider',
  `callback_url` varchar(255) DEFAULT NULL COMMENT '回调地址',
  `callback_secret` varchar(255) DEFAULT NULL COMMENT '回调密钥',
  `extra_config` json DEFAULT NULL COMMENT '额外配置JSON',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1启用 0禁用',
  `dept_id` bigint NOT NULL DEFAULT 0 COMMENT '部门ID',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `updated_by` bigint NOT NULL DEFAULT 0 COMMENT '更新人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ai_engine_config_engine_code` (`engine_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI执行引擎配置表';
```

### 4.4 `system_role_ai_engine`

```sql
CREATE TABLE `system_role_ai_engine` (
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `engine_code` varchar(32) NOT NULL COMMENT '引擎编码',
  PRIMARY KEY (`role_id`, `engine_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色可用AI执行引擎';
```

### 4.5 `ai_task`

```sql
CREATE TABLE `ai_task` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `title` varchar(255) NOT NULL COMMENT '任务标题',
  `engine_code` varchar(32) NOT NULL COMMENT '执行引擎',
  `role_id` bigint DEFAULT NULL COMMENT '发起时角色ID',
  `user_id` bigint NOT NULL COMMENT '发起用户ID',
  `project_id` bigint DEFAULT NULL COMMENT '项目ID, 可空',
  `repo_path` varchar(500) NOT NULL COMMENT '仓库路径',
  `worktree_path` varchar(500) DEFAULT NULL COMMENT '执行工作目录',
  `branch_name` varchar(255) DEFAULT NULL COMMENT '分支名称',
  `instruction` longtext NOT NULL COMMENT '用户指令',
  `engine_config_snapshot` json DEFAULT NULL COMMENT '执行时配置快照',
  `request_payload` json DEFAULT NULL COMMENT '发送给引擎的请求体',
  `response_summary` longtext COMMENT '执行结果摘要',
  `error_message` longtext COMMENT '错误信息',
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/cancelled',
  `started_at` datetime DEFAULT NULL COMMENT '开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '结束时间',
  `dept_id` bigint NOT NULL DEFAULT 0 COMMENT '部门ID',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `updated_by` bigint NOT NULL DEFAULT 0 COMMENT '更新人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_ai_task_user_id` (`user_id`),
  KEY `idx_ai_task_engine_code` (`engine_code`),
  KEY `idx_ai_task_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI执行任务表';
```

### 4.6 `ai_task_log`

```sql
CREATE TABLE `ai_task_log` (
  `id` bigint NOT NULL COMMENT '主键ID',
  `task_id` bigint NOT NULL COMMENT '任务ID',
  `seq` int NOT NULL DEFAULT 0 COMMENT '日志序号',
  `log_type` varchar(32) NOT NULL DEFAULT 'stdout' COMMENT 'stdout/stderr/system/event',
  `content` longtext NOT NULL COMMENT '日志内容',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ai_task_log_task_id` (`task_id`),
  KEY `idx_ai_task_log_task_seq` (`task_id`, `seq`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI任务执行日志';
```

## 5. 后端设计

### 5.1 system 模块

职责：

- 角色默认引擎字段管理
- 角色引擎授权

建议扩展：

- `admin-go\app\system\api\system\v1\role.go`
- `admin-go\app\system\internal\controller\role\role.go`
- `admin-go\app\system\internal\service\role.go`
- `admin-go\app\system\internal\logic\role\role.go`
- `admin-go\app\system\internal\model\role.go`
- 新增 `role_ai_engine` 对应 dao/do/entity

新增接口：

- `POST /api/system/role/grant-ai-engine`
- `GET /api/system/role/ai-engine-codes`

### 5.2 ai 模块

职责：

- 管理执行引擎配置
- 接收执行任务
- 路由到对应执行器
- 写入任务日志和状态

建议新增模块：

- `engine`
- `task`
- `executor`

新增接口：

- `GET /api/ai/engine/list`
- `GET /api/ai/engine/detail`
- `POST /api/ai/engine/update`
- `POST /api/ai/engine/test-connection`
- `POST /api/ai/task/execute`
- `GET /api/ai/task/list`
- `GET /api/ai/task/detail`
- `GET /api/ai/task/logs`
- `POST /api/ai/task/cancel`

### 5.3 执行器抽象

```go
type IExecutor interface {
    Code() string
    TestConnection(ctx context.Context, cfg *entity.AiEngineConfig) error
    Execute(ctx context.Context, task *entity.AiTask, cfg *entity.AiEngineConfig) error
    Cancel(ctx context.Context, task *entity.AiTask, cfg *entity.AiEngineConfig) error
}
```

建议实现：

- `AiderExecutor`
- `OpenHandsExecutor`
- `Factory`

### 5.4 安全要求

- 执行仓库路径必须在白名单根目录内
- 密钥建议加密存储
- 日志输出要脱敏
- 任务必须保留完整审计

## 6. 前端设计

### 6.1 角色管理页增强

在 `系统管理 / 角色管理` 中增加：

- `默认引擎` 字段
- `AI引擎权限` 按钮

建议新增文件：

- `vue-vben-admin\apps\web-antd\src\views\system\role\modules\grant-ai-engine.vue`

### 6.2 AI 管理页新增执行引擎配置

建议新增：

- `vue-vben-admin\apps\web-antd\src\views\ai\engine\index.vue`
- `vue-vben-admin\apps\web-antd\src\api\ai\engine\index.ts`
- `vue-vben-admin\apps\web-antd\src\api\ai\engine\types.ts`

字段建议：

- 引擎名称
- 引擎编码
- 启用状态
- 默认模型
- 超时时间
- 最大步数
- 工作区根目录
- Base URL
- API Key
- 命令模板
- 额外配置 JSON

### 6.3 AI 任务页

建议新增：

- `vue-vben-admin\apps\web-antd\src\views\ai\task\index.vue`
- `vue-vben-admin\apps\web-antd\src\views\ai\task\modules\execute-form.vue`
- `vue-vben-admin\apps\web-antd\src\views\ai\task\modules\log-drawer.vue`
- `vue-vben-admin\apps\web-antd\src\api\ai\task\index.ts`
- `vue-vben-admin\apps\web-antd\src\api\ai\task\types.ts`

## 7. 权限与菜单

建议新增权限：

```text
ai:engine:list
ai:engine:detail
ai:engine:update
ai:engine:test
ai:task:execute
ai:task:list
ai:task:detail
ai:task:log
ai:task:cancel
system:role:grant-ai-engine
system:role:view-ai-engine
```

建议新增菜单：

- `AI管理 / 执行引擎配置`
- `AI管理 / 执行任务`

## 8. 推荐实施顺序

1. 完成数据库结构与 DAO 生成
2. 完成后端 role + engine + task API
3. 完成执行器骨架与任务流转
4. 完成前端角色授权、引擎配置、任务页面
5. 补菜单、权限、构建验证

## 9. V1 范围

V1 必做：

- 角色默认引擎
- 角色引擎授权
- 引擎配置管理
- 统一任务执行入口
- Aider/OpenHands 执行器骨架
- 任务日志与状态页

V1 暂不做：

- 角色级模型权限
- 项目级引擎策略
- 多租户隔离
- 审批流

## 10. 开发注意事项

- 该需求涉及后端、前端、菜单、权限、数据库多处联动，开始开发前必须先阅读本文档
- 角色中只放“权限和默认值”，连接配置统一在 AI 配置中心
- 所有执行任务必须记录配置快照，保证历史审计可追溯
