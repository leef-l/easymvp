//go:build integration

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func skipIfUnreachable(t *testing.T, url string) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Skipf("service unreachable at %s: %v", url, err)
	}
	resp.Body.Close()
}

func TestIntegration_CoreHealthz(t *testing.T) {
	t.Parallel()
	const url = "http://127.0.0.1:8000/api/v3/system/healthz"
	skipIfUnreachable(t, url)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_BrainHealth(t *testing.T) {
	t.Parallel()
	const url = "http://127.0.0.1:7701/v1/health"
	skipIfUnreachable(t, url)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_ResolveClientConfig(t *testing.T) {
	t.Parallel()
	skipIfUnreachable(t, "http://127.0.0.1:8000/api/v3/system/healthz")

	svc := &sEasyMVPBrain{}
	cfg, err := svc.ResolveClientConfig(context.Background())
	if err != nil {
		t.Fatalf("ResolveClientConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Mode != "local-sidecar" {
		t.Fatalf("expected mode local-sidecar, got %q", cfg.Mode)
	}
	if cfg.LocalEndpoint == "" {
		t.Fatal("expected non-empty LocalEndpoint")
	}
}

func TestIntegration_BrainJSONRPCPing(t *testing.T) {
	t.Parallel()
	skipIfUnreachable(t, "http://127.0.0.1:7701/v1/health")

	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "brain/execute",
		"params": map[string]any{
			"instruction": "ping",
			"budget":      map[string]any{"max_turns": 1},
		},
	}
	raw, _ := json.Marshal(body)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post("http://127.0.0.1:7701/rpc", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("POST rpc failed: %v", err)
	}
	defer resp.Body.Close()

	var rpcResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if rpcResp["jsonrpc"] != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %v", rpcResp["jsonrpc"])
	}
}
