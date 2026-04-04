# GitWorktree任务级环境隔离设计文档

## 一、文档目标

本文档定义 EasyMVP 在当前 WorkflowRun 与多执行器架构下，基于 `git worktree` 的任务级环境隔离方案。

目标：

1. 降低任务执行对主工作区的污染风险。
2. 支持 Aider、OpenHands、Claude Code、Codex CLI、Gemini CLI 等写仓执行器的安全运行。
3. 在不引入过重容器复杂度的前提下，尽快为任务执行提供隔离能力。

---

## 二、为什么现在需要隔离

当前系统的写仓任务主要直接在项目 `work_dir` 上执行。

现有保护手段包括：

1. 批次门控
2. 依赖检查
3. `affected_resources` 资源锁

这些能力可以减少并发文件冲突，但不能解决以下问题：

1. 任务误改未声明文件
2. 临时文件残留污染主工作区
3. 依赖安装、副作用命令污染环境
4. 失败任务留下半成品状态
5. 多执行器混跑时互相影响

因此，需要为“写仓任务”引入任务级隔离。

---

## 三、为什么选择 Git Worktree

相较于“直接主工作区运行”和“每任务容器沙箱”，`git worktree` 更适合当前阶段。

### 3.1 相较直接主工作区执行

优点：

1. 每个任务有独立目录
2. 主工作区不会被直接污染
3. 任务失败后可以直接清理工作树

### 3.2 相较容器隔离

优点：

1. 实施成本更低
2. 不需要先解决镜像、挂载、容器编排问题
3. 更容易接入现有 Aider/CLI 执行器
4. 调试成本低

### 3.3 适用范围

适合：

1. Git 仓库类项目
2. 文件编辑型任务
3. 本地命令行执行器

不适合：

1. 非 Git 工作目录
2. 强依赖系统级副作用的任务
3. 必须完全隔离进程/网络/依赖的重型 Agent

---

## 四、目标架构

### 4.1 核心思想

主工作区只作为“源仓库入口”，任务不直接在主工作区执行。

每个写仓任务执行前：

1. 为任务分配独立 `worktree`
2. 执行器在该 `worktree` 中运行
3. 任务结束后提取结果
4. 根据策略决定：
   - 回写主工作区
   - 保留隔离目录用于排查
   - 或直接清理

### 4.2 结构示意

```text
project work_dir (主工作区)
└── .git

task worktrees
└── .mvp-worktrees/
    ├── task-315xxxx001/
    ├── task-315xxxx002/
    └── task-315xxxx003/
```

---

## 五、隔离范围

### 5.1 必须启用 worktree 的任务

1. `aider`
2. `openhands`
3. `claude_code`
4. `codex_cli`
5. `gemini_cli`

这些执行器都属于“写仓任务执行器”。

### 5.2 可先不启用的任务

1. `chat`
2. 纯审核任务
3. 纯分析任务
4. 不落地文件的对话型任务

### 5.3 后续可配置化

建议支持：

- 项目级配置
- 执行器级默认配置
- 任务级强制开关

---

## 六、目录规划

建议在项目根目录下维护专用工作树目录：

```text
{work_dir}/.mvp-worktrees/
```

每个任务目录命名建议：

```text
task-{task_id}
```

示例：

```text
/workspace/my-project/.mvp-worktrees/task-315667201073876992
```

这样便于：

1. 排查任务问题
2. 清理孤儿工作树
3. 与任务 ID 直接映射

---

## 七、生命周期设计

### 7.1 创建阶段

任务调度进入执行前：

1. 校验主工作区是否是合法 Git 仓库
2. 选择隔离基线分支/提交
3. 创建 worktree
4. 记录 worktree 元数据
5. 将执行器工作目录切换到 worktree

### 7.2 执行阶段

执行器只在 worktree 内运行：

1. 读取文件
2. 修改文件
3. 生成中间产物
4. 产生日志

不得直接回到主工作区执行。

### 7.3 结束阶段

任务结束后：

1. 收集变更文件
2. 生成 diff / 结果摘要
3. 根据策略回写主工作区或保留 worktree
4. 标记 worktree 状态

### 7.4 清理阶段

默认策略建议：

1. 成功任务：回写后清理
2. 失败任务：保留一段时间
3. 超时/取消任务：保留待诊断

---

## 八、数据模型建议

建议新增表：

### 8.1 `mvp_task_workspace`

字段建议：

- `id`
- `task_id`
- `workflow_run_id`
- `project_id`
- `workspace_type`
- `workspace_path`
- `base_ref`
- `status`
- `cleanup_status`
- `created_at`
- `updated_at`
- `deleted_at`

### 8.2 字段含义

- `workspace_type`
  - `git_worktree`
- `status`
  - `creating`
  - `ready`
  - `running`
  - `completed`
  - `failed`
  - `canceled`
- `cleanup_status`
  - `pending`
  - `done`
  - `retained`
  - `failed`

---

## 九、核心组件设计

### 9.1 `TaskWorkspaceManager`

建议新增：

```text
admin-go/app/mvp/internal/workspace/
```

建议核心组件：

1. `manager.go`
2. `git_worktree.go`
3. `cleanup.go`
4. `workspace_repo.go`

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

