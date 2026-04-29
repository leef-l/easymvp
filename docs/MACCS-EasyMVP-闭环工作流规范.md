# MACCS — EasyMVP 闭环工作流规范

> **版本**: v2.0.0  
> **日期**: 2026-04-29  
> **依赖**: `MACCS-架构总纲-v2.md`, `MACCS-中央大脑智能化编排规范.md`  
> **范围**: EasyMVP 全链路：需求 → 方案 → 审核 → 修正 → 确认 → 执行 → 验收

---

## 1. 愿景：一键交付

用户输入一句话需求，系统自动完成从需求理解到项目交付的全流程，中间无需人工干预，除非系统主动请求确认。

```
用户: "帮我做一个贪吃蛇游戏，有计分、有难度等级、画面好看一点"

系统:
  ├─► 需求分析（Central）
  ├─► 方案设计（Central）
  ├─► 方案审核（Verifier）→ 发现问题 → 自动修正 → 重新审核 → 通过
  ├─► 任务分解（Central）
  │     ├─► 前端页面（Code + Browser 验证）
  │     ├─► 游戏逻辑（Code）
  │     ├─► 计分系统（Code）
  │     └─► 难度等级（Code）
  ├─► 并行执行（Code Brain 集群）
  ├─► 整合测试（Verifier + Browser）
  ├─► 验收测试（Verifier + Browser + Fault）
  └─► 交付用户

用户收到:
  - 可运行的贪吃蛇游戏
  - 项目代码
  - 测试报告
  - 部署说明
```

---

## 2. 七阶段闭环

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            EasyMVP 七阶段闭环                                │
│                                                                              │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐      │
│  │ Phase 1 │──►│ Phase 2 │──►│ Phase 3 │──►│ Phase 4 │──►│ Phase 5 │      │
│  │ 需求理解│   │ 方案设计│   │ 方案审核│   │ 任务执行│   │ 验收测试│      │
│  └─────────┘   └─────────┘   └────┬────┘   └─────────┘   └────┬────┘      │
│       │                            │                           │           │
│       │                            ▼                           ▼           │
│       │                       ┌─────────┐                 ┌─────────┐      │
│       │                       │ 修正循环│                 │ 修复循环│      │
│       │                       │ (自动)  │                 │ (自动)  │      │
│       │                       └────┬────┘                 └────┬────┘      │
│       │                            │                           │           │
│       └────────────────────────────┴───────────────────────────┘           │
│                                    │                                       │
│                                    ▼                                       │
│                              ┌─────────┐                                   │
│                              │ Phase 6 │                                   │
│                              │ 项目交付│                                   │
│                              └────┬────┘                                   │
│                                   │                                        │
│                                   ▼                                        │
│                              ┌─────────┐                                   │
│                              │ Phase 7 │                                   │
│                              │ 复盘学习│                                   │
│                              └─────────┘                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Phase 1: 需求理解（Requirement Analysis）

### 3.1 输入

- 用户自然语言需求
- 可选：参考项目、设计稿、API 文档

### 3.2 处理流程

```
用户输入
  │
  ▼
Central Brain 解析需求
  ├─► 提取核心功能点
  ├─► 识别非功能需求（性能、安全、UI）
  ├─► 识别隐含需求（"好看一点"→需要 CSS/动画）
  ├─► 识别技术栈偏好
  └─► 生成结构化需求文档
  
  │
  ▼
用户确认（系统呈现需求文档，用户可修改）
  │
  ▼
需求锁定 → 进入 Phase 2
```

### 3.3 输出

