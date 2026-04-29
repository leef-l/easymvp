package consts

import "time"

// =============================================================================
// Brain / AI 运行时
// =============================================================================

// BrainErrorCodeContractInvalid contract envelope missing required fields
const BrainErrorCodeContractInvalid = "BRN_001"

// BrainErrorCodeUnsupportedKind unsupported kind in contract
const BrainErrorCodeUnsupportedKind = "BRN_002"

// BrainErrorCodeConfigInvalid brain client configuration invalid
const BrainErrorCodeConfigInvalid = "BRN_003"

// BrainErrorCodeExecuteFailed brain execution failed
const BrainErrorCodeExecuteFailed = "BRN_004"

// BrainErrorCodeDecodeFailed brain response decode failed
const BrainErrorCodeDecodeFailed = "BRN_005"

// BrainDefaultTimeout default timeout for brain calls
const BrainDefaultTimeout = 30 * time.Second

// BrainDefaultMaxTurns default max conversation turns
const BrainDefaultMaxTurns = 6

// BrainInstructionPrefixReview standard instruction prefix for brain review calls
const BrainInstructionPrefixReview = "Return only a valid easymvp-brain contract envelope JSON."

// =============================================================================
// Startup / 配置源
// =============================================================================

// StartupConfigSourceCLI value from CLI flag
const StartupConfigSourceCLI = "cli"

// StartupConfigSourceConfig value from config file
const StartupConfigSourceConfig = "config"

// StartupConfigSourceDefault built-in default value
const StartupConfigSourceDefault = "default"

// =============================================================================
// 数据库表名
// =============================================================================

// TableRepairPlanDrafts repair plan drafts table
const TableRepairPlanDrafts = "repair_plan_drafts"

// =============================================================================
// 常用 limit / 默认值（跨 service 使用）
// =============================================================================

// ProjectWorkspaceActivityLimit project workspace activity feed limit
const ProjectWorkspaceActivityLimit = 12

// ProjectWorkspaceInboxLimit project workspace action inbox limit
const ProjectWorkspaceInboxLimit = 10

// WorkspaceHomeProjectLimit workspace home projects list limit
const WorkspaceHomeProjectLimit = 12

// WorkspaceHomeAttentionLimit workspace home attention items limit
const WorkspaceHomeAttentionLimit = 10

// WorkspaceHomeActivityLimit workspace home activity feed limit
const WorkspaceHomeActivityLimit = 12

// WorkspaceHomeReleaseLimit workspace home release items limit
const WorkspaceHomeReleaseLimit = 8

// WorkspaceEventDefaultLimit default SSE event batch limit
const WorkspaceEventDefaultLimit = 50

// ReplayTimelineDefaultLimit default replay timeline entries limit
const ReplayTimelineDefaultLimit = 50

// ReplayTimelineMaxLimit maximum replay timeline entries limit
const ReplayTimelineMaxLimit = 200

// ReplayRawDefaultLimit default replay raw data limit (bytes)
const ReplayRawDefaultLimit = 8192

// ReplayRawMaxLimit maximum replay raw data limit (bytes)
const ReplayRawMaxLimit = 65536

// LogSegmentDefaultLimit default log segment entries limit
const LogSegmentDefaultLimit = 50

// LogSegmentMaxLimit maximum log segment entries limit
const LogSegmentMaxLimit = 200

// RuntimeRunEventDefaultLimit default runtime run event limit
const RuntimeRunEventDefaultLimit = 50

// RuntimeRunEventMaxLimit maximum runtime run event limit
const RuntimeRunEventMaxLimit = 200

// ExecutionRecentBindingLimit recent execution binding count limit
const ExecutionRecentBindingLimit = 8

// ExecutionReplayDefaultLimit default execution replay artifact limit
const ExecutionReplayDefaultLimit = 10

// ExecutionLogDefaultLimit default execution log segment limit
const ExecutionLogDefaultLimit = 10

// PlanTaskProjectionLimit maximum tasks in plan projection
const PlanTaskProjectionLimit = 64

// SnapshotFreshnessWindow workspace snapshot freshness window
const SnapshotFreshnessWindow = 45 * time.Second

// =============================================================================
// 默认 Profile 版本
// =============================================================================

// DefaultCategoryProfileVersion default category profile version
const DefaultCategoryProfileVersion = "default/v1"

// DefaultAcceptanceProfileVersion default acceptance profile version
const DefaultAcceptanceProfileVersion = "default/v1"

// DefaultRoleProfileVersion default role profile version
const DefaultRoleProfileVersion = "default/v1"
