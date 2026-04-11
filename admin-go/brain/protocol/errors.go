// Package protocol implements the stdio wire protocol defined in
// 20-协议规格.md — the frozen v1 contract between the BrainKernel host
// process and each brain sidecar.
//
// The package is deliberately split into five files that mirror the spec
// layout:
//
//   - frame.go     — §2 transport-layer framing (header block + body).
//   - rpc.go       — §3 and §4 JSON-RPC 2.0 session layer and the
//     bidirectional id namespace.
//   - lifecycle.go — §6 initialize / shutdown handshake payloads.
//   - methods.go   — §10 frozen v1 method-name namespace.
//   - errors.go    — §9 helpers that wrap protocol-layer failures into the
//     shared *errors.BrainError carrier.
//
// protocol sits just above the errors subpackage in the dependency topology:
// it imports brain/errors to build typed wire-layer failures, but it MUST
// NOT import any other brain subpackage. Higher-level layers (agent, loop,
// runner) wire the protocol package up by composition — see
// 20-协议规格.md §1 for the rationale of keeping the wire protocol a leaf.
//
// This package has zero external dependencies and uses only the Go standard
// library, per the rules in brain骨架实施计划.md §4.6.
package protocol

import (
	brainerrors "easymvp/brain/errors"
)

// NewProtocolError wraps a raw protocol-layer failure into a
// *errors.BrainError so that callers above the wire layer see the same
// BrainError contract regardless of whether the failure originated on the
// frame reader, the JSON-RPC dispatcher, or the lifecycle handshake.
//
// The helper is the single entry point through which wire-layer code is
// allowed to construct BrainErrors; see 20-协议规格.md §9 for the wire
// encoding of the resulting error frame and 21-错误模型.md §3.3 for the
// BrainError construction contract that the returned value MUST satisfy.
func NewProtocolError(code string, msg string) *brainerrors.BrainError {
	panic("unimplemented: 20-协议规格.md §9 NewProtocolError")
}

// WrapFrameError converts a low-level frame reader/writer failure (for
// example a malformed header block, an oversized body, or a write-timeout)
// into a *errors.BrainError tagged with the appropriate frame.* error_code
// from the reserved clause in 21-错误模型.md §5. The raw cause is preserved
// on BrainError.Cause to comply with 20-协议规格.md §9.1 MUST rule 1.
func WrapFrameError(code string, msg string, cause error) *brainerrors.BrainError {
	panic("unimplemented: 20-协议规格.md §9 WrapFrameError")
}

// WrapRPCError converts an inbound JSON-RPC error object (the
// wire representation defined in RPCError) into a *errors.BrainError that
// the host-process business layer can consume uniformly. The mapping from
// the JSON-RPC numeric code to ErrorClass / error_code follows
// 20-协议规格.md §9.2.
func WrapRPCError(rpcErr *RPCError) *brainerrors.BrainError {
	panic("unimplemented: 20-协议规格.md §9 WrapRPCError")
}
