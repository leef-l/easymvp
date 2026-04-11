package errors

import "time"

// Decision is the output of the Decide retry/circuit-breaker engine defined
// in 21-错误模型.md §8.1. Schedulers, Runners, and Orchestrators MUST route
// every failure through Decide and MUST NOT write ad-hoc if/else retry
// logic — the decision matrix in 21 §8.2 is the single source of truth.
//
// This skeleton exposes only the three fields needed by the first-round
// callers; later waves will add Action / MaxRetries / DecideContext when the
// runner package lands.
type Decision struct {
	// Retry tells the caller whether to re-execute the same request. When
	// true, the caller MUST honor BackoffHint before retrying. See 21 §8.2.
	Retry bool

	// BackoffHint is the recommended delay before the next attempt. Zero
	// means "retry immediately" but the caller SHOULD still apply jitter.
	// See 21 §8.3 for the exponential backoff parameters.
	BackoffHint time.Duration

	// Reason is a human-readable explanation written to logs. It is
	// informational only — machine consumers MUST look at Retry instead.
	Reason string
}

// Decide applies the 21-错误模型.md §8.2 decision matrix to err and returns
// the retry/backoff verdict. attempt is the 1-based count of prior tries so
// Decide can apply the exponential backoff curve and the per-class budgets
// described in §8.3 / §8.4.
//
// Implementation is deferred to a later wave — see the spec for the
// normative matrix.
func Decide(err *BrainError, attempt int) Decision {
	panic("unimplemented: 21-错误模型.md §8 Decide")
}
