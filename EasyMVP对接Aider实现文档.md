# EasyMVP 对接 Aider 实现文档

> 版本：v1.0 | 更新日期：2026-04-02
>
> 本文档详细记录 EasyMVP 如何集成 Aider CLI 工具，让 AI 实施员角色能够真实编辑代码文件。

---

## 1. 为什么选择 Aider

### 1.1 背景问题

EasyMVP 最初的 Executor 只支持 ChatStream 对话模式——AI 只是"说"出代码方案，但不能真正修改文件。实施员和其他角色一样，只是产生一段文本回复，用户需要手动复制代码到项目中。

### 1.2 方案对比

| 方案 | 优点 | 缺点 | 复杂度 |
|------|------|------|--------|
| **自建 Tool Use（Function Calling）** | 完全可控 | 需要实现 read_file/write_file/exec_cmd 等工具，沙箱安全问题 | 高 |
| **Claude Code CLI** | 功能强大，自带终端 | 仅支持 Anthropic 模型，需要付费 | 中 |
| **Aider CLI** | 成熟稳定，多模型支持，Git 感知，多文件编辑 | CLI 调用，需要解析输出 | 中 |
| **OpenHands/Devin 模式** | 完整沙箱环境 | 架构过重，部署复杂 | 高 |

### 1.3 选择 Aider 的理由

1. **多模型支持**：通过 litellm，支持 Anthropic / OpenAI / 任何 OpenAI 兼容 API
2. **真实代码编辑**：直接修改文件，不是输出文本
3. **Git 感知**：自动检测仓库状态，可选自动提交
4. **多文件编辑**：`--file` 指定需要编辑的文件，`--read` 指定只读参考文件
5. **非交互模式**：`--message` 单次执行，适合程序化调用
6. **零侵入**：不需要修改现有 AI Provider 代码

---

## 2. Aider 安装与配置

### 2.1 安装

```bash
pip3 install aider-chat
```

验证：
```bash
aider --version
# aider 0.86.2
```

### 2.2 腾讯云 Coding Plan 配置

EasyMVP 当前使用腾讯云 Coding Plan 作为 AI 供应商，所有模型走 Anthropic 兼容协议。

**关键发现：** 腾讯云 Coding Plan 的 API 地址是 `https://api.lkeap.cloud.tencent.com/coding/anthropic/v1`，但 Aider 的 litellm 会自动拼接 `/v1/messages`，所以需要去掉末尾的 `/v1`。

**正确配置：**
```bash
# 环境变量
ANTHROPIC_API_KEY=sk-sp-xxxxx
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic  # 注意：不含 /v1

# 模型名前缀
--model anthropic/tc-code-latest
--model anthropic/glm-5
--model anthropic/hunyuan-t1
```

**错误配置（会导致 404）：**
```bash
# 错误：包含 /v1
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic/v1
```

### 2.3 验证连通性

```bash
# 测试 tc-code-latest（架构师模型）
ANTHROPIC_API_KEY=sk-sp-xxxxx \
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic \
aider --model anthropic/tc-code-latest \
  --no-auto-commits --no-show-model-warnings \
  --no-pretty --no-stream --yes-always \
  --message "say hello" --no-git

# 测试 glm-5（实施员模型）
ANTHROPIC_API_KEY=sk-sp-xxxxx \
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic \
aider --model anthropic/glm-5 \
  --no-auto-commits --no-show-model-warnings \
  --no-pretty --no-stream --yes-always \
  --message "say hello" --no-git

# 测试 hunyuan-t1（轻量实施员模型）
ANTHROPIC_API_KEY=sk-sp-xxxxx \
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic \
aider --model anthropic/hunyuan-t1 \
  --no-auto-commits --no-show-model-warnings \
  --no-pretty --no-stream --yes-always \
  --message "say hello" --no-git
```

### 2.4 真实代码编辑测试

