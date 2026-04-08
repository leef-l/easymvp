package provider

import "testing"

func TestDecodeSupportedProtocolsIgnoresVendorTypeFallback(t *testing.T) {
	got := DecodeSupportedProtocols("", "tencent_coding")
	if len(got) != 0 {
		t.Fatalf("DecodeSupportedProtocols() = %#v", got)
	}
}

func TestResolveProtocolPrefersAnthropicBaseURL(t *testing.T) {
	got := ResolveProtocol(Config{
		ProviderType:       "tencent_coding",
		SupportedProtocols: []string{"anthropic", "openai_compatible"},
		BaseURL:            "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1",
	})
	if got != TypeAnthropic {
		t.Fatalf("ResolveProtocol() = %q, want %q", got, TypeAnthropic)
	}
}

func TestResolveProtocolSupportsVendorWithOpenAIOnly(t *testing.T) {
	got := ResolveProtocol(Config{
		ProviderType:       "tencent_coding",
		SupportedProtocols: []string{"openai_compatible"},
		BaseURL:            "https://example.com/v1",
	})
	if got != TypeOpenAICompatible {
		t.Fatalf("ResolveProtocol() = %q, want %q", got, TypeOpenAICompatible)
	}
}

func TestResolveBaseURLForProtocolSupportsTencentCodingOpenAI(t *testing.T) {
	got := ResolveBaseURLForProtocol(Config{
		ProviderType:       "tencent_coding",
		SupportedProtocols: []string{"anthropic", "openai_compatible"},
		BaseURL:            "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1",
	}, TypeOpenAICompatible)
	want := "https://api.lkeap.cloud.tencent.com/coding/v3"
	if got != want {
		t.Fatalf("ResolveBaseURLForProtocol() = %q, want %q", got, want)
	}
}

func TestResolveBaseURLForProtocolSupportsTencentCodingAnthropic(t *testing.T) {
	got := ResolveBaseURLForProtocol(Config{
		ProviderType:       "tencent_coding",
		SupportedProtocols: []string{"anthropic", "openai_compatible"},
		BaseURL:            "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1",
	}, TypeAnthropic)
	want := "https://api.lkeap.cloud.tencent.com/coding/anthropic"
	if got != want {
		t.Fatalf("ResolveBaseURLForProtocol() = %q, want %q", got, want)
	}
}

func TestResolveBaseURLForProtocolStripsGenericAnthropicV1(t *testing.T) {
	got := ResolveBaseURLForProtocol(Config{
		ProviderType: "anthropic",
		BaseURL:      "https://api.anthropic.com/v1",
	}, TypeAnthropic)
	want := "https://api.anthropic.com"
	if got != want {
		t.Fatalf("ResolveBaseURLForProtocol() = %q, want %q", got, want)
	}
}

func TestAnthropicProviderRequestUsesV1Messages(t *testing.T) {
	p := NewAnthropic(Config{
		ProviderType: "tencent_coding",
		BaseURL:      "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1",
	})
	req, err := p.newHTTPRequest(t.Context(), []byte(`{}`))
	if err != nil {
		t.Fatalf("newHTTPRequest() error = %v", err)
	}
	want := "https://api.lkeap.cloud.tencent.com/coding/anthropic/v1/messages"
	if req.URL.String() != want {
		t.Fatalf("newHTTPRequest() URL = %q, want %q", req.URL.String(), want)
	}
}
