package braintesting

import (
	"context"
	"time"
)

// ComplianceTest is a single entry in the frozen v1 conformance suite
// defined in 25-测试策略.md §3 and enumerated in 25 Appendix A. Every SDK
// implementation (Go, TypeScript, Python, Rust) MUST pass the full set of
// ~150 tests to be considered protocol-compliant per 28-SDK交付规范.md.
//
// ID follows the numbering scheme from 25 §3.1 — e.g. "C-01" for protocol
// tests, "C-E-01" for error-model tests, "C-L-01" for Agent-Loop tests.
type ComplianceTest struct {
	// ID is the stable test identifier from 25 Appendix A.
	ID string

	// Description is a short human-readable summary of what the test
	// asserts. It is surfaced in compliance reports.
	Description string

	// Category groups tests for reporting purposes. See 25 §3.1 for the
	// allowed values: protocol / error / loop / security / persistence /
	// observability / cli.
	Category string
}

// TestResult captures the outcome of a single ComplianceTest invocation.
// See 25-测试策略.md §3 for the required result schema used by the CI
// reporter and the cross-language dashboard in 28 §5.
type TestResult struct {
	// TestID matches ComplianceTest.ID.
	TestID string

	// Status is one of "pass", "fail" or "skipped". The three values are
	// frozen v1 — no other statuses are permitted.
	Status string

	// DurationMS is the wall-clock duration of the test in milliseconds.
	DurationMS int64

	// Error is the stringified BrainError.Code:Message for failing tests,
	// empty for passes. Implementations MUST populate this field whenever
	// Status is "fail" so that the cross-SDK diff report in 28 §5 is
	// actionable.
	Error string
}

// Summary aggregates TestResult counts for a single ComplianceReport.
// See 25-测试策略.md §3 for the canonical summary format.
type Summary struct {
	// Total is the number of tests executed (passed + failed + skipped).
	Total int

	// Passed is the number of tests whose Status is "pass".
	Passed int

	// Failed is the number of tests whose Status is "fail".
	Failed int

	// Skipped is the number of tests whose Status is "skipped".
	Skipped int
}

// ComplianceReport is the aggregate output of ComplianceRunner.RunAll and
// the serialised artifact uploaded to the cross-SDK dashboard in
// 28-SDK交付规范.md §5. Fields are stable v1 and MUST NOT be renamed.
type ComplianceReport struct {
	// SDKLanguage is the target SDK identifier, e.g. "go", "ts", "py".
	SDKLanguage string

	// SDKVersion is the semver of the SDK under test.
	SDKVersion string

	// KernelVersion is the semver of the BrainKernel build used to host
	// the tests.
	KernelVersion string

	// ProtocolVersion is the wire protocol version from 20-协议规格.md §2.
	ProtocolVersion string

	// RunAt is the wall-clock time at which RunAll started.
	RunAt time.Time

	// Results maps ComplianceTest.ID → TestResult for every test that was
	// executed. Skipped tests MUST also appear so consumers can
	// distinguish "not implemented" from "not run".
	Results map[string]*TestResult

	// Summary is the pre-aggregated count matching Results.
	Summary Summary
}

// ComplianceRunner executes the frozen v1 compliance suite defined in
// 25-测试策略.md §3. List enumerates the tests registered with this runner,
// Run executes a single test by ID, and RunAll executes the full suite and
// returns an aggregate report suitable for upload to the cross-SDK
// dashboard (28-SDK交付规范.md §5).
//
// Implementations MUST be deterministic: running the same suite against the
// same Kernel build twice MUST produce identical Results maps (barring
// DurationMS). Non-determinism is a compliance failure in its own right.
type ComplianceRunner interface {
	// List returns the full set of ComplianceTest entries this runner
	// knows about, in stable ID order.
	List() []ComplianceTest

	// Run executes a single test by ID and returns its result. It MUST
	// return a Permanent BrainError when testID is unknown.
	Run(ctx context.Context, testID string) (*TestResult, error)

	// RunAll executes every test returned by List and aggregates the
	// outcomes into a ComplianceReport. It MUST NOT stop on the first
	// failure; every test MUST be attempted so the report is complete.
	RunAll(ctx context.Context) (*ComplianceReport, error)
}

// NewComplianceRunner returns a new ComplianceRunner bound to the current
// Kernel build. The concrete wiring is filled in by cmd/brain compliance
// in wave 3; this skeleton exists so that 25 §3 and 28 §5 consumers can
// compile-time reference the entry point.
func NewComplianceRunner() ComplianceRunner {
	panic("unimplemented: 25-测试策略.md §3 NewComplianceRunner")
}