```go
type RequirementDoc struct {
    ProjectID       string            `json:"project_id"`
    RawInput        string            `json:"raw_input"`
    
    // 功能需求
    CoreFeatures    []Feature         `json:"core_features"`
    
    // 非功能需求
    NonFunctional   NonFunctionalReq  `json:"non_functional"`
    
    // 技术栈
    TechStack       TechStack         `json:"tech_stack"`
    
    // 约束
    Constraints     []string          `json:"constraints"`
    
    // 用户确认
    UserConfirmed   bool              `json:"user_confirmed"`
    ConfirmedAt     *time.Time        `json:"confirmed_at,omitempty"`
}

type Feature struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Priority    string   `json:"priority"` // must/should/could
    Acceptance  []string `json:"acceptance"` // 验收标准
}

type NonFunctionalReq struct {
    Performance string `json:"performance,omitempty"`
    Security    string `json:"security,omitempty"`
    UIQuality   string `json:"ui_quality,omitempty"`
    BrowserSupport []string `json:"browser_support,omitempty"`
}

type TechStack struct {
    Frontend    string `json:"frontend,omitempty"`
    Backend     string `json:"backend,omitempty"`
    Database    string `json:"database,omitempty"`
    Language    string `json:"language,omitempty"`
}
```

### 3.4 关键能力

- **隐含需求识别**: 用户说"好看一点"，系统要自动推断出需要 CSS 样式、动画效果、响应式布局
- **技术栈推荐**: 基于项目类型推荐最佳技术栈（小游戏→纯前端；后台系统→React+Go）
- **需求澄清**: 对模糊需求主动追问（"难度等级具体要几级？"）

---

## 4. Phase 2: 方案设计（Solution Design）

### 4.1 设计内容

Central Brain 基于需求文档，设计完整的技术方案：

```
方案设计输出：
  ├─► 系统架构图（文字描述 + Mermaid 图）
  ├─► 模块划分与职责
  ├─► 数据模型设计
  ├─► API 接口设计（如需要后端）
  ├─► 页面结构/组件树
  ├─► 关键算法/逻辑描述
  ├─► 文件目录结构
  └─► 任务分解清单（初步）
```

### 4.2 输出格式

```go
type SolutionDesign struct {
    DesignID    string `json:"design_id"`
    ProjectID   string `json:"project_id"`
    Version     int    `json:"version"`
    
    // 架构
    Architecture string `json:"architecture"` // Markdown 描述
    
    // 模块
    Modules     []Module `json:"modules"`
    
    // 数据模型
    DataModels  []DataModel `json:"data_models,omitempty"`
    
    // 页面/组件
    Pages       []PageDesign `json:"pages,omitempty"`
    
    // 任务分解（初步，Phase 3 审核后细化）
    TaskDrafts  []TaskDraft `json:"task_drafts"`
    
    // 用户确认
    UserConfirmed bool `json:"user_confirmed"`
}

type Module struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Responsibilities []string `json:"responsibilities"`
    Files       []string `json:"files"` // 预计要创建的文件
}

type DataModel struct {
    Name   string   `json:"name"`
    Fields []Field  `json:"fields"`
}

type Field struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Required bool   `json:"required"`
}

type PageDesign struct {
    Name       string   `json:"name"`
    Route      string   `json:"route,omitempty"`
    Components []string `json:"components"`
    Features   []string `json:"features"`
}

type TaskDraft struct {
    Name        string     `json:"name"`
    Description string     `json:"description"`
    Kind        string     `json:"kind"` // code/browser/verifier
    EstimatedHours float64 `json:"estimated_hours"`
}
```

### 4.3 用户确认

系统呈现方案给用户，用户可以：
- ✅ 确认通过 → 进入 Phase 3
- 📝 提出修改意见 → Central Brain 修改 → 重新呈现
- ❌ 完全不接受 → 回到 Phase 1 重新理解需求

---

## 5. Phase 3: 方案审核 — 自动化闭环（核心创新）

### 5.1 审核员角色

**不是人工审核，是 Verifier Brain 自动审核。**