```bash
# 创建测试项目
mkdir -p /tmp/aider-test && cd /tmp/aider-test && git init

# 创建测试文件
cat > hello.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}
EOF

# 用 Aider 编辑
ANTHROPIC_API_KEY=sk-sp-xxxxx \
ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic \
aider --model anthropic/glm-5 \
  --no-auto-commits --no-show-model-warnings \
  --no-pretty --no-stream --yes-always \
  --file hello.go \
  --message "把 hello world 改成 hello from EasyMVP，添加 add(a, b int) int 函数"

# 验证文件被修改
cat hello.go
```

**预期输出：**
```go
package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func main() {
    fmt.Println("hello from EasyMVP")
}
```

---

## 3. 架构设计

### 3.1 整体流程

```
任务调度器 (Scheduler)
  │
  ├─ 分配 implementer 任务
  │
  ▼
任务执行器 (Executor)
  │
  ├─ role_type == "implementer"?
  │   ├─ YES → executeWithAider()
  │   └─ NO  → executeWithChat() (ChatStream 对话)
  │
  ▼
executeWithAider()
  │
  ├─ 1. 查询项目 work_dir
  ├─ 2. 构建 taskPrompt（任务描述 + 前置任务结果）
  ├─ 3. 解析 affected_resources → 文件列表
  ├─ 4. 创建对话记录（用于前端展示）
  │
  ▼
AiderRunner.RunTask()
  │
  ├─ 5. BuildConfigFromModel() → 构建 AiderConfig
  │     ├─ 去掉 BaseURL 末尾的 /v1
  │     ├─ 设置 ProviderType 对应的环境变量
  │     └─ 格式化 model 名称（加 anthropic/ 前缀）
  │
  ├─ 6. buildArgs() → 构建 CLI 参数
  │     --model anthropic/glm-5
  │     --no-auto-commits
  │     --yes-always
  │     --chat-language Chinese
  │     --file src/api/user.go
  │     --message "## 角色设定\n...\n## 任务指令\n..."
  │
  ├─ 7. buildEnv() → 构建环境变量
  │     ANTHROPIC_API_KEY=sk-xxx
  │     ANTHROPIC_BASE_URL=https://...
  │
  ├─ 8. exec.CommandContext("aider", args...)
  │     cmd.Dir = workDir
  │     cmd.Env = env
  │     cmd.Stdout / cmd.Stderr → 捕获输出
  │
  ▼
  9. 检查 ExitCode
     ├─ 0 → 成功 → 保存输出 → completed
     └─ 非0 → 失败 → failTask()
```

### 3.2 文件结构

```
admin-go/app/mvp/internal/engine/
├── executor.go        # 任务执行器（修改：增加 Aider 分支）
├── aider_runner.go    # Aider CLI 封装（新增）
└── ...
```

### 3.3 角色→执行方式映射

| 角色 | 执行方式 | 说明 |
|------|---------|------|
| architect | ChatStream | 需求分析、方案设计，只需要对话 |
| **implementer** | **Aider CLI** | 真实代码编辑 |
| auditor | ChatStream | 代码审查，只需要分析文本 |
| coordinator | ChatStream | 协调总结，只需要对话 |

---

## 4. 核心代码详解

### 4.1 AiderRunner（aider_runner.go）

完整文件位置：`admin-go/app/mvp/internal/engine/aider_runner.go`

#### 4.1.1 数据结构

```go
// AiderConfig 运行配置
type AiderConfig struct {
    ModelCode    string        // "tc-code-latest" / "glm-5" / "hunyuan-t1"
    APIKey       string        // API Key
    BaseURL      string        // API Base URL（不含 /v1）
    ProviderType string        // "anthropic" / "openai_compatible"
    SystemPrompt string        // 系统提示词（角色定义）
    WorkDir      string        // 代码仓库根目录
    Files        []string      // 需要编辑的文件
    ReadFiles    []string      // 只读参考文件
    Message      string        // 任务指令
    MaxTokens    int           // 最大输出 token
    AutoCommit   bool          // 是否自动 git commit（默认 false）
    Timeout      time.Duration // 超时（默认 10 分钟）
}

// AiderResult 执行结果
type AiderResult struct {
    Output   string        // Aider 完整输出（stdout + stderr）
    ExitCode int           // 进程退出码（0=成功）
    Error    error         // Go 层面的错误
    Duration time.Duration // 执行耗时
}
```

