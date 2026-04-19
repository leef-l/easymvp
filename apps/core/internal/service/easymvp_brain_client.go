package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

const (
	easyMVPBrainErrorCodeConfigInvalid  = "BRN_003"
	easyMVPBrainErrorCodeExecuteFailed  = "BRN_004"
	easyMVPBrainErrorCodeDecodeFailed   = "BRN_005"
	easyMVPBrainDefaultTimeout          = 30 * time.Second
	easyMVPBrainDefaultMaxTurns         = 6
	easyMVPBrainInstructionPrefixReview = "Return only a valid easymvp-brain contract envelope JSON."
)

type EasyMVPBrainClientConfig struct {
	Mode           string
	LocalEndpoint  string
	RemoteEndpoint string
	APIKey         string
	Timeout        time.Duration
	MaxTurns       int
}

type EasyMVPBrainExecuteCommand struct {
	TaskID       string
	Instruction  string
	ContextJSON  []byte
	ContractKind string
}

type EasyMVPBrainExecuteResult struct {
	Status   string                                `json:"status"`
	Summary  string                                `json:"summary,omitempty"`
	Error    string                                `json:"error,omitempty"`
	Turns    int                                   `json:"turns"`
	Envelope *braincontracts.BrainContractEnvelope `json:"envelope,omitempty"`
}

type easyMVPBrainJSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type easyMVPBrainJSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   any             `json:"error,omitempty"`
}

type easyMVPBrainExecuteRequest struct {
	TaskID      string                     `json:"task_id,omitempty"`
	Instruction string                     `json:"instruction"`
	Context     json.RawMessage            `json:"context,omitempty"`
	Budget      *easyMVPBrainExecuteBudget `json:"budget,omitempty"`
}

type easyMVPBrainExecuteBudget struct {
	MaxTurns int `json:"max_turns,omitempty"`
}

type easyMVPBrainExecuteResponse struct {
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
	Error   string `json:"error,omitempty"`
	Turns   int    `json:"turns"`
}

