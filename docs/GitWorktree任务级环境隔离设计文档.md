# GitWorktree任务级环境隔离设计文档

> 更新日期：2026-04-09

本文只描述当前 `internal/workspace` 包已经实现的内容，不再保留“建议新增”的隔离方案。

## 1. 当前实现概览

EasyMVP 现在已经把写仓执行器接到 `git worktree` 隔离目录中。核心入口在：

- `admin-go/app/mvp/internal/workspace/manager.go`
- `admin-go/app/mvp/internal/workspace/git_worktree.go`
- `admin-go/app/mvp/internal/workspace/cleanup.go`
- `admin-go/app/mvp/internal/workspace/workspace_repo.go`

隔离目录规则：

```text
{work_dir}/.mvp-worktrees/task-{task_id}
```

分支命名规则：

```text
mvp-task-{task_id}
```

## 2. 哪些执行器会使用 worktree

当前需要隔离工作区的执行模式：

- `aider`
- `openhands`
- `claude_code`
- `codex_cli`
- `gemini_cli`
- `auto`

当前不需要隔离工作区的模式：

- `chat`

执行器是否使用隔离，不再靠文档约定，而是由 `Executor.NeedsWorkspace()` 在运行时决定。

## 3. 创建流程

当前 `Prepare` 的真实流程如下：

1. 校验 `workDir`
2. 如果目录不是 Git 仓库，则自动：
   - 创建目录
   - `git init`
   - 写入本地 `user.name` / `user.email`
   - 创建空提交，满足 `git worktree` 的基线要求
3. 读取主工作区 `HEAD`
4. 在 `mvp_task_workspace` 写入一条 `creating` 记录
5. 创建 `.mvp-worktrees/task-{taskID}`
6. 通过 `git worktree add -b mvp-task-{taskID}` 建立隔离目录
7. 更新状态为 `ready`

这意味着编码类项目即使用户给的是空目录，也能被系统自动补成最小可运行的 Git 工作区。

## 4. 运行与收尾

### 4.1 运行阶段

当执行器真正开始执行时，会：

- 把工作目录切到 `WorkspacePath`
- 将工作空间状态从 `ready` 标成 `running`

### 4.2 成功阶段

当前成功收尾会做两件事：

1. 调用 `Finalize`
2. 采集 `git diff --stat` 结果写入 `diff_summary`

随后，大多数写仓执行器会直接触发 `Cleanup`，把 worktree 和临时分支删除。

### 4.3 失败阶段

失败路径会：

- 将工作空间标记为 `failed`
- 记录 `error_message`
- 不立即清理，留给后台保留期策略处理

### 4.4 定时清理

`RunCleanup` 当前默认保留期：

- 成功：24 小时
- 失败：72 小时
- 取消：24 小时

但要注意：

- 成功任务通常会被执行器立即清理
- 这个保留策略主要影响失败、取消或显式保留的工作空间

## 5. 数据模型

`mvp_task_workspace` 当前已经落库，关键字段包括：

- `task_id`
- `workflow_run_id`
- `project_id`
- `workspace_type`
- `workspace_path`
- `base_ref`
- `status`
- `cleanup_status`
- `diff_summary`
- `error_message`

当前状态值：

- `creating`
- `ready`
- `running`
- `completed`
- `failed`
- `canceled`

当前清理状态值：

- `pending`
- `done`
- `retained`
- `failed`

## 6. 当前边界

更新本文档时，和旧版本相比有一个必须明确写出来的边界：

- 当前 `workspace` 包已经负责：
  - 准备隔离目录
  - 收集 `diff_summary / patch_ref`
  - 根据风险矩阵决定 `patch / pr / manual` 交付路径
  - 在 `sync_strategy=auto_apply` 时把 worktree 变更同步回主工作区
  - 记录交付事件与清理状态
- 当前 `workspace` 包仍然不负责：
  - 容器或沙箱级安全隔离
  - 真正远端 PR 创建
  - 复杂冲突协商后的自动合并策略

因此，当前系统已经具备“受控自动回写”的能力，但不是“所有任务默认直接 merge 回主工作区”。

### 9.2 核心接口

```go
type WorkspaceManager interface {
    PrepareTaskWorkspace(ctx context.Context, taskID int64, projectID int64, workDir string) (*TaskWorkspace, error)
    GetTaskWorkspace(ctx context.Context, taskID int64) (*TaskWorkspace, error)
    FinalizeTaskWorkspace(ctx context.Context, taskID int64, result WorkspaceFinalizeRequest) error
    CleanupTaskWorkspace(ctx context.Context, taskID int64) error
}
```

