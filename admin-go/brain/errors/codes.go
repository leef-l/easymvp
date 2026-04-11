package errors

// The constants in this file are the reserved error_code clause from
// 21-错误模型.md §5 — the v1 frozen namespace that every sidecar, runner, and
// orchestrator is allowed to emit on the wire. Any error_code that appears in
// a BrainError MUST be registered here (or in a future additive extension);
// emitting an unregistered code is a protocol violation per 21 §5.1.
//
// Naming rules (21 §5.1):
//   - snake_case ASCII, length <= 64
//   - domain-prefixed for clustering: sidecar.* / tool.* / llm.* / frame.* /
//     brain.* / db.* / internal.*
//   - no vague suffixes like "failed"/"error"/"problem" without a subject
//
// The brain/errors package keeps these as plain string constants (not a typed
// alias) so that cross-language sidecars can copy the list verbatim into
// their own SDKs — see 21 §3.4 for the cross-language equivalence rule.

// sidecar.* — sidecar process / handshake lifecycle.
const (
	// CodeSidecarCrashed indicates the sidecar process exited unexpectedly
	// (non-zero status without a protocol shutdown frame). See 21 §4.3 mapping
	// for os/exec.ExitError.
	CodeSidecarCrashed = "sidecar.crashed"

	// CodeSidecarTimeout indicates the sidecar did not produce a response
	// within the configured timeout for a request that had already been
	// accepted. Typically Transient.
	CodeSidecarTimeout = "sidecar.timeout"

	// CodeSidecarHandshakeFail indicates the initialize handshake defined in
	// 20-协议规格.md §4 failed (bad version, missing fields, schema mismatch).
	CodeSidecarHandshakeFail = "sidecar.handshake_fail"

	// CodeSidecarStartFailed indicates the subprocess could not even be
	// spawned (binary missing, permissions, fork error).
	CodeSidecarStartFailed = "sidecar.start_failed"

	// CodeSidecarStdoutEOF indicates the Kernel read EOF from the sidecar's
	// stdout unexpectedly. Mapped from io.EOF in 21 §4.3.
	CodeSidecarStdoutEOF = "sidecar.stdout_eof"
)

// tool.* — tool invocation failures.
const (
	// CodeToolDenied indicates a tool call was rejected by policy or the
	// sandbox (out-of-zone write, capability missing, risk gate veto).
	CodeToolDenied = "tool.denied"

	// CodeToolTimeout indicates a tool call exceeded its per-call budget.
	CodeToolTimeout = "tool.timeout"

	// CodeToolArgsInvalid indicates tool arguments failed schema validation
	// against the tool's declared ToolSchema.
	CodeToolArgsInvalid = "tool.args_invalid"

	// CodeToolRuntime indicates the tool's own runtime raised an error that
	// is not classifiable as denied / timeout / args_invalid.
	CodeToolRuntime = "tool.runtime"
)

// llm.* — LLM provider / adapter failures.
const (
	// CodeLLMRateLimit indicates short-term rate limiting (retry-after within
	// seconds). Maps to ClassTransient per 21 §4.3.
	CodeLLMRateLimit = "llm.rate_limit"

	// CodeLLMQuota indicates daily/monthly quota exhaustion. Maps to
	// ClassQuotaExceeded per 21 §4.3.
	CodeLLMQuota = "llm.quota"

	// CodeLLMBadRequest indicates the LLM provider rejected the request as
	// malformed (context too long, invalid role sequence, ...).
	CodeLLMBadRequest = "llm.bad_request"

	// CodeLLMSafety indicates a safety refusal from the provider. Maps to
	// ClassSafetyRefused per 21 §4.3.
	CodeLLMSafety = "llm.safety"

	// CodeLLMTimeout indicates the LLM provider did not respond in time.
	CodeLLMTimeout = "llm.timeout"

	// CodeLLMNetwork indicates a network-level failure reaching the provider
	// (dial/TLS/DNS). Typically ClassTransient.
	CodeLLMNetwork = "llm.network"
)

// frame.* — wire protocol framing errors (see 20-协议规格.md §3, §9).
const (
	// CodeFrameCorrupt indicates a malformed Content-Length frame or a JSON
	// payload that failed schema validation. Maps from json.SyntaxError per
	// 21 §4.3.
	CodeFrameCorrupt = "frame.corrupt"

	// CodeFrameTooLarge indicates a frame exceeded the Kernel's maximum
	// allowed payload size.
	CodeFrameTooLarge = "frame.too_large"
)

// brain.* — BrainKernel runtime invariants.
const (
	// CodeBrainBudget indicates a run/turn/tool/llm-call budget was exhausted
	// mid-execution. See 22-推理循环.md for the four-layer budget model.
	CodeBrainBudget = "brain.budget_exhausted"

	// CodeBrainLoop indicates the loop detector aborted a run because the
	// same tool call repeated past the allowed threshold.
	CodeBrainLoop = "brain.loop_detected"

	// CodeBrainInvalidState indicates the Kernel rejected an operation
	// because it was called in the wrong lifecycle state.
	CodeBrainInvalidState = "brain.invalid_state"
)

// db.* — persistence-layer failures.
const (
	// CodeDBConflict indicates a unique/constraint violation or optimistic
	// concurrency conflict. Generally ClassUserFault for uniqueness,
	// ClassTransient for deadlocks — callers MUST use the Go mapping table in
	// 21 §4.3 to distinguish.
	CodeDBConflict = "db.conflict"

	// CodeDBUnavailable indicates the database was unreachable or under
	// failover. Typically ClassTransient.
	CodeDBUnavailable = "db.unavailable"
)

// internal.* — Kernel-side bugs. These MUST only ever appear with
// ClassInternalBug and MUST trigger an alert per 21 §8.2.
const (
	// CodeInternalBug is the generic fallback for unclassified bugs when no
	// more specific code applies.
	CodeInternalBug = "internal.bug"

	// CodeInternalPanic indicates a recovered panic from Go code. See 21 §4.3
	// mapping for "panic recovered".
	CodeInternalPanic = "internal.panic"
)
