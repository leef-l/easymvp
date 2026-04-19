package service

import (
	"testing"
)

func TestRuntimeResumeCommandUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("EASYMVP_BRAIN_CMD", `"C:\Program Files\Brain\brain.exe"`)
	t.Setenv("EASYMVP_BRAIN_ARGS", `--profile default --config "C:\Users\me\brain config.json"`)

	command, args := runtimeResumeCommand()
	if command != `C:\Program Files\Brain\brain.exe` {
		t.Fatalf("unexpected command: got %q", command)
	}
	if len(args) != 4 {
		t.Fatalf("unexpected arg count: got %d want %d", len(args), 4)
	}
	if args[0] != "--profile" || args[1] != "default" {
		t.Fatalf("unexpected leading args: %#v", args)
	}
	if args[2] != "--config" || args[3] != `C:\Users\me\brain config.json` {
		t.Fatalf("unexpected config args: %#v", args)
	}
}

func TestRuntimeResumeCommandFallsBackToBrainBinary(t *testing.T) {
	t.Setenv("EASYMVP_BRAIN_CMD", "")
	t.Setenv("EASYMVP_BRAIN_ARGS", "")

	command, args := runtimeResumeCommand()
	if command != "brain" {
		t.Fatalf("unexpected fallback command: got %q", command)
	}
	if len(args) != 0 {
		t.Fatalf("unexpected fallback args: %#v", args)
	}
}

func TestSplitCommandArgsPreservesQuotedSegments(t *testing.T) {
	t.Parallel()

	got := splitCommandArgs(`--flag value --path "C:\Program Files\Easy MVP\config.json" --literal 'two words'`)
	if len(got) != 6 {
		t.Fatalf("unexpected arg count: got %d want %d", len(got), 6)
	}
	if got[0] != "--flag" || got[1] != "value" {
		t.Fatalf("unexpected leading args: %#v", got)
	}
	if got[2] != "--path" || got[3] != `C:\Program Files\Easy MVP\config.json` {
		t.Fatalf("unexpected path args: %#v", got)
	}
	if got[4] != "--literal" || got[5] != "two words" {
		t.Fatalf("unexpected literal args: %#v", got)
	}
}

func TestStripWrappedCommandTrimsMatchingQuotes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  string
		expect string
	}{
		{input: `"brain.exe"`, expect: "brain.exe"},
		{input: `'brain.exe'`, expect: "brain.exe"},
		{input: ` brain.exe `, expect: "brain.exe"},
	}

	for _, tc := range cases {
		if got := stripWrappedCommand(tc.input); got != tc.expect {
			t.Fatalf("unexpected normalized command for %q: got %q want %q", tc.input, got, tc.expect)
		}
	}
}