---

## 十、基线选择策略

当前实现基于主工作区当前 `HEAD` 创建 worktree。

这是代码中的实际行为：

1. `Prepare` 时先确保主目录可作为 Git 基线使用
2. `git worktree add -b mvp-task-{taskID}` 从当前 `HEAD` 拉出任务分支
3. `Finalize` 时按交付策略决定写 patch、生成 PR 草稿说明，或自动回写主工作区

当前没有引入 `workflow_run` / `batch` 级独立基线版本管理。

---

## 十一、回写策略

当前实现不是方案讨论，而是明确的交付矩阵：

1. 所有成功任务都会优先尝试生成 `diff_summary` 与 `patch_ref`
2. 风险矩阵默认值为：
   - `low -> patch + auto_apply`
   - `medium -> patch + manual`
   - `high -> manual + manual`
3. 当 `delivery_mode=pr` 时，会额外生成 PR 草稿说明文件
4. 当 `sync_strategy=auto_apply` 时，会通过 `syncWorktreeCommit(...)` 将变更同步回主工作区
5. 当 `sync_strategy=manual` 时，只保留交付物和待审核/待回写状态，不直接改主工作区

因此，当前系统已经不再是“只保留 worktree 结果，不做主工作区沉淀”的第一阶段实现。

---

## 十二、与执行器的衔接方式

### 12.1 Aider

将 `cfg.WorkDir` 替换为任务 worktree 路径即可。

### 12.2 OpenHands

如果暂不做容器隔离，可先让 OpenHands 在 worktree 中运行。  
后续如果升级成 sandbox/container，worktree 可以作为宿主侧源目录。

### 12.3 Claude Code / Codex CLI / Gemini CLI

均可统一按：

- `workDir = task workspace`

执行。

---

## 十三、与调度器的衔接点

当前衔接点已经落地为：

1. 任务 `pending -> running` 前调用 `Prepare`
2. 执行器通过 `workspace_path` 在 task worktree 中运行
3. 任务成功后调用 `Finalize`：
   - 生成 `diff_summary / patch_ref / delivery_ref`
   - 记录交付事件
   - 视 `sync_strategy` 决定自动回写或待人工回写
4. 任务失败后也会进入 `Finalize`，并将交付状态标为 `skipped/failed`
5. `Cleanup` 与定时清理负责收尾目录和状态

---

## 十四、清理策略

当前清理语义已经体现在 `cleanup_status` 中：

1. 默认进入 `pending`
2. 显式保留时写为 `retained`
3. 清理成功后写为 `done`
4. 清理失败时写为 `failed`

失败任务默认不会立即删除 worktree，而是留给后续清理链路补偿处理。

---

## 十五、异常场景

### 15.1 主工作区不是 Git 仓库

当前实现会自动初始化最小 Git 基线，而不是直接降级回非隔离模式。

### 15.2 worktree 创建失败

任务不进入 running，直接失败回写。

### 15.3 worktree 清理失败

记录为 `cleanup_status=failed`，后续由清理任务补偿。

### 15.4 回写主工作区冲突

当前 `auto_apply` 回写失败会直接把 `sync_status` 置为 `failed`，并保留交付物供人工处理。

---

## 十六、并发与容量影响

引入 worktree 后：

### 好处

1. 降低主工作区污染
2. 降低执行器互相踩文件的风险

### 成本

1. 磁盘占用上升
2. Git 元数据操作增加
3. 创建/清理成本增加

在 9950X + 48G 环境下，worktree 隔离通常比容器隔离便宜得多，适合作为当前阶段默认方案。

---

## 十七、安全边界

worktree 隔离不是安全沙箱，它只解决：

1. 工作目录污染
2. 文件改动隔离

它不解决：

1. 进程隔离
2. 网络隔离
3. 系统调用隔离
4. 依赖安装污染宿主机

所以对 OpenHands 这类更重执行器，worktree 只是第一步，不是最终形态。

---

## 十八、当前结论

对于当前 EasyMVP 代码状态，`git worktree` 已不是待落地方案，而是正式使用中的任务级隔离实现。

它当前已经解决：

1. 写仓任务不直接污染主工作区
2. 不同任务的文件改动隔离
3. patch / PR 草稿 / 人工交付三种结果形态
4. 低风险自动回写与中高风险人工复核的分流

它当前仍不解决：

1. 容器级或沙箱级隔离
2. 远端 PR 创建与仓库托管集成
3. 复杂冲突下的全自动合并策略
