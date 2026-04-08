# GitWorktree任务级环境隔离设计文档

> 更新日期：2026-04-08

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

- 当前 `workspace` 包负责的是“准备隔离目录、记录 diff、清理目录”
- 当前 `workspace` 包本身没有实现“把 worktree 变更自动 merge / cherry-pick 回主工作区”

因此，本文档不再写“系统已经自动回写主工作区”。如果后续要把 worktree 变更正式沉淀回项目主目录，需要在工作流主链里单独补齐合入策略。

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

### 10.1 默认策略

基于主工作区当前 HEAD 创建 worktree。

适合当前阶段，实施简单。

### 10.2 后续增强

可升级为基于：

1. `workflow_run` 开始时的基线提交
2. `batch` 开始时的基线提交
3. 上游已确认变更后的最新提交

当前建议先用默认策略，避免过早引入复杂的多版本仓库管理。

---

## 十一、回写策略

这里是最关键的设计点。

### 11.1 方案 A：直接保留 worktree 结果，不自动回写

优点：

1. 安全
2. 实现简单

缺点：

1. 结果不能自动进入主工作区
2. 需要人工处理

### 11.2 方案 B：生成 diff，再自动应用到主工作区

优点：

1. 主工作区仍是统一结果源
2. 可控

缺点：

1. 需要处理 patch 冲突

### 11.3 方案 C：成功任务直接 merge/worktree 回写

优点：

1. 自动化程度高

缺点：

1. 风险高
2. 与当前主链不够匹配

### 11.4 当前推荐

建议分阶段：

1. 第一阶段：先记录 worktree + diff，不自动 merge
2. 第二阶段：对低风险任务开放自动 patch 回写
3. 第三阶段：再考虑更复杂的自动合并

如果你们想要最快可用，当前也可以采用：

1. Aider 先继续直接写 worktree
2. 执行完成后使用受控 patch 回写主工作区

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

### 13.1 调度前准备

在任务 `pending -> running` 之间插入：

1. `PrepareTaskWorkspace`

### 13.2 执行器运行

执行器获取 workspace 路径后启动。

### 13.3 任务完成/失败

在：

- `OnTaskCompleted`
- `OnTaskFailed`

之后补：

1. `FinalizeTaskWorkspace`
2. `CleanupTaskWorkspace` 或保留

---

## 十四、清理策略

### 14.1 默认保留策略

1. 成功：保留 24 小时
2. 失败：保留 72 小时
3. 取消：保留 24 小时

### 14.2 定时清理

建议增加定时任务：

1. 清理已失效 worktree
2. 清理 orphan 目录
3. 清理失效数据库记录

---

## 十五、异常场景

### 15.1 主工作区不是 Git 仓库

策略：

1. 直接降级回现有模式
2. 或阻止需要隔离的执行器运行

建议对写仓任务默认阻止，并返回清晰错误。

### 15.2 worktree 创建失败

任务不进入 running，直接失败回写。

### 15.3 worktree 清理失败

记录为 `cleanup_status=failed`，后续由清理任务补偿。

### 15.4 回写主工作区冲突

策略：

1. 标记任务为 `completed_with_conflict` 或 `failed`
2. 保留 worktree
3. 通知人工处理

当前阶段建议直接保留 worktree 并标记失败/警告。

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

## 十八、实施计划

### 阶段 1：基础设施

1. 新增 `mvp_task_workspace`
2. 新增 `workspace/` 模块
3. 实现 `git worktree prepare/cleanup`

### 阶段 2：Aider 接入

1. Aider 改为使用 task workspace
2. 记录日志、diff、产物

### 阶段 3：CLI 执行器接入

1. Claude Code
2. Codex CLI
3. Gemini CLI

### 阶段 4：OpenHands 过渡接入

1. 先用 worktree 作为目录隔离
2. 后续升级为容器隔离

---

## 十九、验收标准

1. 写仓任务默认不直接在主工作区运行
2. 每个任务都能定位到独立 workspace
3. 任务失败后主工作区不残留半成品污染
4. 可清理、可保留、可诊断
5. 与 WorkflowRun 主链兼容

---

## 二十、结论

对于当前 EasyMVP 阶段，`git worktree` 是最合适的第一版任务级环境隔离方案。

它不是最终形态，但它能用较低成本解决当前最现实的问题：

1. 主工作区污染
2. 多任务执行互相影响
3. 写仓执行器缺少隔离

推荐策略：

1. 现在先上 `git worktree`
2. OpenHands 后续再升级到更重的 sandbox/container
