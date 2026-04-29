# EasyMVP V3 easymvp-brain 合同 JSON Schema 与 DTO 映射终稿

> 更新时间：2026-04-24
> 上游文档：[钱学森总纲设计/easymvp-brain-职责与边界定义](./钱学森总纲设计/easymvp-brain-职责与边界定义.md)
> 关联文档：[钱学森总纲设计/easymvp-brain-输入输出契约](./钱学森总纲设计/easymvp-brain-输入输出契约.md)
> 关联文档：[EasyMVP-V3计划数据模型与表结构设计](./EasyMVP-V3计划数据模型与表结构设计.md)
> 目标：把 `easymvp-brain` 六类领域合同落成可直接实现的 JSON Schema 基线，并明确与 Go DTO / SQLite 字段的映射关系。

## 1. 设计结论

`easymvp-brain` 的所有合同都必须满足：

1. 输入可跨进程序列化
2. 输出可直接落库
3. 输出可被工作台解释
4. 输出不依赖本地隐式上下文

统一约束：

1. 顶层返回采用 envelope
2. `result_json` 承载领域对象正文
3. `source_refs` 负责版本回溯
4. `decision_summary` 负责工作台与审计摘要
5. 所有运行时、浏览器、验证、故障信息都必须先被 EasyMVP 归一化后才能进入领域合同

## 1.1 与 `brain-v3` 最新工具协议的对齐要求

合同设计必须承认 `brain-v3` 当前的工具协议与能力边界：

1. 内置脑通过结构化 `tools/list` 暴露 `name / description / input_schema / output_schema`
2. 内置脑通过结构化 `tools/call` 返回 `output`，兼容层 `content[]` 不作为领域落库主载荷
3. 高风险工具可能显式返回 `unsupported` / `denied`
4. `easymvp-brain` 输入合同只接受 EasyMVP 归一化后的 `success / failure / unsupported_or_denied`
5. 合同 Schema 不得要求调用方传入具体内置脑工具名或原始工具 payload

## 2. 统一 Envelope Schema

```json
{
  "type": "object",
  "required": [
    "schema_version",
    "result_kind",
    "result_version",
    "source_refs",
    "decision_summary",
    "result_json"
  ],
  "properties": {
    "schema_version": {"type": "integer", "minimum": 1},
    "result_kind": {"type": "string", "minLength": 1},
    "result_version": {"type": "integer", "minimum": 1},
    "source_refs": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["kind", "id", "version"],
        "properties": {
          "kind": {"type": "string"},
          "id": {"type": "string"},
          "version": {"type": "integer", "minimum": 1}
        }
      }
    },
    "decision_summary": {"type": "string", "minLength": 1},
    "trace_id": {"type": "string"},
    "deployment_mode": {
      "type": "string",
      "enum": ["local-sidecar", "remote-service"]
    },
    "brain_endpoint": {"type": "string"},
    "normalized_status": {
      "type": "string",
      "enum": ["success", "failure", "unsupported_or_denied"]
    },
    "result_json": {"type": "object"}
  }
}
```

建议 Go DTO：

```go
type BrainContractEnvelope struct {
    SchemaVersion  int              `json:"schema_version"`
    ResultKind     string           `json:"result_kind"`
    ResultVersion  int              `json:"result_version"`
    SourceRefs     []BrainSourceRef `json:"source_refs"`
    DecisionSummary string          `json:"decision_summary"`
    TraceID        string           `json:"trace_id,omitempty"`
    DeploymentMode string           `json:"deployment_mode,omitempty"`
    BrainEndpoint  string           `json:"brain_endpoint,omitempty"`
    NormalizedStatus string         `json:"normalized_status,omitempty"`
    ResultJSON     json.RawMessage  `json:"result_json"`
}

type BrainSourceRef struct {
    Kind    string `json:"kind"`
    ID      string `json:"id"`
    Version int    `json:"version"`
}
```

补充约束：

1. `result_json` 是领域合同的 canonical payload
2. 如果上游执行链只返回兼容层 `content[]`，必须先由 EasyMVP 适配层转成结构化 `result_json`
3. `normalized_status=unsupported_or_denied` 时，不允许把该次结果伪装成正常成功输出