func (s *sEasyMVPBrain) ResolveClientConfig(ctx context.Context) (*EasyMVPBrainClientConfig, error) {
	var (
		mode           = strings.TrimSpace(g.Cfg().MustGet(ctx, "easymvp.easymvpBrain.mode", "local-sidecar").String())
		localEndpoint  = strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "easymvp.easymvpBrain.localEndpoint", "").String()), "/")
		remoteEndpoint = strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "easymvp.easymvpBrain.remoteEndpoint", "").String()), "/")
		apiKey         = strings.TrimSpace(g.Cfg().MustGet(ctx, "easymvp.easymvpBrain.apiKey", "").String())
		timeoutText    = strings.TrimSpace(g.Cfg().MustGet(ctx, "easymvp.easymvpBrain.timeout", "30s").String())
		maxTurns       = g.Cfg().MustGet(ctx, "easymvp.easymvpBrain.maxTurns", easyMVPBrainDefaultMaxTurns).Int()
	)

	if !isAllowedValue(mode, "local-sidecar", "remote-service") {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvpBrain.mode is invalid", nil)
	}

	timeout := easyMVPBrainDefaultTimeout
	if timeoutText != "" {
		parsed, err := time.ParseDuration(timeoutText)
		if err != nil {
			return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvpBrain.timeout is invalid", err)
		}
		timeout = parsed
	}
	if maxTurns <= 0 {
		maxTurns = easyMVPBrainDefaultMaxTurns
	}

	cfg := &EasyMVPBrainClientConfig{
		Mode:           mode,
		LocalEndpoint:  localEndpoint,
		RemoteEndpoint: remoteEndpoint,
		APIKey:         apiKey,
		Timeout:        timeout,
		MaxTurns:       maxTurns,
	}
	if err := validateEasyMVPBrainClientConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *sEasyMVPBrain) ExecuteContract(ctx context.Context, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error) {
	cfg, err := s.ResolveClientConfig(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := cfg.LocalEndpoint
	if cfg.Mode == "remote-service" {
		endpoint = cfg.RemoteEndpoint
	}
	return callEasyMVPBrainExecute(ctx, cfg, endpoint, cmd)
}

func callEasyMVPBrainExecute(ctx context.Context, cfg *EasyMVPBrainClientConfig, endpoint string, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error) {
	if cfg == nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain config is required", nil)
	}
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain endpoint is empty", nil)
	}
	if strings.TrimSpace(cmd.Instruction) == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "instruction is required", nil)
	}

	payload := easyMVPBrainExecuteRequest{
		TaskID:      strings.TrimSpace(cmd.TaskID),
		Instruction: buildEasyMVPBrainInstruction(cmd.ContractKind, cmd.Instruction),
		Context:     json.RawMessage(cmd.ContextJSON),
		Budget:      &easyMVPBrainExecuteBudget{MaxTurns: cfg.MaxTurns},
	}

	body := easyMVPBrainJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      time.Now().UnixNano(),
		Method:  "brain/execute",
		Params:  payload,
	}
	rawRequest, err := json.Marshal(body)
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "marshal easymvp-brain request failed", err)
	}

	client := &http.Client{Timeout: cfg.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/rpc", bytes.NewReader(rawRequest))
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "build easymvp-brain request failed", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		recordDiagnostic(ctx, "easymvp_brain.execute", "warning", easyMVPBrainErrorCodeExecuteFailed, "call easymvp-brain failed", map[string]any{
			"contract_kind": cmd.ContractKind,
			"endpoint":      endpoint,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"error":         err.Error(),
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "call easymvp-brain failed", err)
	}
	defer resp.Body.Close()

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "read easymvp-brain response failed", err)
	}
	if resp.StatusCode >= 300 {
		recordDiagnostic(ctx, "easymvp_brain.execute", "warning", easyMVPBrainErrorCodeExecuteFailed, "easymvp-brain returned non-success status", map[string]any{
			"contract_kind": cmd.ContractKind,
			"endpoint":      endpoint,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"status_code":   resp.StatusCode,
			"body":          strings.TrimSpace(string(rawResponse)),
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "easymvp-brain returned non-success status", gerror.Newf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(rawResponse))))
	}

	var rpcResp easyMVPBrainJSONRPCResponse
	if err = json.Unmarshal(rawResponse, &rpcResp); err != nil {
		recordDiagnostic(ctx, "easymvp_brain.decode", "warning", easyMVPBrainErrorCodeDecodeFailed, "decode easymvp-brain rpc response failed", map[string]any{
			"contract_kind": cmd.ContractKind,
			"endpoint":      endpoint,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"error":         err.Error(),
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeDecodeFailed, "decode easymvp-brain rpc response failed", err)
	}
	if rpcResp.Error != nil {
		recordDiagnostic(ctx, "easymvp_brain.execute", "warning", easyMVPBrainErrorCodeExecuteFailed, "easymvp-brain rpc returned error", map[string]any{
			"contract_kind": cmd.ContractKind,
			"endpoint":      endpoint,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"rpc_error":     rpcResp.Error,
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "easymvp-brain rpc returned error", gerror.Newf("%v", rpcResp.Error))
	}

	var execResp easyMVPBrainExecuteResponse
	if err = json.Unmarshal(rpcResp.Result, &execResp); err != nil {
		recordDiagnostic(ctx, "easymvp_brain.decode", "warning", easyMVPBrainErrorCodeDecodeFailed, "decode easymvp-brain execute result failed", map[string]any{
			"contract_kind": cmd.ContractKind,
			"endpoint":      endpoint,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"error":         err.Error(),
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeDecodeFailed, "decode easymvp-brain execute result failed", err)
	}

	result := &EasyMVPBrainExecuteResult{
		Status:  execResp.Status,
		Summary: execResp.Summary,
		Error:   execResp.Error,
		Turns:   execResp.Turns,
	}
	if strings.TrimSpace(execResp.Summary) != "" {
		envelope, err := decodeEasyMVPBrainEnvelope(execResp.Summary)
		if err != nil {
			recordDiagnostic(ctx, "easymvp_brain.decode", "warning", easyMVPBrainErrorCodeDecodeFailed, "decode easymvp-brain envelope from summary failed", map[string]any{
				"contract_kind": cmd.ContractKind,
				"endpoint":      endpoint,
				"task_id":       strings.TrimSpace(cmd.TaskID),
				"error":         err.Error(),
			})
			return nil, err
		}
		result.Envelope = envelope
	}
	return result, nil
}

func decodeEasyMVPBrainEnvelope(summary string) (*braincontracts.BrainContractEnvelope, error) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil, nil
	}
	var envelope braincontracts.BrainContractEnvelope
	if err := json.Unmarshal([]byte(summary), &envelope); err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeDecodeFailed, "decode easymvp-brain envelope from summary failed", err)
	}
	return &envelope, nil
}

func validateEasyMVPBrainClientConfig(cfg *EasyMVPBrainClientConfig) error {
	if cfg == nil {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain client config is required", nil)
	}
	if !isAllowedValue(cfg.Mode, "local-sidecar", "remote-service") {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain client mode is invalid", nil)
	}
	if cfg.Timeout <= 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain client timeout must be positive", nil)
	}
	if cfg.MaxTurns <= 0 {
		return wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain client max turns must be positive", nil)
	}
	switch cfg.Mode {
	case "local-sidecar":
		if cfg.LocalEndpoint == "" {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "local-sidecar mode requires localEndpoint", nil)
		}
	case "remote-service":
		if cfg.RemoteEndpoint == "" {
			return wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "remote-service mode requires remoteEndpoint", nil)
		}
	}
	return nil
}

func buildEasyMVPBrainInstruction(contractKind, instruction string) string {
	instruction = strings.TrimSpace(instruction)
	contractKind = strings.TrimSpace(contractKind)
	if contractKind == "" {
		return easyMVPBrainInstructionPrefixReview + "\n" + instruction
	}
	return easyMVPBrainInstructionPrefixReview + "\nContract Kind: " + contractKind + "\n" + instruction
}
