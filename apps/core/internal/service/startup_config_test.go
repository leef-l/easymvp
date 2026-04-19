package service

import "testing"

func TestNormalizeServerAddress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty uses default", input: "", want: ":8000"},
		{name: "plain port gets colon", input: "8080", want: ":8080"},
		{name: "trimmed port gets colon", input: " 9000 ", want: ":9000"},
		{name: "host and port kept", input: "127.0.0.1:8080", want: "127.0.0.1:8080"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeServerAddress(tt.input); got != tt.want {
				t.Fatalf("normalizeServerAddress(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseBoolOption(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty means true flag", input: "", want: true},
		{name: "true", input: "true", want: true},
		{name: "yes", input: "yes", want: true},
		{name: "off", input: "off", want: false},
		{name: "false", input: "false", want: false},
		{name: "zero", input: "0", want: false},
		{name: "unknown defaults true", input: "unexpected", want: true},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := parseBoolOption(tt.input); got != tt.want {
				t.Fatalf("parseBoolOption(%q) = %t, want %t", tt.input, got, tt.want)
			}
		})
	}
}