```
方案设计
  │
  ▼
Verifier Brain 审核
  ├─► 架构合理性检查
  ├─► 技术选型评估
  ├─► 可扩展性分析
  ├─► 安全风险评估
  ├─► 与需求匹配度检查
  └─► 生成审核报告
  
  │
  ├─► [通过] ──► 进入 Phase 4
  │
  └─► [不通过] ──► 生成问题清单 ──► 自动修正 ──► 重新审核
                            │
                            └──► 最多 5 轮循环
```

### 5.2 审核维度

| 维度 | 检查内容 | 审核 Brain |
|------|----------|-----------|
| 架构合理性 | 模块划分是否清晰、职责是否单一 | Verifier |
| 技术选型 | 技术栈是否匹配项目类型 | Verifier |
| 安全 | 是否有明显的安全漏洞设计 | Verifier |
| 性能 | 是否考虑了性能瓶颈 | Verifier |
| 需求匹配 | 是否覆盖了所有需求 | Verifier |
| 可测试性 | 设计是否便于测试 | Verifier |
| 可维护性 | 代码结构是否清晰 | Verifier |

### 5.3 审核报告

```go
type DesignReviewReport struct {
    ReportID    string `json:"report_id"`
    DesignID    string `json:"design_id"`
    
    // 总体评价
    Passed      bool   `json:"passed"`
    Score       int    `json:"score"` // 0-100
    
    // 各维度评分
    Dimensions  []DimensionScore `json:"dimensions"`
    
    // 问题清单
    Issues      []DesignIssue `json:"issues"`
    
    // 建议
    Suggestions []string `json:"suggestions"`
}

type DimensionScore struct {
    Name    string `json:"name"`
    Score   int    `json:"score"` // 0-100
    Weight  int    `json:"weight"`
    Comment string `json:"comment"`
}

type DesignIssue struct {
    ID          string `json:"id"`
    Severity    string `json:"severity"` // critical/major/minor
    Category    string `json:"category"`
    Description string `json:"description"`
    Location    string `json:"location,omitempty"` // 设计文档中的位置
    SuggestedFix string `json:"suggested_fix"`
}
```

### 5.4 自动修正循环

```go
func (c *ReviewLoopController) RunDesignReview(
    ctx context.Context, 
    design *SolutionDesign,
) (*SolutionDesign, *DesignReviewReport, error) {
    
    for round := 1; round <= c.MaxRounds; round++ {
        // 1. 提交审核
        report := c.verifier.ReviewDesign(design)
        
        // 2. 记录审核历史
        design.ReviewHistory = append(design.ReviewHistory, ReviewCycle{
            Round: round,
            Report: report,
        })
        
        // 3. 检查是否通过
        if report.Passed {
            return design, report, nil
        }
        
        // 4. 检查是否收敛（问题数不再显著减少）
        if round > 1 && !c.isConverging(design.ReviewHistory) {
            return design, report, fmt.Errorf("审核无法收敛，剩余 %d 个问题", len(report.Issues))
        }
        
        // 5. 生成修正任务
        fixTasks := c.generateFixTasks(report.Issues)
        
        // 6. 执行修正（Central Brain 自己修改设计文档）
        fixedDesign, err := c.applyFixes(design, fixTasks)
        if err != nil {
            return design, report, err
        }
        
        design = fixedDesign
        design.Version++
    }
    
    return design, nil, fmt.Errorf("达到最大审核轮数 %d", c.MaxRounds)
}
```

### 5.5 收敛条件

- **硬性通过**: Score >= 85 且没有 critical issue
- **软性收敛**: 连续两轮问题数减少 < 20%，且剩余问题都是 minor
- **失败**: 达到最大轮数（5 轮）仍未通过

### 5.6 人工介入点

系统自动循环，但以下情况需要人工确认：
- 审核 3 轮仍未通过 → 提示用户"方案可能有根本性缺陷，建议重新设计"
- 发现 critical security issue → 立即停止，通知用户
- 系统不确定如何修正 → 请求用户指导

---

## 6. Phase 4: 任务执行（Task Execution）

### 6.1 任务细化