#### 4.1.2 Run() 核心执行方法

```go
func (r *AiderRunner) Run(ctx context.Context, cfg *AiderConfig) *AiderResult {
    // 1. 构建命令行参数
    args := r.buildArgs(cfg)
    
    // 2. 构建环境变量
    env := r.buildEnv(cfg)
    
    // 3. 设置超时
    cmdCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
    defer cancel()
    
    // 4. 创建命令
    cmd := exec.CommandContext(cmdCtx, "aider", args...)
    cmd.Dir = cfg.WorkDir
    cmd.Env = append(cmd.Environ(), env...)
    
    // 5. 捕获输出
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // 6. 执行
    err := cmd.Run()
    
    // 7. 返回结果
    return &AiderResult{
        Output:   stdout.String() + stderr.String(),
        ExitCode: exitCode,
        Error:    err,
        Duration: time.Since(start),
    }
}
```

#### 4.1.3 buildArgs() 参数构建

```go
func (r *AiderRunner) buildArgs(cfg *AiderConfig) []string {
    args := []string{
        "--model", r.formatModel(cfg),  // 如 "anthropic/glm-5"
        "--no-auto-commits",            // 不自动 git commit
        "--no-show-model-warnings",     // 不显示模型警告
        "--no-pretty",                  // 纯文本输出
        "--no-stream",                  // 非流式（等待完整响应）
        "--yes-always",                 // 自动确认所有操作
        "--chat-language", "Chinese",   // 中文对话
    }
    
    // 任务指令
    if cfg.Message != "" {
        args = append(args, "--message", cfg.Message)
    }
    
    // 需要编辑的文件
    for _, f := range cfg.Files {
        args = append(args, "--file", f)
    }
    
    // 只读参考文件
    for _, f := range cfg.ReadFiles {
        args = append(args, "--read", f)
    }
    
    return args
}
```

**关键参数说明：**

| 参数 | 作用 | 为什么需要 |
|------|------|-----------|
| `--model anthropic/glm-5` | 指定 AI 模型 | 必须加 `anthropic/` 前缀 |
| `--no-auto-commits` | 不自动 git commit | 让 EasyMVP 控制提交时机 |
| `--no-show-model-warnings` | 跳过未知模型警告 | 腾讯云模型不在 Aider 内置列表 |
| `--yes-always` | 自动确认 | 非交互模式必须 |
| `--no-pretty` | 纯文本输出 | 便于程序解析 |
| `--no-stream` | 等待完整响应 | 简化输出处理 |
| `--message` | 单次指令 | 执行完自动退出 |
| `--file` | 可编辑文件 | 来自任务的 affected_resources |
| `--read` | 只读参考 | 提供上下文但不允许修改 |
| `--chat-language Chinese` | 中文交互 | 输出中文说明 |

#### 4.1.4 buildEnv() 环境变量

```go
func (r *AiderRunner) buildEnv(cfg *AiderConfig) []string {
    switch cfg.ProviderType {
    case "anthropic":
        return []string{
            "ANTHROPIC_API_KEY=" + cfg.APIKey,
            "ANTHROPIC_BASE_URL=" + cfg.BaseURL,
        }
    case "openai_compatible":
        return []string{
            "OPENAI_API_KEY=" + cfg.APIKey,
            "OPENAI_API_BASE=" + cfg.BaseURL,
        }
    }
}
```

#### 4.1.5 formatModel() 模型名格式化

