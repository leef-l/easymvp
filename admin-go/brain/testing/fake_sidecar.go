package braintesting

import (
	"context"
	"time"
)

// FakeSidecar is the minimal scripted sidecar described in 25-测试策略.md
// Appendix B. It implements enough of the stdio wire protocol from
// 20-协议规格.md to satisfy Unit and Contract tests without spawning a real
// Python/TS brain process, without hitting a real LLM, and without actually
// executing tools.
//
// A FakeSidecar walks through its Script one SidecarStep at a time, applying
// deterministic side effects on every step (reply with canned result, inject
// a reverse-RPC call, crash, sleep, write a malformed frame, etc.). Tests
// drive the kernel under test, the kernel talks to the fake over stdio, and
// assertions are made on the kernel's observable behaviour.
//
// Implementations MUST treat Script as frozen at Start time: mutating the
// script mid-run is a programming error and SHOULD panic.
type FakeSidecar interface {
	// Start brings the fake online: open the stdio pipes, write the
	// protocol `hello` frame (20 §4), and park at Script[0].
	Start(ctx context.Context) error

	// Stop drains any in-flight frame, closes stdio and releases all
	// resources. It MUST be idempotent.
	Stop(ctx context.Context) error

	// SendFrame writes a single wire frame to the kernel side of the
	// transport. The frame argument is left as interface{} so that tests
	// can inject raw bytes, RPC messages or malformed payloads alike.
	SendFrame(ctx context.Context, frame interface{}) error

	// ReceiveFrame reads the next frame observed from the kernel side.
	// It MUST block until a frame is available, the context is cancelled,
	// or Stop has been called.
	ReceiveFrame(ctx context.Context) (interface{}, error)
}

// SidecarStep is a single scripted action applied by a FakeSidecar during
// Run. See 25-测试策略.md Appendix B for the taxonomy of built-in steps and
// the contract each one MUST satisfy.
//
// Tests MAY define their own steps by implementing this interface; doing so
// keeps the script ergonomic while avoiding a combinatorial explosion in the
// built-in catalogue.
type SidecarStep interface {
	// Apply mutates the FakeSidecar state to perform this step's effect.
	// Apply MUST return a Permanent BrainError when the step cannot be
	// satisfied (e.g. the kernel has not yet sent the expected request).
	Apply(s FakeSidecar) error
}

// ReplyStep instructs the FakeSidecar to answer a previously received
// request with a canned Result or Error, as documented in 25 Appendix B.
// Exactly one of Result and Error MUST be set; both set or both nil is a
// programming error.
type ReplyStep struct {
	// RequestID is the JSON-RPC id of the request being answered. See
	// 20-协议规格.md §5 for the id format ("k:<n>" for kernel-initiated
	// requests, "b:<n>" for brain-initiated).
	RequestID string

	// Result is the success payload to return. Typed as interface{} so
	// that tests can pass any of the Result structs from 20 §6.
	Result interface{}

	// Error is the failure payload to return. When non-nil it MUST be a
	// *errors.BrainError so that cross-language error fidelity is
	// preserved; typed as interface{} here only to keep this package
	// free of an errors import cycle.
	Error interface{}
}

// Apply satisfies SidecarStep for ReplyStep. See 25 Appendix B.
func (ReplyStep) Apply(s FakeSidecar) error {
	panic("unimplemented: 25-测试策略.md Appendix B ReplyStep.Apply")
}

// ReverseRPCStep instructs the FakeSidecar to initiate a brain → kernel
// reverse RPC call, simulating what a real sidecar does when it needs
// services like llm.complete or tool.invoke. See 25 Appendix B.
type ReverseRPCStep struct {
	// Method is one of the reverse-RPC method names enumerated in
	// 20-协议规格.md §7 (e.g. "llm.complete", "tool.invoke",
	// "artifact.put").
	Method string

	// Params is the untyped JSON-serialisable parameter object for the
	// call. The fake MUST marshal it via encoding/json before writing
	// the frame.
	Params interface{}
}

// Apply satisfies SidecarStep for ReverseRPCStep. See 25 Appendix B.
func (ReverseRPCStep) Apply(s FakeSidecar) error {
	panic("unimplemented: 25-测试策略.md Appendix B ReverseRPCStep.Apply")
}

// DelayStep pauses script execution for a fixed duration. Useful for
// regression tests of timeout, watchdog and circuit-breaker behaviour.
// See 25 Appendix B.
type DelayStep struct {
	// Duration is the amount of time to sleep before moving to the next
	// step. Implementations MUST honour context cancellation during the
	// sleep so that tests do not hang.
	Duration time.Duration
}

// Apply satisfies SidecarStep for DelayStep. See 25 Appendix B.
func (DelayStep) Apply(s FakeSidecar) error {
	panic("unimplemented: 25-测试策略.md Appendix B DelayStep.Apply")
}

// CrashStep forces the FakeSidecar to exit with a specific code, letting
// tests assert the kernel's Restart / Dead-letter handling from
// 02-BrainKernel设计.md §11. See 25 Appendix B.
type CrashStep struct {
	// ExitCode is the process exit code to report back through the
	// transport. Values 0..255 are honoured.
	ExitCode int
}

// Apply satisfies SidecarStep for CrashStep. See 25 Appendix B.
func (CrashStep) Apply(s FakeSidecar) error {
	panic("unimplemented: 25-测试策略.md Appendix B CrashStep.Apply")
}

// SendMalformedFrameStep writes an arbitrary byte sequence onto the wire
// without going through the normal framing layer. Used to exercise the
// kernel's 20-协议规格.md §3 frame-parser error paths. See 25 Appendix B.
type SendMalformedFrameStep struct {
	// Raw is the exact byte sequence to write, including any
	// intentionally corrupt Content-Length headers.
	Raw []byte
}

// Apply satisfies SidecarStep for SendMalformedFrameStep. See 25 Appendix B.
func (SendMalformedFrameStep) Apply(s FakeSidecar) error {
	panic("unimplemented: 25-测试策略.md Appendix B SendMalformedFrameStep.Apply")
}

// NewFakeSidecar returns a new FakeSidecar preloaded with the given script.
// The script is evaluated lazily: nothing happens until Start is called.
// See 25 Appendix B for the usage example.
func NewFakeSidecar(script []SidecarStep) FakeSidecar {
	panic("unimplemented: 25-测试策略.md Appendix B NewFakeSidecar")
}
