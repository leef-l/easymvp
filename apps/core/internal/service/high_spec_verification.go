package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// HighSpecVerificationChannel implements Engineering 铁律 4:
// final verification must execute on a high-specification environment.
// When enabled and reachable, the system routes verification work to
// this channel instead of falling back to github_actions or manual_review.
type HighSpecVerificationChannel struct {
	enabled         bool
	endpoint        string
	healthCheckPath string
	timeout         time.Duration
}

func newHighSpecVerificationChannelFromConfig(ctx context.Context) *HighSpecVerificationChannel {
	cfg := g.Cfg()
	return &HighSpecVerificationChannel{
		enabled:         cfg.MustGet(ctx, "easymvp.highSpecVerification.enabled", false).Bool(),
		endpoint:        strings.TrimRight(strings.TrimSpace(cfg.MustGet(ctx, "easymvp.highSpecVerification.endpoint", "").String()), "/"),
		healthCheckPath: strings.TrimSpace(cfg.MustGet(ctx, "easymvp.highSpecVerification.healthCheckPath", "/health").String()),
		timeout:         cfg.MustGet(ctx, "easymvp.highSpecVerification.timeout", "10s").Duration(),
	}
}

// Available returns true when the high-spec verification channel is configured
// and responds to health checks within the configured timeout.
func (c *HighSpecVerificationChannel) Available(ctx context.Context) bool {
	if !c.enabled || c.endpoint == "" {
		return false
	}
	url := c.endpoint + c.healthCheckPath
	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// PreferredChannel returns the channel name used in verification contracts.
func (c *HighSpecVerificationChannel) PreferredChannel() string {
	return "high_spec_remote"
}

// deriveVerificationCurrentChannelWithHighSpec returns the current verification channel,
// preferring high_spec_remote when it is available.
func deriveVerificationCurrentChannelWithHighSpec(ctx context.Context, manualReviewRequired bool) string {
	highSpec := newHighSpecVerificationChannelFromConfig(ctx)
	if highSpec.Available(ctx) {
		return highSpec.PreferredChannel()
	}
	if manualReviewRequired {
		return "manual_review"
	}
	return "github_actions"
}