```go
func (r *AiderRunner) formatModel(cfg *AiderConfig) string {
    switch cfg.ProviderType {
    case "anthropic":
        // glm-5 → anthropic/glm-5
        if !strings.HasPrefix(cfg.ModelCode, "anthropic/") {
            return "anthropic/" + cfg.ModelCode
        }
    case "openai_compatible":
        // gpt-4 → openai/gpt-4
        if !strings.HasPrefix(cfg.ModelCode, "openai/") {
            return "openai/" + cfg.ModelCode
        }
    }
    return cfg.ModelCode
}
```

#### 4.1.6 BuildConfigFromModel() 从数据库配置构建

```go
func (r *AiderRunner) BuildConfigFromModel(ctx context.Context, modelInfo *ModelInfo, workDir string) *AiderConfig {
    // 关键：去掉 BaseURL 末尾的 /v1
    // 数据库存的是 https://api.lkeap.cloud.tencent.com/coding/anthropic/v1
    // Aider 需要的是 https://api.lkeap.cloud.tencent.com/coding/anthropic
    baseURL := modelInfo.BaseURL
    baseURL = strings.TrimSuffix(baseURL, "/v1")
    baseURL = strings.TrimSuffix(baseURL, "/")
    
    return &AiderConfig{
        ModelCode:    modelInfo.ModelCode,
        APIKey:       modelInfo.APIKey,
        BaseURL:      baseURL,
        ProviderType: modelInfo.ProviderType,
        SystemPrompt: modelInfo.SystemPrompt,
        WorkDir:      workDir,
        AutoCommit:   false,
        Timeout:      10 * time.Minute,
    }
}
```

#### 4.1.7 RunTask() 完整任务执行

```go
func (r *AiderRunner) RunTask(
    ctx context.Context,
    projectID int64,
    taskID int64,
    modelInfo *ModelInfo,
    taskPrompt string,
    workDir string,
    files []string,
    readFiles []string,
) *AiderResult {
    cfg := r.BuildConfigFromModel(ctx, modelInfo, workDir)
    cfg.Message = taskPrompt
    cfg.Files = files
    cfg.ReadFiles = readFiles
    
    // 如果有 systemPrompt，拼到 message 前面
    if cfg.SystemPrompt != "" {
        cfg.Message = fmt.Sprintf(
            "## 角色设定\n%s\n\n## 任务指令\n%s",
            cfg.SystemPrompt,
            taskPrompt,
        )
    }
    
    return r.Run(ctx, cfg)
}
```

### 4.2 Executor 修改（executor.go）

#### 4.2.1 Execute() 分发逻辑

**修改前：** 所有角色都走 ChatStream。

**修改后：**
```go
func (e *Executor) Execute(ctx context.Context, projectID int64, taskID int64) {
    // ... 查询任务、获取模型配置 ...
    
    // implementer 角色 → Aider 代码编辑模式
    if roleType == "implementer" {
        e.executeWithAider(ctx, projectID, taskID, task, modelInfo)
        return
    }
    
    // 其他角色 → ChatStream 对话模式
    // ... 原有逻辑不变 ...
}
```

#### 4.2.2 executeWithAider() 完整实现