## 3. `plan_review` 合同终稿

### 输入 Schema

```json
{
  "type": "object",
  "required": [
    "plan_draft_id",
    "plan_draft_version",
    "plan_draft_json",
    "project_category",
    "category_profile_version",
    "category_profile_json"
  ],
  "properties": {
    "plan_draft_id": {"type": "string"},
    "plan_draft_version": {"type": "integer", "minimum": 1},
    "plan_draft_json": {"type": "object"},
    "project_category": {"type": "string"},
    "category_profile_version": {"type": "integer", "minimum": 1},
    "category_profile_json": {"type": "object"},
    "project_context_json": {"type": "object"},
    "artifact_refs": {"type": "array", "items": {"type": "object"}},
    "normalized_inputs": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["status"],
        "properties": {
          "status": {
            "type": "string",
            "enum": ["success", "failure", "unsupported_or_denied"]
          },
          "summary_json": {"type": "object"}
        }
      }
    }
  }
}
```

### 输出 `result_json` Schema

```json
{
  "type": "object",
  "required": [
    "review_result_id",
    "review_version",
    "decision",
    "compile_allowed",
    "blocking_issues",
    "advisory_issues"
  ],
  "properties": {
    "review_result_id": {"type": "string"},
    "review_version": {"type": "integer", "minimum": 1},
    "decision": {"type": "string", "enum": ["approved", "approved_with_advisory", "rejected"]},
    "compile_allowed": {"type": "boolean"},
    "blocking_issues": {"type": "array", "items": {"type": "object"}},
    "advisory_issues": {"type": "array", "items": {"type": "object"}},
    "rewrite_hints": {"type": "array", "items": {"type": "string"}}
  }
}
```

当前总纲口径补充：

- 这份 `result_json` 可以继续保留为旧 adjudication 合同形态
- 但 `final_status = production_passed / released_by_human` 不应再被解释为最终业务完成语义
- 更准确的方向应逐步收口到：
  - `decision`
  - `completed`
  - `reason_summary`
  - `manual_release_required`

### SQLite 映射

1. `workflow_plan_review_results.id` ← `review_result_id`
2. `workflow_plan_review_results.review_version` ← `review_version`
3. `workflow_plan_review_results.decision` ← `decision`
4. `workflow_plan_review_results.compile_allowed` ← `compile_allowed`
5. `workflow_plan_review_results.blocking_issues_json` ← `blocking_issues`
6. `workflow_plan_review_results.advisory_issues_json` ← `advisory_issues`

## 4. `plan_compile` 合同终稿

### 输入 Schema

```json
{
  "type": "object",
  "required": [
    "plan_draft_json",
    "plan_review_result_json",
    "category_profile_json",
    "role_context_json"
  ],
  "properties": {
    "plan_draft_json": {"type": "object"},
    "plan_review_result_json": {"type": "object"},
    "category_profile_json": {"type": "object"},
    "role_context_json": {"type": "object"},
    "artifact_refs": {"type": "array", "items": {"type": "object"}},
    "normalized_inputs": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["status"],
        "properties": {
          "status": {
            "type": "string",
            "enum": ["success", "failure", "unsupported_or_denied"]
          },
          "summary_json": {"type": "object"}
        }
      }
    }
  }
}
```

### 输出 `result_json` Schema

```json
{
  "type": "object",
  "required": [
    "compiled_plan_id",
    "compiled_version",
    "compiled_tasks",
    "risk_summary"
  ],
  "properties": {
    "compiled_plan_id": {"type": "string"},
    "compiled_version": {"type": "integer", "minimum": 1},
    "compiled_tasks": {
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "compiled_task_id",
          "name",
          "role_type",
          "brain_kind",
          "delivery_contract",
          "verification_contract",
          "risk_level"
        ],
        "properties": {
          "compiled_task_id": {"type": "string"},
          "name": {"type": "string"},
          "role_type": {"type": "string"},
          "brain_kind": {"type": "string"},
          "delivery_contract": {"type": "object"},
          "verification_contract": {"type": "object"},
          "risk_level": {"type": "string"}
        }
      }
    },
    "risk_summary": {"type": "object"}
  }
}
```

