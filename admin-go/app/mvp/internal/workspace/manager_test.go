package workspace

import "testing"

func TestNeedsIsolation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		mode string
		want bool
	}{
		{mode: "aider", want: true},
		{mode: "codex_cli", want: true},
		{mode: "gemini_cli", want: true},
		{mode: "manual", want: false},
		{mode: "", want: false},
	}

	for _, tc := range cases {
		if got := NeedsIsolation(tc.mode); got != tc.want {
			t.Fatalf("NeedsIsolation(%q) = %v, want %v", tc.mode, got, tc.want)
		}
	}
}
