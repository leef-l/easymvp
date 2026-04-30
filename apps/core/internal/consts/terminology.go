package consts

// =============================================================================
// 统一术语常量 (C-05: 前后端术语统一)
//
// 所有模块在日志、API 响应和 UI 标签中应引用这些常量，
// 避免同义但不同名导致的前后端理解歧义。
// =============================================================================

// TermVerification 验证 — 代码级自动检查（单元测试、集成测试、安全扫描等）。
const TermVerification = "验证"

// TermAcceptance 验收 — 多层综合评估（功能验收 + 生产验收 + 人工审查）。
const TermAcceptance = "验收"

// TermRework 返工 — 结构化修复流程（由 repair_design 驱动）。
const TermRework = "返工"

// TermCompletion 完成 — 经裁决确认的最终状态（completion_adjudication 通过后）。
const TermCompletion = "完成"

// TermManualGate 人工检查点 — 需要人工介入才能推进的门控节点。
const TermManualGate = "人工检查点"

// TermAlternateChannel 替代验证通道 — 首选通道不可用时的降级路径（如 github_actions → manual_review）。
const TermAlternateChannel = "替代验证通道"

// ---------------------------------------------------------------------------
// 英文标识（用于 JSON / API 字段值）
// ---------------------------------------------------------------------------

// TermKeyVerification is the canonical English key for verification.
const TermKeyVerification = "verification"

// TermKeyAcceptance is the canonical English key for acceptance.
const TermKeyAcceptance = "acceptance"

// TermKeyRework is the canonical English key for rework.
const TermKeyRework = "rework"

// TermKeyCompletion is the canonical English key for completion.
const TermKeyCompletion = "completion"

// TermKeyManualGate is the canonical English key for manual gate.
const TermKeyManualGate = "manual_gate"

// TermKeyAlternateChannel is the canonical English key for alternate channel.
const TermKeyAlternateChannel = "alternate_channel"