```go
func (e *Executor) executeWithAider(
    ctx context.Context,
    projectID int64,
    taskID int64,
    task gdb.Record,
    modelInfo *ModelInfo,
) {
    // 1. 获取项目工作目录
    project, err := g.DB().Model("mvp_project").Where("id", projectID).One()
    workDir := project["work_dir"].String()
    if workDir == "" {
        workDir = "/www/wwwroot/project/easymvp"  // 兜底默认值
    }
    
    // 2. 构建任务 prompt
    taskPrompt := e.buildTaskPrompt(task)
    // taskPrompt 格式：
    // ## 任务
    // {task.name}
    //
    // ## 任务描述
    // {task.description}
    //
    // ## 前置任务结果（供参考）
    // ### {dep_name}
    // {dep_result}
    
    // 3. 解析 affected_resources 为文件列表
    resources := parseResources(task["affected_resources"].String())
    // 例如：["src/api/user.go", "src/model/user.go"]
    
    // 4. 创建对话记录（用于前端展示 Aider 过程）
    conversationID, _ := e.ensureConversation(ctx, projectID, taskID, "implementer")
    g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
        "conversation_id": conversationID,
    })
    
    // 5. 保存用户指令消息
    userMsgID := int64(snowflake.Generate())
    g.DB().Model("mvp_message").Insert(g.Map{
        "id":              userMsgID,
        "conversation_id": conversationID,
        "role":            "user",
        "content":         taskPrompt,
        "status":          "completed",
        ...
    })
    
    // 6. 调用 Aider
    result := GetAiderRunner().RunTask(
        ctx, projectID, taskID,
        modelInfo,        // 模型配置
        taskPrompt,       // 任务指令
        workDir,          // 代码仓库路径
        resources,        // 需要编辑的文件
        nil,              // 只读参考文件（暂无）
    )
    
    // 7. 保存 Aider 输出为 AI 回复消息
    replyStatus := "completed"
    if result.Error != nil {
        replyStatus = "failed"
    }
    g.DB().Model("mvp_message").Insert(g.Map{
        "id":              replyID,
        "conversation_id": conversationID,
        "role":            "assistant",
        "content":         result.Output,  // Aider 的完整输出
        "model_id":        modelInfo.ModelID,
        "status":          replyStatus,
        ...
    })
    
    // 8. 判断结果
    if result.Error != nil {
        e.failTask(ctx, projectID, taskID,
            fmt.Sprintf("Aider执行失败(code=%d): %v", result.ExitCode, result.Error))
        return
    }
    
    // 9. 更新任务为完成
    g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
        "result":       result.Output,
        "status":       "completed",
        "completed_at": gtime.Now(),
    })
    
    // 10. 压缩上下文
    go GetCompressor().CompressTaskContext(context.Background(), projectID, taskID)
    
    // 11. 创建审计任务
    go e.createAuditTask(ctx, projectID, taskID, task)
    
    // 12. 通知调度器
    e.scheduler.OnTaskCompleted(projectID, taskID)
}
```

### 4.3 数据库变更

#### mvp_project 新增 work_dir 字段

```sql
ALTER TABLE mvp_project
ADD COLUMN work_dir varchar(500) DEFAULT NULL
COMMENT '项目代码工作目录（Aider执行路径）'
AFTER architect_model_id;
```

#### 受影响的代码层

| 层 | 文件 | 变更 |
|----|------|------|
| DAO | `dao/internal/mvp_project.go` | 新增 `WorkDir` 列定义 |
| API | `api/mvp/v1/project.go` | Create/Update 请求加 `WorkDir` |
| Model | `model/project.go` | 所有 Input/Output 加 `WorkDir` |
| Logic | `logic/project/project.go` | Create/Update 写入 `work_dir` |
| 前端 | `views/mvp/project/modules/form.vue` | 编辑表单加"代码工作目录"字段 |
| 前端 | `api/mvp/project/types.ts` | 类型加 `workDir` |

---

## 5. 模型→角色→Aider 配置链路

### 5.1 完整数据流

