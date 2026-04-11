package protocol

import "context"

// Frame is the wire representation of a single stdio message as defined in
// 20-协议规格.md §2.2. It carries the exact bytes that cross the
// host/sidecar boundary: a Content-Length-prefixed header block followed by
// a UTF-8 JSON body.
//
// Construction rules (20-协议规格.md §2.3):
//
//   - ContentLength MUST equal len(Body) in bytes (not runes).
//   - ContentType MUST be "application/vscode-jsonrpc; charset=utf-8" on the
//     wire for v1 producers; readers MAY accept an omitted Content-Type and
//     treat it as equivalent to the canonical value.
//   - Body MUST NOT carry a BOM and MUST be the raw UTF-8 JSON payload —
//     the framing layer does NOT parse JSON itself.
type Frame struct {
	// ContentLength is the exact byte length of Body, as declared in the
	// Content-Length header. The frame reader MUST set this from the
	// header and the frame writer MUST set it before serialization.
	ContentLength int

	// ContentType is the Content-Type header value, canonically
	// "application/vscode-jsonrpc; charset=utf-8" per 20-协议规格.md §2.3.
	ContentType string

	// Body is the raw UTF-8 JSON bytes of the frame body. The framing
	// layer treats Body as opaque; higher layers unmarshal it into an
	// RPCMessage (see rpc.go).
	Body []byte
}

// FrameReader is the pull-based interface implemented by the host-side and
// sidecar-side stdio framers. Implementations MUST obey the parser state
// machine in 20-协议规格.md §2.6, including the 8 KiB header-line cap and
// the 16 MiB body cap from §2.4.
//
// ReadFrame is the only way higher layers obtain an inbound Frame; the
// framer MUST NOT expose the underlying io.Reader directly.
type FrameReader interface {
	// ReadFrame reads exactly one frame from the underlying stream. It
	// returns a non-nil *Frame on success. On stream termination (peer
	// EOF, ctx cancellation, oversized body, malformed header) it MUST
	// return a *errors.BrainError built by WrapFrameError so the caller
	// can drive the half-close state machine in 20-协议规格.md §6.
	ReadFrame(ctx context.Context) (*Frame, error)
}

// FrameWriter is the push-based interface implemented by the host-side and
// sidecar-side stdio framers. Implementations MUST serialize all frames
// through a single writer goroutine per stream, as mandated by
// 20-协议规格.md §2.7 — concurrent writes to the same stdio stream cause
// header/body interleaving and immediate protocol corruption.
//
// WriteFrame is the only way higher layers emit an outbound Frame.
type FrameWriter interface {
	// WriteFrame serializes frame onto the underlying stream. It MUST
	// honour the write-timeout and back-pressure rules in
	// 20-协议规格.md §5.2; on timeout or peer hang it MUST return a
	// BrainError built by WrapFrameError so the caller can transition
	// the sidecar into Draining / Closed.
	WriteFrame(ctx context.Context, frame *Frame) error
}
