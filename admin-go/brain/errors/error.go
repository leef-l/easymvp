package errors

import "time"

// BrainError is the v1 frozen carrier struct defined in 21-错误模型.md §3.1.
//
// BrainError is the only error type allowed to cross Runner/Kernel/sidecar
// boundaries inside BrainKernel. Business code MUST return *BrainError and
// MUST NOT let bare fmt.Errorf / errors.New values leak past those
// boundaries. Construction goes through the errors.New / errors.Wrap builders
// (see 21 §3.3) — the fields below are exported only so that sidecar SDKs in
// other languages can decode/encode wire frames symmetrically (21 §3.4).
//
// JSON encoding preserves snake_case for cross-language equivalence. The
// InternalOnly field is tagged json:"-" and MUST NEVER appear on the wire or
// in an HTTP body; it exists only so the host process can attach stack
// traces and raw stderr for local logging per 21 §3.2 / §7.4.
type BrainError struct {
	// Class is the six-value taxonomy that drives the retry/circuit-breaker
	// matrix. MUST be one of the ClassXxx constants from class.go. See
	// 21 §3.2 and §4.
	Class ErrorClass `json:"class"`

	// ErrorCode is a snake_case identifier from the reserved registry in
	// codes.go. MUST be <= 64 chars and MUST exist in the v1 clause before
	// being emitted on the wire. See 21 §5.
	ErrorCode string `json:"error_code"`

	// Retryable is the boolean hint the scheduler consults. It MUST be
	// consistent with Class (see the matrix in 21 appendix B); e.g.
	// ClassTransient + Retryable=false is a protocol violation.
	Retryable bool `json:"retryable"`

	// Fingerprint is the 16-hex-char aggregation key produced by the
	// Fingerprint function. See 21 §6.
	Fingerprint string `json:"fingerprint"`

	// TraceID is the W3C Trace Context trace ID. It is allowed on internal
	// JSON but MUST be stripped before reaching HTTP bodies — it may only be
	// surfaced in a response header. See 21 §7.4.
	TraceID string `json:"trace_id,omitempty"`

	// SpanID is the W3C Trace Context span ID companion to TraceID.
	SpanID string `json:"span_id,omitempty"`

	// Message is a single human-readable sentence describing the failure.
	// MUST be non-empty and SHOULD be <= 200 chars. This is the string
	// surfaced in logs and UI after sanitization.
	Message string `json:"message"`

	// Hint is an optional self-service troubleshooting note. Still public —
	// do not put secrets here.
	Hint string `json:"hint,omitempty"`

	// Cause is the upstream BrainError that produced this failure. Each
	// propagation layer MUST wrap rather than swallow (21 §7.1 rule 1).
	Cause *BrainError `json:"cause,omitempty"`

	// InternalOnly carries host-side debug information that MUST NOT be
	// serialized. The json:"-" tag enforces this at the encoder level.
	InternalOnly *InternalDetail `json:"-"`

	// BrainID is the identity of the brain that raised the error (central,
	// code, browser, ...). Useful for fingerprint input and per-brain
	// budgets.
	BrainID string `json:"brain_id,omitempty"`

	// SidecarPID is the OS pid of the sidecar process at failure time. Not
	// used in fingerprints (pids change across restarts).
	SidecarPID int `json:"sidecar_pid,omitempty"`

	// OccurredAt is the UTC timestamp at which the error was constructed.
	// MUST be set by the constructor.
	OccurredAt time.Time `json:"occurred_at"`

	// Attempt is the 1-based retry counter attached by the scheduler. Not
	// used in fingerprints.
	Attempt int `json:"attempt,omitempty"`

	// Suggestions is a list of actionable remediation steps shown to the
	// user. MUST be executable advice, not "please check the config".
	Suggestions []string `json:"suggestions,omitempty"`
}

// InternalDetail carries fields that live only inside the host process and
// MUST NOT be serialized to any wire or HTTP body. See 21 §3.1 and §7.4.
type InternalDetail struct {
	// Stack is the full Go stack (or the sidecar's language-native stack) at
	// the point of failure.
	Stack string

	// RawStderr is the sidecar's un-redacted stderr tail. Host-side logging
	// sinks may record it but MUST NOT forward it to end users.
	RawStderr string

	// DebugFlags is a free-form bag of structured debugging fields. It is
	// intentionally string->string so it can be safely dumped to log lines.
	DebugFlags map[string]string
}

// Error implements the Go error interface by returning the Message field, as
// required by 21 §3.1. Callers that need structured access MUST type-assert
// back to *BrainError rather than parse this string.
func (e *BrainError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}
