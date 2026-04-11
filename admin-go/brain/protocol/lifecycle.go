package protocol

// InitializeRequest is the payload of the host→sidecar `initialize`
// request that opens every sidecar session, as specified in
// 20-协议规格.md §6 (lifecycle) and the method table in §10.1.
//
// The struct mirrors the wire schema so that Go and non-Go brain SDKs can
// decode/encode symmetrically. Fields that appear only in specific
// LLMAccess modes are documented individually below; the host MUST NOT
// populate them in modes where they are forbidden (see 23-安全模型.md).
type InitializeRequest struct {
	// ProtocolVersion is the host-advertised wire protocol version
	// (semver). The sidecar MUST refuse to initialize if it cannot speak
	// the requested version. See 20-协议规格.md §6.4 Starting rules.
	ProtocolVersion string

	// KernelVersion is the BrainKernel build version, advertised for
	// diagnostics and compatibility gating. See 20-协议规格.md §6.4.
	KernelVersion string

	// Capabilities is the host-advertised capability set as a flat
	// map[string]bool. The sidecar compares it against its own
	// capability table and replies with BrainCapabilities in the
	// InitializeResponse.
	Capabilities map[string]bool

	// LLMConfig carries non-secret LLM configuration (model name,
	// temperature caps, provider name, ...). It is always populated.
	LLMConfig map[string]interface{}

	// LLMCredentials carries short-lived LLM API credentials. It is
	// populated ONLY when the LLMAccess mode is direct or hybrid, per
	// 23-安全模型.md; proxy mode MUST leave this nil to keep secrets out
	// of the sidecar address space.
	LLMCredentials map[string]string

	// WorkspacePath is the absolute path to the per-Run workspace root
	// that the sidecar is allowed to read and write. The host MUST
	// enforce the sandbox boundary; the field here is informational.
	WorkspacePath string

	// RunContext carries per-Run metadata the sidecar needs for tracing
	// and audit (run id, parent trace ids, tenant hints, ...).
	RunContext map[string]interface{}
}

// InitializeResponse is the sidecar→host reply to `initialize`. It
// advertises the sidecar's effective capabilities so the host can reject
// incompatible sessions before any business request is sent. See
// 20-协议规格.md §6.4 Starting and the method index in §10.1.
type InitializeResponse struct {
	// ProtocolVersion is the sidecar-chosen wire protocol version,
	// which MUST be a version the host advertised in the request.
	ProtocolVersion string

	// BrainVersion is the sidecar build version used for diagnostics.
	BrainVersion string

	// BrainCapabilities is the effective capability set the sidecar
	// commits to for the duration of the session.
	BrainCapabilities map[string]bool

	// SupportedTools is the list of tool identifiers the sidecar is
	// willing to handle via tool.invoke. Tools not listed here MUST
	// cause the host to reject matching tool.invoke requests locally.
	SupportedTools []string
}
