package errors

// Fingerprint computes the 16-hex-char aggregation key defined in
// 21-错误模型.md §6.1.
//
// The algorithm takes (error_code, brain_id, cause.error_code, normalized
// message), joins them with "|", hashes the canonical string with SHA-256,
// and returns the first 8 bytes as lowercase hex. The message normalizer
// (21 §6.2) replaces dynamic segments — numbers, UUIDs, paths, timestamps,
// hex blobs, quoted strings — so that 1000 variants of the same error
// collapse into one fingerprint. Fields that are volatile across otherwise
// identical failures (occurred_at, trace_id/span_id, sidecar_pid, attempt,
// full stack, request ids) MUST NOT be included; see 21 §6.3.
//
// Implementation is deferred to a later wave — this skeleton only fixes the
// signature so callers can compile against it.
func Fingerprint(err *BrainError) string {
	panic("unimplemented: 21-错误模型.md §6 Fingerprint")
}