### SQLite 映射

1. `workflow_compiled_plans.id` ← `compiled_plan_id`
2. `workflow_compiled_plans.compiled_version` ← `compiled_version`
3. `workflow_compiled_tasks.id` ← `compiled_task_id`
4. `workflow_compiled_tasks.delivery_contract_json` ← `delivery_contract`
5. `workflow_compiled_tasks.verification_contract_json` ← `verification_contract`
6. `workflow_compiled_tasks.risk_level` ← `risk_level`

## 5. `repair_design` 合同终稿

### 输入 Schema

```json
{
  "type": "object",
  "required": [
    "failed_task_context_json",
    "failure_reason_json",
    "original_contracts_json",
    "runtime_summary_json"
  ],
  "properties": {
    "failed_task_context_json": {"type": "object"},
    "failure_reason_json": {"type": "object"},
    "original_contracts_json": {"type": "object"},
    "runtime_summary_json": {"type": "object"},
    "artifact_refs": {"type": "array", "items": {"type": "object"}},
    "normalized_failure_status": {
      "type": "string",
      "enum": ["success", "failure", "unsupported_or_denied"]
    }
  }
}
```

### 输出 `result_json` Schema

```json
{
  "type": "object",
  "required": [
    "repair_plan_draft_id",
    "repair_plan_json",
    "repair_reasoning_summary"
  ],
  "properties": {
    "repair_plan_draft_id": {"type": "string"},
    "repair_plan_json": {"type": "object"},
    "repair_reasoning_summary": {"type": "string"},
    "replaced_constraints": {"type": "array", "items": {"type": "string"}}
  }
}
```

### SQLite 映射建议

第一版建议：

1. 新建 `repair_plan_drafts`
2. 或复用 `workflow_plan_drafts`，通过 `draft_kind=repair` 区分

## 6. `acceptance_mapping` 合同终稿

### 输入 Schema

```json
{
  "type": "object",
  "required": [
    "project_category",
    "category_profile_json",
    "artifact_summary_json",
    "coverage_summary_json"
  ],
  "properties": {
    "project_category": {"type": "string"},
    "category_profile_json": {"type": "object"},
    "artifact_summary_json": {"type": "object"},
    "coverage_summary_json": {"type": "object"},
    "normalized_inputs": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["status"],
        "properties": {
          "status": {
            "type": "string",
            "enum": ["success", "failure", "unsupported_or_denied"]
          },
          "summary_json": {"type": "object"}
        }
      }
    }
  }
}
```

### 输出 `result_json` Schema

```json
{
  "type": "object",
  "required": [
    "acceptance_profile_id",
    "production_acceptance_profile_id",
    "required_surfaces",
    "required_journeys",
    "required_evidence"
  ],
  "properties": {
    "acceptance_profile_id": {"type": "string"},
    "production_acceptance_profile_id": {"type": "string"},
    "required_surfaces": {"type": "array", "items": {"type": "string"}},
    "required_journeys": {"type": "array", "items": {"type": "string"}},
    "required_evidence": {"type": "array", "items": {"type": "string"}},
    "release_requirements": {"type": "array", "items": {"type": "string"}}
  }
}
```

### SQLite 映射

1. `acceptance_profiles.id` ← `acceptance_profile_id`
2. `production_acceptance_profiles.id` ← `production_acceptance_profile_id`
3. 相关结构写入 `required_*_json`

## 7. `completion_adjudication` 合同终稿

### 输入 Schema

```json
{
  "type": "object",
  "required": [
    "execution_summary_json",
    "delivery_summary_json",
    "verification_summary_json",
    "acceptance_summary_json",
    "manual_release_state_json"
  ],
  "properties": {
    "execution_summary_json": {"type": "object"},
    "delivery_summary_json": {"type": "object"},
    "verification_summary_json": {"type": "object"},
    "acceptance_summary_json": {"type": "object"},
    "manual_release_state_json": {"type": "object"},
    "normalized_execution_status": {
      "type": "string",
      "enum": ["success", "failure", "unsupported_or_denied"]
    },
    "normalized_verification_status": {
      "type": "string",
      "enum": ["success", "failure", "unsupported_or_denied"]
    }
  }
}
```

