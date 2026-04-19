# EasyMVP V3 路径检查接口设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-路径选择与仓库检测组件规范](./EasyMVP-V3-路径选择与仓库检测组件规范.md)
> 关联文档：[EasyMVP-V3-创建项目流程与页面设计](./EasyMVP-V3-创建项目流程与页面设计.md)
> 目标：定义创建项目时本地路径检查接口、返回模型和错误分类。

## 1. 接口

建议：

`POST /api/v3/workspace/path-check`

## 2. 请求体

```json
{
  "source_mode": "use_existing_repository",
  "path": "/Users/demo/projects/my-app"
}
```

## 3. 返回体

复用 `WorkspacePathCheckResult`：

```json
{
  "path": "/Users/demo/projects/my-app",
  "status": "warning",
  "can_continue": true,
  "checks": []
}
```

## 4. 错误分类

建议：

1. `path_not_found`
2. `path_not_writable`
3. `workspace_conflict`
4. `git_repo_required`

## 5. 不该怎么做

不应该：

1. 让前端自行猜错误
2. 返回不稳定字段