方案审核通过后，Central Brain 基于最终方案生成详细的 TaskPlan：

```go
func (c *CentralBrain) GenerateTaskPlan(design *SolutionDesign) *kernel.TaskPlan {
    plan := &kernel.TaskPlan{
        ProjectID: design.ProjectID,
        Goal: design.Architecture,
    }
    
    // 为每个模块创建任务
    for _, module := range design.Modules {
        task := kernel.SubTask{
            Name: module.Name,
            Kind: agent.KindCode,
            Instruction: fmt.Sprintf("Implement module %s with the following responsibilities:\n%s",
                module.Name, strings.Join(module.Responsibilities, "\n")),
            VerificationCriteria: []string{
                fmt.Sprintf("All files for %s are created", module.Name),
                fmt.Sprintf("Module compiles without errors"),
                fmt.Sprintf("Unit tests pass"),
            },
        }
        plan.SubTasks = append(plan.SubTasks, task)
    }
    
    // 构建依赖关系
    plan.Dependencies = c.analyzeDependencies(plan.SubTasks)
    
    // 拓扑分层
    plan.ParallelLayers = computeTopologicalLayers(plan.Dependencies)
    
    // 动态预算分配
    plan.Budget = c.allocateBudget(plan.SubTasks)
    
    return plan
}
```

### 6.2 执行调度

```
TaskPlan
  │
  ▼
Topological Layers:
  Layer 0: [Task A, Task B, Task C] ──► 并行启动 3 个 Code Brain runs
  Layer 1: [Task D] ──────────────────► 等待 A,B,C 完成后启动（依赖 A,B,C）
  Layer 2: [Task E, Task F] ──────────► 等待 D 完成后并行启动
  Layer 3: [Task G] ──────────────────► 整合测试
```

### 6.3 进度同步

**实时进度推送到 EasyMVP**：

```go
// EasyMVP 订阅 Brain 事件
events.Subscribe(projectID, func(evt Event) {
    switch evt.Type {
    case "task.started":
        updateTaskStatus(evt.TaskID, "running")
        
    case "task.progress":
        updateTaskProgress(evt.TaskID, evt.Progress)
        
    case "task.completed":
        updateTaskStatus(evt.TaskID, "completed")
        checkLayerCompletion(evt.LayerIndex)
        
    case "task.failed":
        updateTaskStatus(evt.TaskID, "failed")
        // Central Brain 自动处理失败
        handleTaskFailure(evt.TaskID, evt.Error)
        
    case "layer.completed":
        startNextLayer(evt.LayerIndex + 1)
        
    case "plan.adjusted":
        notifyUserPlanChanged(evt.Reason)
    }
})
```

### 6.4 动态调整示例

**场景 1: 任务进度停滞**
```
Task B: 写游戏逻辑
  Turn 10/50: "正在实现蛇的移动……"
  Turn 20/50: "正在实现蛇的移动……"（没有实质进展）
  
Central 检测到:
  - 10 轮没有新文件创建
  - 输出内容重复度高
  
决策:
  - 发送 InterruptSignal 给 Task B
  - 分析原因：可能是需求 unclear
  - 调整指令：拆分为更小的子任务
    - Sub-task B1: 实现蛇的基本移动
    - Sub-task B2: 实现方向控制
    - Sub-task B3: 实现碰撞检测
```

**场景 2: 并发冲突**
```
Task A: 创建 index.html（包含全局样式）
Task B: 创建 style.css
  
Task A 和 Task B 同时修改了颜色相关的 CSS
  
Central 检测到冲突:
  - 两个任务的产出物都涉及颜色定义
  
决策:
  - 发送 InterruptSignal 给 Task B
  - 添加依赖：Task B 依赖 Task A
  - 调整指令：Task B 使用 Task A 定义的颜色变量
```

---

## 7. Phase 5: 验收测试（Acceptance Testing）

### 7.1 多层验收

