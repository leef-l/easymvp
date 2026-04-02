# GBaseAdmin - System 应用

系统管理应用，提供基础的权限管理功能。

## 数据库表

| 表名 | 说明 |
|------|------|
| `system_dept` | 部门表 |
| `system_role` | 角色表 |
| `system_menu` | 菜单表 |
| `system_users` | 用户表 |
| `system_user_role` | 用户-角色关联表 |
| `system_role_menu` | 角色-菜单关联表 |
| `system_role_dept` | 角色-部门关联表 |
| `system_user_dept` | 用户-部门关联表 |

## API 路由

所有接口前缀：`/system/`

| 模块 | 路径 | 说明 |
|------|------|------|
| 认证 | `/system/auth/login` | 登录 |
| 认证 | `/system/auth/info` | 当前用户信息 |
| 认证 | `/system/auth/menus` | 当前用户菜单 |
| 部门 | `/system/dept/list,create,update,delete,detail,tree` | 部门管理 |
| 角色 | `/system/role/list,create,update,delete,detail,tree` | 角色管理 |
| 菜单 | `/system/menu/list,create,update,delete,detail,tree` | 菜单管理 |
| 用户 | `/system/users/list,create,update,delete,detail` | 用户管理 |

## 权限模型

- 角色设置 `is_admin=1` 为超级管理员，跳过权限判断，可查看所有菜单和数据
- 普通角色通过 `system_role_menu` 关联表控制菜单权限
- 数据权限通过 `system_role_dept` 关联表控制（`data_scope=5` 时生效）
