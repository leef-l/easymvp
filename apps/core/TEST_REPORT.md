# 阶段1测试执行报告

## 执行命令
```bash
cd C:\Users\Public\project\easymvp\apps\core
go test ./...
go test ./... -v
```

## 执行结果

**全部通过，无需修复。**

### 包覆盖情况

| 类别 | 数量 | 说明 |
|------|------|------|
| 无测试文件 | 35 个包 | api/*, controller/*, dao, model/*, cmd, consts, packed 等 |
| 有测试并通过 | 1 个包 | `internal/service` |

### `internal/service` 测试详情

共 **78 个测试用例**，全部 **PASS**，包括：

- 数据库相关测试（SQLite 嵌入式，正常通过）：
  - `TestApplyManualReleasePersistsAllAffectedRows`
  - `TestApplyManualReleaseReturnsExistingApprovalWithoutDuplicatingRows`
  - `TestAdjudicateAcceptanceAggregatePersistsAwaitingManualReleaseState`
  - `TestAdjudicateAcceptanceAggregatePersistsCompletedState`
  - `TestAdjudicateAcceptanceAggregatePersistsFailedStateAndTriggersRepairDraft`
  - `TestQueryPlanBaselineUsesExpectedIndexes`
  - `TestFindReusableRunBindingForTaskReturnsActiveBinding`
  - `TestAppendRunEventIndexDeduplicatesRepeatedPayload`
  - `TestAppendRunCheckpointForStateDeduplicatesUnchangedState`

- 纯逻辑测试（全部通过）：
  - `TestDeriveAcceptanceRunStatusFromAdjudication`（含 3 个子测试）
  - `TestBuildProjectStageProgressDefaultsUnknownStatusToDesign`
  - `TestBuildProjectActionInboxPrioritizesRepairDraftAndRespectsLimit`
  - `TestBuildWorkspaceCompletionVerdictRequiresManualReviewForManualRelease`
  - `TestBuildWorkspaceCompletionVerdictCompletesCleanPass`
  - `TestBuildProjectWorkspaceExplanationUsesDeniedFallbackWhenRuntimePolicyBlocks`
  - `TestBuildProjectWorkspaceExplanationUsesUnsupportedFallbackWhenCapabilityMissing`
  - `TestRuntimeResumeCommandUsesEnvironmentOverrides`
  - `TestRuntimeResumeCommandFallsBackToBrainBinary`
  - ...（其余详见完整输出文件 `test_output_v.txt`）

- 子测试全部通过：
  - `TestDeriveAcceptanceRunStatusFromAdjudication` (3 个子测试)
  - `TestAcceptanceProfilesNeedRefresh` (4 个子测试)
  - `TestNormalizeAcceptanceRunMode` (4 个子测试)
  - `TestParseBoolOption` (7 个子测试)
  - `TestClassifyLogStreamKindUsesFilenameHints` (4 个子测试)
  - `TestClassifyReplayKindUsesFilenameHints` (4 个子测试)
  - `TestClassifyRunArtifactPathIgnoresCheckpointAndMetaFiles` (4 个子测试)
  - `TestNormalizeServerAddress` (4 个子测试)

## 修复记录

**无需修复** — 所有测试均已通过，未发现失败的、有问题的测试。

## 外部依赖说明

未发现明显依赖外部服务（如 brain-v3）的测试。测试中涉及的 brain 相关逻辑均为纯函数/数据转换，无实际网络调用。

## 完整输出文件

- `test_output.txt` — `go test ./...` 标准输出
- `test_output_v.txt` — `go test ./... -v` 详细输出