不是一次性的"看看能不能跑"，而是系统化的多层验收：

```
Layer 1: 单元测试（Verifier Brain）
  ├─► 每个模块的单元测试是否通过
  ├─► 代码覆盖率是否达标（>= 80%）
  └─► 边界条件是否覆盖

Layer 2: 集成测试（Verifier Brain）
  ├─► 模块间接口是否正确
  ├─► 数据流是否正确
  └─► 错误处理是否完善

Layer 3: 端到端测试（Browser Brain）
  ├─► 在真实浏览器中运行
  ├─► UI 是否符合设计
  ├─► 交互是否流畅
  └─► 截图对比（如有设计稿）

Layer 4: 安全测试（Verifier Brain）
  ├─► 输入验证
  ├─► XSS 防护
  └─► 敏感信息泄露检查

Layer 5: 性能测试（Verifier Brain）
  ├─► 页面加载时间
  ├─► 内存占用
  └─► 响应速度
```

### 7.2 验收标准

```go
type AcceptanceCriteria struct {
    Layers []AcceptanceLayer `json:"layers"`
}

type AcceptanceLayer struct {
    Name      string   `json:"name"`
    Required  bool     `json:"required"`
    Criteria  []Criterion `json:"criteria"`
}

// 每个项目类型有默认的验收标准
type ProjectType string
const (
    ProjectWebApp      ProjectType = "web_app"
    ProjectGame        ProjectType = "game"
    ProjectAPI         ProjectType = "api"
    ProjectDataPipeline ProjectType = "data_pipeline"
)

func DefaultAcceptanceCriteria(projectType ProjectType) AcceptanceCriteria {
    switch projectType {
    case ProjectGame:
        return AcceptanceCriteria{
            Layers: []AcceptanceLayer{
                {
                    Name: "功能验收",
                    Required: true,
                    Criteria: []Criterion{
                        {Name: "核心玩法可运行", Weight: 40},
                        {Name: "计分系统正确", Weight: 20},
                        {Name: "难度系统工作", Weight: 20},
                        {Name: "无崩溃/卡死", Weight: 20},
                    },
                },
                {
                    Name: "UI 验收",
                    Required: true,
                    Criteria: []Criterion{
                        {Name: "视觉效果符合预期", Weight: 50},
                        {Name: "响应式布局", Weight: 30},
                        {Name: "动画流畅", Weight: 20},
                    },
                },
                {
                    Name: "Browser 测试",
                    Required: true,
                    Criteria: []Criterion{
                        {Name: "Chrome 正常", Weight: 40},
                        {Name: "Firefox 正常", Weight: 30},
                        {Name: "Safari 正常", Weight: 30},
                    },
                },
            },
        }
    // ... 其他项目类型
    }
}
```

### 7.3 验收失败处理

```
验收测试
  │
  ├─► [全部通过] ──► Phase 6 交付
  │
  └─► [有未通过项]
        │
        ▼
    生成修复任务清单
        │
        ▼
    Central Brain 评估：
      ├─► 小问题（如样式调整）→ 直接分配给 Code Brain
      ├─► 功能缺陷 → 分配给 Code Brain + Verifier 验证
      └─► 架构问题 → 回到 Phase 2 重新设计
        │
        ▼
    执行修复 → 重新验收（最多 3 轮）
```

---

## 8. Phase 6: 项目交付（Delivery）

### 8.1 交付物

