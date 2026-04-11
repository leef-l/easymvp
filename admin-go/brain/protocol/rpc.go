package protocol

import (
	"context"
	"encoding/json"
)

// RPCMessage is the decoded JSON-RPC 2.0 envelope as defined in
// 20-协议规格.md §3.1. A single RPCMessage can represent any of the four
// JSON-RPC shapes — Request, Notification, Success response, Error
// response — distinguished by which of Method, Result, Error is present.
//
// Field rules (20-协议规格.md §3 and §4):
//
//   - JSONRPC MUST be the literal "2.0" on the wire.
//   - ID MUST carry the string-prefix tuple `k:<seq>` or `s:<seq>` per
//     §4.2 for Requests and responses; it MUST be omitted (or JSON null)
//     for Notifications per §4.4.
//   - Params, Result, and Error are kept as json.RawMessage so the
//     dispatcher can route frames without eagerly decoding payloads,
//     preserving zero-copy forwarding when needed.
type RPCMessage struct {
	// JSONRPC is the literal "2.0" version tag required by JSON-RPC
	// 2.0 and §3.1.
	JSONRPC string `json:"jsonrpc"`

	// ID is the string-prefix tuple defined in 20-协议规格.md §4.2.
	// Empty means "this is a notification" per §4.4.
	ID string `json:"id,omitempty"`

	// Method is the RPC method name from the methods.go namespace.
	// Non-empty on Requests and Notifications; empty on responses.
	Method string `json:"method,omitempty"`

	// Params is the raw JSON payload of a Request or Notification. The
	// dispatcher unmarshals it lazily inside the handler that owns the
	// method.
	Params json.RawMessage `json:"params,omitempty"`

	// Result is the raw JSON payload of a success response. Mutually
	// exclusive with Error per JSON-RPC 2.0.
	Result json.RawMessage `json:"result,omitempty"`

	// Error carries the structured error object documented in
	// 20-协议规格.md §9. Mutually exclusive with Result.
	Error *RPCError `json:"error,omitempty"`
}

// RPCError is the v1 JSON-RPC error object with the strict `data`
// schema from 20-协议规格.md §9.1. Data is intentionally kept as
// json.RawMessage so the wire layer neither validates nor mutates the
// BrainError payload embedded inside it — higher layers use
// errors.WrapRPCError to materialize the typed *BrainError.
type RPCError struct {
	// Code is the numeric error code, drawn from the reserved ranges in
	// 20-协议规格.md §9.2 (-32700..-32600 JSON-RPC standard,
	// -32099..-32000 BrainKernel reserved, -32800 cancelled).
	Code int `json:"code"`

	// Message is a short human-readable description. §9.1 requires it
	// to be populated for every error response.
	Message string `json:"message"`

	// Data carries the BrainError payload (class / retryable /
	// error_code / fingerprint / trace_id / cause / suggestions) per
	// 20-协议规格.md §9.1 MUST rules 1-6. Empty for standard JSON-RPC
	// errors where no BrainError context is available.
	Data json.RawMessage `json:"data,omitempty"`
}

// BidirRPC is the full-duplex JSON-RPC abstraction that sits on top of a
// FrameReader/FrameWriter pair. Implementations MUST obey the
// bidirectional id-namespace rules from 20-协议规格.md §4 — every
// outbound Request uses its own prefix (`k:` for host, `s:` for
// sidecar) and inbound frames are dispatched by prefix inspection per
// §4.3.
//
// BidirRPC is the only way higher layers emit Requests, Notifications,
// and register inbound handlers; the FrameReader/Writer are plumbing and
// should not be touched directly above this layer.
type BidirRPC interface {
	// Call sends a Request and blocks until the matching response
	// arrives or ctx is done. It MUST enforce the in-flight window
	// from 20-协议规格.md §4.6 and return a BrainError built by
	// WrapRPCError when the response carries a non-nil Error.
	Call(ctx context.Context, method string, params interface{}, result interface{}) error

	// Notify sends a Notification (no id, no response). It MUST NOT
	// allocate a seq from the id counter per §4.4.
	Notify(ctx context.Context, method string, params interface{}) error

	// Handle registers a server-side handler for the given method. It
	// is idempotent per method name but MUST panic if the same method
	// is registered twice, because that indicates a programming bug
	// rather than a recoverable condition (21-错误模型.md §3 programmer
	// errors).
	Handle(method string, handler HandlerFunc)
}

// HandlerFunc is the server-side handler signature invoked by BidirRPC
// when an inbound Request or Notification matches a registered method.
// The returned interface{} is marshalled into RPCMessage.Result; a
// non-nil error is converted into RPCMessage.Error via WrapRPCError /
// NewProtocolError, so handlers MAY return a plain error and let the
// dispatcher route it into the BrainError taxonomy.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)
