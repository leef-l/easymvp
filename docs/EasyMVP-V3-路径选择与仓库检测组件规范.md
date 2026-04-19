# EasyMVP V3 路径选择与仓库检测组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 关联文档：[EasyMVP-V3-创建项目弹层线框图设计](./EasyMVP-V3-创建项目弹层线框图设计.md)
> 关联文档：[EasyMVP-V3-本地目录与项目工作区规范](./EasyMVP-V3-本地目录与项目工作区规范.md)
> 目标：定义创建项目时路径选择、仓库识别、路径健康检查和错误提示的组件行为规范。

## 1. 设计结论

路径选择不是普通输入框。

它是创建项目过程中最容易出错的一步，所以必须做到：

1. 选择清楚
2. 检查即时
3. 提示直接
4. 失败可恢复

## 2. 组件职责

该组件至少负责：

1. 路径输入与浏览
2. 工作区来源切换
3. 基础文件系统检查
4. 仓库识别
5. 风险提示

## 3. 组件结构

建议组件由 5 个子区组成：

1. `source_selector`
2. `path_input`
3. `browse_action`
4. `check_summary`
5. `detail_warnings`

## 4. 来源模式

建议固定三种来源模式：

1. `create_new_workspace`
2. `use_existing_folder`
3. `use_existing_repository`

### 4.1 `create_new_workspace`

表示：

1. 选择一个父目录
2. 系统在其中创建新的项目目录

### 4.2 `use_existing_folder`

表示：

1. 使用一个已存在的本地目录
2. 不要求该目录必须是 git 仓库

### 4.3 `use_existing_repository`

表示：

1. 使用一个已存在的 git 仓库目录
2. 必须通过 git 识别检查

## 5. 检查项

建议至少做以下检查：

1. `path_exists`
2. `path_readable`
3. `path_writable`
4. `workspace_conflict`
5. `git_repository_detected`
6. `project_root_detected`
7. `large_noise_directories_detected`

## 6. 检查结果展示

### 6.1 一行摘要

建议路径输入框下方直接显示摘要状态：

1. `✓ Writable`
2. `✓ Git Repo`
3. `! Already linked to another project`
4. `! Large build directories detected`

### 6.2 详细说明

必要时展开详细说明区，显示：

1. 检查项名称
2. 当前结果
3. 风险说明
4. 建议动作

## 7. 状态定义

建议组件状态固定为：

1. `idle`
2. `checking`
3. `valid`
4. `warning`
5. `invalid`

### 7.1 `idle`

尚未输入路径。

### 7.2 `checking`

路径变化后正在检查。

### 7.3 `valid`

可直接用于创建。

### 7.4 `warning`

可创建，但存在噪音或非阻塞风险。

### 7.5 `invalid`

不可继续。

## 8. 建议返回模型

建议定义 `WorkspacePathCheckResult`：

```json
{
  "path": "/Users/demo/projects/my-app",
  "source_mode": "use_existing_repository",
  "status": "warning",
  "checks": [
    {
      "check_key": "path_writable",
      "status": "pass",
      "summary": "Path is writable"
    },
    {
      "check_key": "git_repository_detected",
      "status": "pass",
      "summary": "Git repository detected"
    },
    {
      "check_key": "large_noise_directories_detected",
      "status": "warning",
      "summary": "Large node_modules directory detected"
    }
  ],
  "can_continue": true
}
```

## 9. 交互规则

### 9.1 输入后自动检查

路径输入变化后，建议短延迟自动检查。

### 9.2 浏览选择后立即检查

用户通过 `Browse` 选择目录后，立即检查，不等用户再点一次按钮。

### 9.3 阻塞错误禁止继续

以下错误必须阻塞创建：

1. 路径不存在
2. 无写权限
3. 已被其他项目绑定
4. 选择仓库模式但不是 git 仓库

### 9.4 warning 允许继续

以下问题允许继续，但必须清晰提示：

1. 大型构建目录存在
2. 仓库根目录噪音较多
3. 路径层级过深

## 10. 文案原则

建议文案直接说人话：

1. `This folder is writable`
2. `Git repository detected`
3. `This folder is already linked to another project`
4. `Large generated folders were detected and may slow down scanning`

不要只显示：

1. `path_not_accessible`
2. `workspace_conflict`
3. `repo_check_failed`

## 11. 视觉表现

建议：

1. 输入框旁放 `Browse`
2. 摘要状态用短标签
3. warning 用柔和告警色
4. invalid 用明确错误色

不建议：

1. 全部提示都做成红字
2. 把所有检查都堆成大表格

## 12. 不该怎么做

这个组件不应该：

1. 只在提交时才检查
2. 检查结果藏在 tooltip
3. 要用户自己理解 git/path 权限错误
4. 把 warning 和 invalid 混在一起

## 13. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-路径检查接口设计](./EasyMVP-V3-路径检查接口设计.md)
2. 本地浏览器桥接规范
3. 仓库噪音目录忽略策略