```go
type ProjectDelivery struct {
    ProjectID     string `json:"project_id"`
    DeliveredAt   time.Time `json:"delivered_at"`
    
    // 代码
    WorkspacePath string `json:"workspace_path"`
    
    // 文档
    Documents     DeliveryDocs `json:"documents"`
    
    // 测试报告
    TestReport    TestReport `json:"test_report"`
    
    // 项目统计
    Statistics    ProjectStats `json:"statistics"`
}

type DeliveryDocs struct {
    README        string `json:"readme"`        // 项目说明
    Architecture  string `json:"architecture"`  // 架构文档
    API_Docs      string `json:"api_docs,omitempty"` // API 文档（如需要）
    Deployment    string `json:"deployment"`    // 部署说明
}

type TestReport struct {
    UnitTests     TestResult `json:"unit_tests"`
    Integration   TestResult `json:"integration"`
    E2E           TestResult `json:"e2e"`
    Security      TestResult `json:"security"`
    Performance   TestResult `json:"performance"`
}

type TestResult struct {
    Total     int `json:"total"`
    Passed    int `json:"passed"`
    Failed    int `json:"failed"`
    Skipped   int `json:"skipped"`
    Coverage  float64 `json:"coverage,omitempty"`
}

type ProjectStats struct {
    TotalTasks     int `json:"total_tasks"`
    CompletedTasks int `json:"completed_tasks"`
    FailedTasks    int `json:"failed_tasks"`
    RetriedTasks   int `json:"retried_tasks"`
    
    TotalTurns     int `json:"total_turns"`
    TotalTokens    int `json:"total_tokens"`
    TotalCostUSD   float64 `json:"total_cost_usd"`
    
    Duration       time.Duration `json:"duration"`
    ReviewRounds   int `json:"review_rounds"`
    
    BrainsUsed     []string `json:"brains_used"`
}
```

### 8.2 用户验收

用户收到交付物后可以：
- ✅ 确认验收 → 项目完成 → 进入 Phase 7 复盘
- 📝 提出修改意见 → 生成修复任务 → 回到 Phase 4/5
- ❌ 拒绝验收 → 分析原因 → 可能回到 Phase 1/2

---

## 9. Phase 7: 复盘学习（Retrospective Learning）

### 9.1 自动复盘

项目完成后，系统自动进行复盘：

```
复盘内容：
  1. 计划 vs 实际
     - 预估时间 vs 实际时间
     - 预估 turn 数 vs 实际 turn 数
     - 预估复杂度 vs 实际复杂度
     
  2. 成功因素
     - 哪些决策是对的？
     - 哪个 brain 表现最好？
     
  3. 失败教训
     - 哪些任务失败了？为什么？
     - 哪些审核轮数过多？
     - 哪些中断是不必要的？
     
  4. 模式提取
     - 这种项目类型的最佳实践
     - 常用模块的模板化
     - 常见错误的预防策略
```

### 9.2 学习数据写入

```go
func (e *LearningEngine) RecordProjectOutcome(project *ProjectDelivery) {
    // L1: 更新各 brain 的能力画像
    for _, brain := range project.Statistics.BrainsUsed {
        e.UpdateBrainProfile(brain, project)
    }
    
    // L2: 记录 DAG 执行序列
    e.RecordSequence(project.TaskSequence)
    
    // L3: 提取项目级模式
    e.ExtractProjectPattern(project)
    
    // 持久化
    e.Save()
}
```

### 9.3 经验复用

下次遇到类似项目时：
- 自动推荐相似项目的方案作为参考
- 自动调整预算预估（基于历史数据）
- 自动预加载常用模块模板

---

## 10. EasyMVP 与 Brain 的接口

### 10.1 新 API 设计