```
mvp_task（任务）
  ├─ role_type: "implementer"
  ├─ role_level: "pro"
  ├─ affected_resources: ["src/user.go"]
  │
  ▼ resolveTaskModel()
mvp_project_role（项目角色配置）
  ├─ role_type: "implementer"
  ├─ role_level: "pro"
  ├─ model_id: 315100000000000003
  ├─ system_prompt: "你是一位专业开发工程师..."
  │
  ▼ getModelInfo()
ai_model（模型）
  ├─ model_code: "glm-5"
  ├─ max_tokens: 4096
  ├─ role_prompt: "..."
  │
  ├─ plan_id → ai_plan
  │   ├─ api_key: "sk-sp-xxx"
  │   └─ api_secret: ""
  │
  └─ provider_id → ai_provider
      ├─ provider_type: "anthropic"
      └─ base_url: "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1"
  │
  ▼ BuildConfigFromModel()
AiderConfig
  ├─ ModelCode: "glm-5"
  ├─ APIKey: "sk-sp-xxx"
  ├─ BaseURL: "https://api.lkeap.cloud.tencent.com/coding/anthropic"  ← 去掉 /v1
  ├─ ProviderType: "anthropic"
  ├─ WorkDir: "/www/wwwroot/project/my-app"
  ├─ Files: ["src/user.go"]
  │
  ▼ buildArgs() + buildEnv()
实际执行命令:
  ANTHROPIC_API_KEY=sk-sp-xxx
  ANTHROPIC_BASE_URL=https://api.lkeap.cloud.tencent.com/coding/anthropic
  aider --model anthropic/glm-5
        --file src/user.go
        --message "## 角色设定\n你是一位...\n\n## 任务指令\n..."
        --no-auto-commits --yes-always ...
```

### 5.2 当前模型→角色→等级映射

| 角色类型 | 等级 | 模型 | model_code | 执行方式 |
|---------|------|------|-----------|---------|
| architect | max | Auto | tc-code-latest | ChatStream |
| architect | pro | Kimi-K2.5 | kimi-k2.5 | ChatStream |
| architect | lite | HY 2.0 Think | hunyuan-2.0-thinking | ChatStream |
| **implementer** | **max** | **GLM-5** | **glm-5** | **Aider** |
| **implementer** | **pro** | **GLM-5** | **glm-5** | **Aider** |
| **implementer** | **lite** | **Hunyuan-T1** | **hunyuan-t1** | **Aider** |
| auditor | max | hunyuan-turbos | hunyuan-turbos | ChatStream |
| auditor | pro | hunyuan-turbos | hunyuan-turbos | ChatStream |
| auditor | lite | HY 2.0 Instruct | hunyuan-2.0-instruct | ChatStream |
| coordinator | lite | MiniMax-M2.5 | minimax-m2.5 | ChatStream |

---

## 6. Aider CLI 参数完整参考

### 6.1 EasyMVP 使用的参数

```bash
aider \
  --model anthropic/glm-5 \         # 模型名（必须加 provider 前缀）
  --no-auto-commits \                # 不自动 git commit
  --no-show-model-warnings \         # 跳过未知模型警告
  --no-pretty \                      # 纯文本输出（非 Rich 格式）
  --no-stream \                      # 非流式（等待完整响应）
  --yes-always \                     # 自动确认所有操作
  --chat-language Chinese \          # 中文交互
  --file src/api/user.go \           # 可编辑文件（可多个）
  --file src/model/user.go \
  --read src/config/config.go \      # 只读参考文件（可多个）
  --message "任务指令..."             # 单次执行指令
```

### 6.2 环境变量

| Provider | 变量名 | 说明 |
|----------|--------|------|
| Anthropic | `ANTHROPIC_API_KEY` | API 密钥 |
| Anthropic | `ANTHROPIC_BASE_URL` | API 地址（不含 /v1） |
| OpenAI | `OPENAI_API_KEY` | API 密钥 |
| OpenAI | `OPENAI_API_BASE` | API 地址 |

### 6.3 可用但未使用的参数

| 参数 | 说明 | 未来可能使用场景 |
|------|------|--------------|
| `--auto-commits` | 自动 git commit | 每个任务自动提交 |
| `--commit-prompt` | 自定义 commit 消息格式 | 标准化提交信息 |
| `--lint-cmd` | 编辑后自动 lint | 代码质量保障 |
| `--test-cmd` | 编辑后自动测试 | 测试驱动开发 |
| `--architect` | 架构师模式（先规划后编辑） | 复杂任务双 AI 协作 |
| `--map-tokens` | Repo map token 数 | 大仓库优化 |
| `--subtree-only` | 只映射当前子目录 | 大仓库优化 |

---

## 7. 错误处理与故障排查

### 7.1 常见错误