### 输出 `result_json` Schema

```json
{
  "type": "object",
  "required": [
    "functional_passed",
    "production_passed",
    "manual_release_required",
    "manual_release_completed",
    "final_status",
    "decision_reason"
  ],
  "properties": {
    "functional_passed": {"type": "boolean"},
    "production_passed": {"type": "boolean"},
    "manual_release_required": {"type": "boolean"},
    "manual_release_completed": {"type": "boolean"},
    "final_status": {
      "type": "string",
      "enum": ["failed", "functional_passed", "production_passed", "released_by_human"]
    },
    "decision_reason": {"type": "string"}
  }
}
```

### SQLite 映射

1. `acceptance_runs.final_status`
2. `acceptance_runs.functional_passed`
3. `acceptance_runs.production_passed`
4. `acceptance_runs.manual_release_required`
5. `acceptance_runs.decision_reason`

补充说明：

- 这些映射反映的是旧表结构与历史兼容现实
- 按当前钱学森总纲，后续不应只围绕 `acceptance_runs.*` 承载最终裁决语义
- 更合理的方向是并行补出 `VerificationResult / CompletionVerdict` 对应的结构化落点

## 8. `workspace_explanation` 合同终稿

### 输入 Schema

```json
{
  "type": "object",
  "required": [
    "workspace_context_json",
    "risk_summary_json",
    "latest_decision_summary_json"
  ],
  "properties": {
    "workspace_context_json": {"type": "object"},
    "risk_summary_json": {"type": "object"},
    "latest_decision_summary_json": {"type": "object"},
    "normalized_inputs": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["status"],
        "properties": {
          "status": {
            "type": "string",
            "enum": ["success", "failure", "unsupported_or_denied"]
          },
          "summary_json": {"type": "object"}
        }
      }
    }
  }
}
```

### 输出 `result_json` Schema

```json
{
  "type": "object",
  "required": [
    "headline",
    "summary",
    "top_blockers",
    "recommended_actions"
  ],
  "properties": {
    "headline": {"type": "string"},
    "summary": {"type": "string"},
    "top_blockers": {"type": "array", "items": {"type": "string"}},
    "recommended_actions": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["action_key", "label", "reason"],
        "properties": {
          "action_key": {"type": "string"},
          "label": {"type": "string"},
          "reason": {"type": "string"},
          "deep_link": {"type": "string"}
        }
      }
    },
    "explain_links": {"type": "array", "items": {"type": "string"}}
  }
}
```

### SQLite 映射建议

建议写入：

1. `workspace_snapshots.snapshot_json`
2. 或单独 `workspace_explanations` 表

## 9. 错误返回合同

`easymvp-brain` 调用失败时，建议统一返回：

```json
{
  "error_code": "brain_contract_invalid",
  "error_message": "missing compiled_tasks",
  "trace_id": "trace_xxx",
  "retryable": false
}
```

如果失败来自 `brain-v3` 内置脑能力不可用或被拒绝，建议统一折叠为：

```json
{
  "error_code": "brain_upstream_unsupported_or_denied",
  "error_message": "upstream tool is unsupported or denied by policy",
  "trace_id": "trace_xxx",
  "retryable": false
}
```

建议 Go DTO：

```go
type BrainContractError struct {
    ErrorCode    string `json:"error_code"`
    ErrorMessage string `json:"error_message"`
    TraceID      string `json:"trace_id,omitempty"`
    Retryable    bool   `json:"retryable"`
}
```

## 10. 第一版实现建议

落地顺序：

1. 先实现 envelope 解析与校验
2. 先落 `plan_review` / `plan_compile` / `acceptance_mapping`
3. 再落 `completion_adjudication`
4. 最后补 `repair_design` / `workspace_explanation`
5. 在所有合同入口统一补 `unsupported_or_denied` 归一化校验，禁止把该状态当成成功态吞掉

## 11. 下游实现位置建议

1. Go DTO：`apps/core/internal/model/braincontracts/`
2. Schema 校验：`apps/core/internal/logic/braincontracts/`
3. 调用封装：`apps/core/internal/service/easymvp_brain.go`
4. 审计映射：`apps/core/internal/service/audit_support_*.go`