```go
// EasyMVP 调用 Brain 的接口

// 创建项目级 Central Brain Session
POST /v1/projects
{
    "name": "snake-game",
    "description": "帮我做一个贪吃蛇游戏",
    "category": "game"
}
→ Response: { "project_id": "proj_xxx", "central_session_id": "sess_xxx" }

// 提交需求（触发 Phase 1）
POST /v1/projects/{id}/requirements
{
    "raw_input": "帮我做一个贪吃蛇游戏，有计分、有难度等级、画面好看一点"
}
→ Response: { "requirement_doc": {...}, "status": "analyzing" }

// 确认需求
POST /v1/projects/{id}/requirements/confirm
→ Response: { "status": "designing" }

// 获取方案（Phase 2 输出）
GET /v1/projects/{id}/design
→ Response: { "solution_design": {...}, "status": "reviewing" }

// 确认方案（触发 Phase 3 审核）
POST /v1/projects/{id}/design/confirm
→ Response: { "review_report": {...}, "status": "reviewing" }

// 获取项目进度（实时）
GET /v1/projects/{id}/progress
→ Response: { "progress": {...}, "active_runs": [...] }

// SSE 订阅项目事件
GET /v1/projects/{id}/events
→ SSE Stream: { "type": "task.completed", "task_id": "...", ... }

// 用户介入（审核不通过时人工决策）
POST /v1/projects/{id}/intervene
{
    "action": "override_review", // 或 "restart_phase", "abort"
    "reason": "我觉得方案没问题"
}
```

### 10.2 数据流

```
EasyMVP                          Brain
  │                               │
  ├─► POST /v1/projects ─────────►├─► 创建 Project + Central Session
  │                               │
  ├─► POST /v1/projects/{id}/requirements ──►├─► Phase 1: 需求分析
  │                               │
  │◄──────────────────────────────├─► 返回 RequirementDoc
  │                               │
  ├─► POST .../requirements/confirm ──►├─► 进入 Phase 2
  │                               │
  │◄──────────────────────────────├─► 返回 SolutionDesign
  │                               │
  ├─► POST .../design/confirm ────►├─► 进入 Phase 3（自动审核）
  │                               │
  │◄──────────────────────────────├─► SSE: 审核进度
  │                               │
  │◄──────────────────────────────├─► 审核通过 → 进入 Phase 4
  │                               │
  ├─► GET .../events ─────────────►├─► SSE: 实时进度推送
  │                               │
  │◄──────────────────────────────├─► Phase 4 完成 → Phase 5
  │                               │
  │◄──────────────────────────────├─► Phase 5 完成 → Phase 6
  │                               │
  ├─► GET .../delivery ───────────►├─► 返回 ProjectDelivery
```

---

## 11. 与当前系统的差异

### 11.1 当前 EasyMVP 的问题

| 问题 | 当前行为 | 目标行为 |
|------|----------|----------|
| 无项目级 Central Session | 每个 task 独立启动 central run | 整个项目一个 Central Session |
| 无方案审核闭环 | 方案直接执行，无审核 | 自动审核 → 修正 → 再审核 |
| 无进度感知 | 只 know 任务成功/失败 | 实时知道每个任务的详细进度 |
| 无动态调整 | 计划一旦确定不变 | 根据反馈实时调整计划 |
| 无并发冲突处理 | 同层任务盲目并行 | 冲突检测 + 仲裁 |
| 无验收测试 | 任务完成即结束 | 多层系统化验收 |
| 无学习复盘 | 项目结束即忘记 | 自动复盘，经验复用 |

### 11.2 迁移路径

```
Step 1: 保留现有 task scheduler，新增 Project Session 层
Step 2: Central Brain 升级支持 TaskPlan / InterruptSignal
Step 3: EasyMVP 接入新的 Project API
Step 4: 逐步替换旧调度逻辑
Step 5: 添加验收测试层
Step 6: 添加学习复盘层
```

---

## 12. 关键成功指标（KPI）

| 指标 | 当前 | 目标 |
|------|------|------|
| 项目成功率 | ~0%（全部失败） | >= 90% |
| 方案审核通过率（首轮） | N/A | >= 60% |
| 平均审核轮数 | N/A | <= 2 |
| 任务执行成功率 | ~0% | >= 95% |
| 验收测试通过率 | N/A | >= 90% |
| 用户满意度 | N/A | >= 4.5/5 |
| 项目平均耗时 | 30min+（全部失败） | <= 20min（成功交付） |
| 成本效率 | N/A | <= $0.5/项目 |

---

*本文档定义了 EasyMVP 闭环工作流的完整规范。所有实现必须遵循本文档。*
