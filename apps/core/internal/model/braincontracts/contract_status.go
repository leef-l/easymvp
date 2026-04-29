package braincontracts

// NormalizedStatus defines the three canonical contract execution outcomes.
// All easymvp-brain contracts MUST return one of these values in the envelope.
type NormalizedStatus string

const (
	// StatusSuccess indicates the contract executed successfully.
	StatusSuccess NormalizedStatus = "success"

	// StatusFailure indicates the contract executed but the domain result
	// represents a failure (e.g. review rejected, completion blocked).
	StatusFailure NormalizedStatus = "failure"

	// StatusUnsupportedOrDenied indicates the brain could not execute the
	// contract because the capability is unsupported or policy-denied.
	// This MUST NOT be swallowed as success.
	StatusUnsupportedOrDenied NormalizedStatus = "unsupported_or_denied"
)

// IsValidNormalizedStatus reports whether s is a recognized status value.
func IsValidNormalizedStatus(s string) bool {
	switch NormalizedStatus(s) {
	case StatusSuccess, StatusFailure, StatusUnsupportedOrDenied:
		return true
	}
	return false
}

// EscalationType defines the six canonical escalation reasons when a brain
// cannot complete its work and control must be transferred.
type EscalationType string

const (
	// EscalationUnsupportedCapability — the target brain does not support
	// the requested capability.
	EscalationUnsupportedCapability EscalationType = "unsupported_capability"

	// EscalationPolicyDenied — the request was denied by runtime policy.
	EscalationPolicyDenied EscalationType = "policy_denied"

	// EscalationVerificationConflict — verification results are in conflict
	// and cannot be auto-resolved.
	EscalationVerificationConflict EscalationType = "verification_conflict"

	// EscalationEnvironmentUnavailable — the required execution environment
	// (browser, remote host, etc.) is not available.
	EscalationEnvironmentUnavailable EscalationType = "environment_unavailable"

	// EscalationManualReviewRequired — the situation requires human judgment
	// and cannot be automated.
	EscalationManualReviewRequired EscalationType = "manual_review_required"

	// EscalationFaultLoopDetected — a failure loop has been detected and
	// repair design is required.
	EscalationFaultLoopDetected EscalationType = "fault_loop_detected"
)

// IsValidEscalationType reports whether t is a recognized escalation type.
func IsValidEscalationType(t string) bool {
	switch EscalationType(t) {
	case EscalationUnsupportedCapability, EscalationPolicyDenied,
		EscalationVerificationConflict, EscalationEnvironmentUnavailable,
		EscalationManualReviewRequired, EscalationFaultLoopDetected:
		return true
	}
	return false
}