| 错误 | 原因 | 解决 |
|------|------|------|
| `404 page not found` | BaseURL 包含 /v1 | 去掉 /v1，用 `/coding/anthropic` |
| `No models match` | 模型名不在内置列表 | 正常现象，加 `--no-show-model-warnings` |
| `exit code 1` | Aider 内部错误 | 查看 result.Output 中的错误信息 |
| `context deadline exceeded` | 超时 | 增大 Timeout 或简化任务 |
| `file not found` | affected_resources 中的文件不存在 | 检查 work_dir 和文件路径 |
| `not a git repository` | work_dir 未初始化 git | 在 work_dir 执行 `git init` |

### 7.2 日志追踪

AiderRunner 会在关键节点输出日志：

```
[AiderRunner] 启动: model=glm-5 workdir=/www/wwwroot/project/my-app files=[src/user.go]
[AiderRunner] 完成: 耗时=45.2s output_len=1234
```

或者失败时：
```
[AiderRunner] 退出 code=1 err=exit status 1
```

### 7.3 看门狗保护

即使 Aider 进程卡死，看门狗也会处理：

1. **超时保护**：AiderConfig.Timeout（默认 10 分钟），超时自动 kill 进程
2. **心跳检测**：Watchdog 每 2 分钟检查 running 任务是否有进展
3. **自动重试**：失败任务自动重启（最多 3 次）
4. **升级处理**：3 次仍失败 → 升级给架构师分析

---

## 8. 未来优化方向

### 8.1 流式输出支持

当前使用 `--no-stream`，Aider 完成后一次性返回输出。未来可以改为：

- 使用 `--stream` + 实时读取 stdout
- 逐行解析，写入 message_chunk
- 通过 SSEHub 实时推送给前端

### 8.2 Git 自动提交

当前 `--no-auto-commits`，未来可以：
- 每个任务完成后自动 `git commit`
- 使用 `--commit-prompt` 自定义提交信息
- 包含任务 ID 和描述

### 8.3 Lint/Test 集成

```bash
aider --lint-cmd "golint ./..." --auto-lint
aider --test-cmd "go test ./..." --auto-test
```

编辑后自动运行 lint 和测试，失败则让 AI 自动修复。

### 8.4 Repo Map 优化

对于大仓库，可以使用 `--map-tokens` 和 `--subtree-only` 减少上下文消耗。

### 8.5 架构师模式

Aider 的 `--architect` 模式使用两个模型：
- 架构师模型规划修改方案
- 编辑器模型执行具体编辑

可以用 EasyMVP 的 max 模型做架构师，lite 模型做编辑器。

### 8.6 只读参考文件

当前 `ReadFiles` 传 nil，未来可以：
- 自动传入依赖任务修改过的文件作为参考
- 传入项目的配置文件（如 go.mod、package.json）
- 传入相关的接口定义文件

---

## 9. 测试验证记录

### 9.1 连通性测试

| 模型 | 命令 | 结果 |
|------|------|------|
| tc-code-latest | `aider --model anthropic/tc-code-latest --message "say hello"` | 成功 |
| glm-5 | `aider --model anthropic/glm-5 --message "say hello"` | 成功 |
| hunyuan-t1 | `aider --model anthropic/hunyuan-t1 --message "say hello"` | 成功 |

### 9.2 代码编辑测试

**测试 1：GLM-5 修改函数**
```
输入：把 hello world 改成 hello from EasyMVP Aider，添加 add 函数
结果：✅ 文件成功修改，函数正确添加
输出：Applied edit to hello.go
Token: 668 sent, 105 received
```

**测试 2：Hunyuan-T1 追加代码**
```
输入：在 main 函数里添加一行 fmt.Println(add(3, 5))
结果：✅ 文件成功修改
输出：Applied edit to hello.go
Token: 751 sent, 189 received
```

### 9.3 编译测试

```bash
cd admin-go && go build ./app/mvp/...
# 编译成功，无错误
```
