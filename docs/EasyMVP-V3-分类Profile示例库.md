# EasyMVP V3 分类Profile示例库

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3分类Profile结构与版本化规范](./EasyMVP-V3分类Profile结构与版本化规范.md)
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 目标：给出 V3 第一版可直接落地的分类 `CategoryProfile` 示例，作为实现和评审时的标准样本库。

## 1. 设计结论

`CategoryProfile` 规范文档解决“结构怎么定义”，本专题解决“第一版到底长什么样”。

V3 第一版建议正式支持 3 类 profile：

1. `web`
2. `game`
3. `video_editing`

## 2. 通用基线

所有分类示例都应继承以下通用基线：

1. 必须走 `PlanDraft -> PlanReviewResult -> CompiledPlan`
2. 必须输出结构化 `delivery_contract`
3. 必须输出结构化 `verification_contract`
4. 最终验收都要区分 `functional_passed` 和 `production_passed`

## 3. Web Profile 示例

### 3.1 示例摘要

```json
{
  "project_category": "web",
  "planning_policy": {
    "plan_style": "web_app",
    "task_split_limits": {
      "max_resources_per_task": 8,
      "must_split_when_blocking_review": true
    }
  },
  "role_policy": {
    "default_roles_by_phase": {
      "Review": "architect_reviewer",
      "Compile": "plan_compiler",
      "Execute": "code_executor",
      "Acceptance": "browser_verifier"
    }
  },
  "verification_policy": {
    "browser_required": true,
    "ci_required": true,
    "runtime_proof_required": false
  },
  "acceptance_policy": {
    "required_surfaces": [
      "user_frontend",
      "admin_backend",
      "api_backend"
    ]
  }
}
```

### 3.2 关键特点

1. 用户端和后台端都属于 blocking surface
2. 浏览器验证和 CI 证据都属于核心要求
3. API 一致性属于 production gate 关键项

## 4. Game Profile 示例

### 4.1 示例摘要

```json
{
  "project_category": "game",
  "planning_policy": {
    "plan_style": "game_runtime",
    "task_split_limits": {
      "max_resources_per_task": 6,
      "must_split_when_blocking_review": true
    }
  },
  "verification_policy": {
    "browser_required": false,
    "ci_required": true,
    "runtime_proof_required": true
  },
  "acceptance_policy": {
    "required_surfaces": [
      "game_runtime",
      "admin_backend"
    ],
    "required_journeys": [
      "boot",
      "core_loop",
      "pause_resume",
      "settlement_restart"
    ]
  }
}
```

### 4.2 关键特点

1. `game_runtime` 是最高优先级 surface
2. 核心玩法循环必须覆盖
3. 引擎运行证据是 production gate 的必选项

## 5. Video Editing Profile 示例

### 5.1 示例摘要

```json
{
  "project_category": "video_editing",
  "planning_policy": {
    "plan_style": "editor_pipeline"
  },
  "verification_policy": {
    "browser_required": false,
    "ci_required": true,
    "runtime_proof_required": true
  },
  "acceptance_policy": {
    "required_surfaces": [
      "editor_runtime",
      "admin_backend"
    ],
    "required_journeys": [
      "import",
      "edit",
      "export"
    ],
    "required_evidence": [
      "runtime_log",
      "export_output",
      "verification_report"
    ]
  }
}
```

### 5.2 关键特点

1. 导入/编辑/导出是主路径
2. 导出产物不只是附件，而是 blocking evidence
3. `editor_runtime` 验证失败时不能进入 `production_passed`

## 6. 示例库使用规则

### 6.1 不作为硬编码常量

示例库是标准样本，不意味着系统里必须把所有规则写死在代码里。

### 6.2 作为评审基线

实现新分类前，必须先提供至少一个 profile 示例。

### 6.3 作为回归样例

分类策略改动后，必须用示例库做回归对比。

## 7. 与其他文档的关系

本专题主要服务：

1. `CategoryProfile` 实现
2. `RoleResolver` 命中样例
3. `AcceptanceProfile` 生成样例
4. 工作台页面的示例数据

## 8. 后续细分专题

本专题后续继续拆：

1. 示例库的 JSON 文件组织方式
2. 示例库的回归测试基线
3. 新分类提案模板
