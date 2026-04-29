package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
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
	easyMVPBrainDefaultTimeout          = 120 * time.Second
	easyMVPBrainDefaultMaxTurns         = 6

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
	baseURL := strings.TrimRight(strings.TrimSpace(CurrentStartupConfig(ctx).BrainServeBaseURL), "/")
	if baseURL == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "brain serve base URL is empty", nil)
	}
	return callEasyMVPBrainExecute(ctx, cfg, baseURL, cmd)
}

func callEasyMVPBrainExecute(ctx context.Context, cfg *EasyMVPBrainClientConfig, baseURL string, cmd EasyMVPBrainExecuteCommand) (*EasyMVPBrainExecuteResult, error) {
	if cfg == nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain config is required", nil)
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "brain serve base URL is empty", nil)
	}
	if strings.TrimSpace(cmd.Instruction) == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "instruction is required", nil)
	}

	payload := map[string]interface{}{
		"brain_kind":    "easymvp",
		"contract_kind": cmd.ContractKind,
		"context_json":  json.RawMessage(cmd.ContextJSON),
		"instruction":   buildEasyMVPBrainInstruction(cmd.ContractKind, cmd.Instruction),
		"max_turns":     cfg.MaxTurns,
		"timeout":       cfg.Timeout.String(),
	}

	rawRequest, err := json.Marshal(payload)
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "marshal contract request failed", err)
	}

	client := &http.Client{Timeout: cfg.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/contracts/execute", bytes.NewReader(rawRequest))
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "build contract request failed", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Retry once on transient network errors (EOF, timeout, connection issues).
		if isRetryableBrainError(err) {
			g.Log().Infof(ctx, "brain-v3 contract execution failed with retryable error, retrying once: %v", err)
			select {
			case <-time.After(2 * time.Second):
				req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/contracts/execute", bytes.NewReader(rawRequest))
				req2.Header.Set("Content-Type", "application/json")
				if cfg.APIKey != "" {
					req2.Header.Set("Authorization", "Bearer "+cfg.APIKey)
				}
				resp, err = client.Do(req2)
			case <-ctx.Done():
				return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "call brain-v3 contract execution failed (context cancelled during retry)", ctx.Err())
			}
		}
		if err != nil {
			recordDiagnostic(ctx, "easymvp_brain.execute", "warning", easyMVPBrainErrorCodeExecuteFailed, "call brain-v3 contract execution failed", map[string]any{
				"contract_kind": cmd.ContractKind,
				"base_url":      baseURL,
				"task_id":       strings.TrimSpace(cmd.TaskID),
				"error":         err.Error(),
			})
			return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "call brain-v3 contract execution failed", err)
		}
	}
	defer resp.Body.Close()

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "read contract response failed", err)
	}
	if resp.StatusCode >= 300 {
		recordDiagnostic(ctx, "easymvp_brain.execute", "warning", easyMVPBrainErrorCodeExecuteFailed, "brain-v3 returned non-success status", map[string]any{
			"contract_kind": cmd.ContractKind,
			"base_url":      baseURL,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"status_code":   resp.StatusCode,
			"body":          strings.TrimSpace(string(rawResponse)),
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "brain-v3 returned non-success status", gerror.Newf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(rawResponse))))
	}

	var execResp easyMVPBrainExecuteResponse
	if err = json.Unmarshal(rawResponse, &execResp); err != nil {
		recordDiagnostic(ctx, "easymvp_brain.decode", "warning", easyMVPBrainErrorCodeDecodeFailed, "decode contract response failed", map[string]any{
			"contract_kind": cmd.ContractKind,
			"base_url":      baseURL,
			"task_id":       strings.TrimSpace(cmd.TaskID),
			"error":         err.Error(),
		})
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeDecodeFailed, "decode contract response failed", err)
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
			recordDiagnostic(ctx, "easymvp_brain.decode", "warning", easyMVPBrainErrorCodeDecodeFailed, "decode contract envelope from summary failed", map[string]any{
				"contract_kind": cmd.ContractKind,
				"base_url":      baseURL,
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
	return nil
}

func buildEasyMVPBrainInstruction(contractKind, instruction string) string {
	instruction = strings.TrimSpace(instruction)
	contractKind = strings.TrimSpace(contractKind)
	if contractKind == "" {
		return instruction
	}
	// brain-v3 /v1/contracts/execute will prepend [contract:xxx]; keep instruction clean.
	return instruction
}

// ─────────────────────────────────────────────────────────────────────────────
// Stage 0: Browser + Verifier run helpers (reuse existing Runtime().StartBrainRun)
// ─────────────────────────────────────────────────────────────────────────────

// BrowserCheckResult holds the outcome of a browser anomaly/understand check.
type BrowserCheckResult struct {
	URL       string   `json:"url"`
	Status    string   `json:"status"`     // passed / failed / skipped / error
	Anomalies []string `json:"anomalies,omitempty"`
	Summary   string   `json:"summary,omitempty"`
}

// VerifierCheck defines a single verification check for the verifier brain.
type VerifierCheck struct {
	Name       string `json:"name"`                 // e.g. "unit_test"
	Type       string `json:"type"`                 // "shell" | "browser" | "file_assert"
	Command    string `json:"command,omitempty"`    // shell command or assertion
	ExpectExit int    `json:"expect_exit,omitempty"`
	ExpectOut  string `json:"expect_output,omitempty"`
}

// VerifierCheckResult holds the outcome of a verifier run.
type VerifierCheckResult struct {
	Status   string            `json:"status"`   // passed / failed / skipped / error
	Checks   []VerifierCheck   `json:"checks,omitempty"`
	Summary  string            `json:"summary,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

// isRetryableBrainError returns true for transient network errors that warrant a retry.
func isRetryableBrainError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "eof") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "no such host") {
		return true
	}
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}
	return false
}


// ─────────────────────────────────────────────────────────────────────────────
// SSE streaming client for brain-v3 /v1/contracts/execute?stream=true
// ─────────────────────────────────────────────────────────────────────────────

// EasyMVPBrainStreamEvent represents a single SSE event from brain-v3.
type EasyMVPBrainStreamEvent struct {
	ExecutionID string          `json:"execution_id"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}

// callEasyMVPBrainExecuteStream calls brain-v3 contracts/execute in SSE streaming mode.
// It returns a channel that yields events until the stream ends or an error occurs.
// The channel is closed when the stream is done. Callers should drain the channel.
func callEasyMVPBrainExecuteStream(ctx context.Context, cfg *EasyMVPBrainClientConfig, baseURL string, cmd EasyMVPBrainExecuteCommand) (<-chan EasyMVPBrainStreamEvent, error) {
	if cfg == nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "easymvp brain config is required", nil)
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "brain serve base URL is empty", nil)
	}
	if strings.TrimSpace(cmd.Instruction) == "" {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeConfigInvalid, "instruction is required", nil)
	}

	payload := map[string]interface{}{
		"brain_kind":    "easymvp",
		"contract_kind": cmd.ContractKind,
		"context_json":  json.RawMessage(cmd.ContextJSON),
		"instruction":   buildEasyMVPBrainInstruction(cmd.ContractKind, cmd.Instruction),
		"max_turns":     cfg.MaxTurns,
		"timeout":       cfg.Timeout.String(),
		"stream":        true,
	}

	rawRequest, err := json.Marshal(payload)
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "marshal contract request failed", err)
	}

	client := &http.Client{Timeout: cfg.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/contracts/execute?stream=true", bytes.NewReader(rawRequest))
	if err != nil {
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "build contract request failed", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		if isRetryableBrainError(err) {
			g.Log().Infof(ctx, "brain-v3 contract stream failed with retryable error, retrying once: %v", err)
			select {
			case <-time.After(2 * time.Second):
				req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/contracts/execute?stream=true", bytes.NewReader(rawRequest))
				req2.Header.Set("Content-Type", "application/json")
				req2.Header.Set("Accept", "text/event-stream")
				if cfg.APIKey != "" {
					req2.Header.Set("Authorization", "Bearer "+cfg.APIKey)
				}
				resp, err = client.Do(req2)
			case <-ctx.Done():
				return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "call brain-v3 contract stream failed (context cancelled during retry)", ctx.Err())
			}
		}
		if err != nil {
			return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "call brain-v3 contract stream failed", err)
		}
	}

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, wrapEasyMVPBrainError(easyMVPBrainErrorCodeExecuteFailed, "brain-v3 returned non-success status", gerror.Newf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body))))
	}

	outCh := make(chan EasyMVPBrainStreamEvent, 16)
	go func() {
		defer close(outCh)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "" {
				continue
			}
			var ev EasyMVPBrainStreamEvent
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				g.Log().Warningf(ctx, "brain-v3 stream event decode failed: %v | raw: %s", err, data)
				continue
			}
			select {
			case outCh <- ev:
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			g.Log().Warningf(ctx, "brain-v3 stream scanner error: %v", err)
		}
	}()

	return outCh, nil
}
